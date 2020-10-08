// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"errors"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var defaultPath = os.ExpandEnv("$HOME/.ssh/known_hosts")

// Use default known hosts files to verify host public key.
func DefaultKnownHosts() (ssh.HostKeyCallback, error) {
	return KnownHosts(defaultPath)
}

// Get known hosts callback from a custom path.
func KnownHosts(file string) (ssh.HostKeyCallback, error) {
	return knownhosts.New(file)
}

// Check is host in known hosts file.
// It returns is the host found in known_hosts file and error,
// If the host found in known_hosts file and error not nil that means public key mismatch, Maybe
// MAN IN THE MIDDLE ATTACK! you should not handshake.
func CheckKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) (found bool, err error) {

	var keyErr *knownhosts.KeyError

	// Fallback to default known_hosts file
	if knownFile == "" {
		knownFile = defaultPath
	}

	// Get host key callback
	callback, err := KnownHosts(knownFile)

	if err != nil {
		return false, err
	}

	// check if host already exists.
	err = callback(host, remote, key)

	// Known host already exists.
	if err == nil {
		return true, nil
	}

	// Make sure that the error returned from the callback is host not in file error.
	// If keyErr.Want is greater than 0 length, that means host is in file with different key.
	if errors.As(err, &keyErr) && len(keyErr.Want) > 0 {
		return true, keyErr
	}

	// Some other error occured and safest way to handle is to pass it back to user.
	if err != nil {
		return false, err
	}

	// Key is not trusted because it is not in the file.
	return false, nil
}

// Add a host to knows hosts
func AddKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) (err error) {

	// Fallback to default known_hosts file
	if knownFile == "" {
		knownFile = defaultPath
	}

	f, err := os.OpenFile(knownFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)

	if err != nil {
		return err
	}

	defer f.Close()

	knownHost := knownhosts.Normalize(remote.String())

	_, err = f.WriteString(knownhosts.Line([]string{knownHost}, key) + "\n")

	return err
}
