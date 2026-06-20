# NojPOS Improvement Plan

Status: proposed incremental recovery and hardening plan  
Snapshot date: 2026-06-19  
Branch observed: `feat/flutter-ui-consistency`  
Commit observed: `de07cbb`  
Worktree state: dirty; execute as a plan, not as proof of current readiness.

Primary sources:

1. `NOJPOS_POS_FLUTTER_BACKEND_PRD.md` v2.0
2. `docs/ARCHITECTURE.md`
3. `docs/ANALYSIS_FINDINGS.md`
4. `docs/IMPROVEMENT_PLAN.md` execution waves
5. `docs/DESIGN.md`
6. `docs/design-tokens.json`

Legacy reference:

- `Task & Subtask Breakdown.md` may be used only after reconciliation. It contains old/future assumptions such as Drift and Next.js admin web, while the current repo has no Drift and no `apps/web`.

## 1. Executive Direction

Do **not** continue broad feature development yet. Stabilize the operational contract first.

The next work should be:

1. Prevent Codex from following outdated architecture assumptions.
2. Harden backend idempotency, audit, terminal actor context, and policies.
3. Make checkout server-priced and quote-bound.
4. Harden payment and shift lifecycle.
5. Migrate Flutter to the safe contracts.
6. Only then continue reports, inventory lifecycle, settings, promotions, or admin web.

Admin web should remain paused until P0 backend contracts are green and Flutter consumes those contracts correctly.

## 2. Execution Principles

### 2.1 Do not do a broad rewrite

Every change must be independently reviewable. The project has already accumulated large files and cross-domain state. A big rewrite would likely create more bugs than it fixes.

Preferred change shape:

```text
one risk -> one service/contract -> one endpoint group -> focused tests -> small Flutter migration if needed
```

### 2.2 Backend contract before UI

UI should render trusted server output. It should not be the source of truth for:

- price,
- discounts,
- tax,
- service charge,
- rounding,
- final grand total,
- cashier authority,
- device authority,
- shift authority,
- payment state,
- stock movement,
- audit event.

### 2.3 Idempotency before retry UX

Do not design retry UX first. Make backend idempotency correct first. Retry UI is only safe after the server can prove that replaying the same key produces exactly one business effect.

### 2.4 Terminal actor before shift/payment hardening

Do not keep accepting free `cashier_id` as authority. Sensitive terminal writes must derive or verify actor context from server-side PIN/terminal session.

### 2.5 Audit inside the transaction

Audit should not be a best-effort log after commit. For mandatory audited actions, the business mutation and audit event must succeed or roll back together.

## 3. Stop/Go Gates

### Stop building new feature surfaces when any of these are true

- Checkout still trusts client `unit_price`, totals, or rounding.
- Terminal writes can submit arbitrary same-business cashier/device/shift IDs.
- Idempotency is lookup-execute-insert rather than atomic reserve-execute-complete.
- Shift close lacks fresh PIN, variance reason, authorizer logic, row lock, or idempotency.
- Payment state can be confirmed without row lock/state-machine validation.
- Mandatory audit can be missing after a committed mutation.
- Flutter still depends on persistent multi-item checkout outbox as the main correctness mechanism.

### Resume feature work only when

- P0 acceptance gate passes.
- Backend and Flutter contracts agree.
- Regression tests pass or failures are explicitly triaged.
- `docs/ARCHITECTURE.md` and `docs/ANALYSIS_FINDINGS.md` are updated after contract changes.

## 4. Recommended Delivery Sequence

### Wave 0 — Documentation and execution guard

Goal: prevent wrong-scope work before coding continues.

Tasks:

1. Mark `Task & Subtask Breakdown.md` as legacy/reconciled reference or regenerate it from current docs.
2. Ensure root/project instructions say:
   - no Next.js/admin web until approved,
   - no Drift/general offline DB,
   - no broad UI rewrite during P0 hardening,
   - docs source hierarchy must be followed.
3. Add or maintain `docs/EXECUTION_CHECKLIST.md`:
   - changed money/shift/payment/stock/auth/device/audit contract?
   - covered idempotency, authorization, transaction, row lock, audit?
   - added same-tenant spoofing, manipulated price/rounding, and concurrency/idempotency tests where relevant?
   - avoided admin web and Drift/offline DB unless explicitly approved?
   - documented tests run and tests not run with reasons?
4. Capture current architecture snapshot with branch/commit/worktree state.

Acceptance:

- Codex prompt source-of-truth is unambiguous.
- Legacy task assumptions cannot accidentally drive implementation.
- No app behavior changed.

### Wave 1 — Atomic idempotency and audit foundation

Goal: build shared infrastructure before hardening individual flows.

Tasks:

1. Create `IdempotencyService`.
2. Change idempotency flow from read-execute-insert to atomic reserve-execute-complete.
3. Track statuses such as `in_progress` and `completed`.
4. Return stored body/status for same key + same hash.
5. Return `409 idempotency_mismatch` for same key + different hash.
6. Apply to all sensitive writes already present:
   - checkout/transaction store,
   - payment store/confirm,
   - void,
   - shift open,
   - shift cash movement,
   - shift close,
   - inventory finalization where present.
7. Create `AuditLogService` with request ID, idempotency key, actor, approver, terminal/device, before, after, event version.
8. Make mandatory audit writes happen inside the same DB transaction as the business effect.

Acceptance:

- Parallel same-key test creates one business effect.
- Same key/different body returns 409.
- Audit failure rolls back mandatory audited mutation.
- One idempotent business effect creates one audit event.
- Existing successful response envelope remains compatible.

### Wave 2 — Terminal actor, device, outlet, and shift authority

Goal: stop same-tenant spoofing.

Tasks:

1. Introduce or harden terminal session concept after PIN verification.
2. Bind terminal session to:
   - authenticated account,
   - acting cashier,
   - business,
   - outlet,
   - enrolled device,
   - permission context,
   - expiry/last activity.
3. Add terminal-bound middleware/guard.
4. Validate device enrollment and outlet assignment.
5. Replace free `cashier_id` authority on sensitive writes with terminal session actor.
6. Verify full tuple for shift-bound actions:
   - shift belongs to business,
   - shift belongs to outlet,
   - shift belongs to device,
   - shift is open/pending state as required,
   - actor is authorized.

Acceptance:

- Cashier A cannot operate as cashier B in same business.
- Device A cannot operate Outlet B unless assigned.
- A valid same-business shift cannot be combined with wrong outlet/device/cashier.
- Revoked/unknown device returns `403 device_not_enrolled`.
- PIN errors remain generic and PIN is never logged.

### Wave 3 — Server-owned quote and checkout contract

Goal: remove financial authority from Flutter/client requests.

Tasks:

1. Create canonical quote input:
   - outlet/order context,
   - product IDs,
   - quantities,
   - customer context,
   - order type,
   - discount/promo requests only if allowed.
2. Server loads:
   - product price,
   - product availability,
   - outlet tax/service/rounding config,
   - discount/promo policy,
   - stock policy.
3. Return final quote snapshot:
   - `quote_id`,
   - `expires_at`,
   - `config_version` or equivalent,
   - frozen totals,
   - line allocations,
   - checkout idempotency key or server-bound command token.
4. Checkout accepts quote reference and creates transaction/payment/stock/audit atomically.
5. Old client fields may remain in response for compatibility, but not as authority.
6. Flutter payment screen renders server quote only.

Acceptance:

- Manipulated `unit_price` cannot change final charge.
- Manipulated `rounding` cannot change final charge.
- Same final quote + key creates one transaction.
- Stale quote returns `409 quote_stale`.
- Checkout rollback test proves no partial transaction/payment/stock/audit.

### Wave 4 — Payment state-machine hardening

Goal: make payment mutation safe and auditable.

Tasks:

1. Create `PaymentService`.
2. Define allowed transaction/payment states.
3. Lock transaction row before adding/confirming payment.
4. Validate remaining payable amount.
5. Separate synchronous cash from async QRIS/EDC/transfer behavior.
6. For async methods, create `payment_pending` and settle only by provider webhook or approved confirmation boundary.
7. Audit payment create/confirm/fail/expire where required.

Acceptance:

- Cannot pay a voided/closed/incompatible transaction.
- Cannot overpay beyond modeled behavior.
- Parallel confirmation settles once.
- Webhook + polling race cannot double-apply.
- Payment state and transaction status roll back together on failure.

Implementation note:

- Until a separate cash tender/change-due contract exists, `payments.amount` is the collected allocation amount and must not exceed the remaining payable total.
- Cash methods settle synchronously as `confirmed`; non-cash methods start as `pending` and are confirmed only through the payment confirmation boundary.

### Wave 5 — Shift and cash drawer hardening

Goal: make cashier money reconciliation trustworthy.

Tasks:

1. Create `ShiftService`.
2. Harden open shift:
   - terminal actor-derived cashier,
   - enrolled device,
   - one active shift per outlet/device,
   - idempotency,
   - audit.
3. Harden cash movement:
   - mandatory reason,
   - positive amount,
   - role/approver rules,
   - idempotency,
   - audit,
   - immutable row.
4. Harden close shift:
   - fresh server summary,
   - fresh PIN,
   - actual cash blank by default in UI,
   - row lock,
   - non-zero variance reason,
   - threshold approver,
   - idempotent close,
   - immutable close report snapshot.
5. Route Flutter to PIN/shift gate after close, never back to active POS.

Acceptance:

- Duplicate open blocked/idempotent.
- Cash movement retry creates one movement.
- Missing reason returns 422.
- Double close cannot overwrite actual cash.
- Variance reason/approver rules enforced.
- Close report remains immutable.

Implementation note:

- Cash movements require a non-empty `reason` and are idempotent via `Idempotency-Key`.
- Shift close requires a fresh actor PIN and stores an immutable `close_report_snapshot`; second close attempts return `SHIFT_ALREADY_CLOSED` without overwriting actual cash, variance, reason, or report.
- PRD v2.0 leaves the rupiah variance threshold as an open business decision. Until outlet policy config exists, every non-zero variance requires `variance_reason`; optional `approver_id` is validated as owner/admin/supervisor when supplied.

### Wave 6 — Checkout retry buffer correction

Goal: align Flutter retry behavior with PRD v2.

Dependency: Wave 1 and Wave 3 must be safe first.

Tasks:

1. Replace persistent multi-item outbox with one bounded in-flight checkout retry record.
2. Retry only the same checkout idempotency key.
3. Block new sale while an in-flight checkout is unresolved.
4. After app restart/process kill, query backend by idempotency key before allowing new sale.
5. Remove queue-like UX language for checkout drafts.

Acceptance:

- Only one unresolved checkout can exist.
- Restart recovery does not blindly replay stale full draft.
- New sale is blocked until prior checkout is resolved.
- Same key returns stored result.
- Different body/key behavior matches backend contract.

### Wave 7 — Outlet timezone and business clock

Goal: make reports, shifts, attendance, and numbering Indonesia-correct.

Tasks:

1. Add `outlets.timezone` default `Asia/Jakarta`.
2. Create `BusinessClock`/equivalent.
3. Use outlet-local boundaries for:
   - shift auto-expiry,
   - shift reports,
   - daily reports,
   - attendance day rules,
   - document numbering month/year,
   - store open/close if enabled later.
4. Store timestamps in UTC; render using outlet timezone.

Acceptance:

- Date boundary tests pass for `Asia/Jakarta`.
- UTC storage remains consistent.
- Reports and document numbers use outlet-local day/month.

### Wave 8 — Incremental Flutter architecture extraction

Goal: reduce regression risk after behavior is stable.

Tasks:

1. Extract one feature at a time from `NojposSessionNotifier`.
2. Start with lower-risk domains or those already contract-stable.
3. Keep existing routes stable while moving state.
4. Add provider/widget tests before each extraction.
5. Split giant widgets only around existing visible behavior.

Suggested order:

1. Attendance provider/screen state.
2. Reports provider/screen state.
3. Inventory provider/screen state.
4. Shift provider/cash drawer/close flow.
5. POS quote/cart/payment state.

Acceptance:

- No feature extraction changes server contract.
- Tests pass after each extraction.
- `NojposSessionNotifier` line count decreases gradually.
- UI behavior stays equivalent unless explicitly changed.

### Wave 9 — Resume product domains

Only after P0 and core P1 gates are green:

1. Parked orders with revision/lease.
2. Settings aggregate/security settings.
3. Complete reports definitions.
4. Inventory count/waste/transfer lifecycle.
5. Promotions quote contract.
6. Refund.
7. Store open/close.
8. Admin web architecture decision and implementation.

## 5. P0 Work Item Matrix

| ID | Problem | Primary fix | Must include tests | Contract impact |
| --- | --- | --- | --- | --- |
| P0-1 | Client price/rounding authority | Server quote snapshot | Tampered price/rounding, stale quote | Backend + Flutter checkout migration |
| P0-2 | Actor/device/outlet/shift spoofing | Terminal session + tuple validation | Same-tenant spoof, revoked device | Auth/shift/checkout/attendance/void |
| P0-3 | Payment mutation unsafe | PaymentService + row locks | Overpay, double confirm, bad state | Payment API behavior |
| P0-4 | Idempotency race/incomplete coverage | Atomic reservation service | Parallel same key, mismatch | Shared middleware/service |
| P0-5 | Shift/cash weak | ShiftService + PIN/variance/audit | Double close, reason, approver | Shift API + Flutter close flow |
| P0-6 | Audit inconsistent | Transactional AuditLogService | Audit rollback, exactly-one audit | Shared service and schema |

## 6. Backend Standards To Adopt

### 6.1 Controller shape

```text
validate -> resolve principal/terminal context -> authorize -> call service -> return resource/envelope
```

Controllers should not own financial calculation or multi-table state transitions.

### 6.2 Service shape

```text
service method
  -> DB::transaction
  -> lock mutable aggregate rows
  -> validate current state
  -> mutate domain rows
  -> write audit event
  -> return response DTO
```

### 6.3 Idempotency shape

```text
reserve key atomically
  if completed same hash -> replay stored response
  if same key different hash -> 409 idempotency_mismatch
  if in progress -> safe retry/409/202 according to endpoint policy
  execute domain effect once
  store response snapshot
```

### 6.4 Audit shape

Event naming should be stable and boring:

```text
auth.login
terminal.pin_verified
shift.open
shift.cash_movement.create
shift.close
checkout.quote.finalized
transaction.create
payment.create
payment.confirm
void.create
refund.create
inventory.purchase.finalize
inventory.count.finalize
inventory.waste.finalize
inventory.transfer.dispatch
inventory.transfer.receive
settings.update
device.revoke
terminal.unlock
```

Each event should carry:

- business,
- outlet if applicable,
- device if applicable,
- actor,
- approver if applicable,
- entity type/id,
- before,
- after,
- request ID,
- idempotency key,
- event version,
- timestamp.

### 6.5 Money shape

- Integer rupiah everywhere.
- Server owns product price and calculation order.
- Client displays server values only.
- Rounding is server config, not arbitrary request input.
- Historical transaction rows keep historical labels/values.

### 6.6 Time shape

- Store timestamps UTC.
- Use outlet timezone for business-day logic.
- Default timezone: `Asia/Jakarta`.
- Tests must cover boundary around midnight outlet time.

## 7. Flutter Standards To Adopt

1. Widget does not serialize authoritative business payloads directly.
2. Feature provider owns one feature state, not all app state.
3. Repository maps API DTOs and typed exceptions.
4. UI renders server-calculated totals.
5. Client creates/retains idempotency key only for server-declared idempotent commands.
6. Error handling maps backend codes to Indonesian user copy at feature boundary.
7. Sensitive PIN values are cleared immediately on success/failure/background/dispose.
8. Checkout retry follows backend idempotency contract, not arbitrary local queue behavior.

## 8. Test Strategy

### 8.1 Backend test categories

| Category | Required examples |
| --- | --- |
| Tenant isolation | Business A cannot access/mutate Business B data |
| Same-tenant actor integrity | User cannot impersonate another cashier/device/shift in same business |
| Idempotency | replay, mismatch, parallel same key, failed effect behavior |
| Financial tampering | client price, rounding, discount, grand total manipulation |
| State machine | payment, transaction, shift, inventory document allowed transitions |
| Row lock/concurrency | double close, double payment confirm, document numbering, stock finalize |
| Audit | event exists, exactly once, rollback with mutation, before/after snapshots |
| Timezone | reports/shift/attendance/numbering boundary in outlet timezone |

### 8.2 Flutter test categories

| Category | Required examples |
| --- | --- |
| Provider state | loading/success/validation/forbidden/conflict/recoverable network |
| DTO mapping | server quote, transaction, payment, shift close report |
| Widget flow | PIN, shift gate, cart, payment, close shift, lock screen |
| Retry behavior | one in-flight checkout, recovery by key, block new sale |
| Error copy | `quote_stale`, `idempotency_mismatch`, `device_not_enrolled`, `order_locked` |

## 9. Migration Strategy

### 9.1 General migration rules

- Additive schema changes first.
- Backfill dirty data before adding strict constraints.
- Keep old response fields temporarily if Flutter needs staged rollout.
- Never silently reinterpret old client price/totals as valid authoritative input.
- Fail old unsafe contract clearly once migration window closes.

### 9.2 Checkout contract migration

Safe sequence:

1. Add new quote endpoint/contract.
2. Add backend tests for tampering and stale quote.
3. Update Flutter to consume new quote.
4. Make checkout use quote reference.
5. Keep old fields read-only or ignored.
6. Reject old authoritative client totals after client migration.

### 9.3 Idempotency migration

Safe sequence:

1. Add idempotency status/reservation columns if needed.
2. Preserve ability to replay old completed keys if current rows exist.
3. Introduce service behind existing middleware name if possible.
4. Add coverage to missing endpoints.
5. Add concurrency tests.

### 9.4 Shift migration

Safe sequence:

1. Add terminal session/device/assignment fields if missing.
2. Backfill synthetic terminal context for existing active/demo data.
3. Update open/current/close endpoints.
4. Update Flutter shift repository.
5. Add close report snapshot.
6. Enforce stricter constraints.

## 10. Codex Execution Template

Use this style per wave. Do not paste the whole plan as one coding task.

```md
You are working on NojPOS. Follow PRD v2.0, docs/ARCHITECTURE.md, docs/ANALYSIS_FINDINGS.md, and docs/IMPROVEMENT_PLAN.md.

Task: <one wave or subtask only>

Hard constraints:
- Do not add apps/web or Drift.
- Do not change unrelated UI.
- Keep changes small and reviewable.
- Add/extend tests for the specific risk.
- Preserve API envelope.
- Sensitive writes must be tenant-scoped, authorized, idempotent, transactional, and audited where required.

Before coding:
1. Inspect current files and tests.
2. Identify existing helpers to reuse.
3. State the exact contract you will change.

After coding:
1. Run targeted tests.
2. Summarize changed files.
3. Mention remaining risks.
```

## 11. What To Avoid

Avoid these until P0 is green:

- building `apps/web`,
- adding Drift/general offline DB,
- implementing promotions UI,
- implementing broad admin dashboards,
- large POS UI rewrite,
- adding more queue/outbox domains,
- starting inventory transfer complexity,
- adding invoice/receivable domain,
- changing design system while backend contract is unsafe.

## 12. Final Recommendation

The best next move is **not** to build more visible features. The best next move is to turn the current baseline into a safe operational core.

Recommended immediate order:

1. Wave 0 documentation guard.
2. Wave 1 idempotency + audit foundation.
3. Wave 2 terminal actor/device authority.
4. Wave 3 server-owned quote/checkout.
5. Wave 4 payment state machine.
6. Wave 5 shift/cash drawer hardening.
7. Wave 6 bounded checkout retry.
8. Wave 7 outlet timezone.
9. Wave 8 Flutter extraction.
10. Resume product expansion/admin web only after gates pass.

This keeps the project from becoming a polished but financially unsafe POS.
