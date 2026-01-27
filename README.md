# Go TProxy - HTTP/HTTPS Proxy Server

This is a Go implementation of the Python `tproxy.py` script, providing an asynchronous HTTP/HTTPS proxy server with configurable routing rules.

## Features

- **HTTP Proxy**: Parses Host headers to determine routing
- **HTTPS Proxy**: Uses SNI (Server Name Indication) parsing for TLS connections
- **Configurable Routing**: YAML-based configuration for proxy rules
- **Multiple Backends**: Support for direct connections, proxy chains, and connection dropping
- **Asynchronous**: Concurrent connection handling using Go's goroutines

## Project Structure

```
tproxy/
├── cmd/
│   └── tproxy/
│       └── main.go          # Main application entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration types and loading
│   ├── proxy/
│   │   ├── proxy.go         # Proxy connection handling
│   │   ├── sni.go           # SNI parsing logic
│   │   ├── http.go          # HTTP parsing logic
│   │   └── proxy_test.go    # Proxy functionality tests
│   └── server/
│       ├── server.go        # HTTP/HTTPS server implementation
│       └── handler.go       # Connection handlers
├── tests/                   # Test data
├── go.mod                   # Go module definition
├── proxy_config.yaml        # Configuration file
├── Dockerfile              # Multi-architecture Docker build
├── Makefile                # Build automation
├── build.sh                # Single architecture build script
├── build-all.sh            # Multi-architecture build script
└── README.md               # Documentation
```

## Configuration

The proxy uses a YAML configuration file. Example `proxy_config.yaml`:

```yaml
listen:
  host: "127.0.0.1"
  https_port: 3130
  http_port: 3131

rules:
  - pattern: ".*\.google\.com"
    proxy: "DIRECT"
  - pattern: ".*\.yandex\.ru"
    proxy: "DIRECT"
  - pattern: ".*\.internal\.com"
    proxy: "DROP"
  - pattern: ".*"
    proxy: "proxy.example.com:8080"
```

### Rule Patterns

- `DIRECT`: Connect directly to the target
- `DROP`: Block the connection
- `proxy_host:port`: Route through the specified proxy server

## Multi-Architecture Builds

The project now supports building for multiple architectures:

- **amd64**: x64 systems
- **arm64**: ARM64 systems
- **arm**: ARMv7 (for Mikrotik HAP AC2 routers)

### Build Options

1. **Build for current architecture only**:
   ```bash
   make build
   # or
   ./build.sh
   ```

2. **Build for all architectures**:
   ```bash
   make build-all
   # or
   ./build-all.sh
   ```

3. **Build for specific architecture**:
   ```bash
   ./build.sh arm    # For Mikrotik HAP AC2
   ./build.sh amd64  # For x64 systems
   ./build.sh arm64  # For ARM64 systems
   ```

4. **Build Docker images**:
   ```bash
   make docker-all   # Build Docker images for all architectures
   make docker-push  # Build and push to registry
   ```

### Usage

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

## API Comparison with Python Version

| Feature | Python tproxy.py | Go tproxy |
|---------|------------------|-----------|
| HTTP Proxy | ✅ | ✅ |
| HTTPS Proxy with SNI | ✅ | ✅ |
| YAML Configuration | ✅ | ✅ |
| Proxy Routing Rules | ✅ | ✅ |
| Direct Connections | ✅ | ✅ |
| Proxy Chains | ✅ | ✅ |
| Connection Dropping | ✅ | ✅ |
| Asynchronous | asyncio | goroutines |

## Key Differences

1. **Platform Support**: The Go version doesn't implement Linux-specific `SO_ORIGINAL_DST` functionality
2. **SNI Parsing**: Uses custom SNI parsing instead of Python's ssl library
3. **Concurrency**: Uses Go's native goroutines instead of asyncio
4. **Error Handling**: Go's error handling patterns vs Python exceptions

## Limitations

- SNI parsing is simplified and may not handle all TLS handshake variations
- IPv6 support in proxy addresses is basic
- No support for authentication in proxy connections
- Limited error recovery and logging compared to production-grade proxies

## Development

To modify or extend the proxy:

1. **Add new proxy types**: Modify `ProxyAction` struct and connection logic
2. **Enhance SNI parsing**: Improve the `parseSNI` function in `proxy.go`
3. **Add authentication**: Extend proxy connection functions
4. **Improve logging**: Add structured logging with levels

### Build System

The build system supports cross-compilation for multiple architectures:

- **Dockerfile**: Updated with `TARGETARCH` and `TARGETVARIANT` support
- **Build scripts**: Handle architecture-specific compilation flags
- **Makefile**: Provides convenient build targets

### Project Structure

The project follows standard Go conventions:

- **cmd/tproxy/**: Contains the main application entry point
- **internal/**: Contains application-specific packages not intended for external use
  - **config/**: Configuration loading and parsing
  - **proxy/**: Proxy connection handling and protocol parsing
  - **server/**: HTTP/HTTPS server implementation
- **tests/**: Test data files

## Testing

Basic functionality can be tested by:

1. Starting the proxy server
2. Configuring a browser or HTTP client to use the proxy
3. Testing different domains to verify routing rules

## License

This project is provided as-is for educational and development purposes.