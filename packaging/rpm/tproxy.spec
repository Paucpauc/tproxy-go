Name: tproxy
Version: 1.0.0
Release: 1%{?dist}
Summary: Transparent HTTP/HTTPS Proxy Server
License: MIT
URL: https://github.com/Paucpauc/tproxy-go/
Source0: %{name}-%{version}.tar.gz

BuildArch: x86_64

%description
A Go implementation of a transparent HTTP/HTTPS proxy server with configurable routing rules.
The proxy extracts SNI from TLS connections and routes traffic to upstream proxies using domain names.

%prep
%setup -q

%build
go build -o tproxy ./cmd/tproxy

%install
mkdir -p %{buildroot}/usr/bin
install -m 755 tproxy %{buildroot}/usr/bin/tproxy

mkdir -p %{buildroot}/etc/tproxy
install -m 644 proxy_config.yaml %{buildroot}/etc/tproxy/

%files
/usr/bin/tproxy
/etc/tproxy/proxy_config.yaml

%doc README.md

%changelog
* Mon Jan 27 2026 Andrey Urbanovich <andrey@urbanovich.net> - 1.0.0-1
- Initial package build