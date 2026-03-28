package onboard

import (
	config "github.com/raynaythegreat/ai-business-hq/pkg/config"
)

// authMethod describes a single way to authenticate with a provider.
type authMethod struct {
	Label     string // shown in menu: "OAuth (browser)", "API key", etc.
	IsOAuth   bool   // triggers OAuth flow instead of API-key prompt
	DevCode   bool   // use device-code flow (no browser)
	OAuthArg  string // provider string passed to auth.LoginProvider
}

// providerInfo describes a provider entry shown in the wizard.
type providerInfo struct {
	Name        string       // Display name
	Models      []string     // ModelConfig.ModelName values (from DefaultConfig)
	Category    string       // section header
	NeedsKey    bool         // false for local providers (Ollama, VLLM)
	Desc        string       // one-line description
	AuthMethods []authMethod // if non-empty, user picks auth method; else plain API key
}

// providerCatalog is the ordered list shown during onboarding.
// ModelName values must match entries in DefaultConfig().ModelList.
var providerCatalog = []providerInfo{
	// ── Popular ──────────────────────────────────────────────────────────────
	{
		Name:     "OpenRouter",
		Models:   config.ModelNamesForProvider("openrouter"),
		Category: "Popular",
		NeedsKey: true,
		Desc:     "100+ models via a single API key — auto-routes to best model",
	},
	{
		Name:     "Anthropic",
		Models:   config.ModelNamesForProvider("anthropic"),
		Category: "Popular",
		NeedsKey: true,
		Desc:     "Claude family — Opus (powerful), Sonnet (balanced), Haiku (fast)",
		AuthMethods: []authMethod{
			{Label: "Setup token (from `claude setup-token`)", IsOAuth: true, OAuthArg: "anthropic"},
			{Label: "API key (from console.anthropic.com)"},
		},
	},
	{
		Name:     "OpenAI",
		Models:   config.ModelNamesForProvider("openai"),
		Category: "Popular",
		NeedsKey: true,
		Desc:     "GPT-5.4, GPT-4o, o3/o4 reasoning models",
		AuthMethods: []authMethod{
			{Label: "OAuth — browser login", IsOAuth: true, OAuthArg: "openai"},
			{Label: "OAuth — device code (no browser)", IsOAuth: true, DevCode: true, OAuthArg: "openai"},
			{Label: "API key (from platform.openai.com)"},
		},
	},
	{
		Name:     "Ollama",
		Models:   config.ModelNamesForProvider("ollama"),
		Category: "Popular",
		NeedsKey: false,
		Desc:     "Local models, completely free — pull any model with `ollama pull`",
	},
	// ── Fast Inference ────────────────────────────────────────────────────────
	{
		Name:     "Groq",
		Models:   config.ModelNamesForProvider("groq"),
		Category: "Fast Inference",
		NeedsKey: true,
		Desc:     "Ultra-fast inference — Llama, Mixtral, Gemma",
	},
	{
		Name:     "Cerebras",
		Models:   config.ModelNamesForProvider("cerebras"),
		Category: "Fast Inference",
		NeedsKey: true,
		Desc:     "Wafer-scale chips — extremely fast inference",
	},
	// ── Reasoning Models ─────────────────────────────────────────────────────
	{
		Name:     "DeepSeek",
		Models:   config.ModelNamesForProvider("deepseek"),
		Category: "Reasoning Models",
		NeedsKey: true,
		Desc:     "DeepSeek Chat + R1 Reasoner — great value with full reasoning",
	},
	{
		Name:     "xAI Grok",
		Models:   config.ModelNamesForProvider("grok"),
		Category: "Reasoning Models",
		NeedsKey: true,
		Desc:     "Grok 4 — xAI's frontier model with fast reasoning",
	},
	// ── Google ────────────────────────────────────────────────────────────────
	{
		Name:     "Google Gemini",
		Models:   config.ModelNamesForProvider("gemini"),
		Category: "Google",
		NeedsKey: true,
		Desc:     "Gemini 2.0 Flash / 2.5 Pro — fast, multimodal, long context",
		AuthMethods: []authMethod{
			{Label: "OAuth — browser login (Google Antigravity)", IsOAuth: true, OAuthArg: "google-antigravity"},
			{Label: "API key (from aistudio.google.com)"},
		},
	},
	// ── Other Cloud ───────────────────────────────────────────────────────────
	{
		Name:     "Mistral AI",
		Models:   config.ModelNamesForProvider("mistral"),
		Category: "Other Cloud",
		NeedsKey: true,
		Desc:     "Mistral Large/Small/Codestral — efficient European models",
	},
	{
		// TODO: sync with defaults.go — "kimi-k2.5" uses scheme "avian/moonshotai/..."
		// not "moonshot/...", so ModelNamesForProvider("moonshot") would drop it.
		Name:     "Moonshot / Kimi",
		Models:   []string{"kimi-k2.5", "kimi-k2-thinking", "kimi-k2-turbo"},
		Category: "Other Cloud",
		NeedsKey: true,
		Desc:     "Kimi K2.5 + Moonshot — long context Chinese AI models",
	},
	{
		Name:     "Together AI",
		Models:   config.ModelNamesForProvider("together"),
		Category: "Other Cloud",
		NeedsKey: true,
		Desc:     "100+ open models — Llama 4, DeepSeek R1, Kimi K2",
	},
	{
		Name:     "Avian",
		Models:   config.ModelNamesForProvider("avian"),
		Category: "Other Cloud",
		NeedsKey: true,
		Desc:     "DeepSeek V3.2 + Kimi K2.5 via Avian gateway",
	},
	{
		Name:     "Perplexity",
		Models:   config.ModelNamesForProvider("perplexity"),
		Category: "Other Cloud",
		NeedsKey: true,
		Desc:     "Web-search augmented responses — always up-to-date",
	},
	{
		Name:     "GitHub Copilot",
		Models:   config.ModelNamesForProvider("github-copilot"),
		Category: "Other Cloud",
		NeedsKey: false,
		Desc:     "GPT-5.4 via GitHub Copilot subscription (OAuth — no extra cost)",
	},
	// ── Local / Self-Hosted ───────────────────────────────────────────────────
	{
		Name:     "VLLM",
		Models:   config.ModelNamesForProvider("vllm"),
		Category: "Local / Self-Hosted",
		NeedsKey: false,
		Desc:     "Any model via local vLLM server (http://localhost:8000)",
	},
	{
		Name:     "LiteLLM",
		Models:   config.ModelNamesForProvider("litellm"),
		Category: "Local / Self-Hosted",
		NeedsKey: true,
		Desc:     "LiteLLM proxy — unified gateway to 100+ providers with cost tracking",
	},
}
