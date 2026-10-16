# F2.5 Backend Offline Checkout Idempotency Implementation Plan

> **For agentic workers:** Execute this plan inline with TDD. The PM requested no commit.

**Goal:** Extend `POST /transactions` with per-store offline checkout idempotency, shortage acceptance, and final stock data.

**Architecture:** Keep the existing handler → service → repository checkout path. Store non-empty client transaction IDs under a partial unique composite index, and make shortage stock handling part of the existing checkout transaction without changing checkout pricing.

**Tech Stack:** Go, Gin, GORM AutoMigrate, SQLite router integration tests.

**Spec:** `PLAN.md`, Unit F2.5; exact API and response contract from the task prompt.

## Global Constraints

- Body JSON fields use camelCase; model response tags remain snake_case.
- All transaction amounts remain integer Rupiah and the current checkout calculations stay unchanged.
- Store scope comes only from JWT claims.
- Use GORM AutoMigrate; do not add SQL migration files or dependencies.
- Do not change reports/dashboard behavior or commit.

## Review Focus

- Repeating one non-empty key within a store returns one transaction and decrements inventory once; test same-key replay.
- Empty or absent keys remain unlimited and do not collide with legacy rows; test two ordinary checkouts and their stock effects.
- The same key in another store is a separate transaction; test with two registered stores.
- Offline shortages preserve requested line quantities and pricing, cap stock at zero, and notify admins; test both shortage and ordinary insufficient-stock paths.
- Variant and product final stock entries retain `variant_id` as a number or JSON `null`; test a product and a variant checkout response.

---

### Task 1: Extend the existing transaction checkout

**Files:**
- Modify: `model/transaction.go`
- Modify: `repo/transaction.go`
- Modify: `service/interfaces.go`
- Modify: `service/transaction.go`
- Modify: `handler/transaction.go`
- Test: `router/offline_checkout_integration_test.go`

**Interfaces:**
- `repo.CheckoutInput` and `service.CheckoutCmd` gain `ClientTxID string`.
- `TrxRepository.Checkout` and `TrxService.Checkout` return `(*model.Trx, created bool, err error)` so the existing route can return `201` for a new checkout and `200` for a replay.
- `model.Trx` gains `ClientTxID string`, `FinalStocks []model.FinalStock`; `FinalStock` serializes `product_id`, nullable `variant_id`, and `stock`.

- [x] **Step 1: Write router integration tests first.** Cover same-key replay, per-store key scope, two checkouts with missing/empty key, sufficient offline stock and final stocks, product and variant shortage, admin notification, repeated shortage replay, and ordinary insufficient-stock rejection.
- [x] **Step 2: Run the new test file and confirm it fails because the contract is not implemented.** Run `CGO_ENABLED=1 go test ./router -run 'TestOfflineCheckout' -count=1` with the prescribed disk cache variables.
- [x] **Step 3: Add the optional `clientTxId` request path, transaction response fields/status, and partial composite unique index.** The index predicate must exclude empty IDs so old/default-empty records and online checkouts never conflict.
- [x] **Step 4: Implement per-store idempotent lookup in the existing repository checkout and carry whether the row was newly created through the service to the handler.** Populate `final_stocks` from the current product/variant stock for both new and replayed responses.
- [x] **Step 5: Implement shortage checkout only when `ClientTxID` is non-empty.** Keep requested item quantities and all current money calculations; decrement only available stock with an atomic non-negative update, record actual stock movement, set `completed_offline_shortage`, and retain the current rejection for checkouts without a key.
- [x] **Step 6: Emit one admin notification through `NotificationService.Emit` for `completed_offline_shortage`, with Indonesian title/message and the transaction reference.** Do not emit a second shortage alert on idempotent replay.
- [x] **Step 7: Run the focused integration test and then `CGO_ENABLED=1 go test ./...`.** Confirm the focused cases and all packages pass; also run `gofmt` and `git diff --check`.
