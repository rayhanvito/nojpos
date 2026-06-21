# Story Backend Laravel API

Lokasi: `apps/backend`  
Stack: Laravel API, Sanctum, tenant-scoped POS domain.  
PRD: `PRDPOSJA.md`

Backend adalah sumber kebenaran untuk tenant, outlet, device, role, uang, stok, laporan, langganan, dan audit.

---

## BE-01 — Auth, Tenant Scope, Role, Outlet, Device Context

Status: `[IN PROGRESS]`  
Priority: `P0`  
Dependency: none

### Tujuan

Semua request protected harus mengetahui user, business, outlet, role, device, dan terminal context secara server-side. Client tidak boleh menentukan tenant/actor sendiri.

### Implemented Sekarang

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/me`
- `GET /api/v1/outlets`
- `POST /api/v1/auth/pin-switch`
- `GET /api/v1/terminal/lock-state`
- `POST /api/v1/terminal/lock`
- `POST /api/v1/terminal/unlock`
- Sanctum auth middleware.
- Role middleware untuk owner/admin/superadmin di beberapa route.

### Task dan Subtask

- `[DONE]` Auth API dasar.
  - `[DONE]` Login.
  - `[DONE]` Logout.
  - `[DONE]` Me/profile.
  - `[DONE]` Outlet list.
- `[IN PROGRESS]` Tenant scope hardening.
  - `[TODO]` Audit semua Eloquent dan query builder yang membaca/mengubah tenant data.
  - `[TODO]` Pastikan `business_id` explicit untuk `DB::table` atau query raw.
  - `[TODO]` Tambah feature test tenant A tidak bisa baca/ubah tenant B.
  - `[TODO]` Tambah test 404 vs 403 sesuai resource dan policy.
- `[READY]` Role/outlet/device policy.
  - `[TODO]` Pastikan owner/admin/kasir/superadmin boundary jelas.
  - `[TODO]` Pastikan outlet-bound endpoint validasi outlet milik business.
  - `[TODO]` Pastikan device-bound endpoint tidak menerima device palsu dari client.
- `[DONE]` Web admin browser session strategy.
  - `[DONE]` Pilih Next.js BFF/server-side route handler dengan HttpOnly sealed session cookie.
  - `[DONE]` Definisikan CSRF/origin check, logout, refresh/expiry, dan forbidden state di `docs/API_CONTRACTS/SESSION.md`.
  - `[DONE]` Dokumentasikan cara Web Admin menjaga backend token di server-only boundary tanpa browser-readable auth secret storage.

### Acceptance Criteria

- Semua route protected gagal `401` tanpa session/token valid.
- Role salah menghasilkan `403`.
- Resource tenant lain tidak dapat diakses.
- Endpoint outlet-bound menolak outlet di luar business.
- Auth/session web admin punya keputusan tertulis sebelum integrasi.

### Bug/Kendala

- `[OPEN]` BUG-BE-01 — perlu audit query builder tenant scope.
- `[RESOLVED]` BUG-INT-01 — session strategy web admin sudah final; implementasi adapter lanjut di WEB-01B.

---

## BE-02 — Master Data Toko: Produk, Kategori, Customer, Staff

Status: `[IN PROGRESS]`  
Priority: `P1`  
Dependency: BE-01 tenant scope hardening minimal untuk read-only.

### Tujuan

Owner/admin dapat mengelola master data toko dengan aman, tenant-scoped, dan siap dipakai mobile/web.

### Implemented Sekarang

Products:

- `GET /api/v1/products`
- `POST /api/v1/products`
- `PUT /api/v1/products/{product}`
- `DELETE /api/v1/products/{product}`

Categories:

- `GET /api/v1/categories`
- `POST /api/v1/categories`
- `PUT /api/v1/categories/{category}`
- `DELETE /api/v1/categories/{category}`

Customers:

- `GET /api/v1/customers`
- `POST /api/v1/customers`
- `PUT /api/v1/customers/{customer}`
- `DELETE /api/v1/customers/{customer}`

Staff:

- `GET /api/v1/staff`
- `POST /api/v1/staff`
- `PUT /api/v1/staff/{staff}`
- `DELETE /api/v1/staff/{staff}`
- `POST /api/v1/staff/{staff}/delete`

### Task dan Subtask

- `[DONE]` CRUD route dasar tersedia.
- `[DONE]` Idempotency middleware untuk write route utama.
- `[READY]` Stabilkan kontrak list untuk web admin.
  - `[TODO]` Query params: search, status, category, outlet, pagination, sort.
  - `[TODO]` Response DTO list/table.
  - `[TODO]` Empty, validation, forbidden, not found state.
- `[READY]` Detail route jika UI detail membutuhkan real data.
  - `[TODO]` `GET /api/v1/products/{product}` bila dibutuhkan.
  - `[TODO]` `GET /api/v1/customers/{customer}` bila dibutuhkan.
  - `[TODO]` `GET /api/v1/staff/{staff}` bila dibutuhkan.
- `[TODO]` Staff delete route cleanup.
  - `[TODO]` Pilih satu bentuk delete untuk web admin.
  - `[TODO]` Dokumentasikan route legacy jika masih dipertahankan.

### Acceptance Criteria

- List endpoint dapat dipakai web admin tanpa membuat query berat.
- Data PII customer dimasking jika endpoint/platform membutuhkan redaction.
- Write master data punya idempotency dan validation error jelas.
- Tidak ada cross-tenant read/write.

### Bug/Kendala

- `[OPEN]` DTO list final belum disepakati untuk web admin.
- `[OPEN]` Delete staff punya dua route; perlu penyederhanaan kontrak.

---

## BE-03 — Operasional POS: Inventory, Shift, Transaction, Payment

Status: `[IN PROGRESS]`  
Priority: `P0`  
Dependency: BE-01.

### Tujuan

Transaksi, stok, shift, pembayaran, void, dan refund aman secara uang/stok: server-calculated, transactional, idempotent, audited, dan tenant/outlet/device scoped.

### Implemented Sekarang

Inventory read:

- `GET /api/v1/inventory`
- `GET /api/v1/inventory/movements`
- `GET /api/v1/inventory/transfers`
- `GET /api/v1/inventory/transfers/in-transit`

Inventory write owner/admin:

- `POST /api/v1/inventory/purchases`
- `POST /api/v1/inventory/counts`
- `POST /api/v1/inventory/waste`
- `POST /api/v1/inventory/transfers`
- `POST /api/v1/inventory/transfers/{transfer}/send`
- `POST /api/v1/inventory/transfers/{transfer}/receive`
- `POST /api/v1/inventory/transfers/{transfer}/cancel`

Shift:

- `POST /api/v1/shifts/open`
- `GET /api/v1/shifts/current`
- `POST /api/v1/shifts/{shift}/cash-movements`
- `POST /api/v1/shifts/{shift}/close`

Orders/transactions:

- `GET /api/v1/parked-orders`
- `POST /api/v1/parked-orders`
- `GET /api/v1/parked-orders/{transaction}`
- `PUT /api/v1/parked-orders/{transaction}`
- `DELETE /api/v1/parked-orders/{transaction}`
- `POST /api/v1/parked-orders/{transaction}/lease/acquire`
- `POST /api/v1/parked-orders/{transaction}/lease/refresh`
- `POST /api/v1/parked-orders/{transaction}/lease/release`
- `POST /api/v1/transactions/quote`
- `POST /api/v1/transactions`
- `GET /api/v1/transactions`
- `GET /api/v1/transactions/{transaction}`
- `GET /api/v1/transactions/recovery/{key}`
- `POST /api/v1/transactions/{transaction}/refund`
- `POST /api/v1/payments`
- `POST /api/v1/voids`
- `POST /api/v1/payment-webhooks/{provider}`

### Task dan Subtask

- `[DONE]` Endpoint inti checkout tersedia.
- `[DONE]` Idempotency untuk transaksi, payment, void, refund, shift, inventory writes.
- `[IN PROGRESS]` Money/stock hardening.
  - `[TODO]` Pastikan quote adalah sumber final price snapshot.
  - `[TODO]` Pastikan transaction store menolak quote stale.
  - `[TODO]` Pastikan stock movement terjadi dalam DB transaction.
  - `[TODO]` Pastikan row lock/state conflict untuk shift/payment/stock.
- `[IN PROGRESS]` Audit hardening.
  - `[TODO]` Void/refund/cash movement/stock adjustment audit.
  - `[TODO]` Audit payload redaction untuk PII/payment reference.
- `[READY]` Web admin read-only.
  - `[DONE]` Transaction list read-only contract finalized in `docs/API_CONTRACTS/TRANSACTIONS_READ.md`.
  - `[READY]` Transaction list BFF integration using existing `GET /api/v1/transactions` with BFF sanitization.
  - `[TODO]` Backend-side transaction filter/pagination hardening for high-volume production.
  - `[TODO]` Transaction detail contract/integration if needed after list.
  - `[TODO]` Inventory list/movements/transfers safe for admin web.
  - `[TODO]` Payment read/list endpoint jika web payments page akan real.
- `[DEFERRED]` Web admin sensitive writes.
  - `[TODO]` Refund/void from web.
  - `[TODO]` Stock adjustment from web.
  - `[TODO]` Shift open/close from web.
  - `[TODO]` Payment retry from web.

### Acceptance Criteria

- Tidak ada client-side final money calculation.
- Idempotency same body returns same result; different body returns conflict.
- Failed payment/stock/shift write rollback bersih.
- Void/refund/cash/stock writes audited.
- Web admin tidak bisa memicu sensitive write sebelum gated phase.

### Bug/Kendala

- `[OPEN]` Aksi sensitif belum boleh diaktifkan dari web admin.
- `[OPEN]` `GET /api/v1/transactions` exists and is tenant-scoped, but currently has limited list filters/pagination for Web Admin. BFF must sanitize response and backend hardening is recommended before high-volume production.
- `[OPEN]` Payment read/list endpoint khusus admin belum jelas.

---

## BE-04 — Reports dan Dashboard Summary

Status: `[READY]`  
Priority: `P1`  
Dependency: BE-01, BE-03 data integrity.

### Tujuan

Owner/admin bisa melihat ringkasan bisnis tanpa web admin memanggil terlalu banyak endpoint dan tanpa menghitung laporan di client.

### Implemented Sekarang

- `GET /api/v1/reports/sales-summary`
- `GET /api/v1/reports/sold-products`
- `GET /api/v1/reports/payment-methods`
- `GET /api/v1/reports/cashier-shifts`
- `GET /api/v1/reports/void-refund-audit`
- `GET /api/v1/reports/top-10`

### Task dan Subtask

- `[DONE]` Endpoint report dasar tersedia.
- `[DONE]` BE-04A Dashboard summary API contract.
  - `[DONE]` Endpoint final: `GET /api/v1/dashboard/summary`.
  - `[DONE]` Auth/role/scope: login, owner/admin only, tenant business scope, outlet validation.
  - `[DONE]` Response DTO: `meta`, `kpis`, `alerts`, `sales_last_7_days`, `payment_methods`, `top_products`, `low_stock_items`, `recent_transactions`, `cashier_performance`, `branch_highlights`, `data_notes`.
  - `[DONE]` Empty state dan error state `401`, `403`, `422`, `500`.
  - `[DONE]` BE-04B acceptance criteria documented: read-only, no sensitive trigger, tenant isolation, forbidden role, empty/normal state tests.
- `[DONE]` BE-04B Dashboard summary endpoint implementation.
  - `[DONE]` Implement `GET /api/v1/dashboard/summary` sesuai `docs/API_CONTRACTS/DASHBOARD_SUMMARY.md`.
  - `[DONE]` KPI: sales today, transaction count, average ticket, gross profit estimate, low stock count, cash difference.
  - `[DONE]` Alerts: low stock, out of stock, unclosed shift, cash difference dari data yang tersedia.
  - `[DONE]` Charts: sales last 7 days, payment methods.
  - `[DONE]` Lists: top products, low stock, recent transactions, cashier performance, branch highlights.
  - `[DONE]` Feature tests: owner/admin, cashier/superadmin forbidden, unauthenticated, invalid filter, outlet tenant isolation, empty state, normal state, cross-tenant exclusion.
- `[READY]` Report filter contract.
  - `[TODO]` Outlet filter.
  - `[TODO]` Date range.
  - `[TODO]` Business-day timezone boundary.

### Acceptance Criteria

- Web dashboard dapat render dari satu endpoint agregat atau mapping reports yang disepakati.
- Laporan dihitung server-side.
- Timezone outlet dipakai untuk hari bisnis.
- Gross profit/cash difference punya label estimasi jika sumber data belum final.

### Bug/Kendala

- Tidak ada blocker baru untuk BE-04B. Session strategy sudah final di WEB-01A; web integration tetap menunggu WEB-01B session/BFF adapter scaffold.

---

## BE-05 — Platform Super Admin Core API

Status: `[IN PROGRESS]`  
Priority: `P1`  
Dependency: BE-01 superadmin role boundary.

### Tujuan

Tim internal NojPOS bisa memantau toko, paket, dan langganan dasar tanpa melanggar tenant isolation.

### Implemented Sekarang

- `GET /api/v1/superadmin/summary`
- `GET /api/v1/superadmin/plans`
- `POST /api/v1/superadmin/plans`
- `PUT /api/v1/superadmin/plans/{plan}`
- `PATCH /api/v1/superadmin/plans/{plan}/status`
- `GET /api/v1/superadmin/businesses`
- `POST /api/v1/superadmin/businesses`
- `GET /api/v1/superadmin/businesses/{business}`
- `PATCH /api/v1/superadmin/businesses/{business}`
- `GET /api/v1/superadmin/businesses/{business}/subscription`
- `POST /api/v1/superadmin/businesses/{business}/subscription`

### Task dan Subtask

- `[DONE]` Core Super Admin route tersedia.
- `[READY]` Read-only integration.
  - `[TODO]` `/platform` summary.
  - `[TODO]` `/platform/businesses` toko list.
  - `[TODO]` `/platform/businesses/{business}` detail ringkasan redacted.
  - `[TODO]` `/platform/plans` paket.
- `[READY]` Privacy boundary.
  - `[TODO]` Detail toko hanya ringkasan dan metadata.
  - `[TODO]` Tidak expose detail transaksi tenant lintas toko tanpa access bantuan/audit.
  - `[TODO]` Mask contact/PII di list.
- `[DEFERRED]` Sensitive writes.
  - `[TODO]` Suspend/activate/archive toko.
  - `[TODO]` Change plan/manual extend.
  - `[TODO]` Akses bantuan.

### Acceptance Criteria

- Hanya role superadmin bisa akses `/superadmin/*`.
- Summary menampilkan agregat platform, bukan detail transaksi tenant.
- Sensitive action butuh audit/reason sebelum aktif.

### Bug/Kendala

- `[OPEN]` Super Admin UI lebih lengkap daripada backend core.
- `[OPEN]` Akses bantuan belum boleh real.

---

## BE-06 — Platform Domains: Billing, Support, Activity, System, Announcement, Operator

Status: `[TODO]`  
Priority: `P2`  
Dependency: BE-05 core stable.

### Tujuan

Melengkapi backend domain untuk halaman Platform Super Admin yang sudah ada di UI preview.

### Domain Yang Belum Lengkap

- Billing SaaS/invoice platform.
- Support tickets/bantuan.
- Platform activity logs/audit read endpoint.
- System health summary for UI.
- Announcement draft/target/channel/delivery.
- Operator internal/user access management.

### Task dan Subtask

- `[TODO]` Billing SaaS.
  - `[TODO]` Invoice model/list/detail.
  - `[TODO]` Payment gateway log read-only.
  - `[TODO]` Subscription change history.
  - `[TODO]` Manual billing action tetap gated.
- `[TODO]` Support/Bantuan.
  - `[TODO]` Support ticket model.
  - `[TODO]` Status, priority, category, assignee.
  - `[TODO]` Internal note read/write dengan audit.
  - `[TODO]` Akses bantuan request lifecycle.
- `[TODO]` Activity/Audit.
  - `[TODO]` `GET /api/v1/superadmin/activity-logs`.
  - `[TODO]` Filter category/security/billing/toko/system.
  - `[TODO]` Redaction PII/payment reference.
- `[TODO]` System.
  - `[TODO]` `GET /api/v1/superadmin/system-health`.
  - `[TODO]` API/database/queue/payment/webhook/notification/backup/error status.
  - `[TODO]` Tidak polling dari frontend.
- `[TODO]` Announcement.
  - `[TODO]` Draft model.
  - `[TODO]` Target audience model.
  - `[TODO]` Channel delivery model.
  - `[TODO]` Broadcast hanya setelah approval/audit.
- `[TODO]` Operator/User access.
  - `[TODO]` Operator internal roles.
  - `[TODO]` Tenant user visibility redacted.
  - `[TODO]` Reset/ban action gated.

### Acceptance Criteria

- Semua domain read-only tersedia dulu sebelum write.
- Semua write platform sensitif punya audit, reason, actor, target, before/after.
- Tidak ada endpoint yang membuat Super Admin bisa bebas melihat detail data tenant.

### Bug/Kendala

- `[OPEN]` BUG-BE-03 — domain ini belum ada/baru sebagian.
