#!/bin/bash

# Build DEB package for tproxy using Docker
# Usage: ./build-deb.sh [arch]

set -e

# Default architecture
DEFAULT_ARCH="amd64"
ARCH=${1:-$DEFAULT_ARCH}

# Supported architectures for DEB packages
SUPPORTED_ARCHS=("amd64" "arm64")

# Validate architecture
if [[ ! " ${SUPPORTED_ARCHS[@]} " =~ " ${ARCH} " ]]; then
    echo "Error: Unsupported architecture '$ARCH' for DEB packages"
    echo "Supported architectures: ${SUPPORTED_ARCHS[*]}"
    exit 1
fi

# Set DEB architecture names
case $ARCH in
    "amd64")
        DEB_ARCH="amd64"
        ;;
    "arm64")
        DEB_ARCH="arm64"
        ;;
esac

# Build directories
BUILD_DIR="build"
PKG_DIR="$BUILD_DIR/packages/deb"
DEB_BUILD_DIR="$BUILD_DIR/deb-build"

echo "Building DEB package for $ARCH..."

# Clean and create directories
rm -rf "$DEB_BUILD_DIR"
mkdir -p "$PKG_DIR" "$DEB_BUILD_DIR"

# Build the binary first
echo "Building binary for $ARCH..."
./build.sh "$ARCH" "$DEB_BUILD_DIR"

# Create DEB package structure
echo "Creating DEB package structure..."
mkdir -p "$DEB_BUILD_DIR/DEBIAN"
mkdir -p "$DEB_BUILD_DIR/usr/bin"
mkdir -p "$DEB_BUILD_DIR/etc/tproxy"
mkdir -p "$DEB_BUILD_DIR/usr/share/doc/tproxy"

# Copy binary
cp "$DEB_BUILD_DIR/tproxy-$ARCH" "$DEB_BUILD_DIR/usr/bin/tproxy"
chmod 755 "$DEB_BUILD_DIR/usr/bin/tproxy"

# Copy configuration
cp proxy_config.yaml "$DEB_BUILD_DIR/etc/tproxy/"
chmod 644 "$DEB_BUILD_DIR/etc/tproxy/proxy_config.yaml"

# Copy documentation
cp README.md "$DEB_BUILD_DIR/usr/share/doc/tproxy/"

# Update control file with correct architecture
cat > "$DEB_BUILD_DIR/DEBIAN/control" << EOF
Package: tproxy
Version: 1.0.0
Section: net
Priority: optional
Architecture: $DEB_ARCH
Maintainer: TProxy Team <tproxy@example.com>
Description: Transparent HTTP/HTTPS Proxy Server
 A Go implementation of a transparent HTTP/HTTPS proxy server with configurable routing rules.
 The proxy extracts SNI from TLS connections and routes traffic to upstream proxies using domain names.
Depends: libc6
EOF

# Create postinst script
cat > "$DEB_BUILD_DIR/DEBIAN/postinst" << 'EOF'
#!/bin/bash
chmod 755 /usr/bin/tproxy
echo "tproxy installed successfully"
echo "Configuration file: /etc/tproxy/proxy_config.yaml"
echo "Usage: tproxy --config /etc/tproxy/proxy_config.yaml"
EOF
chmod 755 "$DEB_BUILD_DIR/DEBIAN/postinst"

# Build DEB package using Docker
echo "Building DEB package using Docker..."
docker run --rm \
    -v "$(pwd)/$DEB_BUILD_DIR:/build" \
    -w /build \
    debian:bullseye \
    bash -c "dpkg-deb --build . /build/tproxy_1.0.0_${DEB_ARCH}.deb"

# Move the package to packages directory
mv "$DEB_BUILD_DIR/tproxy_1.0.0_${DEB_ARCH}.deb" "$PKG_DIR/"

echo "DEB package built: $PKG_DIR/tproxy_1.0.0_${DEB_ARCH}.deb"

# Cleanup
rm -rf "$DEB_BUILD_DIR"

echo "DEB package build completed successfully!"