# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Module

`github.com/Alijeyrad/simorq_backend`

---

## Common Commands

| Task | Command |
|---|---|
| Run app locally | `make run` (or `make run CONFIG=./config.yaml`) |
| Full dev environment (Docker + hot reload) | `make dev` |
| Build binary | `make build` |
| Run tests | `make test` |
| Run single test | `go test -v ./internal/service/payment/...` |
| Format | `make fmt` |
| Vet | `make vet` |
| **Generate Ent code** | `make entgen` |
| **Run DB migrations** | `make migrate` |
| Start Postgres only | `make db-start` |
| Start Redis only | `make redis-start` |
| Open psql shell | `make db-shell` |
| Stop everything | `make down` |

> Always use `make entgen` (not `go generate`) and `make migrate` (not raw SQL). The DB must be running before `make migrate`.

---

## Architecture Overview

### Entry Point & CLI

`main.go` → Cobra CLI in `cmd/`. Two relevant commands:
- `cmd/http/start.go` — boots the Fiber v3 HTTP server via Uber Fx
- `cmd/system/migrate.go` — runs Ent auto-migrate

### Dependency Injection (Uber Fx)

All wiring lives in `internal/app/`:
- `infra.go` — `InfraModule`: DB client, Redis, Casbin enforcer, S3, ZarinPal, email, SMS, observability
- `services.go` — `ServiceModule`: all domain services
- `internal/api/http/router/router.go` — `Params` struct receives all services via `fx.In`

When adding a new service: add provider to `services.go`, add field to `router.go`'s `Params`.

### Database (Ent ORM)

**Schemas** live in `internal/schema/*.go` — these are the source of truth.
**NEVER read `internal/repo/**`** — the generated code is very large and auto-synced. Infer generated types from the schemas instead:
- Schema struct name → generated type name (e.g. `type Patient struct` → `repo.Patient`)
- Filename does NOT matter; struct name does (see `patient_assessment.go` which contains `PatientTest` struct → `repo.PatientTest`)
- File names ending in `_test.go` are skipped by Go tooling — rename to `_catalog.go` or similar

**Mixins available** (`internal/schema/mixin.go`):
- `UUIDV7Mixin{}` — adds `id uuid` (UUIDv7, primary key)
- `TimeStampedMixin{}` — adds `created_at`, `updated_at`
- `CreatedAtMixin{}` — adds `created_at` only (for immutable/append-only records)
- `SoftDeleteMixin{}` — adds `deleted_at` with soft-delete interceptors

**Ordering**: use `sql.OrderDesc()` / `sql.OrderAsc()` from `entgo.io/ent/dialect/sql`:
```go
repo.ByCreatedAt(sql.OrderDesc())
```

After writing or modifying schemas: run `make entgen`, then `make migrate`.

### HTTP Layer (Fiber v3)

- Handlers in `internal/api/http/handler/<name>.go`
- Routers in `internal/api/http/router/<name>.go` (registered from `router.go`'s `Register()`)
- Middleware in `internal/api/http/middleware/`

**Middleware locals keys** (set by `ClinicContext` / `ClinicHeader`):
- `middleware.LocalsClinicID` → `"clinic_id"`
- `middleware.LocalsMemberRole` → `"member_role"`
- `middleware.LocalsMemberID` → `"member_id"` (this is `clinic_members.id`, not `users.id`)

**Critical handler convention**: never name a local variable `ok` — it shadows the `ok(c, data)` response helper. Use `valid`, `found`, etc. instead.

### Services

Each service lives in `internal/service/<name>/<name>.go` with:
- An exported `Service` interface
- A concrete struct implementing it
- A `New(...)` constructor
- `errors.go` with sentinel errors

### Authorization (Casbin RBAC)

Domain-scoped policies: `clinic:<uuid>`. Enforcer setup in `pkg/authorize/`.

Resource constants in `pkg/authorize/constants.go` (e.g. `ResourcePatient`, `ResourceAppointment`, `ResourceSchedule`, `ResourceWallet`, etc.). Seed policies in `pkg/authorize/seed.go`.

Use `middleware.RequirePermission(auth, resource, action)` in route registration.

### Key Packages

| Package | Purpose |
|---|---|
| `pkg/authorize` | Casbin RBAC — `IAuthorization`, `NewEnforcer`, resource/action constants |
| `pkg/paseto` | PASETO v4 token manager |
| `pkg/crypto` | AES-256-GCM encryption (used for national_id, IBAN) |
| `pkg/s3` | S3-compatible file storage |
| `pkg/zarinpal` | ZarinPal payment gateway HTTP client |
| `pkg/sms` | SMS provider |
| `pkg/email` | Email client |
| `config/types.go` | Central config structs — all infra config lives here |

### Config

`config.yaml` → loaded into `config.Config` via Viper. Struct in `config/types.go`. Config key for each infra client matches the mapstructure tag on its config sub-struct.

---

## Conventions

- All monetary amounts stored in **Rials (IRR)** as `int64`
- All entity IDs are **UUIDv7** — use `uuid.NewV7()` for new records
- Encrypted sensitive fields (national_id, IBAN): AES-256-GCM via `pkg/crypto`, SHA-256 hash stored alongside for lookup
- Clinic multi-tenancy: all clinic-scoped routes include `X-Clinic-ID` header → `ClinicHeader` middleware populates locals
- Appointments use snapshot pricing — `session_price` copied from slot at booking time
