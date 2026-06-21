import { describe, expect, it } from 'vitest';

import type { ApiEnvelope } from '../api/envelope';
import { mapReadOnlyEnvelope, parseReadOnlyQuery, type ReadOnlyListConfig } from './admin-readonly-list';
import { createSessionPayload } from './session-cookie';

const config: ReadOnlyListConfig = {
  resource: 'staff',
  backendPath: '/staff',
  allowedParams: ['search', 'status', 'page', 'per_page'],
  statusValues: ['active', 'inactive', 'all'],
  title: 'Staff',
  fallbackDescription: 'Fallback',
  columns: ['Nama'],
  fixtureTable: { columns: ['Nama'], rows: [] },
  fixtureMetrics: [],
  note: 'Read-only note',
  rowMapper(row, businessId) {
    if (typeof row.business_id === 'string' && row.business_id !== businessId) return null;
    return { id: row.id, name: row.name, can_delete: false };
  },
  tableMapper(rows) {
    return rows.map((row) => [String(row.name ?? '-')]);
  },
  totalsMapper(rows) {
    return { staff_count: rows.length };
  },
};

const session = createSessionPayload({
  token: 'secret-list-token',
  user: { id: 'owner-1', name: 'Owner', role: 'owner', business_id: 'business-1', is_superadmin: false },
  business: { id: 'business-1', name: 'Kopi Senja' },
  device_uuid: 'web-device-1',
});

describe('admin read-only list helper', () => {
  it('parses whitelisted query only', () => {
    expect(parseReadOnlyQuery(new URLSearchParams('search=Rina&status=active&page=2&per_page=10'), config)).toEqual({ search: 'Rina', status: 'active', page: 2, per_page: 10 });
  });

  it('rejects unknown query params', () => {
    expect(() => parseReadOnlyQuery(new URLSearchParams('debug=true'), config)).toThrow('Filter read-only tidak valid.');
  });

  it('maps rows without token and drops other tenant rows', () => {
    const envelope: ApiEnvelope<Record<string, unknown>> = {
      data: { staff: [{ id: 'staff-1', business_id: 'business-1', name: 'Rina' }, { id: 'staff-2', business_id: 'business-2', name: 'Other Tenant' }] },
      meta: { request_id: 'req-list' },
    };
    const mapped = mapReadOnlyEnvelope(envelope, { page: 1, per_page: 20 }, session, config);
    const serialized = JSON.stringify(mapped);

    expect(mapped.data.rows).toHaveLength(1);
    expect(mapped.data.rows[0].name).toBe('Rina');
    expect(serialized).not.toContain('secret-list-token');
    expect(serialized).not.toContain('Other Tenant');
  });
});
