// Copyright 2026 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"io"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Option is a functional option for configuring a Client.
type Option func(*Client, *ssh.ClientConfig) error

// CmdOption configures a Cmd after creation.
type CmdOption func(*Cmd)

// WithPassword sets password authentication.
func WithPassword(password string) Option {
	return func(c *Client, config *ssh.ClientConfig) error {
		config.Auth = append(config.Auth, ssh.Password(password))
		return nil
	}
}

// WithKeyboardInteractive sets keyboard interactive authentication using a
// user provided handler, the handler is called once for each prompt the server
// sends, and returns the answer or an error.
func WithKeyboardInteractive(handler func(user, instruction, question string, echo bool) (string, error)) Option {
	return func(c *Client, config *ssh.ClientConfig) error {
		config.Auth = append(config.Auth,
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i, q := range questions {
					answer, err := handler(user, instruction, q, echos[i])
					if err != nil {
						return nil, err
					}
					answers[i] = answer
				}
				return answers, nil
			}),
		)
		return nil
	}
}

// WithKeyFile sets public key authentication from a private key file.
func WithKeyFile(keyFile string, passphrase string) Option {

	return func(c *Client, config *ssh.ClientConfig) error {

		signer, err := ParseKeyFile(keyFile, passphrase)
		if err != nil {
			return err
		}

		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
		return nil
	}
}

// WithKey sets public key authentication from raw PEM encoded key bytes.
func WithKey(pemBytes []byte, passphrase string) Option {

	return func(c *Client, config *ssh.ClientConfig) error {

		signer, err := ParseKey(pemBytes, passphrase)
		if err != nil {
			return err
		}

		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
		return nil
	}
}

// WithAgent sets SSH agent authentication from an existing agent net.Conn.
// The conn must be connected to an SSH agent (Unix socket, Windows pipe, etc.).
// Silently skips if conn is nil.
func WithAgent(conn net.Conn) Option {

	return func(c *Client, config *ssh.ClientConfig) error {

		if conn != nil {
			config.Auth = append(config.Auth, ssh.PublicKeysCallback(agent.NewClient(conn).Signers))
		}

		return nil
	}
}

// WithAgentSocket sets SSH agent authentication from a Unix socket path.
// Silently skips if the agent socket is unavailable.
// The socket is dialed lazily during the SSH handshake, so this option is
// safe to apply multiple times (e.g. for connection key calculation).
func WithAgentSocket(socket string) Option {

	return func(c *Client, config *ssh.ClientConfig) error {

		if socket == "" {
			return nil
		}

		config.Auth = append(config.Auth, ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			conn, err := net.Dial("unix", socket)
			if err != nil {
				// Silently skip the agent if the socket is not reachable.
				return nil, nil
			}
			return agent.NewClient(conn).Signers()
		}))

		return nil
	}
}

// WithDefaultAgent sets SSH agent authentication using the SSH_AUTH_SOCK environment variable.
func WithDefaultAgent() Option {
	return WithAgentSocket(os.Getenv("SSH_AUTH_SOCK"))
}

// WithAuth appends a custom ssh.AuthMethod.
func WithAuth(method ssh.AuthMethod) Option {

	return func(c *Client, config *ssh.ClientConfig) error {
		config.Auth = append(config.Auth, method)
		return nil
	}
}

// WithSigner appends public key authentication from an ssh.Signer.
func WithSigner(signer ssh.Signer) Option {

	return func(c *Client, config *ssh.ClientConfig) error {
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
		return nil
	}
}

// WithPort sets the SSH port.
func WithPort(port uint) Option {

	return func(c *Client, config *ssh.ClientConfig) error {
		c.Port = port
		return nil
	}
}

// WithTimeout sets the connection timeout.
func WithTimeout(d time.Duration) Option {

	return func(c *Client, config *ssh.ClientConfig) error {
		config.Timeout = d
		return nil
	}
}

// WithKnownHosts uses the known hosts file for host key verification.
func WithKnownHosts(path string) Option {

	return func(c *Client, config *ssh.ClientConfig) error {

		cb, err := KnownHosts(path)
		if err != nil {
			return err
		}

		config.HostKeyCallback = cb
		return nil
	}
}

// WithInsecureIgnoreHostKey disables host key verification.
func WithInsecureIgnoreHostKey() Option {

	return func(c *Client, config *ssh.ClientConfig) error {
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		return nil
	}
}

// WithHostKeyCallback sets a custom host key callback.
func WithHostKeyCallback(cb ssh.HostKeyCallback) Option {

	return func(c *Client, config *ssh.ClientConfig) error {
		config.HostKeyCallback = cb
		return nil
	}
}

// WithBannerCallback sets a banner callback.
func WithBannerCallback(cb ssh.BannerCallback) Option {

	return func(c *Client, config *ssh.ClientConfig) error {
		config.BannerCallback = cb
		return nil
	}
}

// WithConfig applies a callback to the underlying *ssh.ClientConfig before dial.
func WithConfig(fn func(*ssh.ClientConfig) error) Option {

	return func(c *Client, config *ssh.ClientConfig) error {
		return fn(config)
	}
}

// WithProxy routes the SSH connection through a SOCKS5 proxy. eg like socks5://127.0.0.1:1080
func WithProxy(socks5URL string) Option {

	return func(c *Client, config *ssh.ClientConfig) error {
		c.ProxyURL = socks5URL
		return nil
	}
}

// WithJump sets a pre connected jump client to tunnel through.
func WithJump(jumpClient *Client) Option {

	return func(c *Client, config *ssh.ClientConfig) error {
		c.Jump = jumpClient
		return nil
	}
}

// WithPath sets the command executable path (default for Script: "/bin/sh").
func WithPath(path string) CmdOption {
	return func(c *Cmd) {
		c.Path = path
	}
}

// WithStdout sets the remote command's stdout writer.
func WithStdout(w io.Writer) CmdOption {
	return func(c *Cmd) {
		c.Stdout = w
	}
}

// WithStderr sets the remote command's stderr writer.
func WithStderr(w io.Writer) CmdOption {
	return func(c *Cmd) {
		c.Stderr = w
	}
}
