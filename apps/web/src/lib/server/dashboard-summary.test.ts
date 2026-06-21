import { describe, expect, it } from 'vitest';

import { BackendApiError, type DashboardSummaryData } from './backend-client';
import { mapDashboardSummaryEnvelopeToViewModel, parseDashboardSummaryQuery } from './dashboard-summary';

function dashboardSummary(): DashboardSummaryData {
  return {
    meta: {
      date: '2026-06-21',
      range: 'last_7_days',
      timezone: 'Asia/Jakarta',
      business_id: 'business-1',
      outlet_id: null,
      outlet_name: null,
      generated_at: '2026-06-21T08:30:00.000000Z',
      currency: 'IDR',
      money_format: 'integer_rupiah',
      data_status: 'real',
      is_empty_today: false,
    },
    kpis: {
      sales_today: { value: 2450000, type: 'money', label: 'Penjualan hari ini', trend_label: '+8% vs kemarin', is_estimate: false },
      transaction_count: { value: 86, type: 'integer', label: 'Jumlah transaksi', trend_label: 'Ramai stabil', is_estimate: false },
      average_transaction: { value: 28500, type: 'money', label: 'Rata-rata transaksi', trend_label: null, is_estimate: false },
      gross_profit_estimate: { value: 0, type: 'money', label: 'Gross profit estimasi', trend_label: null, is_estimate: true },
      low_stock_count: { value: 2, type: 'integer', label: 'Stok hampir habis', trend_label: 'Butuh cek', is_estimate: false },
      cash_difference: { value: 0, type: 'money', label: 'Selisih kas', trend_label: null, is_estimate: true },
    },
    alerts: [
      {
        id: 'low-stock-1',
        type: 'low_stock',
        severity: 'warning',
        title: '2 produk stok menipis',
        message: 'Cek stok sebelum jam ramai.',
        outlet_id: null,
        outlet_name: null,
        entity_type: 'inventory',
        entity_id: null,
        action_label: 'Cek inventori',
        action_path: '/inventory',
      },
    ],
    sales_last_7_days: [{ date: '2026-06-21', label: 'Min', sales: 2450000, transaction_count: 86, average_transaction: 28500 }],
    payment_methods: [{ method: 'qris', label: 'QRIS', amount: 1320000, transaction_count: 41, share_percent: 53.88 }],
    top_products: [{ product_id: 'product-1', name: 'Es Kopi', qty_sold: 34, sales: 612000, share_percent: 86, outlet_id: null, outlet_name: null }],
    low_stock_items: [
      { product_id: 'product-2', name: 'Cup 16 oz', sku: null, outlet_id: 'outlet-1', outlet_name: 'Cabang Utama', remaining_stock: 0, threshold: 30, unit: 'pcs', status: 'out' },
    ],
    recent_transactions: [
      { transaction_id: 'trx-1', code: 'TRX-1028', time: '15:42', occurred_at: '2026-06-21T08:42:00.000000Z', outlet_id: 'outlet-1', outlet_name: 'Cabang Utama', cashier_name: 'Ayu', payment_method: 'QRIS', total: 38000, status: 'paid' },
    ],
    cashier_performance: [
      { cashier_id: 'cashier-1', name: 'Ayu', transaction_count: 31, sales: 920000, average_transaction: 29677, void_count: 0, refund_count: 0, cash_difference: 0, note: 'Shift pagi rapi' },
    ],
    branch_highlights: [
      { outlet_id: 'outlet-1', name: 'Cabang Utama', status: 'normal', summary: 'Penjualan stabil.', sales_today: 1750000, transaction_count: 58, low_stock_count: 4, open_shift_count: 1, cash_difference: 0, severity: 'success' },
    ],
    data_notes: [{ key: 'gross_profit_estimate', message: 'Gross profit butuh data cost lengkap.', severity: 'info' }],
  };
}

describe('dashboard summary mapping', () => {
  it('maps backend dashboard summary into stable dashboard view data', () => {
    const viewModel = mapDashboardSummaryEnvelopeToViewModel({ data: dashboardSummary(), meta: { request_id: 'req-1' } });

    expect(viewModel.kpiCards).toHaveLength(6);
    expect(viewModel.kpiCards[0]).toMatchObject({ label: 'Penjualan hari ini', value: 'Rp2.450.000', tone: 'success' });
    expect(viewModel.alerts[0]).toMatchObject({ title: '2 produk stok menipis', tone: 'warning' });
    expect(viewModel.salesLast7Days[0]).toMatchObject({ day: 'Min', sales: 2450000, label: 'Rp2,5 jt' });
    expect(viewModel.paymentMethods[0]).toMatchObject({ method: 'QRIS', total: 1320000 });
    expect(viewModel.lowStockItems[0]).toMatchObject({ name: 'Cup 16 oz', stock: '0 pcs', status: 'Habis' });
    expect(viewModel.recentTransactions[0]).toMatchObject({ code: 'TRX-1028', method: 'QRIS', total: 'Rp38.000' });
    expect(viewModel.dataNotes).toEqual(['Gross profit butuh data cost lengkap.']);
  });

  it('accepts only dashboard query parameters from the contract', () => {
    const query = parseDashboardSummaryQuery(new URLSearchParams('date=2026-06-21&range=last_7_days&outlet_id=135159a6-d523-470f-93e1-5a3f3706a001'));

    expect(query).toEqual({ date: '2026-06-21', outlet_id: '135159a6-d523-470f-93e1-5a3f3706a001', range: 'last_7_days' });
  });

  it('rejects unsupported dashboard query parameters', () => {
    expect(() => parseDashboardSummaryQuery(new URLSearchParams('debug=true'))).toThrow(BackendApiError);
  });
});
