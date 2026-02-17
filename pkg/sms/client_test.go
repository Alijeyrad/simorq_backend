package sms

import (
	"context"
	"testing"

	"github.com/Alijeyrad/simorq_backend/config"
)

func TestNewFromConfig_Disabled(t *testing.T) {
	cfg := config.SMSConfig{
		Enabled: false,
	}

	client, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig failed: %v", err)
	}

	if client.IsEnabled() {
		t.Error("Expected client to be disabled")
	}
}

func TestNewFromConfig_EnabledWithoutAPIKey(t *testing.T) {
	cfg := config.SMSConfig{
		Enabled: true,
		SMSIR: config.SMSIRConfig{
			APIKey:     "",
			SecretKey:  "",
			TemplateID: "test-template",
		},
	}

	_, err := NewFromConfig(cfg)
	if err == nil {
		t.Error("Expected error when API key is missing")
	}
}

func TestNewFromConfig_EnabledWithAPIKey(t *testing.T) {
	cfg := config.SMSConfig{
		Enabled: true,
		SMSIR: config.SMSIRConfig{
			APIKey:     "test-api-key",
			SecretKey:  "test-secret-key",
			TemplateID: "test-template",
		},
	}

	client, err := NewFromConfig(cfg)
	if err != nil {
		t.Fatalf("NewFromConfig failed: %v", err)
	}

	if !client.IsEnabled() {
		t.Error("Expected client to be enabled")
	}
}

func TestSendOTP_DisabledClient(t *testing.T) {
	client := &Client{enabled: false}

	err := client.SendOTP(context.Background(), "+989121234567", "template-id", "123456")
	if err != nil {
		t.Errorf("Expected no error for disabled client, got: %v", err)
	}
}

func TestSendOTP_Validation(t *testing.T) {
	client := &Client{enabled: true}

	tests := []struct {
		name        string
		phone       string
		templateID  string
		otpCode     string
		expectError bool
	}{
		{
			name:        "empty phone number",
			phone:       "",
			templateID:  "template-id",
			otpCode:     "123456",
			expectError: true,
		},
		{
			name:        "empty template ID",
			phone:       "+989121234567",
			templateID:  "",
			otpCode:     "123456",
			expectError: true,
		},
		{
			name:        "empty OTP code",
			phone:       "+989121234567",
			templateID:  "template-id",
			otpCode:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.SendOTP(context.Background(), tt.phone, tt.templateID, tt.otpCode)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled client", true},
		{"disabled client", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{enabled: tt.enabled}
			if client.IsEnabled() != tt.enabled {
				t.Errorf("Expected IsEnabled() = %v, got %v", tt.enabled, client.IsEnabled())
			}
		})
	}
}
