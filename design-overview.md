# Simorgh — Complete Platform Design Document

> **Version:** 1.0
> **Date:** February 2026
> **Domain:** simorqcare.com
> **Stack:** Nuxt 4 + Nuxt UI · Go + Fiber v3 · Ent ORM · PostgreSQL · Redis · NATS · Casbin · PASETO · ZarinPal · ArvanCloud S3 + VPS
> **Backend Phases Complete:** 1 (Foundation) · 2 (Clinical Core) · 3 (Scheduling & Payments)

---

## 1. System Overview

Simorgh is a **multi-tenant SaaS platform** for psychology and therapy clinics. It connects **clinics** (therapists, admins, interns) with **clients** (patients/مراجعان) through a single web application — no subdomains, one unified deployment.

### Core Value Flows

```
Client → Browse therapists → Book slot → Pay reservation fee → Attend session
Therapist → Manage schedule → Conduct session → Write report → Upload files
Intern → Login with given credentials → Complete profile → Submit tasks → View assigned patients
Owner → Create clinic → Add therapists → Assign permissions → Set policies
Platform → Take commission → Manage payouts → Handle support tickets
```

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        CDN (ArvanCloud)                      │
│                     simorqcare.com                           │
└───────────────┬─────────────────────────────┬───────────────┘
                │                             │
    ┌───────────▼───────────┐     ┌───────────▼───────────┐
    │   Nuxt 4 (SSR/SPA)   │     │   Go API (Fiber v3)   │
    │   Port 3000           │     │   Port 8080            │
    │   ─ Nuxt UI           │     │   ─ REST + SSE         │
    │   ─ Pinia             │     │   ─ Casbin RBAC        │
    │   ─ Vazirmatn RTL     │     │   ─ PASETO auth        │
    │   ─ PWA (vite-pwa)    │     │   ─ ZarinPal gateway   │
    └───────────┬───────────┘     └──┬──────┬──────┬───────┘
                │                    │      │      │
          ┌─────▼─────┐    ┌────────▼┐ ┌───▼───┐ ┌▼────────┐
          │   Caddy    │    │Postgres │ │ Redis │ │  NATS   │
          │  (reverse  │    │  15     │ │       │ │ JetStr. │
          │   proxy)   │    └─────────┘ └───────┘ └─────────┘
          └────────────┘         │
                            ┌────▼────┐
                            │ArvanS3  │
                            │(files)  │
                            └─────────┘
```

---

## 2. Roles & Permission Model (Casbin)

### 2.1 Role Hierarchy

```
platform_superadmin          ← Simorgh platform operator (you)
  └── clinic_owner           ← Creates the clinic, full control
        └── clinic_admin     ← Delegated by owner, near-full control
              └── therapist  ← Manages own patients, schedule, reports
                    └── intern  ← Limited access, task-based
client                       ← Patient, highest access to own data
```

### 2.2 Casbin Model (RBAC with resource ownership + tenant isolation)

```ini
# model.conf
[request_definition]
r = sub, tenant, obj, act

[policy_definition]
p = sub, tenant, obj, act

[role_definition]
g = _, _, _    # user, role, tenant

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.tenant) && r.tenant == p.tenant && keyMatch2(r.obj, p.obj) && r.act == p.act
```

### 2.3 Permission Matrix

| Resource              | Owner | Admin | Therapist         | Intern          | Client           |
| --------------------- | ----- | ----- | ----------------- | --------------- | ---------------- |
| Clinic settings       | CRUD  | CRUD  | R                 | —              | —               |
| Therapist profiles    | CRUD  | CRUD  | R(own)+U(own)     | R               | R                |
| Client profiles       | CRUD  | CRUD  | R(assigned)       | R(assigned*)    | R(own)+U(own)    |
| Patient files/reports | CRUD  | CRUD  | CRUD(assigned)    | CR(assigned*)   | R(own)           |
| Appointments          | CRUD  | CRUD  | CRUD(own)         | R(assigned*)    | CRUD(own)        |
| Schedule/slots        | CRUD  | CRUD  | CRUD(own)         | R               | R                |
| Payments/wallet       | CRUD  | R     | R(own)            | —              | R(own)           |
| Tickets               | CRUD  | CRUD  | CRUD(own)         | CRUD(own)       | CRUD(own)        |
| Chat messages         | CRUD  | CRUD  | CRUD(own convos)  | —              | CRUD(own convos) |
| Intern tasks          | CRUD  | CRUD  | CRUD(own interns) | CRUD(own tasks) | —               |
| Commission settings   | —    | —    | —                | —              | —               |
| Platform admin        | CRUD  | —    | —                | —              | —               |

> `*assigned` = only when therapist explicitly grants access to that intern for specific patients.

### 2.4 Per-Entity Permission Overrides

Owners and admins can override the default matrix per-user. The `/panel/permissions` page exposes this UI.

```sql
-- custom_permissions table
clinic_id | user_id | resource_type | resource_id | action | granted
```

---

## 3. Entity-Relationship Design (ERD)

### 3.1 Core Entities

```
┌─────────────────────────────────────────────────────────────────┐
│                          TENANT LAYER                            │
├─────────────────────────────────────────────────────────────────┤
│  clinics                                                         │
│  ├── clinic_members (user + role per clinic)                     │
│  ├── clinic_settings (policies, branding)                        │
│  └── clinic_invitations                                          │
├─────────────────────────────────────────────────────────────────┤
│                          USER LAYER                              │
├─────────────────────────────────────────────────────────────────┤
│  users (unified: one account can be client + therapist)          │
│  ├── user_profiles (personal info per role)                      │
│  ├── user_notification_prefs (per-user notification settings)    │
│  ├── user_sessions (PASETO token tracking)                       │
│  └── user_devices (push notification tokens)                     │
├─────────────────────────────────────────────────────────────────┤
│                       CLINICAL LAYER                             │
├─────────────────────────────────────────────────────────────────┤
│  patients (per-clinic patient record)                            │
│  ├── patient_files (documents, uploads)                          │
│  ├── patient_reports (session + standalone)                      │
│  ├── patient_prescriptions                                       │
│  ├── patient_tests                                               │
│  ├── patient_test_results                                        │
│  └── patient_notes                                               │
├─────────────────────────────────────────────────────────────────┤
│                      SCHEDULING LAYER                            │
├─────────────────────────────────────────────────────────────────┤
│  time_slots (therapist availability)                             │
│  ├── recurring_rules                                             │
│  appointments                                                    │
│  └── appointment_cancellations                                   │
├─────────────────────────────────────────────────────────────────┤
│                       FINANCIAL LAYER                            │
├─────────────────────────────────────────────────────────────────┤
│  wallets                                                         │
│  ├── transactions                                                │
│  ├── payment_requests (ZarinPal)                                 │
│  ├── withdrawal_requests                                         │
│  └── commission_rules                                            │
├─────────────────────────────────────────────────────────────────┤
│                    COMMUNICATION LAYER                            │
├─────────────────────────────────────────────────────────────────┤
│  conversations                                                   │
│  ├── messages                                                    │
│  tickets                                                         │
│  ├── ticket_messages                                             │
│  └── notifications                                               │
├─────────────────────────────────────────────────────────────────┤
│                       INTERN LAYER                               │
├─────────────────────────────────────────────────────────────────┤
│  intern_profiles                                                 │
│  ├── intern_tasks                                                │
│  ├── intern_task_files                                           │
│  ├── intern_task_reviews                                         │
│  └── intern_patient_access                                       │
├─────────────────────────────────────────────────────────────────┤
│                         FILE LAYER                               │
├─────────────────────────────────────────────────────────────────┤
│  files (unified upload record, ArvanCloud S3)                    │
│  └── linked polymorphically to patient_files, messages, tickets  │
└─────────────────────────────────────────────────────────────────┘
```

## 4. API Design

### 4.1 Base URL & Versioning

```
API:  https://simorqcare.com/api/v1/
Web:  https://simorqcare.com/
```

### 4.4 Pagination & Filtering Convention

```json
// GET /api/v1/patients?page=1&per_page=20&sort=created_at&order=desc&status=active

{
  "data": [...],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

Patient list filters:

- `therapist_id`, `status`, `payment_status`, `has_discount`
- `sort` = `created_at` | `appointment_date` | `start_time`
- `order` = `asc` | `desc`

---

## 5. NATS Event Topology

### 5.1 Subjects

```
simorgh.appointment.created.{clinic_id}
simorgh.appointment.cancelled.{clinic_id}
simorgh.appointment.completed.{clinic_id}
simorgh.payment.received.{clinic_id}
simorgh.payment.withdrawal.{clinic_id}
simorgh.message.new.{conversation_id}
simorgh.ticket.new.{clinic_id}
simorgh.ticket.replied.{ticket_id}
simorgh.notification.push.{user_id}
simorgh.report.created.{clinic_id}
simorgh.intern.task.submitted.{clinic_id}
simorgh.intern.task.reviewed.{clinic_id}
```

### 5.2 Consumers

```
notification-worker   → All events → creates notification records + push
sms-worker            → appointment.created, appointment.cancelled → SMS
wallet-worker         → payment.received → splits client → clinic wallet, commission → platform wallet
analytics-worker      → (future) aggregate stats
```

---

### Key Go Dependencies

```
github.com/gofiber/fiber/v3
entgo.io/ent                          ← ORM; generates repo/ from internal/schema/
go.uber.org/fx                        ← dependency injection
github.com/casbin/casbin/v2
github.com/redis/go-redis/v9
github.com/aws/aws-sdk-go-v2          ← S3-compatible (ArvanCloud)
github.com/google/uuid                ← UUIDv7 generation
log/slog                              ← structured logging (stdlib)
```

## 8. Security Design

### 8.1 Authentication (PASETO v4)

- **Access token:** 15-minute expiry, contains `{user_id, roles: [{clinic_id, role}]}`.
- **Refresh token:** 7-day expiry, `httpOnly` cookie + DB for revocation.
- **Token rotation:** Each refresh invalidates the old refresh token.
- **Auth middleware runs client-side only** — protected routes are `ssr: false`, so Pinia state is always available.

### 8.2 Sensitive Data Encryption

| Field       | Storage                    | Lookup        |
| ----------- | -------------------------- | ------------- |
| National ID | AES-256-GCM encrypted      | Argon2id hash |
| Phone       | Plaintext (needed for SMS) | Indexed       |
| Password    | Argon2id hash              | —            |
| IBAN        | AES-256-GCM encrypted      | —            |

### 8.3 File Security

- All uploads → ArvanCloud S3, **private ACL**.
- Downloads via `/api/v1/files/:key` or `/api/v1/patients/:id/files/:fid/download` — API checks permissions, returns **presigned URL** (5-minute TTL).

### 8.4 API Security

- **Rate limiting:** Redis-backed, per-IP + per-user.
- **CORS:** `simorqcare.com` origin only.
- **CSP:** Set dynamically in `server/middleware/csp.ts` using `runtimeConfig.public.apiBase`.
- **Input validation:** `go-playground/validator`.
- **SQL injection:** Parameterized queries via pgx.
- **Open redirect:** Frontend always validates `protocol === 'https:'` before `window.location.href` (payment redirect).
- **No `v-html`:** Vue auto-escapes all template interpolations.

### 8.5 PWA Security

- Service worker caches only static assets + Google Fonts — never API responses.
- Install prompt enabled; push notifications require user permission + device token registration.

---

## 11. Key Technical Decisions

| Decision       | Choice                        | Rationale                                             |
| -------------- | ----------------------------- | ----------------------------------------------------- |
| Auth tokens    | Pinia (client-only)           | No SSR for protected routes — tokens never on server |
| ORM            | Ent ORM (entgo.io/ent)        | Type-safe Go from schemas; auto-migrate; no SQL files |
| File storage   | ArvanCloud S3                 | Private ACL + presigned URLs                          |
| Date handling  | Gregorian in DB, Jalali in UI | Standard DB ops, convert at display layer             |
| Multi-tenancy  | Shared DB + clinic_id column  | Casbin enforces isolation                             |
| Real-time (v1) | HTTP polling                  | Simplest; NATS ready for upgrade                      |
| i18n           | Static Persian text           | No runtime locale switching needed                    |
| PWA            | vite-pwa + Workbox            | Offline shell; push notification support              |
| SMS provider   | TBD (Kavenegar / Ghasedak)    | Iranian provider required                             |
| Sessions       | In-person only                | No video infrastructure needed for v1                 |
