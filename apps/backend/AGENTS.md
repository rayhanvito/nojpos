# NojPOS Laravel API Guide

Read [root guide](../../AGENTS.md), [PRDPOSJA](../../PRDPOSJA.md), [docs index](../../docs/README.md), [story progress](../../docs/STORY_PROGRESS.md), [backend story](../../docs/STORY_BACKEND.md), [integration story](../../docs/STORY_INTEGRATION.md), and [bug tracker](../../docs/BUG_TRACKER.md) before editing.

Backend work must follow the story order. Do not jump to Super Admin sensitive actions, billing writes, support writes, announcement broadcast, or web admin integration before their dependencies are ready.

## Run And Verify

```powershell
cd apps/backend
composer install
Copy-Item .env.example .env
php artisan key:generate
php artisan migrate:fresh --seed
php artisan serve --host=0.0.0.0 --port=8000
php artisan test
php artisan route:list --path=api/v1
```

## Local Structure

```text
routes/api.php                    /api/v1 contract
app/Http/Controllers/Api/V1/      thin request adapters
app/Http/Middleware/              role and idempotency middleware
app/Http/Requests/                FormRequests for new endpoints
app/Http/Resources/               API resources for output contracts
app/Policies/                     policy classes for domain actions
app/Services/                     transactional business services
app/Models/                       UUID tenant models
app/Models/Concerns/              tenant/UUID traits
app/Models/Scopes/                business scope
app/Support/                      response, audit, shared infrastructure
database/migrations/              additive schema changes
tests/Feature/                    API/tenant/security tests
tests/Unit/                       pure service/value-object tests
```

## MUST Rules

- All API routes stay under `/api/v1`, require `Accept: application/json`, and return the standard envelope.
- Derive `business_id` from `$request->user()`. Every query-builder read/write MUST explicitly scope tenant data.
- A free `business_id`, `cashier_id`, outlet, device, shift ID, or role from client input is never proof of authority.
- Use `controller -> request validation -> authorization/policy -> service transaction -> model/query -> resource/envelope`.
- Controllers must not own multi-table money/stock logic.
- Money uses integer rupiah only.
- Store UTC timestamps and use outlet timezone for reports, shifts, attendance, and document-day boundaries.
- Sensitive writes must be DB-transactional, idempotent, authorized, and audited in the same transaction.
- Sensitive writes include money, stock, shift, cash, payment, void/refund, user access, toko status, paket/langganan changes, access bantuan, and announcement broadcast.
- Never log passwords, PINs, bearer tokens, full PII, receipt content, or payment references.

## Current Known Gaps

Track and update these in `../../docs/BUG_TRACKER.md`:

- Tenant scope query-builder audit is still required.
- Web admin session strategy is not final.
- Dashboard summary endpoint is missing.
- Platform billing/support/activity/system/announcement/operator domains are missing or incomplete.
- Some sensitive POS flows need stronger audit/idempotency/concurrency verification.

## Add An Endpoint Safely

1. Confirm story and dependency in `../../docs/STORY_BACKEND.md`.
2. Confirm integration phase in `../../docs/STORY_INTEGRATION.md` if endpoint is for web/mobile integration.
3. Add FormRequest, policy/action authorization, service, resource, and route.
4. Scope every resource by authenticated business/outlet/device context.
5. Use DB transaction and row lock for mutable financial/stock/state transitions.
6. Add idempotency middleware/service for write commands.
7. Add audit event for privileged/money/stock/state changes.
8. Add tests for 401/403/404/409/422, tenant/outlet isolation, policy, idempotency, rollback, and concurrency when relevant.
9. Update story docs and bug tracker.

## Definition Of Done

Run:

```powershell
php artisan test
php artisan route:list --path=api/v1
```

For money/stock/privileged endpoints, done requires tests or documented manual verification for:

- tenant and outlet isolation;
- role and terminal-actor authorization;
- idempotent retry and same-key/different-body conflict;
- atomic rollback;
- state/concurrency conflict;
- audit event persistence.

Final report format:

```text
Built: <changed>
Verified: <commands and results>
Blocker: <precise blocker or none>
```
