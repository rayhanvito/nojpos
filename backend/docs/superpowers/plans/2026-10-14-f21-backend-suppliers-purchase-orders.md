# F2.1 Backend Supplier and Purchase Orders Implementation Plan

> **Scope:** Backend only, per task prompt. No commit.

**Goal:** Add store-scoped supplier CRUD and purchase order draft, ordering, receipt, and cancellation flows with atomic stock movements.

**Architecture:** Add supplier and purchase order models, register them in GORM AutoMigrate, and keep HTTP parsing, business validation, and persistence in the existing handler/service/repo layers. Validate catalog ownership through the existing product repository. Each receipt updates PO items, stock, movements, and PO status in one GORM transaction.

**Tech:** Go, Gin, GORM AutoMigrate, SQLite router integration tests.

**Spec:** `PLAN.md`, Unit F2.1; detailed backend requirements in the user task.

## Global Constraints

- Paths stay under `/api/v1`; store scope comes from JWT claims; reads require login and mutations are admin-only.
- Request JSON uses camelCase; responses follow model JSON tags; list responses use `{"items": [...]}`.
- Money is integer Rupiah; quantities must be positive; buy prices must be nonnegative.
- New schemas use GORM AutoMigrate only; no dependencies or SQL migration files.
- Every receipt changes stock and creates a `purchase` movement atomically; cancel never changes stock.
- Leave the work uncommitted and do not change frontend or deployment files.
- The `YYYYMMDD` segment of a PO number uses the store's configured calendar timezone; the existing `Asia/Makassar` setting is the fallback.

## Review Focus

No spec-derived failure modes remain outside the integration tests in the tasks below.

## Tasks

### Task 1: Supplier model and CRUD

**Files:**
- Create `model/supplier.go`, `repo/supplier.go`, `service/supplier.go`, `handler/supplier.go`.
- Modify `db/db.go`, `service/interfaces.go`, and `router/router.go`.
- Test in `router/purchase_order_integration_test.go`.

**Produces:** `SupplierRepository.List(ctx context.Context, storeID uint) ([]*model.Supplier, error)`, `GetByID(ctx context.Context, storeID, id uint) (*model.Supplier, error)`, `Create(ctx context.Context, storeID uint, supplier *model.Supplier) (*model.Supplier, error)`, `Update(ctx context.Context, storeID, id uint, supplier *model.Supplier) (*model.Supplier, error)`, and soft `Delete(ctx context.Context, storeID, id uint) error`, all store-scoped; authenticated `GET /suppliers` and `GET /suppliers/:id`; admin-only `POST /suppliers`, `PUT /suppliers/:id`, and `DELETE /suppliers/:id`.

- [ ] **Step 1: Write `TestSupplierCRUDIsStoreScopedAndAdminOnly`.** Cover `{items: [...]}`, detail, create/update/delete, foreign-store not-found, cashier reads, and cashier mutation 403s.
- [ ] **Step 2: Verify RED.** Run `CGO_ENABLED=1 go test ./router -run '^TestSupplierCRUDIsStoreScopedAndAdminOnly$' -count=1`. Expected: FAIL because supplier routes do not exist.
- [ ] **Step 3: Add the soft-delete supplier model to `db.Migrate`; implement store-scoped persistence, name validation, Indonesian errors, route wiring, and role grouping.** Supplier fields are `name`, `contact`, and `address`; store ID is always from claims.
- [ ] **Step 4: Verify GREEN.** Repeat the focused test command. Expected: PASS.

### Task 2: Purchase order draft and order transition

**Files:**
- Create `model/purchase_order.go`, `repo/purchase_order.go`, `service/purchase_order.go`, and `handler/purchase_order.go`.
- Modify `model/stock.go`, `db/db.go`, `service/interfaces.go`, and `router/router.go`.
- Extend `router/purchase_order_integration_test.go`.

**Produces:** `PurchaseOrderRepository.List(ctx context.Context, storeID uint, f repo.PurchaseOrderFilter) ([]*model.PurchaseOrder, error)`, `GetByID(ctx context.Context, storeID, id uint) (*model.PurchaseOrder, error)`, `CreateDraft(ctx context.Context, storeID, supplierID uint, items []model.PurchaseOrderItem, datePrefix string) (*model.PurchaseOrder, error)`, `UpdateDraft(ctx context.Context, storeID, id, supplierID uint, items []model.PurchaseOrderItem) (*model.PurchaseOrder, error)`, `DeleteDraft(ctx context.Context, storeID, id uint) error`, and `Order(ctx context.Context, storeID, id uint) (*model.PurchaseOrder, error)`; `GET /purchase-orders` with optional `status` and `supplierId` filters; scoped detail containing supplier and items; admin-only create/update/delete draft operations; admin-only `POST /purchase-orders/:id/order`. A PO number uses `PO-YYYYMMDD-####`, unique per store and daily sequence. Request fields are `supplierId`, `items`, `productId`, optional `variantId`, `qty`, and `buyPrice`; `supplierId`, `productId`, `qty`, and `buyPrice` are required. A product that has variants must use a variant, following F1.3 inventory ownership.

- [ ] **Step 1: Write `TestPurchaseOrderDraftNumberingAndOrderLock` and `TestPurchaseOrderRejectsForeignCatalogAndInvalidItems`.** Cover number format, daily sequence and per-store isolation, deleted draft numbers remaining consumed, detail with supplier/items after supplier soft deletion, status/supplier filters, full draft item replacement, draft deletion, `draft -> ordered`, ordered PUT/DELETE rejection, cashier read/admin-only mutation access, foreign supplier/product/variant, wrong-product variant, missing fields, empty item list, nonpositive quantity, negative buy price, and required variant selection for variant products.
- [ ] **Step 2: Verify RED.** Run `CGO_ENABLED=1 go test ./router -run 'TestPurchaseOrder(DraftNumberingAndOrderLock|RejectsForeignCatalogAndInvalidItems)$' -count=1`. Expected: FAIL because purchase order routes do not exist.
- [ ] **Step 3: Add `purchase_orders` and `purchase_order_items` models and register them with AutoMigrate.** Use a unique `(store_id, number)` index, exact status values, nullable variant ID, integer quantities and prices, `qty_received` defaulting to zero, and hard-replace draft item rows on PUT.
- [ ] **Step 4: Implement scoped PO persistence and service validation.** Validate supplier/product/variant ownership through repositories; create/update/delete only drafts; generate per-store daily numbers; atomically transition `draft -> ordered`; return Indonesian 400 errors for invalid data/state and a clear 409 error when ordered items are locked.
- [ ] **Step 5: Add handlers and routes.** Authenticated reads return `{items: [...]}` and detail includes `supplier` plus `items`; admin group owns all mutations.
- [ ] **Step 6: Verify GREEN.** Repeat the focused test command. Expected: both tests PASS.

### Task 3: Atomic receipt and cancellation

**Files:** Extend `repo/purchase_order.go`, `service/purchase_order.go`, and `handler/purchase_order.go`; extend `router/purchase_order_integration_test.go`.

**Produces:** Adds `PurchaseOrderRepository.Receive(ctx context.Context, storeID, id uint, items []repo.PurchaseOrderReceiptItem) (*model.PurchaseOrder, error)` and `Cancel(ctx context.Context, storeID, id uint) (*model.PurchaseOrder, error)`; admin-only `POST /purchase-orders/:id/receive` and `POST /purchase-orders/:id/cancel`. Receipt is allowed only from `ordered` or `partially_received`, never exceeds ordered quantity cumulatively, increments product or variant stock and `qty_received`, records a positive `purchase` movement, and sets the status to `partially_received` or `fully_received`. Cancellation is allowed only from `draft` or `ordered` and has no stock effect.

- [ ] **Step 1: Write `TestPurchaseOrderPartialAndFullReceipt`, `TestPurchaseOrderOverflowReceiptRollsBack`, and `TestPurchaseOrderCancelAndInvalidReceiptStates`.** Assert partial/full quantities and statuses, product/variant stock, immutable buy-price snapshot, positive `purchase` movements, total rollback on overflow, and no stock changes on cancellation. Assert receipts from draft/cancelled/fully-received and cancellation from partially-received return 400.
- [ ] **Step 2: Verify RED.** Run `CGO_ENABLED=1 go test ./router -run 'TestPurchaseOrder(PartialAndFullReceipt|OverflowReceiptRollsBack|CancelAndInvalidReceiptStates)$' -count=1`. Expected: FAIL because receive/cancel actions are not registered.
- [ ] **Step 3: Implement receipt and cancellation with locked PO state checks inside one transaction.** Update received quantities, product/variant stock, `purchase` movements, and PO status atomically; any invalid/overflow line rolls back the entire request.
- [ ] **Step 4: Verify GREEN.** Repeat the focused test command. Expected: all three tests PASS.
- [ ] **Step 5: Format and verify.** Run `gofmt -w` on changed Go files, ensure `gofmt -l` reports none, then run `CGO_ENABLED=1 go test ./...` with the configured on-disk Go cache. Review the complete diff and leave all changes uncommitted.
