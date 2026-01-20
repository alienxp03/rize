package main

import (
	"fmt"
	"os"

	"github.com/alienxp03/rize/internal/commands"
	"github.com/alienxp03/rize/internal/ui"
)

func main() {
	if err := run(); err != nil {
		ui.Error("%v", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]

	// Default to help if no args
	if len(args) == 0 {
		commands.Help()
		return nil
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "shell":
		return commands.Shell()

	case "claude":
		return commands.Agent("claude", commandArgs)

	case "codex":
		return commands.Agent("codex", commandArgs)

	case "opencode":
		return commands.Agent("opencode", commandArgs)

	case "gemini":
		return commands.Agent("gemini", commandArgs)

	case "exec":
		if len(commandArgs) == 0 {
			return fmt.Errorf("exec requires a command")
		}
		return commands.Exec(commandArgs)

	case "services":
		if len(commandArgs) == 0 {
			return fmt.Errorf("services requires a subcommand (up, down, ps, logs, restart)")
		}
		return handleServicesCommand(commandArgs)

	case "init":
		return commands.Init()

	case "install":
		return commands.Install()

	case "update":
		return commands.Update()

	case "uninstall":
		return commands.Uninstall()

	case "help":
		commands.Help()
		return nil

	default:
		ui.Error("Unknown command: %s", command)
		fmt.Println()
		commands.Help()
		return fmt.Errorf("unknown command: %s", command)
	}
}

func handleServicesCommand(args []string) error {
	subcommand := args[0]
	subcommandArgs := args[1:]

	switch subcommand {
	case "up":
		return commands.ServicesUp()

	case "down":
		return commands.ServicesDown()

	case "ps":
		return commands.ServicesPs()

	case "logs":
		follow := false
		if len(subcommandArgs) > 0 && subcommandArgs[0] == "-f" {
			follow = true
		}
		return commands.ServicesLogs(follow)

	case "restart":
		return commands.ServicesRestart()

	default:
		return fmt.Errorf("unknown services subcommand: %s", subcommand)
	}
}
