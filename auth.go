// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Auth []ssh.AuthMethod

// Get auth method from raw password.
func Password(pass string) Auth {
	return Auth{
		ssh.Password(pass),
	}
}

// Get auth method from private key with or without passphrase.
func Key(prvFile string, passphrase string) Auth {

	signer, err := GetSigner(prvFile, passphrase)

	if err != nil {
		panic(err)
	}

	return Auth{
		ssh.PublicKeys(signer),
	}
}

func HasAgent() bool {
	return os.Getenv("SSH_AUTH_SOCK") != ""
}

// UseAgent auth via ssh agent, (Unix systems only)
func UseAgent() Auth {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		panic(fmt.Errorf("could not find ssh agent: %w", err))
	}
	return Auth{
		ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers),
	}
}

// Get private key signer.
func GetSigner(prvFile string, passphrase string) (ssh.Signer, error) {

	var (
		err    error
		signer ssh.Signer
	)

	privateKey, err := ioutil.ReadFile(prvFile)

	if err != nil {

		return nil, err

	} else if passphrase != "" {

		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(passphrase))

	} else {

		signer, err = ssh.ParsePrivateKey(privateKey)
	}

	return signer, err
}
