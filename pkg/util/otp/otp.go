package otp

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var (
	ErrInvalidLength = errors.New("OTP length must be between 4 and 10")
	ErrMismatch      = errors.New("OTP does not match")
)

const (
	DefaultLength = 6
	MinLength     = 4
	MaxLength     = 10
)

// Generate creates a cryptographically secure numeric OTP of the specified length.
// Length must be between 4 and 10 digits.
func Generate(length int) (string, error) {
	if length < MinLength || length > MaxLength {
		return "", ErrInvalidLength
	}

	// Calculate the maximum value for the given length (e.g., 999999 for 6 digits)
	max := new(big.Int)
	max.Exp(big.NewInt(10), big.NewInt(int64(length)), nil)

	// Generate a random number in [0, max)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	// Format with leading zeros
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n), nil
}

// GenerateDefault creates a 6-digit OTP.
func GenerateDefault() (string, error) {
	return Generate(DefaultLength)
}

// Hash creates a SHA-256 hash of the OTP code.
// The hash is returned as a hex-encoded string.
func Hash(code string) string {
	// Normalize: trim whitespace
	code = strings.TrimSpace(code)

	h := sha256.New()
	h.Write([]byte(code))
	return hex.EncodeToString(h.Sum(nil))
}

// Verify compares a plaintext OTP code against a hash.
// Returns nil if they match, ErrMismatch if they don't.
func Verify(hash, code string) error {
	// Normalize: trim whitespace
	code = strings.TrimSpace(code)

	computedHash := Hash(code)

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(hash), []byte(computedHash)) != 1 {
		return ErrMismatch
	}

	return nil
}

// GenerateAlphanumeric creates a cryptographically secure alphanumeric code.
// Useful for verification tokens, invitation codes, etc.
// Length specifies the number of characters.
func GenerateAlphanumeric(length int) (string, error) {
	if length < 1 {
		return "", errors.New("length must be at least 1")
	}

	// Character set: uppercase + digits, excluding ambiguous characters (0, O, I, 1, L)
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

	result := make([]byte, length)
	max := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("failed to generate random character: %w", err)
		}
		result[i] = charset[n.Int64()]
	}

	return string(result), nil
}

// GenerateHex creates a cryptographically secure hex string.
// byteLength specifies the number of random bytes (output will be 2x this length).
func GenerateHex(byteLength int) (string, error) {
	if byteLength < 1 {
		return "", errors.New("byte length must be at least 1")
	}

	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return hex.EncodeToString(b), nil
}
