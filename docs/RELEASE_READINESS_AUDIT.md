# Release Readiness Audit

Audit date: 2026-06-20  
Scope: implementation evidence for the Wave 1-9 and PRD v2 release gate.

## Verified foundation

The following commands completed successfully on the current workspace:

- `apps/backend`: `php artisan test` — 69 tests, 572 assertions.
- `apps/backend`: `php artisan route:list --path=api/v1` — 45 API routes.
- `apps/cashier`: `flutter analyze` — no issues.
- `apps/cashier`: `flutter test` — 65 tests.
- `apps/cashier`: `flutter build apk --debug` — APK built.

The current test suite contains evidence for the core Wave 1-7 hardening work, including server-owned quote price/rounding, terminal-context spoof rejection, idempotency replay/conflict, payment overpay and confirmation guards, shift close rollback/immutability, bounded checkout recovery, and outlet-local time boundaries.

## Wave 9 evidence gap

PRD v2 still requires these release domains, but the current backend has no matching route or migration/model evidence:

| PRD domain | Current evidence | Status |
| --- | --- | --- |
| Parked order revision and lease | No `parked_orders` schema, lease fields, order route, or `ORDER_LOCKED`/`CONFLICT_REVISION` contract | Not implemented |
| Settings/security aggregate | Outlet and payment configuration fragments exist; no settings aggregate/security policy contract | Partial |
| Complete reports | Sales summary exists; sold-product, top-10, shift, and void/refund audit reports are absent | Partial |
| Inventory count/waste/transfer | Purchase and stock-on-hand exist; count, waste, transfer, and in-transit reconciliation are absent | Not implemented |
| Promotion quote contract | No promotion quote contract or terminal promotion flow | Not implemented |
| Refund | No refund route, model/migration, or refund state/movement tests | Not implemented |
| Store open/close | No store-state route, model/migration, or unresolved-shift guard | Not implemented |
| Admin web | Deliberately absent; architecture decision is required before creating `apps/web` | Pending decision |

## Release decision

The build is green, but the full PRD v2 release gate is not green because the domains above are still absent or partial. Do not mark Wave 9 or a production pilot as complete until the relevant rows have implementation and acceptance-test evidence.

## Next safe task

Implement one bounded domain at a time. The PRD sequence recommends starting with **parked orders with server revision and 90-second lease lock**. Locked PRD §18 decisions 12 (lease/concurrency) and 16 (numbering) apply. It must be implemented as a Laravel contract first, then consumed by Flutter; it must not reuse local order numbering as authority.

## Required product-owner decisions before release

PRD §18 is locked with product-owner-approved implementation defaults. Future changes to those policies require an approved PRD revision.
