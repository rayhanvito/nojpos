# Contract - Tenant Inventory Read-only

Status: Contract ready, implementation pending.  
Owner: Backend Core + Web Admin  
Related stories: BE-03, WEB-03, INT-03  
Task: WEB-03A — Finalize Inventory read-only contract

## Goal

Menyediakan kontrak read-only untuk halaman Tenant Admin `/inventory` agar owner/admin toko dapat melihat stok saat ini, pergerakan stok, dan transfer dalam perjalanan secara aman melalui Next.js BFF.

Kontrak ini tidak mengaktifkan stock adjustment, purchase, count, waste, transfer create/send/receive/cancel, export, atau action sensitif lain. Web Admin hanya boleh membaca data yang sudah dihitung backend.

## Endpoint Backend Nyata

Endpoint berikut sudah ada di Laravel dan protected oleh `auth:sanctum`:

| Method | Path | Status | Notes |
| --- | --- | --- | --- |
| `GET` | `/api/v1/inventory` | exists | Mengembalikan overview stok dan movements ringkas. Tenant-scoped by authenticated `business_id`. |
| `GET` | `/api/v1/inventory/movements` | exists | Mengembalikan stock movements. Tenant-scoped by authenticated `business_id`. |
| `GET` | `/api/v1/inventory/transfers` | exists | Mengembalikan transfer inventory. Tenant-scoped by authenticated `business_id`. |
| `GET` | `/api/v1/inventory/transfers/in-transit` | exists | Mengembalikan transfer berstatus `in_transit`. Tenant-scoped by authenticated `business_id`. |

Backend juga memiliki endpoint write inventory seperti purchases, counts, waste, transfer create/send/receive/cancel. Endpoint write tersebut tidak boleh dipanggil oleh Web Admin read-only integration.

## Endpoint BFF Target

Browser/UI hanya boleh memakai endpoint same-origin Next.js atau server-side helper yang sama:

| Method | Path | Purpose | WEB task |
| --- | --- | --- | --- |
| `GET` | `/api/admin/inventory` | List stok saat ini untuk `/inventory`. | WEB-03B |
| `GET` | `/api/admin/inventory/movements` | List pergerakan stok read-only. | WEB-03B atau follow-up kecil setelah list |
| `GET` | `/api/admin/inventory/transfers/in-transit` | Transfer dalam perjalanan read-only. | Optional follow-up jika UI membutuhkan panel transfer |

BFF wajib membaca sealed HttpOnly session cookie sesuai `SESSION.md`, memanggil Laravel server-side, dan tidak pernah mengirim token/session raw ke browser.

## Auth, Role, dan Scope

- Harus login dengan session Web Admin valid.
- Hanya tenant `owner` dan `admin` yang boleh akses.
- `cashier` harus mendapat `403 FORBIDDEN` dari BFF.
- `superadmin` tidak memakai endpoint tenant inventory ini; platform memakai endpoint platform/superadmin sendiri.
- Semua data wajib scoped ke `business_id` dari user/session server-side.
- Client tidak boleh mengirim `business_id`.
- `outlet_id`, jika dikirim, wajib divalidasi sebagai outlet milik business user.
- `category_id`, jika dipakai, wajib kategori milik business user.
- BFF wajib melakukan sanitization dan tidak boleh meneruskan cost, supplier, audit, atau internal data yang tidak diperlukan browser.

## Query Parameters - Inventory List

Target BFF: `GET /api/admin/inventory`

| Field | Type | Required | Default | Validation | Notes |
| --- | --- | --- | --- | --- | --- |
| `outlet_id` | UUID string | no | all allowed outlets | UUID valid dan milik business user | Kosong berarti semua outlet yang diizinkan. |
| `category_id` | UUID string | no | all categories | UUID valid dan milik business user | Backend saat ini memakai `product_category_id`; BFF boleh filter dari payload sampai backend filter final. |
| `stock_status` | string enum | no | `all` | `all`, `in_stock`, `low`, `out`, `negative`, `not_tracked` | `low` membutuhkan threshold; jika threshold belum tersedia, BFF wajib memberi data note. |
| `search` | string | no | empty | max 100 chars | Cari nama produk atau SKU/barcode aman. |
| `page` | integer | no | `1` | min `1` | Pagination BFF response. |
| `per_page` | integer | no | `20` | min `1`, max `100` | Jangan return list besar ke browser. |

Invalid filter harus menghasilkan `422 VALIDATION_ERROR` dengan detail field yang aman.

## Query Parameters - Movements

Target BFF: `GET /api/admin/inventory/movements`

| Field | Type | Required | Default | Validation | Notes |
| --- | --- | --- | --- | --- | --- |
| `outlet_id` | UUID string | no | all allowed outlets | UUID valid dan milik business user | Scope outlet. |
| `product_id` | UUID string | no | all products | UUID valid dan milik business user | Filter produk. |
| `type` | string enum | no | all movement types | `sale`, `purchase`, `count`, `waste`, `transfer_out`, `transfer_in`, `adjustment`, `return`, `other` | BFF boleh map dari backend type yang tersedia. |
| `from` | string | no | none | `YYYY-MM-DD` | Awal tanggal movement. |
| `to` | string | no | none | `YYYY-MM-DD`, tidak boleh sebelum `from` | Akhir tanggal movement. |
| `page` | integer | no | `1` | min `1` | Pagination BFF response. |
| `per_page` | integer | no | `20` | min `1`, max `100` | Backend saat ini limit by `per_page`; BFF tetap stabilkan pagination. |

## Inventory Success Response JSON

```json
{
  "data": {
    "meta": {
      "timezone": "Asia/Jakarta",
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "currency": "IDR",
      "quantity_format": "integer_units",
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
      "tracked_product_count": 2,
      "low_stock_count": 1,
      "out_of_stock_count": 0,
      "negative_stock_count": 0,
      "in_transit_count": 1,
      "is_estimate": true
    },
    "rows": [
      {
        "product_id": "c313ba03-7c3a-49b7-bcf1-37ad94a2a001",
        "sku": "SKU-KOPI-001",
        "name": "Kopi Susu",
        "category": {
          "id": "0a1c8c52-8c27-428d-94bc-97c9f50fa001",
          "name": "Minuman"
        },
        "outlet": {
          "id": "135159a6-d523-470f-93e1-5a3f3706a001",
          "name": "Cabang Utama"
        },
        "stock_on_hand": 8,
        "available_stock": 8,
        "reserved_stock": 0,
        "in_transit_out": 0,
        "in_transit_in": 4,
        "unit": "pcs",
        "low_stock_threshold": 10,
        "track_stock": true,
        "status": "low",
        "status_label": "Stok hampir habis",
        "updated_at": "2026-06-21T08:20:00.000000Z",
        "can_adjust": false
      }
    ],
    "data_notes": [
      {
        "key": "threshold_partial",
        "message": "Backend inventory saat ini belum memiliki threshold final per produk/outlet, sehingga status low stock dapat berlabel estimasi.",
        "severity": "info"
      }
    ]
  },
  "meta": {
    "request_id": "req_01J0INVENTORY001",
    "generated_at": "2026-06-21T08:30:00.000000Z"
  }
}
```

## Inventory Field Definition

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `product_id` | UUID string | no | `c313...` | Tenant-scoped product id. |
| `sku` | string | yes | `SKU-KOPI-001` | Dari barcode/SKU backend; safe display. |
| `name` | string | no | `Kopi Susu` | Product display name. |
| `category.id` | UUID string | yes | `0a1c...` | Nullable jika produk tanpa kategori. |
| `category.name` | string | yes | `Minuman` | Safe display category. |
| `outlet.id` | UUID string | no | `1351...` | Harus outlet milik business user. |
| `outlet.name` | string | no | `Cabang Utama` | Jangan bocorkan outlet tenant lain. |
| `stock_on_hand` | integer | no | `8` | Dari backend stock movements sum. |
| `available_stock` | integer | no | `8` | Saat ini sama dengan on hand jika reserved belum tersedia. |
| `reserved_stock` | integer | no | `0` | Safe default 0 jika backend belum punya reservation. |
| `in_transit_out` | integer | no | `0` | Dari transfer in transit source side. |
| `in_transit_in` | integer | no | `4` | Dari transfer in transit destination side. |
| `unit` | string | yes | `pcs` | Nullable jika backend belum punya unit. |
| `low_stock_threshold` | integer | yes | `10` | Nullable bila threshold belum tersedia. |
| `track_stock` | boolean | no | `true` | Produk yang tidak tracked harus status `not_tracked`. |
| `status` | string enum | no | `low` | `in_stock`, `low`, `out`, `negative`, `not_tracked`. |
| `status_label` | string | no | `Stok hampir habis` | Label Indonesia dari BFF. |
| `updated_at` | ISO-8601 UTC string | yes | `2026-06-21T08:20:00.000000Z` | Nullable jika backend tidak punya timestamp. |
| `can_adjust` | boolean | no | `false` | Harus `false` sampai stock adjustment contract aman tersedia. |

## Movements Success Response JSON

```json
{
  "data": {
    "meta": {
      "timezone": "Asia/Jakarta",
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "data_status": "partial",
      "generated_at": "2026-06-21T08:30:00.000000Z"
    },
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 1,
      "total_pages": 1,
      "has_next_page": false,
      "has_previous_page": false
    },
    "rows": [
      {
        "movement_id": "e1916f9e-83f5-40c5-9110-f516ddaca001",
        "occurred_at": "2026-06-21T07:12:00.000000Z",
        "time_label": "14:12",
        "product": { "id": "c313ba03-7c3a-49b7-bcf1-37ad94a2a001", "name": "Kopi Susu" },
        "outlet": { "id": "135159a6-d523-470f-93e1-5a3f3706a001", "name": "Cabang Utama" },
        "type": { "value": "sale", "label": "Penjualan" },
        "quantity_delta": -2,
        "before_quantity": 10,
        "after_quantity": 8,
        "reference": { "type": "transaction", "id": "9d56a07f-6d26-4582-8fda-7f630a9da001", "label": "Transaksi" },
        "actor": { "id": "6774cdcd-1cb5-4e2e-9886-62f98d19a001", "name": "Ayu" },
        "reason": null,
        "can_open_reference": true
      }
    ],
    "data_notes": []
  },
  "meta": { "request_id": "req_01J0MOVEMENTS001" }
}
```

## Movement Field Definition

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `movement_id` | UUID string | no | `e191...` | Tenant-scoped movement id. |
| `occurred_at` | ISO-8601 UTC string | no | `2026-06-21T07:12:00.000000Z` | Dari `stock_movements.created_at`. |
| `time_label` | string | no | `14:12` | Display label timezone outlet/business. |
| `product.id` | UUID string | no | `c313...` | Tenant product only. |
| `product.name` | string | no | `Kopi Susu` | Safe display. |
| `outlet.id` | UUID string | no | `1351...` | Tenant outlet only. |
| `outlet.name` | string | no | `Cabang Utama` | Safe display. |
| `type.value` | string | no | `sale` | Backend movement type normalized by BFF. |
| `type.label` | string | no | `Penjualan` | Label Indonesia. |
| `quantity_delta` | integer | no | `-2` | Server value. Client tidak menghitung stok akhir. |
| `before_quantity` | integer | yes | `10` | Nullable jika backend tidak punya snapshot. |
| `after_quantity` | integer | yes | `8` | Nullable jika backend tidak punya snapshot. |
| `reference.type` | string | yes | `transaction` | `transaction`, `inventory_purchase`, `inventory_count`, `inventory_waste`, `inventory_transfer`, atau null. |
| `reference.id` | UUID string | yes | `9d56...` | Tenant-scoped reference id. |
| `reference.label` | string | yes | `Transaksi` | Safe label. |
| `actor.id` | UUID string | yes | `6774...` | Nullable. Tidak expose kontak atau credential actor. |
| `actor.name` | string | yes | `Ayu` | Safe display actor. |
| `reason` | string | yes | `Stok opname` | Reason boleh ditampilkan jika tidak mengandung PII; BFF boleh mask bila perlu. |
| `can_open_reference` | boolean | no | `true` | Read-only detail only. Tidak boleh membuka write action. |

## Transfer In-transit Read-only Shape

Jika WEB-03B/follow-up membutuhkan panel transfer, BFF boleh menyediakan `GET /api/admin/inventory/transfers/in-transit` dengan shape stabil:

```json
{
  "data": {
    "rows": [
      {
        "transfer_id": "80e2b595-6451-4d98-ac3b-a9d65df6a001",
        "number": "TRF-20260621-ABCD",
        "source_outlet": { "id": "outlet-a", "name": "Cabang A" },
        "destination_outlet": { "id": "outlet-b", "name": "Cabang B" },
        "status": "in_transit",
        "sent_at": "2026-06-21T06:30:00.000000Z",
        "received_at": null,
        "age_seconds": 7200,
        "line_count": 2,
        "in_transit_quantity": 12,
        "can_receive": false,
        "can_cancel": false
      }
    ],
    "data_notes": []
  },
  "meta": {}
}
```

`can_receive` dan `can_cancel` wajib `false` di Web Admin read-only sampai safe write framework dan contract transfer action tersedia.

## Field Yang Tidak Boleh Keluar ke Browser

BFF response tidak boleh memuat raw token/session secret, product cost, supplier price, purchase cost, supplier contact, raw audit payload, staff credential fields, device id, shift id, idempotency key, checkout/quote token, full payment reference, atau internal deleted metadata yang tidak perlu UI.

## Empty State

Saat tidak ada item atau movement sesuai filter, response tetap `200` dengan `rows: []`, totals nol, pagination stabil, dan `data_notes` aman. UI harus menampilkan empty state yang jelas dan tidak mencampur data contoh dengan data real tanpa label.

## Error State

| HTTP | Code | Condition | Notes |
| --- | --- | --- | --- |
| `401` | `UNAUTHENTICATED` | Session missing/expired atau backend token invalid. | BFF harus clear cookie jika session invalid. |
| `403` | `FORBIDDEN` | Role bukan owner/admin, outlet/category di luar scope, atau backend menolak. | Cashier dan superadmin forbidden untuk tenant inventory. |
| `422` | `VALIDATION_ERROR` | Query filter invalid. | Detail field aman. |
| `500` | `BACKEND_UNAVAILABLE` | Backend tidak bisa dihubungi atau response tidak sesuai kontrak. | Jangan expose raw backend exception. |

## Data Source Backend

Kandidat sumber data backend yang sudah ada:

- `products` untuk product id, name, barcode/SKU, category, track_stock, updated_at.
- `product_categories` untuk kategori produk.
- `outlets` untuk outlet scope dan nama outlet.
- `stock_movements` untuk on hand quantity, movement history, before/after quantity, reference id, actor id.
- `inventory_transfers` dan `inventory_transfer_lines` untuk in-transit quantity dan transfer rows.
- `users` untuk safe actor display name pada movement.

Backend saat ini sudah membaca sumber tersebut di `InventoryService`, tetapi pagination/list totals/filter status masih perlu distabilkan untuk Web Admin production volume.

## Tenant Isolation Rules

- Semua query Laravel wajib scoped ke `business_id` dari token/session user.
- BFF tidak boleh menerima `business_id` dari browser.
- `outlet_id` harus divalidasi milik business user; jika tidak, return `403` atau `422` mengikuti pattern backend/BFF.
- `category_id` dan `product_id` harus milik business user bila dipakai.
- Jangan bocorkan produk, stok, movement, transfer, actor, atau outlet tenant lain.
- Jangan bypass global tenant scope tanpa explicit query `business_id` guard.

## Backend Gap / Data Notes

- Backend `GET /api/v1/inventory` sudah tenant-scoped, tetapi response belum punya pagination total stabil dan masih menggabungkan movements ringkas.
- Backend inventory list belum punya filter final `category_id` dan `stock_status`; BFF boleh filter/mapping sementara untuk WEB-03B.
- Backend low-stock threshold masih nullable sehingga status `low` dapat menjadi estimasi. BFF wajib menambahkan `data_notes` jika threshold belum final.
- Backend movement endpoint sudah limit `per_page`, tetapi belum expose `total/total_pages`; BFF wajib menstabilkan pagination response.

## Acceptance Criteria untuk WEB-03B

- BFF route `GET /api/admin/inventory` dibuat dan read-only.
- BFF route `GET /api/admin/inventory/movements` dibuat jika halaman membutuhkan movement table.
- Browser/UI tidak memanggil Laravel langsung.
- Token backend hanya dipakai server-side, tidak masuk Client Component, props, fixtures, logs, atau error response.
- Owner/admin dapat membaca inventory melalui session BFF.
- Cashier dan superadmin mendapat `403` untuk tenant inventory.
- Query whitelist menolak parameter liar dengan `422`.
- BFF response mengikuti shape kontrak: `meta`, `pagination`, `totals`, `rows`, `data_notes`.
- BFF men-sanitize cost/supplier/actor/internal refs yang tidak aman.
- Empty state stabil tanpa data contoh yang tercampur dengan data real.
- Stock adjustment, purchase, count, waste, transfer create/send/receive/cancel, export, dan action sensitif lain tetap disabled/preview-only.
- Tests minimal mencakup: 401 tanpa session, 403 role salah, query liar 422, no token in response, safe mapping rows, empty state, tenant isolation/cross-tenant data tidak tampil.
- Web validation lengkap pass: typecheck, lint, test, build, Playwright list, boundary scan.
