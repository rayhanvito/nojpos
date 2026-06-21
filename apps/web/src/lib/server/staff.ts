import { NextRequest, NextResponse } from 'next/server';

import { staffDirectoryTable, type PreviewMetricFixture } from '../../fixtures/preview';
import { getReadOnlyPageModel, handleReadOnlyList, type ReadOnlyListConfig, type ReadOnlyPageModel, type ReadOnlyRow } from './admin-readonly-list';

const staffPreviewMetrics: readonly PreviewMetricFixture[] = [
  { label: 'Staf', value: 'Preview', description: 'Data contoh' },
  { label: 'Admin', value: 'Preview', description: 'Data contoh' },
  { label: 'Kasir', value: 'Preview', description: 'Data contoh' },
  { label: 'Aksi', value: 'Disabled', description: 'Read-only' },
];

const staffConfig: ReadOnlyListConfig = {
  resource: 'staff',
  backendPath: '/staff',
  allowedParams: ['search', 'outlet_id', 'role', 'status', 'page', 'per_page'],
  statusValues: ['active', 'inactive', 'all'],
  title: 'Staff',
  fallbackDescription: 'Server admin belum bisa menghubungi backend staff.',
  columns: ['Nama', 'Role', 'Outlet', 'Status'],
  fixtureTable: staffDirectoryTable,
  fixtureMetrics: staffPreviewMetrics,
  note: 'Staff ditampilkan read-only. Reset password, ban, delete, dan perubahan role tetap disabled.',
  rowMapper(row, sessionBusinessId) {
    if (typeof row.business_id === 'string' && row.business_id !== sessionBusinessId) return null;
    return {
      id: String(row.id ?? ''),
      name: String(row.name ?? 'Staff'),
      role: String(row.role ?? 'staff'),
      status: String(row.status ?? (row.is_active === false ? 'inactive' : 'active')),
      outlet_ids: Array.isArray(row.outlet_ids) ? row.outlet_ids.map(String) : row.outlet_id ? [String(row.outlet_id)] : [],
      outlet_names: Array.isArray(row.outlet_names) ? row.outlet_names.map(String) : row.outlet_name ? [String(row.outlet_name)] : [],
      last_activity_at: typeof row.last_activity_at === 'string' ? row.last_activity_at : null,
      can_edit: false,
      can_reset_password: false,
      can_delete: false,
    };
  },
  tableMapper(rows) {
    return rows.map((row) => [String(row.name ?? 'Staff'), String(row.role ?? '-'), Array.isArray(row.outlet_names) && row.outlet_names.length > 0 ? row.outlet_names.join(', ') : 'Semua outlet', String(row.status ?? '-')]);
  },
  totalsMapper(rows) {
    return {
      staff_count: rows.length,
      active_count: rows.filter((row) => row.status === 'active').length,
      admin_count: rows.filter((row) => row.role === 'admin' || row.role === 'owner').length,
      cashier_count: rows.filter((row) => row.role === 'cashier').length,
    };
  },
};

export function handleAdminStaff(request: NextRequest): Promise<NextResponse> {
  return handleReadOnlyList(request, staffConfig);
}

export function getStaffPageModel(): Promise<ReadOnlyPageModel> {
  return getReadOnlyPageModel(staffConfig);
}

export { staffConfig };
export type { ReadOnlyPageModel as StaffPageModel, ReadOnlyRow as StaffReadOnlyRow };
