#!/bin/bash

# Docker-based package building script for tproxy using minimal images
# Builds DEB and RPM packages using minimal Docker containers
# Uses alpine for DEB packages (dpkg only) instead of debian:bullseye-slim

set -e

# Default values
DEFAULT_OUTPUT="build/packages"
OUTPUT_DIR=${1:-$DEFAULT_OUTPUT}

# Supported architectures for packages
PACKAGE_ARCHS=("amd64")

echo "Building packages for architecture: ${PACKAGE_ARCHS[*]}"
echo "Output directory: $OUTPUT_DIR"

# Create output directories
mkdir -p "$OUTPUT_DIR/deb"
mkdir -p "$OUTPUT_DIR/rpm"

# Build packages for each architecture
for ARCH in "${PACKAGE_ARCHS[@]}"; do
    echo "=== Building packages for $ARCH ==="
    
    # Set GOARCH and RPMARCH based on architecture
    case $ARCH in
        "amd64")
            GOARCH="amd64"
            RPMARCH="x86_64"
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
    
    # Build RPM package using minimal Fedora image
    echo "Building RPM package for $ARCH using minimal Fedora image..."
    docker run --rm \
        -v "$(pwd)":/app \
        -w /app \
        -e ARCH=$ARCH \
        -e RPMARCH=$RPMARCH \
        fedora:latest \
        sh -c "
            dnf install -y --setopt=install_weak_deps=False rpm-build sed && \
            mkdir -p /tmp/rpm-build/{BUILD,RPMS,SOURCES,SPECS,SRPMS} && \
            cp packaging/rpm/tproxy.spec /tmp/rpm-build/SPECS/ && \
            # Update BuildArch in spec file for current architecture
            sed -i \"s/BuildArch: x86_64/BuildArch: $RPMARCH/g\" /tmp/rpm-build/SPECS/tproxy.spec && \
            # Create source tarball
            mkdir -p /tmp/tproxy-1.0.0 && \
            cp build/tproxy-$ARCH /tmp/tproxy-1.0.0/tproxy && \
            cp proxy_config.yaml /tmp/tproxy-1.0.0/ && \
            cp README.md /tmp/tproxy-1.0.0/ && \
            cp packaging/tproxy.service /tmp/tproxy-1.0.0/ && \
            cd /tmp && \
            tar -czf /tmp/rpm-build/SOURCES/tproxy-1.0.0.tar.gz tproxy-1.0.0 && \
            cd /tmp/rpm-build && \
            rpmbuild -bb --define \"_topdir /tmp/rpm-build\" --define \"debug_package %{nil}\" --define \"_build_id_links none\" SPECS/tproxy.spec && \
            cp RPMS/$RPMARCH/*.rpm /app/$OUTPUT_DIR/rpm/
        "
    
    echo "✓ RPM package built: $OUTPUT_DIR/rpm/tproxy-*$RPMARCH.rpm"
done

echo ""
echo "=== Package Build Summary ==="
echo "All packages built successfully using minimal Docker images:"
echo "DEB packages (built with alpine:latest): $OUTPUT_DIR/deb/"
ls -la "$OUTPUT_DIR/deb/"
echo ""
echo "RPM packages: $OUTPUT_DIR/rpm/"
ls -la "$OUTPUT_DIR/rpm/"

echo ""
echo "Minimal Docker-based package build completed successfully!"
echo "DEB packages now use alpine:latest + dpkg instead of debian:bullseye-slim + dpkg-dev"