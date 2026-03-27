// AI Business HQ - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 AI Business HQ contributors

package ui

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) newHomePage() tview.Primitive {
	// ── Status Cards ──────────────────────────────────────────
	makeCard := func(title, value string, valueColor string) *tview.TextView {
		tv := tview.NewTextView().
			SetDynamicColors(true).
			SetText(fmt.Sprintf("[#7B6F8E]%s[-]\n[%s::b]%s[-]", title, valueColor, value))
		tv.SetBorder(true).
			SetBorderColor(tcell.NewHexColor(0x2D1B4E)).
			SetBackgroundColor(tcell.NewHexColor(0x12101F))
		return tv
	}

	gatewayCard := makeCard("GATEWAY", "CHECKING...", "#7B6F8E")
	modelCard   := makeCard("MODEL", a.cfg.CurrentModelLabel(), "#A855F7")
	modeCard    := makeCard("MODE", string(a.currentMode), "#34D399")
	versionCard := makeCard("VERSION", "v"+"1.0.0", "#1E0F3D")

	cardRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(gatewayCard, 0, 1, false).
		AddItem(tview.NewBox().SetBackgroundColor(tcell.NewHexColor(0x0A0A12)), 1, 0, false).
		AddItem(modelCard, 0, 1, false).
		AddItem(tview.NewBox().SetBackgroundColor(tcell.NewHexColor(0x0A0A12)), 1, 0, false).
		AddItem(modeCard, 0, 1, false).
		AddItem(tview.NewBox().SetBackgroundColor(tcell.NewHexColor(0x0A0A12)), 1, 0, false).
		AddItem(versionCard, 0, 1, false)

	// ── Action List ────────────────────────────────────────────
	actions := tview.NewList()
	actions.SetBorder(true).
		SetTitle(" [#A855F7::b] QUICK ACTIONS ").
		SetTitleColor(tcell.NewHexColor(0xA855F7)).
		SetBorderColor(tcell.NewHexColor(0x2D1B4E))
	actions.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))
	actions.SetMainTextColor(tcell.NewHexColor(0xE8E0F0))
	actions.SetSecondaryTextColor(tcell.NewHexColor(0x7B6F8E))
	actions.SetSelectedStyle(
		tcell.StyleDefault.Background(tcell.NewHexColor(0x1E0F3D)).Foreground(tcell.NewHexColor(0xA855F7)),
	)
	actions.SetHighlightFullLine(true)

	actions.AddItem("[#34D399::b]▶  LAUNCH CHAT[-]", "Start an interactive AI agent session", 'c', func() {
		if a.currentMode == ModePlan {
			a.showError("[#A855F7::b]PLAN MODE[-]\n\nSwitch to [#F87171::b]BUILD[-] mode (Tab) to launch the agent.")
			return
		}
		a.tapp.Suspend(func() {
			cmd := exec.Command("aibhq", "agent")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			_ = cmd.Run()
		})
	})
	actions.AddItem("[#F87171::b]■  QUIT SYSTEM[-]", "Exit OctAi Launcher", 'q', func() {
		a.tapp.Stop()
	})

	// ── Outer Layout ───────────────────────────────────────────
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewBox().SetBackgroundColor(tcell.NewHexColor(0x0A0A12)), 1, 0, false).
		AddItem(cardRow, 6, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(tcell.NewHexColor(0x0A0A12)), 1, 0, false).
		AddItem(actions, 0, 1, true)

	// ── Live Gateway Status Update ─────────────────────────────
	updateGateway := func() {
		s := getGatewayStatus()
		if s.running {
			gatewayCard.SetText(fmt.Sprintf("[#7B6F8E]GATEWAY[-]\n[#34D399::b]RUNNING  PID:%-6d[-]", s.pid))
		} else {
			gatewayCard.SetText("[#7B6F8E]GATEWAY[-]\n[#F87171::b]STOPPED[-]")
		}
	}
	updateGateway()

	// Update mode card live when Tab is pressed
	a.pageRefreshFns["home"] = func() {
		updateGateway()
		if a.currentMode == ModePlan {
			modeCard.SetText("[#7B6F8E]MODE[-]\n[#A855F7::b]PLAN[-]")
		} else {
			modeCard.SetText("[#7B6F8E]MODE[-]\n[#F87171::b]BUILD[-]")
		}
		modelCard.SetText(fmt.Sprintf("[#7B6F8E]MODEL[-]\n[#A855F7::b]%s[-]", a.cfg.CurrentModelLabel()))
	}

	// Background ticker to refresh gateway status every 3 seconds
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.tapp.QueueUpdateDraw(updateGateway)
			case <-done:
				return
			}
		}
	}()

	// Stop ticker when leaving home (via input capture on the flex)
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event
	})
	_ = done // ticker runs until app exits (home is always in stack)

	return a.buildShell(
		"home",
		flex,
		" [#7B6F8E]1-8:[-] nav  [#7B6F8E]c:[-] chat  [#A855F7]Tab:[-] mode  [#F87171]q:[-] quit ",
	)
}
