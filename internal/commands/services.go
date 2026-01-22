package commands

import (
	"fmt"
	"os"

	"github.com/alienxp03/rize/internal/config"
	"github.com/alienxp03/rize/internal/docker"
	"github.com/alienxp03/rize/internal/ui"
)

// ServicesUp starts all enabled services
func ServicesUp() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return err
	}

	enabledServices := cfg.GetEnabledServices()
	if len(enabledServices) == 0 {
		ui.Info("No services enabled")
		return nil
	}

	ui.Info("Starting services: %v", enabledServices)

	if err := docker.ComposeUp(cfg); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	ui.Success("Services started")
	return nil
}

// ServicesDown stops all services
func ServicesDown() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return err
	}

	if err := docker.EnsureCompose(cfg); err != nil {
		return err
	}

	ui.Info("Stopping services...")

	if err := docker.ComposeDown(); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	ui.Success("Services stopped")
	return nil
}

// ServicesPs lists running services
func ServicesPs() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return err
	}

	if err := docker.EnsureCompose(cfg); err != nil {
		return err
	}

	return docker.ComposePs()
}

// ServicesLogs shows service logs
func ServicesLogs(follow bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return err
	}

	if err := docker.EnsureCompose(cfg); err != nil {
		return err
	}

	return docker.ComposeLogs(follow)
}

// ServicesRestart restarts services
func ServicesRestart() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return err
	}

	if err := docker.EnsureCompose(cfg); err != nil {
		return err
	}

	ui.Info("Restarting services...")

	if err := docker.ComposeRestart(); err != nil {
		return fmt.Errorf("failed to restart services: %w", err)
	}

	ui.Success("Services restarted")
	return nil
}

// autoStartServices automatically starts enabled services if they're not running
func autoStartServices(cfg *config.Config) error {
	enabledServices := cfg.GetEnabledServices()
	if len(enabledServices) == 0 {
		return nil
	}

	// Try to start services, but don't fail if it doesn't work
	// This will fail silently as per requirements
	docker.ComposeUpQuiet(cfg)

	return nil
}
