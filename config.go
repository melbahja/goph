// Copyright 2020 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license,
// license that can be found in the LICENSE file.

package ssh

import (
	"net"
	"fmt"
	gossh "golang.org/x/crypto/ssh"
)

type Config struct {
	Net string
	Env Env
	Addr string
	Port int
	User string
	Auth Auth
	Config ClientConfig
}


func (d Config) GetNet() string {

	if d.Net == "" {
		return TCP
	}

	return 	d.Net
}

func (d Config) GetAddr() string {

	port := 22

	if d.Port != 0 {
		port = d.Port
	}

	return fmt.Sprintf("%s:%d", d.Addr, port)
}


func (d Config) GetEnv() Env {
	return d.Env
}

func (d Config) GetClientConfig() ClientConfig {

	if d.Config == nil {
		d.Config = &gossh.ClientConfig{}
	}

	d.Config.User = d.User
	d.Config.Auth = d.Auth

	if d.Config.HostKeyCallback == nil {
		d.Config.HostKeyCallback = gossh.HostKeyCallback(func(h string, r net.Addr, k gossh.PublicKey) error {
			return nil
		})
	}

	return d.Config
}
