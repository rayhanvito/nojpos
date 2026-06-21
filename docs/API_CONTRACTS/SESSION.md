# Contract - Web Admin Session

Status: Session/BFF adapter scaffold ready, dashboard integration pending  
Owner: Integration + Web Admin + Backend  
Related stories: INT-01, WEB-01, BE-01  
Decision task: WEB-01A  
Implementation scaffold task: WEB-01B

## Goal

Web Admin needs a safe browser session for two separated surfaces:

- Tenant Admin at `/dashboard` and tenant routes for `owner` / `admin`.
- Platform Super Admin at `/platform` and platform routes for `superadmin` / internal roles.

The web app is still UI preview until the approved session boundary is implemented. Runtime integration must not start from pages/components directly.

## Current Backend Auth Findings

Backend routes under `/api/v1` currently use Laravel Sanctum token auth:

- Public:
  - `POST /api/v1/auth/login`
- Protected by `auth:sanctum`:
  - `POST /api/v1/auth/logout`
  - `GET /api/v1/me`
  - `GET /api/v1/outlets`
  - `GET /api/v1/subscription`
  - `GET /api/v1/settings`
  - `GET /api/v1/dashboard/summary`
  - `GET /api/v1/superadmin/*` routes behind `role:superadmin`

Current login behavior is token-based and mobile/POS oriented:

- `POST /api/v1/auth/login` requires `email`, `password`, and `device_uuid`.
- On success it creates a Sanctum personal access token and returns the plain token in the JSON `data.token` field.
- It may create or reuse a device row for the supplied `device_uuid`.
- Protected requests authenticate through `auth:sanctum`.
- Logout deletes the current access token.
- Role checks are enforced by backend role middleware such as `role:owner,admin` and `role:superadmin`.
- Tenant context is derived from the authenticated user `business_id`; clients must not send or choose tenant ownership.

## Options Compared

| Option | Summary | Pros | Cons | Decision |
| --- | --- | --- | --- | --- |
| A. Direct browser fetch to Laravel Sanctum cookie SPA | Browser uses Laravel Sanctum SPA cookies directly against Laravel API. | Mature Sanctum SPA pattern; no token JSON handled by web app. | Current backend login returns token and requires `device_uuid`; needs CSRF/cookie domain/CORS setup and backend changes before use. More moving parts for separate Next + Laravel origins. | Not chosen for this phase. Reconsider only if backend is changed to first-class SPA cookie auth. |
| B. Direct bearer token in browser | Browser stores/uses backend access token and components attach bearer auth. | Simple technically. | Exposes token to browser JavaScript and UI code, increases XSS blast radius, encourages auth logic in components, and violates Web Admin security rules. | Rejected. |
| C. Next.js BFF/server-side route handler | Browser only talks to same-origin Next routes under `/api/admin/*`. Next server talks to Laravel and keeps backend token/session material inside an HttpOnly server-bound cookie or server session. | Keeps backend token away from Client Components, centralizes error handling and role checks, avoids browser CORS for Laravel, fits current token backend without changing runtime backend now. | Requires careful cookie sealing, CSRF/origin checks for POST routes, and a small server-only adapter layer. | Chosen. |

## Chosen Strategy

Use a **Next.js BFF/server-side route handler strategy**.

Browser traffic must target only same-origin Next.js route handlers under `/api/admin/*`. Those route handlers are the only Web Admin boundary allowed to call Laravel API endpoints. The Laravel access token returned by backend login must never be exposed to Client Components, page props, browser storage, logs, fixtures, or UI state.

Recommended session storage for WEB-01B:

- Store a sealed session payload in an HttpOnly cookie, or store an opaque session id in an HttpOnly cookie with the token kept server-side.
- For the initial scaffold, prefer a sealed HttpOnly cookie because it works without a new external session store or package.
- The sealed payload may include only what the BFF needs to call Laravel, such as backend token, user id, role, business id, expiry timestamp, and minimal context.
- The browser-readable UI session endpoint must return only safe user/business/outlet/role data, never raw backend token or cookie internals.

## Why This Strategy

This strategy is chosen because current backend auth already returns a Sanctum token, while Web Admin needs browser-safe session behavior. A BFF lets the web app use the existing backend contract without placing backend credentials in browser JavaScript.

Security reasons:

- Browser storage must not be used for auth secrets.
- UI components must not build `Authorization` headers.
- Client Components must not receive raw backend tokens.
- Role and tenant checks stay server-side at both BFF and backend layers.
- All Laravel API access is centralized in server-only code, making boundary scans and reviews easier.

## Final Route Boundary

### Browser-facing Next.js routes

Implemented in WEB-01B for session scaffold unless marked future.

| Next route | Method | Purpose | Calls backend |
| --- | --- | --- | --- |
| `/api/admin/auth/login` | `POST` | Validate login form, call Laravel login, create HttpOnly session cookie. | `POST /api/v1/auth/login` |
| `/api/admin/auth/logout` | `POST` | Call Laravel logout if possible, then clear Web Admin session cookie. | `POST /api/v1/auth/logout` |
| `/api/admin/session` | `GET` | Return safe session bootstrap for shell routing. | `GET /api/v1/me`, `GET /api/v1/outlets` |
| `/api/admin/dashboard/summary` | `GET` | Future WEB-01C proxy owner/admin dashboard summary read. | `GET /api/v1/dashboard/summary` |

### Server-only library files

Implemented in WEB-01B:

- `apps/web/src/lib/server/session-cookie.ts`
- `apps/web/src/lib/server/backend-client.ts`
- `apps/web/src/lib/server/admin-session.ts`
- `apps/web/src/app/api/admin/auth/login/route.ts`
- `apps/web/src/app/api/admin/auth/logout/route.ts`
- `apps/web/src/app/api/admin/session/route.ts`

Future tasks may add route guards, role helpers, and dashboard proxy helpers without moving backend tokens into UI code.

Client-side UI code may call only the safe same-origin `/api/admin/*` routes or receive already-sanitized data from Server Components after the adapter is implemented. Client-side UI code must never call Laravel directly.

## Auth Flows

### 1. Login Flow

1. Browser submits email/password to Next route `POST /api/admin/auth/login`.
2. Next route validates the request body and checks Origin/CSRF rules for browser POST.
3. Next route generates or reuses a server-controlled browser device label/uuid for the backend login request.
4. Next route calls Laravel `POST /api/v1/auth/login` from the server.
5. Laravel validates credentials and returns user/business/outlet data plus a Sanctum token.
6. Next route validates the returned role:
   - `owner` / `admin` may enter Tenant Admin.
   - `superadmin` may enter Platform Admin.
   - `cashier` must be rejected for Web Admin.
7. Next route creates the HttpOnly web session cookie.
8. Next route returns only safe session data and suggested redirect target.

Safe response example from Next route:

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "name": "Owner Toko",
      "role": "owner",
      "business_id": "uuid",
      "is_superadmin": false
    },
    "business": {
      "id": "uuid",
      "name": "Kopi Senja"
    },
    "redirect_to": "/dashboard"
  },
  "meta": {
    "session_mode": "next_bff_http_only_cookie"
  }
}
```

### 2. Me / Session Bootstrap Flow

1. Web shell or route guard calls Next route `GET /api/admin/session`.
2. Next route reads and verifies the HttpOnly web session.
3. Next route calls Laravel `GET /api/v1/me` using the backend token server-side.
4. Next route may call `GET /api/v1/outlets` if outlet selection is needed.
5. Next route returns only sanitized session context:
   - user id, name, role, business id
   - business summary
   - outlet list safe fields
   - permissions/route access flags
   - default route
6. If backend returns `401`, Next clears the cookie and returns `401` with `UNAUTHENTICATED`.
7. If backend returns `403`, Next returns `403` and does not redirect users into another area.

### 3. Dashboard Summary Read Flow

1. Browser or Server Component requests Next route `GET /api/admin/dashboard/summary`.
2. Next route verifies the HttpOnly web session.
3. Next route enforces `owner` / `admin` before proxying.
4. Next route forwards only allowed query params: `date`, `outlet_id`, and `range`.
5. Next route calls Laravel `GET /api/v1/dashboard/summary` server-side.
6. Next route returns the backend envelope `{ data, meta }` unchanged except for safe error normalization if needed.

### 4. Logout Flow

1. Browser submits `POST /api/admin/auth/logout`.
2. Next route checks Origin/CSRF rules.
3. Next route reads the HttpOnly web session.
4. If a backend token exists, Next route calls Laravel `POST /api/v1/auth/logout` server-side.
5. Next route clears the Web Admin session cookie even if Laravel logout is already expired.
6. Next route returns `{ data: { logged_out: true }, meta: {} }`.

### 5. Expired Session / 401 Flow

- Any BFF route receiving Laravel `401` must clear the web session cookie.
- Tenant pages should redirect unauthenticated users to the login screen with a safe message such as `Sesi berakhir. Silakan login ulang.`
- Platform pages follow the same behavior and must not fall back to tenant preview data after auth expiry.

### 6. Forbidden Role / 403 Flow

- `/dashboard` and tenant routes require `owner` or `admin`.
- `/platform` and platform routes require `superadmin`.
- `cashier` must not access Web Admin.
- Forbidden users should see a forbidden state or be redirected to their allowed surface only when the session bootstrap clearly identifies an allowed surface.
- Never redirect a forbidden user into another tenant or leak existence of another tenant resource.

## Cookie / Session Rules

| Rule | Required behavior |
| --- | --- |
| Cookie visibility | Web session cookie must be `HttpOnly`. |
| Production transport | Cookie must be `Secure` in production. |
| SameSite | Use `SameSite=Lax` by default for same-site admin navigation. Use stricter settings if deployment allows. |
| Path | Scope cookie to `/` or `/api/admin` plus admin pages as needed. Avoid sharing with unrelated apps. |
| Expiry | Use a finite expiry. Initial recommendation: 8-12 hours absolute expiry or backend token expiry, whichever is sooner. |
| Refresh behavior | No silent refresh is required until backend supports a dedicated refresh flow. Expired sessions must ask user to login again. |
| Cookie contents | Prefer sealed/encrypted payload or opaque id. Never expose raw token to browser-readable JavaScript. |
| Logging | Do not log cookie values, backend tokens, full email, password, PII, payment reference, or receipt data. |

## CSRF Consideration

Because the browser talks to same-origin Next routes, Laravel CSRF cookies are not required for browser-to-Laravel communication in this strategy. However, Next route handlers that mutate auth/session state must protect browser POST routes.

Required for WEB-01B:

- Check `Origin` / `Host` on login and logout POST routes.
- Add a CSRF token or double-submit mechanism before adding broader write routes.
- Keep all early dashboard integration read-only.
- Do not add sensitive writes in the BFF scaffold.

## CORS Consideration

- Browser should not call Laravel directly, so browser CORS to Laravel is not required for Web Admin runtime.
- Next server-side route handlers call Laravel using a server-only base URL.
- Local dev needs a server-only backend URL such as `BACKEND_INTERNAL_API_URL=http://127.0.0.1:8000/api/v1`.
- Do not introduce public browser API base URL env vars for Web Admin integration.

## Environment / Config Needed Later

Server-only variables for WEB-01B:

| Variable | Purpose | Public? |
| --- | --- | --- |
| `BACKEND_INTERNAL_API_URL` | Laravel API base URL for Next server route handlers. | No |
| `WEB_SESSION_COOKIE_NAME` | Cookie name, e.g. `nojpos_web_session`. | No |
| `WEB_SESSION_SECRET` | Secret for sealing/signing cookie payloads. | No |
| `WEB_SESSION_TTL_SECONDS` | Session TTL override. | No |
| `WEB_LOGIN_DEVICE_UUID_NAMESPACE` | Optional stable prefix/namespace for browser login device ids. | No |

## Backend Requirements / Gaps

Current backend is enough to start a BFF scaffold because it already has:

- token login
- logout current token
- `/me`
- `/outlets`
- role middleware
- owner/admin dashboard summary endpoint
- superadmin route group

Gaps to track during implementation:

1. Login is currently mobile/POS oriented because it requires `device_uuid` and may create a device. WEB-01B must either generate a server-controlled web device uuid safely or request a future backend web-login variant if product owner wants no device row for web sessions.
2. Login returns the raw Sanctum token. With the chosen BFF strategy, this is acceptable only inside server route handlers and must never cross to UI/client code.
3. `/me` currently returns the fields available from backend. If web shell later needs richer permissions, add a backend contract task instead of guessing in the UI.
4. Cookie sealing/encryption must be implemented with existing platform APIs or approved dependencies only. Since this task forbids new dependencies, WEB-01B must use existing runtime capability or stop with a blocker.
5. Backend does not currently define a refresh endpoint. Web session expiry should be explicit re-login for now.

No backend runtime code should be changed as part of WEB-01A.

## API Wrapper Boundary Rules

Allowed in WEB-01B:

- Server-only adapter code under `apps/web/src/lib/server/**`.
- Next route handlers under `apps/web/src/app/api/admin/**`.
- Tests proving UI components do not access backend credentials.
- Boundary scan that blocks direct Laravel API calls from `src/app`, `src/components`, and fixtures.

Not allowed in WEB-01B unless the task explicitly expands scope:

- Replacing dashboard UI fixture with live data.
- Adding direct Laravel fetches from pages/components.
- Adding sensitive write routes.
- Adding new packages.
- Storing auth secrets in browser-readable storage.
- Emitting raw backend tokens into logs or test snapshots.

## Web Route Protection

| Route area | Required role | Behavior when unauthenticated | Behavior when forbidden |
| --- | --- | --- | --- |
| `/dashboard` and tenant routes | `owner`, `admin` | Redirect to login or show login-required state. | Show forbidden state or redirect only to known allowed platform area for `superadmin`. |
| `/platform` and platform routes | `superadmin` | Redirect to login or show login-required state. | Show forbidden state; do not expose platform data to tenant roles. |
| `/` landing | Public preview/marketing or redirect based on session if bootstrap is available. | N/A | N/A |

Role protection must happen server-side in the BFF and must still rely on backend role middleware as the final authority.

## Error Handling Flow

BFF routes should preserve the backend envelope shape where possible:

```json
{
  "error": {
    "code": "UNAUTHENTICATED",
    "message": "Sesi tidak valid atau sudah berakhir.",
    "details": {}
  }
}
```

Required mappings:

| Status | Meaning | Web behavior |
| --- | --- | --- |
| `401` | Session/token missing or expired. | Clear cookie, show login-required/expired session state. |
| `403` | Role, tenant, outlet, or policy denied. | Show forbidden state; do not retry as another role. |
| `422` | Invalid login or filter validation. | Show field/filter validation message. |
| `500` | Backend/server error. | Show unavailable state and avoid fixture fallback pretending to be live data. |

## Security Rules

- No browser storage for auth secrets.
- No raw backend token in Client Components, React props, fixtures, snapshots, logs, or browser-visible JSON.
- No `Authorization` header construction in `src/app/**` or `src/components/**`.
- Laravel backend access must live in server-only adapter or Next route handlers.
- Browser-facing routes return sanitized user/session data only.
- Role checks must happen before proxying tenant/platform data.
- Tenant context must come from backend authenticated user/session, never from browser-selected business id.
- Do not trust `outlet_id` until backend validates it against the authenticated business.
- Login/logout POST routes must have Origin/CSRF protection.
- Sensitive actions remain disabled and out of scope.

## Acceptance Criteria For WEB-01B

- Session/BFF adapter scaffold exists under the approved server-only boundary.
- Login route can call backend login server-side and create an HttpOnly web session cookie, or WEB-01B records a blocker if cookie sealing cannot be implemented safely with existing dependencies.
- Logout route clears the web session cookie and calls backend logout when possible.
- Session route returns safe bootstrap data without raw tokens.
- Cashier is rejected from Web Admin.
- Tenant route guard recognizes `owner` / `admin`.
- Platform route guard recognizes `superadmin`.
- Boundary scan confirms no auth secret storage and no direct backend auth in UI components.
- No dashboard data integration yet unless a later task explicitly changes the queue.
- No new packages or dependency changes.
- Tests/build/lint required by `apps/web/AGENTS.md` pass for the changed scaffold.

## Decision Outcome

WEB-01A final decision: **Next.js BFF/server-side route handler with HttpOnly sealed session cookie**.

WEB-01B implementation outcome:

- `POST /api/admin/auth/login` creates an encrypted/sealed HttpOnly cookie and returns safe user/business/outlet context only.
- `POST /api/admin/auth/logout` calls backend logout when a session exists, then clears the local web session cookie even if backend logout fails.
- `GET /api/admin/session` reads the sealed cookie, calls backend `/me` and `/outlets` server-side, and returns safe context only.
- Cashier role is rejected from Web Admin.
- Dashboard integration is not implemented yet.

Implementation must proceed in this order:

1. WEB-01B — implement Web Admin session/BFF adapter scaffold. `[DONE]`
2. WEB-01C — integrate dashboard read-only through the BFF. `[READY]`
3. Later read-only pages.
4. Safe writes.
5. Sensitive actions only after policy, idempotency, audit, and approval are ready.
