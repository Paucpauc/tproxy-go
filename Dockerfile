# Multi-stage Dockerfile for building a static Go binary
# Supports multiple architectures: amd64 (x64) and armv7 (for Mikrotik HAP AC2)
ARG TARGETARCH
ARG TARGETVARIANT

FROM golang:1.21-alpine AS builder

# Install required packages for building
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build static binary with optimizations for target architecture
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT:-7} \
    go build -a -installsuffix cgo -ldflags='-w -s -extldflags "-static"' -o tproxy ./cmd/tproxy

# Create minimal runtime image
FROM scratch

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy SSL certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the static binary
COPY --from=builder /app/tproxy /tproxy

# Copy default configuration
COPY proxy_config.yaml /proxy_config.yaml

# Expose default ports
EXPOSE 3130 3131

# Set the entrypoint
ENTRYPOINT ["/tproxy"]

# Default command with config path
CMD ["-config", "/proxy_config.yaml"]