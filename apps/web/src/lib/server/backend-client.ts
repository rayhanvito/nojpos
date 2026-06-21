import { randomUUID } from 'node:crypto';
import type { ApiEnvelope } from '../api/envelope';
import type { SafeOutletContext, WebAdminSessionPayload } from './session-cookie';

export type BackendApiErrorCode = 'UNAUTHENTICATED' | 'FORBIDDEN' | 'VALIDATION_ERROR' | 'BACKEND_UNAVAILABLE' | string;

export class BackendApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: BackendApiErrorCode,
    message: string,
    public readonly details: unknown = {},
  ) {
    super(message);
    this.name = 'BackendApiError';
  }
}

type BackendErrorEnvelope = {
  readonly error?: {
    readonly code?: unknown;
    readonly message?: unknown;
    readonly details?: unknown;
  };
};

type BackendLoginData = {
  readonly token: string;
  readonly user: WebAdminSessionPayload['user'];
  readonly business?: WebAdminSessionPayload['business'];
  readonly outlets?: readonly SafeOutletContext[];
};

type BackendMeData = {
  readonly user: WebAdminSessionPayload['user'];
  readonly business?: WebAdminSessionPayload['business'];
  readonly outlets?: readonly SafeOutletContext[];
  readonly permissions?: readonly string[];
};

type BackendOutletsData = {
  readonly outlets?: readonly SafeOutletContext[];
};

export type DashboardSummaryQuery = {
  readonly date?: string;
  readonly outlet_id?: string;
  readonly range?: 'last_7_days';
};

export type DashboardSummaryKpi = {
  readonly value: number;
  readonly type: 'money' | 'integer' | 'percent' | string;
  readonly label: string;
  readonly trend_label: string | null;
  readonly is_estimate: boolean;
};

export type DashboardSummaryData = {
  readonly meta: {
    readonly date: string;
    readonly range: string;
    readonly timezone: string;
    readonly business_id: string;
    readonly outlet_id: string | null;
    readonly outlet_name: string | null;
    readonly generated_at: string;
    readonly currency: string;
    readonly money_format: string;
    readonly data_status: 'real' | 'partial' | 'unavailable' | string;
    readonly is_empty_today: boolean;
  };
  readonly kpis: {
    readonly sales_today: DashboardSummaryKpi;
    readonly transaction_count: DashboardSummaryKpi;
    readonly average_transaction: DashboardSummaryKpi;
    readonly gross_profit_estimate: DashboardSummaryKpi;
    readonly low_stock_count: DashboardSummaryKpi;
    readonly cash_difference: DashboardSummaryKpi;
  };
  readonly alerts: readonly {
    readonly id: string;
    readonly type: string;
    readonly severity: string;
    readonly title: string;
    readonly message: string;
    readonly outlet_id: string | null;
    readonly outlet_name: string | null;
    readonly entity_type: string | null;
    readonly entity_id: string | null;
    readonly action_label: string | null;
    readonly action_path: string | null;
  }[];
  readonly sales_last_7_days: readonly {
    readonly date: string;
    readonly label: string;
    readonly sales: number;
    readonly transaction_count: number;
    readonly average_transaction: number;
  }[];
  readonly payment_methods: readonly {
    readonly method: string;
    readonly label: string;
    readonly amount: number;
    readonly transaction_count: number;
    readonly share_percent: number;
  }[];
  readonly top_products: readonly {
    readonly product_id: string | null;
    readonly name: string;
    readonly qty_sold: number;
    readonly sales: number;
    readonly share_percent: number;
    readonly outlet_id: string | null;
    readonly outlet_name: string | null;
  }[];
  readonly low_stock_items: readonly {
    readonly product_id: string;
    readonly name: string;
    readonly sku: string | null;
    readonly outlet_id: string;
    readonly outlet_name: string;
    readonly remaining_stock: number;
    readonly threshold: number | null;
    readonly unit: string | null;
    readonly status: 'low' | 'out' | string;
  }[];
  readonly recent_transactions: readonly {
    readonly transaction_id: string;
    readonly code: string;
    readonly time: string;
    readonly occurred_at: string;
    readonly outlet_id: string;
    readonly outlet_name: string;
    readonly cashier_name: string | null;
    readonly payment_method: string | null;
    readonly total: number;
    readonly status: string;
  }[];
  readonly cashier_performance: readonly {
    readonly cashier_id: string;
    readonly name: string;
    readonly transaction_count: number;
    readonly sales: number;
    readonly average_transaction: number;
    readonly void_count: number;
    readonly refund_count: number;
    readonly cash_difference: number | null;
    readonly note: string | null;
  }[];
  readonly branch_highlights: readonly {
    readonly outlet_id: string;
    readonly name: string;
    readonly status: string;
    readonly summary: string;
    readonly sales_today: number;
    readonly transaction_count: number;
    readonly low_stock_count: number;
    readonly open_shift_count: number;
    readonly cash_difference: number | null;
    readonly severity: string;
  }[];
  readonly data_notes: readonly {
    readonly key: string;
    readonly message: string;
    readonly severity: string;
  }[];
};

export type BackendClient = {
  readonly login: (input: { readonly email: string; readonly password: string; readonly deviceUuid?: string }) => Promise<BackendLoginData>;
  readonly logout: (token: string) => Promise<void>;
  readonly me: (token: string) => Promise<BackendMeData>;
  readonly outlets: (token: string) => Promise<BackendOutletsData>;
  readonly dashboardSummary: (token: string, query?: DashboardSummaryQuery) => Promise<ApiEnvelope<DashboardSummaryData>>;
};

export function createBackendClient(baseUrl = getBackendBaseUrl()): BackendClient {
  return {
    login: async (input) => {
      const envelope = await backendRequest<BackendLoginData>(baseUrl, '/auth/login', {
        method: 'POST',
        body: {
          email: input.email,
          password: input.password,
          device_uuid: input.deviceUuid ?? createWebLoginDeviceUuid(),
        },
      });

      if (!envelope.data.token || !envelope.data.user) {
        throw new BackendApiError(500, 'INVALID_BACKEND_RESPONSE', 'Backend login response is missing required session fields.');
      }

      return envelope.data;
    },
    logout: async (token) => {
      await backendRequest(baseUrl, '/auth/logout', {
        method: 'POST',
        token,
      });
    },
    me: async (token) => {
      const envelope = await backendRequest<BackendMeData>(baseUrl, '/me', {
        method: 'GET',
        token,
      });

      if (!envelope.data.user) {
        throw new BackendApiError(500, 'INVALID_BACKEND_RESPONSE', 'Backend me response is missing user context.');
      }

      return envelope.data;
    },
    outlets: async (token) => {
      const envelope = await backendRequest<BackendOutletsData>(baseUrl, '/outlets', {
        method: 'GET',
        token,
      });

      return envelope.data;
    },
    dashboardSummary: async (token, query = {}) => {
      const path = appendQuery('/dashboard/summary', query);

      return backendRequest<DashboardSummaryData>(baseUrl, path, {
        method: 'GET',
        token,
      });
    },
  };
}

export function getBackendBaseUrl(): string {
  const configured = process.env.BACKEND_INTERNAL_API_URL?.trim();

  if (!configured) {
    throw new BackendApiError(
      500,
      'BACKEND_CONFIG_MISSING',
      'BACKEND_INTERNAL_API_URL is required for Web Admin server-side API access.',
    );
  }

  return configured.replace(/\/+$/, '');
}

export function createWebLoginDeviceUuid(): string {
  const namespace = process.env.WEB_LOGIN_DEVICE_UUID_NAMESPACE?.trim() || 'web-admin';

  return `${namespace}:${randomUUID()}`;
}

export function toSafeErrorResponse(error: unknown): { readonly body: { readonly error: { readonly code: string; readonly message: string; readonly details: unknown } }; readonly status: number } {
  if (error instanceof BackendApiError) {
    return {
      status: error.status,
      body: {
        error: {
          code: normalizeErrorCode(error.status, error.code),
          message: error.message,
          details: error.details ?? {},
        },
      },
    };
  }

  return {
    status: 500,
    body: {
      error: {
        code: 'BACKEND_UNAVAILABLE',
        message: 'Server admin belum bisa menghubungi backend. Coba lagi nanti.',
        details: {},
      },
    },
  };
}

async function backendRequest<TData>(
  baseUrl: string,
  path: string,
  options: {
    readonly method: 'GET' | 'POST';
    readonly token?: string;
    readonly body?: Record<string, unknown>;
  },
): Promise<ApiEnvelope<TData>> {
  const response = await fetch(`${baseUrl}${path}`, {
    method: options.method,
    headers: createBackendHeaders(options.token),
    body: options.body ? JSON.stringify(options.body) : undefined,
    cache: 'no-store',
  });

  const payload = await readJson(response);

  if (!response.ok) {
    throw toBackendApiError(response.status, payload);
  }

  if (!isApiEnvelope<TData>(payload)) {
    throw new BackendApiError(500, 'INVALID_BACKEND_RESPONSE', 'Backend response does not match the expected API envelope.');
  }

  return payload;
}

function appendQuery(path: string, query: DashboardSummaryQuery): string {
  const params = new URLSearchParams();

  if (query.date) {
    params.set('date', query.date);
  }

  if (query.outlet_id) {
    params.set('outlet_id', query.outlet_id);
  }

  if (query.range) {
    params.set('range', query.range);
  }

  const serialized = params.toString();

  return serialized ? `${path}?${serialized}` : path;
}

function createBackendHeaders(token?: string): Headers {
  const headers = new Headers({
    Accept: 'application/json',
    'Content-Type': 'application/json',
  });

  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  return headers;
}

async function readJson(response: Response): Promise<unknown> {
  try {
    return await response.json();
  } catch {
    return null;
  }
}

function toBackendApiError(status: number, payload: unknown): BackendApiError {
  const envelope = payload as BackendErrorEnvelope | null;
  const backendError = envelope && typeof envelope === 'object' ? envelope.error : undefined;

  const code = typeof backendError?.code === 'string' ? backendError.code : normalizeErrorCode(status);
  const message = typeof backendError?.message === 'string' ? backendError.message : defaultErrorMessage(status);

  return new BackendApiError(status, code, message, backendError?.details ?? {});
}

function normalizeErrorCode(status: number, code?: string): string {
  if (code) {
    return code;
  }

  if (status === 401) {
    return 'UNAUTHENTICATED';
  }

  if (status === 403) {
    return 'FORBIDDEN';
  }

  if (status === 422) {
    return 'VALIDATION_ERROR';
  }

  return 'BACKEND_UNAVAILABLE';
}

function defaultErrorMessage(status: number): string {
  if (status === 401) {
    return 'Sesi tidak valid atau sudah berakhir.';
  }

  if (status === 403) {
    return 'Akses tidak diizinkan.';
  }

  if (status === 422) {
    return 'Data yang dikirim tidak valid.';
  }

  return 'Backend belum tersedia. Coba lagi nanti.';
}

function isApiEnvelope<TData>(payload: unknown): payload is ApiEnvelope<TData> {
  return typeof payload === 'object' && payload !== null && 'data' in payload;
}
