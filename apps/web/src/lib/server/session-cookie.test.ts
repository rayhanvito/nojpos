import { describe, expect, it, vi } from 'vitest';
import {
  assertNoTokenInSafeContext,
  createSessionPayload,
  getSessionCookieOptions,
  openSessionCookie,
  sealSessionCookie,
  toSafeSessionContext,
} from './session-cookie';

const secret = 'test-secret-for-web-admin-session-cookie-32';

vi.stubEnv('WEB_SESSION_SECRET', secret);
vi.stubEnv('WEB_SESSION_TTL_SECONDS', '3600');

describe('Web Admin sealed session cookie', () => {
  it('seals token material without plaintext browser-readable content', () => {
    const payload = createSessionPayload({
      token: 'plain-backend-token-value',
      user: {
        id: 'user-1',
        name: 'Owner Toko',
        role: 'owner',
        business_id: 'business-1',
      },
      business: {
        id: 'business-1',
        name: 'Kopi Senja',
      },
      device_uuid: 'web-admin:test-device',
      now: new Date(),
    });

    const sealed = sealSessionCookie(payload, secret);

    expect(sealed).not.toContain('plain-backend-token-value');
    expect(sealed).not.toContain('Owner Toko');
    expect(openSessionCookie(sealed, secret)).toEqual(payload);
  });

  it('returns safe context without token fields', () => {
    const safeContext = toSafeSessionContext({
      user: {
        id: 'user-1',
        name: 'Admin Toko',
        role: 'admin',
        business_id: 'business-1',
      },
      business: {
        id: 'business-1',
        name: 'Kopi Senja',
      },
      outlets: [
        { id: 'outlet-1', business_id: 'business-1', name: 'Cabang Utama', timezone: 'Asia/Jakarta' },
        { id: 'outlet-other', business_id: 'business-other', name: 'Cabang Tenant Lain', timezone: 'Asia/Jakarta' },
      ],
      permissions: ['admin'],
    });

    expect(assertNoTokenInSafeContext(safeContext)).toBe(true);
    expect(JSON.stringify(safeContext)).not.toContain('plain-backend-token-value');
    expect(safeContext.access).toEqual({ tenant_admin: true, platform_admin: false });
    expect(safeContext.outlets).toEqual([
      { id: 'outlet-1', business_id: 'business-1', name: 'Cabang Utama', timezone: 'Asia/Jakarta' },
    ]);
    expect(safeContext.redirect_to).toBe('/dashboard');
  });

  it('uses HttpOnly Lax cookie settings with finite expiry', () => {
    vi.stubEnv('WEB_SESSION_COOKIE_NAME', 'nojpos_test_session');

    expect(getSessionCookieOptions(new Date('2026-06-21T00:00:00.000Z'))).toMatchObject({
      name: 'nojpos_test_session',
      httpOnly: true,
      sameSite: 'lax',
      path: '/',
      maxAgeSeconds: 3600,
    });
  });
});
