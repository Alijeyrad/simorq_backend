package otp

import "github.com/Alijeyrad/simorq_backend/config"

// Config holds OTP generation settings
type Config struct {
	// DefaultLength is the default OTP length (typically 6)
	DefaultLength int

	// MinLength is the minimum allowed OTP length
	MinLength int

	// MaxLength is the maximum allowed OTP length
	MaxLength int

	// HashAlgorithm specifies the hashing algorithm (e.g., "sha256")
	HashAlgorithm string
}

// DefaultConfig returns sensible defaults for OTP generation
func DefaultConfig() Config {
	return Config{
		DefaultLength: 6,
		MinLength:     4,
		MaxLength:     10,
		HashAlgorithm: "sha256",
	}
}

// Validate checks if the config values are valid
func (c Config) Validate() error {
	if c.DefaultLength < c.MinLength || c.DefaultLength > c.MaxLength {
		return ErrInvalidLength
	}
	if c.MinLength < 1 {
		return ErrInvalidLength
	}
	if c.MaxLength < c.MinLength {
		return ErrInvalidLength
	}
	return nil
}

// FromCentralConfig converts central config.OTPConfig to package Config
func FromCentralConfig(c config.OTPConfig) Config {
	return Config{
		DefaultLength: c.DefaultLength,
		MinLength:     c.MinLength,
		MaxLength:     c.MaxLength,
		HashAlgorithm: c.HashAlgorithm,
	}
}
