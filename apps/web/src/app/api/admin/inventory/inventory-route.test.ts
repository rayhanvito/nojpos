import { NextRequest } from 'next/server';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createSessionPayload, sealSessionCookie } from '../../../../lib/server/session-cookie';
import { GET as inventoryRoute } from './route';

const secret = 'test-secret-for-web-admin-session-cookie-32';
const cookieName = 'nojpos_test_session';
const ownerToken = 'owner-inventory-token';

function adminRequest(path: string, cookie?: string): NextRequest {
  return new NextRequest(`http://localhost:3000${path}`, {
    method: 'GET',
    headers: {
      host: 'localhost:3000',
      ...(cookie ? { cookie } : {}),
    },
  });
}

function sessionCookie(role: 'owner' | 'admin' | 'cashier' | 'superadmin' = 'owner'): string {
  const sealed = sealSessionCookie(
    createSessionPayload({
      token: role === 'owner' ? ownerToken : `${role}-backend-token`,
      user: {
        id: `${role}-user`,
        name: `${role} User`,
        role,
        business_id: role === 'superadmin' ? null : 'business-1',
        is_superadmin: role === 'superadmin',
      },
      business: role === 'superadmin' ? null : { id: 'business-1', name: 'Kopi Senja' },
      device_uuid: 'web-device-1',
    }),
  );

  return `${cookieName}=${sealed}`;
}

function jsonResponse(payload: unknown, status = 200): Response {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function inventoryPayload() {
  return {
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
  };
}

describe('Web Admin BFF inventory route', () => {
  beforeEach(() => {
    vi.stubEnv('BACKEND_INTERNAL_API_URL', 'http://backend.test/api/v1');
    vi.stubEnv('WEB_SESSION_SECRET', secret);
    vi.stubEnv('WEB_SESSION_COOKIE_NAME', cookieName);
    vi.stubEnv('WEB_SESSION_TTL_SECONDS', '3600');
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it('returns 401 without session cookie', async () => {
    const response = await inventoryRoute(adminRequest('/api/admin/inventory'));
    const payload = await response.json();

    expect(response.status).toBe(401);
    expect(payload.error.code).toBe('UNAUTHENTICATED');
    expect(response.headers.get('set-cookie')).toContain('Max-Age=0');
  });

  it('rejects non-tenant admin roles before proxying to backend', async () => {
    const backend = vi.fn();
    vi.stubGlobal('fetch', backend);

    const response = await inventoryRoute(adminRequest('/api/admin/inventory', sessionCookie('cashier')));
    const payload = await response.json();

    expect(response.status).toBe(403);
    expect(payload.error.code).toBe('FORBIDDEN');
    expect(backend).not.toHaveBeenCalled();
  });

  it('does not forward unknown query params to backend', async () => {
    const backend = vi.fn();
    vi.stubGlobal('fetch', backend);

    const response = await inventoryRoute(adminRequest('/api/admin/inventory?debug=true', sessionCookie()));
    const payload = await response.json();

    expect(response.status).toBe(422);
    expect(payload.error.code).toBe('VALIDATION_ERROR');
    expect(backend).not.toHaveBeenCalled();
  });

  it('proxies allowed server-side request and never returns token or other outlet rows', async () => {
    const backend = vi.fn().mockResolvedValue(jsonResponse(inventoryPayload()));
    vi.stubGlobal('fetch', backend);

    const response = await inventoryRoute(adminRequest('/api/admin/inventory?stock_status=low&page=1&per_page=20', sessionCookie()));
    const payload = await response.json();
    const serializedPayload = JSON.stringify(payload);
    const [backendUrl, backendInit] = backend.mock.calls[0] as [string, RequestInit];

    expect(response.status).toBe(200);
    expect(backendUrl).toBe('http://backend.test/api/v1/inventory');
    expect(backendInit.headers).toBeInstanceOf(Headers);
    expect((backendInit.headers as Headers).get('Authorization')).toBe(`Bearer ${ownerToken}`);
    expect(serializedPayload).not.toContain(ownerToken);
    expect(serializedPayload.toLowerCase()).not.toContain('bearer');
    expect(serializedPayload).not.toContain('Tenant Lain');
    expect(payload.data.rows).toHaveLength(1);
    expect(payload.data.rows[0].name).toBe('Kopi Susu');
    expect(payload.data.rows[0].can_adjust).toBe(false);
  });
});
