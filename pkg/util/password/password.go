package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("invalid password hash format")
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
	ErrMismatch            = errors.New("password does not match")
)

// Params defines the Argon2id parameters.
type Params struct {
	Memory      uint32 // Memory in KiB
	Iterations  uint32 // Number of iterations
	Parallelism uint8  // Degree of parallelism
	SaltLength  uint32 // Length of salt in bytes
	KeyLength   uint32 // Length of generated key in bytes
}

// DefaultParams returns secure default parameters for Argon2id.
// These are based on OWASP recommendations for password storage.
func DefaultParams() *Params {
	return &Params{
		Memory:      64 * 1024, // 64 MiB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// LowMemoryParams returns parameters suitable for memory-constrained environments.
func LowMemoryParams() *Params {
	return &Params{
		Memory:      32 * 1024, // 32 MiB
		Iterations:  4,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
}

var defaultParams = DefaultParams()

// Hash generates an Argon2id hash of the password using default parameters.
func Hash(password string) (string, error) {
	return HashWithParams(password, defaultParams)
}

// HashWithParams generates an Argon2id hash of the password using custom parameters.
func HashWithParams(password string, p *Params) (string, error) {
	if p == nil {
		p = defaultParams
	}

	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		p.Iterations,
		p.Memory,
		p.Parallelism,
		p.KeyLength,
	)

	// Encode to PHC string format:
	// $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.Memory,
		p.Iterations,
		p.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encoded, nil
}

// Verify compares a password against an Argon2id hash.
// Returns nil if they match, ErrMismatch if they don't, or another error if the hash is invalid.
func Verify(hash, password string) error {
	p, salt, hashBytes, err := decodeHash(hash)
	if err != nil {
		return err
	}

	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		p.Iterations,
		p.Memory,
		p.Parallelism,
		p.KeyLength,
	)

	if subtle.ConstantTimeCompare(hashBytes, otherHash) != 1 {
		return ErrMismatch
	}

	return nil
}

// NeedsRehash checks if a hash was created with outdated parameters.
// Returns true if the hash should be regenerated with current default parameters.
func NeedsRehash(hash string) bool {
	p, _, _, err := decodeHash(hash)
	if err != nil {
		return true
	}

	return p.Memory != defaultParams.Memory ||
		p.Iterations != defaultParams.Iterations ||
		p.Parallelism != defaultParams.Parallelism ||
		p.KeyLength != defaultParams.KeyLength
}

// Generate creates a random password of the specified length.
// Uses URL-safe base64 characters (a-z, A-Z, 0-9, -, _).
func Generate(length int) string {
	if length <= 0 {
		length = 16
	}

	// Generate enough random bytes to produce the desired length after base64 encoding
	byteLen := (length*6 + 7) / 8
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("failed to generate random password: %w", err))
	}

	encoded := base64.RawURLEncoding.EncodeToString(b)
	if len(encoded) > length {
		return encoded[:length]
	}
	return encoded
}

// Match is a convenience wrapper that returns true if password matches hash.
func Match(hash, password string) bool {
	return Verify(hash, password) == nil
}

func decodeHash(encodedHash string) (*Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	var p Params
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	p.SaltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	p.KeyLength = uint32(len(hash))

	return &p, salt, hash, nil
}
