import { describe, expect, it } from 'vitest';
import { ApiClientError, parseApiEnvelope } from './envelope';

describe('parseApiEnvelope', () => {
  it('returns data and meta from the Laravel success envelope', () => {
    expect(parseApiEnvelope({ data: { total: 12500 }, meta: { outlet_id: 'outlet-1' } })).toEqual({
      data: { total: 12500 },
      meta: { outlet_id: 'outlet-1' },
    });
  });

  it('maps a Laravel forbidden error without inventing a domain rule', () => {
    expect(() => parseApiEnvelope({ error: { code: 'FORBIDDEN', message: 'Tidak diizinkan', details: null } })).toThrow(ApiClientError);
  });
});
