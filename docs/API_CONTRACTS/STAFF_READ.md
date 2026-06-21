# Staff Read-only API Contract

Status: Contract ready, implementation pending.

## Purpose

Tenant Admin Web needs a safe read-only staff list before any staff create/update/reset-password/ban action is allowed.

## Backend Endpoint Candidate

`GET /api/v1/staff`

Existing backend route is available and must remain read-only for this integration.

## BFF Endpoint Target

`GET /api/admin/staff`

Browser/UI must call only the Next.js BFF route. Laravel access and bearer token handling stay server-side.

## Auth And Roles

- Requires authenticated Web Admin session.
- Tenant `owner` and `admin` may read staff in their business.
- `cashier` must receive `403`.
- `superadmin` must not use tenant staff endpoint; platform users use platform-specific contracts.

## Query Parameters

Whitelist only:

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| `search` | string | No | Name/role search. Trim and limit length server-side/BFF. |
| `outlet_id` | UUID | No | Must belong to authenticated business. |
| `role` | enum | No | `owner`, `admin`, `cashier`, `staff` if backend supports it. |
| `status` | enum | No | `active`, `inactive`, `all`; default `all`. |
| `page` | integer | No | Default `1`. |
| `per_page` | integer | No | Default `20`, max `50`. |

Unknown parameters must return `422` from BFF before proxying.

## Response Shape

```json
{
  "data": {
    "meta": {
      "business_id": "biz_123",
      "outlet_id": null,
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
      "staff_count": 1,
      "active_count": 1,
      "admin_count": 0,
      "cashier_count": 1
    },
    "rows": [
      {
        "id": "user_123",
        "name": "Rina Putri",
        "role": "cashier",
        "status": "active",
        "outlet_ids": ["outlet_123"],
        "outlet_names": ["Outlet Pusat"],
        "last_activity_at": null,
        "can_edit": false,
        "can_reset_password": false,
        "can_delete": false
      }
    ],
    "data_notes": [
      {
        "key": "staff_read_only",
        "message": "Staff ditampilkan read-only. Reset password, ban, delete, dan perubahan role tetap disabled.",
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
| `id` | string | No | User/staff id scoped to business. |
| `name` | string | No | Safe display name. |
| `role` | string | No | Role label. |
| `status` | string | No | Active/inactive display state. |
| `outlet_ids` | string[] | No | Only outlets under same business. |
| `outlet_names` | string[] | No | Safe outlet names. |
| `last_activity_at` | ISO datetime | Yes | Null if unavailable. |
| `can_edit` | boolean | No | Must be `false` in read-only phase. |
| `can_reset_password` | boolean | No | Must be `false`. |
| `can_delete` | boolean | No | Must be `false`. |

Never expose password hash, PIN hash, raw PIN, auth token, remember token, full audit payload, or private contact fields unless a future contract explicitly allows them.

## Empty State

Return `200` with empty `rows`, zero totals, stable pagination, and data note explaining no staff matched the filter.

## Error State

- `401 UNAUTHENTICATED` for missing/expired web session.
- `403 FORBIDDEN` for cashier/superadmin/wrong business.
- `422 VALIDATION_ERROR` for invalid query/filter/outlet.
- `500 SERVER_ERROR` or safe backend unavailable response.

## Tenant Isolation

- All staff rows must be scoped to authenticated `business_id`.
- `outlet_id` must be validated as owned by the business.
- BFF must drop or reject any backend row whose `business_id` conflicts with session business.

## Acceptance Criteria For WEB-06B

- BFF route `GET /api/admin/staff` exists.
- UI `/staff` uses BFF/server helper only.
- No token in client component/props/fixtures/logs.
- Reset password/ban/delete/edit actions remain disabled.
- Tests cover 401, 403, unknown query, no token leak, and safe mapping.
