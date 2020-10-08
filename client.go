// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	*ssh.Client
	Config *Config
	sftp   *sftp.Client
}

type Config struct {
	Auth     Auth
	User     string
	Addr     string
	Port     uint
	Timeout  time.Duration
	Callback ssh.HostKeyCallback
}

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

// Get new client connection.
func NewConn(config *Config) (c *Client, err error) {

	c = &Client{
		Config: config,
	}

	c.Client, err = Dial("tcp", config)
	return
}

func Dial(proto string, c *Config) (*ssh.Client, error) {
	return ssh.Dial(proto, fmt.Sprintf("%s:%d", c.Addr, c.Port), &ssh.ClientConfig{
		User:            c.User,
		Auth:            c.Auth,
		Timeout:         c.Timeout,
		HostKeyCallback: c.Callback,
	})
}

// Run a command over ssh connection
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
	}, nil
}

func (c Client) NewSftp(opts ...sftp.ClientOption) (*sftp.Client, error) {
	return sftp.NewClient(c.Client, opts...)
}

// Close client net connection.
func (c Client) Close() error {

	if c.sftp != nil {
		c.sftp.Close()
	}

	return c.Client.Close()
}

func (c *Client) ftp() *sftp.Client {

	if c.sftp == nil {
		sftp, err := c.NewSftp()
		if err != nil {
			panic(err)
		}
		c.sftp = sftp
	}

	return c.sftp
}

// Upload a local file to remote machine!
func (c Client) Upload(localPath string, remotePath string) (err error) {

	local, err := os.Open(localPath)

	if err != nil {
		return
	}

	defer local.Close()

	remote, err := c.ftp().Create(remotePath)

	if err != nil {
		return
	}

	defer remote.Close()

	_, err = io.Copy(remote, local)
	return
}

// Download file from remote machine!
func (c Client) Download(remotePath string, localPath string) (err error) {

	local, err := os.Create(localPath)

	if err != nil {
		return
	}

	defer local.Close()

	remote, err := c.ftp().Open(remotePath)

	if err != nil {
		return
	}

	defer remote.Close()

	if _, err = io.Copy(local, remote); err != nil {
		return
	}

	return local.Sync()
}
