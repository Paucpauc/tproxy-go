package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"tproxy/internal/config"
	"tproxy/internal/proxy"
)

// Mock connection for testing
type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
	mu       sync.Mutex
}

func newMockConn() *mockConn {
	return &mockConn{
		readBuf:  &bytes.Buffer{},
		writeBuf: &bytes.Buffer{},
		closed:   false,
	}
}

func (m *mockConn) Read(b []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, fmt.Errorf("connection closed")
	}
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, fmt.Errorf("connection closed")
	}
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(192, 168, 1, 1), Port: 12345}
}

func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func (m *mockConn) WriteData(data []byte) {
	m.readBuf.Write(data)
}

func (m *mockConn) GetWrittenData() []byte {
	return m.writeBuf.Bytes()
}

func TestProxyConnection_Direct(t *testing.T) {
	// Create mock client connection
	clientConn := newMockConn()

	// Create a simple HTTP request data
	httpRequest := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	clientConn.WriteData([]byte(httpRequest))

	// Since we can't easily test the actual network connection in unit tests,
	// we'll test the proxyConnection function with a mock that simulates failure
	// due to missing network connectivity in test environment
	targetHost := "example.com"
	targetPort := 80
	originalIP := "192.168.1.1"
	clientIP := "192.168.1.2"
	proxyAction := &config.ProxyAction{Type: "DIRECT"}
	initialData := []byte(httpRequest)
	// This should attempt to connect and fail (which is expected in test environment)
	proxyConnection(targetHost, targetPort, originalIP, clientIP, clientConn, proxyAction, initialData, 30) // 30 second timeout

	// Verify the connection was attempted (connection will be closed)
	if !clientConn.closed {
		t.Error("Expected client connection to be closed")
	}
}

func TestProxyConnection_Drop(t *testing.T) {
	clientConn := newMockConn()

	httpRequest := "GET / HTTP/1.1\r\nHost: blocked.com\r\n\r\n"
	clientConn.WriteData([]byte(httpRequest))

	proxyAction := &config.ProxyAction{Type: "DROP"}
	targetHost := "blocked.com"
	targetPort := 80
	originalIP := "192.168.1.1"
	clientIP := "192.168.1.2"
	initialData := []byte(httpRequest)
	proxyConnection(targetHost, targetPort, originalIP, clientIP, clientConn, proxyAction, initialData, 30) // 30 second timeout

	// For DROP action, the connection should be handled (may not necessarily close immediately in mock)
	// We'll verify the function executed without panicking
}

func TestProxyConnection_Proxy(t *testing.T) {
	clientConn := newMockConn()

	httpRequest := "GET / HTTP/1.1\r\nHost: proxied.com\r\n\r\n"
	clientConn.WriteData([]byte(httpRequest))

	proxyAction := &config.ProxyAction{
		Type: "PROXY",
		Host: "proxy.example.com",
		Port: 3128,
	}
	targetHost := "proxied.com"
	targetPort := 80
	originalIP := "192.168.1.1"
	clientIP := "192.168.1.2"
	initialData := []byte(httpRequest)
	// This will attempt proxy connection and fail (expected in test)
	proxyConnection(targetHost, targetPort, originalIP, clientIP, clientConn, proxyAction, initialData, 30) // 30 second timeout

	// For PROXY action, connection attempt will fail in test environment
	// We'll verify the function executed without panicking
}

func TestParseHTTPHostFromRequest(t *testing.T) {
	tests := []struct {
		name         string
		request      string
		expectedHost string
		expectedPort int
	}{
		{
			name:         "Basic HTTP request",
			request:      "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expectedHost: "example.com",
			expectedPort: 80,
		},
		{
			name:         "HTTP request with port",
			request:      "GET / HTTP/1.1\r\nHost: example.com:8080\r\n\r\n",
			expectedHost: "example.com",
			expectedPort: 8080,
		},
		{
			name:         "No host header",
			request:      "GET / HTTP/1.1\r\nContent-Type: text/html\r\n\r\n",
			expectedHost: "",
			expectedPort: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port := proxy.ParseHTTPHost([]byte(tt.request))
			if host != tt.expectedHost {
				t.Errorf("Expected host %s, got %s", tt.expectedHost, host)
			}
			if port != tt.expectedPort {
				t.Errorf("Expected port %d, got %d", tt.expectedPort, port)
			}
		})
	}
}

func TestParseSNIFromTLSHandshake(t *testing.T) {
	// Use the existing test file infrastructure
	filename := "../../tests/sni_play.googleapis.com.hex"
	packetData, err := os.ReadFile(filename)
	if err != nil {
		t.Skipf("Test file not found: %v", err)
		return
	}

	// Convert hex string to bytes (similar to proxy_test.go)
	hexStr := strings.ReplaceAll(string(packetData), " ", "")
	hexStr = strings.ReplaceAll(hexStr, "\n", "")
	hexStr = strings.ReplaceAll(hexStr, "\r", "")
	hexStr = strings.ReplaceAll(hexStr, "\t", "")

	packetBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	// Extract TLS payload by finding TLS handshake
	var tlsData []byte
	for i := 0; i < len(packetBytes)-2; i++ {
		if packetBytes[i] == 0x16 && packetBytes[i+1] == 0x03 &&
			(packetBytes[i+2] == 0x01 || packetBytes[i+2] == 0x02 || packetBytes[i+2] == 0x03) {
			tlsData = packetBytes[i:]
			break
		}
	}

	if tlsData == nil {
		t.Fatal("Failed to extract TLS payload")
	}

	sni := proxy.ParseSNI(tlsData)

	if sni != "play.googleapis.com" {
		t.Errorf("Expected SNI play.googleapis.com, got %s", sni)
	}
}

func TestStartServers_InvalidPorts(t *testing.T) {
	// Test with invalid configuration
	invalidConfig := &config.Config{
		Listen: config.ListenConfig{
			Host:      "invalid-host",
			HTTPSPort: -1, // Invalid port
			HTTPPort:  -1,
		},
		Rules: []config.Rule{},
	}

	err := StartServers(invalidConfig)
	if err == nil {
		t.Error("Expected StartServers to fail with invalid ports")
	}
}

// Test helper functions
func TestCreateMockServer(t *testing.T) {
	// Test creating a simple mock server for integration testing
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("Cannot create test listener: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			t.Logf("Listener close error: %v", err)
		}
	}()

	// Start server in background
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				t.Logf("Connection close error: %v", err)
			}
		}()

		// Simple echo server
		reader := bufio.NewReader(conn)
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		_, err = conn.Write([]byte("ECHO: " + line))
		if err != nil {
			return
		}
	}()

	// Test connection to mock server
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to mock server: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Logf("Connection close error: %v", err)
		}
	}()

	// Send test data
	testData := "Hello, Server!\n"
	_, err = conn.Write([]byte(testData))
	if err != nil {
		t.Fatalf("Failed to write to server: %v", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read from server: %v", err)
	}

	expected := "ECHO: " + testData
	if response != expected {
		t.Errorf("Expected response %q, got %q", expected, response)
	}

	// Cleanup
	if err := conn.Close(); err != nil {
		t.Logf("Connection close error: %v", err)
	}
	if err := listener.Close(); err != nil {
		t.Logf("Listener close error: %v", err)
	}
	wg.Wait()
}

// Test context cancellation in pipe operations
func TestContextCancellation(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer func() {
		if err := clientConn.Close(); err != nil {
			t.Logf("Client connection close error: %v", err)
		}
	}()
	defer func() {
		if err := serverConn.Close(); err != nil {
			t.Logf("Server connection close error: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	// Start pipe operation
	go proxy.Pipe(ctx, clientConn, serverConn, &wg)

	// Cancel context immediately
	cancel()

	// Wait for pipe to finish
	wg.Wait()

	// Verify pipe stopped
	_, err := clientConn.Write([]byte("test"))
	if err == nil {
		t.Error("Expected write to fail after context cancellation")
	}
}
