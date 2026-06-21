import { cookies } from 'next/headers';
import { NextRequest, NextResponse } from 'next/server';

import { inventoryDetailLinks, inventoryMetrics, inventoryResources } from '../../fixtures/preview';
import type { PreviewLane, PreviewMetricFixture, PreviewTableFixture } from '../../fixtures/preview';
import type { ApiEnvelope } from '../api/envelope';
import {
  BackendApiError,
  createBackendClient,
  toSafeErrorResponse,
  type BackendInventoryData,
  type BackendInventoryItem,
  type InventoryReadQuery,
} from './backend-client';
import { clearSessionCookie, readSessionFromRequest } from './admin-session';
import { getSessionCookieName, openSessionCookie, type WebAdminSessionPayload } from './session-cookie';

const SAFE_JSON_HEADERS = {
  'Cache-Control': 'no-store',
} as const;

const ALLOWED_INVENTORY_QUERY_PARAMS = new Set([
  'outlet_id',
  'category_id',
  'stock_status',
  'search',
  'page',
  'per_page',
]);
const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{12}$/i;
const STOCK_STATUSES = new Set(['all', 'in_stock', 'low', 'out', 'negative', 'not_tracked']);

export type InventoryRow = {
  readonly product_id: string;
  readonly sku: string | null;
  readonly name: string;
  readonly category: { readonly id: string | null; readonly name: string | null };
  readonly outlet: { readonly id: string; readonly name: string };
  readonly stock_on_hand: number;
  readonly available_stock: number;
  readonly reserved_stock: number;
  readonly in_transit_out: number;
  readonly in_transit_in: number;
  readonly unit: string | null;
  readonly low_stock_threshold: number | null;
  readonly track_stock: boolean;
  readonly status: 'in_stock' | 'low' | 'out' | 'negative' | 'not_tracked';
  readonly status_label: string;
  readonly updated_at: string | null;
  readonly can_adjust: false;
};

export type InventoryReadData = {
  readonly meta: {
    readonly timezone: string;
    readonly business_id: string;
    readonly outlet_id: string | null;
    readonly currency: 'IDR';
    readonly quantity_format: 'integer_units';
    readonly data_status: 'real' | 'partial' | 'unavailable';
    readonly generated_at: string;
  };
  readonly pagination: {
    readonly page: number;
    readonly per_page: number;
    readonly total: number;
    readonly total_pages: number;
    readonly has_next_page: boolean;
    readonly has_previous_page: boolean;
  };
  readonly totals: {
    readonly product_count: number;
    readonly tracked_product_count: number;
    readonly low_stock_count: number;
    readonly out_of_stock_count: number;
    readonly negative_stock_count: number;
    readonly in_transit_count: number;
    readonly is_estimate: boolean;
  };
  readonly rows: readonly InventoryRow[];
  readonly data_notes: readonly { readonly key: string; readonly message: string; readonly severity: 'info' | 'warning' | 'danger' }[];
};

export type InventoryViewModel = {
  readonly metrics: readonly PreviewMetricFixture[];
  readonly lanes: readonly PreviewLane[];
  readonly table: PreviewTableFixture;
  readonly detailLinks: readonly { href: string; label: string; description: string }[];
  readonly dataNotes: readonly string[];
};

export type InventoryPageState = 'real' | 'preview' | 'session_required' | 'forbidden' | 'error';

export type InventoryPageModel = {
  readonly state: InventoryPageState;
  readonly data: InventoryViewModel;
  readonly sourceLabel: string;
  readonly sourceTone: 'success' | 'warning' | 'danger' | 'info' | 'neutral';
  readonly title: string;
  readonly description: string;
  readonly detail: string;
};

export async function handleAdminInventory(request: NextRequest): Promise<NextResponse> {
  const session = readSessionFromRequest(request);

  if (!session) {
    const response = unauthenticatedResponse();
    clearSessionCookie(response);

    return response;
  }

  if (!isTenantAdmin(session)) {
    return forbiddenResponse('Inventaris toko hanya untuk owner atau admin.');
  }

  try {
    const query = parseInventoryReadQuery(request.nextUrl.searchParams);
    const envelope = await createBackendClient().inventory(session.token, query);
    const safeEnvelope = mapInventoryEnvelope(envelope, query, session);

    return NextResponse.json(safeEnvelope, { headers: SAFE_JSON_HEADERS });
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

export async function getInventoryPageModel(): Promise<InventoryPageModel> {
  const session = await readSessionFromRequestCookies();

  if (!session) {
    return createPreviewPageModel({
      state: 'session_required',
      sourceLabel: 'Data contoh fallback',
      sourceTone: 'warning',
      title: 'Inventaris membutuhkan sesi admin toko',
      description: 'Silakan login saat auth UI aktif. Untuk sementara, daftar di bawah memakai data contoh berlabel jelas.',
      detail: 'Sesi admin belum tersedia atau sudah berakhir.',
    });
  }

  if (!isTenantAdmin(session)) {
    return createPreviewPageModel({
      state: 'forbidden',
      sourceLabel: 'Akses ditolak',
      sourceTone: 'danger',
      title: 'Inventaris toko hanya untuk owner/admin',
      description: 'Role saat ini tidak boleh membuka inventaris tenant. Tidak ada data backend toko yang ditampilkan.',
      detail: 'Jika kamu superadmin, gunakan area Platform. Jika kamu kasir, gunakan aplikasi kasir.',
    });
  }

  try {
    const envelope = await createBackendClient().inventory(session.token, { page: 1, per_page: 20, stock_status: 'all' });
    const safeEnvelope = mapInventoryEnvelope(envelope, { page: 1, per_page: 20, stock_status: 'all' }, session);
    const viewModel = mapInventoryReadToViewModel(safeEnvelope.data);

    return {
      state: 'real',
      data: viewModel,
      sourceLabel: safeEnvelope.data.meta.data_status === 'partial' ? 'Data backend parsial' : 'Data backend',
      sourceTone: safeEnvelope.data.meta.data_status === 'partial' ? 'warning' : 'success',
      title: safeEnvelope.data.rows.length === 0 ? 'Belum ada data stok' : 'Inventaris toko',
      description: safeEnvelope.data.rows.length === 0
        ? 'Backend sudah terhubung, tetapi belum ada stok sesuai filter saat ini.'
        : 'Daftar stok dibaca read-only melalui BFF. Aksi adjustment, transfer, purchase, count, waste, dan export tetap dikunci.',
      detail: `Diperbarui ${formatDateTime(safeEnvelope.data.meta.generated_at)} · ${safeEnvelope.data.meta.timezone}`,
    };
  } catch (error) {
    if (error instanceof BackendApiError && error.status === 401) {
      return createPreviewPageModel({
        state: 'session_required',
        sourceLabel: 'Sesi berakhir',
        sourceTone: 'warning',
        title: 'Sesi admin sudah berakhir',
        description: 'Silakan login ulang saat auth UI aktif. Data inventory real tidak ditampilkan tanpa sesi valid.',
        detail: 'Data contoh fallback ditampilkan hanya untuk menjaga layout tetap terbaca.',
      });
    }

    if (error instanceof BackendApiError && error.status === 403) {
      return createPreviewPageModel({
        state: 'forbidden',
        sourceLabel: 'Akses ditolak',
        sourceTone: 'danger',
        title: 'Backend menolak akses inventaris',
        description: 'Role atau scope sesi ini tidak diizinkan membuka inventaris tenant.',
        detail: 'Data contoh fallback ditampilkan dengan label agar tidak tercampur dengan data real.',
      });
    }

    return createPreviewPageModel({
      state: 'error',
      sourceLabel: 'Backend belum tersedia',
      sourceTone: 'warning',
      title: 'Inventory backend belum bisa dibaca',
      description: 'Ada kendala saat menghubungi backend inventory. Data produksi tidak ditampilkan.',
      detail: error instanceof BackendApiError ? error.message : 'Server admin belum bisa menghubungi backend. Coba lagi nanti.',
    });
  }
}

export function parseInventoryReadQuery(searchParams: URLSearchParams): InventoryReadQuery {
  const unknownParams = [...searchParams.keys()].filter((key) => !ALLOWED_INVENTORY_QUERY_PARAMS.has(key));

  if (unknownParams.length > 0) {
    throw validationError({ query: [`Parameter tidak didukung: ${unknownParams.join(', ')}`] });
  }

  const outletId = searchParams.get('outlet_id') ?? undefined;
  const categoryId = searchParams.get('category_id') ?? undefined;
  const stockStatus = searchParams.get('stock_status') ?? undefined;
  const search = searchParams.get('search') ?? undefined;
  const page = parseOptionalInteger(searchParams.get('page'), 'page', 1, 1, 1000000);
  const perPage = parseOptionalInteger(searchParams.get('per_page'), 'per_page', 20, 1, 100);

  if (outletId && !UUID_PATTERN.test(outletId)) {
    throw validationError({ outlet_id: ['outlet_id harus UUID valid.'] });
  }

  if (categoryId && !UUID_PATTERN.test(categoryId)) {
    throw validationError({ category_id: ['category_id harus UUID valid.'] });
  }

  if (stockStatus && !STOCK_STATUSES.has(stockStatus)) {
    throw validationError({ stock_status: ['stock_status tidak didukung.'] });
  }

  if (search && search.length > 100) {
    throw validationError({ search: ['Search maksimal 100 karakter.'] });
  }

  return {
    ...(outletId ? { outlet_id: outletId } : {}),
    ...(categoryId ? { category_id: categoryId } : {}),
    stock_status: (stockStatus as InventoryReadQuery['stock_status'] | undefined) ?? 'all',
    ...(search ? { search } : {}),
    page,
    per_page: perPage,
  };
}

export function mapInventoryEnvelope(
  envelope: ApiEnvelope<BackendInventoryData>,
  query: InventoryReadQuery,
  session: WebAdminSessionPayload,
): ApiEnvelope<InventoryReadData> {
  const allRows = extractInventoryItems(envelope.data)
    .map((row) => mapInventoryRow(row))
    .filter((row): row is InventoryRow => Boolean(row))
    .filter((row) => rowMatchesQuery(row, query));

  const page = query.page ?? 1;
  const perPage = query.per_page ?? 20;
  const total = allRows.length;
  const start = (page - 1) * perPage;
  const rows = allRows.slice(start, start + perPage);
  const totalPages = Math.max(1, Math.ceil(total / perPage));
  const generatedAt = new Date().toISOString();
  const notes = [
    {
      key: 'read_only_boundary',
      message: 'Inventory Web Admin hanya membaca stok melalui BFF; adjustment, transfer, purchase, count, waste, dan export tetap dikunci.',
      severity: 'info' as const,
    },
    {
      key: 'backend_filter_partial',
      message: 'Sebagian filter dan pagination distabilkan di BFF sampai filter inventory backend final untuk volume produksi besar.',
      severity: 'info' as const,
    },
  ];

  return {
    data: {
      meta: {
        timezone: 'Asia/Jakarta',
        business_id: session.business?.id ?? session.user.business_id ?? 'unknown-business',
        outlet_id: query.outlet_id ?? null,
        currency: 'IDR',
        quantity_format: 'integer_units',
        data_status: 'partial',
        generated_at: generatedAt,
      },
      pagination: {
        page,
        per_page: perPage,
        total,
        total_pages: totalPages,
        has_next_page: page < totalPages,
        has_previous_page: page > 1,
      },
      totals: createInventoryTotals(allRows),
      rows,
      data_notes: notes,
    },
    meta: {
      ...(envelope.meta ?? {}),
      generated_at: generatedAt,
    },
  };
}

export function mapInventoryReadToViewModel(data: InventoryReadData): InventoryViewModel {
  return {
    metrics: [
      { label: 'Total item', value: String(data.totals.product_count), badge: 'Backend', tone: 'success', description: `${data.totals.tracked_product_count} item tracked. Data dibaca read-only dari backend.` },
      { label: 'Low stock', value: String(data.totals.low_stock_count), badge: data.totals.is_estimate ? 'Estimasi' : 'Real', tone: data.totals.low_stock_count > 0 ? 'warning' : 'success', description: 'Stok hampir habis berdasarkan data backend/BFF.' },
      { label: 'Transfer', value: `${data.totals.in_transit_count} in transit`, badge: 'Read-only', description: 'Transfer hanya dipantau, tidak ada create/send/receive/cancel dari halaman ini.' },
    ],
    lanes: [
      { label: 'Stock health', items: [`${data.totals.low_stock_count} low stock`, `${data.totals.out_of_stock_count} stok habis`, `${data.totals.negative_stock_count} stok negatif`] },
      { label: 'Read-only guard', items: ['Adjustment disabled', 'Transfer action disabled', 'Export disabled'] },
      { label: 'Scope', items: [`${data.pagination.total} item sesuai filter`, `Timezone ${data.meta.timezone}`, 'Tenant-scoped by session'] },
    ],
    table: {
      columns: ['Produk', 'SKU', 'Outlet', 'Stok', 'Status'],
      rows: data.rows.length > 0
        ? data.rows.map((row) => [row.name, row.sku ?? '—', row.outlet.name, `${row.stock_on_hand} ${row.unit ?? ''}`.trim(), row.status_label])
        : [['Belum ada stok', '—', '—', '0', 'Kosong']],
      caption: `Inventaris read-only · ${data.pagination.total} item · ${data.meta.data_status}`,
      primaryColumn: 0,
      statusColumn: 4,
      numericColumns: [3],
    },
    detailLinks: inventoryDetailLinks,
    dataNotes: data.data_notes.map((note) => note.message),
  };
}

function extractInventoryItems(data: BackendInventoryData): readonly BackendInventoryItem[] {
  if (Array.isArray(data.inventory)) {
    return data.inventory;
  }

  if (Array.isArray(data.rows)) {
    return data.rows;
  }

  if (Array.isArray(data.items)) {
    return data.items;
  }

  return [];
}

function mapInventoryRow(row: BackendInventoryItem): InventoryRow | null {
  const product = row.product ?? null;
  const outlet = row.outlet ?? null;
  const productId = asString(row.product_id ?? row.id ?? product?.id);
  const outletId = asString(row.outlet_id ?? outlet?.id);

  if (!productId || !outletId) {
    return null;
  }

  const outletName = asString(outlet?.name ?? row.outlet_name) ?? 'Outlet';
  const stockOnHand = asNumber(row.stock_on_hand ?? row.quantity ?? row.current_stock, 0);
  const trackStock = asBoolean(row.track_stock ?? product?.track_stock, true);
  const threshold = asNullableNumber(row.low_stock_threshold ?? row.threshold);
  const status = normalizeStockStatus(row.status, stockOnHand, threshold, trackStock);

  return {
    product_id: productId,
    sku: asString(row.sku ?? row.barcode ?? product?.sku ?? product?.barcode),
    name: asString(row.name ?? product?.name) ?? 'Produk tanpa nama',
    category: {
      id: asString(row.category?.id ?? product?.category?.id),
      name: asString(row.category?.name ?? product?.category?.name),
    },
    outlet: { id: outletId, name: outletName },
    stock_on_hand: stockOnHand,
    available_stock: asNumber(row.available_stock, stockOnHand),
    reserved_stock: asNumber(row.reserved_stock, 0),
    in_transit_out: asNumber(row.in_transit_out, 0),
    in_transit_in: asNumber(row.in_transit_in, 0),
    unit: asString(row.unit ?? product?.unit),
    low_stock_threshold: threshold,
    track_stock: trackStock,
    status,
    status_label: stockStatusLabel(status),
    updated_at: asString(row.updated_at),
    can_adjust: false,
  };
}

function rowMatchesQuery(row: InventoryRow, query: InventoryReadQuery): boolean {
  if (query.outlet_id && row.outlet.id !== query.outlet_id) {
    return false;
  }

  if (query.category_id && row.category.id !== query.category_id) {
    return false;
  }

  if (query.stock_status && query.stock_status !== 'all' && row.status !== query.stock_status) {
    return false;
  }

  if (query.search) {
    const haystack = `${row.name} ${row.sku ?? ''}`.toLowerCase();

    if (!haystack.includes(query.search.toLowerCase())) {
      return false;
    }
  }

  return true;
}

function createInventoryTotals(rows: readonly InventoryRow[]): InventoryReadData['totals'] {
  return {
    product_count: rows.length,
    tracked_product_count: rows.filter((row) => row.track_stock).length,
    low_stock_count: rows.filter((row) => row.status === 'low').length,
    out_of_stock_count: rows.filter((row) => row.status === 'out').length,
    negative_stock_count: rows.filter((row) => row.status === 'negative').length,
    in_transit_count: rows.filter((row) => row.in_transit_in > 0 || row.in_transit_out > 0).length,
    is_estimate: rows.some((row) => row.low_stock_threshold === null),
  };
}

function createPreviewPageModel(input: Omit<InventoryPageModel, 'data'>): InventoryPageModel {
  return {
    ...input,
    data: {
      metrics: inventoryMetrics,
      lanes: [
        { label: 'Preview mode', items: inventoryResources.map((resource) => resource.label) },
        { label: 'Guard rails', items: ['Adjustment disabled', 'Transfer disabled', 'No browser stock calculation'] },
        { label: 'Session', items: [input.sourceLabel, input.detail] },
      ],
      table: {
        columns: ['Resource', 'Path', 'Mode'],
        rows: inventoryResources.map((resource) => [resource.label, resource.path, 'Read-only preview']),
        caption: 'Data contoh preview, backend belum tersedia atau sesi belum valid.',
        primaryColumn: 0,
        statusColumn: 2,
      },
      detailLinks: inventoryDetailLinks,
      dataNotes: [input.detail],
    },
  };
}

function parseOptionalInteger(value: string | null, field: string, fallback: number, min: number, max: number): number {
  if (!value) {
    return fallback;
  }

  if (!/^\d+$/.test(value)) {
    throw validationError({ [field]: [`${field} harus berupa angka.`] });
  }

  const parsed = Number(value);

  if (parsed < min || parsed > max) {
    throw validationError({ [field]: [`${field} harus di antara ${min} dan ${max}.`] });
  }

  return parsed;
}

function validationError(details: Record<string, readonly string[]>): BackendApiError {
  return new BackendApiError(422, 'VALIDATION_ERROR', 'Filter inventory tidak valid.', details);
}

function isTenantAdmin(session: WebAdminSessionPayload): boolean {
  return session.user.role === 'owner' || session.user.role === 'admin';
}

async function readSessionFromRequestCookies(): Promise<WebAdminSessionPayload | null> {
  const store = await cookies();
  const sealed = store.get(getSessionCookieName())?.value;

  return sealed ? openSessionCookie(sealed) : null;
}

function unauthenticatedResponse(): NextResponse {
  return NextResponse.json(
    { error: { code: 'UNAUTHENTICATED', message: 'Sesi admin tidak ditemukan atau sudah berakhir.', details: {} } },
    { status: 401, headers: SAFE_JSON_HEADERS },
  );
}

function forbiddenResponse(message: string): NextResponse {
  return NextResponse.json(
    { error: { code: 'FORBIDDEN', message, details: {} } },
    { status: 403, headers: SAFE_JSON_HEADERS },
  );
}

function asString(value: unknown): string | null {
  return typeof value === 'string' && value.trim() ? value : null;
}

function asNumber(value: unknown, fallback: number): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : fallback;
}

function asNullableNumber(value: unknown): number | null {
  return typeof value === 'number' && Number.isFinite(value) ? value : null;
}

function asBoolean(value: unknown, fallback: boolean): boolean {
  return typeof value === 'boolean' ? value : fallback;
}

function normalizeStockStatus(value: unknown, stock: number, threshold: number | null, trackStock: boolean): InventoryRow['status'] {
  if (!trackStock) {
    return 'not_tracked';
  }

  if (typeof value === 'string' && STOCK_STATUSES.has(value) && value !== 'all') {
    return value as InventoryRow['status'];
  }

  if (stock < 0) {
    return 'negative';
  }

  if (stock === 0) {
    return 'out';
  }

  if (threshold !== null && stock <= threshold) {
    return 'low';
  }

  return 'in_stock';
}

function stockStatusLabel(status: InventoryRow['status']): string {
  if (status === 'low') {
    return 'Stok hampir habis';
  }

  if (status === 'out') {
    return 'Stok habis';
  }

  if (status === 'negative') {
    return 'Stok negatif';
  }

  if (status === 'not_tracked') {
    return 'Tidak dilacak';
  }

  return 'Stok aman';
}

function formatDateTime(value: string): string {
  try {
    return new Intl.DateTimeFormat('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'short',
      timeZone: 'Asia/Jakarta',
    }).format(new Date(value));
  } catch {
    return value;
  }
}
