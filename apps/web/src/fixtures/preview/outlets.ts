import type { DetailPreviewFixture, PreviewLane, PreviewMetricFixture, PreviewTableFixture } from './types';

export const outletMetrics: readonly PreviewMetricFixture[] = [
  { label: 'Outlet registry', value: '—', description: 'Jumlah outlet berasal dari tenant API nanti.', badge: 'Preview' },
  { label: 'Open/close', value: 'Read-only', description: 'Aksi buka/tutup outlet disabled.', badge: 'Gated', tone: 'warning' },
  { label: 'Terminals', value: '—', description: 'Lock/unlock terminal menunggu backend.', badge: 'Disabled' },
] as const;

export const outletLanes: readonly PreviewLane[] = [
  { label: 'Directory', items: ['Outlet cards', 'Operational status', 'Terminal summary'] },
  { label: 'Detail', items: ['Address preview', 'Payment rails', 'Terminal lock gate'] },
  { label: 'Controls', items: ['Open disabled', 'Close disabled', 'Terminal action disabled'] },
] as const;

export const outletsTable: PreviewTableFixture = {
  columns: ['Outlet', 'Lokasi', 'Terminal', 'Status'],
  rows: [
    ['Outlet preview pusat', 'Alamat dari backend', 'Terminal preview', 'Preview'],
    ['Outlet satelit', '—', 'Lock/unlock gated', 'Read-only'],
    ['Outlet onboarding', 'Menunggu konfigurasi', '—', 'Backend gated'],
  ],
  caption: 'Outlet list hanya untuk preview. Open/close outlet dan terminal action tetap disabled.',
} as const;

export const outletDetailPreviews: readonly DetailPreviewFixture[] = [
  {
    id: 'outlet-preview-pusat',
    title: 'Outlet preview pusat',
    kicker: 'Outlet detail',
    badge: 'Read-only',
    description: 'Detail outlet menampilkan struktur profil, payment setup, dan terminal summary tanpa mengubah status operasional.',
    backHref: '/outlets',
    backLabel: 'Kembali ke outlets',
    actions: [
      { label: 'Open outlet', status: 'disabled' },
      { label: 'Close outlet', status: 'disabled' },
      { label: 'Lock terminal', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Outlet profile',
        rows: [
          { label: 'Outlet ID', value: '—', helper: 'ID real berasal dari backend.' },
          { label: 'Address', value: 'Alamat dari backend', helper: 'Alamat fixture tidak dipakai sebagai data final.' },
          { label: 'Timezone', value: 'Tenant policy', helper: 'Timezone final dari backend.' },
        ],
      },
      {
        title: 'Operational gate',
        rows: [
          { label: 'Open status', value: 'Read-only', helper: 'Buka/tutup outlet butuh endpoint dan audit.' },
          { label: 'Terminal lock', value: 'Disabled', helper: 'Lock/unlock terminal tidak aktif.' },
          { label: 'Payment rails', value: 'Preview only', helper: 'Konfigurasi pembayaran dari backend.' },
        ],
      },
    ],
    timeline: ['Outlet detail route tersedia.', 'Open/close outlet disabled.', 'Terminal action disabled.'],
  },
  {
    id: 'outlet-satelit',
    title: 'Outlet satelit',
    kicker: 'Outlet terminal preview',
    badge: 'Backend gated',
    description: 'Preview outlet satelit untuk melihat penempatan terminal dan permission state.',
    backHref: '/outlets',
    backLabel: 'Kembali ke outlets',
    actions: [
      { label: 'Edit outlet', status: 'disabled' },
      { label: 'Unlock terminal', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Terminal summary',
        rows: [
          { label: 'Terminal', value: 'Terminal preview', helper: 'Device registry belum dihubungkan.' },
          { label: 'Lock state', value: 'Backend gated', helper: 'Server menentukan status terminal.' },
          { label: 'Last sync', value: '—', helper: 'Tidak ada device polling di UI-first.' },
        ],
      },
      {
        title: 'Access',
        rows: [
          { label: 'Manager scope', value: 'Backend policy', helper: 'Outlet permission final dari backend.' },
          { label: 'Edit', value: 'Disabled', helper: 'Butuh save contract.' },
          { label: 'Audit', value: 'Required', helper: 'Perubahan outlet harus tercatat.' },
        ],
      },
    ],
    timeline: ['Terminal layout tersedia.', 'Lock/unlock tidak aktif.', 'Tidak ada polling device.'],
  },
] as const;

export const outletDetailLinks = outletDetailPreviews.map((detail) => ({ href: `/outlets/${detail.id}`, label: detail.title, description: detail.description })) as readonly { href: string; label: string; description: string }[];

export function getOutletDetailPreview(id: string) {
  return outletDetailPreviews.find((detail) => detail.id === id);
}
