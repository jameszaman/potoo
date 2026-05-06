# Architecture

## Overview

NotifyLayer is a notification orchestration layer. It does not own delivery infrastructure. It accepts send requests from customer applications, routes them through customer-supplied provider credentials, and returns normalized delivery events.

---

## System components

```
                          +----------------------+
                          |      Dashboard       |
                          | Next.js + TypeScript |
                          +----------+-----------+
                                     |  Clerk JWT
                                     v
+----------------+      +-----------+------------+
| Customer Apps  | ----> |  Go API (chi)          |
| SDK or REST    |       |  OpenAPI generated     |
| Bearer API key |       +-----------+------------+
+----------------+                   |
                  +-----------------+------------------+
                  |                 |                  |
                  v                 v                  v
        +--------------+   +--------------+   +--------------+
        |  PostgreSQL  |   | Redis+Asynq  |   | Object Store |
        |  (primary)   |   | (job queue)  |   | (future)     |
        +--------------+   +------+-------+   +--------------+
                                  |
                                  v
                       +----------+-----------+
                       |  Worker process(es)  |
                       |  provider adapters   |
                       +----------+-----------+
                                  |
        +--------+--------+-------+-------+--------+
        |        |        |       |       |        |
        v        v        v       v       v        v
     Resend  SendGrid    SES   Twilio    FCM    Webhook

Provider delivery webhooks
  -> API webhook endpoints
  -> signature verification
  -> event normalization
  -> delivery timeline update
  -> customer webhook fanout
```

---

## Request lifecycle

### Send notification

1. Customer calls `POST /v1/notifications` with an API key.
2. API middleware authenticates the key (hash lookup), resolves org/project/environment, enforces rate limits.
3. Request validated against OpenAPI schema.
4. Idempotency key checked — return cached response if duplicate.
5. Template resolved, variables validated.
6. `notifications` and `deliveries` records created in PostgreSQL.
7. Delivery job enqueued in Redis via Asynq.
8. API returns `{ notification_id, delivery_id, status: "queued" }`.
9. Worker dequeues job, renders template, checks suppression list and preferences.
10. Worker selects provider connection, calls provider adapter.
11. Adapter sends to provider API, records `provider_message_id` and initial status.
12. Provider delivers and fires a webhook event.
13. API receives webhook, verifies signature, normalizes event into `delivery_events`.
14. Delivery status updated in `deliveries`.
15. Customer webhook endpoint called if configured.

---

## Multi-tenancy

Every database query is scoped by `organization_id`. API keys resolve to an `(organization_id, project_id, environment_id)` tuple. This tuple is stored in request context and threaded through every service call and database query. Cross-tenant access is prevented at the query layer, not just the handler layer.

---

## Authentication

Two separate auth systems:

| Caller | Mechanism | Used for |
|---|---|---|
| Dashboard users | Clerk JWT | Dashboard API routes and session management |
| Customer apps | API key (hashed) | All `POST /v1/notifications` and public API calls |

API keys are generated with a random secret, stored as a hash, and exposed only once at creation. The `key_prefix` field (first 8 chars) is stored in plaintext for log correlation.

Clerk handles dashboard signup. On `user.created` webhook from Clerk, the API bootstraps an `organizations` record, a default `projects` record, and three `environments` (development, staging, production).

---

## Provider adapter pattern

All provider-specific code lives in `internal/providers/`. No service outside that package imports a provider SDK directly.

Every email provider implements:

```go
type EmailProvider interface {
    SendEmail(ctx context.Context, input SendEmailInput) (SendResult, error)
    VerifyWebhook(ctx context.Context, headers map[string]string, body []byte) error
    ParseWebhook(ctx context.Context, headers map[string]string, body []byte) ([]NormalizedProviderEvent, error)
}
```

Provider credentials are stored encrypted in `provider_connections.encrypted_config`. The worker decrypts at runtime using the `ENCRYPTION_MASTER_KEY`.

---

## Queue and worker

Asynq uses Redis as a broker. Jobs are serialized as JSON. The worker process runs separately from the API and can be scaled independently.

Retry policy per job type:

```
email delivery:   up to 5 attempts, exponential backoff
customer webhook: up to 10 attempts, exponential backoff
```

Failed jobs after max retries move to `failed_permanent` state and are visible in the dashboard for manual inspection and replay.

---

## Delivery state machine

```
queued -> processing -> sent -> delivered
                     -> bounced
                     -> complained
              -> failed_temporary -> queued (retry)
                                  -> failed_permanent
queued -> cancelled
queued -> suppressed
```

Every state transition creates a `delivery_events` record. Provider events are stored even if they do not change the final state. All event processing is idempotent on `provider_event_id`.

---

## Database

PostgreSQL 16. All tables use ULID primary keys for sortability and URL-safe public IDs.

Migrations managed by goose in `internal/db/migrations/`. All migrations must be reversible. Run with `make migrate`.

Queries written as plain SQL in `internal/db/queries/`. sqlc generates type-safe Go from these. Run `make generate` after any query change.

---

## Observability

Every request log includes: `request_id`, `organization_id`, `project_id`, `environment_id`, `api_key_prefix`.

Every delivery log includes: `notification_id`, `delivery_id`, `channel`, `provider`, `provider_message_id`, `status`, `attempt_count`, `latency_ms`, `error_code`.

OpenTelemetry is used for traces and metrics from Phase 1 onward. Structured JSON logs via `slog`.

---

## Security

- API keys hashed with bcrypt or argon2 before storage.
- Provider credentials encrypted at rest (AES-256-GCM, envelope encryption).
- Provider webhook signatures verified before any processing.
- Outbound customer webhooks signed with HMAC-SHA256.
- All secrets redacted from logs before writing.
- Tenant isolation enforced at the database query layer.
- Rate limits enforced per API key and per project.

---

## Local development

```bash
make setup    # install deps and hooks
make dev      # start DB + Redis, migrate, generate code
make api      # :8080
make worker   # background worker
make web      # :3000
```

See `README.md` section 10 for full OS-specific setup instructions.
