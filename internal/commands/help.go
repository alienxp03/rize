package commands

import (
	"fmt"

	"github.com/alienxp03/rize/internal/ui"
)

// Help displays the help message
func Help() {
	fmt.Printf("%s - Secure AI Agent Sandbox\n\n", ui.Blue("Rize"))

	fmt.Println("Usage: rize [command] [args...]")
	fmt.Println()

	fmt.Println("Commands:")
	fmt.Println("  shell              Start interactive shell (zsh)")
	fmt.Println("  claude [args...]   Run Claude Code agent")
	fmt.Println("  codex [args...]    Run OpenAI Codex agent")
	fmt.Println("  opencode [args...] Run OpenCode agent")
	fmt.Println("  gemini [args...]   Run Gemini agent")
	fmt.Println("  exec <cmd...>      Run a shell command directly")
	fmt.Println()

	fmt.Println("Service Management:")
	fmt.Println("  services up        Start all enabled services")
	fmt.Println("  services down      Stop all services")
	fmt.Println("  services ps        List running services")
	fmt.Println("  services logs      View service logs")
	fmt.Println("  services restart   Restart services")
	fmt.Println()

	fmt.Println("Configuration:")
	fmt.Println("  init               Create default config file")
	fmt.Println()

	fmt.Println("Installation:")
	fmt.Println("  install            Install rize to /usr/local/bin")
	fmt.Println("  update             Update image and binary")
	fmt.Println("  uninstall          Remove rize")
	fmt.Println()

	fmt.Println("Other:")
	fmt.Println("  help               Show this help message")
	fmt.Println()

	fmt.Println("Environment Variables:")
	fmt.Println("  RIZE_IMAGE         Docker image to use (default: alienxp03/rize:latest)")
	fmt.Println()

	fmt.Println("Config Files:")
	fmt.Println("  ~/.rize/config.yml                        Global environment variables")
	fmt.Println("  ~/.rize/projects/{name}/config.yml        Per-project config (services + env)")
	fmt.Println("  ~/.config/rize/docker-compose.yml         Shared services")
	fmt.Println()
}
