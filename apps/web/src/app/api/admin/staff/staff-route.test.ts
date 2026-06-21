import { NextRequest } from 'next/server';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createSessionPayload, sealSessionCookie } from '../../../../lib/server/session-cookie';
import { GET as staffRoute } from './route';

const secret = 'test-secret-for-web-admin-session-cookie-32';
const cookieName = 'nojpos_test_session';
const ownerToken = 'owner-staff-token';

function adminRequest(path: string, cookie?: string): NextRequest {
  return new NextRequest(`http://localhost:3000${path}`, {
    method: 'GET',
    headers: { host: 'localhost:3000', ...(cookie ? { cookie } : {}) },
  });
}

function sessionCookie(role: 'owner' | 'cashier' = 'owner'): string {
  const sealed = sealSessionCookie(createSessionPayload({
    token: role === 'owner' ? ownerToken : 'cashier-token',
    user: { id: `${role}-user`, name: `${role} User`, role, business_id: 'business-1', is_superadmin: false },
    business: { id: 'business-1', name: 'Kopi Senja' },
    device_uuid: 'web-device-1',
  }));
  return `${cookieName}=${sealed}`;
}

function jsonResponse(payload: unknown): Response {
  return new Response(JSON.stringify(payload), { status: 200, headers: { 'Content-Type': 'application/json' } });
}

describe('Web Admin BFF staff route', () => {
  beforeEach(() => {
    vi.stubEnv('BACKEND_INTERNAL_API_URL', 'http://backend.test/api/v1');
    vi.stubEnv('WEB_SESSION_SECRET', secret);
    vi.stubEnv('WEB_SESSION_COOKIE_NAME', cookieName);
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it('returns 401 without session cookie', async () => {
    const response = await staffRoute(adminRequest('/api/admin/staff'));
    const payload = await response.json();

    expect(response.status).toBe(401);
    expect(payload.error.code).toBe('UNAUTHENTICATED');
  });

  it('rejects cashier before proxying', async () => {
    const backend = vi.fn();
    vi.stubGlobal('fetch', backend);

    const response = await staffRoute(adminRequest('/api/admin/staff', sessionCookie('cashier')));

    expect(response.status).toBe(403);
    expect(backend).not.toHaveBeenCalled();
  });

  it('rejects unknown query params', async () => {
    const backend = vi.fn();
    vi.stubGlobal('fetch', backend);

    const response = await staffRoute(adminRequest('/api/admin/staff?debug=true', sessionCookie()));
    const payload = await response.json();

    expect(response.status).toBe(422);
    expect(payload.error.code).toBe('VALIDATION_ERROR');
    expect(backend).not.toHaveBeenCalled();
  });

  it('proxies server-side and never returns token', async () => {
    const backend = vi.fn().mockResolvedValue(jsonResponse({ data: { staff: [{ id: 'staff-1', business_id: 'business-1', name: 'Rina', role: 'cashier' }] }, meta: { request_id: 'req-staff-test' } }));
    vi.stubGlobal('fetch', backend);

    const response = await staffRoute(adminRequest('/api/admin/staff?search=Rina&page=1&per_page=20', sessionCookie()));
    const payload = await response.json();
    const serialized = JSON.stringify(payload);
    const [backendUrl, backendInit] = backend.mock.calls[0] as [string, RequestInit];

    expect(response.status).toBe(200);
    expect(backendUrl).toBe('http://backend.test/api/v1/staff?search=Rina');
    expect((backendInit.headers as Headers).get('Authorization')).toBe(`Bearer ${ownerToken}`);
    expect(serialized).not.toContain(ownerToken);
    expect(serialized.toLowerCase()).not.toContain('bearer');
    expect(payload.data.rows[0].can_delete).toBe(false);
  });
});
