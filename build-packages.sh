#!/bin/bash

# Build packages for all architectures using minimal approach
# Usage: ./build-packages.sh

set -e

echo "Building packages for all architectures using minimal approach..."

# Build DEB packages using minimal approach
echo "=== Building DEB packages (minimal) ==="
./build-deb-minimal.sh amd64
./build-deb-minimal.sh arm64

# Build RPM packages (amd64 only - cross-compilation for arm64 requires special setup)
echo "=== Building RPM packages ==="
./build-rpm.sh amd64
echo "Note: RPM package for arm64 requires native arm64 build environment"

echo "All packages built successfully!"
echo "Packages location: build/packages/"
echo "  DEB packages: build/packages/deb/ (built with minimal alpine image)"
echo "  RPM packages: build/packages/rpm/ (amd64 only)"