import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/features/reports/repositories/report_repository.dart';
import 'package:nojpos_tablet_ui/features/reports/widgets/report_panels.dart';

void main() {
  test('repository parses sales summary', () async {
    final repository = _repositoryWith({
      'GET /reports/sales-summary': _envelope({
        'total_sales': 33000,
        'transaction_count': 2,
        'payment_totals': [
          {'method': 'cash', 'amount': 11000},
          {'method': 'qris', 'amount': 22000},
        ],
        'total_discount': 3000,
        'total_service_charge': 1500,
        'total_tax': 3000,
        'total_rounding': -1500,
        'total_void': 33000,
        'average_transaction_value': 16500,
        'chart': [
          {
            'label': '2026-06-20 10:00',
            'amount': 33000,
            'transaction_count': 2,
          },
        ],
      }),
    });

    final summary = await repository.salesSummary(date: DateTime(2026, 6, 20));

    expect(summary.totalSales, 33000);
    expect(summary.totalTax, 3000);
    expect(summary.totalServiceCharge, 1500);
    expect(summary.totalRounding, -1500);
    expect(summary.averageTransactionValue, 16500);
    expect(summary.paymentTotals.last.method, 'qris');
    expect(summary.chartPoints.single.transactionCount, 2);
  });

  test('repository parses sold products', () async {
    final repository = _repositoryWith({
      'GET /reports/sold-products': _envelope({
        'rows': [
          {
            'product_id': 'product-id',
            'product_name': 'Nasi Goreng',
            'category_name': 'Makanan',
            'quantity_sold': 3,
            'gross_sales': 75000,
            'discount_total': 3000,
            'net_sales': 72000,
            'outlet_id': 'outlet-id',
            'outlet_name': 'Outlet Utama',
          },
        ],
      }),
    });

    final report = await repository.soldProducts();

    expect(report.rows.single.productName, 'Nasi Goreng');
    expect(report.rows.single.quantitySold, 3);
    expect(report.rows.single.grossSales, 75000);
    expect(report.rows.single.discountTotal, 3000);
    expect(report.rows.single.netSales, 72000);
  });

  test('repository parses payment methods', () async {
    final repository = _repositoryWith({
      'GET /reports/payment-methods': _envelope({
        'rows': [
          {
            'method': 'qris',
            'payment_method_config_id': 'config-id',
            'transaction_count': 4,
            'gross_amount': 125000,
            'net_amount': 125000,
          },
        ],
      }),
    });

    final report = await repository.paymentMethods();

    expect(report.rows.single.method, 'qris');
    expect(report.rows.single.paymentMethodConfigId, 'config-id');
    expect(report.rows.single.transactionCount, 4);
    expect(report.rows.single.grossAmount, 125000);
  });

  test('repository parses cashier shifts', () async {
    final repository = _repositoryWith({
      'GET /reports/cashier-shifts': _envelope({
        'rows': [
          {
            'shift_id': 'shift-id',
            'shift_number': 'shift-id',
            'cashier': {'id': 'cashier-id', 'name': 'Siti Kasir'},
            'outlet': {'id': 'outlet-id', 'name': 'Outlet Utama'},
            'opened_at': '2026-06-20T01:00:00Z',
            'closed_at': '2026-06-20T09:00:00Z',
            'opening_cash': 100000,
            'expected_cash': 125000,
            'declared_closing_cash': 123000,
            'variance': -2000,
            'cash_in': 30000,
            'cash_out': 5000,
            'refunds': 0,
            'status': 'closed',
          },
        ],
      }),
    });

    final report = await repository.cashierShifts();

    expect(report.rows.single.cashier.name, 'Siti Kasir');
    expect(report.rows.single.outlet.name, 'Outlet Utama');
    expect(report.rows.single.expectedCash, 125000);
    expect(report.rows.single.variance, -2000);
    expect(report.rows.single.cashIn, 30000);
  });

  test('repository parses void refund audit', () async {
    final repository = _repositoryWith({
      'GET /reports/void-refund-audit': _envelope({
        'rows': [
          {
            'transaction_number': 'TRX-001',
            'original_order_time': '2026-06-20T01:00:00Z',
            'void_or_refund_time': '2026-06-20T02:00:00Z',
            'type': 'void',
            'cashier': 'Siti Kasir',
            'authorizer': null,
            'reason': 'Salah input',
            'order_type': null,
            'amount': 15000,
            'refund_amount': null,
          },
        ],
        'refund_rows': [],
      }),
    });

    final report = await repository.voidRefundAudit();

    expect(report.rows.single.transactionNumber, 'TRX-001');
    expect(report.rows.single.reason, 'Salah input');
    expect(report.rows.single.amount, 15000);
    expect(report.refundRows, isEmpty);
  });

  test(
    'repository parses top ten best selling and gross profit availability',
    () async {
      final repository = _repositoryWith({
        'GET /reports/top-10': _envelope({
          'best_selling': [
            {
              'product_id': 'product-id',
              'product_name': 'Best Seller',
              'quantity_sold': 5,
            },
          ],
          'highest_order_type_value': [],
          'low_stock': [],
          'highest_gross_profit': [],
          'data_availability': {'gross_profit': 'cost_coverage_incomplete'},
        }),
      });

      final report = await repository.topTen();

      expect(report.bestSelling.single.productName, 'Best Seller');
      expect(report.bestSelling.single.quantitySold, 5);
      expect(report.highestGrossProfit, isEmpty);
      expect(
        report.dataAvailability['gross_profit'],
        'cost_coverage_incomplete',
      );
    },
  );

  testWidgets('empty report panels render empty states', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: Column(
            children: [
              SoldProductsReportPanel(report: SoldProductsReport(rows: [])),
              PaymentMethodsReportPanel(report: PaymentMethodsReport(rows: [])),
              CashierShiftsReportPanel(report: CashierShiftsReport(rows: [])),
              VoidRefundAuditReportPanel(
                report: VoidRefundAuditReport(rows: [], refundRows: []),
              ),
              TopTenReportPanel(
                report: TopTenReport(
                  bestSelling: [],
                  highestOrderTypeValue: [],
                  lowStock: [],
                  highestGrossProfit: [],
                  dataAvailability: {},
                ),
              ),
            ],
          ),
        ),
      ),
    );

    expect(find.text('Belum ada produk terjual'), findsOneWidget);
    expect(find.text('Belum ada pembayaran terkonfirmasi'), findsOneWidget);
    expect(find.text('Belum ada shift pada scope ini'), findsOneWidget);
    expect(find.text('Belum ada void atau refund'), findsOneWidget);
    expect(find.text('Belum ada data top 10'), findsOneWidget);
  });

  testWidgets(
    'cashier filter chips do not expose cross-outlet or cross-cashier filters',
    (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(body: ReportsFilterChips(isCashier: true)),
        ),
      );

      expect(find.text('Shift aktif kasir'), findsOneWidget);
      expect(find.text('Outlet saat ini'), findsOneWidget);
      expect(find.text('Semua outlet'), findsNothing);
      expect(find.text('Semua kasir'), findsNothing);
    },
  );
}

ApiReportRepository _repositoryWith(
  Map<String, Map<String, Object?>> responses,
) {
  return ApiReportRepository(
    apiClient: ApiClient(
      dio: Dio()..httpClientAdapter = _RouteAdapter(responses),
    ),
  );
}

Map<String, Object?> _envelope(Map<String, Object?> data) => {
  'data': data,
  'meta': {},
};

class _RouteAdapter implements HttpClientAdapter {
  _RouteAdapter(this.responses);

  final Map<String, Map<String, Object?>> responses;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final key = '${options.method} ${options.path}';
    final response = responses[key];
    if (response == null) {
      return ResponseBody.fromString(
        jsonEncode({
          'error': {'code': 'NOT_FOUND', 'message': key, 'details': {}},
        }),
        404,
      );
    }

    return ResponseBody.fromString(
      jsonEncode(response),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
