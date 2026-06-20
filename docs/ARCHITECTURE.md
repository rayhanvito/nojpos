# NojPOS Architecture Map

Status: read-only architecture assessment, strengthened for execution review  
Snapshot date: 2026-06-19  
Branch observed: `feat/flutter-ui-consistency`  
Commit observed: `de07cbb`  
Worktree state: dirty; this is a snapshot of the current local workspace, not necessarily a clean committed architecture.

Primary product reference: `NOJPOS_POS_FLUTTER_BACKEND_PRD.md` v2.0  
Risk reference: `docs/ANALYSIS_FINDINGS.md`  
Execution reference: `docs/IMPROVEMENT_PLAN.md`

## 1. Executive Architecture Verdict

The current repository is a **two-application workspace**:

1. Flutter cashier terminal.
2. Laravel API backend.

There is **no active Next.js/admin web application** in the current app tree. There is also **no Drift/offline database implementation** in the Flutter cashier. Any document or prompt that assumes `apps/web`, Next.js admin, or Drift must be reconciled before execution.

The main architecture risk is not missing folders. The main risk is that the existing money/shift/payment flows have begun to grow before the backend contract is fully hardened around server-side pricing, terminal actor context, atomic idempotency, transactional audit, row locks, and policy enforcement.

## 2. Current Workspace Reality

| Surface | Location | Actual stack | Entry point | Current verification command |
| --- | --- | --- | --- | --- |
| Cashier terminal | `apps/cashier` | Flutter, Dart, Riverpod, GoRouter, Dio | `lib/main.dart` -> `NojposApp` | `flutter analyze`, `flutter test`, `flutter build apk --debug` |
| API backend | `apps/backend` | PHP `^8.3`, Laravel `^13.8`, Sanctum `^4.3`, PHPUnit | `public/index.php`, `routes/api.php` | `php artisan test`, `php artisan route:list --path=api/v1` |
| Admin web | Not present | Not implemented | Not available | Do not run/build |
| Shared package | Not present | Not implemented | Not available | Do not assume |

Important observations:

- `apps/backend/package.json` belongs to Laravel/Vite tooling; it is not a standalone admin web app.
- There is no `apps/web`, `apps/admin`, or Next.js app in the current app tree.
- Flutter does not use Drift. `apps/cashier/pubspec.yaml` does not show a Drift dependency, and local/offline database is outside current architecture constraints.
- The only persistent client-side write recovery currently observed is checkout outbox under `apps/cashier/lib/core/outbox/`.
- The Git metadata includes `.git/worktrees/nojpos_final_fix-admin`, but that is worktree metadata, not an active web app directory in the current workspace.

## 3. Source-of-Truth Rules

When documents disagree, apply this order:

1. PRD v2.0 for intended product behavior.
2. This architecture map for current repo reality.
3. Analysis findings for risk classification.
4. Improvement plan for execution sequencing.
5. Design system docs for UI behavior/visual consistency.
6. Legacy task breakdown only after reconciliation.

`Task & Subtask Breakdown.md` must not be pasted to Codex blindly because it contains old/future assumptions: Drift, Next.js admin web, and possibly broader offline/admin scope than the current repo supports.

## 4. Directory Map

### 4.1 Root

```text
nojpos_final_fix/
  AGENTS.md
  NOJPOS_POS_FLUTTER_BACKEND_PRD.md
  Task & Subtask Breakdown.md
  docs/
    ANALYSIS_FINDINGS.md
    ARCHITECTURE.md
    DESIGN.md
    IMPROVEMENT_PLAN.md
    design-tokens.json
  apps/
    backend/
    cashier/
  Flow dan Contoh UI/
```

### 4.2 Laravel API

```text
apps/backend/
  routes/api.php                 API v1 route composition and middleware assignment
  app/Http/Controllers/Api/V1/   Operational controllers
  app/Http/Middleware/           Role and idempotency middleware
  app/Models/                    Tenant-domain Eloquent models
  app/Models/Concerns/           Business scope and UUID concerns
  app/Models/Scopes/             BusinessScope
  app/Services/                  TransactionQuoteService and subscription/superadmin services
  app/Support/                   ApiResponse and NojPOS audit helper
  database/migrations/           Operational and subscription schema
  database/seeders/              Demo/seed data
  tests/Feature/                 API, tenant, subscription, superadmin tests
  tests/Unit/                    Unit tests and Laravel examples
```

`routes/api.php` is the public API contract. It mounts `/api/v1`, applies Sanctum to protected routes, applies `role:superadmin` to superadmin routes, and applies `idempotency` selectively to transactions, payments, voids, shift close, and selected superadmin writes.

### 4.3 Flutter Cashier

```text
apps/cashier/lib/
  main.dart                      ProviderScope bootstrap
  app/                            NojposApp, router, theme, session provider
  core/network/                  Dio API client and envelope exceptions
  core/storage/                  Token and device persistence
  core/outbox/                   Persistent checkout retry implementation
  core/printing/                 Receipt/printer adapters
  core/share/                    Share/copy adapters
  features/auth/                 Login, outlet selection, PIN, splash, sync
  features/shift/                Shift gate and shift API repository
  features/pos/                  Catalog, cart, order panel, operations, dialogs
  features/payment/              Payment page and payment flow
  features/transactions/         Transaction repository and success screen
  features/orders/               Local order repository
  features/customers/            Customer repository
  features/inventory/            Inventory repository/screens
  features/attendance/           Attendance repository/screens
  features/reports/              Report repository/provider
  features/staff/                Staff repository/widgets
  features/notifications/        Checkout outbox/retry center
  shared/models/                 Cross-feature DTOs/models
```

Intended Flutter layering is:

```text
Widget/Page -> Feature provider/notifier -> Repository interface/implementation -> ApiClient -> Laravel API
```

The current major exception is `NojposSessionNotifier`, which orchestrates many unrelated domains and becomes a cross-feature god object.

## 5. Current Runtime Lifecycle

### 5.1 Authentication and terminal context as implemented

1. Flutter posts email/password/device UUID to `POST /api/v1/auth/login`.
2. Backend validates credentials, creates/discovers a device, creates a Sanctum token, audits `auth.login`, and returns session data.
3. Flutter persists token/device context and routes through outlet selection and PIN.
4. `POST /auth/pin-switch` verifies a PIN and returns a cashier payload.
5. The selected cashier is then held in Flutter session state.

Critical gap: PIN switch does not establish a durable server-side terminal actor session that future sensitive writes must derive from. Later writes still accept actor/context IDs from the client in several places.

### 5.2 Backend write lifecycle as implemented

Typical current path:

```text
Route middleware -> controller validation -> business_id from token -> direct query/controller logic -> optional DB::transaction -> optional Nojpos::audit -> ApiResponse
```

This is a workable baseline, but it is not yet strong enough for POS money flow because the following concerns are inconsistent:

- terminal actor/session resolution,
- full tuple validation of outlet/device/cashier/shift,
- idempotency reservation before effect,
- row locks for mutable state,
- audit inside same DB transaction,
- service-layer separation,
- policy/authorization matrix.

### 5.3 Flutter write lifecycle as implemented

Typical current path:

```text
Widget -> session notifier or feature provider -> repository -> ApiClient -> backend -> notifier state update
```

Checkout-specific path:

```text
Payment UI -> CheckoutOutboxController.submitOrQueue -> TransactionRepository -> ApiClient -> Backend / one-record checkout recovery
```

Checkout recovery keeps at most one unresolved idempotency-key-bound record. On app resume or restart it queries the tenant-scoped backend recovery endpoint before allowing another sale; it never automatically replays a stored draft after process recovery.

## 6. Current Domain Source-of-Truth Map

| Domain | Backend source | Flutter source | Current verdict |
| --- | --- | --- | --- |
| Business/tenant | `businesses`, token principal | Auth/session state | Baseline exists |
| Outlet | `outlets.timezone` | Outlet selection/session | Outlet-local business clock baseline exists; broader settings remain incomplete |
| Device | `devices` | persisted device UUID/ID | Partial; full enrollment/revoke/assignment guard missing |
| Acting cashier | `users` + PIN response | Flutter session selected cashier | Risky; server-side terminal actor context missing |
| Shift | `shift_sessions`, `cash_movements` | ShiftRepository + session notifier | Partial; close/cash/open need hardening |
| Catalog | `products`, `product_categories` | Product/catalog providers | Baseline exists |
| Cart | None until quote/checkout | local cart state | Expected local draft, but totals must come from server |
| Quote | `TransactionQuoteService` | transaction repository/session | Exists but trusts client price/rounding |
| Transaction | `transactions`, `transaction_items` | TransactionRepository | Exists, but context and pricing contract unsafe |
| Payment | `payments` | payment screen/repository | Exists, but state transition service missing |
| Parked/saved order | `transactions.status=held` and local repository | local saved orders/session | Competing sources of truth |
| Inventory | `stock_movements`, purchases | InventoryRepository | Partial; document lifecycle missing |
| Attendance | `attendance_records` | AttendanceRepository | Partial; global-open rule/policy incomplete |
| Reports | aggregate queries + `BusinessClock` | ReportRepository | Outlet-local day/week/month windows exist; definitions/scalability remain partial |
| Settings | outlet columns/payment configs | mixed UI/config state | No full settings aggregate/security model |
| Audit | `audit_logs` via helper | not authoritative | Inconsistent and often outside transaction |
| Idempotency | `idempotency_keys` via middleware | UUID headers from repositories | Race window and incomplete coverage |

## 7. Current vs Target Backend Architecture

### 7.1 Current shape

```text
Controller
  validates request
  reads business_id from user
  often performs domain logic directly
  sometimes opens DB transaction
  sometimes writes audit
  returns ApiResponse
```

### 7.2 Target shape for sensitive writes

```text
Route
  -> auth:sanctum
  -> device/terminal guard when terminal-bound
  -> idempotency atomic reservation
  -> FormRequest/validated DTO
  -> Policy/ActionAuthorizer
  -> Domain Service
      -> DB::transaction
      -> row locks for mutable aggregates
      -> domain mutation
      -> audit inside same transaction
      -> response snapshot
  -> ApiResponse/Resource
```

### 7.3 Required service boundaries

| Service | Responsibility |
| --- | --- |
| `IdempotencyService` | Atomic reserve/replay/mismatch/in-progress handling |
| `AuditLogService` | Mandatory immutable audit inside domain transaction |
| `TerminalSessionService` | Bind token, cashier PIN, outlet, device, shift, permissions |
| `CheckoutQuoteService` | Server-owned pricing, tax, service, rounding, final quote snapshot |
| `CheckoutService` | Atomic transaction, payment, stock movement, audit, idempotency result |
| `PaymentService` | Payment state machine, row locks, provider confirmation boundary |
| `ShiftService` | Open/current/cash movement/close/pending_close lifecycle |
| `InventoryDocumentService` | Purchase/count/waste/transfer finalization and stock movement invariants |
| `BusinessClock` | Outlet timezone boundaries for reports, shifts, document numbers |
| `DocumentNumberService` | Server-side unique numbering by business/outlet/type/month |

## 8. Current vs Target Flutter Architecture

### 8.1 Current strengths

- Riverpod is present.
- GoRouter is present.
- Dio API client centralizes envelope parsing.
- Feature folders exist.
- Repositories exist for several domains.
- Some retry/printing/share abstractions exist.

### 8.2 Current risks

- `NojposSessionNotifier` handles too many domains.
- Large screens/dialog files increase regression risk.
- Checkout retry is limited to one unresolved idempotency-key-bound recovery record.
- Flutter sends fields that should not be authoritative: cashier/device/shift context, unit price, rounding, client totals.
- Payment and shift UI currently mirror backend weaknesses.

### 8.3 Target shape

```text
features/auth          AuthSessionProvider, OutletSelectionProvider, PinProvider
features/terminal      TerminalSessionProvider, DeviceProvider, ConnectivityProvider
features/shift         ShiftProvider, CashDrawerProvider, CloseShiftProvider
features/pos           CatalogProvider, CartProvider, QuoteProvider, ParkedOrderProvider
features/payment       PaymentProvider, CheckoutRetryProvider
features/inventory     InventoryDocument providers
features/reports       Report filter/results providers
```

Target rule: one feature owns its loading/error/result state. Avoid a single global busy/error channel for unrelated operations.

## 9. API Contract Conventions Actually in Use

| Concern | Current convention | Risk/status |
| --- | --- | --- |
| Success envelope | `{data, meta}` via `ApiResponse` | Good baseline |
| Error envelope | `{error:{code,message,details}}` | Good baseline, codes need standardization |
| Auth | Sanctum bearer token | Good baseline |
| Tenant scope | `business_id` and global scope/manual query predicates | Baseline, but actor/outlet/device tuple incomplete |
| UUIDs | `UsesUuid` and manual UUID generation | Good baseline |
| Money | integer rupiah in storage/DTOs | Good baseline, but server must own all calculation |
| Timezone | UTC storage plus `outlets.timezone` defaulting to `Asia/Jakarta` | `BusinessClock` owns outlet-local query boundaries and document dates |
| Idempotency | middleware with response snapshot | Needs atomic reservation and more coverage |
| Audit | static direct helper | Needs transactional service/context |
| Flutter API | Dio + bearer token + typed exceptions | Needs timeout/request-id/observability policy |

## 10. Anti-Pattern Inventory To Remove Gradually

Do not remove all at once. Each needs tests first.

| Anti-pattern | Why dangerous | Removal approach |
| --- | --- | --- |
| Client-supplied `unit_price`/rounding | Price tampering | Server quote snapshot |
| Client-supplied cashier/device/shift authority | Actor spoofing | Terminal session and tuple validation |
| Read-execute-insert idempotency | Race/double effect | Atomic reservation |
| Audit after commit | Missing audit on committed effect | Audit inside DB transaction |
| Controller domain logic | Inconsistent policy/locks/audit | Service extraction |
| God session notifier | Cross-feature regression | Feature provider extraction |
| One-record checkout retry buffer | Indeterminate checkout only | Lookup by idempotency key before another sale; manual retry keeps the same key |
| Local/server order split | Conflicting order lifecycle | Server parked order aggregate |
| Missing outlet-local lifecycle coverage | Shift expiry/store lifecycle can still drift | Extend `BusinessClock` as those lifecycle features are introduced |
| Legacy prompt assumptions | Wrong architecture generation | Reconciled task prompts only |

## 11. Data/ERD-Level Summary

High-level current domain graph:

```text
Business
  |- Outlet -- Device
  |- User
  |- ProductCategory -- Product
  |- Customer
  |- ShiftSession -- CashMovement
  |                -- Transaction -- TransactionItem
  |                                  -- Payment
  |                -- StockMovement
  |- InventoryPurchase -- InventoryPurchaseItem
  |                       -- StockMovement
  |- AttendanceRecord
  |- PaymentMethodConfig
  |- AuditLog
  |- IdempotencyKey
  |- Subscription -- Plan
```

Target additions or hardening areas:

```text
Business
  |- OutletSetting / SecuritySetting / ReceiptSetting
  |- TerminalSession
  |- DocumentCounter
  |- FinalQuote / PromotionQuote
  |- ParkedOrder -- ParkedOrderItem -- Lease
  |- Refund
  |- InventoryDocument family: Purchase, Count, Waste, Transfer
  |- StoreState
```

## 12. Build/Test Baseline

Known from the assessment snapshot:

- Backend `php artisan test` previously passed: 39 tests, 425 assertions.
- Flutter `flutter analyze` + `flutter test` did not complete within the prior audit timeout. Do not infer correctness from timeout.
- Source file sizes show Flutter maintainability pressure:
  - `operations_screen.dart`: 2,754 lines.
  - `pos_dialogs.dart`: 1,906 lines.
  - `payment_screen.dart`: 1,490 lines.
  - `NojposSessionNotifier`: 1,102 lines.

Minimum gate before production-like pilot:

```text
apps/backend: php artisan test
apps/backend: php artisan route:list --path=api/v1
apps/cashier: flutter analyze
apps/cashier: flutter test
apps/cashier: flutter build apk --debug
```

For P0 work, add targeted tests for:

- tampered price/rounding,
- same-tenant actor spoofing,
- revoked/unknown device,
- idempotency parallel requests,
- payment double confirm/overpay,
- shift close double submit,
- audit rollback,
- outlet timezone boundary.

## 13. Admin Web Status

Admin web is a future surface, not an implemented current surface.

Current facts:

- No `apps/web` folder.
- No `apps/admin` folder.
- No Next.js app package in the app tree.
- `apps/backend/package.json` is Laravel/Vite tooling.

Recommendation:

- Do not start admin web until P0 backend contracts are green.
- When web is approved, create an explicit architecture decision first: app location, framework version, auth/session model, API consumption, shared design tokens, route protection, test/build command, and deployment target.

## 14. Documentation Update Triggers

Update this architecture map whenever any of these change:

- new app surface is added (`apps/web`, shared package, worker, etc.),
- API route contract changes,
- idempotency/audit/session model changes,
- checkout quote contract changes,
- payment/shift state model changes,
- migrations add core domain tables,
- Flutter routing or provider architecture changes,
- production build/test command changes.

## 15. Architecture Decision Summary

Current recommended architecture direction:

1. Keep current two-app workspace: Flutter cashier + Laravel API.
2. Do not add Next.js admin web until backend P0 is green.
3. Do not add Drift/general offline DB.
4. Harden backend contracts before UI expansion.
5. Move sensitive writes to services with idempotency, row locks, policies, and transactional audit.
6. Migrate Flutter to server-owned quote/checkout contracts.
7. Refactor Flutter state/UI incrementally after behavior is protected by tests.

## 16. Final Architecture Verdict

The repository has a valid starting structure, but the architecture is currently **baseline, not hardened**. The next architectural work should focus on correctness contracts and service boundaries, not on adding new product surfaces. The biggest mistake would be to build a polished admin web or broad feature set before the terminal/payment/shift/checkout contract is safe.
