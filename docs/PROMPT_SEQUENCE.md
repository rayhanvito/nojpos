# NojPOS Prompt Sequence — Single-Agent Copy-Paste Plan

Dokumen ini adalah daftar prompt operasional untuk mengendalikan implementasi NojPOS secara **pelan, aman, berurutan, dan tidak loncat-loncat**.

Gunakan aturan ini:

1. Jalankan **satu prompt saja** setiap sesi agent.
2. Jangan lanjut prompt berikutnya sebelum laporan akhir prompt sekarang jelas dan validasi lolos.
3. Jika prompt gagal validasi, buat prompt perbaikan kecil dulu; jangan loncat ke fitur baru.
4. Semua agent wajib baca `AGENTS.md`, `PRDPOSJA.md`, `docs/SINGLE_AGENT_WORKFLOW.md`, `docs/TASK_QUEUE.md`, story terkait, dan bug tracker.
5. Setiap prompt harus update docs/story/task queue sesuai hasil.
6. Setiap milestone besar harus dibuat checkpoint commit.
7. Fitur sensitif seperti refund, void, payment retry, shift close/open, stock adjustment, cash reconciliation, akses bantuan/impersonate, suspend/delete toko, reset password, ban user, dan broadcast pengumuman tetap deferred sampai framework audit/idempotency siap.

## Status Saat Dokumen Ini Dibuat

Sudah selesai dari percakapan sebelumnya:

- `BE-04A` — kontrak `GET /api/v1/dashboard/summary`.
- `BE-04B` — endpoint backend dashboard summary.
- `WEB-01A` — keputusan session strategy Web Admin.
- `WEB-01B` — scaffold Next.js BFF/session HttpOnly sealed cookie.
- `WEB-01C` — integrasi `/dashboard` read-only lewat BFF.

Next aman:

- `PROMPT 006` — checkpoint commit setelah dashboard integration.
- `PROMPT 007` — kontrak transaksi read-only.

---

# Phase A — Checkpoint Dashboard Integration

## PROMPT 006 — Checkpoint Commit After Dashboard Integration

```text
Kamu mengerjakan project NojPOS.

Mode kerja:
- Single-agent sequential workflow.
- Kerjakan hanya task checkpoint ini.
- Jangan implement fitur baru.
- Jangan refactor.
- Jangan ubah backend/web/mobile kecuali docs kecil jika status belum sesuai.
- Fokus memastikan state sekarang valid, lalu commit dan push.

Lokasi repo:
C:\laragon\www\nojpos_final_fix

Baca dulu:
1. AGENTS.md
2. PRDPOSJA.md
3. docs/SINGLE_AGENT_WORKFLOW.md
4. docs/TASK_QUEUE.md
5. docs/STORY_PROGRESS.md
6. docs/STORY_BACKEND.md
7. docs/STORY_WEB_ADMIN.md
8. docs/STORY_INTEGRATION.md
9. docs/BUG_TRACKER.md
10. docs/VALIDATION_CHECKLIST.md

Task ID:
CHK-01 — Checkpoint commit after dashboard backend + web integration

Tujuan:
Membuat commit bersih setelah milestone:
- BE-04B dashboard summary backend endpoint selesai.
- WEB-01A session strategy selesai.
- WEB-01B BFF/session scaffold selesai.
- WEB-01C dashboard read-only integration selesai.

Scope boleh dilakukan:
- Inspect git status.
- Jalankan validasi backend.
- Jalankan validasi web.
- Update docs/TASK_QUEUE.md atau docs/STORY_PROGRESS.md hanya jika status belum sesuai.
- Commit semua perubahan relevan.
- Push ke GitHub branch aktif.

Scope tidak boleh:
- Jangan ubah fitur runtime.
- Jangan ubah mobile cashier kecuali melaporkan file dirty yang sudah ada.
- Jangan membuat endpoint baru.
- Jangan membuat UI baru.
- Jangan menghapus file besar sembarangan.
- Jangan commit node_modules, .next, vendor, build cache, atau artifact lokal.

Langkah kerja:
1. Cek status repo:
   git status --short

2. Pastikan file noise tidak ikut commit:
   - node_modules
   - .next
   - vendor jika tidak seharusnya tracked
   - tsconfig.tsbuildinfo
   - storage/cache/log
   - file lokal legacy yang sudah di-ignore

3. Validasi backend dari apps/backend:
   php artisan test
   php artisan route:list --path=api/v1

4. Validasi web dari apps/web:
   npm run typecheck
   npm run lint
   npm test
   npm run build
   npx playwright test --list

5. Boundary scan web dari apps/web:
   rg -n "fetch\(|XMLHttpRequest|sessionStorage|localStorage|NEXT_PUBLIC_API_BASE_URL" src/app src/components src/fixtures --glob '!*.test.ts' --glob '!*.test.tsx' || true

6. Tambahan scan:
   Cari Authorization/Bearer di app/components/fixtures.
   Jika muncul di UI page/component, stop dan laporkan.

7. Update docs jika perlu:
   - docs/TASK_QUEUE.md: BE-04B, WEB-01A, WEB-01B, WEB-01C harus DONE.
   - docs/STORY_PROGRESS.md: dashboard backend + web read-only integration tercatat selesai.
   - docs/BUG_TRACKER.md: jangan hapus bug existing.

8. Commit:
   Message: checkpoint-dashboard-summary-bff-integration

9. Push:
   Push ke branch aktif. Jika push gagal karena auth/remote, laporkan jujur.

Jika validasi gagal:
- Jangan commit.
- Laporkan command yang gagal.
- Jelaskan error singkat.
- Jangan lanjut task lain.

Laporan akhir wajib:
1. Git status sebelum commit.
2. File utama yang masuk commit.
3. Hasil validasi backend.
4. Hasil validasi web.
5. Boundary scan.
6. Commit hash.
7. Push status.
8. Working tree setelah commit.
9. Next recommended task.
10. Blocker jika ada.
```

---

# Phase B — Tenant Admin Read-only Integration

## PROMPT 007 — Transactions Read-only Contract

```text
Kamu mengerjakan project NojPOS.

Mode kerja:
- Single-agent sequential workflow.
- Kerjakan hanya task kontrak ini.
- Jangan coding runtime dulu.
- Jangan ubah apps/web page.
- Jangan ubah apps/backend runtime.
- Jangan ubah apps/cashier.
- Fokus finalisasi kontrak transaksi read-only untuk Web Admin.

Lokasi repo:
C:\laragon\www\nojpos_final_fix

Baca dulu:
1. AGENTS.md
2. PRDPOSJA.md
3. docs/SINGLE_AGENT_WORKFLOW.md
4. docs/TASK_QUEUE.md
5. docs/STORY_WEB_ADMIN.md
6. docs/STORY_BACKEND.md
7. docs/STORY_INTEGRATION.md
8. docs/BUG_TRACKER.md
9. docs/API_CONTRACTS/README.md
10. apps/backend/routes/api.php
11. apps/web/AGENTS.md
12. apps/web/src/app jika route transaksi sudah ada

Task ID:
WEB-02A1 — Finalize Transactions Read-only Contract

Tujuan:
Membuat kontrak read-only untuk halaman transaksi Tenant Admin Web.

Scope boleh diedit:
- docs/API_CONTRACTS/TRANSACTIONS_READ.md
- docs/API_CONTRACTS/README.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_BACKEND.md jika ada backend gap
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md
- docs/BUG_TRACKER.md jika ada blocker

Scope tidak boleh diedit:
- apps/backend runtime
- apps/web runtime
- apps/cashier
- package/dependency

Yang harus dianalisis:
1. Route backend nyata terkait transaksi.
2. Controller/resource/model transaksi yang sudah ada.
3. Halaman web admin transaksi jika sudah ada.
4. Kebutuhan owner/admin:
   - daftar transaksi
   - filter tanggal
   - filter outlet
   - filter status
   - filter metode pembayaran
   - search kode transaksi
   - ringkasan total read-only
   - detail ringkas transaksi tanpa PII/payment ref sensitif

Output kontrak:
1. Endpoint backend yang dipakai atau endpoint baru yang dibutuhkan.
2. Endpoint BFF yang akan dibuat:
   GET /api/admin/transactions
3. Role:
   owner/admin only.
4. Query params:
   - date_from
   - date_to
   - outlet_id
   - status
   - payment_method
   - q
   - page
   - per_page
5. Response shape:
   - meta pagination
   - summary
   - rows
   - data_notes
6. Field rows:
   - id
   - code
   - outlet_name
   - cashier_name
   - transaction_time
   - payment_method_label
   - status
   - total_amount
   - item_count
   - can_view_detail false/true based on policy
7. Empty state.
8. Error state 401/403/422/500.
9. Tenant isolation.
10. Acceptance criteria untuk implementasi berikutnya.

Update docs:
- Mark WEB-02A1 done jika kontrak selesai.
- Next task WEB-02A2 Backend route check/adapter if needed atau WEB-02A3 BFF+page integration jika backend route cukup.

Validasi:
- Docs read check.
- Tidak perlu test/build runtime.

Laporan akhir:
1. File diubah.
2. Kontrak final.
3. Backend route existing atau gap.
4. Next task.
5. Blocker.
```

---

## PROMPT 008 — Transactions Backend Readiness Check and Tests

```text
Kamu mengerjakan project NojPOS.

Mode kerja:
- Single-agent sequential workflow.
- Kerjakan hanya backend readiness transaksi read-only.
- Jangan integrasi web dulu.
- Jangan ubah mobile.
- Jangan buat write action transaksi.
- Jangan buat refund/void/reprint/export real.

Lokasi repo:
C:\laragon\www\nojpos_final_fix

Baca dulu:
1. AGENTS.md
2. PRDPOSJA.md
3. docs/API_CONTRACTS/TRANSACTIONS_READ.md
4. docs/STORY_BACKEND.md
5. docs/STORY_WEB_ADMIN.md
6. docs/STORY_INTEGRATION.md
7. docs/TASK_QUEUE.md
8. docs/BUG_TRACKER.md
9. apps/backend/AGENTS.md
10. apps/backend/routes/api.php

Task ID:
WEB-02A2 — Transactions Backend Readiness Check

Tujuan:
Memastikan backend punya endpoint transaksi read-only yang sesuai kontrak. Jika sudah ada, tambahkan test/penyesuaian kecil. Jika belum cukup, implement endpoint read-only minimal.

Scope boleh diedit:
- apps/backend/routes/api.php
- apps/backend/app/Http/Controllers/Api/V1/**
- apps/backend/app/Http/Requests/**
- apps/backend/app/Services/**
- apps/backend/app/Http/Resources/**
- apps/backend/tests/**
- docs/STORY_BACKEND.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md
- docs/BUG_TRACKER.md

Scope tidak boleh diedit:
- apps/web
- apps/cashier
- package/dependency
- sensitive write flows

Requirement:
1. Endpoint read-only transaksi harus tenant-scoped.
2. Role owner/admin boleh akses.
3. Cashier forbidden untuk web admin view jika kontrak begitu.
4. Filter harus divalidasi.
5. Pagination harus stabil.
6. Response tidak boleh expose token, PIN, full PII, payment gateway ref sensitif.
7. Tidak ada write database.
8. Tidak trigger refund/void/reprint/export.

Test wajib:
- owner/admin can list transactions.
- cashier forbidden.
- guest 401.
- cross-tenant data excluded.
- invalid filter 422.
- empty state stable.
- pagination shape stable.

Validasi:
cd apps/backend
php artisan test
php artisan route:list --path=api/v1

Docs:
- Update TASK_QUEUE, STORY_BACKEND, STORY_WEB_ADMIN.
- Next task harus WEB-02A3 Transactions BFF + page integration.

Laporan akhir:
1. File diubah.
2. Endpoint backend dipakai/dibuat.
3. Test dibuat.
4. Hasil validasi.
5. Security tenant note.
6. Next task.
```

---

## PROMPT 009 — Transactions BFF and Page Integration

```text
Kamu mengerjakan project NojPOS.

Mode kerja:
- Single-agent sequential workflow.
- Kerjakan hanya integrasi transaksi read-only.
- Jangan ubah backend.
- Jangan ubah mobile.
- Jangan membuat action real: refund, void, reprint, export.
- Semua action sensitif tetap disabled/gated.

Lokasi repo:
C:\laragon\www\nojpos_final_fix

Baca dulu:
1. AGENTS.md
2. PRDPOSJA.md
3. docs/API_CONTRACTS/SESSION.md
4. docs/API_CONTRACTS/TRANSACTIONS_READ.md
5. docs/STORY_WEB_ADMIN.md
6. docs/STORY_INTEGRATION.md
7. docs/TASK_QUEUE.md
8. docs/BUG_TRACKER.md
9. apps/web/AGENTS.md
10. apps/web/src/lib/server/backend-client.ts
11. apps/web/src/lib/server/admin-session.ts
12. apps/web/src/app/dashboard/page.tsx sebagai pattern server integration

Task ID:
WEB-02A3 — Transactions BFF + Page Integration

Tujuan:
Mengintegrasikan halaman transaksi Tenant Admin Web secara read-only melalui BFF.

Scope boleh diedit:
- apps/web/src/app/api/admin/transactions/route.ts
- apps/web/src/lib/server/transactions.ts
- apps/web/src/app/transactions/** atau route transaksi existing
- apps/web/src/components/** jika komponen list diperlukan
- apps/web/src/fixtures/preview/** hanya fallback berlabel jelas
- apps/web/src/**/*.test.ts
- docs/STORY_WEB_ADMIN.md
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md
- docs/BUG_TRACKER.md

Scope tidak boleh diedit:
- apps/backend
- apps/cashier
- package/dependency

Requirement:
1. Buat BFF route GET /api/admin/transactions.
2. BFF membaca session sealed cookie.
3. Only owner/admin.
4. Forward whitelist query params saja.
5. UI page server-side memakai helper server-only, bukan direct Laravel.
6. Loading/error/empty/forbidden/session-required state ada.
7. Preview fallback boleh hanya jika dilabel jelas.
8. Action refund/void/reprint/export disabled atau preview-only.
9. Tidak ada token di props/client/log.

Test:
- BFF 401 tanpa session.
- BFF 403 non owner/admin.
- Query whitelist.
- Response tidak mengandung token.
- Page render empty/fallback.
- Route manifest jika route baru.

Validasi dari apps/web:
npm run typecheck
npm run lint
npm test
npm run build
npx playwright test --list
rg -n "fetch\(|XMLHttpRequest|sessionStorage|localStorage|NEXT_PUBLIC_API_BASE_URL" src/app src/components src/fixtures --glob '!*.test.ts' --glob '!*.test.tsx' || true

Docs:
- WEB-02A3 DONE.
- Next task CHK-02 checkpoint transaksi.

Laporan akhir wajib:
1. File diubah.
2. Route BFF dibuat.
3. Cara page mengambil data.
4. Fallback/error state.
5. Test dan validasi.
6. Security note.
7. Next task.
```

---

## PROMPT 010 — Checkpoint Transactions Integration

```text
Kamu mengerjakan project NojPOS.

Task ID:
CHK-02 — Checkpoint after transactions read-only integration

Mode kerja:
- Jangan implement fitur baru.
- Validasi backend dan web.
- Commit dan push jika semua lolos.

Baca:
- AGENTS.md
- docs/TASK_QUEUE.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_BACKEND.md
- docs/STORY_INTEGRATION.md
- docs/BUG_TRACKER.md

Langkah:
1. git status --short
2. Backend validation jika backend berubah:
   cd apps/backend
   php artisan test
   php artisan route:list --path=api/v1
3. Web validation:
   cd apps/web
   npm run typecheck
   npm run lint
   npm test
   npm run build
   npx playwright test --list
4. Boundary scan web.
5. Update docs status kalau perlu.
6. Commit message:
   checkpoint-transactions-readonly-integration
7. Push branch aktif.

Jangan commit artifact/cache.
Jika validasi gagal, jangan commit.

Laporan akhir:
- status sebelum
- validasi
- commit hash
- push status
- next task: inventory read-only contract
```

---

## PROMPT 011 — Inventory Read-only Contract

```text
Kamu mengerjakan project NojPOS.

Task ID:
WEB-02B1 — Finalize Inventory Read-only Contract

Mode kerja:
- Dokumentasi kontrak saja.
- Jangan coding runtime.
- Jangan ubah web/backend/mobile runtime.

Baca:
- AGENTS.md
- PRDPOSJA.md
- docs/TASK_QUEUE.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_BACKEND.md
- docs/API_CONTRACTS/README.md
- apps/backend/routes/api.php
- apps/backend model/controller inventory terkait
- apps/web route inventory jika ada

Tujuan:
Kontrak halaman Inventory read-only untuk owner/admin.

Scope boleh:
- docs/API_CONTRACTS/INVENTORY_READ.md
- docs/API_CONTRACTS/README.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_BACKEND.md
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md
- docs/BUG_TRACKER.md

Isi kontrak:
1. Endpoint backend existing/needed.
2. BFF route: GET /api/admin/inventory.
3. Role owner/admin.
4. Query: outlet_id, q, stock_status, category_id, page, per_page.
5. Response: summary, rows, pagination, data_notes.
6. Rows: product id/name/category, outlet, stock_on_hand, low_stock_threshold, status, last_movement_at.
7. Empty/error states.
8. Tenant/outlet isolation.
9. Tidak boleh stock adjustment/write.
10. Acceptance criteria backend + web.

Validasi docs read.

Laporan akhir:
- file diubah
- kontrak final
- backend route gap
- next task
```

---

## PROMPT 012 — Inventory Backend Readiness Check

```text
Kamu mengerjakan project NojPOS.

Task ID:
WEB-02B2 — Inventory Backend Readiness Check

Mode kerja:
- Backend read-only only.
- Jangan web integration dulu.
- Jangan mobile.
- Jangan stock adjustment real.

Baca:
- apps/backend/AGENTS.md
- docs/API_CONTRACTS/INVENTORY_READ.md
- docs/STORY_BACKEND.md
- apps/backend/routes/api.php

Scope boleh:
- apps/backend routes/controllers/requests/services/resources/tests terkait inventory read-only
- docs story/task/bug

Requirement:
- endpoint/list inventory tenant-scoped.
- owner/admin allowed.
- cashier forbidden for web admin route if applicable.
- outlet filter validated.
- no cross-tenant product/stock.
- no write.
- no stock adjustment.

Test wajib:
- owner/admin list.
- guest 401.
- forbidden role 403.
- cross tenant excluded.
- outlet other business rejected.
- empty state.

Validasi:
cd apps/backend
php artisan test
php artisan route:list --path=api/v1

Update docs; next WEB-02B3.
```

---

## PROMPT 013 — Inventory BFF and Page Integration

```text
Kamu mengerjakan project NojPOS.

Task ID:
WEB-02B3 — Inventory BFF + Page Integration

Mode:
- Web only.
- Read-only.
- Jangan backend/mobile.
- Jangan stock adjustment real.

Baca:
- docs/API_CONTRACTS/INVENTORY_READ.md
- docs/API_CONTRACTS/SESSION.md
- apps/web/AGENTS.md
- dashboard/transactions integration pattern

Scope boleh:
- apps/web/src/app/api/admin/inventory/route.ts
- apps/web/src/lib/server/inventory.ts
- apps/web/src/app/inventory/**
- apps/web/src/components/** terkait inventory
- tests
- docs story/task/bug

Requirement:
- GET /api/admin/inventory BFF.
- sealed session, owner/admin only.
- query whitelist.
- page server-side helper.
- empty/error/forbidden/session state.
- stock adjustment buttons disabled/gated.
- fallback preview berlabel jika dipakai.

Validasi web full + boundary scan.

Docs next: CHK-03.
```

---

## PROMPT 014 — Checkpoint Inventory Integration

```text
Task ID:
CHK-03 — Checkpoint after inventory read-only integration

Mode:
- Validasi, docs check, commit, push.
- Jangan implement fitur baru.

Run:
- backend tests jika backend berubah.
- web typecheck/lint/test/build/playwright list/boundary scan.
- update task queue.
- commit message: checkpoint-inventory-readonly-integration
- push branch aktif.

Report commit hash, push status, next task: catalog read-only contract.
```

---

## PROMPT 015 — Catalog Products Categories Read-only Contract

```text
Task ID:
WEB-02C1 — Catalog Read-only Contract

Mode:
- Docs contract only.

Baca:
- PRDPOSJA.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_BACKEND.md
- apps/backend/routes/api.php
- backend product/category controllers/models
- web catalog route

Buat:
- docs/API_CONTRACTS/CATALOG_READ.md

Kontrak harus mencakup:
1. Products list.
2. Categories list.
3. BFF routes:
   - GET /api/admin/catalog/products
   - GET /api/admin/catalog/categories
4. Query products: q, category_id, status, page, per_page.
5. Response rows: id, sku, name, category, price, status, stock_summary, updated_at.
6. Categories rows: id, name, product_count, status.
7. No create/update/delete.
8. Tenant isolation.
9. Empty/error states.
10. Acceptance criteria backend/web.

Update docs; next backend readiness.
```

---

## PROMPT 016 — Catalog Backend Readiness Check

```text
Task ID:
WEB-02C2 — Catalog Backend Readiness Check

Mode:
- Backend read-only only.
- Jangan web/mobile.
- Jangan create/update/delete produk.

Implement/verify product/category list endpoints sesuai kontrak.

Test:
- owner/admin list products/categories.
- guest 401.
- forbidden role.
- tenant isolation.
- filters.
- empty state.

Validasi backend full.
Update docs; next web integration.
```

---

## PROMPT 017 — Catalog BFF and Page Integration

```text
Task ID:
WEB-02C3 — Catalog BFF + Page Integration

Mode:
- Web read-only only.
- Jangan backend/mobile.
- Jangan CRUD.

Buat BFF:
- GET /api/admin/catalog/products
- GET /api/admin/catalog/categories

Integrasi page catalog/products/categories existing.

Requirements:
- owner/admin only.
- query whitelist.
- no token to client.
- create/edit/delete buttons disabled or route preview only.
- empty/error/session states.
- preview fallback labeled.

Validasi web full + boundary scan.
Update docs; next checkpoint.
```

---

## PROMPT 018 — Checkpoint Catalog Integration

```text
Task ID:
CHK-04 — Checkpoint after catalog read-only integration

Run backend/web validation as applicable.
Commit: checkpoint-catalog-readonly-integration
Push branch aktif.
Next: reports read-only contract.
```

---

## PROMPT 019 — Reports Read-only Contract

```text
Task ID:
WEB-02D1 — Reports Read-only Contract

Mode:
- Docs contract only.

Tujuan:
Kontrak halaman laporan owner/admin read-only.

Buat:
- docs/API_CONTRACTS/REPORTS_READ.md

Cakupan:
1. Sales summary.
2. Top products.
3. Payment methods.
4. Cashier performance.
5. Outlet performance.
6. Date range filtering.

BFF routes:
- GET /api/admin/reports/sales-summary
- GET /api/admin/reports/top-products
- GET /api/admin/reports/payment-methods
- GET /api/admin/reports/cashiers
- GET /api/admin/reports/outlets

Rules:
- owner/admin only.
- integer rupiah.
- server-calculated.
- no export real yet.
- no client-side final calculation.
- tenant/outlet scoped.

Update docs; next backend readiness.
```

---

## PROMPT 020 — Reports Backend Readiness Check

```text
Task ID:
WEB-02D2 — Reports Backend Readiness Check

Mode:
- Backend read-only only.
- Jangan web/mobile.
- Jangan export real.

Verify/adjust existing reports endpoints to satisfy contract.

Test:
- owner/admin access.
- guest 401.
- forbidden role.
- tenant isolation.
- date range validation.
- empty state.

Validasi backend full.
Update docs; next web integration.
```

---

## PROMPT 021 — Reports BFF and Page Integration

```text
Task ID:
WEB-02D3 — Reports BFF + Page Integration

Mode:
- Web read-only only.
- Jangan backend/mobile.
- Jangan export real.

Implement BFF report routes and integrate reports page.

Rules:
- no direct Laravel from UI.
- no token to client.
- owner/admin only.
- charts/list use server data.
- export/download buttons disabled/gated.
- empty/error/session states.

Validasi web full + boundary scan.
Update docs; next checkpoint.
```

---

## PROMPT 022 — Checkpoint Reports Integration

```text
Task ID:
CHK-05 — Checkpoint after reports read-only integration

Run validation.
Commit: checkpoint-reports-readonly-integration
Push.
Next: staff/customers/outlets read-only planning.
```

---

## PROMPT 023 — Staff Customers Outlets Read-only Contract

```text
Task ID:
WEB-02E1 — Staff Customers Outlets Read-only Contract

Mode:
- Docs only.

Buat:
- docs/API_CONTRACTS/STAFF_CUSTOMERS_OUTLETS_READ.md

Cakupan:
- Staff list read-only.
- Customers list read-only.
- Outlets list read-only.

BFF routes:
- GET /api/admin/staff
- GET /api/admin/customers
- GET /api/admin/outlets

Rules:
- owner/admin only.
- customer PII masked/minimized as needed.
- no reset password/ban/deactivate real.
- no create/update/delete yet.
- tenant isolation.

Update docs; next backend readiness.
```

---

## PROMPT 024 — Staff Customers Outlets Backend Readiness

```text
Task ID:
WEB-02E2 — Staff Customers Outlets Backend Readiness

Mode:
- Backend read-only.
- No web/mobile.
- No reset/ban/user destructive action.

Verify/implement read-only list endpoints.

Test:
- owner/admin.
- guest.
- forbidden.
- tenant isolation.
- PII minimization for customer/user data.
- empty state.

Validasi backend full.
Update docs; next web integration.
```

---

## PROMPT 025 — Staff Customers Outlets BFF and Page Integration

```text
Task ID:
WEB-02E3 — Staff Customers Outlets BFF + Page Integration

Mode:
- Web read-only.
- No backend/mobile.
- No reset password/ban/deactivate/create/update/delete real.

Create BFF routes and integrate pages:
- /staff
- /customers
- /outlets

Requirements:
- owner/admin only.
- token hidden server-side.
- no PII leak beyond contract.
- actions disabled/gated.
- empty/error/session states.

Validasi web full + boundary scan.
Update docs.
```

---

## PROMPT 026 — Checkpoint Tenant Admin Read-only Baseline

```text
Task ID:
CHK-06 — Checkpoint Tenant Admin Read-only Baseline

Mode:
- No new feature.
- Validate everything.
- Commit and push.

Milestone includes:
- Dashboard real read-only.
- Transactions read-only.
- Inventory read-only.
- Catalog read-only.
- Reports read-only.
- Staff/customers/outlets read-only.

Validation:
Backend full tests.
Web full tests/build/playwright list/boundary scan.
Docs status check.

Commit message:
checkpoint-tenant-admin-readonly-baseline

Next recommended:
Safe CRUD planning, not sensitive actions.
```

---

# Phase C — Tenant Admin Safe CRUD, No Sensitive Actions

## PROMPT 027 — Low-risk CRUD Planning Contract

```text
Task ID:
WEB-04A1 — Low-risk CRUD Planning Contract

Mode:
- Docs only.
- Jangan coding CRUD dulu.

Tujuan:
Rencanakan CRUD yang aman:
- products create/update basic info
- categories create/update
- customers create/update
- settings safe update

Tidak termasuk:
- stock adjustment
- price mass update
- payment
- refund/void
- shift/cash
- user ban/reset

Buat/update:
- docs/API_CONTRACTS/SAFE_CRUD.md
- story docs
- task queue

Isi:
- endpoint candidates
- fields allowed
- fields forbidden
- validation
- idempotency need or not
- audit need or not
- acceptance criteria

Next: products CRUD backend.
```

---

## PROMPT 028 — Products Safe CRUD Backend

```text
Task ID:
WEB-04A2 — Products Safe CRUD Backend

Mode:
- Backend only.
- Safe fields only.
- Jangan stock adjustment.
- Jangan inventory movement.

Implement/verify create/update product basic info per SAFE_CRUD contract.

Test:
- owner/admin allowed.
- cashier forbidden.
- tenant isolation.
- validation.
- cannot modify stock through product update.
- audit if required by PRD.

Validasi backend full.
Update docs.
```

---

## PROMPT 029 — Products Safe CRUD Web Integration

```text
Task ID:
WEB-04A3 — Products Safe CRUD Web Integration

Mode:
- Web only.
- Use BFF.
- No direct Laravel.
- No stock adjustment.

Implement product create/update UI through BFF.

Rules:
- form submit real only for approved safe fields.
- loading/error/validation/forbidden states.
- no token client.
- destructive delete remains disabled/gated.

Validasi web full + boundary scan.
Update docs.
```

---

## PROMPT 030 — Categories and Customers Safe CRUD

```text
Task ID:
WEB-04B — Categories and Customers Safe CRUD

Mode:
- Split into backend then web within this task only if small; otherwise stop and request split.
- No sensitive actions.

Implement safe create/update for:
- categories
- customers

Rules:
- tenant scoped.
- validation.
- no PII overexposure.
- delete/archive destructive stays gated unless explicitly approved.

Validation backend + web as applicable.
Update docs.
```

---

## PROMPT 031 — Settings Safe Update Planning and Integration

```text
Task ID:
WEB-04C — Settings Safe Update

Mode:
- Low-risk settings only.
- Jangan payment gateway real.
- Jangan auth/security destructive changes.

Allowed settings:
- business display name
- address/contact basic
- receipt text preview if already supported
- outlet basic metadata if safe

Forbidden for now:
- payment gateway credentials
- tax final rules if not approved
- user password reset
- role permission changes

Implement with backend tests and web BFF integration if endpoints ready.
If endpoints not ready, make contract first and stop.

Validation full.
Update docs.
```

---

## PROMPT 032 — Checkpoint Safe CRUD Baseline

```text
Task ID:
CHK-07 — Checkpoint Safe CRUD Baseline

Validate backend/web.
Commit: checkpoint-tenant-admin-safe-crud-baseline
Push.
Next: Platform Super Admin read-only contracts.
```

---

# Phase D — Platform Super Admin Read-only

## PROMPT 033 — Platform Overview Read-only Contract

```text
Task ID:
PLATFORM-01A1 — Platform Overview Read-only Contract

Mode:
- Docs only.
- No runtime code.

Buat:
- docs/API_CONTRACTS/PLATFORM_OVERVIEW_READ.md

Cakupan:
- Ringkasan Platform
- Toko status
- Paket status
- Tagihan summary
- Bantuan summary
- Sistem summary
- Aktivitas terbaru aggregated/safe

Rules:
- superadmin/internal only.
- agregat platform boleh.
- detail toko sensitif harus minimal/gated.
- no akses bantuan real.
- no payment retry/manual mark paid.

Define BFF routes:
- GET /api/admin/platform/overview

Backend endpoint candidates:
- existing /superadmin/summary or new read endpoint.

Update docs.
```

---

## PROMPT 034 — Platform Overview Backend Endpoint

```text
Task ID:
PLATFORM-01A2 — Platform Overview Backend Endpoint

Mode:
- Backend read-only only.
- No web integration.
- No sensitive action.

Implement/verify superadmin overview endpoint.

Requirements:
- auth sanctum.
- superadmin only.
- aggregate only for transactions/GMV.
- no cross-tenant detail dump.
- response per contract.

Test:
- superadmin allowed.
- owner/admin/cashier forbidden.
- guest 401.
- aggregate shape stable.

Validasi backend full.
Update docs.
```

---

## PROMPT 035 — Platform Overview Web Integration

```text
Task ID:
PLATFORM-01A3 — Platform Overview Web Integration

Mode:
- Web only.
- BFF only.
- No backend/mobile.
- No sensitive action.

Implement BFF:
- GET /api/admin/platform/overview

Integrate /platform Ringkasan page.

Rules:
- superadmin only.
- owner/admin forbidden from platform.
- no token client.
- preview fallback labeled.
- privacy copy remains.

Validate web full + boundary scan.
Update docs.
```

---

## PROMPT 036 — Platform Toko and Paket Read-only Contract

```text
Task ID:
PLATFORM-02A1 — Platform Toko and Paket Read-only Contract

Mode:
- Docs only.

Buat:
- docs/API_CONTRACTS/PLATFORM_STORES_PLANS_READ.md

Cakupan:
- /platform/businesses Toko list/detail summary
- /platform/plans Paket list

Rules:
- superadmin only.
- contact masked.
- no access bantuan real.
- suspend/archive disabled.
- publish/edit package disabled until safe write later.

BFF routes:
- GET /api/admin/platform/stores
- GET /api/admin/platform/stores/{id}
- GET /api/admin/platform/plans

Update docs.
```

---

## PROMPT 037 — Platform Toko and Paket Backend Readiness

```text
Task ID:
PLATFORM-02A2 — Platform Toko and Paket Backend Readiness

Mode:
- Backend read-only.
- No web.

Verify/adjust existing superadmin businesses/plans endpoints.

Tests:
- superadmin allowed.
- tenant roles forbidden.
- contact masking if response resource owns it.
- no detailed transaction leak.
- pagination/filter stable.

Validate backend full.
Update docs.
```

---

## PROMPT 038 — Platform Toko and Paket Web Integration

```text
Task ID:
PLATFORM-02A3 — Platform Toko and Paket Web Integration

Mode:
- Web only.
- Read-only.

Integrate:
- /platform/businesses
- /platform/plans

Through BFF routes only.

Rules:
- actions disabled/gated.
- no access bantuan real.
- no suspend/archive real.
- no package publish real.
- privacy copy visible.

Validate web full + boundary scan.
Update docs.
```

---

## PROMPT 039 — Platform Billing Read-only Contract

```text
Task ID:
PLATFORM-03A1 — Platform Billing Read-only Contract

Mode:
- Docs only.

Buat:
- docs/API_CONTRACTS/PLATFORM_BILLING_READ.md

Cakupan:
- Langganan & Tagihan
- MRR summary
- unpaid bills
- failed payments
- trial ending
- billing rows
- payment status read-only

Rules:
- gateway not connected if not real.
- no manual mark paid.
- no retry payment.
- no refund.
- no raw gateway ref if sensitive.

Update docs; mark backend gaps clearly.
```

---

## PROMPT 040 — Platform Billing Backend Read-only

```text
Task ID:
PLATFORM-03A2 — Platform Billing Backend Read-only

Mode:
- Backend read-only.
- No web.
- No payment action.

Implement minimal billing read endpoint if models support it.
If invoice domain missing, create safe read model only if already exists; otherwise stop and document blocker.

Tests:
- superadmin only.
- tenant roles forbidden.
- no payment write.
- response shape stable.

Validate backend full.
Update docs/bug if missing invoice domain.
```

---

## PROMPT 041 — Platform Billing Web Integration

```text
Task ID:
PLATFORM-03A3 — Platform Billing Web Integration

Mode:
- Web read-only.
- No payment retry/manual mark paid/refund.

Integrate /platform/revenue as Langganan & Tagihan through BFF.

Rules:
- all payment actions disabled/gated.
- gateway copy honest.
- no raw payment secret/reference if sensitive.

Validate web full + boundary scan.
Update docs.
```

---

## PROMPT 042 — Platform Support Activity System Announcement Contracts

```text
Task ID:
PLATFORM-04A1 — Platform Ops Contracts

Mode:
- Docs only.

Buat:
- docs/API_CONTRACTS/PLATFORM_SUPPORT_READ.md
- docs/API_CONTRACTS/PLATFORM_ACTIVITY_READ.md
- docs/API_CONTRACTS/PLATFORM_SYSTEM_READ.md
- docs/API_CONTRACTS/PLATFORM_ANNOUNCEMENTS_READ.md

Rules:
Support:
- read-only tickets first.
- no assign/tandai selesai real.
Activity:
- audit/activity read-only.
- no before/after huge sensitive dump.
System:
- no polling.
- read-only service status.
Announcement:
- drafts/list read-only first.
- no broadcast real.

Update story/task/bug.
```

---

## PROMPT 043 — Platform Ops Backend Read-only

```text
Task ID:
PLATFORM-04A2 — Platform Ops Backend Read-only

Mode:
- Backend read-only.
- No web.
- No support assignment.
- No announcement send.

Implement or identify blockers for:
- support tickets read-only
- activity logs read-only
- system status read-only
- announcements read-only/draft list if model exists

If domain models missing, implement only safe available endpoints and document blockers.

Tests:
- superadmin only.
- tenant forbidden.
- no write side effects.

Validate backend full.
Update docs.
```

---

## PROMPT 044 — Platform Ops Web Integration

```text
Task ID:
PLATFORM-04A3 — Platform Ops Web Integration

Mode:
- Web read-only.

Integrate:
- /platform/support
- /platform/audit
- /platform/system-health
- /platform/announcements

Through BFF only.

Rules:
- actions disabled/gated.
- no assignment real.
- no broadcast real.
- no polling.
- privacy/governance copy visible.

Validate web full + boundary scan.
Update docs.
```

---

## PROMPT 045 — Checkpoint Platform Read-only Baseline

```text
Task ID:
CHK-08 — Checkpoint Platform Read-only Baseline

Validate backend/web.
Commit: checkpoint-platform-readonly-baseline
Push.
Next: Mobile Kasir hardening audit.
```

---

# Phase E — Mobile Kasir Hardening

## PROMPT 046 — Mobile Kasir Integration Audit

```text
Task ID:
MOB-01A — Mobile Kasir Current Integration Audit

Mode:
- Audit/docs only.
- Jangan coding mobile dulu.

Baca:
- apps/cashier/AGENTS.md
- PRDPOSJA.md
- docs/STORY_MOBILE_KASIR.md
- docs/BUG_TRACKER.md
- apps/cashier/lib structure

Tujuan:
Audit kondisi mobile kasir:
- auth/session
- outlet/device
- PIN switch
- shift
- checkout
- payment method
- printer/share receipt
- inventory/staff/reports
- error/loading/offline states

Output:
- update docs/STORY_MOBILE_KASIR.md
- update docs/BUG_TRACKER.md
- update docs/TASK_QUEUE.md

No runtime change.
Next: shift/session hardening.
```

---

## PROMPT 047 — Mobile Shift and Terminal Session Hardening

```text
Task ID:
MOB-03A — Shift and Terminal Session Hardening

Mode:
- Mobile only unless backend bug is blocking.
- Jangan ubah web.
- Jangan create backend endpoint baru tanpa prompt terpisah.

Focus:
- shift status display
- terminal actor context
- safe error state
- no client-trusted cashier id for sensitive writes
- no fake close/open if backend rejects

Validation:
cd apps/cashier
flutter analyze
flutter test
flutter build apk --debug if environment supports it

Update docs/bug.
```

---

## PROMPT 048 — Mobile Checkout Validation Hardening

```text
Task ID:
MOB-02A — POS Checkout Validation Hardening

Mode:
- Mobile only.
- No backend unless blocker.

Focus:
- checkout UI sends draft to backend.
- UI does not calculate final totals as source of truth.
- display backend totals.
- robust error states 401/403/409/422/500.
- no duplicate submit.
- no unbounded offline queue.

Validation Flutter analyze/test/build debug.
Update docs.
```

---

## PROMPT 049 — Mobile Receipt Print Share Hardening

```text
Task ID:
MOB-05A — Receipt Print/Share Hardening

Mode:
- Mobile only.

Focus:
- receipt content from backend transaction data.
- no PII/payment sensitive reference in logs.
- print/share error state.
- retry limited to UI action, not background queue.

Validation Flutter.
Update docs.
```

---

## PROMPT 050 — Mobile QA Checkpoint

```text
Task ID:
CHK-09 — Mobile Kasir Hardening Checkpoint

Run:
cd apps/cashier
flutter analyze
flutter test
flutter build apk --debug if possible

Also docs check.
Commit: checkpoint-mobile-kasir-hardening
Push.
Next: QA regression pass.
```

---

# Phase F — Regression, Polish, and Production Gates

## PROMPT 051 — Full Regression QA Pass

```text
Task ID:
QA-01 — Full Regression QA Pass

Mode:
- QA only.
- Jangan implement fitur baru.
- Fix only small obvious test/doc issues; otherwise report blocker.

Run:
Backend:
cd apps/backend
php artisan test
php artisan route:list --path=api/v1

Web:
cd apps/web
npm run typecheck
npm run lint
npm test
npm run build
npx playwright test --list
Boundary scan.

Mobile:
cd apps/cashier
flutter analyze
flutter test
flutter build apk --debug if possible

Docs:
- verify TASK_QUEUE statuses.
- verify BUG_TRACKER active bugs.
- verify PRDPOSJA consistency.

Report all failures and next fix prompts.
```

---

## PROMPT 052 — UI Polish Pass Tenant Admin

```text
Task ID:
UI-01 — Tenant Admin UI Polish Pass

Mode:
- Web UI only.
- No backend changes.
- No new API.
- No sensitive actions.

Focus:
- responsive no horizontal overflow.
- loading/empty/error/forbidden consistency.
- Indonesian copy clarity.
- actions disabled/gated where not implemented.
- dashboard/transactions/inventory/catalog/reports/staff/customers/outlets visual consistency.

Validation web full + Playwright list + boundary scan.
Update docs.
```

---

## PROMPT 053 — UI Polish Pass Platform Super Admin

```text
Task ID:
UI-02 — Platform Super Admin UI Polish Pass

Mode:
- Web UI only.
- No backend.
- No sensitive action.

Focus:
- Ringkasan/Toko/Paket/Langganan/Bantuan/Aktivitas/Sistem/Pengumuman.
- Indonesia-friendly copy.
- privacy/governance notes.
- responsive no overflow.
- disabled/gated actions clear.

Validation web full + boundary scan.
Update docs.
```

---

## PROMPT 054 — Backend Security and Tenant Scope Audit

```text
Task ID:
SEC-01 — Backend Security and Tenant Scope Audit

Mode:
- Audit + tests.
- No feature expansion.

Focus:
- business_id scoping.
- outlet validation.
- role middleware.
- forbidden responses.
- PII/log redaction.
- idempotency coverage for sensitive writes existing.

Allowed:
- add tests.
- fix obvious tenant-scope bugs.
- update BUG_TRACKER.

Validation backend full.
Report risks.
```

---

## PROMPT 055 — Sensitive Action Framework Contract

```text
Task ID:
SAFE-01A — Sensitive Action Framework Contract

Mode:
- Docs only.
- Do not implement sensitive action yet.

Buat:
- docs/API_CONTRACTS/SENSITIVE_ACTIONS.md

Cakupan framework:
- reason required.
- audit log required.
- idempotency required.
- role/policy required.
- confirmation token if needed.
- before/after summary.
- approval flow if high risk.

Actions covered later:
- refund
- void
- payment retry
- manual mark paid
- shift close/open
- stock adjustment
- cash reconciliation
- access bantuan
- suspend/archive toko
- reset password
- ban user
- announcement broadcast

Update story/task.
```

---

## PROMPT 056 — Final Documentation Sync

```text
Task ID:
DOC-99 — Final Documentation Sync

Mode:
- Docs only.

Goal:
Make docs match actual implemented app.

Check/update:
- PRDPOSJA.md
- docs/README.md
- docs/STORY_PROGRESS.md
- docs/STORY_BACKEND.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_MOBILE_KASIR.md
- docs/STORY_INTEGRATION.md
- docs/BUG_TRACKER.md
- docs/API_CONTRACTS/*
- docs/VALIDATION_CHECKLIST.md
- docs/TASK_QUEUE.md

No runtime code.
Report final next phase.
```

---

## PROMPT 057 — Final Release Candidate Checkpoint

```text
Task ID:
RC-01 — Release Candidate Checkpoint

Mode:
- Validation + commit/push/tag suggestion only.
- No feature coding.

Run all validation:
Backend test + route list.
Web typecheck/lint/test/build/playwright list/boundary scan.
Mobile analyze/test/build debug.
Docs consistency check.
Git status.

If all pass:
- Commit final pending changes.
- Push.
- Suggest tag name, but do not create tag unless user explicitly approves.

Report:
- pass/fail matrix.
- known bugs.
- production blockers.
- recommended next milestone.
```

---

# Prompt Repair Templates

Gunakan prompt repair ini ketika prompt utama gagal validasi.

## REPAIR PROMPT A — Fix Failed Validation Only

```text
Kamu mengerjakan project NojPOS.

Mode:
- Repair only.
- Jangan tambah fitur baru.
- Perbaiki hanya error dari validasi terakhir.

Input error terakhir:
[PASTE ERROR DI SINI]

Scope:
- Hanya file yang terkait langsung dengan error.
- Jangan refactor besar.
- Jangan ubah kontrak kecuali error membuktikan kontrak salah.

Validasi ulang:
[JALANKAN COMMAND YANG GAGAL + COMMAND TERKAIT]

Laporan:
1. Penyebab error.
2. File diubah.
3. Validasi ulang.
4. Apakah task utama sekarang bisa dianggap selesai.
```

## REPAIR PROMPT B — Resolve Merge/Dirty Working Tree

```text
Kamu mengerjakan project NojPOS.

Mode:
- Git hygiene only.
- Jangan coding fitur.

Tujuan:
Membersihkan status repo sebelum lanjut task berikutnya.

Langkah:
1. git status --short
2. Kelompokkan perubahan:
   - perubahan task sebelumnya
   - artifact/cache/noise
   - perubahan tidak dikenal
3. Jangan hapus perubahan user.
4. Update .gitignore hanya jika jelas noise.
5. Commit hanya jika validasi terkait sudah pernah lolos.

Laporan:
- file dirty
- tindakan
- commit hash jika commit
- blocker jika ada
```

## REPAIR PROMPT C — Docs Status Mismatch

```text
Kamu mengerjakan project NojPOS.

Mode:
- Docs repair only.
- Jangan runtime.

Tujuan:
Sinkronkan status docs setelah task selesai.

Cek/update:
- docs/TASK_QUEUE.md
- docs/STORY_PROGRESS.md
- story area terkait
- docs/BUG_TRACKER.md

Pastikan:
- hanya satu next task READY.
- task selesai ditandai DONE.
- blocker tidak dihapus tanpa alasan.

Validasi:
Docs read check.
```
