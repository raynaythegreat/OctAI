// OctAi - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 OctAi contributors

package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	tuicfg "github.com/raynaythegreat/ai-business-hq/cmd/aibhq-launcher/config"
)

func (a *App) newUsersPage(schemeName string) tview.Primitive {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)
	table.SetBorder(true).
		SetTitle(fmt.Sprintf(" [#A855F7::b] USERS · %s ", schemeName)).
		SetTitleColor(tcell.NewHexColor(0xA855F7)).
		SetBorderColor(tcell.NewHexColor(0x2D1B4E))
	table.SetSelectedStyle(
		tcell.StyleDefault.Background(tcell.NewHexColor(0x1E0F3D)).Foreground(tcell.NewHexColor(0xA855F7)),
	)
	table.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))

	visibleUsers := func() []tuicfg.User {
		var out []tuicfg.User
		for _, u := range a.cfg.Provider.Users {
			if u.Scheme == schemeName {
				out = append(out, u)
			}
		}
		return out
	}

	findUserGlobalIdx := func(userName string) int {
		for i, u := range a.cfg.Provider.Users {
			if u.Scheme == schemeName && u.Name == userName {
				return i
			}
		}
		return -1
	}

	rowToVisIdx := func(row int) int { return row / 2 }

	selectedUserName := func() string {
		row, _ := table.GetSelection()
		users := visibleUsers()
		visIdx := rowToVisIdx(row)
		if visIdx >= 0 && visIdx < len(users) {
			return users[visIdx].Name
		}
		return ""
	}

	rebuild := func() {
		selName := selectedUserName()
		table.Clear()
		users := visibleUsers()
		for i, u := range users {
			nameRow := i * 2
			detailRow := nameRow + 1

			table.SetCell(nameRow, 0,
				tview.NewTableCell(" "+u.Name).
					SetTextColor(tcell.NewHexColor(0xE8E0F0)).
					SetExpansion(1).
					SetSelectable(true),
			)
			table.SetCell(nameRow, 1,
				tview.NewTableCell("").
					SetSelectable(false),
			)

			models := a.cachedModels(schemeName, u.Name)
			var detailText string
			if len(models) > 0 {
				detailText = fmt.Sprintf("  [#34D399]%d models available[-]", len(models))
			} else {
				detailText = "  [#F87171]Inactive / No Access[-]"
			}
			table.SetCell(detailRow, 0,
				tview.NewTableCell(detailText).
					SetTextColor(tcell.NewHexColor(0x7B6F8E)).
					SetExpansion(1).
					SetSelectable(false),
			)
			table.SetCell(detailRow, 1,
				tview.NewTableCell("[#A855F7]"+u.Type+"  ").
					SetAlign(tview.AlignRight).
					SetSelectable(false),
			)
		}
		if selName != "" {
			for i, u := range users {
				if u.Name == selName {
					table.Select(i*2, 0)
					return
				}
			}
		}
		if table.GetRowCount() > 0 {
			table.Select(0, 0)
		}
	}
	rebuild()

	a.refreshModelCache(rebuild)
	a.pageRefreshFns["users"] = func() { a.refreshModelCache(rebuild) }

	table.SetSelectedFunc(func(row, _ int) {
		visIdx := rowToVisIdx(row)
		users := visibleUsers()
		if visIdx < 0 || visIdx >= len(users) {
			return
		}
		uName := users[visIdx].Name
		scheme := a.cfg.Provider.SchemeByName(schemeName)
		if scheme == nil {
			a.showError(fmt.Sprintf("Scheme %q not found", schemeName))
			return
		}
		a.navigateTo("models", a.newModelsPage(schemeName, uName, scheme.BaseURL))
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()
		visIdx := rowToVisIdx(row)
		users := visibleUsers()
		switch event.Rune() {
		case 'a':
			a.showUserForm(schemeName, nil, func(u tuicfg.User) {
				a.cfg.Provider.Users = append(a.cfg.Provider.Users, u)
				a.save()
				a.refreshModelCache(rebuild)
			})
			return nil
		case 'e':
			if visIdx < 0 || visIdx >= len(users) {
				return nil
			}
			origName := users[visIdx].Name
			orig := a.cfg.Provider.Users[findUserGlobalIdx(origName)]
			a.showUserForm(schemeName, &orig, func(u tuicfg.User) {
				cfgIdx := findUserGlobalIdx(origName)
				if cfgIdx < 0 {
					a.showError(fmt.Sprintf("User %q no longer exists", origName))
					return
				}
				a.cfg.Provider.Users[cfgIdx] = u
				a.save()
				a.refreshModelCache(func() {
					rebuild()
					for i, usr := range visibleUsers() {
						if usr.Name == u.Name {
							table.Select(i*2, 0)
							break
						}
					}
				})
			})
			return nil
		case 'd':
			if visIdx < 0 || visIdx >= len(users) {
				return nil
			}
			uName := users[visIdx].Name
			a.confirmDelete(fmt.Sprintf("user %q", uName), func() {
				cfgIdx := findUserGlobalIdx(uName)
				if cfgIdx < 0 {
					return
				}
				all := a.cfg.Provider.Users
				a.cfg.Provider.Users = append(all[:cfgIdx], all[cfgIdx+1:]...)
				a.save()
				a.refreshModelCache(rebuild)
			})
			return nil
		}
		return event
	})

	return a.buildShell(
		"users",
		table,
		" [#7B6F8E]a:[-] add  [#7B6F8E]e:[-] edit  [#F87171]d:[-] delete  [#7B6F8E]Enter:[-] models  [#F87171]ESC:[-] back ",
	)
}

func (a *App) showUserForm(schemeName string, existing *tuicfg.User, onSave func(tuicfg.User)) {
	name := ""
	userType := "key"
	key := ""
	title := " [#A855F7::b] ADD USER "

	if existing != nil {
		name = existing.Name
		userType = existing.Type
		key = existing.Key
		title = " [#A855F7::b] EDIT USER "
	}

	typeOptions := []string{"key", "OAuth"}
	typeIdx := 0
	for i, t := range typeOptions {
		if t == userType {
			typeIdx = i
			break
		}
	}

	form := tview.NewForm()
	form.
		AddInputField("Name", name, 20, nil, func(text string) { name = text }).
		AddDropDown("Type", typeOptions, typeIdx, func(option string, _ int) { userType = option }).
		AddPasswordField("Key", key, 28, '*', func(text string) { key = text }).
		AddButton("SAVE", func() {
			if name == "" {
				a.showError("Name is required")
				return
			}
			if existing == nil {
				for _, u := range a.cfg.Provider.Users {
					if u.Scheme == schemeName && u.Name == name {
						a.showError(fmt.Sprintf("User name %q already exists for this scheme", name))
						return
					}
				}
			}
			a.hideModal("user-form")
			onSave(tuicfg.User{Name: name, Scheme: schemeName, Type: userType, Key: key})
		}).
		AddButton("CANCEL", func() {
			a.hideModal("user-form")
		})

	form.SetBorder(true).
		SetTitle(title).
		SetTitleColor(tcell.NewHexColor(0xA855F7)).
		SetBorderColor(tcell.NewHexColor(0x2D1B4E))
	form.SetBackgroundColor(tcell.NewHexColor(0x12101F))
	form.SetFieldBackgroundColor(tcell.NewHexColor(0x1A1230))
	form.SetFieldTextColor(tcell.NewHexColor(0xE8E0F0))
	form.SetLabelColor(tcell.NewHexColor(0xE8E0F0))
	form.SetButtonBackgroundColor(tcell.NewHexColor(0x1E0F3D))
	form.SetButtonTextColor(tcell.NewHexColor(0xA855F7))
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.hideModal("user-form")
			return nil
		}
		return event
	})

	a.showModal("user-form", centeredForm(form, 4, 13))
}
