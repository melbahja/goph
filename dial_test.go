package goph

import (
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestDialSetsDefaultUser(t *testing.T) {
	c := &Client{User: "admin", Addr: "invalid.invalid.invalid", Port: 22}
	config := &ssh.ClientConfig{}

	// Dial will fail because the address is invalid
	// but it should still set config.User first.
	_ = Dial(c, config)

	if config.User != "admin" {
		t.Fatalf("expected config.User to be 'admin', got %q", config.User)
	}
}

func TestDialDefaultsToCurrentUser(t *testing.T) {
	c := &Client{Addr: "invalid.invalid.invalid", Port: 22}
	config := &ssh.ClientConfig{}

	_ = Dial(c, config)

	if config.User == "" {
		t.Fatal("expected config.User to be set to current OS user")
	}
}
