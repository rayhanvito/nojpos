# Roadmap Status

Date: 2026-06-20
Project: NOJ POS P0 Production Readiness

## Done Verified

- Parked orders revision + lease
  - Backend revision field, 90-second lease, acquire/refresh/release routes, conflict handling, tests
  - Flutter lease lifecycle, heartbeat, release, conflict UX, tests
- Settings aggregate + security settings
  - Backend settings aggregate and security policy contract
  - Flutter safe read-only settings summary
- Async payment webhook/polling/expiry
  - Backend async QRIS/EDC/transfer pending status, webhook verification/idempotency, expiry, settlement
  - Flutter pending payment polling and status UI
- Reports backend + Flutter
  - Backend reports contract for sales summary, sold products, payment methods, cashier shifts, void/refund audit, top-10
  - Flutter DTO/repository/UI panels
- Inventory count/waste/transfer/in-transit
  - Backend stock count, waste, transfer lifecycle, in-transit reconciliation
  - Flutter DTO/repository/provider and minimal inventory UI
- Promotion quote + UI promo
  - Backend promotion quote endpoint/contract, token/hash, checkout stale quote protection
  - Flutter promo code panel, applied/rejected promotions, checkout quote token forwarding

## In Progress

- GitHub checkpoint after Promotion
  - Audit workspace
  - Create progress docs
  - Configure remote
  - Create checkpoint branch
  - Commit and push checkpoint

## Ready Next

- Refund
  - Backend contract first
  - Refund mutation, refund payment/stock/audit behavior
  - Flutter only after backend tests pass

## Backlog

- Store open/close
- Screen lock
- Admin web decision / apps-web
- Release candidate QA

## Release Candidate Gates

- Backend full test pass
- Flutter analyze pass
- Flutter full test pass
- API route list reviewed
- No secrets/cache/build artifacts committed
- Tenant and outlet isolation verified for P0 flows
- Role guards verified for privileged actions
- Idempotency verified for sensitive writes
- Audit trail verified for sensitive/stock/payment mutations
- Manual smoke test on POS flow
- Seed/demo data reviewed
- Production environment and secret handling reviewed
- Deployment notes prepared
- GitHub checkpoint branch merged only after review