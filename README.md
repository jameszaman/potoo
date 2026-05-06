# NotifyLayer

A developer-first notification orchestration platform.

One API and one dashboard for email, push, SMS, WhatsApp, webhooks, and in-app notifications — with templates, routing, retries, preferences, logs, and provider failover. Bring your own providers.

---

## 1. What This Is

An orchestration layer that sits above delivery providers. It does not own SMTP, SMS telecom routing, or push gateways. It makes SendGrid, Resend, SES, Twilio, and FCM easier to use together.

```txt
Customer app
  -> NotifyLayer API
  -> Template rendering
  -> Preference checks
  -> Routing rules
  -> Queue and retry
  -> Provider adapter
  -> SendGrid / Resend / SES / Twilio / FCM / webhook
  -> Delivery events
  -> Logs, replay, webhooks
```

---

## 2. Product Focus

**Target:** SaaS teams with 10–200 employees where engineering owns customer notifications and has pain around logs, retries, templates, and multiple providers.

**Initial use cases:**

```txt
Welcome emails
Password resets
OTP and verification
Billing notifications
Product activity alerts
Team invites and mentions
Webhook notifications
Push notifications with email fallback
```

**Avoid initially:**

```txt
Bulk newsletters
Cold email
High-volume marketing campaigns
```

---

## 3. Differentiation

```txt
We make customer notifications reliable, observable, and easy to manage across all of your providers.
```

1. Provider-agnostic routing
2. Bring-your-own-provider setup
3. OpenAPI-first developer experience
4. Logs, traces, and replay for every notification
5. Idempotent send API
6. Template versioning and environment promotion
7. Preferences and suppression lists
8. Reliable queue, retry, and failure handling
9. Webhook ingestion and customer webhook fanout
10. AI-assisted debugging and template QA (later)

---

## 4. Technology Stack

```txt
Frontend:         Next.js + React + TypeScript
Backend:          Go 1.22+ + chi
API contract:     OpenAPI-first (spec in api/openapi.yaml)
Code generation:  oapi-codegen (Go types + server interface)
Docs UI:          Swagger UI at /docs
Database:         PostgreSQL 16+
Database access:  sqlc + pgx/v5
Migrations:       goose
Queue:            Redis + Asynq
Billing:          Stripe
Dashboard auth:   Clerk
Observability:    OpenTelemetry + structured logs
Deployment:       Docker Compose locally, managed cloud later
```

Minimum Go version is **1.22**. Set it explicitly in `go.mod` and in any Dockerfiles.

---

## 5. Architecture

```txt
                          +----------------------+
                          |      Dashboard       |
                          | Next.js + TypeScript |
                          +----------+-----------+
                                     |
                                     v
+----------------+      +-----------+------------+
| Customer Apps  | ----> |  Go API (chi)          |
| SDK or REST    |       |  OpenAPI generated     |
+----------------+       +-----------+------------+
                                     |
              +----------------------+----------------------+
              |                      |                      |
              v                      v                      v
    +----------------+     +----------------+     +----------------+
    | PostgreSQL     |     | Redis + Asynq  |     | Object Storage |
    | source of truth|     | background jobs|     | exports later  |
    +----------------+     +--------+-------+     +----------------+
                                     |
                                     v
                          +----------+-----------+
                          |  Worker processes    |
                          |  provider adapters   |
                          +----------+-----------+
                                     |
        +----------------------------+----------------------------+
        |              |             |             |              |
        v              v             v             v              v
     Resend        SendGrid         SES          Twilio          FCM

Provider webhooks -> API -> event parser -> delivery events -> customer webhooks
```

---

## 6. Core Runtime Flow

### Send notification

```txt
1.  Customer sends POST /v1/notifications
2.  API authenticates the API key
3.  API resolves organization, project, environment
4.  API validates request against OpenAPI schema
5.  API checks idempotency key
6.  API validates template and required variables
7.  API creates notification and delivery records
8.  API enqueues delivery job
9.  API returns notification_id and delivery_id
10. Worker renders template
11. Worker checks preferences and suppression lists
12. Worker chooses provider
13. Worker sends through provider adapter
14. Worker records provider message ID and status
15. Provider sends webhook event
16. API verifies webhook signature
17. API normalizes provider event
18. API updates delivery event timeline
19. API triggers customer webhook if configured
```

### Idempotency

Every public send endpoint must support an `Idempotency-Key` header.

Deduplication key: `organization_id + project_id + endpoint + idempotency_key`

Return `409 Conflict` if the same key is reused with a different request body hash.

---

## 7. Repository Structure

```txt
.
├── api/
│   ├── openapi.yaml          # source of truth — edit this first
│   ├── oapi-codegen.yaml     # code generation config
│   └── examples/
├── apps/
│   └── web/                  # Next.js dashboard
├── cmd/
│   ├── api/main.go
│   ├── worker/main.go
│   └── migrate/main.go
├── internal/
│   ├── auth/
│   ├── billing/
│   ├── config/
│   ├── db/
│   │   ├── migrations/       # goose migration files
│   │   ├── queries/          # sqlc input SQL
│   │   └── sqlc/             # generated — do not edit
│   ├── deliveries/
│   ├── domain/
│   ├── events/
│   ├── gen/openapi/          # generated — do not edit
│   ├── http/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   └── server/
│   ├── notifications/
│   ├── organizations/
│   ├── preferences/
│   ├── providers/
│   │   ├── email/{resend,sendgrid,ses}
│   │   ├── push/fcm
│   │   └── sms/twilio
│   ├── queue/
│   ├── templates/
│   ├── usage/
│   ├── webhooks/
│   └── workers/
├── pkg/sdk/go/
├── sdks/{typescript,python,go}
├── scripts/
└── docs/
    └── runbooks/
```

---

## 8. AI Coding Agent Instructions

### General rules

1. Read `README.md`, `api/openapi.yaml`, and relevant docs before editing code.
2. Never implement public API behavior without updating `api/openapi.yaml` first.
3. Never edit generated files in `internal/gen/` or `internal/db/sqlc/`.
4. Run `make generate` after changing OpenAPI or SQL queries.
5. Run tests before claiming a task is complete.
6. Keep handlers thin. Business logic belongs in services.
7. Keep provider-specific code inside `internal/providers`.
8. Do not hardcode provider credentials.
9. Do not put secrets in tests, fixtures, logs, or snapshots.
10. Every public endpoint must have authentication, tenant isolation, request validation, and tests.
11. Every send operation must be idempotent.
12. Every provider webhook must verify signatures where the provider supports it.
13. Every database migration must be reversible or have a documented irreversible reason.
14. Prefer small, reviewable changes over large rewrites.

### Workflow for any feature

```txt
1.  Restate the feature in one paragraph.
2.  Identify affected files.
3.  Update api/openapi.yaml if public API changes.
4.  Add or update migrations.
5.  Add or update sqlc queries.
6.  Run make generate.
7.  Implement domain types.
8.  Implement service logic.
9.  Implement HTTP handler.
10. Implement worker logic if needed.
11. Add unit tests.
12. Add integration tests where useful.
13. Update docs.
14. Run make lint test.
15. Summarize changes and remaining risks.
```

---

## 9. Package Management Rules

Never add packages or modules by editing manifest files manually. Always use the canonical tool for the language.

### Go

```bash
go get github.com/some/package           # add or upgrade
go get github.com/some/package@v1.2.3    # pin to version
go mod tidy                              # remove unused
```

Never edit `go.mod` or `go.sum` by hand.

### Node / Frontend

```bash
pnpm add <package>       # runtime dependency
pnpm add -D <package>    # dev dependency
pnpm remove <package>    # remove
```

Never edit `package.json` dependency fields by hand. Never use `npm` or `yarn`.

---

## 10. Developer Setup

### Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| Go | 1.22+ | Backend |
| Node.js | 20+ | Frontend runtime |
| pnpm | 9+ | Frontend package manager |
| Docker Desktop | Latest | PostgreSQL + Redis |
| Git | 2.x | Version control |
| pre-commit | Latest | Git hook runner |
| golangci-lint | 2.x | Go linter |
| hadolint | Latest | Dockerfile linter |
| shellcheck | Latest | Shell script linter |
| gitleaks | Latest | Secrets detection |
| sqlfluff | Latest | SQL linter |
| govulncheck | Latest | Go CVE scanner |
| spectral | Latest | OpenAPI linter |

---

### macOS

**1. Install Homebrew** (if not installed):

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

**2. Install core tools:**

```bash
brew install go node git
```

**3. Install pnpm:**

```bash
npm install -g pnpm
```

**4. Install Docker Desktop:**

Download from https://www.docker.com/products/docker-desktop and start it.

**5. Install code quality tools:**

```bash
brew install pre-commit golangci-lint hadolint shellcheck gitleaks sqlfluff
go install golang.org/x/vuln/cmd/govulncheck@latest
npm install -g @stoplight/spectral-cli --prefix ~/.npm-global
```

Add to `~/.zshrc`:

```bash
export PATH="$PATH:$HOME/go/bin:$HOME/.npm-global/bin"
```

**6. Clone and set up:**

```bash
git clone https://github.com/notifylayer/notifylayer.git
cd notifylayer
make setup
cp .env.example .env
make db-up
make migrate
```

**7. Start services** (separate terminals):

```bash
make api       # :8080
make worker
make web       # :3000
```

---

### Linux (Ubuntu / Debian)

**1. Install core tools:**

```bash
sudo apt update && sudo apt install -y git curl wget build-essential

# Go
wget https://go.dev/dl/go1.22.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.linux-amd64.tar.gz
echo 'export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"' >> ~/.bashrc
source ~/.bashrc

# Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# pnpm
npm install -g pnpm
```

**2. Install Docker:**

```bash
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER && newgrp docker
sudo apt install -y docker-compose-plugin
```

**3. Install code quality tools:**

```bash
# pre-commit
pip3 install pre-commit

# golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# shellcheck, hadolint, gitleaks, sqlfluff
sudo apt install -y shellcheck
wget -O /usr/local/bin/hadolint https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Linux-x86_64
chmod +x /usr/local/bin/hadolint
GITLEAKS_VER=$(curl -s https://api.github.com/repos/gitleaks/gitleaks/releases/latest | grep tag_name | cut -d '"' -f4 | tr -d v)
wget -O /tmp/gitleaks.tar.gz https://github.com/gitleaks/gitleaks/releases/latest/download/gitleaks_${GITLEAKS_VER}_linux_x64.tar.gz
sudo tar -C /usr/local/bin -xzf /tmp/gitleaks.tar.gz gitleaks
pip3 install sqlfluff

# govulncheck, spectral
go install golang.org/x/vuln/cmd/govulncheck@latest
npm install -g @stoplight/spectral-cli
```

**4. Clone and set up:**

```bash
git clone https://github.com/notifylayer/notifylayer.git
cd notifylayer
make setup
cp .env.example .env
make db-up
make migrate
```

---

### Windows (WSL 2)

Native Windows is not supported. Use WSL 2.

**1. Enable WSL 2** (PowerShell as Administrator):

```powershell
wsl --install
```

Restart your machine.

**2. Install Docker Desktop** from https://www.docker.com/products/docker-desktop. Enable the WSL 2 backend. In **Settings → Resources → WSL Integration**, enable your Ubuntu distro.

**3.** Open your Ubuntu WSL terminal and follow the **Linux** instructions above.

**Recommended editor:** VS Code with the Remote - WSL extension. Open the project from inside WSL with `code .`.

---

### Verifying setup

```bash
pre-commit run --all-files
```

All hooks should pass or be skipped. The only expected failure is `no-commit-to-branch` when on the `main` branch — that is correct behavior.

---

## 11. Pre-Commit Hooks

Hooks run automatically on every `git commit` across all languages in this project.

| Hook | Tool | Covers |
|---|---|---|
| Hygiene | pre-commit-hooks | Whitespace, EOF, YAML/JSON/TOML syntax, merge conflicts, large files |
| Branch protection | pre-commit-hooks | Blocks direct commits to `main` |
| Secrets | gitleaks | Hardcoded keys and tokens across all file types |
| Shell | shellcheck | `scripts/*.sh` |
| Dockerfiles | hadolint | Best practices and security |
| Go format | go fmt | Canonical formatting |
| Go vet | go vet | Correctness issues |
| Go lint | golangci-lint | Style, complexity, security (see `.golangci.yml`) |
| Go CVEs | govulncheck | Known vulnerabilities in dependencies |
| SQL | sqlfluff | `internal/db/migrations/*.sql` style and formatting |
| OpenAPI | spectral | OAS3 compliance for `api/openapi.yaml` |
| Frontend lint | ESLint via pnpm | `apps/web/` TS/TSX/JSX |
| Frontend format | Prettier via pnpm | `apps/web/` TS/CSS/JSON |
| CI workflows | check-jsonschema | GitHub Actions YAML |

**Config files:**

```txt
.pre-commit-config.yaml   hook definitions
.golangci.yml             Go linter rules
.sqlfluff                 SQL style rules
.spectral.yaml            OpenAPI lint rules
.hadolint.yaml            Dockerfile lint rules
.gitleaks.toml            Secrets allowlist
```

**Useful commands:**

```bash
pre-commit run --all-files     # run everything manually
pre-commit run <hook-id>       # run one hook
pre-commit autoupdate          # bump hook versions
```

Do not use `git commit --no-verify`. Fix the underlying issue instead.

---

## 12. Makefile

```bash
make setup             # install Go deps, frontend packages, pre-commit hooks
make hooks             # install pre-commit hooks only
make dev               # start DB, run migrations, generate code
make api               # run Go API server
make worker            # run background worker
make web               # run Next.js dev server
make db-up             # start PostgreSQL and Redis
make db-down           # stop containers
make migrate           # run database migrations
make generate          # regenerate from OpenAPI spec and SQL queries
make lint              # run all linters
make fmt               # format all code
make test              # run unit tests
make test-integration  # run integration tests
make clean             # delete build artifacts
```

---

## 13. Environment Variables

Copy `.env.example` to `.env` and fill in values for local development.

```env
APP_ENV=development
APP_NAME=notification-platform
API_BASE_URL=http://localhost:8080
WEB_BASE_URL=http://localhost:3000

DATABASE_URL=postgres://postgres:postgres@localhost:5432/notify?sslmode=disable
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

CLERK_SECRET_KEY=
CLERK_PUBLISHABLE_KEY=
CLERK_JWT_ISSUER=
CLERK_JWT_AUDIENCE=
CLERK_WEBHOOK_SECRET=

STRIPE_SECRET_KEY=
STRIPE_WEBHOOK_SECRET=

RESEND_API_KEY=
SENDGRID_API_KEY=
AWS_REGION=
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
TWILIO_ACCOUNT_SID=
TWILIO_AUTH_TOKEN=
FIREBASE_SERVICE_ACCOUNT_JSON=

ENCRYPTION_MASTER_KEY=
WEBHOOK_SIGNING_SECRET=
```

Secrets must be encrypted at rest when stored in the database. Use envelope encryption or a managed KMS in production.

---

## 14. Domain Model

Use ULID primary keys (`github.com/oklog/ulid/v2`). ULIDs are sortable and match the prefixed ID format used in public APIs (`not_01HX...`).

### Core entities

```txt
users
organizations
organization_members
projects
environments
api_keys
provider_connections
sender_identities
contacts
contact_channels
contact_preferences
suppression_entries
templates
template_versions
notifications
deliveries
delivery_events
webhook_endpoints
webhook_deliveries
idempotency_keys
usage_events
billing_customers
audit_logs
```

### Key table schemas

#### organizations
```txt
id, name, slug, created_at, updated_at
```

#### projects
```txt
id, organization_id, name, slug, created_at, updated_at
```

#### environments
```txt
id, project_id, name (development/staging/production), created_at, updated_at
```

#### api_keys
```txt
id, organization_id, project_id, environment_id
name, key_prefix, key_hash, scopes
last_used_at, revoked_at, created_at
```
Never store raw API keys.

#### provider_connections
```txt
id, organization_id, project_id, environment_id
provider_type (resend/sendgrid/ses/twilio/fcm)
channel (email/sms/push/whatsapp)
display_name, encrypted_config, is_default, is_active
created_at, updated_at
```

#### templates / template_versions
```txt
templates:        id, organization_id, project_id, key, name, channel
template_versions: id, template_id, version_number, subject
                   html_body, text_body, push_title, push_body, sms_body
                   variables_schema, status (draft/active/archived)
                   created_by, created_at
```

#### notifications
```txt
id, organization_id, project_id, environment_id
external_id, template_key, channel, recipient_ref
status, metadata, created_at, updated_at
```

#### deliveries
```txt
id, notification_id, organization_id, project_id, environment_id
channel, provider_type, provider_connection_id, provider_message_id
status, attempt_count, last_error_code, last_error_message
scheduled_at, sent_at, delivered_at, failed_at, created_at, updated_at
```

#### delivery_events
```txt
id, delivery_id, event_type, provider_type
provider_event_id, payload, occurred_at, created_at
```

#### idempotency_keys
```txt
id, organization_id, project_id, environment_id
endpoint, key, request_hash
response_status, response_body, created_at, expires_at
```

---

## 15. Delivery State Machine

```txt
Statuses:
  queued, processing, sent, delivered
  failed_temporary, failed_permanent
  bounced, complained, suppressed, cancelled

Transitions:
  queued          -> processing
  processing      -> sent | failed_temporary | failed_permanent
  sent            -> delivered | bounced | complained
  failed_temporary -> queued | failed_permanent
  queued          -> cancelled | suppressed
```

Rules:
1. Do not allow invalid state transitions.
2. Store every provider event even if it does not change final state.
3. Make event processing idempotent using provider event ID.
4. Preserve raw provider payloads.
5. Redact secrets and sensitive headers before storing payloads.

---

## 16. Provider Adapter Interface

No service outside `internal/providers` should import provider SDKs directly.

```go
type EmailProvider interface {
    SendEmail(ctx context.Context, input SendEmailInput) (SendResult, error)
    VerifyWebhook(ctx context.Context, headers map[string]string, body []byte) error
    ParseWebhook(ctx context.Context, headers map[string]string, body []byte) ([]NormalizedProviderEvent, error)
}

type SendEmailInput struct {
    To, From, ReplyTo, Subject, HTML, Text string
    Headers, Metadata map[string]string
    ScheduledAt *time.Time
}

type SendResult struct {
    Status            SendStatus   // accepted | rejected | unknown
    ProviderType      ProviderType
    ProviderMessageID string
    Raw               json.RawMessage
}

type NormalizedProviderEvent struct {
    EventID, ProviderMessageID, EventType string
    ProviderType ProviderType
    OccurredAt   time.Time
    Raw          json.RawMessage
}
```

---

## 17. OpenAPI-First API Contract

`api/openapi.yaml` is the source of truth. Never use Swaggo — this project is spec-first, not code-first.

Run `make generate` after any change to the spec.

### Public endpoints

```txt
GET    /v1/health
GET    /v1/docs

POST   /v1/notifications
GET    /v1/notifications/{notification_id}
GET    /v1/deliveries/{delivery_id}
GET    /v1/deliveries/{delivery_id}/events

GET    /v1/templates
POST   /v1/templates
GET    /v1/templates/{template_key}
POST   /v1/templates/{template_key}/versions
POST   /v1/templates/{template_key}/render
POST   /v1/templates/{template_key}/activate

GET    /v1/provider-connections
POST   /v1/provider-connections
GET    /v1/provider-connections/{id}
PATCH  /v1/provider-connections/{id}
DELETE /v1/provider-connections/{id}

GET    /v1/contacts/{contact_id}
PUT    /v1/contacts/{contact_id}
GET    /v1/contacts/{contact_id}/preferences
PUT    /v1/contacts/{contact_id}/preferences

GET    /v1/webhook-endpoints
POST   /v1/webhook-endpoints
PATCH  /v1/webhook-endpoints/{id}
DELETE /v1/webhook-endpoints/{id}

POST   /v1/provider-webhooks/resend
POST   /v1/provider-webhooks/sendgrid
POST   /v1/provider-webhooks/ses
POST   /v1/provider-webhooks/twilio
POST   /v1/provider-webhooks/fcm
```

### Send notification

Request:
```json
{
  "channel": "email",
  "recipient": { "email": "user@example.com", "external_id": "user_123" },
  "template": { "key": "welcome_email", "data": { "first_name": "James" } },
  "routing": { "provider": "resend" },
  "metadata": { "source": "signup_flow" }
}
```

Response:
```json
{
  "notification_id": "not_01HX...",
  "delivery_id": "del_01HX...",
  "status": "queued"
}
```

---

## 18. Feature Roadmap

### P0 — Email MVP

Goal: connect a provider, create a template, send an email, see delivery logs.

```txt
OpenAPI spec + generated Go types
Go API with chi
PostgreSQL schema and migrations
sqlc query layer
API key authentication
Organizations, projects, environments
Provider connections (Resend, SendGrid)
Template CRUD and versioning
Template rendering with {{variable}} support
POST /v1/notifications with idempotency
Redis + Asynq queue and worker
Delivery records and event timeline
Provider webhook ingestion (Resend, SendGrid)
Dashboard: API keys, providers, templates, logs
Swagger UI at /docs
Docker Compose environment
Unit and integration tests
Structured logs
Rate limiting per API key
Sandbox mode
```

### P1 — Operational reliability

```txt
Amazon SES adapter
Customer webhook endpoints with retries and replay
Suppression list and contact preferences
Template preview and test send
Usage metering and Stripe billing
Audit logs and provider health checks
Provider fallback routing
TypeScript SDK
```

### P2 — Multi-channel

```txt
FCM push adapter
Twilio SMS and WhatsApp adapters
Channel fallback rules
Delayed and scheduled sends
Simple workflows (send, wait, condition, webhook steps)
In-app notification inbox API and React component
Python and Go SDKs
Template environment promotion and approval flow
Team roles and permissions
```

### P3 — Strategic moat

```txt
AI template QA and failure explanations
Notification fatigue detection
Anomaly detection for bounce and complaint spikes
Advanced provider scoring and routing
Data exports and Terraform provider
SOC 2 preparation
Enterprise SSO
Regional data controls
```

---

## 19. Development Phases

Each phase ships as a small sequence of pull requests. Do not skip tests.

### Phase 0: Repository bootstrap
Initialize Go module, Next.js app, Docker Compose, Makefile, `.env.example`, `AGENTS.md`, `CLAUDE.md`, CI.
**Done when:** `make setup && make db-up && make test` all pass.

### Phase 1: OpenAPI and API skeleton
Define OpenAPI base, health endpoint, auth scheme, notification schemas. Configure oapi-codegen, create chi server, serve Swagger UI, add request ID and logging middleware.
**Done when:** `GET /v1/health` returns 200, `GET /docs` loads Swagger UI, `make generate` works.

### Phase 2: Database foundation
Migrations for all core entities, sqlc config and queries, repository layer, test helper, seed script.
**Done when:** `make migrate && sqlc generate && go test ./internal/db/...` pass.

### Phase 3: Authentication and tenant isolation
API key generation (hash only, never store raw), scopes, middleware to resolve org/project/environment, Clerk webhook handler for `user.created` to bootstrap org and project records, cross-tenant access tests.
**Done when:** authenticated requests resolve tenant context; unauthenticated and cross-tenant requests fail.

### Phase 4: Templates
Template CRUD, version management, active version selection, minimal renderer, variable schema validation, render preview endpoint, dashboard template editor.
**Done when:** create template, activate version, render with test data, see missing variable errors.

### Phase 5: Email provider adapters
Provider connection CRUD with encrypted credentials, Resend and SendGrid adapters, mocked HTTP tests, error normalization, dashboard provider setup.
**Done when:** test email sends through Resend or SendGrid via the same internal interface.

### Phase 6: Notification send pipeline
Idempotency service, `POST /v1/notifications`, delivery records, Asynq client and worker, delivery state machine, retry policy, delivery event timeline, dashboard logs.
**Done when:** `POST /v1/notifications` returns queued; worker sends email; dashboard shows full delivery status.

### Phase 7: Provider webhooks
Webhook endpoints for Resend and SendGrid, signature verification, event normalization, deduplication, delivery state updates, raw payload storage with redaction.
**Done when:** provider webhooks create normalized delivery events and update status idempotently.

### Phase 8: Dashboard MVP
Clerk auth in Next.js, dashboard shell with org/project switcher, pages for API keys, providers, templates, test send, delivery logs, delivery detail, settings.
**Done when:** beta user can complete the full setup flow from the dashboard.

### Phase 9: Webhooks and operational controls
Customer webhook endpoint CRUD, signed outbound webhooks, retry queue, replay, delivery logs, rate limiting, suppression list, audit logs.
**Done when:** customers can receive, inspect, and replay delivery webhooks.

### Phase 10: Billing and beta readiness
Stripe customer mapping, subscription plans, usage events, monthly usage dashboard, free tier limits, plan enforcement, acceptable use policy.
**Done when:** customer can sign up, connect a provider, send real notifications, view usage, and be billed.

### Phase 11: Multi-channel expansion
FCM push, Twilio SMS, channel schemas, delivery workers, channel preferences, fallback routing.
**Done when:** same send API delivers email, push, and SMS through separate adapters.

### Phase 12: Simple workflows
Workflow tables, trigger endpoint, step schema, execution engine (send, wait, condition, webhook steps), workflow dashboard.
**Done when:** customer can define an event-triggered workflow with send and wait steps.

---

## 20. API Style Guide

### ID prefixes

```txt
org_   prj_   env_   key_
tpl_   not_   del_   evt_
wh_
```

### Error shape

```json
{
  "error": {
    "code": "template_missing_variable",
    "message": "Template is missing required variable: first_name",
    "request_id": "req_01HX...",
    "docs_url": "https://docs.notifylayer.com/errors/template_missing_variable"
  }
}
```

### Error codes

```txt
auth_invalid_api_key       auth_missing_scope
tenant_not_found           template_not_found
template_missing_variable  provider_not_configured
provider_rejected_request  provider_temporary_failure
rate_limit_exceeded        idempotency_conflict
validation_failed
```

### Pagination

Cursor-based:

```txt
GET /v1/deliveries?limit=50&cursor=...
```

```json
{ "data": [], "next_cursor": "..." }
```

---

## 21. Security Requirements

1. Hash API keys with a strong one-way hash. Never store raw keys.
2. Encrypt provider secrets at rest using envelope encryption or KMS.
3. Redact secrets from logs.
4. Verify provider webhook signatures.
5. Sign outbound customer webhooks.
6. Enforce tenant isolation in every database query.
7. Rate limit per API key and project.
8. Store audit logs for sensitive actions.
9. Support API key revocation.
10. Separate development, staging, and production environments.
11. Sandbox mode for tests.
12. Suppression list support.
13. Unsubscribe and preference controls before lifecycle messaging.
14. Acceptable use policy before public launch.

**Abuse signals to block or flag:**
```txt
cold outreach, scraped recipient lists, high bounce/complaint rates
rapid account creation, suspicious send spikes, unverified domains
```

---

## 22. Observability

Every request log must include:
```txt
request_id, organization_id, project_id, environment_id, api_key_prefix
```

Every delivery log must include:
```txt
notification_id, delivery_id, channel, provider, provider_message_id
status, attempt_count, latency_ms, error_code
```

Use OpenTelemetry for traces, metrics, and logs from day one.

### Required metrics

```txt
notifications_created_total       deliveries_sent_total
deliveries_failed_total           deliveries_bounced_total
deliveries_complained_total       provider_latency_ms
worker_job_duration_ms            worker_job_failures_total
webhook_deliveries_total          webhook_delivery_failures_total
api_rate_limited_total
```

---

## 23. Testing Strategy

**Unit tests:** template renderer, variable validation, state machine transitions, provider event normalization, idempotency logic, rate-limit logic, provider adapter request construction, error mapping.

**Integration tests:** database repositories, API key auth, tenant isolation, notification send endpoint, queue and worker processing, provider webhook ingestion, customer webhook delivery.

**End-to-end (Playwright):** login, provider setup, template creation, test send, delivery log inspection, webhook replay.

Do not call real providers in default tests. Use mocked HTTP servers. Real provider tests must be opt-in and gated by environment variables.

---

## 24. Dashboard Pages

Build in this order:

```txt
/login
/app
/app/projects
/app/api-keys
/app/providers
/app/templates
/app/templates/:template_id
/app/send-test
/app/deliveries
/app/deliveries/:delivery_id
/app/webhooks
/app/usage
/app/settings
```

Principles:
1. Logs are the core product UI.
2. Every delivery must have a timeline.
3. Every failed delivery must show the reason and next step.
4. Provider setup must have a test connection button.
5. Templates must have preview and test send.
6. Webhooks must have replay.

---

## 25. Final Instruction For Agents

Start with Phase 0. Do not build workflows, AI features, SMS, WhatsApp, or in-app inbox until the email send path is reliable.

The first real milestone:

```txt
A customer can connect Resend or SendGrid, create a template,
send a transactional email through the API, see a full delivery
timeline, and debug failures from the dashboard.
```

That is the MVP.
