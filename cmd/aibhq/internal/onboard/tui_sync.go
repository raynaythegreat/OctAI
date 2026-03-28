package onboard

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/raynaythegreat/ai-business-hq/pkg/config"

	tuicfg "github.com/raynaythegreat/ai-business-hq/cmd/aibhq-launcher/config"
)

// syncToTUIConfig writes ~/.octai/tui.toml from the just-configured main config
// so that the TUI launcher is ready to use immediately after onboarding.
func syncToTUIConfig(cfg *config.Config, primary string) {
	tuiPath := tuicfg.DefaultConfigPath()
	tuiCfg, err := tuicfg.Load(tuiPath)
	if err != nil {
		tuiCfg = tuicfg.DefaultConfig()
	}

	// Rebuild schemes/users from scratch based on what was just configured.
	tuiCfg.Provider.Schemes = nil
	tuiCfg.Provider.Users = nil
	tuiCfg.Provider.Current = tuicfg.ProviderCurrent{}

	// Track which base-URLs we've already added as schemes (one Scheme per endpoint).
	schemeByBase := map[string]string{} // baseURL → schemeName

	for _, mc := range cfg.ModelList {
		if mc.IsVirtual() {
			continue
		}
		if mc.APIBase == "" {
			// CLI-based / OAuth providers (GitHub Copilot, etc.) — not representable in tui.toml
			continue
		}

		key := mc.APIKey()
		// For no-key providers (Ollama, VLLM) the key is intentionally empty — allow them.
		// For key-required providers, skip models that were never configured.
		if key == "" && !isNoKeyProvider(mc.APIBase) {
			continue
		}

		schemeName, exists := schemeByBase[mc.APIBase]
		if !exists {
			schemeName = schemeNameFromURL(mc.APIBase)
			// Ensure unique scheme names (e.g., two providers on different ports).
			schemeName = uniqueSchemeName(schemeName, tuiCfg)

			schemeByBase[mc.APIBase] = schemeName
			tuiCfg.Provider.Schemes = append(tuiCfg.Provider.Schemes, tuicfg.Scheme{
				Name:    schemeName,
				BaseURL: mc.APIBase,
				Type:    schemeTypeFromModel(mc.Model),
			})
			tuiCfg.Provider.Users = append(tuiCfg.Provider.Users, tuicfg.User{
				Name:   "default",
				Scheme: schemeName,
				Type:   "key",
				Key:    key,
			})
		}

		// Set Current to the primary model.
		if mc.ModelName == primary && tuiCfg.Provider.Current.Model == "" {
			tuiCfg.Provider.Current = tuicfg.ProviderCurrent{
				Scheme: schemeName,
				User:   "default",
				Model:  modelIDFromModel(mc.Model),
			}
		}
	}

	if err := tuicfg.Save(tuiPath, tuiCfg); err != nil {
		fmt.Printf("%s  warning: could not write %s: %v%s\n", yellow, tuiPath, err, reset)
		return
	}
	fmt.Printf("  TUI config:     %s%s%s\n", dim, tuiPath, reset)
}

// schemeNameFromURL derives a short, human-readable scheme name from an API base URL.
func schemeNameFromURL(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "custom"
	}
	host := strings.ToLower(u.Hostname())
	port := u.Port()

	switch {
	case strings.Contains(host, "openai.com"):
		return "openai"
	case strings.Contains(host, "anthropic.com"):
		return "anthropic"
	case strings.Contains(host, "openrouter.ai"):
		return "openrouter"
	case strings.Contains(host, "groq.com"):
		return "groq"
	case strings.Contains(host, "deepseek.com"):
		return "deepseek"
	case strings.Contains(host, "mistral.ai"):
		return "mistral"
	case strings.Contains(host, "together.xyz") || strings.Contains(host, "together.ai"):
		return "together"
	case strings.Contains(host, "cerebras.ai"):
		return "cerebras"
	case strings.Contains(host, "perplexity.ai"):
		return "perplexity"
	case strings.Contains(host, "x.ai") || strings.Contains(host, "xai.com"):
		return "grok"
	case strings.Contains(host, "googleapis.com") || strings.Contains(host, "generativelanguage"):
		return "gemini"
	case strings.Contains(host, "github.com") || strings.Contains(host, "githubcopilot.com"):
		return "github-copilot"
	case strings.Contains(host, "moonshot.cn") || strings.Contains(host, "moonshotai"):
		return "moonshot"
	case strings.Contains(host, "avian.io"):
		return "avian"
	case strings.Contains(host, "litellm"):
		return "litellm"
	case host == "localhost" || strings.HasPrefix(host, "127.") || host == "0.0.0.0":
		switch port {
		case "11434":
			return "ollama"
		case "8000":
			return "vllm"
		case "4000":
			return "litellm"
		default:
			if port != "" {
				return "local-" + port
			}
			return "local"
		}
	default:
		// Use the host with dots replaced for unknown providers.
		name := strings.ReplaceAll(host, ".", "-")
		if port != "" {
			name += "-" + port
		}
		return name
	}
}

// schemeTypeFromModel returns the TUI scheme type for a given model string.
func schemeTypeFromModel(model string) string {
	if strings.HasPrefix(model, "anthropic/") {
		return "anthropic"
	}
	return "openai-compatible"
}

// modelIDFromModel strips the protocol prefix from a model identifier.
// "openai/gpt-4o" → "gpt-4o", "gpt-4o" → "gpt-4o"
func modelIDFromModel(model string) string {
	if idx := strings.Index(model, "/"); idx >= 0 {
		return model[idx+1:]
	}
	return model
}

// isNoKeyProvider returns true for providers that work without an API key.
func isNoKeyProvider(baseURL string) bool {
	u, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == "localhost" || strings.HasPrefix(host, "127.") || host == "0.0.0.0"
}

// uniqueSchemeName ensures the returned name is not already used in tuiCfg.
func uniqueSchemeName(name string, tuiCfg *tuicfg.TUIConfig) string {
	used := map[string]bool{}
	for _, s := range tuiCfg.Provider.Schemes {
		used[s.Name] = true
	}
	if !used[name] {
		return name
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", name, i)
		if !used[candidate] {
			return candidate
		}
	}
}
