# AGENTS.md — OpenPOS (pos-pakesigap)

Panduan wajib untuk setiap agen (Codex) dan manusia yang mengubah project ini.
Project: OpenPOS di https://pos.pakesigap.com — backend Go (repo `backend/`,
github.com/adrr-dev/openPOS) + frontend React/Vite (repo `frontend/`,
github.com/0xMinomus/OpenPOS), di-deploy dari /opt/pos-pakesigap di VPS 2.
Pemilik project: Rayhan. Orkestrator & QA: Muse (PM). Agen pengkode: Codex.

## Aturan emas
1. **UI/UX TIDAK BERUBAH.** Kembangkan dari tampilan yang ada: shell, layout,
   komponen, spacing, warna, dan pola interaksi saat ini adalah acuan final.
   Fitur baru menempel pada pola halaman yang sudah ada (daftar, form, modal,
   tombol) — dilarang mendesain ulang, mengganti design system, atau membawa
   gaya visual baru.
2. **Business flow harus jelas dan tidak ambigu.** Setiap fitur punya satu
   alur yang tegas: siapa boleh, urutan statusnya apa, angka dihitung dari
   mana. Kalau sebuah aturan bisnis belum diputuskan di PLAN.md, JANGAN
   mengarang — berhenti dan laporkan sebagai pertanyaan ke PM.
3. **Anti-slop.** Tidak ada kode placeholder, stub "TODO nanti", fungsi mati,
   komentar generik hasil template, dependensi baru tanpa alasan tertulis,
   atau penamaan ulang massal. Setiap baris yang ditambah harus punya alasan
   yang bisa dijelaskan dalam satu kalimat.
4. **Clean code & struktur tetap seperti sekarang.** Ikuti layering dan pola
   yang sudah ada di repo (lihat bagian Arsitektur). Perubahan sekecil
   mungkin untuk mencapai tujuan; jangan refactor di luar scope unit kerja.
5. **Bahasa UI: Bahasa Indonesia**, sama seperti aplikasi sekarang. Pesan
   error ke user juga Bahasa Indonesia, jelas dan bisa ditindaklanjuti.

## Kontrak yang tidak boleh dilanggar
- API di bawah awalan `/api/v1`. Respons error selalu `{"error": "..."}` dan,
  bila sudah ada polanya, sertakan `code` mesin (mis. `PASSCODE_REQUIRED`).
- List selalu dibungkus `{"items": [...]}` (paginasi `page`/`limit` bila ada).
- Body request JSON memakai camelCase; field respons mengikuti tag JSON pada
  model (snake_case dari gorm.Model). Jangan ubah nama field yang sudah ada —
  frontend bergantung padanya.
- Semua uang adalah integer Rupiah (int64), tidak pernah float/desimal.
- Pajak bersifat ADITIF: `tax = round_half_up((subtotal - diskon) * taxPct / 100)`
  dan `total = subtotal - diskon + tax`. Rumus ini sudah terverifikasi live;
  jangan diubah dan jangan diganti rumus ekstraksi tax-inclusive.
- Metode pembayaran adalah enum persis: `Cash`, `Bank Transfer`, `QRIS`,
  `E-Wallet`, `Card`. Penambahan metode harus lewat keputusan PM di PLAN.md.
- RBAC: role `admin` (pemilik) dan `cashier`. Endpoint admin-only harus
  menolak kasir dengan 403. Data selalu ter-scope per toko (storeID dari
  token), tidak pernah dari input user.
- Auth: JWT access 15 menit + refresh 7 hari (rotasi refresh token), passcode
  PIN 5 digit untuk switch akun. Jangan melemahkan mekanisme ini.
- File deployment di root project (`docker-compose.yml`, `Dockerfile.backend`,
  `.env.deploy`, `bootstrap-account.txt`) adalah milik PM — JANGAN diubah,
  dipindah, atau di-commit isinya. `.env.deploy` berisi rahasia.

## Cara kerja per unit (wajib)
- **Superpowers selalu dipakai.** Mulai setiap sesi dengan skill
  `using-superpowers`, lalu ikuti disiplinnya untuk unit ini: susun rencana
  kerja unit dari spec PLAN.md (pola `writing-plans`/`executing-plans`),
  kerjakan dengan TDD bila perilaku baru ditambahkan (pola
  `test-driven-development`), dan sebelum menyatakan unit selesai jalankan
  verifikasi menyeluruh (pola `verification-before-completion`) plus
  self-review ala `requesting-code-review`. Spec di PLAN.md adalah sumber
  kebenaran — skill mengatur CARA bekerja, bukan mengubah spec.
- **Ponytail selalu aktif (mode full).** Perubahan terkecil yang benar-benar
  menyelesaikan unit: pakai ulang komponen/fungsi/pola yang sudah ada sebelum
  membuat yang baru, stdlib dan dependensi yang sudah terpasang sebelum
  menambah apa pun, tanpa abstraksi untuk kebutuhan yang belum ada (YAGNI),
  tanpa fitur di luar unit. Bila ada jalan yang lebih pendek dan tetap benar
  menurut spec, jalan itulah yang diambil — dan sebutkan di ringkasan akhir.
- Kerjakan HANYA unit yang tertulis di prompt tugasmu, dari PLAN.md.
- Sebelum selesai: jalankan test suite repo terkait (backend: `go test ./...`;
  frontend: `npm run build` harus lulus termasuk `tsc`). Tambahkan test untuk
  perilaku baru mengikuti gaya test yang sudah ada (backend: integrasi HTTP
  di paket router dengan sqlite).
- Fitur baru yang menyentuh skema: gunakan AutoMigrate GORM (model baru/field
  baru), tanpa file migrasi SQL manual, dan pastikan data lama tetap valid
  (default/backfill yang aman).
- Jangan commit. PM yang meninjau, meng-QA, dan commit per unit.
- Akhiri dengan ringkasan: file yang diubah, keputusan yang diambil, hasil
  test, dan hal yang belum pasti (bila ada) — jujur, tanpa melebih-lebihkan.

## Arsitektur backend (Go, repo backend/)
- Layering ketat: `handler/` (HTTP, parsing, respons) → `service/` (aturan
  bisnis) → `repo/` (akses data GORM) → `model/` (struct & konstanta).
  Handler tidak menulis query; repo tidak tahu HTTP.
- Routing terpusat di `router/router.go`; grup admin dipisah dari grup kasir.
- Setiap fitur baru mengikuti pembagian file yang sama (mis. fitur X →
  handler/x.go, service/x.go, repo/x.go, model/x.go bila perlu).

## Arsitektur frontend (React+Vite+TS, repo frontend/web/)
- Halaman di `src/pages/`, komponen bersama di `src/components/` (termasuk
  `components/ui` bergaya shadcn — PAKAI komponen ini, jangan buat sistem
  tombol/form/modal baru), pemanggilan API hanya lewat `src/lib/api.ts`.
- Halaman baru harus mengikuti pola AppShell dan pola halaman terdekat yang
  sudah ada (struktur daftar/form/aksi yang sama). Tanpa library UI baru,
  tanpa ikon set baru, tanpa font baru.
- Folder `android/` (Capacitor) TIDAK disentuh dalam fase pengembangan ini.
  Electron sudah dihapus dan tidak boleh dihidupkan lagi.
- **DILARANG build ke `web/dist`.** Folder `web/dist` adalah PRODUKSI LIVE
  yang disajikan nginx ke pengguna. Verifikasi build oleh agen pengkode
  WAJIB memakai outDir sementara di luar dist (mis.
  `npm run build -- --outDir /tmp/build-cek`, lalu folder itu dihapus).
  Deploy ke dist hanya dilakukan PM sesudah QA. (Aturan ini lahir dari
  insiden 2026-10-15: build verifikasi membocorkan halaman Promo yang
  backend-nya belum ada ke produksi.)
- Sebelum menambah kontrol UI baru (tombol/menu/aksi), PERIKSA dulu apakah
  halaman itu sudah punya kontrol dengan fungsi yang sama atau mirip.
  Dilarang ada dua kontrol untuk satu fungsi: fitur baru yang tumpang
  tindih dengan kontrol lama harus MELEBUR ke kontrol lama, bukan berdiri
  di sebelahnya. (Insiden yang sama: tombol "Ekspor CSV" baru berdiri di
  samping menu "Export" bawaan.)
- Cache build/test (Go/npm) JANGAN ditaruh di /tmp — /tmp di server ini tmpfs kecil (3,9 GB) dan pernah penuh sampai menggagalkan test & deploy. Pakai folder di disk, mis. GOCACHE=/opt/.cache-go-build, GOTMPDIR=/opt/.cache-go-tmp, npm --cache /opt/.cache-npm.
