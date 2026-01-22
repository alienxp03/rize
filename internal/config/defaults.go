package config

// DefaultGlobalConfig returns the default global configuration
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Environment: map[string]string{
			"ANTHROPIC_API_KEY": "",
			"OPENAI_API_KEY":    "",
			"GOOGLE_API_KEY":    "",
		},
	}
}

// DefaultProjectConfig returns the default project configuration
func DefaultProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		Services: map[string]bool{
			"postgres":   false, // Disabled by default
			"redis":      false, // Disabled by default
			"mitmproxy":  true,  // Enabled by default
			"playwright": true,  // Enabled by default
		},
		Environment: map[string]string{},
	}
}

// DefaultNetworkConfig returns the default network configuration
func DefaultNetworkConfig() NetworkConfig {
	return NetworkConfig{
		Name:   "rize",
		Driver: "bridge",
	}
}
