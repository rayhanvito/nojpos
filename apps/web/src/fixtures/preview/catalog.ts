import type { DetailPreviewFixture, PreviewLane, PreviewTableFixture } from './types';

export const catalogLanes: readonly PreviewLane[] = [
  { label: 'Products', items: ['Compact table', 'Category badge', 'Price placeholder'] },
  { label: 'Variants', items: ['Modifier slot', 'Availability state', 'Image asset later'] },
  { label: 'Publish', items: ['Server validation later', 'Outlet scope later', 'Read-only preview'] },
] as const;

export const catalogProductsTable: PreviewTableFixture = {
  columns: ['Produk', 'Kategori', 'Harga', 'Ketersediaan', 'Status'],
  rows: [
    ['Kopi Susu', 'Minuman', 'Rp —', 'Tersedia', 'Preview'],
    ['Roti Bakar', 'Makanan', 'Rp —', 'Tersedia', 'Preview'],
    ['Teh Lemon', 'Minuman', 'Rp —', 'Nonaktif', 'Read-only'],
    ['Bundling promo', 'Menunggu contract', 'Rp —', 'Belum tersedia', 'Backend gated'],
  ],
} as const;

export const catalogDetailPreviews: readonly DetailPreviewFixture[] = [
  {
    id: 'kopi-susu',
    title: 'Kopi Susu',
    kicker: 'Product detail preview',
    badge: 'Read-only',
    description: 'Detail produk mengikuti gaya premium template, tetapi harga, stok, outlet availability, dan publish status belum ditarik dari backend.',
    backHref: '/catalog',
    backLabel: 'Kembali ke katalog',
    actions: [
      { label: 'Edit produk', status: 'disabled' },
      { label: 'Publish', status: 'backend gated' },
    ],
    sections: [
      {
        title: 'Identitas produk',
        rows: [
          { label: 'Nama', value: 'Kopi Susu', helper: 'Nama hanya fixture preview.' },
          { label: 'Kategori', value: 'Minuman', helper: 'Kategori final berasal dari catalog backend.' },
          { label: 'SKU', value: '—', helper: 'SKU belum dikontrak untuk admin web.' },
        ],
      },
      {
        title: 'Commercial guard',
        rows: [
          { label: 'Harga', value: 'Rp —', helper: 'Tidak ada kalkulasi atau harga produksi di browser.' },
          { label: 'Ketersediaan', value: 'Preview', helper: 'Outlet visibility akan dikontrol server.' },
          { label: 'Modifier', value: 'Backend gated', helper: 'Modifier dan variant belum aktif.' },
        ],
      },
    ],
    timeline: [
      'Draft UI detail siap sebagai route static preview.',
      'Validation, price policy, dan outlet scope menunggu contract backend.',
      'Audit publish akan dibuat ketika API disetujui.',
    ],
  },
  {
    id: 'roti-bakar',
    title: 'Roti Bakar',
    kicker: 'Product detail preview',
    badge: 'Preview',
    description: 'Contoh detail makanan untuk menguji density card, section, dan disabled action tanpa CRUD aktif.',
    backHref: '/catalog',
    backLabel: 'Kembali ke katalog',
    actions: [
      { label: 'Edit produk', status: 'disabled' },
      { label: 'Arsipkan', status: 'backend gated' },
    ],
    sections: [
      {
        title: 'Identitas produk',
        rows: [
          { label: 'Nama', value: 'Roti Bakar', helper: 'Fixture untuk layout saja.' },
          { label: 'Kategori', value: 'Makanan', helper: 'Data kategori real belum dibaca.' },
          { label: 'Status', value: 'Preview', helper: 'Status final dari backend.' },
        ],
      },
      {
        title: 'Commercial guard',
        rows: [
          { label: 'Harga', value: 'Rp —', helper: 'Nilai uang tidak diisi sebelum API.' },
          { label: 'Tax/service', value: '—', helper: 'Tidak dihitung di browser.' },
          { label: 'Bundle', value: 'Tidak aktif', helper: 'Bundling menunggu contract promotion.' },
        ],
      },
    ],
    timeline: [
      'Detail route tersedia untuk validasi layout.',
      'Edit dan archive tetap disabled.',
      'Server contract dibutuhkan sebelum data produksi tampil.',
    ],
  },
] as const;

export const catalogDetailLinks = catalogDetailPreviews.map((detail) => ({
  href: `/catalog/${detail.id}`,
  label: detail.title,
  description: detail.description,
})) as readonly { href: string; label: string; description: string }[];

export function getCatalogDetailPreview(id: string) {
  return catalogDetailPreviews.find((detail) => detail.id === id);
}
