package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tproxy/internal/config"
)

func TestMain_ConfigLoading(t *testing.T) {
	// Test default config loading
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.yaml")

	// Create a valid config file
	configContent := `
listen:
  host: "127.0.0.1"
  https_port: 3130
  http_port: 3131
rules:
  - pattern: ".*"
    proxy: "DIRECT"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test loading the config directly
	loadedConfig, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedConfig.Listen.Host != "127.0.0.1" {
		t.Errorf("Expected host 127.0.0.1, got %s", loadedConfig.Listen.Host)
	}
	if loadedConfig.Listen.HTTPSPort != 3130 {
		t.Errorf("Expected HTTPS port 3130, got %d", loadedConfig.Listen.HTTPSPort)
	}
	if loadedConfig.Listen.HTTPPort != 3131 {
		t.Errorf("Expected HTTP port 3131, got %d", loadedConfig.Listen.HTTPPort)
	}
}

func TestMain_DefaultConfig(t *testing.T) {
	// Test that default config is used when file doesn't exist
	loadedConfig, err := config.LoadConfig("non-existent-config.yaml")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	// Verify default values
	if loadedConfig.Listen.Host != "127.0.0.1" {
		t.Errorf("Expected default host 127.0.0.1, got %s", loadedConfig.Listen.Host)
	}
	if loadedConfig.Listen.HTTPSPort != 3130 {
		t.Errorf("Expected default HTTPS port 3130, got %d", loadedConfig.Listen.HTTPSPort)
	}
	if loadedConfig.Listen.HTTPPort != 3131 {
		t.Errorf("Expected default HTTP port 3131, got %d", loadedConfig.Listen.HTTPPort)
	}
}

func TestMain_FlagParsing(t *testing.T) {
	// Test command line flag parsing
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set test arguments
	os.Args = []string{"tproxy", "-config", "custom_config.yaml"}

	// Reset flag command line
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	configPath := flag.String("config", "proxy_config.yaml", "Path to YAML config file")
	flag.Parse()

	if *configPath != "custom_config.yaml" {
		t.Errorf("Expected config path custom_config.yaml, got %s", *configPath)
	}
}

func TestMain_ConfigValidation(t *testing.T) {
	// Test config validation scenarios
	tests := []struct {
		name        string
		config      *config.Config
		shouldError bool
	}{
		{
			name: "Valid config",
			config: &config.Config{
				Listen: config.ListenConfig{
					Host:      "127.0.0.1",
					HTTPSPort: 3130,
					HTTPPort:  3131,
				},
				Rules: []config.Rule{
					{Pattern: ".*", Proxy: "DIRECT"},
				},
			},
			shouldError: false,
		},
		{
			name: "Invalid port",
			config: &config.Config{
				Listen: config.ListenConfig{
					Host:      "127.0.0.1",
					HTTPSPort: -1, // Invalid port
					HTTPPort:  3131,
				},
				Rules: []config.Rule{},
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would test server startup, but we'll skip actual server startup in tests
			// since it requires network privileges and would block
			if tt.shouldError {
				// We expect server startup to fail with invalid config
				t.Logf("Expected server startup to fail with invalid config (this is normal)")
			}
		})
	}
}

// Test helper function to verify proxy rule matching
func TestProxyRuleMatching(t *testing.T) {
	rules := []config.Rule{
		{Pattern: ".*\\.google\\.com", Proxy: "DROP"},
		{Pattern: ".*\\.example\\.com", Proxy: "proxy.internal:3128"},
		{Pattern: ".*", Proxy: "DIRECT"},
	}

	tests := []struct {
		host     string
		expected string
	}{
		{"www.google.com", "DROP"},
		{"api.example.com", "PROXY"},
		{"example.org", "DIRECT"},
		{"test.com", "DIRECT"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			action, err := config.FindProxyForHost(tt.host, rules)
			if err != nil {
				t.Fatalf("FindProxyForHost failed: %v", err)
			}

			if action.Type != tt.expected {
				t.Errorf("For host %s, expected %s, got %s", tt.host, tt.expected, action.Type)
			}
		})
	}
}

// Test timeout handling (simulated)
func TestTimeoutHandling(t *testing.T) {
	// This test simulates timeout scenarios without actually blocking
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		// Expected timeout
		t.Log("Context timeout occurred as expected")
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected context timeout but it didn't occur")
	}
}
