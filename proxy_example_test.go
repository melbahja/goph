package goph_test

import (
	"testing"

	"github.com/melbahja/goph"
)

// Example demonstrating SOCKS5 proxy usage
func ExampleNewSOCKS5ProxyConn() {
	// This is just an example - won't actually connect without a real SSH server and proxy
	auth := goph.Password("password")

	// Connect to SSH server through SOCKS5 proxy
	client, err := goph.NewSOCKS5ProxyUnknown("user", "example.com", auth, "proxy.example.com", 1080, "", "")
	if err != nil {
		// handle error
		return
	}
	defer client.Close()

	// Use client for SSH operations
	_ = client
}

// Example demonstrating HTTP CONNECT proxy usage
func ExampleNewHTTPProxyConn() {
	// This is just an example - won't actually connect without a real SSH server and proxy
	auth := goph.Password("password")

	// Connect to SSH server through HTTP CONNECT proxy
	client, err := goph.NewHTTPProxyUnknown("user", "example.com", auth, "proxy.example.com", 8080, "proxyuser", "proxypass")
	if err != nil {
		// handle error
		return
	}
	defer client.Close()

	// Use client for SSH operations
	_ = client
}

// Example demonstrating manual proxy configuration
func ExampleProxyConfig() {
	// This is just an example - won't actually connect without a real SSH server and proxy
	config := &goph.Config{
		User:     "user",
		Addr:     "example.com",
		Port:     22,
		Auth:     goph.Password("password"),
		Callback: goph.InsecureIgnoreHostKey(),
		Proxy: &goph.ProxyConfig{
			Type:     goph.ProxyTypeSOCKS5,
			Addr:     "proxy.example.com",
			Port:     1080,
			User:     "", // optional
			Password: "", // optional
		},
	}

	client, err := goph.NewConn(config)
	if err != nil {
		// handle error
		return
	}
	defer client.Close()

	// Use client for SSH operations
	_ = client
}

// Test that proxy configuration compiles correctly
func TestProxyConfiguration(t *testing.T) {
	// Test SOCKS5 proxy config
	socks5Config := &goph.ProxyConfig{
		Type:     goph.ProxyTypeSOCKS5,
		Addr:     "127.0.0.1",
		Port:     1080,
		User:     "testuser",
		Password: "testpass",
	}

	if socks5Config.Type != goph.ProxyTypeSOCKS5 {
		t.Error("SOCKS5 proxy type not set correctly")
	}

	// Test HTTP proxy config
	httpConfig := &goph.ProxyConfig{
		Type:     goph.ProxyTypeHTTP,
		Addr:     "127.0.0.1",
		Port:     8080,
		User:     "testuser",
		Password: "testpass",
	}

	if httpConfig.Type != goph.ProxyTypeHTTP {
		t.Error("HTTP proxy type not set correctly")
	}

	// Test config with proxy
	config := &goph.Config{
		User:     "test",
		Addr:     "127.0.0.1",
		Port:     22,
		Auth:     goph.Password("test"),
		Callback: goph.InsecureIgnoreHostKey(),
		Proxy:    socks5Config,
	}

	if config.Proxy == nil {
		t.Error("Proxy configuration not set")
	}

	if config.Proxy.Type != goph.ProxyTypeSOCKS5 {
		t.Error("Proxy type not preserved in config")
	}
}
