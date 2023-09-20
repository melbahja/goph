// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"io"
	"net"
	"os"
	"time"

	"github.com/kevinburke/ssh_config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
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
}

// DefaultTimeout is the timeout of ssh client connection.
var DefaultTimeout = 20 * time.Second

func GetAuth(host string) (Auth, error) {
	auth, err := Key(getSshConfig(host, "IdentityFile"), "")
	if err != nil {
		return nil, fmt.Errorf("get key: %w", err)
	}
	return auth, nil
}

func getSshConfig(alias string, key string) string {
	// to fix https://github.com/kevinburke/ssh_config/issues/61 issue
	return strings.Trim(ssh_config.Get(alias, key), "\"")
}

func NewWithConfigFileHostKeyCallback(name string, callback ssh.HostKeyCallback) (*Client, error){
	auth, err := GetAuth(name)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(getSshConfig(name, "Port"))
	if err != nil {
		return nil, err
	}

	return NewConn(&Config{
			User:     getSshConfig(name, "User"),
			Addr:     getSshConfig(name, "HostName"),
			Port:     uint(port),
			Auth:     auth,
			Timeout:  DefaultTimeout,
			Callback: callback,
	})

}

// NewWithConfigFile starts a new ssh connection with given config file, the host public key must be in known hosts.
func NewWithConfigFile(name string) (*Client, error) {
	callback, err := DefaultKnownHosts()
	if err != nil {
		return nil, err
	}
	return NewWithConfigFileHostKeyCallback(name, callback)
}

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

func NewWithConfigFileUnknown(name string) (*Client, error) {
	return NewWithConfigFileHostKeyCallback(name, ssh.InsecureIgnoreHostKey())
}

// NewConn returns new client and error if any.
func NewConn(config *Config) (c *Client, err error) {

	c = &Client{
		Config: config,
	}

	c.Client, err = Dial("tcp", config)
	return
}

// Dial starts a client connection to SSH server based on config.
func Dial(proto string, c *Config) (*ssh.Client, error) {
	return ssh.Dial(proto, net.JoinHostPort(c.Addr, fmt.Sprint(c.Port)), &ssh.ClientConfig{
		User:            c.User,
		Auth:            c.Auth,
		Timeout:         c.Timeout,
		HostKeyCallback: c.Callback,
		BannerCallback:  c.BannerCallback,
	})
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
