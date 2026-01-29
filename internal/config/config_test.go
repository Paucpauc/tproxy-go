package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_DefaultConfig(t *testing.T) {
	// Test loading default config when file doesn't exist
	config, err := LoadConfig("non-existent-file.yaml")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify default values
	if config.Listen.Host != "127.0.0.1" {
		t.Errorf("Expected host 127.0.0.1, got %s", config.Listen.Host)
	}
	if config.Listen.HTTPSPort != 3130 {
		t.Errorf("Expected HTTPS port 3130, got %d", config.Listen.HTTPSPort)
	}
	if config.Listen.HTTPPort != 3131 {
		t.Errorf("Expected HTTP port 3131, got %d", config.Listen.HTTPPort)
	}
	if len(config.Rules) != 1 {
		t.Errorf("Expected 1 default rule, got %d", len(config.Rules))
	}
	if config.Rules[0].Pattern != ".*" {
		t.Errorf("Expected pattern .*, got %s", config.Rules[0].Pattern)
	}
	if config.Rules[0].Proxy != "DIRECT" {
		t.Errorf("Expected proxy DIRECT, got %s", config.Rules[0].Proxy)
	}
}

func TestLoadConfig_ValidConfigFile(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
listen:
  host: "0.0.0.0"
  https_port: 8443
  http_port: 8080
rules:
  - pattern: ".*\\.example\\.com"
    proxy: "proxy.example.com:3128"
  - pattern: ".*\\.google\\.com"
    proxy: "DROP"
  - pattern: ".*"
    proxy: "DIRECT"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded values
	if config.Listen.Host != "0.0.0.0" {
		t.Errorf("Expected host 0.0.0.0, got %s", config.Listen.Host)
	}
	if config.Listen.HTTPSPort != 8443 {
		t.Errorf("Expected HTTPS port 8443, got %d", config.Listen.HTTPSPort)
	}
	if config.Listen.HTTPPort != 8080 {
		t.Errorf("Expected HTTP port 8080, got %d", config.Listen.HTTPPort)
	}
	if len(config.Rules) != 3 {
		t.Errorf("Expected 3 rules, got %d", len(config.Rules))
	}
}

func TestLoadConfig_InvalidConfigFile(t *testing.T) {
	// Create invalid config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
invalid: yaml: content
  - this is not valid yaml
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected LoadConfig to fail with invalid YAML")
	}
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	// Create config with partial values
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
listen:
  host: "192.168.1.1"
rules:
  - pattern: "specific\\.com"
    proxy: "proxy:8080"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded values with defaults for missing fields
	if config.Listen.Host != "192.168.1.1" {
		t.Errorf("Expected host 192.168.1.1, got %s", config.Listen.Host)
	}
	if config.Listen.HTTPSPort != 3130 {
		t.Errorf("Expected default HTTPS port 3130, got %d", config.Listen.HTTPSPort)
	}
	if config.Listen.HTTPPort != 3131 {
		t.Errorf("Expected default HTTP port 3131, got %d", config.Listen.HTTPPort)
	}
	if config.Listen.Timeout != DEFAULT_TIMEOUT {
		t.Errorf("Expected default timeout %d, got %d", DEFAULT_TIMEOUT, config.Listen.Timeout)
	}
	if len(config.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(config.Rules))
	}
}

func TestFindProxyForHost_DirectMatch(t *testing.T) {
	rules := []Rule{
		{Pattern: ".*\\.example\\.com", Proxy: "proxy.example.com:3128"},
		{Pattern: ".*\\.google\\.com", Proxy: "DROP"},
		{Pattern: ".*", Proxy: "DIRECT"},
	}

	tests := []struct {
		host     string
		expected string
	}{
		{"api.example.com", "PROXY"},
		{"www.google.com", "DROP"},
		{"example.org", "DIRECT"},
		{"sub.domain.com", "DIRECT"},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			action, err := FindProxyForHost(tt.host, rules)
			if err != nil {
				t.Fatalf("FindProxyForHost failed: %v", err)
			}

			if action.Type != tt.expected {
				t.Errorf("Expected action type %s, got %s", tt.expected, action.Type)
			}
		})
	}
}

func TestFindProxyForHost_ProxyDetails(t *testing.T) {
	rules := []Rule{
		{Pattern: "api\\.example\\.com", Proxy: "proxy.internal:8080"},
		{Pattern: ".*\\.test\\.com", Proxy: "proxy.test.com"},
	}

	tests := []struct {
		host         string
		expectedType string
		expectedHost string
		expectedPort int
	}{
		{"api.example.com", "PROXY", "proxy.internal", 8080},
		{"app.test.com", "PROXY", "proxy.test.com", 3128}, // Default port
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			action, err := FindProxyForHost(tt.host, rules)
			if err != nil {
				t.Fatalf("FindProxyForHost failed: %v", err)
			}

			if action.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, action.Type)
			}
			if action.Host != tt.expectedHost {
				t.Errorf("Expected host %s, got %s", tt.expectedHost, action.Host)
			}
			if action.Port != tt.expectedPort {
				t.Errorf("Expected port %d, got %d", tt.expectedPort, action.Port)
			}
		})
	}
}

func TestFindProxyForHost_InvalidRegex(t *testing.T) {
	rules := []Rule{
		{Pattern: "[invalid-regex", Proxy: "DIRECT"}, // Invalid regex pattern
		{Pattern: ".*", Proxy: "proxy:8080"},
	}

	// Should not fail, just skip invalid patterns
	action, err := FindProxyForHost("example.com", rules)
	if err != nil {
		t.Fatalf("FindProxyForHost failed: %v", err)
	}

	if action.Type != "PROXY" {
		t.Errorf("Expected PROXY action from fallback rule, got %s", action.Type)
	}
}

func TestParseProxyAddress_Basic(t *testing.T) {
	tests := []struct {
		input        string
		expectedHost string
		expectedPort int
	}{
		{"proxy.example.com:3128", "proxy.example.com", 3128},
		{"proxy.internal", "proxy.internal", 3128}, // Default port
		{"192.168.1.1:8080", "192.168.1.1", 8080},
		{"[::1]:8443", "[::1]", 8443},
		{"host:invalid", "host", 3128}, // Invalid port falls back to default
		{"", "", 3128},                 // Empty string
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			host, port := parseProxyAddress(tt.input)
			if host != tt.expectedHost {
				t.Errorf("Expected host %s, got %s", tt.expectedHost, host)
			}
			if port != tt.expectedPort {
				t.Errorf("Expected port %d, got %d", tt.expectedPort, port)
			}
		})
	}
}

func TestParseProxyAddress_IPv6(t *testing.T) {
	// Test IPv6 address parsing
	host, port := parseProxyAddress("[2001:db8::1]:3128")
	if host != "[2001:db8::1]" {
		t.Errorf("Expected host [2001:db8::1], got %s", host)
	}
	if port != 3128 {
		t.Errorf("Expected port 3128, got %d", port)
	}
}

func TestLoadConfig_TimeoutDefaults(t *testing.T) {
	// Test that default timeout is set when not specified
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
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

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify default timeout is set
	if config.Listen.Timeout != DEFAULT_TIMEOUT {
		t.Errorf("Expected default timeout %d, got %d", DEFAULT_TIMEOUT, config.Listen.Timeout)
	}
}
