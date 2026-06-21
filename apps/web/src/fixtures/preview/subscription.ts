import type { PreviewLane, PreviewMetricFixture, PreviewTableFixture } from './types';

export const subscriptionMetrics: readonly PreviewMetricFixture[] = [
  { label: 'Tenant plan', value: 'Read-only', description: 'Plan tenant ditampilkan sebagai preview.', badge: 'Preview' },
  { label: 'Renewal', value: '—', description: 'Tanggal renewal berasal dari billing backend.', badge: 'Gated', tone: 'warning' },
  { label: 'Change plan', value: 'Read-only', description: 'Upgrade/downgrade disabled sampai contract billing siap.', badge: 'Disabled' },
] as const;

export const subscriptionLanes: readonly PreviewLane[] = [
  { label: 'Plan', items: ['Current package', 'Included limits', 'Feature gates'] },
  { label: 'Billing', items: ['Renewal placement', 'Invoice placeholder', 'Payment method slot'] },
  { label: 'Controls', items: ['Change plan disabled', 'Cancel disabled', 'Export invoice disabled'] },
] as const;

export const subscriptionTable: PreviewTableFixture = {
  columns: ['Item', 'Value', 'Owner', 'Status'],
  rows: [
    ['Current plan', 'Business preview', 'Platform backend', 'Preview'],
    ['Renewal', '—', 'Billing backend', 'Read-only'],
    ['Change plan', 'Disabled', 'Platform policy', 'Backend gated'],
  ],
  caption: 'Subscription tenant hanya status preview. Tidak ada billing action aktif.',
} as const;

export const tenantSubscriptionStatus = {
  plan: 'Business preview plan',
  status: 'Change plan disabled',
  renewal: '—',
  note: 'Subscription card memperlihatkan posisi UI untuk renewal, invoice, dan package limit. Semua perubahan plan tetap platform-gated.',
} as const;
