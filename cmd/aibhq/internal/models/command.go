package models

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal"
	"github.com/raynaythegreat/ai-business-hq/pkg/auth"
	"github.com/raynaythegreat/ai-business-hq/pkg/config"
)

// NewModelsCommand creates the "aibhq models" subcommand.
// Subcommands:
//   - aibhq models list      — list all configured models with auth status
//   - aibhq models scan      — check which models have valid API keys
//   - aibhq models status    — show model chain + auth health (--check for CI)
//   - aibhq models fallbacks — manage fallback chains
func NewModelsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "Manage and inspect configured models",
		Long: `List, scan, and manage configured models.

Examples:
  aibhq models list               # List all configured models with auth status
  aibhq models scan               # Show which models are ready to use
  aibhq models status             # Show model chain and auth health
  aibhq models status --check     # Exit 1 if expired/missing, 2 if expiring soon
  aibhq models fallbacks list     # List fallback chain for the default model
  aibhq models fallbacks add claude-haiku-4-5   # Append a fallback
  aibhq models fallbacks remove claude-haiku-4-5
  aibhq models fallbacks clear    # Remove all fallbacks`,
	}

	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newScanCommand())
	cmd.AddCommand(newStatusCommand())
	cmd.AddCommand(newFallbacksCommand())

	return cmd
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured models",
		Long:  "List all models defined in model_list, grouped by provider protocol.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := internal.GetConfigPath()
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			printModelList(cfg)
			return nil
		},
	}
}

func newScanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan which models have API keys configured",
		Long:  "For each model in model_list, check whether an API key is present.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := internal.GetConfigPath()
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			printModelScan(cfg)
			return nil
		},
	}
}

func printModelList(cfg *config.Config) {
	if len(cfg.ModelList) == 0 {
		fmt.Println("No models configured in model_list.")
		return
	}

	defaultModel := cfg.Agents.Defaults.ModelName

	// Group by protocol prefix (part before "/" in Model field)
	byProtocol := make(map[string][]*config.ModelConfig)
	var protocolOrder []string
	seen := make(map[string]bool)

	for _, m := range cfg.ModelList {
		proto := extractProtocol(m.Model)
		if !seen[proto] {
			seen[proto] = true
			protocolOrder = append(protocolOrder, proto)
		}
		byProtocol[proto] = append(byProtocol[proto], m)
	}

	fmt.Printf("%-30s %-40s %-15s %s\n", "NAME", "MODEL", "PROVIDER", "AUTH")
	fmt.Println(strings.Repeat("-", 95))

	for _, proto := range protocolOrder {
		for _, m := range byProtocol[proto] {
			marker := " "
			if m.ModelName == defaultModel {
				marker = "*"
			}
			auth := "configured"
			if m.APIKey() == "" {
				auth = "missing"
			}
			fmt.Printf("%s%-29s %-40s %-15s %s\n",
				marker, m.ModelName, m.Model, proto, auth)
		}
	}

	fmt.Printf("\n* = default model\n")
}

func printModelScan(cfg *config.Config) {
	if len(cfg.ModelList) == 0 {
		fmt.Println("No models configured in model_list.")
		return
	}

	var ready, missing []string
	for _, m := range cfg.ModelList {
		if m.APIKey() != "" {
			ready = append(ready, fmt.Sprintf("  [OK]  %s (%s)", m.ModelName, m.Model))
		} else {
			missing = append(missing, fmt.Sprintf("  [--]  %s (%s)", m.ModelName, m.Model))
		}
	}

	fmt.Printf("Scanned %d models:\n\n", len(cfg.ModelList))

	if len(ready) > 0 {
		fmt.Println("Ready to use:")
		for _, line := range ready {
			fmt.Println(line)
		}
		fmt.Println()
	}

	if len(missing) > 0 {
		fmt.Println("Missing API key:")
		for _, line := range missing {
			fmt.Println(line)
		}
		fmt.Println()
	}

	fmt.Printf("Summary: %d ready, %d missing API key\n", len(ready), len(missing))
}

// newStatusCommand implements "aibhq models status [--check]".
func newStatusCommand() *cobra.Command {
	var checkMode bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show model chain and auth health",
		Long: `Display the resolved primary model, fallback chain, image model, and OAuth credential health.

With --check, exits with:
  0 — all credentials are valid and not expiring soon
  1 — one or more credentials are missing or expired
  2 — one or more OAuth credentials expire within 24 hours`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModelsStatus(checkMode)
		},
	}
	cmd.Flags().BoolVar(&checkMode, "check", false, "Machine-readable exit code (0=ok, 1=expired/missing, 2=expiring)")
	return cmd
}

func runModelsStatus(checkMode bool) error {
	configPath := internal.GetConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	d := cfg.Agents.Defaults
	primary := d.ModelName
	if primary == "" {
		primary = "(none)"
	}

	fmt.Println("Model Status")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("  Primary:    %s\n", primary)

	if len(d.ModelFallbacks) == 0 {
		fmt.Println("  Fallbacks:  (none)")
	} else {
		for i, fb := range d.ModelFallbacks {
			if i == 0 {
				fmt.Printf("  Fallbacks:  %s\n", fb)
			} else {
				fmt.Printf("              %s\n", fb)
			}
		}
	}

	if d.ImageModel != "" {
		fmt.Printf("  Image:      %s\n", d.ImageModel)
		if len(d.ImageModelFallbacks) > 0 {
			fmt.Printf("  Image FB:   %s\n", strings.Join(d.ImageModelFallbacks, ", "))
		}
	}

	// Auth health.
	fmt.Println()
	fmt.Println("Auth Health")
	fmt.Println(strings.Repeat("─", 50))

	store, err := auth.LoadStore()
	if err != nil {
		if !checkMode {
			fmt.Println("  (could not load auth store)")
		}
	}

	exitCode := 0
	now := time.Now()

	if store != nil && len(store.Credentials) > 0 {
		for provider, cred := range store.Credentials {
			tag := ""
			switch {
			case cred.IsExpired():
				tag = " [EXPIRED]"
				if exitCode < 1 {
					exitCode = 1
				}
			case cred.NeedsRefresh() || (!cred.ExpiresAt.IsZero() && cred.ExpiresAt.Sub(now) < 24*time.Hour):
				tag = " [EXPIRING SOON]"
				if exitCode < 2 {
					exitCode = 2
				}
			default:
				tag = " [ok]"
			}
			email := ""
			if cred.Email != "" {
				email = " (" + cred.Email + ")"
			}
			fmt.Printf("  %-20s %s%s%s\n", provider, cred.AuthMethod, tag, email)
		}
	} else {
		fmt.Println("  (no credentials stored — run: aibhq auth login)")
		if exitCode < 1 {
			exitCode = 1
		}
	}

	if checkMode {
		os.Exit(exitCode)
	}
	return nil
}

// newFallbacksCommand implements "aibhq models fallbacks <list|add|remove|clear>".
func newFallbacksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fallbacks",
		Short: "Manage the default model fallback chain",
		Long: `Manage ordered fallback models used when the primary model fails.

Examples:
  aibhq models fallbacks list
  aibhq models fallbacks add claude-haiku-4-5
  aibhq models fallbacks remove claude-haiku-4-5
  aibhq models fallbacks clear`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List fallback chain for the default model",
			RunE: func(cmd *cobra.Command, args []string) error {
				return runFallbacksList()
			},
		},
		&cobra.Command{
			Use:   "add <model-name>",
			Short: "Append a model to the fallback chain",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runFallbacksAdd(args[0])
			},
		},
		&cobra.Command{
			Use:   "remove <model-name>",
			Short: "Remove a model from the fallback chain",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return runFallbacksRemove(args[0])
			},
		},
		&cobra.Command{
			Use:   "clear",
			Short: "Remove all fallbacks from the default model",
			RunE: func(cmd *cobra.Command, args []string) error {
				return runFallbacksClear()
			},
		},
	)

	return cmd
}

func runFallbacksList() error {
	configPath := internal.GetConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	primary := cfg.Agents.Defaults.ModelName
	fallbacks := cfg.Agents.Defaults.ModelFallbacks
	fmt.Printf("Primary: %s\n", primary)
	if len(fallbacks) == 0 {
		fmt.Println("Fallbacks: (none)")
		return nil
	}
	fmt.Println("Fallbacks:")
	for i, fb := range fallbacks {
		fmt.Printf("  %d. %s\n", i+1, fb)
	}
	return nil
}

func runFallbacksAdd(modelName string) error {
	configPath := internal.GetConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	// Prevent duplicates.
	for _, fb := range cfg.Agents.Defaults.ModelFallbacks {
		if fb == modelName {
			fmt.Printf("%s is already in the fallback chain\n", modelName)
			return nil
		}
	}
	cfg.Agents.Defaults.ModelFallbacks = append(cfg.Agents.Defaults.ModelFallbacks, modelName)
	if err := config.SaveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("Added %s to fallback chain\n", modelName)
	return runFallbacksList()
}

func runFallbacksRemove(modelName string) error {
	configPath := internal.GetConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	var updated []string
	found := false
	for _, fb := range cfg.Agents.Defaults.ModelFallbacks {
		if fb == modelName {
			found = true
			continue
		}
		updated = append(updated, fb)
	}
	if !found {
		fmt.Printf("%s is not in the fallback chain\n", modelName)
		return nil
	}
	cfg.Agents.Defaults.ModelFallbacks = updated
	if err := config.SaveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("Removed %s from fallback chain\n", modelName)
	return runFallbacksList()
}

func runFallbacksClear() error {
	configPath := internal.GetConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	count := len(cfg.Agents.Defaults.ModelFallbacks)
	cfg.Agents.Defaults.ModelFallbacks = nil
	if err := config.SaveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("Cleared %d fallback(s)\n", count)
	return nil
}

// extractProtocol returns the protocol prefix from a model identifier.
// For example, "openai/gpt-4o" returns "openai", "anthropic/claude" returns "anthropic".
// If no "/" is present the full string is returned.
func extractProtocol(model string) string {
	if idx := strings.Index(model, "/"); idx >= 0 {
		return model[:idx]
	}
	return model
}
