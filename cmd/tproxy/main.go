package main

import (
	"flag"
	"fmt"
	"log"

	"tproxy/internal/config"
	"tproxy/internal/server"
)

func main() {
	configPath := flag.String("config", "proxy_config.yaml", "Path to YAML config file")
	flag.Parse()

	// Load configuration
	config, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Starting proxy server with config from %s\n", *configPath)

	// Start servers
	if err := server.StartServers(config); err != nil {
		log.Fatalf("Failed to start servers: %v", err)
	}
}
