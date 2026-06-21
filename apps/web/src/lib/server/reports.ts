import { cookies } from 'next/headers';
import { NextRequest, NextResponse } from 'next/server';

import { reportLanes, reportResources } from '../../fixtures/preview';
import type { ApiEnvelope } from '../api/envelope';
import { BackendApiError, getBackendBaseUrl, toSafeErrorResponse } from './backend-client';
import { clearSessionCookie, readSessionFromRequest } from './admin-session';
import { getSessionCookieName, openSessionCookie, type WebAdminSessionPayload } from './session-cookie';

const SAFE_JSON_HEADERS = { 'Cache-Control': 'no-store' } as const;
const ALLOWED_PARAMS = new Set(['date_from', 'date_to', 'outlet_id', 'range']);
const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{12}$/i;
const RANGE_VALUES = new Set(['day', 'week', 'month']);

export type ReportKind = 'sales-summary' | 'top-products' | 'payment-methods' | 'cashier-shifts';
export type ReportsReadQuery = { readonly date_from?: string; readonly date_to?: string; readonly outlet_id?: string; readonly range?: 'day' | 'week' | 'month' };
export type ReportRow = Record<string, unknown>;
export type ReportsReadData = {
  readonly meta: { readonly business_id: string; readonly outlet_id: string | null; readonly currency: 'IDR'; readonly date_from: string; readonly date_to: string; readonly range: 'day' | 'week' | 'month'; readonly generated_at: string };
  readonly summary?: Record<string, unknown>;
  readonly chart?: readonly ReportRow[];
  readonly rows: readonly ReportRow[];
  readonly data_notes: readonly { readonly key: string; readonly message: string; readonly severity: 'info' | 'warning' | 'danger' }[];
};
export type ReportsPageModel = { readonly state: 'real' | 'preview' | 'session_required' | 'forbidden' | 'error'; readonly sourceLabel: string; readonly sourceTone: 'success' | 'warning' | 'danger' | 'info' | 'neutral'; readonly title: string; readonly description: string; readonly metrics: readonly { label: string; value: string; description: string }[]; readonly rows: readonly (readonly string[])[]; readonly dataNotes: readonly string[] };

export async function handleAdminReport(request: NextRequest, kind: ReportKind): Promise<NextResponse> {
  const session = readSessionFromRequest(request);
  if (!session) {
    const response = unauthenticatedResponse();
    clearSessionCookie(response);
    return response;
  }
  if (!isTenantAdmin(session)) return forbiddenResponse('Laporan toko hanya untuk owner atau admin.');

  try {
    const query = parseReportsReadQuery(request.nextUrl.searchParams);
    const envelope = await fetchBackendReport(session.token, kind, query);
    return NextResponse.json(mapReportEnvelope(envelope, query, session, kind), { headers: SAFE_JSON_HEADERS });
  } catch (error) {
    if (error instanceof BackendApiError && error.status === 401) {
      const response = unauthenticatedResponse();
      clearSessionCookie(response);
      return response;
    }
    const safeError = toSafeErrorResponse(error);
    return NextResponse.json(safeError.body, { status: safeError.status, headers: SAFE_JSON_HEADERS });
  }
}

export async function getReportsPageModel(): Promise<ReportsPageModel> {
  const session = await readSessionFromRequestCookies();
  if (!session) return previewModel('session_required', 'Data contoh fallback', 'warning', 'Laporan membutuhkan sesi admin toko', 'Silakan login saat auth UI aktif. Untuk sementara, daftar laporan memakai data contoh berlabel jelas.');
  if (!isTenantAdmin(session)) return previewModel('forbidden', 'Akses ditolak', 'danger', 'Laporan toko hanya untuk owner/admin', 'Role saat ini tidak boleh membuka laporan tenant. Tidak ada data backend toko yang ditampilkan.');

  try {
    const query: ReportsReadQuery = { range: 'day' };
    const sales = mapReportEnvelope(await fetchBackendReport(session.token, 'sales-summary', query), query, session, 'sales-summary').data;
    const payments = mapReportEnvelope(await fetchBackendReport(session.token, 'payment-methods', query), query, session, 'payment-methods').data;
    const topProducts = mapReportEnvelope(await fetchBackendReport(session.token, 'top-products', query), query, session, 'top-products').data;
    const shifts = mapReportEnvelope(await fetchBackendReport(session.token, 'cashier-shifts', query), query, session, 'cashier-shifts').data;
    return mapReportsToPageModel(sales, payments, topProducts, shifts);
  } catch (error) {
    if (error instanceof BackendApiError && error.status === 403) return previewModel('forbidden', 'Akses ditolak', 'danger', 'Backend menolak akses laporan', 'Role atau scope sesi ini tidak diizinkan membuka laporan tenant.');
    return previewModel('error', 'Backend belum tersedia', 'warning', 'Reports backend belum bisa dibaca', error instanceof BackendApiError ? error.message : 'Server admin belum bisa menghubungi backend laporan.');
  }
}

export function parseReportsReadQuery(searchParams: URLSearchParams): ReportsReadQuery {
  const unknown = [...searchParams.keys()].filter((key) => !ALLOWED_PARAMS.has(key));
  if (unknown.length > 0) throw validationError({ query: [`Parameter tidak didukung: ${unknown.join(', ')}`] });
  const dateFrom = searchParams.get('date_from') ?? undefined;
  const dateTo = searchParams.get('date_to') ?? undefined;
  const outletId = searchParams.get('outlet_id') ?? undefined;
  const range = searchParams.get('range') ?? undefined;
  if ((dateFrom && !dateTo) || (!dateFrom && dateTo)) throw validationError({ date_to: ['date_from dan date_to harus dikirim bersama.'] });
  if (dateFrom && !isIsoDate(dateFrom)) throw validationError({ date_from: ['date_from harus format YYYY-MM-DD.'] });
  if (dateTo && !isIsoDate(dateTo)) throw validationError({ date_to: ['date_to harus format YYYY-MM-DD.'] });
  if (dateFrom && dateTo && dateFrom > dateTo) throw validationError({ date_to: ['date_to harus setelah atau sama dengan date_from.'] });
  if (dateFrom && dateTo && daysBetween(dateFrom, dateTo) > 31) throw validationError({ date_to: ['Range laporan maksimal 31 hari.'] });
  if (outletId && !UUID_PATTERN.test(outletId)) throw validationError({ outlet_id: ['outlet_id harus UUID valid.'] });
  if (range && !RANGE_VALUES.has(range)) throw validationError({ range: ['range tidak didukung.'] });
  return { ...(dateFrom ? { date_from: dateFrom } : {}), ...(dateTo ? { date_to: dateTo } : {}), ...(outletId ? { outlet_id: outletId } : {}), range: (range as ReportsReadQuery['range'] | undefined) ?? 'day' };
}

export function mapReportEnvelope(envelope: ApiEnvelope<Record<string, unknown>>, query: ReportsReadQuery, session: WebAdminSessionPayload, kind: ReportKind): ApiEnvelope<ReportsReadData> {
  const data = envelope.data ?? {};
  const today = new Date().toISOString().slice(0, 10);
  const rows = pickRows(data, kind);
  const summary = typeof data.summary === 'object' && data.summary !== null ? data.summary as Record<string, unknown> : kind === 'sales-summary' ? pickSalesSummary(data) : undefined;
  const chart = Array.isArray(data.chart) ? data.chart.filter(isRecord) : [];
  return {
    data: {
      meta: { business_id: getSessionBusinessId(session), outlet_id: query.outlet_id ?? null, currency: 'IDR', date_from: query.date_from ?? today, date_to: query.date_to ?? query.date_from ?? today, range: query.range ?? 'day', generated_at: new Date().toISOString() },
      ...(summary ? { summary } : {}),
      ...(chart.length > 0 ? { chart } : {}),
      rows,
      data_notes: [{ key: 'reports_read_only', message: 'Laporan dibaca server-side; export dan aksi sensitif tetap disabled.', severity: 'info' }],
    },
    meta: envelope.meta ?? { request_id: 'web-admin-reports' },
  };
}

function mapReportsToPageModel(sales: ReportsReadData, payments: ReportsReadData, topProducts: ReportsReadData, shifts: ReportsReadData): ReportsPageModel {
  const summary = sales.summary ?? {};
  return {
    state: 'real', sourceLabel: 'Data backend', sourceTone: 'success', title: 'Laporan read-only aktif', description: 'Ringkasan laporan dibaca melalui BFF. Export dan aksi sensitif tetap dikunci.',
    metrics: [
      { label: 'Gross sales', value: formatRupiah(toNumber(summary.gross_sales ?? summary.total_sales)), description: 'Dari backend' },
      { label: 'Net sales', value: formatRupiah(toNumber(summary.net_sales ?? summary.sales)), description: 'Server-side' },
      { label: 'Transaksi', value: String(toNumber(summary.transaction_count)), description: 'Tidak dihitung di client' },
      { label: 'Metode bayar', value: String(payments.rows.length), description: 'Read-only' },
    ],
    rows: [
      ['Produk terlaris', String(topProducts.rows.length), topProducts.rows[0] ? String(topProducts.rows[0].product_name ?? topProducts.rows[0].name ?? '-') : 'Belum ada'],
      ['Metode pembayaran', String(payments.rows.length), payments.rows[0] ? String(payments.rows[0].label ?? payments.rows[0].method ?? '-') : 'Belum ada'],
      ['Shift kasir', String(shifts.rows.length), shifts.rows[0] ? String((shifts.rows[0].cashier as { name?: string } | undefined)?.name ?? shifts.rows[0].cashier_name ?? '-') : 'Belum ada'],
    ],
    dataNotes: [...sales.data_notes, ...payments.data_notes, ...topProducts.data_notes, ...shifts.data_notes].map((note) => note.message),
  };
}

async function fetchBackendReport(token: string, kind: ReportKind, query: ReportsReadQuery): Promise<ApiEnvelope<Record<string, unknown>>> {
  const response = await fetch(`${getBackendBaseUrl()}${backendReportPath(kind, query)}`, { method: 'GET', headers: backendHeaders(token), cache: 'no-store' });
  const payload = await response.json().catch(() => null);
  if (!response.ok) throw toBackendError(response.status, payload);
  if (!payload || typeof payload !== 'object' || !('data' in payload)) throw new BackendApiError(500, 'INVALID_BACKEND_RESPONSE', 'Backend report response tidak sesuai envelope.');
  return payload as ApiEnvelope<Record<string, unknown>>;
}

function backendReportPath(kind: ReportKind, query: ReportsReadQuery): string {
  const path = kind === 'top-products' ? '/reports/sold-products' : `/reports/${kind}`;
  const params = new URLSearchParams();
  if (query.date_from) params.set('date_from', query.date_from);
  if (query.date_to) params.set('date_to', query.date_to);
  if (query.outlet_id) params.set('outlet_id', query.outlet_id);
  if (query.range) params.set('range', query.range);
  const serialized = params.toString();
  return serialized ? `${path}?${serialized}` : path;
}

function backendHeaders(token: string): Headers {
  return new Headers({ Accept: 'application/json', 'Content-Type': 'application/json', Authorization: `Bearer ${token}` });
}

function toBackendError(status: number, payload: unknown): BackendApiError {
  const error = isRecord(payload) && isRecord(payload.error) ? payload.error : {};
  return new BackendApiError(status, typeof error.code === 'string' ? error.code : 'BACKEND_ERROR', typeof error.message === 'string' ? error.message : 'Backend report request gagal.', isRecord(error.details) ? error.details : {});
}

function pickRows(data: Record<string, unknown>, kind: ReportKind): readonly ReportRow[] {
  const candidates = kind === 'sales-summary' ? [data.rows, data.transactions] : [data.rows, data.items, data.products, data.payment_methods, data.methods, data.shifts, data.data];
  for (const candidate of candidates) if (Array.isArray(candidate)) return candidate.filter(isRecord);
  return [];
}

function pickSalesSummary(data: Record<string, unknown>): Record<string, unknown> {
  return { gross_sales: toNumber(data.gross_sales ?? data.total_sales), net_sales: toNumber(data.net_sales ?? data.sales), transaction_count: toNumber(data.transaction_count), average_transaction: toNumber(data.average_transaction), discount_total: toNumber(data.discount_total) };
}

function previewModel(state: ReportsPageModel['state'], sourceLabel: string, sourceTone: ReportsPageModel['sourceTone'], title: string, description: string): ReportsPageModel {
  return { state, sourceLabel, sourceTone, title, description, metrics: [{ label: 'Sales', value: 'Preview', description: 'Data contoh' }, { label: 'Produk', value: 'Preview', description: 'Data contoh' }, { label: 'Metode bayar', value: 'Preview', description: 'Data contoh' }, { label: 'Export', value: 'Disabled', description: 'Read-only' }], rows: reportResources.map((resource) => [resource.label, resource.path, 'Data contoh fallback']), dataNotes: ['Data contoh preview, backend belum tersedia atau sesi belum aktif.', ...reportLanes.flatMap((lane) => lane.items)] };
}

function unauthenticatedResponse(): NextResponse { return NextResponse.json({ error: { code: 'UNAUTHENTICATED', message: 'Sesi admin tidak valid atau sudah berakhir.', details: {} } }, { status: 401, headers: SAFE_JSON_HEADERS }); }
function forbiddenResponse(message: string): NextResponse { return NextResponse.json({ error: { code: 'FORBIDDEN', message, details: {} } }, { status: 403, headers: SAFE_JSON_HEADERS }); }
function validationError(details: Record<string, readonly string[]>): BackendApiError { return new BackendApiError(422, 'VALIDATION_ERROR', 'Filter laporan tidak valid.', details); }
function isTenantAdmin(session: WebAdminSessionPayload): boolean { return session.user.role === 'owner' || session.user.role === 'admin'; }
function getSessionBusinessId(session: WebAdminSessionPayload): string { return session.business?.id ?? session.user.business_id ?? 'unknown-business'; }
function isIsoDate(value: string): boolean { return /^\d{4}-\d{2}-\d{2}$/.test(value); }
function daysBetween(from: string, to: string): number { return Math.floor((Date.parse(`${to}T00:00:00Z`) - Date.parse(`${from}T00:00:00Z`)) / 86400000) + 1; }
function isRecord(value: unknown): value is Record<string, unknown> { return typeof value === 'object' && value !== null && !Array.isArray(value); }
function toNumber(value: unknown): number { return typeof value === 'number' && Number.isFinite(value) ? value : 0; }
function formatRupiah(value: number): string { return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value); }
async function readSessionFromRequestCookies(): Promise<WebAdminSessionPayload | null> { const cookieStore = await cookies(); const sealed = cookieStore.get(getSessionCookieName())?.value; return sealed ? openSessionCookie(sealed) : null; }
