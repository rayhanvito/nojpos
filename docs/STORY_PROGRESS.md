# NojPOS Story Progress

Board ini adalah urutan kerja aktif agar implementasi tidak loncat-loncat. Detail task harian ada di `docs/TASK_QUEUE.md`.

## Current Product Status

| Area | Status | Ringkasan |
| --- | --- | --- |
| Backend Core POS | `[IN PROGRESS]` | Route inti auth, master data, inventory, shift, transaction, payment, reports, superadmin core, dan dashboard summary read-only sudah ada. Perlu hardening tenant scope, policy, audit, dan endpoint platform tambahan. |
| Mobile Kasir | `[IN PROGRESS]` | Feature Flutter utama sudah ada. Perlu QA end-to-end, hardening checkout retry, device/session, shift, print. |
| Web Admin Tenant | `[INTEGRATED READ-ONLY]` | Dashboard, transactions, inventory, dan catalog owner/admin sudah membaca data read-only melalui Next.js BFF; reports/customers/staff/outlets masih preview/read-only planning. |
| Platform Super Admin | `[PREVIEW]` | UI sederhana Indonesia-friendly sudah preview. Backend baru mendukung sebagian kecil. |
| Integrasi Web Admin | `[IN PROGRESS]` | Session strategy dan BFF adapter sudah selesai; dashboard, transactions, inventory, dan catalog read-only integration selesai; reports read-only contract ready, lanjut reports BFF/page integration. |
| Billing SaaS | `[TODO]` | Belum ada invoice/platform billing domain lengkap. |
| Support/Bantuan | `[TODO]` | Belum ada support ticket backend domain. |
| Pengumuman | `[TODO]` | Belum ada announcement backend domain. |

## Working Mode

Status: `[IN PROGRESS]`

NojPOS sekarang memakai **single active coding agent**.

Aturan aktif:

- Semua task diambil dari `docs/TASK_QUEUE.md`.
- Hanya satu task boleh `[IN PROGRESS]`.
- Jangan menjalankan beberapa coding agent paralel dulu.
- Backend, web, dan mobile dikerjakan berurutan sesuai dependency.
- Sensitive actions tetap deferred sampai audit/policy/idempotency/reason/approval siap.

## Implementation Order Wajib

### Phase 0 — Docs dan Kontrak

Status: `[DONE]`

- `[DONE]` Pindahkan web admin ke `apps/web`.
- `[DONE]` Buat `PRDPOSJA.md` sebagai PRD aktif.
- `[DONE]` Rapikan docs menjadi story-based tracker.
- `[DONE]` Ganti workflow menjadi single-agent sequential workflow.
- `[DONE]` Buat `docs/TASK_QUEUE.md` final.
- `[READY AFTER DOC-01]` Review story dengan product owner.
- `[DONE]` Finalisasi kontrak dashboard summary.
- `[DONE]` Finalisasi session strategy web admin.

### Phase 1 — Backend Foundation Hardening

Status: `[IN PROGRESS]`

Dependency: Phase 0 minimal selesai.

- Story: `BE-01`, `BE-04`.
- Fokus:
  - tenant scope audit,
  - role/outlet/device context,
  - standard error/envelope,
  - `[DONE]` dashboard summary endpoint,
  - policy test untuk endpoint yang akan dibuka ke web admin.

### Phase 2 — Tenant Dashboard Real Data

Status: `[DONE]`

Dependency: backend dashboard endpoint + web session strategy.

- Story: `BE-04`, `WEB-01`, `INT-02`.
- Fokus:
  - `[DONE]` `GET /api/v1/dashboard/summary`,
  - `[DONE]` `/dashboard` read-only integration via BFF,
  - `[DONE]` loading/empty/error/forbidden/session-required state,
  - `[DONE]` labels untuk estimasi gross profit dan selisih kas.

### Phase 3 — Tenant Admin Read-only Pages

Status: `[IN PROGRESS]`

- Story: `WEB-02`, `INT-03`.
- Urutan:
  1. `[DONE]` Transactions read-only contract.
  2. `[DONE]` Transactions BFF + page read-only integration.
  3. `[DONE]` Inventory read-only contract.
  4. `[DONE]` Inventory BFF + page read-only integration.
  5. `[DONE]` Catalog/products/categories read-only contract.
  6. `[DONE]` Catalog/products/categories BFF + page read-only integration.
  7. `[DONE]` Reports read-only contract.
  8. `[DONE]` Reports BFF + page read-only integration.
  9. `[DONE]` Staff/outlets/customer read-only contracts.
  10. `[READY]` Staff/outlets/customer BFF + page read-only integration.
  11. Attendance.
  12. Settings/subscription.

### Phase 4 — Mobile Kasir Hardening

Status: `[READY]`

Dependency: backend auth/checkout/shift endpoint stabil.

- Story: `MOB-01`, `MOB-02`, `MOB-03`.
- Fokus:
  - login/outlet/device/PIN switch QA,
  - checkout quote/store/payment QA,
  - shift open/current/close QA,
  - bounded retry tidak berubah menjadi offline queue.

### Phase 5 — Tenant Admin Safe Writes

Status: `[DEFERRED]`

Dependency: read-only pages stabil, policy/idempotency/test siap.

- Story: `INT-04`.
- Boleh mulai dari:
  - products/categories/customer create/update,
  - settings update,
  - staff update.
- Jangan dulu:
  - refund,
  - void,
  - shift open/close from web,
  - stock adjustment,
  - payment retry,
  - export besar.

### Phase 6 — Platform Super Admin Read-only

Status: `[DEFERRED]`

- Story: `WEB-03`, `WEB-04`, `WEB-05`, `WEB-06`, `INT-05`, `BE-05`, `BE-06`.
- Urutan:
  1. `/platform` summary.
  2. `/platform/businesses` toko.
  3. `/platform/plans` paket.
  4. `/platform/revenue` partial subscription/tagihan.
  5. Support/activity/system/announcement setelah backend domain ada.

### Phase 7 — Sensitive Actions

Status: `[DEFERRED]`

- Story: `INT-06`.
- Hanya setelah audit, approval, reason, idempotency, dan policy siap.

## Bug Aktif Ringkas

Lihat detail di `docs/BUG_TRACKER.md`.

| Bug | Status | Dampak |
| --- | --- | --- |
| BUG-BE-01 | `[OPEN]` | Perlu audit tenant scope query builder. |
| BUG-BE-03 | `[OPEN]` | Platform billing/support/activity/system/announcement backend belum ada. |
| BUG-WEB-02 | `[OPEN]` | Folder lama `nojpos_admin_web/apps/web` masih menyisakan cache/dependency terkunci proses. |
| BUG-BE-03 | `[OPEN]` | Platform billing/support/activity/system/announcement backend belum ada. |
| BUG-MOB-01 | `[OPEN]` | Mobile butuh QA device untuk lock/PIN/print/retry. |

## Definition of Ready Untuk Task Baru

Task boleh dikerjakan jika:

- Ada di `docs/TASK_QUEUE.md`.
- Story dan dependency jelas.
- Endpoint/route target jelas jika menyentuh API.
- Risiko tenant/money/stock/action sensitif sudah dipetakan.
- Acceptance criteria ada.
- Test/validasi yang harus dijalankan jelas.

## Definition of Done Untuk Task Baru

Task selesai jika:

- Implementasi sesuai `PRDPOSJA.md`.
- Test relevan lolos.
- Tidak ada regression di area terkait.
- Docs/story/bug tracker diperbarui.
- Laporan akhir mencantumkan Built, Verified, Blocker, dan Next.
