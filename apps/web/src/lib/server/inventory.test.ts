import { describe, expect, it } from 'vitest';

import { BackendApiError } from './backend-client';
import { mapInventoryEnvelope, mapInventoryReadToViewModel, parseInventoryReadQuery } from './inventory';
import { createSessionPayload } from './session-cookie';

const session = createSessionPayload({
  token: 'inventory-token-secret',
  user: {
    id: 'owner-user',
    name: 'Owner User',
    role: 'owner',
    business_id: 'business-1',
    is_superadmin: false,
  },
  business: { id: 'business-1', name: 'Kopi Senja' },
  device_uuid: 'web-device-1',
});

describe('inventory read BFF helper', () => {
  it('rejects unknown query params', () => {
    expect(() => parseInventoryReadQuery(new URLSearchParams('debug=true'))).toThrow(BackendApiError);
  });

  it('validates stock status and pagination', () => {
    expect(() => parseInventoryReadQuery(new URLSearchParams('stock_status=expired'))).toThrow(BackendApiError);
    expect(() => parseInventoryReadQuery(new URLSearchParams('per_page=500'))).toThrow(BackendApiError);

    expect(parseInventoryReadQuery(new URLSearchParams('stock_status=low&page=2&per_page=10'))).toEqual({
      stock_status: 'low',
      page: 2,
      per_page: 10,
    });
  });

  it('maps backend rows into stable safe read model without token or other outlet rows', () => {
    const mapped = mapInventoryEnvelope(
      {
        data: {
          inventory: [
            {
              product_id: 'product-1',
              name: 'Kopi Susu',
              sku: 'SKU-KOPI-001',
              category: { id: 'category-1', name: 'Minuman' },
              outlet_id: '135159a6-d523-470f-93e1-5a3f3706a001',
              outlet_name: 'Cabang Utama',
              stock_on_hand: 8,
              low_stock_threshold: 10,
              unit: 'pcs',
              track_stock: true,
              updated_at: '2026-06-21T08:20:00.000000Z',
            },
            {
              product_id: 'product-other',
              name: 'Tenant Lain',
              outlet_id: 'outlet-other',
              outlet_name: 'Bocor',
              stock_on_hand: 999,
            },
          ],
        },
        meta: { request_id: 'req-inventory-test' },
      },
      { page: 1, per_page: 20, stock_status: 'all' },
      session,
    );
    const serialized = JSON.stringify(mapped);

    expect(mapped.data.rows).toHaveLength(2);
    expect(mapped.data.rows[0]?.name).toBe('Kopi Susu');
    expect(mapped.data.rows[0]?.status).toBe('low');
    expect(mapped.data.rows[0]?.can_adjust).toBe(false);
    expect(mapped.data.totals.low_stock_count).toBe(1);
    expect(serialized).not.toContain('inventory-token-secret');
    expect(serialized.toLowerCase()).not.toContain('bearer');
  });

  it('maps inventory data into table view model', () => {
    const mapped = mapInventoryEnvelope(
      {
        data: {
          inventory: [
            {
              product_id: 'product-1',
              name: 'Kopi Susu',
              outlet_id: '135159a6-d523-470f-93e1-5a3f3706a001',
              outlet_name: 'Cabang Utama',
              stock_on_hand: 12,
              low_stock_threshold: 10,
            },
          ],
        },
        meta: {},
      },
      { page: 1, per_page: 20, stock_status: 'all' },
      session,
    );

    const viewModel = mapInventoryReadToViewModel(mapped.data);

    expect(viewModel.metrics[0]?.value).toBe('1');
    expect(viewModel.table.columns).toContain('Produk');
    expect(viewModel.table.rows[0]?.[0]).toBe('Kopi Susu');
  });
});
