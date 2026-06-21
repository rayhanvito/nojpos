# Contract - Tenant Dashboard Summary

Status: Draft
Owner: Backend Core + Web Admin
Related stories: BE-04, WEB-01, INT-02

## Goal

Menyediakan satu endpoint agregat ringan untuk dashboard Owner/Admin Toko agar web tidak perlu memanggil banyak endpoint sekaligus.

## Endpoint

- Method: `GET`
- Path: `/api/v1/dashboard/summary`
- Auth: required
- Roles: owner, admin, manager if approved
- Tenant scope: authenticated `business_id`
- Outlet scope: optional query, must be verified server-side

## Query Params

| Field | Type | Required | Notes |
| --- | --- | --- | --- |
| `date` | string `YYYY-MM-DD` | optional | default outlet business day today |
| `outlet_id` | uuid or `all` | optional | default allowed outlets/all based on role |

## Success Response

All money values are integer rupiah. Client formats display.

```json
{
  "data": {
    "context": {
      "date": "2026-06-21",
      "timezone": "Asia/Jakarta",
      "business_id": "uuid",
      "outlet_id": "all",
      "is_preview": false
    },
    "kpis": {
      "sales_today": 2450000,
      "transaction_count": 86,
      "average_transaction": 28500,
      "gross_profit_estimate": 740000,
      "low_stock_count": 12,
      "cash_difference": 45000
    },
    "alerts": [
      {
        "id": "low-stock",
        "type": "stock",
        "severity": "warning",
        "title": "12 produk stok menipis",
        "message": "Cek stok sebelum jam ramai."
      }
    ],
    "sales_last_7_days": [
      {
        "label": "Sen",
        "date": "2026-06-15",
        "sales": 2100000,
        "transactions": 74
      }
    ],
    "payment_methods": [
      {
        "method": "QRIS",
        "amount": 1320000,
        "transactions": 41
      }
    ],
    "top_products": [
      {
        "name": "Es Kopi Susu",
        "qty_sold": 42,
        "sales": 756000
      }
    ],
    "low_stock_items": [
      {
        "name": "Cup 16oz",
        "remaining_stock": 8,
        "status": "low"
      }
    ],
    "recent_transactions": [
      {
        "code": "TRX-1024",
        "time": "14:30",
        "payment_method": "QRIS",
        "total": 58000
      }
    ],
    "cashier_performance": [
      {
        "name": "Rani",
        "transaction_count": 32,
        "sales": 890000,
        "note": "Perlu cek void/refund jika ada"
      }
    ],
    "branch_highlights": [
      {
        "name": "Cabang Utama",
        "status": "normal",
        "summary": "Penjualan stabil"
      }
    ]
  },
  "meta": {}
}
```

## Error Response

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Anda tidak memiliki akses ke cabang ini.",
    "details": {}
  }
}
```

Possible codes:

- `UNAUTHENTICATED`
- `FORBIDDEN`
- `VALIDATION_ERROR`
- `NOT_FOUND`

## Empty State

If no transactions yet:

- kpi money values return `0`
- lists return `[]`
- alerts may include setup guidance
- do not return fake production data

## Business Rules

- Sales and transaction counts must be calculated server-side.
- Gross profit is estimate if cost data is incomplete; response may include note if needed.
- Cash difference must come from server-side shift/cash data.
- Low stock uses server-side threshold.
- Outlet timezone controls business day boundary.

## Client UI Mapping

- Dashboard KPI cards use `kpis`.
- Alerts section uses `alerts`.
- Recharts sales chart uses `sales_last_7_days`.
- Recharts payment chart uses `payment_methods`.
- Operational lists use the matching list arrays.

## Tests Required

Backend:

- owner sees only own business data
- admin sees only allowed outlet/business data
- forbidden for another outlet/business
- empty business returns zero/empty state
- date validation
- timezone business day test

Web:

- loading state
- empty state
- forbidden state
- error state
- no horizontal overflow
- no client-side money calculation beyond formatting
