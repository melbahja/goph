// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Auth represents ssh auth methods.
type Auth []ssh.AuthMethod

// Password returns password auth method.
func Password(pass string) Auth {
	return Auth{
		ssh.Password(pass),
	}
}

// KeyboardInteractive returns password keyboard interactive auth method as fallback of password auth method.
func KeyboardInteractive(pass string) Auth {
	return Auth{
		ssh.Password(pass),
		ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			for _, q := range questions {
				if strings.Contains(strings.ToLower(q), "password") {
					answers = append(answers, pass)
				} else {
					answers = append(answers, "")
				}
			}
			return answers, nil
		}),
	}
}

// Key returns auth method from private key with or without passphrase.
func Key(prvFile string, passphrase string) (Auth, error) {

	signer, err := GetSigner(prvFile, passphrase)

	if err != nil {
		return nil, err
	}

	return Auth{
		ssh.PublicKeys(signer),
	}, nil
}

// KeyContent returns auth method from private key content (byte slice) with optional passphrase.
// This is useful when the private key is stored in memory rather than a file.
func KeyContent(privateKey []byte, passphrase string) (Auth, error) {
	signer, err := GetSignerForRawKey(privateKey, passphrase)
	if err != nil {
		return nil, err
	}

	return Auth{
		ssh.PublicKeys(signer),
	}, nil
}

// RawKey returns auth method from private key string content with optional passphrase.
// Deprecated: Use KeyContent instead which accepts []byte directly.
func RawKey(privateKey string, passphrase string) (Auth, error) {
	return KeyContent([]byte(privateKey), passphrase)
}

// HasAgent checks if ssh agent exists.
func HasAgent() bool {
	return os.Getenv("SSH_AUTH_SOCK") != ""
}

// UseAgent auth via ssh agent, (Unix systems only)
func UseAgent() (Auth, error) {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, fmt.Errorf("could not find ssh agent: %w", err)
	}
	return Auth{
		ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers),
	}, nil
}

// GetSigner returns ssh signer from private key file.
func GetSigner(prvFile string, passphrase string) (ssh.Signer, error) {
	if isRawPrivateKey(prvFile) {
		return GetSignerForRawKey([]byte(prvFile), passphrase)
	}

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

func isRawPrivateKey(input string) bool {
	trimmed := strings.TrimSpace(input)
	return strings.HasPrefix(trimmed, "-----BEGIN") && strings.Contains(trimmed, "PRIVATE KEY-----")
}

// GetSignerForRawKey returns ssh signer from private key file.
func GetSignerForRawKey(privateKey []byte, passphrase string) (ssh.Signer, error) {

	var (
		err    error
		signer ssh.Signer
	)

	if passphrase != "" {

		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(passphrase))

	} else {

		signer, err = ssh.ParsePrivateKey(privateKey)
	}

	return signer, err
}
