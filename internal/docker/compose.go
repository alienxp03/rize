package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alienxp03/rize/internal/config"
	"gopkg.in/yaml.v3"
)

// ComposeFile represents a docker-compose.yml structure
type ComposeFile struct {
	Version  string                       `yaml:"version,omitempty"`
	Services map[string]ComposeService    `yaml:"services"`
	Networks map[string]ComposeNetwork    `yaml:"networks"`
	Volumes  map[string]ComposeVolume     `yaml:"volumes"`
}

type ComposeService struct {
	Image       string            `yaml:"image"`
	Ports       []string          `yaml:"ports,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	Networks    []string          `yaml:"networks,omitempty"`
	HealthCheck *ComposeHealthCheck `yaml:"healthcheck,omitempty"`
}

type ComposeHealthCheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval"`
	Timeout  string   `yaml:"timeout"`
	Retries  int      `yaml:"retries"`
}

type ComposeNetwork struct {
	Driver string `yaml:"driver,omitempty"`
}

type ComposeVolume struct {
	Driver string `yaml:"driver,omitempty"`
}

// GenerateComposeFile generates a docker-compose.yml from config
func GenerateComposeFile(cfg *config.Config) (*ComposeFile, error) {
	compose := &ComposeFile{
		Services: make(map[string]ComposeService),
		Networks: make(map[string]ComposeNetwork),
		Volumes:  make(map[string]ComposeVolume),
	}

	// Add network
	compose.Networks[cfg.Network.Name] = ComposeNetwork{
		Driver: cfg.Network.Driver,
	}

	// Add services
	for name, svc := range cfg.Services {
		if !svc.Enabled {
			continue
		}

		composeSvc := ComposeService{
			Image:       svc.Image,
			Ports:       svc.Ports,
			Environment: svc.Environment,
			Volumes:     svc.Volumes,
			Networks:    []string{cfg.Network.Name},
		}

		if svc.HealthCheck != nil {
			composeSvc.HealthCheck = &ComposeHealthCheck{
				Test:     svc.HealthCheck.Test,
				Interval: svc.HealthCheck.Interval,
				Timeout:  svc.HealthCheck.Timeout,
				Retries:  svc.HealthCheck.Retries,
			}
		}

		compose.Services[name] = composeSvc
	}

	// Add volumes
	for _, vol := range cfg.Volumes {
		compose.Volumes[vol] = ComposeVolume{}
	}

	return compose, nil
}

// WriteComposeFile writes the compose file to disk
func WriteComposeFile(compose *ComposeFile, path string) error {
	data, err := yaml.Marshal(compose)
	if err != nil {
		return fmt.Errorf("failed to marshal compose file: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write compose file: %w", err)
	}

	return nil
}

// GetComposePath returns the path to the docker-compose.yml file
func GetComposePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "rize")
	composePath := filepath.Join(configDir, "docker-compose.yml")

	return composePath, nil
}

// EnsureCompose generates and writes the docker-compose.yml file
func EnsureCompose(cfg *config.Config) error {
	composePath, err := GetComposePath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(composePath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	compose, err := GenerateComposeFile(cfg)
	if err != nil {
		return err
	}

	return WriteComposeFile(compose, composePath)
}

// ComposeUp starts the services
func ComposeUp(cfg *config.Config) error {
	if err := EnsureCompose(cfg); err != nil {
		return err
	}

	composePath, err := GetComposePath()
	if err != nil {
		return err
	}

	cmd := exec.Command("docker", "compose", "-f", composePath, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ComposeDown stops the services
func ComposeDown() error {
	composePath, err := GetComposePath()
	if err != nil {
		return err
	}

	cmd := exec.Command("docker", "compose", "-f", composePath, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ComposePs lists running services
func ComposePs() error {
	composePath, err := GetComposePath()
	if err != nil {
		return err
	}

	cmd := exec.Command("docker", "compose", "-f", composePath, "ps")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ComposeLogs shows service logs
func ComposeLogs(follow bool) error {
	composePath, err := GetComposePath()
	if err != nil {
		return err
	}

	args := []string{"compose", "-f", composePath, "logs"}
	if follow {
		args = append(args, "-f")
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ComposeRestart restarts services
func ComposeRestart() error {
	composePath, err := GetComposePath()
	if err != nil {
		return err
	}

	cmd := exec.Command("docker", "compose", "-f", composePath, "restart")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
