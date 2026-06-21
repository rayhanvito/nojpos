export type PreviewAuthPolicy = {
  readonly mode: 'preview-only';
  readonly browserCredentialStorage: 'disabled';
  readonly approvedForRealLogin: false;
  readonly note: string;
};

export const previewAuthPolicy: PreviewAuthPolicy = {
  mode: 'preview-only',
  browserCredentialStorage: 'disabled',
  approvedForRealLogin: false,
  note: 'Admin web preview does not store browser credentials or perform real login. Auth design is future-gated.',
};

export function assertRealLoginUnavailable(): never {
  throw new Error('Real admin-web login is not available in preview mode. Auth integration has not been approved.');
}
