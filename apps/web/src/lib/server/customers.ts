import { NextRequest, NextResponse } from 'next/server';

import { customerMetrics, customersTable } from '../../fixtures/preview';
import { getReadOnlyPageModel, handleReadOnlyList, type ReadOnlyListConfig, type ReadOnlyPageModel, type ReadOnlyRow } from './admin-readonly-list';

const customersConfig: ReadOnlyListConfig = {
  resource: 'customers',
  backendPath: '/customers',
  allowedParams: ['search', 'group', 'status', 'page', 'per_page'],
  statusValues: ['active', 'archived', 'all'],
  title: 'Pelanggan',
  fallbackDescription: 'Server admin belum bisa menghubungi backend pelanggan.',
  columns: ['Nama', 'Kontak', 'Group', 'Status'],
  fixtureTable: customersTable,
  fixtureMetrics: customerMetrics,
  note: 'Customer ditampilkan read-only dengan kontak masked. Export dan delete tetap disabled.',
  rowMapper(row, sessionBusinessId) {
    if (typeof row.business_id === 'string' && row.business_id !== sessionBusinessId) return null;
    return {
      id: String(row.id ?? ''),
      name: typeof row.name === 'string' && row.name.trim() ? row.name : 'Pelanggan',
      phone_masked: maskPhone(row.phone),
      email_masked: maskEmail(row.email),
      group: typeof row.group === 'string' ? row.group : null,
      status: String(row.status ?? (row.deleted_at ? 'archived' : 'active')),
      last_transaction_at: typeof row.last_transaction_at === 'string' ? row.last_transaction_at : null,
      can_edit: false,
      can_delete: false,
      can_export: false,
    };
  },
  tableMapper(rows) {
    return rows.map((row) => [String(row.name ?? 'Pelanggan'), String(row.phone_masked ?? row.email_masked ?? '-'), String(row.group ?? '-'), String(row.status ?? '-')]);
  },
  totalsMapper(rows) {
    return {
      customer_count: rows.length,
      active_count: rows.filter((row) => row.status === 'active').length,
      archived_count: rows.filter((row) => row.status === 'archived').length,
    };
  },
};

export function handleAdminCustomers(request: NextRequest): Promise<NextResponse> {
  return handleReadOnlyList(request, customersConfig);
}

export function getCustomersPageModel(): Promise<ReadOnlyPageModel> {
  return getReadOnlyPageModel(customersConfig);
}

function maskPhone(value: unknown): string | null {
  if (typeof value !== 'string' || value.length < 6) return null;
  return `${value.slice(0, 4)}****${value.slice(-4)}`;
}

function maskEmail(value: unknown): string | null {
  if (typeof value !== 'string' || !value.includes('@')) return null;
  const [name, domain] = value.split('@');
  return `${name.slice(0, 1)}***@${domain}`;
}

export { customersConfig };
export type { ReadOnlyPageModel as CustomersPageModel, ReadOnlyRow as CustomerReadOnlyRow };
