import type { DetailPreviewFixture, PreviewLane, PreviewTableFixture } from './types';

export const auditLanes: readonly PreviewLane[] = [
  { label: 'Capture', items: ['Actor from server', 'Time from server', 'Outlet scope'] },
  { label: 'Review', items: ['Filter placement', 'Detail drawer later', 'Error state ready'] },
  { label: 'Governance', items: ['Policy from backend', 'Readonly timeline', 'Export gated'] },
] as const;

export const auditTrailTable: PreviewTableFixture = {
  columns: ['Waktu', 'Aktor', 'Aksi', 'Status'],
  rows: [
    ['—', 'Backend actor', 'Akan tampil setelah contract audit tersedia', 'Preview'],
    ['—', 'Policy backend', 'Forbidden dan error state sudah disiapkan', 'Read-only'],
  ],
} as const;

export const auditDetailPreviews: readonly DetailPreviewFixture[] = [
  {
    id: 'policy-preview',
    title: 'Policy backend',
    kicker: 'Audit detail preview',
    badge: 'Read-only',
    description: 'Detail audit disiapkan untuk review governance view. Actor, timestamp, IP, outlet, dan payload harus berasal dari backend.',
    backHref: '/audit',
    backLabel: 'Kembali ke audit',
    actions: [
      { label: 'Export log', status: 'backend gated' },
      { label: 'Mark reviewed', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Event identity',
        rows: [
          { label: 'Event ID', value: '—', helper: 'ID audit belum dibaca dari server.' },
          { label: 'Actor', value: 'Backend actor', helper: 'Actor tidak disimulasikan di browser.' },
          { label: 'Waktu', value: '—', helper: 'Timestamp final dari server.' },
        ],
      },
      {
        title: 'Governance guard',
        rows: [
          { label: 'Payload', value: 'Hidden', helper: 'Payload real tidak dipalsukan.' },
          { label: 'Review status', value: 'Preview only', helper: 'Review action tetap disabled.' },
          { label: 'Retention', value: 'Backend policy', helper: 'Retention bukan tanggung jawab client.' },
        ],
      },
    ],
    timeline: [
      'Audit detail route tersedia untuk inspect layout.',
      'Export dan mark reviewed tidak aktif.',
      'Server menentukan actor, ordering, dan retention.',
    ],
  },
  {
    id: 'contract-missing',
    title: 'Contract missing event',
    kicker: 'Audit unavailable preview',
    badge: 'Unavailable state',
    description: 'Contoh detail ketika contract audit belum tersedia, sehingga UI menahan data dan menampilkan status aman.',
    backHref: '/audit',
    backLabel: 'Kembali ke audit',
    actions: [
      { label: 'Retry load', status: 'disabled' },
      { label: 'Open contract', status: 'docs only' },
    ],
    sections: [
      {
        title: 'Unavailable contract',
        rows: [
          { label: 'Endpoint', value: 'TBD', helper: 'Endpoint final harus dicatat di matrix.' },
          { label: 'Error code', value: '—', helper: 'Kode error dari backend belum tersedia.' },
          { label: 'Fallback', value: 'Safe empty', helper: 'Tidak menampilkan data palsu.' },
        ],
      },
      {
        title: 'Review notes',
        rows: [
          { label: 'Owner', value: 'Backend + admin web', helper: 'Contract harus disepakati bersama.' },
          { label: 'Pagination', value: 'Required', helper: 'Pagination audit wajib ada.' },
          { label: 'Filter', value: 'Required', helper: 'Filter actor/outlet/action wajib dikontrak.' },
        ],
      },
    ],
    timeline: [
      'Detail unavailable memastikan UI tidak blank.',
      'Contract matrix harus lengkap sebelum integrasi.',
      'Audit data tidak boleh direka di frontend.',
    ],
  },
] as const;

export const auditDetailLinks = auditDetailPreviews.map((detail) => ({
  href: `/audit/${detail.id}`,
  label: detail.title,
  description: detail.description,
})) as readonly { href: string; label: string; description: string }[];

export function getAuditDetailPreview(id: string) {
  return auditDetailPreviews.find((detail) => detail.id === id);
}
