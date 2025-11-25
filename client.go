// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

// Client represents Goph client.
type Client struct {
	*ssh.Client
	Config *Config
}

// Config for Client.
type Config struct {
	Auth           Auth
	User           string
	Addr           string
	Port           uint
	Timeout        time.Duration
	Callback       ssh.HostKeyCallback
	BannerCallback ssh.BannerCallback
	Proxy          *ProxyConfig
}

// ProxyConfig holds proxy configuration.
type ProxyConfig struct {
	Type     ProxyType
	Addr     string
	Port     uint
	User     string
	Password string
}

// ProxyType represents the type of proxy.
type ProxyType int

const (
	// ProxyTypeNone indicates no proxy
	ProxyTypeNone ProxyType = iota
	// ProxyTypeSOCKS5 indicates SOCKS5 proxy
	ProxyTypeSOCKS5
	// ProxyTypeHTTP indicates HTTP CONNECT proxy
	ProxyTypeHTTP
)

// DefaultTimeout is the timeout of ssh client connection.
var DefaultTimeout = 20 * time.Second

// New starts a new ssh connection, the host public key must be in known hosts.
func New(user string, addr string, auth Auth) (c *Client, err error) {

	callback, err := DefaultKnownHosts()

	if err != nil {
		return
	}

	c, err = NewConn(&Config{
		User:     user,
		Addr:     addr,
		Port:     22,
		Auth:     auth,
		Timeout:  DefaultTimeout,
		Callback: callback,
	})
	return
}

// NewUnknown starts a ssh connection get client without cheking knownhosts.
// PLEASE AVOID USING THIS, UNLESS YOU KNOW WHAT ARE YOU DOING!
// if there a "man in the middle proxy", this can harm you!
// You can add the key to know hosts and use New() func instead!
func NewUnknown(user string, addr string, auth Auth) (*Client, error) {
	return NewConn(&Config{
		User:     user,
		Addr:     addr,
		Port:     22,
		Auth:     auth,
		Timeout:  DefaultTimeout,
		Callback: ssh.InsecureIgnoreHostKey(),
	})
}

// NewConn returns new client and error if any.
func NewConn(config *Config) (c *Client, err error) {

	c = &Client{
		Config: config,
	}

	c.Client, err = Dial("tcp", config)
	return
}

// NewProxyConn returns new client with proxy configuration and error if any.
func NewProxyConn(config *Config, proxyType ProxyType, proxyAddr string, proxyPort uint, proxyUser, proxyPass string) (*Client, error) {
	config.Proxy = &ProxyConfig{
		Type:     proxyType,
		Addr:     proxyAddr,
		Port:     proxyPort,
		User:     proxyUser,
		Password: proxyPass,
	}
	return NewConn(config)
}

// NewSOCKS5ProxyConn returns new client with SOCKS5 proxy configuration.
func NewSOCKS5ProxyConn(user, addr string, auth Auth, proxyAddr string, proxyPort uint, proxyUser, proxyPass string) (*Client, error) {
	callback, _ := DefaultKnownHosts()
	return NewProxyConn(&Config{
		User:     user,
		Addr:     addr,
		Port:     22,
		Auth:     auth,
		Timeout:  DefaultTimeout,
		Callback: callback,
	}, ProxyTypeSOCKS5, proxyAddr, proxyPort, proxyUser, proxyPass)
}

// NewHTTPProxyConn returns new client with HTTP CONNECT proxy configuration.
func NewHTTPProxyConn(user, addr string, auth Auth, proxyAddr string, proxyPort uint, proxyUser, proxyPass string) (*Client, error) {
	callback, _ := DefaultKnownHosts()
	return NewProxyConn(&Config{
		User:     user,
		Addr:     addr,
		Port:     22,
		Auth:     auth,
		Timeout:  DefaultTimeout,
		Callback: callback,
	}, ProxyTypeHTTP, proxyAddr, proxyPort, proxyUser, proxyPass)
}

// NewSOCKS5ProxyUnknown returns new client with SOCKS5 proxy and insecure host key checking.
func NewSOCKS5ProxyUnknown(user, addr string, auth Auth, proxyAddr string, proxyPort uint, proxyUser, proxyPass string) (*Client, error) {
	return NewProxyConn(&Config{
		User:     user,
		Addr:     addr,
		Port:     22,
		Auth:     auth,
		Timeout:  DefaultTimeout,
		Callback: ssh.InsecureIgnoreHostKey(),
	}, ProxyTypeSOCKS5, proxyAddr, proxyPort, proxyUser, proxyPass)
}

// NewHTTPProxyUnknown returns new client with HTTP CONNECT proxy and insecure host key checking.
func NewHTTPProxyUnknown(user, addr string, auth Auth, proxyAddr string, proxyPort uint, proxyUser, proxyPass string) (*Client, error) {
	return NewProxyConn(&Config{
		User:     user,
		Addr:     addr,
		Port:     22,
		Auth:     auth,
		Timeout:  DefaultTimeout,
		Callback: ssh.InsecureIgnoreHostKey(),
	}, ProxyTypeHTTP, proxyAddr, proxyPort, proxyUser, proxyPass)
}

// Dial starts a client connection to SSH server based on config.
func Dial(proto string, c *Config) (*ssh.Client, error) {
	var conn net.Conn
	var err error

	if c.Proxy != nil && c.Proxy.Type != ProxyTypeNone {
		conn, err = dialThroughProxy(c)
		if err != nil {
			return nil, err
		}
	} else {
		conn, err = net.DialTimeout(proto, net.JoinHostPort(c.Addr, fmt.Sprint(c.Port)), c.Timeout)
		if err != nil {
			return nil, err
		}
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, net.JoinHostPort(c.Addr, fmt.Sprint(c.Port)), &ssh.ClientConfig{
		User:            c.User,
		Auth:            c.Auth,
		Timeout:         c.Timeout,
		HostKeyCallback: c.Callback,
		BannerCallback:  c.BannerCallback,
	})
	if err != nil {
		conn.Close()
		return nil, err
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}

// dialThroughProxy establishes a connection through the configured proxy.
func dialThroughProxy(c *Config) (net.Conn, error) {
	proxyAddr := net.JoinHostPort(c.Proxy.Addr, fmt.Sprint(c.Proxy.Port))

	switch c.Proxy.Type {
	case ProxyTypeSOCKS5:
		return dialSOCKS5Proxy(proxyAddr, c.Proxy.User, c.Proxy.Password, c.Addr, c.Port, c.Timeout)
	case ProxyTypeHTTP:
		return dialHTTPProxy(proxyAddr, c.Proxy.User, c.Proxy.Password, c.Addr, c.Port, c.Timeout)
	default:
		return nil, fmt.Errorf("unsupported proxy type: %v", c.Proxy.Type)
	}
}

// dialSOCKS5Proxy establishes a connection through a SOCKS5 proxy.
func dialSOCKS5Proxy(proxyAddr, proxyUser, proxyPass, targetAddr string, targetPort uint, timeout time.Duration) (net.Conn, error) {
	var auth *proxy.Auth
	if proxyUser != "" {
		auth = &proxy.Auth{
			User:     proxyUser,
			Password: proxyPass,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	target := net.JoinHostPort(targetAddr, fmt.Sprint(targetPort))

	// Set timeout on the dialer if it's supported
	if timeoutDialer, ok := dialer.(interface{ SetTimeout(time.Duration) }); ok {
		timeoutDialer.SetTimeout(timeout)
	}

	return dialer.Dial("tcp", target)
}

// dialHTTPProxy establishes a connection through an HTTP CONNECT proxy.
func dialHTTPProxy(proxyAddr, proxyUser, proxyPass, targetAddr string, targetPort uint, timeout time.Duration) (net.Conn, error) {
	// First connect to the proxy
	conn, err := net.DialTimeout("tcp", proxyAddr, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to HTTP proxy: %w", err)
	}

	target := net.JoinHostPort(targetAddr, fmt.Sprint(targetPort))

	// Send CONNECT request
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\n", target)
	connectReq += fmt.Sprintf("Host: %s\r\n", target)

	if proxyUser != "" {
		// Basic auth for proxy
		auth := proxyUser + ":" + proxyPass
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
		connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", encodedAuth)
	}

	connectReq += "\r\n"

	if _, err := conn.Write([]byte(connectReq)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send CONNECT request: %w", err)
	}

	// Read response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read proxy response: %w", err)
	}

	response := string(buffer[:n])
	if !strings.Contains(response, "200") {
		conn.Close()
		return nil, fmt.Errorf("proxy CONNECT failed: %s", response)
	}

	return conn, nil
}

// Run starts a new SSH session and runs the cmd, it returns CombinedOutput and err if any.
func (c Client) Run(cmd string) ([]byte, error) {

	var (
		err  error
		sess *ssh.Session
	)

	if sess, err = c.NewSession(); err != nil {
		return nil, err
	}

	defer sess.Close()

	return sess.CombinedOutput(cmd)
}

// Run starts a new SSH session with context and runs the cmd. It returns CombinedOutput and err if any.
func (c Client) RunContext(ctx context.Context, name string) ([]byte, error) {
	cmd, err := c.CommandContext(ctx, name)
	if err != nil {
		return nil, err
	}

	return cmd.CombinedOutput()
}

// Command returns new Cmd and error if any.
func (c Client) Command(name string, args ...string) (*Cmd, error) {

	var (
		sess *ssh.Session
		err  error
	)

	if sess, err = c.NewSession(); err != nil {
		return nil, err
	}

	return &Cmd{
		Path:    name,
		Args:    args,
		Session: sess,
		Context: context.Background(),
	}, nil
}

// Command returns new Cmd with context and error, if any.
func (c Client) CommandContext(ctx context.Context, name string, args ...string) (*Cmd, error) {
	cmd, err := c.Command(name, args...)
	if err != nil {
		return cmd, err
	}

	cmd.Context = ctx

	return cmd, nil
}

// NewSftp returns new sftp client and error if any.
func (c Client) NewSftp(opts ...sftp.ClientOption) (*sftp.Client, error) {
	return sftp.NewClient(c.Client, opts...)
}

// Close client net connection.
func (c Client) Close() error {
	return c.Client.Close()
}

// Upload a local file to remote server!
func (c Client) Upload(localPath string, remotePath string) (err error) {

	local, err := os.Open(localPath)
	if err != nil {
		return
	}
	defer local.Close()

	ftp, err := c.NewSftp()
	if err != nil {
		return
	}
	defer ftp.Close()

	remote, err := ftp.Create(remotePath)
	if err != nil {
		return
	}
	defer remote.Close()

	_, err = io.Copy(remote, local)
	return
}

// Download file from remote server!
func (c Client) Download(remotePath string, localPath string) (err error) {

	local, err := os.Create(localPath)
	if err != nil {
		return
	}
	defer local.Close()

	ftp, err := c.NewSftp()
	if err != nil {
		return
	}
	defer ftp.Close()

	remote, err := ftp.Open(remotePath)
	if err != nil {
		return
	}
	defer remote.Close()

	if _, err = io.Copy(local, remote); err != nil {
		return
	}

	return local.Sync()
}
