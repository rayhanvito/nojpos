# F2.6 Backend Multi-Outlet Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add outlet-aware sessions, inventory, operations, and reports while preserving the current single-outlet API behavior.

**Architecture:** Keep the current handler → service → repo → model layering. Add `outlets` and `product_stocks`, backfill through `db.Migrate`, and route stock reads/writes through shared repository helpers that also mirror default-outlet stock into the legacy columns. Carry the active outlet in access and refresh sessions; use the existing outlet binding for cashier sessions and preserve store scoping in every query.

**Tech Stack:** Go, Gin, GORM, SQLite integration tests, JWT.

**Spec:** `PLAN.md`, Unit F2.6, plus the pinned backend contract in the task prompt.

## Global Constraints

- Outlet data is always scoped by `store_id` from the token.
- Existing one-outlet behavior and existing response field names remain unchanged.
- All existing transactions, shifts, products, variants, users, purchase orders, stock opname sessions, and stock movements receive the store's default outlet.
- `product_stocks` is authoritative; legacy `products.stock` and `product_variants.stock` remain mirrored from the default outlet and are not removed.
- Access tokens contain `outletId`; refresh rotation preserves the selected outlet.
- Admins may use any outlet in their store; cashiers remain bound to one outlet.
- Dashboard and reports aggregate all outlets for admins when `outletId` is absent; cashiers always see their active outlet.
- Request bodies use camelCase; list responses use `{"items": [...]}`; user-facing errors are Indonesian.
- Do not commit.

## Review Focus

- Re-running migration after partial or completed backfill preserves exactly one default outlet per store and does not overwrite per-outlet stock already created.
- Product and variant stock with `variant_id IS NULL` remains distinct from variant stock at the same outlet.
- A cashier cannot switch outlet or use a query parameter to escape the cashier's bound outlet.
- An outlet ID from another store never returns that store's sales, inventory, or operational records.
- Refund, offline shortage, adjustment, PO receipt, and opname posting each update the active outlet and preserve the default-stock mirror.
- Stock opname keeps F2.2's rule of at most one active counting session per store.

---

### Task 1: Outlet schema and safe backfill

**Files:**
- Create: `model/outlet.go`
- Modify: `model/catalog.go`, `model/inventory.go`, `model/transaction.go`, `model/shift.go`, `model/stock.go`, `model/purchase_order.go`, `model/stock_opname.go`, `db/db.go`
- Test: `db/multi_outlet_migration_test.go`

**Interfaces:**
- Produces `model.Outlet`, `model.ProductStock`, nullable `OutletID` fields on outlet-owned records, and `db.Migrate` backfill behavior.
- Default outlet name is the existing store name; use `Pusat` only when a store name is blank.

- [x] Write a SQLite migration test with pre-existing store/product/variant/transaction/open-shift/user rows; assert one default outlet, all legacy records linked, product/variant stock seeded, and the open shift left open.
- [x] Run `GOCACHE=... GOTMPDIR=... TMPDIR=... go test ./db -run TestMigrateBackfillsDefaultOutletIdempotently -count=1`; observe the missing schema/backfill failure.
- [x] Add outlet and product-stock models; AutoMigrate them before dependent tables; add `outlet_id` columns to the legacy rows and run an idempotent backfill that only seeds missing product-stock rows.
- [x] Re-run the focused migration test and assert a second call preserves one default and existing `product_stocks.qty` values.

### Task 2: Outlet API, cashier binding, and session claims

**Files:**
- Create: `repo/outlet.go`, `service/outlet.go`, `handler/outlet.go`
- Modify: `model/models.go`, `model/outlet.go`, `repo/user.go`, `repo/cashier.go`, `repo/refresh.go`, `service/auth.go`, `service/interfaces.go`, `service/user.go`, `handler/auth.go`, `handler/user.go`, `router/router.go`
- Test: `router/multi_outlet_integration_test.go`

**Interfaces:**
- Add `GET /api/v1/outlets`, `POST /api/v1/outlets`, `PUT /api/v1/outlets/:id`, and `POST /api/v1/outlets/switch` with `{ "outletId": number }`.
- `POST /outlets/switch` returns the existing auth response shape with a rotated token pair and the active outlet on `user`.
- Add optional `outletId` to cashier create/update requests; omission selects the current outlet on create and leaves the assignment unchanged on update.
- `service.AuthService.issueTokens(ctx, ownerID, cashierID, outletID)` stores the same outlet in JWT and refresh rows.

- [x] Write router tests for list wrapping, admin-only create/update, admin switch to a same-store outlet, cashier switch denial, cashier outlet assignment, and foreign-store outlet rejection.
- [x] Run the focused router test and verify the expected missing endpoint/session behavior.
- [x] Implement the outlet repository/service/handler using existing response and RBAC patterns; persist outlet selection in JWT and refresh rotation; enforce cashier binding and same-store lookup.
- [x] Re-run the focused router test and the auth, user, and outlet tests.

### Task 3: Outlet-aware catalog, checkout, refund, and adjustments

**Files:**
- Modify: `repo/stock.go` (create), `repo/product.go`, `repo/inventory.go`, `repo/transaction.go`, `service/interfaces.go`, `service/catalog.go`, `service/transaction.go`, `handler/catalog.go`, `handler/transaction.go`, `handler/stock.go`, `handler/export.go`
- Test: `router/multi_outlet_integration_test.go`, `router/offline_checkout_integration_test.go`, `router/inventory_integration_test.go`

**Interfaces:**
- Repository stock helpers read/set/adjust `(outletID, productID, optional variantID)` in `product_stocks` and mirror the legacy column only for the default outlet.
- Catalog listing/detail/export and transaction `final_stocks` return the active outlet's stock.
- Checkout and refund receive the active outlet ID from claims and update only its stock row.

- [x] Add the two-outlet scenario: create a product with stock 5 in outlet A, stock 7 in outlet B, sell in A, and assert A decrements while B and `final_stocks` remain correct.
- [x] Run the focused test and confirm the current single-column implementation fails the outlet isolation assertion.
- [x] Route catalog, checkout, refund, and adjustment stock operations through `product_stocks`; include new product/variant initial stock in the selected outlet and update the legacy mirror only when that outlet is default.
- [x] Re-run focused outlet, offline checkout, inventory, promo, loyalty, and refund tests.

### Task 4: Outlet-aware purchase orders and stock opname

**Files:**
- Modify: `repo/purchase_order.go`, `repo/stock_opname.go`, `repo/inventory.go`, `service/purchase_order.go`, `service/stock_opname.go`, `service/interfaces.go`, `service/catalog.go`, `handler/purchase_order.go`, `handler/stock_opname.go`
- Test: `router/multi_outlet_integration_test.go`, `router/purchase_order_integration_test.go`, `router/stock_opname_integration_test.go`

**Interfaces:**
- PO creation/receipt and opname create/list/detail/post receive an active outlet ID.
- `purchase_orders`, `stock_opname_sessions`, and `stock_movements` are scoped by their outlet; their stock changes use the shared product-stock helpers.

- [x] Add assertions that a PO receipt and opname posting affect only the selected outlet, preserve movement records with the outlet ID, and retain existing default-outlet behavior.
- [x] Run focused PO/opname tests and observe the outlet scope failure.
- [x] Persist outlet IDs on PO/opname/movement rows and use the shared stock helpers for receipt/posting.
- [x] Re-run focused PO, opname, adjustment, and movement tests.

### Task 5: Outlet filtering for dashboard, reports, and tax report

**Files:**
- Modify: `repo/report.go`, `service/settings.go`, `service/tax_report.go`, `service/interfaces.go`, `handler/settings.go`, `handler/tax_report.go`
- Test: `router/multi_outlet_integration_test.go`, `router/tax_report_integration_test.go`, `router/refund_report_integration_test.go`

**Interfaces:**
- Dashboard, report, and tax report accept an optional outlet filter; absent outlet means active outlet for cashier and all outlets for admin.
- A requested outlet is validated against the token store before report queries execute.

- [x] Add test sales in two outlets; assert admin aggregate, admin single-outlet, cashier-bound totals, and a foreign outlet ID with no cross-store data.
- [x] Run the focused test and observe the missing aggregate/filter behavior.
- [x] Apply outlet scoping to dashboard, reports including F2.4 tax report, and stock valuation; preserve existing date, timezone, refund, and transaction semantics.
- [x] Re-run focused report/tax/refund tests.

### Task 6: Full regression and review

**Files:**
- Test: all backend packages, especially `router/`

- [ ] Run `gofmt` on every changed Go file and `git diff --check`.
- [ ] Run `CGO_ENABLED=1 GOCACHE=... GOTMPDIR=... TMPDIR=... go test ./...` and read the full result.
- [ ] Review all direct legacy `stock` reads/writes and confirm each is either a schema mirror or display from outlet-aware `product_stocks`.
- [ ] Review the complete diff against Unit F2.6, including migration safety, authorization, store isolation, and old one-outlet behavior; record any remaining limits for the PM.
