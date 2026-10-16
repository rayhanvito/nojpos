# F2.7 Backend Shift Reporting Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Record cash drawer close counts and provide outlet-aware shift summaries, detail, force-close, and CSV export.

**Architecture:** Keep the existing handler → service → repo → model layers. Compute sales and refunds from transactions owned by the shift cashier within the shift time range, persist close values atomically, and use the existing F2.6 outlet scope.

**Tech Stack:** Go, Gin, GORM, SQLite router integration tests.

**Spec:** `PLAN.md`, Unit F2.7 backend; task requirements in the user prompt.

## Global Constraints

- Work only on the F2.7 backend; do not modify frontend files or commit.
- API routes remain under `/api/v1`; errors use `{"error":"..."}`; request JSON uses camelCase.
- Cash amounts are integer Rupiah (`int64`); payment method labels come from `model.PayMethod`.
- Admin reports without `outletId` combine outlets; cashiers remain bound to their active outlet.
- AutoMigrate only; do not add manual SQL migrations.
- Go test caches use `/opt/.cache-go-build` and `/opt/.cache-go-tmp`, or repo-local disk caches if those paths are unavailable; never use `/tmp`.

## Review Focus

- Missing or negative cashier `closingCash` is rejected without closing the shift — `TestShiftCloseRequiresNonnegativeClosingCash`.
- A fully refunded cash sale remains in gross sales, its refund is shown separately, and cash expected is reduced by that refund — `TestShiftCloseStoresCashSummaryWithCashRefund`.
- Switching the active cashier by PIN does not transfer another cashier's shift transactions — `TestShiftOwnershipStaysWithOpeningCashierAfterPINSwitch`.
- Admin list, detail, and export respect optional same-store `outletId`, while an omitted filter combines outlets — `TestShiftAdminReportsFilterOutletAndPaginateDetail`.
- Force-close without cash leaves actual and variance null and records the specified default note; with cash it calculates variance — `TestForceCloseShiftByAdminAndOptionalCash`.

---

### Task 1: Cashier shift summary and counted close

**Files:** `model/shift.go`, `repo/shift.go`, `service/interfaces.go`, `service/shift.go`, `handler/shift.go`, `router/shift_integration_test.go`.

**Interfaces:**
- `repo.ShiftStats` carries `Sales`, `TrxCount`, `SalesByMethod`, `RefundTotal`, `CashRefundTotal`, `Items`, and `Transactions`.
- `ShiftRepository.GetShiftStats(ctx, shift, endTime)`, `CloseShift(ctx, storeID, outletID, cashierID, closingCash, note, closedByName)`, `ForceCloseShift(ctx, storeID, shiftID, closingCash, note, closedByName)`, `GetShiftByID(ctx, storeID, shiftID, outletID)`, and `ListShiftTransactions(ctx, storeID, shift, page, limit)` are the service-facing operations; outlet filters are `*uint`.
- `POST /cashier/shift/close` consumes `{ "closingCash": int64, "note"?: string }`.
- Current shift and close summary retain existing fields and add `sales_by_method`, `refund_total`, `expected_cash`, `closing_cash_actual`, `cash_variance`, `close_note`, and `closed_by_name` where applicable.

- [ ] Add router integration tests for all payment methods, cash refund arithmetic, close response persistence, missing/negative cash, and PIN-switch attribution.
- [ ] Run those tests and confirm they fail for the missing summary/validation behavior.
- [ ] Add nullable close fields to `CashierShift`; rely on existing `db.Migrate` AutoMigrate.
- [ ] Compute gross completed/refunded sales and refunds only for the shift cashier and inclusive shift time range; calculate expected cash from opening cash plus Cash sales minus Cash refunds.
- [ ] Validate required nonnegative `closingCash`, persist the close summary, and return it from both current-shift review and close.
- [ ] Run the focused shift tests and inspect the stored row and HTTP responses.

### Task 2: Admin shift list and paginated detail

**Files:** `repo/shift.go`, `service/interfaces.go`, `service/shift.go`, `handler/shift.go`, `router/router.go`, `router/shift_integration_test.go`.

**Interfaces:**
- Admin `GET /shifts?outletId=` returns enriched shift rows, retaining the existing `shifts` key and including the standard `items` list key.
- Admin `GET /shifts/:id?page=&limit=&outletId=` returns one summary and a paginated `transactions` object (`items`, `page`, `limit`, `total`).
- An omitted `outletId` means all store outlets; a supplied outlet must belong to the token store.
- `ShiftRepository.ListStoreShifts(ctx, storeID, outletID)` accepts an optional outlet filter and `GetShiftByID` applies the same filter.

- [ ] Add integration assertions for aggregate/single-outlet listing, cross-store isolation, open-shift null close fields, and detail pagination.
- [ ] Run the tests and confirm they fail on missing routes, totals, or outlet filtering.
- [ ] Add store-scoped shift lookup/list and transaction pagination using the shift's cashier and inclusive time bounds.
- [ ] Wire the outlet filter with the existing F2.6 `reportOutletFilter` pattern and register the detail route before `:id` ambiguity can arise.
- [ ] Re-run the focused list/detail tests.

### Task 3: Admin force-close

**Files:** `repo/shift.go`, `service/interfaces.go`, `service/shift.go`, `handler/shift.go`, `router/router.go`, `router/shift_integration_test.go`.

**Interfaces:**
- Admin-only `POST /shifts/:id/force-close` accepts optional `{ "closingCash"?: int64, "note"?: string }`.
- Missing cash stores expected cash, with actual and variance null; default note is `Ditutup paksa oleh admin`. A supplied note is used as-is.

- [ ] Add tests for force-close with and without cash, duplicate close rejection, negative cash rejection, and cashier 403 on the admin route.
- [ ] Run the tests and observe the missing route/behavior.
- [ ] Close only an open shift from the same store, persist the same summary calculation, and register the admin-only route.
- [ ] Re-run the force-close tests.

### Task 4: Shift CSV export

**Files:** `service/shift.go`, `handler/shift.go`, `router/router.go`, `router/shift_integration_test.go`.

**Interfaces:**
- Admin-only `GET /shifts/export?outletId=` emits CSV with one sales column for each currently supported `model.PayMethod`, plus refund, totals, count, expected, actual, and variance.
- CSV timestamps use the store timezone and follow existing F0.2 download conventions.

- [ ] Add CSV integration assertions for header order, a populated shift row, and outlet filtering.
- [ ] Run the test and confirm the route/export is missing.
- [ ] Reuse `writeCSV`, `safeCSVText`, and `exportLocation`; register the admin route.
- [ ] Re-run the CSV integration test.

### Task 5: Full verification and review

**Files:** all F2.7 backend files above.

- [ ] Run `gofmt` on changed Go files and `git diff --check`.
- [ ] Run `CGO_ENABLED=1 go test ./...` with Go caches on disk and read the full output.
- [ ] Review F2.7 requirements line by line, especially refund timing, cashier ownership, store/outlet scope, close nullability, and CSV values.
- [ ] Do not commit; report changed files, API shapes, verification, and any remaining uncertainty.
