import { NextRequest } from 'next/server';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { POST as loginRoute } from './login/route';
import { POST as logoutRoute } from './logout/route';
import { GET as sessionRoute } from '../session/route';

const secret = 'test-secret-for-web-admin-session-cookie-32';

function adminRequest(
  path: string,
  init?: {
    readonly method?: string;
    readonly body?: BodyInit | null;
    readonly headers?: Record<string, string>;
  },
): NextRequest {
  return new NextRequest(`http://localhost:3000${path}`, {
    method: init?.method,
    body: init?.body,
    headers: {
      host: 'localhost:3000',
      origin: 'http://localhost:3000',
      ...(init?.headers ?? {}),
    },
  });
}

function jsonResponse(payload: unknown, status = 200): Response {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('Web Admin BFF session routes', () => {
  beforeEach(() => {
    vi.stubEnv('BACKEND_INTERNAL_API_URL', 'http://backend.test/api/v1');
    vi.stubEnv('WEB_SESSION_SECRET', secret);
    vi.stubEnv('WEB_SESSION_COOKIE_NAME', 'nojpos_test_session');
    vi.stubEnv('WEB_SESSION_TTL_SECONDS', '3600');
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it('login route creates a safe session response without returning backend token', async () => {
    const backend = vi.fn().mockResolvedValue(
      jsonResponse({
        data: {
          token: 'plain-backend-token-value',
          user: {
            id: 'user-1',
            name: 'Owner Toko',
            email: 'owner@example.test',
            role: 'owner',
            business_id: 'business-1',
          },
          business: {
            id: 'business-1',
            name: 'Kopi Senja',
          },
          outlets: [{ id: 'outlet-1', business_id: 'business-1', name: 'Cabang Utama', timezone: 'Asia/Jakarta' }],
        },
        meta: {},
      }),
    );
    vi.stubGlobal('fetch', backend);

    const response = await loginRoute(
      adminRequest('/api/admin/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email: 'owner@example.test', password: 'secret' }),
      }),
    );
    const payload = await response.json();
    const serializedPayload = JSON.stringify(payload);

    expect(response.status).toBe(200);
    expect(serializedPayload).not.toContain('plain-backend-token-value');
    expect(serializedPayload.toLowerCase()).not.toContain('token');
    expect(payload.data.redirect_to).toBe('/dashboard');
    expect(response.headers.get('set-cookie')).toContain('HttpOnly');
    expect(response.headers.get('set-cookie')).not.toContain('plain-backend-token-value');
  });

  it('login route rejects cashier role before creating a web admin session', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse({
          data: {
            token: 'cashier-token',
            user: {
              id: 'user-2',
              name: 'Kasir',
              role: 'cashier',
              business_id: 'business-1',
            },
            business: {
              id: 'business-1',
              name: 'Kopi Senja',
            },
            outlets: [],
          },
          meta: {},
        }),
      ),
    );

    const response = await loginRoute(
      adminRequest('/api/admin/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email: 'cashier@example.test', password: 'secret' }),
      }),
    );
    const payload = await response.json();

    expect(response.status).toBe(403);
    expect(payload.error.code).toBe('FORBIDDEN');
    expect(response.headers.get('set-cookie')).toBeNull();
  });

  it('logout route clears cookie even when no backend session is present', async () => {
    const response = await logoutRoute(adminRequest('/api/admin/auth/logout', { method: 'POST' }));
    const payload = await response.json();

    expect(response.status).toBe(200);
    expect(payload).toEqual({ data: { logged_out: true }, meta: {} });
    expect(response.headers.get('set-cookie')).toContain('Max-Age=0');
  });

  it('session route returns 401 and clears cookie without a session cookie', async () => {
    const response = await sessionRoute(adminRequest('/api/admin/session'));
    const payload = await response.json();

    expect(response.status).toBe(401);
    expect(payload.error.code).toBe('UNAUTHENTICATED');
    expect(response.headers.get('set-cookie')).toContain('Max-Age=0');
  });
});
