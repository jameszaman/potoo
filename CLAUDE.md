# NotifyLayer — Claude Code Instructions

---

## Project

Bring-your-own-provider notification orchestration platform.

- **Backend:** Go 1.22+ with chi, OpenAPI-first (`api/openapi.yaml` is source of truth)
- **Frontend:** Next.js + TypeScript (`apps/web/`)
- **Database:** PostgreSQL 16, goose migrations, sqlc + pgx/v5
- **Queue:** Redis + Asynq
- **Auth:** Clerk (dashboard), API keys (customer API)

---

## Session guidance

Keep sessions focused on one phase or one feature. Open only the files needed for the current task. Long sessions degrade when context fills up — use this file to re-orient at the start of each session.

---

## Commands

```bash
make setup             # install deps and hooks — run once after cloning
make dev               # start DB, migrate, generate
make api               # Go API server :8080
make worker            # background worker
make web               # Next.js :3000
make db-up / db-down   # start / stop containers
make migrate           # run migrations
make generate          # regenerate from OpenAPI + SQL
make lint              # all linters
make fmt               # format all code
make test              # unit tests
make test-integration  # integration tests
```

---

## Hard rules

1. Read `README.md` and `api/openapi.yaml` before editing code.
2. Never implement public API behavior without updating `api/openapi.yaml` first.
3. Never edit `internal/gen/` — it is generated. (sqlc output will live in `internal/db/sqlc/` once Phase 2 begins.)
4. Run `make generate` after changing the OpenAPI spec or SQL queries.
5. Run `make lint test` before reporting a task complete.
6. Handlers must be thin — business logic goes in services.
7. All provider-specific code goes inside `internal/providers/`.
8. No hardcoded credentials. No secrets in tests, fixtures, or logs.
9. Every public endpoint needs: auth, tenant isolation, validation, tests.
10. Every send is idempotent. Every provider webhook verifies its signature.
11. Every migration is reversible, or documents why it is not.

---

## Package management

- **Go:** `go get <pkg>` to add, `go mod tidy` to clean. Never edit `go.mod` by hand.
- **Frontend:** `pnpm add <pkg>` to add, `pnpm remove <pkg>` to remove. Never use `npm` or `yarn`.

---

## Key file locations

```
api/openapi.yaml              API contract — edit this first
api/oapi-codegen.yaml         Code generation config
internal/gen/openapi/         Generated Go types and server interface (do not edit)
internal/http/handlers/       HTTP handlers (thin — business logic goes in services)
internal/http/middleware/     Request ID, logging, recovery middleware
internal/http/server/         chi router setup
apps/web/                     Next.js dashboard
cmd/api/main.go               API server entry point
cmd/worker/main.go            Worker entry point
cmd/migrate/main.go           Migration runner entry point
```

Packages such as `internal/db/`, `internal/providers/`, `internal/notifications/` etc. are created when the corresponding phase begins — not before.

---

## Architecture in one paragraph

Customer apps call `POST /v1/notifications`. The Go API authenticates the API key, resolves the tenant, validates the request, checks idempotency, creates notification and delivery records, and enqueues a job in Redis via Asynq. The worker picks up the job, renders the template, checks preferences and suppression lists, selects a provider, and sends via the provider adapter. The provider fires a delivery webhook back to the API, which verifies the signature, normalizes the event, updates the delivery timeline, and fires any configured customer webhooks.

---

## Current phase

**Phase 1 complete.** `GET /v1/health` returns 200 and `GET /docs` loads Swagger UI.

**Next: Phase 2 — Database foundation.**

Tasks:
1. Add goose migrations for all core entities.
2. Configure sqlc and write initial queries.
3. Implement a repository layer.
4. Add a test helper and seed script.

Done when `make migrate && sqlc generate && go test ./internal/db/...` pass.
