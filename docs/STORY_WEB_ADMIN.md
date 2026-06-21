# Story Web Admin Next.js

Lokasi: `apps/web`  
Stack: Next.js, TypeScript, fixture preview, Recharts, Vitest, Playwright.  
PRD: `PRDPOSJA.md`

Web Admin memiliki dua area yang harus tetap terpisah: Tenant Admin untuk owner/admin toko dan Platform Super Admin untuk tim internal NojPOS.

---

## WEB-01 — Tenant Dashboard Owner/Admin

Status: `[INTEGRATED READ-ONLY]`  
Priority: `P1`  
Dependency: INT-01, BE-04.

### Tujuan

Owner/admin toko bisa memahami kondisi bisnis hari ini: penjualan, transaksi, rata-rata transaksi, gross profit estimasi, stok kritis, selisih kas, produk laku, kasir, dan cabang.

### Implemented Sekarang

- Route: `/dashboard`.
- UI dashboard sudah sederhana dan Indonesia-friendly.
- Data real read-only sekarang dibaca melalui Next.js BFF route `GET /api/admin/dashboard/summary`.
- Fixture lokal `apps/web/src/fixtures/preview/dashboard.ts` tetap ada hanya sebagai fallback/preview berlabel jelas saat session/backend belum tersedia.
- Chart Recharts client component:
  - `DashboardSalesChart`.
  - `DashboardPaymentChart`.
- Quick links statis ke transaksi, inventory, catalog, reports, outlets, staff.

### Task dan Subtask

- `[DONE]` Header ringkas dan context chip.
- `[DONE]` 6 KPI utama.
- `[DONE]` Alert penting “Perlu dicek”.
- `[DONE]` Chart sales 7 hari dan metode pembayaran.
- `[DONE]` Ringkasan operasional.
- `[DONE]` Quick links statis.
- `[DONE]` WEB-01A session strategy decision documented in `docs/API_CONTRACTS/SESSION.md`.
- `[DONE]` WEB-01B implement session/BFF adapter scaffold.
  - `[DONE]` Buat server-only sealed HttpOnly session adapter sesuai strategi Next.js BFF.
  - `[DONE]` Buat route handler login/logout/session tanpa integrasi dashboard.
  - `[DONE]` Pastikan tidak ada token mentah ke Client Component atau browser-readable storage.
- `[DONE]` WEB-01C integrasi read-only dashboard melalui BFF.
  - `[DONE]` Session/auth strategy web admin selesai di WEB-01A.
  - `[DONE]` Implementasi `GET /api/v1/dashboard/summary` selesai di BE-04B; kontrak BE-04A sudah final di `docs/API_CONTRACTS/DASHBOARD_SUMMARY.md`.
  - `[DONE]` Buat adapter data dashboard sesuai kontrak final melalui BFF.
  - `[DONE]` Tambah empty state dari backend dan fallback session-required.
  - `[DONE]` Tambah error/server unavailable state dengan label data contoh fallback.
  - `[DONE]` Tambah 401/403 state tanpa menampilkan data backend tenant.
- `[DEFERRED]` Action real dari dashboard.
  - `[TODO]` Jangan aktifkan shift close/open.
  - `[TODO]` Jangan aktifkan stock adjustment.
  - `[TODO]` Jangan aktifkan export/report generation berat.

### Acceptance Criteria

- Dashboard real data hanya melalui BFF/server-side helper yang disetujui.
- Gross profit dan selisih kas tetap berlabel estimasi jika backend belum final.
- Tidak ada horizontal overflow mobile/desktop.
- Tidak ada API call di random component.
- Token backend tidak masuk Client Component, props UI, fixture, atau browser-readable storage.

### Bug/Kendala

- `[RESOLVED]` Dashboard tidak lagi hanya fixture lokal; WEB-01C membaca backend summary melalui BFF dan memakai fixture hanya sebagai fallback berlabel jelas.
- `[RESOLVED]` Session strategy Web Admin sudah diputuskan di WEB-01A: Next.js BFF/server-side route handler dengan HttpOnly sealed session cookie.
- `[RESOLVED]` WEB-01B session/BFF adapter scaffold sudah tersedia: `/api/admin/auth/login`, `/api/admin/auth/logout`, dan `/api/admin/session`.
- `[RESOLVED]` Endpoint dashboard summary sudah diimplementasikan di BE-04B dan Web Dashboard read-only integration selesai di WEB-01C.

---

## WEB-02 — Tenant Admin Management Pages

Status: `[PREVIEW]`  
Priority: `P1`  
Dependency: INT-01, BE-02, BE-03, BE-04.

### Tujuan

Owner/admin bisa mengakses transaksi, inventori, produk, laporan, cabang, staff, customer, pembayaran, promosi, absensi, subscription, dan pengaturan dari web admin.

### Implemented Sekarang

Route web admin tersedia:

- `/transactions`, `/transactions/[transactionId]`
- `/inventory`, `/inventory/[resourceId]`
- `/catalog`, `/catalog/[productId]`
- `/customers`, `/customers/[customerId]`
- `/staff`, `/staff/[staffId]`
- `/attendance`, `/attendance/[recordId]`
- `/reports`
- `/payments`, `/payments/[paymentId]`
- `/promotions`, `/promotions/[promotionId]`
- `/outlets`, `/outlets/[outletId]`
- `/settings`, `/settings/business`, `/settings/outlet`, `/settings/payments`, `/settings/receipt`
- `/subscription`
- `/audit`, `/audit/[eventId]`

### Task dan Subtask

- `[DONE]` Shell Tenant Admin preview.
- `[DONE]` Route manifest mengenali Tenant Admin routes.
- `[READY]` Read-only integration order.
  - `[DONE]` Transactions list contract in `docs/API_CONTRACTS/TRANSACTIONS_READ.md`.
  - `[DONE]` Transactions list via BFF `GET /api/admin/transactions`.
  - `[TODO]` Transactions detail contract/integration if needed after list.
  - `[DONE]` Inventory list/movements/transfers contract in `docs/API_CONTRACTS/INVENTORY_READ.md`.
  - `[DONE]` Inventory list via BFF `GET /api/admin/inventory`.
  - `[DONE]` Products/categories contract in `docs/API_CONTRACTS/CATALOG_READ.md`.
  - `[DONE]` Products/categories list via BFF `GET /api/admin/catalog/products` and `GET /api/admin/catalog/categories`.
  - `[TODO]` Customers list/detail.
  - `[TODO]` Staff list/detail.
  - `[TODO]` Attendance list.
  - `[DONE]` Reports contract in `docs/API_CONTRACTS/REPORTS_READ.md`.
  - `[DONE]` Reports via BFF `/api/admin/reports/*`.
  - `[TODO]` Settings/outlets/subscription.
  - `[TODO]` Payments read-only jika backend read endpoint tersedia.
- `[READY]` API state pattern.
  - `[TODO]` Loading.
  - `[TODO]` Empty.
  - `[TODO]` Error.
  - `[TODO]` 401.
  - `[TODO]` 403.
  - `[TODO]` 404.
  - `[TODO]` 409/422 for future forms.
- `[DEFERRED]` Safe write phase.
  - `[TODO]` Product/category/customer create/update.
  - `[TODO]` Settings update owner/admin.
  - `[TODO]` Staff update owner/admin.
- `[BLOCKED]` Sensitive write phase.
  - `[TODO]` Refund/void.
  - `[TODO]` Reprint/export.
  - `[TODO]` Shift close/open.
  - `[TODO]` Stock adjustment.
  - `[TODO]` Payment retry.
  - `[TODO]` Cash reconciliation.

### Acceptance Criteria

- Tenant Admin tidak memakai endpoint Platform Super Admin.
- Route read-only tidak mengubah state.
- Sensitive action tetap disabled/gated sampai phase khusus.
- PII/customer/payment data ditampilkan sesuai policy.

### Bug/Kendala

- `[RESOLVED]` Session/BFF adapter awal sudah tersedia di server-only boundary; page/component tetap belum memakai API langsung.
- `[OPEN]` Transactions backend route exists, but current backend list filter/pagination is limited; `GET /api/admin/transactions` now sanitizes/maps/paginates safely in BFF and backend-side filter/pagination hardening remains recommended before high-volume production.
- `[OPEN]` Inventory backend read routes exist, but pagination totals, `category_id`/`stock_status` filters, and low-stock threshold are not final; `GET /api/admin/inventory` now sanitizes/maps/paginates safely in BFF and backend-side hardening remains recommended before high-volume production.
- `[OPEN]` Catalog backend read routes exist, but products/categories pagination, `category_id` filter, status, updated_at, and stock summary are limited; `GET /api/admin/catalog/products` and `GET /api/admin/catalog/categories` now sanitize/map/paginate safely in BFF and keep write actions disabled.
- `[OPEN]` Payment/promotions web route belum jelas endpoint backend realnya.

---

## WEB-03 — Platform Super Admin Ringkasan

Status: `[PREVIEW]`  
Priority: `P1`  
Dependency: INT-01, BE-05.

### Tujuan

Tim internal NojPOS bisa melihat toko, paket, langganan/tagihan, bantuan, aktivitas, sistem, dan pengumuman dengan bahasa Indonesia sederhana.

### Implemented Sekarang

- Route: `/platform`.
- Sidebar utama:
  - Ringkasan.
  - Toko.
  - Paket.
  - Langganan & Tagihan.
  - Bantuan.
  - Aktivitas.
  - Sistem.
  - Pengumuman.
- Fixture lokal: `apps/web/src/fixtures/preview/platform.ts`.
- Chart Recharts mini: `PlatformMiniChart`.
- CSS fix untuk platform shell overflow.

### Task dan Subtask

- `[DONE]` UI Ringkasan Platform.
- `[DONE]` KPI maksimal 8.
- `[DONE]` Yang perlu dicek.
- `[DONE]` Privacy/tenant isolation card.
- `[DONE]` Status toko dan status tagihan chart kecil.
- `[DONE]` Layout overflow fix Super Admin.
- `[READY]` Read-only integration.
  - `[TODO]` Map summary ke `GET /api/v1/superadmin/summary`.
  - `[TODO]` Tambah loading/empty/error/forbidden state.
  - `[TODO]` Pastikan agregat tidak membuka transaksi detail lintas toko.

### Acceptance Criteria

- Ringkasan hanya menampilkan agregat platform.
- Copy menyatakan data sensitif toko terlindungi.
- Route tidak overflow di mobile/desktop.

### Bug/Kendala

- `[RESOLVED]` Platform shell overflow/blank kanan sudah dipatch.
- `[OPEN]` Super Admin masih fixture lokal.

---

## WEB-04 — Platform Toko dan Paket

Status: `[PREVIEW]`  
Priority: `P1`  
Dependency: INT-01, BE-05.

### Tujuan

Tim internal bisa memantau daftar toko dan paket NojPOS tanpa mengintip data sensitif toko.

### Implemented Sekarang

- `/platform/businesses` label UI: Toko.
- `/platform/plans` label UI: Paket.
- Action sensitif disabled/gated.
- Istilah “impersonate” diganti menjadi “Akses bantuan”.

### Task dan Subtask

- `[DONE]` Toko table/list preview.
- `[DONE]` Filter chips status toko.
- `[DONE]` Gate copy akses bantuan.
- `[DONE]` Paket cards Free/Starter/Pro/Business.
- `[READY]` Integrasi read-only.
  - `[TODO]` Toko: `GET /api/v1/superadmin/businesses`.
  - `[TODO]` Detail toko: `GET /api/v1/superadmin/businesses/{business}`.
  - `[TODO]` Paket: `GET /api/v1/superadmin/plans`.
  - `[TODO]` Subscription per business: `GET /api/v1/superadmin/businesses/{business}/subscription`.
- `[DEFERRED]` Write actions.
  - `[TODO]` Update status toko dengan audit.
  - `[TODO]` Publish/edit paket dengan audit.
  - `[TODO]` Akses bantuan dengan reason/duration/audit.

### Acceptance Criteria

- Data kontak toko dimasking jika perlu.
- Tidak ada detail transaksi tenant lintas toko.
- Akses bantuan tetap disabled sampai backend approval/audit siap.

### Bug/Kendala

- `[OPEN]` UI detail toko belum boleh membuka data transaksi detail lintas toko tanpa alasan/audit.

---

## WEB-05 — Platform Langganan & Tagihan

Status: `[PREVIEW]`  
Priority: `P2`  
Dependency: BE-05, BE-06.

### Tujuan

Tim internal bisa melihat paket aktif, jatuh tempo, status pembayaran, dan tagihan bermasalah.

### Implemented Sekarang

- `/platform/revenue` dipakai sebagai “Langganan & Tagihan”.
- `/platform/subscriptions` tetap ada sebagai route lama.
- Actions manual mark paid, retry, extend, kirim ulang tagihan disabled/gated.

### Task dan Subtask

- `[DONE]` KPI tagihan preview.
- `[DONE]` Table tagihan preview.
- `[DONE]` Copy gateway payment belum terhubung.
- `[READY]` Integrasi partial.
  - `[TODO]` Map subscription per business ke endpoint yang sudah ada jika konteks toko dipilih.
  - `[TODO]` Backend perlu endpoint billing/invoice list platform.
  - `[TODO]` Backend perlu payment gateway logs read-only.
- `[BLOCKED]` Billing actions.
  - `[TODO]` Tandai lunas.
  - `[TODO]` Perpanjang manual.
  - `[TODO]` Kirim ulang tagihan.
  - `[TODO]` Payment retry.

### Acceptance Criteria

- Gateway payment data tidak diklaim real sebelum endpoint tersedia.
- Manual billing actions tidak aktif tanpa audit/idempotency.
- Billing table dapat dipakai read-only terlebih dahulu.

### Bug/Kendala

- `[OPEN]` Backend belum punya invoice/platform billing list endpoint.
- `[OPEN]` Payment retry/manual mark paid belum boleh real.

---

## WEB-06 — Platform Bantuan, Aktivitas, Sistem, Pengumuman, User

Status: `[PREVIEW]`  
Priority: `P2`  
Dependency: BE-06.

### Tujuan

Super Admin punya pusat kerja internal NojPOS untuk bantuan pelanggan, riwayat aktivitas, sistem, pengumuman, dan user/access.

### Implemented Sekarang

- `/platform/support`.
- `/platform/audit` label UI: Aktivitas.
- `/platform/system-health` label UI: Sistem.
- `/platform/announcements` label UI: Pengumuman.
- `/platform/users` tetap ada sebagai route utility.

### Task dan Subtask

- `[DONE]` Support/Bantuan UI preview.
- `[DONE]` Aktivitas UI preview.
- `[DONE]` Sistem UI preview.
- `[DONE]` Pengumuman UI preview.
- `[DONE]` User & Akses UI preview.
- `[TODO]` Backend support ticket domain.
- `[TODO]` Backend platform activity logs endpoint.
- `[TODO]` Backend system health endpoint.
- `[TODO]` Backend announcement draft/broadcast domain.
- `[TODO]` Backend operator/user access endpoint.
- `[BLOCKED]` Sensitive actions.
  - `[TODO]` Assign/tandai selesai ticket.
  - `[TODO]` Minta akses bantuan.
  - `[TODO]` Broadcast pengumuman.
  - `[TODO]` Reset password.
  - `[TODO]` Ban user.

### Acceptance Criteria

- Semua halaman bisa read-only dulu.
- Write/action sensitif tetap disabled sampai backend domain dan audit siap.
- Aktivitas menampilkan alasan, pelaku, target, waktu, level, device/IP bila tersedia.

### Bug/Kendala

- `[OPEN]` Semua halaman ini masih fixture lokal.
- `[OPEN]` Announcement broadcast, akses bantuan, reset password, ban user harus tetap disabled sampai backend/audit siap.

---

## WEB-07 — Web Admin Quality Gate

Status: `[IN PROGRESS]`  
Priority: `P0`  
Dependency: none.

### Tujuan

Menjaga web admin tetap stabil, responsive, dan tidak mengaktifkan backend/API secara sembarangan.

### Task dan Subtask

- `[DONE]` `apps/web` sudah dipindahkan ke monorepo `apps/web`.
- `[DONE]` Typecheck/lint/test/build/Playwright list pernah lolos dari lokasi baru.
- `[DONE]` Boundary scan app/component/fixture bersih dari fetch/storage/auth import saat UI preview.
- `[READY]` API integration boundary.
  - `[TODO]` Setelah integrasi dimulai, API adapter hanya boleh berada di layer yang disetujui.
  - `[TODO]` Update boundary scan agar page/component tidak langsung memakai fetch/storage.
  - `[TODO]` Tambah test loading/error/forbidden state.
- `[READY]` Responsive QA.
  - `[TODO]` Tenant Admin route no horizontal overflow.
  - `[TODO]` Platform Super Admin route no horizontal overflow.

### Acceptance Criteria

- `npm run typecheck`, `npm run lint`, `npm test`, `npm run build`, dan `npx playwright test --list` lolos sebelum laporan done.
- Jika integrasi API dimulai, layer API jelas dan tidak menyebar ke UI component.

### Bug/Kendala

- `[OPEN]` Folder lama `nojpos_admin_web/apps/web` masih menyisakan cache/dependency terkunci proses.
