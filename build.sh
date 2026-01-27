#!/bin/bash

# Build script for tproxy - supports multiple architectures
# Usage: ./build.sh [arch] [output_dir]

set -e

# Default values
DEFAULT_ARCH="amd64"
DEFAULT_OUTPUT="build"

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
echo "Building tproxy for $ARCH (GOARCH=$GOARCH, GOARM=${GOARM:-not set})"
echo "Output directory: $OUTPUT_DIR"

# Build the binary
CGO_ENABLED=0 GOOS=linux GOARCH=$GOARCH GOARM=$GOARM \
    go build -a -installsuffix cgo -ldflags='-extldflags "-static"' \
    -o "$OUTPUT_DIR/tproxy-$ARCH" ./cmd/tproxy

# Create a symlink for the current architecture
if [ "$ARCH" = "$(./detect-arch.sh)" ]; then
    ln -sf "tproxy-$ARCH" "$OUTPUT_DIR/tproxy"
fi

echo "Build completed: $OUTPUT_DIR/tproxy-$ARCH"

# Build Docker image if requested
if [ "$3" = "--docker" ]; then
    echo "Building Docker image for $ARCH..."
    docker build --build-arg TARGETARCH=$ARCH --build-arg TARGETVARIANT=$GOARM -t tproxy:$ARCH .
    echo "Docker image built: tproxy:$ARCH"
fi