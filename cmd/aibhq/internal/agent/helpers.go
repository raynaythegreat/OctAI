package agent

import (
	"context"
	"fmt"

	"github.com/raynaythegreat/ai-business-hq/cmd/aibhq/internal"
	"github.com/raynaythegreat/ai-business-hq/pkg/agent"
	"github.com/raynaythegreat/ai-business-hq/pkg/bus"
	"github.com/raynaythegreat/ai-business-hq/pkg/logger"
	"github.com/raynaythegreat/ai-business-hq/pkg/providers"
)

func agentCmd(message, sessionKey, model string, debug bool) error {
	if sessionKey == "" {
		sessionKey = "cli:default"
	}

	cfg, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	logger.ConfigureFromEnv()

	if debug {
		logger.SetLevel(logger.DEBUG)
		fmt.Println("🔍 Debug mode enabled")
	}

	if model != "" {
		cfg.Agents.Defaults.ModelName = model
	}

	provider, modelID, err := providers.CreateProvider(cfg)
	if err != nil {
		return fmt.Errorf("error creating provider: %w", err)
	}

	// Use the resolved model ID from provider creation
	if modelID != "" {
		cfg.Agents.Defaults.ModelName = modelID
	}

	msgBus := bus.NewMessageBus()
	defer msgBus.Close()
	agentLoop := agent.NewAgentLoop(cfg, msgBus, provider)
	defer agentLoop.Close()

	// Print agent startup info (only for interactive mode)
	startupInfo := agentLoop.GetStartupInfo()
	logger.InfoCF("agent", "Agent initialized",
		map[string]any{
			"tools_count":      startupInfo["tools"].(map[string]any)["count"],
			"skills_total":     startupInfo["skills"].(map[string]any)["total"],
			"skills_available": startupInfo["skills"].(map[string]any)["available"],
		})

	if message != "" {
		ctx := context.Background()
		response, err := agentLoop.ProcessDirect(ctx, message, sessionKey)
		if err != nil {
			return fmt.Errorf("error processing message: %w", err)
		}
		fmt.Printf("\n%s %s\n", internal.Logo, response)
		return nil
	}

	return interactiveMode(agentLoop, sessionKey, cfg.Agents.Defaults.ModelName)
}

func interactiveMode(agentLoop *agent.AgentLoop, sessionKey string, modelName string) error {
	ctx := context.Background()
	ui := newChatUI(modelName, sessionKey, agentLoop)
	return ui.run(ctx, agentLoop)
}

func simpleInteractiveMode(agentLoop *agent.AgentLoop, sessionKey string) error {
	return interactiveMode(agentLoop, sessionKey, "OctAi")
}

// RunInteractive starts the interactive agent chat UI.
// Called directly by the tui command to avoid subprocess issues.
func RunInteractive(sessionKey string) error {
	return agentCmd("", sessionKey, "", false)
}
