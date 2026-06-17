import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/nojpos_models.dart';

final inventoryRepositoryProvider = Provider<InventoryRepository>(
  (ref) => const MockInventoryRepository(),
);

abstract interface class InventoryRepository {
  InventoryPurchase createPurchase({
    required int sequence,
    required String supplierName,
    required int total,
  });
}

class MockInventoryRepository implements InventoryRepository {
  const MockInventoryRepository();

  @override
  InventoryPurchase createPurchase({
    required int sequence,
    required String supplierName,
    required int total,
  }) {
    final createdAt = DateTime.now();
    return InventoryPurchase(
      id: 'purchase-${createdAt.microsecondsSinceEpoch}',
      number:
          'PO/${createdAt.year}${createdAt.month.toString().padLeft(2, '0')}/${sequence.toString().padLeft(4, '0')}',
      supplierName: supplierName,
      total: total,
      createdAt: createdAt,
    );
  }
}
