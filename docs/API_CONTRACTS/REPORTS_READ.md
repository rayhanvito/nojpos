# Contract - Tenant Reports Read-only

Status: Contract ready, implementation pending.  
Owner: Web Admin + Backend Core  
Related stories: WEB-05, INT-04  
Task: WEB-05A — Reports read-only contract

## Goal

Menyediakan kontrak read-only untuk halaman Tenant Admin `/reports` agar owner/admin toko dapat membaca ringkasan penjualan, produk terlaris, metode pembayaran, dan shift kasir melalui Next.js BFF. Kontrak ini tidak mengaktifkan export, refund, void, reprint, cash reconciliation, shift close/open, payment retry, atau write action apa pun.

## Endpoint Backend Nyata

Endpoint berikut sudah ada di Laravel dan protected oleh `auth:sanctum`:

| Method | Path | Status | Notes |
| --- | --- | --- | --- |
| `GET` | `/api/v1/reports/sales-summary` | exists | Ringkasan penjualan dan chart. |
| `GET` | `/api/v1/reports/sold-products` | exists | Produk terjual berdasarkan transaction items. |
| `GET` | `/api/v1/reports/payment-methods` | exists | Agregasi metode pembayaran. |
| `GET` | `/api/v1/reports/cashier-shifts` | exists | Ringkasan shift kasir dan kas. |
| `GET` | `/api/v1/reports/top-10` | exists | Top ten report; BFF boleh memakai endpoint ini untuk widget top products jika lebih sesuai. |
| `GET` | `/api/v1/reports/void-refund-audit` | exists | Read-only audit; tidak dipakai untuk WEB-05B awal agar tidak membuka aksi sensitif. |

Backend `ReportRequest` saat ini menerima `date`, `date_from`, `date_to`, `range`, `shift_id`, `outlet_id`, dan `export`. Untuk Web Admin BFF awal, query dipersempit agar stabil dan aman.

## Endpoint BFF Target

Browser/UI hanya boleh memakai endpoint same-origin Next.js atau server-side helper yang sama:

| Method | Path | Backend target | Purpose | WEB task |
| --- | --- | --- | --- | --- |
| `GET` | `/api/admin/reports/sales-summary` | `/api/v1/reports/sales-summary` | KPI dan chart sales untuk `/reports`. | WEB-05B |
| `GET` | `/api/admin/reports/top-products` | `/api/v1/reports/sold-products` atau `/api/v1/reports/top-10` | Produk terlaris read-only. | WEB-05B |
| `GET` | `/api/admin/reports/payment-methods` | `/api/v1/reports/payment-methods` | Breakdown metode pembayaran. | WEB-05B |
| `GET` | `/api/admin/reports/cashier-shifts` | `/api/v1/reports/cashier-shifts` | Performa shift dan selisih kas read-only. | WEB-05B |

BFF wajib membaca sealed HttpOnly session cookie sesuai `SESSION.md`, memanggil Laravel server-side, dan tidak pernah mengirim token/session raw ke browser.

## Auth, Role, dan Scope

- Harus login dengan session Web Admin valid.
- Hanya tenant `owner` dan `admin` yang boleh akses.
- `cashier` harus mendapat `403 FORBIDDEN` dari BFF.
- `superadmin` tidak memakai endpoint tenant reports ini; platform memakai endpoint platform/superadmin sendiri.
- Semua data wajib scoped ke `business_id` dari user/session server-side.
- Client tidak boleh mengirim `business_id`, `cashier_id`, atau `shift_id` untuk WEB-05B awal.
- `outlet_id`, jika dikirim, wajib divalidasi sebagai outlet milik business user oleh backend atau BFF guard.
- Field PII/customer/payment reference/token/raw audit payload tidak boleh dikirim ke browser.

## Query Parameters

Semua BFF report endpoint memakai whitelist query yang sama untuk fase awal:

| Field | Type | Required | Default | Validation | Notes |
| --- | --- | --- | --- | --- | --- |
| `date_from` | date string | no | today in outlet/business timezone | `YYYY-MM-DD`; required with `date_to` | Awal range inklusif berdasarkan business day. |
| `date_to` | date string | no | same as `date_from` | `YYYY-MM-DD`; after/equal `date_from`; max 31 days | Akhir range inklusif. |
| `outlet_id` | UUID string | no | all outlets | UUID valid dan harus milik business user | Tenant isolation wajib. |
| `range` | enum | no | `day` | `day`, `week`, `month` | Diteruskan hanya jika dibutuhkan backend chart bucket. |

Parameter lain harus ditolak `422 VALIDATION_ERROR`. BFF tidak boleh meneruskan `export`, `cashier_id`, `shift_id`, `business_id`, atau parameter liar dari browser.

## Success Response - Sales Summary

`GET /api/admin/reports/sales-summary`

```json
{
  "data": {
    "meta": {
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "currency": "IDR",
      "date_from": "2026-06-21",
      "date_to": "2026-06-21",
      "range": "day",
      "generated_at": "2026-06-21T08:30:00.000000Z"
    },
    "summary": {
      "gross_sales": 1250000,
      "net_sales": 1180000,
      "discount_total": 70000,
      "tax_total": 0,
      "service_total": 0,
      "transaction_count": 42,
      "average_transaction": 28095,
      "refund_total": 0,
      "void_count": 0
    },
    "chart": [
      { "label": "2026-06-21", "gross_sales": 1250000, "net_sales": 1180000, "transaction_count": 42 }
    ],
    "data_notes": [
      { "key": "server_totals", "message": "Nominal rupiah berasal dari backend; UI tidak menghitung total final.", "severity": "info" }
    ]
  },
  "meta": { "request_id": "req_abc" }
}
```

## Success Response - Top Products

`GET /api/admin/reports/top-products`

```json
{
  "data": {
    "meta": {
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "currency": "IDR",
      "date_from": "2026-06-21",
      "date_to": "2026-06-21",
      "generated_at": "2026-06-21T08:30:00.000000Z"
    },
    "rows": [
      {
        "product_id": "c313ba03-7c3a-49b7-bcf1-37ad94a2a001",
        "product_name": "Kopi Susu",
        "category_name": "Minuman",
        "quantity_sold": 28,
        "gross_sales": 504000,
        "discount_total": 20000,
        "net_sales": 484000,
        "outlet": { "id": "f1111111-1111-4111-8111-111111111111", "name": "Outlet Pusat" }
      }
    ],
    "data_notes": []
  },
  "meta": { "request_id": "req_abc" }
}
```

## Success Response - Payment Methods

`GET /api/admin/reports/payment-methods`

```json
{
  "data": {
    "meta": {
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "currency": "IDR",
      "date_from": "2026-06-21",
      "date_to": "2026-06-21",
      "generated_at": "2026-06-21T08:30:00.000000Z"
    },
    "rows": [
      { "method": "cash", "label": "Tunai", "amount": 620000, "transaction_count": 18, "share_percent": 52.54 }
    ],
    "data_notes": []
  },
  "meta": { "request_id": "req_abc" }
}
```

## Success Response - Cashier Shifts

`GET /api/admin/reports/cashier-shifts`

```json
{
  "data": {
    "meta": {
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "currency": "IDR",
      "date_from": "2026-06-21",
      "date_to": "2026-06-21",
      "generated_at": "2026-06-21T08:30:00.000000Z"
    },
    "rows": [
      {
        "shift_id": "s1111111-1111-4111-8111-111111111111",
        "cashier": { "id": "u1111111-1111-4111-8111-111111111111", "name": "Rina" },
        "outlet": { "id": "f1111111-1111-4111-8111-111111111111", "name": "Outlet Pusat" },
        "opened_at": "2026-06-21T01:00:00.000000Z",
        "closed_at": null,
        "status": "open",
        "opening_cash": 300000,
        "expected_cash": 920000,
        "declared_closing_cash": 0,
        "variance": 0,
        "cash_in": 0,
        "cash_out": 0
      }
    ],
    "data_notes": [
      { "key": "read_only_shift", "message": "Data shift hanya dibaca; endpoint ini tidak menutup shift atau rekonsiliasi kas.", "severity": "info" }
    ]
  },
  "meta": { "request_id": "req_abc" }
}
```

## Field Rules

- Semua nominal uang adalah integer rupiah dari backend/BFF mapping: `gross_sales`, `net_sales`, `discount_total`, `tax_total`, `service_total`, `average_transaction`, `refund_total`, `amount`, `opening_cash`, `expected_cash`, `declared_closing_cash`, `variance`, `cash_in`, dan `cash_out`.
- UI boleh format Rupiah untuk display, tetapi tidak boleh menghitung total final sendiri.
- `share_percent` boleh angka decimal untuk visual chart; nilai harus berasal dari backend atau BFF mapping server-side, bukan Client Component.
- `cashier.name` boleh ditampilkan; email, phone, PIN, token, dan raw audit payload tidak boleh tampil.
- `shift_id` boleh dipakai untuk read-only drilldown nanti, tetapi WEB-05B tidak membuat shift close/open/cash reconciliation.

## Empty State

Jika tidak ada transaksi atau shift pada filter:

```json
{
  "data": {
    "meta": {
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "currency": "IDR",
      "date_from": "2026-06-21",
      "date_to": "2026-06-21",
      "generated_at": "2026-06-21T08:30:00.000000Z"
    },
    "summary": {
      "gross_sales": 0,
      "net_sales": 0,
      "discount_total": 0,
      "tax_total": 0,
      "service_total": 0,
      "transaction_count": 0,
      "average_transaction": 0,
      "refund_total": 0,
      "void_count": 0
    },
    "chart": [],
    "rows": [],
    "data_notes": []
  },
  "meta": { "request_id": "req_empty" }
}
```

Untuk endpoint list (`top-products`, `payment-methods`, `cashier-shifts`), `rows` kosong dan summary boleh tidak ada sesuai endpoint.

## Error State

| HTTP | Code | When | UI behavior |
| --- | --- | --- | --- |
| `401` | `UNAUTHENTICATED` | Session missing/expired/backend token invalid | Clear web session cookie; tampilkan session-required state. |
| `403` | `FORBIDDEN` | Role bukan owner/admin, outlet bukan milik tenant, superadmin/cashier mencoba tenant report | Tampilkan forbidden state; jangan tampilkan data tenant. |
| `422` | `VALIDATION_ERROR` | Query invalid, range terlalu panjang, parameter liar | Tampilkan filter error ramah. |
| `500` | `SERVER_ERROR` / `BACKEND_UNAVAILABLE` | Backend unavailable atau error tidak terduga | Tampilkan error card dan fallback preview berlabel. |

## Data Source Backend

Kandidat sumber data existing:

- `transactions` untuk sales, count, status, cashier/outlet/shift scope.
- `transaction_items` untuk sold/top products.
- `payments` untuk payment methods.
- `shift_sessions` dan `cash_movements` untuk cashier shifts, expected/actual cash, variance, cash in/out.
- `users` untuk nama kasir scoped ke business.
- `outlets` untuk nama outlet dan timezone.
- Existing `ReportController` dan `ReportService` untuk agregasi.

## Acceptance Criteria WEB-05B

- BFF routes dibuat untuk empat endpoint target.
- `/reports` mengambil data read-only melalui server-side helper/BFF.
- Tidak ada fetch Laravel langsung dari Client Component.
- Tidak ada token di props, fixtures, browser-readable storage, atau logs.
- Query whitelist hanya `date_from`, `date_to`, `outlet_id`, dan `range`.
- BFF menolak `export`, `cashier_id`, `shift_id`, `business_id`, dan parameter liar.
- Export/download tetap disabled/preview-only.
- UI menampilkan session-required, forbidden, empty, error, dan backend data state.
- Test mencakup unauthenticated, forbidden role, query whitelist, no token leak, mapping stable, dan fallback page render.
- Boundary scan clean untuk `fetch(`, `XMLHttpRequest`, `localStorage`, `sessionStorage`, `NEXT_PUBLIC_API_BASE_URL`, `Authorization`, dan `Bearer` di app/components/fixtures.

## Data Notes / Known Gaps

- Backend report endpoints sudah ada, tetapi Web Admin BFF tetap perlu menstabilkan DTO agar halaman tidak tergantung langsung pada variasi payload backend.
- Backend menerima `export`, tetapi Web Admin read-only fase ini wajib menolak export dari browser.
- `void-refund-audit` tidak diintegrasikan pada WEB-05B awal karena dekat dengan aksi sensitif; bisa direncanakan setelah sensitive-action contract siap.
