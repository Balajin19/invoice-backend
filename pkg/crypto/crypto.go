package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// key returns the 32-byte AES-256 key from the ACCOUNT_NUMBER_KEY env variable.
// The env value must be a 64-character hex string (32 bytes).
func key() ([]byte, error) {
	raw := strings.TrimSpace(os.Getenv("ACCOUNT_NUMBER_KEY"))
	if raw == "" {
		return nil, errors.New("ACCOUNT_NUMBER_KEY env variable is not set")
	}
	k, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("ACCOUNT_NUMBER_KEY is not valid hex: %w", err)
	}
	if len(k) != 32 {
		return nil, fmt.Errorf("ACCOUNT_NUMBER_KEY must be 32 bytes (64 hex chars), got %d bytes", len(k))
	}
	return k, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns a hex-encoded string
// in the format "<nonce_hex>:<ciphertext_hex>".
func Encrypt(plaintext string) (string, error) {
	k, err := key()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(nonce) + ":" + hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a value produced by Encrypt.
func Decrypt(encoded string) (string, error) {
	k, err := key()
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(encoded, ":", 2)
	if len(parts) != 2 {
		return "", errors.New("invalid encrypted format: expected nonce:ciphertext")
	}

	nonce, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid nonce hex: %w", err)
	}

	ciphertext, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext hex: %w", err)
	}

	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}
