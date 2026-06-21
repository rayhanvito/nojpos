# API Contracts

Folder ini menyimpan kontrak API sebelum backend endpoint diimplementasikan atau sebelum web/mobile mengonsumsi endpoint.

## Aturan

1. Contract dibuat sebelum implementasi backend atau integrasi client.
2. Contract harus jelas tentang auth, role, tenant scope, request, response, error, pagination, dan audit.
3. Client tidak boleh menebak bentuk response dari backend.
4. Jika backend berubah, contract harus ikut diperbarui di PR/branch yang sama.
5. Sensitive write harus punya bagian audit, idempotency, dan rollback.

## Contract Aktif

| Contract | Status | Dipakai oleh | Catatan |
| --- | --- | --- | --- |
| `SESSION.md` | Decision ready, implementation pending | Web Admin | Next.js BFF/server-side route handler with HttpOnly sealed session cookie |
| `DASHBOARD_SUMMARY.md` | Contract ready, implementation complete | Backend + Web | Dashboard owner/admin read-only endpoint implemented in BE-04B |
| `TRANSACTIONS_READ.md` | Contract ready, implementation complete | Backend + Web | Tenant Admin transactions read-only list through BFF `/api/admin/transactions` |
| `INVENTORY_READ.md` | Contract ready, implementation complete | Backend + Web | Tenant Admin inventory read-only list through BFF `/api/admin/inventory` |
| `CATALOG_READ.md` | Contract ready, implementation complete | Backend + Web | Tenant Admin products/categories read-only through BFF `/api/admin/catalog/*` |
| `REPORTS_READ.md` | Contract ready, implementation complete | Backend + Web | Tenant Admin reports read-only through BFF `/api/admin/reports/*` |
| `STAFF_READ.md` | Contract ready, implementation pending | Backend + Web | Tenant Admin staff read-only through BFF `/api/admin/staff` |
| `CUSTOMERS_READ.md` | Contract ready, implementation pending | Backend + Web | Tenant Admin customers read-only through BFF `/api/admin/customers` |
| `OUTLETS_READ.md` | Contract ready, implementation pending | Backend + Web | Tenant Admin outlets read-only through BFF `/api/admin/outlets` |
| `SETTINGS_READ.md` | Contract ready, implementation pending | Backend + Web | Tenant Admin settings read-only |
| `SUBSCRIPTION_READ.md` | Contract ready, implementation pending | Backend + Web | Tenant Admin subscription read-only |
| `PLATFORM_READ.md` | Draft needed | Backend Platform + Web | Super Admin read-only |

## Template Ringkas

```md
# Contract - <Name>

Status: Draft / Approved / Implemented
Owner: <agent/team>
Related stories: <IDs>

## Endpoint

- Method:
- Path:
- Auth:
- Roles:
- Tenant scope:
- Rate limit:

## Request

Query/body params:

| Field | Type | Required | Notes |
| --- | --- | --- | --- |

## Success Response

```json
{
  "data": {},
  "meta": {}
}
```

## Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "...",
    "details": {}
  }
}
```

## Empty State

## Forbidden State

## Audit And Idempotency

## Client UI Mapping

## Tests Required
```
