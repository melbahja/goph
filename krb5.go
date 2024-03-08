// Copyright 2024 Mohammed El Bahja. All rights reserved.
// Use of this source code is governed by a MIT license.

package goph

import (
	"strings"

	"gopkg.in/jcmturner/gokrb5.v7/client"
	"gopkg.in/jcmturner/gokrb5.v7/gssapi"
	"gopkg.in/jcmturner/gokrb5.v7/spnego"
	"gopkg.in/jcmturner/gokrb5.v7/types"
)

type krb5Client struct {
	client *client.Client
	skey   types.EncryptionKey
	gen    bool
}

func newKrb5Client(c *client.Client) (*krb5Client, error) {
	if err := c.Login(); err != nil {
		return nil, err
	}
	return &krb5Client{client: c}, nil
}

func (c *krb5Client) InitSecContext(target string, token []byte, isGSSDelegCreds bool) (outputToken []byte, needContinue bool, err error) {
	if c.gen {
		return nil, false, nil
	}

	t := strings.Replace(target, "@", "/", 1)
	tkt, skey, err := c.client.GetServiceTicket(t)
	if err != nil {
		return nil, false, err
	}
	c.skey = skey

	gssApiFlags := []int{gssapi.ContextFlagInteg, gssapi.ContextFlagMutual}
	if isGSSDelegCreds {
		gssApiFlags = append(gssApiFlags, gssapi.ContextFlagDeleg)
	}

	krb5Tkn, err := spnego.NewKRB5TokenAPREQ(c.client, tkt, skey, gssApiFlags, nil)
	if err != nil {
		return nil, false, err
	}

	outputToken, err = krb5Tkn.Marshal()
	if err != nil {
		return nil, false, err
	}
	c.gen = true

	return outputToken, true, nil
}

func (c *krb5Client) GetMIC(micFiled []byte) ([]byte, error) {
	micTkn, err := gssapi.NewInitiatorMICToken(micFiled, c.skey)
	if err != nil {
		return nil, err
	}
	return micTkn.Marshal()
}

func (c *krb5Client) DeleteSecContext() error {
	c.client.Destroy()
	return nil
}
