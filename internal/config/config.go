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
	return mergeWithDefaults(&cfg), nil
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
		for name, svc := range defaults.Services {
			if _, exists := cfg.Services[name]; !exists {
				cfg.Services[name] = svc
			}
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
