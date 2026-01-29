# Go TProxy - Transparent HTTP/HTTPS Proxy Server

This is a Go implementation of a transparent HTTP/HTTPS proxy server with configurable routing rules. The proxy is unique in its ability to extract SNI (Server Name Indication) from TLS connections and route traffic to upstream proxies using domain names instead of IP addresses, unlike traditional proxies.

## 📚 Documentation

**Complete documentation is now available:**

- **[WIKI.md](WIKI.md)** - Comprehensive documentation with installation guides, configuration examples, and advanced usage
- **[QUICK_START.md](QUICK_START.md)** - Quick start guide for immediate deployment
- **[CONFIGURATION.md](CONFIGURATION.md)** - Detailed configuration reference and examples

## 🚀 Quick Start

### Installation Options

1. **DEB Package (Debian/Ubuntu)**:
   ```bash
   wget https://github.com/Paucpauc/tproxy-go/releases/latest/download/tproxy_1.0.0_amd64.deb
   sudo dpkg -i tproxy_1.0.0_amd64.deb
   ```


3. **Docker**:
   ```bash
   docker run -d --name tproxy --network host ghcr.io/paucpauc/tproxy:latest
   ```

4. **Mikrotik Container**:
   ```bash
   /container add remote-image=ghcr.io/paucpauc/tproxy:latest-arm
   ```

### Basic Configuration

Create `/etc/tproxy/proxy_config.yaml`:
```yaml
listen:
  host: "127.0.0.1"
  https_port: 3130
  http_port: 3131
  timeout: 900

rules:
  - pattern: ".*\\.google\\.com"
    proxy: "DIRECT"
  - pattern: ".*"
    proxy: "proxy.example.com:8080"
```

## ✨ Features

- **Transparent Proxy**: Intercepts traffic transparently without client configuration
- **SNI Extraction**: Extracts domain names from TLS handshakes for intelligent routing
- **Domain-Based Routing**: Routes to upstream proxies using domain names instead of IP addresses
- **HTTP Proxy**: Parses Host headers to determine routing
- **HTTPS Proxy**: Uses SNI (Server Name Indication) parsing for TLS connections
- **Configurable Routing**: YAML-based configuration for proxy rules
- **Multiple Backends**: Support for direct connections, proxy chains, and connection dropping
- **Asynchronous**: Concurrent connection handling using Go's goroutines
- **Multi-Architecture**: Builds for amd64, arm64, and arm (ARMv7) architectures

## 🏗️ Build System

The project supports multiple build methods:

### GitHub Actions (Automated CI/CD)
- **Continuous Integration**: Tests run on every push and pull request
- **Automated Releases**: Binaries, packages, and Docker images built automatically on tag pushes
- **Multi-Architecture**: Builds for amd64, arm64, and arm (ARMv7) architectures
- **Nightly Builds**: Comprehensive builds and security scans run daily

### Docker-Based Builds (Recommended - No Local Tools Required)
```bash
# Build binaries for all architectures
make docker-build-all

# Build packages for all architectures
make docker-packages

# Build and push Docker images
make docker-push
```

### Traditional Builds (Requires Local Go Installation)
```bash
# Build for current architecture
make build

# Build for all architectures
make build-all

# Build packages
make packages
```

## 📦 Supported Platforms

| Platform | Architecture | Package Type | Container Support |
|----------|--------------|--------------|-------------------|
| Debian/Ubuntu | amd64, arm64 | DEB | ✅ |
| Generic Linux | amd64, arm64, arm | Binary | ✅ |
| Mikrotik RouterOS | arm (ARMv7) | Binary/Container | ✅ |
| Docker | Multi-arch | Image | ✅ |

## 🔧 Configuration

### Rule Patterns
- `DIRECT`: Connect directly to the target
- `DROP`: Block the connection
- `proxy_host:port`: Route through the specified proxy server

### Example Configuration
```yaml
listen:
  host: "127.0.0.1"
  https_port: 3130
  http_port: 3131
  timeout: 900

rules:
  - pattern: ".*\\.internal\\.com"
    proxy: "DIRECT"
  - pattern: ".*\\.google\\.com"
    proxy: "DIRECT"
  - pattern: ".*\\.blocked\\.com"
    proxy: "DROP"
  - pattern: ".*"
    proxy: "proxy.example.com:8080"
```

## 🐳 Docker Deployment

### Basic Docker Usage
```bash
# Run with default configuration
docker run -d --name tproxy --network host ghcr.io/paucpauc/tproxy:latest

# Run with custom configuration
docker run -d --name tproxy --network host \
  -v /path/to/config.yaml:/proxy_config.yaml \
  ghcr.io/paucpauc/tproxy:latest
```

### Docker Compose
```yaml
version: '3.8'
services:
  tproxy:
    image: ghcr.io/paucpauc/tproxy:latest
    network_mode: host
    volumes:
      - ./config:/etc/tproxy
    restart: unless-stopped
```

## 🔍 Transparent Proxy Setup

To use this proxy in transparent mode, configure iptables rules:

```bash
# Redirect HTTPS traffic from specific clients to SNI proxy
# Replace 192.168.1.0/24 with your client network
iptables -t nat -I PREROUTING -s 192.168.1.0/24 -p tcp --dport 443 \
  -j DNAT --to-destination 127.0.0.1:3130

# Redirect HTTP traffic from specific clients to HTTP proxy
iptables -t nat -I PREROUTING -s 192.168.1.0/24 -p tcp --dport 80 \
  -j DNAT --to-destination 127.0.0.1:3131

# Exclude proxy server traffic from redirection (prevent dead loop)
iptables -t nat -I PREROUTING -s 127.0.0.1 -j ACCEPT
iptables -t nat -I PREROUTING -s 192.168.1.100 -j ACCEPT  # Replace with proxy server IP
```

**Key Advantage**: Unlike traditional transparent proxies that route based on IP addresses, this proxy extracts the SNI from TLS handshakes and uses the actual domain name when connecting to upstream proxies, providing more accurate routing and better compatibility with modern web services.

## 📖 Usage

1. **Run with default configuration**:
   ```bash
   ./tproxy
   ```

2. **Run with custom configuration file**:
   ```bash
   ./tproxy --config /path/to/config.yaml
   ```

3. **Run ARM version on Mikrotik**:
   ```bash
   ./build/tproxy-arm --config proxy_config.yaml
   ```

## 🧪 Testing

```bash
# Run tests using local Go
make test

# Run tests using Docker
make docker-test

# Run comprehensive tests
make test-all
```

## 📊 Project Structure

```
tproxy-go/
├── cmd/tproxy/          # Main application entry point
├── internal/
│   ├── config/          # Configuration loading and parsing
│   ├── proxy/           # Proxy connection handling
│   └── server/          # HTTP/HTTPS server implementation
├── packaging/           # DEB package definitions
├── tests/               # Test data files
├── WIKI.md             # Comprehensive documentation
├── QUICK_START.md      # Quick start guide
└── CONFIGURATION.md    # Configuration reference
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request
5. Ensure all tests pass

### GitHub Actions Workflows

The project uses GitHub Actions for automated testing and deployment:

- **CI Workflow**: Runs on every push and pull request
  - Unit tests with multiple Go versions
  - Linting and code validation
  - Build validation for all architectures
  - Security scanning

- **Build Workflow**: Runs on tag pushes (`v*`)
  - Builds binaries for amd64, arm64, and arm
  - Creates DEB packages
  - Builds and pushes multi-arch Docker images
  - Automatically creates GitHub releases

- **Nightly Workflow**: Runs daily at 2 AM UTC
  - Comprehensive builds and tests
  - Security vulnerability scanning
  - Dependency update checks

## 📄 License

This project is provided as-is for educational and development purposes.

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/Paucpauc/tproxy-go/issues)
- **Documentation**: [WIKI.md](WIKI.md)
- **Releases**: [GitHub Releases](https://github.com/Paucpauc/tproxy-go/releases)

---

*For detailed documentation, see [WIKI.md](WIKI.md)*  
*For quick setup, see [QUICK_START.md](QUICK_START.md)*  
*For configuration details, see [CONFIGURATION.md](CONFIGURATION.md)*