# NojPOS Docs Aktif

Dokumen ini adalah index aktif untuk memantau progress NojPOS. Semua kerja baru harus mengikuti PRD, single-agent workflow, task queue, dan story docs.

## Source of Truth

1. `PRDPOSJA.md` — PRD utama NojPOS.
2. `docs/SINGLE_AGENT_WORKFLOW.md` — aturan kerja single active coding agent.
3. `docs/TASK_QUEUE.md` — queue task final yang dikerjakan berurutan.
4. `docs/STORY_PROGRESS.md` — progress global dan urutan fase.
5. Story area:
   - `docs/STORY_BACKEND.md`
   - `docs/STORY_MOBILE_KASIR.md`
   - `docs/STORY_WEB_ADMIN.md`
   - `docs/STORY_INTEGRATION.md`
6. `docs/API_CONTRACTS/` — kontrak API sebelum backend/client integrasi.
7. `docs/BUG_TRACKER.md` — bug dan blocker aktif.
8. `docs/VALIDATION_CHECKLIST.md` — validasi wajib sebelum commit/merge.
9. Kode dan test di `apps/backend`, `apps/cashier`, `apps/web`.

## Dokumen Aktif

| Dokumen | Fungsi |
| --- | --- |
| `docs/SINGLE_AGENT_WORKFLOW.md` | Aturan single-agent, scope, validasi, dan laporan akhir |
| `docs/TASK_QUEUE.md` | Task aktif yang dikerjakan dari atas ke bawah |
| `docs/STORY_PROGRESS.md` | Status fase dan progres global |
| `docs/STORY_BACKEND.md` | Story, task, subtask, dan bug backend |
| `docs/STORY_MOBILE_KASIR.md` | Story, task, subtask, dan bug mobile kasir |
| `docs/STORY_WEB_ADMIN.md` | Story, task, subtask, dan bug web admin |
| `docs/STORY_INTEGRATION.md` | Urutan integrasi backend/web/mobile |
| `docs/API_CONTRACTS/README.md` | Aturan membuat kontrak API |
| `docs/BUG_TRACKER.md` | Bug dan blocker aktif |
| `docs/VALIDATION_CHECKLIST.md` | Checklist validasi per area |

## Format Status

| Status | Arti |
| --- | --- |
| `[DONE]` | Sudah diimplementasikan dan tervalidasi sesuai scope saat itu |
| `[REVIEW]` | Selesai dikerjakan, menunggu review/cek manual |
| `[IN PROGRESS]` | Sedang ada implementasi; hanya boleh ada satu task aktif |
| `[READY]` | Siap dikerjakan setelah dependency terpenuhi |
| `[TODO]` | Belum mulai |
| `[BLOCKED]` | Tidak boleh lanjut sebelum blocker selesai |
| `[BUG]` | Ada bug aktif yang harus dibereskan |
| `[PREVIEW]` | UI/dokumen/konsep sudah ada, belum backend/API real |
| `[DEFERRED]` | Ditunda dari release dekat |

## Cara AI/Engineer Memakai Docs

1. Baca `AGENTS.md`.
2. Baca `PRDPOSJA.md`.
3. Baca `docs/SINGLE_AGENT_WORKFLOW.md`.
4. Baca `docs/TASK_QUEUE.md`.
5. Pilih task paling atas yang statusnya `[READY]` atau lanjutkan task `[IN PROGRESS]` yang sedang berjalan.
6. Baca story area yang relevan.
7. Baca kontrak API jika task menyentuh integrasi.
8. Kerjakan hanya file yang masuk allowed scope task.
9. Jalankan validasi sesuai `docs/VALIDATION_CHECKLIST.md`.
10. Update task/story/bug tracker sebelum laporan akhir.

## Prinsip Integrasi Berikutnya

- Web Admin masih UI preview sampai session strategy web disetujui.
- Tenant Admin dan Platform Super Admin tetap terpisah.
- Mobile Kasir tetap online-first, tanpa general offline database/write queue.
- Backend tetap sumber kebenaran untuk tenant, outlet, device, role, uang, stok, dan laporan.
- Sensitive writes harus paling akhir setelah policy, audit, idempotency, reason, approval, dan tests siap.

## Urutan Aman Saat Ini

1. Selesaikan clean docs baseline dan commit bersih.
2. Finalisasi kontrak `GET /api/v1/dashboard/summary`.
3. Implement backend dashboard summary.
4. Test backend.
5. Putuskan web session strategy.
6. Integrasi `/dashboard` web read-only.
7. Test web.
8. Lanjut Tenant Admin read-only pages.

Jangan menjalankan banyak coding agent paralel dulu. Gunakan satu active coding agent agar perubahan tidak tabrakan.
