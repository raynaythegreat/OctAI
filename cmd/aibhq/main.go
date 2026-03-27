// OctAi - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 OctAi contributors

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/agent"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/auth"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/cron"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/gateway"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/loop"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/migrate"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/model"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/models"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/onboard"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/skills"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/status"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/tui"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/version"
	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal/web"
	"github.com/raynaythegreat/ai-business-hq/pkg/config"
)

func NewPicoclawCommand() *cobra.Command {
	short := fmt.Sprintf("%s OctAi - Personal AI Assistant v%s\n\n", internal.Logo, config.GetVersion())

	cmd := &cobra.Command{
		Use:     "aibhq",
		Short:   short,
		Example: "aibhq version",
	}

	cmd.AddCommand(
		onboard.NewOnboardCommand(),
		agent.NewAgentCommand(),
		auth.NewAuthCommand(),
		gateway.NewGatewayCommand(),
		status.NewStatusCommand(),
		cron.NewCronCommand(),
		loop.NewLoopCommand(),
		migrate.NewMigrateCommand(),
		skills.NewSkillsCommand(),
		model.NewModelCommand(),
		models.NewModelsCommand(),
		web.NewWebCommand(),
		tui.NewTUICommand(),
		version.NewVersionCommand(),
	)

	return cmd
}

const (
	colorPurple = "\033[1;38;2;168;85;247m"
	banner      = "\r\n" +
		colorPurple + " ██████╗  ██████╗████████╗ █████╗ ██╗\n" +
		colorPurple + "██╔═══██╗██╔════╝╚══██╔══╝██╔══██╗██║\n" +
		colorPurple + "██║   ██║██║        ██║   ███████║██║\n" +
		colorPurple + "╚██████╔╝╚██████╗   ██║   ██╔══██║██║\n" +
		colorPurple + " ╚═════╝  ╚═════╝   ╚═╝   ╚═╝  ╚═╝╚═╝\n" +
		"\033[0m\r\n"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "agent" {
		fmt.Printf("%s", banner)
	}
	cmd := NewPicoclawCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
