package config

// GlobalConfig represents global configuration (environment variables only)
// Location: ~/.rize/config.yml
type GlobalConfig struct {
	Environment map[string]string `yaml:"environment"`
}

// ProjectConfig represents per-project configuration
// Location: ~/.rize/projects/{name}-{hash}/config.yml
type ProjectConfig struct {
	Services    map[string]bool   `yaml:"services"`    // Service name -> enabled/disabled
	Environment map[string]string `yaml:"environment"` // Overrides global environment
}

// Config represents the merged configuration (global + project)
type Config struct {
	Services    map[string]bool   // Merged service toggles
	Environment map[string]string // Merged environment (project overrides global)
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Name   string
	Driver string
}
