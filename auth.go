// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"gopkg.in/jcmturner/gokrb5.v7/client"
	"gopkg.in/jcmturner/gokrb5.v7/config"
	"gopkg.in/jcmturner/gokrb5.v7/keytab"
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

func RawKey(privateKey string, passphrase string) (Auth, error) {
	signer, err := GetSignerForRawKey([]byte(privateKey), passphrase)
	if err != nil {
		return nil, err
	}

	return Auth{
		ssh.PublicKeys(signer),
	}, nil
}

// KerberosWithPassword returns a kerberos auth with password.
func KerberosWithPassword(username, password, realm, target string, krb5cfg io.Reader) (Auth, error) {
	cfg, err := config.NewConfigFromReader(krb5cfg)
	if err != nil {
		return nil, err
	}

	cl := client.NewClientWithPassword(username, realm, password, cfg, client.DisablePAFXFAST(true))
	c, err := newKrb5Client(cl)
	if err != nil {
		return nil, err
	}
	return Auth{
		ssh.GSSAPIWithMICAuthMethod(c, target),
	}, nil
}

// KerberosWithKeytab returns a kerberos auth with keytab.
func KerberosWithKeytab(username, keytabFile, realm, target string, krb5cfg io.Reader) (Auth, error) {
	kt, err := keytab.Load(keytabFile)
	if err != nil {
		return nil, err
	}
	cfg, err := config.NewConfigFromReader(krb5cfg)
	if err != nil {
		return nil, err
	}

	cl := client.NewClientWithKeytab(username, realm, kt, cfg, client.DisablePAFXFAST(true))
	c, err := newKrb5Client(cl)
	if err != nil {
		return nil, err
	}
	return Auth{
		ssh.GSSAPIWithMICAuthMethod(c, target),
	}, nil
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
