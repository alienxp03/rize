package commands

import (
	"github.com/alienxp03/rize/internal/config"
	"github.com/alienxp03/rize/internal/docker"
	"github.com/alienxp03/rize/internal/ui"
)

// Shell starts an interactive shell
func Shell() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Auto-start enabled services
	if err := autoStartServices(cfg); err != nil {
		ui.Warning("Failed to start services: %v", err)
	}

	ui.Info("Starting shell...")

	client, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	return client.RunContainer(cfg, []string{"/bin/zsh"}, true)
}
