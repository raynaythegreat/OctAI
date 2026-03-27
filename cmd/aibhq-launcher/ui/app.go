// AI Business HQ - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 AI Business HQ contributors

package ui

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	tuicfg "github.com/raynaythegreat/ai-business-hq/cmd/aibhq-launcher/config"
)

// AppMode represents the current operating mode of the TUI.
type AppMode string

const (
	ModePlan  AppMode = "plan"
	ModeBuild AppMode = "build"
)

// App is the root TUI application.
type App struct {
	tapp           *tview.Application
	pages          *tview.Pages
	pageStack      []string
	cfg            *tuicfg.TUIConfig
	configPath     string
	pageRefreshFns map[string]func()
	headerModelTV  *tview.TextView
	headerModeTV   *tview.TextView // shows current Plan/Build mode
	modalOpen      map[string]bool
	currentMode    AppMode // "plan" or "build"

	// OnModelSelected is called when a model is selected in the UI.
	// Can be nil to disable.
	OnModelSelected func(scheme tuicfg.Scheme, user tuicfg.User, modelID string)

	modelCache   map[string][]modelEntry
	modelCacheMu sync.RWMutex
	refreshMu    sync.Mutex
}

// setMode switches between Plan and Build mode and refreshes the mode indicator.
func (a *App) setMode(mode AppMode) {
	a.currentMode = mode
	if a.headerModeTV != nil {
		a.headerModeTV.SetText(a.modeBadge())
	}
}

// toggleMode flips between Plan and Build mode.
func (a *App) toggleMode() {
	if a.currentMode == ModePlan {
		a.setMode(ModeBuild)
	} else {
		a.setMode(ModePlan)
	}
}

// modeBadge returns the colored mode indicator string for the header.
func (a *App) modeBadge() string {
	if a.currentMode == ModePlan {
		return "[#0A0A12:#A855F7:b] PLAN [:-:-] "
	}
	return "[#0A0A12:#F87171:b] BUILD [:-:-] "
}

// cacheKey returns the map key for a (scheme, user) pair.
func cacheKey(schemeName, userName string) string {
	return fmt.Sprintf("%s/%s", schemeName, userName)
}

// cachedModels returns a defensive copy of the cached model list for a user (may be nil).
func (a *App) cachedModels(schemeName, userName string) []modelEntry {
	a.modelCacheMu.RLock()
	defer a.modelCacheMu.RUnlock()
	entries := a.modelCache[cacheKey(schemeName, userName)]
	return append([]modelEntry(nil), entries...)
}

// refreshModelCache fetches models for every user in the config concurrently.
// Serialized by refreshMu so concurrent calls don't race on the cache map.
// When all fetches complete it calls onDone via QueueUpdateDraw.
func (a *App) refreshModelCache(onDone func()) {
	go func() {
		a.refreshMu.Lock()
		defer a.refreshMu.Unlock()

		users := a.cfg.Provider.Users
		schemes := a.cfg.Provider.Schemes

		schemeURL := make(map[string]string, len(schemes))
		for _, s := range schemes {
			schemeURL[s.Name] = s.BaseURL
		}

		var wg sync.WaitGroup
		for _, u := range users {
			baseURL, ok := schemeURL[u.Scheme]
			if !ok || baseURL == "" {
				continue
			}
			if u.Key == "" {
				a.modelCacheMu.Lock()
				if a.modelCache == nil {
					a.modelCache = make(map[string][]modelEntry)
				}
				a.modelCache[cacheKey(u.Scheme, u.Name)] = nil
				a.modelCacheMu.Unlock()
				continue
			}
			wg.Add(1)
			bURL := baseURL
			go func() {
				defer wg.Done()
				entries, err := fetchModels(bURL, u.Key)
				a.modelCacheMu.Lock()
				if a.modelCache == nil {
					a.modelCache = make(map[string][]modelEntry)
				}
				if err != nil || len(entries) == 0 {
					a.modelCache[cacheKey(u.Scheme, u.Name)] = nil
				} else {
					a.modelCache[cacheKey(u.Scheme, u.Name)] = entries
				}
				a.modelCacheMu.Unlock()
			}()
		}
		wg.Wait()

		if onDone != nil {
			a.tapp.QueueUpdateDraw(onDone)
		}
	}()
}

// New creates and wires up the TUI application.
func New(cfg *tuicfg.TUIConfig, configPath string) *App {
	// OpenClaw Warm Dark Theme
	tview.Styles.PrimitiveBackgroundColor    = tcell.NewHexColor(0x0A0A12) // Near black
	tview.Styles.ContrastBackgroundColor     = tcell.NewHexColor(0x12101F) // Panel bg
	tview.Styles.MoreContrastBackgroundColor = tcell.NewHexColor(0x1A1230) // Card bg
	tview.Styles.BorderColor                 = tcell.NewHexColor(0x2D1B4E) // Dark gray border
	tview.Styles.TitleColor                  = tcell.NewHexColor(0xA855F7) // Amber/gold
	tview.Styles.GraphicsColor               = tcell.NewHexColor(0xF87171) // Orange
	tview.Styles.PrimaryTextColor            = tcell.NewHexColor(0xE8E0F0) // Warm cream
	tview.Styles.SecondaryTextColor          = tcell.NewHexColor(0xA855F7) // Amber
	tview.Styles.TertiaryTextColor           = tcell.NewHexColor(0x34D399) // Mint green
	tview.Styles.InverseTextColor            = tcell.NewHexColor(0x0A0A12) // Near black
	tview.Styles.ContrastSecondaryTextColor  = tcell.NewHexColor(0xF87171) // Orange

	a := &App{
		tapp:           tview.NewApplication(),
		pages:          tview.NewPages(),
		pageStack:      []string{},
		cfg:            cfg,
		configPath:     configPath,
		pageRefreshFns: make(map[string]func()),
		modalOpen:      make(map[string]bool),
		currentMode:    ModeBuild, // default to Build mode
	}

	a.tapp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Tab toggles between Plan and Build mode globally.
		if event.Key() == tcell.KeyTab {
			a.toggleMode()
			a.tapp.QueueUpdateDraw(nil)
			return nil
		}
		// Number keys jump directly to sections (only when no modal is open and no form active)
		if len(a.modalOpen) == 0 {
			switch event.Rune() {
			case '1':
				if len(a.pageStack) > 0 {
					// Navigate to home: pop back to home
					for len(a.pageStack) > 1 {
						popped := a.pageStack[len(a.pageStack)-1]
						a.pageStack = a.pageStack[:len(a.pageStack)-1]
						a.pages.RemovePage(popped)
					}
					a.pages.SwitchToPage("home")
					if fn, ok := a.pageRefreshFns["home"]; ok { fn() }
				}
				return nil
			case '2':
				a.navigateTo("gateway", a.newGatewayPage())
				return nil
			case '3':
				a.navigateTo("schemes", a.newSchemesPage())
				return nil
			case '4':
				a.navigateTo("channels", a.newChannelsPage())
				return nil
			case '5':
				a.navigateTo("teams", a.newTeamsPage())
				return nil
			case '6':
				a.navigateTo("skills", a.newSkillsPage())
				return nil
			case '7':
				a.navigateTo("tools", a.newToolsPage())
				return nil
			case '8':
				a.navigateTo("logs", a.newLogsPage())
				return nil
			}
		}
		if event.Key() == tcell.KeyEscape {
			if len(a.modalOpen) > 0 {
				return event
			}
			return a.goBack()
		}
		return event
	})

	a.buildPages()
	return a
}

// Run starts the TUI event loop.
func (a *App) Run() error {
	return a.tapp.SetRoot(a.pages, true).EnableMouse(true).Run()
}

func (a *App) buildPages() {
	a.pages.AddPage("home", a.newHomePage(), true, true)
	a.pageStack = []string{"home"}
}

func (a *App) navigateTo(name string, page tview.Primitive) {
	a.pages.RemovePage(name)
	a.pages.AddPage(name, page, true, false)
	a.pageStack = append(a.pageStack, name)
	a.pages.SwitchToPage(name)
}

func (a *App) goBack() *tcell.EventKey {
	if len(a.pageStack) <= 1 {
		return nil
	}
	popped := a.pageStack[len(a.pageStack)-1]
	a.pageStack = a.pageStack[:len(a.pageStack)-1]
	a.pages.RemovePage(popped)
	prev := a.pageStack[len(a.pageStack)-1]
	if fn, ok := a.pageRefreshFns[prev]; ok {
		fn()
	}
	if prev == "home" && a.headerModelTV != nil {
		a.headerModelTV.SetText(a.cfg.CurrentModelLabel() + "  ")
	}
	a.pages.SwitchToPage(prev)
	return nil
}

func (a *App) showModal(name string, primitive tview.Primitive) {
	a.modalOpen[name] = true
	a.pages.AddPage(name, primitive, true, true)
}

func (a *App) hideModal(name string) {
	delete(a.modalOpen, name)
	a.pages.HidePage(name)
	a.pages.RemovePage(name)
}

func (a *App) save() {
	if err := tuicfg.Save(a.configPath, a.cfg); err != nil {
		a.showError("save failed: " + err.Error())
	}
}

func (a *App) showError(msg string) {
	modal := tview.NewModal().
		SetText(" [red::b]ERROR[-::-]\n\n" + msg).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			a.hideModal("error")
		})
	modal.SetBackgroundColor(tcell.NewHexColor(0x1A1230))
	modal.SetTextColor(tcell.NewHexColor(0xE8E0F0))
	modal.SetButtonBackgroundColor(tcell.NewHexColor(0xF87171)) // coral red
	modal.SetButtonTextColor(tcell.NewHexColor(0x0A0A12))
	a.showModal("error", modal)
}

func (a *App) confirmDelete(label string, onConfirm func()) {
	modal := tview.NewModal().
		SetText(" [red::b]DELETE WARNING[-::-]\n\nDelete " + label + "?\n[gray]This action cannot be undone.[-]").
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			a.hideModal("confirm-delete")
			if buttonLabel == "Delete" {
				onConfirm()
			}
		})
	modal.SetBackgroundColor(tcell.NewHexColor(0x1A1230))
	modal.SetTextColor(tcell.NewHexColor(0xE8E0F0))
	modal.SetButtonBackgroundColor(tcell.NewHexColor(0xF87171)) // coral red
	modal.SetButtonTextColor(tcell.NewHexColor(0x0A0A12))
	a.showModal("confirm-delete", modal)
}

func centeredForm(form *tview.Form, widthPct, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(form, height, 1, true).
			AddItem(tview.NewBox(), 0, 1, false), 0, widthPct, true).
		AddItem(tview.NewBox(), 0, 1, false)
}

func hintBar(text string) *tview.TextView {
	tv := tview.NewTextView().
		SetText(text).
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.NewHexColor(0x7B6F8E)) // muted gray
	tv.SetBackgroundColor(tcell.NewHexColor(0x12101F))
	return tv
}

func (a *App) buildShell(pageID string, content tview.Primitive, hint string) tview.Primitive {
	// Model display (right side of header)
	var modelTV *tview.TextView
	if pageID == "home" {
		if a.headerModelTV == nil {
			a.headerModelTV = tview.NewTextView()
			a.headerModelTV.SetTextAlign(tview.AlignRight).
				SetTextColor(tcell.NewHexColor(0x7B6F8E)).
				SetDynamicColors(true).
				SetBackgroundColor(tcell.NewHexColor(0x0A0A12))
		}
		modelTV = a.headerModelTV
		modelTV.SetText(a.cfg.CurrentModelLabel() + "  ")
	} else {
		modelTV = tview.NewTextView()
		modelTV.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))
		modelTV.SetTextColor(tcell.NewHexColor(0x7B6F8E))
		modelTV.SetText(a.cfg.CurrentModelLabel() + "  ")
	}

	// Mode badge
	if a.headerModeTV == nil {
		a.headerModeTV = tview.NewTextView().SetDynamicColors(true)
		a.headerModeTV.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))
	}
	a.headerModeTV.SetText(a.modeBadge())

	// Header left — OpenClaw style: bold amber app name
	headerLeft := tview.NewTextView().
		SetText("  [#A855F7::b]OCTAI[-]  [#2D1B4E]·[-]").
		SetDynamicColors(true).
		SetBackgroundColor(tcell.NewHexColor(0x0A0A12))

	header := tview.NewFlex().
		AddItem(headerLeft, 0, 2, false).
		AddItem(a.headerModeTV, 10, 0, false).
		AddItem(modelTV, 0, 1, false)

	// Sidebar — OpenClaw uses → cursor, dim inactive items
	sidebar := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)
	sidebar.SetBorder(false)
	sidebar.SetBackgroundColor(tcell.NewHexColor(0x12101F))

	activePrefix   := "[#A855F7::b]→  "   // amber arrow
	activeSuffix   := "[-]"
	inactivePrefix := "[#7B6F8E]   "        // dim gray
	inactiveSuffix := "[-]"

	menuItem := func(id, label string) string {
		active := pageID == id ||
			(id == "model" && (pageID == "schemes" || pageID == "users" || pageID == "models"))
		if active {
			return activePrefix + label + activeSuffix + "\n\n"
		}
		return inactivePrefix + label + inactiveSuffix + "\n\n"
	}

	sbText := "\n"
	sbText += menuItem("home",     "HOME")
	sbText += menuItem("gateway",  "GATEWAY")
	sbText += menuItem("model",    "MODEL")
	sbText += menuItem("channels", "CHANNELS")
	sbText += "[#2D1B4E]──────────────────────[-]\n\n"
	sbText += menuItem("teams",    "TEAMS")
	sbText += menuItem("skills",   "SKILLS")
	sbText += menuItem("tools",    "TOOLS")
	sbText += menuItem("logs",     "LOGS")

	sidebar.SetText(sbText)

	footer := hintBar(hint)

	grid := tview.NewGrid().
		SetRows(1, 0, 1).
		SetColumns(24, 0).
		AddItem(header,  0, 0, 1, 2, 0, 0, false).
		AddItem(sidebar, 1, 0, 1, 1, 0, 0, false).
		AddItem(content, 1, 1, 1, 1, 0, 0, true).
		AddItem(footer,  2, 0, 1, 2, 0, 0, false)

	return grid
}
