package goph

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func generateTestPrivateKeyPEM(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}

	return string(pem.EncodeToMemory(block))
}

func TestGetSignerFromRawPrivateKeyContent(t *testing.T) {
	privateKey := generateTestPrivateKeyPEM(t)

	signer, err := GetSigner(privateKey, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if signer == nil {
		t.Fatalf("expected non-nil signer")
	}
}

func TestGetSignerFromMissingPrivateKeyFile(t *testing.T) {
	_, err := GetSigner(filepath.Join(t.TempDir(), "missing-id-rsa"), "")
	if err == nil {
		t.Fatalf("expected error for missing key file")
	}
}

func TestKeyFromRawPrivateKeyContent(t *testing.T) {
	privateKey := generateTestPrivateKeyPEM(t)

	auth, err := Key(privateKey, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(auth) == 0 {
		t.Fatalf("expected at least one auth method")
	}
}

func TestGetSignerFromPrivateKeyFile(t *testing.T) {
	privateKey := generateTestPrivateKeyPEM(t)
	keyPath := filepath.Join(t.TempDir(), "id_rsa")
	if err := os.WriteFile(keyPath, []byte(privateKey), 0o600); err != nil {
		t.Fatalf("failed to write private key file: %v", err)
	}

	signer, err := GetSigner(keyPath, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if signer == nil {
		t.Fatalf("expected non-nil signer")
	}
}
