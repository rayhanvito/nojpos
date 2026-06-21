# Contract - Tenant Dashboard Summary

Status: Contract ready, implementation pending.  
Owner: Backend Core + Web Admin  
Related stories: BE-04, WEB-01, INT-02  
Task: BE-04A — Finalize Dashboard Summary API Contract

## Goal

Menyediakan satu endpoint agregat read-only untuk dashboard Owner/Admin Toko agar Web Admin `/dashboard` dapat membaca data real dari backend tanpa memanggil banyak endpoint dan tanpa menghitung laporan final di client.

Endpoint ini hanya kontrak. Implementasi backend dilakukan di task berikutnya: `BE-04B — Implement dashboard summary endpoint`.

## Endpoint Final

- Method: `GET`
- Path: `/api/v1/dashboard/summary`
- Status: Contract ready, implementation pending.
- Response envelope sukses: `{ "data": ..., "meta": ... }`
- Response error: `{ "error": { "code": "...", "message": "...", "details": ... } }`
- Semua nilai uang adalah integer rupiah. Client hanya memformat tampilan.
- Endpoint bersifat read-only dan tidak boleh mengubah data apa pun.

## Auth, Role, dan Scope

- Harus login dengan session/token valid.
- Hanya tenant `owner` dan `admin` yang boleh akses.
- `manager` belum masuk kontrak final kecuali nanti diputuskan di task terpisah.
- `cashier` tidak boleh akses endpoint ini kecuali nanti diputuskan khusus.
- `superadmin` tidak memakai endpoint ini untuk platform dashboard; Superadmin memakai endpoint platform seperti `/api/v1/superadmin/summary`.
- Semua data wajib scoped ke `business_id` dari user yang sedang login.
- Client tidak boleh mengirim atau memilih `business_id`.

## Query Parameters

Jangan dibuat kompleks dulu. Filter final untuk BE-04B:

| Field | Type | Required | Default | Validation | Notes |
| --- | --- | --- | --- | --- | --- |
| `date` | string | no | business-day today | `YYYY-MM-DD` | Anchor hari dashboard dan akhir chart 7 hari. Business-day dihitung dengan timezone outlet. |
| `outlet_id` | UUID string | no | semua outlet milik business | UUID valid dan harus milik `business_id` user | Jika diberikan, semua agregat hanya untuk outlet itu. Jangan pakai `all`; kosong berarti semua outlet yang diizinkan. |
| `range` | string enum | no | `last_7_days` | `last_7_days` saja untuk kontrak awal | Dipakai untuk chart 7 hari. Nilai lain ditolak `422` sampai kontrak berikutnya. |

Catatan:

- Jika `outlet_id` tidak dikirim, dashboard menggabungkan semua outlet dalam business user.
- Jika `outlet_id` dikirim tetapi bukan milik business user, return `403 FORBIDDEN`.
- Jika format filter salah, return `422 VALIDATION_ERROR`.
- Date range custom, export, pagination, dan drill-down tidak termasuk kontrak ini.

## Success Response JSON Final

```json
{
  "data": {
    "meta": {
      "date": "2026-06-21",
      "range": "last_7_days",
      "timezone": "Asia/Jakarta",
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "outlet_name": null,
      "generated_at": "2026-06-21T08:30:00.000000Z",
      "currency": "IDR",
      "money_format": "integer_rupiah",
      "data_status": "real",
      "is_empty_today": false
    },
    "kpis": {
      "sales_today": {
        "value": 2450000,
        "type": "money",
        "label": "Penjualan hari ini",
        "trend_label": "+8% vs kemarin",
        "is_estimate": false
      },
      "transaction_count": {
        "value": 86,
        "type": "integer",
        "label": "Jumlah transaksi",
        "trend_label": "Ramai stabil",
        "is_estimate": false
      },
      "average_transaction": {
        "value": 28500,
        "type": "money",
        "label": "Rata-rata transaksi",
        "trend_label": null,
        "is_estimate": false
      },
      "gross_profit_estimate": {
        "value": 740000,
        "type": "money",
        "label": "Gross profit estimasi",
        "trend_label": null,
        "is_estimate": true
      },
      "low_stock_count": {
        "value": 12,
        "type": "integer",
        "label": "Stok hampir habis",
        "trend_label": "Butuh cek",
        "is_estimate": false
      },
      "cash_difference": {
        "value": 45000,
        "type": "money",
        "label": "Selisih kas",
        "trend_label": "Perlu konfirmasi",
        "is_estimate": true
      }
    },
    "alerts": [
      {
        "id": "low-stock-2026-06-21",
        "type": "low_stock",
        "severity": "warning",
        "title": "12 produk stok menipis",
        "message": "Cek stok sebelum jam ramai agar kasir tidak menjual item yang kosong.",
        "outlet_id": null,
        "outlet_name": null,
        "entity_type": "inventory",
        "entity_id": null,
        "action_label": "Cek inventori",
        "action_path": "/inventory"
      }
    ],
    "sales_last_7_days": [
      {
        "date": "2026-06-15",
        "label": "Sen",
        "sales": 2100000,
        "transaction_count": 74,
        "average_transaction": 28378
      }
    ],
    "payment_methods": [
      {
        "method": "qris",
        "label": "QRIS",
        "amount": 1320000,
        "transaction_count": 41,
        "share_percent": 53.88
      }
    ],
    "top_products": [
      {
        "product_id": "3eb675c1-e31a-4d3f-9d1b-b04c0567a001",
        "name": "Es Kopi Susu Gula Aren",
        "qty_sold": 34,
        "sales": 612000,
        "share_percent": 86.0,
        "outlet_id": null,
        "outlet_name": null
      }
    ],
    "low_stock_items": [
      {
        "product_id": "5369b506-7386-4f4f-9fdf-583d5ad0a001",
        "name": "Cup 16 oz",
        "sku": null,
        "outlet_id": "135159a6-d523-470f-93e1-5a3f3706a001",
        "outlet_name": "Cabang Utama",
        "remaining_stock": 22,
        "threshold": 30,
        "unit": "pcs",
        "status": "low"
      }
    ],
    "recent_transactions": [
      {
        "transaction_id": "9d56a07f-6d26-4582-8fda-7f630a9da001",
        "code": "TRX-1028",
        "time": "15:42",
        "occurred_at": "2026-06-21T08:42:00.000000Z",
        "outlet_id": "135159a6-d523-470f-93e1-5a3f3706a001",
        "outlet_name": "Cabang Utama",
        "cashier_name": "Ayu",
        "payment_method": "QRIS",
        "total": 38000,
        "status": "paid"
      }
    ],
    "cashier_performance": [
      {
        "cashier_id": "6774cdcd-1cb5-4e2e-9886-62f98d19a001",
        "name": "Ayu",
        "transaction_count": 31,
        "sales": 920000,
        "average_transaction": 29677,
        "void_count": 0,
        "refund_count": 0,
        "cash_difference": 0,
        "note": "Shift pagi rapi"
      }
    ],
    "branch_highlights": [
      {
        "outlet_id": "135159a6-d523-470f-93e1-5a3f3706a001",
        "name": "Cabang Utama",
        "status": "normal",
        "summary": "Penjualan stabil, stok cup perlu dicek.",
        "sales_today": 1750000,
        "transaction_count": 58,
        "low_stock_count": 4,
        "open_shift_count": 1,
        "cash_difference": 0,
        "severity": "success"
      }
    ],
    "data_notes": [
      {
        "key": "gross_profit_estimate",
        "message": "Gross profit masih estimasi jika data cost produk belum lengkap.",
        "severity": "info"
      },
      {
        "key": "cash_difference",
        "message": "Selisih kas berasal dari shift yang sudah ditutup atau estimasi shift berjalan bila tersedia.",
        "severity": "warning"
      }
    ]
  },
  "meta": {
    "request_id": "req_01J0DASHBOARD001",
    "generated_at": "2026-06-21T08:30:00.000000Z"
  }
}
```

## Field Definition Detail

### `data.meta`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `date` | string `YYYY-MM-DD` | no | `2026-06-21` | Anchor hari dashboard dalam timezone outlet/business. |
| `range` | string | no | `last_7_days` | Kontrak awal hanya 7 hari terakhir. |
| `timezone` | string | no | `Asia/Jakarta` | Dari outlet bila `outlet_id` ada, default business/outlet fallback `Asia/Jakarta`. |
| `business_id` | UUID string | no | `8f7f...` | Diambil dari user login; hanya business sendiri. |
| `outlet_id` | UUID string | yes | `1351...` | `null` berarti semua outlet dalam business. |
| `outlet_name` | string | yes | `Cabang Utama` | `null` bila semua outlet. |
| `generated_at` | ISO-8601 UTC string | no | `2026-06-21T08:30:00.000000Z` | Waktu server membuat response. |
| `currency` | string | no | `IDR` | Tetap `IDR`. |
| `money_format` | string | no | `integer_rupiah` | Client tidak menghitung uang final. |
| `data_status` | string enum | no | `real` | `real`, `partial`, atau `unavailable`. Jangan return data palsu produksi. |
| `is_empty_today` | boolean | no | `false` | `true` jika belum ada transaksi pada `date`. |

### `kpis`

Semua KPI memakai object stabil dengan field:

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `value` | integer | no | `2450000` | Money dalam rupiah integer atau count integer. |
| `type` | string enum | no | `money` | `money`, `integer`, atau `percent`. |
| `label` | string | no | `Penjualan hari ini` | Label UI Indonesia. |
| `trend_label` | string | yes | `+8% vs kemarin` | Server boleh `null` jika belum dihitung. |
| `is_estimate` | boolean | no | `true` | Wajib `true` untuk gross profit estimasi dan cash difference jika belum final. |

KPI keys wajib ada walau nilainya `0`:

| Key | Source Candidate | Privacy/Scope Notes |
| --- | --- | --- |
| `sales_today` | `transactions.grand_total` status paid/partially refunded/refunded dikurangi refund bila tersedia | Scoped business/outlet/date. |
| `transaction_count` | `transactions` paid pada tanggal bisnis | Scoped business/outlet/date. |
| `average_transaction` | `sales_today / transaction_count`, dihitung server | Return `0` jika count `0`. |
| `gross_profit_estimate` | kandidat: `transaction_items`, product cost jika nanti tersedia | Wajib `is_estimate: true` sampai COGS/cost final. |
| `low_stock_count` | kandidat: `products.track_stock`, `stock_movements`, threshold setting | Tidak expose data toko lain. |
| `cash_difference` | `shift_sessions.cash_difference`, `cash_movements`, payment cash totals | Wajib jelas estimasi/final dari backend. |

### `alerts[]`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `id` | string | no | `low-stock-2026-06-21` | Stabil untuk React key; tidak harus UUID. |
| `type` | string enum | no | `low_stock` | `low_stock`, `out_of_stock`, `unclosed_shift`, `cash_difference`, `subscription_expiring`, `system_note`. |
| `severity` | string enum | no | `warning` | `info`, `success`, `warning`, `danger`. |
| `title` | string | no | `12 produk stok menipis` | Tidak mengandung PII/payment reference. |
| `message` | string | no | `Cek stok sebelum jam ramai...` | Copy aman untuk owner/admin. |
| `outlet_id` | UUID string | yes | `1351...` | `null` untuk alert business-wide. |
| `outlet_name` | string | yes | `Cabang Utama` | Hanya outlet dalam business. |
| `entity_type` | string | yes | `inventory` | `inventory`, `shift`, `cash`, `subscription`, atau `transaction`. |
| `entity_id` | UUID string | yes | `5369...` | Jangan expose resource luar tenant. |
| `action_label` | string | yes | `Cek inventori` | UI boleh menampilkan link read-only. |
| `action_path` | string | yes | `/inventory` | Path web internal; jangan trigger write. |

### `sales_last_7_days[]`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `date` | string `YYYY-MM-DD` | no | `2026-06-15` | Local business date. |
| `label` | string | no | `Sen` | Label pendek untuk chart. |
| `sales` | integer money | no | `2100000` | Server-calculated. |
| `transaction_count` | integer | no | `74` | Server-calculated. |
| `average_transaction` | integer money | no | `28378` | Server-calculated; `0` jika tidak ada transaksi. |

Array wajib 7 item untuk `range=last_7_days`, termasuk hari tanpa transaksi dengan nilai `0`.

### `payment_methods[]`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `method` | string | no | `qris` | Key normalized dari `payments.method`. |
| `label` | string | no | `QRIS` | Dari config jika ada, fallback uppercase. |
| `amount` | integer money | no | `1320000` | Sum payment confirmed/settled. |
| `transaction_count` | integer | no | `41` | Count distinct transaction. |
| `share_percent` | number | no | `53.88` | Server-calculated untuk chart; `0` jika total `0`. |

Jangan return `payments.reference` atau detail payment provider.

### `top_products[]`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `product_id` | UUID string | yes | `3eb6...` | Bisa `null` jika item historical tidak lagi punya product record. |
| `name` | string | no | `Es Kopi Susu Gula Aren` | Dari snapshot `transaction_items.name`. |
| `qty_sold` | integer | no | `34` | Sum quantity. |
| `sales` | integer money | no | `612000` | Net/gross sesuai implementasi BE-04B, harus didokumentasikan di resource. |
| `share_percent` | number | no | `86.0` | Untuk progress bar relatif top item, server-calculated. |
| `outlet_id` | UUID string | yes | `1351...` | `null` jika semua outlet/agregat. |
| `outlet_name` | string | yes | `Cabang Utama` | Hanya outlet scoped. |

Limit awal: maksimal 5 item.

### `low_stock_items[]`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `product_id` | UUID string | no | `5369...` | Product dalam business. |
| `name` | string | no | `Cup 16 oz` | Tidak PII. |
| `sku` | string | yes | `CUP-16` | `null` jika belum ada SKU. |
| `outlet_id` | UUID string | no | `1351...` | Wajib karena stok per outlet. |
| `outlet_name` | string | no | `Cabang Utama` | Hanya outlet scoped. |
| `remaining_stock` | integer | no | `22` | Berdasarkan stock server. |
| `threshold` | integer | yes | `30` | `null` jika threshold belum ada; status bisa dari default threshold. |
| `unit` | string | yes | `pcs` | `null` jika unit belum ada. |
| `status` | string enum | no | `low` | `low` atau `out`. |

Limit awal: maksimal 5 item paling kritis.

### `recent_transactions[]`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `transaction_id` | UUID string | no | `9d56...` | Hanya transaksi dalam scope. |
| `code` | string | no | `TRX-1028` | Dari `transactions.number`. |
| `time` | string `HH:mm` | no | `15:42` | Waktu local timezone. |
| `occurred_at` | ISO-8601 UTC string | no | `2026-06-21T08:42:00.000000Z` | Raw server timestamp. |
| `outlet_id` | UUID string | no | `1351...` | Scope outlet. |
| `outlet_name` | string | no | `Cabang Utama` | Scope outlet. |
| `cashier_name` | string | yes | `Ayu` | Staff name dalam business; jangan expose PIN/email. |
| `payment_method` | string | yes | `QRIS` | Label metode utama; jangan expose reference. |
| `total` | integer money | no | `38000` | Grand total server. |
| `status` | string | no | `paid` | Status transaksi. |

Limit awal: maksimal 5 transaksi terbaru.

### `cashier_performance[]`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `cashier_id` | UUID string | no | `6774...` | User/staff dalam business. |
| `name` | string | no | `Ayu` | Jangan expose email/phone/PIN. |
| `transaction_count` | integer | no | `31` | Count transaksi paid. |
| `sales` | integer money | no | `920000` | Sum transaksi paid. |
| `average_transaction` | integer money | no | `29677` | Server-calculated. |
| `void_count` | integer | no | `0` | Count void jika tersedia. |
| `refund_count` | integer | no | `0` | Count refund jika tersedia. |
| `cash_difference` | integer money | yes | `0` | `null` jika shift belum ditutup atau belum tersedia. |
| `note` | string | yes | `Shift pagi rapi` | Optional summary aman dari backend. |

Limit awal: maksimal 5 kasir aktif/teratas pada tanggal tersebut.

### `branch_highlights[]`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `outlet_id` | UUID string | no | `1351...` | Outlet dalam business. |
| `name` | string | no | `Cabang Utama` | Nama outlet. |
| `status` | string enum | no | `normal` | `normal`, `needs_stock`, `unclosed_shift`, `cash_issue`, `inactive`. |
| `summary` | string | no | `Penjualan stabil...` | Copy pendek aman. |
| `sales_today` | integer money | no | `1750000` | Scoped outlet. |
| `transaction_count` | integer | no | `58` | Scoped outlet. |
| `low_stock_count` | integer | no | `4` | Scoped outlet. |
| `open_shift_count` | integer | no | `1` | Scoped outlet. |
| `cash_difference` | integer money | yes | `0` | `null` jika belum tersedia. |
| `severity` | string enum | no | `success` | `success`, `info`, `warning`, `danger`. |

Return `[]` jika business hanya punya satu outlet dan tidak ada highlight penting; UI boleh tetap render empty state.

### `data_notes[]`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `key` | string | no | `gross_profit_estimate` | Mengacu field yang perlu disclaimer. |
| `message` | string | no | `Gross profit masih estimasi...` | Tidak mengandung data sensitif. |
| `severity` | string enum | no | `info` | `info`, `warning`, `danger`. |

Gunakan `data_notes`, bukan `preview_notes`, karena endpoint ini untuk data real. `data_status` dapat menjadi `partial` jika beberapa sumber belum tersedia.

## Empty State

Jika belum ada transaksi pada `date`, endpoint tetap return `200` dengan shape lengkap:

```json
{
  "data": {
    "meta": {
      "date": "2026-06-21",
      "range": "last_7_days",
      "timezone": "Asia/Jakarta",
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "outlet_name": null,
      "generated_at": "2026-06-21T01:00:00.000000Z",
      "currency": "IDR",
      "money_format": "integer_rupiah",
      "data_status": "real",
      "is_empty_today": true
    },
    "kpis": {
      "sales_today": { "value": 0, "type": "money", "label": "Penjualan hari ini", "trend_label": null, "is_estimate": false },
      "transaction_count": { "value": 0, "type": "integer", "label": "Jumlah transaksi", "trend_label": null, "is_estimate": false },
      "average_transaction": { "value": 0, "type": "money", "label": "Rata-rata transaksi", "trend_label": null, "is_estimate": false },
      "gross_profit_estimate": { "value": 0, "type": "money", "label": "Gross profit estimasi", "trend_label": null, "is_estimate": true },
      "low_stock_count": { "value": 0, "type": "integer", "label": "Stok hampir habis", "trend_label": null, "is_estimate": false },
      "cash_difference": { "value": 0, "type": "money", "label": "Selisih kas", "trend_label": null, "is_estimate": true }
    },
    "alerts": [],
    "sales_last_7_days": [
      { "date": "2026-06-15", "label": "Sen", "sales": 0, "transaction_count": 0, "average_transaction": 0 }
    ],
    "payment_methods": [],
    "top_products": [],
    "low_stock_items": [],
    "recent_transactions": [],
    "cashier_performance": [],
    "branch_highlights": [],
    "data_notes": []
  },
  "meta": {
    "request_id": "req_01J0DASHBOARDEMPTY",
    "generated_at": "2026-06-21T01:00:00.000000Z"
  }
}
```

Rules empty state:

- Jangan return data palsu sebagai data produksi.
- Semua KPI key wajib tetap ada.
- `sales_last_7_days` tetap 7 item agar chart stabil.
- List operasional boleh `[]`.
- Alert setup boleh ditambahkan hanya jika berasal dari rule backend nyata dan aman.

## Error States

### `401 UNAUTHENTICATED`

```json
{
  "error": {
    "code": "UNAUTHENTICATED",
    "message": "Sesi tidak valid atau sudah berakhir.",
    "details": {}
  }
}
```

Use when session/token missing, invalid, or expired.

### `403 FORBIDDEN`

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Anda tidak memiliki akses ke dashboard toko ini.",
    "details": {}
  }
}
```

Use when role is cashier/superadmin/unsupported, or `outlet_id` is outside the authenticated business.

### `422 VALIDATION_ERROR`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Filter dashboard tidak valid.",
    "details": {
      "date": ["The date field must match the format Y-m-d."],
      "range": ["Range dashboard hanya mendukung last_7_days untuk kontrak awal."]
    }
  }
}
```

Use for invalid `date`, invalid UUID format, unsupported `range`, or malformed query.

### `500 SERVER_ERROR`

```json
{
  "error": {
    "code": "SERVER_ERROR",
    "message": "Ringkasan dashboard belum bisa dimuat. Coba lagi nanti.",
    "details": {}
  }
}
```

Use for unexpected server failure. Logs must redact token, PIN, customer PII, receipt content, and payment references.

## Data Source Backend Candidates

BE-04B boleh memakai sumber data yang sudah ada, tanpa membuat write flow baru:

| Response Area | Candidate Source |
| --- | --- |
| `kpis.sales_today`, `transaction_count`, `average_transaction` | `transactions` filtered by `business_id`, optional `outlet_id`, business-day UTC window, paid/partially refunded/refunded status. |
| Refund/net adjustment | `refunds` joined to `transactions` if needed. |
| `gross_profit_estimate` | `transaction_items`, `products`, future cost/COGS source if available. If cost missing, return estimate note. |
| `low_stock_count`, `low_stock_items` | `products.track_stock`, `stock_movements`, inventory service aggregation, default/setting threshold if no threshold table exists. |
| `cash_difference` | `shift_sessions.cash_difference`, `expected_cash`, `actual_cash`, `cash_movements`, cash payment totals. |
| `alerts` | Derived from low/out stock, open/unclosed shifts, cash difference, subscription expiring, and unavailable data notes. |
| `sales_last_7_days` | `transactions` grouped by local business date over seven-day window. |
| `payment_methods` | `payments` joined to `transactions`, optionally `payment_method_configs` for label/is_cash. |
| `top_products` | `transaction_items` joined to `transactions`; product metadata optional. |
| `recent_transactions` | `transactions` joined to `users`, `outlets`, and confirmed/settled `payments` summary. |
| `cashier_performance` | `transactions` grouped by `cashier_id`, joined to `users`; void/refund counts from `transactions`/`refunds`. |
| `branch_highlights` | `outlets`, transaction aggregates, shift sessions, low stock counts per outlet. |
| Subscription alert | `subscriptions` if tenant subscription data is available in backend. |

Do not implement in BE-04A. This section only guides BE-04B implementation.

## Existing Backend Route/Report Context

Routes already present in `apps/backend/routes/api.php` that can inform BE-04B:

- `GET /api/v1/reports/sales-summary`
- `GET /api/v1/reports/sold-products`
- `GET /api/v1/reports/payment-methods`
- `GET /api/v1/reports/cashier-shifts`
- `GET /api/v1/reports/void-refund-audit`
- `GET /api/v1/reports/top-10`
- `GET /api/v1/transactions`
- `GET /api/v1/inventory`
- `GET /api/v1/shifts/current`
- `GET /api/v1/subscription`

BE-04B should prefer shared backend service/query logic over duplicating inconsistent calculations, while still returning this contract shape in one dashboard endpoint.

## Tenant Isolation Rules

- Every query must include `business_id` from authenticated user context.
- Never trust `business_id` from query/body/header.
- `outlet_id` must be validated as belonging to the authenticated business before use.
- Superadmin must not use this endpoint to access tenant dashboards.
- Cashier must be forbidden unless a future contract explicitly allows a limited cashier dashboard.
- Do not expose customer PII, staff PIN/password, bearer tokens, payment references, receipt content, or data from another business.
- If data source uses query builder, add explicit `business_id` conditions; do not rely only on global scopes.
- For all-outlet summary, aggregate only outlets in the authenticated business and any future outlet permission policy.

## Security and Non-Goals

Endpoint must not:

- create, update, delete, export, or enqueue data;
- trigger payment, refund, void, shift open/close, cash movement, stock adjustment, payment retry, or report export;
- expose payment provider reference values;
- create fake production data;
- bypass role middleware/policy.

## Acceptance Criteria for BE-04B Implementation

- Adds `GET /api/v1/dashboard/summary` under `/api/v1` with auth and owner/admin authorization.
- Endpoint is read-only and does not mutate database state.
- Does not trigger payment/refund/void/shift close/open/cash movement/stock adjustment/export.
- Response follows this contract and standard envelope.
- All money values are integer rupiah.
- All dashboard calculations are server-side.
- Tenant isolation is enforced for every source query.
- `outlet_id` outside business returns `403 FORBIDDEN`.
- Cashier and superadmin roles receive `403 FORBIDDEN`.
- Empty state returns stable shape with zero KPI values and arrays.
- Tests cover: unauthenticated, forbidden role, outlet tenant isolation, invalid filter `422`, empty state, and normal state.
- Tests include at least one check that data from tenant A does not appear for tenant B.

## Web Mapping Notes

Current Web Admin `/dashboard` preview expects these sections:

- KPI cards: `kpis`
- Alert penting: `alerts`
- Sales chart: `sales_last_7_days`
- Payment chart: `payment_methods`
- Produk terlaris: `top_products`
- Stok kritis: `low_stock_items`
- Transaksi terbaru: `recent_transactions`
- Performa kasir: `cashier_performance`
- Cabang yang perlu dicek: `branch_highlights`
- Preview/disclaimer copy: `data_notes` and `data.meta.data_status`

Web integration remains blocked until session strategy and BE-04B implementation are done.
