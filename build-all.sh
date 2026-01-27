#!/bin/bash

# Multi-architecture build script for tproxy
# Builds for amd64, arm64, and arm (ARMv7 for Mikrotik HAP AC2)

set -e

# Default values
DEFAULT_OUTPUT="build"
OUTPUT_DIR=${1:-$DEFAULT_OUTPUT}

# Supported architectures
ARCHS=("amd64" "arm64" "arm")

echo "Building tproxy for multiple architectures: ${ARCHS[*]}"
echo "Output directory: $OUTPUT_DIR"

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
    
    # Build the binary
    CGO_ENABLED=0 GOOS=linux GOARCH=$GOARCH GOARM=$GOARM \
        go build -a -installsuffix cgo -ldflags='-buildid= -extldflags "-static"' \
        -o "$OUTPUT_DIR/tproxy-$ARCH" ./cmd/tproxy
    
    echo "✓ Built: $OUTPUT_DIR/tproxy-$ARCH"
    
    # Create a symlink for the current architecture
    if [ "$ARCH" = "$(./detect-arch.sh)" ]; then
        ln -sf "tproxy-$ARCH" "$OUTPUT_DIR/tproxy"
        echo "✓ Created symlink: $OUTPUT_DIR/tproxy -> tproxy-$ARCH"
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
echo "Build completed successfully!"