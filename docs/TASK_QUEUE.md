# NojPOS Task Queue

Dokumen ini adalah queue kerja aktif. Kerjakan dari atas ke bawah. Jangan mengambil task berikutnya sebelum task saat ini selesai atau diblokir dengan alasan jelas.

## Queue Rules

- Hanya satu task boleh `[IN PROGRESS]`.
- Task harus mengikuti `PRDPOSJA.md` dan `docs/SINGLE_AGENT_WORKFLOW.md`.
- Jangan mengubah area di luar allowed scope task.
- Update status task, story, dan bug tracker setelah selesai.
- Action sensitif tetap disabled/gated sampai fase akhir.

## Current Active Lane

| Field | Value |
| --- | --- |
| Mode | Single active coding agent |
| Current phase | Phase 3 — Tenant Admin Read-only Pages |
| Current task | `CHK-03` done; inventory read-only checkpoint pushed |
| Next implementation target | `WEB-04A` Catalog/products/categories read-only contract |
| Do not start yet | Tenant pages beyond dashboard, Super Admin writes, payment/broadcast/impersonation real |

## Phase 0 — Clean Baseline And Planning

### DOC-01 — Switch docs to single-agent workflow

Status: `[DONE]`
Priority: P0
Area: Docs
Depends on: none

Goal:
- Replace old parallel workflow with safer single-agent sequential workflow.
- Create this final task queue.
- Remove obsolete workflow docs/scripts that can confuse future AI runs.
- Commit clean baseline to GitHub.

Allowed scope:
- `AGENTS.md`
- `apps/*/AGENTS.md`
- `docs/**`
- `scripts/**` for obsolete workflow setup script cleanup only

Subtasks:
- [x] Create `docs/SINGLE_AGENT_WORKFLOW.md`.
- [x] Create `docs/TASK_QUEUE.md`.
- [x] Update `docs/README.md`.
- [x] Update `docs/STORY_PROGRESS.md`.
- [x] Update root/package `AGENTS.md` references.
- [x] Remove obsolete workflow docs.
- [x] Remove obsolete setup scripts.
- [ ] Commit and push clean baseline if Git remote/auth is available.

Acceptance criteria:
- Active docs point to single-agent sequential workflow as the default process.
- `docs/TASK_QUEUE.md` clearly shows the next sequential work.
- Git working tree is committed as a fresh baseline, or blocker explains why not.

Validation:
- Docs read check.
- `git status --short` before and after commit.

---

### DOC-02 — Review PRD and story consistency

Status: `[READY AFTER DOC-01]`
Priority: P0
Area: Docs
Depends on: `DOC-01`

Goal:
- Re-read `PRDPOSJA.md` against current story files.
- Ensure all stories have dependency, task, subtask, acceptance criteria, and blocker notes.
- Mark only one next implementation task as `[READY]`.

Allowed scope:
- `PRDPOSJA.md`
- `docs/STORY_*.md`
- `docs/BUG_TRACKER.md`
- `docs/TASK_QUEUE.md`

Subtasks:
- [ ] Check Backend story order.
- [ ] Check Mobile Kasir story order.
- [ ] Check Web Admin story order.
- [ ] Check Integration story order.
- [ ] Move unclear items to `[BLOCKED]`.

Acceptance criteria:
- Story docs do not encourage jumping directly to sensitive action or cross-area integration.
- Next task remains backend dashboard summary, not web integration.

Validation:
- Docs read check.

## Phase 1 — Backend Dashboard Summary

### BE-04A — Finalize dashboard summary API contract

Status: `[DONE]`
Priority: P0
Area: Backend/API Contract
Depends on: `DOC-01`

Goal:
- Finalize contract for owner/admin dashboard summary.
- Ensure response covers current `/dashboard` UI without requiring multiple client calls.

Allowed scope:
- `docs/API_CONTRACTS/DASHBOARD_SUMMARY.md`
- `docs/STORY_BACKEND.md`
- `docs/STORY_INTEGRATION.md`
- `docs/BUG_TRACKER.md`
- `docs/TASK_QUEUE.md`

Subtasks:
- [x] Confirm endpoint path and method.
- [x] Define response DTO for KPI, alerts, sales chart, payment methods, top products, low stock, recent transactions, cashiers, branches.
- [x] Define role/tenant/outlet authorization.
- [x] Define empty/error/forbidden states.
- [x] Define calculation source and preview/estimate labels.

Acceptance criteria:
- Web admin can implement `/dashboard` without guessing response shape.
- No client-side final money/stock calculation required.

Validation:
- Docs read check.

---

### BE-04B — Implement `GET /api/v1/dashboard/summary`

Status: `[DONE]`
Priority: P0
Area: Backend
Depends on: `BE-04A`

Goal:
- Add a tenant-scoped dashboard summary endpoint for owner/admin web dashboard.

Allowed scope:
- `apps/backend/**`
- `docs/STORY_BACKEND.md`
- `docs/STORY_INTEGRATION.md`
- `docs/BUG_TRACKER.md`
- `docs/TASK_QUEUE.md`

Subtasks:
- [x] Add route under `/api/v1` with auth.
- [x] Add thin controller or report/dashboard controller method.
- [x] Add service/resource if needed.
- [x] Aggregate data server-side using authenticated business/outlet context.
- [x] Add feature tests for auth, tenant isolation, empty state, and response shape.
- [x] Update docs/story/bug tracker.

Acceptance criteria:
- Endpoint returns standard envelope `{ data, meta }`.
- No cross-tenant data leaks.
- Existing backend tests still pass.

Validation:
```bash
cd apps/backend
php artisan test
php artisan route:list --path=api/v1
```

## Phase 2 — Web Dashboard Read-only Integration

### WEB-01A — Decide and document web admin session strategy

Status: `[DONE]`
Priority: P0
Area: Integration/Web Docs
Depends on: `BE-04B`

Goal:
- Decide how browser session works before real API calls.

Decision:
- Next.js BFF/server-side route handler.
- Browser talks only to same-origin `/api/admin/*` routes.
- Backend token/session material stays behind HttpOnly sealed cookie or server-only session boundary.
- No auth secret storage in browser-readable storage.
- No Authorization/Bearer construction in UI components.

Acceptance criteria:
- [x] `docs/API_CONTRACTS/SESSION.md` defines final strategy.
- [x] Web implementation rules forbid browser-readable auth secret storage.
- [x] Next task is session/BFF adapter scaffold, not dashboard integration.

---

### WEB-01B — Implement Web Admin session/BFF adapter scaffold

Status: `[DONE]`
Priority: P0
Area: Web Admin / Integration
Depends on: `WEB-01A`, `BE-04B`

Goal:
- Implement the approved server-only session/BFF boundary before any dashboard data integration.

Allowed scope:
- `apps/web/src/lib/server/**`
- `apps/web/src/app/api/admin/**`
- `apps/web/src/lib/auth/**` only if needed for shared safe types/tests
- `apps/web/src/lib/api/**` only if needed for server adapter envelope/types/tests
- `docs/STORY_WEB_ADMIN.md`
- `docs/STORY_INTEGRATION.md`
- `docs/BUG_TRACKER.md`
- `docs/TASK_QUEUE.md`

Subtasks:
- [x] Add server-only session cookie/sealing helper with existing dependencies only.
- [x] Add server-only Laravel backend client.
- [x] Add Next route handler `POST /api/admin/auth/login`.
- [x] Add Next route handler `POST /api/admin/auth/logout`.
- [x] Add Next route handler `GET /api/admin/session`.
- [x] Reject cashier from Web Admin.
- [x] Keep dashboard fixture integration untouched.
- [x] Add tests and boundary scan for no raw token in UI/client code.

Validation:
```bash
cd apps/web
npm run typecheck
npm run lint
npm test
npm run build
npx playwright test --list
```

Boundary scan:
```bash
rg -n "fetch\(|XMLHttpRequest|sessionStorage|localStorage|Authorization|Bearer |NEXT_PUBLIC_API_BASE_URL|@/lib/api|@/lib/auth" src/app src/components src/fixtures --glob '!*.test.ts' --glob '!*.test.tsx' || true
```

---

### WEB-01C — Integrate `/dashboard` read-only data

Status: `[DONE]`
Priority: P1
Area: Web Admin
Depends on: `WEB-01B`, `BE-04B`

Goal:
- Replace dashboard fixture with server data behind approved API/session boundary.

Allowed scope:
- `apps/web/**`
- `docs/STORY_WEB_ADMIN.md`
- `docs/STORY_INTEGRATION.md`
- `docs/BUG_TRACKER.md`
- `docs/TASK_QUEUE.md`

Subtasks:
- [x] Add dashboard summary adapter through `/api/admin/dashboard/summary` according to approved BFF strategy.
- [x] Add error, empty, forbidden, and session-required state.
- [x] Keep preview fallback only with clear label when session/backend is unavailable.
- [x] Keep action buttons static/disabled.
- [x] Run full web validation.

Validation:
```bash
cd apps/web
npm run typecheck
npm run lint
npm test
npm run build
npx playwright test --list
```

Boundary scan:
```bash
rg -n "fetch\(|XMLHttpRequest|sessionStorage|localStorage|Authorization|Bearer |NEXT_PUBLIC_API_BASE_URL|@/lib/api|@/lib/auth" src/app src/components src/fixtures --glob '!*.test.ts' --glob '!*.test.tsx' || true
```

## Phase 3 — Tenant Admin Read-only Pages

### WEB-02A — Transactions read-only contract

Status: `[DONE]`
Priority: P1
Area: Web Admin / API Contract
Depends on: dashboard/session read-only success

Goal:
- Finalize transactions list read-only contract before inventory/catalog.

Subtasks:
- [x] Finalize transaction list contract in `docs/API_CONTRACTS/TRANSACTIONS_READ.md`.
- [x] Backend route shape check for `GET /api/v1/transactions`.
- [x] Document BFF target `GET /api/admin/transactions`.
- [x] Document loading/error/empty/forbidden states for implementation.

---

### WEB-02B — Transactions BFF + page read-only integration

Status: `[DONE]`
Priority: P1
Area: Web Admin
Depends on: `WEB-02A`, dashboard/session read-only success

Goal:
- Integrate transactions list read-only through BFF before inventory/catalog.

Subtasks:
- [x] Add BFF route `GET /api/admin/transactions`.
- [x] Add server-only transactions mapping helper.
- [x] Integrate `/transactions` page with session/error/empty/forbidden state.
- [x] Keep transaction actions disabled.
- [x] Run full web validation and boundary scan.

---

### CHK-02 — Transactions read-only checkpoint

Status: `[DONE]`
Priority: P1
Area: Checkpoint
Depends on: `WEB-02B`

Goal:
- Commit and push transactions read-only contract, BFF route, page integration, tests, and docs.

---

### WEB-03A — Inventory read-only contract

Status: `[DONE]`
Priority: P1
Area: Web Admin / API Contract
Depends on: transactions read-only checkpoint

Goal:
- Finalize inventory read-only contract before BFF/page integration.

Subtasks:
- [x] Document backend routes `GET /api/v1/inventory`, `/inventory/movements`, `/inventory/transfers`, and `/inventory/transfers/in-transit`.
- [x] Document BFF targets `/api/admin/inventory` and `/api/admin/inventory/movements`.
- [x] Document query whitelist, response shape, field privacy, tenant/outlet isolation, empty/error state, and WEB-03B acceptance criteria.

---

### WEB-03B — Inventory BFF + page read-only integration

Status: `[DONE]`
Priority: P1
Area: Web Admin
Depends on: `WEB-03A`, transactions read-only checkpoint

Goal:
- Integrate inventory list read-only through BFF before catalog.

Subtasks:
- [x] Add BFF route `GET /api/admin/inventory`.
- [x] Add server-only inventory mapping helper.
- [x] Integrate `/inventory` page with session/error/empty/forbidden state.
- [x] Keep stock adjustment/purchase/count/waste/transfer actions disabled.
- [x] Run full web validation and boundary scan.

---

### CHK-03 — Inventory read-only checkpoint

Status: `[DONE]`
Priority: P1
Area: Checkpoint
Depends on: `WEB-03B`

Goal:
- Commit and push inventory read-only contract, BFF route, page integration, tests, and docs.

---

### WEB-04A — Catalog/products/categories read-only contract

Status: `[READY]`
Priority: P1
Area: Web Admin
Depends on: inventory read-only checkpoint

Goal:
- Integrate catalog read-only.

## Phase 4 — Deferred Safe Writes

### WEB-04A — Low-risk CRUD planning

Status: `[DEFERRED]`
Priority: P2
Area: Integration
Depends on: Tenant admin read-only pages stable

Goal:
- Plan create/update for products, categories, customers, settings.
- No money/stock/shift/payment sensitive action yet.

## Phase 5 — Platform Super Admin Read-only

### PLATFORM-01A — Platform overview read-only contract

Status: `[DEFERRED]`
Priority: P2
Area: Platform Backend/Web
Depends on: tenant admin read-only baseline stable

Goal:
- Add read-only contracts for platform summary, toko, paket, and partial billing.

## Phase 6 — Sensitive Actions

### SAFE-01A — Sensitive action framework

Status: `[DEFERRED]`
Priority: P3
Area: Backend/Web/Mobile
Depends on: audit, policy, reason, idempotency, approval design

Goal:
- Create framework before any real refund, void, payment retry, access bantuan, suspend, ban, reset, or broadcast action.
