# Goph Proxy Testing Guide

This document provides comprehensive guidance for testing the TCP proxy functionality in the goph SSH library.

## Overview

The proxy testing suite includes:

- **Unit tests** for configuration and basic functionality
- **Integration tests** with mock proxy servers
- **End-to-end tests** demonstrating real-world usage
- **Test utilities** for easy proxy testing setup

## Test Files

### Core Test Files

1. **`proxy_test.go`** - Comprehensive unit tests and examples
2. **`proxy_test_utils.go`** - Mock proxy server implementations
3. **`proxy_integration_test.go`** - Full integration tests
4. **`proxy_test_helpers.go`** - Testing utilities and fluent API

### Modified Files

- **`client.go`** - Added proxy support
- **`goph_test.go`** - Added basic proxy config tests
- **`README.md`** - Updated with proxy usage examples
- **`examples/goph/main.go`** - Added proxy command-line options

## Running Tests

### Run All Proxy Tests

```bash
go test -v -run=".*[Pp]roxy.*"
```

### Run Specific Test Categories

```bash
# Unit tests
go test -v -run="TestProxyConfiguration"

# Integration tests
go test -v -run="TestSOCKS5ProxyIntegration"

# All proxy tests
go test -v ./... -run=".*[Pp]roxy.*"
```

### Run Benchmarks

```bash
go test -bench=".*[Pp]roxy.*" -benchmem
```

## Test Structure

### Unit Tests (`proxy_test.go`)

- **TestProxyConfiguration** - Validates proxy config structures
- **TestProxyTypeValidation** - Tests proxy type validation
- **TestProxyErrorHandling** - Tests various error scenarios
- **TestProxyTimeoutBehavior** - Tests timeout handling

### Mock Servers (`proxy_test_utils.go`)

- **MockProxyServer** - Configurable mock proxy server
- **NewMockSOCKS5Server** - SOCKS5 proxy mock
- **NewMockHTTPProxyServer** - HTTP CONNECT proxy mock
- **ProxyTestHelper** - Test helper utilities

### Integration Tests (`proxy_integration_test.go`)

- **TestSOCKS5ProxyIntegration** - End-to-end SOCKS5 testing
- **TestHTTPProxyIntegration** - End-to-end HTTP proxy testing
- **TestProxyAuthentication** - Authentication testing
- **TestConcurrentProxyConnections** - Concurrent connection testing

### Test Helpers (`proxy_test_helpers.go`)

- **TestProxySetup** - Complete testing environment
- **ProxyTestRunner** - Fluent testing API
- **TestHelperFunctions** - Utility functions

## Test Examples

### Basic Proxy Testing

```go
func TestMyProxyFunctionality(t *testing.T) {
    // Create test environment
    setup := NewTestProxySetup(t)
    defer setup.Close()

    // Test SOCKS5 proxy
    setup.TestProxyType(t, goph.ProxyTypeSOCKS5)

    // Test HTTP proxy
    setup.TestProxyType(t, goph.ProxyTypeHTTP)
}
```

### Using Test Runner (Fluent API)

```go
func TestProxyWithRunner(t *testing.T) {
    runner := NewProxyTestRunner(t)
    defer runner.Close()

    // Configure and run tests
    runner.WithAuth("user", "pass").
        WithTimeout(5 * time.Second).
        TestSOCKS5().
        TestHTTP().
        TestConvenienceFunctions()
}
```

### Manual Proxy Testing

```go
func TestManualProxySetup(t *testing.T) {
    // Create mock servers
    socks5Server := NewMockSOCKS5Server(t, &ProxyAuth{"user", "pass"})
    defer socks5Server.Close()

    httpServer := NewMockHTTPProxyServer(t, &ProxyAuth{"user", "pass"})
    defer httpServer.Close()

    // Test SOCKS5
    config := &goph.Config{
        User: "testuser",
        Addr: "127.0.0.1",
        Port: 22,
        Auth: goph.Password("testpass"),
        Callback: ssh.InsecureIgnoreHostKey(),
        Proxy: &goph.ProxyConfig{
            Type: goph.ProxyTypeSOCKS5,
            Addr: socks5Server.Host(),
            Port: socks5Server.Port(),
            User: "user",
            Password: "pass",
        },
    }

    _, err := goph.NewConn(config)
    AssertSSHError(t, err) // Should fail at SSH level, not proxy level
}
```

## Understanding Test Results

### Expected Behaviors

1. **Proxy Connection Success**: Tests should establish proxy connections successfully
2. **SSH Connection Failure**: Tests should fail at SSH handshake (expected, since no real SSH server)
3. **Proxy Authentication**: Tests should handle auth correctly
4. **Timeout Handling**: Tests should respect timeout settings

### Common Test Scenarios

#### 1. Successful Proxy Connection (fails at SSH)
```
Expected: SSH handshake error
Actual: "ssh: handshake failed" or "connection refused"
```

#### 2. Proxy Authentication Failure
```
Expected: Proxy auth error
Actual: "dial tcp: connection refused" or proxy-specific error
```

#### 3. Timeout Scenarios
```
Expected: Quick failure
Actual: Test completes in < 200ms with timeout error
```

## Real-World Testing

### Testing with Real Proxies

```go
func TestRealProxy(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping real proxy test in short mode")
    }

    // Test with real SOCKS5 proxy
    client, err := goph.NewSOCKS5ProxyConn(
        "sshuser",
        "real-ssh-server.com",
        goph.Password("sshpass"),
        "proxy.example.com",
        1080,
        "proxyuser",    // optional
        "proxypass",    // optional
    )

    if err != nil {
        t.Fatalf("failed to connect via proxy: %v", err)
    }
    defer client.Close()

    // Test SSH operations
    output, err := client.Run("echo 'Proxy connection successful'")
    if err != nil {
        t.Errorf("SSH command failed: %v", err)
    }

    expected := "Proxy connection successful"
    if string(output) != expected {
        t.Errorf("expected %q, got %q", expected, string(output))
    }
}
```

### Testing Different Proxy Types

```go
func TestMultipleProxyTypes(t *testing.T) {
    proxyConfigs := []struct {
        name string
        createClient func() (*goph.Client, error)
    }{
        {
            name: "SOCKS5",
            createClient: func() (*goph.Client, error) {
                return goph.NewSOCKS5ProxyConn(
                    "user", "host.com", goph.Password("pass"),
                    "socks5.proxy.com", 1080, "", "",
                )
            },
        },
        {
            name: "HTTP",
            createClient: func() (*goph.Client, error) {
                return goph.NewHTTPProxyConn(
                    "user", "host.com", goph.Password("pass"),
                    "http.proxy.com", 8080, "user", "pass",
                )
            },
        },
    }

    for _, pc := range proxyConfigs {
        t.Run(pc.name, func(t *testing.T) {
            client, err := pc.createClient()
            if err != nil {
                t.Logf("Proxy connection failed (expected for test): %v", err)
                return
            }
            defer client.Close()

            // Test operations
            _, err = client.Run("uname -a")
            if err != nil {
                t.Errorf("SSH operation failed: %v", err)
            }
        })
    }
}
```

## Debugging Tests

### Enable Debug Logging

```go
func init() {
    // Enable debug logging for proxy connections
    log.SetFlags(log.LstdFlags | log.Lshortfile)
}
```

### Test with Verbose Output

```bash
go test -v -run="TestSOCKS5ProxyIntegration" 2>&1 | grep -E "(PASS|FAIL|RUN)"
```

### Network Debugging

```bash
# Monitor network connections during tests
netstat -tlnp | grep :1080  # SOCKS5 default port
netstat -tlnp | grep :8080  # HTTP proxy default port

# Use tcpdump to capture proxy traffic (if needed)
sudo tcpdump -i lo port 1080 -w proxy_traffic.pcap
```

## Continuous Integration

### CI Configuration Example

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.19

    - name: Run proxy tests
      run: go test -v -run=".*[Pp]roxy.*" -timeout=5m

    - name: Run benchmarks
      run: go test -bench=".*[Pp]roxy.*" -benchmem
```

## Performance Testing

### Benchmark Results

```bash
go test -bench="BenchmarkProxyConnectionCreation" -benchmem
```

Expected output:
```
BenchmarkProxyConnectionCreation-8   	  100000	     15467 ns/op	    3456 B/op	      67 allocs/op
```

### Load Testing

```go
func BenchmarkConcurrentProxyConnections(b *testing.B) {
    setup := NewTestProxySetup(b)
    defer setup.Close()

    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            config := *setup.Config
            config.Proxy = setup.ProxyConfigs[goph.ProxyTypeSOCKS5]

            _, err := goph.NewConn(&config)
            if err == nil {
                b.Error("expected connection to fail")
            }
        }
    })
}
```

## Troubleshooting

### Common Issues

1. **Port conflicts**: Mock servers may fail if ports are in use
2. **Timeout issues**: Tests may fail if system is slow
3. **Network restrictions**: Firewall may block test connections
4. **Go version compatibility**: Ensure Go 1.13+ for module support

### Debugging Steps

1. **Check port availability**:
   ```bash
   netstat -tlnp | grep :1080
   ```

2. **Increase timeouts**:
   ```go
   config.Timeout = 10 * time.Second
   ```

3. **Enable test logging**:
   ```go
   t.Logf("Debug: attempting connection to %s:%d", host, port)
   ```

4. **Test network connectivity**:
   ```bash
   telnet 127.0.0.1 1080
   ```

## Contributing

When adding new proxy tests:

1. Follow the existing naming conventions
2. Add tests to appropriate files based on scope
3. Include both positive and negative test cases
4. Document test expectations clearly
5. Update this README if adding new test patterns

## Test Coverage

Current test coverage includes:

- ✅ SOCKS5 proxy connections
- ✅ HTTP CONNECT proxy connections
- ✅ Proxy authentication (both types)
- ✅ Error handling and validation
- ✅ Timeout behavior
- ✅ Concurrent connections
- ✅ Convenience functions
- ✅ Configuration validation
- ✅ Mock server implementations
- ✅ Integration testing utilities

## Future Enhancements

Potential areas for additional testing:

- Real proxy server integration tests
- Proxy chaining scenarios
- IPv6 proxy support
- Proxy failover testing
- Performance benchmarking under load
- Memory leak detection
- Fuzz testing for proxy protocols
