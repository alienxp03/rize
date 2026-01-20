package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/alienxp03/rize/internal/docker"
	"github.com/alienxp03/rize/internal/ui"
)

const (
	installPath     = "/usr/local/bin/rize"
	githubReleaseURL = "https://github.com/alienxp03/rize/releases/latest/download/rize-%s-%s"
)

// Install installs rize to /usr/local/bin
func Install() error {
	ui.Info("Installing rize to %s", installPath)

	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Check if we can write to install directory
	installDir := filepath.Dir(installPath)
	if !isWritable(installDir) {
		// Need sudo
		ui.Warning("Installing to %s requires elevated permissions", installPath)
		return installWithSudo(exePath)
	}

	// Copy file
	if err := copyFile(exePath, installPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Make executable
	if err := os.Chmod(installPath, 0755); err != nil {
		return fmt.Errorf("failed to make executable: %w", err)
	}

	ui.Success("Installed rize to %s", installPath)
	return nil
}

// Update updates rize and pulls the latest image
func Update() error {
	ui.Info("Updating rize image...")

	client, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.PullImage(); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	ui.Success("Image updated")

	// Try to update binary if installed
	if _, err := os.Stat(installPath); err == nil {
		ui.Info("Updating rize binary...")
		if err := downloadLatestBinary(); err != nil {
			ui.Warning("Failed to update binary: %v", err)
			ui.Info("To update manually, run: rize install")
		} else {
			ui.Success("Binary updated")
		}
	}

	return nil
}

// Uninstall removes rize
func Uninstall() error {
	ui.Warning("This will remove the rize image and binary")
	ui.Info("Continue? [y/N]")

	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "y" && confirm != "Y" {
		ui.Info("Uninstall cancelled")
		return nil
	}

	// Remove image
	ui.Info("Removing Docker image...")
	client, err := docker.NewClient()
	if err == nil {
		if err := client.RemoveImage(); err != nil {
			ui.Warning("Failed to remove image: %v", err)
		} else {
			ui.Success("Image removed")
		}
		client.Close()
	}

	// Remove binary
	if _, err := os.Stat(installPath); err == nil {
		ui.Info("Removing binary...")
		if !isWritable(filepath.Dir(installPath)) {
			cmd := exec.Command("sudo", "rm", "-f", installPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				ui.Warning("Failed to remove binary: %v", err)
			} else {
				ui.Success("Binary removed")
			}
		} else {
			if err := os.Remove(installPath); err != nil {
				ui.Warning("Failed to remove binary: %v", err)
			} else {
				ui.Success("Binary removed")
			}
		}
	}

	ui.Success("Uninstall complete")
	return nil
}

// Helper functions

func isWritable(path string) bool {
	file, err := os.OpenFile(filepath.Join(path, ".rize-test"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(filepath.Join(path, ".rize-test"))
	return true
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func installWithSudo(exePath string) error {
	cmd := exec.Command("sudo", "install", "-m", "755", exePath, installPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install with sudo: %w", err)
	}

	ui.Success("Installed rize to %s", installPath)
	return nil
}

func downloadLatestBinary() error {
	url := fmt.Sprintf(githubReleaseURL, runtime.GOOS, runtime.GOARCH)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: %s", resp.Status)
	}

	tmpFile, err := os.CreateTemp("", "rize-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return err
	}

	tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return err
	}

	// Install
	if !isWritable(filepath.Dir(installPath)) {
		cmd := exec.Command("sudo", "install", "-m", "755", tmpFile.Name(), installPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return copyFile(tmpFile.Name(), installPath)
}
