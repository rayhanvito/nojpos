# Backend F2.2 Stock Opname Implementation Plan

> **For agentic workers:** Execute this unit inline with TDD. Do not commit; AGENTS.md assigns commits to the PM.

**Goal:** Add store-scoped stock opname sessions that snapshot stock, accept partial counts, post counted stock with movements, and support cancellation.

**Architecture:** Follow the existing handler → service → repo → model layers and the purchase-order transaction/state pattern. Store session and item tables through `db.Migrate` AutoMigrate; return related current product and variant names in detail responses.

**Tech Stack:** Go, Gin, GORM, SQLite integration tests; existing dependencies only.

**Spec:** `PLAN.md` — Unit F2.2, refined by the user's backend-only task prompt.

## Global Constraints

- API routes live below `/api/v1`; errors use `{"error":"..."}` and list responses use `{"items":[...]}`.
- Request JSON uses camelCase; model response fields use snake_case; store scope comes only from token claims.
- Both admin and cashier may create and count sessions; at most one `counting` session exists per store.
- Every counted stock change posts a `stock_movements` row of type `opname` in the same transaction as the stock update and session state change.
- Session snapshots come from active products and their active variants at creation; uncounted items are not changed or moved.
- A parent-product snapshot cannot post if variants have since been added, because the parent is no longer the stock target used by checkout.
- Use AutoMigrate and existing model soft-delete conventions; do not add dependencies or commit.

## Review Focus

Checked against the task prompt: active-session uniqueness, snapshot-versus-live stock, product and variant targets, partial counts, immutable posted state, cancellation, invalid/foreign item IDs, role access, and store scoping are all assigned integration assertions below. No additional business behavior is selected beyond the spec.

---

### Task 1: Stock opname sessions and posting

**Files:**
- Create: `model/stock_opname.go`
- Modify: `model/stock.go`, `db/db.go`, `service/interfaces.go`, `router/router.go`
- Create: `repo/stock_opname.go`, `service/stock_opname.go`, `handler/stock_opname.go`
- Test: `router/stock_opname_integration_test.go`

**Interfaces:**
- Consumes: `model.Product`, `model.ProductVariant`, `model.Movement`, `middleware.ClaimsFrom`, and the existing purchase-order transaction/state pattern.
- Produces: authenticated routes `GET/POST /stock-opnames`, `GET /stock-opnames/:id`, `PUT /stock-opnames/:id/items`, `POST /stock-opnames/:id/post`, and `POST /stock-opnames/:id/cancel`.

- [x] **Step 1: Write integration tests first**

  Add tests proving snapshot-based differences after live stock changes, exact product and variant correction with `opname` movements, one active session per store, uncounted rows unchanged, posted sessions locked, cancellation leaves stock and movements untouched, incremental counts, negative counts rejected, role/store isolation, and rollback when a snapshotted parent product gains variants before posting.

- [x] **Step 2: Run tests and verify RED**

  Run `CGO_ENABLED=1 GOCACHE=/opt/.cache-go-build GOTMPDIR=/opt/.cache-go-tmp go test ./router -run StockOpname -count=1`.
  Expected: compile succeeds and integration assertions fail because the routes are not registered yet.

- [x] **Step 3: Implement the model and transactional repository**

  Add soft-deletable session/item models, a nullable counted stock and computed nullable difference, and the `opname` movement constant. Register both models in AutoMigrate. Repository operations must create the active-product snapshot, update only session-owned items, and atomically post stock, movements, and status.

- [x] **Step 4: Implement service, handler, and authenticated routes**

  Validate nonnegative counts and duplicate/foreign item IDs; map not-found, invalid-state, and active-session conflicts to clear Indonesian errors. Keep all routes accessible to authenticated admin and cashier roles.

- [x] **Step 5: Run the targeted tests and verify GREEN**

  Run `CGO_ENABLED=1 GOCACHE=/opt/.cache-go-build GOTMPDIR=/opt/.cache-go-tmp go test ./router -run StockOpname -count=1`.
  Expected: all stock opname integration tests pass.

- [x] **Step 6: Verify the full backend suite and formatting**

  Run `gofmt -w` on changed Go files, then `CGO_ENABLED=1 GOCACHE=/opt/.cache-go-build GOTMPDIR=/opt/.cache-go-tmp go test ./...`.
  Expected: suite exits 0 and changed Go files have no gofmt diff.

**Execution note:** The sandbox denied writes to `/opt/.cache-go-build` and `/opt/.cache-go-tmp`; Go cache/temp files were placed under the ignored repo-local `.superpowers/sdd/2026-10-15-f22-backend-stock-opname/` directory, then removed after verification. No cache was placed in `/tmp`.
