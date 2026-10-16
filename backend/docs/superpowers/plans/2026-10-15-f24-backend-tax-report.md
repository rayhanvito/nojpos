# F2.4 Backend Tax Report Implementation Plan

> **For agentic workers:** Use superpowers:executing-plans inline. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an admin-only monthly tax summary endpoint and a matching CSV export, scoped to the store timezone.

**Architecture:** Add a tax summary query to `ReportRepo`, expose it through a small reporting service that resolves and validates the local month, and serve JSON and CSV from one handler. Reuse the existing refund aggregation, admin route group, and CSV writer conventions.

**Tech Stack:** Go, Gin, GORM, SQLite integration tests, `encoding/csv`; no new dependencies or schema changes.

**Spec:** `PLAN.md`, Unit F2.4, backend scope from the task prompt.

## Global Constraints

- Routes stay under `/api/v1`, are admin-only, and use `storeID` from JWT claims.
- Monthly boundaries use the store timezone; missing `month` means the current local month; invalid `YYYY-MM` returns HTTP 400.
- Gross sales, refund total, net sales, and refund count follow the existing F0.1 report netting and refund-date attribution.
- DPP is the sum of `total - tax` for counted transactions; PPN is the sum of their `tax` fields; PPh Final is rounded half-up from net omzet at 0.5%.
- CSV uses `encoding/csv`, Indonesian headers, integer Rupiah values, and a filename containing the requested month.
- Do not change existing dashboard/report behavior, frontend files, schema, or commit changes.
- Run Go with `CGO_ENABLED=1 GOCACHE=/opt/.cache-go-build GOTMPDIR=/opt/.cache-go-tmp TMPDIR=/opt/.cache-go-tmp`.

## Review Focus

- A month boundary that differs between UTC and the store timezone must include the correct rows; integration test places sale and refund near a Tokyo local month boundary.
- A fully empty month must return HTTP 200 with every numeric field zero; integration test queries an empty month.
- A malformed or out-of-range month must return HTTP 400 with the standard JSON error shape; integration test checks `2026-13`.
- Partial refunds must use F0.1's proportional-tax amount and refund-created month; integration test asserts gross, refund, net, DPP, PPN, counts, and PPh values.
- CSV headers and row values must describe the same report as JSON; integration test parses the CSV and compares exact cells.

---

### Task 1: Monthly tax summary and CSV export

**Files:**
- Modify: `model/reporting.go` — define the response fields and JSON names.
- Modify: `service/interfaces.go` — expose the repository summary query.
- Modify: `repo/report.go` — aggregate completed/refunded transactions and reuse `refunds` for refund totals and counts.
- Create: `service/tax_report.go` — resolve store timezone, default/validate `YYYY-MM`, build local month bounds, and calculate the PPh estimate.
- Create: `handler/tax_report.go` — return JSON and write a one-row Indonesian CSV.
- Modify: `router/router.go` — wire the handler and register `/reports/tax` plus `/reports/tax/export` in the existing admin group.
- Test: `router/tax_report_integration_test.go` — exercise checkout/refund, local-month boundaries, empty and invalid months, CSV parity, and cashier denial.

**Interfaces:**
- `ReportRepository.TaxReport(ctx context.Context, storeID uint, start, end time.Time, timezone string) (*model.TaxReport, error)` returns aggregate values for the half-open period.
- `TaxReportService.Report(ctx context.Context, storeID uint, month string) (*model.TaxReport, error)` returns the validated local month and its aggregate.
- JSON fields: `month`, `omzet_bruto`, `refund_total`, `omzet_neto`, `dpp`, `ppn_terkumpul`, `trx_count`, `refund_count`, `pph_final_estimate`.
- CSV header: `bulan,omzet bruto,total refund,omzet neto,DPP,PPN terkumpul,jumlah transaksi,jumlah refund,estimasi PPh Final`.

- [x] Write the integration test with a Rp1.000 item, 10% additive tax, a two-item sale, and a one-item refund; assert November 2026 values of gross 2,200, refund 1,100, net 1,100, DPP 2,000, PPN 200, transaction count 1, refund count 1, and PPh 6.
- [x] Run `CGO_ENABLED=1 GOCACHE=/opt/.cache-go-build GOTMPDIR=/opt/.cache-go-tmp TMPDIR=/opt/.cache-go-tmp go test ./router -run TestTaxReportAndCSV -count=1`; it compiled and failed at HTTP 404 before route registration.
- [x] Implement the report repository query using the existing completed/refunded status set and refund aggregation, with UTC query bounds derived from the store-local month.
- [x] Implement month validation/defaulting and the integer half-up 0.5% calculation in the service; keep errors in the existing JSON error shape.
- [x] Add JSON and CSV handlers, wire both admin routes, and ensure both routes call the same service result.
- [x] Re-run the focused integration test; exact amounts, Tokyo month boundary, empty/invalid month, CSV parity, default month, and cashier role assertions passed.
- [x] Run `gofmt` on changed Go files, confirm no gofmt diff, run `git diff --check`, then run the full backend suite; all packages passed.
- [x] Review the working diff against Unit F2.4 and leave changes uncommitted.

**Execution note:** The sandbox rejects writes to `/opt/.cache-go-build` and `/opt/.cache-go-tmp`; Go cache and temp files were kept in this ignored workspace on the `/opt` disk, not `/tmp`.
