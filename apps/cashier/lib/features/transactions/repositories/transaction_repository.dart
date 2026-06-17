import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/nojpos_models.dart';

final transactionRepositoryProvider = Provider<TransactionRepository>(
  (ref) => const MockTransactionRepository(),
);

abstract interface class TransactionRepository {
  SalesTransaction createTransaction({
    required SalesOrder order,
    required List<PaymentLine> payments,
    required Employee cashier,
  });
}

class MockTransactionRepository implements TransactionRepository {
  const MockTransactionRepository();

  @override
  SalesTransaction createTransaction({
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
