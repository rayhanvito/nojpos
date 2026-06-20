import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/features/payment/widgets/pending_payment_status_view.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';

void main() {
  for (final entry in {
    'qris': 'Menunggu konfirmasi pembayaran QRIS...',
    'edc': 'Menunggu konfirmasi mesin EDC...',
    'transfer': 'Menunggu konfirmasi transfer...',
  }.entries) {
    testWidgets('${entry.key} pending message', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: PendingPaymentStatusView(
            transaction: transaction('payment_pending', entry.key),
            onCancel: () {},
          ),
        ),
      );
      expect(find.text(entry.value), findsOneWidget);
    });
  }
  testWidgets('failed and expired messages are clear', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: PendingPaymentStatusView(
          transaction: transaction('payment_failed', 'qris'),
          onCancel: () {},
        ),
      ),
    );
    expect(
      find.text('Pembayaran gagal. Pilih metode pembayaran lagi.'),
      findsOneWidget,
    );
    await tester.pumpWidget(
      MaterialApp(
        home: PendingPaymentStatusView(
          transaction: transaction('expired', 'qris'),
          onCancel: () {},
        ),
      ),
    );
    expect(
      find.text('Waktu pembayaran habis. Pilih metode pembayaran lagi.'),
      findsOneWidget,
    );
  });
}

CheckoutTransaction transaction(String status, String method) =>
    CheckoutTransaction(
      id: 'id',
      number: 'TRX',
      status: status,
      grandTotal: 10000,
      payments: [
        CheckoutPaymentLine(
          id: 'payment',
          transactionId: 'id',
          method: method,
          amount: 10000,
          status: 'pending',
          isCash: false,
          confirmExpiresAt: '2026-06-20T10:00:00Z',
        ),
      ],
    );
