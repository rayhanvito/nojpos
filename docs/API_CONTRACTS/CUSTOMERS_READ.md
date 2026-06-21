# Customers Read-only API Contract

Status: Contract ready, implementation pending.

## Purpose

Tenant Admin Web needs a safe read-only customer list. This phase must not enable customer create, update, delete, loyalty mutation, export, or broad PII exposure.

## Backend Endpoint Candidate

`GET /api/v1/customers`

Existing backend route is available and must be consumed as read-only by Web Admin.

## BFF Endpoint Target

`GET /api/admin/customers`

Browser/UI must call only the Next.js BFF route. Laravel token handling remains server-side.

## Auth And Roles

- Requires authenticated Web Admin session.
- Tenant `owner` and `admin` may read customers in their business.
- `cashier` must receive `403`.
- `superadmin` must not use tenant customer endpoint.

## Query Parameters

Whitelist only:

| Name | Type | Required | Notes |
| --- | --- | --- | --- |
| `search` | string | No | Name/contact search if backend supports it; trim and limit length. |
| `group` | string | No | Customer group/segment if available. |
| `status` | enum | No | `active`, `archived`, `all`; default `all`. |
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
      "customer_count": 1,
      "active_count": 1,
      "archived_count": 0
    },
    "rows": [
      {
        "id": "cust_123",
        "name": "Pelanggan",
        "phone_masked": "0812****7890",
        "email_masked": "m***@example.invalid",
        "group": "member",
        "status": "active",
        "last_transaction_at": null,
        "can_edit": false,
        "can_delete": false,
        "can_export": false
      }
    ],
    "data_notes": [
      {
        "key": "customers_privacy",
        "message": "Customer ditampilkan read-only dengan kontak masked. Export dan delete tetap disabled.",
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
| `id` | string | No | Customer id scoped to business. |
| `name` | string | Yes | Safe display name; fallback `Pelanggan` if null. |
| `phone_masked` | string | Yes | Masked. Never expose raw contact in Web Admin list phase. |
| `email_masked` | string | Yes | Masked. Never expose raw contact in Web Admin list phase. |
| `group` | string | Yes | Segment label. |
| `status` | string | No | Active/archived display state. |
| `last_transaction_at` | ISO datetime | Yes | Null if unavailable. |
| `can_edit` | boolean | No | Must be `false` in read-only phase. |
| `can_delete` | boolean | No | Must be `false`. |
| `can_export` | boolean | No | Must be `false`. |

Never expose raw contact details, notes with sensitive content, addresses, payment references, tokens, or audit payloads in this contract.

## Empty State

Return `200` with empty `rows`, zero totals, stable pagination, and a data note that no customers matched the filter.

## Error State

- `401 UNAUTHENTICATED` for missing/expired web session.
- `403 FORBIDDEN` for cashier/superadmin/wrong business.
- `422 VALIDATION_ERROR` for invalid query/filter.
- `500 SERVER_ERROR` or safe backend unavailable response.

## Tenant Isolation

- All customer rows must be scoped to authenticated `business_id`.
- BFF must drop or reject backend rows from other businesses.
- Client may not submit or override `business_id`.

## Acceptance Criteria For WEB-06B

- BFF route `GET /api/admin/customers` exists.
- UI `/customers` uses BFF/server helper only.
- Customer contact details are masked in returned Web Admin payload.
- No token/PII leakage in client props/fixtures/logs.
- Add/edit/delete/export actions remain disabled.
- Tests cover 401, 403, unknown query, masking, no token leak, and safe mapping.
