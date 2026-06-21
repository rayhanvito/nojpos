import type { DetailPreviewFixture, PreviewLane, PreviewMetricFixture, PreviewTableFixture } from './types';

export const platformOverviewMetrics: readonly PreviewMetricFixture[] = [
  { label: 'Total toko', value: '128', description: 'Jumlah toko contoh di platform NojPOS.', badge: 'Data contoh' },
  { label: 'Toko aktif', value: '94', description: 'Toko yang masih aktif memakai NojPOS.', badge: 'Preview' },
  { label: 'Trial', value: '21', description: 'Toko yang sedang mencoba paket.', badge: 'Perlu follow-up' },
  { label: 'Expired', value: '9', description: 'Toko yang masa aktifnya sudah habis.', badge: 'Cek tagihan', tone: 'warning' },
  { label: 'MRR', value: 'Rp18,4 jt', description: 'Contoh agregat bulanan. Belum terhubung billing.', badge: 'Agregat contoh' },
  { label: 'Tagihan belum dibayar', value: '7', description: 'Tagihan contoh yang perlu dicek tim finance.', badge: 'Perlu dicek', tone: 'warning' },
  { label: 'Bantuan terbuka', value: '12', description: 'Tiket bantuan contoh dari toko.', badge: 'Support' },
  { label: 'Gangguan sistem', value: '1', description: 'Status contoh, bukan monitoring produksi.', badge: 'Preview', tone: 'warning' },
] as const;

export const platformAlerts = [
  { title: '4 toko trial akan habis minggu ini', description: 'Tim bisa menyiapkan follow-up tanpa membuka data transaksi toko.', tone: 'warning' },
  { title: '7 tagihan belum dibayar', description: 'Masuk daftar cek finance. Tandai lunas masih dikunci.', tone: 'warning' },
  { title: '3 tiket bantuan urgent', description: 'Kendala printer dan QRIS/payment perlu diprioritaskan.', tone: 'warning' },
  { title: '1 gangguan webhook payment', description: 'Status gateway hanya preview dan retry belum aktif.', tone: 'neutral' },
  { title: '2 toko butuh verifikasi data', description: 'Akses detail toko membutuhkan alasan dan tercatat di aktivitas.', tone: 'neutral' },
] as const;

export const platformTenantStatus = [
  { name: 'Aktif', value: 94 },
  { name: 'Trial', value: 21 },
  { name: 'Expired', value: 9 },
  { name: 'Suspended', value: 4 },
] as const;

export const platformBillingStatus = [
  { name: 'Lunas', value: 32 },
  { name: 'Belum dibayar', value: 7 },
  { name: 'Gagal', value: 3 },
] as const;

export const platformActivityMetrics: readonly PreviewMetricFixture[] = [
  { label: 'Total transaksi agregat', value: '18.240', description: 'Agregat contoh. Tidak membuka detail transaksi toko.', badge: 'Privacy' },
  { label: 'GMV agregat', value: 'Rp284,5 jt', description: 'GMV ditampilkan agregat untuk preview platform.', badge: 'Agregat' },
  { label: 'Payment failure', value: '3', description: 'Pembayaran gagal contoh. Retry tetap dikunci.', badge: 'Gated', tone: 'warning' },
  { label: 'Toko aktif hari ini', value: '76', description: 'Aktivitas toko contoh untuk pusat kerja internal.', badge: 'Data contoh' },
] as const;

export const platformPrivacyNotes = [
  'Data transaksi dan GMV ditampilkan secara agregat.',
  'Akses detail toko membutuhkan alasan dan tercatat di aktivitas.',
  'Akses bantuan belum aktif pada preview UI.',
  'Aksi sensitif akan membutuhkan audit log saat backend aktif.',
] as const;

export const platformLanes: readonly PreviewLane[] = [
  { label: 'Toko', items: ['Status aktif/trial/expired', 'Kontak owner dimasking', 'Akses bantuan dikunci'] },
  { label: 'Langganan & Tagihan', items: ['Paket aktif', 'Jatuh tempo', 'Pembayaran bermasalah'] },
  { label: 'Operasional NojPOS', items: ['Bantuan', 'Aktivitas penting', 'Kondisi sistem'] },
] as const;

export const platformOperationalLinks = [
  { href: '/platform/businesses', label: 'Toko', description: 'Cek toko aktif, trial, expired, dan perlu verifikasi.' },
  { href: '/platform/plans', label: 'Paket', description: 'Review konsep paket Free, Starter, Pro, dan Business.' },
  { href: '/platform/revenue', label: 'Langganan & Tagihan', description: 'Pantau jatuh tempo, tagihan, dan pembayaran gagal.' },
  { href: '/platform/support', label: 'Bantuan', description: 'Prioritaskan tiket toko yang urgent.' },
  { href: '/platform/audit', label: 'Aktivitas', description: 'Lihat riwayat aksi penting dan akses bantuan.' },
  { href: '/platform/system-health', label: 'Sistem', description: 'Cek layanan yang perlu perhatian.' },
  { href: '/platform/announcements', label: 'Pengumuman', description: 'Siapkan info maintenance, promo, atau update fitur.' },
] as const;

export const platformOverviewTable: PreviewTableFixture = {
  columns: ['Hal yang perlu dicek', 'Area', 'Dampak', 'Status'],
  rows: platformAlerts.map((item) => [item.title, item.title.includes('tagihan') ? 'Tagihan' : item.title.includes('bantuan') ? 'Bantuan' : item.title.includes('webhook') ? 'Sistem' : 'Toko', item.description, 'Preview']),
  caption: 'Data transaksi dan GMV ditampilkan secara agregat. Akses detail toko membutuhkan alasan dan tercatat di aktivitas.',
  primaryColumn: 'Hal yang perlu dicek',
  statusColumn: 'Status',
  metaColumns: ['Area', 'Dampak'],
} as const;

export const platformStoresTable: PreviewTableFixture = {
  columns: ['Nama toko', 'Owner', 'Kontak', 'Status', 'Paket', 'Cabang', 'User', 'Last active', 'Tagihan', 'Aksi'],
  rows: [
    ['Kopi Senja Sample', 'Raka Owner', 'ra***@sample.id · 0812****221', 'Aktif', 'Pro', '3', '12', 'Hari ini', 'Lunas', 'Lihat ringkasan · Akses bantuan dikunci'],
    ['Laundry Kita Preview', 'Maya Owner', 'ma***@sample.id · 0857****119', 'Trial', 'Starter', '1', '4', 'Kemarin', 'Trial', 'Akses bantuan butuh alasan'],
    ['Retail Barat Sample', 'Dimas Owner', 'di***@sample.id · 0813****552', 'Expired', 'Business', '5', '25', '3 hari lalu', 'Belum dibayar', 'Nonaktifkan sementara disabled'],
    ['Warung Ibu Demo', 'Sari Owner', 'sa***@sample.id · 0821****778', 'Suspended', 'Free', '1', '2', '7 hari lalu', 'Grace period', 'Arsipkan disabled'],
    ['Bakery Utara Sample', 'Arman Owner', 'ar***@sample.id · 0878****664', 'Perlu dicek', 'Pro', '2', '9', 'Hari ini', 'Gagal', 'Verifikasi data preview'],
  ],
  caption: 'Akses bantuan membutuhkan alasan, durasi akses, dan akan tercatat di aktivitas saat backend aktif. Nonaktifkan dan arsipkan tidak bekerja di preview.',
  primaryColumn: 'Nama toko',
  statusColumn: 'Status',
  metaColumns: ['Owner', 'Kontak', 'Paket', 'Cabang', 'User', 'Last active', 'Tagihan', 'Aksi'],
} as const;

export const platformPlansTable: PreviewTableFixture = {
  columns: ['Paket', 'Status', 'Harga', 'Cabang', 'User', 'Produk', 'Fitur utama', 'Dipakai toko', 'Aksi'],
  rows: [
    ['Free', 'Draft', 'Belum final', '1', '2', 'Dasar', 'Kasir · Produk dasar', '18 contoh', 'Edit paket disabled'],
    ['Starter', 'Preview', 'Belum final', '1', '5', 'Menengah', 'Kasir · Produk & inventori · Laporan', '41 contoh', 'Duplikat disabled'],
    ['Pro', 'Published preview', 'Belum final', '3', '15', 'Lebih besar', 'Multi-cabang · Laporan · Tim & absensi', '52 contoh', 'Publish disabled'],
    ['Business', 'Preview', 'Belum final', 'Custom', 'Custom', 'Custom', 'Bantuan prioritas · Payment integration', '17 contoh', 'Edit paket disabled'],
  ],
  caption: 'Harga dan limit final belum aktif. Perubahan paket membutuhkan backend billing.',
  primaryColumn: 'Paket',
  statusColumn: 'Status',
  metaColumns: ['Harga', 'Cabang', 'User', 'Produk', 'Fitur utama', 'Dipakai toko', 'Aksi'],
} as const;

export const platformBillingRows: PreviewTableFixture = {
  columns: ['Toko', 'Paket', 'Jatuh tempo', 'Status pembayaran', 'Nominal', 'Update terakhir', 'Aksi'],
  rows: [
    ['Kopi Senja Sample', 'Pro', '01 Jul 2026', 'Lunas', 'Preview', '21 Jun 2026 09:10', 'Lihat tagihan disabled'],
    ['Laundry Kita Preview', 'Starter', '24 Jun 2026', 'Trial', 'Preview', 'Belum ada tagihan', 'Perpanjang manual gated'],
    ['Retail Barat Sample', 'Business', '01 Jun 2026', 'Gagal', 'Preview', '20 Jun 2026 18:22', 'Kirim ulang tagihan disabled'],
    ['Bakery Utara Sample', 'Pro', '28 Jun 2026', 'Belum dibayar', 'Preview', '20 Jun 2026 16:44', 'Tandai lunas gated'],
    ['Warung Ibu Demo', 'Free', '15 Jun 2026', 'Grace period', 'Preview', '18 Jun 2026 12:20', 'Manual review'],
  ],
  caption: 'Gateway payment belum terhubung. Data tagihan hanya preview UI. Tandai lunas, perpanjang manual, dan kirim ulang tagihan tidak bekerja.',
  primaryColumn: 'Toko',
  statusColumn: 'Status pembayaran',
  metaColumns: ['Paket', 'Jatuh tempo', 'Nominal', 'Update terakhir', 'Aksi'],
} as const;

export const platformSupportTickets: PreviewTableFixture = {
  columns: ['ID', 'Toko', 'Masalah', 'Kategori', 'Prioritas', 'Status', 'Ditangani oleh', 'Update terakhir'],
  rows: [
    ['BNT-120', 'Kopi Senja Sample', 'Printer struk tidak keluar', 'Printer', 'Urgent', 'Terbuka', 'Rizky Support', '21 Jun 2026 09:12'],
    ['BNT-119', 'Retail Barat Sample', 'Pembayaran QRIS belum masuk', 'QRIS/Payment', 'Tinggi', 'Menunggu cek', 'Fina Finance', '21 Jun 2026 08:40'],
    ['BNT-118', 'Laundry Kita Preview', 'Owner lupa akses akun', 'Akun', 'Sedang', 'Triage', 'Nadia Ops', '20 Jun 2026 17:01'],
    ['BNT-117', 'Bakery Utara Sample', 'Stok minus setelah koreksi', 'Stok', 'Sedang', 'Terbuka', 'Rizky Support', '20 Jun 2026 16:44'],
  ],
  caption: 'Catatan internal dan assignment belum tersimpan karena backend belum aktif. Minta akses bantuan tetap disabled/preview.',
  primaryColumn: 'ID',
  statusColumn: 'Prioritas',
  metaColumns: ['Toko', 'Masalah', 'Kategori', 'Status', 'Ditangani oleh', 'Update terakhir'],
} as const;

export const platformActivityLogs: PreviewTableFixture = {
  columns: ['Waktu', 'Pelaku', 'Peran', 'Aksi', 'Toko', 'IP/device', 'Level', 'Alasan'],
  rows: [
    ['21 Jun 2026 09:20', 'Rizky Support', 'support', 'Akses bantuan diminta', 'Kopi Senja Sample', '192.0.2.10 · Chrome', 'Tinggi', 'Alasan wajib saat backend aktif'],
    ['21 Jun 2026 08:44', 'Fina Finance', 'finance', 'Status pembayaran dicek', 'Retail Barat Sample', '198.51.100.8 · Desktop', 'Sedang', 'Tandai lunas disabled'],
    ['20 Jun 2026 17:12', 'Nadia Ops', 'owner', 'Perubahan paket preview', 'Laundry Kita Preview', '203.0.113.4 · Chrome', 'Sedang', 'Backend billing required'],
    ['20 Jun 2026 16:30', 'Sistem preview', 'system', 'Webhook payment gagal contoh', 'Retail Barat Sample', 'Gateway fixture', 'Perlu dicek', 'Retry belum aktif'],
  ],
  caption: 'Akses bantuan wajib tercatat sebelum fitur digunakan. Semua event memakai fixture lokal.',
  primaryColumn: 'Waktu',
  statusColumn: 'Level',
  metaColumns: ['Pelaku', 'Peran', 'Aksi', 'Toko', 'IP/device', 'Alasan'],
} as const;

export const platformSystemServices: PreviewTableFixture = {
  columns: ['Layanan', 'Status', 'Dampak', 'Update terakhir', 'Penanggung jawab'],
  rows: [
    ['API', 'Normal', 'Tampilan admin tetap bisa direview', 'Preview static', 'Platform'],
    ['Database', 'Normal', 'Belum terhubung data real', 'Preview static', 'Engineering'],
    ['Antrian', 'Perlu dicek', 'Billing job contoh belum berjalan', '21 Jun 2026 09:00', 'Ops'],
    ['Payment gateway', 'Perlu dicek', 'Gateway belum terhubung', '21 Jun 2026 08:40', 'Finance'],
    ['Webhook', 'Gangguan', '1 webhook payment contoh gagal', '21 Jun 2026 08:22', 'Finance'],
    ['Email/WhatsApp', 'Normal', 'Broadcast belum aktif', 'Preview static', 'Support'],
    ['Backup', 'Normal', 'Status backup hanya contoh', 'Preview static', 'Engineering'],
    ['Error', 'Perlu dicek', '1 error contoh untuk review UI', '21 Jun 2026 08:30', 'Ops'],
  ],
  caption: 'Status sistem hanya mock/fixture. Tidak ada polling, fetch, atau raw log produksi.',
  primaryColumn: 'Layanan',
  statusColumn: 'Status',
  metaColumns: ['Dampak', 'Update terakhir', 'Penanggung jawab'],
} as const;

export const platformAnnouncements: PreviewTableFixture = {
  columns: ['Judul', 'Target', 'Jenis', 'Channel', 'Status', 'Jadwal'],
  rows: [
    ['Maintenance dashboard preview', 'Semua toko', 'Maintenance', 'In-app · Email preview', 'Draft', 'Belum dijadwalkan'],
    ['Promo paket Pro preview', 'Trial', 'Promo', 'In-app preview', 'Terjadwal', '28 Jun 2026'],
    ['Update fitur multi-cabang', 'Paket tertentu', 'Update fitur', 'Email · WhatsApp preview', 'Terkirim sample', '18 Jun 2026'],
    ['Pengingat tagihan', 'Expired', 'Tagihan', 'In-app preview', 'Draft', 'Belum dijadwalkan'],
  ],
  caption: 'Pengumuman belum mengirim email, WhatsApp, atau notifikasi in-app nyata. Simpan draft belum aktif.',
  primaryColumn: 'Judul',
  statusColumn: 'Status',
  metaColumns: ['Target', 'Jenis', 'Channel', 'Jadwal'],
} as const;

export const platformUsersTable: PreviewTableFixture = {
  columns: ['Nama', 'Tipe', 'Toko', 'Peran', 'Last login', 'Scope/device', 'Status', 'Aksi'],
  rows: [
    ['Nadia Ops', 'Operator NojPOS', 'Platform', 'owner', 'Hari ini', 'Owner scope', 'Aktif', 'Reset password disabled'],
    ['Rizky Support', 'Operator NojPOS', 'Platform', 'support', 'Kemarin', 'Ringkasan toko saja', 'Aktif', 'Lihat security log'],
    ['Fina Finance', 'Operator NojPOS', 'Platform', 'finance', '2 hari lalu', 'Tagihan saja', 'Aktif', 'Ban disabled'],
    ['Raka Tenant Owner', 'User toko', 'Kopi Senja Sample', 'owner', 'Hari ini', 'Chrome sample', 'Aktif', 'Security log only'],
    ['Siti Cashier', 'User toko', 'Laundry Kita Preview', 'cashier', '3 hari lalu', 'Android POS sample', 'Perlu dicek', 'Reset gated'],
  ],
  caption: 'Akses user toko dibatasi dan membutuhkan alasan operasional. Reset password dan ban user hanya preview.',
  primaryColumn: 'Nama',
  statusColumn: 'Status',
  metaColumns: ['Tipe', 'Toko', 'Peran', 'Last login', 'Scope/device', 'Aksi'],
} as const;

export const platformHealthChecks = [
  { label: 'Mode preview', status: 'Aman', detail: 'Tidak ada API, auth real, storage, payment, atau broadcast.' },
  { label: 'Aksi sensitif', status: 'Dikunci', detail: 'Akses bantuan, suspend, mark paid, reset password, dan broadcast disabled.' },
  { label: 'Privasi toko', status: 'Dijaga', detail: 'Transaksi dan GMV tampil agregat, bukan detail lintas toko.' },
  { label: 'Backend', status: 'Belum terhubung', detail: 'Laravel API, database, gateway, dan announcement job belum aktif.' },
] as const;

export const platformBusinessDetailPreviews: readonly DetailPreviewFixture[] = [
  {
    id: 'kopi-senja-sample',
    title: 'Kopi Senja Sample',
    kicker: 'Ringkasan toko preview',
    badge: 'Aktif sample',
    description: 'Ringkasan toko untuk kebutuhan bantuan. Detail transaksi/GMV toko tidak dibuka bebas dan membutuhkan alasan serta catatan aktivitas saat backend aktif.',
    backHref: '/platform/businesses',
    backLabel: 'Kembali ke Toko',
    actions: [
      { label: 'Lihat ringkasan', status: 'preview-only' },
      { label: 'Akses bantuan', status: 'disabled · alasan wajib' },
      { label: 'Nonaktifkan sementara', status: 'disabled · approval required' },
      { label: 'Arsipkan', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Identitas toko',
        rows: [
          { label: 'Owner', value: 'Raka Owner', helper: 'Kontak owner tetap dimasking di daftar toko.' },
          { label: 'Paket', value: 'Pro preview', helper: 'Paket final mengikuti backend billing.' },
          { label: 'Status', value: 'Aktif sample', helper: 'Status sample untuk UI preview.' },
        ],
      },
      {
        title: 'Batas akses',
        rows: [
          { label: 'Akses bantuan', value: 'Belum aktif', helper: 'Membutuhkan alasan, durasi akses, dan aktivitas tercatat.' },
          { label: 'Aksi sensitif', value: 'Disabled', helper: 'Suspend/delete membutuhkan approval saat backend aktif.' },
          { label: 'Privacy', value: 'Agregat saja', helper: 'Data transaksi dan GMV tidak dibuka detail lintas toko.' },
        ],
      },
    ],
    timeline: ['Ringkasan toko preview dibuka.', 'Akses bantuan dikunci.', 'Tidak ada aksi real dijalankan.'],
  },
  {
    id: 'retail-barat-sample',
    title: 'Retail Barat Sample',
    kicker: 'Tagihan perlu dicek',
    badge: 'Past due sample',
    description: 'Preview risiko tagihan toko dengan privacy copy dan action gate.',
    backHref: '/platform/businesses',
    backLabel: 'Kembali ke Toko',
    actions: [
      { label: 'Tandai lunas', status: 'disabled · backend billing required' },
      { label: 'Aktifkan toko', status: 'disabled · approval required' },
    ],
    sections: [
      {
        title: 'Ringkasan tagihan',
        rows: [
          { label: 'Status pembayaran', value: 'Gagal sample', helper: 'Status belum terhubung gateway.' },
          { label: 'Tagihan terakhir', value: 'INV-SAMPLE-098', helper: 'Invoice sample tidak bisa ditandai lunas.' },
          { label: 'Risiko', value: 'Manual review', helper: 'Risk queue hanya fixture lokal.' },
        ],
      },
      {
        title: 'Kebutuhan aktivitas',
        rows: [
          { label: 'Alasan', value: 'Wajib nanti', helper: 'Akses detail toko membutuhkan alasan operasional.' },
          { label: 'Before/after', value: 'Wajib nanti', helper: 'Perubahan status wajib masuk aktivitas.' },
          { label: 'Tenant isolation', value: 'Dijaga oleh desain', helper: 'Platform tidak memakai konteks outlet admin toko.' },
        ],
      },
    ],
    timeline: ['Pembayaran gagal contoh terdeteksi.', 'Tandai lunas tetap disabled.', 'Aktivitas wajib sebelum aksi aktif.'],
  },
] as const;

export const platformBusinessDetailLinks = platformBusinessDetailPreviews.map((detail) => ({ href: `/platform/businesses/${detail.id}`, label: detail.title, description: detail.description })) as readonly { href: string; label: string; description: string }[];

export function getPlatformBusinessDetailPreview(id: string) {
  return platformBusinessDetailPreviews.find((detail) => detail.id === id);
}

// Backward-compatible aliases for older platform preview pages/tests.
export const platformTenantHealthMetrics = platformOverviewMetrics.slice(0, 4);
export const platformRevenueBillingMetrics = platformOverviewMetrics.slice(4, 7);
export const platformMetrics = platformOverviewMetrics.slice(0, 4);
export const platformRevenueMetrics = platformOverviewMetrics.slice(4, 8);
export const platformHealthMetrics: readonly PreviewMetricFixture[] = [
  { label: 'API', value: 'Normal', description: 'Status contoh. Tidak ada health check real.', badge: 'Preview' },
  { label: 'Database', value: 'Normal', description: 'Belum terhubung data produksi.', badge: 'Mock' },
  { label: 'Webhook', value: 'Gangguan', description: '1 contoh webhook perlu dicek.', badge: 'Perlu dicek', tone: 'warning' },
  { label: 'Backup', value: 'Normal', description: 'Status backup hanya fixture.', badge: 'Data contoh' },
] as const;
export const platformRiskQueue = platformAlerts;
export const platformBusinessesTable = platformStoresTable;
export const platformRevenueTable = platformBillingRows;
export const platformGatewayLogsTable: PreviewTableFixture = {
  columns: ['Gateway', 'Reference', 'Webhook', 'Status'],
  rows: [['Payment preview', 'ref_sample_001', 'Gagal contoh', 'Retry disabled']],
  caption: 'Gateway logs hanya preview dan tidak melakukan retry.',
  primaryColumn: 'Gateway',
  statusColumn: 'Status',
  metaColumns: ['Reference', 'Webhook'],
};
export const platformRevenueCards = [
  { title: 'MRR', value: 'Rp18,4 jt', description: 'Contoh agregat bulanan, belum billing real.' },
  { title: 'Tagihan belum dibayar', value: '7', description: 'Perlu dicek finance.' },
  { title: 'Pembayaran gagal', value: '3', description: 'Retry tetap disabled.' },
  { title: 'Trial segera habis', value: '4', description: 'Follow-up preview.' },
] as const;
export const platformSubscriptionsTable = platformBillingRows;
export const platformAuditTable = platformActivityLogs;
export const platformSystemHealthTable = platformSystemServices;
export const platformIncidentsTable = platformSystemServices;
export const platformSupportTable = platformSupportTickets;
export const platformAnnouncementsTable = platformAnnouncements;
