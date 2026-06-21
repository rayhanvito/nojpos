import { describe, expect, it } from 'vitest';

import type { ApiEnvelope } from '../api/envelope';
import type { BackendCatalogProductsData } from './backend-client';
import { mapCatalogProductsEnvelope, mapCatalogReadToViewModel, parseCatalogReadQuery } from './catalog';
import { createSessionPayload } from './session-cookie';

const session = createSessionPayload({
  token: 'secret-token',
  user: { id: 'owner-1', name: 'Owner', role: 'owner', business_id: 'business-1', is_superadmin: false },
  business: { id: 'business-1', name: 'Kopi Senja' },
  device_uuid: 'web-device-1',
});

function backendEnvelope(): ApiEnvelope<BackendCatalogProductsData> {
  return {
    data: {
      categories: [{ id: 'category-1', business_id: 'business-1', name: 'Minuman' }],
      products: [
        { id: 'product-1', business_id: 'business-1', product_category_id: 'category-1', name: 'Kopi Susu', barcode: 'SKU-KOPI-001', price: 18000, track_stock: true },
        { id: 'product-2', business_id: 'business-2', product_category_id: 'category-2', name: 'Tenant Lain', price: 999999 },
      ],
    },
    meta: { request_id: 'req-catalog-unit' },
  };
}

describe('catalog read-only mapper', () => {
  it('parses whitelisted query only', () => {
    const query = parseCatalogReadQuery(new URLSearchParams('search=Kopi&status=active&page=2&per_page=10'));

    expect(query).toEqual({ search: 'Kopi', status: 'active', page: 2, per_page: 10 });
  });

  it('rejects unknown query params', () => {
    expect(() => parseCatalogReadQuery(new URLSearchParams('debug=true'))).toThrow('Filter catalog tidak valid.');
  });

  it('maps products without token and drops other tenant rows', () => {
    const mapped = mapCatalogProductsEnvelope(backendEnvelope(), { page: 1, per_page: 20, status: 'all' }, session);
    const serialized = JSON.stringify(mapped);

    expect(mapped.data.rows).toHaveLength(1);
    expect(mapped.data.rows[0].name).toBe('Kopi Susu');
    expect(mapped.data.rows[0].price).toBe(18000);
    expect(mapped.data.rows[0].can_edit).toBe(false);
    expect(mapped.data.rows[0].can_delete).toBe(false);
    expect(serialized).not.toContain('secret-token');
    expect(serialized).not.toContain('Tenant Lain');
  });

  it('maps a stable preview table view model', () => {
    const mapped = mapCatalogProductsEnvelope(backendEnvelope(), { page: 1, per_page: 20, status: 'all' }, session);
    const viewModel = mapCatalogReadToViewModel(mapped.data);

    expect(viewModel.table.columns).toEqual(['Produk', 'Kategori', 'Harga', 'Ketersediaan', 'Status']);
    expect(viewModel.table.rows[0]).toContain('Kopi Susu');
    expect(viewModel.metrics[0].label).toBe('Produk');
  });
});
