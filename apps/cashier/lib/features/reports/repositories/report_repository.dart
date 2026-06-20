import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/network/api_client.dart';

final reportRepositoryProvider = Provider<ReportRepository>(
  (ref) => ApiReportRepository(apiClient: ref.watch(apiClientProvider)),
);

abstract interface class ReportRepository {
  Future<SalesSummary> salesSummary({
    DateTime? date,
    String? shiftId,
    String? outletId,
  });

  Future<SoldProductsReport> soldProducts({
    DateTime? date,
    String? shiftId,
    String? outletId,
  });

  Future<PaymentMethodsReport> paymentMethods({
    DateTime? date,
    String? shiftId,
    String? outletId,
  });

  Future<CashierShiftsReport> cashierShifts({
    DateTime? date,
    String? shiftId,
    String? outletId,
  });

  Future<VoidRefundAuditReport> voidRefundAudit({
    DateTime? date,
    String? shiftId,
    String? outletId,
  });

  Future<TopTenReport> topTen({
    DateTime? date,
    String? shiftId,
    String? outletId,
  });
}

class ApiReportRepository implements ReportRepository {
  const ApiReportRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<SalesSummary> salesSummary({
    DateTime? date,
    String? shiftId,
    String? outletId,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/reports/sales-summary',
      queryParameters: _query(date: date, shiftId: shiftId, outletId: outletId),
    );
    return SalesSummary.fromJson(response.data);
  }

  @override
  Future<SoldProductsReport> soldProducts({
    DateTime? date,
    String? shiftId,
    String? outletId,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/reports/sold-products',
      queryParameters: _query(date: date, shiftId: shiftId, outletId: outletId),
    );
    return SoldProductsReport.fromJson(response.data);
  }

  @override
  Future<PaymentMethodsReport> paymentMethods({
    DateTime? date,
    String? shiftId,
    String? outletId,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/reports/payment-methods',
      queryParameters: _query(date: date, shiftId: shiftId, outletId: outletId),
    );
    return PaymentMethodsReport.fromJson(response.data);
  }

  @override
  Future<CashierShiftsReport> cashierShifts({
    DateTime? date,
    String? shiftId,
    String? outletId,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/reports/cashier-shifts',
      queryParameters: _query(date: date, shiftId: shiftId, outletId: outletId),
    );
    return CashierShiftsReport.fromJson(response.data);
  }

  @override
  Future<VoidRefundAuditReport> voidRefundAudit({
    DateTime? date,
    String? shiftId,
    String? outletId,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/reports/void-refund-audit',
      queryParameters: _query(date: date, shiftId: shiftId, outletId: outletId),
    );
    return VoidRefundAuditReport.fromJson(response.data);
  }

  @override
  Future<TopTenReport> topTen({
    DateTime? date,
    String? shiftId,
    String? outletId,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/reports/top-10',
      queryParameters: _query(date: date, shiftId: shiftId, outletId: outletId),
    );
    return TopTenReport.fromJson(response.data);
  }

  Map<String, Object?> _query({
    DateTime? date,
    String? shiftId,
    String? outletId,
  }) {
    return {
      if (date != null) 'date': DateFormat('yyyy-MM-dd').format(date),
      if (shiftId != null && shiftId.isNotEmpty) 'shift_id': shiftId,
      if (outletId != null && outletId.isNotEmpty) 'outlet_id': outletId,
    };
  }
}

class ReportsDashboard {
  const ReportsDashboard({
    required this.salesSummary,
    required this.soldProducts,
    required this.paymentMethods,
    required this.cashierShifts,
    required this.voidRefundAudit,
    required this.topTen,
  });

  final SalesSummary salesSummary;
  final SoldProductsReport soldProducts;
  final PaymentMethodsReport paymentMethods;
  final CashierShiftsReport cashierShifts;
  final VoidRefundAuditReport voidRefundAudit;
  final TopTenReport topTen;
}

class SalesSummary {
  const SalesSummary({
    required this.totalSales,
    required this.transactionCount,
    required this.paymentTotals,
    required this.totalDiscount,
    required this.totalVoid,
    required this.averageTransactionValue,
    required this.chartPoints,
    this.totalServiceCharge = 0,
    this.totalTax = 0,
    this.totalRounding = 0,
  });

  factory SalesSummary.empty() {
    return const SalesSummary(
      totalSales: 0,
      transactionCount: 0,
      paymentTotals: [],
      totalDiscount: 0,
      totalVoid: 0,
      averageTransactionValue: 0,
      chartPoints: [],
    );
  }

  factory SalesSummary.fromJson(Map<String, Object?> json) {
    return SalesSummary(
      totalSales: _int(json['total_sales']),
      transactionCount: _int(json['transaction_count']),
      paymentTotals: [
        for (final value in json['payment_totals'] as List? ?? const [])
          PaymentTotal.fromJson(_asMap(value)),
      ],
      totalDiscount: _int(json['total_discount']),
      totalVoid: _int(json['total_void']),
      averageTransactionValue: _int(json['average_transaction_value']),
      totalServiceCharge: _int(json['total_service_charge']),
      totalTax: _int(json['total_tax']),
      totalRounding: _int(json['total_rounding']),
      chartPoints: [
        for (final value in json['chart'] as List? ?? const [])
          SalesChartPoint.fromJson(_asMap(value)),
      ],
    );
  }

  final int totalSales;
  final int transactionCount;
  final List<PaymentTotal> paymentTotals;
  final int totalDiscount;
  final int totalVoid;
  final int averageTransactionValue;
  final int totalServiceCharge;
  final int totalTax;
  final int totalRounding;
  final List<SalesChartPoint> chartPoints;
}

class PaymentTotal {
  const PaymentTotal({required this.method, required this.amount});

  factory PaymentTotal.fromJson(Map<String, Object?> json) {
    return PaymentTotal(
      method: (json['method'] as String?) ?? '',
      amount: _int(json['amount']),
    );
  }

  final String method;
  final int amount;
}

class SalesChartPoint {
  const SalesChartPoint({
    required this.label,
    required this.amount,
    this.transactionCount = 0,
  });

  factory SalesChartPoint.fromJson(Map<String, Object?> json) {
    return SalesChartPoint(
      label: (json['label'] as String?) ?? '',
      amount: _int(json['amount']),
      transactionCount: _int(json['transaction_count']),
    );
  }

  final String label;
  final int amount;
  final int transactionCount;
}

class SoldProductsReport {
  const SoldProductsReport({required this.rows});

  factory SoldProductsReport.fromJson(Map<String, Object?> json) {
    return SoldProductsReport(
      rows: [
        for (final value in json['rows'] as List? ?? const [])
          SoldProductReportRow.fromJson(_asMap(value)),
      ],
    );
  }

  final List<SoldProductReportRow> rows;
}

class SoldProductReportRow {
  const SoldProductReportRow({
    required this.productId,
    required this.productName,
    required this.categoryName,
    required this.quantitySold,
    required this.grossSales,
    required this.discountTotal,
    required this.netSales,
    required this.outletId,
    required this.outletName,
  });

  factory SoldProductReportRow.fromJson(Map<String, Object?> json) {
    return SoldProductReportRow(
      productId: (json['product_id'] as String?) ?? '',
      productName: (json['product_name'] as String?) ?? '',
      categoryName: json['category_name'] as String?,
      quantitySold: _int(json['quantity_sold']),
      grossSales: _int(json['gross_sales']),
      discountTotal: _int(json['discount_total']),
      netSales: _int(json['net_sales']),
      outletId: (json['outlet_id'] as String?) ?? '',
      outletName: (json['outlet_name'] as String?) ?? '',
    );
  }

  final String productId;
  final String productName;
  final String? categoryName;
  final int quantitySold;
  final int grossSales;
  final int discountTotal;
  final int netSales;
  final String outletId;
  final String outletName;
}

class PaymentMethodsReport {
  const PaymentMethodsReport({required this.rows});

  factory PaymentMethodsReport.fromJson(Map<String, Object?> json) {
    return PaymentMethodsReport(
      rows: [
        for (final value in json['rows'] as List? ?? const [])
          PaymentMethodReportRow.fromJson(_asMap(value)),
      ],
    );
  }

  final List<PaymentMethodReportRow> rows;
}

class PaymentMethodReportRow {
  const PaymentMethodReportRow({
    required this.method,
    required this.transactionCount,
    required this.grossAmount,
    required this.netAmount,
    this.paymentMethodConfigId,
  });

  factory PaymentMethodReportRow.fromJson(Map<String, Object?> json) {
    return PaymentMethodReportRow(
      method: (json['method'] as String?) ?? '',
      paymentMethodConfigId: json['payment_method_config_id'] as String?,
      transactionCount: _int(json['transaction_count']),
      grossAmount: _int(json['gross_amount']),
      netAmount: _int(json['net_amount']),
    );
  }

  final String method;
  final String? paymentMethodConfigId;
  final int transactionCount;
  final int grossAmount;
  final int netAmount;
}

class CashierShiftsReport {
  const CashierShiftsReport({required this.rows});

  factory CashierShiftsReport.fromJson(Map<String, Object?> json) {
    return CashierShiftsReport(
      rows: [
        for (final value in json['rows'] as List? ?? const [])
          CashierShiftReportRow.fromJson(_asMap(value)),
      ],
    );
  }

  final List<CashierShiftReportRow> rows;
}

class CashierShiftReportRow {
  const CashierShiftReportRow({
    required this.shiftId,
    required this.shiftNumber,
    required this.cashier,
    required this.outlet,
    required this.openedAt,
    required this.closedAt,
    required this.openingCash,
    required this.expectedCash,
    required this.declaredClosingCash,
    required this.variance,
    required this.cashIn,
    required this.cashOut,
    required this.refunds,
    required this.status,
  });

  factory CashierShiftReportRow.fromJson(Map<String, Object?> json) {
    return CashierShiftReportRow(
      shiftId: (json['shift_id'] as String?) ?? '',
      shiftNumber: (json['shift_number'] as String?) ?? '',
      cashier: ReportActor.fromJson(_asMap(json['cashier'])),
      outlet: ReportOutlet.fromJson(_asMap(json['outlet'])),
      openedAt: json['opened_at'] as String?,
      closedAt: json['closed_at'] as String?,
      openingCash: _int(json['opening_cash']),
      expectedCash: _int(json['expected_cash']),
      declaredClosingCash: _int(json['declared_closing_cash']),
      variance: _int(json['variance']),
      cashIn: _int(json['cash_in']),
      cashOut: _int(json['cash_out']),
      refunds: _int(json['refunds']),
      status: (json['status'] as String?) ?? '',
    );
  }

  final String shiftId;
  final String shiftNumber;
  final ReportActor cashier;
  final ReportOutlet outlet;
  final String? openedAt;
  final String? closedAt;
  final int openingCash;
  final int expectedCash;
  final int declaredClosingCash;
  final int variance;
  final int cashIn;
  final int cashOut;
  final int refunds;
  final String status;
}

class ReportActor {
  const ReportActor({required this.id, required this.name});

  factory ReportActor.fromJson(Map<String, Object?> json) {
    return ReportActor(
      id: (json['id'] as String?) ?? '',
      name: (json['name'] as String?) ?? '',
    );
  }

  final String id;
  final String name;
}

class ReportOutlet {
  const ReportOutlet({required this.id, required this.name});

  factory ReportOutlet.fromJson(Map<String, Object?> json) {
    return ReportOutlet(
      id: (json['id'] as String?) ?? '',
      name: (json['name'] as String?) ?? '',
    );
  }

  final String id;
  final String name;
}

class VoidRefundAuditReport {
  const VoidRefundAuditReport({required this.rows, required this.refundRows});

  factory VoidRefundAuditReport.fromJson(Map<String, Object?> json) {
    return VoidRefundAuditReport(
      rows: [
        for (final value in json['rows'] as List? ?? const [])
          VoidRefundAuditRow.fromJson(_asMap(value)),
      ],
      refundRows: [
        for (final value in json['refund_rows'] as List? ?? const [])
          VoidRefundAuditRow.fromJson(_asMap(value)),
      ],
    );
  }

  final List<VoidRefundAuditRow> rows;
  final List<VoidRefundAuditRow> refundRows;
}

class VoidRefundAuditRow {
  const VoidRefundAuditRow({
    required this.transactionNumber,
    required this.originalOrderTime,
    required this.voidOrRefundTime,
    required this.type,
    required this.cashier,
    required this.authorizer,
    required this.reason,
    required this.orderType,
    required this.amount,
    required this.refundAmount,
  });

  factory VoidRefundAuditRow.fromJson(Map<String, Object?> json) {
    return VoidRefundAuditRow(
      transactionNumber: (json['transaction_number'] as String?) ?? '',
      originalOrderTime: json['original_order_time'] as String?,
      voidOrRefundTime: json['void_or_refund_time'] as String?,
      type: (json['type'] as String?) ?? '',
      cashier: json['cashier'] as String?,
      authorizer: json['authorizer'] as String?,
      reason: json['reason'] as String?,
      orderType: json['order_type'] as String?,
      amount: _int(json['amount']),
      refundAmount: json['refund_amount'] == null
          ? null
          : _int(json['refund_amount']),
    );
  }

  final String transactionNumber;
  final String? originalOrderTime;
  final String? voidOrRefundTime;
  final String type;
  final String? cashier;
  final String? authorizer;
  final String? reason;
  final String? orderType;
  final int amount;
  final int? refundAmount;
}

class TopTenReport {
  const TopTenReport({
    required this.bestSelling,
    required this.highestOrderTypeValue,
    required this.lowStock,
    required this.highestGrossProfit,
    required this.dataAvailability,
  });

  factory TopTenReport.fromJson(Map<String, Object?> json) {
    return TopTenReport(
      bestSelling: [
        for (final value in json['best_selling'] as List? ?? const [])
          ProductRankingRow.fromJson(_asMap(value)),
      ],
      highestOrderTypeValue: [
        for (final value
            in json['highest_order_type_value'] as List? ?? const [])
          ProductRankingRow.fromJson(_asMap(value)),
      ],
      lowStock: [
        for (final value in json['low_stock'] as List? ?? const [])
          ProductRankingRow.fromJson(_asMap(value)),
      ],
      highestGrossProfit: [
        for (final value in json['highest_gross_profit'] as List? ?? const [])
          ProductRankingRow.fromJson(_asMap(value)),
      ],
      dataAvailability: Map<String, String>.from(
        _asMap(
          json['data_availability'] ?? const <String, Object?>{},
        ).map((key, value) => MapEntry(key, '$value')),
      ),
    );
  }

  final List<ProductRankingRow> bestSelling;
  final List<ProductRankingRow> highestOrderTypeValue;
  final List<ProductRankingRow> lowStock;
  final List<ProductRankingRow> highestGrossProfit;
  final Map<String, String> dataAvailability;
}

class ProductRankingRow {
  const ProductRankingRow({
    required this.productId,
    required this.productName,
    required this.quantitySold,
    required this.amount,
  });

  factory ProductRankingRow.fromJson(Map<String, Object?> json) {
    return ProductRankingRow(
      productId: (json['product_id'] as String?) ?? '',
      productName:
          (json['product_name'] as String?) ?? (json['name'] as String?) ?? '',
      quantitySold: _int(json['quantity_sold']),
      amount: _int(json['amount'] ?? json['gross_profit'] ?? json['total']),
    );
  }

  final String productId;
  final String productName;
  final int quantitySold;
  final int amount;
}

int _int(Object? value) => (value as num?)?.toInt() ?? 0;

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}
