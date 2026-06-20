import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/core/outbox/checkout_outbox.dart';
import 'package:nojpos_tablet_ui/features/notifications/widgets/checkout_outbox_center.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';

void main() {
  testWidgets('notification center reflects outbox status and retries', (
    tester,
  ) async {
    final item = CheckoutOutboxItem.fromDraft(
      _draft(),
      now: DateTime(2026, 6, 18, 10),
      error: 'Timeout',
    ).copyWith(status: CheckoutOutboxStatus.needsAction, retryCount: 5);
    CheckoutOutboxItem? retried;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: CheckoutOutboxCenter(
            items: [item],
            onRetry: (item) => retried = item,
          ),
        ),
      ),
    );

    expect(find.text('Perlu tindakan kasir'), findsOneWidget);
    expect(find.textContaining('Timeout'), findsOneWidget);
    expect(find.textContaining('Percobaan kirim ulang: 5'), findsOneWidget);

    await tester.tap(find.text('Kirim ulang'));
    expect(retried?.idempotencyKey, 'outbox-key');
  });
}

CheckoutDraft _draft() {
  return const CheckoutDraft(
    idempotencyKey: 'outbox-key',
    outletId: 'outlet-id',
    deviceId: 'device-id',
    cashierId: 'cashier-id',
    shiftId: 'shift-id',
    items: [
      CheckoutItem(
        productId: 'product-id',
        name: 'Kopi',
        quantity: 1,
        unitPrice: 18000,
      ),
    ],
    payments: [CheckoutPayment(method: 'Tunai', amount: 18000, isCash: true)],
    paidAmount: 18000,
  );
}
