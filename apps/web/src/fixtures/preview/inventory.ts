import type { DetailPreviewFixture, PreviewMetricFixture, ReadonlyResourceFixture } from './types';

export const inventoryMetrics: readonly PreviewMetricFixture[] = [
  { label: 'Total item', value: '—', description: 'Jumlah produk dan stok akan berasal dari inventory backend.' },
  { label: 'Low stock', value: '—', badge: 'No browser calc', tone: 'warning', description: 'Threshold stok tidak dihitung di browser preview.' },
  { label: 'Transfer', value: 'Read-only', badge: 'Preview', description: 'Transfer hanya placeholder layout sampai contract tersedia.' },
] as const;

export const inventoryResources: readonly ReadonlyResourceFixture[] = [
  { label: 'Stok saat ini', path: '/inventory' },
  { label: 'Pergerakan stok', path: '/inventory/movements' },
  { label: 'Transfer dalam perjalanan', path: '/inventory/transfers/in-transit' },
] as const;

export const inventoryDetailPreviews: readonly DetailPreviewFixture[] = [
  {
    id: 'stock-current',
    title: 'Stok saat ini',
    kicker: 'Inventory detail preview',
    badge: 'No browser calc',
    description: 'Detail stok disiapkan untuk layout observability. Jumlah stok, threshold, dan movement tetap harus datang dari server.',
    backHref: '/inventory',
    backLabel: 'Kembali ke inventaris',
    actions: [
      { label: 'Adjust stok', status: 'blocked' },
      { label: 'Export CSV', status: 'backend gated' },
    ],
    sections: [
      {
        title: 'Stock snapshot',
        rows: [
          { label: 'Outlet', value: 'Tenant scope', helper: 'Outlet scope real belum dibaca.' },
          { label: 'Jumlah item', value: '—', helper: 'Tidak ada kalkulasi stok di browser.' },
          { label: 'Low stock', value: '—', helper: 'Threshold berasal dari backend.' },
        ],
      },
      {
        title: 'Guard rails',
        rows: [
          { label: 'Mutation', value: 'Disabled', helper: 'Stock mutation tidak dibuka di admin preview.' },
          { label: 'Transfer', value: 'Preview only', helper: 'Transfer workflow menunggu endpoint.' },
          { label: 'Audit', value: 'Server-owned', helper: 'Semua perubahan stok harus diaudit backend.' },
        ],
      },
    ],
    timeline: [
      'Inventory detail route tersedia untuk layout review.',
      'Stock adjustment tetap diblokir sampai contract siap.',
      'Audit dan movement harus diverifikasi oleh backend.',
    ],
  },
  {
    id: 'movement-log',
    title: 'Pergerakan stok',
    kicker: 'Movement detail preview',
    badge: 'Backend gated',
    description: 'Movement log dipakai sebagai placeholder struktur timeline, bukan data mutasi produksi.',
    backHref: '/inventory',
    backLabel: 'Kembali ke inventaris',
    actions: [
      { label: 'Filter movement', status: 'disabled' },
      { label: 'Download', status: 'backend gated' },
    ],
    sections: [
      {
        title: 'Movement scope',
        rows: [
          { label: 'Periode', value: '—', helper: 'Filter tanggal belum aktif.' },
          { label: 'Produk', value: '—', helper: 'Produk real menunggu API.' },
          { label: 'Aktor', value: 'Backend actor', helper: 'Actor harus dari audit backend.' },
        ],
      },
      {
        title: 'State coverage',
        rows: [
          { label: 'Loading', value: 'Ready', helper: 'State sudah tersedia di matrix.' },
          { label: 'Empty', value: 'Ready', helper: 'Empty log sudah disiapkan.' },
          { label: 'Error', value: 'Ready', helper: 'Error log tidak menampilkan data palsu.' },
        ],
      },
    ],
    timeline: [
      'Movement route dipakai untuk review density timeline.',
      'Filter dan download tetap disabled.',
      'Server contract menentukan pagination dan ordering.',
    ],
  },
] as const;

export const inventoryDetailLinks = inventoryDetailPreviews.map((detail) => ({
  href: `/inventory/${detail.id}`,
  label: detail.title,
  description: detail.description,
})) as readonly { href: string; label: string; description: string }[];

export function getInventoryDetailPreview(id: string) {
  return inventoryDetailPreviews.find((detail) => detail.id === id);
}
