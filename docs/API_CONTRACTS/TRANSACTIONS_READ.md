# Contract - Tenant Transactions Read-only

Status: Contract ready, implementation pending.  
Owner: Backend Core + Web Admin  
Related stories: BE-03, WEB-02, INT-03  
Task: WEB-02A — Finalize Transactions read-only contract

## Goal

Menyediakan kontrak read-only untuk halaman Tenant Admin `/transactions` agar owner/admin toko dapat melihat daftar transaksi secara aman melalui Next.js BFF tanpa membuka action sensitif seperti refund, void, reprint, export, payment retry, shift close/open, atau cash reconciliation.

Kontrak ini menargetkan integrasi Web Admin berikutnya. UI tidak boleh memanggil Laravel langsung dan tidak boleh menghitung total transaksi final di browser.

## Endpoint Target

### Backend Laravel

- Method: `GET`
- Path: `/api/v1/transactions`
- Current status: route exists, read-only, protected by `auth:sanctum`.
- Current backend note: endpoint saat ini sudah tenant-scoped by authenticated `business_id`, tetapi filter/pagination backend masih terbatas. Kontrak Web Admin membutuhkan filter dan pagination stabil; WEB-02B boleh melakukan mapping/sanitization di BFF, tetapi backend-side filter/pagination tetap direkomendasikan sebelum volume produksi besar.

### Next.js BFF

- Method: `GET`
- Path: `/api/admin/transactions`
- Browser/UI hanya boleh memanggil endpoint BFF ini atau server-side helper yang sama.
- BFF membaca sealed HttpOnly session cookie sesuai `SESSION.md`.
- BFF memanggil Laravel server-side dan tidak pernah mengirim token/session raw ke browser.

## Auth, Role, dan Scope

- Harus login dengan session Web Admin valid.
- Hanya tenant `owner` dan `admin` yang boleh akses.
- `cashier` harus mendapat `403 FORBIDDEN` dari BFF.
- `superadmin` tidak memakai endpoint tenant ini; platform memakai endpoint `/api/v1/superadmin/*`.
- Semua data wajib scoped ke `business_id` dari user/session server-side.
- Client tidak boleh mengirim `business_id`.
- `outlet_id`, jika dikirim, wajib divalidasi sebagai outlet milik business user.
- BFF wajib melakukan sanitization dan tidak boleh meneruskan field sensitif dari payload backend ke browser.

## Query Parameters

| Field | Type | Required | Default | Validation | Notes |
| --- | --- | --- | --- | --- | --- |
| `date_from` | string | no | business-day today atau default UI | `YYYY-MM-DD` | Awal filter tanggal dalam timezone outlet/business. |
| `date_to` | string | no | sama dengan `date_from` atau default UI | `YYYY-MM-DD`, tidak boleh sebelum `date_from` | Akhir filter tanggal. |
| `outlet_id` | UUID string | no | semua outlet milik business | UUID valid dan harus milik business user | Kosong berarti semua outlet yang diizinkan. |
| `cashier_id` | UUID string | no | semua kasir | UUID valid; harus user/staff dalam business jika backend mendukung validasi | Untuk filter kasir. |
| `payment_method` | string enum | no | semua metode | contoh: `cash`, `qris`, `card`, `transfer`, `ewallet`, `other` | BFF boleh menormalkan label, tapi value disimpan sebagai backend method. |
| `status` | string enum | no | semua status | `paid`, `partial`, `unpaid`, `held`, `payment_pending`, `payment_failed`, `voided`, `refunded`, `pending` | `pending` berarti transaksi dengan payment pending mengikuti pola backend saat ini. |
| `search` | string | no | empty | max 100 chars | Cari nomor transaksi/code, nama customer yang aman, atau nama kasir. Jangan cari payment reference sensitif. |
| `page` | integer | no | `1` | min `1` | Pagination BFF response. |
| `per_page` | integer | no | `20` | min `1`, max `100` | Jangan return list besar ke browser. |

Invalid filter harus menghasilkan `422 VALIDATION_ERROR` dengan detail field yang aman.

## Success Response JSON

```json
{
  "data": {
    "meta": {
      "date_from": "2026-06-21",
      "date_to": "2026-06-21",
      "timezone": "Asia/Jakarta",
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "currency": "IDR",
      "money_format": "integer_rupiah",
      "data_status": "real",
      "generated_at": "2026-06-21T08:30:00.000000Z"
    },
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 42,
      "total_pages": 3,
      "has_next_page": true,
      "has_previous_page": false
    },
    "totals": {
      "transaction_count": 42,
      "gross_sales": 1850000,
      "discount_total": 75000,
      "service_charge_total": 0,
      "tax_total": 0,
      "rounding_total": 0,
      "net_sales": 1775000,
      "paid_amount": 1775000,
      "pending_amount": 0,
      "refund_amount": 0,
      "void_count": 0,
      "is_estimate": false
    },
    "rows": [
      {
        "id": "9d56a07f-6d26-4582-8fda-7f630a9da001",
        "code": "TRX-20260621-ABCD",
        "occurred_at": "2026-06-21T08:42:00.000000Z",
        "business_date": "2026-06-21",
        "time_label": "15:42",
        "outlet": {
          "id": "135159a6-d523-470f-93e1-5a3f3706a001",
          "name": "Cabang Utama"
        },
        "cashier": {
          "id": "6774cdcd-1cb5-4e2e-9886-62f98d19a001",
          "name": "Ayu"
        },
        "customer": {
          "id": "751a5fbe-86ac-4421-8ef7-6ce5b3c1a001",
          "name": "Budi",
          "display_label": "Budi"
        },
        "payment_method": {
          "value": "qris",
          "label": "QRIS",
          "is_mixed": false
        },
        "total": 38000,
        "subtotal": 40000,
        "discount_total": 2000,
        "item_count": 2,
        "status": "paid",
        "status_label": "Lunas",
        "can_open_detail": true
      }
    ],
    "data_notes": [
      {
        "key": "backend_pagination",
        "message": "Backend transactions endpoint saat ini perlu filter dan pagination backend-side untuk volume produksi besar.",
        "severity": "info"
      }
    ]
  },
  "meta": {
    "request_id": "req_01J0TRANSACTIONS001",
    "generated_at": "2026-06-21T08:30:00.000000Z"
  }
}
```

## Field Definition

### `data.meta`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `date_from` | string `YYYY-MM-DD` | yes | `2026-06-21` | Null jika tidak ada tanggal yang dipilih. |
| `date_to` | string `YYYY-MM-DD` | yes | `2026-06-21` | Null jika tidak ada tanggal yang dipilih. |
| `timezone` | string | no | `Asia/Jakarta` | Dari outlet/business fallback. |
| `business_id` | UUID string | no | `8f7f...` | Dari session user, bukan client input. |
| `outlet_id` | UUID string | yes | `1351...` | Null berarti semua outlet dalam business. |
| `currency` | string | no | `IDR` | Tetap IDR. |
| `money_format` | string | no | `integer_rupiah` | Semua money integer rupiah. |
| `data_status` | string enum | no | `real` | `real`, `partial`, atau `unavailable`. |
| `generated_at` | ISO-8601 UTC string | no | `2026-06-21T08:30:00.000000Z` | Waktu server/BFF membuat response. |

### `pagination`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `page` | integer | no | `1` | Halaman saat ini. |
| `per_page` | integer | no | `20` | Jumlah row per halaman. |
| `total` | integer | no | `42` | Total row setelah filter. |
| `total_pages` | integer | no | `3` | Minimal `1` untuk response stabil, meski kosong. |
| `has_next_page` | boolean | no | `true` | Dipakai UI pagination. |
| `has_previous_page` | boolean | no | `false` | Dipakai UI pagination. |

### `totals`

| Field | Type | Nullable | Example | Notes |
| --- | --- | --- | --- | --- |
| `transaction_count` | integer | no | `42` | Jumlah transaksi setelah filter. |
| `gross_sales` | integer money | no | `1850000` | Total sebelum discount/service/tax/rounding jika tersedia. |
| `discount_total` | integer money | no | `75000` | Total discount dari backend. |
| `service_charge_total` | integer money | no | `0` | Dari backend. |
| `tax_total` | integer money | no | `0` | Dari backend. |
| `rounding_total` | integer money | no | `0` | Bisa negatif/positif. |
| `net_sales` | integer money | no | `1775000` | Total akhir dari backend transaction grand total; BFF hanya menjumlahkan row yang sudah server-calculated bila backend belum menyediakan aggregate. |
| `paid_amount` | integer money | no | `1775000` | Jumlah transaksi/payment paid/confirmed jika tersedia. |
| `pending_amount` | integer money | no | `0` | Payment pending jika tersedia. |
| `refund_amount` | integer money | no | `0` | Hanya read-only; tidak trigger refund. |
| `void_count` | integer | no | `0` | Hanya read-only. |
| `is_estimate` | boolean | no | `false` | `true` bila totals berasal dari fallback BFF aggregation karena backend belum punya aggregate khusus. |

### `rows[]`

| Field | Type | Nullable | Example | Privacy/Scope Notes |
| --- | --- | --- | --- | --- |
| `id` | UUID string | no | `9d56...` | Tenant-scoped transaction id. |
| `code` | string | no | `TRX-20260621-ABCD` | Dari `transactions.number`. |
| `occurred_at` | ISO-8601 UTC string | no | `2026-06-21T08:42:00.000000Z` | Dari `transactions.created_at`. |
| `business_date` | string `YYYY-MM-DD` | no | `2026-06-21` | Tanggal bisnis sesuai timezone. |
| `time_label` | string | no | `15:42` | Label display; client boleh render ulang dari timestamp dan timezone. |
| `outlet.id` | UUID string | yes | `1351...` | Hanya outlet milik business user. |
| `outlet.name` | string | yes | `Cabang Utama` | Jangan tampilkan data outlet tenant lain. |
| `cashier.id` | UUID string | yes | `6774...` | Tidak expose email/PIN. |
| `cashier.name` | string | yes | `Ayu` | Safe display name. |
| `customer.id` | UUID string | yes | `751a...` | Nullable. |
| `customer.name` | string | yes | `Budi` | Safe display only; boleh null/masqued bila policy berubah. |
| `customer.display_label` | string | yes | `Budi` | Jangan expose phone/email penuh. |
| `payment_method.value` | string | yes | `qris` | Jika multiple payment, value `mixed`. |
| `payment_method.label` | string | yes | `QRIS` | Display label. |
| `payment_method.is_mixed` | boolean | no | `false` | True jika lebih dari satu method. |
| `total` | integer money | no | `38000` | Dari `grand_total`, bukan hitungan UI. |
| `subtotal` | integer money | no | `40000` | Dari backend. |
| `discount_total` | integer money | no | `2000` | Dari backend. |
| `item_count` | integer | no | `2` | Dari transaction items. |
| `status` | string enum | no | `paid` | Backend status. |
| `status_label` | string | no | `Lunas` | BFF boleh map label Indonesia. |
| `can_open_detail` | boolean | no | `true` | Hanya untuk read-only detail. |

## Field Yang Tidak Boleh Keluar ke Browser

BFF response tidak boleh memuat:

- raw bearer token/session secret;
- `payments.reference`;
- `payments.provider_reference` atau `payment_ref`;
- `payments.failure_reason` bila mengandung detail provider sensitif;
- raw receipt lengkap;
- full email/phone customer;
- staff/cashier email, PIN, password hash;
- `device_id`, `shift_id`, internal lease owner id, atau checkout/quote token;
- payload audit atau request body asli.

## Empty State

Saat tidak ada transaksi sesuai filter, response tetap `200`:

```json
{
  "data": {
    "meta": {
      "date_from": "2026-06-21",
      "date_to": "2026-06-21",
      "timezone": "Asia/Jakarta",
      "business_id": "8f7f1d44-4b20-4f06-a09b-fefccf05a001",
      "outlet_id": null,
      "currency": "IDR",
      "money_format": "integer_rupiah",
      "data_status": "real",
      "generated_at": "2026-06-21T08:30:00.000000Z"
    },
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 0,
      "total_pages": 1,
      "has_next_page": false,
      "has_previous_page": false
    },
    "totals": {
      "transaction_count": 0,
      "gross_sales": 0,
      "discount_total": 0,
      "service_charge_total": 0,
      "tax_total": 0,
      "rounding_total": 0,
      "net_sales": 0,
      "paid_amount": 0,
      "pending_amount": 0,
      "refund_amount": 0,
      "void_count": 0,
      "is_estimate": false
    },
    "rows": [],
    "data_notes": []
  },
  "meta": {
    "request_id": "req_01J0TRANSACTIONS_EMPTY",
    "generated_at": "2026-06-21T08:30:00.000000Z"
  }
}
```

## Error States

| HTTP | Code | Use | Notes |
| --- | --- | --- | --- |
| `401` | `UNAUTHENTICATED` | Session/cookie invalid atau expired | BFF harus clear invalid session jika perlu. |
| `403` | `FORBIDDEN` | Role bukan owner/admin, outlet bukan milik business, atau resource denied | Cashier dan superadmin ditolak untuk tenant transactions. |
| `422` | `VALIDATION_ERROR` | Query invalid | Return detail field aman. |
| `500` | `BACKEND_UNAVAILABLE` atau `SERVER_ERROR` | Laravel unavailable atau unexpected error | Jangan expose stack trace/token. |

## Data Source Backend

Kandidat sumber data untuk implementasi/mapping:

- `transactions`: id, number, status, subtotal, discount_total, service_charge_total, tax_total, rounding_total, grand_total, created_at, outlet_id, cashier_id, customer_id.
- `transaction_items`: item count dan ringkasan item bila detail dibutuhkan.
- `payments`: method, amount, status, is_cash untuk summary metode bayar; reference/provider reference harus disaring.
- `outlets`: outlet name dan timezone, scoped by business.
- `users`: cashier display name; jangan expose email/PIN.
- `customers`: optional display label; jangan expose full phone/email.

## Tenant Isolation Rules

- Semua query backend dan BFF wajib memakai `business_id` dari user/session.
- `outlet_id` harus divalidasi milik business user sebelum dipakai filter.
- `cashier_id` harus divalidasi milik business user jika filter diaktifkan.
- Transaction detail link hanya boleh membuka transaksi dalam business yang sama.
- Superadmin tidak boleh memakai endpoint tenant transactions untuk melihat toko lain.
- BFF tidak boleh cache response lintas business tanpa key tenant aman.

## Client UI Mapping

- `/transactions` memakai `rows` untuk table/list.
- KPI ringkas memakai `totals` dari response, bukan hitungan di Client Component.
- Filter UI hanya boleh mengirim query whitelist kontrak.
- Action sensitif tetap disabled/gated: refund, void, reprint, export, payment retry, shift close/open, cash reconciliation.
- Detail row boleh link ke `/transactions/{id}` untuk read-only detail, tetapi detail contract dapat dibuat terpisah jika UI membutuhkan field lebih banyak.

## Current Backend Alignment Notes

- `GET /api/v1/transactions` route sudah ada dan read-only.
- Saat kontrak ini dibuat, backend list hanya membaca `status` secara eksplisit dan mengembalikan payload transaksi lengkap tanpa pagination khusus Web Admin.
- WEB-02B harus memastikan BFF response tersanitasi dan tidak membocorkan field sensitif dari payload backend.
- Untuk production scale, backend perlu mendukung filter tanggal/outlet/cashier/payment/search dan pagination secara server-side. Ini dicatat sebagai gap backend read-only, bukan blocker untuk finalisasi kontrak.

## Acceptance Criteria untuk WEB-02B

- BFF route `GET /api/admin/transactions` dibuat.
- BFF hanya menerima query whitelist: `date_from`, `date_to`, `outlet_id`, `cashier_id`, `payment_method`, `status`, `search`, `page`, `per_page`.
- BFF membaca session server-side dan menolak unauthenticated `401`.
- BFF menolak role selain owner/admin dengan `403`.
- BFF memanggil Laravel server-side; token tidak keluar ke browser.
- Response BFF mengikuti shape kontrak ini.
- BFF menghapus payment reference, provider reference, raw receipt, email/PIN, token, dan PII berlebih.
- `/transactions` tetap read-only dan menampilkan loading/empty/error/session-required/forbidden state.
- Tidak ada refund/void/reprint/export/payment retry action real.
- Test mencakup 401, 403, validation query, mapping/sanitization, empty state, dan boundary scan.

## Validation Required for Implementation

```bash
cd apps/web
npm run typecheck
npm run lint
npm test
npm run build
npx playwright test --list
rg -n "fetch\(|XMLHttpRequest|sessionStorage|localStorage|NEXT_PUBLIC_API_BASE_URL" src/app src/components src/fixtures --glob '!*.test.ts' --glob '!*.test.tsx' || true
rg -n "Authorization|Bearer " src/app src/components src/fixtures --glob '!*.test.ts' --glob '!*.test.tsx' || true
```
