#!/bin/bash

# Docker-based multi-architecture binary build script for tproxy
# Builds Go binaries for all supported architectures using Docker containers
# No local Go installation required

set -e

# Default values
DEFAULT_OUTPUT="build"
OUTPUT_DIR=${1:-$DEFAULT_OUTPUT}
IMAGE_NAME="golang:1.21-alpine"

# Supported architectures
ARCHS=("amd64" "arm64" "arm")

echo "Building tproxy for multiple architectures using Docker: ${ARCHS[*]}"
echo "Output directory: $OUTPUT_DIR"
echo "Using Docker image: $IMAGE_NAME"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build for each architecture
for ARCH in "${ARCHS[@]}"; do
    echo "=== Building for $ARCH ==="
    
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
            go build -buildvcs=false -a -installsuffix cgo -ldflags='-w -s -extldflags \"-static\"' \
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
done

echo ""
echo "=== Build Summary ==="
echo "All binaries built successfully in $OUTPUT_DIR:"
ls -la "$OUTPUT_DIR"/tproxy-*

# Create checksums
echo ""
echo "Creating checksums..."
cd "$OUTPUT_DIR"
for file in tproxy-*; do
    if [ -f "$file" ] && [[ "$file" != *.sha256 ]]; then
        sha256sum "$file" > "$file.sha256"
        echo "✓ Created checksum: $file.sha256"
    fi
done
cd -

echo ""
echo "Docker-based build completed successfully!"