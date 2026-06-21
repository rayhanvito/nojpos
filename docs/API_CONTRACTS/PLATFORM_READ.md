# Contract - Platform Super Admin Read APIs

Status: Draft
Owner: Backend Platform + Web Admin
Related stories: BE-05, BE-06, WEB-03, WEB-04, WEB-05, WEB-06, INT-05

## Goal

Menyediakan endpoint read-only untuk Super Admin NojPOS dengan bahasa UI sederhana: Ringkasan, Toko, Paket, Langganan & Tagihan, Bantuan, Aktivitas, Sistem, Pengumuman.

Super Admin boleh melihat agregat platform, tetapi akses detail toko harus gated dan tercatat.

## Shared Rules

- Auth: required.
- Role: superadmin/internal operator only.
- Tenant detail data must be redacted unless endpoint explicitly allows summary.
- Transaction and GMV data are aggregate only.
- No sensitive action in this contract.
- All timestamps UTC in API; UI renders local.

## Endpoint 1 - Platform Overview

- Method: `GET`
- Path: `/api/v1/superadmin/overview`

Success:

```json
{
  "data": {
    "metrics": {
      "total_stores": 128,
      "active_stores": 94,
      "trial_stores": 21,
      "expired_stores": 9,
      "mrr": 18400000,
      "unpaid_bills": 7,
      "open_support_tickets": 12,
      "system_incidents": 1
    },
    "alerts": [
      {
        "id": "trial-expiring",
        "severity": "warning",
        "title": "4 toko trial akan habis minggu ini",
        "message": "Follow up toko yang aktif mencoba paket."
      }
    ],
    "activity_metrics": {
      "aggregate_transactions": 18240,
      "aggregate_gmv": 284500000,
      "payment_failures": 5,
      "active_stores_today": 67
    },
    "store_status_chart": [
      { "label": "Aktif", "value": 94 },
      { "label": "Trial", "value": 21 },
      { "label": "Expired", "value": 9 },
      { "label": "Suspended", "value": 4 }
    ],
    "billing_status_chart": [
      { "label": "Lunas", "value": 83 },
      { "label": "Belum dibayar", "value": 7 },
      { "label": "Gagal", "value": 5 }
    ]
  },
  "meta": {
    "privacy_note": "Data transaksi dan GMV ditampilkan secara agregat."
  }
}
```

## Endpoint 2 - Stores

- Method: `GET`
- Path: `/api/v1/superadmin/stores`
- Existing possible route to reconcile: `/api/v1/superadmin/businesses`

Query:

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `status` | string | optional | active, trial, expired, suspended, needs_review |
| `search` | string | optional | name/owner/email masked search |
| `page` | integer | optional | paginated |

Success item:

```json
{
  "id": "uuid",
  "name": "Kopi Senja",
  "owner_name": "Rama",
  "contact_masked": "ra***@example.com / 08******123",
  "status": "active",
  "plan": "Pro",
  "outlet_count": 3,
  "user_count": 12,
  "last_active_at": "2026-06-21T07:00:00Z",
  "billing_status": "lunas",
  "needs_review": false
}
```

## Endpoint 3 - Plans

- Method: `GET`
- Path: `/api/v1/superadmin/plans`

Success item:

```json
{
  "id": "uuid",
  "name": "Pro",
  "status": "preview",
  "price_label": "Belum final",
  "outlet_limit": 5,
  "user_limit": 20,
  "product_limit": 2000,
  "features": ["Kasir", "Produk & inventori", "Multi-cabang", "Laporan"],
  "store_count": 42
}
```

## Endpoint 4 - Billing And Subscriptions

- Method: `GET`
- Path: `/api/v1/superadmin/billing`

Query:

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `status` | string | optional | paid, unpaid, failed, trial, grace_period, cancelled |
| `page` | integer | optional | paginated |

Success item:

```json
{
  "id": "uuid",
  "store_name": "Kopi Senja",
  "plan": "Pro",
  "due_date": "2026-06-26",
  "payment_status": "unpaid",
  "amount": 199000,
  "updated_at": "2026-06-21T07:00:00Z"
}
```

## Endpoint 5 - Support Tickets

- Method: `GET`
- Path: `/api/v1/superadmin/support-tickets`

Success item:

```json
{
  "id": "SUP-1021",
  "store_name": "Kopi Senja",
  "issue": "Printer tidak keluar struk",
  "category": "Printer",
  "priority": "urgent",
  "status": "open",
  "assigned_to": "Ops NojPOS",
  "updated_at": "2026-06-21T07:00:00Z",
  "contact_masked": "08******123"
}
```

## Endpoint 6 - Activity Logs

- Method: `GET`
- Path: `/api/v1/superadmin/activity-logs`

Success item:

```json
{
  "id": "uuid",
  "time": "2026-06-21T07:00:00Z",
  "actor_name": "Support NojPOS",
  "actor_role": "support",
  "action": "Akses bantuan diminta",
  "store_name": "Kopi Senja",
  "ip_device": "Jakarta / Chrome",
  "level": "info",
  "reason": "Membantu cek printer"
}
```

## Endpoint 7 - System Status

- Method: `GET`
- Path: `/api/v1/superadmin/system-status`

Success item:

```json
{
  "service": "Webhook payment",
  "status": "perlu_dicek",
  "impact": "Sebagian pembayaran terlambat update",
  "updated_at": "2026-06-21T07:00:00Z",
  "owner": "Ops"
}
```

## Endpoint 8 - Announcements

- Method: `GET`
- Path: `/api/v1/superadmin/announcements`

Success item:

```json
{
  "id": "uuid",
  "title": "Maintenance dashboard malam ini",
  "target": "Semua toko",
  "type": "Maintenance",
  "channel": ["in-app", "email"],
  "status": "draft",
  "scheduled_at": "2026-06-22T15:00:00Z"
}
```

## Error Response

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Akses hanya untuk operator internal NojPOS.",
    "details": {}
  }
}
```

Possible codes:

- `UNAUTHENTICATED`
- `FORBIDDEN`
- `VALIDATION_ERROR`
- `NOT_FOUND`

## Client UI Mapping

| UI route | Endpoint |
| --- | --- |
| `/platform` | `/superadmin/overview` |
| `/platform/businesses` | `/superadmin/stores` or existing `/superadmin/businesses` after naming decision |
| `/platform/plans` | `/superadmin/plans` |
| `/platform/revenue` | `/superadmin/billing` |
| `/platform/support` | `/superadmin/support-tickets` |
| `/platform/audit` | `/superadmin/activity-logs` |
| `/platform/system-health` | `/superadmin/system-status` |
| `/platform/announcements` | `/superadmin/announcements` |

## Sensitive Actions Out Of Scope

- akses bantuan real
- suspend/delete toko
- manual mark paid
- retry payment
- reset password/ban user
- broadcast pengumuman

These require separate contracts with reason, approval, audit, idempotency, and tests.

## Tests Required

Backend:

- tenant user cannot access superadmin endpoints
- superadmin can access read endpoints
- aggregated data does not expose transaction detail
- contact fields masked
- pagination/filter validation

Web:

- loading state
- empty state
- forbidden state
- error state
- preview/disabled action states remain safe
- no horizontal overflow
