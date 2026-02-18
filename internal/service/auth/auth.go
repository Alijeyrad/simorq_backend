package auth

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/Alijeyrad/simorq_backend/config"
	"github.com/Alijeyrad/simorq_backend/internal/repo"
	entuser "github.com/Alijeyrad/simorq_backend/internal/repo/user"
	"github.com/Alijeyrad/simorq_backend/pkg/crypto"
	pasetotoken "github.com/Alijeyrad/simorq_backend/pkg/paseto"
	"github.com/Alijeyrad/simorq_backend/pkg/sms"
	"github.com/Alijeyrad/simorq_backend/pkg/util/otp"
	"github.com/Alijeyrad/simorq_backend/pkg/util/password"
)

const (
	maxOTPAttempts   = 5
	accountLockMins  = 15
	maxLoginAttempts = 5
)

// redisKeyOTP returns the Redis key for the OTP hash associated with a phone.
func redisKeyOTP(phone string) string { return "otp:" + phone }

// redisKeyOTPAttempts returns the Redis key for OTP attempt counter.
func redisKeyOTPAttempts(phone string) string { return "otp:attempts:" + phone }

// redisKeySession returns the Redis key for a session.
func redisKeySession(sessionID string) string { return "session:" + sessionID }

var reNationalID = regexp.MustCompile(`^\d{10}$`)
var rePhone = regexp.MustCompile(`^09\d{9}$`)

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

type RegisterRequest struct {
	Phone      string
	Password   string
	FirstName  string
	LastName   string
	NationalID string // optional; raw digits
}

type VerifyOTPRequest struct {
	Phone string
	Code  string
}

type LoginRequest struct {
	Phone      string // one of Phone or NationalID must be set
	NationalID string
	Password   string
}

type InternSetupRequest struct {
	FirstName      string
	LastName       string
	InternshipYear int
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // seconds until access token expires
}

// ---------------------------------------------------------------------------
// Service interface
// ---------------------------------------------------------------------------

type Service interface {
	Register(ctx context.Context, req RegisterRequest) error
	VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*AuthTokens, error)
	Login(ctx context.Context, req LoginRequest) (*AuthTokens, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*AuthTokens, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	InternSetup(ctx context.Context, userID uuid.UUID, req InternSetupRequest) (*repo.User, error)
}

// ---------------------------------------------------------------------------
// Implementation
// ---------------------------------------------------------------------------

type authService struct {
	db     *repo.Client
	rdb    *redis.Client
	sms    *sms.Client
	paseto *pasetotoken.Manager
	cfg    *config.Config
	encKey []byte // AES-256 key for national_id encryption
}

func New(
	db *repo.Client,
	rdb *redis.Client,
	smsCli *sms.Client,
	paseto *pasetotoken.Manager,
	cfg *config.Config,
) (Service, error) {
	encKey, err := crypto.KeyFromHex(cfg.Authentication.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("auth service: invalid encryption key: %w", err)
	}
	return &authService{
		db:     db,
		rdb:    rdb,
		sms:    smsCli,
		paseto: paseto,
		cfg:    cfg,
		encKey: encKey,
	}, nil
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

func (s *authService) Register(ctx context.Context, req RegisterRequest) error {
	// Normalise
	req.Phone = strings.TrimSpace(req.Phone)
	req.NationalID = strings.TrimSpace(req.NationalID)

	// Validate phone
	if !rePhone.MatchString(req.Phone) {
		return ErrInvalidPhone
	}
	// Validate national_id when provided
	if req.NationalID != "" && !reNationalID.MatchString(req.NationalID) {
		return ErrInvalidNationalID
	}
	if len(req.Password) < 8 {
		return ErrPasswordTooShort
	}

	// Check phone uniqueness
	exists, err := s.db.User.Query().Where(entuser.Phone(req.Phone), entuser.DeletedAtIsNil()).Exist(ctx)
	if err != nil {
		return fmt.Errorf("check phone: %w", err)
	}
	if exists {
		return ErrPhoneAlreadyExists
	}

	// Hash password
	passHash, err := password.Hash(req.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Encrypt national_id + compute hash for lookups
	var encNatID, natIDHash *string
	if req.NationalID != "" {
		enc, err := crypto.Encrypt(s.encKey, req.NationalID)
		if err != nil {
			return fmt.Errorf("encrypt national_id: %w", err)
		}
		h := crypto.Hash(req.NationalID)

		// Check national_id uniqueness
		hashExists, err := s.db.User.Query().
			Where(entuser.NationalIDHash(h), entuser.DeletedAtIsNil()).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("check national_id: %w", err)
		}
		if hashExists {
			return ErrNationalIDExists
		}

		encNatID = &enc
		natIDHash = &h
	}

	// Create user
	q := s.db.User.Create().
		SetPhone(req.Phone).
		SetPasswordHash(passHash).
		SetMustChangePassword(false).
		SetPhoneVerified(false).
		SetStatus("ACTIVE")

	if req.FirstName != "" {
		q = q.SetFirstName(req.FirstName)
	}
	if req.LastName != "" {
		q = q.SetLastName(req.LastName)
	}
	if encNatID != nil {
		q = q.SetNationalID(*encNatID)
	}
	if natIDHash != nil {
		q = q.SetNationalIDHash(*natIDHash)
	}

	_, err = q.Save(ctx)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	// Generate and send OTP
	return s.sendOTP(ctx, req.Phone)
}

// ---------------------------------------------------------------------------
// VerifyOTP
// ---------------------------------------------------------------------------

func (s *authService) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*AuthTokens, error) {
	req.Phone = strings.TrimSpace(req.Phone)
	req.Code = strings.TrimSpace(req.Code)

	// Get stored OTP hash
	otpHash, err := s.rdb.Get(ctx, redisKeyOTP(req.Phone)).Result()
	if err == redis.Nil {
		return nil, ErrOTPExpired
	}
	if err != nil {
		return nil, fmt.Errorf("redis get otp: %w", err)
	}

	// Check attempt count
	attempts, _ := s.rdb.Get(ctx, redisKeyOTPAttempts(req.Phone)).Int()
	if attempts >= maxOTPAttempts {
		return nil, ErrOTPMaxAttempts
	}

	// Verify code
	if err := otp.Verify(otpHash, req.Code); err != nil {
		s.rdb.Incr(ctx, redisKeyOTPAttempts(req.Phone))
		return nil, ErrOTPInvalid
	}

	// Clean up OTP keys
	s.rdb.Del(ctx, redisKeyOTP(req.Phone), redisKeyOTPAttempts(req.Phone))

	// Mark phone as verified
	u, err := s.db.User.Query().Where(entuser.Phone(req.Phone), entuser.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	_, err = s.db.User.UpdateOne(u).SetPhoneVerified(true).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update phone_verified: %w", err)
	}

	return s.createSession(ctx, u)
}

// ---------------------------------------------------------------------------
// Login
// ---------------------------------------------------------------------------

func (s *authService) Login(ctx context.Context, req LoginRequest) (*AuthTokens, error) {
	req.Phone = strings.TrimSpace(req.Phone)
	req.NationalID = strings.TrimSpace(req.NationalID)

	// Find user by phone or national_id_hash
	var u *repo.User
	var err error

	if req.Phone != "" {
		u, err = s.db.User.Query().
			Where(entuser.Phone(req.Phone), entuser.DeletedAtIsNil()).
			Only(ctx)
	} else if req.NationalID != "" {
		if !reNationalID.MatchString(req.NationalID) {
			return nil, ErrInvalidCredentials
		}
		h := crypto.Hash(req.NationalID)
		u, err = s.db.User.Query().
			Where(entuser.NationalIDHash(h), entuser.DeletedAtIsNil()).
			Only(ctx)
	} else {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	// Check account status
	if u.Status == "SUSPENDED" {
		return nil, ErrAccountSuspended
	}
	if !u.PhoneVerified {
		return nil, ErrPhoneNotVerified
	}

	// Check lockout
	if u.LockedUntil != nil && time.Now().Before(*u.LockedUntil) {
		return nil, ErrAccountLocked
	}

	// Verify password
	if u.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}
	if err := password.Verify(*u.PasswordHash, req.Password); err != nil {
		s.recordFailedLogin(ctx, u)
		return nil, ErrInvalidCredentials
	}

	// Reset failure counters
	now := time.Now()
	s.db.User.UpdateOne(u).
		SetFailedLoginAttempts(0).
		SetNillableLockedUntil(nil).
		SetLastLoginAt(now).
		Save(ctx)

	return s.createSession(ctx, u)
}

// ---------------------------------------------------------------------------
// RefreshTokens
// ---------------------------------------------------------------------------

func (s *authService) RefreshTokens(ctx context.Context, refreshToken string) (*AuthTokens, error) {
	claims, err := s.paseto.Verify(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if claims.Type != pasetotoken.TokenTypeRefresh {
		return nil, ErrInvalidToken
	}
	if claims.SessionID == nil {
		return nil, ErrInvalidToken
	}

	sessionKey := redisKeySession(claims.SessionID.String())

	// Check session exists
	if err := s.rdb.Get(ctx, sessionKey).Err(); err == redis.Nil {
		return nil, ErrSessionNotFound
	} else if err != nil {
		return nil, fmt.Errorf("redis get session: %w", err)
	}

	// Extend session TTL
	refreshTTL := time.Duration(s.cfg.Authentication.Paseto.RefreshTTLDays) * 24 * time.Hour
	s.rdb.Expire(ctx, sessionKey, refreshTTL)

	// Issue new access token only (refresh token stays the same until logout)
	accessTTL := time.Duration(s.cfg.Authentication.Paseto.AccessTTLMinutes) * time.Minute
	accessToken, err := s.paseto.IssueAccess(claims.UserID, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // unchanged
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, nil
}

// ---------------------------------------------------------------------------
// Logout
// ---------------------------------------------------------------------------

func (s *authService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	deleted, err := s.rdb.Del(ctx, redisKeySession(sessionID.String())).Result()
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	if deleted == 0 {
		// Session already expired — not an error from the client's perspective
		slog.Debug("logout: session not found in Redis (already expired)", "session_id", sessionID)
	}

	// Mark revoked in DB (best-effort; not critical path)
	now := time.Now()
	s.db.UserSession.Update().
		Where().
		SetRevokedAt(now).
		Save(ctx)

	return nil
}

// ---------------------------------------------------------------------------
// InternSetup
// ---------------------------------------------------------------------------

func (s *authService) InternSetup(ctx context.Context, userID uuid.UUID, req InternSetupRequest) (*repo.User, error) {
	u, err := s.db.User.Get(ctx, userID)
	if err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	upd := s.db.User.UpdateOne(u).
		SetMustChangePassword(false)
	if req.FirstName != "" {
		upd = upd.SetFirstName(req.FirstName)
	}
	if req.LastName != "" {
		upd = upd.SetLastName(req.LastName)
	}

	return upd.Save(ctx)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (s *authService) sendOTP(ctx context.Context, phone string) error {
	code, err := otp.GenerateDefault()
	if err != nil {
		return fmt.Errorf("generate OTP: %w", err)
	}

	otpTTL := time.Duration(s.cfg.Authentication.OTPTTLMinutes) * time.Minute
	if otpTTL <= 0 {
		otpTTL = 5 * time.Minute
	}

	// Store hash
	if err := s.rdb.Set(ctx, redisKeyOTP(phone), otp.Hash(code), otpTTL).Err(); err != nil {
		return fmt.Errorf("store OTP: %w", err)
	}
	// Reset attempts
	s.rdb.Set(ctx, redisKeyOTPAttempts(phone), "0", otpTTL+5*time.Minute)

	// Send via SMS.ir
	templateID := s.cfg.SMS.SMSIR.TemplateID
	if err := s.sms.SendOTP(ctx, phone, templateID, code); err != nil {
		// Log but don't fail — SMS failure shouldn't block registration
		slog.Warn("failed to send OTP SMS", "phone", phone, "error", err)
	}

	return nil
}

func (s *authService) createSession(ctx context.Context, u *repo.User) (*AuthTokens, error) {
	sessionID := uuid.Must(uuid.NewV7())

	refreshTTL := time.Duration(s.cfg.Authentication.Paseto.RefreshTTLDays) * 24 * time.Hour
	accessTTL := time.Duration(s.cfg.Authentication.Paseto.AccessTTLMinutes) * time.Minute

	// Store in Redis
	sessionKey := redisKeySession(sessionID.String())
	if err := s.rdb.Set(ctx, sessionKey, u.ID.String(), refreshTTL).Err(); err != nil {
		return nil, fmt.Errorf("store session: %w", err)
	}

	// Issue tokens
	access, err := s.paseto.IssueAccess(u.ID, &sessionID)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}
	refresh, err := s.paseto.IssueRefresh(u.ID, &sessionID)
	if err != nil {
		return nil, fmt.Errorf("issue refresh token: %w", err)
	}

	// Persist session record to DB (audit, best-effort)
	expiresAt := time.Now().Add(refreshTTL)
	refreshHash := crypto.Hash(refresh) // SHA-256 of refresh token
	s.db.UserSession.Create().
		SetUserID(u.ID).
		SetSessionID(sessionID.String()).
		SetRefreshTokenHash(refreshHash).
		SetExpiresAt(expiresAt).
		Save(ctx)

	return &AuthTokens{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, nil
}

func (s *authService) recordFailedLogin(ctx context.Context, u *repo.User) {
	attempts := u.FailedLoginAttempts + 1
	upd := s.db.User.UpdateOne(u).
		SetFailedLoginAttempts(attempts).
		SetLastFailedLoginAt(time.Now())

	if attempts >= maxLoginAttempts {
		lockUntil := time.Now().Add(accountLockMins * time.Minute)
		upd = upd.SetLockedUntil(lockUntil)
	}
	upd.Save(ctx)
}
