# NotifyLayer

A developer-first notification orchestration platform. One API for email — with templates, routing, retries, and provider failover.

---

## What It Does

- Send email notifications via a single API call
- Bring your own provider — Resend and SendGrid supported
- Template management with versioning and variable rendering
- Background delivery with automatic retries (Asynq + Redis)
- Inbound provider webhooks update the delivery timeline in real time
- Dashboard to manage providers, templates, send tests, and view delivery logs

---

## Stack

| Layer | Technology |
|---|---|
| API | Go 1.25+, chi, OpenAPI 3.0 |
| Database | PostgreSQL 16, goose, sqlc + pgx/v5 |
| Queue | Redis + Asynq |
| Dashboard | Next.js 16, TypeScript, Tailwind CSS |
| Auth | Clerk (dashboard), API keys (customer API) |

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

### Running the services

```bash
make api         # API server on :8080
make worker      # background worker
make web         # Next.js dashboard on :3000
```

### Dashboard setup

Copy the environment file and fill in your [Clerk](https://clerk.com) keys:

```bash
cp apps/web/.env.local.example apps/web/.env.local
```

---

## API

Base URL: `http://localhost:8080`

Authenticate all requests with an API key:

```
Authorization: Bearer <api_key>
```

Interactive docs available at [http://localhost:8080/docs](http://localhost:8080/docs).

### Key endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/notifications` | Send a notification |
| `GET` | `/v1/notifications/:id` | Get a notification |
| `GET` | `/v1/deliveries/:id` | Get a delivery |
| `GET` | `/v1/deliveries/:id/events` | Get the delivery event timeline |
| `POST` | `/v1/webhooks/email/:provider` | Inbound provider webhook (no auth) |
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
  db/                  # Migrations, queries, sqlc output, repos
  gen/                 # Generated code (do not edit)
  providers/           # Resend and SendGrid adapters
  queue/               # Asynq client and task definitions
  template/            # Template renderer
  worker/              # Email delivery worker
```
