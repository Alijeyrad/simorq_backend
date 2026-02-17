package database

import (
	"time"

	"github.com/Alijeyrad/simorq_backend/config"
)

// Config holds database connection and behavior settings
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string

	// Connection pooling
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetimeMin int

	// Migration control
	AutoMigrate bool
	SafeMode    bool

	// Query logging
	EnableLogging        bool
	SlowQueryThresholdMs int
}

// DSN returns a PostgreSQL connection string
func (c Config) DSN() string {
	return buildDSN(c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// ConnMaxLifetime returns the connection max lifetime as a duration
func (c Config) ConnMaxLifetime() time.Duration {
	if c.ConnMaxLifetimeMin <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(c.ConnMaxLifetimeMin) * time.Minute
}

// DefaultConfig returns sensible defaults for database configuration
func DefaultConfig() Config {
	return Config{
		Host:                 "localhost",
		Port:                 5432,
		SSLMode:              "disable",
		MaxOpenConns:         25,
		MaxIdleConns:         5,
		ConnMaxLifetimeMin:   5,
		AutoMigrate:          false,
		SafeMode:             true,
		EnableLogging:        false,
		SlowQueryThresholdMs: 200,
	}
}

// FromCentralConfig converts central config.DatabaseConfig to package Config
func FromCentralConfig(c config.DatabaseConfig) Config {
	return Config{
		Host:                 c.Host,
		Port:                 c.Port,
		User:                 c.User,
		Password:             c.Password,
		DBName:               c.DBName,
		SSLMode:              c.SSLMode,
		MaxOpenConns:         c.Pool.MaxOpenConns,
		MaxIdleConns:         c.Pool.MaxIdleConns,
		ConnMaxLifetimeMin:   c.Pool.ConnMaxLifetimeMin,
		AutoMigrate:          c.Migrations.AutoMigrate,
		SafeMode:             c.Migrations.SafeMode,
		EnableLogging:        c.Logging.Enabled,
		SlowQueryThresholdMs: c.Logging.SlowQueryThresholdMs,
	}
}

// NewDSN creates a DSN string from central config.DatabaseConfig
func NewDSN(c config.DatabaseConfig) string {
	return FromCentralConfig(c).DSN()
}
