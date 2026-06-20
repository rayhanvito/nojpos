# NojPOS Execution Checklist

Use this checklist before reporting Codex completion or opening a PR. It is intentionally strict because the current repo is not production-safe for real money flow until P0 backend contracts are green.

## Scope guard

- [ ] Did this task stay inside the requested wave/subtask?
- [ ] Were app code changes kept scoped and reviewable?
- [ ] Was broad formatting-only churn avoided?
- [ ] Was admin web avoided unless explicitly approved?
- [ ] Was `apps/web`, Next.js, or any admin web package avoided unless explicitly approved?
- [ ] Was Drift, SQLite, Hive, generic sync, or any general offline database avoided unless explicitly approved?
- [ ] Were package/dependency changes avoided unless the task explicitly required them?

## Sensitive-write trigger

Mark this task sensitive if it touched any of these:

- [ ] Money / totals / price / discount / tax / service / rounding
- [ ] Checkout / transaction / payment / void / refund
- [ ] Shift open / shift close / cash movement / store open-close
- [ ] Stock / inventory movement / document finalization
- [ ] Auth / PIN / role / permission / policy
- [ ] Device / terminal session / outlet assignment
- [ ] Audit / idempotency / request replay / retry behavior

If any item above is checked, the hardening checklist below is mandatory.

## Sensitive-write hardening checklist

- [ ] Tenant scope comes from the authenticated principal, not client-supplied tenant data.
- [ ] Same-tenant actor spoofing is prevented; `cashier_id`, `device_id`, `shift_id`, and `outlet_id` from the client are not treated as proof of authority.
- [ ] Server-side authorization/policy/action guard is present.
- [ ] Idempotency is atomic: reserve before effect, replay same key/body, reject same key/different body.
- [ ] State changes happen inside a database transaction.
- [ ] Mutable aggregate rows are locked or conditionally updated where concurrency can race.
- [ ] Mandatory audit is written inside the same transaction as the business effect.
- [ ] Server is authoritative for price, totals, tax, service, rounding, stock, and report calculations.
- [ ] Client-rendered totals are treated as display only.
- [ ] PIN, tokens, PII, and payment references are not logged, audited raw, or exposed in errors/fixtures.

## Required tests for sensitive work

- [ ] Tenant isolation tests are included.
- [ ] Same-tenant spoofing tests are included for actor/device/outlet/shift combinations.
- [ ] Authorization/role/outlet policy tests are included.
- [ ] Manipulated price, manipulated totals, and manipulated rounding tests are included for checkout/quote work.
- [ ] Idempotent replay and same-key/different-body conflict tests are included.
- [ ] Concurrency/idempotency tests are included for sensitive writes.
- [ ] Transaction rollback tests prove no partial money/stock/payment/audit effect.
- [ ] Audit tests prove exactly one audit event per idempotent business effect.
- [ ] Payment/shift/stock state-machine conflict tests are included where relevant.
- [ ] Timezone boundary tests are included when business-day, shift, report, or document numbering logic changes.

## Documentation and contract checks

- [ ] API envelope stayed `{ "data", "meta" }` for success and `{ "error": { "code", "message", "details" } }` for failure.
- [ ] Standard error codes were preserved or documented.
- [ ] Docs were updated if routes, DTOs, migrations, idempotency, audit, terminal session, or app structure changed.
- [ ] `docs/ARCHITECTURE.md` was updated if repo structure or current reality changed.
- [ ] `docs/ANALYSIS_FINDINGS.md` was updated if a P0/P1 finding was fixed, downgraded, or newly discovered.
- [ ] `docs/IMPROVEMENT_PLAN.md` was updated if execution sequence or gates changed.
- [ ] `docs/DESIGN.md` and `docs/design-tokens.json` were updated if UI/design contracts changed.

## Validation report

Fill this in the Codex completion message:

- Files changed:
- App code touched? yes/no. If yes, list paths and justify scope:
- Sensitive-write checklist required? yes/no:
- Tests run:
- Tests not run and why:
- Remaining blockers/risks:
- Next recommended wave/subtask:
