package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	tuicfg "github.com/raynaythegreat/ai-business-hq/cmd/aibhq-launcher/config"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq-launcher/ui"
)

func NewTUICommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tui [config.json]",
		Short: "Open the terminal UI config editor",
		Long:  "Open the OctAi terminal UI for visual agent and model configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := tuicfg.DefaultConfigPath()
			if len(args) > 0 {
				configPath = args[0]
			}

			configDir := filepath.Dir(configPath)
			if _, err := os.Stat(configDir); os.IsNotExist(err) {
				onboard := exec.Command("aibhq", "onboard")
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
		},
	}
}
