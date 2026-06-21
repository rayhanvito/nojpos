import { cookies } from 'next/headers';
import { NextRequest, NextResponse } from 'next/server';

import type {
  DashboardAlert,
  DashboardBranchHighlight,
  DashboardCashierPerformance,
  DashboardKpiCard,
  DashboardLowStockItem,
  DashboardPaymentMethod,
  DashboardRecentTransaction,
  DashboardSalesPoint,
  DashboardTone,
  DashboardTopProduct,
} from '../../fixtures/preview/dashboard';
import {
  dashboardAlerts,
  dashboardBranchHighlights,
  dashboardCashierPerformance,
  dashboardKpiCards,
  dashboardLowStockItems,
  dashboardPaymentMethods,
  dashboardRecentTransactions,
  dashboardSalesLast7Days,
  dashboardTopProducts,
} from '../../fixtures/preview/dashboard';
import type { ApiEnvelope } from '../api/envelope';
import {
  BackendApiError,
  createBackendClient,
  toSafeErrorResponse,
  type DashboardSummaryData,
  type DashboardSummaryKpi,
  type DashboardSummaryQuery,
} from './backend-client';
import { clearSessionCookie, readSessionFromRequest } from './admin-session';
import { getSessionCookieName, openSessionCookie, type WebAdminSessionPayload } from './session-cookie';

const SAFE_JSON_HEADERS = {
  'Cache-Control': 'no-store',
} as const;

const ALLOWED_DASHBOARD_QUERY_PARAMS = new Set(['date', 'outlet_id', 'range']);
const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;
const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export type DashboardSummaryViewModel = {
  readonly kpiCards: readonly DashboardKpiCard[];
  readonly alerts: readonly DashboardAlert[];
  readonly salesLast7Days: readonly DashboardSalesPoint[];
  readonly paymentMethods: readonly DashboardPaymentMethod[];
  readonly topProducts: readonly DashboardTopProduct[];
  readonly lowStockItems: readonly DashboardLowStockItem[];
  readonly recentTransactions: readonly DashboardRecentTransaction[];
  readonly cashierPerformance: readonly DashboardCashierPerformance[];
  readonly branchHighlights: readonly DashboardBranchHighlight[];
  readonly dataNotes: readonly string[];
  readonly isEmptyToday: boolean;
};

export type DashboardPageState = 'real' | 'preview' | 'session_required' | 'forbidden' | 'error';

export type DashboardPageModel = {
  readonly state: DashboardPageState;
  readonly data: DashboardSummaryViewModel;
  readonly sourceLabel: string;
  readonly sourceTone: DashboardTone;
  readonly title: string;
  readonly description: string;
  readonly detail: string;
  readonly contextBadges: readonly { readonly label: string; readonly tone: DashboardTone }[];
  readonly backendMeta?: DashboardSummaryData['meta'];
};

export async function handleAdminDashboardSummary(request: NextRequest): Promise<NextResponse> {
  const session = readSessionFromRequest(request);

  if (!session) {
    const response = unauthenticatedResponse();
    clearSessionCookie(response);

    return response;
  }

  if (!isTenantAdmin(session)) {
    return forbiddenResponse('Dashboard toko hanya untuk owner atau admin.');
  }

  try {
    const query = parseDashboardSummaryQuery(request.nextUrl.searchParams);
    const envelope = await createBackendClient().dashboardSummary(session.token, query);

    return NextResponse.json(envelope, { headers: SAFE_JSON_HEADERS });
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

export async function getDashboardPageModel(): Promise<DashboardPageModel> {
  const session = await readSessionFromRequestCookies();

  if (!session) {
    return createPreviewPageModel({
      state: 'session_required',
      sourceLabel: 'Data contoh fallback',
      sourceTone: 'warning',
      title: 'Dashboard membutuhkan sesi admin toko',
      description: 'Silakan login saat auth UI aktif. Untuk sementara, tampilan di bawah memakai data contoh yang diberi label jelas.',
      detail: 'Sesi admin belum tersedia atau sudah berakhir.',
    });
  }

  if (!isTenantAdmin(session)) {
    return createPreviewPageModel({
      state: 'forbidden',
      sourceLabel: 'Akses ditolak',
      sourceTone: 'danger',
      title: 'Dashboard toko hanya untuk owner/admin',
      description: 'Role saat ini tidak boleh membuka dashboard tenant. Tidak ada data backend toko yang ditampilkan.',
      detail: 'Jika kamu superadmin, gunakan area Platform. Jika kamu kasir, gunakan aplikasi kasir.',
    });
  }

  try {
    const envelope = await createBackendClient().dashboardSummary(session.token);
    const viewModel = mapDashboardSummaryToViewModel(envelope.data);

    return {
      state: 'real',
      data: viewModel,
      sourceLabel: envelope.data.meta.data_status === 'partial' ? 'Data backend parsial' : 'Data backend',
      sourceTone: envelope.data.meta.data_status === 'partial' ? 'warning' : 'success',
      title: envelope.data.meta.is_empty_today ? 'Belum ada transaksi hari ini' : 'Ringkasan toko hari ini',
      description: envelope.data.meta.is_empty_today
        ? 'Backend sudah terhubung, tetapi belum ada transaksi pada tanggal dashboard.'
        : 'Pantau penjualan, stok, kas, dan aktivitas toko dari data backend read-only.',
      detail: `Diperbarui ${formatDateTime(envelope.data.meta.generated_at)} · ${envelope.data.meta.timezone}`,
      contextBadges: [
        { label: envelope.data.meta.date, tone: 'neutral' },
        { label: envelope.data.meta.outlet_name ?? 'Semua cabang', tone: 'neutral' },
        { label: envelope.data.meta.data_status === 'partial' ? 'Backend parsial' : 'Backend real', tone: envelope.data.meta.data_status === 'partial' ? 'warning' : 'success' },
      ],
      backendMeta: envelope.data.meta,
    };
  } catch (error) {
    if (error instanceof BackendApiError && error.status === 401) {
      return createPreviewPageModel({
        state: 'session_required',
        sourceLabel: 'Sesi berakhir',
        sourceTone: 'warning',
        title: 'Sesi admin sudah berakhir',
        description: 'Silakan login ulang saat auth UI aktif. Data real tidak ditampilkan tanpa sesi valid.',
        detail: 'Data contoh fallback ditampilkan hanya untuk menjaga layout tetap terbaca.',
      });
    }

    if (error instanceof BackendApiError && error.status === 403) {
      return createPreviewPageModel({
        state: 'forbidden',
        sourceLabel: 'Akses ditolak',
        sourceTone: 'danger',
        title: 'Backend menolak akses dashboard',
        description: 'Role atau scope sesi ini tidak diizinkan membuka dashboard tenant.',
        detail: 'Data contoh fallback ditampilkan dengan label agar tidak tercampur dengan data real.',
      });
    }

    return createPreviewPageModel({
      state: 'error',
      sourceLabel: 'Backend belum tersedia',
      sourceTone: 'warning',
      title: 'Dashboard backend belum bisa dibaca',
      description: 'Ada kendala saat menghubungi backend dashboard summary. Data produksi tidak ditampilkan.',
      detail: error instanceof BackendApiError ? error.message : 'Server admin belum bisa menghubungi backend. Coba lagi nanti.',
    });
  }
}

export function parseDashboardSummaryQuery(searchParams: URLSearchParams): DashboardSummaryQuery {
  const unknownParams = [...searchParams.keys()].filter((key) => !ALLOWED_DASHBOARD_QUERY_PARAMS.has(key));

  if (unknownParams.length > 0) {
    throw new BackendApiError(422, 'VALIDATION_ERROR', 'Filter dashboard tidak valid.', {
      query: [`Parameter tidak didukung: ${unknownParams.join(', ')}`],
    });
  }

  const date = searchParams.get('date') ?? undefined;
  const outletId = searchParams.get('outlet_id') ?? undefined;
  const range = searchParams.get('range') ?? undefined;

  if (date && (!DATE_PATTERN.test(date) || Number.isNaN(Date.parse(`${date}T00:00:00.000Z`)))) {
    throw new BackendApiError(422, 'VALIDATION_ERROR', 'Filter dashboard tidak valid.', {
      date: ['Format date harus YYYY-MM-DD.'],
    });
  }

  if (outletId && !UUID_PATTERN.test(outletId)) {
    throw new BackendApiError(422, 'VALIDATION_ERROR', 'Filter dashboard tidak valid.', {
      outlet_id: ['outlet_id harus UUID valid.'],
    });
  }

  if (range && range !== 'last_7_days') {
    throw new BackendApiError(422, 'VALIDATION_ERROR', 'Filter dashboard tidak valid.', {
      range: ['range hanya boleh last_7_days.'],
    });
  }

  return {
    ...(date ? { date } : {}),
    ...(outletId ? { outlet_id: outletId } : {}),
    ...(range ? { range: 'last_7_days' } : {}),
  };
}

export function mapDashboardSummaryEnvelopeToViewModel(envelope: ApiEnvelope<DashboardSummaryData>): DashboardSummaryViewModel {
  return mapDashboardSummaryToViewModel(envelope.data);
}

export function mapDashboardSummaryToViewModel(summary: DashboardSummaryData): DashboardSummaryViewModel {
  const topProductMaxShare = Math.max(1, ...summary.top_products.map((product) => product.share_percent));

  return {
    kpiCards: [
      mapKpi(summary.kpis.sales_today, 'success'),
      mapKpi(summary.kpis.transaction_count, 'info', ' transaksi'),
      mapKpi(summary.kpis.average_transaction, 'neutral'),
      mapKpi(summary.kpis.gross_profit_estimate, 'success'),
      mapKpi(summary.kpis.low_stock_count, 'warning', ' item'),
      mapKpi(summary.kpis.cash_difference, summary.kpis.cash_difference.value === 0 ? 'neutral' : 'warning'),
    ],
    alerts: summary.alerts.map((alert) => ({
      title: alert.title,
      description: alert.message,
      tone: severityToTone(alert.severity),
    })),
    salesLast7Days: summary.sales_last_7_days.map((point) => ({
      day: point.label,
      sales: point.sales,
      label: formatCompactRupiah(point.sales),
    })),
    paymentMethods: summary.payment_methods.map((method) => ({
      method: method.label,
      total: method.amount,
      label: formatCompactRupiah(method.amount),
    })),
    topProducts: summary.top_products.map((product) => ({
      name: product.name,
      quantity: `${product.qty_sold} terjual`,
      total: formatRupiah(product.sales),
      share: clampPercent((product.share_percent / topProductMaxShare) * 100),
    })),
    lowStockItems: summary.low_stock_items.map((item) => ({
      name: item.name,
      stock: `${item.remaining_stock}${item.unit ? ` ${item.unit}` : ''}`,
      status: item.status === 'out' ? 'Habis' : 'Hampir habis',
    })),
    recentTransactions: summary.recent_transactions.map((transaction) => ({
      code: transaction.code,
      time: transaction.time,
      method: transaction.payment_method ?? '-',
      total: formatRupiah(transaction.total),
    })),
    cashierPerformance: summary.cashier_performance.map((cashier) => ({
      name: cashier.name,
      transactions: `${cashier.transaction_count} trx`,
      total: formatRupiah(cashier.sales),
      note: cashier.note ?? `Rata-rata ${formatRupiah(cashier.average_transaction)}`,
    })),
    branchHighlights: summary.branch_highlights.map((branch) => ({
      name: branch.name,
      status: branch.status,
      summary: branch.summary,
      tone: severityToTone(branch.severity),
    })),
    dataNotes: summary.data_notes.map((note) => note.message),
    isEmptyToday: summary.meta.is_empty_today,
  };
}

function createPreviewPageModel(input: {
  readonly state: Exclude<DashboardPageState, 'real' | 'preview'>;
  readonly sourceLabel: string;
  readonly sourceTone: DashboardTone;
  readonly title: string;
  readonly description: string;
  readonly detail: string;
}): DashboardPageModel {
  return {
    state: input.state,
    data: createPreviewViewModel(),
    sourceLabel: input.sourceLabel,
    sourceTone: input.sourceTone,
    title: input.title,
    description: input.description,
    detail: input.detail,
    contextBadges: [
      { label: 'Fallback preview', tone: 'warning' },
      { label: 'Bukan data produksi', tone: 'neutral' },
      { label: input.state === 'forbidden' ? '403' : input.state === 'session_required' ? '401' : 'Error', tone: input.sourceTone },
    ],
  };
}

function createPreviewViewModel(): DashboardSummaryViewModel {
  return {
    kpiCards: dashboardKpiCards,
    alerts: dashboardAlerts,
    salesLast7Days: dashboardSalesLast7Days,
    paymentMethods: dashboardPaymentMethods,
    topProducts: dashboardTopProducts,
    lowStockItems: dashboardLowStockItems,
    recentTransactions: dashboardRecentTransactions,
    cashierPerformance: dashboardCashierPerformance,
    branchHighlights: dashboardBranchHighlights,
    dataNotes: ['Data contoh fallback. Backend real belum terbaca untuk sesi ini.'],
    isEmptyToday: false,
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

function isTenantAdmin(session: WebAdminSessionPayload): boolean {
  return session.user.role === 'owner' || session.user.role === 'admin';
}

function mapKpi(kpi: DashboardSummaryKpi, tone: DashboardTone, integerSuffix = ''): DashboardKpiCard {
  return {
    label: kpi.label,
    value: kpi.type === 'money' ? formatRupiah(kpi.value) : `${formatInteger(kpi.value)}${integerSuffix}`,
    helper: kpi.is_estimate ? `${kpi.label} masih berlabel estimasi dari backend.` : `${kpi.label} dari backend read-only.`,
    trend: kpi.trend_label ?? (kpi.is_estimate ? 'Estimasi backend' : 'Data backend'),
    tone,
  };
}

function severityToTone(severity: string): DashboardTone {
  if (severity === 'danger') {
    return 'danger';
  }

  if (severity === 'warning') {
    return 'warning';
  }

  if (severity === 'success') {
    return 'success';
  }

  if (severity === 'info') {
    return 'info';
  }

  return 'neutral';
}

function formatRupiah(value: number): string {
  return `Rp${new Intl.NumberFormat('id-ID').format(value)}`;
}

function formatInteger(value: number): string {
  return new Intl.NumberFormat('id-ID').format(value);
}

function formatCompactRupiah(value: number): string {
  if (value >= 1000000) {
    return `Rp${(value / 1000000).toFixed(value % 1000000 === 0 ? 0 : 1).replace('.', ',')} jt`;
  }

  if (value >= 1000) {
    return `Rp${Math.round(value / 1000)} rb`;
  }

  return formatRupiah(value);
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

function clampPercent(value: number): number {
  return Math.max(0, Math.min(100, Math.round(value)));
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
