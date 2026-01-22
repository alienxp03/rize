package commands

import (
	"fmt"
	"os"

	"github.com/alienxp03/rize/internal/config"
	"github.com/alienxp03/rize/internal/ui"
)

// Init initializes the rize configuration
func Init() error {
	// Initialize global config
	globalPath, err := config.GlobalConfigPath()
	if err != nil {
		return err
	}

	// Check if global config already exists
	if _, err := os.Stat(globalPath); err == nil {
		ui.Warning("Global config file already exists at %s", globalPath)
	} else {
		ui.Info("Creating global config file at %s", globalPath)
		cfg := config.DefaultGlobalConfig()
		if err := config.SaveGlobal(cfg); err != nil {
			return fmt.Errorf("failed to save global config: %w", err)
		}
		ui.Success("Global config file created successfully")
	}

	// Initialize per-project config
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	projectPath, err := config.ProjectConfigPath(cwd)
	if err != nil {
		return err
	}

	// Check if project config already exists
	if _, err := os.Stat(projectPath); err == nil {
		ui.Warning("Project config file already exists at %s", projectPath)
	} else {
		ui.Info("Creating project config file at %s", projectPath)
		cfg := config.DefaultProjectConfig()
		if err := config.SaveProject(cfg, cwd); err != nil {
			return fmt.Errorf("failed to save project config: %w", err)
		}
		ui.Success("Project config file created successfully")
	}

	ui.Info("Edit the config files to customize your environment")
	return nil
}
