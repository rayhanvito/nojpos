# Story Integrasi NojPOS

PRD: `PRDPOSJA.md`  
Tujuan dokumen ini adalah membuat urutan integrasi jelas supaya implementasi AI/engineer tidak loncat-loncat.

Integrasi harus dimulai read-only, lalu safe writes, lalu sensitive actions paling akhir.

---

## INT-01 — Web Admin Session Integration

Status: `[BLOCKED]`  
Priority: `P0`  
Dependency: BE-01 decision.

### Tujuan

Web Admin dapat mengenali user, role, business, outlet list, dan akses tenant/platform secara aman tanpa menyimpan token di `localStorage`/`sessionStorage`.

### Keputusan Yang Harus Dibuat

- `[TODO]` Pilih model session:
  - cookie HttpOnly/Sanctum SPA, atau
  - Next.js BFF/server boundary, atau
  - model lain yang tetap tidak menyimpan token di browser storage.
- `[TODO]` CSRF strategy.
- `[TODO]` Logout/expiry behavior.
- `[TODO]` 401/403 UI behavior.
- `[TODO]` Role routing:
  - owner/admin ke Tenant Admin,
  - superadmin ke Platform Admin.

### Task dan Subtask

- `[TODO]` Buat dokumen keputusan session singkat.
- `[TODO]` Siapkan API adapter layer yang disetujui.
- `[TODO]` Integrasi read-only:
  - `GET /api/v1/me`.
  - `GET /api/v1/outlets`.
  - `GET /api/v1/subscription`.
  - `GET /api/v1/settings`.
- `[TODO]` Tambah loading/empty/error/forbidden state di shell.
- `[TODO]` Test session expired.
- `[TODO]` Test role denied.

### Acceptance Criteria

- Tidak ada token di localStorage/sessionStorage.
- Page/component tidak fetch langsung jika boundary melarang.
- Tenant Admin dan Platform Super Admin routing jelas.
- Logout membersihkan session dengan aman.

### Bug/Kendala

- `[OPEN]` BUG-INT-01 — session strategy belum final.

---

## INT-02 — Tenant Dashboard Read-only Integration

Status: `[READY AFTER INT-01]`  
Priority: `P1`  
Dependency: INT-01, BE-04, WEB-01.

### Tujuan

Mengganti fixture `/dashboard` dengan data backend read-only tanpa mengaktifkan action sensitif.

### Pilihan Integrasi

Preferred:

- `[TODO]` Buat `GET /api/v1/dashboard/summary`.

Fallback sementara jika endpoint agregat belum dibuat:

- `[TODO]` Map beberapa report endpoint:
  - `/reports/sales-summary`
  - `/reports/top-10`
  - `/reports/payment-methods`
  - `/reports/cashier-shifts`
  - `/transactions`
  - `/inventory`

### Task dan Subtask

- `[TODO]` Definisikan DTO dashboard summary.
- `[TODO]` Backend endpoint/mapping.
- `[TODO]` Web adapter.
- `[TODO]` Replace fixture with server data.
- `[TODO]` Keep local fixture only for tests/storybook-like preview jika perlu.
- `[TODO]` Add loading/empty/error/forbidden.
- `[TODO]` QA responsive.

### Acceptance Criteria

- Dashboard memuat data sesuai outlet/business user.
- Web tidak menghitung laporan final.
- Gross profit/cash difference diberi label estimasi bila sesuai.
- Tidak ada action real dari dashboard.

### Bug/Kendala

- `[OPEN]` Dashboard summary endpoint belum ada.

---

## INT-03 — Tenant Admin Read-only Pages

Status: `[READY AFTER INT-02]`  
Priority: `P1`  
Dependency: INT-01, BE-02, BE-03, BE-04, WEB-02.

### Tujuan

Menghubungkan halaman tenant admin ke backend secara read-only dulu.

### Urutan Integrasi

1. Transactions list/detail.
2. Inventory list/movements/transfers.
3. Catalog/products/categories.
4. Customers.
5. Staff/attendance.
6. Reports.
7. Settings/outlets/subscription.
8. Payments read-only jika endpoint tersedia.
9. Promotions hanya setelah backend jelas.

### Task dan Subtask

- `[TODO]` Buat typed DTO per endpoint.
- `[TODO]` Buat adapter per feature.
- `[TODO]` Tambah state loading/empty/error/forbidden/not found.
- `[TODO]` Tambah test route render.
- `[TODO]` Tambah responsive QA jika layout berubah.
- `[TODO]` Update docs kontrak endpoint yang dipakai.

### Acceptance Criteria

- Semua integration read-only tidak mengubah state backend.
- Tenant data scoped by session.
- 403 dan 404 dibedakan dengan benar.
- UI tetap usable saat data kosong.

### Bug/Kendala

- `[OPEN]` DTO/pagination/filter belum final untuk semua list.
- `[OPEN]` Payments/promotions route belum punya backend mapping jelas.

---

## INT-04 — Tenant Admin Safe Writes

Status: `[DEFERRED]`  
Priority: `P2`  
Dependency: INT-03 stable, backend policy/idempotency/test ready.

### Tujuan

Mengaktifkan write rendah risiko untuk owner/admin setelah read-only stabil.

### Boleh Dipertimbangkan Lebih Dulu

- Product create/update.
- Category create/update.
- Customer create/update.
- Settings update.
- Staff update owner/admin.

### Jangan Dulu

- Refund.
- Void.
- Reprint jika trigger sensitive flow.
- Shift close/open.
- Stock adjustment.
- Payment retry.
- Cash reconciliation.
- Export/report heavy job.

### Task dan Subtask

- `[TODO]` Pilih satu feature write pertama.
- `[TODO]` Pastikan endpoint idempotent.
- `[TODO]` Pastikan policy test ada.
- `[TODO]` Tambah form validation client.
- `[TODO]` Tambah server validation mapping.
- `[TODO]` Tambah success/error state.
- `[TODO]` Tambah audit jika action termasuk privileged.

### Acceptance Criteria

- Double submit tidak membuat data duplikat.
- Validation error 422 tampil jelas.
- Forbidden 403 tampil jelas.
- Tidak mengaktifkan sensitive action.

### Bug/Kendala

- `[OPEN]` Belum semua endpoint write punya kontrak UI/error mapping yang stabil.

---

## INT-05 — Super Admin Read-only Integration

Status: `[READY AFTER INT-01]`  
Priority: `P2`  
Dependency: INT-01, BE-05, WEB-03, WEB-04.

### Tujuan

Menghubungkan Platform Super Admin ke backend read-only tanpa membuka detail tenant sensitif.

### Endpoint Yang Sudah Bisa Dipetakan

- `GET /api/v1/superadmin/summary` → `/platform`.
- `GET /api/v1/superadmin/businesses` → `/platform/businesses`.
- `GET /api/v1/superadmin/businesses/{business}` → detail toko/ringkasan.
- `GET /api/v1/superadmin/plans` → `/platform/plans`.
- `GET /api/v1/superadmin/businesses/{business}/subscription` → subscription per toko.

### Endpoint Yang Masih Dibutuhkan

- Platform billing/invoice list.
- Support tickets.
- Activity logs.
- System health.
- Announcements.
- Operator/user access.

### Task dan Subtask

- `[TODO]` Integrasi summary read-only.
- `[TODO]` Integrasi businesses read-only.
- `[TODO]` Integrasi plans read-only.
- `[TODO]` Integrasi subscription per business jika context tersedia.
- `[TODO]` Pertahankan preview/empty unavailable untuk domain yang belum ada backend.
- `[TODO]` Mask contact/PII.

### Acceptance Criteria

- Platform route hanya superadmin.
- Summary hanya agregat platform.
- Detail toko tidak membuka data transaksi detail tanpa access bantuan/audit.
- Domain tanpa endpoint tetap jelas sebagai belum tersedia, bukan data palsu.

### Bug/Kendala

- `[OPEN]` Backend belum mencakup banyak halaman Platform Super Admin.

---

## INT-06 — Super Admin Sensitive Actions

Status: `[DEFERRED]`  
Priority: `P3`  
Dependency: INT-05 stable, BE-06 audit/approval/idempotency ready.

### Tujuan

Mengaktifkan aksi internal NojPOS secara aman setelah backend audit dan approval siap.

### Aksi Ditunda

- Akses bantuan.
- Suspend/activate/archive toko.
- Edit/publish paket.
- Change plan/manual extend.
- Manual mark paid.
- Retry payment.
- Assign/tandai selesai support.
- Reset password.
- Ban user.
- Announcement broadcast.

### Subtask Wajib Sebelum Aktif

- `[TODO]` Reason wajib.
- `[TODO]` Duration/scope wajib untuk akses bantuan.
- `[TODO]` Approval flow jika diperlukan.
- `[TODO]` Audit log immutable.
- `[TODO]` Before/after snapshot redacted.
- `[TODO]` Idempotency.
- `[TODO]` Policy and role test.
- `[TODO]` UI confirmation/gated state.

### Acceptance Criteria

- Tidak ada sensitive action yang bisa jalan tanpa reason/audit.
- Semua action bisa dilacak ke actor, target, waktu, IP/device, dan hasil.
- Error state tidak menyebabkan aksi partial tanpa rollback.

### Bug/Kendala

- `[OPEN]` Belum ada backend domain lengkap untuk banyak aksi ini.

---

## INT-07 — Mobile ↔ Backend Hardening

Status: `[READY]`  
Priority: `P0`  
Dependency: BE-01, BE-03, MOB-01, MOB-02, MOB-03.

### Tujuan

Memastikan mobile kasir dan backend benar-benar aman untuk checkout, shift, terminal, payment, dan retry.

### Task dan Subtask

- `[TODO]` End-to-end login → outlet → shift open → checkout → payment → receipt.
- `[TODO]` End-to-end checkout retry/recovery.
- `[TODO]` End-to-end shift close.
- `[TODO]` Test 401/403/409/422 di mobile.
- `[TODO]` Test no client money calculation.
- `[TODO]` Test device/terminal lock/PIN switch.
- `[TODO]` Test print/share jika hardware tersedia.

### Acceptance Criteria

- Mobile happy path berjalan di emulator/device.
- Conflict/retry tidak membuat transaksi ganda.
- Role/outlet/device policy enforced server-side.
- No general offline queue/database ditambahkan.

### Bug/Kendala

- `[OPEN]` Butuh QA device/emulator nyata untuk alur terminal dan print.
