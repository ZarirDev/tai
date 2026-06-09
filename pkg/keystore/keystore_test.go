package keystore_test

import (
	"testing"

	"github.com/zarirdev/tai/pkg/keystore"
)

func TestEncryptDecrypt(t *testing.T) {
	original := "gsk_mySecretApiKey12345"
	enc, err := keystore.Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if enc == "" {
		t.Fatal("Encrypted string is empty")
	}
	if enc == original {
		t.Fatal("Encrypted string should not equal original")
	}

	dec, err := keystore.Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if dec != original {
		t.Errorf("Decrypted = %q; want %q", dec, original)
	}
}

func TestDecryptEmpty(t *testing.T) {
	dec, err := keystore.Decrypt("")
	if err != nil {
		t.Fatalf("Decrypt of empty string should not error: %v", err)
	}
	if dec != "" {
		t.Errorf("Decrypt of empty string should return empty, got %q", dec)
	}
}

func TestDecryptCorrupted(t *testing.T) {
	_, err := keystore.Decrypt("not-valid-base64")
	if err == nil {
		t.Fatal("Expected error for corrupted input")
	}
}

func TestEncryptDifferentOutputs(t *testing.T) {
	// Same plaintext should produce different ciphertexts due to random nonce
	p1, _ := keystore.Encrypt("hello")
	p2, _ := keystore.Encrypt("hello")
	if p1 == p2 {
		t.Errorf("Encrypt should produce different outputs each time (nonce)")
	}
}
