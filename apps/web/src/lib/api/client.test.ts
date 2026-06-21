import { describe, expect, it } from 'vitest';
import { assertApiIntegrationUnavailable, defineApiContract } from './client';

describe('pre-API contract marker', () => {
  it('describes endpoint intent without creating a runtime response', () => {
    expect(
      defineApiContract<{ rows: unknown[] }>({
        feature: 'reports.salesSummary',
        method: 'GET',
        path: '/api/v1/reports/sales-summary',
      }),
    ).toEqual({
      feature: 'reports.salesSummary',
      method: 'GET',
      path: '/api/v1/reports/sales-summary',
      response: null,
      status: 'future-gated',
      note: 'Preview mode only. No network client is available until API integration is explicitly approved.',
    });
  });

  it('fails closed when a preview route tries to use real integration', () => {
    expect(() => assertApiIntegrationUnavailable('reports.salesSummary')).toThrow(
      'reports.salesSummary is still in preview mode.',
    );
  });
});
