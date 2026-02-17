package password

import "github.com/Alijeyrad/simorq_backend/config"

// Config holds Argon2id password hashing parameters
type Config struct {
	// Algorithm should be "argon2id" (default and recommended)
	Algorithm string

	// Memory usage in KiB (64 MiB default, OWASP recommended)
	MemoryKiB uint32

	// Number of iterations (3 default, OWASP recommended)
	Iterations uint32

	// Degree of parallelism (2 default, OWASP recommended)
	Parallelism uint8

	// Length of random salt in bytes (16 default)
	SaltLength uint32

	// Length of derived key in bytes (32 default)
	KeyLength uint32

	// LowMemoryMode reduces memory to 32 MiB for constrained environments
	LowMemoryMode bool
}

// ToParams converts Config to Params for the password package
func (c Config) ToParams() *Params {
	memory := c.MemoryKiB
	if c.LowMemoryMode && memory > 32*1024 {
		memory = 32 * 1024 // 32 MiB
	}

	return &Params{
		Memory:      memory,
		Iterations:  c.Iterations,
		Parallelism: c.Parallelism,
		SaltLength:  c.SaltLength,
		KeyLength:   c.KeyLength,
	}
}

// DefaultConfig returns OWASP-recommended defaults for password hashing
func DefaultConfig() Config {
	return Config{
		Algorithm:     "argon2id",
		MemoryKiB:     64 * 1024, // 64 MiB
		Iterations:    3,
		Parallelism:   2,
		SaltLength:    16,
		KeyLength:     32,
		LowMemoryMode: false,
	}
}

// LowMemoryConfig returns parameters for memory-constrained environments
func LowMemoryConfig() Config {
	return Config{
		Algorithm:     "argon2id",
		MemoryKiB:     32 * 1024, // 32 MiB
		Iterations:    4,         // Increase iterations to compensate
		Parallelism:   2,
		SaltLength:    16,
		KeyLength:     32,
		LowMemoryMode: true,
	}
}

// FromCentralConfig converts central config.PasswordConfig to package Config
func FromCentralConfig(c config.PasswordConfig) Config {
	return Config{
		Algorithm:     c.Algorithm,
		MemoryKiB:     c.MemoryKiB,
		Iterations:    c.Iterations,
		Parallelism:   c.Parallelism,
		SaltLength:    c.SaltLength,
		KeyLength:     c.KeyLength,
		LowMemoryMode: c.LowMemoryMode,
	}
}
