# NojPOS Monorepo Guide

Read this file first, then the closest package guide.

## Product And Source Of Truth

NojPOS is an online-first POS SaaS for Indonesian UMKM. Cashiers operate shared terminals; owners and admins control business operations; superadmins operate the SaaS layer.

Source order:

1. Latest user instruction.
2. Closest `AGENTS.md`.
3. [PRDPOSJA](PRDPOSJA.md).
4. [Active docs index](docs/README.md).
5. [Single-agent workflow](docs/SINGLE_AGENT_WORKFLOW.md) and [task queue](docs/TASK_QUEUE.md).
6. [Story progress board](docs/STORY_PROGRESS.md).
7. Story files by area: [Backend](docs/STORY_BACKEND.md), [Mobile Kasir](docs/STORY_MOBILE_KASIR.md), [Web Admin](docs/STORY_WEB_ADMIN.md), and [Integration](docs/STORY_INTEGRATION.md).
8. [API contracts](docs/API_CONTRACTS/README.md), [validation checklist](docs/VALIDATION_CHECKLIST.md), and [bug tracker](docs/BUG_TRACKER.md).
9. Actual code and tests for implementation details and current constraints.
10. [Task breakdown](Task%20%26%20Subtask%20Breakdown.md) only as a legacy/reconciled reference. Do not execute it blindly.

When business rules are ambiguous, use `PRDPOSJA.md` first. State which default was used in the change summary; do not invent a different rule. When execution sequence is ambiguous, follow `docs/SINGLE_AGENT_WORKFLOW.md`, `docs/TASK_QUEUE.md`, and the story docs under `docs/`, not the legacy task breakdown. Use one active coding agent and one active task at a time; do not run parallel coding agents for this repo until the product owner explicitly changes the workflow.

## Active Packages

| Package | Purpose | Guide | Setup and verification |
| --- | --- | --- | --- |
| `apps/backend` | Laravel API, Sanctum, tenant data, money/stock rules | [backend guide](apps/backend/AGENTS.md) | `composer install`; `php artisan migrate:fresh --seed`; `php artisan test`; `php artisan route:list --path=api/v1` |
| `apps/cashier` | Flutter shared-terminal cashier app | [cashier guide](apps/cashier/AGENTS.md) | `flutter pub get`; `flutter analyze`; `flutter test`; `flutter build apk --debug` |
| `apps/web` | Next.js Admin Web preview for Tenant Admin and Platform Super Admin | [admin web guide](apps/web/AGENTS.md) | `npm ci`; `npm test`; `npm run typecheck`; `npm run lint`; `npm run build`; `npx playwright test --list` |

Admin Web now lives in `apps/web` as a UI-first preview surface. Backend integration starts only after the session and integration gates in `docs/STORY_INTEGRATION.md` are approved.

Local API for Android emulator: `http://10.0.2.2:8000/api/v1`.

## MUST Invariants

1. **Tenant and actor:** Every domain query MUST scope `business_id` from the authenticated principal. Clients MUST NOT choose tenant ownership or acting `cashier_id`; server-side terminal context/policy MUST derive or verify actor, outlet, device, and shift together.
2. **Money and stock:** All money MUST be integer rupiah. The server is the only calculator for price, discount, promotion, tax, service, rounding, totals, and stock. Flutter renders server results.
3. **Sensitive writes:** Money, stock, shift, cash, payment, refund, void, and privileged writes MUST be database-transactional, idempotent by `business_id + endpoint + request_hash`, and create immutable audit events.
4. **Authorization:** Server-side policy/gate/action authorization is mandatory. Hiding a Flutter control is never security.
5. **Terminal:** Shift, cash, lock, checkout, and attendance writes MUST use a valid enrolled `device_id` and server-verified terminal actor context.
6. **Time:** Store timestamps in UTC. Render and calculate business-day boundaries using the outlet timezone, defaulting to `Asia/Jakarta` when no configured outlet timezone exists.
7. **Response envelope:** Success is `{ "data": ..., "meta": ... }`; failure is `{ "error": { "code": "...", "message": "...", "details": ... } }`.
8. **PIN and PII:** PINs MUST be hashed; raw PIN, password, token, customer PII, and payment references MUST NEVER enter logs, audit payloads, screenshots, fixtures, or error messages.
9. **Online-first:** NEVER add Drift, a general offline database, `/sync/pull`, `/sync/push`, or a general write queue. Checkout retry behavior is only the bounded PRD §2.1 mechanism.

### Standard Error Codes

| HTTP | Use |
| --- | --- |
| `401` | `UNAUTHENTICATED` - session/token invalid |
| `403` | `FORBIDDEN` - tenant, outlet, role, or policy denied |
| `404` | `NOT_FOUND` - resource absent within allowed scope |
| `409` | `IDEMPOTENCY_CONFLICT`, `CONFLICT_REVISION`, `ORDER_LOCKED`, `QUOTE_STALE`, concurrent state conflict |
| `422` | `VALIDATION_ERROR`, invalid state, required PIN/reason, `SHIFT_NOT_OPEN` |

## Architecture Rules

### Flutter

MUST follow `ui -> provider/notifier -> repository interface -> ApiClient`.

```dart
// DO: feature notifier calls a repository.
final result = await ref.read(transactionRepositoryProvider).createTransaction(draft);

// DON'T: create Dio or serialize HTTP directly in a widget.
await Dio().post('/transactions');
```

- Put new code in `apps/cashier/lib/features/<feature>/{pages,widgets,providers,repositories,models}`.
- Widgets MUST NOT call HTTP, persist tokens, calculate money, or own idempotency retry behavior.
- Feature state MUST own its own busy/error/result state. Do not add unrelated work to `NojposSessionNotifier`; it is a known fix-forward hotspot.
- Repositories expose an interface and an API implementation. DTO parsing belongs in repositories/models, not widgets.

### Laravel

MUST use `controller -> authorization -> service/transaction -> model/query -> ApiResponse`.

```php
// DO: controller validates, authorizes, then delegates.
$data = $request->validated();
$this->authorize('close', $shift);
return ApiResponse::success($service->close($context, $shift, $data));

// DON'T: accept a free cashier ID and perform multi-table money writes in a controller.
DB::table('transactions')->insert($request->all());
```

- Controllers live in `app/Http/Controllers/Api/V1/`; keep them thin.
- Form requests live in `app/Http/Requests/`; policies in `app/Policies/`; business services in `app/Services/`; API output formatters/resources in `app/Http/Resources/`.
- New domain models MUST use UUIDs, timestamps, soft delete where required, and tenant scope according to the domain model.
- Query-builder reads MUST explicitly add `business_id`; do not assume Eloquent global scopes apply to `DB::table`.
- A service handling a sensitive write MUST wrap state, ledger/movement, and audit insert in one `DB::transaction` and lock mutable aggregates as needed.

## Conventions And Recipes

### Add an endpoint

1. Confirm PRD scope and `[DECISION]` rule.
2. Add a FormRequest and policy/action authorization.
3. Implement transaction/business logic in a service.
4. Scope every resource by token-derived business and verified outlet/device/actor relation.
5. Add route under `/api/v1`, then apply auth, role/policy, and idempotency middleware where required.
6. Return only the standard envelope through `ApiResponse` or an API Resource.
7. Add feature tests for tenant isolation, authorization, validation, idempotency, rollback, and concurrency when money/stock/state changes.

### Add a Flutter feature

1. Create the feature folder and repository interface/API implementation.
2. Add a narrow provider/notifier; do not add new feature state to the global session notifier.
3. Use typed `ApiException` and render loading, empty, validation, forbidden, conflict, and retry states.
4. Use integer values for money and display rupiah only at the presentation boundary.
5. Add unit/provider tests, a widget test for interaction, and an integration test when the flow crosses authentication, shift, checkout, or inventory.

### Add a migration, audit, report, or idempotent write

- Migrations MUST be additive and reversible. Inspect existing data before adding constraints. Never silently mutate financial history.
- Audit events use `domain.action`, actor, business, outlet/device/request context, before/after, and are written in the same transaction as the effect.
- Reports MUST aggregate server-side, paginate lists, apply outlet timezone, and never calculate totals in Flutter.
- Idempotent writes MUST reserve the key atomically before executing, store an in-progress/completed response, and return the saved response for the same body.

## Testing And Definition Of Done

Every money, stock, or privileged change MUST test: tenant isolation, role/outlet authorization, idempotency, transaction rollback, and concurrent/state-conflict behavior. Track planned and completed work in the story docs before reporting Codex completion.

Before declaring a change done:

- [ ] Scope is limited and no unrelated dirty work was reverted.
- [ ] Backend: `php artisan test` and route check pass when API changed.
- [ ] Flutter: `flutter analyze`, `flutter test`, `flutter build apk --debug`; install on `emulator-5554` when available.
- [ ] Formatting/lint is clean.
- [ ] Migration is additive and rollback considered.
- [ ] Audit exists for privileged/money/stock action.
- [ ] No client-side money or stock calculation was added.
- [ ] `PRDPOSJA.md`, story docs, and bug tracker are updated when contract changed.

## Security And Logging

Log structured identifiers only: `request_id`, `idempotency_key`, `business_id`, `outlet_id`, `device_id`, resource ID, and error code. Redact bearer token, password, PIN, full phone/email, receipt content, and payment reference.

Do not commit `.env`, database dumps, device tokens, or production credentials. Demo credentials may appear only in dev seed documentation, never production code.

## Current Fix-Forward Exceptions

These are known gaps, not patterns to copy: client-trusted checkout price/rounding, client-trusted cashier context, persistent multi-item checkout outbox, controller-heavy writes, missing policy layer, UTC-only outlet behavior, and incomplete audit/idempotency coverage. P0 production blockers must be fixed before broad feature expansion. Sensitive writes must be tenant-scoped, authorized, idempotent, transactional, row-locked where state can race, and audited inside the same transaction. Track active gaps in `docs/BUG_TRACKER.md` and the relevant story file.

Final report format:

```text
Built: <changed>
Verified: <commands and results>
Blocker: <precise blocker or none>
```
