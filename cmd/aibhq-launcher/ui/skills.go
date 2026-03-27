// AI Business HQ - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 AI Business HQ contributors

package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func skillsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".aibhq", "workspace", "skills")
}

func (a *App) newSkillsPage() tview.Primitive {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)
	table.SetBorder(true).
		SetTitle(" [#A855F7::b] AGENT SKILLS ").
		SetTitleColor(tcell.NewHexColor(0xA855F7)).
		SetBorderColor(tcell.NewHexColor(0x2D1B4E))
	table.SetSelectedStyle(
		tcell.StyleDefault.Background(tcell.NewHexColor(0x1E0F3D)).Foreground(tcell.NewHexColor(0xA855F7)),
	)
	table.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))

	type skillEntry struct {
		name string
		path string
	}
	var skills []skillEntry

	rebuild := func() {
		table.Clear()
		skills = skills[:0]

		// Header row
		table.SetCell(0, 0, tview.NewTableCell(" [#A855F7]NAME[-]").SetSelectable(false).SetExpansion(2))
		table.SetCell(0, 1, tview.NewTableCell(" [#A855F7]PATH[-]").SetSelectable(false).SetExpansion(3))

		dir := skillsDir()
		entries, err := os.ReadDir(dir)
		if err != nil {
			table.SetCell(1, 0,
				tview.NewTableCell(fmt.Sprintf(" [#7B6F8E]%s[-]", err.Error())).
					SetSelectable(false).SetExpansion(1),
			)
			return
		}

		row := 1
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			skillPath := filepath.Join(dir, e.Name())
			skills = append(skills, skillEntry{name: e.Name(), path: skillPath})
			table.SetCell(row, 0,
				tview.NewTableCell(" "+e.Name()).
					SetTextColor(tcell.NewHexColor(0xE8E0F0)).
					SetExpansion(2).
					SetSelectable(true),
			)
			table.SetCell(row, 1,
				tview.NewTableCell(" "+skillPath).
					SetTextColor(tcell.NewHexColor(0x7B6F8E)).
					SetExpansion(3).
					SetSelectable(true),
			)
			row++
		}

		if len(skills) == 0 {
			table.SetCell(1, 0,
				tview.NewTableCell(" [#7B6F8E](no skills found)[-]").
					SetSelectable(false).SetExpansion(1),
			)
		} else if table.GetRowCount() > 1 {
			table.Select(1, 0)
		}
	}
	rebuild()

	a.pageRefreshFns["skills"] = rebuild

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'r':
			rebuild()
			return nil
		case 'd':
			row, _ := table.GetSelection()
			idx := row - 1
			if idx < 0 || idx >= len(skills) {
				return nil
			}
			sk := skills[idx]
			a.confirmDelete(fmt.Sprintf("skill %q", sk.name), func() {
				if err := os.RemoveAll(sk.path); err != nil {
					a.showError("delete failed: " + err.Error())
				}
				rebuild()
			})
			return nil
		}
		return event
	})

	return a.buildShell("skills", table, " [#7B6F8E]r:[-] refresh  [#F87171]d:[-] delete  [#F87171]ESC:[-] back ")
}
