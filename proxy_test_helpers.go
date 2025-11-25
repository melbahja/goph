package goph_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

// TestProxySetup represents a complete proxy testing setup
type TestProxySetup struct {
	Helper       *ProxyTestHelper
	Config       *goph.Config
	ProxyConfigs map[goph.ProxyType]*goph.ProxyConfig
}

// NewTestProxySetup creates a complete proxy testing environment
func NewTestProxySetup(t *testing.T) *TestProxySetup {
	helper := NewProxyTestHelper(t)

	baseConfig := &goph.Config{
		User:     "testuser",
		Addr:     "127.0.0.1",
		Port:     22,
		Auth:     goph.Password("testpass"),
		Callback: ssh.InsecureIgnoreHostKey(),
		Timeout:  2 * time.Second,
	}

	proxyConfigs := map[goph.ProxyType]*goph.ProxyConfig{
		goph.ProxyTypeSOCKS5: {
			Type:     goph.ProxyTypeSOCKS5,
			Addr:     helper.SOCKS5Server.Host(),
			Port:     helper.SOCKS5Server.Port(),
			User:     "testuser",
			Password: "testpass",
		},
		goph.ProxyTypeHTTP: {
			Type:     goph.ProxyTypeHTTP,
			Addr:     helper.HTTPServer.Host(),
			Port:     helper.HTTPServer.Port(),
			User:     "testuser",
			Password: "testpass",
		},
	}

	return &TestProxySetup{
		Helper:       helper,
		Config:       baseConfig,
		ProxyConfigs: proxyConfigs,
	}
}

// Close cleans up the test setup
func (s *TestProxySetup) Close() {
	s.Helper.Close()
}

// TestProxyType tests a specific proxy type
func (s *TestProxySetup) TestProxyType(t *testing.T, proxyType goph.ProxyType) {
	t.Helper()

	config := *s.Config // Copy base config
	config.Proxy = s.ProxyConfigs[proxyType]

	err := s.Helper.TestProxyConnection(t, proxyType, &config)
	AssertSSHError(t, err)
}

// TestAllProxyTypes tests all available proxy types
func (s *TestProxySetup) TestAllProxyTypes(t *testing.T) {
	t.Helper()

	for proxyType := range s.ProxyConfigs {
		t.Run(fmt.Sprintf("ProxyType_%v", proxyType), func(t *testing.T) {
			s.TestProxyType(t, proxyType)
		})
	}
}

// TestProxyWithoutAuth tests proxy connections without authentication
func (s *TestProxySetup) TestProxyWithoutAuth(t *testing.T, proxyType goph.ProxyType) {
	t.Helper()

	// Create a server without auth requirements
	var server *MockProxyServer
	switch proxyType {
	case goph.ProxyTypeSOCKS5:
		server = NewMockSOCKS5Server(t, nil)
	case goph.ProxyTypeHTTP:
		server = NewMockHTTPProxyServer(t, nil)
	default:
		t.Fatalf("unsupported proxy type: %v", proxyType)
	}
	defer server.Close()

	config := *s.Config
	config.Proxy = &goph.ProxyConfig{
		Type: proxyType,
		Addr: server.Host(),
		Port: server.Port(),
		// No auth provided
	}

	err := s.Helper.TestProxyConnection(t, proxyType, &config)
	AssertSSHError(t, err)
}

// TestProxyWithWrongAuth tests proxy connections with wrong authentication
func (s *TestProxySetup) TestProxyWithWrongAuth(t *testing.T, proxyType goph.ProxyType) {
	t.Helper()

	config := *s.Config
	config.Proxy = &goph.ProxyConfig{
		Type:     proxyType,
		Addr:     s.ProxyConfigs[proxyType].Addr,
		Port:     s.ProxyConfigs[proxyType].Port,
		User:     "wronguser",
		Password: "wrongpass",
	}

	err := s.Helper.TestProxyConnection(t, proxyType, &config)
	if err == nil {
		t.Error("expected authentication error but got nil")
	}
	// Note: The exact error depends on proxy implementation
}

// ProxyTestRunner provides a fluent interface for running proxy tests
type ProxyTestRunner struct {
	setup *TestProxySetup
	t     *testing.T
}

// NewProxyTestRunner creates a new test runner
func NewProxyTestRunner(t *testing.T) *ProxyTestRunner {
	return &ProxyTestRunner{
		setup: NewTestProxySetup(t),
		t:     t,
	}
}

// Close cleans up the test runner
func (r *ProxyTestRunner) Close() {
	r.setup.Close()
}

// WithAuth configures the test to use authentication
func (r *ProxyTestRunner) WithAuth(username, password string) *ProxyTestRunner {
	for _, config := range r.setup.ProxyConfigs {
		config.User = username
		config.Password = password
	}
	return r
}

// WithTimeout sets the connection timeout
func (r *ProxyTestRunner) WithTimeout(timeout time.Duration) *ProxyTestRunner {
	r.setup.Config.Timeout = timeout
	return r
}

// TestSOCKS5 runs SOCKS5 proxy tests
func (r *ProxyTestRunner) TestSOCKS5() *ProxyTestRunner {
	r.t.Run("SOCKS5_Proxy", func(t *testing.T) {
		r.setup.TestProxyType(t, goph.ProxyTypeSOCKS5)
	})
	return r
}

// TestHTTP runs HTTP proxy tests
func (r *ProxyTestRunner) TestHTTP() *ProxyTestRunner {
	r.t.Run("HTTP_Proxy", func(t *testing.T) {
		r.setup.TestProxyType(t, goph.ProxyTypeHTTP)
	})
	return r
}

// TestAll runs tests for all proxy types
func (r *ProxyTestRunner) TestAll() *ProxyTestRunner {
	r.t.Run("All_Proxy_Types", func(t *testing.T) {
		r.setup.TestAllProxyTypes(t)
	})
	return r
}

// TestWithoutAuth runs tests without proxy authentication
func (r *ProxyTestRunner) TestWithoutAuth(proxyType goph.ProxyType) *ProxyTestRunner {
	r.t.Run("No_Auth", func(t *testing.T) {
		r.setup.TestProxyWithoutAuth(t, proxyType)
	})
	return r
}

// TestWrongAuth runs tests with wrong proxy authentication
func (r *ProxyTestRunner) TestWrongAuth(proxyType goph.ProxyType) *ProxyTestRunner {
	r.t.Run("Wrong_Auth", func(t *testing.T) {
		r.setup.TestProxyWithWrongAuth(t, proxyType)
	})
	return r
}

// TestConvenienceFunctions tests the convenience constructor functions
func (r *ProxyTestRunner) TestConvenienceFunctions() *ProxyTestRunner {
	r.t.Run("Convenience_Functions", func(t *testing.T) {
		// Test SOCKS5 convenience function
		_, err := goph.NewSOCKS5ProxyUnknown(
			"testuser",
			"127.0.0.1",
			goph.Password("testpass"),
			r.setup.Helper.SOCKS5Server.Host(),
			r.setup.Helper.SOCKS5Server.Port(),
			"testuser",
			"testpass",
		)
		if err == nil {
			t.Error("expected error due to no SSH server")
		} else {
			AssertSSHError(t, err)
		}

		// Test HTTP convenience function
		_, err = goph.NewHTTPProxyUnknown(
			"testuser",
			"127.0.0.1",
			goph.Password("testpass"),
			r.setup.Helper.HTTPServer.Host(),
			r.setup.Helper.HTTPServer.Port(),
			"testuser",
			"testpass",
		)
		if err == nil {
			t.Error("expected error due to no SSH server")
		} else {
			AssertSSHError(t, err)
		}
	})
	return r
}

// Example usage of the proxy testing framework
func ExampleProxyTestingFramework() {
	// This example demonstrates how to use the proxy testing framework

	// Create a test runner (would be done in a test function)
	// runner := NewProxyTestRunner(t)
	// defer runner.Close()

	// Configure authentication
	// runner.WithAuth("testuser", "testpass").WithTimeout(5 * time.Second)

	// Run various tests
	// runner.TestSOCKS5().TestHTTP().TestAll()

	// Test edge cases
	// runner.TestWithoutAuth(goph.ProxyTypeSOCKS5)
	// runner.TestWrongAuth(goph.ProxyTypeHTTP)

	// Test convenience functions
	// runner.TestConvenienceFunctions()
}

// TestHelperFunctions demonstrates usage of the testing helpers
func TestHelperFunctions(t *testing.T) {
	runner := NewProxyTestRunner(t)
	defer runner.Close()

	// Test with authentication
	runner.WithAuth("testuser", "testpass").
		WithTimeout(3 * time.Second).
		TestSOCKS5().
		TestHTTP()

	// Test convenience functions
	runner.TestConvenienceFunctions()
}

// TestNetworkConnectivity tests basic network connectivity to proxy servers
func TestNetworkConnectivity(t *testing.T) {
	helper := NewProxyTestHelper(t)
	defer helper.Close()

	// Test SOCKS5 server connectivity
	t.Run("SOCKS5_Server_Connectivity", func(t *testing.T) {
		conn, err := net.DialTimeout("tcp",
			net.JoinHostPort(helper.SOCKS5Server.Host(), fmt.Sprintf("%d", helper.SOCKS5Server.Port())),
			1*time.Second)
		if err != nil {
			t.Fatalf("failed to connect to SOCKS5 server: %v", err)
		}
		conn.Close()
	})

	// Test HTTP server connectivity
	t.Run("HTTP_Server_Connectivity", func(t *testing.T) {
		conn, err := net.DialTimeout("tcp",
			net.JoinHostPort(helper.HTTPServer.Host(), fmt.Sprintf("%d", helper.HTTPServer.Port())),
			1*time.Second)
		if err != nil {
			t.Fatalf("failed to connect to HTTP server: %v", err)
		}
		conn.Close()
	})
}

// TestProxyConfigValidation tests configuration validation
func TestProxyConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *goph.ProxyConfig
		valid   bool
		errorContains string
	}{
		{
			name: "Valid SOCKS5 config",
			config: &goph.ProxyConfig{
				Type: goph.ProxyTypeSOCKS5,
				Addr: "127.0.0.1",
				Port: 1080,
			},
			valid: true,
		},
		{
			name: "Valid HTTP config with auth",
			config: &goph.ProxyConfig{
				Type: goph.ProxyTypeHTTP,
				Addr: "proxy.example.com",
				Port: 8080,
				User: "user",
				Password: "pass",
			},
			valid: true,
		},
		{
			name: "Invalid proxy type",
			config: &goph.ProxyConfig{
				Type: goph.ProxyType(999),
				Addr: "127.0.0.1",
				Port: 1080,
			},
			valid: false,
			errorContains: "unsupported proxy type",
		},
		{
			name: "Empty address",
			config: &goph.ProxyConfig{
				Type: goph.ProxyTypeSOCKS5,
				Addr: "",
				Port: 1080,
			},
			valid: false,
		},
		{
			name: "Zero port",
			config: &goph.ProxyConfig{
				Type: goph.ProxyTypeSOCKS5,
				Addr: "127.0.0.1",
				Port: 0,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a config with the proxy
			config := &goph.Config{
				User:     "test",
				Addr:     "127.0.0.1",
				Port:     22,
				Auth:     goph.Password("test"),
				Callback: ssh.InsecureIgnoreHostKey(),
				Proxy:    tt.config,
			}

			_, err := goph.NewConn(config)

			if tt.valid {
				// For valid configs, we expect SSH-related errors (since no real SSH server)
				// but not proxy configuration errors
				if err != nil && !isSSHError(err) && !isNetworkError(err) {
					t.Errorf("expected valid config but got unexpected error: %v", err)
				}
			} else {
				// For invalid configs, we expect errors
				if err == nil {
					t.Error("expected error for invalid config but got nil")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

// Helper functions for error checking
func isSSHError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsString(errStr, "ssh") || containsString(errStr, "handshake")
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsString(errStr, "connection") || containsString(errStr, "dial")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsString(s[1:], substr))
}
