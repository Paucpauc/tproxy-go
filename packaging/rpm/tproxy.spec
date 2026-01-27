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
# Binary is pre-built, no compilation needed

%install
mkdir -p %{buildroot}/usr/bin
install -m 755 tproxy %{buildroot}/usr/bin/tproxy

mkdir -p %{buildroot}/etc/tproxy
install -m 644 proxy_config.yaml %{buildroot}/etc/tproxy/

mkdir -p %{buildroot}/usr/lib/systemd/system
install -m 644 packaging/tproxy.service %{buildroot}/usr/lib/systemd/system/tproxy.service

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