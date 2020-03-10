package ssh

import (
	"io/ioutil"
	gossh "golang.org/x/crypto/ssh"
)

const (
	UDP string = "udp"
	TCP string = "tcp"
)

type Network string

type Auth []gossh.AuthMethod

type ClientConfig *gossh.ClientConfig

type Env map[string]string

type ConnectionConfig interface {

	GetNet() string

	GetAddr() string

	GetEnv() Env

	GetClientConfig() ClientConfig
}


func New(c ConnectionConfig) (*Client, error) {

	client, err := gossh.Dial(c.GetNet(), c.GetAddr(), c.GetClientConfig())

	if err != nil {
		return nil, err
	}

	return &Client{
		Client: client,
		Conn: c,
	}, nil
}


func Password(pass string) Auth {
	return Auth{
		gossh.Password(pass),
	}
}

func Key(path string, passphrase string) Auth {

	var (
		err error
		signer gossh.Signer
	)

	privateKey, err := ioutil.ReadFile(path)

	if err != nil {
	
		panic(err)
	
	} else if passphrase != "" {

		signer, err = gossh.ParsePrivateKeyWithPassphrase(privateKey, []byte(passphrase))
	
	} else {

		signer, err = gossh.ParsePrivateKey(privateKey)
	}

	if err != nil {
		panic(err)
	}

	return Auth{
		gossh.PublicKeys(signer),
	}
}
