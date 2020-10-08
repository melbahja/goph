// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Cmd struct {
	Path    string
	Args    []string
	Env     []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Session *ssh.Session
}

func (c *Cmd) CombinedOutput() ([]byte, error) {
	return c.init().CombinedOutput()
}

func (c *Cmd) Output() ([]byte, error) {
	return c.init().Output()
}

func (c *Cmd) Run() error {
	return c.init().Run()
}

func (c *Cmd) Start() error {
	return c.init().Start()
}

func (c *Cmd) StderrPipe() (io.Reader, error) {
	return c.Session.StderrPipe()
}

func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	return c.Session.StdinPipe()
}

func (c *Cmd) StdoutPipe() (io.Reader, error) {
	return c.Session.StdoutPipe()
}

func (c *Cmd) String() string {
	return fmt.Sprintf("%s %s", c.Path, strings.Join(c.Args, " "))
}

func (c *Cmd) Wait() error {
	return c.Session.Wait()
}

func (c *Cmd) init() *Cmd {

	// Set session env vars
	var envParts []string
	for _, env := range c.Env {
		envParts = strings.Split(env, "=")
		c.Session.Setenv(envParts[0], strings.Join(envParts[1:], "="))
	}

	// Set session stdio
	c.Session.Stdin = c.Stdin
	c.Session.Stdout = c.Stdout
	c.Session.Stderr = c.Stderr
	return c
}
