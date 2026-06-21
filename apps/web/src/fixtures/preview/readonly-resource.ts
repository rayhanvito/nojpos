import type { PreviewTableFixture } from './types';

const inventoryTable: PreviewTableFixture = {
  columns: ['Product Name', 'On Hand Quantity', 'Outlet', 'Status'],
  rows: [
    ['Item preview', '—', 'Backend gated', 'Preview'],
    ['Item kontrak stok', '—', 'Backend gated', 'Read-only'],
  ],
} as const;

const reportTable: PreviewTableFixture = {
  columns: ['Metric', 'Value', 'Period', 'Status'],
  rows: [
    ['Data laporan', '—', 'Dari backend', 'Preview'],
    ['Export', 'Belum tersedia', 'Contract gated', 'Read-only'],
  ],
} as const;

const defaultTable: PreviewTableFixture = {
  columns: ['Label', 'Value', 'Scope', 'Status'],
  rows: [
    ['Data preview', 'Akan diisi backend', 'Tenant admin', 'Read-only'],
  ],
} as const;

export function readonlyResourceTableFor(path: string): PreviewTableFixture {
  if (path.includes('inventory')) {
    return inventoryTable;
  }

  if (path.includes('reports')) {
    return reportTable;
  }

  return defaultTable;
}
