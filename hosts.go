// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"os"
	"strings"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Use default known hosts files to verify host public key.
func DefaultKnownHosts() (ssh.HostKeyCallback, error) {

	return KnownHosts(strings.Join([]string{os.Getenv("HOME"), ".ssh", "known_hosts"}, "/"))
}

// Get known hosts callback from a custom path.
func KnownHosts(kh string) (ssh.HostKeyCallback, error) {

	return knownhosts.New(kh)
}
