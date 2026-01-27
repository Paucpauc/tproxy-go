#!/bin/bash

# Build packages for all architectures using minimal approach
# Usage: ./build-packages.sh

set -e

echo "Building packages for all architectures using minimal approach..."

# Build DEB packages using minimal approach
echo "=== Building DEB packages (minimal) ==="
./build-deb-minimal.sh amd64
./build-deb-minimal.sh arm64

# Build RPM packages
echo "=== Building RPM packages ==="
./build-rpm.sh amd64
./build-rpm.sh arm64

echo "All packages built successfully!"
echo "Packages location: build/packages/"
echo "  DEB packages: build/packages/deb/ (built with minimal alpine image)"
echo "  RPM packages: build/packages/rpm/"