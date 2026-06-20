# NojPOS Flutter Cashier Guide

Read [root guide](../../AGENTS.md), [PRD v2](../../NOJPOS_POS_FLUTTER_BACKEND_PRD.md), [architecture](../../docs/ARCHITECTURE.md), [analysis findings](../../docs/ANALYSIS_FINDINGS.md), and [improvement plan](../../docs/IMPROVEMENT_PLAN.md) before editing. For the current repo, Wave 0 -> Wave 1 -> Wave 2 is mandatory before broad cashier feature expansion.

## Run And Verify

```powershell
cd apps/cashier
C:\flutter\bin\flutter.bat pub get
C:\flutter\bin\flutter.bat analyze
C:\flutter\bin\flutter.bat test
C:\flutter\bin\flutter.bat build apk --debug
C:\flutter\bin\flutter.bat install -d emulator-5554 --use-application-binary build\app\outputs\flutter-apk\app-debug.apk
```

Use the last command only when the emulator exists. The Android-emulator API base URL is `http://10.0.2.2:8000/api/v1`; override it only with `--dart-define=API_BASE_URL=...`.

## Local Structure

```text
lib/app/                 app shell, router, theme, session state
lib/core/network/        Dio client and API errors
lib/core/storage/        secure token storage
lib/core/outbox/         checkout retry mechanism only
lib/core/printing/       receipt printer adapter
lib/core/share/          OS share adapter
lib/features/<feature>/  pages, widgets, providers, repositories, models
lib/shared/models/       DTOs used by multiple features
test/                    unit/provider/widget tests
integration_test/        device/runtime flows
```

## MUST Rules

- Current cashier work is Flutter terminal only. Do not add admin web, Next.js, Drift, SQLite, Hive, generic sync, or any general offline database from cashier tasks.
- Admin web remains paused until P0 backend contracts are green and explicitly approved.
- Use Riverpod. A widget MUST call a feature notifier/provider, which calls a repository interface, which calls `ApiClient`.
- Use `ApiClient` from `lib/core/network/api_client.dart`; it supplies the envelope, bearer token, and typed failures.
- Keep money as `int`; format with the POS formatter only in UI. Do not calculate total/tax/service/rounding/promotion or stock locally.
- Treat server quote and transaction response as authoritative.
- Send an idempotency key only for commands that the backend contract declares idempotent. Preserve the same key for retry of the same command body.
- Clear PIN/password input after submit, failure, background, and dispose. Never log PIN, token, or PII.
- Do not add Drift, SQLite, Hive, generic sync, or any offline database. `core/database/` and `core/sync/` are forbidden.
- If a task touches checkout, shift, payment, stock, auth, device, audit, or retry behavior, use `docs/EXECUTION_CHECKLIST.md` and verify backend contract coverage, not just widget/provider behavior.

```dart
// DO: repository owns the HTTP call.
final quote = await ref.read(transactionRepositoryProvider).quoteTransaction(draft);

// DON'T: widget makes HTTP or trusts a computed total.
final total = cart.fold(0, (sum, item) => sum + item.total);
await Dio().post('/transactions', data: {'grand_total': total});
```

## Feature Pattern

For a new feature, create:

```text
features/<feature>/
  models/          feature DTO/view state when not shared
  repositories/    abstract interface + Api implementation + provider
  providers/       focused Notifier/AsyncNotifier
  pages/           route-level screens
  widgets/         presentation-only widgets
```

Keep route definitions in `lib/app/router/app_router.dart`. Reuse `shared/models` only after a DTO is genuinely cross-feature.

## Checkout, Shift, And Outbox

The intended PRD behavior is a single bounded retry for an indeterminate checkout. The current persistent multi-item `CheckoutOutbox` is a **fix-forward exception** documented in `docs/ANALYSIS_FINDINGS.md`; do not extend it to shifts, cash movements, inventory, attendance, settings, or other writes.

Do not add work to `NojposSessionNotifier` unless it is truly session-wide. It is already a large cross-feature coordinator. Prefer a feature controller and let the session expose only authenticated terminal context.

The terminal must render server-verified outlet, device, cashier, shift, and policy state. Client-held IDs are not proof of authorization.

## Common Pitfalls

- `operations_screen.dart`, `pos_dialogs.dart`, `payment_screen.dart`, and `NojposSessionNotifier` are large. Extract one tested slice at a time; never rewrite them wholesale.
- `LocalOrderRepository` is temporary local display state, not authority for persisted orders. New parked-order work belongs to the server order contract.
- Do not show a fixed `Online` status; connection state needs a real backend heartbeat per PRD §5.9.
- Do not prefill physical opening or closing cash with a sample business amount.
- Do not hide a button and assume the backend is protected.

## Tests

- Repository: request serialization, envelope mapping, conflict/error mapping.
- Provider: state transition, retry/idempotency-key retention, no duplicate submit.
- Widget: controls, field validation, loading/error/empty/forbidden states.
- Integration: login -> outlet -> PIN -> shift -> checkout and recovery flows.

For money/stock/shift UI, test both normal result and server rejection. A green widget test alone does not prove tenant or actor safety; backend feature tests are required.

## Definition Of Done

- `flutter analyze`, `flutter test`, and debug APK build pass.
- New widgets have stable layout on phone and tablet and use existing theme/icons.
- No widget calls Dio or directly persists auth data.
- No client-side financial calculation or offline queue added.
- Error state is actionable and does not expose secrets/PII.
- Any changed API contract is reflected in repository tests and PRD/task docs.
