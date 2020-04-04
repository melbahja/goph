// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"time"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	Port  int
	Auth  Auth
	Addr  string
	User  string
	Conn  *ssh.Client
	Proto string
}

// Connect to ssh and get client
// The key must be in known hosts
func New(user string, addr string, auth Auth) (c *Client, err error) {

	callback, err := DefaultKnownHosts()

	if err != nil {
		return
	}

	c, err = NewConn(user, addr, auth, callback)
	return
}

// Connect to ssh and get client without cheking knownhosts
//
// PLEASE AVOID USING THIS, UNLESS YOU KNOW WHAT ARE YOU DOING!
// if there a "man in the middle proxy", this can harm you!
// You can add the key to know hosts and use New() func instead!
func NewUnknown(user string, addr string, auth Auth) (*Client, error) {

	return NewConn(user, addr, auth, ssh.InsecureIgnoreHostKey())
}

// Get new client connection
func NewConn(user string, addr string, auth Auth, callback ssh.HostKeyCallback) (c *Client, err error) {

	c = &Client{
		User: user,
		Addr: addr,
		Auth: auth,
	}

	err = Conn(c, &ssh.ClientConfig{
		User:            c.User,
		Auth:            c.Auth,
		Timeout:         20 * time.Second,
		HostKeyCallback: callback,
	})

	return
}

// Get new ssh session from client connection
// See: https://pkg.go.dev/golang.org/x/crypto/ssh?tab=doc#Session
func (c Client) NewSession() (*ssh.Session, error) {

	return c.Conn.NewSession()
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

// Upload a local file to remote machine!
func (c Client) Upload(localPath string, remotePath string) error {

	return Upload(c.Conn, localPath, remotePath)
}

// Download file from remote machine!
func (c Client) Download(remotePath string, localPath string) error {

	return Download(c.Conn, remotePath, localPath)
}
