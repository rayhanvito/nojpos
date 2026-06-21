# Subscription Read-only API Contract

Status: Contract ready, implementation pending.

## Purpose

Tenant Admin Web needs a read-only subscription page that shows plan status, period, limits, and usage. This phase must not enable plan or billing changes.

## Backend Endpoint Candidate

Use the existing subscription or business data source if available. If no tenant subscription read route exists yet, WEB-07B must render labelled fallback and record a backend gap instead of guessing.

Candidate target:

- `GET /api/v1/subscription`

## BFF Endpoint Target

`GET /api/admin/subscription`

Browser/UI must call only the Next.js BFF route or server-side helper.

## Auth And Roles

- Requires authenticated Web Admin session.
- Tenant `owner` and `admin` may read subscription status for their business.
- `cashier` must receive `403`.
- `superadmin` must use platform contracts, not this tenant endpoint.

## Query Parameters

None. Unknown parameters must return `422`.

## Response Shape

```json
{
  "data": {
    "meta": {
      "business_id": "biz_123",
      "generated_at": "2026-06-21T10:00:00.000Z",
      "source": "backend"
    },
    "subscription": {
      "plan_name": "Starter",
      "status": "active",
      "period_start": "2026-06-01",
      "period_end": "2026-06-30",
      "limits": {
        "outlets": 3,
        "staff": 10
      },
      "usage": {
        "outlets": 1,
        "staff": 4
      },
      "can_change_plan": false,
      "can_manage_billing": false
    },
    "data_notes": [
      {
        "key": "subscription_read_only",
        "message": "Subscription ditampilkan read-only. Plan dan billing changes tetap disabled.",
        "severity": "info"
      }
    ]
  },
  "meta": { "request_id": "req_123" }
}
```

## Field Rules

- `plan_name`: safe plan label.
- `status`: active/trialing/past_due/grace/cancelled/unknown.
- `period_start` and `period_end`: date or null.
- `limits`: safe numeric plan limits.
- `usage`: safe usage counters.
- `can_change_plan`: must be `false`.
- `can_manage_billing`: must be `false`.

Do not expose raw invoice payloads, provider secrets, or payment references.

## Error State

- `401 UNAUTHENTICATED` for missing or expired session.
- `403 FORBIDDEN` for wrong role or business.
- `422 VALIDATION_ERROR` for unknown query.
- `500 SERVER_ERROR` or safe backend unavailable response.

## Tenant Isolation

- Subscription must be scoped to authenticated `business_id`.
- Client may not pass or override `business_id`.
- Platform subscription management remains outside this tenant contract.

## Acceptance Criteria For WEB-07B

- BFF route `GET /api/admin/subscription` exists, or clear labelled fallback if backend route is unavailable.
- UI `/subscription` uses BFF/server helper only.
- Plan and billing changes remain disabled.
- No token or private billing data reaches browser, props, fixtures, logs, or error responses.
- Tests cover 401, 403, unknown query, no token leak, and fallback if backend route is unavailable.
