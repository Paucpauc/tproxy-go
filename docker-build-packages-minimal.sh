#!/bin/bash

# Docker-based package building script for tproxy using minimal images
# Builds DEB packages using minimal Docker containers
# Uses alpine for DEB packages (dpkg only) instead of debian:bullseye-slim

set -e

# Default values
DEFAULT_OUTPUT="build/packages"
OUTPUT_DIR=${1:-$DEFAULT_OUTPUT}

# Supported architectures for packages
PACKAGE_ARCHS=("amd64" "arm64")

echo "Building packages for architecture: ${PACKAGE_ARCHS[*]}"
echo "Output directory: $OUTPUT_DIR"

# Create output directories
mkdir -p "$OUTPUT_DIR/deb"

# Build packages for each architecture
for ARCH in "${PACKAGE_ARCHS[@]}"; do
    echo "=== Building packages for $ARCH ==="
    
    # Set GOARCH based on architecture
    case $ARCH in
        "amd64")
            GOARCH="amd64"
            ;;
        "arm64")
            GOARCH="arm64"
            ;;
    esac
    
    # First, build the binary using Docker
    echo "Building binary for $ARCH..."
    docker run --rm \
        -v "$(pwd)":/app \
        -w /app \
        -e CGO_ENABLED=0 \
        -e GOOS=linux \
        -e GOARCH=$GOARCH \
        golang:1.21-alpine \
        sh -c "
            apk add --no-cache git ca-certificates tzdata && \
            go build -buildvcs=false -a -installsuffix cgo -ldflags='-buildid= -extldflags \"-static\"' \
            -o /app/build/tproxy-$ARCH ./cmd/tproxy
        "
    
    echo "✓ Binary built: build/tproxy-$ARCH"
    
    # Build DEB package using minimal Alpine image
    echo "Building DEB package for $ARCH using minimal Alpine image..."
    docker run --rm \
        -v "$(pwd)":/app \
        -w /app \
        -e ARCH=$ARCH \
        alpine:latest \
        sh -c "
            apk add --no-cache dpkg sed && \
            mkdir -p /tmp/deb-build && \
            mkdir -p /tmp/deb-build/DEBIAN && \
            cp packaging/deb/control /tmp/deb-build/DEBIAN/ && \
            cp packaging/deb/postinst /tmp/deb-build/DEBIAN/ && \
            cp packaging/deb/prerm /tmp/deb-build/DEBIAN/ && \
            # Update Architecture field in control file
            sed -i \"s/Architecture: amd64/Architecture: $ARCH/g\" /tmp/deb-build/DEBIAN/control && \
            mkdir -p /tmp/deb-build/usr/bin && \
            cp build/tproxy-$ARCH /tmp/deb-build/usr/bin/tproxy && \
            mkdir -p /tmp/deb-build/etc/tproxy && \
            cp proxy_config.yaml /tmp/deb-build/etc/tproxy/ && \
            mkdir -p /tmp/deb-build/usr/share/doc/tproxy && \
            cp README.md /tmp/deb-build/usr/share/doc/tproxy/ && \
            mkdir -p /tmp/deb-build/lib/systemd/system && \
            cp packaging/tproxy.service /tmp/deb-build/lib/systemd/system/ && \
            cd /tmp/deb-build && \
            dpkg-deb --build . /app/$OUTPUT_DIR/deb/tproxy_${ARCH}.deb
        "
    
    echo "✓ DEB package built: $OUTPUT_DIR/deb/tproxy_${ARCH}.deb"
done

echo ""
echo "=== Package Build Summary ==="
echo "All packages built successfully using minimal Docker images:"
echo "DEB packages (built with alpine:latest): $OUTPUT_DIR/deb/"
ls -la "$OUTPUT_DIR/deb/"

echo ""
echo "Minimal Docker-based package build completed successfully!"
echo "DEB packages now use alpine:latest + dpkg instead of debian:bullseye-slim + dpkg-dev"