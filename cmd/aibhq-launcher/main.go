// OctAi - Ultra-lightweight personal AI agent
// License: MIT
//
// Copyright (c) 2026 OctAi contributors

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tuicfg "github.com/raynaythegreat/ai-business-hq/cmd/aibhq-launcher/config"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq-launcher/ui"
)

func main() {
	configPath := tuicfg.DefaultConfigPath()
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		cmd := exec.Command("octai", "onboard")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	}

	cfg, err := tuicfg.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aibhq-launcher: %v\n", err)
		os.Exit(1)
	}

	app := ui.New(cfg, configPath)
	// Bind model selection hook to sync to main config
	app.OnModelSelected = func(scheme tuicfg.Scheme, user tuicfg.User, modelID string) {
		_ = tuicfg.SyncSelectedModelToMainConfig(scheme, user, modelID)
	}
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "aibhq-launcher: %v\n", err)
		os.Exit(1)
	}
}
