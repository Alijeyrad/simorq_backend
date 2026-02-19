# ZarinPal Payment Gateway — Integration Reference

Source: official ZarinPal API docs (v4)

---

## Endpoints

| Environment | Base URL |
|---|---|
| Production | `https://payment.zarinpal.com/pg/v4` |
| Sandbox | `https://sandbox.zarinpal.com/pg/v4` |

All requests: `Content-Type: application/json`, `Accept: application/json`

Redirect URL pattern: `https://payment.zarinpal.com/pg/StartPay/{authority}`
(replace domain with `sandbox.zarinpal.com` for sandbox)

Sandbox authorities always start with the letter **S**.

---

## Flow

```
1. POST /payment/request.json   → get authority
2. Redirect user to StartPay/{authority}
3. ZarinPal redirects back to callback_url?Authority=...&Status=OK|NOK
4. If Status=OK → POST /payment/verify.json
5. Check response code (100=success, 101=already verified)
```

---

## 1. Request Payment

**POST** `/payment/request.json`

```json
{
  "merchant_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "amount": 10000,
  "currency": "IRR",
  "description": "...",
  "callback_url": "https://yoursite.com/payments/verify",
  "metadata": {
    "mobile": "09121234567",
    "email": "user@example.com",
    "order_id": "optional"
  }
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `merchant_id` | string | yes | 36-char UUID from panel |
| `amount` | integer | yes | Amount in Rials (IRR) or Tomans (IRT) |
| `currency` | string | no | `"IRR"` (default) or `"IRT"` |
| `description` | string | yes | Shown on payment page |
| `callback_url` | string | yes | Must match registered domain |
| `metadata.mobile` | string | no | Buyer phone |
| `metadata.email` | string | no | Buyer email |
| `referrer_id` | string | no | Referral code |

**Response:**
```json
{
  "data": {
    "code": 100,
    "message": "Success",
    "authority": "A0000000000000000000000000000wwOGYpd",
    "fee_type": "Merchant",
    "fee": 100
  },
  "errors": []
}
```

On `code=100`: redirect user to `StartPay/{authority}`.

---

## 2. ZarinPal Callback

ZarinPal sends user back to `callback_url` with query params:

```
GET {callback_url}?Authority=A000...&Status=OK
GET {callback_url}?Authority=A000...&Status=NOK
```

- `Status=NOK` → payment failed or cancelled. Do **not** call verify. Mark as failed.
- `Status=OK` → proceed to verify.

---

## 3. Verify Payment

**POST** `/payment/verify.json`

```json
{
  "merchant_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "amount": 10000,
  "authority": "A0000000000000000000000000000wwOGYpd"
}
```

**Response:**
```json
{
  "data": {
    "code": 100,
    "message": "Verified",
    "ref_id": 201,
    "card_pan": "502229******5995",
    "card_hash": "1EBE3E...89B69",
    "fee_type": "Merchant",
    "fee": 0
  },
  "errors": []
}
```

| Field | Notes |
|---|---|
| `ref_id` | Bank reference ID — show to user as receipt |
| `card_pan` | Masked card number |
| `card_hash` | SHA-256 hash of card number |

**Important:** `amount` in verify must exactly match `amount` in request (code `-50` if not).

---

## Response Codes

### Success
| Code | Meaning |
|---|---|
| `100` | Success |
| `101` | Already verified (idempotent — treat as success) |

### Request errors
| Code | Meaning |
|---|---|
| `-9` | Validation error (merchant_id / callback_url / description / amount invalid) |
| `-10` | IP or merchant_id incorrect |
| `-11` | Merchant inactive |
| `-14` | Callback URL domain mismatch with registered domain |
| `-41` | Amount exceeds 100M Tomans |

### Verify errors
| Code | Meaning |
|---|---|
| `-50` | Amount mismatch between request and verify |
| `-51` | Payment failed / session not active |
| `-54` | Invalid authority |
| `-55` | Authority not found |

---

## Implementation in this project

`pkg/zarinpal/zarinpal.go` — thin HTTP client wrapping the two endpoints.

Config keys (`config.yaml` → `zarinpal:`):
```yaml
zarinpal:
  merchant_id: ""          # 36-char UUID from ZarinPal panel
  callback_url: "https://simorqcare.com/api/v1/payments/verify"
  sandbox: true            # set false in production
```

Payment flow wired in `internal/service/payment/payment.go`:
- `InitiatePayment()` → creates `payment_requests` row (pending), calls `RequestPayment`, stores authority
- `VerifyPayment()` → checks `Status` param, calls `VerifyPayment`, updates row, sets `appointment.payment_status=reservation_paid`
