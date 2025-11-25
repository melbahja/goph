package goph_test

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/melbahja/goph"
)

// MockProxyServer provides a configurable mock proxy server for testing
type MockProxyServer struct {
	Listener net.Listener
	Type     goph.ProxyType
	Auth     *ProxyAuth
	server   interface{} // Either *http.Server or custom TCP server
}

// ProxyAuth represents proxy authentication credentials
type ProxyAuth struct {
	Username string
	Password string
}

// NewMockSOCKS5Server creates a mock SOCKS5 proxy server
func NewMockSOCKS5Server(t *testing.T, auth *ProxyAuth) *MockProxyServer {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create SOCKS5 server: %v", err)
	}

	server := &MockProxyServer{
		Listener: listener,
		Type:     goph.ProxyTypeSOCKS5,
		Auth:     auth,
	}

	go server.handleSOCKS5Connections(t)

	return server
}

// NewMockHTTPProxyServer creates a mock HTTP CONNECT proxy server
func NewMockHTTPProxyServer(t *testing.T, auth *ProxyAuth) *MockProxyServer {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create HTTP proxy server: %v", err)
	}

	server := &MockProxyServer{
		Listener: listener,
		Type:     goph.ProxyTypeHTTP,
		Auth:     auth,
	}

	go server.handleHTTPConnections(t)

	return server
}

// Address returns the server address as "host:port"
func (s *MockProxyServer) Address() string {
	return s.Listener.Addr().String()
}

// Host returns just the host part
func (s *MockProxyServer) Host() string {
	return s.Listener.Addr().(*net.TCPAddr).IP.String()
}

// Port returns the port number
func (s *MockProxyServer) Port() uint {
	return uint(s.Listener.Addr().(*net.TCPAddr).Port)
}

// Close shuts down the mock server
func (s *MockProxyServer) Close() error {
	return s.Listener.Close()
}

// handleSOCKS5Connections handles incoming SOCKS5 connections
func (s *MockProxyServer) handleSOCKS5Connections(t *testing.T) {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			return // Server closed
		}
		go s.handleSOCKS5Connection(conn, t)
	}
}

// handleSOCKS5Connection handles a single SOCKS5 connection
func (s *MockProxyServer) handleSOCKS5Connection(conn net.Conn, t *testing.T) {
	defer conn.Close()

	// Read SOCKS5 greeting
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Logf("failed to read SOCKS5 greeting: %v", err)
		return
	}

	version, nmethods := buf[0], buf[1]
	if version != 0x05 {
		t.Logf("invalid SOCKS5 version: %d", version)
		return
	}

	// Read methods
	methods := make([]byte, nmethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		t.Logf("failed to read SOCKS5 methods: %v", err)
		return
	}

	var response byte = 0x00 // No auth required
	if s.Auth != nil {
		// Check if username/password auth is supported
		hasUserPass := false
		for _, method := range methods {
			if method == 0x02 { // Username/password auth
				hasUserPass = true
				response = 0x02
				break
			}
		}
		if !hasUserPass {
			response = 0xFF // No acceptable methods
		}
	}

	// Send method selection response
	if _, err := conn.Write([]byte{0x05, response}); err != nil {
		t.Logf("failed to send SOCKS5 method response: %v", err)
		return
	}

	if response == 0xFF {
		return // No acceptable auth methods
	}

	// Handle authentication if required
	if s.Auth != nil && response == 0x02 {
		if !s.handleSOCKS5Auth(conn, t) {
			return
		}
	}

	// Read connect request
	buf = make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Logf("failed to read SOCKS5 connect request: %v", err)
		return
	}

	if buf[0] != 0x05 || buf[1] != 0x01 {
		t.Logf("invalid SOCKS5 connect request")
		return
	}

	// Read address
	var addr string
	switch buf[3] {
	case 0x01: // IPv4
		addrBuf := make([]byte, 6) // 4 bytes IP + 2 bytes port
		if _, err := io.ReadFull(conn, addrBuf); err != nil {
			t.Logf("failed to read IPv4 address: %v", err)
			return
		}
		addr = fmt.Sprintf("%d.%d.%d.%d:%d",
			addrBuf[0], addrBuf[1], addrBuf[2], addrBuf[3],
			int(addrBuf[4])<<8|int(addrBuf[5]))
	case 0x03: // Domain name
		addrLenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, addrLenBuf); err != nil {
			t.Logf("failed to read domain length: %v", err)
			return
		}
		addrLen := int(addrLenBuf[0])
		addrBuf := make([]byte, addrLen+2)
		if _, err := io.ReadFull(conn, addrBuf); err != nil {
			t.Logf("failed to read domain address: %v", err)
			return
		}
		addr = fmt.Sprintf("%s:%d", string(addrBuf[:addrLen]), int(addrBuf[addrLen])<<8|int(addrBuf[addrLen+1]))
	default:
		t.Logf("unsupported address type: %d", buf[3])
		return
	}

	t.Logf("SOCKS5 proxy request for: %s", addr)

	// Send success response (we don't actually connect anywhere)
	response = []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if _, err := conn.Write(response); err != nil {
		t.Logf("failed to send SOCKS5 connect response: %v", err)
		return
	}

	// Keep connection alive briefly to simulate successful proxy
	time.Sleep(200 * time.Millisecond)
}

// handleSOCKS5Auth handles SOCKS5 username/password authentication
func (s *MockProxyServer) handleSOCKS5Auth(conn net.Conn, t *testing.T) bool {
	// Read auth request
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Logf("failed to read auth version: %v", err)
		return false
	}

	if buf[0] != 0x01 {
		t.Logf("invalid auth version: %d", buf[0])
		return false
	}

	userLen := int(buf[1])
	if userLen == 0 {
		t.Logf("empty username")
		return false
	}

	userBuf := make([]byte, userLen)
	if _, err := io.ReadFull(conn, userBuf); err != nil {
		t.Logf("failed to read username: %v", err)
		return false
	}

	passLenBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, passLenBuf); err != nil {
		t.Logf("failed to read password length: %v", err)
		return false
	}

	passLen := int(passLenBuf[0])
	passBuf := make([]byte, passLen)
	if _, err := io.ReadFull(conn, passBuf); err != nil {
		t.Logf("failed to read password: %v", err)
		return false
	}

	username := string(userBuf)
	password := string(passBuf)

	// Check credentials
	status := byte(0x00) // Success
	if username != s.Auth.Username || password != s.Auth.Password {
		status = 0x01 // Failure
		t.Logf("auth failed for user: %s", username)
	}

	// Send auth response
	if _, err := conn.Write([]byte{0x01, status}); err != nil {
		t.Logf("failed to send auth response: %v", err)
		return false
	}

	return status == 0x00
}

// handleHTTPConnections handles incoming HTTP CONNECT connections
func (s *MockProxyServer) handleHTTPConnections(t *testing.T) {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			return // Server closed
		}
		go s.handleHTTPConnection(conn, t)
	}
}

// handleHTTPConnection handles a single HTTP CONNECT connection
func (s *MockProxyServer) handleHTTPConnection(conn net.Conn, t *testing.T) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read CONNECT request
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		t.Logf("failed to read CONNECT request: %v", err)
		return
	}

	parts := strings.Fields(strings.TrimSpace(requestLine))
	if len(parts) != 3 || parts[0] != "CONNECT" {
		t.Logf("invalid CONNECT request: %s", requestLine)
		return
	}

	target := parts[1]

	// Read headers
	authHeader := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Logf("failed to read header: %v", err)
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break // End of headers
		}

		if strings.HasPrefix(strings.ToLower(line), "proxy-authorization:") {
			authHeader = strings.TrimSpace(line[len("proxy-authorization:"):])
		}
	}

	// Check authentication if required
	if s.Auth != nil {
		if !s.checkHTTPAuth(authHeader, t) {
			// Send 407 Proxy Authentication Required
			response := "HTTP/1.1 407 Proxy Authentication Required\r\n"
			response += "Proxy-Authenticate: Basic realm=\"Proxy\"\r\n"
			response += "\r\n"
			conn.Write([]byte(response))
			return
		}
	}

	t.Logf("HTTP proxy CONNECT request for: %s", target)

	// Send 200 Connection established
	response := "HTTP/1.1 200 Connection established\r\n\r\n"
	if _, err := conn.Write([]byte(response)); err != nil {
		t.Logf("failed to send CONNECT response: %v", err)
		return
	}

	// Keep connection alive briefly to simulate successful proxy
	time.Sleep(200 * time.Millisecond)
}

// checkHTTPAuth validates HTTP Basic authentication
func (s *MockProxyServer) checkHTTPAuth(authHeader string, t *testing.T) bool {
	if authHeader == "" {
		return false
	}

	if !strings.HasPrefix(strings.ToLower(authHeader), "basic ") {
		return false
	}

	encoded := strings.TrimSpace(authHeader[6:])
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Logf("failed to decode auth: %v", err)
		return false
	}

	credentials := string(decoded)
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return false
	}

	username, password := parts[0], parts[1]
	return username == s.Auth.Username && password == s.Auth.Password
}

// ProxyTestHelper provides utilities for testing proxy connections
type ProxyTestHelper struct {
	SOCKS5Server *MockProxyServer
	HTTPServer    *MockProxyServer
}

// NewProxyTestHelper creates a new test helper with mock servers
func NewProxyTestHelper(t *testing.T) *ProxyTestHelper {
	return &ProxyTestHelper{
		SOCKS5Server: NewMockSOCKS5Server(t, &ProxyAuth{Username: "testuser", Password: "testpass"}),
		HTTPServer:    NewMockHTTPProxyServer(t, &ProxyAuth{Username: "testuser", Password: "testpass"}),
	}
}

// Close shuts down all mock servers
func (h *ProxyTestHelper) Close() {
	if h.SOCKS5Server != nil {
		h.SOCKS5Server.Close()
	}
	if h.HTTPServer != nil {
		h.HTTPServer.Close()
	}
}

// TestProxyConnection tests a proxy connection with the given configuration
func (h *ProxyTestHelper) TestProxyConnection(t *testing.T, proxyType goph.ProxyType, config *goph.Config) error {
	// Set proxy configuration based on type
	switch proxyType {
	case goph.ProxyTypeSOCKS5:
		config.Proxy = &goph.ProxyConfig{
			Type:     goph.ProxyTypeSOCKS5,
			Addr:     h.SOCKS5Server.Host(),
			Port:     h.SOCKS5Server.Port(),
			User:     "testuser",
			Password: "testpass",
		}
	case goph.ProxyTypeHTTP:
		config.Proxy = &goph.ProxyConfig{
			Type:     goph.ProxyTypeHTTP,
			Addr:     h.HTTPServer.Host(),
			Port:     h.HTTPServer.Port(),
			User:     "testuser",
			Password: "testpass",
		}
	}

	// Set short timeout for testing
	config.Timeout = 500 * time.Millisecond

	// Attempt connection (will fail at SSH level but proxy should work)
	_, err := goph.NewConn(config)
	return err
}

// AssertProxyError asserts that an error is proxy-related (not SSH-related)
func AssertProxyError(t *testing.T, err error, expectedProxyType string) {
	if err == nil {
		t.Error("expected proxy error but got nil")
		return
	}

	errorStr := err.Error()

	// Should contain proxy-related error messages, not SSH handshake errors
	switch expectedProxyType {
	case "socks5":
		if strings.Contains(errorStr, "ssh") || strings.Contains(errorStr, "handshake") {
			t.Errorf("expected SOCKS5 proxy error, but got SSH error: %v", err)
		}
	case "http":
		if strings.Contains(errorStr, "ssh") || strings.Contains(errorStr, "handshake") {
			t.Errorf("expected HTTP proxy error, but got SSH error: %v", err)
		}
	}
}

// AssertSSHError asserts that an error is SSH-related (not proxy-related)
func AssertSSHError(t *testing.T, err error) {
	if err == nil {
		t.Error("expected SSH error but got nil")
		return
	}

	errorStr := err.Error()

	// Should contain SSH-related error messages
	if !strings.Contains(errorStr, "ssh") && !strings.Contains(errorStr, "handshake") &&
	   !strings.Contains(errorStr, "connection") {
		t.Errorf("expected SSH-related error, got: %v", err)
	}
}
