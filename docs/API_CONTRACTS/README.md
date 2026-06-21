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
| `SESSION.md` | Draft needed | Web Admin | Web auth/session bootstrap |
| `DASHBOARD_SUMMARY.md` | Draft needed | Backend + Web | Dashboard owner/admin read-only |
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
