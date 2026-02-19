package payment

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entappt "github.com/Alijeyrad/simorq_backend/internal/repo/appointment"
	entpayment "github.com/Alijeyrad/simorq_backend/internal/repo/paymentrequest"
	enttransaction "github.com/Alijeyrad/simorq_backend/internal/repo/transaction"
	entwallet "github.com/Alijeyrad/simorq_backend/internal/repo/wallet"
	"github.com/Alijeyrad/simorq_backend/pkg/crypto"
	zarinpalpkg "github.com/Alijeyrad/simorq_backend/pkg/zarinpal"
)

// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

type Service interface {
	// ZarinPal flow
	InitiatePayment(ctx context.Context, clinicID, userID uuid.UUID, apptID *uuid.UUID, amount int64, desc string) (payURL string, err error)
	VerifyPayment(ctx context.Context, authority string, status string) (*repo.PaymentRequest, error)

	// Wallet management
	GetOrCreateWallet(ctx context.Context, ownerType string, ownerID uuid.UUID) (*repo.Wallet, error)
	GetTransactions(ctx context.Context, walletID uuid.UUID, page, perPage int) ([]*repo.Transaction, error)
	SetIBAN(ctx context.Context, walletID uuid.UUID, iban, accountHolder string) error
	RequestWithdrawal(ctx context.Context, walletID, clinicID uuid.UUID, amount int64) (*repo.WithdrawalRequest, error)
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type paymentService struct {
	db  *repo.Client
	zp  *zarinpalpkg.Client
	cfg *config.Config
}

func New(db *repo.Client, zp *zarinpalpkg.Client, cfg *config.Config) Service {
	return &paymentService{db: db, zp: zp, cfg: cfg}
}

// ---------------------------------------------------------------------------
// ZarinPal flow
// ---------------------------------------------------------------------------

func (s *paymentService) InitiatePayment(ctx context.Context, clinicID, userID uuid.UUID, apptID *uuid.UUID, amount int64, desc string) (string, error) {
	// Create a pending payment_request record first
	c := s.db.PaymentRequest.Create().
		SetClinicID(clinicID).
		SetUserID(userID).
		SetAmount(amount).
		SetDescription(desc).
		SetNillableAppointmentID(apptID)

	pr, err := c.Save(ctx)
	if err != nil {
		return "", fmt.Errorf("create payment request: %w", err)
	}

	// Call ZarinPal
	authority, payURL, err := s.zp.RequestPayment(ctx, amount, "IRR", desc, s.cfg.ZarinPal.CallbackURL)
	if err != nil {
		// Mark as failed
		_ = s.db.PaymentRequest.UpdateOne(pr).
			SetStatus(entpayment.StatusFailed).
			Exec(ctx)
		return "", fmt.Errorf("%w: %v", ErrZarinPalFailure, err)
	}

	// Store authority
	if err := s.db.PaymentRequest.UpdateOne(pr).
		SetZarinpalAuthority(authority).
		Exec(ctx); err != nil {
		return "", fmt.Errorf("store authority: %w", err)
	}

	return payURL, nil
}

func (s *paymentService) VerifyPayment(ctx context.Context, authority string, status string) (*repo.PaymentRequest, error) {
	if status != "OK" {
		// Payment was cancelled/failed by user — find request and mark failed
		pr, err := s.db.PaymentRequest.Query().
			Where(entpayment.ZarinpalAuthority(authority)).
			Only(ctx)
		if err == nil {
			_ = s.db.PaymentRequest.UpdateOne(pr).
				SetStatus(entpayment.StatusFailed).
				Exec(ctx)
		}
		return nil, ErrPaymentFailed
	}

	pr, err := s.db.PaymentRequest.Query().
		Where(entpayment.ZarinpalAuthority(authority)).
		Only(ctx)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrPaymentNotFound
		}
		return nil, fmt.Errorf("get payment request: %w", err)
	}

	// Call ZarinPal verify
	refID, cardPan, _, err := s.zp.VerifyPayment(ctx, authority, pr.Amount)
	if err != nil {
		_ = s.db.PaymentRequest.UpdateOne(pr).
			SetStatus(entpayment.StatusFailed).
			Exec(ctx)
		return nil, fmt.Errorf("%w: %v", ErrZarinPalFailure, err)
	}

	// Mark as success
	now := time.Now()
	upd := s.db.PaymentRequest.UpdateOne(pr).
		SetStatus(entpayment.StatusSuccess).
		SetZarinpalRefID(strconv.FormatInt(refID, 10)).
		SetPaidAt(now)
	if cardPan != "" {
		upd = upd.SetZarinpalCardPan(cardPan)
	}
	verified, err := upd.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update payment request: %w", err)
	}

	// Update appointment payment status if linked
	if pr.AppointmentID != nil {
		_ = s.db.Appointment.Update().
			Where(entappt.ID(*pr.AppointmentID)).
			SetPaymentStatus(entappt.PaymentStatusReservationPaid).
			Exec(ctx)
	}

	return verified, nil
}

// ---------------------------------------------------------------------------
// Wallet
// ---------------------------------------------------------------------------

func (s *paymentService) GetOrCreateWallet(ctx context.Context, ownerType string, ownerID uuid.UUID) (*repo.Wallet, error) {
	w, err := s.db.Wallet.Query().
		Where(
			entwallet.OwnerTypeEQ(entwallet.OwnerType(ownerType)),
			entwallet.OwnerID(ownerID),
		).
		Only(ctx)
	if err == nil {
		return w, nil
	}
	if !repo.IsNotFound(err) {
		return nil, fmt.Errorf("get wallet: %w", err)
	}

	w, err = s.db.Wallet.Create().
		SetOwnerType(entwallet.OwnerType(ownerType)).
		SetOwnerID(ownerID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return w, nil
}

func (s *paymentService) GetTransactions(ctx context.Context, walletID uuid.UUID, page, perPage int) ([]*repo.Transaction, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	txns, err := s.db.Transaction.Query().
		Where(enttransaction.WalletID(walletID)).
		Order(enttransaction.ByCreatedAt(sql.OrderDesc())).
		Offset(offset).
		Limit(perPage).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	return txns, nil
}

func (s *paymentService) SetIBAN(ctx context.Context, walletID uuid.UUID, iban, accountHolder string) error {
	key, err := crypto.KeyFromHex(s.cfg.Authentication.EncryptionKey)
	if err != nil {
		return fmt.Errorf("encryption key: %w", err)
	}

	encrypted, err := crypto.Encrypt(key, iban)
	if err != nil {
		return fmt.Errorf("encrypt iban: %w", err)
	}
	hash := crypto.Hash(iban)

	wallet, err := s.db.Wallet.Get(ctx, walletID)
	if err != nil {
		if repo.IsNotFound(err) {
			return ErrWalletNotFound
		}
		return fmt.Errorf("get wallet: %w", err)
	}

	return s.db.Wallet.UpdateOne(wallet).
		SetIbanEncrypted(encrypted).
		SetIbanHash(hash).
		SetAccountHolder(accountHolder).
		Exec(ctx)
}

func (s *paymentService) RequestWithdrawal(ctx context.Context, walletID, clinicID uuid.UUID, amount int64) (*repo.WithdrawalRequest, error) {
	wallet, err := s.db.Wallet.Get(ctx, walletID)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrWalletNotFound
		}
		return nil, fmt.Errorf("get wallet: %w", err)
	}

	if wallet.Balance < amount {
		return nil, ErrInsufficientFunds
	}

	if wallet.IbanEncrypted == nil {
		return nil, fmt.Errorf("no IBAN set for this wallet")
	}

	accountHolder := ""
	if wallet.AccountHolder != nil {
		accountHolder = *wallet.AccountHolder
	}

	wr, err := s.db.WithdrawalRequest.Create().
		SetWalletID(walletID).
		SetClinicID(clinicID).
		SetAmount(amount).
		SetIbanEncrypted(*wallet.IbanEncrypted).
		SetAccountHolder(accountHolder).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create withdrawal request: %w", err)
	}

	// Record a pending debit transaction (actual balance deduction happens on processing)
	_ = s.db.Transaction.Create().
		SetWalletID(walletID).
		SetType(enttransaction.TypeDebit).
		SetAmount(amount).
		SetBalanceBefore(wallet.Balance).
		SetBalanceAfter(wallet.Balance). // balance not yet deducted (pending)
		SetEntityType("withdrawal_request").
		SetEntityID(wr.ID).
		SetDescription("Withdrawal request pending").
		Exec(ctx)

	return wr, nil
}
