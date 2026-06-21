import { NextRequest, NextResponse } from 'next/server';
import { BackendApiError, createBackendClient, toSafeErrorResponse } from './backend-client';
import {
  assertNoTokenInSafeContext,
  createSessionPayload,
  getSessionCookieName,
  getSessionCookieOptions,
  openSessionCookie,
  sealSessionCookie,
  toSafeSessionContext,
  type SafeWebAdminSessionContext,
  type WebAdminSessionPayload,
} from './session-cookie';

const SAFE_JSON_HEADERS = {
  'Cache-Control': 'no-store',
} as const;

type LoginRequestBody = {
  readonly email?: unknown;
  readonly password?: unknown;
};

export type LoginResponseData = {
  readonly user: SafeWebAdminSessionContext['user'];
  readonly business: SafeWebAdminSessionContext['business'];
  readonly outlets: SafeWebAdminSessionContext['outlets'];
  readonly access: SafeWebAdminSessionContext['access'];
  readonly redirect_to: SafeWebAdminSessionContext['redirect_to'];
};

export async function handleAdminLogin(request: NextRequest): Promise<NextResponse> {
  const originError = assertAllowedOrigin(request);
  if (originError) {
    return originError;
  }

  try {
    const body = (await request.json()) as LoginRequestBody;
    const validation = validateLoginBody(body);

    if (validation) {
      return validation;
    }

    const email = body.email;
    const password = body.password;

    if (typeof email !== 'string' || typeof password !== 'string') {
      return validationResponse({ email: ['Email wajib diisi.'], password: ['Password wajib diisi.'] });
    }

    const backend = createBackendClient();
    const login = await backend.login({ email, password });
    const role = login.user.role;

    if (role === 'cashier') {
      return forbiddenResponse('Kasir tidak boleh mengakses Web Admin.');
    }

    if (role !== 'owner' && role !== 'admin' && role !== 'superadmin') {
      return forbiddenResponse('Role user belum diizinkan untuk Web Admin.');
    }

    const payload = createSessionPayload({
      token: login.token,
      user: login.user,
      business: login.business ?? null,
      device_uuid: crypto.randomUUID(),
    });
    const sealed = sealSessionCookie(payload);
    const safeContext = toSafeSessionContext({
      user: login.user,
      business: login.business ?? null,
      outlets: login.outlets ?? [],
      permissions: [role],
    });

    if (!assertNoTokenInSafeContext(safeContext)) {
      throw new Error('Unsafe session context attempted to expose token material.');
    }

    const response = NextResponse.json(
      {
        data: toLoginResponseData(safeContext),
        meta: {
          session_mode: 'next_bff_http_only_cookie',
        },
      },
      { headers: SAFE_JSON_HEADERS },
    );
    const options = getSessionCookieOptions();
    response.cookies.set(options.name, sealed, {
      httpOnly: options.httpOnly,
      secure: options.secure,
      sameSite: options.sameSite,
      path: options.path,
      maxAge: options.maxAgeSeconds,
      expires: options.expiresAt,
    });

    return response;
  } catch (error) {
    const safeError = toSafeErrorResponse(error);

    return NextResponse.json(safeError.body, { status: safeError.status, headers: SAFE_JSON_HEADERS });
  }
}

export async function handleAdminLogout(request: NextRequest): Promise<NextResponse> {
  const originError = assertAllowedOrigin(request);
  if (originError) {
    return originError;
  }

  const session = readSessionFromRequest(request);

  if (session?.token) {
    try {
      await createBackendClient().logout(session.token);
    } catch {
      // Logout must clear the local browser session even if the backend token is already expired.
    }
  }

  const response = NextResponse.json(
    {
      data: { logged_out: true },
      meta: {},
    },
    { headers: SAFE_JSON_HEADERS },
  );
  clearSessionCookie(response);

  return response;
}

export async function handleAdminSession(request: NextRequest): Promise<NextResponse> {
  const session = readSessionFromRequest(request);

  if (!session) {
    const response = unauthenticatedResponse();
    clearSessionCookie(response);

    return response;
  }

  try {
    const backend = createBackendClient();
    const me = await backend.me(session.token);
    const outletData = await backend.outlets(session.token);
    const safeContext = toSafeSessionContext({
      user: me.user,
      business: me.business ?? session.business ?? null,
      outlets: outletData.outlets ?? me.outlets ?? [],
      permissions: me.permissions ?? [me.user.role],
    });

    if (!assertNoTokenInSafeContext(safeContext)) {
      throw new Error('Unsafe session context attempted to expose token material.');
    }

    return NextResponse.json(
      {
        data: safeContext,
        meta: {
          session_mode: 'next_bff_http_only_cookie',
        },
      },
      { headers: SAFE_JSON_HEADERS },
    );
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

export function readSessionFromRequest(request: NextRequest): WebAdminSessionPayload | null {
  const cookieValue = request.cookies.get(getSessionCookieName())?.value;

  if (!cookieValue) {
    return null;
  }

  return openSessionCookie(cookieValue);
}

export function clearSessionCookie(response: NextResponse): void {
  response.cookies.set(getSessionCookieName(), '', {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    maxAge: 0,
    expires: new Date(0),
  });
}

function validateLoginBody(body: LoginRequestBody): NextResponse | null {
  if (typeof body.email !== 'string' || typeof body.password !== 'string') {
    return validationResponse({ email: ['Email wajib diisi.'], password: ['Password wajib diisi.'] });
  }

  if (!body.email.includes('@') || body.password.length < 1) {
    return validationResponse({ email: ['Format email tidak valid.'], password: ['Password wajib diisi.'] });
  }

  return null;
}

function assertAllowedOrigin(request: NextRequest): NextResponse | null {
  const origin = request.headers.get('origin');
  const host = request.headers.get('host');

  if (!origin || !host) {
    return validationResponse({ origin: ['Origin dan Host wajib tersedia untuk request browser.'] });
  }

  try {
    const originHost = new URL(origin).host;

    if (originHost === host) {
      return null;
    }
  } catch {
    return validationResponse({ origin: ['Origin tidak valid.'] });
  }

  return forbiddenResponse('Origin request tidak diizinkan.');
}

function toLoginResponseData(context: SafeWebAdminSessionContext): LoginResponseData {
  return {
    user: context.user,
    business: context.business,
    outlets: context.outlets,
    access: context.access,
    redirect_to: context.redirect_to,
  };
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

function validationResponse(details: unknown): NextResponse {
  return NextResponse.json(
    {
      error: {
        code: 'VALIDATION_ERROR',
        message: 'Request Web Admin tidak valid.',
        details,
      },
    },
    { status: 422, headers: SAFE_JSON_HEADERS },
  );
}
