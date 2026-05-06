# NotifyLayer

A developer-first notification orchestration platform. One API for email, SMS, push, and webhooks — with templates, routing, retries, preferences, and provider failover.

---

## What This Is

An orchestration layer that sits above delivery providers. It does not own SMTP or telecom routing. It makes SendGrid, Resend, SES, Twilio, and FCM easier to use together reliably.

---

## Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.25+, chi |
| API contract | OpenAPI 3.0 (`api/openapi.yaml` is source of truth) |
| Code generation | oapi-codegen v2 |
| Frontend | Next.js 16, TypeScript, Tailwind CSS |
| Database | PostgreSQL 16, goose migrations, sqlc + pgx/v5 |
| Queue | Redis + Asynq |
| Dashboard auth | Clerk |
| Docs UI | Swagger UI at `/docs` |

---

## Repository Structure

```
.
├── api/
│   ├── openapi.yaml          # source of truth — edit this first
│   └── oapi-codegen.yaml     # code generation config
├── apps/
│   └── web/                  # Next.js dashboard
├── cmd/
│   ├── api/main.go           # API server entry point
│   ├── worker/main.go        # background worker entry point
│   └── migrate/main.go       # migration runner entry point
├── docs/
│   └── architecture.md
├── internal/
│   ├── api/
│   │   ├── handlers/         # thin HTTP handlers (implement generated interface)
│   │   ├── middleware/       # request ID, logging, recovery
│   │   └── server/           # chi router setup
│   └── gen/
│       └── openapi/          # generated from api/openapi.yaml — do not edit
├── .github/workflows/        # CI (go.yml, web.yml)
├── docker-compose.yml        # PostgreSQL 16 + Redis 7
├── Makefile
└── go.mod
```

New packages (`internal/db/`, `internal/providers/`, etc.) are added as each phase is implemented — not before.

---

## Developer Setup

### Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| Go | 1.25+ | Backend |
| Node.js | 20+ | Frontend runtime |
| pnpm | 9+ | Frontend package manager |
| Docker Desktop | Latest | PostgreSQL + Redis |
| pre-commit | Latest | Git hook runner |
| golangci-lint | 2.x | Go linter |
| hadolint | Latest | Dockerfile linter |
| shellcheck | Latest | Shell script linter |
| gitleaks | Latest | Secrets detection |
| sqlfluff | Latest | SQL linter |
| govulncheck | Latest | Go CVE scanner |
| spectral | Latest | OpenAPI linter |

### macOS

```bash
# Core tools
brew install go node git pre-commit golangci-lint hadolint shellcheck gitleaks sqlfluff

# pnpm (standalone installer — avoids npm permission issues)
curl -fsSL https://get.pnpm.io/install.sh | sh -

# Go tools
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# spectral
npm install -g @stoplight/spectral-cli --prefix ~/.npm-global
```

Add to `~/.zshrc`:

```bash
export PATH="$PATH:$HOME/go/bin:$HOME/.npm-global/bin"
```

### Linux (Ubuntu / Debian)

```bash
# Core tools
sudo apt update && sudo apt install -y git curl wget build-essential shellcheck

# Go — download from https://go.dev/dl/
wget https://go.dev/dl/go1.25.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
echo 'export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"' >> ~/.bashrc && source ~/.bashrc

# Node.js
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash - && sudo apt install -y nodejs

# pnpm
curl -fsSL https://get.pnpm.io/install.sh | sh -

# pre-commit
pip3 install pre-commit

# golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# hadolint
wget -O /usr/local/bin/hadolint https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Linux-x86_64
chmod +x /usr/local/bin/hadolint

# gitleaks
GITLEAKS_VER=$(curl -s https://api.github.com/repos/gitleaks/gitleaks/releases/latest | grep tag_name | cut -d'"' -f4 | tr -d v)
wget -O /tmp/gitleaks.tar.gz https://github.com/gitleaks/gitleaks/releases/latest/download/gitleaks_${GITLEAKS_VER}_linux_x64.tar.gz
sudo tar -C /usr/local/bin -xzf /tmp/gitleaks.tar.gz gitleaks

# sqlfluff, govulncheck, spectral, oapi-codegen
pip3 install sqlfluff
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
npm install -g @stoplight/spectral-cli
```

### Windows

Use WSL 2. Enable it from PowerShell (as Administrator):

```powershell
wsl --install
```

Then follow the Linux instructions above inside your WSL terminal.

### First-time setup

```bash
git clone https://github.com/notifylayer/notifylayer.git
cd notifylayer
make setup
make db-up
```

Verify everything works:

```bash
pre-commit run --all-files
make test
```

---

## Makefile

```bash
make setup             # install Go deps, frontend packages, pre-commit hooks
make api               # run Go API server on :8080
make worker            # run background worker
make web               # run Next.js dev server on :3000
make db-up             # start PostgreSQL and Redis (Docker)
make db-down           # stop containers
make migrate           # run database migrations
make generate          # regenerate from OpenAPI spec
make lint              # run all linters
make fmt               # format all code
make test              # run unit tests
make test-integration  # run integration tests
```

---

## Pre-Commit Hooks

Hooks run automatically on every `git commit`.

| Hook | Tool | Covers |
|---|---|---|
| Hygiene | pre-commit-hooks | Whitespace, EOF, YAML/JSON syntax, merge conflicts, large files |
| Branch protection | pre-commit-hooks | Blocks direct commits to `main` |
| Secrets | gitleaks | Hardcoded keys and tokens |
| Shell | shellcheck | `scripts/*.sh` |
| Dockerfiles | hadolint | Best practices and security |
| Go format | go-fmt | Canonical formatting |
| Go vet | go vet | Correctness issues |
| Go lint | golangci-lint | Style, complexity, security |
| Go CVEs | govulncheck | Known vulnerabilities in dependencies |
| SQL | sqlfluff | `internal/db/migrations/*.sql` |
| OpenAPI | spectral | OAS3 compliance for `api/openapi.yaml` |
| Frontend lint | ESLint via pnpm | `apps/web/` TS/TSX/JSX |
| Frontend format | Prettier via pnpm | `apps/web/` TS/CSS/JSON |
| CI workflows | check-jsonschema | GitHub Actions YAML syntax |

Config files: `.pre-commit-config.yaml`, `.golangci.yml`, `.sqlfluff`, `.spectral.yaml`, `.hadolint.yaml`, `.gitleaks.toml`

```bash
pre-commit run --all-files    # run manually
pre-commit run <hook-id>      # run one hook
pre-commit autoupdate         # bump hook versions
```

Do not use `git commit --no-verify`. Fix the underlying issue instead.

---

## Package Management

Never edit manifest files by hand.

**Go:**
```bash
go get github.com/some/package     # add or upgrade
go mod tidy                        # remove unused
```

**Frontend:**
```bash
pnpm add <package>      # runtime dependency
pnpm add -D <package>   # dev dependency
pnpm remove <package>   # remove
```

Never use `npm` or `yarn`.

---

## Development Phases

| Phase | Goal | Done when |
|---|---|---|
| 0 | Repository bootstrap | `make setup && make test` pass |
| 1 | OpenAPI + API skeleton | `GET /v1/health` → 200, `/docs` loads Swagger UI |
| 2 | Database foundation | migrations, sqlc, repository layer, `make migrate` passes |
| 3 | Auth + tenant isolation | API key auth, org/project/environment resolved per request |
| 4 | Templates | CRUD, versioning, rendering, variable validation |
| 5 | Email providers | Resend + SendGrid adapters, webhook ingestion |
| 6 | Send pipeline | `POST /v1/notifications`, Asynq worker, delivery state machine |
| 7 | Provider webhooks | signature verification, event normalization, delivery updates |
| 8 | Dashboard MVP | Clerk auth, provider setup, templates, test send, delivery logs |

**Current phase: Phase 1 complete.**

MVP target: a customer can connect Resend or SendGrid, create a template, send an email via the API, see the full delivery timeline, and debug failures from the dashboard.

---

## Hard Rules

1. Read `api/openapi.yaml` before touching any API code.
2. Never implement public API behavior without updating `api/openapi.yaml` first.
3. Never edit `internal/gen/` — it is generated.
4. Run `make generate` after changing the OpenAPI spec.
5. Run `make lint test` before marking any task complete.
6. Handlers must be thin — business logic goes in services.
7. All provider-specific code goes in `internal/providers/`.
8. No hardcoded credentials. No secrets in tests, fixtures, or logs.
9. Every public endpoint needs: auth, tenant isolation, validation, tests.
10. Every send is idempotent. Every provider webhook verifies its signature.
11. Every migration is reversible, or documents why it is not.
