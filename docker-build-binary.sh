#!/bin/bash

# Docker-based binary build script for tproxy
# Builds Go binaries using Docker containers - no local Go installation required
# Supports multiple architectures: amd64, arm64, and arm

set -e

# Default values
DEFAULT_ARCH="amd64"
DEFAULT_OUTPUT="build"
IMAGE_NAME="golang:1.21-alpine"

# Parse arguments
ARCH=${1:-$DEFAULT_ARCH}
OUTPUT_DIR=${2:-$DEFAULT_OUTPUT}

# Supported architectures
SUPPORTED_ARCHS=("amd64" "arm64" "arm")

# Validate architecture
if [[ ! " ${SUPPORTED_ARCHS[@]} " =~ " ${ARCH} " ]]; then
    echo "Error: Unsupported architecture '$ARCH'"
    echo "Supported architectures: ${SUPPORTED_ARCHS[*]}"
    exit 1
fi

# Set GOARCH and GOARM based on architecture
case $ARCH in
    "amd64")
        GOARCH="amd64"
        GOARM=""
        ;;
    "arm64")
        GOARCH="arm64"
        GOARM=""
        ;;
    "arm")
        GOARCH="arm"
        GOARM="7"  # For Mikrotik HAP AC2 (ARMv7)
        ;;
esac

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build info
echo "Building tproxy for $ARCH using Docker (GOARCH=$GOARCH, GOARM=${GOARM:-not set})"
echo "Output directory: $OUTPUT_DIR"
echo "Using Docker image: $IMAGE_NAME"

# Build the binary using Docker
docker run --rm \
    -v "$(pwd)":/app \
    -w /app \
    -e CGO_ENABLED=0 \
    -e GOOS=linux \
    -e GOARCH=$GOARCH \
    -e GOARM=$GOARM \
    $IMAGE_NAME \
    sh -c "
        apk add --no-cache git ca-certificates tzdata && \
        go build -a -installsuffix cgo -ldflags='-w -s -extldflags \"-static\"' \
        -o /app/$OUTPUT_DIR/tproxy-$ARCH ./cmd/tproxy
    "

echo "✓ Built: $OUTPUT_DIR/tproxy-$ARCH"

# Create a symlink for the current architecture if we can detect it
if command -v uname >/dev/null 2>&1; then
    CURRENT_ARCH=$(uname -m)
    case $CURRENT_ARCH in
        "x86_64") DETECTED_ARCH="amd64" ;;
        "aarch64") DETECTED_ARCH="arm64" ;;
        "armv7l"|"armv7") DETECTED_ARCH="arm" ;;
        *) DETECTED_ARCH="" ;;
    esac
    
    if [ "$ARCH" = "$DETECTED_ARCH" ]; then
        cd "$OUTPUT_DIR"
        ln -sf "tproxy-$ARCH" tproxy
        cd -
        echo "✓ Created symlink: $OUTPUT_DIR/tproxy -> tproxy-$ARCH"
    fi
fi

echo "Build completed successfully: $OUTPUT_DIR/tproxy-$ARCH"