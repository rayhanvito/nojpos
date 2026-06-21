# NojPOS Bug Tracker

PRD: `PRDPOSJA.md`  
Gunakan file ini untuk bug, blocker, dan gap yang menghambat implementasi story. Jangan biarkan bug penting hanya tersimpan di chat.

## Bug/Kendala Aktif

| ID | Area | Severity | Status | Ringkasan | Story Terkait | Next Action |
| --- | --- | --- | --- | --- | --- | --- |
| BUG-INT-01 | Integration/Web | P0 | `[OPEN]` | Session strategy web admin belum final. | INT-01, BE-01, WEB-07 | Putuskan cookie HttpOnly/Sanctum SPA/BFF dan dokumentasikan. |
| BUG-BE-01 | Backend | P0 | `[OPEN]` | Perlu audit tenant scope untuk semua query builder/read-write tenant data. | BE-01 | Audit endpoint yang akan dibuka ke web admin. |
| BUG-BE-02 | Backend | P1 | `[OPEN]` | Endpoint `GET /api/v1/dashboard/summary` belum ada. | BE-04, INT-02, WEB-01 | Buat kontrak DTO dan endpoint agregat. |
| BUG-BE-03 | Backend | P2 | `[OPEN]` | Platform billing/support/activity/system/announcement/operator backend belum lengkap. | BE-06, WEB-05, WEB-06, INT-05 | Pecah domain read-only satu per satu. |
| BUG-WEB-01 | Web Admin | P1 | `[OPEN]` | Web admin masih fixture lokal dan belum punya API adapter aktif. | WEB-01..WEB-07, INT-01 | Tunggu session strategy, lalu integrasi read-only. |
| BUG-WEB-02 | Web Admin | P3 | `[OPEN]` | Folder lama `nojpos_admin_web/apps/web` masih menyisakan cache/dependency terkunci proses. | WEB-07 | Hapus manual setelah dev server/editor lama ditutup. |
| BUG-MOB-01 | Mobile Kasir | P1 | `[OPEN]` | Perlu QA device/emulator untuk lock, PIN switch, checkout retry, dan print. | MOB-01, MOB-02, MOB-05, INT-07 | Jalankan QA di emulator/device target. |
| BUG-MOB-02 | Mobile Kasir | P0 | `[OPEN]` | Perlu verifikasi final bahwa semua total transaksi final berasal dari server. | MOB-02, BE-03 | Audit POS/payment UI dan repository. |
| BUG-BE-04 | Backend | P1 | `[OPEN]` | Payment read/list endpoint untuk web admin belum jelas. | BE-03, WEB-02 | Tentukan apakah pakai transactions/payment report atau endpoint baru. |

## Bug Resolved

| ID | Area | Status | Ringkasan | Bukti/Notes |
| --- | --- | --- | --- | --- |
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
