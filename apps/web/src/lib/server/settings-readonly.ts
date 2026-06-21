import { cookies } from 'next/headers';
import { NextRequest, NextResponse } from 'next/server';

import { settingsSummaryTable, subscriptionMetrics, subscriptionTable, type PreviewMetricFixture, type PreviewTableFixture } from '../../fixtures/preview';
import type { ApiEnvelope } from '../api/envelope';
import { BackendApiError, getBackendBaseUrl, toSafeErrorResponse } from './backend-client';
import { clearSessionCookie, readSessionFromRequest } from './admin-session';
import { getSessionCookieName, openSessionCookie, type WebAdminSessionPayload } from './session-cookie';

const SAFE_JSON_HEADERS = { 'Cache-Control': 'no-store' } as const;

export type SettingsKind = 'business' | 'outlets' | 'payments' | 'receipt';
export type SettingsReadData = { readonly meta: { readonly business_id: string; readonly generated_at: string; readonly source: 'backend' }; readonly section: string; readonly fields: readonly { readonly label: string; readonly value: string }[]; readonly can_edit: false; readonly data_notes: readonly { readonly key: string; readonly message: string; readonly severity: 'info' | 'warning' | 'danger' }[] };
export type SettingsPageModel = { readonly sourceLabel: string; readonly sourceTone: 'success' | 'warning' | 'danger' | 'info' | 'neutral'; readonly title: string; readonly description: string; readonly table: PreviewTableFixture; readonly dataNotes: readonly string[] };
export type SubscriptionReadData = { readonly meta: { readonly business_id: string; readonly generated_at: string; readonly source: 'backend' | 'fallback' }; readonly subscription: Record<string, unknown>; readonly data_notes: readonly { readonly key: string; readonly message: string; readonly severity: 'info' | 'warning' | 'danger' }[] };
export type SubscriptionPageModel = { readonly sourceLabel: string; readonly sourceTone: 'success' | 'warning' | 'danger' | 'info' | 'neutral'; readonly title: string; readonly description: string; readonly metrics: readonly PreviewMetricFixture[]; readonly table: PreviewTableFixture; readonly dataNotes: readonly string[] };

export async function handleAdminSettings(request: NextRequest, kind: SettingsKind): Promise<NextResponse> {
  const session = readSessionFromRequest(request);
  if (!session) {
    const response = unauthenticatedResponse();
    clearSessionCookie(response);
    return response;
  }
  if (!isTenantAdmin(session)) return forbiddenResponse('Settings tenant hanya untuk owner atau admin.');
  if ([...request.nextUrl.searchParams.keys()].length > 0) return validationResponse();

  try {
    const envelope = await fetchBackend(session.token, `/settings/${kind}`);
    return NextResponse.json(mapSettingsEnvelope(envelope, session, kind), { headers: SAFE_JSON_HEADERS });
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

export async function handleAdminSubscription(request: NextRequest): Promise<NextResponse> {
  const session = readSessionFromRequest(request);
  if (!session) {
    const response = unauthenticatedResponse();
    clearSessionCookie(response);
    return response;
  }
  if (!isTenantAdmin(session)) return forbiddenResponse('Subscription tenant hanya untuk owner atau admin.');
  if ([...request.nextUrl.searchParams.keys()].length > 0) return validationResponse();

  try {
    const envelope = await fetchBackend(session.token, '/subscription');
    return NextResponse.json(mapSubscriptionEnvelope(envelope, session), { headers: SAFE_JSON_HEADERS });
  } catch (error) {
    if (error instanceof BackendApiError && error.status === 401) {
      const response = unauthenticatedResponse();
      clearSessionCookie(response);
      return response;
    }
    if (error instanceof BackendApiError && error.status === 404) return NextResponse.json(fallbackSubscriptionEnvelope(session), { headers: SAFE_JSON_HEADERS });
    const safeError = toSafeErrorResponse(error);
    return NextResponse.json(safeError.body, { status: safeError.status, headers: SAFE_JSON_HEADERS });
  }
}

export async function getSettingsPageModel(): Promise<SettingsPageModel> {
  const session = await readSessionFromRequestCookies();
  if (!session) return previewSettings('Data contoh fallback', 'warning', 'Pengaturan membutuhkan sesi admin toko', 'Silakan login saat auth UI aktif.');
  if (!isTenantAdmin(session)) return previewSettings('Akses ditolak', 'danger', 'Pengaturan hanya untuk owner/admin', 'Role saat ini tidak boleh membuka settings tenant.');
  try {
    const mapped = mapSettingsEnvelope(await fetchBackend(session.token, '/settings/business'), session, 'business').data;
    return { sourceLabel: 'Data backend', sourceTone: 'success', title: 'Pengaturan read-only aktif', description: 'Settings dibaca server-side. Tombol simpan tetap disabled.', table: { columns: ['Field', 'Value'], rows: mapped.fields.map((field) => [field.label, field.value]), caption: 'Settings read-only' }, dataNotes: mapped.data_notes.map((note) => note.message) };
  } catch {
    return previewSettings('Backend belum tersedia', 'warning', 'Settings backend belum bisa dibaca', 'Data contoh fallback ditampilkan dengan label jelas.');
  }
}

export async function getSubscriptionPageModel(): Promise<SubscriptionPageModel> {
  const session = await readSessionFromRequestCookies();
  if (!session) return previewSubscription('Data contoh fallback', 'warning', 'Langganan membutuhkan sesi admin toko', 'Silakan login saat auth UI aktif.');
  if (!isTenantAdmin(session)) return previewSubscription('Akses ditolak', 'danger', 'Langganan hanya untuk owner/admin', 'Role saat ini tidak boleh membuka subscription tenant.');
  try {
    const mapped = mapSubscriptionEnvelope(await fetchBackend(session.token, '/subscription'), session).data;
    return { sourceLabel: mapped.meta.source === 'backend' ? 'Data backend' : 'Data fallback', sourceTone: mapped.meta.source === 'backend' ? 'success' : 'warning', title: 'Langganan read-only aktif', description: 'Plan dan billing changes tetap disabled.', metrics: subscriptionMetrics, table: subscriptionTable, dataNotes: mapped.data_notes.map((note) => note.message) };
  } catch {
    return previewSubscription('Backend belum tersedia', 'warning', 'Subscription backend belum bisa dibaca', 'Data contoh fallback ditampilkan dengan label jelas.');
  }
}

export function mapSettingsEnvelope(envelope: ApiEnvelope<Record<string, unknown>>, session: WebAdminSessionPayload, kind: SettingsKind): ApiEnvelope<SettingsReadData> {
  const data = envelope.data ?? {};
  const fields = Object.entries(data).filter(([key, value]) => !key.toLowerCase().includes('token') && !key.toLowerCase().includes('secret') && typeof value !== 'object').slice(0, 12).map(([key, value]) => ({ label: key.replaceAll('_', ' '), value: String(value ?? '-') }));
  return { data: { meta: { business_id: getSessionBusinessId(session), generated_at: new Date().toISOString(), source: 'backend' }, section: kind, fields, can_edit: false, data_notes: [{ key: 'settings_read_only', message: 'Settings ditampilkan read-only. Tombol simpan tetap disabled.', severity: 'info' }] }, meta: envelope.meta ?? { request_id: 'web-admin-settings' } };
}

export function mapSubscriptionEnvelope(envelope: ApiEnvelope<Record<string, unknown>>, session: WebAdminSessionPayload): ApiEnvelope<SubscriptionReadData> {
  return { data: { meta: { business_id: getSessionBusinessId(session), generated_at: new Date().toISOString(), source: 'backend' }, subscription: envelope.data ?? {}, data_notes: [{ key: 'subscription_read_only', message: 'Subscription ditampilkan read-only. Plan dan billing changes tetap disabled.', severity: 'info' }] }, meta: envelope.meta ?? { request_id: 'web-admin-subscription' } };
}

function fallbackSubscriptionEnvelope(session: WebAdminSessionPayload): ApiEnvelope<SubscriptionReadData> { return { data: { meta: { business_id: getSessionBusinessId(session), generated_at: new Date().toISOString(), source: 'fallback' }, subscription: { plan_name: 'Unknown', status: 'unknown', can_change_plan: false, can_manage_billing: false }, data_notes: [{ key: 'subscription_backend_gap', message: 'Backend subscription read route belum tersedia; fallback read-only ditampilkan.', severity: 'warning' }] }, meta: { request_id: 'web-admin-subscription-fallback' } }; }
async function fetchBackend(token: string, path: string): Promise<ApiEnvelope<Record<string, unknown>>> { const response = await fetch(`${getBackendBaseUrl()}${path}`, { method: 'GET', headers: new Headers({ Accept: 'application/json', Authorization: `Bearer ${token}` }), cache: 'no-store' }); const payload = await response.json().catch(() => null); if (!response.ok) throw toBackendError(response.status, payload); if (!payload || typeof payload !== 'object' || !('data' in payload)) throw new BackendApiError(500, 'INVALID_BACKEND_RESPONSE', 'Backend response tidak sesuai envelope.'); return payload as ApiEnvelope<Record<string, unknown>>; }
function toBackendError(status: number, payload: unknown): BackendApiError { const error = isRecord(payload) && isRecord(payload.error) ? payload.error : {}; return new BackendApiError(status, typeof error.code === 'string' ? error.code : 'BACKEND_ERROR', typeof error.message === 'string' ? error.message : 'Backend request gagal.', isRecord(error.details) ? error.details : {}); }
function previewSettings(sourceLabel: string, sourceTone: SettingsPageModel['sourceTone'], title: string, description: string): SettingsPageModel { return { sourceLabel, sourceTone, title, description, table: settingsSummaryTable, dataNotes: ['Data contoh preview, backend belum tersedia atau sesi belum aktif.'] }; }
function previewSubscription(sourceLabel: string, sourceTone: SubscriptionPageModel['sourceTone'], title: string, description: string): SubscriptionPageModel { return { sourceLabel, sourceTone, title, description, metrics: subscriptionMetrics, table: subscriptionTable, dataNotes: ['Data contoh preview, backend belum tersedia atau sesi belum aktif.'] }; }
function unauthenticatedResponse(): NextResponse { return NextResponse.json({ error: { code: 'UNAUTHENTICATED', message: 'Sesi admin tidak valid atau sudah berakhir.', details: {} } }, { status: 401, headers: SAFE_JSON_HEADERS }); }
function forbiddenResponse(message: string): NextResponse { return NextResponse.json({ error: { code: 'FORBIDDEN', message, details: {} } }, { status: 403, headers: SAFE_JSON_HEADERS }); }
function validationResponse(): NextResponse { return NextResponse.json({ error: { code: 'VALIDATION_ERROR', message: 'Filter settings tidak valid.', details: { query: ['Parameter tidak didukung.'] } } }, { status: 422, headers: SAFE_JSON_HEADERS }); }
function isTenantAdmin(session: WebAdminSessionPayload): boolean { return session.user.role === 'owner' || session.user.role === 'admin'; }
function getSessionBusinessId(session: WebAdminSessionPayload): string { return session.business?.id ?? session.user.business_id ?? 'unknown-business'; }
function isRecord(value: unknown): value is Record<string, unknown> { return typeof value === 'object' && value !== null && !Array.isArray(value); }
async function readSessionFromRequestCookies(): Promise<WebAdminSessionPayload | null> { const cookieStore = await cookies(); const sealed = cookieStore.get(getSessionCookieName())?.value; return sealed ? openSessionCookie(sealed) : null; }
