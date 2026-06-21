# Contract - Tenant Catalog Read-only

Status: Contract ready, implementation pending.  
Owner: Web Admin + Backend Core  
Related stories: WEB-04, INT-04  
Task: WEB-04A — Catalog/products/categories read-only contract

## Goal

Menyediakan kontrak read-only untuk halaman Tenant Admin `/catalog` agar owner/admin toko dapat melihat produk dan kategori secara aman melalui Next.js BFF sebelum fitur create/update/delete katalog dibuka.

Kontrak ini tidak mengaktifkan product create/update/delete, category create/update/delete, price update, stock adjustment, import/export, atau publish action. Semua write action tetap disabled/gated sampai fase safe write resmi.

## Endpoint Backend Nyata

Endpoint berikut sudah ada di Laravel dan protected oleh `auth:sanctum`:

| Method | Path | Status | Notes |
| --- | --- | --- | --- |
| `GET` | `/api/v1/products` | exists | Mengembalikan `categories` dan `products` dalam satu payload. Tenant-scoped dari authenticated `business_id`. |
| `GET` | `/api/v1/categories` | exists | Mengembalikan kategori tenant. Tenant-scoped dari authenticated `business_id`. |

Backend juga memiliki endpoint write `POST/PUT/DELETE /products` dan `POST/PUT/DELETE /categories`. Web Admin read-only integration tidak boleh memanggil endpoint write tersebut.

## Endpoint BFF Target

Browser/UI hanya boleh memakai endpoint same-origin Next.js atau server-side helper yang sama:

| Method | Path | Purpose | WEB task |
| --- | --- | --- | --- |
| `GET` | `/api/admin/catalog/products` | List produk read-only untuk `/catalog`. | WEB-04B |
| `GET` | `/api/admin/catalog/categories` | List kategori read-only/filter options untuk `/catalog`. | WEB-04B |

BFF wajib membaca sealed HttpOnly session cookie sesuai `SESSION.md`, memanggil Laravel server-side, dan tidak pernah mengirim token/session raw ke browser.

## Auth, Role, dan Scope

- Harus login dengan session Web Admin valid.
- Hanya tenant `owner` dan `admin` yang boleh akses.
- `cashier` harus mendapat `403 FORBIDDEN` dari BFF.
- `superadmin` tidak memakai endpoint tenant catalog ini; platform memakai endpoint platform/superadmin sendiri.
- Semua data wajib scoped ke `business_id` dari user/session server-side.
- Client tidak boleh mengirim `business_id`.
- `outlet_id`, jika dikirim, wajib divalidasi sebagai outlet milik business user oleh backend atau BFF guard.
- `category_id`, jika dikirim, wajib kategori milik business user. Karena backend product endpoint saat ini memakai filter `category` by name, BFF boleh filter `category_id` dari payload secara server-side sampai backend filter final tersedia.
- Field internal seperti cost/modal, supplier secret, audit payload, token, dan payment references tidak boleh dikirim ke browser.

## Query Parameters - Products

Target BFF: `GET /api/admin/catalog/products`

| Field | Type | Required | Default | Validation | Notes |
| --- | --- | --- | --- | --- | --- |
| `search` | string | no | empty | max 100 chars | Cari nama produk, SKU/barcode aman. |
| `category_id` | UUID string | no | all categories | UUID valid dan harus milik business user | BFF dapat filter dari payload jika backend belum punya filter id. |
| `outlet_id` | UUID string | no | all/global products | UUID valid dan harus milik business user | Backend products supports `outlet_id`; global product `outlet_id = null` tetap boleh muncul. |
| `status` | string enum | no | `all` | `all`, `active`, `inactive`, `archived` | Backend saat ini tidak punya status final di product list; BFF harus memakai safe fallback/data note. |
| `page` | integer | no | `1` | min `1` | Pagination BFF response. |
| `per_page` | integer | no | `20` | min `1`, max `100` | Jangan return list besar ke browser. |

Parameter lain harus ditolak dengan `422 VALIDATION_ERROR`.

## Query Parameters - Categories

Target BFF: `GET /api/admin/catalog/categories`

| Field | Type | Required | Default | Validation | Notes |
| --- | --- | --- | --- | --- | --- |
| `search` | string | no | empty | max 100 chars | Cari nama kategori. |
| `status` | string enum | no | `all` | `all`, `active`, `inactive`, `archived` | Jika backend belum punya status, BFF return `active` safe fallback dengan data note. |
| `page` | integer | no | `1` | min `1` | Pagination BFF response. |
| `per_page` | integer | no | `50` | min `1`, max `100` | Kategori biasanya kecil, tetapi tetap paginated. |

## Products Success Response JSON

```json
{
  "data": {
    "meta": {
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "currency": "IDR",
      "data_status": "partial",
      "generated_at": "2026-06-21T08:30:00.000000Z"
    },
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 2,
      "total_pages": 1,
      "has_next_page": false,
      "has_previous_page": false
    },
    "totals": {
      "product_count": 2,
      "active_count": 2,
      "inactive_count": 0,
      "tracked_stock_count": 1,
      "category_count": 1,
      "is_estimate": true
    },
    "rows": [
      {
        "product_id": "c313ba03-7c3a-49b7-bcf1-37ad94a2a001",
        "sku": "KOPI-001",
        "name": "Kopi Susu",
        "category": {
          "id": "0a1c8c52-8c27-428d-94bc-97c9f50fa001",
          "name": "Minuman"
        },
        "outlet": {
          "id": null,
          "name": "Semua cabang"
        },
        "price": 18000,
        "price_label": "Rp18.000",
        "track_stock": true,
        "stock_summary": {
          "status": "unknown",
          "label": "Stok lihat di Inventory",
          "quantity": null,
          "is_estimate": true
        },
        "status": "active",
        "status_label": "Aktif",
        "updated_at": null,
        "can_edit": false,
        "can_delete": false
      }
    ],
    "data_notes": [
      {
        "key": "catalog_status_partial",
        "message": "Backend product list saat ini belum menyediakan status aktif/arsip dan updated_at final; BFF memakai fallback aman sampai backend difinalkan.",
        "severity": "info"
      }
    ]
  },
  "meta": {
    "request_id": "req_01J0CATALOG001",
    "generated_at": "2026-06-21T08:30:00.000000Z"
  }
}
```

## Product Field Definition

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `product_id` | UUID string | no | `c313...` | Tenant-scoped product id. |
| `sku` | string | yes | `KOPI-001` | Backend currently exposes `barcode`; BFF may map to SKU display. |
| `name` | string | no | `Kopi Susu` | Product display name. |
| `category.id` | UUID string | yes | `0a1c...` | Nullable for uncategorized products. |
| `category.name` | string | yes | `Minuman` | Safe category name. |
| `outlet.id` | UUID string | yes | `1351...` | `null` means global/all outlets. |
| `outlet.name` | string | no | `Semua cabang` | Do not leak other tenant outlet names. |
| `price` | integer rupiah | no | `18000` | Server value only; client only formats display. |
| `price_label` | string | no | `Rp18.000` | Optional BFF convenience label; price integer remains source. |
| `track_stock` | boolean | no | `true` | Product stock tracking flag. |
| `stock_summary.status` | string enum | no | `unknown` | `in_stock`, `low`, `out`, `not_tracked`, `unknown`. |
| `stock_summary.quantity` | integer | yes | `12` | Nullable if product endpoint has no stock quantity. |
| `status` | string enum | no | `active` | `active`, `inactive`, `archived`, `unknown`. |
| `updated_at` | ISO-8601 UTC string | yes | `2026-06-21T08:20:00.000000Z` | Nullable if backend does not expose it. |
| `can_edit` | boolean | no | `false` | Must remain `false` until safe write contract is approved. |
| `can_delete` | boolean | no | `false` | Must remain `false` until safe write contract is approved. |

## Categories Success Response JSON

```json
{
  "data": {
    "meta": {
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "data_status": "partial",
      "generated_at": "2026-06-21T08:30:00.000000Z"
    },
    "pagination": {
      "page": 1,
      "per_page": 50,
      "total": 1,
      "total_pages": 1,
      "has_next_page": false,
      "has_previous_page": false
    },
    "rows": [
      {
        "category_id": "0a1c8c52-8c27-428d-94bc-97c9f50fa001",
        "name": "Minuman",
        "product_count": 12,
        "status": "active",
        "status_label": "Aktif",
        "updated_at": null,
        "can_edit": false,
        "can_delete": false
      }
    ],
    "data_notes": []
  },
  "meta": { "request_id": "req_01J0CATEGORIES001" }
}
```

## Category Field Definition

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `category_id` | UUID string | no | `0a1c...` | Tenant-scoped category id. |
| `name` | string | no | `Minuman` | Safe display name. |
| `product_count` | integer | no | `12` | BFF may compute from product list if backend category endpoint lacks count. |
| `status` | string enum | no | `active` | `active`, `inactive`, `archived`, `unknown`. |
| `updated_at` | ISO-8601 UTC string | yes | `2026-06-21T08:20:00.000000Z` | Nullable if backend does not expose it. |
| `can_edit` | boolean | no | `false` | Must remain `false`. |
| `can_delete` | boolean | no | `false` | Must remain `false`. |

## Empty State

If no products match filters, return stable shape with `rows: []`, totals zero, pagination totals zero, and `data_notes` explaining whether backend returned empty data or filters excluded all data. UI copy: `Belum ada produk yang cocok dengan filter.`

If no categories exist, return `rows: []` for categories and allow product rows to show uncategorized items with `category: null`.

## Error State

| HTTP | Code | When | Browser behavior |
| --- | --- | --- | --- |
| `401` | `UNAUTHENTICATED` | Missing/expired session | Show session-required state and do not call Laravel from UI. |
| `403` | `FORBIDDEN` | Role not owner/admin or tenant scope denied | Show forbidden state. |
| `422` | `VALIDATION_ERROR` | Unknown query param, invalid enum/date/UUID/page | Show filter error safely. |
| `500` | `BACKEND_UNAVAILABLE` | Laravel unavailable/unexpected response | Show friendly error card and optional preview fallback labeled clearly. |

## Data Source Candidates

- `products`: id, business_id, outlet_id, product_category_id, name, barcode/SKU, price, track_stock, deleted_at.
- `product_categories`: id, business_id, name, deleted_at.
- `outlets`: id/name validation and display for outlet-specific products.
- `stock_movements` or inventory endpoint: optional stock summary if WEB-04B chooses to enrich products; do not calculate final stock in Client Component.

## Backend Gaps / Data Notes

- Existing `GET /api/v1/products` returns all products/categories without backend pagination; BFF must paginate/sanitize for Web Admin until backend hardening exists.
- Existing product list supports `search`, `barcode`, `category` by name, and `outlet_id`; it does not yet support `category_id`, status, or updated_at filtering directly.
- Existing product list does not expose stock quantity summary. Use `unknown`/`not_tracked` labels or a server-side inventory helper later; do not invent stock numbers.
- Existing product/category write routes exist but are out of scope and must not be called by Web Admin read-only integration.

## Tenant Isolation Rules

- All data must be scoped to authenticated user `business_id`.
- Client must never send or override `business_id`.
- `outlet_id` and `category_id` filters must be validated against the same business scope.
- BFF must not expose `business_id` in each row unless contract explicitly needs meta; row-level tenant ids are unnecessary for UI.
- Product/category rows from other business must be dropped and logged only with safe identifiers if detected.

## UI Mapping For `/catalog`

- KPI/summary cards: product_count, active_count, tracked_stock_count, category_count.
- Table columns: product name, SKU, category, price, stock status, status, outlet, updated_at.
- Category filter options: categories endpoint rows.
- Buttons: `Tambah produk`, `Edit`, `Hapus`, `Import`, `Export` remain disabled/preview-only with explicit label until safe write contract.
- Preview fixture may remain only as labeled fallback when session/backend unavailable.

## Acceptance Criteria For WEB-04B

- Add BFF routes `GET /api/admin/catalog/products` and `GET /api/admin/catalog/categories`.
- BFF reads sealed session, allows only owner/admin, rejects cashier/superadmin.
- BFF whitelists query params and returns `422` for unknown/invalid filters.
- BFF calls Laravel only server-side and never returns token/session raw.
- `/catalog` renders backend data when session/backend succeeds.
- `/catalog` shows session-required, forbidden, empty, validation, and backend unavailable states.
- Product/category write actions remain disabled/preview-only.
- Tests cover no session, forbidden role, query validation, mapping shape, and no token in response.
- Web validation passes: typecheck, lint, tests, build, Playwright list, boundary scan.
