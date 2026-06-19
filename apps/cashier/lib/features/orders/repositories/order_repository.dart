import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/nojpos_models.dart';
import '../../pos/models/cart_item.dart';

final orderRepositoryProvider = Provider<OrderRepository>(
  (ref) => const LocalOrderRepository(),
);

abstract interface class OrderRepository {
  SalesOrder createOrder({
    required String outletId,
    required List<CartItem> cartItems,
    required OrderType type,
    required OrderStatus status,
    required int sequence,
    Customer? customer,
  });
}

class LocalOrderRepository implements OrderRepository {
  const LocalOrderRepository();

  @override
  SalesOrder createOrder({
    required String outletId,
    required List<CartItem> cartItems,
    required OrderType type,
    required OrderStatus status,
    required int sequence,
    Customer? customer,
  }) {
    final createdAt = DateTime.now();
    // Local order numbers are only for cart/held display; paid receipts use
    // the server transaction number applied in NojposSessionNotifier.
    return SalesOrder(
      id: 'order-${createdAt.millisecondsSinceEpoch}',
      number: _numberFor(createdAt, sequence, outletId),
      type: type,
      customer: customer,
      lines: [
        for (final item in cartItems)
          OrderLine(
            productId: item.product.id,
            name: item.product.name,
            quantity: item.quantity,
            unitPrice: item.product.price,
            discount: item.discount,
          ),
      ],
      status: status,
      createdAt: createdAt,
    );
  }

  String _numberFor(DateTime date, int sequence, String outletId) {
    final yy = (date.year % 100).toString().padLeft(2, '0');
    final mm = date.month.toString().padLeft(2, '0');
    final dd = date.day.toString().padLeft(2, '0');
    final outletNumber = outletId.replaceAll('outlet-', '');
    return 'CS/$outletNumber/$yy$mm$dd/${sequence.toString().padLeft(4, '0')}';
  }
}
