package sms

import (
	"context"
	"fmt"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/arsmn/go-smsir/smsir"
)

// Client provides SMS sending functionality via sms.ir.
type Client struct {
	client  *smsir.Client
	enabled bool
}

// NewFromConfig creates a new SMS client from the application configuration.
// If SMS is disabled, returns a client that no-ops on all operations.
func NewFromConfig(cfg config.SMSConfig) (*Client, error) {
	if !cfg.Enabled {
		return &Client{enabled: false}, nil
	}

	if cfg.SMSIR.APIKey == "" {
		return nil, fmt.Errorf("sms.ir API key required when SMS enabled")
	}

	client := smsir.NewClient().WithAuthentication(cfg.SMSIR.APIKey, cfg.SMSIR.SecretKey)

	return &Client{
		client:  client,
		enabled: true,
	}, nil
}

// SendOTP sends an OTP code to the specified phone number using the configured template.
// If SMS is disabled, this is a no-op and returns nil.
//
// Parameters:
//   - ctx: Context for the request
//   - phoneNumber: Recipient phone number (E.164 format recommended)
//   - templateID: sms.ir template ID to use
//   - otpCode: The OTP code to send
//
// The template must have a parameter named "code" for the OTP value.
func (c *Client) SendOTP(ctx context.Context, phoneNumber, templateID, otpCode string) error {
	if !c.enabled {
		// No-op when disabled (useful for development)
		return nil
	}

	if phoneNumber == "" {
		return fmt.Errorf("phone number is required")
	}
	if templateID == "" {
		return fmt.Errorf("template ID is required")
	}
	if otpCode == "" {
		return fmt.Errorf("OTP code is required")
	}

	req := &smsir.UltraFastSendRequest{
		Mobile:     phoneNumber,
		TemplateID: templateID,
		Parameters: []smsir.UltraFastParameter{
			{Key: "code", Value: otpCode},
		},
	}

	_, err := c.client.Verification.UltraFastSend(ctx, req)
	if err != nil {
		return fmt.Errorf("sms.ir send failed: %w", err)
	}

	return nil
}

// IsEnabled returns whether SMS sending is enabled.
func (c *Client) IsEnabled() bool {
	return c.enabled
}
