# PRDPOSJA — Product Requirements Document NojPOS

**Status:** Active source of truth  
**Version:** 3.0 — implementation roadmap edition  
**Updated:** 2026-06-21  
**Scope:** Laravel Backend API, Flutter Mobile Kasir, Next.js Web Admin, Platform Super Admin, dan integrasi antar-aplikasi  
**Audience:** Product owner, AI coding agent, backend engineer, Flutter engineer, web admin engineer, QA

---

## 1. Ringkasan Produk

NojPOS adalah POS SaaS online-first untuk UMKM Indonesia. Produk ini membantu pemilik toko, kasir, admin, dan tim internal NojPOS menjalankan penjualan, stok, shift, laporan, paket langganan, bantuan pelanggan, dan operasional platform secara aman.

NojPOS terdiri dari tiga aplikasi utama:

| Aplikasi | Lokasi | Peran |
| --- | --- | --- |
| Backend API | `apps/backend` | Laravel API, autentikasi, tenant scope, transaksi, stok, laporan, langganan, platform admin |
| Mobile Kasir | `apps/cashier` | Flutter app untuk kasir shared-terminal, checkout, shift, pembayaran, struk |
| Web Admin | `apps/web` | Next.js admin untuk owner/admin toko dan Platform Super Admin NojPOS |

NojPOS harus terasa sederhana untuk market Indonesia: bahasa jelas, alur pendek, rupiah, cocok untuk toko retail, kafe, laundry, barbershop, booth minuman, dan usaha lokal.

---

## 2. Prinsip Produk Yang Tidak Boleh Dilanggar

1. **Online-first.** Semua write operasional harus lewat backend. Tidak ada general offline database, `/sync/pull`, `/sync/push`, atau write queue umum.
2. **Server adalah kalkulator.** Client tidak menghitung total final, diskon, pajak, service, rounding, stok akhir, refund final, atau laporan final.
3. **Tenant isolation.** `business_id` selalu berasal dari user/session server-side. Client tidak boleh memilih tenant sendiri.
4. **Outlet dan device diverifikasi.** Checkout, shift, cash, attendance, terminal lock, dan payment harus memverifikasi outlet, device, user, dan role.
5. **Uang integer rupiah.** Semua money disimpan dan dikirim sebagai integer rupiah.
6. **Aksi sensitif harus aman.** Payment, refund, void, stock movement, shift close, cash reconciliation, access bantuan, suspend toko, reset password, dan broadcast announcement harus authorized, idempotent, audited, dan transactional.
7. **Traceability.** Perubahan uang, stok, user, toko, paket, dan akses bantuan harus masuk audit/activity log.
8. **UI boleh preview, tapi jujur.** Web admin yang belum integrasi backend harus menyatakan data contoh/preview dan tidak membuat action real.
9. **Bahasa Indonesia sederhana.** Gunakan istilah Toko, Paket, Langganan & Tagihan, Bantuan, Aktivitas, Sistem, Pengumuman, Akses bantuan.
10. **Jangan loncat fase.** Integrasi harus mengikuti story order di `docs/STORY_PROGRESS.md` dan `docs/STORY_INTEGRATION.md`.

---

## 3. Persona dan Kebutuhan Utama

### 3.1 Kasir

Kasir memakai mobile/tablet shared terminal untuk jualan cepat.

Kebutuhan:

- Login atau PIN switch cepat.
- Buka shift sebelum jualan.
- Cari produk dan customer dengan cepat.
- Quote transaksi dari server.
- Terima pembayaran.
- Cetak/bagikan struk.
- Tutup shift dan laporkan kas aktual.

### 3.2 Owner/Admin Toko

Owner/admin memantau bisnis dan mengelola data toko.

Kebutuhan:

- Lihat ringkasan hari ini: penjualan, transaksi, rata-rata, gross profit estimasi, stok kritis, selisih kas.
- Kelola produk, kategori, customer, staff, cabang, pengaturan.
- Lihat transaksi, inventory, laporan, pembayaran.
- Tidak menjalankan aksi sensitif dari web sebelum backend aman.

### 3.3 Tim Internal NojPOS / Platform Super Admin

Tim internal memantau platform SaaS.

Kebutuhan:

- Lihat toko, paket, langganan/tagihan, bantuan, aktivitas, sistem, pengumuman.
- Melihat data agregat platform, bukan bebas mengintip detail toko.
- Akses bantuan hanya dengan alasan, durasi, dan audit log.
- Action sensitif gated sampai backend siap.

---

## 4. Ruang Lingkup Release

### 4.1 Sudah Diimplementasikan / Sedang Ada

Backend:

- Auth dasar, Sanctum protected routes, role middleware.
- Products, categories, customers, staff CRUD dasar.
- Inventory read/write dasar.
- Shift open/current/cash movement/close.
- Transactions quote/create/list/detail/recovery.
- Payments create, void, refund.
- Reports dasar.
- Superadmin summary, plans, businesses, subscription per business.
- Payment webhook route.

Mobile Kasir:

- Struktur feature Flutter sudah ada untuk auth, POS, payment, shift, inventory, reports, staff, attendance, store, lock, notifications.
- Flow inti checkout/shift sudah tersambung sebagian ke backend.

Web Admin:

- Tenant Admin UI preview lengkap di `apps/web`.
- Dashboard owner/admin sudah dipolish.
- Platform Super Admin UI preview sudah dipolish dan Indonesia-friendly.
- Belum ada fetch/API/storage/auth integration di page/component.

### 4.2 Belum Boleh Diaktifkan

- Web admin action real untuk refund, void, reprint, export, shift open/close, stock adjustment, payment retry, cash reconciliation.
- Platform access bantuan/impersonation real.
- Platform suspend/delete toko real.
- Platform reset password/ban user real.
- Platform announcement broadcast real.
- Manual mark paid / retry payment real.

### 4.3 Belum Ada Domain Backend Penuh

- Dashboard summary agregat satu endpoint.
- SaaS invoice/billing platform list.
- Support tickets/bantuan.
- Platform activity logs endpoint lengkap.
- System health endpoint untuk UI.
- Announcement draft/target/delivery.
- Operator internal/user access management.

---

## 5. Arsitektur Target

### 5.1 Backend Laravel

Pola wajib:

```text
controller -> request validation -> authorization/policy -> service transaction -> model/query -> resource/envelope
```

Syarat:

- Semua response success: `{ "data": ..., "meta": ... }`.
- Semua error: `{ "error": { "code": "...", "message": "...", "details": ... } }`.
- Semua query tenant harus scoped by authenticated context.
- Sensitive writes memakai idempotency, DB transaction, row lock/state check jika perlu, dan audit event.
- Uang integer rupiah.
- Timezone outlet dipakai untuk business-day boundary.

### 5.2 Flutter Mobile Kasir

Pola wajib:

```text
ui -> provider/notifier -> repository interface -> ApiClient
```

Syarat:

- Widget tidak membuat HTTP langsung.
- Widget tidak menghitung total uang final.
- Bounded checkout retry hanya untuk final checkout request.
- Tidak ada general offline database/write queue.
- Error state harus jelas: loading, empty, 401, 403, 409 conflict, 422 validation, offline/server unavailable.

### 5.3 Next.js Web Admin

Pola target setelah integrasi disetujui:

```text
page/server boundary or client view -> feature adapter -> typed API contract -> backend API
```

Syarat:

- Tidak langsung fetch dari random component.
- Auth/session harus diputuskan dulu: cookie HttpOnly/Sanctum SPA/BFF.
- Token tidak boleh disimpan di `localStorage` atau `sessionStorage`.
- Tenant Admin dan Platform Super Admin tetap dipisah.
- Route wajib tetap:
  - Landing: `/`
  - Tenant Admin: `/dashboard`
  - Platform Super Admin: `/platform`

---

## 6. Modul Produk dan Story Utama

Story detail ada di folder `docs/`.

| Area | Story | Tujuan |
| --- | --- | --- |
| Backend | BE-01 | Auth, tenant scope, role, outlet, device context |
| Backend | BE-02 | Master data toko: produk, kategori, customer, staff |
| Backend | BE-03 | Operasional POS: inventory, shift, transaction, payment |
| Backend | BE-04 | Reports dan dashboard summary |
| Backend | BE-05 | Platform Super Admin API |
| Backend | BE-06 | Billing SaaS, support, announcement, activity domain |
| Mobile | MOB-01 | Auth, outlet, device, PIN, terminal session |
| Mobile | MOB-02 | POS checkout |
| Mobile | MOB-03 | Shift dan kas |
| Mobile | MOB-04 | Inventory, attendance, staff, reports |
| Mobile | MOB-05 | Receipt, print/share, notification/retry center |
| Web | WEB-01 | Tenant dashboard owner/admin |
| Web | WEB-02 | Tenant admin management pages |
| Web | WEB-03 | Platform Super Admin ringkasan |
| Web | WEB-04 | Platform toko dan paket |
| Web | WEB-05 | Platform langganan & tagihan |
| Web | WEB-06 | Platform bantuan, aktivitas, sistem, pengumuman, user |
| Web | WEB-07 | Web admin quality gate |
| Integration | INT-01..INT-07 | Urutan integrasi agar tidak loncat-loncat |

---

## 7. Urutan Implementasi Wajib

AI/engineer harus mengikuti urutan ini kecuali product owner memberi instruksi eksplisit.

### Phase 0 — Stabilkan dokumentasi dan kontrak

1. Pakai `PRDPOSJA.md` sebagai PRD utama.
2. Pakai `docs/STORY_PROGRESS.md` untuk urutan kerja.
3. Jangan implementasi API web admin sebelum session strategy jelas.
4. Update bug tracker setiap menemukan blocker.

### Phase 1 — Backend foundation hardening

1. Audit tenant scope dan query builder.
2. Finalisasi auth/session web admin.
3. Pastikan standard envelope dan error code konsisten.
4. Tambahkan policy/test untuk endpoint yang akan dibuka ke web admin.

### Phase 2 — Mobile kasir hardening

1. QA login/outlet/device/PIN switch.
2. QA checkout quote/store/payment dengan idempotency.
3. QA shift gate dan terminal lock.
4. Pastikan no client-side final calculation.

### Phase 3 — Web Admin integration foundation

1. Implement session bootstrap read-only.
2. Integrasikan `/me`, `/outlets`, `/subscription`, `/settings` read-only.
3. Tambahkan loading/empty/error/forbidden state.

### Phase 4 — Tenant Dashboard read-only

1. Buat backend `GET /api/v1/dashboard/summary` atau mapping reports final.
2. Hubungkan `/dashboard` ke data real.
3. Tetap label estimasi untuk gross profit dan kas sampai perhitungan final.

### Phase 5 — Tenant Admin read-only pages

Urutan:

1. Transactions.
2. Inventory.
3. Catalog/products/categories.
4. Customers.
5. Staff/attendance.
6. Reports.
7. Settings/outlets/subscription.

### Phase 6 — Tenant Admin safe writes

Mulai dari low-risk:

1. Product/category/customer create/update.
2. Settings update.
3. Staff update.

Tunda:

- Refund.
- Void.
- Shift open/close from web.
- Stock adjustment.
- Payment retry.
- Export/report generation heavy.

### Phase 7 — Platform Super Admin read-only

1. Summary.
2. Businesses/toko.
3. Plans/paket.
4. Subscription per business.
5. Billing/support/activity/system/announcement setelah endpoint domain ada.

### Phase 8 — Platform Super Admin sensitive actions

Hanya setelah audit, approval, reason, idempotency, dan policy siap:

- Akses bantuan.
- Suspend/activate/archive toko.
- Change plan/manual extend.
- Manual mark paid/retry payment.
- Assign/tandai selesai support.
- Reset password/ban user.
- Announcement broadcast.

---

## 8. Acceptance Criteria Global

Sebuah story dinyatakan selesai jika:

1. Scope sesuai PRD dan story file.
2. Tidak melanggar tenant isolation.
3. Sensitive write memakai policy, idempotency, transaction, audit.
4. Response/error sesuai envelope.
5. UI punya loading, empty, error, forbidden, conflict state jika sudah integrasi.
6. Test relevan berjalan.
7. Docs/story/bug tracker diperbarui.
8. Tidak ada package baru tanpa alasan dan persetujuan.
9. Tidak menyentuh area lain tanpa kebutuhan.

---

## 9. Definition of Done per Aplikasi

### Backend

- `php artisan test` lolos.
- Route list sesuai kontrak.
- Migration reversible dan tidak merusak data historis.
- Policy/test tenant isolation ada.
- Audit/idempotency ada untuk write sensitif.

### Mobile Kasir

- `flutter analyze` lolos.
- `flutter test` lolos.
- Build debug bisa dibuat.
- Flow utama diuji di emulator/device jika tersedia.
- Tidak ada money final calculation di client.

### Web Admin

- `npm run typecheck` lolos.
- `npm run lint` lolos.
- `npm test` lolos.
- `npm run build` lolos.
- `npx playwright test --list` lolos.
- Responsive no horizontal overflow untuk route utama.
- API integration hanya di layer yang disetujui.

---

## 10. Bug dan Risk Register Utama

Risk terbesar saat ini:

1. Web admin session strategy belum final.
2. Dashboard summary endpoint belum ada.
3. Super Admin UI lebih lengkap daripada backend domain.
4. Billing SaaS belum punya invoice/payment monitoring domain penuh.
5. Support/announcement/platform activity belum punya backend domain.
6. Sensitive action belum boleh real sampai audit/approval siap.
7. Mobile perlu QA device untuk terminal lock, PIN switch, print, checkout retry.
8. Tenant scope harus diuji ulang sebelum membuka web admin real data.

Semua risk aktif harus dicatat di `docs/BUG_TRACKER.md` atau story terkait.

---

## 11. AI Implementation Rules

Saat AI mengerjakan task NojPOS:

1. Baca `AGENTS.md`, `PRDPOSJA.md`, `docs/STORY_PROGRESS.md`, lalu story area yang relevan.
2. Pilih satu story dan satu task saja.
3. Jangan mengerjakan task downstream jika dependency belum selesai.
4. Jangan mengaktifkan action sensitif hanya karena UI sudah ada.
5. Jangan membuat endpoint baru tanpa menaruhnya di story/backend contract.
6. Jangan mengubah mobile, backend, dan web sekaligus kecuali story integrasi memang memerlukan itu.
7. Setelah selesai, update story status dan bug tracker.
8. Laporan akhir wajib menyebut Built, Verified, Blocker.

---

## 12. Current Next Best Work

Urutan kerja paling aman setelah dokumen ini:

1. Finalisasi web admin session strategy di `INT-01`.
2. Hardening backend tenant scope untuk endpoint read-only yang akan dibuka.
3. Buat endpoint `GET /api/v1/dashboard/summary`.
4. Integrasikan `/dashboard` read-only.
5. Integrasikan Tenant Admin read-only pages.
6. Baru lanjut Platform Super Admin read-only.

Jangan mulai sensitive writes sebelum phase read-only stabil.
