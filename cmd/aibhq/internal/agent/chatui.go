package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	pkgagent "github.com/raynaythegreat/ai-business-hq/pkg/agent"
)

// cardWidth is the total display width of a tool execution card (including │ borders).
const cardWidth = 50

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type chatUI struct {
	app    *tview.Application
	pages  *tview.Pages
	layout *tview.Flex

	header     *tview.TextView
	chatLog    *tview.TextView
	statusLine *tview.TextView
	footer     *tview.TextView
	input      *tview.InputField

	modelName  string
	sessionKey string
	agentLoop  *pkgagent.AgentLoop

	mu         sync.Mutex
	busy       bool
	spinIdx    int
	startTime  time.Time
	lastTool   string
	history    []string
	histIdx    int
	ctx        context.Context // set during run(), used by sendToLoop
	skillNames map[string]bool // lowercase skill name → true
}

func newChatUI(modelName, sessionKey string, agentLoop *pkgagent.AgentLoop) *chatUI {
	c := &chatUI{
		modelName:  modelName,
		sessionKey: sessionKey,
		agentLoop:  agentLoop,
	}
	c.buildLayout()
	return c
}

func (c *chatUI) shortSession() string {
	if len(c.sessionKey) > 8 {
		return c.sessionKey[:8]
	}
	return c.sessionKey
}

func (c *chatUI) buildLayout() {
	// ── Header ─────────────────────────────────────────────────
	c.header = tview.NewTextView().SetDynamicColors(true)
	c.header.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))

	// ── Chat log ────────────────────────────────────────────────
	c.chatLog = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	c.chatLog.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))
	c.chatLog.SetTextColor(tcell.NewHexColor(0xE8E0F0))

	// ── Status line ─────────────────────────────────────────────
	c.statusLine = tview.NewTextView().SetDynamicColors(true)
	c.statusLine.SetBackgroundColor(tcell.NewHexColor(0x12101F))
	c.statusLine.SetText("  [#7B6F8E]idle[-]")

	// ── Footer ──────────────────────────────────────────────────
	c.footer = tview.NewTextView().SetDynamicColors(true)
	c.footer.SetBackgroundColor(tcell.NewHexColor(0x12101F))

	// ── Input field ─────────────────────────────────────────────
	c.input = tview.NewInputField()
	c.input.SetLabel("  [#A855F7]›[-] ")
	c.input.SetLabelColor(tcell.NewHexColor(0xA855F7))
	c.input.SetFieldBackgroundColor(tcell.NewHexColor(0x0A0A12))
	c.input.SetFieldTextColor(tcell.NewHexColor(0xE8E0F0))
	c.input.SetBackgroundColor(tcell.NewHexColor(0x0A0A12))
	c.input.SetBorder(true)
	c.input.SetBorderColor(tcell.NewHexColor(0x2D1B4E))

	c.updateHeader()
	c.updateFooter()

	// ── Root layout ─────────────────────────────────────────────
	c.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(c.header, 1, 0, false).
		AddItem(c.chatLog, 0, 1, false).
		AddItem(c.statusLine, 1, 0, false).
		AddItem(c.footer, 1, 0, false).
		AddItem(c.input, 3, 0, true)

	c.pages = tview.NewPages()
	c.pages.AddPage("main", c.layout, true, true)

	c.app = tview.NewApplication()
	c.app.SetRoot(c.pages, true).EnableMouse(false)
	c.app.SetFocus(c.input)
}

func (c *chatUI) updateHeader() {
	c.header.SetText(fmt.Sprintf(
		"  [#A855F7::b]OCTAI[-]  [#2D1B4E]·[-]  [#7B6F8E]%s[-]  [#2D1B4E]·[-]  [#7B6F8E]session:%s[-]",
		c.modelName, c.shortSession(),
	))
}

func (c *chatUI) updateFooter() {
	c.footer.SetText(fmt.Sprintf(
		"  [#7B6F8E]%s[-]  [#2D1B4E]·[-]  [#7B6F8E]session:%s[-]  [#2D1B4E]·[-]  [#A855F7]Ctrl+L[-][#7B6F8E]:models  /help[-]",
		c.modelName, c.shortSession(),
	))
}

func (c *chatUI) printWelcome() {
	fmt.Fprintf(c.chatLog, "\n")
	fmt.Fprintf(c.chatLog, "[#A855F7::b]  █████╗ ██╗██████╗ ██╗  ██╗ ██████╗ [-]\n")
	fmt.Fprintf(c.chatLog, "[#A855F7::b] ██╔══██╗██║██╔══██╗██║  ██║██╔═══██╗[-]\n")
	fmt.Fprintf(c.chatLog, "[#A855F7::b] ███████║██║██████╔╝███████║██║   ██║[-]\n")
	fmt.Fprintf(c.chatLog, "[#A855F7::b] ██╔══██║██║██╔══██╗██╔══██║██║▄▄ ██║[-]\n")
	fmt.Fprintf(c.chatLog, "[#A855F7::b] ██║  ██║██║██████╔╝██║  ██║╚██████╔╝[-]\n")
	fmt.Fprintf(c.chatLog, "[#A855F7::b] ╚═╝  ╚═╝╚═╝╚═════╝ ╚═╝  ╚═╝ ╚══▀▀═╝[-]\n")
	fmt.Fprintf(c.chatLog, "\n")
	fmt.Fprintf(c.chatLog, "  [#2D1B4E]────────────────────────────────────────[-]\n")
	fmt.Fprintf(c.chatLog, "  [#7B6F8E]model:[-] [#A855F7]%s[-]  [#7B6F8E]· session:[-] [#A855F7]%s[-]\n",
		c.modelName, c.shortSession())
	fmt.Fprintf(c.chatLog, "  [#7B6F8E]type a message to begin · /help for commands[-]\n")
	fmt.Fprintf(c.chatLog, "  [#2D1B4E]────────────────────────────────────────[-]\n\n")
}

func (c *chatUI) run(ctx context.Context, loop *pkgagent.AgentLoop) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.ctx = ctx

	// Cache installed skill names for shorthand dispatch
	c.skillNames = make(map[string]bool)
	if info := loop.GetStartupInfo(); info != nil {
		if skills, ok := info["skills"].(map[string]any); ok {
			if names, ok := skills["names"].([]string); ok {
				for _, n := range names {
					c.skillNames[strings.ToLower(n)] = true
				}
			}
		}
	}

	sub := loop.SubscribeEvents(64)
	defer loop.UnsubscribeEvents(sub.ID)
	go c.handleEvents(ctx, sub.C)

	go c.runSpinner(ctx)

	// Global key capture
	c.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyCtrlD:
			cancel()
			c.app.Stop()
			return nil
		case tcell.KeyCtrlL:
			c.app.QueueUpdateDraw(func() { c.showModelPicker() })
			return nil
		case tcell.KeyEscape:
			if c.pages.HasPage("model-picker") {
				c.hideModelPicker()
				return nil
			}
		}
		return event
	})

	// Input: up/down history
	c.input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		c.mu.Lock()
		defer c.mu.Unlock()
		switch event.Key() {
		case tcell.KeyUp:
			if c.histIdx > 0 {
				c.histIdx--
				text := c.history[c.histIdx]
				c.app.QueueUpdateDraw(func() { c.input.SetText(text) })
			}
			return nil
		case tcell.KeyDown:
			if c.histIdx < len(c.history)-1 {
				c.histIdx++
				text := c.history[c.histIdx]
				c.app.QueueUpdateDraw(func() { c.input.SetText(text) })
			} else {
				c.histIdx = len(c.history)
				c.app.QueueUpdateDraw(func() { c.input.SetText("") })
			}
			return nil
		}
		return event
	})

	// Input: submit on Enter
	c.input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		text := strings.TrimSpace(c.input.GetText())
		if text == "" {
			return
		}
		c.input.SetText("")
		if strings.HasPrefix(text, "/") {
			c.app.QueueUpdateDraw(func() { c.handleSlashCommand(text) })
			return
		}
		c.app.QueueUpdateDraw(func() { c.sendToLoop(text) })
	})

	c.printWelcome()
	return c.app.Run()
}

// handleSlashCommand handles TUI-local commands and forwards the rest to the agent loop.
func (c *chatUI) handleSlashCommand(raw string) {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return
	}
	cmd := strings.ToLower(parts[0])

	// ── TUI-local commands (never forwarded to the agent loop) ──
	switch cmd {
	case "/help":
		c.showHelp()
		return
	case "/clear":
		c.chatLog.Clear()
		c.printWelcome()
		return
	case "/exit", "/quit":
		c.app.Stop()
		return
	case "/session":
		if len(parts) > 1 {
			c.sessionKey = parts[1]
			c.updateHeader()
			c.updateFooter()
			c.appendSystemMessage("Session set to: " + c.sessionKey)
		} else {
			c.appendSystemMessage("Current session: " + c.sessionKey)
		}
		return
	case "/model":
		if len(parts) > 1 {
			c.modelName = parts[1]
			c.updateHeader()
			c.updateFooter()
			c.appendSystemMessage("Model set to: " + c.modelName)
		} else {
			c.appendSystemMessage("Current model: " + c.modelName)
		}
		return
	case "/skills":
		// Alias for /list skills
		raw = "/list skills"
	}

	// ── Skill shorthand: /<skillname> [msg] → /use <skillname> [msg] ──
	cmdWithoutSlash := strings.TrimPrefix(cmd, "/")
	if c.skillNames[cmdWithoutSlash] {
		rest := ""
		if len(parts) > 1 {
			rest = " " + strings.Join(parts[1:], " ")
		}
		raw = "/use " + cmdWithoutSlash + rest
	}

	// ── Forward to agent loop (/use, /list, /show, /status, /think, /fast, …) ──
	c.sendToLoop(raw)
}

// sendToLoop sends text to the agent loop and displays the response in the chat.
// Must be called from within a QueueUpdateDraw callback.
func (c *chatUI) sendToLoop(text string) {
	c.mu.Lock()
	c.history = append(c.history, text)
	c.histIdx = len(c.history)
	c.busy = true
	c.startTime = time.Now()
	c.lastTool = ""
	c.mu.Unlock()

	c.appendUserMessage(text)

	go func() {
		resp, err := c.agentLoop.ProcessDirect(c.ctx, text, c.sessionKey)
		c.mu.Lock()
		c.busy = false
		c.mu.Unlock()
		c.app.QueueUpdateDraw(func() {
			if err != nil {
				c.appendError(err.Error())
			} else if resp != "" {
				c.appendAssistantMessage(resp)
			}
			c.statusLine.SetText("  [#7B6F8E]idle[-]")
			c.chatLog.ScrollToEnd()
		})
	}()
}

func (c *chatUI) showHelp() {
	c.appendSystemMessage(
		"TUI commands:\n" +
			"  /help               this message\n" +
			"  /clear              clear chat log\n" +
			"  /session [key]      show or change session\n" +
			"  /model [name]       show or switch model\n" +
			"  /exit  /quit        exit\n" +
			"\n" +
			"Agent commands (forwarded to loop):\n" +
			"  /skills             list installed skills\n" +
			"  /use <skill> [msg]  invoke a skill\n" +
			"  /<skill> [msg]      shorthand for /use\n" +
			"  /status             agent status\n" +
			"  /list models        list models\n" +
			"  /show model         show current model\n" +
			"  /think              toggle extended thinking\n" +
			"  /fast               toggle fast mode\n" +
			"  /memory <query>     search memory\n" +
			"\n" +
			"Keys:\n" +
			"  Ctrl+L              model picker\n" +
			"  ↑↓                  input history\n" +
			"  Ctrl+C              quit",
	)
}

func (c *chatUI) showModelPicker() {
	if c.pages.HasPage("model-picker") {
		return
	}

	list := tview.NewList()
	list.ShowSecondaryText(true)
	list.SetBackgroundColor(tcell.NewHexColor(0x12101F))
	list.SetMainTextColor(tcell.NewHexColor(0xE8E0F0))
	list.SetSecondaryTextColor(tcell.NewHexColor(0x7B6F8E))
	list.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.NewHexColor(0x1E0F3D)).
		Foreground(tcell.NewHexColor(0xA855F7)))
	list.SetHighlightFullLine(true)
	list.SetBorder(true)
	list.SetBorderColor(tcell.NewHexColor(0xA855F7))
	list.SetTitle(" [#A855F7::b]MODEL PICKER[-]  [#7B6F8E]↑↓ select · Esc close[-] ")
	list.SetTitleColor(tcell.NewHexColor(0xA855F7))

	cfg := c.agentLoop.GetConfig()
	for _, m := range cfg.ModelList {
		name := m.ModelName
		modelID := m.Model
		list.AddItem(name, modelID, 0, func() {
			c.modelName = name
			c.updateHeader()
			c.updateFooter()
			c.hideModelPicker()
			c.appendSystemMessage("Model switched to: " + name)
		})
	}

	// Center at 60% width, capped at 30 rows
	overlay := tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(list, 30, 0, true).
			AddItem(tview.NewBox(), 0, 1, false),
			0, 3, true).
		AddItem(tview.NewBox(), 0, 1, false)

	c.pages.AddPage("model-picker", overlay, true, true)
	c.app.SetFocus(list)
}

func (c *chatUI) hideModelPicker() {
	c.pages.RemovePage("model-picker")
	c.app.SetFocus(c.input)
}

func (c *chatUI) handleEvents(ctx context.Context, ch <-chan pkgagent.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			switch evt.Kind {
			case pkgagent.EventKindToolExecStart:
				if p, ok := evt.Payload.(pkgagent.ToolExecStartPayload); ok {
					var argParts []string
					for k, v := range p.Arguments {
						argParts = append(argParts, fmt.Sprintf("%s=%v", k, v))
					}
					argsStr := strings.Join(argParts, ", ")
					c.mu.Lock()
					c.lastTool = p.Tool
					c.mu.Unlock()
					tool, args := p.Tool, argsStr
					c.app.QueueUpdateDraw(func() { c.appendToolStart(tool, args) })
				}
			case pkgagent.EventKindToolExecEnd:
				if p, ok := evt.Payload.(pkgagent.ToolExecEndPayload); ok {
					isErr, dur, tool := p.IsError, p.Duration, p.Tool
					c.app.QueueUpdateDraw(func() { c.appendToolEnd(tool, dur, isErr) })
				}
			case pkgagent.EventKindLLMRetry:
				if p, ok := evt.Payload.(pkgagent.LLMRetryPayload); ok {
					reason, attempt := p.Reason, p.Attempt
					c.app.QueueUpdateDraw(func() {
						fmt.Fprintf(c.chatLog, "  [#F87171]↺  retrying: %s (attempt %d)[-]\n",
							tview.Escape(reason), attempt)
					})
				}
			case pkgagent.EventKindError:
				if p, ok := evt.Payload.(pkgagent.ErrorPayload); ok {
					msg, stage := p.Message, p.Stage
					c.app.QueueUpdateDraw(func() {
						c.appendError(fmt.Sprintf("[%s] %s", stage, msg))
					})
				}
			}
		}
	}
}

func (c *chatUI) runSpinner(ctx context.Context) {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			busy := c.busy
			tool := c.lastTool
			elapsed := time.Since(c.startTime)
			frame := spinnerFrames[c.spinIdx%len(spinnerFrames)]
			if busy {
				c.spinIdx++
			}
			c.mu.Unlock()

			if busy {
				var status string
				if tool != "" {
					status = fmt.Sprintf("  [#A855F7]%s[-]  [#7B6F8E]%s · %.1fs[-]", frame, tool, elapsed.Seconds())
				} else {
					status = fmt.Sprintf("  [#A855F7]%s[-]  [#7B6F8E]thinking · %.1fs[-]", frame, elapsed.Seconds())
				}
				c.app.QueueUpdateDraw(func() { c.statusLine.SetText(status) })
			}
		}
	}
}

func (c *chatUI) appendUserMessage(text string) {
	fmt.Fprintf(c.chatLog, "\n[#A855F7:#1E0F3D:b] You [-:-:-]\n")
	for _, line := range strings.Split(text, "\n") {
		fmt.Fprintf(c.chatLog, "[#E8E0F0:#1E0F3D:-] %s [-:-:-]\n", tview.Escape(line))
	}
	fmt.Fprintf(c.chatLog, "\n")
	c.chatLog.ScrollToEnd()
}

func (c *chatUI) appendAssistantMessage(text string) {
	fmt.Fprintf(c.chatLog, "[#A855F7]●[-]  [#A855F7::b]OctAi[-]\n\n")
	fmt.Fprintf(c.chatLog, "%s\n\n", tview.Escape(text))
}

func (c *chatUI) appendSystemMessage(text string) {
	lines := strings.Split(text, "\n")
	fmt.Fprintf(c.chatLog, "\n")
	for _, line := range lines {
		fmt.Fprintf(c.chatLog, "  [#7B6F8E]%s[-]\n", line)
	}
	fmt.Fprintf(c.chatLog, "\n")
	c.chatLog.ScrollToEnd()
}

// appendToolStart draws the top border and args line of a tool card.
// appendToolEnd must be called to close it.
func (c *chatUI) appendToolStart(tool, args string) {
	const inner = cardWidth - 2 // inner display width between │ chars

	// Title bar: ┌─ tool_name ──...──┐
	if len(tool) > inner-6 {
		tool = tool[:inner-9] + "..."
	}
	titlePart := "─ " + tool + " "
	dashes := inner - 1 - len(titlePart) // -1 for the ┌ char
	if dashes < 0 {
		dashes = 0
	}
	fmt.Fprintf(c.chatLog, "  [#2D1B4E]┌[#7B6F8E::i]%s[-][#2D1B4E]%s┐[-]\n",
		titlePart, strings.Repeat("─", dashes))

	// Args row: │ ⟳  args...  │
	if len(args) > inner-5 {
		args = args[:inner-8] + "..."
	}
	argPad := inner - 5 - len(args) // 5 = len(" ⟳  ")
	if argPad < 0 {
		argPad = 0
	}
	fmt.Fprintf(c.chatLog, "  [#2D1B4E]│[-][#7B6F8E] ⟳  %s%s[#2D1B4E]│[-]\n",
		tview.Escape(args), strings.Repeat(" ", argPad))
}

// appendToolEnd draws the result line and bottom border of a tool card.
func (c *chatUI) appendToolEnd(_ string, dur time.Duration, isErr bool) {
	const inner = cardWidth - 2

	durStr := fmt.Sprintf("%.2fs", dur.Seconds())
	// 5 = len(" ✓  ") or len(" ✗  ")
	durPad := inner - 5 - len(durStr)
	if durPad < 0 {
		durPad = 0
	}

	if isErr {
		fmt.Fprintf(c.chatLog, "  [#2D1B4E]│[-][#F87171] ✗  %s%s[#2D1B4E]│[-]\n",
			durStr, strings.Repeat(" ", durPad))
	} else {
		fmt.Fprintf(c.chatLog, "  [#2D1B4E]│[-][#34D399] ✓[-][#7B6F8E]  %s%s[#2D1B4E]│[-]\n",
			durStr, strings.Repeat(" ", durPad))
	}
	fmt.Fprintf(c.chatLog, "  [#2D1B4E]└%s┘[-]\n\n", strings.Repeat("─", cardWidth-2))
}

func (c *chatUI) appendError(msg string) {
	fmt.Fprintf(c.chatLog, "\n  [#F87171]✗  %s[-]\n\n", tview.Escape(msg))
	c.chatLog.ScrollToEnd()
}
