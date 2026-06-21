import { cookies } from 'next/headers';
import { NextRequest, NextResponse } from 'next/server';

import type { PreviewMetricFixture, PreviewTableFixture } from '../../fixtures/preview';
import type { ApiEnvelope } from '../api/envelope';
import { BackendApiError, getBackendBaseUrl, toSafeErrorResponse } from './backend-client';
import { clearSessionCookie, readSessionFromRequest } from './admin-session';
import { getSessionCookieName, openSessionCookie, type WebAdminSessionPayload } from './session-cookie';

const SAFE_JSON_HEADERS = { 'Cache-Control': 'no-store' } as const;
const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{12}$/i;

export type ReadOnlyListQuery = { readonly search?: string; readonly status?: string; readonly role?: string; readonly group?: string; readonly outlet_id?: string; readonly page?: number; readonly per_page?: number };
type MutableReadOnlyListQuery = { search?: string; status?: string; role?: string; group?: string; outlet_id?: string; page?: number; per_page?: number };
export type ReadOnlyRow = Record<string, unknown>;
export type ReadOnlyListData = {
  readonly meta: { readonly business_id: string; readonly outlet_id: string | null; readonly generated_at: string; readonly source: 'backend' };
  readonly pagination: { readonly page: number; readonly per_page: number; readonly total: number; readonly has_next_page: boolean };
  readonly totals: Record<string, number>;
  readonly rows: readonly ReadOnlyRow[];
  readonly data_notes: readonly { readonly key: string; readonly message: string; readonly severity: 'info' | 'warning' | 'danger' }[];
};
export type ReadOnlyPageModel = { readonly sourceLabel: string; readonly sourceTone: 'success' | 'warning' | 'danger' | 'info' | 'neutral'; readonly title: string; readonly description: string; readonly metrics: readonly PreviewMetricFixture[]; readonly table: PreviewTableFixture; readonly dataNotes: readonly string[] };

export type ReadOnlyListConfig = {
  readonly resource: 'staff' | 'customers' | 'outlets';
  readonly backendPath: string;
  readonly allowedParams: readonly string[];
  readonly statusValues?: readonly string[];
  readonly title: string;
  readonly fallbackDescription: string;
  readonly columns: readonly string[];
  readonly rowMapper: (row: ReadOnlyRow, sessionBusinessId: string) => ReadOnlyRow | null;
  readonly tableMapper: (rows: readonly ReadOnlyRow[]) => readonly (readonly string[])[];
  readonly totalsMapper: (rows: readonly ReadOnlyRow[]) => Record<string, number>;
  readonly note: string;
  readonly fixtureTable: PreviewTableFixture;
  readonly fixtureMetrics: readonly PreviewMetricFixture[];
};

export async function handleReadOnlyList(request: NextRequest, config: ReadOnlyListConfig): Promise<NextResponse> {
  const session = readSessionFromRequest(request);
  if (!session) {
    const response = unauthenticatedResponse();
    clearSessionCookie(response);
    return response;
  }
  if (!isTenantAdmin(session)) return forbiddenResponse(`${config.title} hanya untuk owner atau admin.`);

  try {
    const query = parseReadOnlyQuery(request.nextUrl.searchParams, config);
    const envelope = await fetchBackendList(session.token, config, query);
    return NextResponse.json(mapReadOnlyEnvelope(envelope, query, session, config), { headers: SAFE_JSON_HEADERS });
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

export async function getReadOnlyPageModel(config: ReadOnlyListConfig): Promise<ReadOnlyPageModel> {
  const session = await readSessionFromRequestCookies();
  if (!session) return previewModel(config, 'Data contoh fallback', 'warning', `${config.title} membutuhkan sesi admin toko`, 'Silakan login saat auth UI aktif. Data contoh diberi label jelas.');
  if (!isTenantAdmin(session)) return previewModel(config, 'Akses ditolak', 'danger', `${config.title} hanya untuk owner/admin`, 'Role saat ini tidak boleh membuka data tenant.');

  try {
    const query: ReadOnlyListQuery = { page: 1, per_page: 20, status: 'all' };
    const data = mapReadOnlyEnvelope(await fetchBackendList(session.token, config, query), query, session, config).data;
    return {
      sourceLabel: 'Data backend', sourceTone: 'success', title: `${config.title} read-only aktif`, description: config.note,
      metrics: Object.entries(data.totals).slice(0, 4).map(([label, value]) => ({ label: label.replaceAll('_', ' '), value: String(value), description: 'Dari backend' })),
      table: { columns: [...config.columns], rows: data.rows.length > 0 ? config.tableMapper(data.rows) : [], caption: config.note, primaryColumn: 0 },
      dataNotes: data.data_notes.map((note) => note.message),
    };
  } catch (error) {
    if (error instanceof BackendApiError && error.status === 403) return previewModel(config, 'Akses ditolak', 'danger', `Backend menolak akses ${config.title}`, 'Role atau scope sesi ini tidak diizinkan membuka data tenant.');
    return previewModel(config, 'Backend belum tersedia', 'warning', `${config.title} backend belum bisa dibaca`, error instanceof BackendApiError ? error.message : config.fallbackDescription);
  }
}

export function parseReadOnlyQuery(searchParams: URLSearchParams, config: ReadOnlyListConfig): ReadOnlyListQuery {
  const allowed = new Set(config.allowedParams);
  const unknown = [...searchParams.keys()].filter((key) => !allowed.has(key));
  if (unknown.length > 0) throw validationError({ query: [`Parameter tidak didukung: ${unknown.join(', ')}`] });
  const query: MutableReadOnlyListQuery = {};
  const search = searchParams.get('search')?.trim();
  if (search) query.search = search.slice(0, 80);
  const status = searchParams.get('status') ?? undefined;
  if (status) {
    if (config.statusValues && !config.statusValues.includes(status)) throw validationError({ status: ['status tidak didukung.'] });
    query.status = status;
  }
  const role = searchParams.get('role') ?? undefined;
  if (role) query.role = role;
  const group = searchParams.get('group') ?? undefined;
  if (group) query.group = group;
  const outletId = searchParams.get('outlet_id') ?? undefined;
  if (outletId) {
    if (!UUID_PATTERN.test(outletId)) throw validationError({ outlet_id: ['outlet_id harus UUID valid.'] });
    query.outlet_id = outletId;
  }
  const page = parsePositiveInt(searchParams.get('page'), 1, 1, 999);
  const perPage = parsePositiveInt(searchParams.get('per_page'), 20, 1, 50);
  query.page = page;
  query.per_page = perPage;
  return query;
}

export function mapReadOnlyEnvelope(envelope: ApiEnvelope<Record<string, unknown>>, query: ReadOnlyListQuery, session: WebAdminSessionPayload, config: ReadOnlyListConfig): ApiEnvelope<ReadOnlyListData> {
  const data = envelope.data ?? {};
  const sessionBusinessId = getSessionBusinessId(session);
  const rows = extractRows(data, config.resource).map((row) => config.rowMapper(row, sessionBusinessId)).filter((row): row is ReadOnlyRow => row !== null);
  const page = query.page ?? 1;
  const perPage = query.per_page ?? 20;
  return {
    data: {
      meta: { business_id: sessionBusinessId, outlet_id: query.outlet_id ?? null, generated_at: new Date().toISOString(), source: 'backend' },
      pagination: { page, per_page: perPage, total: rows.length, has_next_page: rows.length >= perPage },
      totals: config.totalsMapper(rows),
      rows,
      data_notes: [{ key: `${config.resource}_read_only`, message: config.note, severity: 'info' }],
    },
    meta: envelope.meta ?? { request_id: `web-admin-${config.resource}` },
  };
}

async function fetchBackendList(token: string, config: ReadOnlyListConfig, query: ReadOnlyListQuery): Promise<ApiEnvelope<Record<string, unknown>>> {
  const response = await fetch(`${getBackendBaseUrl()}${backendListPath(config.backendPath, query)}`, { method: 'GET', headers: backendHeaders(token), cache: 'no-store' });
  const payload = await response.json().catch(() => null);
  if (!response.ok) throw toBackendError(response.status, payload);
  if (!payload || typeof payload !== 'object' || !('data' in payload)) throw new BackendApiError(500, 'INVALID_BACKEND_RESPONSE', 'Backend response tidak sesuai envelope.');
  return payload as ApiEnvelope<Record<string, unknown>>;
}

function backendListPath(path: string, query: ReadOnlyListQuery): string {
  const params = new URLSearchParams();
  if (query.search) params.set('search', query.search);
  if (query.status && query.status !== 'all') params.set('status', query.status);
  if (query.role) params.set('role', query.role);
  if (query.group) params.set('group', query.group);
  if (query.outlet_id) params.set('outlet_id', query.outlet_id);
  const serialized = params.toString();
  return serialized ? `${path}?${serialized}` : path;
}

function extractRows(data: Record<string, unknown>, resource: ReadOnlyListConfig['resource']): readonly ReadOnlyRow[] {
  const candidates = [data[resource], data.rows, data.items, data.data, data.outlets];
  for (const candidate of candidates) if (Array.isArray(candidate)) return candidate.filter(isRecord);
  return [];
}

function previewModel(config: ReadOnlyListConfig, sourceLabel: string, sourceTone: ReadOnlyPageModel['sourceTone'], title: string, description: string): ReadOnlyPageModel {
  return { sourceLabel, sourceTone, title, description, metrics: config.fixtureMetrics, table: config.fixtureTable, dataNotes: ['Data contoh preview, backend belum tersedia atau sesi belum aktif.', config.note] };
}

async function readSessionFromRequestCookies(): Promise<WebAdminSessionPayload | null> { const cookieStore = await cookies(); const sealed = cookieStore.get(getSessionCookieName())?.value; return sealed ? openSessionCookie(sealed) : null; }
function backendHeaders(token: string): Headers { return new Headers({ Accept: 'application/json', 'Content-Type': 'application/json', Authorization: `Bearer ${token}` }); }
function toBackendError(status: number, payload: unknown): BackendApiError { const error = isRecord(payload) && isRecord(payload.error) ? payload.error : {}; return new BackendApiError(status, typeof error.code === 'string' ? error.code : 'BACKEND_ERROR', typeof error.message === 'string' ? error.message : 'Backend request gagal.', isRecord(error.details) ? error.details : {}); }
function validationError(details: Record<string, readonly string[]>): BackendApiError { return new BackendApiError(422, 'VALIDATION_ERROR', 'Filter read-only tidak valid.', details); }
function unauthenticatedResponse(): NextResponse { return NextResponse.json({ error: { code: 'UNAUTHENTICATED', message: 'Sesi admin tidak valid atau sudah berakhir.', details: {} } }, { status: 401, headers: SAFE_JSON_HEADERS }); }
function forbiddenResponse(message: string): NextResponse { return NextResponse.json({ error: { code: 'FORBIDDEN', message, details: {} } }, { status: 403, headers: SAFE_JSON_HEADERS }); }
function isTenantAdmin(session: WebAdminSessionPayload): boolean { return session.user.role === 'owner' || session.user.role === 'admin'; }
function getSessionBusinessId(session: WebAdminSessionPayload): string { return session.business?.id ?? session.user.business_id ?? 'unknown-business'; }
function parsePositiveInt(value: string | null, fallback: number, min: number, max: number): number { if (!value) return fallback; const parsed = Number(value); if (!Number.isInteger(parsed) || parsed < min || parsed > max) throw validationError({ pagination: [`Nilai harus ${min}-${max}.`] }); return parsed; }
function isRecord(value: unknown): value is Record<string, unknown> { return typeof value === 'object' && value !== null && !Array.isArray(value); }
