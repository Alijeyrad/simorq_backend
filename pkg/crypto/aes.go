// Package crypto provides AES-256-GCM encryption/decryption helpers for
// storing sensitive fields (national_id, IBAN) encrypted at rest.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKey         = errors.New("encryption key must be 32 bytes")
	ErrCiphertextTooShort = errors.New("ciphertext too short")
)

// KeyFromHex decodes a 64-char hex string into a 32-byte AES-256 key.
func KeyFromHex(hexKey string) ([]byte, error) {
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}
	if len(b) != 32 {
		return nil, ErrInvalidKey
	}
	return b, nil
}

// Encrypt encrypts plaintext using AES-256-GCM with a random nonce.
// Returns a base64-encoded string: nonce || ciphertext.
func Encrypt(key []byte, plaintext string) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidKey
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded AES-256-GCM ciphertext produced by Encrypt.
func Decrypt(key []byte, encoded string) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidKey
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	if len(data) < gcm.NonceSize() {
		return "", ErrCiphertextTooShort
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// Hash returns the SHA-256 hex digest of value.
// Used for national_id_hash: deterministic, allows indexed uniqueness lookups
// without storing the plaintext.
func Hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
