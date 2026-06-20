import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/features/payment/providers/pending_payment_controller.dart';
import 'package:nojpos_tablet_ui/features/payment/repositories/pending_payment_repository.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';

void main() {
  test('polling stops after the server reports paid', () async {
    final repository = _FakePendingPaymentRepository([
      'payment_pending',
      'paid',
    ]);
    final container = ProviderContainer(
      overrides: [
        pendingPaymentRepositoryProvider.overrideWithValue(repository),
        pendingPaymentPollIntervalProvider.overrideWithValue(
          const Duration(milliseconds: 5),
        ),
      ],
    );
    addTearDown(container.dispose);

    container
        .read(pendingPaymentControllerProvider.notifier)
        .start(_transaction('payment_pending'));
    await Future<void>.delayed(const Duration(milliseconds: 30));

    expect(
      container.read(pendingPaymentControllerProvider).transaction?.status,
      'paid',
    );
    expect(container.read(pendingPaymentControllerProvider).active, isFalse);
    final callsAfterPaid = repository.calls;
    await Future<void>.delayed(const Duration(milliseconds: 20));
    expect(repository.calls, callsAfterPaid);
  });

  test('polling error retains server pending status', () async {
    final repository = _FakePendingPaymentRepository([], throwOnRead: true);
    final container = ProviderContainer(
      overrides: [
        pendingPaymentRepositoryProvider.overrideWithValue(repository),
        pendingPaymentPollIntervalProvider.overrideWithValue(
          const Duration(seconds: 1),
        ),
      ],
    );
    addTearDown(container.dispose);

    container
        .read(pendingPaymentControllerProvider.notifier)
        .start(_transaction('payment_pending'));
    await Future<void>.delayed(Duration.zero);

    final state = container.read(pendingPaymentControllerProvider);
    expect(state.transaction?.status, 'payment_pending');
    expect(state.errorMessage, isNotNull);
  });
}

CheckoutTransaction _transaction(String status) => CheckoutTransaction(
  id: 'transaction-id',
  number: 'TRX-001',
  status: status,
  grandTotal: 10000,
);

class _FakePendingPaymentRepository implements PendingPaymentRepository {
  _FakePendingPaymentRepository(this._statuses, {this.throwOnRead = false});

  final List<String> _statuses;
  final bool throwOnRead;
  int calls = 0;

  @override
  Future<CheckoutTransaction> getTransaction(String transactionId) async {
    calls++;
    if (throwOnRead) throw StateError('network');
    final status = _statuses.removeAt(0);
    return _transaction(status);
  }
}
