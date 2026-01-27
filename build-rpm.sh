#!/bin/bash

# Build RPM package for tproxy using Docker
# Usage: ./build-rpm.sh [arch]

set -e

# Default architecture
DEFAULT_ARCH="amd64"
ARCH=${1:-$DEFAULT_ARCH}

# Supported architectures for RPM packages
SUPPORTED_ARCHS=("amd64" "arm64")

# Validate architecture
if [[ ! " ${SUPPORTED_ARCHS[@]} " =~ " ${ARCH} " ]]; then
    echo "Error: Unsupported architecture '$ARCH' for RPM packages"
    echo "Supported architectures: ${SUPPORTED_ARCHS[*]}"
    exit 1
fi

# Set RPM architecture names
case $ARCH in
    "amd64")
        RPM_ARCH="x86_64"
        ;;
    "arm64")
        RPM_ARCH="aarch64"
        ;;
esac

# Build directories
BUILD_DIR="build"
PKG_DIR="$BUILD_DIR/packages/rpm"
RPM_BUILD_DIR="$BUILD_DIR/rpm-build"

echo "Building RPM package for $ARCH ($RPM_ARCH)..."

# Clean and create directories
rm -rf "$RPM_BUILD_DIR"
mkdir -p "$PKG_DIR" "$RPM_BUILD_DIR"

# Build the binary first
echo "Building binary for $ARCH..."
./build.sh "$ARCH" "$RPM_BUILD_DIR"

# Create RPM build structure
echo "Creating RPM build structure..."
mkdir -p "$RPM_BUILD_DIR/SOURCES"
mkdir -p "$RPM_BUILD_DIR/SPECS"
mkdir -p "$RPM_BUILD_DIR/BUILD"
mkdir -p "$RPM_BUILD_DIR/RPMS"

# Copy source files
cp "$RPM_BUILD_DIR/tproxy-$ARCH" "$RPM_BUILD_DIR/SOURCES/tproxy"
cp proxy_config.yaml "$RPM_BUILD_DIR/SOURCES/"
cp README.md "$RPM_BUILD_DIR/SOURCES/"
cp packaging/tproxy.service "$RPM_BUILD_DIR/SOURCES/"

# Create tarball for RPM build
cd "$RPM_BUILD_DIR/SOURCES"
tar -czf tproxy-1.0.0.tar.gz tproxy proxy_config.yaml README.md tproxy.service
cd -

# Create spec file
cat > "$RPM_BUILD_DIR/SPECS/tproxy.spec" << EOF
Name: tproxy
Version: 1.0.0
Release: 1%{?dist}
Summary: Transparent HTTP/HTTPS Proxy Server
License: MIT
URL: https://github.com/Paucpauc/tproxy-go/
Source0: tproxy-1.0.0.tar.gz

BuildArch: $RPM_ARCH

%description
A Go implementation of a transparent HTTP/HTTPS proxy server with configurable routing rules.
The proxy extracts SNI from TLS connections and routes traffic to upstream proxies using domain names.

%prep
%setup -q

%build
# Binary is already built, just copy it
# Disable build ID generation for static Go binaries
%global _build_id_links none
cp tproxy %{_builddir}/tproxy

%install
mkdir -p %{buildroot}/usr/bin
install -m 755 %{_builddir}/tproxy %{buildroot}/usr/bin/tproxy

mkdir -p %{buildroot}/etc/tproxy
install -m 644 proxy_config.yaml %{buildroot}/etc/tproxy/

mkdir -p %{buildroot}/usr/lib/systemd/system
install -m 644 tproxy.service %{buildroot}/usr/lib/systemd/system/tproxy.service

%files
/usr/bin/tproxy
/etc/tproxy/proxy_config.yaml
/usr/lib/systemd/system/tproxy.service

%doc README.md

%post
%{_bindir}/systemctl daemon-reload >/dev/null 2>&1 || :
%{_bindir}/systemctl enable tproxy.service >/dev/null 2>&1 || :

%preun
%{_bindir}/systemctl --no-reload disable tproxy.service >/dev/null 2>&1 || :
%{_bindir}/systemctl stop tproxy.service >/dev/null 2>&1 || :

%postun
%{_bindir}/systemctl daemon-reload >/dev/null 2>&1 || :

%changelog
* Mon Jan 27 2025 Andrey Urbanovich <andrey@urbanovich.net> - 1.0.0-1
- Initial package build
- Added systemd service support
EOF

# Build RPM package using Docker
echo "Building RPM package using Docker..."
docker run --rm \
    -v "$(pwd)/$RPM_BUILD_DIR:/rpmbuild" \
    -w /rpmbuild \
    centos:7 \
    bash -c "
        yum install -y rpm-build && \
        rpmbuild -bb \
            --define '_topdir /rpmbuild' \
            --define '_builddir %{_topdir}/BUILD' \
            --define '_rpmdir %{_topdir}/RPMS' \
            --define '_sourcedir %{_topdir}/SOURCES' \
            --define '_specdir %{_topdir}/SPECS' \
            --define '_srcrpmdir %{_topdir}/SRPMS' \
            SPECS/tproxy.spec
    "

# Find and copy the built RPM
RPM_FILE=$(find "$RPM_BUILD_DIR/RPMS" -name "tproxy-1.0.0-1.*.rpm" | head -1)
if [ -n "$RPM_FILE" ]; then
    cp "$RPM_FILE" "$PKG_DIR/"
    echo "RPM package built: $PKG_DIR/$(basename "$RPM_FILE")"
else
    echo "Error: RPM package not found"
    exit 1
fi

# Cleanup
rm -rf "$RPM_BUILD_DIR"

echo "RPM package build completed successfully!"