import type { DetailPreviewFixture, PreviewLane, PreviewMetricFixture, PreviewTableFixture } from './types';

export const paymentMetrics: readonly PreviewMetricFixture[] = [
  { label: 'Payment events', value: '—', description: 'Event pembayaran berasal dari backend.', badge: 'Preview' },
  { label: 'Gateway status', value: 'Read-only', description: 'Status gateway tidak dipolling di UI-first.', badge: 'Gated', tone: 'warning' },
  { label: 'Payment action', value: 'Read-only', description: 'Capture, refund, void, retry tetap disabled.', badge: 'Disabled' },
] as const;

export const paymentLanes: readonly PreviewLane[] = [
  { label: 'Event list', items: ['Gateway label', 'Masked reference', 'Settlement state'] },
  { label: 'Detail', items: ['Tender metadata', 'Transaction link slot', 'Action gate'] },
  { label: 'Controls', items: ['Refund disabled', 'Void disabled', 'Retry disabled'] },
] as const;

export const paymentsTable: PreviewTableFixture = {
  columns: ['Payment', 'Method', 'Reference', 'Status'],
  rows: [
    ['PAY-PREVIEW-001', 'QRIS preview', 'Masked by backend', 'Preview'],
    ['PAY-PREVIEW-002', 'Card preview', '—', 'Read-only'],
    ['PAY-PREVIEW-003', 'Settlement gated', 'Gateway ref', 'Backend gated'],
  ],
  caption: 'Payment action tidak aktif. Tidak ada capture/refund/void/retry di browser.',
  primaryColumn: 'Payment',
  statusColumn: 'Status',
  metaColumns: ['Method', 'Reference'],
} as const;

export const paymentStatusPreview = {
  label: 'Payment controls',
  status: 'Disabled',
  description: 'Semua payment action menunggu contract gateway, idempotency, audit, dan permission server.',
} as const;

export const paymentDetailPreviews: readonly DetailPreviewFixture[] = [
  {
    id: 'pay-preview-001',
    title: 'PAY-PREVIEW-001',
    kicker: 'Payment detail',
    badge: 'Read-only',
    description: 'Detail pembayaran preview untuk settlement dan gateway metadata tanpa payment action aktif.',
    backHref: '/payments',
    backLabel: 'Kembali ke payments',
    actions: [
      { label: 'Refund payment', status: 'disabled' },
      { label: 'Void payment', status: 'disabled' },
      { label: 'Retry payment', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Payment identity',
        rows: [
          { label: 'Payment ID', value: 'PAY-PREVIEW-001', helper: 'ID real dari backend.' },
          { label: 'Method', value: 'QRIS preview', helper: 'Method berasal dari payment contract.' },
          { label: 'Reference', value: 'Masked', helper: 'Gateway ref harus dimasking.' },
        ],
      },
      {
        title: 'Action boundary',
        rows: [
          { label: 'Refund', value: 'Disabled', helper: 'Butuh endpoint dan idempotency key.' },
          { label: 'Void', value: 'Disabled', helper: 'Butuh state transition contract.' },
          { label: 'Settlement', value: '—', helper: 'Status settlement final dari backend.' },
        ],
      },
    ],
    timeline: ['Payment detail route tersedia.', 'Refund/void/retry disabled.', 'Gateway reference tetap masked.'],
  },
  {
    id: 'pay-preview-002',
    title: 'PAY-PREVIEW-002',
    kicker: 'Card payment preview',
    badge: 'Backend gated',
    description: 'Preview card payment untuk UI metadata dan unavailable state tanpa integrasi gateway.',
    backHref: '/payments',
    backLabel: 'Kembali ke payments',
    actions: [
      { label: 'Open gateway event', status: 'disabled' },
      { label: 'Export payment log', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Gateway metadata',
        rows: [
          { label: 'Gateway', value: 'Backend gateway', helper: 'Tidak ada SDK gateway di preview.' },
          { label: 'Token', value: 'Not stored', helper: 'Tidak ada credential payment di browser.' },
          { label: 'Trace', value: '—', helper: 'Trace dari backend saat tersedia.' },
        ],
      },
      {
        title: 'Controls',
        rows: [
          { label: 'Retry', value: 'Disabled', helper: 'Retry action butuh server job.' },
          { label: 'Export', value: 'Disabled', helper: 'Export menunggu permission.' },
          { label: 'Audit', value: 'Required', helper: 'Payment action harus tercatat.' },
        ],
      },
    ],
    timeline: ['Card payment area disiapkan.', 'Tidak ada gateway call.', 'Payment action gated.'],
  },
] as const;

export const paymentDetailLinks = paymentDetailPreviews.map((detail) => ({ href: `/payments/${detail.id}`, label: detail.title, description: detail.description })) as readonly { href: string; label: string; description: string }[];

export function getPaymentDetailPreview(id: string) {
  return paymentDetailPreviews.find((detail) => detail.id === id);
}
