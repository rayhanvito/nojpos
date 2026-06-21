import { expect, it } from 'vitest';
import { assertRealLoginUnavailable, previewAuthPolicy } from './session';

it('keeps browser credential storage disabled in preview mode', () => {
  expect(previewAuthPolicy).toEqual({
    mode: 'preview-only',
    browserCredentialStorage: 'disabled',
    approvedForRealLogin: false,
    note: 'Admin web preview does not store browser credentials or perform real login. Auth design is future-gated.',
  });
});

it('fails closed when real login is requested before approval', () => {
  expect(() => assertRealLoginUnavailable()).toThrow('Real admin-web login is not available in preview mode.');
});
