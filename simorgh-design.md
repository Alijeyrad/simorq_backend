# Simorgh — Complete Platform Design Document

> **Version:** 2.0
> **Date:** February 2026
> **Domain:** simorqcare.com
> **Stack:** Nuxt 4 + Nuxt UI · Go + Fiber v3 · PostgreSQL · Redis · NATS · Casbin · PASETO · ZarinPal · ArvanCloud S3 + VPS

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
CREATE TABLE time_slots (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id           UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    therapist_id        UUID NOT NULL REFERENCES clinic_members(id),

    slot_date           DATE NOT NULL,
    start_time          TIME NOT NULL,
    end_time            TIME NOT NULL,
    duration_minutes    INTEGER NOT NULL DEFAULT 60,
    price               BIGINT NOT NULL,   -- رزرو نوبت fee (Rials)

    is_available        BOOLEAN DEFAULT TRUE,
    is_booked           BOOLEAN DEFAULT FALSE,

    recurring_rule_id   UUID REFERENCES recurring_rules(id),

    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_time_slots_therapist_date ON time_slots(therapist_id, slot_date);
CREATE INDEX idx_time_slots_clinic_date ON time_slots(clinic_id, slot_date);
CREATE INDEX idx_time_slots_available ON time_slots(clinic_id, is_available, is_booked, slot_date);
```

#### **recurring_rules**

```sql
CREATE TABLE recurring_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id       UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    therapist_id    UUID NOT NULL REFERENCES clinic_members(id),

    day_of_week     INTEGER NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Saturday
    start_time      TIME NOT NULL,
    end_time        TIME NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 60,
    price           BIGINT NOT NULL,

    effective_from  DATE NOT NULL,
    effective_until DATE,
    is_active       BOOLEAN DEFAULT TRUE,

    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

#### **appointments**

```sql
CREATE TABLE appointments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id           UUID NOT NULL REFERENCES clinics(id),
    patient_id          UUID NOT NULL REFERENCES patients(id),
    therapist_id        UUID NOT NULL REFERENCES clinic_members(id),
    time_slot_id        UUID REFERENCES time_slots(id),

    appointment_date    DATE NOT NULL,
    start_time          TIME NOT NULL,
    duration_minutes    INTEGER NOT NULL,
    price               BIGINT NOT NULL,   -- reservation fee paid

    status              VARCHAR(30) NOT NULL DEFAULT 'pending_payment'
                        CHECK (status IN (
                            'pending_payment',
                            'reserved',
                            'completed',
                            'cancelled_by_client',
                            'cancelled_by_therapist',
                            'no_show'
                        )),

    reservation_fee     BIGINT DEFAULT 0,
    reservation_paid    BOOLEAN DEFAULT FALSE,
    session_paid        BOOLEAN DEFAULT FALSE,  -- settled directly with therapist

    session_number      INTEGER,

    cancelled_at        TIMESTAMPTZ,
    cancel_reason       TEXT,
    cancellation_fee    BIGINT DEFAULT 0,

    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_appointments_clinic ON appointments(clinic_id);
CREATE INDEX idx_appointments_patient ON appointments(patient_id);
CREATE INDEX idx_appointments_therapist ON appointments(therapist_id, appointment_date);
CREATE INDEX idx_appointments_status ON appointments(clinic_id, status);
CREATE INDEX idx_appointments_date ON appointments(appointment_date);
```

#### **wallets**

```sql
CREATE TABLE wallets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_type      VARCHAR(20) NOT NULL CHECK (owner_type IN ('user', 'clinic', 'platform')),
    owner_id        UUID NOT NULL,
    balance         BIGINT NOT NULL DEFAULT 0,
    iban            VARCHAR(26),   -- شماره شبا (AES-256-GCM encrypted)

    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(owner_type, owner_id)
);
```

#### **transactions**

```sql
CREATE TABLE transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id           UUID NOT NULL REFERENCES wallets(id),

    type                VARCHAR(20) NOT NULL CHECK (type IN (
                            'deposit', 'withdrawal', 'commission',
                            'reservation_fee', 'cancellation_fee', 'refund'
                        )),
    amount              BIGINT NOT NULL,
    direction           VARCHAR(6) NOT NULL CHECK (direction IN ('credit', 'debit')),
    balance_after       BIGINT NOT NULL,

    appointment_id      UUID REFERENCES appointments(id),
    payment_request_id  UUID,

    description         TEXT,
    status              VARCHAR(20) DEFAULT 'completed'
                        CHECK (status IN ('pending', 'completed', 'failed')),

    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_transactions_wallet ON transactions(wallet_id);
CREATE INDEX idx_transactions_created ON transactions(created_at);
```

#### **payment_requests** (ZarinPal)

```sql
CREATE TABLE payment_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    appointment_id  UUID REFERENCES appointments(id),

    amount          BIGINT NOT NULL,
    authority       VARCHAR(100),
    ref_id          VARCHAR(100),
    status          VARCHAR(20) DEFAULT 'pending'
                    CHECK (status IN ('pending', 'paid', 'failed', 'refunded')),
    gateway_data    JSONB,

    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
```

#### **withdrawal_requests**

```sql
CREATE TABLE withdrawal_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id       UUID NOT NULL REFERENCES wallets(id),
    amount          BIGINT NOT NULL,
    iban            VARCHAR(26) NOT NULL,
    status          VARCHAR(20) DEFAULT 'pending'
                    CHECK (status IN ('pending', 'processing', 'completed', 'rejected')),
    processed_at    TIMESTAMPTZ,
    admin_note      TEXT,

    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

#### **commission_rules**

```sql
CREATE TABLE commission_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clinic_id       UUID REFERENCES clinics(id),   -- NULL = platform-wide default
    percentage      NUMERIC(5,2) NOT NULL,
    flat_fee        BIGINT DEFAULT 0,
    effective_from  DATE NOT NULL,
    effective_until DATE,
    is_active       BOOLEAN DEFAULT TRUE,

    created_at      TIMESTAMPTZ DEFAULT NOW()
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
│   └── GET    /:slug/slots                ← public; available slots for booking
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
simorgh-api/
├── cmd/api/main.go
├── internal/
│   ├── config/config.go
│   ├── server/
│   │   ├── server.go
│   │   ├── routes.go
│   │   └── middleware/
│   │       ├── auth.go          ← PASETO validation
│   │       ├── rbac.go          ← Casbin enforcement
│   │       ├── tenant.go        ← clinic context from JWT
│   │       ├── ratelimit.go
│   │       └── cors.go
│   ├── domain/
│   │   ├── user.go
│   │   ├── clinic.go
│   │   ├── patient.go
│   │   ├── appointment.go
│   │   ├── wallet.go
│   │   └── ...
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── clinic_handler.go
│   │   ├── patient_handler.go
│   │   ├── appointment_handler.go
│   │   ├── schedule_handler.go
│   │   ├── payment_handler.go
│   │   ├── chat_handler.go
│   │   ├── ticket_handler.go
│   │   ├── file_handler.go
│   │   ├── contact_handler.go
│   │   ├── intern_handler.go
│   │   ├── notification_handler.go
│   │   └── admin_handler.go
│   ├── service/
│   │   ├── auth_service.go
│   │   ├── patient_service.go
│   │   ├── booking_service.go
│   │   ├── payment_service.go
│   │   ├── file_service.go
│   │   ├── notification_service.go
│   │   └── ...
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── clinic_repo.go
│   │   ├── patient_repo.go
│   │   └── ...
│   ├── worker/
│   │   ├── notification_worker.go
│   │   ├── sms_worker.go
│   │   └── wallet_worker.go
│   └── pkg/
│       ├── paseto/
│       ├── crypto/             ← AES-256-GCM for national_id, IBAN
│       ├── jalali/
│       ├── zarinpal/
│       ├── s3/                 ← ArvanCloud S3 client + presigned URLs
│       ├── sms/
│       ├── validator/
│       └── pagination/
├── migrations/
├── casbin/model.conf + policy.csv
├── docker-compose.yml
├── docker-compose.prod.yml
├── Dockerfile
├── Caddyfile
├── go.mod
└── go.sum
```

### Key Go Dependencies

```
github.com/gofiber/fiber/v3
github.com/jackc/pgx/v5
github.com/redis/go-redis/v9
github.com/nats-io/nats.go
github.com/casbin/casbin/v2
github.com/casbin/pgx-adapter
github.com/vk-rv/pvx                  ← PASETO v4
github.com/aws/aws-sdk-go-v2          ← S3-compatible (ArvanCloud)
github.com/golang-migrate/migrate/v4
github.com/go-playground/validator/v10
golang.org/x/crypto                   ← argon2id
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

### Completed (Frontend UI done, API needed)

- [X] All layouts: default, auth, dashboard, panel
- [X] Public pages: home, therapists, therapist profile, clinics, clinic detail, faq, privacy, terms, about, contact
- [X] Login + signup pages
- [X] Dashboard: appointments (upcoming, finalize, history), messages, payments, profile, tickets, settings (notifications, security)
- [X] Panel: appointments, patient list, patient detail (reports, files, tests, prescriptions), messages, finances, profile, schedule, members, permissions, settings, tickets
- [X] Shared components: AppointmentCard, TherapistCard, ConfirmModal, JalaliDatePicker, FileUploader, WalletCard, TransactionCard

### Phase 1 — Go API Foundation

- [ ] Project scaffold (Go + Fiber v3)
- [ ] Docker Compose dev environment
- [ ] Database migrations (all tables above)
- [ ] User auth: register, OTP verify, login, refresh, change-password (PASETO)
- [ ] Casbin setup with role model + clinic tenant isolation
- [ ] Basic clinic CRUD + settings
- [ ] Clinic member management (list, add, remove)
- [ ] Clinic permissions endpoint

### Phase 2 — Clinical Core

- [ ] Patient CRUD (all fields including child psych)
- [ ] Patient reports, files (S3 upload + presigned download), prescriptions, tests
- [ ] File upload unified endpoint (`POST /files/upload`)
- [ ] File serve endpoint (`GET /files/:key`)

### Phase 3 — Scheduling & Payments

- [ ] Time slot CRUD + recurring rules + slot generation
- [ ] Schedule toggle (`PATCH /schedule/toggle`)
- [ ] Public slot listing (`GET /therapists/:slug/slots`)
- [ ] Appointment booking + cancellation + completion
- [ ] ZarinPal integration (`POST /payments/pay` + verify callback)
- [ ] Wallet + transaction history
- [ ] IBAN management + withdrawal requests
- [ ] Commission calculation (NATS wallet-worker)

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
| ORM vs raw SQL | Raw SQL + pgx                 | Full control; sqlc optional                           |
| File storage   | ArvanCloud S3                 | Private ACL + presigned URLs                          |
| Date handling  | Gregorian in DB, Jalali in UI | Standard DB ops, convert at display layer             |
| Multi-tenancy  | Shared DB + clinic_id column  | Casbin enforces isolation                             |
| Real-time (v1) | HTTP polling                  | Simplest; NATS ready for upgrade                      |
| i18n           | Static Persian text           | No runtime locale switching needed                    |
| PWA            | vite-pwa + Workbox            | Offline shell; push notification support              |
| SMS provider   | TBD (Kavenegar / Ghasedak)    | Iranian provider required                             |
| Sessions       | In-person only                | No video infrastructure needed for v1                 |
