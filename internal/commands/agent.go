package commands

import (
	"github.com/alienxp03/rize/internal/config"
	"github.com/alienxp03/rize/internal/docker"
	"github.com/alienxp03/rize/internal/ui"
)

// Agent runs a specific AI agent
func Agent(name string, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Auto-start enabled services
	if err := autoStartServices(cfg); err != nil {
		ui.Warning("Failed to start services: %v", err)
	}

	ui.Info("Running %s...", name)

	client, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	// Build command based on agent
	cmd := buildAgentCommand(name, args)

	return client.RunContainer(cfg, cmd, true)
}

// buildAgentCommand builds the command for the specific agent
func buildAgentCommand(name string, args []string) []string {
	switch name {
	case "claude":
		cmd := []string{"claude", "--dangerously-skip-permissions"}
		return append(cmd, args...)
	case "codex":
		cmd := []string{"codex"}
		return append(cmd, args...)
	case "opencode":
		cmd := []string{"opencode"}
		return append(cmd, args...)
	case "gemini":
		cmd := []string{"gemini"}
		return append(cmd, args...)
	default:
		return []string{name}
	}
}
