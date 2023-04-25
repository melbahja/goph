// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type remoteScriptType byte

const (
	cmdLine remoteScriptType = iota
	rawScript
	scriptFile
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

	_type remoteScriptType

	script *bytes.Buffer

	// scriptFile string

	stdout io.Writer
	stderr io.Writer
}

func appendStringToBuffer(str string, buf bytes.Buffer) *bytes.Buffer {
	parentStr := buf.String()
	parentStr = fmt.Sprintf("%v \n %v", str, parentStr)
	return bytes.NewBufferString(parentStr)
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

// Output runs cmd on the remote host and returns its stdout.
func (c *Cmd) Output() ([]byte, error) {
	if err := c.init(); err != nil {
		return nil, errors.Wrap(err, "cmd init")
	}

	return c.runWithContext(func() ([]byte, error) {
		return c.Session.Output(c.String())
	})
}

// Output runs cmd on the remote host and returns its stdout.
func (c *Cmd) ScriptOutput() ([]byte, error) {
	if err := c.init(); err != nil {
		return nil, errors.Wrap(err, "cmd init")
	}
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	c.stdout = &stdout
	c.stderr = &stderr
	return c.runWithContext(func() ([]byte, error) {
		err := c.run()
		if err != nil {
			return stderr.Bytes(), err
		}
		return stdout.Bytes(), err
	})
}

func (c *Cmd) run() error {
	if c._type == cmdLine {
		return c.runCmds()
	} else if c._type == rawScript {
		return c.runScript()
	} else if c._type == scriptFile {
		return nil
		// return c.runScriptFile()
	} else {
		return errors.New("Not supported RemoteScript type")
	}
}

func (c *Cmd) runCmd(cmd string) error {
	c.Session.Stdout = c.stdout
	c.Session.Stderr = c.stderr

	if err := c.Session.Run(cmd); err != nil {
		return err
	}
	return nil
}

func (c *Cmd) runCmds() error {
	for {
		statment, err := c.script.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := c.runCmd(statment); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cmd) runScript() error {
	envs := strings.Join(c.Env, "\n")
	c.script = appendStringToBuffer(fmt.Sprintf("%v \n", envs), *c.script)
	c.Session.Stdin = c.script
	c.Session.Stdout = c.stdout
	c.Session.Stderr = c.stderr
	if err := c.Session.Shell(); err != nil {
		return err
	}
	if err := c.Session.Wait(); err != nil {
		return err
	}
	return nil
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

	if c._type == rawScript {
		return
	}
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
