# Makefile for tproxy - Multi-architecture builds

# Variables
BINARY_NAME = tproxy
BUILD_DIR = build
IMAGE_NAME = tproxy
TAG = latest

# Supported architectures
ARCHS = amd64 arm64 arm

.PHONY: all build build-all docker docker-all packages packages-all clean help

# Default target
all: build

# Build for current architecture only
build:
	@echo "Building for current architecture..."
	@./build.sh

# Build for all architectures
build-all:
	@echo "Building for all architectures..."
	@./build-all.sh

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

# Build packages for current architecture
packages:
	@echo "Building packages for current architecture..."
	@./build-deb.sh $(shell ./detect-arch.sh)
	@./build-rpm.sh $(shell ./detect-arch.sh)

# Build packages for all architectures
packages-all:
	@echo "Building packages for all architectures..."
	@./build-packages.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "✓ Clean completed"

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build for current architecture"
	@echo "  build-all   - Build for all architectures (amd64, arm64, arm)"
	@echo "  docker      - Build Docker image for current architecture"
	@echo "  docker-all  - Build Docker images for all architectures"
	@echo "  docker-push - Build and push Docker images for all architectures"
	@echo "  packages    - Build DEB/RPM packages for current architecture"
	@echo "  packages-all - Build DEB/RPM packages for all architectures"
	@echo "  clean       - Clean build artifacts"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Architecture support:"
	@echo "  - amd64: x64 systems"
	@echo "  - arm64: ARM64 systems"
	@echo "  - arm:   ARMv7 (for Mikrotik HAP AC2 routers)"
	@echo ""
	@echo "Package support:"
	@echo "  - DEB packages: amd64, arm64"
	@echo "  - RPM packages: amd64, arm64"
	@echo ""
	@echo "Usage examples:"
	@echo "  make build-all          # Build binaries for all architectures"
	@echo "  make docker-all         # Build Docker images for all architectures"
	@echo "  make packages-all       # Build DEB/RPM packages for all architectures"
	@echo "  ./build.sh arm          # Build specifically for ARM"