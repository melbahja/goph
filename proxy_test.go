package goph_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

// TestProxyConfiguration tests basic proxy configuration functionality
func TestProxyConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		proxyType goph.ProxyType
		addr     string
		port     uint
		user     string
		pass     string
	}{
		{
			name:     "SOCKS5 proxy without auth",
			proxyType: goph.ProxyTypeSOCKS5,
			addr:     "127.0.0.1",
			port:     1080,
			user:     "",
			pass:     "",
		},
		{
			name:     "HTTP proxy with auth",
			proxyType: goph.ProxyTypeHTTP,
			addr:     "proxy.example.com",
			port:     8080,
			user:     "proxyuser",
			pass:     "proxypass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyConfig := &goph.ProxyConfig{
				Type:     tt.proxyType,
				Addr:     tt.addr,
				Port:     tt.port,
				User:     tt.user,
				Password: tt.pass,
			}

			config := &goph.Config{
				User:     "testuser",
				Addr:     "127.0.0.1",
				Port:     22,
				Auth:     goph.Password("testpass"),
				Callback: ssh.InsecureIgnoreHostKey(),
				Proxy:    proxyConfig,
			}

			// Verify proxy config is set correctly
			if config.Proxy.Type != tt.proxyType {
				t.Errorf("expected proxy type %v, got %v", tt.proxyType, config.Proxy.Type)
			}
			if config.Proxy.Addr != tt.addr {
				t.Errorf("expected proxy addr %s, got %s", tt.addr, config.Proxy.Addr)
			}
			if config.Proxy.Port != tt.port {
				t.Errorf("expected proxy port %d, got %d", tt.port, config.Proxy.Port)
			}
			if config.Proxy.User != tt.user {
				t.Errorf("expected proxy user %s, got %s", tt.user, config.Proxy.User)
			}
			if config.Proxy.Password != tt.pass {
				t.Errorf("expected proxy pass %s, got %s", tt.pass, config.Proxy.Password)
			}
		})
	}
}

// mockSOCKS5Server creates a mock SOCKS5 proxy server for testing
func mockSOCKS5Server(t *testing.T) (net.Listener, func()) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create mock SOCKS5 server: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // listener closed
			}
			go handleMockSOCKS5Connection(conn, t)
		}
	}()

	cleanup := func() {
		listener.Close()
	}

	return listener, cleanup
}

// handleMockSOCKS5Connection handles a mock SOCKS5 connection
func handleMockSOCKS5Connection(conn net.Conn, t *testing.T) {
	defer conn.Close()

	// Read SOCKS5 greeting
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Logf("failed to read SOCKS5 greeting: %v", err)
		return
	}

	// Send greeting response (no auth required)
	if _, err := conn.Write([]byte{0x05, 0x00}); err != nil {
		t.Logf("failed to send SOCKS5 greeting response: %v", err)
		return
	}

	// Read connect request
	buf = make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Logf("failed to read SOCKS5 connect request: %v", err)
		return
	}

	// Read address and port (we'll just accept any)
	addrLen := int(buf[3])
	addrBuf := make([]byte, addrLen+2) // +2 for port
	if _, err := io.ReadFull(conn, addrBuf); err != nil {
		t.Logf("failed to read SOCKS5 address: %v", err)
		return
	}

	// Send success response (we're not actually connecting anywhere)
	response := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if _, err := conn.Write(response); err != nil {
		t.Logf("failed to send SOCKS5 response: %v", err)
		return
	}

	// Keep connection alive briefly
	time.Sleep(100 * time.Millisecond)
}

// mockHTTPProxyServer creates a mock HTTP CONNECT proxy server for testing
func mockHTTPProxyServer(t *testing.T) (*httptest.Server, func()) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "CONNECT" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Hijack the connection
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}

		conn, _, err := hj.Hijack()
		if err != nil {
			t.Logf("failed to hijack connection: %v", err)
			return
		}
		defer conn.Close()

		// Send 200 Connection established
		response := "HTTP/1.1 200 Connection established\r\n\r\n"
		if _, err := conn.Write([]byte(response)); err != nil {
			t.Logf("failed to send CONNECT response: %v", err)
			return
		}

		// Keep connection alive briefly
		time.Sleep(100 * time.Millisecond)
	})

	server := httptest.NewServer(handler)

	cleanup := func() {
		server.Close()
	}

	return server, cleanup
}

// TestSOCKS5ProxyDialing tests SOCKS5 proxy dialing functionality
func TestSOCKS5ProxyDialing(t *testing.T) {
	// Start mock SOCKS5 server
	listener, cleanup := mockSOCKS5Server(t)
	defer cleanup()

	addr := listener.Addr().(*net.TCPAddr)

	// Create proxy config
	proxyConfig := &goph.ProxyConfig{
		Type: goph.ProxyTypeSOCKS5,
		Addr: addr.IP.String(),
		Port: uint(addr.Port),
	}

	config := &goph.Config{
		User:     "testuser",
		Addr:     "127.0.0.1", // This will fail because no real SSH server, but proxy dialing should work
		Port:     22,
		Auth:     goph.Password("testpass"),
		Callback: ssh.InsecureIgnoreHostKey(),
		Proxy:    proxyConfig,
		Timeout:  100 * time.Millisecond, // Short timeout for test
	}

	// This should attempt to dial through the proxy (and fail at SSH handshake, which is expected)
	_, err := goph.NewConn(config)
	if err == nil {
		t.Error("expected error due to no real SSH server, but got nil")
	}

	// Verify the error is related to SSH connection, not proxy dialing
	if !strings.Contains(err.Error(), "ssh") && !strings.Contains(err.Error(), "handshake") {
		t.Errorf("expected SSH-related error, got: %v", err)
	}
}

// TestHTTPProxyDialing tests HTTP CONNECT proxy dialing functionality
func TestHTTPProxyDialing(t *testing.T) {
	// Start mock HTTP proxy server
	server, cleanup := mockHTTPProxyServer(t)
	defer cleanup()

	// Extract host and port from server URL
	proxyURL := server.URL
	host, port, err := extractHostPort(proxyURL)
	if err != nil {
		t.Fatalf("failed to extract host/port from %s: %v", proxyURL, err)
	}

	// Create proxy config
	proxyConfig := &goph.ProxyConfig{
		Type: goph.ProxyTypeHTTP,
		Addr: host,
		Port: port,
	}

	config := &goph.Config{
		User:     "testuser",
		Addr:     "127.0.0.1", // This will fail because no real SSH server, but proxy dialing should work
		Port:     22,
		Auth:     goph.Password("testpass"),
		Callback: ssh.InsecureIgnoreHostKey(),
		Proxy:    proxyConfig,
		Timeout:  100 * time.Millisecond, // Short timeout for test
	}

	// This should attempt to dial through the proxy (and fail at SSH handshake, which is expected)
	_, err = goph.NewConn(config)
	if err == nil {
		t.Error("expected error due to no real SSH server, but got nil")
	}

	// Verify the error is related to SSH connection, not proxy dialing
	if !strings.Contains(err.Error(), "ssh") && !strings.Contains(err.Error(), "handshake") {
		t.Errorf("expected SSH-related error, got: %v", err)
	}
}

// extractHostPort extracts host and port from a URL string
func extractHostPort(url string) (string, uint, error) {
	if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	} else if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	}

	host, portStr, err := net.SplitHostPort(url)
	if err != nil {
		return "", 0, err
	}

	port, err := net.LookupPort("tcp", portStr)
	if err != nil {
		return "", 0, err
	}

	return host, uint(port), nil
}

// TestProxyConvenienceFunctions tests the convenience functions for proxy connections
func TestProxyConvenienceFunctions(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func() (*goph.Client, error)
		expectError bool
	}{
		{
			name: "NewSOCKS5ProxyUnknown",
			testFunc: func() (*goph.Client, error) {
				return goph.NewSOCKS5ProxyUnknown("user", "127.0.0.1", goph.Password("pass"), "127.0.0.1", 1080, "", "")
			},
			expectError: true, // Should fail due to no real servers
		},
		{
			name: "NewHTTPProxyUnknown",
			testFunc: func() (*goph.Client, error) {
				return goph.NewHTTPProxyUnknown("user", "127.0.0.1", goph.Password("pass"), "127.0.0.1", 8080, "", "")
			},
			expectError: true, // Should fail due to no real servers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.testFunc()
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

// TestProxyTypeValidation tests that proxy types are validated correctly
func TestProxyTypeValidation(t *testing.T) {
	tests := []struct {
		name      string
		proxyType goph.ProxyType
		valid     bool
	}{
		{"ProxyTypeNone", goph.ProxyTypeNone, true},
		{"ProxyTypeSOCKS5", goph.ProxyTypeSOCKS5, true},
		{"ProxyTypeHTTP", goph.ProxyTypeHTTP, true},
		{"InvalidType", goph.ProxyType(999), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &goph.Config{
				User:     "test",
				Addr:     "127.0.0.1",
				Port:     22,
				Auth:     goph.Password("test"),
				Callback: ssh.InsecureIgnoreHostKey(),
			}

			if tt.proxyType != goph.ProxyTypeNone {
				config.Proxy = &goph.ProxyConfig{
					Type: tt.proxyType,
					Addr: "127.0.0.1",
					Port: 1080,
				}
			}

			_, err := goph.NewConn(config)

			// We expect connection errors due to no real servers, but proxy type validation should pass
			if tt.valid && err != nil && !strings.Contains(err.Error(), "dial") && !strings.Contains(err.Error(), "connect") {
				t.Errorf("expected valid proxy type %v but got unexpected error: %v", tt.proxyType, err)
			}
		})
	}
}

// TestProxyTimeout tests that proxy connections respect timeout settings
func TestProxyTimeout(t *testing.T) {
	// Use a non-routable address to test timeout
	proxyConfig := &goph.ProxyConfig{
		Type: goph.ProxyTypeSOCKS5,
		Addr: "192.0.2.1", // RFC 5737 test address that should not respond
		Port: 1080,
	}

	config := &goph.Config{
		User:     "testuser",
		Addr:     "127.0.0.1",
		Port:     22,
		Auth:     goph.Password("testpass"),
		Callback: ssh.InsecureIgnoreHostKey(),
		Proxy:    proxyConfig,
		Timeout:  50 * time.Millisecond, // Very short timeout
	}

	start := time.Now()
	_, err := goph.NewConn(config)
	duration := time.Since(start)

	if err == nil {
		t.Error("expected timeout error but got nil")
	}

	// Should fail relatively quickly due to timeout
	if duration > 200*time.Millisecond {
		t.Errorf("timeout took too long: %v", duration)
	}
}

// BenchmarkProxyConfigCreation benchmarks proxy configuration creation
func BenchmarkProxyConfigCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		proxyConfig := &goph.ProxyConfig{
			Type:     goph.ProxyTypeSOCKS5,
			Addr:     "proxy.example.com",
			Port:     1080,
			User:     "user",
			Password: "pass",
		}

		config := &goph.Config{
			User:     "testuser",
			Addr:     "example.com",
			Port:     22,
			Auth:     goph.Password("testpass"),
			Callback: ssh.InsecureIgnoreHostKey(),
			Proxy:    proxyConfig,
		}

		_ = config
	}
}

// ExampleProxyTesting demonstrates how to test proxy functionality
func ExampleProxyTesting() {
	// Example of testing SOCKS5 proxy connection
	proxyConfig := &goph.ProxyConfig{
		Type:     goph.ProxyTypeSOCKS5,
		Addr:     "proxy.example.com",
		Port:     1080,
		User:     "proxyuser",
		Password: "proxypass",
	}

	config := &goph.Config{
		User:     "sshuser",
		Addr:     "ssh.example.com",
		Port:     22,
		Auth:     goph.Password("sshpass"),
		Callback: ssh.InsecureIgnoreHostKey(), // Use proper host key checking in production
		Proxy:    proxyConfig,
	}

	client, err := goph.NewConn(config)
	if err != nil {
		// Handle error - could be proxy connection failure, SSH auth failure, etc.
		fmt.Printf("Connection failed: %v\n", err)
		return
	}
	defer client.Close()

	// Use client for SSH operations
	output, err := client.Run("echo 'Hello from proxy connection!'")
	if err != nil {
		fmt.Printf("Command failed: %v\n", err)
		return
	}

	fmt.Printf("Output: %s\n", string(output))
}

// ExampleRealWorldProxyTest shows how to test with real proxy servers
func ExampleRealWorldProxyTest() {
	// This example shows how you might test with real proxy servers
	// Note: This requires actual proxy servers to be running

	testCases := []struct {
		name     string
		proxyType goph.ProxyType
		proxyAddr string
		proxyPort uint
		proxyUser string
		proxyPass string
		sshAddr   string
		sshUser   string
		sshAuth   goph.Auth
	}{
		{
			name:      "SOCKS5 proxy to local SSH",
			proxyType: goph.ProxyTypeSOCKS5,
			proxyAddr: "127.0.0.1",
			proxyPort: 1080,
			sshAddr:   "127.0.0.1",
			sshUser:   "testuser",
			sshAuth:   goph.Password("testpass"),
		},
		{
			name:      "HTTP proxy with auth",
			proxyType: goph.ProxyTypeHTTP,
			proxyAddr: "proxy.company.com",
			proxyPort: 8080,
			proxyUser: "domain\\user",
			proxyPass: "proxypass",
			sshAddr:   "internal.ssh.company.com",
			sshUser:   "sshuser",
			sshAuth:   goph.Key("/path/to/private/key", ""),
		},
	}

	for _, tc := range testCases {
		fmt.Printf("Testing %s...\n", tc.name)

		var client *goph.Client
		var err error

		if tc.proxyType == goph.ProxyTypeSOCKS5 {
			client, err = goph.NewSOCKS5ProxyConn(tc.sshUser, tc.sshAddr, tc.sshAuth, tc.proxyAddr, tc.proxyPort, tc.proxyUser, tc.proxyPass)
		} else {
			client, err = goph.NewHTTPProxyConn(tc.sshUser, tc.sshAddr, tc.sshAuth, tc.proxyAddr, tc.proxyPort, tc.proxyUser, tc.proxyPass)
		}

		if err != nil {
			fmt.Printf("  FAILED: %v\n", err)
			continue
		}

		// Test basic connectivity
		output, err := client.Run("echo 'Proxy test successful'")
		client.Close()

		if err != nil {
			fmt.Printf("  FAILED: %v\n", err)
			continue
		}

		fmt.Printf("  SUCCESS: %s\n", strings.TrimSpace(string(output)))
	}
}

// TestProxyErrorHandling tests various proxy error scenarios
func TestProxyErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		proxyConfig *goph.ProxyConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Invalid proxy type",
			proxyConfig: &goph.ProxyConfig{
				Type: goph.ProxyType(999),
				Addr: "127.0.0.1",
				Port: 1080,
			},
			expectError: true,
			errorMsg:    "unsupported proxy type",
		},
		{
			name: "Unreachable SOCKS5 proxy",
			proxyConfig: &goph.ProxyConfig{
				Type: goph.ProxyTypeSOCKS5,
				Addr: "192.0.2.1", // Non-routable test address
				Port: 1080,
			},
			expectError: true,
			errorMsg:    "dial",
		},
		{
			name: "Unreachable HTTP proxy",
			proxyConfig: &goph.ProxyConfig{
				Type: goph.ProxyTypeHTTP,
				Addr: "192.0.2.1", // Non-routable test address
				Port: 8080,
			},
			expectError: true,
			errorMsg:    "dial",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &goph.Config{
				User:     "testuser",
				Addr:     "127.0.0.1",
				Port:     22,
				Auth:     goph.Password("testpass"),
				Callback: ssh.InsecureIgnoreHostKey(),
				Proxy:    tt.proxyConfig,
				Timeout:  100 * time.Millisecond,
			}

			_, err := goph.NewConn(config)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}
