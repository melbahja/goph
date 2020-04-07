// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"io"
	"os"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Upload local file to remote.
func Upload(c *ssh.Client, src string, dest string) (err error) {

	client, err := sftp.NewClient(c)

	if err != nil {
		return
	}

	defer client.Close()

	srcFile, err := os.Open(src)

	if err != nil {
		return
	}

	defer srcFile.Close()

	destFile, err := client.Create(dest)

	if err != nil {
		return
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return
}

// Download remote file to local.
func Download(c *ssh.Client, src string, dest string) (err error) {

	client, err := sftp.NewClient(c)

	if err != nil {
		return
	}

	destFile, err := os.Create(dest)

	if err != nil {
		return
	}

	defer destFile.Close()

	defer client.Close()

	srcFile, err := client.Open(src)

	if err != nil {
		return
	}

	defer srcFile.Close()

	if _, err = io.Copy(destFile, srcFile); err != nil {
		return
	}

	return destFile.Sync()
}
