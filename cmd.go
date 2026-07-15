// Copyright 2026 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"strings"
	"sync/atomic"
)

// Cmd it's like os/exec.Cmd but for ssh session.
type Cmd struct {

	// SSH session.
	*ssh.Session

	// Path to command executable filename
	Path string

	// Command args.
	Args []string

	// Session env vars.
	Env []string

	// Cancel is called when Context is done.
	// If non-nil, it replaces the default ssh.SIGINT signal.
	// If nil, ssh.SIGINT is sent on cancellation.
	Cancel func() error

	initialized atomic.Bool

	// ctx for cancellation
	ctx context.Context
}

// CombinedOutput runs cmd on the remote host and returns its combined stdout and stderr.
func (c *Cmd) CombinedOutput() ([]byte, error) {

	return c.runInContext(func() ([]byte, error) {
		return c.Session.CombinedOutput(c.String())
	})
}

// Output runs cmd on the remote host and returns its stdout.
func (c *Cmd) Output() ([]byte, error) {

	return c.runInContext(func() ([]byte, error) {
		return c.Session.Output(c.String())
	})
}

// Run runs cmd on the remote host.
func (c *Cmd) Run() (err error) {

	_, err = c.runInContext(func() ([]byte, error) {
		return nil, c.Session.Run(c.String())
	})
	return
}

// Start runs the command on the remote host.
func (c *Cmd) Start() (err error) {

	_, err = c.runInContext(func() ([]byte, error) {
		return nil, c.Session.Start(c.String())
	})
	return
}

// String returns the command line string.
//
// WARNING: Path and Args are joined as is without any shell escaping.
// If Path or Args contain untrusted input, remote command injection is possible.
// Sanitize or shell quote untrusted values yourself before building the Cmd,
// or use a shell quoting helper from your application.
func (c *Cmd) String() string {
	return fmt.Sprintf("%s %s", c.Path, strings.Join(c.Args, " "))
}

// Init inits and sets session env vars.
func (c *Cmd) init() (err error) {

	if c.initialized.Load() {
		return nil
	}

	// Set session env vars
	var env []string
	for _, value := range c.Env {
		env = strings.Split(value, "=")
		if err = c.Setenv(env[0], strings.Join(env[1:], "=")); err != nil {
			return
		}
	}

	c.initialized.Store(true)
	return nil
}

// Executes the given callback within session. Sends SIGINT when the context is canceled.
func (c *Cmd) runInContext(callback func() ([]byte, error)) (out []byte, err error) {

	if err = c.init(); err != nil {
		return nil, fmt.Errorf("cmd init: %w", err)
	}

	done := make(chan struct{}, 1)
	go func() {
		out, err = callback()
		done <- struct{}{}
	}()

	select {

	case <-c.ctx.Done():

		if c.Cancel != nil {
			c.Cancel()
		} else {
			c.Signal(ssh.SIGINT)
		}

		return nil, c.ctx.Err()

	case <-done:
		return out, err
	}
}
