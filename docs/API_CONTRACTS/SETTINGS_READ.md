# Settings Read-only API Contract

Status: Contract ready, implementation pending.

## Purpose

Tenant Admin Web needs read-only settings pages for business profile, outlet defaults, payment display config, receipt display config, and security flags. This phase must not enable save/update actions.

## Backend Endpoint Candidates

- `GET /api/v1/settings/business`
- `GET /api/v1/settings/outlets`
- `GET /api/v1/settings/payments`
- `GET /api/v1/settings/receipt`

If a route is missing or not stable, BFF must return labelled fallback with `data_notes`.

## BFF Endpoint Targets

- `GET /api/admin/settings/business`
- `GET /api/admin/settings/outlets`
- `GET /api/admin/settings/payments`
- `GET /api/admin/settings/receipt`

Browser/UI must call only same-origin BFF routes or server-side helpers. Backend auth material remains server-side.

## Auth And Roles

- Requires authenticated Web Admin session.
- Tenant `owner` and `admin` may read settings for their business.
- `cashier` must receive `403`.
- `superadmin` must not use tenant settings endpoints.

## Query Parameters

No query parameters for initial read-only phase, except optional `outlet_id` on outlet/receipt reads if needed. Unknown parameters must return `422`.

## Business Settings Response

```json
{
  "data": {
    "meta": {
      "business_id": "biz_123",
      "generated_at": "2026-06-21T10:00:00.000Z",
      "source": "backend"
    },
    "business": {
      "id": "biz_123",
      "name": "Kopi Senja",
      "timezone": "Asia/Jakarta",
      "currency": "IDR",
      "tax_mode": "inclusive",
      "status": "active",
      "can_edit": false
    },
    "data_notes": [
      {
        "key": "settings_read_only",
        "message": "Settings ditampilkan read-only. Tombol simpan tetap disabled.",
        "severity": "info"
      }
    ]
  },
  "meta": { "request_id": "req_123" }
}
```

## Safe Field Rules

- Business: `id`, `name`, `timezone`, `currency`, `tax_mode`, `status`, `can_edit`.
- Outlets: `id`, `name`, `timezone`, `address_short`, `status`, display flags.
- Payments: `method`, `label`, `status`, `provider_label`, `is_enabled`, `masked_identifier`, `can_edit`.
- Receipt: header/footer text, outlet name, tax label, and print layout flags.

Do not include private provider config, raw auth material, payment references, or customer PII in this contract.

## Error State

- `401 UNAUTHENTICATED` for missing or expired session.
- `403 FORBIDDEN` for wrong role/business.
- `422 VALIDATION_ERROR` for invalid optional filters.
- `500 SERVER_ERROR` or safe backend unavailable response.

## Tenant Isolation

- All settings must be scoped to authenticated `business_id`.
- `outlet_id`, if accepted, must belong to the business.
- BFF must not allow client-provided `business_id`.

## Acceptance Criteria For WEB-07B

- BFF routes for settings reads exist.
- UI settings pages render backend data or labelled fallback.
- Save/update buttons remain disabled in this phase.
- No raw auth material reaches browser, props, fixtures, logs, or error responses.
- Tests cover 401, 403, query whitelist, masking, and no auth-material leakage.
