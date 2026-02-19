// Package zarinpal provides a minimal HTTP client for the ZarinPal v4 payment gateway.
package zarinpal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Alijeyrad/simorq_backend/config"
)

var (
	ErrPaymentFailed      = errors.New("zarinpal: payment failed or cancelled by user")
	ErrValidation         = errors.New("zarinpal: validation error")
	ErrAmountMismatch     = errors.New("zarinpal: amount does not match original request")
	ErrInvalidAuthority   = errors.New("zarinpal: invalid authority")
	ErrAuthorityNotFound  = errors.New("zarinpal: authority not found")
	ErrUnexpectedResponse = errors.New("zarinpal: unexpected response from gateway")
)

// Client is a lightweight ZarinPal HTTP client.
type Client struct {
	merchantID  string
	baseURL     string
	startPayURL string
	httpClient  *http.Client
}

// New creates a Client from config. Uses sandbox endpoints when cfg.Sandbox is true.
func New(cfg config.ZarinPalConfig) *Client {
	baseURL := "https://payment.zarinpal.com/pg/v4"
	startPayURL := "https://payment.zarinpal.com/pg/StartPay/"
	if cfg.Sandbox {
		baseURL = "https://sandbox.zarinpal.com/pg/v4"
		startPayURL = "https://sandbox.zarinpal.com/pg/StartPay/"
	}
	return &Client{
		merchantID:  cfg.MerchantID,
		baseURL:     baseURL,
		startPayURL: startPayURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// RequestPayment initiates a payment and returns (authority, paymentPageURL, error).
// amount should be in Rials (currency="IRR") or Tomans (currency="IRT").
func (c *Client) RequestPayment(ctx context.Context, amount int64, currency, desc, callbackURL string) (authority string, payURL string, err error) {
	reqBody := map[string]any{
		"merchant_id":  c.merchantID,
		"amount":       amount,
		"currency":     currency,
		"description":  desc,
		"callback_url": callbackURL,
	}

	var resp struct {
		Data struct {
			Code      int    `json:"code"`
			Authority string `json:"authority"`
			Fee       int    `json:"fee"`
			Message   string `json:"message"`
		} `json:"data"`
		Errors any `json:"errors"`
	}

	if err := c.post(ctx, "/payment/request.json", reqBody, &resp); err != nil {
		return "", "", fmt.Errorf("zarinpal request: %w", err)
	}

	switch resp.Data.Code {
	case 100:
		// success
	case -9:
		return "", "", ErrValidation
	default:
		return "", "", fmt.Errorf("%w (code=%d, msg=%s)", ErrUnexpectedResponse, resp.Data.Code, resp.Data.Message)
	}

	if resp.Data.Authority == "" {
		return "", "", ErrUnexpectedResponse
	}

	return resp.Data.Authority, c.startPayURL + resp.Data.Authority, nil
}

// VerifyPayment verifies a payment after the user returns from the gateway.
// Returns (refID, cardPan, alreadyVerified, error).
// alreadyVerified=true means code 101 (idempotent verify — treat as success).
func (c *Client) VerifyPayment(ctx context.Context, authority string, amount int64) (refID int64, cardPan string, alreadyVerified bool, err error) {
	reqBody := map[string]any{
		"merchant_id": c.merchantID,
		"amount":      amount,
		"authority":   authority,
	}

	var resp struct {
		Data struct {
			Code     int    `json:"code"`
			RefID    int64  `json:"ref_id"`
			CardPan  string `json:"card_pan"`
			CardHash string `json:"card_hash"`
			Message  string `json:"message"`
		} `json:"data"`
		Errors any `json:"errors"`
	}

	if err := c.post(ctx, "/payment/verify.json", reqBody, &resp); err != nil {
		return 0, "", false, fmt.Errorf("zarinpal verify: %w", err)
	}

	switch resp.Data.Code {
	case 100:
		return resp.Data.RefID, resp.Data.CardPan, false, nil
	case 101:
		return resp.Data.RefID, resp.Data.CardPan, true, nil
	case -9:
		return 0, "", false, ErrValidation
	case -50:
		return 0, "", false, ErrAmountMismatch
	case -51:
		return 0, "", false, ErrPaymentFailed
	case -54:
		return 0, "", false, ErrInvalidAuthority
	case -55:
		return 0, "", false, ErrAuthorityNotFound
	default:
		return 0, "", false, fmt.Errorf("%w (code=%d, msg=%s)", ErrUnexpectedResponse, resp.Data.Code, resp.Data.Message)
	}
}

// post sends a JSON POST request to baseURL+path and decodes the JSON response into out.
func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
