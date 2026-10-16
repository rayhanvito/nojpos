# PLAN.md — Pengembangan NOJPOS Fase 0–2

Spec resmi dari PM (Muse) untuk agen pengkode (Codex). Aturan umum di
AGENTS.md berlaku penuh. Setiap unit di bawah ini adalah satu tugas terpisah:
kerjakan berurutan, jangan melompat unit, jangan menggabungkan unit.
Status tiap unit dicatat PM di `.loop/state.md` — agen tidak mengubah file itu.

Aturan bisnis global (berlaku lintas unit, tidak boleh ditafsirkan lain):
- Semua angka uang integer Rupiah. Semua perhitungan diskon/pajak/promo
  dilakukan di SERVER; frontend hanya menampilkan hasil server.
- Urutan hitung checkout (final, berlaku setelah F1.1):
  subtotal → potongan promo/voucher (bila ada) → diskon manual TIDAK bisa
  digabung dengan promo/voucher (lihat F1.1) → pajak aditif dari
  (subtotal − potongan) → total.
- Setiap perubahan stok WAJIB tercatat sebagai stock movement dengan tipe &
  alasan (pola yang sudah ada: adjust/sale/refund/...), tidak ada perubahan
  stok senyap.

---

## FASE 0 — Fondasi

### Unit F0.1 (backend) — Semantik refund yang benar di laporan
Latar: model `Refund` sudah ada (model/transaction.go) dan refund sudah
mengembalikan stok, TETAPI agregasi di `repo/report.go` (Dashboard & Report)
mengabaikan refund: omzet & items_sold tetap bruto (terverifikasi live).
Yang dikerjakan:
1. Dashboard (`ReportRepo.Dashboard`) dan Report (`ReportRepo.Report`):
   omzet, trx_count (tetap menghitung transaksi), items_sold, dan
   gross_profit dihitung NET: penjualan completed dikurangi refund yang
   `created_at`-nya jatuh di periode yang sama (refund diatribusikan ke
   tanggal refund terjadi, BUKAN tanggal penjualan asal).
   - Nilai refund per item = harga snapshot item di transaksi asal × qty
     refund; pajak tidak dihitung ulang (refund mengurangi omzet sebesar
     nilai item + proporsi pajaknya: proporsi = total_pajak_transaksi /
     subtotal_setelah_diskon_transaksi × nilai item; bila penyebut 0 → 0).
     Bulatkan round-half-up ke integer.
   - Gross profit net = (omzet net) − (HPP barang terjual − HPP barang
     refund) dengan HPP dari snapshot buy_price.
2. Report summary menambah field `refund_total` dan `refund_count` untuk
   periode itu; Dashboard `today` menambah `refund_total`. Field lama tidak
   berubah nama/tipe.
3. GET /transactions/:id menyertakan daftar refund transaksi itu
   (field `refunds`: id, created_at, items, reason, actor) dari tabel refunds.
4. Status transaksi TIDAK diubah oleh unit ini (tetap seperti sekarang).
Test: integrasi router — buat penjualan, refund sebagian, assert dashboard &
report net + refund_total benar; refund penuh; refund di hari berbeda tidak
mengurangi omzet hari penjualan.

### Unit F0.2 (backend+frontend) — Ekspor CSV
1. Backend: `GET /transactions/export?date=YYYY-MM-DD&period=` (ikuti pola
   parameter Report: period today/yesterday/week/month/all + date) →
   text/csv, kolom: id, tanggal (ISO, zona toko), kasir, pelanggan, metode,
   status, subtotal, diskon, pajak, total, dibayar, kembalian, jumlah item.
   `GET /products/export` → kolom: id, nama, sku, barcode, kategori, harga
   beli, harga jual, stok, satuan, aktif. Keduanya admin-only, ter-scope toko,
   Content-Disposition attachment dengan nama file bertanggal.
2. Frontend: tombol "Ekspor CSV" di halaman Transaksi (mengunduh sesuai
   filter tanggal yang sedang aktif di halaman itu) dan di halaman Produk —
   mengikuti pola tombol yang sudah ada, tanpa mengubah layout halaman.
Test: endpoint mengembalikan CSV dengan header benar dan baris sesuai data;
kasir ditolak 403 pada kedua endpoint.

---

## FASE 1 — Standar pasar

### Unit F1.1 (backend+frontend) — Promo/voucher engine + diskon terjadwal
Model baru `promos`: nama, tipe (`percent` | `amount`), nilai, cakupan
(`all` | `category` dengan category_id | `product` dengan product_id),
kode voucher (nullable; null = promo otomatis), jadwal: `starts_at`,
`ends_at`, dan jendela jam harian opsional (`happy_start`, `happy_end`,
format HH:MM, zona toko), `min_purchase`, `max_discount` (untuk percent),
`active`. CRUD admin penuh di `/promos` (+ PATCH active), pola endpoint
mengikuti categories/products.
Aturan checkout (TEGAS, sesuai urutan global di atas):
- Kasir boleh memasukkan SATU kode voucher. Bila kode valid (ada, aktif,
  dalam jadwal, min_purchase terpenuhi) → promo voucher dipakai.
- Bila tidak ada kode: sistem mengevaluasi promo OTOMATIS yang sedang
  berlaku; bila lebih dari satu cocok, dipakai yang potongannya TERBESAR
  (satu saja, tidak bertumpuk).
- Bila promo/voucher dipakai, field `discount` manual dari kasir DIABAIKAN
  (diperlakukan 0) — tidak ada gabungan promo + diskon manual.
- Potongan dihitung per cakupan: percent × nilai item yang tercakup (dibatasi
  max_discount), atau amount tetap (tidak melebihi nilai item tercakup).
- Transaksi menyimpan snapshot: `promo_id`, `promo_name`, `promo_discount`
  (terpisah dari `discount` manual). Respons checkout & struk menampilkan
  potongan promo sebagai baris sendiri.
Frontend: halaman admin "Promo" (CRUD + toggle aktif, pola halaman Produk);
di halaman POS: kolom input kode voucher + tampilkan promo yang terpakai dan
besar potongannya (dari respons server) sebelum bayar.
Test: voucher valid/kedaluwarsa/salah kode; happy hour di dalam & di luar
jendela; dua promo otomatis → yang terbesar; promo + diskon manual → manual
diabaikan; cakupan kategori hanya memotong item kategori itu.
Keputusan PM (2026-10-15, mengikat — Codex sempat berhenti menanyakan ini):
- Nilai promo `percent` BOLEH desimal (mis. 12.5), mengikuti pola `taxPct`
  yang float di codebase ini; rentang sah percent: 0 < nilai ≤ 100. Hasil
  potongan uang tetap dibulatkan round-half-up ke integer Rupiah.
- Happy hour dengan happy_start > happy_end adalah SAH dan berarti lintas
  tengah malam (mis. 22:00–02:00 = pukul 22:00–24:00 dan 00:00–02:00, zona
  toko). Jendela normal (start ≤ end) berlaku [start, end).

### Unit F1.2 (backend+frontend) — CRM pelanggan + loyalty poin
Model baru `customers`: nama, telepon (unik per toko bila diisi), email
opsional, alamat opsional, `points` (saldo poin), timestamps, soft delete.
CRUD admin/kasir di `/customers` (kasir boleh membaca & membuat, tidak boleh
menghapus — mengikuti pola RBAC yang ada). Checkout menerima `customerId`
opsional (field `customer` teks lama tetap ada untuk walk-in tanpa akun).
Aturan poin (TEGAS):
- Pengaturan toko menambah: `loyaltyEnabled` (bool), `loyaltyEarnPer`
  (Rupiah per 1 poin; default 10000), `loyaltyPointValue` (nilai tukar
  1 poin dalam Rupiah; default 100).
- Dapat poin: hanya bila loyaltyEnabled dan transaksi completed dengan
  customerId: poin = floor(total_setelah_pajak / loyaltyEarnPer).
  Refund mengurangi kembali poin proporsional dari transaksi itu
  (floor dari nilai refund / loyaltyEarnPer), saldo tidak boleh negatif
  (mentok 0).
- Tukar poin: kasir boleh mengirim `pointsRedeem` saat checkout; nilai
  potongannya = pointsRedeem × loyaltyPointValue, tidak boleh melebihi
  total sebelum tukar dan tidak melebihi saldo pelanggan. Potongan poin
  diperlakukan seperti potongan non-promo SEBELUM pajak? — TIDAK: tukar
  poin adalah pembayaran: ia mengurangi jumlah yang harus dibayar SETELAH
  pajak (seperti tunai), dicatat terpisah `points_redeemed` &
  `points_discount` pada transaksi, tidak mengubah dasar pajak.
- Riwayat: `GET /customers/:id/transactions` (transaksi pelanggan itu) dan
  saldo poin ikut di respons detail pelanggan.
Frontend: halaman "Pelanggan" (CRUD + riwayat + saldo poin); di POS: pemilih
pelanggan (pencarian nama/telepon) + opsi tukar poin bila saldo cukup;
pengaturan loyalty di halaman Pengaturan (pola form settings yang ada).
Test: poin bertambah sesuai rumus; refund mengurangi poin; tukar poin tidak
melebihi saldo/total; pelanggan walk-in (tanpa customerId) tidak terpengaruh;
telepon duplikat per toko ditolak.

### Unit F1.3 (backend+frontend) — Inventori tahap 1
Tiga sub-fitur, satu unit, urutan bebas di dalam unit:
1. **Varian produk**: model `product_variants` (product_id, nama varian,
   sku unik per toko, buy_price, sell_price, stock). Produk tanpa varian
   berperilaku persis seperti sekarang. Produk bervarian: stok & harga
   hidup di varian; checkout menerima `variantId` opsional per item
   (item tanpa variantId untuk produk bervarian → 400). Movement mencatat
   varian. Frontend: editor varian di form produk + pemilih varian di POS.
2. **Multi-satuan & harga bertingkat**: field produk `unit` yang ada tetap
   satuan jual terkecil. Tabel `price_tiers` per produk (min_qty, sell_price):
   saat checkout, harga item = tier dengan min_qty terbesar yang ≤ qty
   (fallback harga normal). Ini menggantikan negosiasi grosir manual dan
   berlaku otomatis, termasuk untuk varian (tier per varian bila variantId
   ada, else per produk). Frontend: editor tier di form produk; POS
   menampilkan harga tier yang sedang berlaku saat qty berubah (dari data
   produk yang sudah dimuat, perhitungan final tetap di server).
3. **Alert stok menipis**: field produk/varian `min_stock` (default 0 =
   nonaktif; default global di settings `lowStockDefault`). Saat stok
   menyentuh ≤ min_stock sesudah transaksi/adjustment/refund, buat notifikasi
   kategori stok (pola notifikasi yang sudah ada), maksimal 1 notifikasi per
   produk per hari. Dashboard `low_stock` yang sudah ada memakai ambang ini.
Test: checkout varian mengurangi stok varian bukan induk; tier qty 10 memakai
harga tier; notifikasi stok menipis muncul sekali; produk lama tanpa varian
& tier tetap berfungsi (regresi checkout dasar).
Keputusan PM (2026-10-15, mengikat — Codex menanyakan ini di akhir F1.3):
DELETE varian = soft delete, SELALU diizinkan (mengikuti pola hapus produk
yang sudah ada); stok tersisa pada varian yang dihapus dianggap pensiun.
Transaksi masa lalu tidak terpengaruh (snapshot nama/harga). Refund atas
item varian yang sudah dihapus tetap diproses; stok dikembalikan ke
record varian tersebut.

---

## FASE 2 — Kedalaman operasional

### Unit F2.1 (backend+frontend) — Supplier + purchase order + penerimaan
Model: `suppliers` (nama, kontak, alamat), `purchase_orders` (supplier_id,
nomor PO otomatis `PO-YYYYMMDD-####` per toko, status:
`draft → ordered → partially_received → fully_received`, atau `cancelled`
dari draft/ordered), `purchase_order_items` (product_id/variant_id, qty
pesan, qty diterima, buy_price snapshot).
Alur (TEGAS): PO dibuat draft → `POST /purchase-orders/:id/order`
(draft→ordered; item terkunci) → penerimaan `POST /purchase-orders/:id/receive`
{items:[{itemId, qty}]} boleh berkali-kali (parsial); qty terima kumulatif
tidak boleh melebihi qty pesan; setiap penerimaan menambah stok (movement
tipe `purchase`) dan memperbarui status; `fully_received` otomatis saat
semua item penuh. Batal hanya dari draft/ordered dan tidak menyentuh stok.
CRUD supplier pola standar; halaman admin "Supplier" dan "Pembelian"
(daftar PO + detail + aksi order/terima/batal) mengikuti pola halaman yang
ada.
Test: alur penuh draft→ordered→terima parsial→penuh dengan stok bertambah
tepat; terima melebihi pesan ditolak; batal tidak mengubah stok.

### Unit F2.2 (backend+frontend) — Sesi stock opname
Model: `stock_opname_sessions` (status `counting → posted | cancelled`,
catatan, dibuat oleh), `stock_opname_items` (product/variant, stok sistem
snapshot saat sesi dibuat, stok hitung, selisih terhitung server).
Alur (TEGAS): buat sesi → sistem mengunci snapshot stok saat itu → kasir/
admin mengisi stok hitung per item (boleh bertahap) → posting: untuk setiap
item berselisih, buat adjustment movement tipe `opname` dengan alasan nomor
sesi, stok menjadi stok hitung; sesi `posted` terkunci selamanya. Sesi yang
sedang `counting` maksimal SATU per toko. Halaman "Opname" pola yang ada.
Test: selisih dihitung dari snapshot (bukan stok live saat posting);
posting mengoreksi stok tepat; sesi kedua saat satu aktif ditolak; sesi
posted tidak bisa diubah.

### Unit F2.3 (backend+frontend) — Hutang-piutang terstruktur
Dua sisi, aturan simetris:
- **Piutang pelanggan**: checkout dengan metode pembayaran baru `Credit`
  (satu-satunya penambahan enum; PM menyetujui di sini) WAJIB customerId;
  membuat `receivables` (customer_id, transaction_id, jumlah = total,
  jatuh tempo dari settings `creditDueDays` default 14 hari, status
  `unpaid → partially_paid → paid`). Pembayaran dicatat lewat
  `POST /receivables/:id/payments` {amount, method (enum standar minus
  Credit)}; kumulatif tidak boleh melebihi sisa.
- **Hutang supplier**: penerimaan PO (F2.1) membuat `payables`
  (supplier_id, po_id, jumlah = nilai barang diterima, status sama);
  pembayaran lewat `POST /payables/:id/payments`.
- Endpoint daftar keduanya + filter status (belum lunas/jatuh tempo),
  halaman "Hutang Piutang" dua tab. Transaksi Credit TIDAK menghitung
  `paid` tunai: checkout Credit mensyaratkan paid = 0 dan change = 0.
Test: jual kredit → piutang lahir sebesar total; bayar sebagian → status
partial; lunasi → paid; bayar melebihi sisa ditolak; PO terima → hutang
lahir; checkout Credit tanpa customerId ditolak.

### Unit F2.4 (backend+frontend) — Laporan pajak-ready + ekspor lanjutan
1. `GET /reports/tax?month=YYYY-MM`: rekap bulan itu — omzet bruto, refund,
   omzet neto, DPP (dasar pengenaan: total sebelum pajak), PPN terkumpul
   (jumlah tax), jumlah transaksi, dan estimasi PPh Final 0,5% × omzet neto
   (dibulatkan ke Rupiah penuh). Semua dari data transaksi completed & refund
   (konsisten dengan netting F0.1), zona waktu toko.
2. Ekspor CSV laporan itu (pola F0.2) + bagian "Pajak" di halaman Laporan
   menampilkan tabel bulanan tahun berjalan.
3. (Bila mudah dari pola yang ada) ekspor CSV stok/produk sudah di F0.2 —
   unit ini tidak mengulanginya.
Test: angka rekap cocok dengan penjualan+refund yang dibuat di bulan itu;
bulan kosong → nol semua, bukan error.

### Unit F2.5 (frontend+backend) — Mode offline yang benar untuk web
Tujuan: kasir tetap bisa menjual saat internet putus; data menyusul sinkron.
- Frontend: service worker (pola Vite yang ada) meng-cache shell aplikasi;
  transaksi yang dibuat saat offline disimpan di antrean IndexedDB
  (halaman offline yang sudah ada adalah modal pola datanya — satukan,
  jangan buat sistem paralel kedua).
- Backend: checkout menerima header/flag `client_tx_id` (UUID dari klien);
  server menyimpan client_tx_id unik per toko → sinkronisasi ulang tidak
  menggandakan transaksi (idempotent). Respons untuk transaksi tersinkron
  menyertakan status stok final.
- Aturan konflik (TEGAS): transaksi offline TIDAK PERNAH ditolak saat dibuat
  di klien; saat sinkron, server memvalidasi stok — bila stok tidak cukup,
  transaksi tetap diterima sebagai `backorder` berstatus khusus
  `completed_offline_shortage` dan memicu notifikasi ke admin; tidak ada
  penghapusan senyap. Antrean klien menampilkan status tiap item
  (menunggu/tersinkron/gagal-perlu-tindakan).
- Kasir tahu ia offline: indikator status koneksi yang jelas di AppShell.
Test: kirim checkout yang sama dua kali dengan client_tx_id sama → satu
transaksi; checkout offline yang disinkron mengurangi stok sekali;
kekurangan stok saat sinkron menghasilkan status khusus + notifikasi.

### Unit F2.6 (backend+frontend) — Fondasi multi-outlet
Refactor model data (unit terberat; kerjakan paling akhir):
- Model `outlets` (store_id, nama, alamat, is_default). Migrasi AutoMigrate +
  backfill: setiap toko yang ada mendapat satu outlet default; semua produk,
  stok, transaksi, shift, user yang ada terhubung ke outlet default itu.
- Stok menjadi per outlet: tabel `product_stocks` (product/variant, outlet,
  qty) menggantikan kolom stock tunggal sebagai sumber kebenaran; kolom lama
  diisi dari outlet default selama masa transisi dan dihapus hanya bila aman
  (bila ragu, biarkan terisi — laporkan ke PM).
- Token/sesi membawa outlet aktif; switch outlet untuk admin; kasir terikat
  satu outlet. Checkout, adjustment, opname, PO semuanya ter-scope outlet.
- Dashboard & laporan bisa difilter per outlet atau gabungan (parameter
  `outletId` opsional; tanpa parameter = outlet aktif untuk kasir, gabungan
  untuk admin).
- Perilaku toko satu-outlet yang ada TIDAK berubah dari sisi user.
Test: regresi penuh alur dasar di outlet default; dua outlet: stok tidak
bercampur, penjualan outlet A tidak mengurangi stok outlet B; dashboard
gabungan menjumlah keduanya.

### Unit F2.7 (backend+frontend) — Shift lengkap: tutup shift terhitung + Laporan Shift
Ditambahkan atas instruksi Rayhan 2026-10-15. Latar terverifikasi: model
`CashierShift` saat ini hanya menyimpan modal awal; endpoint admin
`GET /shifts` sudah menghitung total penjualan + jumlah transaksi per shift,
tetapi tidak ada halaman laporannya, tidak ada rincian per metode bayar, dan
uang akhir laci tidak disimpan. Unit ini melengkapinya TANPA absensi
(absensi ditunda oleh Rayhan) dan TANPA komisi (paket terpisah).
Backend:
1. `CashierShift` ditambah: `closing_cash_actual` (uang fisik yang dihitung
   kasir saat tutup; null selama shift terbuka), `expected_cash` (dihitung
   server saat tutup = opening_cash + penjualan metode Cash dalam shift −
   refund atas transaksi Cash dalam shift itu), `cash_variance`
   (= closing_cash_actual − expected_cash; tersimpan), `close_note`
   (opsional), `closed_by_name`. AutoMigrate saja.
2. `POST /cashier/shift/close` menerima `closingCash` (wajib, ≥ 0) + `note`
   opsional. `GET /cashier/shift` (shift aktif) diperkaya sebagai bahan
   review sebelum tutup: total penjualan, jumlah transaksi, rincian per
   metode pembayaran (semua metode enum yang berlaku), total refund dalam
   shift, dan expected_cash sementara. Respons close memuat summary lengkap
   termasuk variance.
3. Atribusi (pola existing, ditegaskan): penjualan shift = transaksi dengan
   cashier_id pemilik shift DAN created_at dalam [started_at, closed_at].
   Ganti operator via PIN tidak memindahkan kepemilikan shift.
4. `GET /shifts` (admin) diperkaya per shift: rincian per metode, refund
   total, expected_cash, closing_cash_actual, cash_variance, close_note
   (shift terbuka: angka berjalan, field tutup null). `GET /shifts/:id`
   (admin): detail satu shift + daftar transaksinya (pola pagination yang
   ada). `GET /shifts/export` (admin): CSV pola F0.2 — id, kasir, mulai,
   selesai, modal awal, penjualan per metode (satu kolom per metode),
   refund, total penjualan, jumlah transaksi, expected, aktual, selisih.
5. Admin boleh force-close shift yang masih terbuka
   (`POST /shifts/:id/force-close`; closingCash opsional — bila kosong,
   closing_cash_actual & variance null, close_note otomatis
   "Ditutup paksa oleh admin"). Kasir hanya bisa menutup shiftnya sendiri.
Frontend:
1. Halaman admin "Shift": daftar shift (pola tabel halaman admin lain);
   klik baris → detail (ringkasan + daftar transaksi); tombol "Ekspor CSV";
   entri menu mengikuti pola navigasi yang ada.
2. Dialog tutup shift pada alur yang ada: sebelum konfirmasi tampilkan
   ringkasan review (rincian per metode, refund, expected cash) + input
   uang aktual + catatan opsional; sesudah tutup tampilkan selisihnya.
   Teks Bahasa Indonesia, komponen yang ada, layout lain tidak berubah.
Urutan: unit ini dikerjakan SESUDAH F2.6 dan memakai scope outlet dari
F2.6 (shift terikat outlet aktif; laporan menerima filter outletId
mengikuti pola F2.6).
Test: close dengan closingCash → expected & variance benar (termasuk
refund tunai); rincian per metode cocok dengan transaksi; force-close
admin; kasir menutup shift orang lain ditolak; ekspor CSV header benar.

### Unit F2.8 (backend+frontend) — Import produk via CSV
Ditambahkan atas instruksi Rayhan 2026-10-15 (ekspor sudah ada di F0.2;
unit ini sisi masuknya, untuk onboarding katalog awal toko).
Backend:
1. `POST /products/import` (admin-only, ter-scope toko): unggah multipart
   CSV dengan kolom PERSIS format ekspor F0.2 (nama, sku, barcode,
   kategori, harga beli, harga jual, stok, satuan, aktif) supaya
   ekspor → sunting → impor bisa bolak-balik.
2. Aturan (TEGAS): impor hanya MEMBUAT produk baru — SKU yang sudah ada
   (di toko atau duplikat di dalam file) membuat baris itu DITOLAK dan
   dilaporkan, tidak pernah menimpa produk/stok yang ada. Kategori disebut
   dengan nama; kategori yang belum ada DIBUAT otomatis. Baris tidak valid
   (nama kosong, harga/stok negatif atau bukan angka) ditolak + dilaporkan.
   Impor bersifat parsial: baris valid tetap masuk, baris gagal tidak.
   Maksimal 1.000 baris data per berkas.
3. Respons: ringkasan {total baris, berhasil, gagal} + daftar kegagalan
   per baris (nomor baris + alasan). Tidak ada kegagalan yang diam.
4. Stok awal dari impor tercatat sebagai stock movement mengikuti pola
   yang ada (tipe adjust, alasan "import"), bukan perubahan stok senyap.
Frontend: tombol "Import CSV" di halaman Produk (admin, pola tombol yang
ada) membuka dialog unggah; sesudah selesai tampilkan ringkasan hasil +
daftar baris gagal dari respons server; dialog menyediakan unduhan
template CSV (header + satu baris contoh, dibuat di client). Layout &
komponen lain tidak berubah.
Test: impor file campuran (valid, SKU duplikat, harga tidak valid) →
produk valid tercipta, yang gagal dilaporkan per baris; kategori baru
terbuat otomatis; stok awal tercatat sebagai movement; kasir ditolak 403.

---

## Selesai berarti
Semua unit F0.1–F2.8: implementasi + test lulus + ter-deploy + terverifikasi
live oleh PM. PM yang menyatakan selesai kepada pemilik project — agen
pengkode tidak menyatakan selesai ke siapa pun.
