# F1.1 Backend Promo Engine Implementation Plan

> **For agentic workers:** Use `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Add store-scoped promo management, read-only promo preview, and checkout promo application with immutable transaction snapshots.

**Architecture:** Add the promo model and CRUD through the existing model/repo/service/handler layers. Keep checkout pricing in the transaction repository, which already owns the atomic checkout calculation; share its promo resolver with the read-only preview route. Register admin management routes separately from authenticated preview.

**Tech Stack:** Go, Gin, GORM AutoMigrate, SQLite router integration tests.

**Spec:** `PLAN.md`, Unit F1.1, plus the backend-specific requirements in the user task.

## Global Constraints

- Only implement the backend part of F1.1.
- All API paths are under `/api/v1`; errors use `{"error":"..."}` and invalid checkout vouchers also use code `PROMO_INVALID`.
- Request JSON uses camelCase; model responses keep their existing JSON tags; lists use `{"items": [...]}`.
- Store scope comes only from authenticated claims.
- Money remains integer Rupiah; promo discounts apply before the existing additive tax calculation.
- Promo or voucher discounts replace manual discounts for that checkout; automatic promos do not stack.
- Schema changes use GORM AutoMigrate; no SQL migration files and no commit.
- Keep current UI and avoid new dependencies.

## Review Focus

- A category or product ID from another store must be rejected, and a store must not list or apply another store's promo; test in Task 1 and Task 2.
- A malformed or incomplete happy-hour window must be rejected; test in Task 1.
- Percentage rounding at a half-Rupiah and a `maxDiscount` cap must produce integer Rupiah; test in Task 3.
- Store-local happy-hour checks must use `Store.Timezone`; test in Task 2.
- Deleting a voucher must allow its code to be reused without changing transaction snapshots; test in Task 1 and Task 3.

---

### Task 1: Promo model and admin CRUD

**Files:**
- Create: `model/promo.go`
- Modify: `db/db.go`
- Create: `repo/promo.go`
- Modify: `service/interfaces.go`
- Create: `service/promo.go`
- Create: `handler/promo.go`
- Modify: `router/router.go`
- Create: `router/promo_integration_test.go`

**Interfaces:**
- Produces `model.Promo` with store ID, name, type, integer value, scope and optional scope ID, nullable voucher code, start/end timestamps, optional happy-hour strings, minimum purchase, maximum discount, and active flag.
- Produces `service.PromoRepository` methods for list, get, create, update, active toggle, and delete, always scoped by store ID.
- Admin routes are `GET/POST /promos`, `GET/PUT/DELETE /promos/:id`, and `PATCH /promos/:id/active`.

- [ ] **Step 1: Write the failing router integration test** `TestPromoCRUDIsAdminOnlyAndStoreScoped`. Verify create/list/get/update/toggle/delete, duplicate voucher code per store, foreign scope IDs rejected, and cashier management routes return 403.
- [ ] **Step 2: Run the test and confirm it fails because promo routes/model are absent.** Run `CGO_ENABLED=1 go test ./router -run TestPromoCRUDIsAdminOnlyAndStoreScoped -count=1`.
- [ ] **Step 3: Add the promo model and AutoMigrate registration.** Use a nullable code with a composite unique index on store ID and voucher code; use integer Rupiah fields and JSON response tags consistent with existing models.
- [ ] **Step 4: Add the scoped repository, service validation, handler, and admin routes.** Validate type/scope/required IDs, nonnegative values, start before end, and both-or-neither `happyStart`/`happyEnd` in `HH:MM`; validate category/product ownership through existing repositories. Create active promos by default. Deleting a promo hard-deletes its row so a voucher code can be reused; transaction snapshots remain independent.
- [ ] **Step 5: Run the focused integration test and confirm it passes.** Run `CGO_ENABLED=1 go test ./router -run TestPromoCRUDIsAdminOnlyAndStoreScoped -count=1`.

### Task 2: Shared promo resolution and read-only preview

**Files:**
- Modify: `repo/transaction.go`
- Modify: `repo/promo.go`
- Modify: `service/interfaces.go`
- Modify: `service/transaction.go`
- Modify: `handler/transaction.go`
- Modify: `router/router.go`
- Modify: `router/promo_integration_test.go`

**Interfaces:**
- Produces `repo.TrxRepo.PreviewPromo(ctx, input) (*model.PromoPreview, error)` and `service.TrxService.PreviewPromo(...)`.
- Preview request is `{items:[{productId,qty}], voucherCode?}` and response is a promo object with `promo_id`, `promo_name`, and `promo_discount`, or JSON `null`.
- Promo evaluation uses current store product prices, scope totals, current time, and `Store.Timezone`; a supplied but invalid code returns `null` from preview and is rejected by checkout in Task 3.

- [ ] **Step 1: Extend the router integration test** `TestPromoPreviewUsesStoreScheduleScopeAndDoesNotWrite`. Cover cashier access, current and expired voucher windows, happy-hour inside/outside, category-only covered value, and unchanged transaction count, stock, and movements after preview.
- [ ] **Step 2: Run the focused test and confirm the preview route is absent.** Run `CGO_ENABLED=1 go test ./router -run TestPromoPreviewUsesStoreScheduleScopeAndDoesNotWrite -count=1`.
- [ ] **Step 3: Implement one shared resolver for voucher and automatic promos.** A voucher selects only its exact code; automatic promos require a null code and the highest computed discount wins, with ID order as a stable tie-break. Start is inclusive and end is exclusive. Happy-hour bounds are interpreted in the store timezone. Percent discounts use round-half-up, apply a positive `maxDiscount` cap when set, and amount discounts are capped at covered item value.
- [ ] **Step 4: Implement preview using read queries only and register `POST /promos/preview` in the authenticated route group.** Build item values from the same store-scoped products checkout uses; do not create transactions, movements, stock updates, or notifications.
- [ ] **Step 5: Run the focused preview test and confirm it passes.** Run `CGO_ENABLED=1 go test ./router -run TestPromoPreviewUsesStoreScheduleScopeAndDoesNotWrite -count=1`.

### Task 3: Checkout application and transaction snapshot

**Files:**
- Modify: `model/transaction.go`
- Modify: `repo/transaction.go`
- Modify: `service/transaction.go`
- Modify: `handler/transaction.go`
- Modify: `db/db.go`
- Modify: `router/promo_integration_test.go`

**Interfaces:**
- Checkout JSON adds optional `voucherCode`; checkout responses and transaction detail expose `promo_id`, `promo_name`, and `promo_discount`.
- Invalid, inactive, out-of-date, out-of-happy-hour, or below-minimum submitted vouchers fail checkout with HTTP 400 and code `PROMO_INVALID`.

- [ ] **Step 1: Extend router integration coverage** `TestCheckoutAppliesPromoSnapshotAndRejectsInvalidVoucher`. Verify valid and expired vouchers, incorrect code error code, automatic winner among two promos, manual discount ignored when a promo applies, category-scoped discount, half-up rounding and cap, additive tax base after promo, and persisted snapshot in checkout/detail responses.
- [ ] **Step 2: Run the focused test and confirm it fails on missing checkout behavior.** Run `CGO_ENABLED=1 go test ./router -run TestCheckoutAppliesPromoSnapshotAndRejectsInvalidVoucher -count=1`.
- [ ] **Step 3: Pass the voucher code into the existing atomic checkout transaction and call the shared resolver after subtotal calculation.** If a promo is selected, set manual `discount` to zero, subtract `promo_discount` before tax, and persist the promo ID/name/discount snapshot; otherwise retain current manual-discount behavior.
- [ ] **Step 4: Return `PROMO_INVALID` through the transaction error handler and ensure persisted transaction detail includes the snapshot.** Keep receipt data sourced from the transaction record.
- [ ] **Step 5: Run the focused checkout integration test and confirm it passes.** Run `CGO_ENABLED=1 go test ./router -run TestCheckoutAppliesPromoSnapshotAndRejectsInvalidVoucher -count=1`.

### Task 4: Final verification and review

**Files:** All files above.

- [ ] Run `gofmt` on changed Go files and verify `gofmt -l` is empty.
- [ ] Run `CGO_ENABLED=1 go test ./...` and read the complete result.
- [ ] Review the complete diff against `PLAN.md` F1.1 and the user task; fix any Critical or Important finding with a failing integration test first.
- [ ] Do not commit; report files, decisions, test status, and remaining uncertainty.
