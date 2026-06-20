# GitHub Issues Plan

Project: NOJ POS P0 Production Readiness
Date: 2026-06-20

Copy each issue block into GitHub as needed.

---

## 1. EPIC: P0 Production Readiness

**Title:** EPIC: P0 Production Readiness

**Labels:** `epic`, `p0`, `production-readiness`

**Priority:** P0

**Area:** Fullstack

**Wave:** Release

**Status:** In Progress

**Summary:**
Track all P0 work required to move NOJ POS from verified milestone implementation toward a release candidate.

**Acceptance Criteria:**
- Parked orders revision + lease verified.
- Settings aggregate + security settings verified.
- Async payment webhook/polling/expiry verified.
- Reports backend + Flutter verified.
- Inventory count/waste/transfer/in-transit verified.
- Promotion quote + UI promo verified.
- Refund completed and verified.
- Store open/close completed and verified.
- Screen lock completed and verified.
- Admin web decision made.
- Release candidate QA completed.

**Verification Commands:**
- `cd apps/backend && php artisan test --compact`
- `cd apps/backend && php artisan route:list --path=api/v1 --except-vendor`
- `cd apps/cashier && C:/flutter/bin/flutter.bat analyze --no-pub`
- `cd apps/cashier && C:/flutter/bin/flutter.bat test --no-pub`

**Definition of Done:**
- All child issues done or intentionally deferred with approval.
- Full backend and Flutter verification passes.
- No secret/cache/build artifacts committed.
- Release candidate checklist completed.

---

## 2. Parked orders revision + lease

**Title:** Parked orders revision + lease

**Labels:** `p0`, `backend`, `flutter`, `done`, `parked-orders`

**Priority:** P0

**Area:** Fullstack

**Wave:** Parked Orders

**Status:** Done

**Summary:**
Add optimistic revision and 90-second lease lifecycle for parked orders, with backend contract and Flutter lifecycle support.

**Acceptance Criteria:**
- Parked order has revision.
- Lease acquire/refresh/release endpoints exist.
- Lease expires logically after 90 seconds.
- Concurrent editor receives conflict/locked response.
- Flutter acquires before edit, heartbeats while active, and releases on finish/cancel/exit.

**Verification Commands:**
- `cd apps/backend && php artisan test --compact tests/Feature/ParkedOrderLeaseTest.php`
- `cd apps/backend && php artisan test --compact`
- `cd apps/cashier && C:/flutter/bin/flutter.bat test --no-pub`

**Definition of Done:**
- Backend and Flutter tests pass.
- Existing held order behavior remains compatible.

---

## 3. Settings aggregate + security settings

**Title:** Settings aggregate + security settings

**Labels:** `p0`, `backend`, `flutter`, `done`, `settings`, `security`

**Priority:** P0

**Area:** Fullstack

**Wave:** Settings

**Status:** Done

**Summary:**
Create tenant-scoped settings aggregate and security policy contract with safe POS read response and owner/admin writes.

**Acceptance Criteria:**
- `GET /settings` returns POS-safe aggregate.
- Owner/admin can update settings.
- Cashier cannot update settings.
- PIN hashes/secrets never appear in response.
- Security mutation audited.

**Verification Commands:**
- `cd apps/backend && php artisan test --compact tests/Feature/SettingsAggregateTest.php`
- `cd apps/backend && php artisan test --compact`
- `cd apps/cashier && C:/flutter/bin/flutter.bat test --no-pub`

**Definition of Done:**
- Tenant/role tests pass.
- Flutter settings summary remains read-only for cashier.

---

## 4. Async payment webhook/polling/expiry

**Title:** Async payment webhook/polling/expiry

**Labels:** `p0`, `backend`, `flutter`, `done`, `payment`

**Priority:** P0

**Area:** Fullstack

**Wave:** Async Payment

**Status:** Done

**Summary:**
Support asynchronous QRIS/EDC/bank transfer payment confirmation, polling, idempotent webhook settlement, and expiry.

**Acceptance Criteria:**
- Cash remains synchronous.
- Async methods create `payment_pending` with reference and expiry.
- Webhook confirms payment exactly once.
- Duplicate webhook is safe.
- Invalid webhook signature rejected.
- Expiry marks failed/expired without stock/cash effects.
- Flutter polls every 3 seconds and stops on terminal state/dispose.

**Verification Commands:**
- `cd apps/backend && php artisan test --compact tests/Feature/AsyncPaymentTest.php`
- `cd apps/backend && php artisan test --compact`
- `cd apps/cashier && C:/flutter/bin/flutter.bat test --no-pub`

**Definition of Done:**
- Settlement is transactional, idempotent, and audited.
- Flutter does not settle client-side.

---

## 5. Reports backend + Flutter

**Title:** Reports backend + Flutter

**Labels:** `p0`, `backend`, `flutter`, `done`, `reports`

**Priority:** P0

**Area:** Fullstack

**Wave:** Reports

**Status:** Done

**Summary:**
Implement backend reports contract and Flutter DTO/repository/UI rendering for P0 reports.

**Acceptance Criteria:**
- Reports include sales summary, sold products, payment methods, cashier shifts, void/refund audit, and top-10.
- Tenant/outlet/cashier scope enforced.
- Paid totals exclude pending/failed/expired/voided.
- Flutter renders server response and does not aggregate totals client-side.

**Verification Commands:**
- `cd apps/backend && php artisan test --compact tests/Feature/ReportsContractTest.php`
- `cd apps/backend && php artisan test --compact`
- `cd apps/cashier && C:/flutter/bin/flutter.bat test --no-pub`

**Definition of Done:**
- Report contract tests pass.
- Flutter parsing and empty states pass.

---

## 6. Inventory operations

**Title:** Inventory count, waste, transfer, in-transit reconciliation

**Labels:** `p0`, `backend`, `flutter`, `done`, `inventory`

**Priority:** P0

**Area:** Fullstack

**Wave:** Inventory

**Status:** Done

**Summary:**
Implement stock count, waste, transfer lifecycle, and in-transit reconciliation with immutable stock movements.

**Acceptance Criteria:**
- Inventory overview exists.
- Movement history exists.
- Count writes adjustment movement.
- Waste reduces stock with reason.
- Transfer requested/sent/received/cancelled lifecycle works.
- Source decreases on send; destination increases only on receive.
- Writes are tenant-scoped, outlet-scoped, transactional, idempotent, and audited.

**Verification Commands:**
- `cd apps/backend && php artisan test --compact tests/Feature/InventoryOperationsTest.php`
- `cd apps/backend && php artisan test --compact`
- `cd apps/cashier && C:/flutter/bin/flutter.bat test --no-pub`

**Definition of Done:**
- Backend stock mutation tests pass.
- Flutter inventory operation tests pass.

---

## 7. Promotion quote + UI promo

**Title:** Promotion quote contract + UI promo

**Labels:** `p0`, `backend`, `flutter`, `done`, `promotion`

**Priority:** P0

**Area:** Fullstack

**Wave:** Promo

**Status:** Done

**Summary:**
Make promotion quote server-authoritative with quote token/hash and minimal Flutter promo UI.

**Acceptance Criteria:**
- Quote endpoint returns token/hash and server-calculated totals.
- Active percentage/fixed promos apply correctly.
- Expired/inactive/outlet/product/min-spend rules tested.
- No-stacking policy explicit.
- Checkout rejects stale quote and does not trust client totals.
- Flutter renders applied/rejected promo and sends latest quote token/hash.

**Verification Commands:**
- `cd apps/backend && php artisan test --compact tests/Feature/PromotionQuoteTest.php`
- `cd apps/backend && php artisan test --compact`
- `cd apps/cashier && C:/flutter/bin/flutter.bat test --no-pub`

**Definition of Done:**
- Backend full suite and Flutter full suite pass.
- No client-side totals calculation added.

---

## 8. Refund

**Title:** Refund contract and POS flow

**Labels:** `p0`, `ready`, `refund`, `backend`, `flutter`

**Priority:** P0

**Area:** Fullstack

**Wave:** Refund

**Status:** Ready for Codex

**Summary:**
Implement refund mutation and POS refund flow after the checkpoint branch is pushed.

**Acceptance Criteria:**
- Backend refund contract is tenant/outlet scoped.
- Refund mutation is role-protected, transactional, idempotent, and audited.
- Refund handles payment and stock effects according to PRD.
- Refund audit report consumes refund rows.
- Flutter renders refund flow only after backend tests pass.

**Verification Commands:**
- `cd apps/backend && php artisan test --compact tests/Feature/<RefundTest>.php`
- `cd apps/backend && php artisan test --compact`
- `cd apps/cashier && C:/flutter/bin/flutter.bat analyze --no-pub`
- `cd apps/cashier && C:/flutter/bin/flutter.bat test --no-pub`

**Definition of Done:**
- Refund backend and Flutter tests pass.
- Existing payment/report/inventory flows remain green.

---

## 9. Store open/close

**Title:** Store open/close lifecycle

**Labels:** `p0`, `backlog`, `store`, `backend`, `flutter`

**Priority:** P0

**Area:** Fullstack

**Wave:** Store

**Status:** Backlog

**Summary:**
Implement store opening and closing lifecycle after Refund is completed and verified.

**Acceptance Criteria:**
- Store state is tenant/outlet scoped.
- Opening/closing is role-protected and audited.
- POS blocks or warns according to PRD when store is closed.
- Flutter renders current store state and allowed actions.

**Verification Commands:**
- `cd apps/backend && php artisan test --compact`
- `cd apps/cashier && C:/flutter/bin/flutter.bat test --no-pub`

**Definition of Done:**
- Store lifecycle tests pass without breaking shift/payment flows.

---

## 10. Screen lock

**Title:** Screen lock and terminal idle policy

**Labels:** `p0`, `backlog`, `security`, `flutter`

**Priority:** P0

**Area:** Flutter

**Wave:** Screen Lock

**Status:** Backlog

**Summary:**
Implement screen lock behavior using the security settings policy after core store/refund flows are complete.

**Acceptance Criteria:**
- Idle lock respects server-provided security setting.
- Unlock requires safe authentication/PIN flow.
- Sensitive information is hidden while locked.
- Tests cover idle lock and unlock behavior.

**Verification Commands:**
- `cd apps/cashier && C:/flutter/bin/flutter.bat analyze --no-pub`
- `cd apps/cashier && C:/flutter/bin/flutter.bat test --no-pub`

**Definition of Done:**
- Flutter lock tests pass.
- Existing POS session flow remains green.

---

## 11. Admin web decision / apps-web

**Title:** Admin web decision / apps-web

**Labels:** `p0`, `backlog`, `admin-web`, `decision`

**Priority:** P1

**Area:** Docs

**Wave:** Admin Web

**Status:** Backlog

**Summary:**
Decide whether `apps/web` should be built for the release candidate or deferred.

**Acceptance Criteria:**
- Decision documented.
- If approved, scope and backend contracts are defined before implementation.
- If deferred, release notes explain POS-only release scope.

**Verification Commands:**
- Documentation review.
- Backend/Flutter tests if implementation is approved later.

**Definition of Done:**
- Explicit approval or deferral recorded.

---

## 12. Release candidate QA

**Title:** Release candidate QA

**Labels:** `p0`, `backlog`, `qa`, `release`

**Priority:** P0

**Area:** QA

**Wave:** Release

**Status:** Backlog

**Summary:**
Run final release candidate QA after all P0 feature milestones are complete.

**Acceptance Criteria:**
- Backend full suite passes.
- Flutter analyze and full tests pass.
- Manual POS smoke test passes.
- API route list reviewed.
- No secret/cache/build artifacts committed.
- Deployment/release notes ready.

**Verification Commands:**
- `cd apps/backend && php artisan test --compact`
- `cd apps/backend && php artisan route:list --path=api/v1 --except-vendor`
- `cd apps/cashier && C:/flutter/bin/flutter.bat analyze --no-pub`
- `cd apps/cashier && C:/flutter/bin/flutter.bat test --no-pub`

**Definition of Done:**
- Release candidate is tagged or ready for PR merge after review.
