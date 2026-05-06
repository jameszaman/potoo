# NotifyLayer

A developer-first notification orchestration platform. One API for email — with templates, routing, retries, and provider failover.

---

## What It Does

- Send email notifications via a single API call
- Bring your own provider — Resend and SendGrid supported
- Template management with versioning and variable rendering
- Background delivery with automatic retries (Asynq + Redis)
- Inbound provider webhooks update the delivery timeline in real time
- Multi-tenant: platform org controls customer organizations
- Dashboard to manage providers, templates, send tests, and view delivery logs

---

## Stack

| Layer | Technology |
|---|---|
| API | Go 1.25+, chi, OpenAPI 3.0 |
| Database | PostgreSQL 16, goose, sqlc + pgx/v5 |
| Queue | Redis + Asynq |
| Dashboard | Next.js 16, TypeScript, Tailwind CSS |
| Auth | JWT + HttpOnly cookies (dashboard), API keys (customer API) |

---

## Running Locally

### Prerequisites

- Go 1.25+
- Node.js 20+ and pnpm
- Docker Desktop (for Postgres + Redis)
- [golangci-lint](https://golangci-lint.run/welcome/install/)
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen): `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html): `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- pre-commit: `pip install pre-commit`

### First-time setup

```bash
git clone https://github.com/notifylayer/notifylayer.git
cd notifylayer
make setup       # install deps and git hooks
make db-up       # start Postgres + Redis
make migrate     # run database migrations
```

### Environment variables

Copy `.env` and fill in values. All variables are required:

```
DATABASE_URL=postgres://user@localhost:5432/notify?sslmode=disable
REDIS_ADDR=localhost:6379
ALLOWED_ORIGIN=http://localhost:3000
JWT_SECRET=change-before-deploy
```

### Running the services

```bash
make api         # API server on :8080
make worker      # background worker
make web         # Next.js dashboard on :3000
```

### First run

Navigate to `http://localhost:3000/setup` to create the platform organization and your owner account. This route is only available once — it returns 410 Gone after setup is complete.

---

## API

Base URL: `http://localhost:8080`

Customer API requests authenticate with an API key:

```
Authorization: Bearer <api_key>
```

Interactive docs at [http://localhost:8080/docs](http://localhost:8080/docs).

### Key endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/notifications` | Send a notification |
| `GET` | `/v1/notifications/:id` | Get a notification |
| `GET` | `/v1/deliveries/:id` | Get a delivery |
| `GET` | `/v1/deliveries/:id/events` | Get the delivery event timeline |
| `POST` | `/v1/webhooks/email/:provider` | Inbound provider webhook |
| `GET/POST` | `/v1/provider-connections` | Manage provider connections |
| `GET/POST` | `/v1/templates` | Manage templates |

### Send a notification

```bash
curl -X POST http://localhost:8080/v1/notifications \
  -H "Authorization: Bearer <api_key>" \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "email",
    "recipient": { "email": "user@example.com" },
    "template": { "key": "welcome_email", "data": { "first_name": "James" } }
  }'
```

---

## Makefile

```bash
make setup             # install deps and hooks
make api               # run API server on :8080
make worker            # run background worker
make web               # run Next.js dev server on :3000
make db-up             # start Postgres + Redis
make db-down           # stop containers
make migrate           # run migrations
make generate          # regenerate from OpenAPI spec and SQL queries
make lint              # run all linters
make fmt               # format all code
make test              # run unit tests
```

---

## Project Structure

```
api/                   # OpenAPI spec and codegen config
apps/web/              # Next.js dashboard
cmd/
  api/                 # API server entry point
  worker/              # Background worker entry point
  migrate/             # Migration runner
internal/
  api/                 # Handlers, middleware, router
  auth/                # API key and session middleware, context helpers
  db/                  # Migrations, queries, sqlc output, repos
  gen/                 # Generated code (do not edit)
  jwtutil/             # JWT signing and verification
  providers/           # Resend and SendGrid adapters
  queue/               # Asynq client and task definitions
  template/            # Template renderer
  worker/              # Email delivery worker
```

---

## Planned

- **Projects and environments** — customer orgs will be able to create multiple projects (e.g. one per app) each with `development`, `staging`, and `production` environments. API keys will be scoped per environment, enabling rate limiting and separate provider configs per env. The database schema (`projects`, `environments` tables) is already in place; the management UI and API endpoints are not yet built.
- SMS and push notification channels
- Suppression lists and recipient preferences
- Customer-configured outbound webhooks
