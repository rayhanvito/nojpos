import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/features/pos/screens/operations_screen.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  testWidgets('should show pending payments list and confirm deliberately', (
    tester,
  ) async {
    final transactions = _FakeTransactionRepository();
    tester.view.physicalSize = const Size(1280, 800);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          transactionRepositoryProvider.overrideWith((ref) => transactions),
        ],
        child: const MaterialApp(
          home: OperationsScreen(type: OperationsPageType.sales),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Pembayaran Pending'));
    await tester.pumpAndSettle();

    expect(transactions.lastListStatus, 'pending');
    expect(find.text('QR-001'), findsOneWidget);
    expect(find.text('Tandai sudah diterima'), findsOneWidget);

    await tester.tap(find.byKey(const ValueKey('pending_payment_payment-pending')));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Tandai diterima'));
    await tester.pumpAndSettle();

    expect(transactions.lastConfirmedPaymentId, 'payment-pending');
    expect(find.textContaining('Status transaksi: paid'), findsOneWidget);
  });
}

class _FakeTransactionRepository implements TransactionRepository {
  String? lastListStatus;
  String? lastConfirmedPaymentId;
  String confirmedTransactionStatus = 'paid';

  @override
  Future<CheckoutQuote> quoteTransaction(CheckoutDraft draft) async {
    throw UnimplementedError();
  }

  @override
  Future<CheckoutTransaction> createTransaction(CheckoutDraft draft) async {
    throw UnimplementedError();
  }

  @override
  Future<CheckoutRecovery> recoverCheckout(String idempotencyKey) async {
    return const CheckoutRecovery(status: CheckoutRecoveryStatus.missing);
  }

  @override
  Future<CheckoutPaymentLine> confirmPayment({
    required String paymentId,
    required String idempotencyKey,
  }) async {
    lastConfirmedPaymentId = paymentId;
    return CheckoutPaymentLine(
      id: paymentId,
      transactionId: 'transaction-pending',
      method: 'QRIS Statis',
      amount: 60000,
      reference: 'QR-001',
      status: 'confirmed',
      isCash: false,
      confirmedBy: 'cashier-id',
      confirmedAt: '2026-06-18T10:00:00Z',
      transactionStatus: confirmedTransactionStatus,
    );
  }

  @override
  Future<List<CheckoutTransaction>> listTransactions({String? status}) async {
    lastListStatus = status;
    if (status != 'pending') return const [];
    return const [
      CheckoutTransaction(
        id: 'transaction-pending',
        number: 'TRX-PENDING',
        status: 'partial',
        grandTotal: 60000,
        createdAt: null,
        cashierName: 'Kasir Demo',
        payments: [
          CheckoutPaymentLine(
            id: 'payment-pending',
            transactionId: 'transaction-pending',
            method: 'QRIS Statis',
            amount: 60000,
            status: 'pending',
            isCash: false,
            reference: 'QR-001',
          ),
        ],
        items: [
          CheckoutItem(
            productId: 'product-id',
            name: 'Kopi Susu',
            quantity: 1,
            unitPrice: 60000,
          ),
        ],
      ),
    ];
  }

  @override
  Future<VoidResult> voidTransaction({
    required String transactionId,
    required String shiftId,
    required String reason,
    required String idempotencyKey,
  }) async {
    throw UnimplementedError();
  }

  @override
  SalesTransaction createLocalTransaction({
    required SalesOrder order,
    required List<PaymentLine> payments,
    required Employee cashier,
  }) {
    throw UnimplementedError();
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}
