import type { DetailPreviewFixture, PreviewLane, PreviewMetricFixture, PreviewTableFixture } from './types';

export const customerMetrics: readonly PreviewMetricFixture[] = [
  { label: 'Customer directory', value: '—', description: 'Jumlah customer berasal dari backend saat integrasi.', badge: 'Preview' },
  { label: 'Segments', value: 'Read-only', description: 'Segment hanya label UI, bukan rule marketing client.', badge: 'Gated', tone: 'warning' },
  { label: 'Consent', value: '—', description: 'Preferensi kontak wajib dari contract backend.', badge: 'Policy' },
] as const;

export const customerLanes: readonly PreviewLane[] = [
  { label: 'Directory', items: ['Search placement', 'Segment chips', 'Consent display'] },
  { label: 'Detail', items: ['Profile facts', 'Activity placeholder', 'Action gate'] },
  { label: 'Governance', items: ['Import disabled', 'Export disabled', 'Privacy policy from server'] },
] as const;

export const customersTable: PreviewTableFixture = {
  columns: ['Customer', 'Segment', 'Kontak', 'Status'],
  rows: [
    ['Guest preview', 'Walk-in', '—', 'Preview'],
    ['Member placeholder', 'Loyalty gated', 'Masked by backend', 'Read-only'],
    ['Corporate contact', 'B2B preview', '—', 'Backend gated'],
  ],
  caption: 'Customer data hanya fixture UI. Import/export dan edit customer tetap disabled.',
  primaryColumn: 'Customer',
  statusColumn: 'Status',
  metaColumns: ['Segment', 'Kontak'],
} as const;

export const customerDetailPreviews: readonly DetailPreviewFixture[] = [
  {
    id: 'guest-preview',
    title: 'Guest preview',
    kicker: 'Customer detail',
    badge: 'Read-only',
    description: 'Profil customer preview untuk menyetujui layout identity, consent, dan activity tanpa menyimpan data personal di browser.',
    backHref: '/customers',
    backLabel: 'Kembali ke customers',
    actions: [
      { label: 'Edit customer', status: 'backend gated' },
      { label: 'Export profile', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Identity',
        rows: [
          { label: 'Customer ID', value: '—', helper: 'ID real berasal dari backend.' },
          { label: 'Display name', value: 'Guest preview', helper: 'Nama hanya placeholder review UI.' },
          { label: 'Contact', value: 'Masked', helper: 'Kontak real tidak disimpan di state browser.' },
        ],
      },
      {
        title: 'Consent & activity',
        rows: [
          { label: 'Marketing consent', value: 'Backend policy', helper: 'Consent wajib server-authoritative.' },
          { label: 'Last visit', value: '—', helper: 'Riwayat transaksi belum dihubungkan.' },
          { label: 'Notes', value: 'Read-only', helper: 'Catatan customer tidak bisa diedit di preview.' },
        ],
      },
    ],
    timeline: ['Customer detail route tersedia.', 'Edit/import/export customer disabled.', 'Backend menentukan visibility dan privacy.'],
  },
  {
    id: 'member-placeholder',
    title: 'Member placeholder',
    kicker: 'Customer loyalty preview',
    badge: 'Backend gated',
    description: 'Contoh detail member untuk posisi loyalty dan consent tanpa menghitung poin atau reward di frontend.',
    backHref: '/customers',
    backLabel: 'Kembali ke customers',
    actions: [
      { label: 'Change segment', status: 'disabled' },
      { label: 'Export activity', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Profile',
        rows: [
          { label: 'Segment', value: 'Loyalty gated', helper: 'Segment final dari backend.' },
          { label: 'Points', value: '—', helper: 'Tidak menghitung loyalty di browser.' },
          { label: 'Tier', value: '—', helper: 'Tier menunggu contract.' },
        ],
      },
      {
        title: 'Restrictions',
        rows: [
          { label: 'Edit', value: 'Disabled', helper: 'Perlu update endpoint dan audit.' },
          { label: 'Delete', value: 'Disabled', helper: 'Perlu policy retention.' },
          { label: 'Export', value: 'Disabled', helper: 'Perlu permission dan masking policy.' },
        ],
      },
    ],
    timeline: ['Loyalty placement disiapkan.', 'Tidak ada kalkulasi reward.', 'Privacy gate tetap terlihat.'],
  },
] as const;

export const customerDetailLinks = customerDetailPreviews.map((detail) => ({ href: `/customers/${detail.id}`, label: detail.title, description: detail.description })) as readonly { href: string; label: string; description: string }[];

export function getCustomerDetailPreview(id: string) {
  return customerDetailPreviews.find((detail) => detail.id === id);
}
