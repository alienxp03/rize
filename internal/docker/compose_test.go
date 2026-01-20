package docker

import (
	"testing"

	"github.com/alienxp03/rize/internal/config"
)

func TestGenerateComposeFile(t *testing.T) {
	cfg := config.DefaultConfig()

	compose, err := GenerateComposeFile(cfg)
	if err != nil {
		t.Fatalf("Failed to generate compose file: %v", err)
	}

	// Check services
	if len(compose.Services) != 4 {
		t.Errorf("Expected 4 services, got %d", len(compose.Services))
	}

	// Check that services exist
	services := []string{"playwright", "postgres", "redis", "mitmproxy"}
	for _, svc := range services {
		if _, exists := compose.Services[svc]; !exists {
			t.Errorf("Service %s not found in compose file", svc)
		}
	}

	// Check network
	if _, exists := compose.Networks["rize"]; !exists {
		t.Error("Network 'rize' not found in compose file")
	}

	// Check volumes
	if len(compose.Volumes) != 3 {
		t.Errorf("Expected 3 volumes, got %d", len(compose.Volumes))
	}
}

func TestGenerateComposeFileWithDisabledService(t *testing.T) {
	cfg := config.DefaultConfig()

	// Disable postgres
	svc := cfg.Services["postgres"]
	svc.Enabled = false
	cfg.Services["postgres"] = svc

	compose, err := GenerateComposeFile(cfg)
	if err != nil {
		t.Fatalf("Failed to generate compose file: %v", err)
	}

	// Should only have 3 services now (4 - 1 disabled)
	if len(compose.Services) != 3 {
		t.Errorf("Expected 3 services, got %d", len(compose.Services))
	}

	// Postgres should not be in the compose file
	if _, exists := compose.Services["postgres"]; exists {
		t.Error("Postgres should not be in compose file when disabled")
	}
}
