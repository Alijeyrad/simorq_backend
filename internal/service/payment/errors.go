package payment

import "errors"

var (
	ErrPaymentNotFound   = errors.New("payment request not found")
	ErrPaymentFailed     = errors.New("payment failed or cancelled by user")
	ErrZarinPalFailure   = errors.New("zarinpal gateway error")
	ErrAmountMismatch    = errors.New("payment amount does not match")
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient wallet balance")
)
