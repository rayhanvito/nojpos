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
| Current phase | Phase 0 — docs clean baseline |
| Current task | `DOC-01` done; commit/push in progress |
| Next implementation target | `BE-04A` dashboard summary contract/API |
| Do not start yet | Web API integration, Super Admin writes, payment/broadcast/impersonation real |

## Phase 0 — Clean Baseline And Planning

### DOC-01 — Switch docs to single-agent workflow

Status: `[DONE]`
Priority: P0
Area: Docs
Depends on: none

Goal:
- Replace multi-agent parallel workflow with safer single-agent sequential workflow.
- Create this final task queue.
- Remove obsolete multi-agent docs/scripts that can confuse future AI runs.
- Commit clean baseline to GitHub.

Allowed scope:
- `AGENTS.md`
- `apps/*/AGENTS.md`
- `docs/**`
- `scripts/**` for obsolete multi-agent setup script cleanup only

Subtasks:
- [x] Create `docs/SINGLE_AGENT_WORKFLOW.md`.
- [x] Create `docs/TASK_QUEUE.md`.
- [x] Update `docs/README.md`.
- [x] Update `docs/STORY_PROGRESS.md`.
- [x] Update root/package `AGENTS.md` references.
- [x] Remove obsolete multi-agent docs.
- [x] Remove obsolete setup worktree scripts.
- [ ] Commit and push clean baseline if Git remote/auth is available.

Acceptance criteria:
- No active docs reference multi-agent workflow, parallel agents, or worktree setup as the default process.
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

Status: `[READY AFTER DOC-01]`
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
- [ ] Confirm endpoint path and method.
- [ ] Define response DTO for KPI, alerts, sales chart, payment methods, top products, low stock, recent transactions, cashiers, branches.
- [ ] Define role/tenant/outlet authorization.
- [ ] Define empty/error/forbidden states.
- [ ] Define calculation source and preview/estimate labels.

Acceptance criteria:
- Web admin can implement `/dashboard` without guessing response shape.
- No client-side final money/stock calculation required.

Validation:
- Docs read check.

---

### BE-04B — Implement `GET /api/v1/dashboard/summary`

Status: `[READY AFTER BE-04A]`
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
- [ ] Add route under `/api/v1` with auth.
- [ ] Add thin controller or report/dashboard controller method.
- [ ] Add service/resource if needed.
- [ ] Aggregate data server-side using authenticated business/outlet context.
- [ ] Add feature tests for auth, tenant isolation, empty state, and response shape.
- [ ] Update docs/story/bug tracker.

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

Status: `[BLOCKED]`
Priority: P0
Area: Integration/Web
Depends on: product owner decision

Goal:
- Decide how browser session works before real API calls.

Options to decide:
- HttpOnly cookie / Sanctum SPA cookie.
- Next.js BFF/server-side adapter.
- Token behavior for local dev only.

Acceptance criteria:
- `docs/API_CONTRACTS/SESSION.md` defines final strategy.
- Web implementation does not use `localStorage` or `sessionStorage` for tokens.

---

### WEB-01B — Integrate `/dashboard` read-only data

Status: `[READY AFTER WEB-01A AND BE-04B]`
Priority: P1
Area: Web Admin
Depends on: `WEB-01A`, `BE-04B`

Goal:
- Replace dashboard fixture with server data behind approved API/session boundary.

Allowed scope:
- `apps/web/**`
- `docs/STORY_WEB_ADMIN.md`
- `docs/STORY_INTEGRATION.md`
- `docs/BUG_TRACKER.md`
- `docs/TASK_QUEUE.md`

Subtasks:
- [ ] Add typed data adapter according to approved session strategy.
- [ ] Add loading, error, empty, forbidden state.
- [ ] Keep preview fallback only if explicitly documented.
- [ ] Keep action buttons static/disabled.
- [ ] Run full web validation.

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

### WEB-02A — Transactions read-only page integration

Status: `[READY AFTER WEB-01B]`
Priority: P1
Area: Web Admin
Depends on: dashboard/session read-only success

Goal:
- Integrate transactions list read-only before inventory/catalog.

Subtasks:
- [ ] Finalize transaction list contract.
- [ ] Backend route shape check.
- [ ] Web read-only integration.
- [ ] Loading/error/empty/forbidden state.

---

### WEB-02B — Inventory read-only page integration

Status: `[READY AFTER WEB-02A]`
Priority: P1
Area: Web Admin
Depends on: transactions read-only success

Goal:
- Integrate inventory list read-only.

---

### WEB-02C — Catalog/products/categories read-only integration

Status: `[READY AFTER WEB-02B]`
Priority: P1
Area: Web Admin
Depends on: inventory read-only success

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
