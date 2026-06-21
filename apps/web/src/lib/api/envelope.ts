export type ApiMeta = Record<string, unknown> | undefined;

export type ApiEnvelope<T> = { data: T; meta?: ApiMeta };

type ApiErrorEnvelope = {
  error: { code: string; message: string; details?: unknown };
};

export class ApiClientError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly status?: number,
    public readonly details?: unknown,
  ) {
    super(message);
    this.name = 'ApiClientError';
  }
}

export function parseApiEnvelope<T>(payload: unknown): ApiEnvelope<T> {
  if (isApiError(payload)) {
    throw new ApiClientError(payload.error.message, payload.error.code, undefined, payload.error.details);
  }

  if (!isApiSuccess(payload)) {
    throw new ApiClientError('Respons API tidak sesuai kontrak.', 'INVALID_ENVELOPE');
  }

  return payload as ApiEnvelope<T>;
}

function isApiError(payload: unknown): payload is ApiErrorEnvelope {
  return typeof payload === 'object' && payload !== null && 'error' in payload;
}

function isApiSuccess(payload: unknown): payload is ApiEnvelope<unknown> {
  return typeof payload === 'object' && payload !== null && 'data' in payload;
}
