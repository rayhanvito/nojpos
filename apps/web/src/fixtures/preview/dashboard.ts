export type DashboardTone = 'neutral' | 'success' | 'warning' | 'danger' | 'info';

export type DashboardKpiCard = {
  readonly label: string;
  readonly value: string;
  readonly helper: string;
  readonly trend: string;
  readonly tone: DashboardTone;
};

export type DashboardAlert = {
  readonly title: string;
  readonly description: string;
  readonly tone: DashboardTone;
};

export type DashboardSalesPoint = {
  readonly day: string;
  readonly sales: number;
  readonly label: string;
};

export type DashboardPaymentMethod = {
  readonly method: string;
  readonly total: number;
  readonly label: string;
};

export type DashboardTopProduct = {
  readonly name: string;
  readonly quantity: string;
  readonly total: string;
  readonly share: number;
};

export type DashboardLowStockItem = {
  readonly name: string;
  readonly stock: string;
  readonly status: 'Hampir habis' | 'Habis';
};

export type DashboardRecentTransaction = {
  readonly code: string;
  readonly time: string;
  readonly method: string;
  readonly total: string;
};

export type DashboardCashierPerformance = {
  readonly name: string;
  readonly transactions: string;
  readonly total: string;
  readonly note: string;
};

export type DashboardBranchHighlight = {
  readonly name: string;
  readonly status: string;
  readonly summary: string;
  readonly tone: DashboardTone;
};

export const dashboardKpiCards: readonly DashboardKpiCard[] = [
  {
    label: 'Penjualan hari ini',
    value: 'Rp2.450.000',
    helper: 'Contoh total penjualan dari transaksi hari ini.',
    trend: '+8% vs kemarin',
    tone: 'success',
  },
  {
    label: 'Jumlah transaksi',
    value: '86 transaksi',
    helper: 'Transaksi yang tercatat pada hari ini.',
    trend: 'Ramai stabil',
    tone: 'info',
  },
  {
    label: 'Rata-rata transaksi',
    value: 'Rp28.500',
    helper: 'Nilai rata-rata dari transaksi hari ini.',
    trend: 'Contoh preview',
    tone: 'neutral',
  },
  {
    label: 'Gross profit',
    value: 'Rp740.000',
    helper: 'Estimasi laba kotor dari data contoh, belum akurat produksi.',
    trend: 'Estimasi UI',
    tone: 'success',
  },
  {
    label: 'Stok hampir habis',
    value: '12 item',
    helper: 'Produk yang perlu dicek ulang sebelum jam ramai.',
    trend: 'Butuh cek',
    tone: 'warning',
  },
  {
    label: 'Selisih kas',
    value: 'Rp45.000',
    helper: 'Perlu dikonfirmasi saat tutup shift.',
    trend: 'Preview kas',
    tone: 'warning',
  },
] as const;

export const dashboardAlerts: readonly DashboardAlert[] = [
  {
    title: '12 produk stok menipis',
    description: 'Cek stok sebelum jam ramai agar kasir tidak menjual item yang kosong.',
    tone: 'warning',
  },
  {
    title: '3 produk habis',
    description: 'Tandai sementara sebagai tidak tersedia setelah backend inventori aktif.',
    tone: 'danger',
  },
  {
    title: '1 shift belum ditutup',
    description: 'Shift sore belum ditutup. Konfirmasi kasir saat tutup toko.',
    tone: 'info',
  },
  {
    title: 'Selisih kas Rp45.000',
    description: 'Selisih kas perlu dikonfirmasi saat tutup shift, bukan angka final.',
    tone: 'warning',
  },
  {
    title: 'Subscription habis 5 hari lagi',
    description: 'Siapkan reminder pembayaran agar operasional tidak terganggu.',
    tone: 'neutral',
  },
] as const;

export const dashboardSalesLast7Days: readonly DashboardSalesPoint[] = [
  { day: 'Sen', sales: 1850000, label: 'Rp1,85 jt' },
  { day: 'Sel', sales: 2120000, label: 'Rp2,12 jt' },
  { day: 'Rab', sales: 1980000, label: 'Rp1,98 jt' },
  { day: 'Kam', sales: 2310000, label: 'Rp2,31 jt' },
  { day: 'Jum', sales: 2450000, label: 'Rp2,45 jt' },
  { day: 'Sab', sales: 2680000, label: 'Rp2,68 jt' },
  { day: 'Min', sales: 2450000, label: 'Rp2,45 jt' },
] as const;

export const dashboardPaymentMethods: readonly DashboardPaymentMethod[] = [
  { method: 'QRIS', total: 1180000, label: 'Rp1,18 jt' },
  { method: 'Tunai', total: 820000, label: 'Rp820 rb' },
  { method: 'Transfer', total: 310000, label: 'Rp310 rb' },
  { method: 'Kartu', total: 140000, label: 'Rp140 rb' },
] as const;

export const dashboardTopProducts: readonly DashboardTopProduct[] = [
  { name: 'Es Kopi Susu Gula Aren', quantity: '34 terjual', total: 'Rp612.000', share: 86 },
  { name: 'Nasi Ayam Sambal Matah', quantity: '21 terjual', total: 'Rp525.000', share: 74 },
  { name: 'Americano Ice', quantity: '18 terjual', total: 'Rp306.000', share: 56 },
  { name: 'Roti Bakar Cokelat', quantity: '15 terjual', total: 'Rp225.000', share: 44 },
  { name: 'Teh Lemon', quantity: '12 terjual', total: 'Rp144.000', share: 35 },
] as const;

export const dashboardLowStockItems: readonly DashboardLowStockItem[] = [
  { name: 'Susu UHT 1L', stock: '4 pcs', status: 'Hampir habis' },
  { name: 'Cup 16 oz', stock: '22 pcs', status: 'Hampir habis' },
  { name: 'Sirup gula aren', stock: '1 botol', status: 'Hampir habis' },
  { name: 'Sedotan hitam', stock: '0 pack', status: 'Habis' },
  { name: 'Tisu meja', stock: '0 pack', status: 'Habis' },
] as const;

export const dashboardRecentTransactions: readonly DashboardRecentTransaction[] = [
  { code: 'TRX-1028', time: '15:42', method: 'QRIS', total: 'Rp38.000' },
  { code: 'TRX-1027', time: '15:36', method: 'Tunai', total: 'Rp52.000' },
  { code: 'TRX-1026', time: '15:22', method: 'QRIS', total: 'Rp26.000' },
  { code: 'TRX-1025', time: '15:10', method: 'Transfer', total: 'Rp88.000' },
  { code: 'TRX-1024', time: '14:58', method: 'Tunai', total: 'Rp19.000' },
] as const;

export const dashboardCashierPerformance: readonly DashboardCashierPerformance[] = [
  { name: 'Ayu', transactions: '31 trx', total: 'Rp920.000', note: 'Shift pagi rapi' },
  { name: 'Dimas', transactions: '28 trx', total: 'Rp780.000', note: 'Cek kas tutup shift' },
  { name: 'Raka', transactions: '18 trx', total: 'Rp515.000', note: 'Banyak transaksi QRIS' },
  { name: 'Mira', transactions: '9 trx', total: 'Rp235.000', note: 'Baru mulai shift' },
] as const;

export const dashboardBranchHighlights: readonly DashboardBranchHighlight[] = [
  {
    name: 'Cabang Utama',
    status: 'Ramai stabil',
    summary: 'Penjualan contoh tertinggi, stok cup perlu dicek.',
    tone: 'success',
  },
  {
    name: 'Booth Kampus',
    status: 'Perlu stok',
    summary: 'Beberapa bahan hampir habis sebelum jam pulang.',
    tone: 'warning',
  },
  {
    name: 'Outlet Laundry',
    status: 'Shift belum tutup',
    summary: 'Perlu konfirmasi kasir saat tutup operasional.',
    tone: 'info',
  },
] as const;

export const dashboardQuickLinks = [
  { label: 'Lihat Transaksi', path: '/transactions' },
  { label: 'Cek Inventori', path: '/inventory' },
  { label: 'Kelola Produk', path: '/catalog' },
  { label: 'Lihat Laporan', path: '/reports' },
  { label: 'Kelola Cabang', path: '/outlets' },
  { label: 'Cek Karyawan', path: '/staff' },
] as const;
