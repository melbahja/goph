// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

const (
	UDP string = "udp"
	TCP string = "tcp"
)

// Set new net connection to a client.
func Conn(c *Client, cfg *ssh.ClientConfig) (err error) {

	if c.Port == 0 {
		c.Port = 22
	}

	if c.Proto == "" {
		c.Proto = TCP
	}

	c.Conn, err = ssh.Dial(c.Proto, fmt.Sprintf("%s:%d", c.Addr, c.Port), cfg)
	return
}
