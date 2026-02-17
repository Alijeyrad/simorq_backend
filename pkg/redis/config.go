package redis

import (
	"time"

	"github.com/Alijeyrad/simorq_backend/config"
)

// Config holds Redis connection settings
type Config struct {
	Addr     string
	DB       int
	Username string
	Password string

	// Connection pool settings
	PoolSize     int
	MinIdleConns int

	// Timeouts
	DialTimeoutSeconds  int
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
}

// DefaultConfig returns sensible defaults for Redis configuration
func DefaultConfig() Config {
	return Config{
		Addr:                "localhost:6379",
		DB:                  0,
		PoolSize:            10,
		MinIdleConns:        2,
		DialTimeoutSeconds:  5,
		ReadTimeoutSeconds:  3,
		WriteTimeoutSeconds: 3,
	}
}

// DialTimeout returns the dial timeout as a duration
func (c Config) DialTimeout() time.Duration {
	if c.DialTimeoutSeconds <= 0 {
		return 5 * time.Second
	}
	return time.Duration(c.DialTimeoutSeconds) * time.Second
}

// ReadTimeout returns the read timeout as a duration
func (c Config) ReadTimeout() time.Duration {
	if c.ReadTimeoutSeconds <= 0 {
		return 3 * time.Second
	}
	return time.Duration(c.ReadTimeoutSeconds) * time.Second
}

// WriteTimeout returns the write timeout as a duration
func (c Config) WriteTimeout() time.Duration {
	if c.WriteTimeoutSeconds <= 0 {
		return 3 * time.Second
	}
	return time.Duration(c.WriteTimeoutSeconds) * time.Second
}

// FromCentralConfig converts central config.RedisConfig to package Config
func FromCentralConfig(c config.RedisConfig) Config {
	cfg := Config{
		Addr:     c.Addr,
		DB:       c.DB,
		Username: c.Username,
		Password: c.Password,
	}

	// Set pool size if configured, otherwise use default
	if c.PoolSize > 0 {
		cfg.PoolSize = c.PoolSize
	} else {
		cfg.PoolSize = DefaultConfig().PoolSize
	}

	// Set min idle conns if configured, otherwise use default
	if c.MinIdleConns > 0 {
		cfg.MinIdleConns = c.MinIdleConns
	} else {
		cfg.MinIdleConns = DefaultConfig().MinIdleConns
	}

	// Set dial timeout if configured, otherwise use default
	if c.DialTimeoutSeconds > 0 {
		cfg.DialTimeoutSeconds = c.DialTimeoutSeconds
	} else {
		cfg.DialTimeoutSeconds = DefaultConfig().DialTimeoutSeconds
	}

	// Set read timeout if configured, otherwise use default
	if c.ReadTimeoutSeconds > 0 {
		cfg.ReadTimeoutSeconds = c.ReadTimeoutSeconds
	} else {
		cfg.ReadTimeoutSeconds = DefaultConfig().ReadTimeoutSeconds
	}

	// Set write timeout if configured, otherwise use default
	if c.WriteTimeoutSeconds > 0 {
		cfg.WriteTimeoutSeconds = c.WriteTimeoutSeconds
	} else {
		cfg.WriteTimeoutSeconds = DefaultConfig().WriteTimeoutSeconds
	}

	return cfg
}
