import type { ApiEnvelope } from './envelope';

export type ApiContractMethod = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';

export type ApiContractStatus = 'future-gated';

export type ApiContractDescriptor<TData = unknown> = {
  readonly feature: string;
  readonly method: ApiContractMethod;
  readonly path: string;
  readonly response: ApiEnvelope<TData> | null;
  readonly status: ApiContractStatus;
  readonly note: string;
};

export function defineApiContract<TData>(
  descriptor: Omit<ApiContractDescriptor<TData>, 'response' | 'status' | 'note'>,
): ApiContractDescriptor<TData> {
  return {
    ...descriptor,
    response: null,
    status: 'future-gated',
    note: 'Preview mode only. No network client is available until API integration is explicitly approved.',
  };
}

export function assertApiIntegrationUnavailable(feature: string): never {
  throw new Error(`${feature} is still in preview mode. Backend integration has not been approved.`);
}
