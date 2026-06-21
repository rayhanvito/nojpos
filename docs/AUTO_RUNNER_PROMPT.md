# NojPOS Auto Runner Prompt — Single-Agent Sequential Executor

Gunakan prompt ini untuk agent coding yang bisa membaca file `.md` otomatis.

Tujuan prompt ini adalah membuat agent:

1. Membaca PRD, workflow, task queue, dan prompt sequence otomatis.
2. Menentukan prompt aktif berikutnya secara aman.
3. Menjalankan **hanya satu prompt/task aktif** dalam satu sesi, kecuali user eksplisit meminta batch kecil.
4. Berhenti jika validasi gagal, scope tidak jelas, atau ada risiko fitur sensitif aktif terlalu cepat.
5. Menghasilkan laporan akhir yang rapi.

> Jangan pakai prompt ini untuk menjalankan semua task sampai selesai tanpa review manusia. NojPOS harus tetap berjalan bertahap agar tidak rusak dan tidak loncat-loncat.

---

## AUTO RUNNER PROMPT — Copy Paste Ke Coding Agent

```text
Kamu mengerjakan project NojPOS.

Mode utama:
- Single-agent sequential executor.
- Baca file markdown project secara otomatis.
- Tentukan task aktif berikutnya dari docs/TASK_QUEUE.md dan docs/PROMPT_SEQUENCE.md.
- Jalankan hanya 1 prompt/task aktif dalam sesi ini.
- Jangan menjalankan banyak prompt sekaligus kecuali user eksplisit menulis: RUN_BATCH_MAX_N.
- Jangan loncat prompt.
- Jangan mengerjakan task yang masih BLOCKED, DEFERRED, atau butuh approval.
- Jangan mengaktifkan fitur sensitif sebelum framework audit/idempotency/permission siap.

Lokasi repo:
C:\laragon\www\nojpos_final_fix

Langkah wajib sebelum mengubah apa pun:
1. Baca AGENTS.md.
2. Baca PRDPOSJA.md.
3. Baca docs/README.md.
4. Baca docs/SINGLE_AGENT_WORKFLOW.md.
5. Baca docs/TASK_QUEUE.md.
6. Baca docs/STORY_PROGRESS.md.
7. Baca docs/BUG_TRACKER.md.
8. Baca docs/PROMPT_SEQUENCE.md.
9. Baca docs/VALIDATION_CHECKLIST.md.
10. Baca story file sesuai area task:
    - Backend: docs/STORY_BACKEND.md dan apps/backend/AGENTS.md.
    - Web: docs/STORY_WEB_ADMIN.md dan apps/web/AGENTS.md.
    - Mobile: docs/STORY_MOBILE_KASIR.md dan apps/cashier/AGENTS.md.
    - Integration: docs/STORY_INTEGRATION.md.
11. Jika task menyentuh API, baca kontrak terkait di docs/API_CONTRACTS/.

Cara menentukan task aktif:
1. Cek docs/TASK_QUEUE.md.
2. Ambil task paling atas yang statusnya READY atau IN PROGRESS.
3. Jika ada task IN PROGRESS, lanjutkan task itu dulu.
4. Jika tidak ada IN PROGRESS, ambil READY paling atas.
5. Cocokkan Task ID itu dengan prompt di docs/PROMPT_SEQUENCE.md.
6. Jika prompt tidak ditemukan, STOP dan laporkan blocker.
7. Jika task status BLOCKED/DEFERRED/TODO tanpa dependency terpenuhi, STOP dan laporkan.
8. Jika working tree tidak bersih, baca status git dan pastikan perubahan existing relevan dengan task aktif. Jangan revert perubahan user/agent lain tanpa izin.

Mode eksekusi:
- Jalankan isi prompt aktif dari docs/PROMPT_SEQUENCE.md secara lengkap.
- Patuhi scope boleh diedit dan scope tidak boleh diedit.
- Jangan membuat keputusan di luar PRD/story/contract.
- Jika kontrak belum ada, jangan langsung implementasi runtime; buat/selesaikan kontrak dulu.
- Jika endpoint backend belum ada, jangan buat web UI menebak response.
- Jika backend/web/mobile mismatch, catat blocker di BUG_TRACKER dan STOP.

Larangan global:
- Jangan install package baru kecuali prompt aktif eksplisit mengizinkan.
- Jangan edit node_modules, vendor, .next, build cache, storage log, atau artifact lokal.
- Jangan simpan token di localStorage/sessionStorage.
- Jangan expose token ke Client Component, React props, fixture, console log, screenshot, atau error response.
- Jangan memakai NEXT_PUBLIC_API_BASE_URL untuk backend secret/server API.
- Jangan membuat UI fetch langsung ke Laravel.
- Jangan menghitung uang/stok final di client.
- Jangan membuat offline database umum/write queue umum di mobile.
- Jangan aktifkan refund, void, payment retry, manual mark paid, shift close/open, stock adjustment, cash reconciliation, akses bantuan/impersonate, suspend/delete toko, reset password, ban user, atau broadcast pengumuman sebelum prompt sensitive-action resmi.

Aturan sensitive action:
Jika task terlihat menyentuh salah satu aksi ini:
- refund
- void
- payment retry
- manual mark paid
- shift close/open
- stock adjustment
- cash reconciliation
- akses bantuan/impersonate
- suspend/delete/activate toko
- reset password
- ban user
- announcement broadcast

Maka cek dulu:
1. Apakah prompt aktif memang fase sensitive action?
2. Apakah docs/API_CONTRACTS/SENSITIVE_ACTIONS.md sudah ready?
3. Apakah audit log, reason, role permission, idempotency, dan tests sudah disyaratkan?
Jika salah satu belum siap, STOP dan catat blocker. Jangan implementasi action real.

Validasi wajib:
Ikuti validasi dari prompt aktif.
Jika prompt aktif tidak mencantumkan validasi, gunakan validasi standar:

Backend:
cd apps/backend
php artisan test
php artisan route:list --path=api/v1

Web:
cd apps/web
npm run typecheck
npm run lint
npm test
npm run build
npx playwright test --list
rg -n "fetch\(|XMLHttpRequest|sessionStorage|localStorage|NEXT_PUBLIC_API_BASE_URL" src/app src/components src/fixtures --glob '!*.test.ts' --glob '!*.test.tsx' || true
rg -n "Authorization|Bearer " src/app src/components src/fixtures --glob '!*.test.ts' --glob '!*.test.tsx' || true

Mobile:
cd apps/cashier
flutter analyze
flutter test
flutter build apk --debug

Docs-only:
Cek file yang diubah bisa dibaca.
Cek tidak ada referensi workflow lama:
rg -n "docs/AGENT_WORKFLOW.md|docs/AGENT_TASK_BOARD.md|docs/AGENT_PROMPT_TEMPLATE.md|docs/AGENT_HANDOFF_TEMPLATE.md|setup-agent-worktrees|worktree setup" docs --glob '!PROMPT_SEQUENCE.md' || true

Aturan jika validasi gagal:
- Jangan commit.
- Jangan lanjut prompt berikutnya.
- Jangan klaim sukses.
- Perbaiki jika masih dalam scope prompt aktif.
- Jika gagal karena blocker di luar scope, update docs/BUG_TRACKER.md dan STOP.
- Laporan akhir harus menyebut command yang gagal dan error ringkas.

Aturan update docs:
Setelah task selesai:
1. Update docs/TASK_QUEUE.md.
2. Update story area terkait.
3. Update docs/STORY_PROGRESS.md jika milestone berubah.
4. Update docs/BUG_TRACKER.md jika ada bug/blocker baru.
5. Update docs/API_CONTRACTS jika kontrak berubah.

Aturan commit:
- Commit hanya jika prompt aktif adalah checkpoint/commit prompt, atau user eksplisit meminta commit.
- Jangan commit otomatis setelah setiap task non-checkpoint kecuali diminta.
- Sebelum commit, cek git status dan pastikan artifact/noise tidak ikut.
- Jangan commit node_modules, .next, vendor, cache, log, secret, .env production.

Mode batch opsional:
Default: jalankan 1 prompt lalu berhenti.
Jika user menulis RUN_BATCH_MAX_N=2 atau RUN_BATCH_MAX_N=3:
- Boleh jalankan maksimal N prompt berurutan.
- Tetap berhenti di setiap checkpoint commit, validasi gagal, blocker, perubahan scope besar, sensitive action, atau prompt docs-only yang membutuhkan keputusan manusia.
- Jangan batch lebih dari 3 prompt.
- Jangan melewati checkpoint commit.

Laporan awal wajib sebelum coding:
Tulis singkat:
1. Prompt aktif yang dipilih.
2. Alasan prompt itu dipilih dari TASK_QUEUE.
3. Scope file yang akan diedit.
4. Validasi yang akan dijalankan.
5. Konfirmasi tidak ada task yang dilompati.

Laporan akhir wajib:
Gunakan format ini:

Built:
- Task/prompt yang dikerjakan.
- File yang diubah.
- Ringkasan implementasi/dokumen.
- Route/endpoint/komponen yang dibuat jika ada.

Verified:
- Command validasi yang dijalankan.
- Hasil setiap command: PASS/FAIL.
- Boundary scan jika web.
- Test count jika tersedia.

Blocker:
- None, atau jelaskan blocker spesifik.
- Jika ada blocker, sebutkan file docs/BUG_TRACKER.md sudah diupdate atau belum.

Next:
- Prompt berikutnya dari docs/PROMPT_SEQUENCE.md.
- Apakah aman lanjut atau harus review manusia dulu.
```

---

## Mode Yang Disarankan Untuk Kamu Pakai

Untuk NojPOS sekarang, pakai default:

```text
Jalankan AUTO RUNNER PROMPT.
Mode: 1 prompt aktif saja.
```

Jangan langsung batch panjang.

Setelah agent selesai, cek laporan akhirnya. Kalau aman, baru jalankan lagi prompt yang sama. Agent akan membaca `TASK_QUEUE.md` dan `PROMPT_SEQUENCE.md`, lalu mengambil prompt berikutnya.

---

## Contoh Instruksi Ke Agent Setelah Paste Auto Runner

Tambahkan satu kalimat ini di bawah prompt kalau ingin benar-benar hanya 1 step:

```text
Jalankan hanya prompt aktif berikutnya, lalu berhenti setelah laporan akhir. Jangan lanjut prompt setelahnya.
```

Kalau ingin batch kecil, gunakan:

```text
RUN_BATCH_MAX_N=2. Jalankan maksimal 2 prompt berurutan, tetapi berhenti jika validasi gagal, masuk checkpoint commit, menemukan blocker, atau scope berubah besar.
```

Rekomendasi saat ini tetap 1 prompt per sesi.
