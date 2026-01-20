package config

// Config represents the main configuration structure
type Config struct {
	Services    map[string]Service    `yaml:"services"`
	Environment map[string]string     `yaml:"environment"`
	Network     NetworkConfig         `yaml:"network"`
	Volumes     []string              `yaml:"volumes"`
}

// Service represents a docker compose service
type Service struct {
	Enabled     bool              `yaml:"enabled"`
	Image       string            `yaml:"image"`
	Ports       []string          `yaml:"ports"`
	Environment map[string]string `yaml:"environment"`
	Volumes     []string          `yaml:"volumes"`
	HealthCheck *HealthCheck      `yaml:"healthcheck,omitempty"`
}

// HealthCheck represents a service health check
type HealthCheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval"`
	Timeout  string   `yaml:"timeout"`
	Retries  int      `yaml:"retries"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	Name   string `yaml:"name"`
	Driver string `yaml:"driver"`
}
