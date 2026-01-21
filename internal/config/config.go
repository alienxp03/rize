package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "rize")
	configFile := filepath.Join(configDir, "config.yml")

	return configFile, nil
}

// Load loads the configuration from the config file
// If the file doesn't exist, it creates it with default configuration
func Load() (*Config, error) {
	configFile, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := DefaultConfig()
		if err := Save(cfg); err != nil {
			// If we can't save, just return the default config in memory
			return cfg, nil
		}
		return cfg, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge with defaults for missing fields
	merged := mergeWithDefaults(&cfg)
	if normalizeLegacyDefaults(merged) {
		_ = Save(merged)
	}

	return merged, nil
}

// Save saves the configuration to the config file
func Save(cfg *Config) error {
	configFile, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// mergeWithDefaults merges the loaded config with default values
func mergeWithDefaults(cfg *Config) *Config {
	defaults := DefaultConfig()

	// Merge services
	if cfg.Services == nil {
		cfg.Services = defaults.Services
	} else {
		// Fill in missing services with defaults
		for name, defaultSvc := range defaults.Services {
			if svc, exists := cfg.Services[name]; exists {
				cfg.Services[name] = mergeServiceWithDefaults(svc, defaultSvc)
				continue
			}
			cfg.Services[name] = defaultSvc
		}
	}

	// Merge environment
	if cfg.Environment == nil {
		cfg.Environment = make(map[string]string)
	}

	// Merge network config
	if cfg.Network.Name == "" {
		cfg.Network = defaults.Network
	}

	// Merge volumes
	if len(cfg.Volumes) == 0 {
		cfg.Volumes = defaults.Volumes
	}

	return cfg
}

func mergeServiceWithDefaults(svc Service, defaults Service) Service {
	if svc.Image == "" {
		svc.Image = defaults.Image
	}

	if svc.Ports == nil {
		svc.Ports = defaults.Ports
	}

	if svc.Command == nil {
		svc.Command = defaults.Command
	}

	if svc.Environment == nil {
		svc.Environment = defaults.Environment
	} else {
		for key, value := range defaults.Environment {
			if _, exists := svc.Environment[key]; !exists {
				svc.Environment[key] = value
			}
		}
	}

	if svc.Volumes == nil {
		svc.Volumes = defaults.Volumes
	}

	if svc.HealthCheck == nil {
		svc.HealthCheck = defaults.HealthCheck
	}

	return svc
}

func normalizeLegacyDefaults(cfg *Config) bool {
	defaults := DefaultConfig()
	legacyPorts := map[string][]string{
		"playwright": {"3000:3000"},
		"postgres":   {"5432:5432"},
		"redis":      {"6379:6379"},
	}
	legacyCommands := map[string][][]string{
		"mitmproxy": {
			{"mitmweb", "--web-host", "0.0.0.0", "--set", "block_global=false"},
			{"mitmweb", "--web-host", "0.0.0.0", "--set", "block_global=false", "--set", "web_username=", "--set", "web_password="},
		},
	}

	changed := false
	for name, legacy := range legacyPorts {
		svc, exists := cfg.Services[name]
		if !exists {
			continue
		}

		if portsEqual(svc.Ports, legacy) {
			svc.Ports = defaults.Services[name].Ports
			cfg.Services[name] = svc
			changed = true
		}
	}

	for name, legacyList := range legacyCommands {
		svc, exists := cfg.Services[name]
		if !exists {
			continue
		}

		for _, legacy := range legacyList {
			if commandsEqual(svc.Command, legacy) {
				svc.Command = defaults.Services[name].Command
				cfg.Services[name] = svc
				changed = true
				break
			}
		}
	}

	return changed
}

func portsEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func commandsEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// GetEnabledServices returns a list of enabled service names
func (c *Config) GetEnabledServices() []string {
	var enabled []string
	for name, svc := range c.Services {
		if svc.Enabled {
			enabled = append(enabled, name)
		}
	}
	return enabled
}
