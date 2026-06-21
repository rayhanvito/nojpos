# Contract - Web Admin Session

Status: Draft
Owner: Integration + Backend + Web Admin
Related stories: INT-01, WEB-01, BE-01

## Goal

Web Admin membutuhkan session bootstrap yang aman untuk Owner/Admin Toko dan Platform Super Admin. Keputusan final session belum boleh menaruh token di `localStorage` atau `sessionStorage`.

## Proposed Endpoints

### Login

- Method: `POST`
- Path: `/api/v1/auth/login`
- Auth: public
- Roles: returns authenticated principal role
- Tenant scope: derived server-side after credential validation

Request body:

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `email` | string | yes | email user/operator |
| `password` | string | yes | never logged |
| `device_name` | string | optional | browser label, no fingerprinting sensitive data |

Success:

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "name": "Owner Toko",
      "email_masked": "ow***@example.com",
      "role": "owner",
      "business_id": "uuid",
      "is_superadmin": false
    }
  },
  "meta": {
    "session_mode": "http_only_cookie_or_server_boundary"
  }
}
```

### Session Bootstrap

- Method: `GET`
- Path: `/api/v1/me`
- Auth: required
- Roles: owner, admin, support, finance, ops, superadmin
- Tenant scope: returned from authenticated principal, not client-selected

Success:

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "name": "Owner Toko",
      "email_masked": "ow***@example.com",
      "role": "owner",
      "business_id": "uuid",
      "is_superadmin": false
    },
    "business": {
      "id": "uuid",
      "name": "Kopi Senja",
      "status": "active",
      "plan": "Pro"
    },
    "permissions": ["dashboard.read", "products.read"],
    "default_outlet_id": "uuid"
  },
  "meta": {}
}
```

### Outlets

- Method: `GET`
- Path: `/api/v1/outlets`
- Auth: required
- Roles: owner, admin, manager if allowed
- Tenant scope: only outlets within authenticated business

Success:

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Cabang Utama",
      "timezone": "Asia/Jakarta",
      "status": "active"
    }
  ],
  "meta": {}
}
```

### Logout

- Method: `POST`
- Path: `/api/v1/auth/logout`
- Auth: required
- Effect: revoke current session/token only

## Error Response

```json
{
  "error": {
    "code": "UNAUTHENTICATED",
    "message": "Sesi tidak valid atau sudah berakhir.",
    "details": {}
  }
}
```

Possible codes:

- `UNAUTHENTICATED`
- `FORBIDDEN`
- `VALIDATION_ERROR`
- `ACCOUNT_DISABLED`
- `BUSINESS_SUSPENDED`

## Client Rules

- Web Admin must not store access token in localStorage/sessionStorage.
- UI components must not manually attach Authorization headers.
- Session adapter must centralize auth handling.
- Platform Super Admin and Tenant Admin shell must remain separated.
- Forbidden state must not redirect users into another tenant/platform area.

## Tests Required

Backend:

- login success/failure
- logout revokes current session only
- `/me` returns current principal and scoped business
- suspended business returns safe error
- tenant user cannot access superadmin route
- superadmin route cannot be used by tenant user

Web:

- unauthenticated state
- forbidden state
- owner/admin route guard
- superadmin route guard
- no token in storage
- no auth header usage in UI components

## Open Decision

Choose one final session mode before runtime integration:

1. Laravel Sanctum SPA cookie flow.
2. Next.js server boundary/BFF that stores token server-side.

Do not start runtime web integration until this decision is marked approved.
