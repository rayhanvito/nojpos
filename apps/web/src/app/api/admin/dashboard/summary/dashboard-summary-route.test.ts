import { NextRequest } from 'next/server';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createSessionPayload, sealSessionCookie } from '../../../../../lib/server/session-cookie';
import { GET as dashboardSummaryRoute } from './route';

const secret = 'test-secret-for-web-admin-session-cookie-32';
const cookieName = 'nojpos_test_session';
const ownerToken = 'owner-backend-token';

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

function dashboardSummaryPayload() {
  return {
    data: {
      meta: {
        date: '2026-06-21',
        range: 'last_7_days',
        timezone: 'Asia/Jakarta',
        business_id: 'business-1',
        outlet_id: null,
        outlet_name: null,
        generated_at: '2026-06-21T08:30:00.000000Z',
        currency: 'IDR',
        money_format: 'integer_rupiah',
        data_status: 'real',
        is_empty_today: false,
      },
      kpis: {
        sales_today: { value: 2450000, type: 'money', label: 'Penjualan hari ini', trend_label: '+8%', is_estimate: false },
        transaction_count: { value: 86, type: 'integer', label: 'Jumlah transaksi', trend_label: null, is_estimate: false },
        average_transaction: { value: 28500, type: 'money', label: 'Rata-rata transaksi', trend_label: null, is_estimate: false },
        gross_profit_estimate: { value: 0, type: 'money', label: 'Gross profit estimasi', trend_label: null, is_estimate: true },
        low_stock_count: { value: 1, type: 'integer', label: 'Stok hampir habis', trend_label: 'Butuh cek', is_estimate: false },
        cash_difference: { value: 0, type: 'money', label: 'Selisih kas', trend_label: null, is_estimate: true },
      },
      alerts: [],
      sales_last_7_days: [],
      payment_methods: [],
      top_products: [],
      low_stock_items: [],
      recent_transactions: [],
      cashier_performance: [],
      branch_highlights: [],
      data_notes: [{ key: 'gross_profit_estimate', message: 'Gross profit butuh data cost lengkap.', severity: 'info' }],
    },
    meta: { request_id: 'req-test' },
  };
}

describe('Web Admin BFF dashboard summary route', () => {
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
    const response = await dashboardSummaryRoute(adminRequest('/api/admin/dashboard/summary'));
    const payload = await response.json();

    expect(response.status).toBe(401);
    expect(payload.error.code).toBe('UNAUTHENTICATED');
    expect(response.headers.get('set-cookie')).toContain('Max-Age=0');
  });

  it('rejects non-tenant admin roles before proxying to backend', async () => {
    const backend = vi.fn();
    vi.stubGlobal('fetch', backend);

    const response = await dashboardSummaryRoute(adminRequest('/api/admin/dashboard/summary', sessionCookie('superadmin')));
    const payload = await response.json();

    expect(response.status).toBe(403);
    expect(payload.error.code).toBe('FORBIDDEN');
    expect(backend).not.toHaveBeenCalled();
  });

  it('does not forward unknown query params to backend', async () => {
    const backend = vi.fn();
    vi.stubGlobal('fetch', backend);

    const response = await dashboardSummaryRoute(adminRequest('/api/admin/dashboard/summary?debug=true', sessionCookie()));
    const payload = await response.json();

    expect(response.status).toBe(422);
    expect(payload.error.code).toBe('VALIDATION_ERROR');
    expect(backend).not.toHaveBeenCalled();
  });

  it('proxies allowed query params server-side and never returns the backend token', async () => {
    const backend = vi.fn().mockResolvedValue(jsonResponse(dashboardSummaryPayload()));
    vi.stubGlobal('fetch', backend);

    const response = await dashboardSummaryRoute(
      adminRequest('/api/admin/dashboard/summary?date=2026-06-21&range=last_7_days', sessionCookie()),
    );
    const payload = await response.json();
    const serializedPayload = JSON.stringify(payload);
    const [backendUrl, backendInit] = backend.mock.calls[0] as [string, RequestInit];

    expect(response.status).toBe(200);
    expect(backendUrl).toBe('http://backend.test/api/v1/dashboard/summary?date=2026-06-21&range=last_7_days');
    expect(backendInit.headers).toBeInstanceOf(Headers);
    expect((backendInit.headers as Headers).get('Authorization')).toBe(`Bearer ${ownerToken}`);
    expect(serializedPayload).not.toContain(ownerToken);
    expect(serializedPayload.toLowerCase()).not.toContain('bearer');
    expect(payload.data).toHaveProperty('kpis');
    expect(payload.data).toHaveProperty('data_notes');
  });
});
