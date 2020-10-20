// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Cmd struct {
	Path string
	Args []string
	Env  []string
	*ssh.Session
}

func (c *Cmd) CombinedOutput() ([]byte, error) {
	return c.init().Session.CombinedOutput(c.String())
}

func (c *Cmd) Output() ([]byte, error) {
	return c.init().Session.Output(c.String())
}

func (c *Cmd) Run() error {
	return c.init().Session.Run(c.String())
}

func (c *Cmd) Start() error {
	return c.init().Session.Start(c.String())
}

func (c *Cmd) String() string {
	return fmt.Sprintf("%s %s", c.Path, strings.Join(c.Args, " "))
}

func (c *Cmd) init() *Cmd {

	// Set session env vars
	var env []string
	for _, value := range c.Env {
		env = strings.Split(value, "=")
		c.Setenv(env[0], strings.Join(env[1:], "="))
	}

	return c
}
