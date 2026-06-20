import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';

void main() {
  test('parses payment pending checkout and terminal polling states', () {
    final pending = CheckoutTransaction.fromJson({
      'id': 'transaction-id',
      'number': 'TRX',
      'status': 'payment_pending',
      'grand_total': 10000,
      'server_time': '2026-06-20T09:00:00Z',
      'payments': [
        {
          'id': 'payment-id',
          'transaction_id': 'transaction-id',
          'method': 'qris',
          'amount': 10000,
          'status': 'pending',
          'is_cash': false,
          'payment_ref': 'opaque-test-ref',
          'confirm_expires_at': '2026-06-20T09:10:00Z',
        },
      ],
    });
    expect(pending.status, 'payment_pending');
    expect(pending.payments.single.id, 'payment-id');
    expect(pending.payments.single.status, 'pending');
    expect(pending.payments.single.confirmExpiresAt, '2026-06-20T09:10:00Z');
    expect(
      CheckoutTransaction.fromJson({
        'id': 'transaction-id',
        'number': 'TRX',
        'status': 'paid',
        'grand_total': 10000,
      }).status,
      'paid',
    );
    expect(
      CheckoutTransaction.fromJson({
        'id': 'transaction-id',
        'number': 'TRX',
        'status': 'payment_failed',
        'grand_total': 10000,
      }).status,
      'payment_failed',
    );
  });
}
