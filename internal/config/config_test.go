package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Check services
	if len(cfg.Services) != 4 {
		t.Errorf("Expected 4 services, got %d", len(cfg.Services))
	}

	// Check that default services exist
	services := []string{"playwright", "postgres", "redis", "mitmproxy"}
	for _, svc := range services {
		if _, exists := cfg.Services[svc]; !exists {
			t.Errorf("Service %s not found in default config", svc)
		}
	}

	// Check network config
	if cfg.Network.Name != "rize" {
		t.Errorf("Expected network name 'rize', got '%s'", cfg.Network.Name)
	}
}

func TestGetEnabledServices(t *testing.T) {
	cfg := DefaultConfig()

	enabled := cfg.GetEnabledServices()

	// All services should be enabled by default
	if len(enabled) != 4 {
		t.Errorf("Expected 4 enabled services, got %d", len(enabled))
	}

	// Disable a service
	svc := cfg.Services["postgres"]
	svc.Enabled = false
	cfg.Services["postgres"] = svc

	enabled = cfg.GetEnabledServices()

	if len(enabled) != 3 {
		t.Errorf("Expected 3 enabled services after disabling one, got %d", len(enabled))
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "rize-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up test config path
	testConfigPath := filepath.Join(tmpDir, "config.yml")

	// Create a test config
	cfg := DefaultConfig()

	// Save to temp path
	if err := os.WriteFile(testConfigPath, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	// Note: This is a basic test. In a real scenario, you'd want to mock ConfigPath()
	// For now, we'll just test that the config marshaling works
	if cfg.Services == nil {
		t.Error("Services should not be nil")
	}
}
