package config

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// GlobalConfigPath returns the path to the global config file
// Location: ~/.rize/config.yml
func GlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".rize")
	configFile := filepath.Join(configDir, "config.yml")

	return configFile, nil
}

// ProjectConfigPath returns the path to the per-project config file
// Location: ~/.rize/projects/{name}-{hash}/config.yml
func ProjectConfigPath(projectPath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		absPath = projectPath
	}

	projectName := filepath.Base(absPath)
	safeName := sanitizeProjectName(projectName)
	hash := shortHash(absPath)

	projectDir := filepath.Join(home, ".rize", "projects", fmt.Sprintf("%s-%s", safeName, hash))
	configFile := filepath.Join(projectDir, "config.yml")

	return configFile, nil
}

// ProjectStateDir returns the path to the per-project state directory
// Location: ~/.rize/projects/{name}-{hash}/state
func ProjectStateDir(projectPath string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		absPath = projectPath
	}

	projectName := filepath.Base(absPath)
	safeName := sanitizeProjectName(projectName)
	hash := shortHash(absPath)

	projectDir := filepath.Join(home, ".rize", "projects", fmt.Sprintf("%s-%s", safeName, hash))
	stateDir := filepath.Join(projectDir, "state")

	return stateDir, nil
}

func sanitizeProjectName(name string) string {
	name = strings.ToLower(name)
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('-')
	}

	safe := strings.Trim(b.String(), "-")
	if safe == "" {
		return "project"
	}

	return safe
}

func shortHash(value string) string {
	sum := sha1.Sum([]byte(value))
	return hex.EncodeToString(sum[:])[:6]
}

// LoadGlobal loads the global configuration from ~/.rize/config.yml
// If the file doesn't exist, it creates it with default configuration
func LoadGlobal() (*GlobalConfig, error) {
	configFile, err := GlobalConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := DefaultGlobalConfig()
		if err := SaveGlobal(cfg); err != nil {
			// If we can't save, just return the default config in memory
			return cfg, nil
		}
		return cfg, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read global config file: %w", err)
	}

	var cfg GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse global config file: %w", err)
	}

	// Ensure environment map is initialized
	if cfg.Environment == nil {
		cfg.Environment = make(map[string]string)
	}

	return &cfg, nil
}

// LoadProject loads the per-project configuration
// If the file doesn't exist, it creates it with default configuration
func LoadProject(projectPath string) (*ProjectConfig, error) {
	configFile, err := ProjectConfigPath(projectPath)
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := DefaultProjectConfig()
		if err := SaveProject(cfg, projectPath); err != nil {
			// If we can't save, just return the default config in memory
			return cfg, nil
		}
		return cfg, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config file: %w", err)
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse project config file: %w", err)
	}

	// Merge with defaults for missing services
	cfg = *mergeProjectWithDefaults(&cfg)

	return &cfg, nil
}

// Load loads and merges global + project configuration
func Load(projectPath string) (*Config, error) {
	global, err := LoadGlobal()
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	project, err := LoadProject(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}

	return MergeConfigs(global, project), nil
}

// MergeConfigs merges global and project configurations
// Project environment variables override global ones
func MergeConfigs(global *GlobalConfig, project *ProjectConfig) *Config {
	merged := &Config{
		Services:    make(map[string]bool),
		Environment: make(map[string]string),
	}

	// Copy services from project
	for name, enabled := range project.Services {
		merged.Services[name] = enabled
	}

	// Merge environment: global first, then project overrides
	for key, value := range global.Environment {
		merged.Environment[key] = value
	}
	for key, value := range project.Environment {
		merged.Environment[key] = value
	}

	return merged
}

// SaveGlobal saves the global configuration
func SaveGlobal(cfg *GlobalConfig) error {
	configFile, err := GlobalConfigPath()
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
		return fmt.Errorf("failed to marshal global config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write global config file: %w", err)
	}

	return nil
}

// SaveProject saves the per-project configuration
func SaveProject(cfg *ProjectConfig, projectPath string) error {
	configFile, err := ProjectConfigPath(projectPath)
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create project config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write project config file: %w", err)
	}

	return nil
}

// mergeProjectWithDefaults merges the loaded project config with default values
func mergeProjectWithDefaults(cfg *ProjectConfig) *ProjectConfig {
	defaults := DefaultProjectConfig()

	// Ensure services map is initialized
	if cfg.Services == nil {
		cfg.Services = make(map[string]bool)
	}

	// Fill in missing services with defaults
	for name, enabled := range defaults.Services {
		if _, exists := cfg.Services[name]; !exists {
			cfg.Services[name] = enabled
		}
	}

	// Ensure environment map is initialized
	if cfg.Environment == nil {
		cfg.Environment = make(map[string]string)
	}

	return cfg
}

// GetEnabledServices returns a list of enabled service names
func (c *Config) GetEnabledServices() []string {
	var enabled []string
	for name, isEnabled := range c.Services {
		if isEnabled {
			enabled = append(enabled, name)
		}
	}
	return enabled
}
