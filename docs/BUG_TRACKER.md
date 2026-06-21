# NojPOS Bug Tracker

PRD: `PRDPOSJA.md`  
Gunakan file ini untuk bug, blocker, dan gap yang menghambat implementasi story. Jangan biarkan bug penting hanya tersimpan di chat.

## Bug/Kendala Aktif

| ID | Area | Severity | Status | Ringkasan | Story Terkait | Next Action |
| --- | --- | --- | --- | --- | --- | --- |
| BUG-BE-01 | Backend | P0 | `[OPEN]` | Perlu audit tenant scope untuk semua query builder/read-write tenant data, termasuk `/me` dan `/outlets` yang dibaca Web Admin session bootstrap. | BE-01 | Audit endpoint yang akan dibuka ke web admin; BFF WEB-01B sudah memfilter outlets by `business_id` sebagai guard tambahan. |
| BUG-BE-03 | Backend | P2 | `[OPEN]` | Platform billing/support/activity/system/announcement/operator backend belum lengkap. | BE-06, WEB-05, WEB-06, INT-05 | Pecah domain read-only satu per satu. |
| BUG-WEB-02 | Web Admin | P3 | `[OPEN]` | Folder lama `nojpos_admin_web/apps/web` masih menyisakan cache/dependency terkunci proses. | WEB-07 | Hapus manual setelah dev server/editor lama ditutup. |
| BUG-MOB-01 | Mobile Kasir | P1 | `[OPEN]` | Perlu QA device/emulator untuk lock, PIN switch, checkout retry, dan print. | MOB-01, MOB-02, MOB-05, INT-07 | Jalankan QA di emulator/device target. |
| BUG-MOB-02 | Mobile Kasir | P0 | `[OPEN]` | Perlu verifikasi final bahwa semua total transaksi final berasal dari server. | MOB-02, BE-03 | Audit POS/payment UI dan repository. |
| BUG-BE-04 | Backend | P1 | `[OPEN]` | Payment read/list endpoint untuk web admin belum jelas. | BE-03, WEB-02 | Tentukan apakah pakai transactions/payment report atau endpoint baru. |
| BUG-BE-05 | Backend/Web | P2 | `[OPEN]` | `GET /api/v1/transactions` sudah ada dan tenant-scoped, tapi filter/pagination backend masih terbatas untuk kebutuhan Web Admin. | BE-03, WEB-02, INT-03 | WEB-02B sudah sanitize/map di BFF; backend-side filter/pagination hardening direkomendasikan sebelum volume produksi besar. |
| BUG-BE-06 | Backend/Web | P2 | `[OPEN]` | Inventory read endpoints sudah ada dan tenant-scoped, tapi pagination total, filter `category_id`/`stock_status`, dan low-stock threshold belum stabil untuk Web Admin produksi. | BE-03, WEB-03, INT-03 | WEB-03B wajib sanitize/map/paginate di BFF; backend-side inventory filter/pagination/threshold hardening direkomendasikan sebelum volume produksi besar. |

## Bug Resolved

| ID | Area | Status | Ringkasan | Bukti/Notes |
| --- | --- | --- | --- | --- |
| BUG-WEB-RES-04 | Integration/Web | `[RESOLVED]` | Dashboard Owner/Admin sudah terintegrasi read-only melalui BFF. | WEB-01C menambahkan route `GET /api/admin/dashboard/summary`, server-side dashboard adapter, mapping response backend ke `/dashboard`, fallback berlabel jelas, dan test/boundary scan. |
| BUG-BE-RES-02 | Backend | `[RESOLVED]` | Endpoint `GET /api/v1/dashboard/summary` sudah tersedia. | BE-04B menambahkan endpoint tenant-scoped owner/admin dan test backend. |
| BUG-WEB-RES-03 | Integration/Web | `[RESOLVED]` | Session/BFF adapter scaffold Web Admin sudah selesai. | WEB-01B menambahkan sealed HttpOnly cookie helper, server-only backend client, route `/api/admin/auth/login`, `/api/admin/auth/logout`, dan `/api/admin/session`; dashboard integration selesai di WEB-01C. |
| BUG-WEB-RES-02 | Integration/Web | `[RESOLVED]` | Session strategy web admin sudah final. | WEB-01A memilih Next.js BFF/server-side route handler dengan HttpOnly sealed session cookie; implementasi adapter selesai di WEB-01B. |
| BUG-WEB-RES-01 | Web Admin | `[RESOLVED]` | Platform shell sempat overflow/blank kanan. | CSS container platform sudah dipatch; build/test/list lolos saat itu. |

## Watchlist

### Backend

- Idempotency conflict harus konsisten untuk body berbeda.
- Audit payload harus redact PII/payment reference.
- Timezone outlet untuk laporan harian belum boleh dilupakan.
- Staff delete punya dua route; kontrak web admin harus pilih satu.

### Mobile Kasir

- Jangan menambah general offline DB/write queue.
- Checkout retry hanya untuk satu transaksi in-flight.
- Print/reprint harus policy-aware.
- Error 401/403/409/422 harus punya copy jelas.

### Web Admin

- Jangan menambah fetch/API langsung di component sembarangan.
- Jangan menyimpan token di localStorage/sessionStorage.
- Jangan mengaktifkan action sensitif dari preview UI.
- Tetap jaga Tenant Admin dan Platform Super Admin terpisah.

### Platform Super Admin

- Jangan membuat UI terlihat bebas mengintip data detail toko.
- Akses bantuan harus butuh alasan, durasi, audit.
- Billing action manual harus idempotent dan audited.
- Announcement broadcast harus melalui approval/audit.

## Template Bug Baru

```md
## BUG-AREA-XX — Judul singkat

Severity: P0/P1/P2/P3  
Status: [OPEN]/[IN PROGRESS]/[RESOLVED]/[WONTFIX]  
Story: BE-xx/MOB-xx/WEB-xx/INT-xx

### Gejala

Apa yang terlihat oleh user/developer.

### Dampak

Risiko bisnis/teknis/security.

### Dugaan Penyebab

Hipotesis awal, jangan klaim jika belum pasti.

### Repro Step

1. ...
2. ...

### Fix Plan

- [ ] Task 1
- [ ] Task 2

### Verifikasi

- Command/test/manual QA yang harus lolos.
```
