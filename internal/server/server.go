package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"syscall"
	"unsafe"

	"tproxy/internal/config"
	"tproxy/internal/proxy"
)

// Constants for SO_ORIGINAL_DST (Linux-specific)
const SO_ORIGINAL_DST = 80 // Typically 80 on Linux systems

// getOriginalDst gets the original destination using SO_ORIGINAL_DST (Linux only)
func getOriginalDst(conn net.Conn) (string, int, error) {
	// Get the underlying file descriptor
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return "", 0, fmt.Errorf("not a TCP connection")
	}

	file, err := tcpConn.File()
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	fd := int(file.Fd())

	// Use getsockopt to get SO_ORIGINAL_DST
	var addr [16]byte
	addrLen := uint32(16)

	_, _, errno := syscall.Syscall6(
		syscall.SYS_GETSOCKOPT,
		uintptr(fd),
		syscall.IPPROTO_IP,
		SO_ORIGINAL_DST,
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(unsafe.Pointer(&addrLen)),
		0,
	)

	if errno != 0 {
		return "", 0, fmt.Errorf("getsockopt failed: %v", errno)
	}

	// Parse the result (similar to Python version)
	// Format: 2 bytes padding, 2 bytes port, 4 bytes IP, 8 bytes padding
	port := int(addr[2])<<8 | int(addr[3])
	ip := fmt.Sprintf("%d.%d.%d.%d", addr[4], addr[5], addr[6], addr[7])

	return ip, port, nil
}

func handleHTTPSClient(conn net.Conn, rules []config.Rule) {
	defer conn.Close()

	clientIP := conn.RemoteAddr().String()
	originalIP := ""
	originalPort := config.DEFAULT_HTTPS_PORT

	// Try to get original destination using SO_ORIGINAL_DST
	ip, port, err := getOriginalDst(conn)
	if err == nil {
		originalIP = ip
		originalPort = port
	} else {
		// Fallback to RemoteAddr if SO_ORIGINAL_DST fails
		if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
			originalIP = tcpAddr.IP.String()
		}
	}

	// Read initial data to parse SNI
	buf := make([]byte, config.BUFFER_SIZE)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return
	}

	initialData := buf[:n]
	sni := proxy.ParseSNI(initialData)

	if sni == "" {
		fmt.Printf("SNI not found from %s -> %s:%d\n", clientIP, originalIP, originalPort)
		// Use original IP as fallback (similar to Python version)
		if originalIP != "" {
			sni = originalIP
		} else {
			return
		}
	}

	proxyAction, err := config.FindProxyForHost(sni, rules)
	if err != nil {
		fmt.Printf("Error finding proxy for %s: %v\n", sni, err)
		return
	}

	proxyConnection(sni, originalPort, originalIP, clientIP, conn, proxyAction, initialData, true)
}

func handleHTTPClient(conn net.Conn, rules []config.Rule) {
	defer conn.Close()

	clientIP := conn.RemoteAddr().String()
	originalIP := ""

	// Try to get original destination using SO_ORIGINAL_DST
	ip, _, err := getOriginalDst(conn)
	if err == nil {
		originalIP = ip
	} else {
		// Fallback to RemoteAddr if SO_ORIGINAL_DST fails
		if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
			originalIP = tcpAddr.IP.String()
		}
	}

	// Read initial data to parse Host header
	buf := make([]byte, config.BUFFER_SIZE)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return
	}

	initialData := buf[:n]
	host, port := proxy.ParseHTTPHost(initialData)

	if host == "" {
		fmt.Printf("Host header not found from %s\n", clientIP)
		return
	}

	proxyAction, err := config.FindProxyForHost(host, rules)
	if err != nil {
		fmt.Printf("Error finding proxy for %s: %v\n", host, err)
		return
	}

	proxyConnection(host, port, originalIP, clientIP, conn, proxyAction, initialData, false)
}

func proxyConnection(
	targetHost string,
	targetPort int,
	originalIP string,
	clientIP string,
	clientConn net.Conn,
	proxyAction *config.ProxyAction,
	initialData []byte,
	isHTTPS bool,
) {
	// If originalIP is not provided, try to extract it from client connection
	if originalIP == "" {
		if tcpAddr, ok := clientConn.RemoteAddr().(*net.TCPAddr); ok {
			originalIP = tcpAddr.IP.String()
		}
	}

	if proxyAction.Type == "DROP" {
		fmt.Printf("%s => %s:%d: Drop for %s:%d\n", clientIP, originalIP, targetPort, targetHost, targetPort)
		return
	}

	var remoteConn net.Conn
	var err error

	if proxyAction.Type == "PROXY" && proxyAction.Host != "" && proxyAction.Port != 0 {
		fmt.Printf("%s => %s:%d: Proxying connection for %s:%d via %s:%d\n",
			clientIP, originalIP, targetPort, targetHost, targetPort, proxyAction.Host, proxyAction.Port)

		remoteConn, err = proxy.ConnectViaProxy(proxyAction.Host, proxyAction.Port, targetHost, targetPort, clientIP)
	} else {
		fmt.Printf("%s => %s:%d: Direct connection for %s:%d\n", clientIP, originalIP, targetPort, targetHost, targetPort)
		remoteConn, err = proxy.ConnectDirect(targetHost, targetPort)
	}

	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		return
	}
	defer remoteConn.Close()

	// Send initial data if we have it
	if len(initialData) > 0 {
		if _, err := remoteConn.Write(initialData); err != nil {
			fmt.Printf("Failed to send initial data: %v\n", err)
			return
		}
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	// Pipe data between client and remote
	go proxy.Pipe(ctx, clientConn, remoteConn, &wg)
	go proxy.Pipe(ctx, remoteConn, clientConn, &wg)

	wg.Wait()
}

func StartServers(config *config.Config) error {
	listenConfig := config.Listen
	rules := config.Rules

	// Start HTTPS server
	httpsListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenConfig.Host, listenConfig.HTTPSPort))
	if err != nil {
		return fmt.Errorf("failed to start HTTPS server: %w", err)
	}
	defer httpsListener.Close()

	// Start HTTP server
	httpListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", listenConfig.Host, listenConfig.HTTPPort))
	if err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	defer httpListener.Close()

	fmt.Printf("SNI proxy (HTTPS) listening on %s:%d\n", listenConfig.Host, listenConfig.HTTPSPort)
	fmt.Printf("Host proxy (HTTP) listening on %s:%d\n", listenConfig.Host, listenConfig.HTTPPort)
	fmt.Println("Routing rules:")
	for i, rule := range rules {
		fmt.Printf("  %d. %s -> %s\n", i+1, rule.Pattern, rule.Proxy)
	}

	// Handle HTTPS connections
	go func() {
		for {
			conn, err := httpsListener.Accept()
			if err != nil {
				fmt.Printf("HTTPS accept error: %v\n", err)
				continue
			}
			go handleHTTPSClient(conn, rules)
		}
	}()

	// Handle HTTP connections
	for {
		conn, err := httpListener.Accept()
		if err != nil {
			fmt.Printf("HTTP accept error: %v\n", err)
			continue
		}
		go handleHTTPClient(conn, rules)
	}
}
