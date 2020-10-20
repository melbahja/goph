// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Cmd it's like os/exec.Cmd but for ssh session.
type Cmd struct {

	// Path to command executable filename
	Path string

	// Command args.
	Args []string

	// Session env vars.
	Env []string

	// ssh session.
	*ssh.Session
}

// CombinedOutput runs cmd on the remote host and returns its combined standard output and standard error.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	return c.init().Session.CombinedOutput(c.String())
}

// Output runs cmd on the remote host and returns its standard output.
func (c *Cmd) Output() ([]byte, error) {
	return c.init().Session.Output(c.String())
}

// Run runs cmd on the remote host.
func (c *Cmd) Run() error {
	return c.init().Session.Run(c.String())
}

// Start runs the command on the remote host.
func (c *Cmd) Start() error {
	return c.init().Session.Start(c.String())
}

// String return the command line string.
func (c *Cmd) String() string {
	return fmt.Sprintf("%s %s", c.Path, strings.Join(c.Args, " "))
}

// init inits Cmd.Env vars.
func (c *Cmd) init() *Cmd {

	// Set session env vars
	var env []string
	for _, value := range c.Env {
		env = strings.Split(value, "=")
		c.Setenv(env[0], strings.Join(env[1:], "="))
	}

	return c
}
