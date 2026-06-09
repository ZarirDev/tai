package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
)

// deriveKey returns a 32-byte AES key based on the machine ID and username.
func deriveKey() ([]byte, error) {
	var seed string

	// Try /etc/machine-id (Linux)
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		seed = strings.TrimSpace(string(data))
	} else {
		// fallback to hostname
		host, _ := os.Hostname()
		seed = host
	}

	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}

	combined := seed + ":" + user + ":tai-salt-2024"
	hash := sha256.Sum256([]byte(combined))
	return hash[:], nil
}

// Encrypt encrypts plaintext using AES-GCM and returns a base64-encoded string.
func Encrypt(plaintext string) (string, error) {
	key, err := deriveKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt takes a base64-encoded encrypted string and returns the plaintext.
func Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}

	key, err := deriveKey()
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
