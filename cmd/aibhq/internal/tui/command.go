package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	tuicfg "github.com/raynaythegreat/ai-business-hq/cmd/aibhq-launcher/config"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq-launcher/ui"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/agent"
)

func NewTUICommand() *cobra.Command {
	var configMode bool

	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Interactive AI chat (use --config for the configuration dashboard)",
		Long:  "Open the OctAi interactive AI chat. Pass --config to open the visual agent and model configuration dashboard instead.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if configMode {
				return runConfigDashboard(args)
			}
			// Run agent chat directly (avoids subprocess/terminal ownership issues)
			return agent.RunInteractive("cli:default")
		},
	}

	cmd.Flags().BoolVar(&configMode, "config", false, "Open the configuration dashboard")

	return cmd
}

func runConfigDashboard(args []string) error {
	configPath := tuicfg.DefaultConfigPath()
	if len(args) > 0 {
		configPath = args[0]
	}

	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		onboard := exec.Command("octai", "onboard")
		onboard.Stdin = os.Stdin
		onboard.Stdout = os.Stdout
		onboard.Stderr = os.Stderr
		_ = onboard.Run()
	}

	cfg, err := tuicfg.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	app := ui.New(cfg, configPath)
	app.OnModelSelected = func(scheme tuicfg.Scheme, user tuicfg.User, modelID string) {
		_ = tuicfg.SyncSelectedModelToMainConfig(scheme, user, modelID)
	}
	return app.Run()
}
