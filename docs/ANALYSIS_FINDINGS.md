# NojPOS Analysis Findings

Status: evidence-based read-only assessment, strengthened for execution review  
Snapshot date: 2026-06-19  
Branch observed: `feat/flutter-ui-consistency`  
Commit observed: `de07cbb`  
Worktree state: dirty; these findings describe the current local workspace, not a clean merge-ready branch.

PRD reference: `NOJPOS_POS_FLUTTER_BACKEND_PRD.md` v2.0  
Architecture reference: `docs/ARCHITECTURE.md`  
Plan reference: `docs/IMPROVEMENT_PLAN.md`

Severity scale:

- **P0** = production blocker for correctness, security, money, stock, actor integrity, or audit integrity.
- **P1** = high-risk architecture, product-contract, scalability, or workflow divergence.
- **P2** = maintainability, DX, observability, documentation, or cleanup issue.

## Executive Verdict

The current NojPOS workspace is **not production-safe for real outlet money flow yet**. It is useful as a strong implementation baseline, but the operational core must not be treated as ready for live UMKM transactions until the P0 issues below are fixed and protected by tests.

The highest-risk area is **not UI**. It is the backend contract for checkout, payment, shift, actor context, idempotency, and audit. Building more screens, reports, promotions, inventory documents, or admin web before these P0s are green will make the product look more complete while increasing the amount of behavior that may later need to be migrated.

## Source-of-Truth Hierarchy

Use this hierarchy when documents disagree:

1. `NOJPOS_POS_FLUTTER_BACKEND_PRD.md` v2.0 for product and behavior.
2. `docs/ARCHITECTURE.md` for current repo reality.
3. This file for current risk findings.
4. `docs/IMPROVEMENT_PLAN.md` for execution sequence.
5. `docs/DESIGN.md` for UI/design contracts.
6. `Task & Subtask Breakdown.md` only as a legacy sequencing reference after reconciliation.

Important: `Task & Subtask Breakdown.md` currently mentions Drift and Next.js admin web, but the current repo has no Drift dependency and no `apps/web`/Next.js app. Do not execute that document blindly.

## P0 Findings

### P0-1: Checkout trusts client price, discount inputs, and rounding

**Evidence:** `TransactionQuoteService` uses request `items.*.unit_price` and `rounding` directly when deriving totals (`apps/backend/app/Services/TransactionQuoteService.php:45-83`). `TransactionController` validates and accepts `items.*.unit_price`, client totals, `rounding`, `rounding_total`, and `grand_total` (`apps/backend/app/Http/Controllers/Api/V1/TransactionController.php:259-290`).

**Risk:** A modified client, patched APK, proxy, or manual API request can submit a lower unit price, arbitrary discount, or negative/incorrect rounding. The server then quotes and commits the manipulated values. A `TOTAL_MISMATCH`-style check cannot protect against tampering if the server quote itself is derived from client-controlled price fields.

**Business impact:** Sales, cash drawer, tax/service, stock valuation, reports, shift close, and future refunds can become financially unreliable.

**PRD divergence:** PRD §2 “Server is the calculator”, §5.4 checkout, §6.4 deterministic calculation, §12.3.

**Required direction:**

- Client sends product IDs, quantity, selected customer/order context, and discount requests only.
- Server loads product sell price, tax/service/rounding config, entitlement, and discount policy.
- Server returns a final quote snapshot with `quote_id`, expiry, revision/config version, frozen totals, and checkout idempotency key.
- Checkout accepts `quote_id` + expected revision/snapshot, not authoritative price/totals from Flutter.
- Old client-total fields may remain temporarily for staged migration, but must no longer be authoritative.

**Acceptance tests:**

- A request with a manipulated lower `unit_price` still charges the server-owned product price.
- A request with negative or arbitrary `rounding` is ignored or rejected according to server config.
- Reusing a stale quote returns `409 quote_stale`.
- Same quote/key submitted twice creates one transaction.

### P0-2: Operational actor, device, outlet, and shift context are client-trusted within a tenant

**Evidence:** Shift open accepts `cashier_id` and only checks that outlet/device/user belong to the same business (`apps/backend/app/Http/Controllers/Api/V1/ShiftController.php:17-54`). Transaction quote/store validates `outlet_id`, `device_id`, `cashier_id`, and `shift_id` independently by same-business membership (`apps/backend/app/Http/Controllers/Api/V1/TransactionController.php:237-248`). PIN switch returns a cashier payload but does not replace the Sanctum principal with a durable terminal actor context (`apps/backend/app/Http/Controllers/Api/V1/AuthController.php:92-156`).

**Risk:** A valid tenant token can nominate a different same-business cashier, device, outlet, or shift if it knows the UUIDs. Tenant isolation tests do not prove actor integrity.

**Business impact:** Audit logs may name the wrong cashier, a cashier may operate another device/outlet, and shift/payment/void responsibility can be disputed.

**PRD divergence:** PRD §4.1 role/authorization, §5.6 shift lifecycle, §5.8 device enrollment, §12.1 authorization.

**Required direction:**

- Introduce a server-side terminal session after PIN verification.
- Terminal session binds account, acting cashier, outlet, enrolled device, and permission context.
- Sensitive writes derive actor/device/outlet from the terminal session or verify the full tuple against the open shift.
- Same-business membership alone is never sufficient authority.
- Device assignment/revocation must be enforced for terminal-bound writes.

**Acceptance tests:**

- Cashier A cannot open/close/use cashier B’s shift in the same business.
- Device assigned to Outlet A cannot submit shift/checkout for Outlet B unless explicitly assigned.
- Transaction cannot combine valid same-business outlet/device/cashier/shift IDs that do not belong together.
- Revoked/unknown device returns `403 device_not_enrolled`.

### P0-3: Payment writes are not atomic and can mutate inappropriate transactions

**Evidence:** `PaymentController@store` creates payment records and then updates/recalculates transaction status as separate logic. `confirm` follows a similar state update path (`apps/backend/app/Http/Controllers/Api/V1/PaymentController.php`). The inspected flow does not sufficiently prove allowed transaction state, remaining amount, shift state, or provider-backed authorization before mutation.

**Risk:** A failure between insert and status update can leave partial state. A caller can attempt to add/confirm payments against held, voided, already paid, overpaid, expired, or wrong-shift transactions unless every path is guarded elsewhere.

**Business impact:** Overpayment, double-confirmation, false paid status, incorrect close-shift cash totals, and unreliable payment reports.

**PRD divergence:** PRD §5.4.8 async payment, §5.4.9 checkout, §12.3 transaction integrity.

**Required direction:**

- Move payment mutation to a `PaymentService`.
- Wrap payment creation/confirmation/status transition/audit in one DB transaction.
- Lock the transaction row before evaluating state.
- Enforce allowed state transitions and remaining amount.
- Async confirmation must be provider/webhook or explicitly approved boundary, not arbitrary client status mutation.

**Acceptance tests:**

- Cannot add payment to voided/closed/paid transaction where not allowed.
- Cannot overpay beyond remaining amount unless explicitly supported and modeled as change.
- Parallel confirm calls settle once.
- Any injected failure rolls back both payment and transaction status.

### P0-4: Idempotency middleware has a race window and incomplete coverage

**Evidence:** `EnsureIdempotency` performs lookup, executes `$next($request)`, then creates the idempotency row (`apps/backend/app/Http/Middleware/EnsureIdempotency.php:31-58`). `POST /shifts/open` and `POST /shifts/{shift}/cash-movements` are not protected by idempotency middleware (`apps/backend/routes/api.php:68-69`).

**Risk:** Concurrent requests with the same key can both execute before the snapshot row exists. A unique-key failure may happen after the business effect is already applied. Shift open and cash movement retries can duplicate money/state.

**Business impact:** Double transactions, double cash movement, double close, duplicate payment state, and irreconcilable cashier reports.

**PRD divergence:** PRD §2.1 online-first retry semantics, §5.6 shift/cash, §12.3, §15.

**Required direction:**

- Implement an idempotency service with atomic reservation before business effect.
- Store status `in_progress`, `completed`, and optionally `failed_retryable`/`failed_terminal`.
- Same key + same request hash returns stored response.
- Same key + different request hash returns `409 idempotency_mismatch`.
- Apply to all money, stock, shift, payment, void/refund, and document-finalization writes.

**Acceptance tests:**

- Parallel same-key requests produce one business effect.
- Same key + different body returns `409 idempotency_mismatch`.
- Retrying after a completed response returns byte-equivalent JSON and original status.
- Shift open and cash movement are idempotent.

### P0-5: Shift close and cash movement lack required authorization, reason, audit, and locking guarantees

**Evidence:** `cashMovement` allows nullable `reason`, inserts directly, has no idempotency middleware, and no audit call (`apps/backend/app/Http/Controllers/Api/V1/ShiftController.php:75-104`). `close` accepts only `actual_cash`, calculates totals, updates status, and audits after update without row locking or fresh PIN/variance/approver logic (`apps/backend/app/Http/Controllers/Api/V1/ShiftController.php:107-162`). Flutter close submits only `actual_cash` (`apps/cashier/lib/features/shift/repositories/shift_repository.dart:75-85`).

**Risk:** Cash movements can be duplicated or unexplained. Shift close can race. An unapproved staff member can close a shift if the endpoint accepts their tenant context. Non-zero variance can be finalized without reason/approval.

**Business impact:** Cash disputes, weak accountability, and unusable shift close report for real-world cashier operations.

**PRD divergence:** PRD §5.6 shift lifecycle, §7.4 cash report, §12.3, §15.

**Required direction:**

- Introduce `ShiftService`.
- Open/cash movement/close use idempotency and DB row locks.
- Cash movement reason is mandatory.
- Fresh PIN is required for close and configured high-risk actions.
- Non-zero variance requires reason; variance beyond threshold requires owner/admin approval.
- Close report snapshot is immutable and includes summary, cashier, device, outlet, time, expected, declared, variance, reasons, and authorizer.

**Acceptance tests:**

- Missing cash movement reason returns 422.
- Duplicate cash movement retry creates one row.
- Double close with new key cannot overwrite declared cash.
- Close with variance requires reason.
- Close beyond threshold requires authorized approver.

### P0-6: Audit is inconsistent and not guaranteed to commit with the business effect

**Evidence:** Transaction creation commits the database transaction, then calls `Nojpos::audit` afterward (`apps/backend/app/Http/Controllers/Api/V1/TransactionController.php:155-224`). Shift open/close call static audit helper, but cash movement and several CRUD paths do not. `Nojpos::audit` is a direct insert helper with limited actor/device/request context (`apps/backend/app/Support/Nojpos.php`).

**Risk:** A business effect can commit while audit fails. Some sensitive actions are not audited. Audit semantics vary by endpoint.

**Business impact:** The system cannot reliably answer who did what, on which device, under which shift, with what before/after state.

**PRD divergence:** PRD §2, §5.6, §12.1, §13.1.

**Required direction:**

- Introduce `AuditLogService` called inside the same DB transaction as the domain mutation.
- Include actor, approver, terminal session, device, outlet, request ID, idempotency key, action name, entity type/id, before, after, event version, timestamp.
- Define mandatory audit actions for checkout, payment, shift open/close, cash movement, void/refund, inventory finalization, PIN/lock/unlock, settings changes, device revoke.

**Acceptance tests:**

- Failure during audit rolls back the business mutation for mandatory audited actions.
- Each sensitive write creates exactly one audit event per idempotent business effect.
- Audit includes actor and device context.

## P1 Findings

### P1-1: Persistent checkout outbox conflicts with PRD v2 bounded retry buffer

**Evidence:** The checkout outbox persists full checkout drafts to `FlutterSecureStorage`, exposes a queue-like retry center, and permits retries across lifecycle events (`apps/cashier/lib/core/outbox/checkout_outbox.dart:151-240`).

**Risk:** This is materially more than the PRD §2.1 single in-flight retry buffer. It retains potentially stale prices/context and creates a multi-item queue concept the product contract excludes.

**Required direction:** Keep one in-memory indeterminate checkout record only. After process restart, query backend by idempotency key before allowing a new sale. Do not extend outbox semantics to other domains.

**Dependency:** Do not remove persistence until backend idempotency lookup/recovery is safe.

### P1-2: Transaction list has an N+1 query pattern

**Evidence:** `TransactionController@index` fetches rows and calls `transactionPayload` for each; payload loads cashier/customer/items/payments per transaction (`apps/backend/app/Http/Controllers/Api/V1/TransactionController.php`).

**Risk:** Busy outlets will degrade as transaction count grows.

**Required direction:** Add pagination and list/detail projections. Batch-load related users/customers/items/payments or use resources with eager loading.

### P1-3: Timezone defaults conflict with Indonesia-first reporting and shift rules

**Evidence:** The Laravel application timezone remains UTC, which is correct for storage. Wave 7 adds `outlets.timezone` (default `Asia/Jakarta`) and `BusinessClock`; sales reports, attendance date filters, and generated purchase/transaction numbers now resolve outlet-local boundaries before querying/storing UTC timestamps.

**Risk:** Future shift auto-expiry and store open/close are not yet implemented, so their lifecycle rules still need to adopt `BusinessClock` when introduced. Existing multi-outlet aggregate reports require an explicit `outlet_id` to select a business timezone.

**Required direction:** Preserve UTC storage and route every new business-day calculation through `BusinessClock`; add a timezone-aware rendering strategy for Flutter surfaces that display historical timestamps.

### P1-4: Controller business logic is too concentrated

**Evidence:** Financial and multi-table behavior lives inside controllers. Examples include transaction store, shift close, payment store/confirm, inventory purchase, and void.

**Risk:** Authorization, transaction boundaries, audit, row locks, and validation are duplicated or missed.

**Required direction:** Move domain effects to services: `CheckoutService`, `PaymentService`, `ShiftService`, `TerminalSessionService`, `InventoryDocumentService`, `AuditLogService`, `IdempotencyService`.

### P1-5: Intended Flutter feature-first layering is bypassed by very large state/UI modules

**Evidence:** `operations_screen.dart` = 2,754 lines, `pos_dialogs.dart` = 1,906 lines, `payment_screen.dart` = 1,490 lines, `NojposSessionNotifier` = 1,102 lines.

**Risk:** Unrelated workflows share global busy/error/session state. Small changes can regress unrelated flows. Tests require broad setup. Review becomes difficult.

**Required direction:** Extract feature controllers/providers incrementally, protected by tests. Do not rewrite the full POS UI in one change.

### P1-6: Local orders and held server transactions are competing sources of truth

**Evidence:** `LocalOrderRepository` generates local IDs/numbers, while held orders also map to `transactions.status=held` in backend flow.

**Risk:** “Saved order” can have two numbering and lifecycle systems, contrary to PRD parked order revision/lease contract.

**Required direction:** Create one server-backed parked order aggregate with server number, revision, lease lock, and checkout handoff. Remove local document-number authority.

### P1-7: Policy layer is incomplete

**Evidence:** No stable policy layer is evident for transaction/shift/inventory/staff/config. Authorization is mostly route role middleware or controller checks.

**Risk:** New endpoints can forget outlet assignment, terminal actor, or approver rules.

**Required direction:** Add explicit policies/action authorizers and test owner/admin/cashier/outlet/device matrices.

### P1-8: Required PRD domains are absent or partial

Missing or materially incomplete domains include:

- terminal session/device enrollment,
- genuine heartbeat,
- store open/close,
- settings aggregate/security settings,
- refund,
- complete sales/product/top-10/void reports,
- inventory count/waste/transfer lifecycle,
- attendance global-open uniqueness/shift relation,
- audited lock/handover,
- frozen promotion quote.

Partial code should not be labeled full PRD delivery just because a similarly named controller/screen exists.

### P1-9: Legacy task breakdown can mislead Codex into wrong architecture

**Evidence:** `Task & Subtask Breakdown.md` states stack assumptions including Drift and Next.js admin web. Current repo reality shows no Drift dependency and no `apps/web`/Next.js application.

**Risk:** Codex may create an unnecessary web app or offline database path, contradicting current architecture and AGENTS constraints.

**Required direction:** Mark the task breakdown as legacy/reconciled reference only. Generate new task prompts from `PRD + ARCHITECTURE + ANALYSIS + IMPROVEMENT_PLAN`, not directly from the old breakdown.

### P1-10: Database constraints and invariants are not yet strong enough as a declared contract

**Risk area:** Several critical invariants appear enforced mostly at controller level, not as a documented database/service invariant.

Examples that need explicit DB/service enforcement:

- one active shift per outlet/device,
- idempotency unique reservation,
- transaction/payment state transitions,
- immutable finalized inventory documents,
- no payment beyond remaining allowed amount,
- no close-shift overwrite,
- document number uniqueness,
- stock movement immutability,
- terminal session validity.

**Required direction:** Add migrations/constraints where safe, plus service-level row locks and tests. Do not retrofit constraints blindly on dirty data; audit/backfill first.

## P2 Findings

### P2-1: Naming and status vocabulary diverge

Examples include `savedOrders` vs held transactions, stock movement `void` vs PRD `void_reversal`, literal `cash` defaults, and roles such as `supervisor` appearing without a settled role model. Standardize PHP enums/value objects and Dart wire enums.

### P2-2: Flutter API client lacks a full operational policy

`ApiClient` centralizes envelope parsing, but timeout, request ID, safe logging/redaction, token-expiry handling, and network observability need a consistent policy after P0 idempotency semantics are settled.

### P2-3: Tests are broad but weak for concurrency, tampering, and same-tenant policy edges

Backend tests pass in the snapshot, but missing/weak cases include:

- same-tenant actor impersonation,
- device/outlet tuple spoofing,
- manipulated price/rounding,
- idempotency reservation race,
- shift close race,
- payment overpay/double-confirm,
- audit rollback,
- timezone date boundaries.

Flutter has repository/outbox coverage, but full analyze/test previously timed out during assessment; do not claim Flutter correctness from that timeout.

### P2-4: Documentation can drift unless tied to gates

The docs are now useful, but they must be updated whenever routes, migrations, contracts, or app structure change. Add a PR checklist that requires doc updates for contract changes.

## Cross-Check Summary Against PRD v2

| PRD area | Matches now | Diverges now | Missing or incomplete now |
| --- | --- | --- | --- |
| §2.1 Online-first retry | Checkout uses idempotency/outbox concepts | Persistent multi-item outbox | Single bounded in-flight buffer and recovery-by-key |
| §4 Roles/authorization | Roles and tenant fields exist | Actor/outlet/device policy not comprehensive | Policy matrix and terminal session |
| §5.4 POS checkout | Catalog/cart/quote/payment/receipt concepts exist | Server trusts client price/rounding; local held order | Frozen quote, lease/revision, async provider boundary |
| §5.6 Shift | Open/current/close endpoints exist | No robust PIN close, row lock, reason, idempotent cash movement | Handover, pending_close, immutable close report |
| §5.8 Device | Device created during login | No full enrollment/revocation/assignment guard | Device admin and terminal write guard |
| §5.9 Connectivity | Some retry/outbox UI exists | Genuine heartbeat not established as source of truth | online/degraded/offline rules |
| §6 Settings | Some outlet/payment config exists | No unified settings/security model | Full settings, audit, outlet overrides |
| §7 Reports | Sales summary/void partial | N+1 and limited definitions | Refund, top-10, product reports, shift reports |
| §8 Inventory | On-hand/purchase partial | No full document lifecycle | Count, waste, transfer, invariants |
| §9 Attendance | PIN attendance partial | Global one-open-record rule not proven | Shift relation/lockout/policy |
| §10 Screen lock | Lock route/PIN concept exists | Full audited lock/handover not proven | Idle timeout, unlock event |
| §12-13 Contracts/data | Envelope, UUIDs, tenant fields, audit/idempotency tables exist | Inconsistent policies/transactions/audit | Standard contract enforcement |

## Release Gate Recommendation

No production pilot with real money should begin until this gate passes:

1. P0-1 to P0-6 are fixed.
2. Backend feature tests cover tampering, idempotency, policy, payment, shift, audit rollback.
3. Concurrency tests exist for idempotency, document numbering, shift close, and payment confirm.
4. Flutter checkout flow uses server quote and one bounded retry buffer.
5. `php artisan test` passes.
6. `flutter analyze` passes.
7. `flutter test` passes or failing tests are explicitly triaged with owners.
8. Architecture and improvement docs are updated with any new contract.

## What Not To Do Next

Do not prioritize these until P0 is green:

- admin web/Next.js implementation,
- promotion/loyalty activation,
- fancy reports/dashboard visuals,
- large Flutter UI rewrite,
- invoice/receivable domain,
- inventory transfer complexity,
- offline sync/general local database.

## Final Review Verdict

The project is promising and already has a meaningful baseline, but the current operational core is **not safe enough for production**. The next engineering focus must be backend correctness, terminal authorization, atomic idempotency, transactional audit, payment state, and shift integrity. UI polish and admin web should wait.
