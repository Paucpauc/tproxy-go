package proxy

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Test utility functions

// loadHexDump loads a hex dump file and converts it to a byte array
func loadHexDump(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Remove all whitespace and convert to single string
	hexStr := strings.ReplaceAll(string(data), " ", "")
	hexStr = strings.ReplaceAll(hexStr, "\n", "")
	hexStr = strings.ReplaceAll(hexStr, "\r", "")
	hexStr = strings.ReplaceAll(hexStr, "\t", "")

	// Decode hex string to bytes
	return hex.DecodeString(hexStr)
}

// extractTLSPayload extracts TLS payload from network packet data
// This removes TCP/IP headers to get the raw TLS handshake data
func extractTLSPayload(packetData []byte) []byte {
	// Look for TLS handshake signature (0x16 followed by version bytes)
	for i := 0; i < len(packetData)-2; i++ {
		if packetData[i] == 0x16 && packetData[i+1] == 0x03 &&
			(packetData[i+2] == 0x01 || packetData[i+2] == 0x02 || packetData[i+2] == 0x03) {
			return packetData[i:]
		}
	}
	return packetData // Return original if no TLS header found
}

// getExpectedSNIFromFilename extracts expected SNI from filename
// Format: sni_<domain>.hex
func getExpectedSNIFromFilename(filename string) string {
	base := filepath.Base(filename)
	if strings.HasPrefix(base, "sni_") && strings.HasSuffix(base, ".hex") {
		sni := base[4 : len(base)-4] // Remove "sni_" prefix and ".hex" suffix
		// Handle cases where filename might have additional parts
		if idx := strings.Index(sni, "."); idx != -1 {
			return sni
		}
	}
	return ""
}

// Test cases

func TestParseSNI_WithHexDumpFiles(t *testing.T) {
	testsDir := "../../tests"

	// Get all .hex files in tests directory
	files, err := os.ReadDir(testsDir)
	if err != nil {
		t.Skipf("Tests directory not found: %v", err)
		return
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".hex") {
			filename := filepath.Join(testsDir, file.Name())
			t.Run(file.Name(), func(t *testing.T) {
				// Load hex dump
				packetData, err := loadHexDump(filename)
				if err != nil {
					t.Fatalf("Failed to load hex dump %s: %v", filename, err)
				}

				// Extract TLS payload
				tlsData := extractTLSPayload(packetData)

				// Parse SNI
				sni := ParseSNI(tlsData)

				// Get expected SNI from filename
				expectedSNI := getExpectedSNIFromFilename(file.Name())

				if expectedSNI != "" {
					if sni != expectedSNI {
						t.Errorf("Expected SNI %s, got %s", expectedSNI, sni)
					} else {
						t.Logf("Successfully extracted SNI: %s", sni)
					}
				} else {
					// For files without expected SNI in filename, just verify we get something
					if sni == "" {
						t.Error("Failed to extract SNI from valid TLS handshake")
					} else {
						t.Logf("Extracted SNI: %s", sni)
					}
				}
			})
		}
	}
}

func TestParseSNI_ValidTLSHandshake(t *testing.T) {
	// Test with the specific play.googleapis.com hex dump
	filename := "../../tests/sni_play.googleapis.com.hex"
	packetData, err := loadHexDump(filename)
	if err != nil {
		t.Skipf("Test file not found: %v", err)
		return
	}

	tlsData := extractTLSPayload(packetData)
	sni := ParseSNI(tlsData)

	expected := "play.googleapis.com"
	if sni != expected {
		t.Errorf("Expected SNI %s, got %s", expected, sni)
	}
}

func TestParseSNI_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectedSNI string
		description string
	}{
		{
			name:        "EmptyData",
			data:        []byte{},
			expectedSNI: "",
			description: "Empty data should return empty SNI",
		},
		{
			name:        "NonTLSData",
			data:        []byte("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n"),
			expectedSNI: "",
			description: "Non-TLS data should return empty SNI",
		},
		{
			name: "IncompleteTLSHeader",
			data: []byte{
				0x16, 0x03, 0x01, // TLS handshake header
				// Missing length bytes
			},
			expectedSNI: "",
			description: "Incomplete TLS header should return empty SNI",
		},
		{
			name: "TLSWithoutSNIExtension",
			data: []byte{
				0x16, 0x03, 0x01, 0x00, 0x20, // TLS record header
				0x01, 0x00, 0x00, 0x1C, // ClientHello
				0x03, 0x03, // TLS 1.2
				// Random (32 bytes) - zeros for simplicity
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00,                   // Session ID length = 0
				0x00, 0x02, 0x00, 0x2F, // Cipher suites length = 2, one suite
				0x00, 0x00, // Compression methods length = 0
				0x00, 0x00, // Extensions length = 0 (no SNI extension)
			},
			expectedSNI: "",
			description: "TLS handshake without SNI extension should return empty SNI",
		},
		{
			name: "ValidSNIExtension",
			data: []byte{
				// TLS record header
				0x16, 0x03, 0x01, 0x00, 0x40,
				// ClientHello
				0x01, 0x00, 0x00, 0x3C, 0x03, 0x03,
				// Random (32 bytes)
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00,                   // Session ID length = 0
				0x00, 0x02, 0x00, 0x2F, // Cipher suites
				0x01, 0x00, // Compression methods
				// Extensions
				0x00, 0x10, // Extensions length = 16
				// SNI extension (type 0x0000)
				0x00, 0x00, 0x00, 0x0E, // Type + length
				// Server Name List
				0x00, 0x0C, // List length = 12
				// Server Name entry
				0x00,       // Name type = host_name
				0x00, 0x09, // Name length = 9
				// Domain name
				'l', 'o', 'c', 'a', 'l', 'h', 'o', 's', 't',
			},
			expectedSNI: "localhost",
			description: "Valid SNI extension should be extracted correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sni := ParseSNI(tt.data)
			if sni != tt.expectedSNI {
				t.Errorf("%s: expected %q, got %q", tt.description, tt.expectedSNI, sni)
			}
		})
	}
}

func TestParseSNI_PartialData(t *testing.T) {
	// Test with progressively larger chunks of data to ensure robustness
	filename := "../../tests/sni_play.googleapis.com.hex"
	packetData, err := loadHexDump(filename)
	if err != nil {
		t.Skipf("Test file not found: %v", err)
		return
	}

	tlsData := extractTLSPayload(packetData)

	// Test with increasing data sizes
	sizes := []int{10, 50, 100, 200, 500, len(tlsData)}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("PartialData_%d_bytes", size), func(t *testing.T) {
			if size > len(tlsData) {
				size = len(tlsData)
			}

			partialData := tlsData[:size]
			sni := ParseSNI(partialData)

			// For small sizes, we might not get the SNI, which is acceptable
			if size >= 100 { // Arbitrary threshold where SNI should be parseable
				if sni == "" {
					t.Logf("No SNI extracted from %d bytes (may be acceptable for very small sizes)", size)
				} else if sni != "play.googleapis.com" {
					t.Errorf("Unexpected SNI %s from %d bytes", sni, size)
				}
			}
		})
	}
}

func BenchmarkParseSNI(b *testing.B) {
	filename := "../../tests/sni_play.googleapis.com.hex"
	packetData, err := loadHexDump(filename)
	if err != nil {
		b.Skipf("Test file not found: %v", err)
		return
	}

	tlsData := extractTLSPayload(packetData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseSNI(tlsData)
	}
}

func TestParseHTTPHost_Basic(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
		port     int
	}{
		{
			name:     "ValidHTTPRequest",
			data:     []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"),
			expected: "example.com",
			port:     80,
		},
		{
			name:     "ValidHTTPRequestWithPort",
			data:     []byte("GET / HTTP/1.1\r\nHost: example.com:8080\r\n\r\n"),
			expected: "example.com",
			port:     8080,
		},
		{
			name:     "NoHostHeader",
			data:     []byte("GET / HTTP/1.1\r\nContent-Type: text/html\r\n\r\n"),
			expected: "",
			port:     80,
		},
		{
			name:     "InvalidPort",
			data:     []byte("GET / HTTP/1.1\r\nHost: example.com:invalid\r\n\r\n"),
			expected: "example.com",
			port:     80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port := ParseHTTPHost(tt.data)
			if host != tt.expected {
				t.Errorf("Expected host %s, got %s", tt.expected, host)
			}
			if port != tt.port {
				t.Errorf("Expected port %d, got %d", tt.port, port)
			}
		})
	}
}

func TestPipe_BasicDataTransfer(t *testing.T) {
	// Create two connected pipes
	clientConn, serverConn := net.Pipe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	// Start piping from client to server
	go Pipe(ctx, clientConn, serverConn, &wg)

	// Write data to client
	testData := []byte("Hello, World!")
	go func() {
		_, err := clientConn.Write(testData)
		if err != nil {
			t.Logf("Write error (expected after close): %v", err)
		}
	}()

	// Read data from server
	buf := make([]byte, len(testData))
	n, err := serverConn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read from server: %v", err)
	}

	if string(buf[:n]) != string(testData) {
		t.Errorf("Expected %q, got %q", string(testData), string(buf[:n]))
	}

	// Close connections and cancel context
	clientConn.Close()
	serverConn.Close()
	cancel()
	wg.Wait()
}

func TestPipe_ContextCancellation(t *testing.T) {
	clientConn, serverConn := net.Pipe()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	// Start piping
	go Pipe(ctx, clientConn, serverConn, &wg)

	// Cancel context immediately
	cancel()

	// Wait for pipe to finish
	wg.Wait()

	// Close connections
	clientConn.Close()
	serverConn.Close()

	// Verify connections are closed or in error state
	_, err := clientConn.Write([]byte("test"))
	if err == nil {
		t.Error("Expected write to fail after context cancellation")
	}
}

func TestConnectDirect_Success(t *testing.T) {
	// Start a test server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer listener.Close()

	// Accept connections in background
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	// Test connection
	host := "127.0.0.1"
	port := listener.Addr().(*net.TCPAddr).Port

	conn, err := ConnectDirect(host, port)
	if err != nil {
		t.Errorf("ConnectDirect failed: %v", err)
		return
	}
	defer conn.Close()

	if conn == nil {
		t.Error("ConnectDirect returned nil connection")
	}
}

func TestConnectDirect_InvalidHost(t *testing.T) {
	conn, err := ConnectDirect("invalid-host-that-does-not-exist", 9999)
	if err == nil {
		conn.Close()
		t.Error("Expected ConnectDirect to fail with invalid host")
	}
}

func TestConnectViaProxy_Success(t *testing.T) {
	// Start a mock proxy server
	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock proxy: %v", err)
	}
	defer proxyListener.Close()

	// Start target server
	targetListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start target server: %v", err)
	}
	defer targetListener.Close()

	// Mock proxy handler
	go func() {
		conn, err := proxyListener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read CONNECT request
		reader := bufio.NewReader(conn)
		request, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		// Verify it's a CONNECT request
		if !strings.HasPrefix(request, "CONNECT") {
			return
		}

		// Send successful response
		response := "HTTP/1.1 200 Connection Established\r\n\r\n"
		conn.Write([]byte(response))
	}()

	// Test proxy connection
	proxyHost := "127.0.0.1"
	proxyPort := proxyListener.Addr().(*net.TCPAddr).Port
	targetHost := "example.com"
	targetPort := 443
	clientIP := "192.168.1.1"

	conn, err := ConnectViaProxy(proxyHost, proxyPort, targetHost, targetPort, clientIP)
	if err != nil {
		t.Errorf("ConnectViaProxy failed: %v", err)
		return
	}
	defer conn.Close()

	if conn == nil {
		t.Error("ConnectViaProxy returned nil connection")
	}
}

func TestConnectViaProxy_ProxyError(t *testing.T) {
	// Start a mock proxy that returns error
	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start mock proxy: %v", err)
	}
	defer proxyListener.Close()

	go func() {
		conn, err := proxyListener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Send error response
		response := "HTTP/1.1 500 Internal Server Error\r\n\r\n"
		conn.Write([]byte(response))
	}()

	proxyHost := "127.0.0.1"
	proxyPort := proxyListener.Addr().(*net.TCPAddr).Port

	conn, err := ConnectViaProxy(proxyHost, proxyPort, "example.com", 443, "192.168.1.1")
	if err == nil {
		conn.Close()
		t.Error("Expected ConnectViaProxy to fail with proxy error")
	}
}

func TestConnectViaProxy_InvalidProxy(t *testing.T) {
	conn, err := ConnectViaProxy("invalid-proxy", 9999, "example.com", 443, "192.168.1.1")
	if err == nil {
		conn.Close()
		t.Error("Expected ConnectViaProxy to fail with invalid proxy")
	}
}
