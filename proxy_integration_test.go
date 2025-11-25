package goph_test

import (
	"strings"
	"testing"
	"time"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

// TestSOCKS5ProxyIntegration performs end-to-end testing of SOCKS5 proxy functionality
func TestSOCKS5ProxyIntegration(t *testing.T) {
	helper := NewProxyTestHelper(t)
	defer helper.Close()

	tests := []struct {
		name      string
		auth      bool
		expectErr bool
	}{
		{"SOCKS5 without auth", false, true}, // Will fail at SSH level, but proxy should work
		{"SOCKS5 with auth", true, true},     // Will fail at SSH level, but proxy should work
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &goph.Config{
				User:     "testuser",
				Addr:     "127.0.0.1",
				Port:     22,
				Auth:     goph.Password("testpass"),
				Callback: ssh.InsecureIgnoreHostKey(),
				Timeout:  1 * time.Second,
			}

			if tt.auth {
				config.Proxy = &goph.ProxyConfig{
					Type:     goph.ProxyTypeSOCKS5,
					Addr:     helper.SOCKS5Server.Host(),
					Port:     helper.SOCKS5Server.Port(),
					User:     "testuser",
					Password: "testpass",
				}
			} else {
				config.Proxy = &goph.ProxyConfig{
					Type: goph.ProxyTypeSOCKS5,
					Addr: helper.SOCKS5Server.Host(),
					Port: helper.SOCKS5Server.Port(),
				}
			}

			err := helper.TestProxyConnection(t, goph.ProxyTypeSOCKS5, config)

			if tt.expectErr && err == nil {
				t.Error("expected error but got nil")
			} else if !tt.expectErr && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}

			// If we got an error, it should be SSH-related, not proxy-related
			if err != nil {
				AssertSSHError(t, err)
			}
		})
	}
}

// TestHTTPProxyIntegration performs end-to-end testing of HTTP CONNECT proxy functionality
func TestHTTPProxyIntegration(t *testing.T) {
	helper := NewProxyTestHelper(t)
	defer helper.Close()

	tests := []struct {
		name      string
		auth      bool
		expectErr bool
	}{
		{"HTTP proxy without auth", false, true}, // Will fail at SSH level, but proxy should work
		{"HTTP proxy with auth", true, true},     // Will fail at SSH level, but proxy should work
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &goph.Config{
				User:     "testuser",
				Addr:     "127.0.0.1",
				Port:     22,
				Auth:     goph.Password("testpass"),
				Callback: ssh.InsecureIgnoreHostKey(),
				Timeout:  1 * time.Second,
			}

			if tt.auth {
				config.Proxy = &goph.ProxyConfig{
					Type:     goph.ProxyTypeHTTP,
					Addr:     helper.HTTPServer.Host(),
					Port:     helper.HTTPServer.Port(),
					User:     "testuser",
					Password: "testpass",
				}
			} else {
				config.Proxy = &goph.ProxyConfig{
					Type: goph.ProxyTypeHTTP,
					Addr: helper.HTTPServer.Host(),
					Port: helper.HTTPServer.Port(),
				}
			}

			err := helper.TestProxyConnection(t, goph.ProxyTypeHTTP, config)

			if tt.expectErr && err == nil {
				t.Error("expected error but got nil")
			} else if !tt.expectErr && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}

			// If we got an error, it should be SSH-related, not proxy-related
			if err != nil {
				AssertSSHError(t, err)
			}
		})
	}
}

// TestProxyConvenienceFunctionsIntegration tests the convenience functions with mock servers
func TestProxyConvenienceFunctionsIntegration(t *testing.T) {
	helper := NewProxyTestHelper(t)
	defer helper.Close()

	tests := []struct {
		name     string
		testFunc func() (*goph.Client, error)
	}{
		{
			name: "NewSOCKS5ProxyUnknown",
			testFunc: func() (*goph.Client, error) {
				return goph.NewSOCKS5ProxyUnknown(
					"testuser",
					"127.0.0.1",
					goph.Password("testpass"),
					helper.SOCKS5Server.Host(),
					helper.SOCKS5Server.Port(),
					"testuser",
					"testpass",
				)
			},
		},
		{
			name: "NewHTTPProxyUnknown",
			testFunc: func() (*goph.Client, error) {
				return goph.NewHTTPProxyUnknown(
					"testuser",
					"127.0.0.1",
					goph.Password("testpass"),
					helper.HTTPServer.Host(),
					helper.HTTPServer.Port(),
					"testuser",
					"testpass",
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.testFunc()
			if err == nil {
				t.Error("expected connection error due to no SSH server, but got nil")
			} else {
				// Should be SSH-related error, not proxy error
				AssertSSHError(t, err)
			}
		})
	}
}

// TestProxyAuthentication tests proxy authentication scenarios
func TestProxyAuthentication(t *testing.T) {
	// Test with no auth required server but providing auth
	t.Run("SOCKS5 auth when not required", func(t *testing.T) {
		helper := NewProxyTestHelper(t)
		defer helper.Close()

		// Create server without auth requirement
		noAuthServer := NewMockSOCKS5Server(t, nil)
		defer noAuthServer.Close()

		config := &goph.Config{
			User:     "testuser",
			Addr:     "127.0.0.1",
			Port:     22,
			Auth:     goph.Password("testpass"),
			Callback: ssh.InsecureIgnoreHostKey(),
			Timeout:  1 * time.Second,
			Proxy: &goph.ProxyConfig{
				Type:     goph.ProxyTypeSOCKS5,
				Addr:     noAuthServer.Host(),
				Port:     noAuthServer.Port(),
				User:     "testuser", // Providing auth when not required
				Password: "testpass",
			},
		}

		err := helper.TestProxyConnection(t, goph.ProxyTypeSOCKS5, config)
		// Should fail at SSH level, not proxy auth level
		AssertSSHError(t, err)
	})

	// Test with auth required but not provided
	t.Run("SOCKS5 auth required but not provided", func(t *testing.T) {
		helper := NewProxyTestHelper(t)
		defer helper.Close()

		config := &goph.Config{
			User:     "testuser",
			Addr:     "127.0.0.1",
			Port:     22,
			Auth:     goph.Password("testpass"),
			Callback: ssh.InsecureIgnoreHostKey(),
			Timeout:  1 * time.Second,
			Proxy: &goph.ProxyConfig{
				Type: goph.ProxyTypeSOCKS5,
				Addr: helper.SOCKS5Server.Host(),
				Port: helper.SOCKS5Server.Port(),
				// No auth provided but server requires it
			},
		}

		err := helper.TestProxyConnection(t, goph.ProxyTypeSOCKS5, config)
		// Should fail at proxy level due to missing auth
		if err == nil || !strings.Contains(err.Error(), "dial") {
			t.Errorf("expected proxy authentication error, got: %v", err)
		}
	})

	// Test HTTP proxy with wrong credentials
	t.Run("HTTP proxy wrong credentials", func(t *testing.T) {
		helper := NewProxyTestHelper(t)
		defer helper.Close()

		config := &goph.Config{
			User:     "testuser",
			Addr:     "127.0.0.1",
			Port:     22,
			Auth:     goph.Password("testpass"),
			Callback: ssh.InsecureIgnoreHostKey(),
			Timeout:  1 * time.Second,
			Proxy: &goph.ProxyConfig{
				Type:     goph.ProxyTypeHTTP,
				Addr:     helper.HTTPServer.Host(),
				Port:     helper.HTTPServer.Port(),
				User:     "wronguser",
				Password: "wrongpass",
			},
		}

		err := helper.TestProxyConnection(t, goph.ProxyTypeHTTP, config)
		// Should fail at proxy auth level
		if err == nil || !strings.Contains(err.Error(), "dial") {
			t.Errorf("expected proxy authentication error, got: %v", err)
		}
	})
}

// TestProxyTimeoutBehavior tests how proxies behave with timeouts
func TestProxyTimeoutBehavior(t *testing.T) {
	helper := NewProxyTestHelper(t)
	defer helper.Close()

	t.Run("SOCKS5 with short timeout", func(t *testing.T) {
		config := &goph.Config{
			User:     "testuser",
			Addr:     "127.0.0.1",
			Port:     22,
			Auth:     goph.Password("testpass"),
			Callback: ssh.InsecureIgnoreHostKey(),
			Timeout:  50 * time.Millisecond, // Very short timeout
			Proxy: &goph.ProxyConfig{
				Type: goph.ProxyTypeSOCKS5,
				Addr: helper.SOCKS5Server.Host(),
				Port: helper.SOCKS5Server.Port(),
				User: "testuser",
				Password: "testpass",
			},
		}

		start := time.Now()
		_, err := goph.NewConn(config)
		duration := time.Since(start)

		// Should timeout relatively quickly
		if duration > 200*time.Millisecond {
			t.Errorf("proxy connection took too long: %v", duration)
		}

		// Should get some kind of error
		if err == nil {
			t.Error("expected timeout error but got nil")
		}
	})

	t.Run("HTTP proxy with short timeout", func(t *testing.T) {
		config := &goph.Config{
			User:     "testuser",
			Addr:     "127.0.0.1",
			Port:     22,
			Auth:     goph.Password("testpass"),
			Callback: ssh.InsecureIgnoreHostKey(),
			Timeout:  50 * time.Millisecond, // Very short timeout
			Proxy: &goph.ProxyConfig{
				Type: goph.ProxyTypeHTTP,
				Addr: helper.HTTPServer.Host(),
				Port: helper.HTTPServer.Port(),
				User: "testuser",
				Password: "testpass",
			},
		}

		start := time.Now()
		_, err := goph.NewConn(config)
		duration := time.Since(start)

		// Should timeout relatively quickly
		if duration > 200*time.Millisecond {
			t.Errorf("proxy connection took too long: %v", duration)
		}

		// Should get some kind of error
		if err == nil {
			t.Error("expected timeout error but got nil")
		}
	})
}

// TestProxyTypeSwitching tests switching between proxy types
func TestProxyTypeSwitching(t *testing.T) {
	helper := NewProxyTestHelper(t)
	defer helper.Close()

	// Test switching from SOCKS5 to HTTP
	t.Run("Switch from SOCKS5 to HTTP", func(t *testing.T) {
		config := &goph.Config{
			User:     "testuser",
			Addr:     "127.0.0.1",
			Port:     22,
			Auth:     goph.Password("testpass"),
			Callback: ssh.InsecureIgnoreHostKey(),
			Timeout:  1 * time.Second,
		}

		// First try SOCKS5
		config.Proxy = &goph.ProxyConfig{
			Type:     goph.ProxyTypeSOCKS5,
			Addr:     helper.SOCKS5Server.Host(),
			Port:     helper.SOCKS5Server.Port(),
			User:     "testuser",
			Password: "testpass",
		}

		err1 := helper.TestProxyConnection(t, goph.ProxyTypeSOCKS5, config)
		AssertSSHError(t, err1) // Should fail at SSH level

		// Then switch to HTTP
		config.Proxy = &goph.ProxyConfig{
			Type:     goph.ProxyTypeHTTP,
			Addr:     helper.HTTPServer.Host(),
			Port:     helper.HTTPServer.Port(),
			User:     "testuser",
			Password: "testpass",
		}

		err2 := helper.TestProxyConnection(t, goph.ProxyTypeHTTP, config)
		AssertSSHError(t, err2) // Should fail at SSH level

		// Both should have similar error patterns (SSH-related)
		if err1 == nil || err2 == nil {
			t.Error("expected both connections to fail at SSH level")
		}
	})
}

// TestConcurrentProxyConnections tests multiple proxy connections simultaneously
func TestConcurrentProxyConnections(t *testing.T) {
	helper := NewProxyTestHelper(t)
	defer helper.Close()

	numConnections := 5
	errors := make(chan error, numConnections)

	for i := 0; i < numConnections; i++ {
		go func(id int) {
			config := &goph.Config{
				User:     "testuser",
				Addr:     "127.0.0.1",
				Port:     22,
				Auth:     goph.Password("testpass"),
				Callback: ssh.InsecureIgnoreHostKey(),
				Timeout:  2 * time.Second,
				Proxy: &goph.ProxyConfig{
					Type:     goph.ProxyTypeSOCKS5,
					Addr:     helper.SOCKS5Server.Host(),
					Port:     helper.SOCKS5Server.Port(),
					User:     "testuser",
					Password: "testpass",
				},
			}

			_, err := goph.NewConn(config)
			errors <- err
		}(i)
	}

	// Collect all errors
	for i := 0; i < numConnections; i++ {
		err := <-errors
		if err == nil {
			t.Errorf("connection %d: expected error but got nil", i)
		} else {
			AssertSSHError(t, err)
		}
	}
}

// BenchmarkProxyConnectionCreation benchmarks proxy connection creation
func BenchmarkProxyConnectionCreation(b *testing.B) {
	helper := NewProxyTestHelper(b)
	defer helper.Close()

	config := &goph.Config{
		User:     "testuser",
		Addr:     "127.0.0.1",
		Port:     22,
		Auth:     goph.Password("testpass"),
		Callback: ssh.InsecureIgnoreHostKey(),
		Timeout:  100 * time.Millisecond,
		Proxy: &goph.ProxyConfig{
			Type:     goph.ProxyTypeSOCKS5,
			Addr:     helper.SOCKS5Server.Host(),
			Port:     helper.SOCKS5Server.Port(),
			User:     "testuser",
			Password: "testpass",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = goph.NewConn(config)
	}
}
