package main

import (
	"fmt"
	"log"

	"github.com/melbahja/goph"
)

// Demo script showing TCP proxy functionality
func main() {
	fmt.Println("=== Goph TCP Proxy Support Demo ===")
	fmt.Println()

	// Demo 1: SOCKS5 Proxy Configuration
	fmt.Println("1. SOCKS5 Proxy Configuration:")
	socks5Config := &goph.ProxyConfig{
		Type:     goph.ProxyTypeSOCKS5,
		Addr:     "proxy.example.com",
		Port:     1080,
		User:     "proxyuser",
		Password: "proxypass",
	}
	fmt.Printf("   Type: SOCKS5\n")
	fmt.Printf("   Address: %s:%d\n", socks5Config.Addr, socks5Config.Port)
	fmt.Printf("   Auth: %s/%s\n", socks5Config.User, socks5Config.Password)
	fmt.Println()

	// Demo 2: HTTP Proxy Configuration
	fmt.Println("2. HTTP CONNECT Proxy Configuration:")
	httpConfig := &goph.ProxyConfig{
		Type:     goph.ProxyTypeHTTP,
		Addr:     "http-proxy.company.com",
		Port:     8080,
		User:     "domain\\user",
		Password: "securepass",
	}
	fmt.Printf("   Type: HTTP CONNECT\n")
	fmt.Printf("   Address: %s:%d\n", httpConfig.Addr, httpConfig.Port)
	fmt.Printf("   Auth: %s/%s\n", httpConfig.User, httpConfig.Password)
	fmt.Println()

	// Demo 3: Full Client Configuration with Proxy
	fmt.Println("3. Complete SSH Client Configuration with Proxy:")
	config := &goph.Config{
		User:     "sshuser",
		Addr:     "internal.server.company.com",
		Port:     22,
		Auth:     goph.Password("sshpassword"),
		Callback: goph.DefaultKnownHosts(),
		Proxy:    socks5Config,
	}
	fmt.Printf("   SSH User: %s\n", config.User)
	fmt.Printf("   SSH Host: %s:%d\n", config.Addr, config.Port)
	fmt.Printf("   Proxy: %s via %s:%d\n",
		getProxyTypeName(config.Proxy.Type),
		config.Proxy.Addr,
		config.Proxy.Port)
	fmt.Println()

	// Demo 4: Convenience Functions
	fmt.Println("4. Convenience Functions for Proxy Connections:")
	fmt.Println("   // SOCKS5 proxy connection")
	fmt.Println("   client, err := goph.NewSOCKS5ProxyConn(")
	fmt.Println("       \"sshuser\", \"ssh.example.com\", goph.Password(\"pass\"),")
	fmt.Println("       \"proxy.example.com\", 1080, \"proxyuser\", \"proxypass\")")
	fmt.Println()
	fmt.Println("   // HTTP proxy connection")
	fmt.Println("   client, err := goph.NewHTTPProxyConn(")
	fmt.Println("       \"sshuser\", \"ssh.example.com\", goph.Password(\"pass\"),")
	fmt.Println("       \"proxy.example.com\", 8080, \"proxyuser\", \"proxypass\")")
	fmt.Println()

	// Demo 5: Error Handling
	fmt.Println("5. Error Handling:")
	fmt.Println("   // The connection will attempt to go through the proxy first")
	fmt.Println("   // If proxy connection fails, you'll get a network-related error")
	fmt.Println("   // If proxy succeeds but SSH fails, you'll get SSH handshake errors")
	fmt.Println("   // This allows you to distinguish proxy vs SSH authentication issues")
	fmt.Println()

	// Demo 6: Testing Approach
	fmt.Println("6. Testing Strategy:")
	fmt.Println("   ✅ Unit tests for configuration structures")
	fmt.Println("   ✅ Mock proxy servers for testing")
	fmt.Println("   ✅ Integration tests with real proxy protocols")
	fmt.Println("   ✅ Authentication testing scenarios")
	fmt.Println("   ✅ Error handling validation")
	fmt.Println("   ✅ Concurrent connection testing")
	fmt.Println()

	// Demo 7: Real Usage Examples
	fmt.Println("7. Real-World Usage Examples:")
	fmt.Println()
	fmt.Println("   // Corporate environment with HTTP proxy")
	fmt.Println("   client, err := goph.NewHTTPProxyConn(")
	fmt.Println("       \"john.doe\", \"git.company.com\", goph.Key(\"/home/user/.ssh/id_rsa\"),")
	fmt.Println("       \"proxy.company.com\", 8080, \"COMPANY\\\\john.doe\", \"corporate_password\")")
	fmt.Println()
	fmt.Println("   // SOCKS5 proxy for development")
	fmt.Println("   client, err := goph.NewSOCKS5ProxyConn(")
	fmt.Println("       \"dev\", \"dev-server.internal\", goph.Password(\"devpass\"),")
	fmt.Println("       \"127.0.0.1\", 1080, \"\", \"\") // No auth for local proxy")
	fmt.Println()
	fmt.Println("   // Manual configuration for complex scenarios")
	fmt.Println("   config := &goph.Config{")
	fmt.Println("       User: \"user\",")
	fmt.Println("       Addr: \"server.com\",")
	fmt.Println("       Auth: goph.Key(\"/path/to/key\"),")
	fmt.Println("       Proxy: &goph.ProxyConfig{")
	fmt.Println("           Type: goph.ProxyTypeSOCKS5,")
	fmt.Println("           Addr: \"proxy.host.com\",")
	fmt.Println("           Port: 1080,")
	fmt.Println("       },")
	fmt.Println("   }")
	fmt.Println("   client, err := goph.NewConn(config)")
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
	fmt.Println("The TCP proxy functionality has been successfully implemented!")
	fmt.Println("See PROXY_TESTING_README.md for comprehensive testing information.")
}

func getProxyTypeName(proxyType goph.ProxyType) string {
	switch proxyType {
	case goph.ProxyTypeSOCKS5:
		return "SOCKS5"
	case goph.ProxyTypeHTTP:
		return "HTTP CONNECT"
	default:
		return "Unknown"
	}
}
