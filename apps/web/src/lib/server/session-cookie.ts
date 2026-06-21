import { createCipheriv, createDecipheriv, createHash, randomBytes, timingSafeEqual } from 'node:crypto';

const SESSION_COOKIE_VERSION = 'v1';
const AES_256_GCM_KEY_BYTES = 32;
const AES_GCM_IV_BYTES = 12;
const AES_GCM_TAG_BYTES = 16;

export const DEFAULT_WEB_SESSION_COOKIE_NAME = 'nojpos_web_session';
export const DEFAULT_WEB_SESSION_TTL_SECONDS = 60 * 60 * 8;

export type WebAdminRole = 'owner' | 'admin' | 'cashier' | 'superadmin' | string;

export type WebAdminSessionPayload = {
  readonly token: string;
  readonly user: {
    readonly id: string;
    readonly name: string;
    readonly email?: string | null;
    readonly role: WebAdminRole;
    readonly business_id?: string | null;
    readonly is_superadmin?: boolean;
  };
  readonly business?: {
    readonly id: string;
    readonly name: string;
  } | null;
  readonly device_uuid: string;
  readonly issued_at: number;
  readonly expires_at: number;
};

export type SafeUserContext = {
  readonly id: string;
  readonly name: string;
  readonly role: WebAdminRole;
  readonly business_id?: string | null;
  readonly is_superadmin: boolean;
};

export type SafeBusinessContext = {
  readonly id: string;
  readonly name: string;
};

export type SafeOutletContext = {
  readonly id: string;
  readonly business_id?: string | null;
  readonly name: string;
  readonly timezone?: string | null;
};

export type SafeWebAdminSessionContext = {
  readonly user: SafeUserContext;
  readonly business: SafeBusinessContext | null;
  readonly outlets: readonly SafeOutletContext[];
  readonly permissions: readonly string[];
  readonly access: {
    readonly tenant_admin: boolean;
    readonly platform_admin: boolean;
  };
  readonly redirect_to: '/dashboard' | '/platform';
};

export type WebSessionCookieOptions = {
  readonly name: string;
  readonly maxAgeSeconds: number;
  readonly expiresAt: Date;
  readonly httpOnly: true;
  readonly secure: boolean;
  readonly sameSite: 'lax';
  readonly path: '/';
};

export function getSessionCookieName(): string {
  return process.env.WEB_SESSION_COOKIE_NAME?.trim() || DEFAULT_WEB_SESSION_COOKIE_NAME;
}

export function getSessionTtlSeconds(): number {
  const configured = Number.parseInt(process.env.WEB_SESSION_TTL_SECONDS ?? '', 10);

  if (Number.isFinite(configured) && configured > 0) {
    return configured;
  }

  return DEFAULT_WEB_SESSION_TTL_SECONDS;
}

export function getSessionCookieOptions(now: Date = new Date()): WebSessionCookieOptions {
  const maxAgeSeconds = getSessionTtlSeconds();

  return {
    name: getSessionCookieName(),
    maxAgeSeconds,
    expiresAt: new Date(now.getTime() + maxAgeSeconds * 1000),
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
  };
}

export function createSessionPayload(input: {
  readonly token: string;
  readonly user: WebAdminSessionPayload['user'];
  readonly business?: WebAdminSessionPayload['business'];
  readonly device_uuid: string;
  readonly now?: Date;
}): WebAdminSessionPayload {
  const now = input.now ?? new Date();
  const issuedAt = Math.floor(now.getTime() / 1000);

  return {
    token: input.token,
    user: input.user,
    business: input.business ?? null,
    device_uuid: input.device_uuid,
    issued_at: issuedAt,
    expires_at: issuedAt + getSessionTtlSeconds(),
  };
}

export function sealSessionCookie(payload: WebAdminSessionPayload, secret = getRequiredSessionSecret()): string {
  validateSessionPayload(payload);

  const key = deriveEncryptionKey(secret);
  const iv = randomBytes(AES_GCM_IV_BYTES);
  const cipher = createCipheriv('aes-256-gcm', key, iv);
  const plaintext = Buffer.from(JSON.stringify(payload), 'utf8');
  const encrypted = Buffer.concat([cipher.update(plaintext), cipher.final()]);
  const tag = cipher.getAuthTag();

  return [
    SESSION_COOKIE_VERSION,
    base64UrlEncode(iv),
    base64UrlEncode(tag),
    base64UrlEncode(encrypted),
  ].join('.');
}

export function openSessionCookie(sealed: string | undefined | null, secret = getRequiredSessionSecret()): WebAdminSessionPayload | null {
  if (!sealed) {
    return null;
  }

  const [version, encodedIv, encodedTag, encodedEncrypted] = sealed.split('.');

  if (version !== SESSION_COOKIE_VERSION || !encodedIv || !encodedTag || !encodedEncrypted) {
    return null;
  }

  try {
    const iv = base64UrlDecode(encodedIv);
    const tag = base64UrlDecode(encodedTag);
    const encrypted = base64UrlDecode(encodedEncrypted);

    if (iv.length !== AES_GCM_IV_BYTES || tag.length !== AES_GCM_TAG_BYTES) {
      return null;
    }

    const key = deriveEncryptionKey(secret);
    const decipher = createDecipheriv('aes-256-gcm', key, iv);
    decipher.setAuthTag(tag);
    const plaintext = Buffer.concat([decipher.update(encrypted), decipher.final()]);
    const payload = JSON.parse(plaintext.toString('utf8')) as unknown;

    if (!isSessionPayload(payload)) {
      return null;
    }

    if (isSessionExpired(payload)) {
      return null;
    }

    return payload;
  } catch {
    return null;
  }
}

export function isSessionExpired(payload: Pick<WebAdminSessionPayload, 'expires_at'>, now: Date = new Date()): boolean {
  return payload.expires_at <= Math.floor(now.getTime() / 1000);
}

export function toSafeSessionContext(input: {
  readonly user: WebAdminSessionPayload['user'];
  readonly business?: WebAdminSessionPayload['business'];
  readonly outlets?: readonly unknown[];
  readonly permissions?: readonly unknown[];
}): SafeWebAdminSessionContext {
  const role = input.user.role;
  const isPlatformAdmin = role === 'superadmin' || input.user.is_superadmin === true;
  const isTenantAdmin = role === 'owner' || role === 'admin';

  return {
    user: {
      id: input.user.id,
      name: input.user.name,
      role,
      business_id: input.user.business_id ?? null,
      is_superadmin: isPlatformAdmin,
    },
    business: sanitizeBusiness(input.business),
    outlets: sanitizeOutlets(input.outlets ?? [], input.user.business_id ?? null, isTenantAdmin),
    permissions: sanitizePermissions(input.permissions ?? [role]),
    access: {
      tenant_admin: isTenantAdmin,
      platform_admin: isPlatformAdmin,
    },
    redirect_to: isPlatformAdmin ? '/platform' : '/dashboard',
  };
}

export function assertNoTokenInSafeContext(context: unknown): boolean {
  const serialized = JSON.stringify(context).toLowerCase();

  return !serialized.includes('token') && !serialized.includes('bearer');
}

function getRequiredSessionSecret(): string {
  const secret = process.env.WEB_SESSION_SECRET?.trim();

  if (!secret || secret.length < 32) {
    throw new Error('WEB_SESSION_SECRET must be at least 32 characters for Web Admin session sealing.');
  }

  return secret;
}

function deriveEncryptionKey(secret: string): Buffer {
  return createHash('sha256').update(secret).digest().subarray(0, AES_256_GCM_KEY_BYTES);
}

function validateSessionPayload(payload: WebAdminSessionPayload): void {
  if (!payload.token || !payload.user?.id || !payload.user?.name || !payload.user?.role || !payload.device_uuid) {
    throw new Error('Cannot seal an incomplete Web Admin session payload.');
  }
}

function isSessionPayload(payload: unknown): payload is WebAdminSessionPayload {
  if (typeof payload !== 'object' || payload === null) {
    return false;
  }

  const candidate = payload as Partial<WebAdminSessionPayload>;

  return (
    typeof candidate.token === 'string' &&
    typeof candidate.device_uuid === 'string' &&
    typeof candidate.issued_at === 'number' &&
    typeof candidate.expires_at === 'number' &&
    typeof candidate.user === 'object' &&
    candidate.user !== null &&
    typeof candidate.user.id === 'string' &&
    typeof candidate.user.name === 'string' &&
    typeof candidate.user.role === 'string'
  );
}

function sanitizeBusiness(business: WebAdminSessionPayload['business'] | undefined): SafeBusinessContext | null {
  if (!business || typeof business.id !== 'string' || typeof business.name !== 'string') {
    return null;
  }

  return {
    id: business.id,
    name: business.name,
  };
}

function sanitizeOutlets(outlets: readonly unknown[], businessId: string | null, allowTenantOutlets: boolean): SafeOutletContext[] {
  if (!allowTenantOutlets || !businessId) {
    return [];
  }

  return outlets.flatMap((outlet): SafeOutletContext[] => {
    if (typeof outlet !== 'object' || outlet === null) {
      return [];
    }

    const candidate = outlet as Partial<SafeOutletContext>;

    if (typeof candidate.id !== 'string' || typeof candidate.name !== 'string' || candidate.business_id !== businessId) {
      return [];
    }

    return [
      {
        id: candidate.id,
        business_id: candidate.business_id,
        name: candidate.name,
        timezone: typeof candidate.timezone === 'string' ? candidate.timezone : null,
      },
    ];
  });
}

function sanitizePermissions(permissions: readonly unknown[]): string[] {
  return permissions.filter((permission): permission is string => typeof permission === 'string');
}

function base64UrlEncode(value: Buffer): string {
  return value.toString('base64url');
}

function base64UrlDecode(value: string): Buffer {
  return Buffer.from(value, 'base64url');
}

export function safeConstantTimeEqual(left: string, right: string): boolean {
  const leftBuffer = Buffer.from(left);
  const rightBuffer = Buffer.from(right);

  if (leftBuffer.length !== rightBuffer.length) {
    return false;
  }

  return timingSafeEqual(leftBuffer, rightBuffer);
}
