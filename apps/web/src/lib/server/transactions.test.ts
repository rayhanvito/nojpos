import { describe, expect, it } from 'vitest';

import { BackendApiError } from './backend-client';
import { mapTransactionsEnvelope, mapTransactionsReadToViewModel, parseTransactionsReadQuery } from './transactions';
import type { WebAdminSessionPayload } from './session-cookie';

const session: WebAdminSessionPayload = {
  token: 'secret-token',
  issued_at: 1782000000,
  expires_at: 1782028800,
  device_uuid: 'web-device-1',
  user: {
    id: 'user-1',
    name: 'Owner',
    role: 'owner',
    business_id: 'business-1',
    is_superadmin: false,
  },
  business: { id: 'business-1', name: 'Kopi Senja' },
};

describe('transactions read-only server helper', () => {
  it('validates query whitelist and supported values', () => {
    expect(parseTransactionsReadQuery(new URLSearchParams('status=paid&page=2&per_page=50'))).toMatchObject({
      status: 'paid',
      page: 2,
      per_page: 50,
    });

    expect(() => parseTransactionsReadQuery(new URLSearchParams('unsafe=true'))).toThrow(BackendApiError);
    expect(() => parseTransactionsReadQuery(new URLSearchParams('date_from=2026-99-99'))).toThrow(BackendApiError);
    expect(() => parseTransactionsReadQuery(new URLSearchParams('per_page=500'))).toThrow(BackendApiError);
  });

  it('maps backend rows into stable safe contract without token or cross-tenant data', () => {
    const envelope = mapTransactionsEnvelope(
      {
        data: {
          transactions: [
            {
              id: 'trx-1',
              business_id: 'business-1',
              outlet_id: 'outlet-1',
              cashier_id: 'cashier-1',
              number: 'TRX-001',
              status: 'paid',
              subtotal: 42000,
              discount_total: 2000,
              service_charge_total: 0,
              tax_total: 0,
              rounding_total: 0,
              grand_total: 40000,
              created_at: '2026-06-21T08:42:00.000000Z',
              cashier: { id: 'cashier-1', name: 'Ayu' },
              customer: { id: 'customer-1', name: 'Budi' },
              items: [{ quantity: 2 }],
              payments: [{ method: 'qris', amount: 40000 }],
            },
            {
              id: 'trx-other',
              business_id: 'other-business',
              number: 'TRX-OTHER',
              status: 'paid',
              grand_total: 999999,
            },
          ],
        },
        meta: { request_id: 'req-test' },
      },
      { page: 1, per_page: 20 },
      session,
    );
    const serialized = JSON.stringify(envelope);

    expect(envelope.data.rows).toHaveLength(1);
    expect(envelope.data.rows[0].code).toBe('TRX-001');
    expect(envelope.data.rows[0].payment_method.label).toBe('QRIS');
    expect(envelope.data.totals.net_sales).toBe(40000);
    expect(serialized).not.toContain(session.token);
    expect(serialized).not.toContain('TRX-OTHER');
  });

  it('maps read response to preview-compatible page model data', () => {
    const envelope = mapTransactionsEnvelope(
      {
        data: {
          transactions: [
            {
              id: 'trx-1',
              business_id: 'business-1',
              number: 'TRX-001',
              status: 'paid',
              subtotal: 10000,
              discount_total: 0,
              grand_total: 10000,
              created_at: '2026-06-21T08:42:00.000000Z',
              payments: [{ method: 'cash', amount: 10000 }],
            },
          ],
        },
      },
      { page: 1, per_page: 20 },
      session,
    );

    const viewModel = mapTransactionsReadToViewModel(envelope.data);

    expect(viewModel.table.rows[0]).toEqual(['TRX-001', 'Semua cabang', 'Tunai', 'Rp10.000', 'Lunas']);
    expect(viewModel.metrics[1].value).toBe('Rp10.000');
    expect(viewModel.detailLinks[0].href).toBe('/transactions/trx-1');
  });
});
