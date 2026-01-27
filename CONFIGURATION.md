# TProxy Configuration Reference

## Configuration File Structure

The TProxy configuration uses YAML format. The main configuration file is typically located at `/etc/tproxy/proxy_config.yaml`.

### Complete Configuration Schema

```yaml
# TProxy Configuration File
# Version: 1.0.0

# Server listening configuration
listen:
  host: "127.0.0.1"      # Interface to bind to (default: 127.0.0.1)
  https_port: 3130       # HTTPS/SNI proxy port (default: 3130)
  http_port: 3131        # HTTP proxy port (default: 3131)

# Logging configuration (optional)
logging:
  level: "info"          # Log level: debug, info, warn, error
  format: "text"         # Log format: text, json

# Routing rules - processed in order
rules:
  - pattern: "regex_pattern"  # Regular expression to match domains
    proxy: "action"           # Action: DIRECT, DROP, or proxy_host:port
    comment: "description"    # Optional description (for documentation)

# Advanced settings (optional)
advanced:
  timeout: 30           # Connection timeout in seconds (default: 30)
  max_connections: 1000 # Maximum concurrent connections (default: 1000)
  buffer_size: 8192     # Buffer size for data transfer (default: 8192)
```

## Configuration Sections

### Listen Section

Defines how TProxy listens for incoming connections.

**Parameters:**
- `host`: IP address or hostname to bind to
  - `"127.0.0.1"`: Listen only on localhost (secure)
  - `"0.0.0.0"`: Listen on all interfaces
  - Specific IP: `"192.168.1.100"`
- `https_port`: Port for HTTPS/SNI proxy (typically 443 redirects here)
- `http_port`: Port for HTTP proxy (typically 80 redirects here)

**Examples:**
```yaml
listen:
  host: "0.0.0.0"      # Listen on all interfaces
  https_port: 8443     # Custom HTTPS port
  http_port: 8080      # Custom HTTP port
```

### Rules Section

Defines the routing logic for incoming connections. Rules are processed in order, and the first matching rule is applied.

**Rule Structure:**
```yaml
- pattern: "regex_pattern"    # PCRE-compatible regular expression
  proxy: "action"             # What to do with matching traffic
  comment: "Human readable description"
```

**Proxy Actions:**
- `"DIRECT"`: Connect directly to the target server
- `"DROP"`: Block the connection entirely
- `"proxy_host:port"`: Route through specified upstream proxy
- `"socks5://proxy:port"`: Use SOCKS5 proxy (if supported)
- `"http://user:pass@proxy:port"`: HTTP proxy with authentication

### Pattern Examples

#### Basic Domain Matching
```yaml
# Exact domain match
- pattern: "example\\.com"
  proxy: "DIRECT"

# Subdomain matching
- pattern: ".*\\.example\\.com"
  proxy: "proxy.internal:8080"

# TLD matching
- pattern: ".*\\.org"
  proxy: "DIRECT"
```

#### Complex Patterns
```yaml
# Multiple domains with OR logic
- pattern: "(google\\.com|youtube\\.com)"
  proxy: "DIRECT"

# Exclude specific subdomains
- pattern: "(?!api\\.).*\\.company\\.com"
  proxy: "DIRECT"

# IP address ranges
- pattern: "192\\.168\\.1\\..*"
  proxy: "DIRECT"
```

#### Priority and Ordering
Rules are processed top-to-bottom. More specific rules should come before general catch-alls.

```yaml
rules:
  # Specific rules first
  - pattern: "api\\.critical\\.com"
    proxy: "secure-proxy:8080"
    comment: "Critical API traffic"
  
  - pattern: ".*\\.internal\\.com"
    proxy: "DIRECT"
    comment: "Internal domains"
  
  # General rules last
  - pattern: ".*"
    proxy: "default-proxy:8080"
    comment: "Everything else"
```

## Complete Configuration Examples

### Corporate Environment
```yaml
listen:
  host: "192.168.1.100"
  https_port: 3130
  http_port: 3131

logging:
  level: "info"
  format: "text"

rules:
  # Internal services - direct access
  - pattern: ".*\\.internal\\.company\\.com"
    proxy: "DIRECT"
    comment: "Internal company domains"
  
  - pattern: ".*\\.local"
    proxy: "DIRECT"
    comment: "Local network resources"
  
  # Trusted external services
  - pattern: ".*\\.google\\.com"
    proxy: "DIRECT"
    comment: "Google services"
  
  - pattern: ".*\\.microsoft\\.com"
    proxy: "DIRECT"
    comment: "Microsoft services"
  
  # Blocked categories
  - pattern: ".*\\.social\\.com"
    proxy: "DROP"
    comment: "Social media blocking"
  
  - pattern: ".*\\.gaming\\.com"
    proxy: "DROP"
    comment: "Gaming sites blocking"
  
  # Department-specific proxies
  - pattern: ".*\\.engineering\\.com"
    proxy: "eng-proxy:8080"
    comment: "Engineering team proxy"
  
  - pattern: ".*\\.marketing\\.com"
    proxy: "marketing-proxy:8080"
    comment: "Marketing team proxy"
  
  # Default corporate proxy
  - pattern: ".*"
    proxy: "corporate-proxy.company.com:8080"
    comment: "Default corporate proxy"

advanced:
  timeout: 60
  max_connections: 2000
```

### Home Network with Parental Controls
```yaml
listen:
  host: "192.168.0.1"
  https_port: 3130
  http_port: 3131

rules:
  # Educational sites - direct access
  - pattern: "(wikipedia\\.org|khanacademy\\.org|.*\\.edu)"
    proxy: "DIRECT"
    comment: "Educational resources"
  
  # Kid-friendly content
  - pattern: ".*\\.pbskids\\.org"
    proxy: "DIRECT"
    comment: "PBS Kids"
  
  # Block inappropriate content
  - pattern: ".*\\.adult\\.com"
    proxy: "DROP"
    comment: "Adult content blocking"
  
  - pattern: ".*\\.gambling\\.com"
    proxy: "DROP"
    comment: "Gambling sites blocking"
  
  # Time-limited access
  - pattern: ".*\\.youtube\\.com"
    proxy: "youtube-proxy:8080"
    comment: "YouTube with time limits"
  
  # General internet access
  - pattern: ".*"
    proxy: "family-proxy:8080"
    comment: "General family filtering"
```

### Development Environment
```yaml
listen:
  host: "127.0.0.1"
  https_port: 3130
  http_port: 3131

rules:
  # Local development
  - pattern: ".*\\.local"
    proxy: "DIRECT"
    comment: "Local development"
  
  - pattern: "localhost"
    proxy: "DIRECT"
    comment: "Localhost"
  
  # Internal services
  - pattern: ".*\\.internal\\.company\\.com"
    proxy: "DIRECT"
    comment: "Internal services"
  
  # Development tools
  - pattern: "(github\\.com|gitlab\\.com|docker\\.com)"
    proxy: "DIRECT"
    comment: "Development tools"
  
  # Testing through specific proxy
  - pattern: ".*\\.test\\.com"
    proxy: "test-proxy:8080"
    comment: "Test environment"
  
  # Default development proxy
  - pattern: ".*"
    proxy: "dev-proxy:8080"
    comment: "Default development proxy"
```

## Environment Variables

TProxy supports configuration through environment variables for containerized deployments:

```bash
# Override configuration values
export TPROXY_HOST="0.0.0.0"
export TPROXY_HTTPS_PORT="8443"
export TPROXY_HTTP_PORT="8080"
export TPROXY_CONFIG_FILE="/app/config.yaml"

# Run with environment variables
tproxy
```

## Validation and Testing

### Configuration Validation
```bash
# Check configuration syntax
tproxy --config /etc/tproxy/proxy_config.yaml --validate

# Dry run to test configuration
tproxy --config /etc/tproxy/proxy_config.yaml --dry-run
```

### Testing Specific Rules
```bash
# Test domain matching
curl -x http://127.0.0.1:3131 http://example.com

# Test HTTPS/SNI extraction
curl -x http://127.0.0.1:3131 https://example.com
```

## Best Practices

### Security
- Use `127.0.0.1` as the default host for security
- Regularly audit and update routing rules
- Use specific patterns rather than broad catch-alls
- Implement proper logging and monitoring

### Performance
- Place frequently matched rules at the top
- Use specific patterns before general ones
- Monitor connection counts and adjust `max_connections` as needed
- Consider timeout values based on network latency

### Maintenance
- Document rules with comments
- Version control configuration files
- Test configuration changes in staging first
- Regularly review and clean up unused rules

---

*For quick start instructions, see [QUICK_START.md](QUICK_START.md)*  
*For complete documentation, see [WIKI.md](WIKI.md)*