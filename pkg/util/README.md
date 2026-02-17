# Package Overview

### `pkg/util/password`

Argon2id password hashing with OWASP-recommended parameters.

```go
import "github.com/Alijeyrad/simorq_backend/pkg/util/password"

// Hash a password
hash, err := password.Hash("mypassword")

// Verify a password
err := password.Verify(hash, "mypassword")
if errors.Is(err, password.ErrMismatch) {
    // wrong password
}

// Convenience boolean check
if password.Match(hash, "mypassword") {
    // correct
}

// Generate random password
randomPwd := password.Generate(16) // 16 chars

// Check if hash needs upgrade
if password.NeedsRehash(hash) {
    newHash, _ := password.Hash(plaintext)
}
```

**Functions:**

- `Hash(password string) (string, error)` - Hash with default params
- `HashWithParams(password string, p *Params) (string, error)` - Hash with custom params
- `Verify(hash, password string) error` - Verify password (constant-time)
- `Match(hash, password string) bool` - Boolean verification
- `NeedsRehash(hash string) bool` - Check if params are outdated
- `Generate(length int) string` - Generate random password

---

### `pkg/util/otp`

Secure OTP generation and verification for authentication flows.

```go
import "github.com/Alijeyrad/simorq_backend/pkg/util/otp"

// Generate 6-digit OTP
code, err := otp.GenerateDefault() // "123456"

// Or custom length (4-10 digits)
code, err := otp.Generate(4) // "1234"

// Hash for storage (never store plaintext!)
hash := otp.Hash(code)

// Verify user input
err := otp.Verify(hash, userInput)
if errors.Is(err, otp.ErrMismatch) {
    // wrong code
}

// Generate alphanumeric (for tokens)
token, _ := otp.GenerateAlphanumeric(8) // "ABCD1234"

// Generate hex string
hex, _ := otp.GenerateHex(16) // 32 hex chars
```

**Functions:**

- `Generate(length int) (string, error)` - Numeric OTP (4-10 digits)
- `GenerateDefault() (string, error)` - 6-digit OTP
- `Hash(code string) string` - SHA-256 hash for storage
- `Verify(hash, code string) error` - Constant-time verification
- `GenerateAlphanumeric(length int) (string, error)` - Alphanumeric code
- `GenerateHex(byteLength int) (string, error)` - Hex token

---

### `pkg/util/phone`

Phone number parsing, validation, and normalization using libphonenumber.

```go
import "github.com/Alijeyrad/simorq_backend/pkg/util/phone"

// Normalize to E.164 format
e164, err := phone.NormalizeE164("09123456789", "IR")
// "+989123456789"

// Iranian convenience function
e164, err := phone.NormalizeIR("0912 345 6789")
// "+989123456789"

// Full parse with metadata
result, err := phone.Parse("09123456789", "IR")
// result.E164 = "+989123456789"
// result.National = "0912 345 6789"
// result.CountryCode = 98
// result.Region = "IR"
// result.IsValid = true
// result.IsMobile = true

// Validation
phone.IsValid("09123456789", "IR") // true
phone.IsMobile("09123456789", "IR") // true

// Masking for display
phone.MaskPhone("+989123456789") // "+9891*****789"
```

**Functions:**

- `NormalizeE164(raw, region string) (string, error)` - E.164 format
- `NormalizeIR(raw string) (string, error)` - Iranian numbers
- `Parse(raw, region string) (*ParseResult, error)` - Full parse
- `IsValid(raw, region string) bool` - Validation check
- `IsMobile(raw, region string) bool` - Mobile check
- `FormatNational(raw, region string) (string, error)` - National format
- `MaskPhone(e164 string) string` - Mask for display
- `ExtractDigits(raw string) string` - Extract digits only

---

### `pkg/util/codes`

Secure code and token generation for referrals, invitations, and verification.

```go
import "github.com/Alijeyrad/simorq_backend/pkg/util/codes"

// Referral code (8 uppercase alphanumeric)
code, err := codes.GenerateReferralCode() // "ABCD1234"

// Invitation code (8 uppercase alphanumeric)
code, err := codes.GenerateInvitationCode() // "XYZ98765"

// Secure tokens (hex, 32 chars)
token, err := codes.GenerateInvitationToken()
token, err := codes.GenerateVerificationToken()

// URL-safe tokens
token, err := codes.GenerateURLSafeToken(16) // base64

// Numeric codes
code, err := codes.GenerateNumericCode(6) // "123456"

// Formatting
formatted := codes.FormatCode("ABCD1234", 4) // "ABCD-1234"
parsed := codes.ParseCode("ABCD-1234") // "ABCD1234"
normalized := codes.NormalizeCode("abcd1234") // "ABCD1234"
```

**Functions:**

- `GenerateReferralCode() (string, error)` - 8-char referral code
- `GenerateInvitationCode() (string, error)` - 8-char invitation code
- `GenerateInvitationToken() (string, error)` - 32-char hex token
- `GenerateVerificationToken() (string, error)` - 32-char hex token
- `GenerateSecureToken(byteLength int) (string, error)` - Custom hex token
- `GenerateURLSafeToken(byteLength int) (string, error)` - URL-safe base64
- `GenerateNumericCode(length int) (string, error)` - Numeric only
- `GenerateCode(length int, charset string) (string, error)` - Custom charset
- `NormalizeCode(code string) string` - Uppercase + trim
- `FormatCode(code string, groupSize int) string` - Add dashes
- `ParseCode(formatted string) string` - Remove formatting

---

### `pkg/util/gqlid`

GraphQL global ID encoding/decoding (Relay-style).

```go
import "github.com/Alijeyrad/simorq_backend/pkg/util/gqlid"

// Encode type + UUID to global ID
gid := gqlid.Encode("User", userUUID) // "VXNlcjo1NTBlODQwMC4uLg"

// Decode back
typ, uuid, err := gqlid.Decode(gid)
// typ = "User", uuid = original UUID
```

---

## Usage in Services

### Auth Service Example

```go
import (
    "github.com/Alijeyrad/simorq_backend/pkg/util/codes"
    "github.com/Alijeyrad/simorq_backend/pkg/util/otp"
    "github.com/Alijeyrad/simorq_backend/pkg/util/password"
    "github.com/Alijeyrad/simorq_backend/pkg/util/phone"
)

func (s *AuthService) RegisterStart(ctx context.Context, input Input) error {
    // Normalize phone
    e164, err := phone.NormalizeE164(input.PhoneRaw, input.PhoneRegion)
    if err != nil {
        return ErrInvalidPhone
    }

    // Hash password
    pwdHash, err := password.Hash(input.Password)
    if err != nil {
        return err
    }

    // Generate referral code for new user
    referralCode, err := codes.GenerateReferralCode()
    if err != nil {
        return err
    }

    // Generate OTP for verification
    otpCode, otpHash := otp.GenerateDefault()
    // Store otpHash in DB, send otpCode via SMS

    // ...
}
```

### Waitlist Service Example

```go
import (
    "github.com/Alijeyrad/simorq_backend/pkg/util/codes"
)

func (s *WaitlistService) Join(ctx context.Context, input Input) error {
    // Generate verification token
    verificationToken, err := codes.GenerateVerificationToken()
    if err != nil {
        return err
    }

    // Generate referral code for this entry
    referralCode, err := codes.GenerateReferralCode()
    if err != nil {
        return err
    }

    // ...
}

func (s *WaitlistService) Invite(ctx context.Context, entryID string) error {
    // Generate invitation token
    invitationToken, err := codes.GenerateInvitationToken()
    if err != nil {
        return err
    }

    // Build URL
    url := fmt.Sprintf("%s/accept-invitation?token=%s", s.cfg.BaseURL, invitationToken)
    // ...
}
```

---

## Dependencies

Add to `go.mod`:

```
golang.org/x/crypto v0.x.x  // for argon2
github.com/nyaruka/phonenumbers v1.x.x  // for phone parsing
```

---

## Security Notes

1. **Never store plaintext passwords or OTPs** - Always use `password.Hash()` and `otp.Hash()`
2. **Use constant-time comparison** - Both packages use `crypto/subtle.ConstantTimeCompare`
3. **Argon2id is the recommended algorithm** - Resistant to GPU/ASIC attacks
4. **OTPs should expire** - Use with time-based expiry (typically 5 minutes)
5. **Tokens are cryptographically random** - Using `crypto/rand`
