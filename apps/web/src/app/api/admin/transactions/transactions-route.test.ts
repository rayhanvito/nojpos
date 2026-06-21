import { NextRequest } from 'next/server';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createSessionPayload, sealSessionCookie } from '../../../../lib/server/session-cookie';
import { GET as transactionsRoute } from './route';

const secret = 'test-secret-for-web-admin-session-cookie-32';
const cookieName = 'nojpos_test_session';
const ownerToken = 'owner-transactions-token';

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

function transactionsPayload() {
  return {
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
          items: [{ product_id: 'product-1', name: 'Kopi', quantity: 2 }],
          payments: [{ method: 'qris', status: 'paid', amount: 40000 }],
        },
        {
          id: 'trx-other',
          business_id: 'business-other',
          number: 'TRX-OTHER',
          status: 'paid',
          grand_total: 999999,
          created_at: '2026-06-21T08:50:00.000000Z',
        },
      ],
    },
    meta: { request_id: 'req-transactions-test' },
  };
}

describe('Web Admin BFF transactions route', () => {
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
    const response = await transactionsRoute(adminRequest('/api/admin/transactions'));
    const payload = await response.json();

    expect(response.status).toBe(401);
    expect(payload.error.code).toBe('UNAUTHENTICATED');
    expect(response.headers.get('set-cookie')).toContain('Max-Age=0');
  });

  it('rejects non-tenant admin roles before proxying to backend', async () => {
    const backend = vi.fn();
    vi.stubGlobal('fetch', backend);

    const response = await transactionsRoute(adminRequest('/api/admin/transactions', sessionCookie('superadmin')));
    const payload = await response.json();

    expect(response.status).toBe(403);
    expect(payload.error.code).toBe('FORBIDDEN');
    expect(backend).not.toHaveBeenCalled();
  });

  it('does not forward unknown query params to backend', async () => {
    const backend = vi.fn();
    vi.stubGlobal('fetch', backend);

    const response = await transactionsRoute(adminRequest('/api/admin/transactions?debug=true', sessionCookie()));
    const payload = await response.json();

    expect(response.status).toBe(422);
    expect(payload.error.code).toBe('VALIDATION_ERROR');
    expect(backend).not.toHaveBeenCalled();
  });

  it('proxies allowed server-side request and never returns token or other tenant rows', async () => {
    const backend = vi.fn().mockResolvedValue(jsonResponse(transactionsPayload()));
    vi.stubGlobal('fetch', backend);

    const response = await transactionsRoute(adminRequest('/api/admin/transactions?status=paid&page=1&per_page=20', sessionCookie()));
    const payload = await response.json();
    const serializedPayload = JSON.stringify(payload);
    const [backendUrl, backendInit] = backend.mock.calls[0] as [string, RequestInit];

    expect(response.status).toBe(200);
    expect(backendUrl).toBe('http://backend.test/api/v1/transactions?status=paid');
    expect(backendInit.headers).toBeInstanceOf(Headers);
    expect((backendInit.headers as Headers).get('Authorization')).toBe(`Bearer ${ownerToken}`);
    expect(serializedPayload).not.toContain(ownerToken);
    expect(serializedPayload.toLowerCase()).not.toContain('bearer');
    expect(serializedPayload).not.toContain('business-other');
    expect(serializedPayload).not.toContain('TRX-OTHER');
    expect(payload.data.rows).toHaveLength(1);
    expect(payload.data.rows[0].code).toBe('TRX-001');
    expect(payload.data.totals.net_sales).toBe(40000);
  });
});
