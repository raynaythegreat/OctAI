// OctAi - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 OctAi contributors

package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/raynaythegreat/ai-business-hq/pkg/config"
)

type toolEntry struct {
	name       string
	getEnabled func(cfg *config.Config) bool
	setEnabled func(cfg *config.Config, v bool)
}

func toolEntries() []toolEntry {
	return []toolEntry{
		{
			name:       "append_file",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.AppendFile.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.AppendFile.Enabled = v },
		},
		{
			name:       "edit_file",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.EditFile.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.EditFile.Enabled = v },
		},
		{
			name:       "find_skills",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.FindSkills.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.FindSkills.Enabled = v },
		},
		{
			name:       "install_skill",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.InstallSkill.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.InstallSkill.Enabled = v },
		},
		{
			name:       "list_dir",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.ListDir.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.ListDir.Enabled = v },
		},
		{
			name:       "message",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.Message.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.Message.Enabled = v },
		},
		{
			name:       "send_file",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.SendFile.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.SendFile.Enabled = v },
		},
		{
			name:       "spawn",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.Spawn.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.Spawn.Enabled = v },
		},
		{
			name:       "spawn_status",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.SpawnStatus.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.SpawnStatus.Enabled = v },
		},
		{
			name:       "subagent",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.Subagent.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.Subagent.Enabled = v },
		},
		{
			name:       "web_fetch",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.WebFetch.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.WebFetch.Enabled = v },
		},
		{
			name:       "write_file",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.WriteFile.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.WriteFile.Enabled = v },
		},
		{
			name:       "i2c",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.I2C.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.I2C.Enabled = v },
		},
		{
			name:       "spi",
			getEnabled: func(cfg *config.Config) bool { return cfg.Tools.SPI.Enabled },
			setEnabled: func(cfg *config.Config, v bool) { cfg.Tools.SPI.Enabled = v },
		},
	}
}

func (a *App) newToolsPage() tview.Primitive {
	cfgPath := mainConfigPath()

	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)
	table.SetBorder(true).
		SetTitle(" [#A855F7::b] TOOL ACCESS ").
		SetTitleColor(tcell.NewHexColor(0xA855F7)).
		SetBorderColor(tcell.NewHexColor(0x2D1B4E))
	table.SetSelectedStyle(
		tcell.StyleDefault.Background(tcell.NewHexColor(0x1E0F3D)).Foreground(tcell.NewHexColor(0xA855F7)),
	)
	table.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))

	entries := toolEntries()

	rebuild := func() {
		table.Clear()

		table.SetCell(0, 0, tview.NewTableCell(" [#A855F7]TOOL[-]").SetSelectable(false).SetExpansion(2))
		table.SetCell(0, 1, tview.NewTableCell(" [#A855F7]STATUS[-]").SetSelectable(false))

		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			cfg = &config.Config{}
		}

		for i, e := range entries {
			row := i + 1
			status := "[#F87171]✗ disabled[-]"
			if e.getEnabled(cfg) {
				status = "[#34D399]✓ enabled[-]"
			}
			table.SetCell(row, 0,
				tview.NewTableCell(" "+e.name).
					SetTextColor(tcell.NewHexColor(0xE8E0F0)).
					SetExpansion(2).
					SetSelectable(true),
			)
			table.SetCell(row, 1,
				tview.NewTableCell(" "+status).
					SetSelectable(true),
			)
		}
		if len(entries) > 0 {
			table.Select(1, 0)
		}
	}
	rebuild()

	toggle := func() {
		row, _ := table.GetSelection()
		idx := row - 1
		if idx < 0 || idx >= len(entries) {
			return
		}
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			cfg = &config.Config{}
		}
		e := entries[idx]
		e.setEnabled(cfg, !e.getEnabled(cfg))
		if err := config.SaveConfig(cfgPath, cfg); err != nil {
			a.showError("save failed: " + err.Error())
			return
		}
		// Refresh just this row
		status := "[#F87171]✗ disabled[-]"
		if e.getEnabled(cfg) {
			status = "[#34D399]✓ enabled[-]"
		}
		table.SetCell(row, 1, tview.NewTableCell(" "+status).SetSelectable(true))
	}

	table.SetSelectedFunc(func(_, _ int) {
		toggle()
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			toggle()
			return nil
		}
		return event
	})

	a.pageRefreshFns["tools"] = rebuild

	return a.buildShell("tools", table, " [#7B6F8E]Enter/Space:[-] toggle  [#F87171]ESC:[-] back ")
}
