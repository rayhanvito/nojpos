import { NextRequest } from 'next/server';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createSessionPayload, sealSessionCookie } from '../../../../../lib/server/session-cookie';
import { GET as businessSettingsRoute } from './route';

const secret = 'test-secret-for-web-admin-session-cookie-32';
const cookieName = 'nojpos_test_session';
const ownerToken = 'owner-settings-token';

function adminRequest(path: string, cookie?: string): NextRequest {
  return new NextRequest(`http://localhost:3000${path}`, {
    method: 'GET',
    headers: { host: 'localhost:3000', ...(cookie ? { cookie } : {}) },
  });
}

function sessionCookie(role: 'owner' | 'cashier' = 'owner'): string {
  const sealed = sealSessionCookie(
    createSessionPayload({
      token: role === 'owner' ? ownerToken : 'cashier-backend-token',
      user: { id: `${role}-user`, name: `${role} User`, role, business_id: 'business-1', is_superadmin: false },
      business: { id: 'business-1', name: 'Kopi Senja' },
      device_uuid: 'web-device-1',
    }),
  );
  return `${cookieName}=${sealed}`;
}

function jsonResponse(payload: unknown, status = 200): Response {
  return new Response(JSON.stringify(payload), { status, headers: { 'Content-Type': 'application/json' } });
}

describe('Web Admin BFF settings route', () => {
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
    const response = await businessSettingsRoute(adminRequest('/api/admin/settings/business'));
    const payload = await response.json();

    expect(response.status).toBe(401);
    expect(payload.error.code).toBe('UNAUTHENTICATED');
  });

  it('rejects cashier before proxying to backend', async () => {
    const backend = vi.fn();
    vi.stubGlobal('fetch', backend);

    const response = await businessSettingsRoute(adminRequest('/api/admin/settings/business', sessionCookie('cashier')));
    const payload = await response.json();

    expect(response.status).toBe(403);
    expect(payload.error.code).toBe('FORBIDDEN');
    expect(backend).not.toHaveBeenCalled();
  });

  it('rejects unknown query params', async () => {
    const backend = vi.fn();
    vi.stubGlobal('fetch', backend);

    const response = await businessSettingsRoute(adminRequest('/api/admin/settings/business?debug=true', sessionCookie()));
    const payload = await response.json();

    expect(response.status).toBe(422);
    expect(payload.error.code).toBe('VALIDATION_ERROR');
    expect(backend).not.toHaveBeenCalled();
  });

  it('proxies server-side and never returns session material', async () => {
    const backend = vi.fn().mockResolvedValue(jsonResponse({ data: { name: 'Kopi Senja', timezone: 'Asia/Jakarta' }, meta: { request_id: 'req-settings-test' } }));
    vi.stubGlobal('fetch', backend);

    const response = await businessSettingsRoute(adminRequest('/api/admin/settings/business', sessionCookie()));
    const payload = await response.json();
    const serializedPayload = JSON.stringify(payload);
    const [backendUrl, backendInit] = backend.mock.calls[0] as [string, RequestInit];

    expect(response.status).toBe(200);
    expect(backendUrl).toBe('http://backend.test/api/v1/settings/business');
    expect((backendInit.headers as Headers).get('Authorization')).toBe(`Bearer ${ownerToken}`);
    expect(serializedPayload).not.toContain(ownerToken);
    expect(serializedPayload.toLowerCase()).not.toContain('bearer');
    expect(payload.data.can_edit).toBe(false);
  });
});
