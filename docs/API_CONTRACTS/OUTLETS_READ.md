# Outlets Read-only API Contract

Status: Contract ready, implementation pending.

## Purpose

Tenant Admin Web needs a safe read-only outlet list and outlet context. This phase must not enable outlet create/update/delete, store open/close, cash reconciliation, or operational writes.

## Backend Endpoint Candidate

`GET /api/v1/outlets`

Existing backend auth route returns outlets for the authenticated business context.

## BFF Endpoint Target

`GET /api/admin/outlets`

Browser/UI must call only the Next.js BFF route. Laravel token handling remains server-side.

## Auth And Roles

- Requires authenticated Web Admin session.
- Tenant `owner` and `admin` may read outlets in their business.
- `cashier` must receive `403`.
- `superadmin` must not use tenant outlet endpoint.

## Query Parameters

Whitelist only:

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| `search` | string | No | Outlet name/address search if backend supports it; trim and limit length. |
| `status` | enum | No | `active`, `inactive`, `all`; default `all`. |
| `page` | integer | No | Default `1`. |
| `per_page` | integer | No | Default `20`, max `50`. |

Unknown parameters must return `422`.

## Response Shape

```json
{
  "data": {
    "meta": {
      "business_id": "biz_123",
      "generated_at": "2026-06-21T10:00:00.000Z",
      "source": "backend"
    },
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 1,
      "has_next_page": false
    },
    "totals": {
      "outlet_count": 1,
      "active_count": 1,
      "inactive_count": 0
    },
    "rows": [
      {
        "id": "outlet_123",
        "name": "Outlet Pusat",
        "address_short": "Jakarta",
        "timezone": "Asia/Jakarta",
        "status": "active",
        "is_default": true,
        "can_edit": false,
        "can_open_store": false,
        "can_close_store": false,
        "can_delete": false
      }
    ],
    "data_notes": [
      {
        "key": "outlets_read_only",
        "message": "Outlet ditampilkan read-only. Store open/close, edit, delete, dan rekonsiliasi kas tetap disabled.",
        "severity": "info"
      }
    ]
  },
  "meta": {
    "request_id": "req_123"
  }
}
```

## Field Rules

| Field | Type | Nullable | Notes |
| --- | --- | --- | --- |
| `id` | string | No | Outlet id scoped to business. |
| `name` | string | No | Safe outlet name. |
| `address_short` | string | Yes | Short display address only. |
| `timezone` | string | No | Default `Asia/Jakarta` if backend missing. |
| `status` | string | No | Active/inactive display state. |
| `is_default` | boolean | No | Display only. |
| `can_edit` | boolean | No | Must be `false` in read-only phase. |
| `can_open_store` | boolean | No | Must be `false`. |
| `can_close_store` | boolean | No | Must be `false`. |
| `can_delete` | boolean | No | Must be `false`. |

Do not expose store cash state, private operational notes, payment terminal secrets, or any write capability in this contract.

## Empty State

Return `200` with empty `rows`, zero totals, stable pagination, and a data note that no outlets matched the filter.

## Error State

- `401 UNAUTHENTICATED` for missing/expired web session.
- `403 FORBIDDEN` for cashier/superadmin/wrong business.
- `422 VALIDATION_ERROR` for invalid query/filter.
- `500 SERVER_ERROR` or safe backend unavailable response.

## Tenant Isolation

- All outlet rows must be scoped to authenticated `business_id`.
- BFF must drop or reject backend rows from other businesses.
- Client may not submit or override `business_id`.

## Acceptance Criteria For WEB-06B

- BFF route `GET /api/admin/outlets` exists.
- UI `/outlets` uses BFF/server helper only.
- Store open/close, edit, delete, cash reconciliation, and operational writes remain disabled.
- No token or operational secret leak.
- Tests cover 401, 403, unknown query, no token leak, and safe mapping.
