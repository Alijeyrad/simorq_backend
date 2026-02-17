package codes

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var (
	ErrInvalidLength = errors.New("invalid code length")
)

const (
	// ReferralCodeLength is the length of generated referral codes
	ReferralCodeLength = 16

	// InvitationCodeLength is the length of generated invitation codes
	InvitationCodeLength = 16

	// ReferralCodeByteLength is the number of random bytes for referral codes (produces 16 base64 chars)
	ReferralCodeByteLength = 12

	// InvitationCodeByteLength is the number of random bytes for invitation codes (produces 16 base64 chars)
	InvitationCodeByteLength = 12

	// TokenByteLength is the number of random bytes for tokens (produces 32 hex chars)
	TokenByteLength = 16

	// Mixed case alphanumeric excluding ambiguous characters
	charsetMixedAlphanumeric = "ABCDEFGHJKMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"
)

// GenerateReferralCode creates a unique referral code.
// Format: 16-character base64 URL-safe string (e.g., "aBcDeFgHiJkLmNoP")
func GenerateReferralCode() (string, error) {
	return GenerateURLSafeToken(ReferralCodeByteLength)
}

// GenerateInvitationCode creates a unique invitation code.
// Format: 16-character base64 URL-safe string (e.g., "XyZaBcDeFgHiJkLm")
func GenerateInvitationCode() (string, error) {
	return GenerateURLSafeToken(InvitationCodeByteLength)
}

// GenerateInvitationToken creates a secure token for invitation URLs.
// Returns a 32-character hex string.
func GenerateInvitationToken() (string, error) {
	return GenerateSecureToken(TokenByteLength)
}

// GenerateVerificationToken creates a secure token for email verification URLs.
// Returns a 32-character hex string.
func GenerateVerificationToken() (string, error) {
	return GenerateSecureToken(TokenByteLength)
}

// GenerateSecureToken creates a cryptographically secure hex token.
// byteLength specifies the number of random bytes (output will be 2x this length in hex).
func GenerateSecureToken(byteLength int) (string, error) {
	if byteLength < 1 {
		return "", ErrInvalidLength
	}

	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return hex.EncodeToString(b), nil
}

// GenerateURLSafeToken creates a URL-safe base64-encoded token.
// byteLength specifies the number of random bytes.
func GenerateURLSafeToken(byteLength int) (string, error) {
	if byteLength < 1 {
		return "", ErrInvalidLength
	}

	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateCode creates a code of specified length from a given character set.
func GenerateCode(length int, charset string) (string, error) {
	if length < 1 {
		return "", ErrInvalidLength
	}
	if len(charset) == 0 {
		return "", errors.New("charset cannot be empty")
	}

	return generateFromCharset(length, charset)
}

// GenerateNumericCode creates a numeric-only code of specified length.
func GenerateNumericCode(length int) (string, error) {
	if length < 1 {
		return "", ErrInvalidLength
	}

	max := new(big.Int)
	max.Exp(big.NewInt(10), big.NewInt(int64(length)), nil)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n), nil
}

// NormalizeCode normalizes a code for comparison (uppercase, trim whitespace).
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// FormatCode formats a code with dashes for readability.
// e.g., "ABCD1234" -> "ABCD-1234" with groupSize=4
func FormatCode(code string, groupSize int) string {
	if groupSize < 1 || len(code) <= groupSize {
		return code
	}

	var parts []string
	for i := 0; i < len(code); i += groupSize {
		end := i + groupSize
		if end > len(code) {
			end = len(code)
		}
		parts = append(parts, code[i:end])
	}

	return strings.Join(parts, "-")
}

// ParseCode removes formatting (dashes, spaces) from a code.
func ParseCode(formatted string) string {
	code := strings.ReplaceAll(formatted, "-", "")
	code = strings.ReplaceAll(code, " ", "")
	return strings.ToUpper(strings.TrimSpace(code))
}

func generateFromCharset(length int, charset string) (string, error) {
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
