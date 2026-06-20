import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/app/providers/nojpos_session_provider.dart';
import 'package:nojpos_tablet_ui/core/share/receipt_share_service.dart';
import 'package:nojpos_tablet_ui/features/transactions/pages/success_screen.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  testWidgets('success screen share button should call native share service', (
    tester,
  ) async {
    final shareService = _RecordingShareService();
    tester.view.physicalSize = const Size(1280, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          receiptShareServiceProvider.overrideWith((ref) => shareService),
          nojposSessionProvider.overrideWith(() => _FakeSessionNotifier()),
        ],
        child: const MaterialApp(home: SuccessScreen()),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Bagikan/WhatsApp'));
    await tester.pumpAndSettle();

    expect(shareService.shareCalls, 1);
    expect(shareService.lastText, contains('TRX-001'));
    expect(shareService.lastText, contains('Kopi Susu'));
  });
}

class _RecordingShareService implements ReceiptShareService {
  int shareCalls = 0;
  String? lastText;

  @override
  Future<void> shareReceipt(String receiptText) async {
    shareCalls += 1;
    lastText = receiptText;
  }
}

class _FakeSessionNotifier extends NojposSessionNotifier {
  @override
  NojposSessionState build() {
    return NojposSessionState.initial().copyWith(
      outlet: const Outlet(
        id: 'outlet-id',
        name: 'Outlet Demo',
        isOnline: true,
      ),
      cashier: const Employee(id: 'cashier-id', name: 'Siti', role: 'cashier'),
      status: SessionStatus.ready,
      lastTransaction: _transaction(),
    );
  }
}

SalesTransaction _transaction() {
  return SalesTransaction(
    id: 'transaction-id',
    number: 'TRX-001',
    order: SalesOrder(
      id: 'order-id',
      number: 'ORD-001',
      type: OrderType.pickup,
      status: OrderStatus.paid,
      lines: const [
        OrderLine(
          productId: 'product-id',
          name: 'Kopi Susu',
          quantity: 1,
          unitPrice: 18000,
        ),
      ],
      createdAt: DateTime(2026, 6, 18, 10),
    ),
    payments: [PaymentLine(method: PaymentMethod.cash, amount: 20000)],
    cashier: const Employee(id: 'cashier-id', name: 'Siti', role: 'cashier'),
    createdAt: DateTime(2026, 6, 18, 10),
  );
}
