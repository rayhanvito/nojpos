import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../shared/models/nojpos_models.dart';

final transactionRepositoryProvider = Provider<TransactionRepository>(
  (ref) => ApiTransactionRepository(apiClient: ref.watch(apiClientProvider)),
);

abstract interface class TransactionRepository {
  Future<CheckoutQuote> quoteTransaction(CheckoutDraft draft);

  Future<CheckoutTransaction> createTransaction(CheckoutDraft draft);

  Future<CheckoutRecovery> recoverCheckout(String idempotencyKey);

  Future<CheckoutPaymentLine> confirmPayment({
    required String paymentId,
    required String idempotencyKey,
  });

  Future<List<CheckoutTransaction>> listTransactions({String? status});

  Future<List<CheckoutTransaction>> listParkedOrders();

  Future<CheckoutTransaction> createParkedOrder(CheckoutDraft draft);

  Future<CheckoutTransaction> updateParkedOrder({
    required String transactionId,
    required String deviceId,
    required int expectedRevision,
    required List<CheckoutItem> items,
    String? customerId,
    String? servedBy,
    int cartDiscount = 0,
    String? notes,
  });

  Future<CheckoutTransaction> acquireParkedOrderLease({
    required String transactionId,
    required String deviceId,
  });

  Future<CheckoutTransaction> refreshParkedOrderLease({
    required String transactionId,
    required String deviceId,
  });

  Future<CheckoutTransaction> releaseParkedOrderLease({
    required String transactionId,
    required String deviceId,
  });

  Future<VoidResult> voidTransaction({
    required String transactionId,
    required String shiftId,
    required String reason,
    required String idempotencyKey,
  });

  SalesTransaction createLocalTransaction({
    required SalesOrder order,
    required List<PaymentLine> payments,
    required Employee cashier,
  });
}

class CheckoutDraft {
  const CheckoutDraft({
    required this.idempotencyKey,
    required this.outletId,
    required this.deviceId,
    required this.cashierId,
    required this.shiftId,
    required this.items,
    required this.payments,
    required this.paidAmount,
    this.customerId,
    this.servedBy,
    this.cartDiscount = 0,
    this.promotionCodes = const [],
    this.appliedPromotionIds = const [],
    this.rounding = 0,
    this.quote,
    this.notes,
    this.status,
  });

  factory CheckoutDraft.fromJson(Map<String, Object?> json) {
    return CheckoutDraft(
      idempotencyKey: json['idempotency_key'] as String,
      outletId: json['outlet_id'] as String,
      deviceId: json['device_id'] as String,
      cashierId: json['cashier_id'] as String,
      shiftId: json['shift_id'] as String,
      customerId: json['customer_id'] as String?,
      servedBy: json['served_by'] as String?,
      items: [
        for (final item in json['items'] as List? ?? const [])
          CheckoutItem.fromJson(_asMap(item)),
      ],
      payments: [
        for (final payment in json['payments'] as List? ?? const [])
          CheckoutPayment.fromJson(_asMap(payment)),
      ],
      paidAmount: (json['paid_amount'] as num).toInt(),
      cartDiscount: (json['cart_discount'] as num?)?.toInt() ?? 0,
      promotionCodes: [
        for (final code in json['promotion_codes'] as List? ?? const [])
          code.toString(),
      ],
      appliedPromotionIds: [
        for (final id in json['applied_promotion_ids'] as List? ?? const [])
          id.toString(),
      ],
      rounding: (json['rounding'] as num?)?.toInt() ?? 0,
      quote: json['quote'] == null
          ? null
          : CheckoutQuote.fromJson(json['quote']),
      notes: json['notes'] as String?,
      status: json['status'] as String?,
    );
  }

  final String idempotencyKey;
  final String outletId;
  final String deviceId;
  final String cashierId;
  final String shiftId;
  final String? customerId;
  final String? servedBy;
  final List<CheckoutItem> items;
  final List<CheckoutPayment> payments;
  final int paidAmount;
  final int cartDiscount;
  final List<String> promotionCodes;
  final List<String> appliedPromotionIds;
  final int rounding;
  final CheckoutQuote? quote;
  final String? notes;
  final String? status;

  int get subtotal => items.fold(0, (sum, item) => sum + item.subtotal);

  int get itemDiscountTotal =>
      items.fold(0, (sum, item) => sum + item.discount);

  int get discountTotal => itemDiscountTotal + cartDiscount;

  int get baseTotal => (subtotal - discountTotal).clamp(0, subtotal);

  int get grandTotal => baseTotal + rounding;

  int get change => (paidAmount - grandTotal).clamp(0, paidAmount);

  Map<String, Object?> toApiJson() {
    return {
      'outlet_id': outletId,
      'device_id': deviceId,
      'cashier_id': cashierId,
      'shift_id': shiftId,
      if (customerId != null && customerId!.isNotEmpty)
        'customer_id': customerId,
      if (servedBy != null && servedBy!.isNotEmpty) 'served_by': servedBy,
      'cart_discount': cartDiscount,
      if (promotionCodes.isNotEmpty) 'promotion_codes': promotionCodes,
      if (appliedPromotionIds.isNotEmpty)
        'applied_promotion_ids': appliedPromotionIds,
      'rounding': rounding,
      if (notes != null && notes!.isNotEmpty) 'notes': notes,
      if (status != null && status!.isNotEmpty) 'status': status,
      if (quote != null) ...quote!.toApiJson(),
      'items': [for (final item in items) item.toJson()],
      if (payments.isNotEmpty)
        'payments': [for (final payment in payments) payment.toJson()],
    };
  }

  Map<String, Object?> toJson() {
    return {
      'idempotency_key': idempotencyKey,
      'outlet_id': outletId,
      'device_id': deviceId,
      'cashier_id': cashierId,
      'shift_id': shiftId,
      if (customerId != null) 'customer_id': customerId,
      if (servedBy != null) 'served_by': servedBy,
      'paid_amount': paidAmount,
      'cart_discount': cartDiscount,
      if (promotionCodes.isNotEmpty) 'promotion_codes': promotionCodes,
      if (appliedPromotionIds.isNotEmpty)
        'applied_promotion_ids': appliedPromotionIds,
      'rounding': rounding,
      if (quote != null) 'quote': quote!.toApiJson(),
      if (notes != null) 'notes': notes,
      if (status != null) 'status': status,
      'items': [for (final item in items) item.toJson()],
      'payments': [for (final payment in payments) payment.toJson()],
    };
  }
}

class CheckoutItem {
  const CheckoutItem({
    required this.productId,
    required this.name,
    required this.quantity,
    required this.unitPrice,
    this.discount = 0,
  });

  factory CheckoutItem.fromJson(Map<String, Object?> json) {
    return CheckoutItem(
      productId: (json['product_id'] ?? json['productId'] ?? json['id'] ?? '')
          .toString(),
      name: (json['name'] as String?) ?? '',
      quantity: (json['quantity'] as num? ?? json['qty'] as num? ?? 0).toInt(),
      unitPrice:
          (json['unit_price'] as num? ??
                  json['unitPrice'] as num? ??
                  json['price'] as num? ??
                  0)
              .toInt(),
      discount: (json['discount'] as num?)?.toInt() ?? 0,
    );
  }

  final String productId;
  final String name;
  final int quantity;
  final int unitPrice;
  final int discount;

  int get subtotal => quantity * unitPrice;

  Map<String, Object?> toJson() {
    return {
      'product_id': productId,
      'quantity': quantity,
      'unit_price': unitPrice,
      'discount': discount,
    };
  }
}

class CheckoutPayment {
  const CheckoutPayment({
    required this.method,
    required this.amount,
    required this.isCash,
    this.reference,
  });

  factory CheckoutPayment.fromJson(Map<String, Object?> json) {
    return CheckoutPayment(
      method: json['method'] as String,
      amount: (json['amount'] as num).toInt(),
      isCash: json['is_cash'] as bool? ?? false,
      reference: json['reference'] as String?,
    );
  }

  final String method;
  final int amount;
  final bool isCash;
  final String? reference;

  Map<String, Object?> toJson() {
    return {
      'method': method,
      'amount': amount,
      if (reference != null && reference!.isNotEmpty) 'reference': reference,
    };
  }
}

class CheckoutQuote {
  const CheckoutQuote({
    required this.subtotal,
    required this.itemDiscountTotal,
    required this.cartDiscountTotal,
    required this.discountTotal,
    required this.serviceChargeTotal,
    required this.taxTotal,
    required this.roundingTotal,
    required this.grandTotal,
    this.quoteId = '',
    this.quoteHash = '',
    this.quoteRevision = '',
    this.expiresAt = '',
    this.serverTime = '',
    this.configVersion = '',
    this.checkoutIdempotencyKey = '',
    this.checkoutToken = '',
    this.promotionDiscountTotal = 0,
    this.manualDiscountTotal = 0,
    this.appliedPromotions = const [],
    this.rejectedPromotions = const [],
  });

  factory CheckoutQuote.fromJson(Object? value) {
    final json = _asMap(value);
    final totals = _asMap(json['totals']);
    int intValue(String key) {
      return (json[key] as num? ?? totals[key] as num? ?? 0).toInt();
    }

    return CheckoutQuote(
      quoteId: (json['quote_id'] as String?) ?? '',
      quoteHash: (json['quote_hash'] as String?) ?? '',
      quoteRevision:
          (json['quote_revision'] as String?) ??
          (json['config_version'] as String?) ??
          '',
      expiresAt: (json['expires_at'] as String?) ?? '',
      serverTime: (json['server_time'] as String?) ?? '',
      configVersion: (json['config_version'] as String?) ?? '',
      checkoutIdempotencyKey:
          (json['checkout_idempotency_key'] as String?) ?? '',
      checkoutToken:
          (json['checkout_token'] as String?) ??
          (json['quote_token'] as String?) ??
          '',
      subtotal: intValue('subtotal'),
      itemDiscountTotal: intValue('item_discount_total'),
      cartDiscountTotal: intValue('cart_discount_total'),
      promotionDiscountTotal: intValue('promotion_discount_total'),
      manualDiscountTotal: intValue('manual_discount_total'),
      discountTotal: intValue('discount_total'),
      serviceChargeTotal: intValue('service_charge_total'),
      taxTotal: intValue('tax_total'),
      roundingTotal: intValue('rounding_total'),
      grandTotal: intValue('grand_total'),
      appliedPromotions: [
        for (final promotion in json['applied_promotions'] as List? ?? const [])
          AppliedPromotion.fromJson(promotion),
      ],
      rejectedPromotions: [
        for (final promotion
            in json['rejected_promotions'] as List? ?? const [])
          RejectedPromotion.fromJson(promotion),
      ],
    );
  }

  final String quoteId;
  final String quoteHash;
  final String quoteRevision;
  final String expiresAt;
  final String serverTime;
  final String configVersion;
  final String checkoutIdempotencyKey;
  final String checkoutToken;
  final int subtotal;
  final int itemDiscountTotal;
  final int cartDiscountTotal;
  final int promotionDiscountTotal;
  final int manualDiscountTotal;
  final int discountTotal;
  final int serviceChargeTotal;
  final int taxTotal;
  final int roundingTotal;
  final int grandTotal;
  final List<AppliedPromotion> appliedPromotions;
  final List<RejectedPromotion> rejectedPromotions;

  Map<String, Object?> toApiJson() {
    return {
      if (quoteId.isNotEmpty) 'quote_id': quoteId,
      if (checkoutToken.isNotEmpty) 'checkout_token': checkoutToken,
      if (checkoutToken.isNotEmpty) 'quote_token': checkoutToken,
      if (quoteHash.isNotEmpty) 'quote_hash': quoteHash,
      if (quoteRevision.isNotEmpty) 'quote_revision': quoteRevision,
      'subtotal': subtotal,
      'item_discount_total': itemDiscountTotal,
      'cart_discount_total': cartDiscountTotal,
      'promotion_discount_total': promotionDiscountTotal,
      'manual_discount_total': manualDiscountTotal,
      'discount_total': discountTotal,
      'service_charge_total': serviceChargeTotal,
      'tax_total': taxTotal,
      'rounding_total': roundingTotal,
      'grand_total': grandTotal,
    };
  }
}

class AppliedPromotion {
  const AppliedPromotion({
    required this.promotionId,
    required this.name,
    required this.discountAmount,
    this.code,
    this.type = '',
    this.affectedItems = const [],
  });

  factory AppliedPromotion.fromJson(Object? value) {
    final json = _asMap(value);
    return AppliedPromotion(
      promotionId: (json['promotion_id'] as String?) ?? '',
      name: (json['name'] as String?) ?? '',
      code: json['code'] as String?,
      type: (json['type'] as String?) ?? '',
      discountAmount: (json['discount_amount'] as num?)?.toInt() ?? 0,
      affectedItems: [
        for (final item in json['affected_items'] as List? ?? const [])
          PromotionAffectedItem.fromJson(item),
      ],
    );
  }

  final String promotionId;
  final String name;
  final String? code;
  final String type;
  final int discountAmount;
  final List<PromotionAffectedItem> affectedItems;
}

class RejectedPromotion {
  const RejectedPromotion({
    required this.reasonCode,
    required this.message,
    this.code,
    this.promotionId,
    this.name,
  });

  factory RejectedPromotion.fromJson(Object? value) {
    final json = _asMap(value);
    return RejectedPromotion(
      code: json['code'] as String?,
      promotionId: json['promotion_id'] as String?,
      name: json['name'] as String?,
      reasonCode: (json['reason_code'] as String?) ?? '',
      message: (json['message'] as String?) ?? '',
    );
  }

  final String? code;
  final String? promotionId;
  final String? name;
  final String reasonCode;
  final String message;
}

class PromotionAffectedItem {
  const PromotionAffectedItem({
    required this.productId,
    required this.discountAmount,
  });

  factory PromotionAffectedItem.fromJson(Object? value) {
    final json = _asMap(value);
    return PromotionAffectedItem(
      productId: (json['product_id'] as String?) ?? '',
      discountAmount: (json['discount_amount'] as num?)?.toInt() ?? 0,
    );
  }

  final String productId;
  final int discountAmount;
}

class CheckoutTransaction {
  const CheckoutTransaction({
    required this.id,
    required this.number,
    required this.status,
    required this.grandTotal,
    this.payments = const [],
    this.items = const [],
    this.itemDiscountTotal = 0,
    this.cartDiscountTotal = 0,
    this.promotionDiscountTotal = 0,
    this.customerId,
    this.customerName,
    this.cashierId,
    this.cashierName,
    this.cashierRole,
    this.createdAt,
    this.notes,
    this.servedBy,
    this.revision = 1,
    this.leaseDeviceId,
    this.leaseUserId,
    this.leaseExpiresAt,
    this.leaseRemainingSeconds = 0,
  });

  factory CheckoutTransaction.fromJson(Object? value) {
    final json = _asMap(value);
    final customer = _asMap(json['customer']);
    final cashier = _asMap(json['cashier'] ?? json['user']);
    return CheckoutTransaction(
      id: (json['id'] as String?) ?? '',
      number: (json['number'] as String?) ?? '',
      status: (json['status'] as String?) ?? '',
      revision: (json['revision'] as num?)?.toInt() ?? 1,
      leaseDeviceId: _asMap(json['lease'])['device_id'] as String?,
      leaseUserId: _asMap(json['lease'])['user_id'] as String?,
      leaseExpiresAt: _asMap(json['lease'])['expires_at'] as String?,
      leaseRemainingSeconds:
          (_asMap(json['lease'])['remaining_seconds'] as num?)?.toInt() ?? 0,
      grandTotal: (json['grand_total'] as num? ?? json['total'] as num? ?? 0)
          .toInt(),
      payments: [
        for (final payment in json['payments'] as List? ?? const [])
          CheckoutPaymentLine.fromJson(payment),
      ],
      items: [
        for (final item
            in (json['items'] as List? ?? json['lines'] as List? ?? const []))
          CheckoutItem.fromJson(_asMap(item)),
      ],
      itemDiscountTotal: (json['item_discount_total'] as num?)?.toInt() ?? 0,
      cartDiscountTotal: (json['cart_discount_total'] as num?)?.toInt() ?? 0,
      promotionDiscountTotal:
          (json['promotion_discount_total'] as num?)?.toInt() ?? 0,
      customerId:
          (customer['id'] as String?) ?? (json['customer_id'] as String?),
      customerName: customer['name'] as String?,
      cashierId: (cashier['id'] as String?) ?? (json['cashier_id'] as String?),
      cashierName: cashier['name'] as String?,
      cashierRole: (cashier['role'] as String?) ?? 'cashier',
      createdAt: _parseDateTime(
        json['created_at'] ?? json['paid_at'] ?? json['updated_at'],
      ),
      notes: json['notes'] as String?,
      servedBy: json['served_by'] as String?,
    );
  }

  final String id;
  final String number;
  final String status;
  final int grandTotal;
  final List<CheckoutPaymentLine> payments;
  final List<CheckoutItem> items;
  final int itemDiscountTotal;
  final int cartDiscountTotal;
  final int promotionDiscountTotal;
  final String? customerId;
  final String? customerName;
  final String? cashierId;
  final String? cashierName;
  final String? cashierRole;
  final DateTime? createdAt;
  final String? notes;
  final String? servedBy;
  final int revision;
  final String? leaseDeviceId;
  final String? leaseUserId;
  final String? leaseExpiresAt;
  final int leaseRemainingSeconds;
}

enum CheckoutRecoveryStatus { missing, inProgress, completed }

class CheckoutRecovery {
  const CheckoutRecovery({required this.status, this.transaction});

  factory CheckoutRecovery.fromJson(Object? value) {
    final json = _asMap(value);
    return CheckoutRecovery(
      status: switch (json['status']) {
        'completed' => CheckoutRecoveryStatus.completed,
        'in_progress' => CheckoutRecoveryStatus.inProgress,
        _ => CheckoutRecoveryStatus.missing,
      },
      transaction: json['transaction'] == null
          ? null
          : CheckoutTransaction.fromJson(json['transaction']),
    );
  }

  final CheckoutRecoveryStatus status;
  final CheckoutTransaction? transaction;
}

class CheckoutPaymentLine {
  const CheckoutPaymentLine({
    required this.id,
    required this.transactionId,
    required this.method,
    required this.amount,
    required this.status,
    required this.isCash,
    this.reference,
    this.paymentRef,
    this.confirmExpiresAt,
    this.confirmedBy,
    this.confirmedAt,
    this.transactionStatus,
  });

  factory CheckoutPaymentLine.fromJson(Object? value) {
    final json = _asMap(value);
    final transaction = _asMap(json['transaction']);
    return CheckoutPaymentLine(
      id: (json['id'] as String?) ?? '',
      transactionId: (json['transaction_id'] as String?) ?? '',
      method: (json['method'] as String?) ?? '',
      amount: (json['amount'] as num?)?.toInt() ?? 0,
      status: (json['status'] as String?) ?? '',
      isCash: json['is_cash'] as bool? ?? false,
      reference: json['reference'] as String?,
      paymentRef: json['payment_ref'] as String?,
      confirmExpiresAt: json['confirm_expires_at'] as String?,
      confirmedBy: json['confirmed_by'] as String?,
      confirmedAt: json['confirmed_at'] as String?,
      transactionStatus:
          (json['transaction_status'] as String?) ??
          (transaction['status'] as String?),
    );
  }

  final String id;
  final String transactionId;
  final String method;
  final int amount;
  final String status;
  final bool isCash;
  final String? reference;
  final String? paymentRef;
  final String? confirmExpiresAt;
  final String? confirmedBy;
  final String? confirmedAt;
  final String? transactionStatus;
}

class VoidResult {
  const VoidResult({
    required this.transactionId,
    required this.status,
    required this.reason,
  });

  factory VoidResult.fromJson(Object? value) {
    final json = _asMap(value);
    return VoidResult(
      transactionId: (json['transaction_id'] as String?) ?? '',
      status: (json['status'] as String?) ?? '',
      reason: (json['reason'] as String?) ?? '',
    );
  }

  final String transactionId;
  final String status;
  final String reason;
}

class ApiTransactionRepository implements TransactionRepository {
  const ApiTransactionRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<CheckoutQuote> quoteTransaction(CheckoutDraft draft) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/transactions/quote',
      data: draft.toApiJson(),
    );
    return CheckoutQuote.fromJson(response.data);
  }

  @override
  Future<CheckoutTransaction> createTransaction(CheckoutDraft draft) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/transactions',
      data: draft.toApiJson(),
      idempotencyKey: draft.idempotencyKey,
    );
    return CheckoutTransaction.fromJson(response.data);
  }

  @override
  Future<CheckoutRecovery> recoverCheckout(String idempotencyKey) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/transactions/recovery/$idempotencyKey',
    );
    return CheckoutRecovery.fromJson(response.data);
  }

  @override
  Future<CheckoutPaymentLine> confirmPayment({
    required String paymentId,
    required String idempotencyKey,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/payments',
      data: {'payment_id': paymentId, 'confirm': true},
      idempotencyKey: idempotencyKey,
    );
    return CheckoutPaymentLine.fromJson(response.data);
  }

  @override
  Future<List<CheckoutTransaction>> listTransactions({String? status}) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/transactions',
      queryParameters: {
        if (status != null && status.isNotEmpty) 'status': status,
      },
    );
    final rows = response.data['transactions'] as List? ?? const [];
    return [for (final row in rows) CheckoutTransaction.fromJson(row)];
  }

  @override
  Future<List<CheckoutTransaction>> listParkedOrders() async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/parked-orders',
    );
    final rows = response.data['parked_orders'] as List? ?? const [];
    return [for (final row in rows) CheckoutTransaction.fromJson(row)];
  }

  @override
  Future<CheckoutTransaction> createParkedOrder(CheckoutDraft draft) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/parked-orders',
      data: draft.toApiJson(),
    );
    return CheckoutTransaction.fromJson(response.data);
  }

  @override
  Future<CheckoutTransaction> updateParkedOrder({
    required String transactionId,
    required String deviceId,
    required int expectedRevision,
    required List<CheckoutItem> items,
    String? customerId,
    String? servedBy,
    int cartDiscount = 0,
    String? notes,
  }) async {
    final response = await _apiClient.put<Map<String, Object?>>(
      '/parked-orders/$transactionId',
      data: {
        'device_id': deviceId,
        'expected_revision': expectedRevision,
        if (customerId != null && customerId.isNotEmpty)
          'customer_id': customerId,
        if (servedBy != null && servedBy.isNotEmpty) 'served_by': servedBy,
        'cart_discount': cartDiscount,
        if (notes != null && notes.isNotEmpty) 'notes': notes,
        'items': [for (final item in items) item.toJson()],
      },
    );
    return CheckoutTransaction.fromJson(response.data);
  }

  @override
  Future<CheckoutTransaction> acquireParkedOrderLease({
    required String transactionId,
    required String deviceId,
  }) {
    return _parkedOrderLeaseAction(
      transactionId: transactionId,
      deviceId: deviceId,
      action: 'acquire',
    );
  }

  @override
  Future<CheckoutTransaction> refreshParkedOrderLease({
    required String transactionId,
    required String deviceId,
  }) {
    return _parkedOrderLeaseAction(
      transactionId: transactionId,
      deviceId: deviceId,
      action: 'refresh',
    );
  }

  @override
  Future<CheckoutTransaction> releaseParkedOrderLease({
    required String transactionId,
    required String deviceId,
  }) {
    return _parkedOrderLeaseAction(
      transactionId: transactionId,
      deviceId: deviceId,
      action: 'release',
    );
  }

  Future<CheckoutTransaction> _parkedOrderLeaseAction({
    required String transactionId,
    required String deviceId,
    required String action,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/parked-orders/$transactionId/lease/$action',
      data: {'device_id': deviceId},
    );
    return CheckoutTransaction.fromJson(response.data);
  }

  @override
  Future<VoidResult> voidTransaction({
    required String transactionId,
    required String shiftId,
    required String reason,
    required String idempotencyKey,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/voids',
      data: {
        'transaction_id': transactionId,
        'shift_id': shiftId,
        'reason': reason,
      },
      idempotencyKey: idempotencyKey,
    );
    return VoidResult.fromJson(response.data);
  }

  @override
  SalesTransaction createLocalTransaction({
    required SalesOrder order,
    required List<PaymentLine> payments,
    required Employee cashier,
  }) {
    return SalesTransaction(
      id: 'txn-${DateTime.now().millisecondsSinceEpoch}',
      number: order.number,
      order: order,
      payments: payments,
      cashier: cashier,
      createdAt: DateTime.now(),
    );
  }
}

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}

DateTime? _parseDateTime(Object? value) {
  if (value is DateTime) return value;
  if (value is String && value.isNotEmpty) return DateTime.tryParse(value);
  return null;
}
