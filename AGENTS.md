# NotifyLayer — Agent Instructions

This file is loaded automatically by OpenAI Codex and compatible agents. For Claude Code, see `CLAUDE.md`.

---

## Repository overview

Bring-your-own-provider notification orchestration platform.

- **Backend:** Go 1.22+ with chi router, OpenAPI-first
- **Frontend:** Next.js + TypeScript in `apps/web/`
- **Database:** PostgreSQL 16 via goose migrations and sqlc
- **Queue:** Redis + Asynq
- **Auth:** Clerk (dashboard), API keys (customer API)

---

## Commands

```bash
make setup             # install all dependencies and pre-commit hooks
make dev               # start DB, run migrations, generate code
make api               # run Go API server (:8080)
make worker            # run background worker
make web               # run Next.js dev server (:3000)
make db-up             # start PostgreSQL and Redis
make migrate           # run database migrations
make generate          # regenerate from OpenAPI spec and SQL queries
make lint              # run all linters
make fmt               # format all code
make test              # run unit tests
make test-integration  # run integration tests
make clean             # delete build artifacts
```

---

## Rules

1. Read `README.md` and `api/openapi.yaml` before editing any code.
2. Never implement public API behavior without updating `api/openapi.yaml` first.
3. Never edit generated files in `internal/gen/` or `internal/db/sqlc/`.
4. Run `make generate` after changing the OpenAPI spec or SQL queries.
5. Run `make lint test` before marking any task complete.
6. Keep handlers thin — business logic belongs in services.
7. Keep provider-specific code inside `internal/providers/`.
8. Never hardcode credentials or secrets. Use environment variables.
9. Never put secrets in tests, fixtures, logs, or snapshots.
10. Every public endpoint must have: authentication, tenant isolation, request validation, and tests.
11. Every send operation must be idempotent.
12. Every provider webhook must verify the provider's signature.
13. Every database migration must be reversible or document why it is not.
14. Prefer small reviewable changes over large rewrites.

---

## Package management

- **Go:** use `go get` and `go mod tidy`. Never edit `go.mod` or `go.sum` by hand.
- **Frontend:** use `pnpm add` / `pnpm remove`. Never edit `package.json` deps by hand. Never use `npm` or `yarn`.

---

## Workflow for any feature

```
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

## Current phase

**Phase 1 — OpenAPI and API skeleton.**

See `README.md` section 19 for full phase details.
