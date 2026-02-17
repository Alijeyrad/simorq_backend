package email

import (
	"time"

	"github.com/Alijeyrad/simorq_backend/config"
)

// Config holds email service configuration
type Config struct {
	Enabled bool
	From    string

	// SMTP settings
	SMTPHost           string
	SMTPPort           int
	SMTPUsername       string
	SMTPPassword       string
	SMTPUseTLS         bool
	SMTPTimeoutSeconds int

	// Template settings
	AppName      string
	BaseURL      string
	SupportEmail string

	// Branding
	PrimaryColor string
	LogoURL      string
}

// DefaultConfig returns sensible defaults for email configuration
func DefaultConfig() Config {
	return Config{
		Enabled:            false,
		SMTPPort:           587,
		SMTPUseTLS:         true,
		SMTPTimeoutSeconds: 30,
		PrimaryColor:       "#007bff",
	}
}

// SMTPTimeout returns the SMTP timeout as a duration
func (c Config) SMTPTimeout() time.Duration {
	if c.SMTPTimeoutSeconds <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.SMTPTimeoutSeconds) * time.Second
}

// FromCentralConfig converts central config.EmailConfig to package Config
func FromCentralConfig(c config.EmailConfig) Config {
	return Config{
		Enabled:            c.Enabled,
		From:               c.From,
		SMTPHost:           c.SMTP.Host,
		SMTPPort:           c.SMTP.Port,
		SMTPUsername:       c.SMTP.Username,
		SMTPPassword:       c.SMTP.Password,
		SMTPUseTLS:         c.SMTP.UseTLS,
		SMTPTimeoutSeconds: c.SMTP.TimeoutSeconds,
	}
}
