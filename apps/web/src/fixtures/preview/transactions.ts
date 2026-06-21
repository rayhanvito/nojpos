import type { DetailPreviewFixture, PreviewLane, PreviewMetricFixture, PreviewTableFixture } from './types';

export const transactionMetrics: readonly PreviewMetricFixture[] = [
  { label: 'Transactions', value: '—', description: 'Jumlah transaksi berasal dari backend.', badge: 'Preview' },
  { label: 'Totals', value: '—', description: 'Total, pajak, service, rounding tidak dihitung di browser.', badge: 'Safe' },
  { label: 'Void/refund', value: 'Read-only', description: 'Void/refund action disabled sampai contract tersedia.', badge: 'Gated', tone: 'warning' },
] as const;

export const transactionLanes: readonly PreviewLane[] = [
  { label: 'List', items: ['Receipt number', 'Outlet scope', 'Payment status'] },
  { label: 'Detail', items: ['Line item placeholder', 'Tender summary', 'Audit timeline'] },
  { label: 'Controls', items: ['Void disabled', 'Refund disabled', 'Export disabled'] },
] as const;

export const transactionsTable: PreviewTableFixture = {
  columns: ['Transaction', 'Outlet', 'Payment', 'Status'],
  rows: [
    ['TRX-PREVIEW-001', 'Outlet preview', 'Payment preview', 'Preview'],
    ['TRX-PREVIEW-002', 'Outlet scope backend', 'Split tender gated', 'Read-only'],
    ['TRX-PREVIEW-003', '—', 'Refund gated', 'Backend gated'],
  ],
  caption: 'Transaksi bersifat read-only. Tidak ada kalkulasi total atau refund di frontend.',
  primaryColumn: 'Transaction',
  statusColumn: 'Status',
  metaColumns: ['Outlet', 'Payment'],
} as const;

export const transactionDetailPreviews: readonly DetailPreviewFixture[] = [
  {
    id: 'trx-preview-001',
    title: 'TRX-PREVIEW-001',
    kicker: 'Transaction detail',
    badge: 'Read-only',
    description: 'Detail transaksi preview untuk layout receipt, tender, dan audit tanpa menghitung total di browser.',
    backHref: '/transactions',
    backLabel: 'Kembali ke transactions',
    actions: [
      { label: 'Void transaction', status: 'disabled' },
      { label: 'Refund transaction', status: 'disabled' },
      { label: 'Export receipt', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Receipt identity',
        rows: [
          { label: 'Receipt', value: 'TRX-PREVIEW-001', helper: 'Nomor receipt real dari backend.' },
          { label: 'Outlet', value: 'Outlet preview', helper: 'Outlet scope final dari backend.' },
          { label: 'Cashier', value: 'Backend actor', helper: 'Actor tidak disimulasikan.' },
        ],
      },
      {
        title: 'Payment & total gate',
        rows: [
          { label: 'Subtotal', value: '—', helper: 'Tidak menghitung nilai transaksi di browser.' },
          { label: 'Payment', value: 'Payment preview', helper: 'Tender detail dari backend.' },
          { label: 'Refund/Void', value: 'Disabled', helper: 'Butuh idempotency, audit, dan server validation.' },
        ],
      },
    ],
    timeline: ['Receipt layout tersedia.', 'Void/refund disabled.', 'Backend tetap sumber total dan status final.'],
  },
  {
    id: 'trx-preview-002',
    title: 'TRX-PREVIEW-002',
    kicker: 'Split tender preview',
    badge: 'Backend gated',
    description: 'Preview transaksi split tender tanpa menyusun logic pembayaran client-side.',
    backHref: '/transactions',
    backLabel: 'Kembali ke transactions',
    actions: [
      { label: 'Open payment trail', status: 'disabled' },
      { label: 'Refund line item', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Tender placement',
        rows: [
          { label: 'Method', value: 'Split tender gated', helper: 'Detail tender menunggu contract.' },
          { label: 'Gateway ref', value: 'Masked', helper: 'Reference pembayaran harus dimasking.' },
          { label: 'Settlement', value: '—', helper: 'Status settlement dari backend.' },
        ],
      },
      {
        title: 'Boundary',
        rows: [
          { label: 'Calculation', value: 'Server only', helper: 'Tidak ada kalkulasi harga/pajak/service.' },
          { label: 'Refund', value: 'Disabled', helper: 'Refund action gated.' },
          { label: 'Export', value: 'Disabled', helper: 'Export menunggu job contract.' },
        ],
      },
    ],
    timeline: ['Split tender area disiapkan.', 'Refund line item disabled.', 'Settlement dari backend.'],
  },
] as const;

export const transactionDetailLinks = transactionDetailPreviews.map((detail) => ({ href: `/transactions/${detail.id}`, label: detail.title, description: detail.description })) as readonly { href: string; label: string; description: string }[];

export function getTransactionDetailPreview(id: string) {
  return transactionDetailPreviews.find((detail) => detail.id === id);
}
