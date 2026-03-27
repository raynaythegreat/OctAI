// OctAi - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 OctAi contributors

package ui

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func logFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	primary := filepath.Join(home, ".aibhq", "logs", "gateway.log")
	if _, err := os.Stat(primary); err == nil {
		return primary
	}
	return filepath.Join(home, ".aibhq", "gateway.log")
}

func readLastLines(path string, n int) string {
	f, err := os.Open(path)
	if err != nil {
		return "[#7B6F8E](log file not found or cannot be read)[-]"
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

func (a *App) newLogsPage() tview.Primitive {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(false)
	tv.SetBorder(true).
		SetTitle(" [#A855F7::b] GATEWAY LOGS ").
		SetTitleColor(tcell.NewHexColor(0xA855F7)).
		SetBorderColor(tcell.NewHexColor(0x2D1B4E))
	tv.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))
	tv.SetTextColor(tcell.NewHexColor(0xE8E0F0))

	loadLogs := func() {
		content := readLastLines(logFilePath(), 200)
		tv.SetText(content)
		tv.ScrollToEnd()
	}
	loadLogs()

	a.pageRefreshFns["logs"] = loadLogs

	tv.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'r':
			loadLogs()
			return nil
		case 'c':
			a.confirmDelete("all gateway logs", func() {
				path := logFilePath()
				if err := os.Truncate(path, 0); err != nil {
					// try creating an empty file if truncate fails
					f, cerr := os.Create(path)
					if cerr != nil {
						a.showError("clear failed: " + err.Error())
						return
					}
					f.Close()
				}
				loadLogs()
			})
			return nil
		}
		return event
	})

	return a.buildShell("logs", tv, " [#7B6F8E]r:[-] refresh  [#F87171]c:[-] clear  [#F87171]ESC:[-] back ")
}
