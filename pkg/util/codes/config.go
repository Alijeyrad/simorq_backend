package codes

import "github.com/Alijeyrad/simorq_backend/config"

// Config holds settings for various code generation utilities
type Config struct {

	// TokenByteLength is the number of random bytes for tokens
	TokenByteLength int

	// URLSafeTokens determines whether to use URL-safe base64 encoding
	URLSafeTokens bool

	// Charset is the character set used for alphanumeric codes
	// If empty, defaults to mixed case alphanumeric without ambiguous chars
	Charset string
}

// DefaultConfig returns sensible defaults for code generation
func DefaultConfig() Config {
	return Config{
		TokenByteLength: 16,
		URLSafeTokens:   true,
		Charset:         "ABCDEFGHJKMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789",
	}
}

// GetCharset returns the configured charset or the default if empty
func (c Config) GetCharset() string {
	if c.Charset == "" {
		return "ABCDEFGHJKMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"
	}
	return c.Charset
}

// FromCentralConfig converts central config.CodesConfig to package Config
func FromCentralConfig(c config.CodesConfig) Config {
	return Config{
		TokenByteLength: c.TokenByteLength,
		URLSafeTokens:   c.URLSafeTokens,
		Charset:         c.Charset,
	}
}
