# NojPOS Laravel API Guide

Read [root guide](../../AGENTS.md), [PRD v2](../../NOJPOS_POS_FLUTTER_BACKEND_PRD.md), [architecture](../../docs/ARCHITECTURE.md), [analysis findings](../../docs/ANALYSIS_FINDINGS.md), and [improvement plan](../../docs/IMPROVEMENT_PLAN.md) before editing a sensitive flow. Follow the Wave 0 -> Wave 1 -> Wave 2 order for the current repo unless a later user instruction explicitly narrows scope without violating P0 guardrails.

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

Laravel requires PHP `^8.3`, Laravel `^13.8`, and Sanctum `^4.3`. `composer run dev` starts Laravel, queue listener, logs, and Vite for local development.

## Local Structure

```text
routes/api.php                    /api/v1 contract
app/Http/Controllers/Api/V1/      thin request adapters
app/Http/Middleware/              role and idempotency middleware
app/Http/Requests/                FormRequests for new endpoints
app/Http/Resources/               API resources for new output contracts
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

`Http/Requests`, `Http/Resources`, and `Policies` are the target locations for new work even though parts of the current code still validate and serialize in controllers.

## MUST Rules

- Current backend work is Laravel API only. Do not start `apps/web`, Next.js admin web, Drift/offline DB, or broad feature expansion from backend tasks.
- P0 production blockers in `docs/ANALYSIS_FINDINGS.md` must be handled before non-critical product expansion. Admin web remains paused until P0 backend contracts are green and explicitly approved.
- All API routes stay under `/api/v1`, require `Accept: application/json`, and return the standard `ApiResponse` envelope.
- Derive `business_id` from `$request->user()`. Every `DB::table` query MUST explicitly scope it. A free `business_id`, `cashier_id`, outlet, device, or shift ID is never proof of authority.
- Use `BelongsToBusiness` only for Eloquent models. It does not protect query-builder calls.
- Validate with a FormRequest; authorize through a policy/action gate; then delegate to a service. Controllers MUST NOT own multi-table money/stock logic.
- Sensitive service writes MUST use `DB::transaction`; lock mutable aggregate rows when concurrent transitions are possible.
- Require atomic idempotency reservation for money, stock, shift, cash, payment, void/refund, and document-finalization writes. Same key/body returns the saved result; same key/different body is `409`.
- Write immutable audit data inside the same transaction as the effect. Include actor, business, outlet, device/terminal, request ID, before/after, and stable `domain.action` name.
- If a task touches money, shift, payment, stock, auth, device, or audit, complete `docs/EXECUTION_CHECKLIST.md` and include the tests run or explicitly not run.
- Money uses integer rupiah only. Load price/config/rounding from server-owned records; never accept client `grand_total`, sell price, or authoritative rounding.
- Store UTC timestamps and use outlet timezone for reports, shifts, attendance, and document-day boundaries.

```php
// DO: tenant-safe query-builder access.
$shift = DB::table('shift_sessions')
    ->where('business_id', $context->businessId)
    ->where('id', $shiftId)
    ->lockForUpdate()
    ->first();

// DON'T: actor or tenant ownership from unverified request data.
DB::table('transactions')->insert([
    'cashier_id' => $request->input('cashier_id'),
    'grand_total' => $request->input('grand_total'),
]);
```

## Add An Endpoint Safely

1. Read the matching PRD section and apply its `[DECISION]` default.
2. Add an additive migration/model only if the domain record is needed.
3. Create a FormRequest, policy, service method, API Resource, and route.
4. Resolve terminal actor/device/outlet/shift relation server-side.
5. Put financial/stock state, movement rows, and audit event in one transaction.
6. Add idempotency middleware/service before exposing the command.
7. Add feature tests for 401/403/404/409/422, tenant/outlet isolation, policy, retry, rollback, and concurrency.

## Reporting And Migrations

- Reports aggregate only on the server, use outlet-local boundaries, paginate list data, and have explicit accounting definitions.
- New migrations MUST be additive and have a meaningful `down`. Inspect real data before adding constraints/indexes or backfills.
- Do not edit historical transaction, payment, cash, stock, or audit rows to “correct” them. Use an explicit reversal/adjustment document.

## Current Fix-Forward Exceptions

Do not copy these patterns: `TransactionQuoteService` currently trusts client unit price/rounding; `ShiftController` accepts client cashier context and has incomplete close/cash protection; `PaymentController` lacks atomic state transitions; `EnsureIdempotency` has a read-execute-insert race; many controller writes bypass FormRequests, policies, services, Resources, or in-transaction audit. Details are in [analysis findings](../../docs/ANALYSIS_FINDINGS.md).

## Tests And Definition Of Done

Run:

```powershell
php artisan test
php artisan route:list --path=api/v1
```

For money/stock/privileged endpoints, a change is done only when tests prove:

- business and outlet isolation;
- role and terminal-actor authorization;
- idempotent retry and same-key/different-body conflict;
- atomic rollback when one write fails;
- state/concurrency conflict behavior;
- audit event persistence with the final business result.

Never log passwords, PINs, bearer tokens, full PII, or payment references. Log only safe correlation fields: `request_id`, `idempotency_key`, `business_id`, `outlet_id`, `device_id`, resource ID, and error code.
