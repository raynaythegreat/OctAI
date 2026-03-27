// OctAi - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 OctAi contributors

package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/raynaythegreat/ai-business-hq/pkg/config"
)

func mainConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".octai", "config.json")
}

func (a *App) newTeamsPage() tview.Primitive {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)
	table.SetBorder(true).
		SetTitle(" [#A855F7::b] AGENT TEAMS ").
		SetTitleColor(tcell.NewHexColor(0xA855F7)).
		SetBorderColor(tcell.NewHexColor(0x2D1B4E))
	table.SetSelectedStyle(
		tcell.StyleDefault.Background(tcell.NewHexColor(0x1E0F3D)).Foreground(tcell.NewHexColor(0xA855F7)),
	)
	table.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))

	cfgPath := mainConfigPath()

	loadConfig := func() *config.Config {
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			return &config.Config{}
		}
		return cfg
	}

	selectedTeamID := func() string {
		row, _ := table.GetSelection()
		idx := row - 1 // row 0 is header
		cfg := loadConfig()
		teams := cfg.Agents.Teams
		if idx >= 0 && idx < len(teams) {
			return teams[idx].ID
		}
		return ""
	}

	rebuild := func() {
		selID := selectedTeamID()
		table.Clear()

		// Header row
		table.SetCell(0, 0, tview.NewTableCell(" [#A855F7]NAME[-]").SetSelectable(false).SetExpansion(2))
		table.SetCell(0, 1, tview.NewTableCell(" [#A855F7]ORCHESTRATOR[-]").SetSelectable(false).SetExpansion(2))
		table.SetCell(0, 2, tview.NewTableCell(" [#A855F7]MEMBERS[-]").SetSelectable(false))
		table.SetCell(0, 3, tview.NewTableCell(" [#A855F7]TOKEN BUDGET[-]").SetSelectable(false))

		cfg := loadConfig()
		teams := cfg.Agents.Teams
		for i, t := range teams {
			row := i + 1
			name := t.Name
			if name == "" {
				name = t.ID
			}
			budget := "-"
			if t.TokenBudget > 0 {
				budget = strconv.Itoa(t.TokenBudget)
			}
			table.SetCell(row, 0,
				tview.NewTableCell(" "+name).
					SetTextColor(tcell.NewHexColor(0xE8E0F0)).
					SetExpansion(2).
					SetSelectable(true),
			)
			table.SetCell(row, 1,
				tview.NewTableCell(" "+t.OrchestratorID).
					SetTextColor(tcell.NewHexColor(0x7B6F8E)).
					SetExpansion(2).
					SetSelectable(true),
			)
			table.SetCell(row, 2,
				tview.NewTableCell(fmt.Sprintf(" %d", len(t.MemberIDs))).
					SetTextColor(tcell.NewHexColor(0x34D399)).
					SetSelectable(true),
			)
			table.SetCell(row, 3,
				tview.NewTableCell(" "+budget).
					SetTextColor(tcell.NewHexColor(0x7B6F8E)).
					SetSelectable(true),
			)
		}

		if selID != "" {
			for i, t := range teams {
				if t.ID == selID {
					table.Select(i+1, 0)
					return
				}
			}
		}
		if len(teams) > 0 {
			table.Select(1, 0)
		}
	}
	rebuild()

	a.pageRefreshFns["teams"] = rebuild

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'a':
			a.showTeamForm(cfgPath, nil, rebuild)
			return nil
		case 'e':
			row, _ := table.GetSelection()
			cfg := loadConfig()
			teams := cfg.Agents.Teams
			idx := row - 1
			if idx < 0 || idx >= len(teams) {
				return nil
			}
			orig := teams[idx]
			a.showTeamForm(cfgPath, &orig, rebuild)
			return nil
		case 'd':
			row, _ := table.GetSelection()
			cfg := loadConfig()
			teams := cfg.Agents.Teams
			idx := row - 1
			if idx < 0 || idx >= len(teams) {
				return nil
			}
			teamID := teams[idx].ID
			label := teams[idx].Name
			if label == "" {
				label = teamID
			}
			a.confirmDelete(fmt.Sprintf("team %q", label), func() {
				cfg2 := loadConfig()
				newTeams := make([]config.TeamConfig, 0, len(cfg2.Agents.Teams))
				for _, t := range cfg2.Agents.Teams {
					if t.ID != teamID {
						newTeams = append(newTeams, t)
					}
				}
				cfg2.Agents.Teams = newTeams
				if err := config.SaveConfig(cfgPath, cfg2); err != nil {
					a.showError("save failed: " + err.Error())
				}
				rebuild()
			})
			return nil
		}
		return event
	})

	return a.buildShell("teams", table, " [#7B6F8E]a:[-] add  [#7B6F8E]e:[-] edit  [#F87171]d:[-] delete  [#F87171]ESC:[-] back ")
}

func (a *App) showTeamForm(cfgPath string, existing *config.TeamConfig, onDone func()) {
	id := ""
	name := ""
	orchestratorID := ""
	memberIDs := ""
	tokenBudget := ""
	maxConcurrent := ""
	title := " ADD TEAM "

	if existing != nil {
		id = existing.ID
		name = existing.Name
		orchestratorID = existing.OrchestratorID
		memberIDs = strings.Join(existing.MemberIDs, ", ")
		if existing.TokenBudget > 0 {
			tokenBudget = strconv.Itoa(existing.TokenBudget)
		}
		if existing.MaxConcurrent > 0 {
			maxConcurrent = strconv.Itoa(existing.MaxConcurrent)
		}
		title = " EDIT TEAM "
	}

	form := tview.NewForm()
	form.
		AddInputField("ID (auto if empty)", id, 30, nil, func(text string) { id = text }).
		AddInputField("Name", name, 30, nil, func(text string) { name = text }).
		AddInputField("Orchestrator ID", orchestratorID, 30, nil, func(text string) { orchestratorID = text }).
		AddInputField("Member IDs (comma-sep)", memberIDs, 40, nil, func(text string) { memberIDs = text }).
		AddInputField("Token Budget", tokenBudget, 10, func(textToCheck string, _ rune) bool {
			return textToCheck == "" || func() bool { _, e := strconv.Atoi(textToCheck); return e == nil }()
		}, func(text string) { tokenBudget = text }).
		AddInputField("Max Concurrent", maxConcurrent, 10, func(textToCheck string, _ rune) bool {
			return textToCheck == "" || func() bool { _, e := strconv.Atoi(textToCheck); return e == nil }()
		}, func(text string) { maxConcurrent = text }).
		AddButton("SAVE", func() {
			if orchestratorID == "" {
				a.showError("Orchestrator ID is required")
				return
			}
			finalID := id
			if finalID == "" {
				finalID = fmt.Sprintf("team-%d", time.Now().UnixMilli())
			}
			budget := 0
			if tokenBudget != "" {
				if v, err := strconv.Atoi(tokenBudget); err == nil {
					budget = v
				}
			}
			maxC := 0
			if maxConcurrent != "" {
				if v, err := strconv.Atoi(maxConcurrent); err == nil {
					maxC = v
				}
			}
			var members []string
			for _, m := range strings.Split(memberIDs, ",") {
				m = strings.TrimSpace(m)
				if m != "" {
					members = append(members, m)
				}
			}
			team := config.TeamConfig{
				ID:             finalID,
				Name:           name,
				OrchestratorID: orchestratorID,
				MemberIDs:      members,
				TokenBudget:    budget,
				MaxConcurrent:  maxC,
			}
			cfg := func() *config.Config {
				c, err := config.LoadConfig(cfgPath)
				if err != nil {
					return &config.Config{}
				}
				return c
			}()
			if existing != nil {
				for i, t := range cfg.Agents.Teams {
					if t.ID == existing.ID {
						cfg.Agents.Teams[i] = team
						break
					}
				}
			} else {
				cfg.Agents.Teams = append(cfg.Agents.Teams, team)
			}
			if err := config.SaveConfig(cfgPath, cfg); err != nil {
				a.showError("save failed: " + err.Error())
				return
			}
			a.hideModal("team-form")
			onDone()
		}).
		AddButton("CANCEL", func() {
			a.hideModal("team-form")
		})

	form.SetBorder(true).
		SetTitle(" [::b]" + title + " ").
		SetTitleColor(tcell.NewHexColor(0xA855F7)).
		SetBorderColor(tcell.NewHexColor(0x2D1B4E))
	form.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))
	form.SetFieldBackgroundColor(tcell.NewHexColor(0x1A1230))
	form.SetFieldTextColor(tcell.NewHexColor(0xE8E0F0))
	form.SetLabelColor(tcell.NewHexColor(0xE8E0F0))
	form.SetButtonBackgroundColor(tcell.NewHexColor(0x1E0F3D))
	form.SetButtonTextColor(tcell.NewHexColor(0xA855F7))
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.hideModal("team-form")
			return nil
		}
		return event
	})

	a.showModal("team-form", centeredForm(form, 4, 16))
}
