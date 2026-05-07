# Potoo — Claude Code Instructions

---

## Project

Bring-your-own-provider notification orchestration platform.

- **Backend:** Go 1.22+ with chi, OpenAPI-first (`docs/spec/openapi.yaml` is source of truth)
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

1. Read `README.md` and `docs/spec/openapi.yaml` before editing code.
2. Never implement public API behavior without updating `docs/spec/openapi.yaml` first — add the path to the correct domain file under `docs/spec/paths/`.
3. After updating the spec: run `make generate` immediately. The spec is embedded in the binary at compile time — Swagger UI at `/docs` will not reflect changes until this is done and the server is rebuilt.
4. After updating the spec: register the new route in `internal/api/server/server.go` under the correct auth tier group.
5. After updating the spec: update the endpoint table in `README.md` (Method, Path, Tier, Description columns) to match.
5. Never edit `internal/gen/openapi/` or `internal/db/sqlc/` — they are generated. Run `make generate` after changing `docs/spec/` or `internal/db/queries/`.
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
docs/spec/openapi.yaml        API contract — edit this first
docs/spec/schemas/            Domain schema files (auth, orgs, providers, templates, …)
docs/spec/paths/              Domain path files (auth, platform, providers, templates, …)
docs/spec/oapi-codegen.yaml   Code generation config
docs/architecture/            Architecture documentation
sqlc.yaml                     sqlc config
internal/gen/openapi/         Generated Go types and server interface (do not edit)
internal/api/handlers/        HTTP handlers (thin — business logic goes in services)
internal/api/middleware/      Request ID, logging, recovery middleware
internal/api/server/          chi router setup
internal/db/migrations/       goose migration files
internal/db/queries/          sqlc input SQL
internal/db/sqlc/             Generated database code (do not edit)
internal/db/repo/             Repository layer wrapping sqlc queries
internal/db/db.go             pgxpool connection helper
apps/web/                     Next.js dashboard
cmd/api/main.go               API server entry point
cmd/worker/main.go            Worker entry point
cmd/migrate/main.go           Migration runner
```

Packages such as `internal/providers/`, `internal/notifications/` etc. are created when the corresponding phase begins — not before.

---

## Architecture in one paragraph

Customer apps call `POST /v1/notifications`. The Go API authenticates the API key, resolves the tenant, validates the request, checks idempotency, creates notification and delivery records, and enqueues a job in Redis via Asynq. The worker picks up the job, renders the template, checks preferences and suppression lists, selects a provider, and sends via the provider adapter. The provider fires a delivery webhook back to the API, which verifies the signature, normalizes the event, updates the delivery timeline, and fires any configured customer webhooks.

---

## Current phase

**Phase 1 complete.** `GET /v1/health` returns 200 and `GET /docs` loads Swagger UI.

**Phases 3 and 4 complete.** Auth middleware hashes the Bearer token with SHA-256, looks it up in `api_keys`, and stores the resolved tenant (org/project/env) in request context. `/v1/health` is public; all other routes require a valid key.

**Phase 5 complete.** `EmailProvider` interface, Resend + SendGrid adapters, `provider_connections` table, CRUD endpoints.

**Next: Phase 6 — Notification send pipeline.**

Tasks:
1. Wire `POST /v1/notifications` — create notification + delivery records, enqueue job.
2. Add Redis + Asynq worker.
3. Worker: resolve provider, render template, send via adapter, update delivery status.
4. Delivery state machine transitions.

Done when `POST /v1/notifications` returns queued and the worker sends a real email.
