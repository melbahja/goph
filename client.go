// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// DefaultClientVersion is the SSH client version sent during handshake.
const DefaultClientVersion = "SSH-2.0-Goph"

// Client represents a Goph SSH client.
type Client struct {
	*ssh.Client

	User     string
	Addr     string
	Port     uint
	ProxyURL string
	Jump     *Client
}

// New starts a new SSH connection.
// By default it uses the default known_hosts file for host key verification,
// port 22, and a 20 second timeout. Override with With* options.
func New(user, addr string, opts ...Option) (*Client, error) {

	c := &Client{
		User: user,
		Addr: addr,
		Port: 22,
	}

	config := &ssh.ClientConfig{
		User:          user,
		Timeout:       20 * time.Second,
		ClientVersion: DefaultClientVersion,
	}

	for _, opt := range opts {
		if err := opt(c, config); err != nil {
			return nil, err
		}
	}

	if err := Dial(c, config); err != nil {
		return nil, err
	}

	return c, nil
}

// Run starts a new SSH session and runs the cmd, it returns CombinedOutput and err if any.
func (c *Client) Run(cmd string) ([]byte, error) {

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

// RunContext starts a new SSH session with context and runs the cmd.
func (c *Client) RunContext(ctx context.Context, name string) ([]byte, error) {
	cmd, err := c.CommandContext(ctx, name)
	if err != nil {
		return nil, err
	}

	return cmd.CombinedOutput()
}

// Command returns new Cmd and error if any.
func (c *Client) Command(name string, args ...string) (*Cmd, error) {
	return c.CommandContext(context.Background(), name, args...)
}

// CommandContext returns new Cmd with context and error, if any.
func (c *Client) CommandContext(ctx context.Context, name string, args ...string) (*Cmd, error) {

	sess, err := c.NewSession()
	if err != nil {
		return nil, err
	}

	return &Cmd{
		ctx:     ctx,
		Path:    name,
		Args:    args,
		Session: sess,
	}, nil
}

// NewSftp returns a new SFTP client.
func (c *Client) NewSftp(opts ...sftp.ClientOption) (*sftp.Client, error) {
	return sftp.NewClient(c.Client, opts...)
}

// Close closes the SSH connection.
func (c *Client) Close() error {
	return c.Client.Close()
}

// Upload a local file to the remote server.
func (c *Client) Upload(localPath string, remotePath string) (err error) {

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

// Download file from remote server.
func (c *Client) Download(remotePath string, localPath string) (err error) {

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

// Script runs a script from an io.Reader on the remote host via /bin/sh you can override (with WithPath).
func (c *Client) Script(ctx context.Context, r io.Reader, opts ...CmdOption) (cmd *Cmd, err error) {

	if cmd, err = c.CommandContext(ctx, "/bin/sh"); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(cmd)
	}
	cmd.Stdin = r

	return cmd, nil
}

// ScriptFile reads the local script file into memory and executes it on the remote.
// Warning: the entire file is loaded into memory. Use Script with an io.Reader directly for large files.
func (c *Client) ScriptFile(ctx context.Context, localPath string, opts ...CmdOption) (*Cmd, error) {

	content, err := os.ReadFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("goph: read script: %w", err)
	}

	return c.Script(ctx, bytes.NewReader(content), opts...)
}
