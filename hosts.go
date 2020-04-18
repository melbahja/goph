// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"os"
	"net"
	"errors"

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

// Add a host to knows hosts
// this function by @dixonwille see: https://github.com/melbahja/goph/issues/2
func AddKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) error {

	var (
		hostErr error
		fileErr error
		keyErr  *knownhosts.KeyError
	)

	// Fallback to default known_hosts file
	if knownFile == "" {
		knownFile = defaultPath
	}

	_, fileErr = os.Stat(knownFile)

	// Create if not exists
	if errors.Is(fileErr, os.ErrNotExist) {

		f, fileErr := os.OpenFile(knownFile, os.O_CREATE, 0600)

		if fileErr != nil {
			return fileErr
		}

		f.Close()
	}

	// Get host key callback
	callback, hostErr := KnownHosts(knownFile)

	if hostErr != nil {
		return hostErr
	}

	// check if host already exists
	hostErr = callback(host, remote, key)

	// Known host already exists
	if hostErr == nil {
		return nil
	}

	// Append new host
	if errors.As(hostErr, &keyErr) && len(keyErr.Want) == 0 {

		f, fileErr := os.OpenFile(knownFile, os.O_WRONLY|os.O_APPEND, 0600)

		if fileErr != nil {
			return fileErr
		}

		defer f.Close()

		knownHost := knownhosts.Normalize(remote.String())

		_, fileErr = f.WriteString(knownhosts.Line([]string{knownHost}, key) + "\n")

		if fileErr != nil {

			return fileErr
		}

		return nil
	}

	return hostErr
}
