import { NextRequest, NextResponse } from 'next/server';

import { outletMetrics, outletsTable } from '../../fixtures/preview';
import { getReadOnlyPageModel, handleReadOnlyList, type ReadOnlyListConfig, type ReadOnlyPageModel, type ReadOnlyRow } from './admin-readonly-list';

const outletsConfig: ReadOnlyListConfig = {
  resource: 'outlets',
  backendPath: '/outlets',
  allowedParams: ['search', 'status', 'page', 'per_page'],
  statusValues: ['active', 'inactive', 'all'],
  title: 'Outlet',
  fallbackDescription: 'Server admin belum bisa menghubungi backend outlet.',
  columns: ['Outlet', 'Alamat', 'Timezone', 'Status'],
  fixtureTable: outletsTable,
  fixtureMetrics: outletMetrics,
  note: 'Outlet ditampilkan read-only. Store open/close, edit, delete, dan rekonsiliasi kas tetap disabled.',
  rowMapper(row, sessionBusinessId) {
    if (typeof row.business_id === 'string' && row.business_id !== sessionBusinessId) return null;
    return {
      id: String(row.id ?? ''),
      name: String(row.name ?? 'Outlet'),
      address_short: typeof row.address_short === 'string' ? row.address_short : typeof row.address === 'string' ? row.address.slice(0, 80) : null,
      timezone: typeof row.timezone === 'string' ? row.timezone : 'Asia/Jakarta',
      status: String(row.status ?? (row.is_active === false ? 'inactive' : 'active')),
      is_default: Boolean(row.is_default),
      can_edit: false,
      can_open_store: false,
      can_close_store: false,
      can_delete: false,
    };
  },
  tableMapper(rows) {
    return rows.map((row) => [String(row.name ?? 'Outlet'), String(row.address_short ?? '-'), String(row.timezone ?? 'Asia/Jakarta'), String(row.status ?? '-')]);
  },
  totalsMapper(rows) {
    return {
      outlet_count: rows.length,
      active_count: rows.filter((row) => row.status === 'active').length,
      inactive_count: rows.filter((row) => row.status === 'inactive').length,
    };
  },
};

export function handleAdminOutlets(request: NextRequest): Promise<NextResponse> {
  return handleReadOnlyList(request, outletsConfig);
}

export function getOutletsPageModel(): Promise<ReadOnlyPageModel> {
  return getReadOnlyPageModel(outletsConfig);
}

export { outletsConfig };
export type { ReadOnlyPageModel as OutletsPageModel, ReadOnlyRow as OutletReadOnlyRow };
