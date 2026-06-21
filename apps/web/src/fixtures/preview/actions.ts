import type { PreviewActionGroupFixture } from './types';

export const previewActionGroups = {
  dashboard: {
    title: 'Aksi Dasbor',
    description: 'Tombol cepat ditampilkan untuk melihat alur kerja. Perilaku asli menunggu koneksi server dan hak akses pengguna.',
    actions: [
      { label: 'Segarkan ringkasan', category: 'read', reason: 'Menunggu data ringkasan dashboard dari server.' },
      { label: 'Buka panel alert', category: 'read', reason: 'Menunggu daftar alert operasional dari server.' },
      { label: 'Export ringkasan', category: 'export', reason: 'Export akan diproses server dan dicatat di audit.' },
    ],
  },
  catalog: {
    title: 'Aksi Produk',
    description: 'Kontrol produk masih dikunci sampai data produk, kategori, media, dan publikasi terhubung ke server.',
    actions: [
      { label: 'Tambah produk', category: 'write', reason: 'Membutuhkan validasi produk dan media dari server.' },
      { label: 'Ubah visibilitas massal', category: 'write', reason: 'Membutuhkan proses massal dan jejak audit dari server.' },
      { label: 'Import katalog', category: 'import', reason: 'Membutuhkan template import dan pemeriksaan data oleh server.' },
      { label: 'Export katalog', category: 'export', reason: 'Export akan mengikuti format dan izin dari server.' },
    ],
  },
  inventory: {
    title: 'Aksi Stok',
    description: 'Aksi stok dikunci karena pergerakan barang harus mengikuti data dan persetujuan dari server.',
    actions: [
      { label: 'Buat penyesuaian', category: 'write', reason: 'Membutuhkan alur mutasi stok dan audit dari server.' },
      { label: 'Transfer stok', category: 'write', reason: 'Membutuhkan aturan transfer antar outlet dari server.' },
      { label: 'Import stok awal', category: 'import', reason: 'Membutuhkan pemeriksaan file dan konflik data oleh server.' },
      { label: 'Export mutasi stok', category: 'export', reason: 'Export mutasi akan diproses server berdasarkan periode.' },
    ],
  },
  reports: {
    title: 'Aksi Laporan',
    description: 'Aksi laporan dikunci sampai server menangani angka, filter, zona waktu, dan export.',
    actions: [
      { label: 'Terapkan filter laporan', category: 'read', reason: 'Membutuhkan skema filter dan pagination dari server.' },
      { label: 'Jadwalkan laporan', category: 'write', reason: 'Membutuhkan job server dan pengaturan pengiriman.' },
      { label: 'Export laporan', category: 'export', reason: 'Export akan mengikuti antrean, izin, dan retensi file dari server.' },
    ],
  },
  settings: {
    title: 'Aksi Pengaturan',
    description: 'Kontrol pengaturan dikunci sampai setiap bagian punya validasi, penyimpanan, dan audit dari server.',
    actions: [
      { label: 'Simpan profil bisnis', category: 'write', reason: 'Membutuhkan validasi profil bisnis dari server.' },
      { label: 'Ubah pengaturan outlet', category: 'write', reason: 'Membutuhkan hak akses outlet dan audit dari server.' },
      { label: 'Lihat contoh struk', category: 'read', reason: 'Preview struk asli akan dibuat oleh server.' },
      { label: 'Export audit pengaturan', category: 'export', reason: 'Export audit menunggu data dari server.' },
    ],
  },
  staff: {
    title: 'Aksi Tim',
    description: 'Manajemen tim masih dikunci sampai role, undangan, scope outlet, dan audit siap.',
    actions: [
      { label: 'Undang staf', category: 'write', reason: 'Membutuhkan alur undangan dan masa berlaku dari server.' },
      { label: 'Ubah role', category: 'write', reason: 'Membutuhkan matriks role dan cek izin server.' },
      { label: 'Nonaktifkan staf', category: 'danger', reason: 'Membutuhkan konfirmasi, perlindungan akun owner, dan audit.' },
      { label: 'Export daftar staf', category: 'export', reason: 'Export staf mengikuti izin dan masking data.' },
    ],
  },
  promotions: {
    title: 'Aksi Promo',
    description: 'Aksi promo dikunci karena harga, syarat promo, dan kuota harus dihitung oleh server.',
    actions: [
      { label: 'Buat kampanye', category: 'write', reason: 'Membutuhkan aturan promo dan cek konflik dari server.' },
      { label: 'Publikasikan kampanye', category: 'write', reason: 'Membutuhkan periode publish, scope outlet, dan audit.' },
      { label: 'Jeda kampanye', category: 'danger', reason: 'Membutuhkan perubahan status dan audit dari server.' },
      { label: 'Export daftar promo', category: 'export', reason: 'Export promo mengikuti izin dan format server.' },
    ],
  },
  audit: {
    title: 'Aksi Aktivitas',
    description: 'Aksi audit fokus untuk baca data dan masih dikunci sampai pencarian, retensi, dan export siap.',
    actions: [
      { label: 'Filter aktivitas', category: 'read', reason: 'Membutuhkan pencarian, retensi, dan pagination dari server.' },
      { label: 'Buka bukti event', category: 'read', reason: 'Membutuhkan detail event dan masking data.' },
      { label: 'Export log audit', category: 'export', reason: 'Export log mengikuti izin, scope, retensi, dan approval.' },
    ],
  },
  customers: {
    title: 'Aksi Pelanggan',
    description: 'Kontrol pelanggan dikunci sampai privasi, persetujuan, import, dan export siap.',
    actions: [
      { label: 'Tambah pelanggan', category: 'write', reason: 'Membutuhkan data pelanggan, persetujuan, dan validasi server.' },
      { label: 'Ubah pelanggan', category: 'write', reason: 'Membutuhkan izin ubah dan masking data.' },
      { label: 'Import pelanggan', category: 'import', reason: 'Membutuhkan template import dan pemeriksaan persetujuan.' },
      { label: 'Export pelanggan', category: 'export', reason: 'Export pelanggan mengikuti izin dan masking data.' },
    ],
  },
  attendance: {
    title: 'Aksi Absensi',
    description: 'Koreksi absensi dikunci karena jam kerja dan dampak payroll harus mengikuti aturan server.',
    actions: [
      { label: 'Buat catatan absensi', category: 'write', reason: 'Membutuhkan catatan, approval, dan audit dari server.' },
      { label: 'Ubah record absensi', category: 'write', reason: 'Membutuhkan alur koreksi dan cek izin server.' },
      { label: 'Import absensi', category: 'import', reason: 'Membutuhkan pemeriksaan file dan konflik data.' },
      { label: 'Export absensi', category: 'export', reason: 'Export absensi membutuhkan izin dan harus aman untuk payroll.' },
    ],
  },
  outlets: {
    title: 'Aksi Outlet',
    description: 'Kontrol operasional outlet dikunci sampai perubahan status, terminal, dan audit siap.',
    actions: [
      { label: 'Tambah outlet', category: 'write', reason: 'Membutuhkan data outlet dan kuota tenant dari server.' },
      { label: 'Ubah outlet', category: 'write', reason: 'Membutuhkan validasi dan hak akses outlet.' },
      { label: 'Buka/tutup outlet', category: 'danger', reason: 'Membutuhkan perubahan status operasional dan audit.' },
      { label: 'Kunci/buka terminal', category: 'danger', reason: 'Membutuhkan status terminal dan izin operator.' },
    ],
  },
  transactions: {
    title: 'Aksi Transaksi',
    description: 'Kontrol transaksi hanya untuk lihat data karena total, pajak, refund, dan void harus diproses server.',
    actions: [
      { label: 'Filter transaksi', category: 'read', reason: 'Membutuhkan filter, pagination, dan scope outlet dari server.' },
      { label: 'Void transaksi', category: 'danger', reason: 'Membutuhkan alur void, pencegahan duplikasi, dan audit.' },
      { label: 'Refund transaksi', category: 'danger', reason: 'Membutuhkan aturan refund, gateway, dan audit.' },
      { label: 'Export transaksi', category: 'export', reason: 'Export transaksi mengikuti izin dan proses server.' },
    ],
  },
  payments: {
    title: 'Aksi Pembayaran',
    description: 'Aksi pembayaran dikunci sampai gateway, settlement, pencegahan duplikasi, dan audit siap.',
    actions: [
      { label: 'Coba ulang pembayaran', category: 'write', reason: 'Membutuhkan status gateway dan proses server.' },
      { label: 'Refund pembayaran', category: 'danger', reason: 'Membutuhkan aturan refund, gateway, dan audit.' },
      { label: 'Void pembayaran', category: 'danger', reason: 'Membutuhkan transisi status dan settlement.' },
      { label: 'Export pembayaran', category: 'export', reason: 'Export pembayaran mengikuti izin dan masking data.' },
    ],
  },
  subscription: {
    title: 'Aksi Langganan',
    description: 'Kontrol langganan tenant dikunci sampai billing dan aturan platform siap.',
    actions: [
      { label: 'Ubah paket', category: 'write', reason: 'Membutuhkan aturan billing, prorata, dan approval owner.' },
      { label: 'Batalkan langganan', category: 'danger', reason: 'Membutuhkan aturan pembatalan dan audit.' },
      { label: 'Ubah metode bayar', category: 'write', reason: 'Membutuhkan handoff billing yang aman.' },
      { label: 'Export invoice', category: 'export', reason: 'Membutuhkan invoice dari server dan izin export.' },
    ],
  },
  platform: {
    title: 'Aksi Platform',
    description: 'Kontrol Platform Super Admin dikunci sampai izin platform, billing, dan audit siap.',
    actions: [
      { label: 'Suspend tenant', category: 'danger', reason: 'Membutuhkan aturan suspend, pengaman, dan audit.' },
      { label: 'Ubah paket', category: 'write', reason: 'Membutuhkan billing platform dan aturan prorata.' },
      { label: 'Buat operator platform', category: 'write', reason: 'Membutuhkan role platform dan alur undangan.' },
      { label: 'Export laporan platform', category: 'export', reason: 'Export finance harus mengikuti izin dan proses server.' },
    ],
  },
  detail: {
    title: 'Aksi Detail',
    description: 'Aksi level record ditampilkan untuk menguji layout. Perubahan asli tetap menunggu koneksi server.',
    actions: [
      { label: 'Ubah record', category: 'write', reason: 'Membutuhkan validasi, izin, dan proses server.' },
      { label: 'Duplikasi record', category: 'write', reason: 'Membutuhkan aturan duplikasi dan cek konflik.' },
      { label: 'Arsipkan record', category: 'danger', reason: 'Membutuhkan perubahan status, konfirmasi, dan audit.' },
      { label: 'Export record', category: 'export', reason: 'Membutuhkan format export dan izin pengguna.' },
    ],
  },
} satisfies Record<string, PreviewActionGroupFixture>;
