import type { DetailPreviewFixture, PreviewLane, PreviewTableFixture } from './types';

export const promotionLanes: readonly PreviewLane[] = [
  { label: 'Campaign', items: ['Name slot', 'Period slot', 'Status badge'] },
  { label: 'Rules', items: ['Rule copy only', 'No client totals', 'Server contract later'] },
  { label: 'Publish', items: ['Approval slot', 'Outlet scope later', 'Audit slot later'] },
] as const;

export const promotionTable: PreviewTableFixture = {
  columns: ['Nama', 'Rule', 'Periode', 'Status'],
  rows: [
    ['Campaign preview', 'Menunggu backend', '—', 'Backend gated'],
    ['Voucher outlet', 'Menunggu contract', '—', 'Belum tersedia'],
    ['Bundling', 'Tidak dihitung di browser', '—', 'Read-only'],
  ],
} as const;

export const promotionDetailPreviews: readonly DetailPreviewFixture[] = [
  {
    id: 'campaign-preview',
    title: 'Campaign preview',
    kicker: 'Promotion detail preview',
    badge: 'Backend gated',
    description: 'Detail promo dipakai untuk review rule layout saja. Discount, voucher, stacking, dan eligibility tidak dihitung di browser.',
    backHref: '/promotions',
    backLabel: 'Kembali ke promo',
    actions: [
      { label: 'Publish campaign', status: 'disabled' },
      { label: 'Duplicate', status: 'backend gated' },
    ],
    sections: [
      {
        title: 'Campaign identity',
        rows: [
          { label: 'Nama', value: 'Campaign preview', helper: 'Nama fixture untuk layout.' },
          { label: 'Periode', value: '—', helper: 'Tanggal tidak aktif sebelum contract.' },
          { label: 'Outlet', value: 'Tenant scope', helper: 'Scope final berasal dari backend.' },
        ],
      },
      {
        title: 'Rule guard',
        rows: [
          { label: 'Rule', value: 'Backend gated', helper: 'Rule tidak dijalankan di browser.' },
          { label: 'Discount', value: '—', helper: 'Tidak ada kalkulasi promosi client-side.' },
          { label: 'Eligibility', value: 'Server-owned', helper: 'Eligibility harus server authoritative.' },
        ],
      },
    ],
    timeline: [
      'Detail promo tersedia untuk desain review.',
      'Publish dan duplicate tetap disabled.',
      'Rule engine wajib berasal dari backend.',
    ],
  },
  {
    id: 'voucher-outlet',
    title: 'Voucher outlet',
    kicker: 'Voucher detail preview',
    badge: 'Preview',
    description: 'Voucher layout disiapkan untuk memastikan section, action, dan audit rail siap tanpa logic voucher aktif.',
    backHref: '/promotions',
    backLabel: 'Kembali ke promo',
    actions: [
      { label: 'Generate voucher', status: 'disabled' },
      { label: 'Deactivate', status: 'backend gated' },
    ],
    sections: [
      {
        title: 'Voucher identity',
        rows: [
          { label: 'Nama', value: 'Voucher outlet', helper: 'Fixture preview.' },
          { label: 'Kode', value: '—', helper: 'Kode voucher tidak dibuat di browser.' },
          { label: 'Kuota', value: '—', helper: 'Kuota menunggu contract.' },
        ],
      },
      {
        title: 'Rule guard',
        rows: [
          { label: 'Limit penggunaan', value: '—', helper: 'Limit final dari server.' },
          { label: 'Stacking', value: 'Disabled', helper: 'Stacking rule belum tersedia.' },
          { label: 'Audit', value: 'Required later', helper: 'Publish/deactivate harus diaudit.' },
        ],
      },
    ],
    timeline: [
      'Voucher detail route siap sebagai placeholder.',
      'Code generation tidak tersedia di preview.',
      'Backend contract menentukan status dan kuota.',
    ],
  },
] as const;

export const promotionDetailLinks = promotionDetailPreviews.map((detail) => ({
  href: `/promotions/${detail.id}`,
  label: detail.title,
  description: detail.description,
})) as readonly { href: string; label: string; description: string }[];

export function getPromotionDetailPreview(id: string) {
  return promotionDetailPreviews.find((detail) => detail.id === id);
}
