package onboard

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"

	authpkg "github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/auth"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal"
	"github.com/raynaythegreat/ai-business-hq/pkg/config"
)

// ANSI helpers
const (
	bold    = "\033[1m"
	dim     = "\033[2m"
	cyan    = "\033[1;38;2;0;200;220m"
	blue    = "\033[1;38;2;62;93;185m"
	red     = "\033[1;38;2;213;70;70m"
	green   = "\033[1;38;2;80;200;120m"
	yellow  = "\033[1;38;2;220;180;50m"
	reset   = "\033[0m"
	divider = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
)

func hr() string { return dim + divider + reset }

// sentinel errors for back/skip navigation
var errBack = errors.New("back")
var errSkip = errors.New("skip")

// runWizard is the full interactive onboarding wizard.
func runWizard(encrypt bool) {
	sc := bufio.NewScanner(os.Stdin)
	configPath := internal.GetConfigPath()

	// ── Welcome ──────────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Printf("%s%s AIBHQ Setup Wizard%s\n", bold, internal.Logo, reset)
	fmt.Println(hr())
	fmt.Println("This wizard configures your AI providers, models, and tools.")
	fmt.Println("You can re-run it any time with " + cyan + "octai onboard" + reset + ".")
	fmt.Println()

	// Check existing config
	configExists := false
	if _, err := os.Stat(configPath); err == nil {
		configExists = true
		fmt.Printf("%sExisting config found at %s%s\n", yellow, configPath, reset)
		fmt.Print("Start fresh? (y/n, default n): ")
		if sc.Scan() && strings.TrimSpace(sc.Text()) == "y" {
			configExists = false
		}
		fmt.Println()
	}

	// Load or create config
	var cfg *config.Config
	if configExists {
		var err error
		cfg, err = config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg = config.DefaultConfig()
	}

	// ── Encryption ───────────────────────────────────────────────────────────
	if encrypt {
		fmt.Println(bold + "Credential Encryption" + reset)
		fmt.Println(hr())
		passphrase, err := promptPassphrase()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		os.Setenv("PICOCLAW_KEY_PASSPHRASE", passphrase)
		if err = setupSSHKey(); err != nil {
			fmt.Printf("Error generating SSH key: %v\n", err)
			os.Exit(1)
		}
		fmt.Println()
	}

	// ── Step-based flow with back/skip ────────────────────────────────────────
	var (
		selectedIdxs []int
		configured   []string
		primary      string
	)

	type stepFn func() error
	steps := []stepFn{
		// Step 1 — Select Providers
		func() error {
			fmt.Println(bold + "Step 1 — Select Providers" + reset)
			fmt.Println(hr())
			idxs, err := selectProviders(sc)
			if err != nil {
				return err
			}
			selectedIdxs = idxs
			fmt.Println()
			return nil
		},
		// Step 2 — API Keys / Auth
		func() error {
			fmt.Println(bold + "Step 2 — API Keys / Auth" + reset)
			fmt.Println(hr())
			if err := enterAPIKeys(sc, cfg, selectedIdxs); err != nil {
				return err
			}
			// Reload config in case OAuth updated it
			if reloaded, err2 := config.LoadConfig(configPath); err2 == nil {
				cfg = reloaded
			}
			configured = configuredModels(cfg, selectedIdxs)
			fmt.Println()
			return nil
		},
		// Step 3 — Primary Model
		func() error {
			if len(configured) == 0 {
				fmt.Println(yellow + "No providers configured — skipping model selection." + reset)
				fmt.Println("You can add providers later by editing ~/.octai/config.json")
				return errSkip
			}
			fmt.Println(bold + "Step 3 — Primary Model" + reset)
			fmt.Println(hr())
			fmt.Println("Which model should agents use by default?")
			p, err := pickOne(sc, configured)
			if err != nil {
				return err
			}
			primary = p
			cfg.Agents.Defaults.ModelName = primary
			fmt.Println()
			return nil
		},
		// Step 4 — Fallback Models
		func() error {
			if len(configured) == 0 {
				return errSkip
			}
			fmt.Println(bold + "Step 4 — Fallback Models" + reset)
			fmt.Println(hr())
			fmt.Printf("When %s%s%s hits a rate limit or error, automatically failover to:\n", cyan, primary, reset)
			fmt.Println(dim + "(These are tried in order — select 0 or more)" + reset)
			remaining := without(configured, primary)
			if len(remaining) > 0 {
				fallbacks, err := pickMany(sc, remaining, true)
				if err != nil {
					return err
				}
				cfg.Agents.Defaults.ModelFallbacks = fallbacks
			} else {
				fmt.Println(dim + "No additional models available for fallback." + reset)
			}
			fmt.Println()
			return nil
		},
		// Step 5 — Smart Routing
		func() error {
			if len(configured) == 0 {
				return errSkip
			}
			fmt.Println(bold + "Step 5 — Smart Routing" + reset)
			fmt.Println(hr())
			fmt.Println("Smart routing sends simple tasks to a cheaper/faster model")
			fmt.Println("and routes complex tasks (code, long context) to your primary model.")
			fmt.Println()
			ans, err := promptLine(sc, "Enable smart routing? (y/n, default n)")
			if err != nil {
				return err
			}
			if strings.ToLower(strings.TrimSpace(ans)) == "y" {
				lightModel := primary
				remaining := without(configured, primary)
				if len(remaining) > 0 {
					fmt.Println()
					fmt.Println("Pick the " + green + "light model" + reset + " for simple tasks:")
					lm, err2 := pickOne(sc, remaining)
					if err2 != nil {
						return err2
					}
					lightModel = lm
				}
				cfg.Agents.Defaults.Routing = &config.RoutingConfig{
					Enabled:    true,
					LightModel: lightModel,
					Threshold:  0.35,
				}
				fmt.Printf("%s✓ Routing enabled%s — simple → %s%s%s, complex → %s%s%s\n",
					green, reset, cyan, lightModel, reset, cyan, primary, reset)
			}
			fmt.Println()
			return nil
		},
		// Step 6 — Tools
		func() error {
			fmt.Println(bold + "Step 6 — Tools" + reset)
			fmt.Println(hr())
			if err := configureTools(sc, cfg); err != nil {
				return err
			}
			fmt.Println()
			return nil
		},
	}

	// Run steps with back/skip support
	i := 0
	for i < len(steps) {
		err := steps[i]()
		if errors.Is(err, errBack) {
			if i > 0 {
				i--
			}
			fmt.Println()
			continue
		}
		if errors.Is(err, errSkip) {
			i++
			continue
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		i++
	}

	// ── Prune unconfigured models ─────────────────────────────────────────────
	pruneUnconfigured(cfg, selectedIdxs)

	// ── Save ──────────────────────────────────────────────────────────────────
	createWorkspaceTemplates(cfg.WorkspacePath())

	if err := config.SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	// ── Done ──────────────────────────────────────────────────────────────────
	fmt.Println(hr())
	fmt.Printf("\n%s%s octai is ready!%s\n\n", green+bold, internal.Logo, reset)

	if cfg.Agents.Defaults.ModelName != "" {
		fmt.Printf("  Primary model:  %s%s%s\n", cyan, cfg.Agents.Defaults.ModelName, reset)
	}
	if len(cfg.Agents.Defaults.ModelFallbacks) > 0 {
		fmt.Printf("  Fallbacks:      %s%s%s\n", dim, strings.Join(cfg.Agents.Defaults.ModelFallbacks, " → "), reset)
	}
	if cfg.Agents.Defaults.Routing != nil && cfg.Agents.Defaults.Routing.Enabled {
		fmt.Printf("  Smart routing:  %senabled%s (light model: %s)\n", green, reset, cfg.Agents.Defaults.Routing.LightModel)
	}
	fmt.Printf("  Config saved:   %s%s%s\n\n", dim, configPath, reset)

	fmt.Println("Next steps:")
	fmt.Printf("  Chat:       %soctai agent -m \"Hello!\"%s\n", cyan, reset)
	fmt.Printf("  Web UI:     %soctai web --console%s\n", cyan, reset)
	fmt.Printf("  TUI editor: %soctai tui%s\n\n", cyan, reset)
}

// ── Provider selection ────────────────────────────────────────────────────────

func selectProviders(sc *bufio.Scanner) ([]int, error) {
	currentCat := ""
	for i, p := range providerCatalog {
		if p.Category != currentCat {
			currentCat = p.Category
			fmt.Printf("\n%s%s%s\n", bold, currentCat, reset)
		}
		keyNote := ""
		if !p.NeedsKey {
			keyNote = dim + " (no API key needed)" + reset
		}
		oauthNote := ""
		if len(p.AuthMethods) > 0 {
			oauthNote = dim + " [OAuth available]" + reset
		}
		fmt.Printf("  %s[%2d]%s %-18s%s%s %s\n",
			cyan, i+1, reset,
			p.Name,
			keyNote,
			oauthNote,
			dim+"— "+p.Desc+reset,
		)
	}
	fmt.Println()
	fmt.Printf("Enter numbers separated by commas (e.g. %s1,2,4%s)\n", cyan, reset)
	fmt.Printf("%s[b=back, s=skip]%s Enter: ", dim, reset)

	if !sc.Scan() {
		return nil, nil
	}
	line := strings.TrimSpace(sc.Text())
	if line == "b" {
		return nil, errBack
	}
	if line == "s" {
		return nil, errSkip
	}
	return parseNumberList(line, len(providerCatalog)), nil
}

// ── API key / OAuth entry ─────────────────────────────────────────────────────

func enterAPIKeys(sc *bufio.Scanner, cfg *config.Config, selectedIdxs []int) error {
	for _, idx := range selectedIdxs {
		p := providerCatalog[idx]

		if !p.NeedsKey {
			fmt.Printf("  %s%-18s%s %sno key needed%s\n", cyan, p.Name, reset, dim, reset)
			continue
		}

		// Provider supports multiple auth methods — let user choose
		if len(p.AuthMethods) > 0 {
			fmt.Printf("\n  %s%s%s — choose auth method:\n", bold, p.Name, reset)
			for i, m := range p.AuthMethods {
				fmt.Printf("    %s[%d]%s %s\n", cyan, i+1, reset, m.Label)
			}
			fmt.Printf("    %s[s]%s skip   %s[b]%s back\n", cyan, reset, cyan, reset)

			chosen := -1
			for chosen < 0 {
				fmt.Print("  Enter choice [1]: ")
				if !sc.Scan() {
					break
				}
				t := strings.TrimSpace(sc.Text())
				switch t {
				case "b":
					return errBack
				case "s":
					chosen = -2 // skip this provider
				case "":
					chosen = 0 // default to first option
				default:
					n, err := strconv.Atoi(t)
					if err != nil || n < 1 || n > len(p.AuthMethods) {
						fmt.Printf("  %sInvalid — enter 1-%d, s, or b.%s\n", yellow, len(p.AuthMethods), reset)
						continue
					}
					chosen = n - 1
				}
			}

			if chosen == -2 || chosen < 0 {
				fmt.Printf("  %sskipped%s\n", yellow, reset)
				continue
			}

			method := p.AuthMethods[chosen]
			if method.IsOAuth {
				fmt.Printf("  %s→ Starting OAuth for %s...%s\n", cyan, p.Name, reset)
				if err := authpkg.LoginProvider(method.OAuthArg, method.DevCode, true); err != nil {
					fmt.Printf("  %s✗ OAuth failed: %v%s\n", red, err, reset)
					fmt.Printf("  %sskipped%s\n", yellow, reset)
				} else {
					fmt.Printf("  %s✓ OAuth login successful for %s%s\n", green, p.Name, reset)
				}
				continue
			}
			// Fall through to plain API-key prompt
		}

		// Plain API-key prompt
		fmt.Printf("  %s%s%s API key: ", bold, p.Name, reset)
		keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil || len(keyBytes) == 0 {
			fmt.Printf("  %sskipped%s\n", yellow, reset)
			continue
		}
		key := strings.TrimSpace(string(keyBytes))
		for _, modelName := range p.Models {
			mc, _ := cfg.GetModelConfig(modelName)
			if mc != nil {
				mc.SetAPIKey(key)
			}
		}
		fmt.Printf("  %s✓ Key set for %s%s\n", green, p.Name, reset)
	}
	return nil
}

// ── configuredModels returns ModelName list for all selected providers ────────

func configuredModels(cfg *config.Config, selectedIdxs []int) []string {
	var names []string
	seen := map[string]bool{}
	for _, idx := range selectedIdxs {
		for _, modelName := range providerCatalog[idx].Models {
			if !seen[modelName] {
				seen[modelName] = true
				names = append(names, modelName)
			}
		}
	}
	return names
}

// ── Tools configuration ───────────────────────────────────────────────────────

type toolToggle struct {
	label   string
	enabled *bool
}

func configureTools(sc *bufio.Scanner, cfg *config.Config) error {
	toggles := []toolToggle{
		{"Web Search (DuckDuckGo + Brave + Tavily)", &cfg.Tools.Web.Enabled},
		{"Code / Shell Execution", &cfg.Tools.Exec.Enabled},
		{"File Operations (read/write/edit)", &cfg.Tools.EditFile.Enabled},
		{"MCP Servers", &cfg.Tools.MCP.Enabled},
	}

	// Enable all by default
	for _, t := range toggles {
		*t.enabled = true
	}

	for {
		fmt.Println("Toggle tools (Enter to keep defaults):")
		for i, t := range toggles {
			state := green + "✓ on" + reset
			if !*t.enabled {
				state = dim + "✗ off" + reset
			}
			fmt.Printf("  %s[%d]%s %-42s %s\n", cyan, i+1, reset, t.label, state)
		}
		fmt.Printf("\n%s[b=back, s/Enter=continue]%s Toggle number or action: ", dim, reset)
		if !sc.Scan() {
			break
		}
		line := strings.TrimSpace(sc.Text())
		if line == "b" {
			return errBack
		}
		if line == "" || line == "s" {
			break
		}
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > len(toggles) {
			fmt.Println(yellow + "Invalid — enter a number from the list, b, or s." + reset)
			continue
		}
		t := toggles[n-1]
		*t.enabled = !*t.enabled
	}
	return nil
}

// ── pruneUnconfigured removes models that were not selected and have no key ───

func pruneUnconfigured(cfg *config.Config, selectedIdxs []int) {
	keep := map[string]bool{}
	for _, idx := range selectedIdxs {
		for _, m := range providerCatalog[idx].Models {
			keep[m] = true
		}
	}
	var filtered []*config.ModelConfig
	for _, mc := range cfg.ModelList {
		if keep[mc.ModelName] || mc.APIKey() != "" {
			filtered = append(filtered, mc)
		}
	}
	cfg.ModelList = filtered
}

// ── Menu helpers ──────────────────────────────────────────────────────────────

// promptLine shows a prompt and returns trimmed input. Returns errBack/errSkip.
func promptLine(sc *bufio.Scanner, prompt string) (string, error) {
	fmt.Printf("%s %s[b=back, s=skip]%s: ", prompt, dim, reset)
	if !sc.Scan() {
		return "", nil
	}
	t := strings.TrimSpace(sc.Text())
	if t == "b" {
		return "", errBack
	}
	if t == "s" {
		return "", errSkip
	}
	return t, nil
}

// pickOne shows a numbered list and returns the chosen item.
func pickOne(sc *bufio.Scanner, items []string) (string, error) {
	for i, item := range items {
		fmt.Printf("  %s[%d]%s %s\n", cyan, i+1, reset, item)
	}
	for {
		fmt.Printf("%s[b=back, s=skip]%s Enter number: ", dim, reset)
		if !sc.Scan() {
			return items[0], nil
		}
		t := strings.TrimSpace(sc.Text())
		if t == "b" {
			return "", errBack
		}
		if t == "s" {
			return "", errSkip
		}
		n, err := strconv.Atoi(t)
		if err != nil || n < 1 || n > len(items) {
			fmt.Printf("%sInvalid — enter 1-%d, b, or s.%s\n", yellow, len(items), reset)
			continue
		}
		return items[n-1], nil
	}
}

// pickMany shows a numbered list for multi-select. allowEmpty skips on blank.
func pickMany(sc *bufio.Scanner, items []string, allowEmpty bool) ([]string, error) {
	for i, item := range items {
		fmt.Printf("  %s[%d]%s %s\n", cyan, i+1, reset, item)
	}
	hint := fmt.Sprintf("comma-separated numbers (e.g. %s1,3%s)", cyan, reset)
	if allowEmpty {
		hint += ", or " + cyan + "Enter" + reset + " to skip"
	}
	fmt.Printf("%s[b=back]%s Enter %s: ", dim, reset, hint)
	if !sc.Scan() {
		return nil, nil
	}
	line := strings.TrimSpace(sc.Text())
	if line == "b" {
		return nil, errBack
	}
	if line == "" {
		return nil, nil
	}
	idxs := parseNumberList(line, len(items))
	var result []string
	for _, i := range idxs {
		result = append(result, items[i])
	}
	return result, nil
}

// parseNumberList parses "1,3,5" into zero-based indices, clamped to max.
func parseNumberList(s string, max int) []int {
	seen := map[int]bool{}
	var result []int
	for _, part := range strings.Split(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 1 || n > max || seen[n-1] {
			continue
		}
		seen[n-1] = true
		result = append(result, n-1)
	}
	return result
}

// without returns items with toRemove excluded.
func without(items []string, toRemove string) []string {
	var result []string
	for _, s := range items {
		if s != toRemove {
			result = append(result, s)
		}
	}
	return result
}
