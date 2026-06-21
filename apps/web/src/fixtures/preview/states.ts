import type { PreviewStateFixture } from './types';

export const requiredPreviewStateTones = ['loading', 'empty', 'error', 'forbidden', 'unavailable'] as const;

const shared = {
  loading: {
    tone: 'loading',
    title: 'Menyiapkan tampilan contoh…',
    message: 'Ruang loading sudah tersedia agar halaman terasa stabil saat koneksi server aktif nanti.',
  },
  empty: {
    tone: 'empty',
    title: 'Belum ada data contoh untuk bagian ini.',
    message: 'Dipakai saat daftar, filter, atau periode tertentu memang belum punya data yang bisa ditampilkan.',
  },
  error: {
    tone: 'error',
    title: 'Tampilan error contoh. Koneksi server belum aktif.',
    message: 'Pesan error disiapkan tanpa membuat data pengganti di browser.',
  },
  forbidden: {
    tone: 'forbidden',
    title: 'Akses ini akan mengikuti hak pengguna saat login asli aktif.',
    message: 'Hak akses final tetap ditentukan oleh role pengguna dan aturan server.',
  },
  unavailable: {
    tone: 'unavailable',
    title: 'Fitur ini belum tersedia di mode contoh.',
    message: 'Aksi dan data asli tetap dikunci sampai koneksi server disetujui.',
  },
} satisfies Record<(typeof requiredPreviewStateTones)[number], PreviewStateFixture>;

function states(overrides: Partial<Record<(typeof requiredPreviewStateTones)[number], Partial<PreviewStateFixture>>> = {}) {
  return requiredPreviewStateTones.map((tone) => ({ ...shared[tone], ...overrides[tone], tone })) satisfies readonly PreviewStateFixture[];
}

export const previewStateMatrix = {
  dashboard: states({
    unavailable: { title: 'Ringkasan dashboard belum terhubung', message: 'KPI, aktivitas, dan alert tenant akan memakai data server saat fase koneksi dibuka.' },
  }),
  catalog: states({
    loading: { title: 'Menyiapkan daftar produk…', message: 'Skeleton tabel produk dan kategori sudah siap untuk koneksi data nanti.' },
    unavailable: { title: 'Editor produk belum tersedia di mode contoh', message: 'Tambah, ubah, publish, dan urutkan produk tetap dikunci sampai koneksi server aktif.' },
  }),
  inventory: states({
    loading: { title: 'Menyiapkan tampilan stok…', message: 'Area stok, pergerakan barang, dan transfer disiapkan tanpa menghitung stok di browser.' },
    forbidden: { title: 'Akses stok mengikuti outlet pengguna', message: 'Hak lihat stok per outlet akan mengikuti aturan server saat login asli aktif.' },
  }),
  reports: states({
    empty: { title: 'Belum ada data laporan untuk filter ini.', message: 'Dipakai saat periode, outlet, atau filter tertentu belum memiliki data.' },
    unavailable: { title: 'Data laporan belum terhubung', message: 'Angka, export, pagination, dan visibilitas laporan menunggu koneksi server.' },
  }),
  settings: states({
    forbidden: { title: 'Pengaturan mengikuti hak pengguna', message: 'Hak edit konfigurasi bisnis dan outlet tetap divalidasi oleh server.' },
    unavailable: { title: 'Form pengaturan masih dikunci', message: 'Tidak ada perubahan konfigurasi yang diproses di mode contoh.' },
  }),
  staff: states({
    forbidden: { title: 'Akses tim mengikuti role pengguna', message: 'Role, outlet, dan izin undangan akan mengikuti aturan server.' },
    unavailable: { title: 'Undangan staf belum tersedia', message: 'Login asli dan sesi pengguna belum aktif di fase UI-first.' },
  }),
  promotions: states({
    error: { title: 'Tampilan error promosi contoh', message: 'Slot ini disiapkan untuk pesan error dari server, bukan simulasi hitungan diskon.' },
    unavailable: { title: 'Manajemen promosi belum tersedia', message: 'Tambah, ubah, dan publikasi promosi tetap dikunci sampai koneksi server aktif.' },
  }),
  audit: states({
    forbidden: { title: 'Log mengikuti hak akses pengguna', message: 'Server tetap menentukan siapa yang boleh melihat log dan detail audit.' },
    unavailable: { title: 'Audit belum terhubung', message: 'Halaman akan memakai data audit asli setelah koneksi server dibuka.' },
  }),
  customers: states({
    unavailable: { title: 'Data pelanggan belum terhubung', message: 'Profil, persetujuan, import, dan export pelanggan masih dikunci di mode contoh.' },
  }),
  attendance: states({
    forbidden: { title: 'Absensi mengikuti hak akses tim', message: 'Koreksi absensi dan visibilitas payroll tetap divalidasi oleh server.' },
    unavailable: { title: 'Data absensi belum terhubung', message: 'Catatan absensi hanya contoh sampai koneksi server aktif.' },
  }),
  outlets: states({
    forbidden: { title: 'Akses outlet mengikuti hak pengguna', message: 'Daftar outlet dan terminal akan mengikuti aturan akses dari server.' },
    unavailable: { title: 'Operasional outlet belum tersedia', message: 'Buka/tutup outlet dan kunci terminal tetap dikunci di mode contoh.' },
  }),
  transactions: states({
    error: { title: 'Tampilan error transaksi contoh', message: 'Slot error disediakan tanpa membuat transaksi palsu di browser.' },
    unavailable: { title: 'Data transaksi belum terhubung', message: 'Total, pajak, refund, void, dan receipt tetap menjadi tanggung jawab server.' },
  }),
  payments: states({
    error: { title: 'Tampilan error pembayaran contoh', message: 'Error gateway dan settlement akan ditampilkan dari server saat tersedia.' },
    unavailable: { title: 'Aksi pembayaran masih dikunci', message: 'Refund, void, retry, dan settlement belum tersedia di mode contoh.' },
  }),
  subscription: states({
    forbidden: { title: 'Billing mengikuti hak pengguna', message: 'Perubahan langganan tenant tetap ditentukan oleh platform dan aturan server.' },
    unavailable: { title: 'Billing langganan belum aktif', message: 'Ubah paket, pembatalan, dan export invoice masih dikunci.' },
  }),
  platform: states({
    loading: { title: 'Menyiapkan control center…', message: 'Skeleton tenant, billing, audit, dan kesehatan sistem siap untuk koneksi data SaaS.' },
    forbidden: { title: 'Akses platform mengikuti role operator', message: 'Owner/operator platform tetap divalidasi server dan tidak memakai konteks outlet tenant.' },
    unavailable: { title: 'Kontrol platform belum terhubung', message: 'Suspend tenant, ubah paket, pendapatan, dan kesehatan sistem menunggu koneksi platform.' },
  }),
} as const satisfies Record<string, readonly PreviewStateFixture[]>;

export type PreviewStateModule = keyof typeof previewStateMatrix;
