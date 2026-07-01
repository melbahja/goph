// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"fmt"
	"net"
	"net/url"
	"os/user"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

// Dialer holds reusable connection options for establishing SSH connections.
//
//	d := goph.NewDialer(goph.WithKeyFile("id_rsa", ""))
//	c1, _ := d.New("root", "host1")
//	c2, _ := d.New("root", "host2", goph.WithPort(2222))
type Dialer struct {
	opts []Option
}

// NewDialer creates a Dialer with the given options.
func NewDialer(opts ...Option) *Dialer {
	return &Dialer{opts: opts}
}

// New establishes an SSH connection using the Dialer options merged with
// per call options. Per call options take precedence (applied after).
func (d Dialer) New(user, addr string, opts ...Option) (*Client, error) {
	return New(user, addr, append(d.opts, opts...)...)
}

// Dial establishes the SSH connection described by c and config.
// It honors c.User, c.ProxyURL, c.Jump, and c.Port, and
// applies the default known hosts callback if HostKeyCallback is nil.
//
// If config.User is empty, it is set from c.User. If c.User is also empty,
// the current OS user is used, matching OpenSSH default behavior.
func Dial(c *Client, config *ssh.ClientConfig) error {

	if c.Jump != nil && c.ProxyURL != "" {
		return fmt.Errorf("goph: cannot use WithProxy and WithJump, put WithProxy on the jump client instead.")
	}

	if config.User == "" {
		config.User = c.User
		if config.User == "" {
			if u, err := user.Current(); err == nil {
				c.User = u.Username
				config.User = u.Username
			}
		}
	}

	if config.HostKeyCallback == nil {
		callback, err := DefaultKnownHosts()
		if err != nil {
			return err
		}
		config.HostKeyCallback = callback
	}

	var (
		err    error
		conn   net.Conn
		target = net.JoinHostPort(c.Addr, fmt.Sprint(c.Port))
	)

	switch {
	case c.Jump != nil:
		conn, err = dialJump(c, target)
	case c.ProxyURL != "":
		conn, err = dialProxy(c.ProxyURL, target)
	default:
		conn, err = net.DialTimeout("tcp", target, config.Timeout)
	}

	if err != nil {
		return err
	}

	cc, chans, reqs, err := ssh.NewClientConn(conn, target, config)
	if err != nil {
		conn.Close()
		return err
	}

	c.Client = ssh.NewClient(cc, chans, reqs)
	return nil
}

// dialProxy returns a TCP connection to addr through a SOCKS5 proxy.
func dialProxy(proxyURL, addr string) (net.Conn, error) {

	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("proxy: invalid proxy URL: %w", err)
	}

	proxyAddr := u.Host
	if proxyAddr == "" {
		proxyAddr = proxyURL
	}

	var auth *proxy.Auth
	if u.User != nil {
		auth = &proxy.Auth{
			User: u.User.Username(),
		}
		if pass, ok := u.User.Password(); ok {
			auth.Password = pass
		}
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("proxy dial: %w", err)
	}

	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("proxy dial: %w", err)
	}

	return conn, nil
}

// dialJump returns a TCP connection to addr through an existing jump client SSH tunnel.
func dialJump(c *Client, addr string) (net.Conn, error) {

	conn, err := c.Jump.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("jump dial: %w", err)
	}

	return conn, nil
}
