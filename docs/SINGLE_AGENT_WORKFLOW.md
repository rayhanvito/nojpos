# NojPOS Single-Agent Workflow

Dokumen ini adalah aturan kerja aktif NojPOS. Tujuannya menjaga implementasi tetap aman, berurutan, dan tidak tabrakan.

## Prinsip Utama

NojPOS saat ini memakai pola **single active coding agent**.

Artinya:

- Hanya satu task yang boleh berstatus `[IN PROGRESS]`.
- Hanya satu area utama yang boleh diubah dalam satu task.
- Backend, web, dan mobile tidak boleh diubah bersamaan kecuali task integrasi eksplisit mengizinkan.
- Task berikutnya baru dimulai setelah validasi task sekarang selesai.
- Semua perubahan harus mengikuti `PRDPOSJA.md` dan `docs/TASK_QUEUE.md`.

## Source Of Truth

Urutan baca wajib sebelum coding:

1. `AGENTS.md`
2. `PRDPOSJA.md`
3. `docs/README.md`
4. `docs/TASK_QUEUE.md`
5. `docs/STORY_PROGRESS.md`
6. Story area yang relevan:
   - `docs/STORY_BACKEND.md`
   - `docs/STORY_MOBILE_KASIR.md`
   - `docs/STORY_WEB_ADMIN.md`
   - `docs/STORY_INTEGRATION.md`
7. `docs/API_CONTRACTS/` jika menyentuh API/integrasi.
8. `docs/BUG_TRACKER.md`
9. `docs/VALIDATION_CHECKLIST.md`
10. `apps/<area>/AGENTS.md`

## Status Yang Dipakai

| Status | Arti |
| --- | --- |
| `[TODO]` | Belum siap atau belum diprioritaskan |
| `[READY]` | Boleh dikerjakan sekarang jika dependency selesai |
| `[IN PROGRESS]` | Sedang dikerjakan; hanya boleh ada satu |
| `[REVIEW]` | Implementasi selesai, menunggu cek validasi/manual review |
| `[DONE]` | Selesai dan validasi sesuai scope lolos |
| `[BLOCKED]` | Tidak boleh dikerjakan sebelum blocker selesai |
| `[DEFERRED]` | Sengaja ditunda dari fase sekarang |
| `[BUG]` | Ada bug aktif yang perlu diperbaiki |
| `[PREVIEW]` | UI/konsep ada, belum data/API real |

## Cara Memilih Task

1. Buka `docs/TASK_QUEUE.md`.
2. Ambil task paling atas yang statusnya `[READY]`.
3. Pastikan tidak ada task lain `[IN PROGRESS]`.
4. Update task menjadi `[IN PROGRESS]` sebelum coding.
5. Kerjakan hanya allowed scope task tersebut.
6. Jalankan validasi yang diminta task.
7. Update task menjadi `[DONE]`, `[REVIEW]`, atau `[BLOCKED]`.
8. Update story dan bug tracker jika ada perubahan.

## Aturan Scope

### Backend Task

Boleh edit:

- `apps/backend/**`
- `docs/STORY_BACKEND.md`
- `docs/STORY_INTEGRATION.md`
- `docs/API_CONTRACTS/**`
- `docs/BUG_TRACKER.md`
- `docs/TASK_QUEUE.md`

Tidak boleh edit tanpa izin eksplisit:

- `apps/web/**`
- `apps/cashier/**`

### Web Admin Task

Boleh edit:

- `apps/web/**`
- `docs/STORY_WEB_ADMIN.md`
- `docs/STORY_INTEGRATION.md`
- `docs/API_CONTRACTS/**`
- `docs/BUG_TRACKER.md`
- `docs/TASK_QUEUE.md`

Tidak boleh edit tanpa izin eksplisit:

- `apps/backend/**`
- `apps/cashier/**`

### Mobile Kasir Task

Boleh edit:

- `apps/cashier/**`
- `docs/STORY_MOBILE_KASIR.md`
- `docs/STORY_INTEGRATION.md`
- `docs/BUG_TRACKER.md`
- `docs/TASK_QUEUE.md`

Tidak boleh edit tanpa izin eksplisit:

- `apps/backend/**`
- `apps/web/**`

### Docs/Planning Task

Boleh edit:

- `PRDPOSJA.md`
- `AGENTS.md`
- `apps/*/AGENTS.md`
- `docs/**`
- `scripts/**` hanya jika task memang menyentuh tooling

Tidak boleh edit runtime app kecuali task menyebutkan.

## Urutan Integrasi Wajib

Jangan loncat urutan ini:

1. Rapikan docs dan task queue.
2. Finalisasi session/auth strategy untuk web admin.
3. Backend dashboard summary endpoint.
4. Test backend.
5. Integrasi `/dashboard` web read-only.
6. Test web.
7. Tenant admin read-only pages.
8. Tenant admin safe writes.
9. Super Admin read-only.
10. Sensitive actions paling akhir.

## Sensitive Actions Yang Dilarang Sampai Fase Akhir

Jangan aktifkan real action berikut sebelum audit, policy, idempotency, reason, dan approval siap:

- refund
- void
- payment retry
- manual mark paid
- akses bantuan/impersonation
- suspend/delete toko
- ban user
- reset password
- announcement broadcast
- shift close/open dari web
- stock adjustment dari web
- cash reconciliation dari web

Untuk UI preview, action tersebut harus disabled/gated dan copy harus jujur bahwa fitur belum aktif.

## Validasi Wajib

Gunakan `docs/VALIDATION_CHECKLIST.md`.

Ringkas:

Backend:

```bash
cd apps/backend
php artisan test
php artisan route:list --path=api/v1
```

Web Admin:

```bash
cd apps/web
npm run typecheck
npm run lint
npm test
npm run build
npx playwright test --list
```

Mobile Kasir:

```bash
cd apps/cashier
flutter analyze
flutter test
```

## Laporan Akhir Wajib

Setiap task harus melaporkan:

```text
Built:
- <apa yang dibuat/diubah>

Verified:
- <command yang dijalankan dan hasilnya>

Blocker:
- <none atau blocker spesifik>

Next:
- <task berikutnya yang disarankan dari TASK_QUEUE>
```

## Commit Rule

Satu commit sebaiknya mewakili satu task yang utuh.

Format commit:

```text
<area>: <ringkasan task>
```

Contoh:

```text
docs: switch to single-agent task queue
backend: add dashboard summary endpoint
web: integrate dashboard summary read-only
```

Jangan commit jika validasi wajib gagal, kecuali commit tersebut memang dokumentasi atau mencatat blocker dengan jelas.
