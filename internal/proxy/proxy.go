package proxy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ParseHTTPHost(data []byte) (string, int) {
	reader := bufio.NewReader(bytes.NewReader(data))

	// Read the first line (request line)
	_, err := reader.ReadString('\n')
	if err != nil {
		return "", 80
	}

	// Read headers until we find Host header
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}

		if strings.HasPrefix(line, "Host: ") {
			hostStr := strings.TrimSpace(strings.TrimPrefix(line, "Host: "))
			parts := strings.Split(hostStr, ":")
			host := parts[0]
			port := 80

			if len(parts) > 1 {
				if p, err := strconv.Atoi(parts[1]); err == nil {
					port = p
				}
			}
			return host, port
		}
	}

	return "", 80
}

func ParseSNI(data []byte) string {
	// Find TLS handshake in the data (skip TCP/IP headers)
	startPos := -1
	for i := 0; i < len(data)-2; i++ {
		if data[i] == 0x16 && data[i+1] == 0x03 && (data[i+2] == 0x01 || data[i+2] == 0x02 || data[i+2] == 0x03) {
			startPos = i
			break
		}
	}

	if startPos == -1 {
		return "" // Not a TLS handshake found
	}

	// Use data starting from TLS handshake
	data = data[startPos:]

	// Check if this is a TLS handshake
	if len(data) < 5 || data[0] != 0x16 {
		return "" // Not a TLS handshake
	}

	// Parse TLS record header
	recordLength := int(data[3])<<8 | int(data[4])
	// Don't require complete record - we only need the beginning to parse SNI
	// if len(data) < 5+recordLength {
	// 	return "" // Incomplete record
	// }

	// Check if this is a ClientHello
	if len(data) < 9 || data[5] != 0x01 {
		return "" // Not a ClientHello or insufficient data
	}

	// Parse handshake length (3 bytes)
	handshakeLength := int(data[6])<<16 | int(data[7])<<8 | int(data[8])

	// The handshake length should be <= recordLength - 4 (handshake type + length bytes)
	if handshakeLength > recordLength-4 {
		// Use record length as fallback if handshake length seems invalid
		handshakeLength = recordLength - 4
	}

	// Validate we have enough data for the handshake
	if len(data) < 9+handshakeLength {
		// If we don't have complete handshake, use available data
		handshakeLength = len(data) - 9
		if handshakeLength < 0 {
			return "" // Not enough data for handshake
		}
	}

	// Start parsing ClientHello at position 9
	pos := 9

	// Skip ClientVersion (2 bytes)
	if pos+2 > len(data) {
		return ""
	}
	pos += 2

	// Skip Random (32 bytes)
	if pos+32 > len(data) {
		return ""
	}
	pos += 32

	// Skip SessionID (1 byte length + session data)
	if pos >= len(data) {
		return ""
	}
	sessionIDLen := int(data[pos])
	pos++
	if pos+sessionIDLen > len(data) {
		return ""
	}
	pos += sessionIDLen

	// Skip CipherSuites (2 bytes length + cipher suites)
	if pos+2 > len(data) {
		return ""
	}
	cipherSuitesLen := int(data[pos])<<8 | int(data[pos+1])
	pos += 2
	if pos+cipherSuitesLen > len(data) {
		return ""
	}
	pos += cipherSuitesLen

	// Skip CompressionMethods (1 byte length + methods)
	if pos >= len(data) {
		return ""
	}
	compressionMethodsLen := int(data[pos])
	pos++
	if pos+compressionMethodsLen > len(data) {
		return ""
	}
	pos += compressionMethodsLen

	// Check if we have extensions
	if pos+2 > len(data) {
		return "" // No extensions
	}

	extensionsLength := int(data[pos])<<8 | int(data[pos+1])
	pos += 2

	// Parse extensions
	extensionsEnd := pos + extensionsLength
	if extensionsEnd > len(data) {
		// If extensions exceed data length, use available data
		extensionsEnd = len(data)
	}

	for pos < extensionsEnd-4 {
		// Parse extension header
		extType := int(data[pos])<<8 | int(data[pos+1])
		extLength := int(data[pos+2])<<8 | int(data[pos+3])
		pos += 4

		if pos+extLength > len(data) {
			break
		}

		// Check for Server Name Indication (type 0x0000)
		if extType == 0x0000 {
			// Parse SNI extension data
			if extLength < 2 {
				break
			}

			// ServerNameList length
			listLength := int(data[pos])<<8 | int(data[pos+1])
			pos += 2

			if listLength < 3 || pos+listLength > len(data) {
				break
			}

			// Parse ServerName entries
			listEnd := pos + listLength
			for pos < listEnd-3 {
				// ServerName entry
				nameType := data[pos]
				nameLength := int(data[pos+1])<<8 | int(data[pos+2])
				pos += 3

				if pos+nameLength > len(data) {
					break
				}

				// Check for host_name type (0x00)
				if nameType == 0x00 {
					return string(data[pos : pos+nameLength])
				}

				pos += nameLength
			}
			break
		}

		pos += extLength
	}

	return ""
}

func Pipe(ctx context.Context, src, dst net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := dst.Close(); err != nil {
			// Connection close errors are expected and can be safely ignored
			_ = err // explicitly ignore the error
		}
	}()

	buf := make([]byte, 4096) // BUFFER_SIZE is now in config package
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := src.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				_, err = dst.Write(buf[:n])
				if err != nil {
					return
				}
			}
		}
	}
}

func ConnectDirect(host string, port int, timeout int) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, err
	}
	
	// Set read/write deadlines
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return nil, err
	}
	
	return conn, nil
}

func ConnectViaProxy(proxyHost string, proxyPort int, targetHost string, targetPort int, clientIP string, timeout int) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(proxyHost, strconv.Itoa(proxyPort)), time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, err
	}
	
	// Set read/write deadlines
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return nil, err
	}
	
	connectRequest := fmt.Sprintf(
		"CONNECT %s:%d HTTP/1.1\r\n"+
			"Host: %s:%d\r\n"+
			"X-Forwarded-For: %s\r\n"+
			"Forwarded: for=%s\r\n"+
			"\r\n",
		targetHost, targetPort,
		targetHost, targetPort,
		clientIP, clientIP,
	)
	
	if _, err := conn.Write([]byte(connectRequest)); err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			// Connection close errors are expected and can be safely ignored
			_ = closeErr // explicitly ignore the error
		}
		return nil, err
	}
	
	// Read proxy response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			// Connection close errors are expected and can be safely ignored
			_ = closeErr // explicitly ignore the error
		}
		return nil, err
	}
	
	if !strings.HasPrefix(response, "HTTP/1.1 200") {
		if closeErr := conn.Close(); closeErr != nil {
			// Connection close errors are expected and can be safely ignored
			_ = closeErr // explicitly ignore the error
		}
		return nil, fmt.Errorf("proxy connection failed: %s", response)
	}
	
	// Read remaining headers until empty line
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}
	}
	
	return conn, nil
}
