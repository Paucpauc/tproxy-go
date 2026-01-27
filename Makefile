# Makefile for tproxy - Multi-architecture builds with Docker support

# Variables
BINARY_NAME = tproxy
BUILD_DIR = build
IMAGE_NAME = tproxy
TAG = latest

# Supported architectures
ARCHS = amd64 arm64 arm

.PHONY: all build build-all docker-build docker-build-all docker docker-all packages packages-all docker-packages clean help

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

# Build packages for current architecture
packages:
	@echo "Building packages for current architecture..."
	@./build-deb.sh $(shell ./detect-arch.sh)
	@./build-rpm.sh $(shell ./detect-arch.sh)

# Build packages for all architectures
packages-all:
	@echo "Building packages for all architectures..."
	@./build-packages.sh

# Build packages for all architectures using Docker (no local tools required)
docker-packages:
	@echo "Building packages for all architectures using Docker..."
	@./docker-build-packages.sh

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
	@echo "  packages        - Build DEB/RPM packages for current architecture"
	@echo "  packages-all    - Build DEB/RPM packages for all architectures"
	@echo "  docker-packages - Build DEB/RPM packages using Docker (no local tools)"
	@echo "  clean           - Clean build artifacts"
	@echo "  help            - Show this help message"
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
	@echo "  make docker-build-all    # Build binaries for all architectures using Docker"
	@echo "  make docker-all          # Build Docker images for all architectures"
	@echo "  make docker-packages     # Build packages using Docker (no local tools)"
	@echo "  make packages-all        # Build DEB/RPM packages for all architectures"
	@echo "  ./docker-build-binary.sh arm  # Build specifically for ARM using Docker"