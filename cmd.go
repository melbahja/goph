// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"strings"
)

// Cmd it's like os/exec.Cmd but for ssh session.
type Cmd struct {

	// Path to command executable filename
	Path string

	// Command args.
	Args []string

	// Session env vars.
	Env []string

	// SSH session.
	*ssh.Session

	// Context for cancellation
	Context context.Context
}

// CombinedOutput runs cmd on the remote host and returns its combined stdout and stderr.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	if err := c.init(); err != nil {
		return nil, errors.Wrap(err, "cmd init")
	}

	return c.runWithContext(func() ([]byte, error) {
		return c.Session.CombinedOutput(c.String())
	})
}

// SeparatedOutput runs cmd on the remote host and returns its stdout and stderr.
func (c *Cmd) SeparatedOutput() ([]byte, []byte, error) {
	var stdout_buff, stderr_buff bytes.Buffer
	c.Session.Stdout = &stdout_buff
	c.Session.Stderr = &stderr_buff

	if err := c.init(); err != nil {
		return nil, nil, errors.Wrap(err, "cmd init")
	}

	return c.runWithContextAndErr(func() ([]byte, []byte, error) {
		err := c.Session.Run(c.String())
		return stdout_buff.Bytes(), stderr_buff.Bytes(), err
	})
}

// Output runs cmd on the remote host and returns its stdout.
func (c *Cmd) Output() ([]byte, error) {
	if err := c.init(); err != nil {
		return nil, errors.Wrap(err, "cmd init")
	}

	return c.runWithContext(func() ([]byte, error) {
		return c.Session.Output(c.String())
	})
}

// Run runs cmd on the remote host.
func (c *Cmd) Run() error {
	if err := c.init(); err != nil {
		return errors.Wrap(err, "cmd init")
	}

	_, err := c.runWithContext(func() ([]byte, error) {
		return nil, c.Session.Run(c.String())
	})

	return err
}

// Start runs the command on the remote host.
func (c *Cmd) Start() error {
	if err := c.init(); err != nil {
		return errors.Wrap(err, "cmd init")
	}
	return c.Session.Start(c.String())
}

// String return the command line string.
func (c *Cmd) String() string {
	return fmt.Sprintf("%s %s", c.Path, strings.Join(c.Args, " "))
}

// Init inits and sets session env vars.
func (c *Cmd) init() (err error) {

	// Set session env vars
	var env []string
	for _, value := range c.Env {
		env = strings.Split(value, "=")
		if err = c.Setenv(env[0], strings.Join(env[1:], "=")); err != nil {
			return
		}
	}

	return nil
}

// Command with context output.
type ctxCmdOutputAndErr struct {
	stdout []byte
	stderr []byte
	err    error
}

// Command with context output.
type ctxCmdOutput struct {
	output []byte
	err    error
}

// Executes the given callback within session. Sends SIGINT when the context is canceled.
func (c *Cmd) runWithContext(callback func() ([]byte, error)) ([]byte, error) {
	outputChan := make(chan ctxCmdOutput)
	go func() {
		output, err := callback()
		outputChan <- ctxCmdOutput{
			output: output,
			err:    err,
		}
	}()

	select {
	case <-c.Context.Done():
		_ = c.Session.Signal(ssh.SIGINT)

		return nil, c.Context.Err()
	case result := <-outputChan:
		return result.output, result.err
	}
}

// Executes the given callback within session. Sends SIGINT when the context is canceled.
func (c *Cmd) runWithContextAndErr(callback func() ([]byte, []byte, error)) ([]byte, []byte, error) {
	outputChan := make(chan ctxCmdOutputAndErr)
	go func() {
		stdout, stderr, err := callback()
		outputChan <- ctxCmdOutputAndErr{
			stdout: stdout,
			stderr: stderr,
			err:    err,
		}
	}()

	select {
	case <-c.Context.Done():
		_ = c.Session.Signal(ssh.SIGINT)

		return nil, nil, c.Context.Err()
	case result := <-outputChan:
		return result.stdout, result.stderr, result.err
	}
}
