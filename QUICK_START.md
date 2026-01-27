# TProxy Quick Start Guide

## One-Minute Setup

### Prerequisites
- Linux system with iptables
- Root access for iptables configuration

### Installation Options

#### Option 1: Direct Binary Download
```bash
# Download binary
wget https://github.com/Paucpauc/tproxy-go/releases/latest/download/tproxy-amd64
chmod +x tproxy-amd64
sudo mv tproxy-amd64 /usr/local/bin/tproxy
```

#### Option 2: DEB Package (Debian/Ubuntu)
```bash
wget https://github.com/Paucpauc/tproxy-go/releases/latest/download/tproxy_1.0.0_amd64.deb
sudo dpkg -i tproxy_1.0.0_amd64.deb
```

#### Option 3: RPM Package (CentOS/RHEL/Fedora)
```bash
wget https://github.com/Paucpauc/tproxy-go/releases/latest/download/tproxy-1.0.0-1.x86_64.rpm
sudo rpm -i tproxy-1.0.0-1.x86_64.rpm
```

#### Option 4: Docker
```bash
docker run -d --name tproxy --network host \
  -v /etc/tproxy:/etc/tproxy \
  ghcr.io/paucpauc/tproxy:latest
```

### Basic Configuration

Create `/etc/tproxy/proxy_config.yaml`:
```yaml
listen:
  host: "127.0.0.1"
  https_port: 3130
  http_port: 3131

rules:
  - pattern: ".*\\.google\\.com"
    proxy: "DIRECT"
  - pattern: ".*"
    proxy: "proxy.example.com:8080"
```

### Start the Proxy

```bash
# Run directly
tproxy --config /etc/tproxy/proxy_config.yaml

# Or as systemd service
sudo systemctl start tproxy
```

### Configure iptables

```bash
# Redirect HTTP traffic
iptables -t nat -I PREROUTING -p tcp --dport 80 \
  -j DNAT --to-destination 127.0.0.1:3131

# Redirect HTTPS traffic
iptables -t nat -I PREROUTING -p tcp --dport 443 \
  -j DNAT --to-destination 127.0.0.1:3130
```

### Verify Operation

```bash
# Check if running
ps aux | grep tproxy

# Test HTTP proxy
curl -x http://127.0.0.1:3131 http://example.com

# Check listening ports
netstat -tlnp | grep 313
```

## Common Commands

### Building from Source
```bash
# Clone and build
git clone https://github.com/Paucpauc/tproxy-go.git
cd tproxy-go
make docker-build-all
```

### Package Management
```bash
# Build packages
make docker-packages

# Install DEB package
sudo dpkg -i build/packages/deb/tproxy_1.0.0_amd64.deb

# Install RPM package
sudo rpm -i build/packages/rpm/tproxy-1.0.0-1.x86_64.rpm
```

### Docker Operations
```bash
# Build image
docker build -t tproxy:latest .

# Run container
docker run -d --name tproxy --network host tproxy:latest

# View logs
docker logs tproxy
```

### Mikrotik Deployment
```bash
# Pull ARMv7 image
/container add remote-image=ghcr.io/paucpauc/tproxy:latest-arm

# Configure iptables
/ip firewall nat add chain=dstnat protocol=tcp dst-port=443 \
  action=dst-nat to-addresses=127.0.0.1 to-ports=3130
```

## Troubleshooting Quick Tips

### Service Not Starting
```bash
# Check configuration
tproxy --config /etc/tproxy/proxy_config.yaml --verbose

# Verify file permissions
ls -la /etc/tproxy/proxy_config.yaml
```

### Traffic Not Redirected
```bash
# Check iptables rules
iptables -t nat -L -n -v

# Verify TProxy is listening
ss -tlnp | grep tproxy
```

### Connection Issues
```bash
# Test upstream proxy
telnet proxy.example.com 8080

# Check DNS resolution
nslookup proxy.example.com
```

## Configuration Examples

### Simple Direct/Drop Rules
```yaml
rules:
  - pattern: ".*\\.trusted\\.com"
    proxy: "DIRECT"
  - pattern: ".*\\.blocked\\.com"
    proxy: "DROP"
  - pattern: ".*"
    proxy: "proxy.internal:8080"
```

### Multi-Proxy Setup
```yaml
rules:
  - pattern: ".*\\.cdn\\.com"
    proxy: "cdn-proxy:3128"
  - pattern: ".*\\.api\\.com"
    proxy: "api-proxy:8080"
  - pattern: ".*"
    proxy: "default-proxy:8080"
```

---

*For detailed documentation, see [WIKI.md](WIKI.md)*