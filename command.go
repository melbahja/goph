package ssh


import (
	"io"
	"fmt"
	"strings"
	gossh "golang.org/x/crypto/ssh"
)

type Command struct {

	// Working directory
	Dir string

	Command string
	Args []string
	Client *Client
	Stdin io.Reader
}


func (cmd Command) Run() (res Result, err error) {
	
	var sess *gossh.Session

	if sess, err = cmd.Client.Client.NewSession(); err != nil {
		return
	}

	defer sess.Close()

	sess.Stdin = cmd.Stdin
	sess.Stdout = &res.Stdout
	sess.Stderr = &res.Stderr

	if cmd.Dir == "" {
		cmd.Dir = "~"
	}

	err = sess.Run(fmt.Sprintf("cd %s && %s %s", cmd.Dir, cmd.Command, strings.Join(cmd.Args, " ")))
	return
}
