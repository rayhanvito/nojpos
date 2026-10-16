# F0.2 Backend CSV Export Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add store-scoped admin CSV downloads for transactions and products.

**Architecture:** Reuse Report's period/date rules and store timezone in a service, with dedicated unpaginated repository queries for transaction and product exports. A handler writes the exact Indonesian headers and rows with `encoding/csv`; both routes live in the existing admin group.

**Tech Stack:** Go, Gin, GORM, `encoding/csv`.

**Spec:** `PLAN.md`, Unit F0.2 (backend only).

## Global Constraints

- API routes stay under `/api/v1`; both exports are admin-only and use `storeID` from JWT claims.
- Transaction periods and optional dates follow `ReportRepo.Report`, including its local-day and invalid-date behavior.
- CSV output uses integer Rupiah values and `encoding/csv`; no frontend or schema changes.
- Do not commit; the user explicitly requested reviewable working-tree changes.

## Review Focus

- Commas, quotes, and embedded newlines in customer, product, or category text must round-trip as CSV fields; assert with `encoding/csv` in the integration test.
- Transaction timestamps and date filtering must use the store timezone; set a timezone and assert RFC3339 local timestamps and selected local-day rows.
- Invalid periods must return the existing JSON error shape with HTTP 400; assert in the integration test.
- Data from another store must not appear in either download; create a second store and assert its records are absent.
- Exports must include all store rows rather than a paginated subset; assert the full set created by the integration test is present.

---

### Task 1: Store-scoped CSV export endpoints

**Files:**
- Create: `service/export.go` — validate periods, resolve store timezone, and call repositories.
- Create: `handler/export.go` — write transaction and product CSV downloads.
- Modify: `repo/transaction.go` — query all store transactions in Report's period and load item quantities.
- Modify: `repo/product.go` — query all store products and category names without pagination.
- Modify: `service/interfaces.go` — expose repository export methods.
- Modify: `router/router.go` — construct and register export handler in the admin group.
- Test: `router/csv_export_integration_test.go` — verify headers, data, escaping, local timestamps/date filtering, store scope, invalid periods, attachment names, and cashier 403s.

**Interfaces:**
- `TrxRepository.ListForExport(ctx, storeID, period, timezone, date) ([]*model.Trx, error)`.
- `ProductRepository.ListForExport(ctx, storeID) ([]*model.Product, error)`.
- `ExportService.Transactions(ctx, storeID, period, date) ([]*model.Trx, string, error)` returns transactions and store timezone.
- `ExportService.Products(ctx, storeID) ([]*model.Product, string, error)` returns products and store timezone.

- [x] **Step 1: Write `TestCSVExportsAreScopedAndAdminOnly`**. Assert exact headers `id,tanggal,kasir,pelanggan,metode,status,subtotal,diskon,pajak,total,dibayar,kembalian,jumlah item` and `id,nama,sku,barcode,kategori,harga beli,harga jual,stok,satuan,aktif`, row values, RFC3339 timezone, CSV text escaping, dated attachment headers, second-store isolation, invalid-period error, and cashier 403 for both endpoints.
- [x] **Step 2: Run the focused integration test** with `go test ./router -run TestCSVExportsAreScopedAndAdminOnly -count=1`; it failed before the routes were implemented.
- [x] **Step 3: Implement repository export queries** using store-scoped GORM filters, Report's `periodStarts` and `parseLocalDay`, full unpaginated results, transaction item quantities, and category names.
- [x] **Step 4: Implement `ExportService`** using existing `validPeriods` and `StoreRepository.GetTimezone` fallback behavior.
- [x] **Step 5: Implement the CSV handler and register both routes**. Use `csv.Writer`, RFC3339 transaction timestamps in the store timezone, summed item quantities, integer money strings, and store-local dates in filenames.
- [x] **Step 6: Run the focused test**, then `gofmt` on changed Go files and `go test ./...`; all passed.
- [x] **Step 7: Review the working diff and leave all changes uncommitted**, as requested.
