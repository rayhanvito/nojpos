import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../shared/models/nojpos_models.dart';

final inventoryRepositoryProvider = Provider<InventoryRepository>(
  (ref) => ApiInventoryRepository(apiClient: ref.watch(apiClientProvider)),
);

abstract interface class InventoryRepository {
  Future<InventorySnapshot> getInventory({
    String? outletId,
    String? productId,
    String? type,
    String? search,
  });

  Future<InventoryMovementHistory> getMovements({
    String? outletId,
    String? productId,
    String? type,
  });

  Future<InventoryPurchase> createPurchase({
    required String outletId,
    String? number,
    required String supplierName,
    String? notes,
    required List<PurchaseItemDraft> items,
  });

  Future<InventoryCountResult> createCount({
    required String outletId,
    String? notes,
    DateTime? countedAt,
    required List<InventoryCountLineDraft> lines,
    required String idempotencyKey,
  });

  Future<InventoryWasteResult> createWaste({
    required String outletId,
    required String productId,
    required int quantity,
    required String reason,
    DateTime? occurredAt,
    required String idempotencyKey,
  });

  Future<InventoryTransferList> listTransfers({
    String? outletId,
    String? status,
  });

  Future<InventoryTransferList> listInTransitTransfers({String? outletId});

  Future<InventoryTransferResult> createTransfer({
    required String sourceOutletId,
    required String destinationOutletId,
    String? notes,
    required List<InventoryTransferLineDraft> lines,
    required String idempotencyKey,
  });

  Future<InventoryTransferResult> sendTransfer({
    required String transferId,
    required String idempotencyKey,
  });

  Future<InventoryTransferResult> receiveTransfer({
    required String transferId,
    required String idempotencyKey,
  });

  Future<InventoryTransferResult> cancelTransfer({
    required String transferId,
    required String idempotencyKey,
  });
}

class ApiInventoryRepository implements InventoryRepository {
  const ApiInventoryRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<InventorySnapshot> getInventory({
    String? outletId,
    String? productId,
    String? type,
    String? search,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/inventory',
      queryParameters: {
        if (outletId != null && outletId.isNotEmpty) 'outlet_id': outletId,
        if (productId != null && productId.isNotEmpty) 'product_id': productId,
        if (type != null && type.isNotEmpty) 'type': type,
        if (search != null && search.trim().isNotEmpty) 'search': search.trim(),
      },
    );
    return InventorySnapshot.fromJson(response.data);
  }

  @override
  Future<InventoryMovementHistory> getMovements({
    String? outletId,
    String? productId,
    String? type,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/inventory/movements',
      queryParameters: {
        if (outletId != null && outletId.isNotEmpty) 'outlet_id': outletId,
        if (productId != null && productId.isNotEmpty) 'product_id': productId,
        if (type != null && type.isNotEmpty) 'type': type,
      },
    );
    return InventoryMovementHistory.fromJson(response.data);
  }

  @override
  Future<InventoryPurchase> createPurchase({
    required String outletId,
    String? number,
    required String supplierName,
    String? notes,
    required List<PurchaseItemDraft> items,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/inventory/purchases',
      data: {
        'outlet_id': outletId,
        if (number != null && number.isNotEmpty) 'number': number,
        'supplier_name': supplierName,
        if (notes != null && notes.isNotEmpty) 'notes': notes,
        'items': [for (final item in items) item.toJson()],
      },
    );
    return InventoryPurchase.fromJson(response.data);
  }

  @override
  Future<InventoryCountResult> createCount({
    required String outletId,
    String? notes,
    DateTime? countedAt,
    required List<InventoryCountLineDraft> lines,
    required String idempotencyKey,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/inventory/counts',
      idempotencyKey: idempotencyKey,
      data: {
        'outlet_id': outletId,
        if (notes != null && notes.isNotEmpty) 'notes': notes,
        if (countedAt != null) 'counted_at': countedAt.toIso8601String(),
        'lines': [for (final line in lines) line.toJson()],
      },
    );
    return InventoryCountResult.fromJson(response.data);
  }

  @override
  Future<InventoryWasteResult> createWaste({
    required String outletId,
    required String productId,
    required int quantity,
    required String reason,
    DateTime? occurredAt,
    required String idempotencyKey,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/inventory/waste',
      idempotencyKey: idempotencyKey,
      data: {
        'outlet_id': outletId,
        'product_id': productId,
        'quantity': quantity,
        'reason': reason,
        if (occurredAt != null) 'occurred_at': occurredAt.toIso8601String(),
      },
    );
    return InventoryWasteResult.fromJson(response.data);
  }

  @override
  Future<InventoryTransferList> listTransfers({
    String? outletId,
    String? status,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/inventory/transfers',
      queryParameters: {
        if (outletId != null && outletId.isNotEmpty) 'outlet_id': outletId,
        if (status != null && status.isNotEmpty) 'status': status,
      },
    );
    return InventoryTransferList.fromJson(response.data);
  }

  @override
  Future<InventoryTransferList> listInTransitTransfers({
    String? outletId,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/inventory/transfers/in-transit',
      queryParameters: {
        if (outletId != null && outletId.isNotEmpty) 'outlet_id': outletId,
      },
    );
    return InventoryTransferList.fromJson(response.data);
  }

  @override
  Future<InventoryTransferResult> createTransfer({
    required String sourceOutletId,
    required String destinationOutletId,
    String? notes,
    required List<InventoryTransferLineDraft> lines,
    required String idempotencyKey,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/inventory/transfers',
      idempotencyKey: idempotencyKey,
      data: {
        'source_outlet_id': sourceOutletId,
        'destination_outlet_id': destinationOutletId,
        if (notes != null && notes.isNotEmpty) 'notes': notes,
        'lines': [for (final line in lines) line.toJson()],
      },
    );
    return InventoryTransferResult.fromJson(response.data);
  }

  @override
  Future<InventoryTransferResult> sendTransfer({
    required String transferId,
    required String idempotencyKey,
  }) {
    return _transferAction('send', transferId, idempotencyKey);
  }

  @override
  Future<InventoryTransferResult> receiveTransfer({
    required String transferId,
    required String idempotencyKey,
  }) {
    return _transferAction('receive', transferId, idempotencyKey);
  }

  @override
  Future<InventoryTransferResult> cancelTransfer({
    required String transferId,
    required String idempotencyKey,
  }) {
    return _transferAction('cancel', transferId, idempotencyKey);
  }

  Future<InventoryTransferResult> _transferAction(
    String action,
    String transferId,
    String idempotencyKey,
  ) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/inventory/transfers/$transferId/$action',
      idempotencyKey: idempotencyKey,
      data: const <String, Object?>{},
    );
    return InventoryTransferResult.fromJson(response.data);
  }
}

class InventorySnapshot {
  const InventorySnapshot({required this.items, required this.movements});

  factory InventorySnapshot.fromJson(Map<String, Object?> json) {
    return InventorySnapshot(
      items: [
        for (final item in json['items'] as List? ?? const [])
          InventoryStockItem.fromJson(item),
      ],
      movements: [
        for (final item in json['movements'] as List? ?? const [])
          StockMovement.fromJson(item),
      ],
    );
  }

  final List<InventoryStockItem> items;
  final List<StockMovement> movements;
}

class InventoryMovementHistory {
  const InventoryMovementHistory({required this.items});

  factory InventoryMovementHistory.fromJson(Object? value) {
    final json = _asMap(value);
    return InventoryMovementHistory(
      items: [
        for (final item
            in json['items'] as List? ?? json['movements'] as List? ?? const [])
          StockMovement.fromJson(item),
      ],
    );
  }

  final List<StockMovement> items;
}

class InventoryStockItem {
  const InventoryStockItem({
    required this.productId,
    required this.name,
    required this.price,
    required this.stockOnHand,
    required this.availableQuantity,
    this.sku,
    this.categoryName,
    this.outletId,
    this.outletName,
    this.inTransitOutQuantity = 0,
    this.inTransitInQuantity = 0,
    this.warning,
  });

  factory InventoryStockItem.fromJson(Object? value) {
    final json = _asMap(value);
    final onHand = _int(json['on_hand_quantity'] ?? json['stock_on_hand']);
    return InventoryStockItem(
      productId: (json['product_id'] as String?) ?? '',
      name:
          (json['product_name'] ?? json['name'] as String?)?.toString() ??
          'Produk',
      price: _int(json['price']),
      stockOnHand: onHand,
      availableQuantity: _int(json['available_quantity'] ?? onHand),
      sku: json['sku'] as String?,
      categoryName: json['category_name'] as String?,
      outletId: json['outlet_id'] as String?,
      outletName: json['outlet_name'] as String?,
      inTransitOutQuantity: _int(json['in_transit_out_quantity']),
      inTransitInQuantity: _int(json['in_transit_in_quantity']),
      warning: json['warning'] as String?,
    );
  }

  final String productId;
  final String name;
  final int price;
  final int stockOnHand;
  final int availableQuantity;
  final String? sku;
  final String? categoryName;
  final String? outletId;
  final String? outletName;
  final int inTransitOutQuantity;
  final int inTransitInQuantity;
  final String? warning;

  bool get isNegative => stockOnHand < 0;
}

class StockMovement {
  const StockMovement({
    required this.id,
    required this.productId,
    required this.type,
    required this.quantityDelta,
    this.productName,
    this.outletId,
    this.outletName,
    this.beforeQuantity,
    this.afterQuantity,
    this.referenceType,
    this.referenceId,
    this.reason,
    this.createdAt,
  });

  factory StockMovement.fromJson(Object? value) {
    final json = _asMap(value);
    return StockMovement(
      id: (json['movement_id'] ?? json['id'] as String?)?.toString() ?? '',
      productId: (json['product_id'] as String?) ?? '',
      productName: json['product_name'] as String?,
      outletId: json['outlet_id'] as String?,
      outletName: json['outlet_name'] as String?,
      type: (json['type'] as String?) ?? '',
      quantityDelta: _int(json['quantity_delta']),
      beforeQuantity: _nullableInt(json['before_quantity']),
      afterQuantity: _nullableInt(json['after_quantity']),
      referenceType: json['reference_type'] as String?,
      referenceId: json['reference_id'] as String?,
      reason: json['reason'] as String?,
      createdAt: DateTime.tryParse((json['created_at'] ?? '').toString()),
    );
  }

  final String id;
  final String productId;
  final String? productName;
  final String? outletId;
  final String? outletName;
  final String type;
  final int quantityDelta;
  final int? beforeQuantity;
  final int? afterQuantity;
  final String? referenceType;
  final String? referenceId;
  final String? reason;
  final DateTime? createdAt;
}

class PurchaseItemDraft {
  const PurchaseItemDraft({
    required this.productId,
    required this.quantity,
    required this.unitCost,
  });

  final String productId;
  final int quantity;
  final int unitCost;

  Map<String, Object?> toJson() => {
    'product_id': productId,
    'quantity': quantity,
    'unit_cost': unitCost,
  };
}

class InventoryCountLineDraft {
  const InventoryCountLineDraft({
    required this.productId,
    required this.countedQuantity,
    this.reason,
  });

  final String productId;
  final int countedQuantity;
  final String? reason;

  Map<String, Object?> toJson() => {
    'product_id': productId,
    'counted_quantity': countedQuantity,
    if (reason != null && reason!.isNotEmpty) 'reason': reason,
  };
}

class InventoryCountResult {
  const InventoryCountResult({
    required this.id,
    required this.number,
    required this.status,
    required this.lines,
  });

  factory InventoryCountResult.fromJson(Object? value) {
    final json = _asMap(value);
    return InventoryCountResult(
      id: (json['id'] as String?) ?? '',
      number: (json['number'] as String?) ?? '',
      status: (json['status'] as String?) ?? '',
      lines: [
        for (final line in json['lines'] as List? ?? const [])
          InventoryCountLineResult.fromJson(line),
      ],
    );
  }

  final String id;
  final String number;
  final String status;
  final List<InventoryCountLineResult> lines;
}

class InventoryCountLineResult {
  const InventoryCountLineResult({
    required this.productId,
    required this.systemQuantity,
    required this.countedQuantity,
    required this.deltaQuantity,
    this.reason,
  });

  factory InventoryCountLineResult.fromJson(Object? value) {
    final json = _asMap(value);
    return InventoryCountLineResult(
      productId: (json['product_id'] as String?) ?? '',
      systemQuantity: _int(json['system_quantity']),
      countedQuantity: _int(json['counted_quantity']),
      deltaQuantity: _int(json['delta_quantity']),
      reason: json['reason'] as String?,
    );
  }

  final String productId;
  final int systemQuantity;
  final int countedQuantity;
  final int deltaQuantity;
  final String? reason;
}

class InventoryWasteResult {
  const InventoryWasteResult({
    required this.id,
    required this.productId,
    required this.quantity,
    required this.reason,
    required this.status,
  });

  factory InventoryWasteResult.fromJson(Object? value) {
    final json = _asMap(value);
    return InventoryWasteResult(
      id: (json['id'] as String?) ?? '',
      productId: (json['product_id'] as String?) ?? '',
      quantity: _int(json['quantity']),
      reason: (json['reason'] as String?) ?? '',
      status: (json['status'] as String?) ?? '',
    );
  }

  final String id;
  final String productId;
  final int quantity;
  final String reason;
  final String status;
}

class InventoryTransferLineDraft {
  const InventoryTransferLineDraft({
    required this.productId,
    required this.quantity,
  });

  final String productId;
  final int quantity;

  Map<String, Object?> toJson() => {
    'product_id': productId,
    'quantity': quantity,
  };
}

class InventoryTransferList {
  const InventoryTransferList({required this.items});

  factory InventoryTransferList.fromJson(Object? value) {
    final json = _asMap(value);
    return InventoryTransferList(
      items: [
        for (final item in json['items'] as List? ?? const [])
          InventoryTransferResult.fromJson(item),
      ],
    );
  }

  final List<InventoryTransferResult> items;
}

class InventoryTransferResult {
  const InventoryTransferResult({
    required this.id,
    required this.number,
    required this.status,
    required this.sourceOutletName,
    required this.destinationOutletName,
    required this.lines,
    this.sentAt,
    this.receivedAt,
    this.ageSeconds = 0,
  });

  factory InventoryTransferResult.fromJson(Object? value) {
    final json = _asMap(value);
    final source = _asMap(json['source_outlet']);
    final destination = _asMap(json['destination_outlet']);
    return InventoryTransferResult(
      id: (json['id'] ?? json['transfer_id'] as String?)?.toString() ?? '',
      number: (json['number'] as String?) ?? '',
      status: (json['status'] as String?) ?? '',
      sourceOutletName: (source['name'] as String?) ?? '',
      destinationOutletName: (destination['name'] as String?) ?? '',
      sentAt: DateTime.tryParse((json['sent_at'] ?? '').toString()),
      receivedAt: DateTime.tryParse((json['received_at'] ?? '').toString()),
      ageSeconds: _int(json['age_seconds']),
      lines: [
        for (final line in json['lines'] as List? ?? const [])
          InventoryTransferLineResult.fromJson(line),
      ],
    );
  }

  final String id;
  final String number;
  final String status;
  final String sourceOutletName;
  final String destinationOutletName;
  final DateTime? sentAt;
  final DateTime? receivedAt;
  final int ageSeconds;
  final List<InventoryTransferLineResult> lines;

  bool get isInTransit => status == 'in_transit';
}

class InventoryTransferLineResult {
  const InventoryTransferLineResult({
    required this.productId,
    required this.productName,
    required this.requestedQuantity,
    required this.sentQuantity,
    required this.receivedQuantity,
    required this.inTransitQuantity,
  });

  factory InventoryTransferLineResult.fromJson(Object? value) {
    final json = _asMap(value);
    return InventoryTransferLineResult(
      productId: (json['product_id'] as String?) ?? '',
      productName: (json['product_name'] as String?) ?? 'Produk',
      requestedQuantity: _int(json['requested_quantity']),
      sentQuantity: _int(json['sent_quantity']),
      receivedQuantity: _int(json['received_quantity']),
      inTransitQuantity: _int(json['in_transit_quantity']),
    );
  }

  final String productId;
  final String productName;
  final int requestedQuantity;
  final int sentQuantity;
  final int receivedQuantity;
  final int inTransitQuantity;
}

int _int(Object? value) => (value as num?)?.toInt() ?? 0;

int? _nullableInt(Object? value) => value == null ? null : _int(value);

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}
