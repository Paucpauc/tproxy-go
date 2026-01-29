package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_HTTPS_PORT = 443
	DEFAULT_HTTP_PORT  = 80
	BUFFER_SIZE        = 4096
	DEFAULT_TIMEOUT     = 900 // seconds
)

type ListenConfig struct {
	Host      string `yaml:"host"`
	HTTPSPort int    `yaml:"https_port"`
	HTTPPort  int    `yaml:"http_port"`
	Timeout   int    `yaml:"timeout"` // Timeout in seconds
}

type Rule struct {
	Pattern string `yaml:"pattern"`
	Proxy   string `yaml:"proxy"`
}

type Config struct {
	Listen ListenConfig `yaml:"listen"`
	Rules  []Rule       `yaml:"rules"`
}

var DefaultConfig = Config{
	Listen: ListenConfig{
		Host:      "127.0.0.1",
		HTTPSPort: 3130,
		HTTPPort:  3131,
		Timeout:   DEFAULT_TIMEOUT,
	},
	Rules: []Rule{
		{Pattern: ".*", Proxy: "DIRECT"},
	},
}

func LoadConfig(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Config file %s not found, using default config\n", configPath)
		return &DefaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge with default config to ensure all required fields exist
	if config.Listen.Host == "" {
		config.Listen.Host = DefaultConfig.Listen.Host
	}
	if config.Listen.HTTPSPort == 0 {
		config.Listen.HTTPSPort = DefaultConfig.Listen.HTTPSPort
	}
	if config.Listen.HTTPPort == 0 {
		config.Listen.HTTPPort = DefaultConfig.Listen.HTTPPort
	}
	if config.Listen.Timeout == 0 {
		config.Listen.Timeout = DefaultConfig.Listen.Timeout
	}
	if len(config.Rules) == 0 {
		config.Rules = DefaultConfig.Rules
	}

	return &config, nil
}

type ProxyAction struct {
	Type string // "DIRECT", "PROXY", "DROP"
	Host string
	Port int
}

func FindProxyForHost(host string, rules []Rule) (*ProxyAction, error) {
	for _, rule := range rules {
		matched, err := regexp.MatchString(rule.Pattern, host)
		if err != nil {
			fmt.Printf("Invalid regex pattern: %s\n", rule.Pattern)
			continue
		}

		if matched {
			switch rule.Proxy {
			case "DIRECT":
				return &ProxyAction{Type: "DIRECT"}, nil
			case "DROP":
				return &ProxyAction{Type: "DROP"}, nil
			default:
				// Parse proxy host:port
				host, port := parseProxyAddress(rule.Proxy)
				return &ProxyAction{
					Type: "PROXY",
					Host: host,
					Port: port,
				}, nil
			}
		}
	}

	// Fallback to DIRECT if no rules match
	return &ProxyAction{Type: "DIRECT"}, nil
}

func parseProxyAddress(proxy string) (string, int) {
	// Simple parsing for host:port format
	// Default to port 3128 if not specified
	host := proxy
	port := 3128

	// Look for the last colon (to handle IPv6 addresses correctly)
	lastColon := -1
	for i := len(proxy) - 1; i >= 0; i-- {
		if proxy[i] == ':' {
			lastColon = i
			break
		}
	}

	if lastColon != -1 {
		host = proxy[:lastColon]
		portStr := proxy[lastColon+1:]
		if portStr != "" {
			// Try to parse port
			if p, err := strconv.Atoi(portStr); err == nil {
				port = p
			}
		}
	}

	return host, port
}
