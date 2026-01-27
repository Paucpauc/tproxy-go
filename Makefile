# Makefile for tproxy - Multi-architecture builds with Docker support

# Variables
BINARY_NAME = tproxy
BUILD_DIR = build
IMAGE_NAME = tproxy
TAG = latest

# Supported architectures
ARCHS = amd64 arm64 arm

.PHONY: all build build-all docker-build docker-build-all docker docker-all packages packages-all docker-packages-minimal clean help test test-all docker-test docker-test-all

# Default target - use Docker-based build
all: docker-build

# Build for current architecture only (using local Go)
build:
	@echo "Building for current architecture using local Go..."
	@./build.sh

# Build for all architectures (using local Go)
build-all:
	@echo "Building for all architectures using local Go..."
	@./build-all.sh

# Build binary for current architecture using Docker (no local Go required)
docker-build:
	@echo "Building for current architecture using Docker..."
	@./docker-build-binary.sh

# Build binaries for all architectures using Docker (no local Go required)
docker-build-all:
	@echo "Building for all architectures using Docker..."
	@./docker-build-all-binary.sh

# Build Docker image for current architecture
docker:
	@echo "Building Docker image for current architecture..."
	@./build.sh $(shell ./detect-arch.sh) $(BUILD_DIR) --docker

# Build Docker images for all architectures
docker-all:
	@echo "Building Docker images for all architectures..."
	@./docker-build.sh $(TAG) false

# Build and push Docker images for all architectures
docker-push:
	@echo "Building and pushing Docker images for all architectures..."
	@./docker-build.sh $(TAG) true

# Build packages for all architectures using minimal approach
packages:
	@echo "Building packages for all architectures using minimal approach..."
	@./build-packages.sh

# Build packages for all architectures using minimal approach (alias)
packages-all:
	@echo "Building packages for all architectures using minimal approach..."
	@./build-packages.sh

# Build packages for all architectures using minimal Docker approach
docker-packages-minimal:
	@echo "Building packages for all architectures using minimal Docker approach..."
	@./docker-build-packages-minimal.sh

# Test targets

# Run tests for current architecture (using local Go)
test:
	@echo "Running tests using local Go..."
	@go test ./...

# Run tests with verbose output and coverage (using local Go)
test-all:
	@echo "Running all tests with verbose output and coverage..."
	@go test -v -cover ./...

# Run tests using Docker (no local Go required)
docker-test:
	@echo "Running tests using Docker..."
	@docker run --rm -v $(PWD):/app -w /app golang:1.21-alpine go test ./...

# Run tests with verbose output and coverage using Docker
docker-test-all:
	@echo "Running all tests with verbose output and coverage using Docker..."
	@docker run --rm -v $(PWD):/app -w /app golang:1.21-alpine go test -v -cover ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "✓ Clean completed"

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build for current architecture using local Go"
	@echo "  build-all       - Build for all architectures using local Go"
	@echo "  docker-build    - Build for current architecture using Docker (no local Go)"
	@echo "  docker-build-all - Build for all architectures using Docker (no local Go)"
	@echo "  docker          - Build Docker image for current architecture"
	@echo "  docker-all      - Build Docker images for all architectures"
	@echo "  docker-push     - Build and push Docker images for all architectures"
	@echo "  packages           - Build DEB packages for all architectures using minimal approach"
	@echo "  packages-all       - Build DEB packages for all architectures (alias)"
	@echo "  docker-packages-minimal - Build DEB packages using minimal Docker approach"
	@echo "  clean           - Clean build artifacts"
	@echo "  test            - Run tests using local Go"
	@echo "  test-all        - Run all tests with verbose output and coverage"
	@echo "  docker-test     - Run tests using Docker (no local Go)"
	@echo "  docker-test-all - Run all tests with verbose output and coverage using Docker"
	@echo "  help            - Show this help message"
	@echo ""
	@echo "Architecture support:"
	@echo "  - amd64: x64 systems"
	@echo "  - arm64: ARM64 systems"
	@echo "  - arm:   ARMv7 (for Mikrotik HAP AC2 routers)"
	@echo ""
	@echo "Package support:"
	@echo "  - DEB packages: amd64, arm64"
	@echo ""
	@echo "Usage examples:"
	@echo "  make docker-build-all    # Build binaries for all architectures using Docker"
	@echo "  make docker-all          # Build Docker images for all architectures"
	@echo "  make docker-packages-minimal  # Build packages using minimal Docker approach"
	@echo "  make packages-all             # Build DEB packages for all architectures"
	@echo "  make test                # Run tests using local Go"
	@echo "  make docker-test         # Run tests using Docker"
	@echo "  ./docker-build-binary.sh arm  # Build specifically for ARM using Docker"