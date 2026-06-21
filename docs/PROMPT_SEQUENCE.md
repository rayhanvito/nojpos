# NojPOS Prompt Sequence v2 — Single-Agent Master Copy-Paste Plan

Dokumen ini adalah **master urutan prompt** untuk mengendalikan implementasi NojPOS secara pelan, aman, berurutan, dan tidak loncat-loncat.

Target dokumen ini:

1. Membantu user tinggal **copy-paste prompt satu per satu** ke coding agent.
2. Membuat coding agent paham scope, file yang boleh diedit, file yang tidak boleh diedit, validasi, dan laporan akhir.
3. Menjaga implementasi tetap sesuai `PRDPOSJA.md`, story docs, API contracts, dan single-agent workflow.
4. Mencegah agent mengaktifkan fitur sensitif terlalu cepat.
5. Menjaga backend, web admin, dan mobile kasir tidak saling tabrak.

---

## Cara Pakai Wajib

1. Jalankan **satu prompt saja** dalam satu sesi agent.
2. Jangan lanjut prompt berikutnya sebelum prompt sekarang selesai, validasi jelas, dan blocker dicatat.
3. Kalau validasi gagal, gunakan prompt repair yang sesuai. Jangan loncat ke fitur baru.
4. Setelah milestone besar, jalankan checkpoint commit.
5. Semua agent wajib update story docs dan task queue saat task selesai.
6. Semua agent wajib melapor dengan format: `Built`, `Verified`, `Blocker`.
7. Jangan menjalankan beberapa coding agent paralel. Workflow repo ini adalah **single active coding agent**.

---

## File Yang Selalu Harus Dibaca Agent

Setiap prompt di bawah mengharuskan agent membaca minimal:

```text
AGENTS.md
PRDPOSJA.md
docs/README.md
docs/SINGLE_AGENT_WORKFLOW.md
docs/TASK_QUEUE.md
docs/STORY_PROGRESS.md
docs/BUG_TRACKER.md
```

Tambahkan file area sesuai task:

```text
Backend: apps/backend/AGENTS.md, docs/STORY_BACKEND.md, docs/API_CONTRACTS/* terkait
Web:     apps/web/AGENTS.md, docs/STORY_WEB_ADMIN.md, docs/API_CONTRACTS/* terkait
Mobile:  apps/cashier/AGENTS.md, docs/STORY_MOBILE_KASIR.md
Integrasi: docs/STORY_INTEGRATION.md, docs/VALIDATION_CHECKLIST.md
```

---

## Larangan Global Untuk Semua Prompt

Larangan ini berlaku kecuali prompt eksplisit sudah masuk fase sensitive action dan framework audit/idempotency sudah siap.

```text
Jangan install package baru tanpa instruksi eksplisit.
Jangan edit node_modules/vendor/build cache.
Jangan simpan token di localStorage/sessionStorage.
Jangan expose token ke Client Component, props React, fixture, log, atau screenshot.
Jangan memakai NEXT_PUBLIC_API_BASE_URL untuk backend secret/server API.
Jangan membuat fetch langsung dari UI ke Laravel.
Jangan membuat action sensitif real sebelum waktunya.
Jangan mengaktifkan refund, void, payment retry, manual mark paid, shift close/open, stock adjustment, cash reconciliation, akses bantuan/impersonate, suspend/delete toko, reset password, ban user, atau broadcast pengumuman sebelum framework audit/idempotency/permission siap.
Jangan mengubah backend, web, dan mobile sekaligus kecuali prompt checkpoint/QA eksplisit memeriksa semua tanpa coding fitur.
Jangan menghapus bug lama dari BUG_TRACKER tanpa bukti validasi.
```

---

## Validasi Standar Per Area

### Backend

```bash
cd apps/backend
php artisan test
php artisan route:list --path=api/v1
```

Jika formatter tersedia dan sesuai guide:

```bash
./vendor/bin/pint --dirty
```

### Web Admin

```bash
cd apps/web
npm run typecheck
npm run lint
npm test
npm run build
npx playwright test --list
```

Boundary scan web:

```bash
rg -n "fetch\(|XMLHttpRequest|sessionStorage|localStorage|NEXT_PUBLIC_API_BASE_URL" src/app src/components src/fixtures --glob '!*.test.ts' --glob '!*.test.tsx' || true
rg -n "Authorization|Bearer " src/app src/components src/fixtures --glob '!*.test.ts' --glob '!*.test.tsx' || true
```

Catatan: `Authorization/Bearer` boleh muncul hanya di server-only helper/route handler bila token dijaga di server. Tidak boleh muncul di UI page/component/fixture.

### Mobile Kasir

```bash
cd apps/cashier
flutter analyze
flutter test
flutter build apk --debug
```

### Docs-only

```bash
rg -n "docs/AGENT_WORKFLOW.md|docs/AGENT_TASK_BOARD.md|docs/AGENT_PROMPT_TEMPLATE.md|docs/AGENT_HANDOFF_TEMPLATE.md|setup-agent-worktrees|worktree setup" docs --glob '!PROMPT_SEQUENCE.md' || true
```

---

## Status Baseline Saat Dokumen Ini Dibuat

Sudah selesai dari percakapan sebelumnya:

```text
PROMPT 001 / BE-04A — kontrak GET /api/v1/dashboard/summary
PROMPT 002 / BE-04B — backend endpoint dashboard summary
PROMPT 003 / WEB-01A — session strategy Web Admin
PROMPT 004 / WEB-01B — BFF/session scaffold
PROMPT 005 / WEB-01C — integrasi /dashboard read-only lewat BFF
```

Prompt aktif berikutnya:

```text
PROMPT 006 — checkpoint commit setelah dashboard integration
```

---

# Phase 0 — Completed Baseline Archive

Bagian ini disimpan supaya dokumen lengkap dari awal. Biasanya tidak perlu dijalankan ulang kecuali repo di-reset.

## PROMPT 001 — BE-04A Finalize Dashboard Summary API Contract

```text
Kamu mengerjakan NojPOS.

Task ID: BE-04A — Finalize Dashboard Summary API Contract.

Mode kerja:
- Docs-only.
- Jangan ubah runtime backend/web/mobile.
- Finalisasi kontrak GET /api/v1/dashboard/summary.

Baca:
- AGENTS.md
- PRDPOSJA.md
- docs/SINGLE_AGENT_WORKFLOW.md
- docs/TASK_QUEUE.md
- docs/STORY_BACKEND.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_INTEGRATION.md
- docs/API_CONTRACTS/DASHBOARD_SUMMARY.md
- apps/backend/routes/api.php
- apps/web/src/app/dashboard/page.tsx
- apps/web/src/fixtures/preview/dashboard.ts

Scope boleh diedit:
- docs/API_CONTRACTS/DASHBOARD_SUMMARY.md
- docs/STORY_BACKEND.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md
- docs/BUG_TRACKER.md jika ada blocker

Output wajib:
- Endpoint final GET /api/v1/dashboard/summary.
- Auth owner/admin tenant, cashier forbidden, superadmin tidak memakai endpoint ini.
- Query date, outlet_id, range=last_7_days.
- Response shape: meta, kpis, alerts, sales_last_7_days, payment_methods, top_products, low_stock_items, recent_transactions, cashier_performance, branch_highlights, data_notes.
- Empty state, error state 401/403/422/500.
- Data source candidates.
- Tenant isolation rules.
- Acceptance criteria BE-04B.

Validasi:
- Docs read check.
- Tidak ada referensi workflow agent paralel lama.

Laporan akhir:
Built, Verified, Blocker, Next recommended task BE-04B.
```

## PROMPT 002 — BE-04B Implement Dashboard Summary Backend Endpoint

```text
Kamu mengerjakan NojPOS.

Task ID: BE-04B — Implement GET /api/v1/dashboard/summary.

Mode kerja:
- Backend only.
- Jangan ubah apps/web atau apps/cashier.
- Endpoint read-only.
- Jangan trigger payment, refund, void, shift close/open, stock adjustment, export, notification.

Baca:
- AGENTS.md
- PRDPOSJA.md
- docs/API_CONTRACTS/DASHBOARD_SUMMARY.md
- docs/STORY_BACKEND.md
- apps/backend/AGENTS.md
- apps/backend/routes/api.php

Scope boleh diedit:
- apps/backend/routes/api.php
- apps/backend/app/Http/Controllers/**
- apps/backend/app/Http/Requests/**
- apps/backend/app/Services/**
- apps/backend/tests/**
- docs/STORY_BACKEND.md
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md

Requirement:
- Route GET /api/v1/dashboard/summary protected auth:sanctum + role owner/admin.
- Cashier/superadmin forbidden.
- Validate date, outlet_id, range.
- Scope semua query ke business_id user.
- outlet_id harus milik business user.
- Response sesuai kontrak.
- Gross profit aman bila COGS belum lengkap.
- Cash difference read-only.
- Jangan hardcode mock sebagai produksi.

Test wajib:
- owner/admin success.
- cashier/superadmin 403.
- guest 401.
- invalid filter 422.
- cross-tenant outlet rejected.
- empty state stable.
- normal shape stable.
- tenant isolation.

Validasi:
- ./vendor/bin/pint --dirty jika ada.
- php artisan test.
- php artisan route:list --path=api/v1.

Laporan akhir:
Built, Verified, Blocker, Next recommended task WEB-01A.
```

## PROMPT 003 — WEB-01A Decide Web Admin Session Strategy

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-01A — Decide and document Web Admin session strategy.

Mode kerja:
- Docs-only.
- Jangan coding runtime.
- Jangan integrasi dashboard.
- Jangan localStorage/sessionStorage.

Baca:
- AGENTS.md
- PRDPOSJA.md
- docs/API_CONTRACTS/SESSION.md
- docs/STORY_WEB_ADMIN.md
- apps/backend/routes/api.php
- apps/web/AGENTS.md

Scope boleh diedit:
- docs/API_CONTRACTS/SESSION.md
- docs/API_CONTRACTS/README.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md
- docs/BUG_TRACKER.md

Analisis opsi:
A. Direct Sanctum SPA cookie.
B. Direct Bearer token di browser.
C. Next.js BFF/server-side route handler + HttpOnly sealed session cookie.

Output wajib:
- Pilih strategi final.
- Jelaskan alasan.
- Login/session/logout flow.
- Cookie rules.
- Role protection /dashboard dan /platform.
- Backend gaps.
- Acceptance criteria WEB-01B.

Validasi:
- Docs read check.
- Tidak ada instruksi memakai localStorage/sessionStorage.

Laporan akhir:
Built, Verified, Blocker, Next recommended task WEB-01B.
```

## PROMPT 004 — WEB-01B Implement Web Admin Session/BFF Scaffold

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-01B — Implement Web Admin session/BFF adapter scaffold.

Mode kerja:
- Web only.
- Jangan integrasi dashboard.
- Jangan ubah backend/mobile.
- Jangan expose token ke browser.

Baca:
- AGENTS.md
- PRDPOSJA.md
- docs/API_CONTRACTS/SESSION.md
- apps/web/AGENTS.md
- apps/web/package.json
- apps/backend/routes/api.php sebagai referensi saja.

Scope boleh diedit:
- apps/web/src/lib/server/**
- apps/web/src/app/api/admin/auth/login/route.ts
- apps/web/src/app/api/admin/auth/logout/route.ts
- apps/web/src/app/api/admin/session/route.ts
- apps/web/src/**/*.test.ts
- apps/web/.env.example
- docs/STORY_WEB_ADMIN.md
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md

Requirement:
- BFF route /api/admin/auth/login, /api/admin/auth/logout, /api/admin/session.
- Token Laravel hanya server-side.
- Cookie HttpOnly, SameSite=Lax, Secure production, TTL jelas.
- Cookie harus sealed/encrypted/HMAC. Jangan plaintext token.
- No localStorage/sessionStorage.
- No token in Client Component/props.
- Safe session context only.
- Cashier forbidden for web admin.

Test:
- Cookie tidak plaintext token.
- Response tidak return token.
- logout clears cookie.
- session 401 tanpa cookie.

Validasi web lengkap.

Laporan akhir:
Built, Verified, Blocker, Next recommended task WEB-01C.
```

## PROMPT 005 — WEB-01C Integrate /dashboard Read-only Through BFF

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-01C — Integrate /dashboard read-only data through BFF.

Mode kerja:
- Web only.
- Jangan ubah backend/mobile.
- Read-only dashboard.
- Jangan action real.

Baca:
- AGENTS.md
- PRDPOSJA.md
- docs/API_CONTRACTS/DASHBOARD_SUMMARY.md
- docs/API_CONTRACTS/SESSION.md
- apps/web/AGENTS.md
- apps/web/src/lib/server/**
- apps/web/src/app/dashboard/page.tsx
- apps/web/src/fixtures/preview/dashboard.ts

Scope boleh diedit:
- apps/web/src/app/api/admin/dashboard/summary/route.ts
- apps/web/src/lib/server/dashboard-summary.ts
- apps/web/src/app/dashboard/page.tsx
- apps/web/src/components/dashboard/**
- tests terkait
- docs/STORY_WEB_ADMIN.md
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md

Requirement:
- Buat GET /api/admin/dashboard/summary.
- BFF memanggil Laravel GET /api/v1/dashboard/summary server-side.
- /dashboard Server Component memakai helper server-side.
- No direct Laravel call dari UI.
- Fallback session missing, forbidden, backend unavailable, empty state.
- Fixture boleh fallback dengan label jelas.
- No token exposed.

Test:
- route 401 tanpa session.
- 403 role non-owner/admin.
- query whitelist.
- response no token.
- mapping stable.

Validasi web lengkap.

Laporan akhir:
Built, Verified, Blocker, Next recommended task CHK-01.
```

---

# Phase A — Checkpoint Current Milestone

## PROMPT 006 — CHK-01 Checkpoint Commit After Dashboard Integration

```text
Kamu mengerjakan project NojPOS.

Task ID: CHK-01 — Checkpoint commit after dashboard backend + web integration.

Mode kerja:
- Checkpoint only.
- Jangan implement fitur baru.
- Jangan refactor.
- Jangan ubah runtime kecuali docs status kecil jika belum sesuai.
- Validasi, commit, push.

Baca:
- AGENTS.md
- PRDPOSJA.md
- docs/SINGLE_AGENT_WORKFLOW.md
- docs/TASK_QUEUE.md
- docs/STORY_PROGRESS.md
- docs/STORY_BACKEND.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_INTEGRATION.md
- docs/BUG_TRACKER.md
- docs/VALIDATION_CHECKLIST.md

Scope boleh:
- Inspect git status.
- Validasi backend dan web.
- Update docs status jika perlu.
- Commit semua perubahan relevan.
- Push ke branch aktif.

Scope tidak boleh:
- Jangan fitur baru.
- Jangan hapus file besar sembarangan.
- Jangan commit node_modules, .next, vendor, storage/log, tsbuildinfo, cache.

Langkah:
1. `git status --short`.
2. Pastikan noise tidak staged.
3. Backend:
   `cd apps/backend && php artisan test && php artisan route:list --path=api/v1`.
4. Web:
   `cd apps/web && npm run typecheck && npm run lint && npm test && npm run build && npx playwright test --list`.
5. Boundary scan web sesuai standar.
6. Pastikan docs/TASK_QUEUE.md menandai BE-04B, WEB-01A, WEB-01B, WEB-01C DONE.
7. Commit message: `checkpoint-dashboard-summary-bff-integration`.
8. Push ke branch aktif.

Jika validasi gagal:
- Jangan commit.
- Laporkan command gagal.
- Jangan lanjut task lain.

Laporan akhir wajib:
- Git status sebelum commit.
- File utama yang masuk commit.
- Hasil validasi backend.
- Hasil validasi web.
- Commit hash.
- Push status.
- Working tree setelah commit.
- Next recommended task PROMPT 007.
- Blocker.
```

---

# Phase B — Tenant Admin Read-only Integration

Tujuan fase ini: membuat Tenant Admin Web membaca data backend secara aman dan read-only untuk halaman penting. Tidak ada write action dulu.

Urutan fase ini:

```text
Transactions -> Inventory -> Catalog -> Reports -> Staff/Customers/Outlets -> Settings/Subscription -> checkpoint
```

## PROMPT 007 — WEB-02A Transactions Read-only Contract

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-02A — Finalize Transactions read-only contract.

Mode kerja:
- Docs-only.
- Jangan coding runtime.
- Jangan ubah backend/web/mobile.

Baca:
- AGENTS.md
- PRDPOSJA.md
- docs/TASK_QUEUE.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_BACKEND.md
- docs/STORY_INTEGRATION.md
- docs/API_CONTRACTS/README.md
- apps/backend/routes/api.php
- apps/web/src/app jika ada route transaksi

Scope boleh diedit:
- docs/API_CONTRACTS/TRANSACTIONS_READ.md
- docs/API_CONTRACTS/README.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_BACKEND.md jika ada gap backend
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md
- docs/BUG_TRACKER.md jika ada blocker

Output kontrak:
- Backend endpoint nyata yang dipakai, misalnya GET /api/v1/transactions jika tersedia.
- BFF endpoint target: GET /api/admin/transactions.
- Auth owner/admin tenant.
- Query: date_from, date_to, outlet_id, cashier_id, payment_method, status, search, page, per_page.
- Response: meta pagination, rows transaction summary, totals safe.
- Field row: id, code, time, outlet, cashier, payment_method, total, status, item_count.
- Jangan expose payment reference sensitif, customer PII penuh, token, raw receipt.
- Empty state.
- Error 401/403/422/500.
- Tenant isolation.
- Acceptance criteria implement prompt berikutnya.

Validasi:
- Docs read check.
- Tidak ada workflow agent paralel lama.

Laporan akhir:
Built, Verified, Blocker, Next recommended task PROMPT 008.
```

## PROMPT 008 — WEB-02B Transactions BFF + Page Read-only Integration

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-02B — Integrate transactions read-only through BFF.

Mode kerja:
- Web only.
- Read-only.
- Jangan ubah backend/mobile.
- Jangan refund/void/reprint/export real.

Baca:
- AGENTS.md
- PRDPOSJA.md
- docs/API_CONTRACTS/TRANSACTIONS_READ.md
- docs/API_CONTRACTS/SESSION.md
- apps/web/AGENTS.md
- apps/web/src/lib/server/**
- apps/web/src/app/transactions jika ada
- apps/backend/routes/api.php sebagai referensi saja

Scope boleh diedit:
- apps/web/src/app/api/admin/transactions/route.ts
- apps/web/src/lib/server/transactions.ts
- apps/web/src/app/transactions/**
- apps/web/src/components/** transaksi terkait
- tests terkait
- docs/STORY_WEB_ADMIN.md
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md
- docs/BUG_TRACKER.md jika ada blocker

Requirement:
- BFF GET /api/admin/transactions.
- Server-only call ke Laravel transactions endpoint.
- Query whitelist.
- Owner/admin only.
- Cashier/superadmin forbidden.
- Page transaksi render loading/session missing/forbidden/backend error/empty/success.
- Semua action sensitif seperti refund, void, reprint, export disabled/preview-only.
- No token/client exposure.

Test:
- 401 tanpa session.
- 403 role salah.
- query liar ditolak.
- no token in response.
- page fallback render.
- mapping rows stable.

Validasi web lengkap.

Laporan akhir:
Built, Verified, Blocker, Next recommended task PROMPT 009.
```

## PROMPT 009 — CHK-02 Transactions Checkpoint Commit

```text
Kamu mengerjakan NojPOS.

Task ID: CHK-02 — Transactions read-only checkpoint.

Mode kerja:
- Checkpoint only.
- Jangan fitur baru.

Baca docs workflow dan story terkait.

Langkah:
1. Cek git status.
2. Validasi web lengkap.
3. Boundary scan web.
4. Update TASK_QUEUE/STORY_WEB_ADMIN jika status belum sesuai.
5. Commit message: `checkpoint-transactions-read-only`.
6. Push branch aktif.

Jika validasi gagal, jangan commit.

Laporan akhir:
Built, Verified, Blocker, Next PROMPT 010.
```

## PROMPT 010 — WEB-03A Inventory Read-only Contract

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-03A — Finalize Inventory read-only contract.

Mode kerja:
- Docs-only.
- Jangan coding runtime.

Baca:
- AGENTS.md
- PRDPOSJA.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_BACKEND.md
- docs/API_CONTRACTS/README.md
- apps/backend/routes/api.php
- apps/backend inventory controller/model jika ada
- apps/web inventory page jika ada

Scope boleh diedit:
- docs/API_CONTRACTS/INVENTORY_READ.md
- docs/API_CONTRACTS/README.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_BACKEND.md
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md
- docs/BUG_TRACKER.md

Output kontrak:
- Backend endpoints nyata untuk inventory list, movements, transfers jika ada.
- BFF endpoints target: /api/admin/inventory, /api/admin/inventory/movements.
- Query: outlet_id, category_id, stock_status, search, page, per_page.
- Response list: product_id, sku, name, category, outlet, stock_on_hand, unit, low_stock_threshold, status.
- Movement row: time, product, outlet, type, quantity, reference, actor safe.
- Read-only only.
- Stock adjustment tetap disabled.
- Tenant/outlet isolation.
- Empty/error states.
- Acceptance criteria implementation.

Validasi docs.

Laporan akhir:
Built, Verified, Blocker, Next PROMPT 011.
```

## PROMPT 011 — WEB-03B Inventory BFF + Page Read-only Integration

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-03B — Integrate inventory read-only through BFF.

Mode kerja:
- Web only.
- Read-only.
- Jangan stock adjustment/transfer real.
- Jangan ubah backend/mobile.

Baca contract INVENTORY_READ, session docs, apps/web guide.

Scope boleh diedit:
- apps/web/src/app/api/admin/inventory/route.ts
- apps/web/src/app/api/admin/inventory/movements/route.ts jika sesuai kontrak
- apps/web/src/lib/server/inventory.ts
- apps/web/src/app/inventory/**
- tests terkait
- docs status

Requirement:
- BFF whitelist query.
- Server-only token.
- Page inventory render state aman.
- Buttons adjustment/transfer/export disabled atau preview-only.
- No client-side stock mutation.

Test:
- 401/403/query validation.
- mapping stable.
- no token leak.
- page fallback render.

Validasi web lengkap.

Laporan akhir:
Built, Verified, Blocker, Next PROMPT 012.
```

## PROMPT 012 — CHK-03 Inventory Checkpoint Commit

```text
Kamu mengerjakan NojPOS.

Task ID: CHK-03 — Inventory read-only checkpoint.

Mode kerja checkpoint only.

Validasi web lengkap, boundary scan, docs status, commit message:
`checkpoint-inventory-read-only`.
Push branch aktif.

Laporan akhir wajib.
```

## PROMPT 013 — WEB-04A Catalog Read-only Contract

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-04A — Finalize Catalog/products/categories read-only contract.

Mode kerja docs-only.

Baca route backend products/categories, web catalog pages, story docs.

Scope boleh diedit:
- docs/API_CONTRACTS/CATALOG_READ.md
- docs/API_CONTRACTS/README.md
- story/task/bug docs

Output kontrak:
- GET products list.
- GET categories list.
- BFF /api/admin/catalog/products dan /api/admin/catalog/categories.
- Query: search, category_id, status, page, per_page.
- Product fields safe: id, sku, name, category, price, stock summary, active status, updated_at.
- Category fields safe.
- No create/update/delete yet.
- Tenant isolation.
- Empty/error states.
- Acceptance criteria implement.

Validasi docs.

Laporan akhir dengan next PROMPT 014.
```

## PROMPT 014 — WEB-04B Catalog BFF + Page Read-only Integration

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-04B — Integrate catalog read-only through BFF.

Mode kerja web only, read-only.

Scope boleh diedit:
- apps/web/src/app/api/admin/catalog/products/route.ts
- apps/web/src/app/api/admin/catalog/categories/route.ts
- apps/web/src/lib/server/catalog.ts
- apps/web/src/app/catalog/**
- tests terkait
- docs status

Requirement:
- Products/categories read-only.
- Add/edit/delete buttons disabled/preview-only.
- No token leak.
- No direct Laravel from UI.
- Empty/error/session states.

Validasi web lengkap.

Laporan akhir next PROMPT 015.
```

## PROMPT 015 — CHK-04 Catalog Checkpoint Commit

```text
Checkpoint catalog read-only.

Validasi web lengkap, boundary scan, docs status, commit:
`checkpoint-catalog-read-only`.
Push.
```

## PROMPT 016 — WEB-05A Reports Read-only Contract

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-05A — Finalize Reports read-only contract.

Mode kerja docs-only.

Analisis backend reports endpoints nyata:
- sales summary
- top products
- payment methods
- cashier shifts
- other existing reports

Scope docs:
- docs/API_CONTRACTS/REPORTS_READ.md
- README contract
- story/task/bug docs

Output kontrak:
- BFF /api/admin/reports/sales-summary.
- BFF /api/admin/reports/top-products.
- BFF /api/admin/reports/payment-methods.
- BFF /api/admin/reports/cashier-shifts.
- Query date_from/date_to/outlet_id.
- Response integer rupiah.
- No client-side totals calculation.
- Export disabled.
- Tenant/outlet isolation.
- Empty/error states.

Validasi docs.
Next PROMPT 017.
```

## PROMPT 017 — WEB-05B Reports BFF + Page Read-only Integration

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-05B — Integrate reports read-only through BFF.

Mode kerja web only.

Scope boleh diedit:
- apps/web/src/app/api/admin/reports/**
- apps/web/src/lib/server/reports.ts
- apps/web/src/app/reports/**
- report chart/list components jika ada
- tests
- docs

Requirement:
- Reports read-only.
- Money display integer rupiah from server.
- No client-side recalculation of totals.
- Export button disabled/preview-only.
- No token leak.
- Session/forbidden/error/empty states.

Validasi web lengkap.
Next PROMPT 018.
```

## PROMPT 018 — CHK-05 Reports Checkpoint Commit

```text
Checkpoint reports read-only.

Validasi web lengkap, commit:
`checkpoint-reports-read-only`.
Push.
```

## PROMPT 019 — WEB-06A Staff, Customers, Outlets Read-only Contract

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-06A — Finalize Staff, Customers, Outlets read-only contract.

Mode docs-only.

Output kontrak:
- docs/API_CONTRACTS/STAFF_READ.md
- docs/API_CONTRACTS/CUSTOMERS_READ.md
- docs/API_CONTRACTS/OUTLETS_READ.md
- Update README contracts.

Staff safe fields:
- id, name, role, status, outlet scope, last activity if available.
- No PIN/password.

Customers safe fields:
- masked phone/email.
- name optional safe.
- No excessive PII.

Outlets fields:
- id, name, address short, timezone, status.

No reset password, ban user, outlet destructive action.
Tenant isolation and role rules.
Acceptance criteria.

Validasi docs.
Next PROMPT 020.
```

## PROMPT 020 — WEB-06B Staff, Customers, Outlets BFF + Page Read-only Integration

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-06B — Integrate staff/customers/outlets read-only.

Mode web only.

Scope:
- apps/web/src/app/api/admin/staff/route.ts
- apps/web/src/app/api/admin/customers/route.ts
- apps/web/src/app/api/admin/outlets/route.ts
- apps/web/src/lib/server/staff.ts
- apps/web/src/lib/server/customers.ts
- apps/web/src/lib/server/outlets.ts
- apps/web/src/app/staff/**
- apps/web/src/app/customers/**
- apps/web/src/app/outlets/**
- tests and docs

Requirement:
- Read-only lists.
- Mask PII where needed.
- Reset password/ban/delete disabled.
- No token leak.
- Session/error/empty states.

Validasi web lengkap.
Next PROMPT 021.
```

## PROMPT 021 — WEB-07A Settings and Subscription Read-only Contract

```text
Kamu mengerjakan NojPOS.

Task ID: WEB-07A — Finalize settings/subscription read-only contract.

Mode docs-only.

Output:
- docs/API_CONTRACTS/SETTINGS_READ.md
- docs/API_CONTRACTS/SUBSCRIPTION_READ.md

Settings:
- business profile read.
- payment methods config read safe.
- security settings safe flags.
- no secret/token.

Subscription:
- active plan, period, status, limits, grace/trial if available.
- no manual payment action.

Acceptance criteria.

Validasi docs.
Next PROMPT 022.
```

## PROMPT 022 — WEB-07B Settings and Subscription Read-only Integration

```text
Task ID: WEB-07B — Integrate settings/subscription read-only.

Mode web only.

Scope:
- BFF routes for settings/subscription.
- server helpers.
- settings/subscription pages.
- tests/docs.

Requirement:
- Read-only display.
- Save buttons disabled unless already safe preview.
- No secrets shown.
- Payment config masked.

Validasi web lengkap.
Next PROMPT 023.
```

## PROMPT 023 — CHK-06 Tenant Admin Read-only Checkpoint

```text
Task ID: CHK-06 — Tenant Admin read-only checkpoint.

Mode checkpoint.

Validate:
- backend route list if backend untouched optional.
- web full validation.
- boundary scan.
- docs status.

Commit message:
`checkpoint-tenant-admin-read-only`.
Push.

Next recommended task PROMPT 024.
```

---

# Phase C — Tenant Admin Safe CRUD

Tujuan fase ini: mengaktifkan write yang risikonya rendah dan punya kontrak jelas. Fitur uang/stok/shift sensitif tetap deferred.

## PROMPT 024 — CRUD-00 Safe Write Policy and Form Pattern Contract

```text
Kamu mengerjakan NojPOS.

Task ID: CRUD-00 — Safe write policy and form pattern contract.

Mode docs-only.

Tujuan:
Menentukan aturan umum sebelum CRUD aktif.

Scope docs:
- docs/API_CONTRACTS/SAFE_CRUD_POLICY.md
- docs/API_CONTRACTS/README.md
- docs/STORY_WEB_ADMIN.md
- docs/STORY_BACKEND.md
- docs/STORY_INTEGRATION.md
- docs/TASK_QUEUE.md

Output:
- Which writes are allowed first: product/category/customer basic/profile/settings low risk.
- Which writes remain forbidden: refund, void, payment retry, stock adjustment, shift close/open, cash reconciliation.
- CSRF/session/BFF pattern for POST/PATCH.
- Validation error format.
- Optimistic UI rules: avoid or keep conservative.
- Audit requirement for safe writes.
- Idempotency requirement for writes that can duplicate.
- Confirmation pattern for destructive actions: disabled until sensitive framework.

Validasi docs.
Next PROMPT 025.
```

## PROMPT 025 — PRODUCT-01 Product Safe Create/Update Contract

```text
Task ID: PRODUCT-01 — Product safe create/update contract.

Mode docs-only.

Output docs/API_CONTRACTS/PRODUCT_WRITE.md:
- Backend endpoints existing or needed.
- BFF POST/PATCH routes.
- Fields allowed: name, sku, category, price, active, basic metadata.
- Stock quantity not editable here unless backend has safe stock adjustment framework.
- Price integer rupiah.
- Validation errors.
- Tenant isolation.
- Audit requirement.
- Acceptance criteria.

No runtime coding.
Next PROMPT 026.
```

## PROMPT 026 — PRODUCT-02 Backend Product Write Hardening If Needed

```text
Task ID: PRODUCT-02 — Backend product write hardening.

Mode backend only.

Run this only if PRODUCT_WRITE contract finds backend gap.
If backend endpoints already safe and tested, mark not needed and update docs.

Scope backend product controller/request/service/tests.
No web/mobile.

Requirement:
- owner/admin only.
- tenant scoped.
- validation.
- price integer.
- no stock mutation.
- audit if product write is privileged in PRD.

Tests:
- authorization.
- tenant isolation.
- validation.
- no cross-tenant update.

Validate backend.
Next PROMPT 027.
```

## PROMPT 027 — PRODUCT-03 Web Product Create/Update Through BFF

```text
Task ID: PRODUCT-03 — Web product create/update through BFF.

Mode web only.

Scope:
- BFF product POST/PATCH.
- catalog product form/page components.
- server helpers.
- tests/docs.

Requirement:
- No direct Laravel UI call.
- No localStorage/sessionStorage.
- Validation errors visible.
- Save action real only for allowed fields.
- Delete remains disabled.
- Stock adjustment remains disabled.

Validate web.
Next PROMPT 028.
```

## PROMPT 028 — CATEGORY-01 Category CRUD Contract and Implementation

```text
Task ID: CATEGORY-01 — Category safe CRUD.

Mode:
1. Docs contract first.
2. Backend only if needed.
3. Web BFF/page integration.
If this becomes too large, split into CATEGORY-01A docs, 01B backend, 01C web.

Allowed:
- create/update category name/status.
- delete only if backend has safe soft-delete/no products guard; otherwise archive disabled.

Tests backend/web as applicable.
No mobile.
Validate all touched area.
Next PROMPT 029.
```

## PROMPT 029 — CUSTOMER-01 Customer Safe CRUD Contract and Implementation

```text
Task ID: CUSTOMER-01 — Customer safe CRUD.

Mode split if large.

Rules:
- Minimize PII.
- Mask phone/email where displayed in lists.
- Validate consent/notes if PRD has it.
- No export real.
- No bulk import.

Implement BFF + web only after contract.
Backend only if gaps.
Validate.
Next PROMPT 030.
```

## PROMPT 030 — SETTINGS-01 Business Settings Safe Update

```text
Task ID: SETTINGS-01 — Business settings safe update.

Rules:
- Allowed: business display name, address, timezone, receipt footer if backend supports.
- Forbidden: payment secrets, destructive reset, ownership transfer.
- Backend only if needed.
- Web via BFF.
- Validate timezone handling.

Validate backend/web as touched.
Next PROMPT 031.
```

## PROMPT 031 — STAFF-01 Staff Safe Update Contract

```text
Task ID: STAFF-01 — Staff safe update contract.

Mode docs-only first.

Important:
- Reset password/ban user/deactivate sensitive remains disabled unless policy ready.
- PIN never exposed.
- Role changes need audit; can be deferred.

Output contract and decide if implementation can proceed.
Next PROMPT 032.
```

## PROMPT 032 — CHK-07 Tenant Safe CRUD Checkpoint

```text
Checkpoint Tenant Admin safe CRUD.

Validate:
- backend tests if touched.
- web full validation.
- boundary scan.
- docs status.

Commit message:
`checkpoint-tenant-safe-crud`.
Push.
Next PROMPT 033.
```

---

# Phase D — Platform Super Admin Read-only

Tujuan fase ini: Super Admin membaca data backend secara aman. Semua action sensitif tetap disabled/gated.

## PROMPT 033 — PLATFORM-01 Overview Read-only Contract

```text
Task ID: PLATFORM-01 — Platform overview read-only contract.

Mode docs-only.

Baca backend superadmin routes, platform UI pages, PRD.

Output docs/API_CONTRACTS/PLATFORM_OVERVIEW_READ.md:
- Backend endpoint existing or needed.
- Recommended GET /api/v1/superadmin/overview if summary existing insufficient.
- BFF /api/admin/platform/overview.
- Auth superadmin only.
- Metrics: total toko, active, trial, expired, MRR, unpaid bills, open support, system issues.
- Aggregated only for GMV/transactions.
- Tenant detail privacy note.
- Empty/error states.
- Acceptance criteria.

Next PROMPT 034.
```

## PROMPT 034 — PLATFORM-02 Backend Platform Overview Endpoint If Needed

```text
Task ID: PLATFORM-02 — Backend platform overview endpoint.

Mode backend only.

Implement only if existing /superadmin/summary cannot satisfy contract.

Requirement:
- superadmin only.
- aggregate only.
- no tenant transaction detail.
- no sensitive actions.
- tests authorization and aggregation.

Validate backend.
Next PROMPT 035.
```

## PROMPT 035 — PLATFORM-03 Web Platform Overview Read-only Integration

```text
Task ID: PLATFORM-03 — Web platform overview read-only.

Mode web only.

Scope:
- BFF /api/admin/platform/overview.
- apps/web/src/app/platform/page.tsx.
- platform chart/model helpers.
- tests/docs.

Requirement:
- superadmin only.
- /platform uses backend aggregate if session superadmin.
- fallback preview labeled if no session/backend.
- No sensitive action.
- No tenant detail leak.

Validate web.
Next PROMPT 036.
```

## PROMPT 036 — PLATFORM-04 Toko/Businesses Read-only Integration

```text
Task ID: PLATFORM-04 — Platform toko read-only.

Mode contract then implementation.
Split if large.

Requirement:
- BFF /api/admin/platform/businesses.
- Backend superadmin businesses existing if sufficient.
- Columns: nama toko, owner, masked contact, status, paket, cabang, user, last active, tagihan.
- Actions Lihat ringkasan/Akses bantuan/Nonaktifkan/Arsipkan disabled/gated.
- Akses detail toko requires reason/audit note.

Validate touched area.
Next PROMPT 037.
```

## PROMPT 037 — PLATFORM-05 Paket/Plans Read-only Integration

```text
Task ID: PLATFORM-05 — Platform paket read-only.

Requirement:
- BFF /api/admin/platform/plans.
- Backend superadmin plans existing if sufficient.
- Cards Free/Starter/Pro/Business from backend if available, otherwise preview fallback labeled.
- Edit/Duplicate/Publish disabled unless safe write phase later.
- Harga final if backend has, otherwise label preview.

Validate web/backend if touched.
Next PROMPT 038.
```

## PROMPT 038 — PLATFORM-06 Langganan & Tagihan Read-only Contract

```text
Task ID: PLATFORM-06 — Platform billing read-only contract.

Mode docs-only.

Output docs/API_CONTRACTS/PLATFORM_BILLING_READ.md:
- Needed backend endpoints for subscription/billing rows.
- Existing superadmin business subscription endpoint mapping.
- Missing invoice/gateway log domain if absent.
- BFF /api/admin/platform/billing.
- Status: paid/unpaid/failed/trial/grace/cancelled.
- Actions manual mark paid/retry/send invoice disabled.
- Gateway payment not connected unless real backend exists.

Next PROMPT 039.
```

## PROMPT 039 — PLATFORM-07 Billing Backend Read-only If Needed

```text
Task ID: PLATFORM-07 — Backend platform billing read-only.

Mode backend only if contract says missing.

Requirement:
- superadmin only.
- read-only billing/subscription aggregate.
- no payment retry/manual mark paid.
- no gateway mutation.
- tests authorization and no tenant detail leak.

Validate backend.
Next PROMPT 040.
```

## PROMPT 040 — PLATFORM-08 Billing Web Read-only Integration

```text
Task ID: PLATFORM-08 — Web platform billing read-only.

Mode web only.

Scope platform revenue page + BFF.

Requirement:
- page label Langganan & Tagihan.
- read-only rows.
- payment actions disabled/gated.
- backend unavailable fallback labeled preview.

Validate web.
Next PROMPT 041.
```

## PROMPT 041 — PLATFORM-09 Support/Bantuan Read-only Contract and Stub Backend Decision

```text
Task ID: PLATFORM-09 — Support read-only contract.

Mode docs-only.

Output docs/API_CONTRACTS/PLATFORM_SUPPORT_READ.md:
- support ticket domain needed.
- If backend missing, mark implementation blocked or preview-only.
- Fields: ID, toko, masalah, kategori, prioritas, status, assigned, last update.
- Internal notes no submit until backend exists.
- Minta akses bantuan disabled.

Next PROMPT 042.
```

## PROMPT 042 — PLATFORM-10 Support Web State Alignment

```text
Task ID: PLATFORM-10 — Support page integration or preview-state alignment.

Mode web only unless backend endpoint exists.

If backend endpoint missing:
- Keep page preview-only.
- Add clear banner: backend support ticket belum aktif.
- No real submit.
- No fake production claim.
If backend endpoint exists:
- Add BFF read-only.

Validate web.
Next PROMPT 043.
```

## PROMPT 043 — PLATFORM-11 Activity/Audit Read-only Integration

```text
Task ID: PLATFORM-11 — Platform aktivitas read-only.

Mode contract then implementation.

Requirement:
- Prefer backend audit_logs endpoint if exists; otherwise contract missing and page preview-only.
- BFF /api/admin/platform/activity.
- superadmin only.
- No before/after sensitive diff in main UI.
- Akses bantuan note wajib.

Validate touched area.
Next PROMPT 044.
```

## PROMPT 044 — PLATFORM-12 System Status Read-only Integration

```text
Task ID: PLATFORM-12 — Platform sistem read-only.

Requirement:
- No polling.
- No fetch from external service client-side.
- Backend endpoint if exists, otherwise preview-only with honest copy.
- Cards API, DB, queue, payment gateway, webhook, email/WA, backup, error.

Validate web/backend if touched.
Next PROMPT 045.
```

## PROMPT 045 — PLATFORM-13 Announcements Read-only/Draft Contract

```text
Task ID: PLATFORM-13 — Announcements contract.

Mode docs-only.

Output docs/API_CONTRACTS/PLATFORM_ANNOUNCEMENTS.md:
- draft/read endpoints needed.
- broadcast/send deferred.
- target audience types.
- channel types.
- audit/approval requirement for send.
- composer remains disabled until backend draft endpoint ready.

Next PROMPT 046.
```

## PROMPT 046 — PLATFORM-14 Announcements Web Preview/Read Alignment

```text
Task ID: PLATFORM-14 — Announcements web state alignment.

Mode web only unless read endpoint exists.

Requirement:
- Keep broadcast disabled.
- If backend missing, honest preview copy.
- If read endpoint exists, BFF read-only list.
- No email/WhatsApp/in-app send.

Validate web.
Next PROMPT 047.
```

## PROMPT 047 — PLATFORM-15 Users & Access Read-only Alignment

```text
Task ID: PLATFORM-15 — Platform users route alignment.

Mode web only unless backend endpoint exists.

Requirement:
- Route /platform/users tetap build.
- Not primary menu.
- Operator internal and user toko read-only/preview.
- Reset password/ban disabled.
- Copy privacy user toko.

Validate web.
Next PROMPT 048.
```

## PROMPT 048 — CHK-08 Platform Read-only Checkpoint

```text
Checkpoint Platform Super Admin read-only.

Validate backend if touched, web full validation, boundary scan, docs.
Commit message:
`checkpoint-platform-read-only`.
Push.
Next PROMPT 049.
```

---

# Phase E — Mobile Kasir Hardening

Tujuan fase ini: audit dan perkuat mobile kasir tanpa merusak backend/web. Fokus online-first, shared terminal, actor context, POS checkout, shift, print/share, dan error handling.

## PROMPT 049 — MOB-01 Mobile Current State Audit

```text
Kamu mengerjakan NojPOS.

Task ID: MOB-01 — Mobile kasir current state audit.

Mode docs-only/read-only audit.

Baca:
- AGENTS.md
- PRDPOSJA.md
- docs/STORY_MOBILE_KASIR.md
- docs/BUG_TRACKER.md
- apps/cashier/AGENTS.md
- apps/cashier/lib

Output:
- Update docs/STORY_MOBILE_KASIR.md with current implemented/gap.
- Update docs/BUG_TRACKER.md for real bugs only.
- Identify risky areas: checkout totals, cashier actor, shift state, offline retry, print/share, token storage, PIN.

No runtime changes.
Validate docs.
Next PROMPT 050.
```

## PROMPT 050 — MOB-02 Auth/Device/Terminal Context Hardening

```text
Task ID: MOB-02 — Auth/device/terminal context hardening.

Mode mobile only, backend only if blocker and explicitly needed.

Requirement:
- Ensure mobile does not trust free cashier_id for sensitive writes.
- Terminal/device/outlet context clear.
- Token storage follows existing secure pattern.
- PIN not logged or exposed.
- Session errors render clearly.

Tests:
- Flutter analyze/test.
- Add unit/widget/provider tests if feasible.

No web.
Next PROMPT 051.
```

## PROMPT 051 — MOB-03 POS Checkout Flow Hardening

```text
Task ID: MOB-03 — POS checkout hardening.

Mode mobile first, backend only if unavoidable.

Requirement:
- Flutter must not calculate authoritative money totals.
- Render server quote/transaction totals.
- Handle validation, stale quote, conflict, network error.
- No general offline DB/write queue.
- Retry only bounded per PRD.

Tests:
- flutter analyze/test.
- widget/provider tests for checkout states.

Next PROMPT 052.
```

## PROMPT 052 — MOB-04 Shift and Cash Flow Hardening

```text
Task ID: MOB-04 — Shift/cash hardening.

Requirement:
- Shift open/close UI respects backend state.
- Cash difference displayed safely.
- No duplicate close/open.
- Error states clear.
- Sensitive cash writes need idempotency/audit backend; if missing, document blocker.

Validate mobile; backend if touched.
Next PROMPT 053.
```

## PROMPT 053 — MOB-05 Receipt Print/Share Hardening

```text
Task ID: MOB-05 — Receipt print/share hardening.

Requirement:
- Receipt content safe.
- No payment reference/token/PII leak.
- Print failure handled.
- Share fallback handled.
- No crash when printer unavailable.

Validate mobile.
Next PROMPT 054.
```

## PROMPT 054 — MOB-06 Inventory/Attendance/Reports Mobile Review

```text
Task ID: MOB-06 — Inventory/attendance/reports mobile review.

Requirement:
- Read-only/report screens show backend values.
- Attendance writes verified by role/device if available.
- No stock mutation unless safe backend exists.
- Error/loading/empty states.

Validate mobile.
Next PROMPT 055.
```

## PROMPT 055 — CHK-09 Mobile Hardening Checkpoint

```text
Checkpoint mobile hardening.

Validate:
- flutter analyze
- flutter test
- flutter build apk --debug
- docs/bug tracker updated

Commit message:
`checkpoint-mobile-kasir-hardening`.
Push.
Next PROMPT 056.
```

---

# Phase F — Backend Security, Audit, and Sensitive Action Foundation

Tujuan fase ini: sebelum action sensitif diaktifkan, pastikan tenant scope, audit, idempotency, policy, dan approval/reason framework siap.

## PROMPT 056 — SEC-01 Backend Tenant Scope Audit

```text
Task ID: SEC-01 — Backend tenant scope audit.

Mode backend audit + tests.

Scope:
- apps/backend tests and policy/query fixes only.
- docs/BUG_TRACKER.md.

Requirement:
- Audit /me, /outlets, dashboard summary, transactions, inventory, reports, superadmin routes.
- Add/strengthen tests for cross-tenant access.
- Fix query builder reads missing business_id.
- Do not add new features.

Validate backend.
Next PROMPT 057.
```

## PROMPT 057 — SEC-02 Audit Log and Idempotency Framework Contract

```text
Task ID: SEC-02 — Audit log and idempotency framework contract.

Mode docs-only unless existing code audit needed.

Output:
- docs/API_CONTRACTS/SENSITIVE_ACTIONS.md
- Define action categories.
- Reason required rules.
- Approval gates.
- Idempotency key rules.
- Audit event schema.
- Rollback/transaction rules.
- Which actions remain deferred.

No runtime action.
Next PROMPT 058.
```

## PROMPT 058 — SEC-03 Backend Audit/Idempotency Foundation Implementation

```text
Task ID: SEC-03 — Backend audit/idempotency foundation.

Mode backend only.

Requirement:
- Implement reusable foundation only if contract ready.
- Do not activate all sensitive actions.
- Add tests for idempotency conflict, audit write in transaction, no token/PII logs.

Validate backend.
Next PROMPT 059.
```

## PROMPT 059 — SENS-01 Refund/Void Contract Only

```text
Task ID: SENS-01 — Refund/void contract only.

Mode docs-only.

Output:
- Refund/void domain rules.
- Allowed roles.
- Reason required.
- Idempotency.
- Audit.
- Payment gateway dependency.
- UI confirmation.
- When still blocked.

No implementation.
Next PROMPT 060.
```

## PROMPT 060 — SENS-02 Stock Adjustment Contract Only

```text
Task ID: SENS-02 — Stock adjustment contract only.

Mode docs-only.

Define stock adjustment rules, approval/reason, audit, idempotency, movement ledger, role, UI.
No runtime implementation.
Next PROMPT 061.
```

## PROMPT 061 — SENS-03 Shift Close/Cash Reconciliation Contract Only

```text
Task ID: SENS-03 — Shift close and cash reconciliation contract.

Mode docs-only.

Define shift state machine, cash difference, actor/device/outlet context, idempotency, audit, conflict states.
No runtime implementation unless later approved.
Next PROMPT 062.
```

## PROMPT 062 — SENS-04 Platform Sensitive Actions Contract Only

```text
Task ID: SENS-04 — Platform sensitive actions contract.

Mode docs-only.

Actions:
- akses bantuan/impersonate.
- suspend/activate/archive toko.
- manual mark paid/payment retry.
- reset password/ban user.
- announcement broadcast.

Define reason, duration, approval, audit, role, emergency revoke, UI gating.
No implementation.
Next PROMPT 063.
```

---

# Phase G — Login UI, Local Dev Setup, and UX Completion

## PROMPT 063 — WEB-LOGIN-01 Login UI Contract and UX

```text
Task ID: WEB-LOGIN-01 — Web Admin login UI contract and UX.

Mode docs/design contract first.

Output:
- Login page route decision.
- Role redirect rules: owner/admin -> /dashboard, superadmin -> /platform.
- Error copy.
- Session expired copy.
- No token exposure.
- BFF route usage.
- Acceptance criteria implementation.

No runtime.
Next PROMPT 064.
```

## PROMPT 064 — WEB-LOGIN-02 Implement Login UI Through BFF

```text
Task ID: WEB-LOGIN-02 — Implement login UI through BFF.

Mode web only.

Requirement:
- Page/form calls same-origin BFF only.
- No token in client.
- No localStorage/sessionStorage.
- Role redirect safe.
- Loading/error states.
- Logout visible in shell if applicable.

Validate web full.
Next PROMPT 065.
```

## PROMPT 065 — DEV-01 Local Development Setup Checklist

```text
Task ID: DEV-01 — Local development setup checklist.

Mode docs-only.

Output:
- docs/LOCAL_DEVELOPMENT.md
- How to run backend, web, mobile.
- Required env variables without secrets.
- BACKEND_API_BASE_URL server-only.
- Laravel database/migrate/seed.
- Android emulator API URL.
- Common errors and fixes.
- Validation commands.

No runtime.
Next PROMPT 066.
```

## PROMPT 066 — UX-01 Tenant Admin UI Polish After Real Data

```text
Task ID: UX-01 — Tenant Admin UI polish after read-only real data.

Mode web only.

Requirement:
- Polish empty/error/loading states.
- No horizontal overflow.
- Mobile responsive.
- Copy Indonesia-friendly.
- Do not add new backend features.
- Do not activate sensitive actions.

Validate web including Playwright list and optionally targeted responsive test.
Next PROMPT 067.
```

## PROMPT 067 — UX-02 Platform Super Admin UI Polish After Real Data

```text
Task ID: UX-02 — Platform UI polish after read-only integration.

Mode web only.

Requirement:
- Indonesia-friendly copy.
- Simple internal NojPOS control center.
- Privacy notes clear.
- No enterprise-heavy clutter.
- Actions gated/disabled.
- No horizontal overflow.

Validate web.
Next PROMPT 068.
```

---

# Phase H — Regression, Deployment Prep, and Release Candidate

## PROMPT 068 — QA-01 Full Backend Regression

```text
Task ID: QA-01 — Full backend regression.

Mode QA/fix only.

Run backend tests, inspect failed tests, fix only regressions within scope.
No new features.
Update BUG_TRACKER.
Validate backend.
Next PROMPT 069.
```

## PROMPT 069 — QA-02 Full Web Regression

```text
Task ID: QA-02 — Full web regression.

Mode QA/fix only.

Run typecheck, lint, tests, build, Playwright list, boundary scan.
Fix only regressions/scope bugs.
No new features.
Update BUG_TRACKER.
Next PROMPT 070.
```

## PROMPT 070 — QA-03 Full Mobile Regression

```text
Task ID: QA-03 — Full mobile regression.

Mode QA/fix only.

Run flutter analyze/test/build apk debug.
Fix only regressions/scope bugs.
No new features.
Update BUG_TRACKER.
Next PROMPT 071.
```

## PROMPT 071 — DEPLOY-01 Staging Deployment Plan

```text
Task ID: DEPLOY-01 — Staging deployment plan.

Mode docs-only.

Output:
- docs/STAGING_DEPLOYMENT.md
- Backend deploy checklist.
- Web deploy checklist.
- Mobile build checklist.
- Env variables.
- Migration/seed rules.
- Rollback plan.
- Smoke tests.
- Security reminders.

No runtime deploy unless explicitly instructed.
Next PROMPT 072.
```

## PROMPT 072 — OBS-01 Observability and Logging Review

```text
Task ID: OBS-01 — Observability/logging review.

Mode audit/fix if small.

Check:
- no token/PIN/password/PII logs.
- structured error codes.
- backend error responses standard.
- web BFF safe errors.
- mobile safe logs.

Update BUG_TRACKER.
Validate touched area.
Next PROMPT 073.
```

## PROMPT 073 — DOC-RECON-01 Final Docs Sync

```text
Task ID: DOC-RECON-01 — Final docs sync.

Mode docs-only.

Update:
- PRDPOSJA.md if product decisions changed.
- STORY_PROGRESS.
- STORY_BACKEND.
- STORY_WEB_ADMIN.
- STORY_MOBILE_KASIR.
- STORY_INTEGRATION.
- TASK_QUEUE.
- BUG_TRACKER.
- API_CONTRACTS README.

Remove obsolete references.
Do not change runtime.
Next PROMPT 074.
```

## PROMPT 074 — RC-01 Release Candidate Validation

```text
Task ID: RC-01 — Release candidate validation.

Mode validation only.

Run:
- backend full validation.
- web full validation.
- mobile full validation.
- docs consistency check.
- git status.

Do not implement features.
If all pass, prepare checkpoint commit.
If fail, create targeted repair prompt.
Next PROMPT 075.
```

## PROMPT 075 — RC-02 Final Release Candidate Commit

```text
Task ID: RC-02 — Final release candidate commit.

Mode checkpoint only.

Prerequisite:
- RC-01 all pass.

Actions:
- git status.
- ensure no artifacts staged.
- commit message: `release-candidate-nojpos-admin-integration`.
- push branch.

Laporan akhir:
- commit hash.
- push status.
- validations summary.
- known bugs remaining.
- recommended next production work.
```

---

# Repair Prompt Templates

Gunakan template ini ketika prompt sebelumnya gagal validasi.

## REPAIR-BACKEND — Fix Backend Validation Failure

```text
Kamu mengerjakan NojPOS.

Task ID: REPAIR-BACKEND — Fix backend validation failure from previous task.

Mode:
- Repair only.
- Jangan fitur baru.
- Jangan ubah web/mobile.

Input dari user:
- Paste error php artisan test / route-list / stack trace.

Tugas:
- Reproduce failure.
- Fix minimal.
- Add/update test if needed.
- Update BUG_TRACKER only if blocker remains.

Validate:
- php artisan test
- php artisan route:list --path=api/v1

Report Built, Verified, Blocker.
```

## REPAIR-WEB — Fix Web Validation Failure

```text
Kamu mengerjakan NojPOS.

Task ID: REPAIR-WEB — Fix web validation failure from previous task.

Mode:
- Repair only.
- Jangan fitur baru.
- Jangan ubah backend/mobile.

Input:
- Paste typecheck/lint/test/build/boundary scan error.

Tugas:
- Fix minimal.
- Keep no token leak.
- Keep no localStorage/sessionStorage.
- Keep no direct Laravel from UI.

Validate web full.
Report Built, Verified, Blocker.
```

## REPAIR-MOBILE — Fix Mobile Validation Failure

```text
Kamu mengerjakan NojPOS.

Task ID: REPAIR-MOBILE — Fix mobile validation failure.

Mode repair only.
No backend/web unless error proves contract mismatch and user approves.

Input:
- Paste flutter analyze/test/build error.

Validate:
- flutter analyze
- flutter test
- flutter build apk --debug

Report Built, Verified, Blocker.
```

## REPAIR-DOCS — Fix Docs Consistency Failure

```text
Kamu mengerjakan NojPOS.

Task ID: REPAIR-DOCS — Fix docs consistency.

Mode docs-only.

Tugas:
- Remove obsolete references.
- Align PRD, story docs, task queue, bug tracker, API contracts.
- Do not change runtime.

Validate:
- Docs read check.
- Scan old workflow references.

Report Built, Verified, Blocker.
```

---

# Final Notes For Coding Agents

1. Jangan pernah menganggap prompt berikutnya otomatis boleh dijalankan.
2. Setiap prompt harus selesai dengan validasi dan laporan akhir.
3. Jika menemukan gap kontrak, stop dan update docs dulu.
4. Jika menemukan bug keamanan/tenant leak, stop fitur baru dan buat repair/security task.
5. Jika action sensitif belum punya audit/idempotency/reason/permission, biarkan disabled/gated.
6. Untuk user-facing copy, pakai Bahasa Indonesia sederhana dan jujur.
7. Untuk Platform Super Admin, jangan membuat UI terlihat bebas mengintip detail toko.
8. Untuk Tenant Admin, jangan hitung uang/stok final di client.
9. Untuk Mobile Kasir, jangan membuat offline database umum atau write queue umum.
10. Untuk semua app, jangan klaim produksi jika masih preview/fallback.
