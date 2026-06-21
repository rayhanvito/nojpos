import { cookies } from 'next/headers';
import { NextRequest, NextResponse } from 'next/server';

import type { PreviewLane, PreviewMetricFixture, PreviewTableFixture } from '../../fixtures/preview';
import { transactionDetailLinks, transactionLanes, transactionMetrics, transactionsTable } from '../../fixtures/preview';
import type { ApiEnvelope } from '../api/envelope';
import {
  BackendApiError,
  createBackendClient,
  toSafeErrorResponse,
  type BackendTransactionRow,
  type BackendTransactionsData,
  type TransactionsReadQuery,
} from './backend-client';
import { clearSessionCookie, readSessionFromRequest } from './admin-session';
import { getSessionCookieName, openSessionCookie, type WebAdminSessionPayload } from './session-cookie';

const SAFE_JSON_HEADERS = {
  'Cache-Control': 'no-store',
} as const;

const ALLOWED_TRANSACTION_QUERY_PARAMS = new Set([
  'date_from',
  'date_to',
  'outlet_id',
  'cashier_id',
  'payment_method',
  'status',
  'search',
  'page',
  'per_page',
]);
const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;
const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const PAYMENT_METHODS = new Set(['cash', 'qris', 'card', 'transfer', 'ewallet', 'other']);
const STATUSES = new Set(['paid', 'partial', 'unpaid', 'held', 'payment_pending', 'payment_failed', 'voided', 'refunded', 'pending']);

export type TransactionRowView = {
  readonly id: string;
  readonly code: string;
  readonly occurredAt: string;
  readonly timeLabel: string;
  readonly outletName: string;
  readonly cashierName: string;
  readonly customerLabel: string;
  readonly paymentLabel: string;
  readonly total: number;
  readonly subtotal: number;
  readonly discountTotal: number;
  readonly itemCount: number;
  readonly status: string;
  readonly statusLabel: string;
  readonly canOpenDetail: boolean;
};

export type TransactionsViewModel = {
  readonly metrics: readonly PreviewMetricFixture[];
  readonly lanes: readonly PreviewLane[];
  readonly table: PreviewTableFixture;
  readonly detailLinks: readonly { href: string; label: string; description: string }[];
  readonly rows: readonly TransactionRowView[];
  readonly dataNotes: readonly string[];
  readonly pagination: {
    readonly page: number;
    readonly perPage: number;
    readonly total: number;
    readonly totalPages: number;
    readonly hasNextPage: boolean;
    readonly hasPreviousPage: boolean;
  };
};

export type TransactionsPageState = 'real' | 'preview' | 'session_required' | 'forbidden' | 'error';

export type TransactionsPageModel = {
  readonly state: TransactionsPageState;
  readonly data: TransactionsViewModel;
  readonly sourceLabel: string;
  readonly sourceTone: 'success' | 'warning' | 'danger' | 'info' | 'neutral';
  readonly title: string;
  readonly description: string;
  readonly detail: string;
};

export async function handleAdminTransactions(request: NextRequest): Promise<NextResponse> {
  const session = readSessionFromRequest(request);

  if (!session) {
    const response = unauthenticatedResponse();
    clearSessionCookie(response);

    return response;
  }

  if (!isTenantAdmin(session)) {
    return forbiddenResponse('Daftar transaksi toko hanya untuk owner atau admin.');
  }

  try {
    const query = parseTransactionsReadQuery(request.nextUrl.searchParams);
    const envelope = await createBackendClient().transactions(session.token, query);
    const safeEnvelope = mapTransactionsEnvelope(envelope, query, session);

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

export async function getTransactionsPageModel(): Promise<TransactionsPageModel> {
  const session = await readSessionFromRequestCookies();

  if (!session) {
    return createPreviewPageModel({
      state: 'session_required',
      sourceLabel: 'Data contoh fallback',
      sourceTone: 'warning',
      title: 'Transaksi membutuhkan sesi admin toko',
      description: 'Silakan login saat auth UI aktif. Untuk sementara, daftar di bawah memakai data contoh berlabel jelas.',
      detail: 'Sesi admin belum tersedia atau sudah berakhir.',
    });
  }

  if (!isTenantAdmin(session)) {
    return createPreviewPageModel({
      state: 'forbidden',
      sourceLabel: 'Akses ditolak',
      sourceTone: 'danger',
      title: 'Transaksi toko hanya untuk owner/admin',
      description: 'Role saat ini tidak boleh membuka transaksi tenant. Tidak ada data backend toko yang ditampilkan.',
      detail: 'Jika kamu superadmin, gunakan area Platform. Jika kamu kasir, gunakan aplikasi kasir.',
    });
  }

  try {
    const envelope = await createBackendClient().transactions(session.token, { page: 1, per_page: 20 });
    const safeEnvelope = mapTransactionsEnvelope(envelope, { page: 1, per_page: 20 }, session);
    const viewModel = mapTransactionsReadToViewModel(safeEnvelope.data);

    return {
      state: 'real',
      data: viewModel,
      sourceLabel: safeEnvelope.data.meta.data_status === 'partial' ? 'Data backend parsial' : 'Data backend',
      sourceTone: safeEnvelope.data.meta.data_status === 'partial' ? 'warning' : 'success',
      title: safeEnvelope.data.rows.length === 0 ? 'Belum ada transaksi' : 'Riwayat transaksi toko',
      description: safeEnvelope.data.rows.length === 0
        ? 'Backend sudah terhubung, tetapi belum ada transaksi sesuai filter saat ini.'
        : 'Daftar transaksi dibaca read-only melalui BFF. Aksi void, refund, reprint, dan export tetap dikunci.',
      detail: `Diperbarui ${formatDateTime(safeEnvelope.data.meta.generated_at)} · ${safeEnvelope.data.meta.timezone}`,
    };
  } catch (error) {
    if (error instanceof BackendApiError && error.status === 401) {
      return createPreviewPageModel({
        state: 'session_required',
        sourceLabel: 'Sesi berakhir',
        sourceTone: 'warning',
        title: 'Sesi admin sudah berakhir',
        description: 'Silakan login ulang saat auth UI aktif. Data transaksi real tidak ditampilkan tanpa sesi valid.',
        detail: 'Data contoh fallback ditampilkan hanya untuk menjaga layout tetap terbaca.',
      });
    }

    if (error instanceof BackendApiError && error.status === 403) {
      return createPreviewPageModel({
        state: 'forbidden',
        sourceLabel: 'Akses ditolak',
        sourceTone: 'danger',
        title: 'Backend menolak akses transaksi',
        description: 'Role atau scope sesi ini tidak diizinkan membuka transaksi tenant.',
        detail: 'Data contoh fallback ditampilkan dengan label agar tidak tercampur dengan data real.',
      });
    }

    return createPreviewPageModel({
      state: 'error',
      sourceLabel: 'Backend belum tersedia',
      sourceTone: 'warning',
      title: 'Transaksi backend belum bisa dibaca',
      description: 'Ada kendala saat menghubungi backend transaksi. Data produksi tidak ditampilkan.',
      detail: error instanceof BackendApiError ? error.message : 'Server admin belum bisa menghubungi backend. Coba lagi nanti.',
    });
  }
}

export function parseTransactionsReadQuery(searchParams: URLSearchParams): TransactionsReadQuery {
  const unknownParams = [...searchParams.keys()].filter((key) => !ALLOWED_TRANSACTION_QUERY_PARAMS.has(key));

  if (unknownParams.length > 0) {
    throw new BackendApiError(422, 'VALIDATION_ERROR', 'Filter transaksi tidak valid.', {
      query: [`Parameter tidak didukung: ${unknownParams.join(', ')}`],
    });
  }

  const dateFrom = searchParams.get('date_from') ?? undefined;
  const dateTo = searchParams.get('date_to') ?? undefined;
  const outletId = searchParams.get('outlet_id') ?? undefined;
  const cashierId = searchParams.get('cashier_id') ?? undefined;
  const paymentMethod = searchParams.get('payment_method') ?? undefined;
  const status = searchParams.get('status') ?? undefined;
  const search = searchParams.get('search') ?? undefined;
  const page = parseOptionalInteger(searchParams.get('page'), 'page', 1, 1, 1000000);
  const perPage = parseOptionalInteger(searchParams.get('per_page'), 'per_page', 20, 1, 100);

  if (dateFrom && !isValidDate(dateFrom)) {
    throw validationError({ date_from: ['Format date_from harus YYYY-MM-DD.'] });
  }

  if (dateTo && !isValidDate(dateTo)) {
    throw validationError({ date_to: ['Format date_to harus YYYY-MM-DD.'] });
  }

  if (dateFrom && dateTo && dateTo < dateFrom) {
    throw validationError({ date_to: ['date_to tidak boleh sebelum date_from.'] });
  }

  if (outletId && !UUID_PATTERN.test(outletId)) {
    throw validationError({ outlet_id: ['outlet_id harus UUID valid.'] });
  }

  if (cashierId && !UUID_PATTERN.test(cashierId)) {
    throw validationError({ cashier_id: ['cashier_id harus UUID valid.'] });
  }

  if (paymentMethod && !PAYMENT_METHODS.has(paymentMethod)) {
    throw validationError({ payment_method: ['Metode pembayaran tidak didukung.'] });
  }

  if (status && !STATUSES.has(status)) {
    throw validationError({ status: ['Status transaksi tidak didukung.'] });
  }

  if (search && search.length > 100) {
    throw validationError({ search: ['Search maksimal 100 karakter.'] });
  }

  return {
    ...(dateFrom ? { date_from: dateFrom } : {}),
    ...(dateTo ? { date_to: dateTo } : {}),
    ...(outletId ? { outlet_id: outletId } : {}),
    ...(cashierId ? { cashier_id: cashierId } : {}),
    ...(paymentMethod ? { payment_method: paymentMethod as TransactionsReadQuery['payment_method'] } : {}),
    ...(status ? { status: status as TransactionsReadQuery['status'] } : {}),
    ...(search ? { search } : {}),
    page,
    per_page: perPage,
  };
}

export type TransactionsReadData = {
  readonly meta: {
    readonly date_from: string | null;
    readonly date_to: string | null;
    readonly timezone: string;
    readonly business_id: string;
    readonly outlet_id: string | null;
    readonly currency: 'IDR';
    readonly money_format: 'integer_rupiah';
    readonly data_status: 'real' | 'partial' | 'unavailable';
    readonly generated_at: string;
  };
  readonly pagination: TransactionsViewModel['pagination'];
  readonly totals: {
    readonly transaction_count: number;
    readonly gross_sales: number;
    readonly discount_total: number;
    readonly service_charge_total: number;
    readonly tax_total: number;
    readonly rounding_total: number;
    readonly net_sales: number;
    readonly paid_amount: number;
    readonly pending_amount: number;
    readonly refund_amount: number;
    readonly void_count: number;
    readonly is_estimate: boolean;
  };
  readonly rows: readonly {
    readonly id: string;
    readonly code: string;
    readonly occurred_at: string;
    readonly business_date: string;
    readonly time_label: string;
    readonly outlet: { readonly id: string | null; readonly name: string | null };
    readonly cashier: { readonly id: string | null; readonly name: string | null };
    readonly customer: { readonly id: string | null; readonly name: string | null; readonly display_label: string | null };
    readonly payment_method: { readonly value: string | null; readonly label: string | null; readonly is_mixed: boolean };
    readonly total: number;
    readonly subtotal: number;
    readonly discount_total: number;
    readonly item_count: number;
    readonly status: string;
    readonly status_label: string;
    readonly can_open_detail: boolean;
  }[];
  readonly data_notes: readonly { readonly key: string; readonly message: string; readonly severity: string }[];
};

export function mapTransactionsEnvelope(
  envelope: ApiEnvelope<BackendTransactionsData>,
  query: TransactionsReadQuery,
  session: WebAdminSessionPayload,
): ApiEnvelope<TransactionsReadData> {
  const allRows = (envelope.data.transactions ?? []).filter((row) => belongsToSessionBusiness(row, session));
  const filteredRows = filterBackendRows(allRows, query);
  const page = query.page ?? 1;
  const perPage = query.per_page ?? 20;
  const pagedRows = filteredRows.slice((page - 1) * perPage, page * perPage);
  const rows = pagedRows.map((row) => mapBackendTransactionRow(row));
  const totalPages = Math.max(1, Math.ceil(filteredRows.length / perPage));
  const generatedAt = new Date().toISOString();

  return {
    data: {
      meta: {
        date_from: query.date_from ?? null,
        date_to: query.date_to ?? null,
        timezone: 'Asia/Jakarta',
        business_id: session.user.business_id ?? session.business?.id ?? '',
        outlet_id: query.outlet_id ?? null,
        currency: 'IDR',
        money_format: 'integer_rupiah',
        data_status: 'partial',
        generated_at: generatedAt,
      },
      pagination: {
        page,
        perPage,
        total: filteredRows.length,
        totalPages,
        hasNextPage: page < totalPages,
        hasPreviousPage: page > 1,
      },
      totals: calculateTotals(filteredRows),
      rows,
      data_notes: [
        {
          key: 'backend_filter_pagination_gap',
          message: 'BFF menormalkan dan mem-paginate response karena backend transactions endpoint belum punya filter/pagination produksi lengkap.',
          severity: 'info',
        },
      ],
    },
    meta: envelope.meta,
  };
}

export function mapTransactionsReadToViewModel(data: TransactionsReadData): TransactionsViewModel {
  return {
    metrics: [
      {
        label: 'Transactions',
        value: String(data.totals.transaction_count),
        description: 'Jumlah transaksi dari backend read-only setelah filter BFF.',
        badge: data.meta.data_status === 'partial' ? 'Partial' : 'Real',
        tone: data.totals.transaction_count > 0 ? 'success' : 'neutral',
      },
      {
        label: 'Net sales',
        value: formatRupiah(data.totals.net_sales),
        description: 'Total akhir berasal dari grand_total backend, bukan hitungan browser.',
        badge: 'Server',
        tone: 'success',
      },
      {
        label: 'Void/refund',
        value: 'Read-only',
        description: 'Void, refund, reprint, dan export tetap disabled sampai kontrak sensitif tersedia.',
        badge: 'Gated',
        tone: 'warning',
      },
    ],
    lanes: transactionLanes,
    table: {
      columns: ['Transaction', 'Outlet', 'Payment', 'Total', 'Status'],
      rows: data.rows.map((row) => [row.code, row.outlet.name ?? 'Semua cabang', row.payment_method.label ?? '-', formatRupiah(row.total), row.status_label]),
      caption: data.rows.length === 0
        ? 'Belum ada transaksi sesuai filter. Tidak ada data contoh yang dicampur dengan data backend.'
        : 'Daftar transaksi backend read-only. Tidak ada action void, refund, reprint, atau export aktif.',
      primaryColumn: 'Transaction',
      statusColumn: 'Status',
      metaColumns: ['Outlet', 'Payment', 'Total'],
      numericColumns: ['Total'],
    },
    detailLinks: data.rows.map((row) => ({
      href: `/transactions/${row.id}`,
      label: row.code,
      description: `${row.status_label} · ${formatRupiah(row.total)} · Detail read-only`,
    })),
    rows: data.rows.map((row) => ({
      id: row.id,
      code: row.code,
      occurredAt: row.occurred_at,
      timeLabel: row.time_label,
      outletName: row.outlet.name ?? '-',
      cashierName: row.cashier.name ?? '-',
      customerLabel: row.customer.display_label ?? '-',
      paymentLabel: row.payment_method.label ?? '-',
      total: row.total,
      subtotal: row.subtotal,
      discountTotal: row.discount_total,
      itemCount: row.item_count,
      status: row.status,
      statusLabel: row.status_label,
      canOpenDetail: row.can_open_detail,
    })),
    dataNotes: data.data_notes.map((note) => note.message),
    pagination: data.pagination,
  };
}

function filterBackendRows(rows: readonly BackendTransactionRow[], query: TransactionsReadQuery): readonly BackendTransactionRow[] {
  const search = query.search?.trim().toLowerCase();

  return rows.filter((row) => {
    if (query.outlet_id && row.outlet_id !== query.outlet_id) return false;
    if (query.cashier_id && row.cashier_id !== query.cashier_id) return false;
    if (query.status && row.status !== query.status) return false;
    if (query.payment_method && !paymentMethods(row).includes(query.payment_method)) return false;

    const occurredDate = businessDate(row.created_at ?? row.updated_at ?? null);
    if (query.date_from && occurredDate < query.date_from) return false;
    if (query.date_to && occurredDate > query.date_to) return false;

    if (search) {
      const haystack = [row.number, row.cashier?.name, row.customer?.name, row.status].filter(Boolean).join(' ').toLowerCase();
      if (!haystack.includes(search)) return false;
    }

    return true;
  });
}

function mapBackendTransactionRow(row: BackendTransactionRow): TransactionsReadData['rows'][number] {
  const firstPayment = row.payments?.[0];
  const methods = paymentMethods(row);
  const value = methods.length > 1 ? 'mixed' : methods[0] ?? firstPayment?.method ?? null;
  const status = row.status ?? 'unknown';
  const occurredAt = row.created_at ?? row.updated_at ?? new Date(0).toISOString();

  return {
    id: row.id,
    code: row.number ?? row.id,
    occurred_at: occurredAt,
    business_date: businessDate(occurredAt),
    time_label: timeLabel(occurredAt),
    outlet: {
      id: row.outlet_id ?? null,
      name: row.outlet_id ? 'Outlet backend' : null,
    },
    cashier: {
      id: row.cashier?.id ?? row.cashier_id ?? null,
      name: row.cashier?.name ?? null,
    },
    customer: {
      id: row.customer?.id ?? row.customer_id ?? null,
      name: row.customer?.name ?? null,
      display_label: row.customer?.name ?? null,
    },
    payment_method: {
      value,
      label: value ? paymentMethodLabel(value) : null,
      is_mixed: methods.length > 1,
    },
    total: money(row.grand_total),
    subtotal: money(row.subtotal),
    discount_total: money(row.discount_total),
    item_count: row.items?.reduce((sum, item) => sum + Math.max(0, Number(item.quantity ?? 0)), 0) ?? 0,
    status,
    status_label: statusLabel(status),
    can_open_detail: true,
  };
}

function calculateTotals(rows: readonly BackendTransactionRow[]): TransactionsReadData['totals'] {
  const grossSales = rows.reduce((sum, row) => sum + money(row.subtotal), 0);
  const discountTotal = rows.reduce((sum, row) => sum + money(row.discount_total), 0);
  const serviceChargeTotal = rows.reduce((sum, row) => sum + money(row.service_charge_total), 0);
  const taxTotal = rows.reduce((sum, row) => sum + money(row.tax_total), 0);
  const roundingTotal = rows.reduce((sum, row) => sum + money(row.rounding_total), 0);
  const netSales = rows.reduce((sum, row) => sum + money(row.grand_total), 0);
  const paidAmount = rows.filter((row) => row.status === 'paid' || row.status === 'partial' || row.status === 'refunded').reduce((sum, row) => sum + money(row.grand_total), 0);
  const pendingAmount = rows.filter((row) => row.status === 'payment_pending' || row.status === 'pending' || row.status === 'unpaid').reduce((sum, row) => sum + money(row.grand_total), 0);
  const refundAmount = rows.filter((row) => row.status === 'refunded').reduce((sum, row) => sum + money(row.grand_total), 0);
  const voidCount = rows.filter((row) => row.status === 'voided').length;

  return {
    transaction_count: rows.length,
    gross_sales: grossSales,
    discount_total: discountTotal,
    service_charge_total: serviceChargeTotal,
    tax_total: taxTotal,
    rounding_total: roundingTotal,
    net_sales: netSales,
    paid_amount: paidAmount,
    pending_amount: pendingAmount,
    refund_amount: refundAmount,
    void_count: voidCount,
    is_estimate: true,
  };
}

function createPreviewPageModel(input: {
  readonly state: Exclude<TransactionsPageState, 'real' | 'preview'>;
  readonly sourceLabel: string;
  readonly sourceTone: TransactionsPageModel['sourceTone'];
  readonly title: string;
  readonly description: string;
  readonly detail: string;
}): TransactionsPageModel {
  return {
    state: input.state,
    data: createPreviewViewModel(),
    sourceLabel: input.sourceLabel,
    sourceTone: input.sourceTone,
    title: input.title,
    description: input.description,
    detail: input.detail,
  };
}

function createPreviewViewModel(): TransactionsViewModel {
  return {
    metrics: transactionMetrics,
    lanes: transactionLanes,
    table: transactionsTable,
    detailLinks: transactionDetailLinks,
    rows: [],
    dataNotes: ['Data contoh fallback. Backend transaksi real belum terbaca untuk sesi ini.'],
    pagination: {
      page: 1,
      perPage: 20,
      total: transactionsTable.rows.length,
      totalPages: 1,
      hasNextPage: false,
      hasPreviousPage: false,
    },
  };
}

async function readSessionFromRequestCookies(): Promise<WebAdminSessionPayload | null> {
  const cookieStore = await cookies();
  const cookieValue = cookieStore.get(getSessionCookieName())?.value;

  if (!cookieValue) {
    return null;
  }

  return openSessionCookie(cookieValue);
}

function parseOptionalInteger(value: string | null, field: string, fallback: number, min: number, max: number): number {
  if (!value) {
    return fallback;
  }

  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed < min || parsed > max) {
    throw validationError({ [field]: [`${field} harus integer ${min}-${max}.`] });
  }

  return parsed;
}

function validationError(details: Record<string, readonly string[]>): BackendApiError {
  return new BackendApiError(422, 'VALIDATION_ERROR', 'Filter transaksi tidak valid.', details);
}

function belongsToSessionBusiness(row: BackendTransactionRow, session: WebAdminSessionPayload): boolean {
  const businessId = session.user.business_id ?? session.business?.id ?? null;

  return !row.business_id || !businessId || row.business_id === businessId;
}

function paymentMethods(row: BackendTransactionRow): string[] {
  const methods = [...new Set((row.payments ?? []).map((payment) => payment.method).filter((method): method is string => typeof method === 'string' && method.length > 0))];

  return methods;
}

function paymentMethodLabel(method: string): string {
  const labels: Record<string, string> = {
    cash: 'Tunai',
    qris: 'QRIS',
    card: 'Kartu',
    transfer: 'Transfer',
    ewallet: 'E-wallet',
    other: 'Lainnya',
    mixed: 'Campuran',
  };

  return labels[method] ?? method.toUpperCase();
}

function statusLabel(status: string): string {
  const labels: Record<string, string> = {
    paid: 'Lunas',
    partial: 'Sebagian',
    unpaid: 'Belum bayar',
    held: 'Ditahan',
    payment_pending: 'Menunggu bayar',
    payment_failed: 'Pembayaran gagal',
    voided: 'Void',
    refunded: 'Refunded',
    pending: 'Pending',
  };

  return labels[status] ?? status;
}

function money(value: unknown): number {
  return Number.isFinite(Number(value)) ? Number(value) : 0;
}

function isValidDate(value: string): boolean {
  return DATE_PATTERN.test(value) && !Number.isNaN(Date.parse(`${value}T00:00:00.000Z`));
}

function businessDate(value: string | null): string {
  const parsed = value ? new Date(value) : null;

  if (!parsed || Number.isNaN(parsed.getTime())) {
    return '1970-01-01';
  }

  return parsed.toISOString().slice(0, 10);
}

function timeLabel(value: string): string {
  const parsed = new Date(value);

  if (Number.isNaN(parsed.getTime())) {
    return '-';
  }

  return new Intl.DateTimeFormat('id-ID', {
    hour: '2-digit',
    minute: '2-digit',
    timeZone: 'Asia/Jakarta',
  }).format(parsed);
}

function formatRupiah(value: number): string {
  return `Rp${new Intl.NumberFormat('id-ID').format(value)}`;
}

function formatDateTime(value: string): string {
  const parsed = new Date(value);

  if (Number.isNaN(parsed.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
    timeZone: 'Asia/Jakarta',
  }).format(parsed);
}

function isTenantAdmin(session: WebAdminSessionPayload): boolean {
  return session.user.role === 'owner' || session.user.role === 'admin';
}

function unauthenticatedResponse(): NextResponse {
  return NextResponse.json(
    {
      error: {
        code: 'UNAUTHENTICATED',
        message: 'Sesi tidak valid atau sudah berakhir.',
        details: {},
      },
    },
    { status: 401, headers: SAFE_JSON_HEADERS },
  );
}

function forbiddenResponse(message: string): NextResponse {
  return NextResponse.json(
    {
      error: {
        code: 'FORBIDDEN',
        message,
        details: {},
      },
    },
    { status: 403, headers: SAFE_JSON_HEADERS },
  );
}
