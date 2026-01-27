#!/bin/bash

# Docker-based package building script for tproxy
# Builds DEB and RPM packages using Docker containers
# No local Go or packaging tools required

set -e

# Default values
DEFAULT_OUTPUT="build/packages"
OUTPUT_DIR=${1:-$DEFAULT_OUTPUT}

# Supported architectures for packages (DEB/RPM don't support ARMv7 well)
PACKAGE_ARCHS=("amd64" "arm64")

echo "Building packages for architectures: ${PACKAGE_ARCHS[*]}"
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
        "arm64")
            GOARCH="arm64"
            RPMARCH="aarch64"
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
            go build -buildvcs=false -a -installsuffix cgo -ldflags='-w -s -extldflags \"-static\"' \
            -o /app/build/tproxy-$ARCH ./cmd/tproxy
        "
    
    echo "✓ Binary built: build/tproxy-$ARCH"
    
    # Build DEB package using Docker
    echo "Building DEB package for $ARCH..."
    docker run --rm \
        -v "$(pwd)":/app \
        -w /app \
        -e ARCH=$ARCH \
        debian:bullseye-slim \
        sh -c "
            apt-get update && \
            apt-get install -y dpkg-dev fakeroot sed && \
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
            cd /tmp/deb-build && \
            dpkg-deb --build . /app/$OUTPUT_DIR/deb/tproxy_${ARCH}.deb
        "
    
    echo "✓ DEB package built: $OUTPUT_DIR/deb/tproxy_${ARCH}.deb"
    
    # Build RPM package using Docker
    echo "Building RPM package for $ARCH..."
    docker run --rm \
        -v "$(pwd)":/app \
        -w /app \
        -e ARCH=$ARCH \
        -e RPMARCH=$RPMARCH \
        fedora:latest \
        sh -c "
            dnf install -y rpm-build sed && \
            mkdir -p /tmp/rpm-build/{BUILD,RPMS,SOURCES,SPECS,SRPMS} && \
            cp packaging/rpm/tproxy.spec /tmp/rpm-build/SPECS/ && \
            # Update BuildArch in spec file
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
            rpmbuild -bb --define \"_topdir /tmp/rpm-build\" --target $RPMARCH SPECS/tproxy.spec && \
            cp RPMS/$RPMARCH/*.rpm /app/$OUTPUT_DIR/rpm/
        "
    
    echo "✓ RPM package built: $OUTPUT_DIR/rpm/tproxy-*${RPMARCH}.rpm"
done

echo ""
echo "=== Package Build Summary ==="
echo "All packages built successfully:"
echo "DEB packages: $OUTPUT_DIR/deb/"
ls -la "$OUTPUT_DIR/deb/"
echo ""
echo "RPM packages: $OUTPUT_DIR/rpm/"
ls -la "$OUTPUT_DIR/rpm/"

echo ""
echo "Docker-based package build completed successfully!"