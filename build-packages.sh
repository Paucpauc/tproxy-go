#!/bin/bash

# Build packages for all architectures
# Usage: ./build-packages.sh

set -e

echo "Building packages for all architectures..."

# Build DEB packages
echo "=== Building DEB packages ==="
./build-deb.sh amd64
./build-deb.sh arm64

# Build RPM packages
echo "=== Building RPM packages ==="
./build-rpm.sh amd64
./build-rpm.sh arm64

echo "All packages built successfully!"
echo "Packages location: build/packages/"
echo "  DEB packages: build/packages/deb/"
echo "  RPM packages: build/packages/rpm/"