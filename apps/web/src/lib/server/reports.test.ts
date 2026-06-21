import { describe, expect, it } from 'vitest';

import type { ApiEnvelope } from '../api/envelope';
import { mapReportEnvelope, parseReportsReadQuery } from './reports';
import { createSessionPayload } from './session-cookie';

const session = createSessionPayload({
  token: 'secret-report-token',
  user: { id: 'owner-1', name: 'Owner', role: 'owner', business_id: 'business-1', is_superadmin: false },
  business: { id: 'business-1', name: 'Kopi Senja' },
  device_uuid: 'web-device-1',
});

function reportEnvelope(): ApiEnvelope<Record<string, unknown>> {
  return {
    data: {
      summary: { gross_sales: 250000, net_sales: 225000, transaction_count: 12 },
      rows: [{ product_name: 'Kopi Susu', quantity: 8, gross_sales: 144000 }],
      chart: [{ date: '2026-06-21', gross_sales: 250000 }],
    },
    meta: { request_id: 'req-report-unit' },
  };
}

describe('reports read-only mapper', () => {
  it('parses whitelisted query only', () => {
    const query = parseReportsReadQuery(new URLSearchParams('date_from=2026-06-01&date_to=2026-06-21&range=week'));

    expect(query).toEqual({ date_from: '2026-06-01', date_to: '2026-06-21', range: 'week' });
  });

  it('rejects unknown query params', () => {
    expect(() => parseReportsReadQuery(new URLSearchParams('export=csv'))).toThrow('Filter laporan tidak valid.');
  });

  it('rejects over-wide ranges', () => {
    expect(() => parseReportsReadQuery(new URLSearchParams('date_from=2026-06-01&date_to=2026-07-15'))).toThrow('Filter laporan tidak valid.');
  });

  it('maps backend reports without leaking token', () => {
    const mapped = mapReportEnvelope(reportEnvelope(), { date_from: '2026-06-01', date_to: '2026-06-21', range: 'week' }, session, 'top-products');
    const serialized = JSON.stringify(mapped);

    expect(mapped.data.meta.business_id).toBe('business-1');
    expect(mapped.data.rows).toHaveLength(1);
    expect(mapped.data.rows[0].product_name).toBe('Kopi Susu');
    expect(mapped.data.data_notes[0].message).toContain('disabled');
    expect(serialized).not.toContain('secret-report-token');
    expect(serialized.toLowerCase()).not.toContain('bearer');
  });
});
