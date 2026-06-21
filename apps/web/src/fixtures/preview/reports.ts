import type { PreviewLane, ReadonlyResourceFixture } from './types';

export const reportResources: readonly ReadonlyResourceFixture[] = [
  { label: 'Ringkasan penjualan', path: '/reports/sales-summary' },
  { label: 'Produk terjual', path: '/reports/sold-products' },
  { label: 'Metode pembayaran', path: '/reports/payment-methods' },
  { label: 'Shift kasir', path: '/reports/cashier-shifts' },
  { label: 'Audit void/refund', path: '/reports/void-refund-audit' },
  { label: 'Top 10', path: '/reports/top-10' },
] as const;

export const reportLanes: readonly PreviewLane[] = [
  { label: 'Filter', items: ['Outlet selector disabled', 'Period control disabled', 'Search placement'] },
  { label: 'Render', items: ['Compact table', 'Empty state', 'Error state'] },
  { label: 'Output', items: ['Download gated', 'Server policy later', 'Log slot later'] },
] as const;
