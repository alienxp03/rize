package docker

import (
	"fmt"
	"io"
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
	Image       string              `yaml:"image"`
	Command     []string            `yaml:"command,omitempty"`
	Ports       []string            `yaml:"ports,omitempty"`
	Environment map[string]string   `yaml:"environment,omitempty"`
	Volumes     []string            `yaml:"volumes,omitempty"`
	Networks    []string            `yaml:"networks,omitempty"`
	HealthCheck *ComposeHealthCheck `yaml:"healthcheck,omitempty"`
}

type ComposeHealthCheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval"`
	Timeout  string   `yaml:"timeout"`
	Retries  int      `yaml:"retries"`
}

type ComposeNetwork struct {
	Name   string `yaml:"name,omitempty"`
	Driver string `yaml:"driver,omitempty"`
}

type ComposeVolume struct {
	Driver string `yaml:"driver,omitempty"`
}

// hardcodedServiceDefinitions returns all available service definitions
func hardcodedServiceDefinitions() map[string]ComposeService {
	networkName := config.DefaultNetworkConfig().Name

	return map[string]ComposeService{
		"playwright": {
			Image: "mcr.microsoft.com/playwright:v1.40.0",
			Ports: []string{"8381:3000"},
			Environment: map[string]string{
				"PLAYWRIGHT_BROWSERS_PATH": "/ms-playwright",
			},
			Networks: []string{networkName},
		},
		"postgres": {
			Image: "postgres:16-alpine",
			Environment: map[string]string{
				"POSTGRES_PASSWORD": "dev",
				"POSTGRES_USER":     "dev",
				"POSTGRES_DB":       "dev",
			},
			Volumes:  []string{"rize-postgres:/var/lib/postgresql/data"},
			Networks: []string{networkName},
			HealthCheck: &ComposeHealthCheck{
				Test:     []string{"CMD-SHELL", "pg_isready -U dev"},
				Interval: "5s",
				Timeout:  "5s",
				Retries:  5,
			},
		},
		"redis": {
			Image:    "redis:7-alpine",
			Volumes:  []string{"rize-redis:/data"},
			Networks: []string{networkName},
			HealthCheck: &ComposeHealthCheck{
				Test:     []string{"CMD", "redis-cli", "ping"},
				Interval: "5s",
				Timeout:  "3s",
				Retries:  5,
			},
		},
		"mitmproxy": {
			Image:    "mitmproxy/mitmproxy:latest",
			Ports:    []string{"8080:8080", "8081:8081"},
			Volumes:  []string{"rize-mitmproxy:/home/mitmproxy/.mitmproxy"},
			Networks: []string{networkName},
			Command: []string{
				"/bin/sh",
				"-c",
				`cat > /tmp/rize-noauth.py <<'PY'
from mitmproxy import ctx

class DisableWebAuth:
    def running(self):
        app = getattr(ctx.master, "app", None)
        if app:
            app.settings["is_valid_password"] = lambda _password: True

addons = [DisableWebAuth()]
PY
exec mitmweb --web-host 0.0.0.0 --set block_global=false --set web_password= --set web_open_browser=false -s /tmp/rize-noauth.py`,
			},
		},
	}
}

// hardcodedVolumes returns all volumes used by services
func hardcodedVolumes() []string {
	return []string{
		"rize-postgres",
		"rize-redis",
		"rize-mitmproxy",
	}
}

// GenerateComposeFile generates a docker-compose.yml from config
func GenerateComposeFile(cfg *config.Config) (*ComposeFile, error) {
	compose := &ComposeFile{
		Services: make(map[string]ComposeService),
		Networks: make(map[string]ComposeNetwork),
		Volumes:  make(map[string]ComposeVolume),
	}

	// Add network
	networkCfg := config.DefaultNetworkConfig()
	compose.Networks[networkCfg.Name] = ComposeNetwork{
		Name:   networkCfg.Name,
		Driver: networkCfg.Driver,
	}

	// Get all service definitions
	allServices := hardcodedServiceDefinitions()

	// Add only enabled services from config
	for name, enabled := range cfg.Services {
		if !enabled {
			continue
		}

		if svc, exists := allServices[name]; exists {
			compose.Services[name] = svc
		}
	}

	// Add volumes
	for _, vol := range hardcodedVolumes() {
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

// ComposeUpQuiet starts services without emitting compose output
func ComposeUpQuiet(cfg *config.Config) error {
	if err := EnsureCompose(cfg); err != nil {
		return err
	}

	composePath, err := GetComposePath()
	if err != nil {
		return err
	}

	cmd := exec.Command("docker", "compose", "-f", composePath, "up", "-d")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

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
