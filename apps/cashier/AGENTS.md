# NojPOS Flutter Cashier Guide

Read [root guide](../../AGENTS.md), [PRDPOSJA](../../PRDPOSJA.md), [docs index](../../docs/README.md), [story progress](../../docs/STORY_PROGRESS.md), [mobile story](../../docs/STORY_MOBILE_KASIR.md), [integration story](../../docs/STORY_INTEGRATION.md), and [bug tracker](../../docs/BUG_TRACKER.md) before editing.

Mobile cashier work must follow story order. Do not add broad features, offline database, or sensitive shortcuts before backend and integration dependencies are ready.

## Run And Verify

```powershell
cd apps/cashier
C:\flutter\bin\flutter.bat pub get
C:\flutter\bin\flutter.bat analyze
C:\flutter\bin\flutter.bat test
C:\flutter\bin\flutter.bat build apk --debug
C:\flutter\bin\flutter.bat install -d emulator-5554 --use-application-binary build\app\outputs\flutter-apk\app-debug.apk
```

Use the install command only when emulator exists. Android emulator API base URL is `http://10.0.2.2:8000/api/v1`; override only with `--dart-define=API_BASE_URL=...`.

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

- Flutter mobile is cashier terminal only. Do not add admin web, Next.js, or platform admin work here.
- Use Riverpod. UI calls provider/notifier; provider calls repository; repository calls `ApiClient`.
- Widgets must not create Dio, persist tokens, or calculate final financial totals.
- Keep money as `int` rupiah; format only at presentation boundary.
- Treat server quote and transaction response as authoritative.
- Send idempotency key only for backend-declared idempotent commands.
- Preserve same idempotency key for retry of the same command body.
- Clear PIN/password input after submit, failure, background, and dispose.
- Never log PIN, token, password, full PII, receipt content, or payment reference.
- Do not add Drift, SQLite, Hive, generic sync, or any offline database.
- Checkout retry is bounded to one in-flight checkout attempt only; do not extend it to shift, inventory, attendance, settings, or other writes.

## Feature Pattern

For a new feature, create:

```text
features/<feature>/
  models/          feature DTO/view state when not shared
  repositories/    abstract interface + API implementation + provider
  providers/       focused Notifier/AsyncNotifier
  pages/           route-level screens
  widgets/         presentation-only widgets
```

Keep route definitions in `lib/app/router/app_router.dart`. Reuse `shared/models` only after a DTO is genuinely cross-feature.

## Current Known Gaps

Track and update these in `../../docs/BUG_TRACKER.md`:

- Device/emulator QA is needed for lock/unlock and PIN switch.
- Checkout retry/recovery needs end-to-end verification.
- Print/share needs hardware QA after target printer is selected.
- Verify all final totals come from server, not client calculation.

## Tests

- Repository: request serialization, envelope mapping, conflict/error mapping.
- Provider: state transition, retry/idempotency-key retention, no duplicate submit.
- Widget: controls, field validation, loading/error/empty/forbidden states.
- Integration: login -> outlet -> PIN -> shift -> checkout -> payment -> receipt and recovery flows.

For money/stock/shift UI, test both normal result and server rejection. A green widget test alone does not prove tenant or actor safety; backend feature tests are required.

## Definition Of Done

Run:

```powershell
C:\flutter\bin\flutter.bat analyze
C:\flutter\bin\flutter.bat test
C:\flutter\bin\flutter.bat build apk --debug
```

Done requires:

- No widget calls Dio directly.
- No client-side financial calculation or general offline queue added.
- New widgets are responsive on phone/tablet.
- Backend contract coverage checked when touching checkout, shift, payment, stock, auth, device, audit, or retry.
- Story docs and bug tracker updated.

Final report format:

```text
Built: <changed>
Verified: <commands and results>
Blocker: <precise blocker or none>
```
