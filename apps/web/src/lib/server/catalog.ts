import { cookies } from 'next/headers';
import { NextRequest, NextResponse } from 'next/server';

import { catalogDetailLinks, catalogLanes, catalogProductsTable } from '../../fixtures/preview';
import type { PreviewLane, PreviewMetricFixture, PreviewTableFixture } from '../../fixtures/preview';
import type { ApiEnvelope } from '../api/envelope';
import {
  BackendApiError,
  createBackendClient,
  toSafeErrorResponse,
  type BackendCatalogProductsData,
  type BackendCatalogProduct,
  type CatalogReadQuery,
} from './backend-client';
import { clearSessionCookie, readSessionFromRequest } from './admin-session';
import { getSessionCookieName, openSessionCookie, type WebAdminSessionPayload } from './session-cookie';

const SAFE_JSON_HEADERS = { 'Cache-Control': 'no-store' } as const;
const ALLOWED_PARAMS = new Set(['search', 'category_id', 'outlet_id', 'status', 'page', 'per_page']);
const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{12}$/i;
const STATUSES = new Set(['all', 'active', 'inactive', 'archived']);

export type CatalogProductRow = {
  readonly product_id: string;
  readonly sku: string | null;
  readonly name: string;
  readonly category: { readonly id: string | null; readonly name: string | null };
  readonly outlet: { readonly id: string | null; readonly name: string };
  readonly price: number;
  readonly price_label: string;
  readonly track_stock: boolean;
  readonly stock_summary: { readonly status: 'unknown' | 'not_tracked'; readonly label: string; readonly quantity: number | null; readonly is_estimate: true };
  readonly status: 'active' | 'inactive' | 'archived' | 'unknown';
  readonly status_label: string;
  readonly updated_at: string | null;
  readonly can_edit: false;
  readonly can_delete: false;
};

export type CatalogCategoryRow = {
  readonly category_id: string;
  readonly name: string;
  readonly product_count: number;
  readonly status: 'active' | 'inactive' | 'archived' | 'unknown';
  readonly status_label: string;
  readonly updated_at: string | null;
  readonly can_edit: false;
  readonly can_delete: false;
};

export type CatalogReadData = {
  readonly meta: { readonly business_id: string; readonly outlet_id: string | null; readonly currency: 'IDR'; readonly data_status: 'real' | 'partial' | 'unavailable'; readonly generated_at: string };
  readonly pagination: { readonly page: number; readonly per_page: number; readonly total: number; readonly total_pages: number; readonly has_next_page: boolean; readonly has_previous_page: boolean };
  readonly totals: { readonly product_count: number; readonly active_count: number; readonly inactive_count: number; readonly tracked_stock_count: number; readonly category_count: number; readonly is_estimate: true };
  readonly rows: readonly CatalogProductRow[];
  readonly categories: readonly CatalogCategoryRow[];
  readonly data_notes: readonly { readonly key: string; readonly message: string; readonly severity: 'info' | 'warning' | 'danger' }[];
};

export type CatalogCategoriesReadData = { readonly meta: CatalogReadData['meta']; readonly pagination: CatalogReadData['pagination']; readonly rows: readonly CatalogCategoryRow[]; readonly data_notes: CatalogReadData['data_notes'] };
export type CatalogViewModel = { readonly metrics: readonly PreviewMetricFixture[]; readonly lanes: readonly PreviewLane[]; readonly table: PreviewTableFixture; readonly detailLinks: readonly { href: string; label: string; description: string }[]; readonly dataNotes: readonly string[] };
export type CatalogPageState = 'real' | 'preview' | 'session_required' | 'forbidden' | 'error';
export type CatalogPageModel = { readonly state: CatalogPageState; readonly data: CatalogViewModel; readonly sourceLabel: string; readonly sourceTone: 'success' | 'warning' | 'danger' | 'info' | 'neutral'; readonly title: string; readonly description: string; readonly detail: string };

export async function handleAdminCatalogProducts(request: NextRequest): Promise<NextResponse> {
  const session = readSessionFromRequest(request);
  if (!session) {
    const response = unauthenticatedResponse();
    clearSessionCookie(response);
    return response;
  }
  if (!isTenantAdmin(session)) return forbiddenResponse('Katalog toko hanya untuk owner atau admin.');

  try {
    const query = parseCatalogReadQuery(request.nextUrl.searchParams);
    const envelope = await createBackendClient().catalogProducts(session.token, query);
    return NextResponse.json(mapCatalogProductsEnvelope(envelope, query, session), { headers: SAFE_JSON_HEADERS });
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

export async function handleAdminCatalogCategories(request: NextRequest): Promise<NextResponse> {
  const session = readSessionFromRequest(request);
  if (!session) {
    const response = unauthenticatedResponse();
    clearSessionCookie(response);
    return response;
  }
  if (!isTenantAdmin(session)) return forbiddenResponse('Kategori toko hanya untuk owner atau admin.');

  try {
    const query = parseCatalogReadQuery(request.nextUrl.searchParams);
    const envelope = await createBackendClient().catalogProducts(session.token, query);
    const mapped = mapCatalogProductsEnvelope(envelope, query, session);
    const page = paginateRows(mapped.data.categories, query.page ?? 1, query.per_page ?? 50);
    const body: ApiEnvelope<CatalogCategoriesReadData> = { data: { meta: mapped.data.meta, pagination: page.pagination, rows: page.rows, data_notes: mapped.data.data_notes }, meta: mapped.meta };
    return NextResponse.json(body, { headers: SAFE_JSON_HEADERS });
  } catch (error) {
    const safeError = toSafeErrorResponse(error);
    return NextResponse.json(safeError.body, { status: safeError.status, headers: SAFE_JSON_HEADERS });
  }
}

export async function getCatalogPageModel(): Promise<CatalogPageModel> {
  const session = await readSessionFromRequestCookies();
  if (!session) return createPreviewPageModel({ state: 'session_required', sourceLabel: 'Data contoh fallback', sourceTone: 'warning', title: 'Katalog membutuhkan sesi admin toko', description: 'Silakan login saat auth UI aktif. Untuk sementara, daftar di bawah memakai data contoh berlabel jelas.', detail: 'Sesi admin belum tersedia atau sudah berakhir.' });
  if (!isTenantAdmin(session)) return createPreviewPageModel({ state: 'forbidden', sourceLabel: 'Akses ditolak', sourceTone: 'danger', title: 'Katalog toko hanya untuk owner/admin', description: 'Role saat ini tidak boleh membuka katalog tenant. Tidak ada data backend toko yang ditampilkan.', detail: 'Jika kamu superadmin, gunakan area Platform. Jika kamu kasir, gunakan aplikasi kasir.' });

  try {
    const query: CatalogReadQuery = { page: 1, per_page: 20, status: 'all' };
    const envelope = await createBackendClient().catalogProducts(session.token, query);
    const mapped = mapCatalogProductsEnvelope(envelope, query, session);
    return { state: 'real', data: mapCatalogReadToViewModel(mapped.data), sourceLabel: 'Data backend parsial', sourceTone: 'warning', title: mapped.data.rows.length === 0 ? 'Belum ada produk' : 'Katalog produk', description: mapped.data.rows.length === 0 ? 'Backend sudah terhubung, tetapi belum ada produk sesuai filter saat ini.' : 'Produk dan kategori dibaca read-only melalui BFF. Aksi tambah, edit, hapus, import, dan export tetap dikunci.', detail: `Diperbarui ${formatDateTime(mapped.data.meta.generated_at)}` };
  } catch (error) {
    if (error instanceof BackendApiError && error.status === 403) return createPreviewPageModel({ state: 'forbidden', sourceLabel: 'Akses ditolak', sourceTone: 'danger', title: 'Backend menolak akses katalog', description: 'Role atau scope sesi ini tidak diizinkan membuka katalog tenant.', detail: 'Data contoh fallback ditampilkan dengan label agar tidak tercampur dengan data real.' });
    return createPreviewPageModel({ state: 'error', sourceLabel: 'Backend belum tersedia', sourceTone: 'warning', title: 'Catalog backend belum bisa dibaca', description: 'Ada kendala saat menghubungi backend katalog. Data produksi tidak ditampilkan.', detail: error instanceof BackendApiError ? error.message : 'Server admin belum bisa menghubungi backend. Coba lagi nanti.' });
  }
}

export function parseCatalogReadQuery(searchParams: URLSearchParams): CatalogReadQuery {
  const unknown = [...searchParams.keys()].filter((key) => !ALLOWED_PARAMS.has(key));
  if (unknown.length > 0) throw validationError({ query: [`Parameter tidak didukung: ${unknown.join(', ')}`] });
  const search = searchParams.get('search') ?? undefined;
  const categoryId = searchParams.get('category_id') ?? undefined;
  const outletId = searchParams.get('outlet_id') ?? undefined;
  const status = searchParams.get('status') ?? undefined;
  const page = parseOptionalInteger(searchParams.get('page'), 'page', 1, 1, 1000000);
  const perPage = parseOptionalInteger(searchParams.get('per_page'), 'per_page', 20, 1, 100);
  if (search && search.length > 100) throw validationError({ search: ['Search maksimal 100 karakter.'] });
  if (categoryId && !UUID_PATTERN.test(categoryId)) throw validationError({ category_id: ['category_id harus UUID valid.'] });
  if (outletId && !UUID_PATTERN.test(outletId)) throw validationError({ outlet_id: ['outlet_id harus UUID valid.'] });
  if (status && !STATUSES.has(status)) throw validationError({ status: ['status tidak didukung.'] });
  return { ...(search ? { search } : {}), ...(categoryId ? { category_id: categoryId } : {}), ...(outletId ? { outlet_id: outletId } : {}), status: (status as CatalogReadQuery['status'] | undefined) ?? 'all', page, per_page: perPage };
}

export function mapCatalogProductsEnvelope(envelope: ApiEnvelope<BackendCatalogProductsData>, query: CatalogReadQuery, session: WebAdminSessionPayload): ApiEnvelope<CatalogReadData> {
  const categories = (envelope.data.categories ?? []).filter((category) => belongsToSession(category.business_id, session));
  const categoryMap = new Map(categories.map((category) => [category.id, category]));
  const allRows = (envelope.data.products ?? []).filter((product) => belongsToSession(product.business_id, session)).map((product) => mapProductRow(product, categoryMap)).filter((row): row is CatalogProductRow => Boolean(row)).filter((row) => matchesCatalogQuery(row, query));
  const categoryRows = categories.filter((category) => category.id && category.name).map((category) => mapCategoryRow(category, allRows)).filter((row) => matchesCategoryQuery(row, query));
  const page = paginateRows(allRows, query.page ?? 1, query.per_page ?? 20);
  return { data: { meta: { business_id: getSessionBusinessId(session), outlet_id: query.outlet_id ?? null, currency: 'IDR', data_status: 'partial', generated_at: new Date().toISOString() }, pagination: page.pagination, totals: { product_count: allRows.length, active_count: allRows.filter((row) => row.status === 'active').length, inactive_count: allRows.filter((row) => row.status === 'inactive').length, tracked_stock_count: allRows.filter((row) => row.track_stock).length, category_count: categoryRows.length, is_estimate: true }, rows: page.rows, categories: categoryRows, data_notes: [{ key: 'catalog_backend_partial', message: 'Backend catalog read belum menyediakan pagination/status/stock summary final; BFF menstabilkan shape read-only.', severity: 'info' }] }, meta: envelope.meta ?? { request_id: 'web-admin-catalog' } };
}

export function mapCatalogReadToViewModel(data: CatalogReadData): CatalogViewModel {
  const rows = data.rows.map((row) => [row.name, row.category.name ?? 'Tanpa kategori', row.price_label, row.stock_summary.label, row.status_label]);
  return { metrics: [{ label: 'Produk', value: String(data.totals.product_count), description: 'Read-only backend' }, { label: 'Aktif', value: String(data.totals.active_count), description: 'Status aman' }, { label: 'Kategori', value: String(data.totals.category_count), description: 'Filter katalog' }, { label: 'Lacak stok', value: String(data.totals.tracked_stock_count), description: 'Stok lihat Inventory' }], lanes: catalogLanes, table: { columns: ['Produk', 'Kategori', 'Harga', 'Ketersediaan', 'Status'], rows: rows.length > 0 ? rows : [['Belum ada produk', '—', 'Rp0', 'Kosong', 'Read-only']] }, detailLinks: catalogDetailLinks, dataNotes: data.data_notes.map((note) => note.message) };
}

function mapProductRow(product: BackendCatalogProduct, categoryMap: Map<string, { readonly name?: string | null }>): CatalogProductRow | null {
  if (!product.id || !product.name) return null;
  const categoryId = product.product_category_id ?? product.category_id ?? product.category?.id ?? null;
  const category = categoryId ? categoryMap.get(categoryId) : undefined;
  const status = normalizeStatus(product.status);
  const price = typeof product.price === 'number' ? product.price : 0;
  const trackStock = product.track_stock === true;
  return { product_id: product.id, sku: product.sku ?? product.barcode ?? null, name: product.name, category: { id: categoryId, name: product.category?.name ?? category?.name ?? null }, outlet: { id: product.outlet_id ?? product.outlet?.id ?? null, name: product.outlet?.name ?? (product.outlet_id ? 'Outlet terpilih' : 'Semua cabang') }, price, price_label: formatRupiah(price), track_stock: trackStock, stock_summary: { status: trackStock ? 'unknown' : 'not_tracked', label: trackStock ? 'Stok lihat di Inventory' : 'Tidak dilacak', quantity: null, is_estimate: true }, status, status_label: statusLabel(status), updated_at: product.updated_at ?? null, can_edit: false, can_delete: false };
}

function mapCategoryRow(category: { readonly id: string; readonly name?: string | null; readonly status?: string | null; readonly updated_at?: string | null }, products: readonly CatalogProductRow[]): CatalogCategoryRow {
  const status = normalizeStatus(category.status);
  return { category_id: category.id, name: category.name ?? 'Tanpa nama', product_count: products.filter((product) => product.category.id === category.id).length, status, status_label: statusLabel(status), updated_at: category.updated_at ?? null, can_edit: false, can_delete: false };
}

function matchesCatalogQuery(row: CatalogProductRow, query: CatalogReadQuery): boolean {
  if (query.category_id && row.category.id !== query.category_id) return false;
  if (query.status && query.status !== 'all' && row.status !== query.status) return false;
  if (!query.search) return true;
  const needle = query.search.toLowerCase();
  return [row.name, row.sku, row.category.name].some((value) => value?.toLowerCase().includes(needle));
}

function matchesCategoryQuery(row: CatalogCategoryRow, query: CatalogReadQuery): boolean {
  if (query.status && query.status !== 'all' && row.status !== query.status) return false;
  return query.search ? row.name.toLowerCase().includes(query.search.toLowerCase()) : true;
}

function paginateRows<TRow>(rows: readonly TRow[], page: number, perPage: number): { readonly rows: readonly TRow[]; readonly pagination: CatalogReadData['pagination'] } {
  const total = rows.length;
  const totalPages = total === 0 ? 0 : Math.ceil(total / perPage);
  return { rows: rows.slice((page - 1) * perPage, (page - 1) * perPage + perPage), pagination: { page, per_page: perPage, total, total_pages: totalPages, has_next_page: totalPages > 0 && page < totalPages, has_previous_page: page > 1 && totalPages > 0 } };
}

function belongsToSession(businessId: string | null | undefined, session: WebAdminSessionPayload): boolean {
  return !businessId || businessId === getSessionBusinessId(session);
}

function getSessionBusinessId(session: WebAdminSessionPayload): string {
  return session.business?.id ?? session.user.business_id ?? 'unknown-business';
}

function normalizeStatus(value: string | null | undefined): CatalogProductRow['status'] {
  if (value === 'inactive' || value === 'archived') return value;
  if (value === 'active' || !value) return 'active';
  return 'unknown';
}

function statusLabel(status: CatalogProductRow['status']): string {
  if (status === 'inactive') return 'Nonaktif';
  if (status === 'archived') return 'Diarsipkan';
  if (status === 'unknown') return 'Status belum final';
  return 'Aktif';
}

function formatRupiah(value: number): string {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(value);
}

function createPreviewPageModel(input: Omit<CatalogPageModel, 'data'>): CatalogPageModel {
  return { ...input, data: { metrics: [{ label: 'Produk', value: 'Preview', description: 'Data contoh' }, { label: 'Kategori', value: 'Preview', description: 'Data contoh' }, { label: 'Harga', value: 'Rp —', description: 'Tidak real' }, { label: 'Aksi', value: 'Disabled', description: 'Read-only' }], lanes: catalogLanes, table: catalogProductsTable, detailLinks: catalogDetailLinks, dataNotes: ['Data contoh preview, backend belum tersedia atau sesi belum aktif.'] } };
}

function validationError(details: Record<string, readonly string[]>): BackendApiError {
  return new BackendApiError(422, 'VALIDATION_ERROR', 'Filter catalog tidak valid.', details);
}

function parseOptionalInteger(value: string | null, field: string, fallback: number, min: number, max: number): number {
  if (!value) return fallback;
  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed < min || parsed > max) throw validationError({ [field]: [`${field} harus integer ${min}-${max}.`] });
  return parsed;
}

function unauthenticatedResponse(): NextResponse {
  return NextResponse.json({ error: { code: 'UNAUTHENTICATED', message: 'Sesi admin tidak valid atau sudah berakhir.', details: {} } }, { status: 401, headers: SAFE_JSON_HEADERS });
}

function forbiddenResponse(message: string): NextResponse {
  return NextResponse.json({ error: { code: 'FORBIDDEN', message, details: {} } }, { status: 403, headers: SAFE_JSON_HEADERS });
}

function isTenantAdmin(session: WebAdminSessionPayload): boolean {
  return session.user.role === 'owner' || session.user.role === 'admin';
}

async function readSessionFromRequestCookies(): Promise<WebAdminSessionPayload | null> {
  const cookieStore = await cookies();
  const sealed = cookieStore.get(getSessionCookieName())?.value;
  return sealed ? openSessionCookie(sealed) : null;
}

function formatDateTime(value: string): string {
  try {
    return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Jakarta' }).format(new Date(value));
  } catch {
    return value;
  }
}
