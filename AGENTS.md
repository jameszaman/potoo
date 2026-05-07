# Potoo — Agent Instructions

This file is loaded automatically by OpenAI Codex and compatible agents. For Claude Code, see `CLAUDE.md`.

---

## Repository overview

Bring-your-own-provider notification orchestration platform.

- **Backend:** Go 1.22+ with chi router, OpenAPI-first (`docs/spec/openapi.yaml` is source of truth)
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
make generate          # regenerate from OpenAPI spec and SQL queries — run after editing docs/spec/ or internal/db/queries/
make lint              # run all linters
make fmt               # format all code
make test              # run unit tests
make test-integration  # run integration tests
make clean             # delete build artifacts
```

---

## Rules

1. Read `README.md` and `docs/spec/openapi.yaml` before editing any code.
2. Never implement public API behavior without updating `docs/spec/openapi.yaml` first — add the path to the correct domain file under `docs/spec/paths/`.
3. After updating the spec: run `make generate` immediately. The spec is embedded in the binary at compile time — Swagger UI at `/docs` will not reflect changes until this is done and the server is rebuilt.
4. After updating the spec: register the new route in `internal/api/server/server.go` under the correct auth tier group.
5. After updating the spec: update the endpoint table in `README.md` (Method, Path, Tier, Description columns) to match.
5. Never edit generated files in `internal/gen/openapi/` or `internal/db/sqlc/` — run `make generate` after changing `docs/spec/` or `internal/db/queries/`.
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
3.  Update the relevant file under docs/spec/paths/ if public API changes; add the new path entry to docs/spec/openapi.yaml.
4.  Run make generate (bundles spec, regenerates Go types and DB queries).
5.  Register the new route in internal/api/server/server.go under the correct auth tier group.
6.  Update the endpoint table in README.md.
7.  Add or update migrations.
8.  Add or update sqlc queries in internal/db/queries/; run make generate again if queries changed.
9.  Implement domain types.
10. Implement service logic.
11. Implement HTTP handler.
12. Implement worker logic if needed.
13. Add unit tests.
14. Add integration tests where useful.
15. Run make lint test.
16. Summarize changes and remaining risks.
```

---

## Current phase

**Phase 1 — OpenAPI and API skeleton.**

See `README.md` section 19 for full phase details.
