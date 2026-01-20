package commands

import (
	"fmt"
	"os"

	"github.com/alienxp03/rize/internal/config"
	"github.com/alienxp03/rize/internal/ui"
)

// Init initializes the rize configuration
func Init() error {
	configPath, err := config.ConfigPath()
	if err != nil {
		return err
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		ui.Warning("Config file already exists at %s", configPath)
		return nil
	}

	ui.Info("Creating config file at %s", configPath)

	cfg := config.DefaultConfig()
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.Success("Config file created successfully")
	ui.Info("Edit the config file to customize your environment")
	return nil
}
