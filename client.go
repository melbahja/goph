package ssh

import (
	"os"
	"io"
	"fmt"
	"bytes"
	// "math/rand"
	"github.com/segmentio/ksuid"
	gossh "golang.org/x/crypto/ssh"
)

type Client struct {
	Client *gossh.Client
	Conn ConnectionConfig
}

type Result struct {

	Stdout bytes.Buffer

	Stderr bytes.Buffer
}

// 
// Get a Command
//
func (c *Client) Command(command string, args []string) *Command {

	return &Command{
		Client: c,
		Command: command,
		Args: args,
	}
}

// 
// Run Cmd command and get Result and error
//
func (c Client) Run(command string) (Result, error) {

	return c.Command(command, []string{}).Run()
}

// 
// Exec command and get output as a string
//
func (c Client) Exec(command string) (string, error) {
	
	out, err := c.Run(command)

	if err != nil {
		return "", err
	}

	return out.Stdout.String(), nil
}

// 
// Copy file from local machine to remote machine
//
func (c Client) CopyFileToRemote(src string, dist string) (err error) {

	var file io.Reader

	if file, err = os.Open(src); err != nil {
		return
	}

	cmd := c.Command("cp", []string{"/dev/stdin", dist})
	cmd.Stdin = file

	_, err = cmd.Run()  
	return
}


//
// Run local executable file in remote machine 
//
func (c Client) RunFile(file string) (res Result, err error) {

	tmp := fmt.Sprintf("/tmp/%X", ksuid.New().String())

	if err = c.CopyFileToRemote(file, tmp); err != nil {
		return
	}

	res, err = c.Run(fmt.Sprintf("/usr/bin/env bash %s && rm -f %s", tmp, tmp))
	return
}

