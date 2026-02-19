# Simorgh — Complete Platform Design Document

> **Version:** 3.0
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

### 3.2 Detailed Table Definitions

#### **users**

```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    national_id     VARCHAR(10) UNIQUE,          -- کد ملی (encrypted at rest)
    national_id_hash VARCHAR(64) NOT NULL UNIQUE, -- argon2 hash for lookups
    phone           VARCHAR(11) NOT NULL UNIQUE,  -- شماره همراه
    phone_verified  BOOLEAN DEFAULT FALSE,
    password_hash   VARCHAR(255) NOT NULL,        -- argon2id
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    gender          VARCHAR(10),
    marital_status  VARCHAR(20),
    birth_year      INTEGER,                      -- سال تولد (Jalali)
    avatar_key      VARCHAR(500),                 -- S3 key
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_national_id_hash ON users(national_id_hash);
```

#### **user_notification_prefs** (new — frontend uses GET/PATCH /users/me/notifications)

```sql
CREATE TABLE user_notification_prefs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Notification channels per event type
    appointment_sms         BOOLEAN DEFAULT TRUE,
    appointment_push        BOOLEAN DEFAULT TRUE,
    message_push            BOOLEAN DEFAULT TRUE,
    ticket_reply_push       BOOLEAN DEFAULT TRUE,

    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);
```

#### **clinics**

```sql
CREATE TABLE clinics (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    description     TEXT,
    logo_key        VARCHAR(500),
    phone           VARCHAR(20),
    address         TEXT,
    city            VARCHAR(100),
    province        VARCHAR(100),
    is_active       BOOLEAN DEFAULT TRUE,
    is_verified     BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
```

#### **clinic_members**

```sql
CREATE TABLE clinic_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id       UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL CHECK (role IN ('owner', 'admin', 'therapist', 'intern')),
    is_active       BOOLEAN DEFAULT TRUE,
    joined_at       TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(clinic_id, user_id)
);

CREATE INDEX idx_clinic_members_clinic ON clinic_members(clinic_id);
CREATE INDEX idx_clinic_members_user ON clinic_members(user_id);
```

#### **clinic_settings**

```sql
CREATE TABLE clinic_settings (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id               UUID NOT NULL UNIQUE REFERENCES clinics(id) ON DELETE CASCADE,

    reservation_fee_amount  BIGINT DEFAULT 0,
    reservation_fee_percent INTEGER DEFAULT 0,
    cancellation_window_hours INTEGER DEFAULT 24,
    cancellation_fee_amount BIGINT DEFAULT 0,
    cancellation_fee_percent INTEGER DEFAULT 0,
    allow_client_self_book  BOOLEAN DEFAULT TRUE,

    default_session_duration_min INTEGER DEFAULT 60,
    default_session_price   BIGINT DEFAULT 0,

    working_hours           JSONB,

    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);
```

#### **clinic_permissions** (per-user overrides, used by /panel/permissions)

```sql
CREATE TABLE clinic_permissions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id       UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resource_type   VARCHAR(50) NOT NULL,
    resource_id     UUID,
    action          VARCHAR(20) NOT NULL,
    granted         BOOLEAN NOT NULL DEFAULT TRUE,

    created_at      TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(clinic_id, user_id, resource_type, COALESCE(resource_id, '00000000-0000-0000-0000-000000000000'::uuid), action)
);

CREATE INDEX idx_clinic_permissions_clinic_user ON clinic_permissions(clinic_id, user_id);
```

#### **therapist_profiles** (extends clinic_members where role=therapist)

```sql
CREATE TABLE therapist_profiles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_member_id    UUID NOT NULL UNIQUE REFERENCES clinic_members(id) ON DELETE CASCADE,
    education           VARCHAR(255),
    psychology_license  VARCHAR(50),
    approach            VARCHAR(255),
    specialties         TEXT[],
    bio                 TEXT,
    rating              NUMERIC(2,1) DEFAULT 0,
    session_price       BIGINT,            -- میانگین هزینه جلسه (NOT reservation fee)
    session_duration_min INTEGER,
    is_accepting        BOOLEAN DEFAULT TRUE,   -- scheduling toggle

    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);
```

#### **patients** (per-clinic patient record)

```sql
CREATE TABLE patients (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id           UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id),
    file_number         VARCHAR(50),

    primary_therapist_id UUID REFERENCES clinic_members(id),

    status              VARCHAR(20) DEFAULT 'active'
                        CHECK (status IN ('active', 'waiting_reservation', 'inactive', 'discharged')),
    session_count       INTEGER DEFAULT 0,

    total_cancellations INTEGER DEFAULT 0,
    last_cancel_reason  TEXT,

    has_discount        BOOLEAN DEFAULT FALSE,
    discount_percent    INTEGER DEFAULT 0,
    payment_status      VARCHAR(20) DEFAULT 'unpaid'
                        CHECK (payment_status IN ('paid', 'unpaid', 'partial')),
    total_paid          BIGINT DEFAULT 0,

    notes               TEXT,
    referral_source     VARCHAR(255),
    chief_complaint     TEXT,

    is_child            BOOLEAN DEFAULT FALSE,
    child_birth_date    DATE,
    child_school        VARCHAR(255),
    child_grade         VARCHAR(50),
    parent_name         VARCHAR(255),
    parent_phone        VARCHAR(11),
    parent_relation     VARCHAR(50),

    developmental_history JSONB,

    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(clinic_id, user_id)
);

CREATE INDEX idx_patients_clinic ON patients(clinic_id);
CREATE INDEX idx_patients_user ON patients(user_id);
CREATE INDEX idx_patients_therapist ON patients(primary_therapist_id);
CREATE INDEX idx_patients_status ON patients(clinic_id, status);
CREATE INDEX idx_patients_file_number ON patients(clinic_id, file_number);
```

#### **patient_reports**

```sql
CREATE TABLE patient_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    clinic_id       UUID NOT NULL REFERENCES clinics(id),
    therapist_id    UUID NOT NULL REFERENCES clinic_members(id),
    appointment_id  UUID REFERENCES appointments(id),

    title           VARCHAR(255),
    content         TEXT,
    report_date     DATE NOT NULL DEFAULT CURRENT_DATE,

    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_patient_reports_patient ON patient_reports(patient_id);
CREATE INDEX idx_patient_reports_clinic ON patient_reports(clinic_id);
CREATE INDEX idx_patient_reports_therapist ON patient_reports(therapist_id);
CREATE INDEX idx_patient_reports_date ON patient_reports(report_date);
```

#### **patient_files** (unified file storage, with download endpoint)

```sql
CREATE TABLE patient_files (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    clinic_id       UUID NOT NULL REFERENCES clinics(id),
    uploaded_by     UUID NOT NULL REFERENCES users(id),

    linked_type     VARCHAR(30),   -- 'report', 'test_result', 'prescription', NULL=standalone
    linked_id       UUID,

    file_name       VARCHAR(255) NOT NULL,
    file_key        VARCHAR(500) NOT NULL,  -- S3 key
    file_size       BIGINT,
    mime_type       VARCHAR(100),
    description     TEXT,

    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_patient_files_patient ON patient_files(patient_id);
CREATE INDEX idx_patient_files_linked ON patient_files(linked_type, linked_id);
```

#### **patient_prescriptions**

```sql
CREATE TABLE patient_prescriptions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    clinic_id       UUID NOT NULL REFERENCES clinics(id),
    therapist_id    UUID NOT NULL REFERENCES clinic_members(id),

    title           VARCHAR(255),
    notes           TEXT,
    file_key        VARCHAR(500),   -- optional attached file (S3 key)
    file_name       VARCHAR(255),
    prescribed_date DATE NOT NULL DEFAULT CURRENT_DATE,

    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
```

#### **psych_tests** (platform-wide test catalog)

```sql
CREATE TABLE psych_tests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    name_fa         VARCHAR(255),
    description     TEXT,
    category        VARCHAR(100),
    age_range       VARCHAR(50),
    schema          JSONB,
    scoring_method  VARCHAR(50),
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

#### **patient_tests** (test results per patient)

```sql
CREATE TABLE patient_tests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id      UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    clinic_id       UUID NOT NULL REFERENCES clinics(id),
    test_id         UUID REFERENCES psych_tests(id),
    administered_by UUID REFERENCES clinic_members(id),

    test_name       VARCHAR(255),   -- free text if test_id is null
    raw_scores      JSONB,
    computed_scores JSONB,
    interpretation  TEXT,
    test_date       DATE NOT NULL DEFAULT CURRENT_DATE,
    status          VARCHAR(20) DEFAULT 'assigned'
                    CHECK (status IN ('assigned', 'completed', 'reviewed')),

    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
```

#### **time_slots** (therapist availability)

```sql
-- Actual schema (Ent ORM). Timestamps stored as TIMESTAMPTZ.
CREATE TABLE time_slots (
    id                  UUID PRIMARY KEY,  -- UUIDv7
    clinic_id           UUID NOT NULL REFERENCES clinics(id),
    therapist_id        UUID NOT NULL REFERENCES clinic_members(id),

    start_time          TIMESTAMPTZ NOT NULL,
    end_time            TIMESTAMPTZ NOT NULL,

    status              VARCHAR(20) NOT NULL DEFAULT 'available'
                        CHECK (status IN ('available', 'booked', 'blocked', 'cancelled')),

    session_price       BIGINT,            -- nullable; per-slot override
    reservation_fee     BIGINT,            -- nullable; per-slot override
    is_recurring        BOOLEAN DEFAULT FALSE,
    recurring_rule_id   UUID,              -- nullable non-FK snapshot ref

    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_time_slots_therapist_start ON time_slots(therapist_id, start_time);
CREATE INDEX idx_time_slots_clinic_status_start ON time_slots(clinic_id, status, start_time);
```

#### **recurring_rules**

```sql
CREATE TABLE recurring_rules (
    id              UUID PRIMARY KEY,  -- UUIDv7
    clinic_id       UUID NOT NULL REFERENCES clinics(id),
    therapist_id    UUID NOT NULL REFERENCES clinic_members(id),

    day_of_week     SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),  -- 0=Sunday
    start_hour      SMALLINT NOT NULL,
    start_minute    SMALLINT NOT NULL,
    end_hour        SMALLINT NOT NULL,
    end_minute      SMALLINT NOT NULL,

    session_price   BIGINT,   -- nullable
    reservation_fee BIGINT,   -- nullable

    valid_from      TIMESTAMPTZ NOT NULL,
    valid_until     TIMESTAMPTZ,          -- nullable
    is_active       BOOLEAN DEFAULT TRUE,

    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_recurring_rules_therapist ON recurring_rules(therapist_id, day_of_week, is_active);
CREATE INDEX idx_recurring_rules_clinic ON recurring_rules(clinic_id);
```

#### **appointments**

```sql
CREATE TABLE appointments (
    id                      UUID PRIMARY KEY,  -- UUIDv7
    clinic_id               UUID NOT NULL REFERENCES clinics(id),
    therapist_id            UUID NOT NULL REFERENCES clinic_members(id),
    patient_id              UUID NOT NULL REFERENCES patients(id),
    time_slot_id            UUID,              -- nullable non-FK snapshot ref (allows slot deletion)

    start_time              TIMESTAMPTZ NOT NULL,
    end_time                TIMESTAMPTZ NOT NULL,

    status                  VARCHAR(20) NOT NULL DEFAULT 'scheduled'
                            CHECK (status IN ('scheduled', 'completed', 'cancelled', 'no_show')),

    session_price           BIGINT NOT NULL,   -- snapshot from slot at booking time
    reservation_fee         BIGINT NOT NULL DEFAULT 0,

    payment_status          VARCHAR(20) NOT NULL DEFAULT 'unpaid'
                            CHECK (payment_status IN (
                                'unpaid', 'reservation_paid', 'fully_paid', 'refunded'
                            )),

    notes                   TEXT,
    cancellation_reason     TEXT,
    cancel_requested_by     VARCHAR(20) CHECK (cancel_requested_by IN ('patient', 'therapist', 'clinic')),
    cancelled_at            TIMESTAMPTZ,
    cancellation_fee        BIGINT DEFAULT 0,
    completed_at            TIMESTAMPTZ,

    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_appointments_clinic_therapist_start ON appointments(clinic_id, therapist_id, start_time);
CREATE INDEX idx_appointments_clinic_patient ON appointments(clinic_id, patient_id);
CREATE INDEX idx_appointments_therapist_status ON appointments(therapist_id, status, start_time);
CREATE INDEX idx_appointments_patient_status ON appointments(patient_id, status);
```

#### **wallets**

```sql
CREATE TABLE wallets (
    id               UUID PRIMARY KEY,  -- UUIDv7
    owner_type       VARCHAR(20) NOT NULL CHECK (owner_type IN ('user', 'clinic', 'platform')),
    owner_id         UUID NOT NULL,
    balance          BIGINT NOT NULL DEFAULT 0,

    -- IBAN stored encrypted (AES-256-GCM); iban_hash is SHA-256 for uniqueness lookups
    iban_encrypted   VARCHAR(1000),
    iban_hash        VARCHAR(64),
    account_holder   VARCHAR(200),

    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(owner_type, owner_id)
);
```

#### **transactions** (append-only ledger)

```sql
CREATE TABLE transactions (
    id              UUID PRIMARY KEY,  -- UUIDv7
    wallet_id       UUID NOT NULL REFERENCES wallets(id),

    type            VARCHAR(10) NOT NULL CHECK (type IN ('credit', 'debit')),
    amount          BIGINT NOT NULL,
    balance_before  BIGINT NOT NULL,
    balance_after   BIGINT NOT NULL,

    -- Polymorphic reference to the entity that caused this transaction
    entity_type     VARCHAR(30),   -- e.g. 'appointment', 'payment_request', 'withdrawal'
    entity_id       UUID,

    description     VARCHAR(500),

    created_at      TIMESTAMPTZ DEFAULT NOW()
    -- No updated_at: transactions are immutable
);

CREATE INDEX idx_transactions_wallet_created ON transactions(wallet_id, created_at DESC);
CREATE INDEX idx_transactions_entity ON transactions(entity_type, entity_id);
```

#### **payment_requests**

```sql
CREATE TABLE payment_requests (
    id                      UUID PRIMARY KEY,  -- UUIDv7
    clinic_id               UUID NOT NULL REFERENCES clinics(id),
    user_id                 UUID NOT NULL REFERENCES users(id),
    appointment_id          UUID,              -- nullable

    amount                  BIGINT NOT NULL,
    description             VARCHAR(500) NOT NULL,

    status                  VARCHAR(20) NOT NULL DEFAULT 'pending'
                            CHECK (status IN ('pending', 'success', 'failed', 'cancelled')),
    source                  VARCHAR(20) NOT NULL DEFAULT 'zarinpal'
                            CHECK (source IN ('zarinpal', 'wallet')),

    -- ZarinPal-specific fields
    zarinpal_authority      VARCHAR(200),
    zarinpal_ref_id         VARCHAR(50),       -- int64 ref_id stored as string
    zarinpal_card_pan       VARCHAR(25),       -- masked e.g. "502229******5995"
    zarinpal_card_hash      VARCHAR(70),

    paid_at                 TIMESTAMPTZ,

    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_payment_requests_user ON payment_requests(user_id, status, created_at DESC);
CREATE INDEX idx_payment_requests_clinic ON payment_requests(clinic_id, status);
CREATE INDEX idx_payment_requests_authority ON payment_requests(zarinpal_authority);
```

#### **withdrawal_requests**

```sql
CREATE TABLE withdrawal_requests (
    id               UUID PRIMARY KEY,  -- UUIDv7
    wallet_id        UUID NOT NULL REFERENCES wallets(id),
    clinic_id        UUID NOT NULL REFERENCES clinics(id),
    amount           BIGINT NOT NULL,

    status           VARCHAR(20) NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),

    -- IBAN snapshot at time of request (encrypted, same key as wallet.iban_encrypted)
    iban_encrypted   VARCHAR(1000) NOT NULL,
    account_holder   VARCHAR(200) NOT NULL,
    bank_ref         VARCHAR(100),             -- bank transfer reference

    requested_at     TIMESTAMPTZ DEFAULT NOW(),
    processed_at     TIMESTAMPTZ,
    failure_reason   TEXT,

    created_at       TIMESTAMPTZ DEFAULT NOW()
    -- No updated_at: CreatedAtMixin only
);

CREATE INDEX idx_withdrawal_requests_wallet ON withdrawal_requests(wallet_id, status);
CREATE INDEX idx_withdrawal_requests_clinic ON withdrawal_requests(clinic_id, status, requested_at DESC);
```

#### **commission_rules**

```sql
CREATE TABLE commission_rules (
    id                      UUID PRIMARY KEY,  -- UUIDv7
    clinic_id               UUID NOT NULL UNIQUE REFERENCES clinics(id),
    platform_fee_percent    INTEGER NOT NULL DEFAULT 0,
    clinic_fee_percent      INTEGER NOT NULL DEFAULT 0,
    is_flat_fee             BOOLEAN NOT NULL DEFAULT FALSE,
    flat_fee_amount         BIGINT NOT NULL DEFAULT 0,
    is_active               BOOLEAN NOT NULL DEFAULT TRUE,

    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);
```

#### **conversations & messages**

```sql
CREATE TABLE conversations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id       UUID NOT NULL REFERENCES clinics(id),
    participant_a   UUID NOT NULL REFERENCES users(id),
    participant_b   UUID NOT NULL REFERENCES users(id),
    patient_id      UUID REFERENCES patients(id),  -- optional patient context

    last_message_at TIMESTAMPTZ,
    is_active       BOOLEAN DEFAULT TRUE,

    created_at      TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(clinic_id, participant_a, participant_b)
);

CREATE TABLE messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id       UUID NOT NULL REFERENCES users(id),

    content         TEXT,
    file_key        VARCHAR(500),
    file_name       VARCHAR(255),
    file_mime       VARCHAR(100),

    is_read         BOOLEAN DEFAULT FALSE,
    read_at         TIMESTAMPTZ,
    is_deleted      BOOLEAN DEFAULT FALSE,

    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at);
```

#### **tickets & ticket_messages**

```sql
CREATE TABLE tickets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id       UUID REFERENCES clinics(id),   -- NULL = platform-level
    user_id         UUID NOT NULL REFERENCES users(id),

    subject         VARCHAR(255) NOT NULL,
    status          VARCHAR(20) DEFAULT 'open'
                    CHECK (status IN ('open', 'answered', 'closed')),
    priority        VARCHAR(10) DEFAULT 'normal'
                    CHECK (priority IN ('low', 'normal', 'high', 'urgent')),

    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE ticket_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id       UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    sender_id       UUID NOT NULL REFERENCES users(id),

    content         TEXT NOT NULL,
    file_key        VARCHAR(500),
    file_name       VARCHAR(255),

    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

#### **notifications**

```sql
CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),

    type            VARCHAR(50) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    body            TEXT,
    data            JSONB,

    is_read         BOOLEAN DEFAULT FALSE,
    is_pushed       BOOLEAN DEFAULT FALSE,

    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read, created_at DESC);
```

#### **contact_messages** (from public contact form)

```sql
CREATE TABLE contact_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    subject         VARCHAR(255) NOT NULL,
    message         TEXT NOT NULL,

    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

#### **intern_profiles, intern_tasks, intern_patient_access**

```sql
CREATE TABLE intern_profiles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_member_id    UUID NOT NULL UNIQUE REFERENCES clinic_members(id) ON DELETE CASCADE,
    internship_year     INTEGER,
    supervisor_ids      UUID[],

    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE intern_tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id       UUID NOT NULL REFERENCES clinics(id),
    intern_id       UUID NOT NULL REFERENCES clinic_members(id),

    title           VARCHAR(255) NOT NULL,
    caption         TEXT,
    submitted_at    TIMESTAMPTZ DEFAULT NOW(),

    reviewed_by     UUID REFERENCES clinic_members(id),
    review_status   VARCHAR(20) DEFAULT 'pending'
                    CHECK (review_status IN ('pending', 'reviewed', 'needs_revision')),
    review_comment  TEXT,
    grade           VARCHAR(10),
    reviewed_at     TIMESTAMPTZ,

    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE intern_task_files (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id         UUID NOT NULL REFERENCES intern_tasks(id) ON DELETE CASCADE,
    file_key        VARCHAR(500) NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    file_size       BIGINT,
    mime_type       VARCHAR(100),

    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE intern_patient_access (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    intern_id       UUID NOT NULL REFERENCES clinic_members(id),
    patient_id      UUID NOT NULL REFERENCES patients(id),
    granted_by      UUID NOT NULL REFERENCES clinic_members(id),

    can_view_files  BOOLEAN DEFAULT TRUE,
    can_write_reports BOOLEAN DEFAULT FALSE,

    granted_at      TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(intern_id, patient_id)
);
```

#### **user_devices** (push notifications / PWA)

```sql
CREATE TABLE user_devices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token    VARCHAR(500) NOT NULL,
    platform        VARCHAR(10) CHECK (platform IN ('web', 'android', 'ios')),
    is_active       BOOLEAN DEFAULT TRUE,

    created_at      TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, device_token)
);
```

---

## 4. API Design

### 4.1 Base URL & Versioning

```
API:  https://simorqcare.com/api/v1/
Web:  https://simorqcare.com/
```

### 4.2 Authentication Flow

```
1. Register:    POST /api/v1/auth/register
   Body: {national_id, phone, password, first_name, last_name}
   → Sends OTP via SMS

2. Verify OTP:  POST /api/v1/auth/verify-otp
   Body: {phone, code}
   → Returns {user: AuthUser, access_token: string}
      Sets httpOnly refresh_token cookie

3. Login:       POST /api/v1/auth/login
   Body: {phone, password}
   → Returns {user: AuthUser, access_token: string}
      Sets httpOnly refresh_token cookie

4. Refresh:     POST /api/v1/auth/refresh
   (uses httpOnly cookie automatically)
   → Returns {user: AuthUser, access_token: string}

5. Change password: POST /api/v1/auth/change-password
   Body: {current_password, new_password}

6. Intern first login: POST /api/v1/auth/intern-setup
   Body: {full_name, internship_year}
   (only works if profile incomplete + role is intern)
```

**AuthUser shape** (returned in login/register/refresh responses):

```ts
interface AuthUser {
  id: string
  firstName: string
  lastName: string
  phone: string
  avatar?: string
  activeRole: 'client' | 'therapist' | 'admin' | 'owner' | 'intern'
  activeClinicId?: string
  activeClinicName?: string
  roles: Array<{ clinicId: string, clinicName: string, role: string }>
  // Therapist-specific fields (returned when activeRole is therapist/admin/owner):
  education?: string
  approach?: string
  specialties?: string[]
}
```

### 4.3 API Route Map

```
/api/v1/
├── auth/
│   ├── POST   register                    ← {national_id, phone, password, first_name, last_name}
│   ├── POST   verify-otp                  ← {phone, code} → {user, access_token}
│   ├── POST   login                       ← {phone, password} → {user, access_token}
│   ├── POST   refresh                     ← (cookie) → {user, access_token}
│   ├── POST   change-password             ← {current_password, new_password}
│   └── POST   intern-setup                ← {full_name, internship_year}
│
├── users/
│   ├── GET    me                          ← current AuthUser + profile
│   ├── PATCH  me                          ← {first_name, last_name, phone, ...}
│   ├── PUT    me/avatar                   ← multipart upload
│   ├── GET    me/notifications            ← notification prefs
│   └── PATCH  me/notifications            ← update prefs
│
├── clinics/
│   ├── GET    /                           ← list (public, paginated)
│   ├── GET    /:slug                      ← detail (public)
│   ├── POST   /                           ← create clinic (becomes owner)
│   ├── GET    /:id                        ← clinic info (used by panel/settings)
│   ├── PATCH  /:id                        ← update {name, description, address, phone, logo_url}
│   ├── GET    /:id/settings               ← clinic policy settings
│   ├── PATCH  /:id/settings               ← update policy settings
│   ├── GET    /:id/members                ← list members [{userId, firstName, lastName, phone, role}]
│   ├── POST   /:id/members                ← invite/add member {phone, role}
│   ├── PATCH  /:id/members/:mid           ← update role
│   ├── DELETE /:id/members/:mid           ← remove member
│   ├── GET    /:id/permissions            ← list permission overrides
│   └── PATCH  /:id/permissions            ← set permission override {user_id, resource_type, action, granted}
│
├── patients/                              (clinic-scoped via auth context)
│   ├── GET    /                           ← list; ?page, per_page, therapist_id, status,
│   │                                        payment_status, has_discount,
│   │                                        sort=(created_at|appointment_date|start_time), order=(asc|desc)
│   ├── POST   /                           ← create patient
│   ├── GET    /:id                        ← patient detail
│   ├── PATCH  /:id                        ← update patient info
│   │
│   ├── GET    /:id/reports                ← list reports
│   ├── POST   /:id/reports                ← create {title, content, report_date}
│   ├── PATCH  /:id/reports/:rid           ← update
│   ├── DELETE /:id/reports/:rid
│   │
│   ├── GET    /:id/files                  ← list files
│   ├── POST   /:id/files                  ← upload via multipart (or link existing file_key)
│   ├── GET    /:id/files/:fid/download    ← presigned S3 download URL
│   ├── DELETE /:id/files/:fid
│   │
│   ├── GET    /:id/tests                  ← list test records
│   ├── POST   /:id/tests                  ← create {test_name, test_date, ...}
│   ├── PATCH  /:id/tests/:tid
│   │
│   ├── GET    /:id/prescriptions          ← list prescriptions
│   ├── POST   /:id/prescriptions          ← create {title, notes, prescribed_date, file_key?}
│   └── PATCH  /:id/prescriptions/:pid
│
├── appointments/
│   ├── GET    /                           ← list; role-filtered, ?status, date_from, date_to
│   ├── POST   /                           ← book {therapist_id, slot_id, date, start_time, duration_minutes}
│   ├── GET    /:id
│   ├── PATCH  /:id/cancel                 ← {reason?}
│   └── PATCH  /:id/complete              ← mark completed
│
├── therapists/
│   └── GET    /:mid/slots                 ← public; available slots by clinic_member UUID
│
├── schedule/
│   ├── PATCH  /toggle                     ← {enabled: bool} — toggle scheduling on/off
│   ├── GET    /slots                      ← therapist's own slots (panel view)
│   ├── POST   /slots                      ← create {date, start_time, duration_minutes, price}
│   ├── DELETE /slots/:id
│   ├── GET    /recurring                  ← recurring rules
│   ├── POST   /recurring
│   └── DELETE /recurring/:id
│
├── payments/
│   ├── POST   /pay                        ← {appointment_id} → {redirect_url}
│   ├── GET    /verify                     ← ZarinPal callback; ?Authority&Status
│   ├── GET    /transactions               ← transaction history
│   ├── GET    /wallet                     ← {balance, iban?}
│   ├── POST   /wallet/iban               ← {iban}
│   └── POST   /withdraw                  ← initiate withdrawal (uses wallet IBAN)
│
├── conversations/
│   ├── GET    /                           ← list conversations (with unread counts)
│   ├── POST   /                           ← start conversation {participant_id, patient_id?}
│   ├── GET    /:id                        ← conversation detail; returns {patient_id, ...}
│   ├── GET    /:id/messages               ← paginated messages
│   ├── POST   /:id/messages               ← {content?, file_key?, file_name?}
│   └── DELETE /:id/messages/:mid
│
├── tickets/
│   ├── GET    /                           ← list; ?status=(open|answered|closed)
│   ├── POST   /                           ← {subject, content, file_key?}
│   ├── GET    /:id                        ← ticket + messages array
│   ├── PATCH  /:id                        ← {status: 'open'|'closed'}
│   ├── GET    /:id/messages
│   └── POST   /:id/messages               ← {content, file_key?}
│
├── files/
│   ├── POST   /upload                     ← multipart; returns {key, name, size, mime}
│   └── GET    /:key                       ← serve/redirect to presigned S3 URL
│
├── notifications/
│   ├── GET    /                           ← paginated
│   ├── PATCH  /:id/read
│   └── POST   /register-device           ← {device_token, platform}
│
├── contact/
│   └── POST   /                          ← {name, email, subject, message}
│
├── interns/
│   ├── GET    /tasks                      ← intern's own tasks
│   ├── POST   /tasks                      ← submit task
│   ├── GET    /tasks/:id
│   ├── PATCH  /tasks/:id
│   ├── GET    /patients                   ← patients I have access to
│   ├── GET    /roster                     ← list interns in clinic (admin/owner)
│   ├── POST   /:id/grant-access           ← {patient_id, can_view_files, can_write_reports}
│   ├── DELETE /:id/revoke-access/:pid
│   ├── GET    /:id/tasks                  ← view intern's tasks
│   └── PATCH  /:id/tasks/:tid/review      ← {review_status, review_comment, grade}
│
├── tests/
│   ├── GET    /                           ← platform test catalog
│   └── GET    /:id
│
└── admin/
    ├── GET    /clinics
    ├── PATCH  /clinics/:id/verify
    ├── GET    /commissions
    ├── POST   /commissions
    ├── GET    /withdrawals
    └── PATCH  /withdrawals/:id
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

## 6. Go Backend Structure

```
simorq_backend/              ← module: github.com/Alijeyrad/simorq_backend
├── main.go                  ← Cobra root entrypoint
├── cmd/
│   ├── root.go
│   ├── http/start.go        ← boots Fiber app via Uber Fx
│   └── system/
│       ├── migrate.go       ← Ent auto-migrate
│       ├── init.go          ← DB initialization
│       └── gendocs.go       ← Cobra CLI docs
│
├── config/
│   ├── config.go            ← Viper loader
│   └── types.go             ← Config struct + sub-structs (Database, Redis, ZarinPal, S3, …)
│
├── internal/
│   ├── app/
│   │   ├── infra.go         ← InfraModule (Fx): DB, Redis, Casbin, S3, ZarinPal, SMS, email, OTel
│   │   └── services.go      ← ServiceModule (Fx): all domain services
│   │
│   ├── schema/              ← Ent schemas — SOURCE OF TRUTH for DB structure
│   │   ├── mixin.go         ← UUIDV7Mixin, TimeStampedMixin, CreatedAtMixin, SoftDeleteMixin
│   │   ├── user.go
│   │   ├── clinic.go
│   │   ├── clinic_member.go
│   │   ├── therapist_profile.go
│   │   ├── patient.go
│   │   ├── patient_assessment.go   ← struct PatientTest (NOT _test.go, avoids Go tooling conflict)
│   │   ├── patient_report.go
│   │   ├── patient_file.go
│   │   ├── patient_prescription.go
│   │   ├── psych_catalog.go        ← struct PsychTest (platform test catalog)
│   │   ├── time_slot.go
│   │   ├── recurring_rule.go
│   │   ├── appointment.go
│   │   ├── wallet.go
│   │   ├── transaction.go
│   │   ├── payment_request.go
│   │   ├── withdrawal_request.go
│   │   └── commission_rule.go
│   │
│   ├── repo/                ← Ent GENERATED code — DO NOT READ (infer from schemas above)
│   │
│   ├── service/
│   │   ├── auth/            ← OTP, PASETO token issue/refresh, change-password
│   │   ├── user/            ← user profile, avatar
│   │   ├── clinic/          ← clinic CRUD, members, permissions
│   │   ├── patient/         ← patient CRUD, reports, files, prescriptions, tests
│   │   ├── file/            ← S3 upload + presigned URL generation
│   │   ├── psychtest/       ← psych test catalog
│   │   ├── scheduling/      ← slots, recurring rules, schedule toggle
│   │   ├── appointment/     ← book, cancel, complete
│   │   └── payment/         ← ZarinPal initiate/verify, wallet, IBAN, withdrawal
│   │
│   └── api/http/
│       ├── handler/         ← one file per domain (auth, user, clinic, patient, file,
│       │                       test, schedule, appointment, payment)
│       ├── middleware/      ← AuthRequired, ClinicContext, ClinicHeader, RequirePermission
│       └── router/          ← route registration split by domain; router.go holds Params struct
│
├── pkg/
│   ├── authorize/           ← Casbin RBAC: IAuthorization, resource/action constants, seed policies
│   ├── paseto/              ← PASETO v4 token manager
│   ├── crypto/              ← AES-256-GCM encryption (national_id, IBAN)
│   ├── database/            ← Ent client factory + DSN builder
│   ├── redis/               ← Redis client factory
│   ├── s3/                  ← ArvanCloud S3 client + presigned URL helpers
│   ├── zarinpal/            ← ZarinPal v4 HTTP client (no SDK; custom net/http)
│   ├── sms/                 ← SMS provider client
│   ├── email/               ← email client
│   └── observability/       ← OpenTelemetry provider (traces + metrics)
│
├── docs/
│   └── zarinpal.md          ← ZarinPal API reference + implementation notes
│
├── Makefile                 ← build, run, test, entgen, migrate, dev, db-*, redis-*, help
├── Dockerfile
├── Dockerfile.dev           ← with Air hot reload
├── docker-compose.yml
├── config.yaml
├── CLAUDE.md
├── go.mod
└── go.sum
```

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

---

## 7. Nuxt 4 Frontend Structure

### 7.1 Actual File Tree

```
app/
├── app.config.ts
├── app.vue
│
├── layouts/
│   ├── default.vue              ← public pages (AppHeader + AppFooter)
│   ├── auth.vue                 ← login/signup/payments/result
│   ├── dashboard.vue            ← client area (top nav + sidebar)
│   ├── panel.vue                ← clinic panel (top nav + sidebar + role switcher)
│   └── docs.vue                 ← (template remnant, not used by Simorq pages)
│
├── pages/
│   │
│   │  ── PUBLIC (prerendered) ──
│   ├── index.vue                ← homepage (content: 0.index.yml)
│   ├── about.vue                ← درباره ما (content: 9.about.yml)
│   ├── contact.vue              ← تماس با ما; POST /api/v1/contact
│   ├── faq.vue                  ← سوالات متداول (content: 6.faq.yml)
│   ├── privacy.vue              ← حریم خصوصی (content: 7.privacy.yml)
│   ├── terms.vue                ← قوانین (content: 8.terms.yml)
│   ├── therapists/
│   │   ├── index.vue            ← browse therapists (content: 5.therapists/*.yml)
│   │   └── [id].vue             ← therapist profile; slots from GET /therapists/:slug/slots
│   └── clinics/
│       ├── index.vue            ← clinic listing (content: 11.clinics/*.yml)
│       └── [id].vue             ← clinic public page
│   │
│   │  ── AUTH (ssr: false) ──
│   ├── login.vue                ← POST /auth/login + POST /auth/verify-otp
│   ├── signup.vue               ← POST /auth/register → OTP step
│   ├── payments/
│   │   └── result.vue           ← after ZarinPal redirect; reads ?status&ref
│   │
│   │  ── DASHBOARD (client, layout: dashboard, middleware: auth) ──
│   ├── dashboard/
│   │   ├── index.vue            ← نوبت‌های رزرو شده; GET /appointments?status=reserved
│   │   ├── finalize.vue         ← نهایی‌کردن نوبت; POST /payments/pay → redirect
│   │   ├── history.vue          ← نوبت‌های گذشته; GET /appointments?status=completed
│   │   ├── messages/
│   │   │   ├── index.vue        ← GET /conversations
│   │   │   └── [id].vue         ← GET+POST /conversations/:id/messages
│   │   ├── payments.vue         ← GET /payments/transactions + /payments/wallet
│   │   ├── profile.vue          ← GET+PATCH /users/me
│   │   ├── tickets/
│   │   │   ├── index.vue        ← GET /tickets; POST /tickets; PATCH /tickets/:id
│   │   │   └── [id].vue         ← GET /tickets/:id; POST /tickets/:id/messages
│   │   └── settings/
│   │       ├── index.vue        ← settings hub (tabs link to sub-pages)
│   │       ├── notifications.vue ← GET+PATCH /users/me/notifications
│   │       └── security.vue     ← POST /auth/change-password
│   │
│   │  ── PANEL (client, layout: panel, middleware: role) ──
│   └── panel/
│       ├── index.vue            ← نوبت‌های رزرو شده; GET /appointments (therapist-scoped)
│       ├── patients/
│       │   ├── index.vue        ← پرونده مراجعان; GET /patients with filters
│       │   └── [id].vue         ← patient detail; tabs: reports, files, tests, prescriptions
│       ├── history.vue          ← نوبت‌های گذشته
│       ├── messages/
│       │   ├── index.vue        ← GET /conversations
│       │   └── [id].vue         ← GET /conversations/:id → patient context; messages
│       ├── finances.vue         ← GET /payments/wallet + /transactions; IBAN; withdraw
│       ├── profile.vue          ← GET+PATCH /users/me (therapist fields)
│       ├── schedule.vue         ← PATCH /schedule/toggle; GET+POST+DELETE /schedule/slots
│       ├── members.vue          ← GET+POST+DELETE /clinics/:id/members
│       ├── permissions.vue      ← GET+PATCH /clinics/:id/permissions
│       ├── settings.vue         ← GET+PATCH /clinics/:id (clinic info)
│       └── tickets/
│           ├── index.vue        ← same as dashboard tickets
│           └── [id].vue         ← same as dashboard ticket detail
│
├── components/
│   ├── AppHeader.vue            ← public header with nav + auth links
│   ├── AppFooter.vue            ← links to /about /contact /privacy /terms /faq
│   ├── AppLogo.vue
│   ├── HeroBackground.vue
│   ├── ImagePlaceholder.vue
│   ├── StarsBg.vue
│   ├── OgImage/OgImageSaas.vue  ← OG image template
│   ├── settings/
│   │   └── MembersList.vue      ← (template remnant — may be removed)
│   └── shared/                  ← auto-imported as Shared*
│       ├── AppointmentCard.vue  ← shows "حضوری" badge, reservation/session payment status
│       ├── CommentsSection.vue  ← therapist/clinic public ratings
│       ├── ConfirmModal.vue     ← destructive action confirmation
│       ├── FileUploader.vue     ← drag-drop upload; POST /files/upload
│       ├── JalaliDatePicker.vue ← Jalali calendar input
│       ├── PatientListItem.vue  ← row in patient list
│       ├── TherapistCard.vue    ← therapist card (public listing)
│       ├── TransactionCard.vue
│       └── WalletCard.vue
│
├── composables/
│   ├── useAuth.ts               ← login, logout, hasRole
│   ├── useClinic.ts             ← switchRole, availableRoles, clinicId
│   ├── useAppointments.ts       ← appointment fetch helpers
│   ├── useDashboard.ts
│   ├── useJalali.ts             ← persianNumber(), formatCurrency(), formatJalali()
│   ├── useMessages.ts           ← conversation/message helpers
│   └── usePermissions.ts        ← client-side role check helpers
│
├── middleware/
│   ├── auth.ts                  ← redirect /login if no token (client-only)
│   └── role.ts                  ← check hasClinicRole for /panel/**
│
├── stores/
│   ├── auth.ts                  ← {user: AuthUser, accessToken} + login/register/verify/refresh
│   └── clinic.ts                ← {isSchedulingEnabled} + switchRole, clinicId getter
│
├── plugins/
│   └── api.ts                   ← $api = $fetch wrapper with Authorization header + error handling
│
├── utils/
│   ├── errors.ts                ← getErrorMessage(e, fallback)
│   └── random.ts
│
└── types/index.d.ts             ← AuthUser, Appointment, Patient, Therapist, Clinic,
                                    TherapistSlot, ClinicSettings, ClinicMember,
                                    Message, Conversation, Transaction, Ticket,
                                    ScheduleSlot, Wallet, PatientReport, PatientFile
```

### 7.2 Content Collections (`content.config.ts`)

| Collection     | Source                 | Type | Used by                               |
| -------------- | ---------------------- | ---- | ------------------------------------- |
| `index`      | `0.index.yml`        | page | `/`                                 |
| `therapists` | `5.therapists/*.yml` | data | `/therapists`, `/therapists/[id]` |
| `faq`        | `6.faq.yml`          | data | `/faq`                              |
| `privacy`    | `7.privacy.yml`      | data | `/privacy`                          |
| `terms`      | `8.terms.yml`        | data | `/terms`                            |
| `about`      | `9.about.yml`        | data | `/about`                            |
| `contact`    | `10.contact.yml`     | data | `/contact` (info section only)      |
| `clinics`    | `11.clinics/*.yml`   | data | `/clinics`, `/clinics/[id]`       |

> Template collections (`docs`, `pricing`, `blog/posts`, `changelog/versions`) remain in `content.config.ts` but are not used by Simorq pages and can be removed during cleanup.

### 7.3 Key nuxt.config.ts Settings

```ts
modules: [
  '@nuxt/eslint',
  '@nuxt/image',
  '@nuxt/ui',              // NOT ui-pro — base Nuxt UI
  '@nuxtjs/google-fonts',  // Vazirmatn (not local fonts)
  '@nuxt/content',
  '@vueuse/nuxt',
  'nuxt-og-image',
  '@pinia/nuxt',
  'nuxt-auth-utils',
  '@vite-pwa/nuxt'         // PWA support
]

routeRules: {
  // Client-only (auth state is client-side Pinia only):
  '/login':        { ssr: false },
  '/dashboard/**': { ssr: false },
  '/panel/**':     { ssr: false },
  '/payments/**':  { ssr: false },

  // Prerendered at build time:
  '/':             { prerender: true },
  '/therapists':   { prerender: true },
  '/therapists/**': { prerender: true },
  '/clinics':      { prerender: true },
  '/clinics/**':   { prerender: true },
  '/faq':          { prerender: true },
  '/privacy':      { prerender: true },
  '/terms':        { prerender: true },
  '/about':        { prerender: true },
  '/contact':      { prerender: true },

  // Security headers on all routes:
  '/**': { headers: { 'X-Frame-Options': 'DENY', ... } }
}

runtimeConfig: {
  public: { apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8000' }
}
```

### 7.4 Design System Quick Reference

| Token              | Value                              | Usage                           |
| ------------------ | ---------------------------------- | ------------------------------- |
| `--ui-primary`   | `#00C16A`                        | Buttons, active nav, highlights |
| `--ui-secondary` | `#8B5CF6`                        | Role switcher, special badges   |
| Font               | Vazirmatn (Google Fonts)           | All text                        |
| Direction          | RTL globally                       | `html[dir=rtl]`               |
| Numbers            | `persianNumber()`                | All displayed numbers           |
| Dates              | `formatJalali()`                 | All displayed dates             |
| Currency           | `formatCurrency()`               | `x٬xxx٬xxx ریال`        |
| Icons              | Heroicons only (`i-heroicons-*`) |                                 |

**Sessions:** In-person only (`حضوری`). No video, no join button anywhere.

---

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

## 9. Server-Side Routes (Nuxt)

```
server/
├── middleware/
│   └── csp.ts           ← Sets Content-Security-Policy header dynamically
└── routes/
    ├── robots.txt.ts    ← Blocks /dashboard/, /panel/, /payments/, /login, /signup
    └── sitemap.xml.ts   ← Dynamic XML; uses queryCollection() for therapist/clinic slugs
```

---

## 10. Deployment Architecture

### 10.1 Docker Compose (Production)

```yaml
services:
  caddy:
    image: caddy:2-alpine
    ports: ["80:80", "443:443"]

  web:
    build: ./simorgh-web
    expose: ["3000"]
    environment:
      NUXT_PUBLIC_API_BASE: https://simorqcare.com/api/v1

  api:
    build: ./simorgh-api
    expose: ["8080"]
    environment:
      DB_URL: postgres://simorgh:xxx@postgres:5432/simorgh
      REDIS_URL: redis://redis:6379
      NATS_URL: nats://nats:4222
      S3_ENDPOINT: https://s3.ir-thr-at1.arvanstorage.ir
      S3_BUCKET: simorgh-files
      ZARINPAL_MERCHANT: ${ZARINPAL_MERCHANT}
      PASETO_KEY: ${PASETO_KEY}
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}

  postgres:
    image: postgres:15-alpine
    # NOT exposed to host

  redis:
    image: redis:7-alpine

  nats:
    image: nats:2-alpine
    command: ["--jetstream", "--store_dir=/data"]
```

### 10.2 Caddyfile

```
simorqcare.com {
    handle /api/* {
        reverse_proxy api:8080
    }
    handle {
        reverse_proxy web:3000
    }
    header {
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        Referrer-Policy strict-origin-when-cross-origin
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
    }
}
```

---

## 11. Phased Delivery Roadmap

### Completed (Frontend UI + API both done)

- [X] All layouts: default, auth, dashboard, panel
- [X] Public pages: home, therapists, therapist profile, clinics, clinic detail, faq, privacy, terms, about, contact
- [X] Login + signup pages
- [X] Dashboard: appointments (upcoming, finalize, history), messages, payments, profile, tickets, settings (notifications, security)
- [X] Panel: appointments, patient list, patient detail (reports, files, tests, prescriptions), messages, finances, profile, schedule, members, permissions, settings, tickets
- [X] Shared components: AppointmentCard, TherapistCard, ConfirmModal, JalaliDatePicker, FileUploader, WalletCard, TransactionCard

### ✅ Phase 1 — Go API Foundation (COMPLETE)

- [x] Project scaffold (Go + Fiber v3 + Uber Fx + Ent ORM)
- [x] Docker Compose dev environment
- [x] Ent auto-migrate (no SQL migration files)
- [x] User auth: register, OTP verify, login, refresh, change-password (PASETO v4)
- [x] Casbin v2 setup with role model + clinic tenant isolation
- [x] Basic clinic CRUD + settings
- [x] Clinic member management (list, add, remove, update role)
- [x] Clinic permissions endpoint
- [x] Therapist profile management

### ✅ Phase 2 — Clinical Core (COMPLETE)

- [x] Patient CRUD (all fields including child psych)
- [x] Patient reports, files (S3 upload + presigned download), prescriptions, tests
- [x] File upload unified endpoint (`POST /files/upload`)
- [x] File serve endpoint (`GET /files/:key`)
- [x] Psych test catalog (`GET /tests`, `GET /tests/:id`)

### ✅ Phase 3 — Scheduling & Payments (COMPLETE)

- [x] Time slot CRUD + recurring rules
- [x] Schedule toggle (`PATCH /schedule/toggle`) — updates `therapist_profiles.is_accepting`
- [x] Public slot listing (`GET /therapists/:mid/slots`) — uses clinic_member UUID
- [x] Appointment booking + cancellation + completion
- [x] Optimistic slot locking on booking (atomic UPDATE WHERE status='available')
- [x] ZarinPal v4 integration (`POST /payments/pay` + `GET /payments/verify` callback)
- [x] Wallet (get/create) + transaction history
- [x] IBAN management (AES-256-GCM encrypted) + withdrawal requests
- [ ] Commission calculation worker (deferred to Phase 4 NATS)

### Phase 4 — Communication

- [ ] Conversations + messages (polling)
- [ ] Patient context on conversation detail
- [ ] Ticket system (create, reply, status change)
- [ ] File attachments in messages + tickets
- [ ] Notification records + `GET /users/me/notifications` prefs
- [ ] PWA push notification registration (`POST /notifications/register-device`)
- [ ] NATS workers: notification, SMS, wallet

### Phase 5 — User Settings & Misc

- [ ] `GET/PATCH /users/me/notifications` (prefs)
- [ ] `POST /auth/change-password`
- [ ] Contact form (`POST /contact`)
- [ ] Intern module (profiles, tasks, patient access)

### Phase 6 — Polish & Ship

- [ ] Admin panel endpoints (`/admin/*`)
- [ ] Sitemap: wire to real DB query (therapists + clinics)
- [ ] Load testing
- [ ] Security audit
- [ ] Deploy to ArvanCloud VPS

### Post-Launch (Backlog)

- [ ] Real-time chat via NATS/WebSocket
- [ ] In-app psychological test conducting
- [ ] Video sessions
- [ ] Mobile apps (Nuxt + Capacitor)
- [ ] Advanced analytics dashboard
- [ ] Insurance integration

---

## 12. Key Technical Decisions

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
