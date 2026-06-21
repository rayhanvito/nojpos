import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/app/providers/nojpos_session_provider.dart';
import 'package:nojpos_tablet_ui/core/outbox/checkout_outbox.dart';
import 'package:nojpos_tablet_ui/features/auth/repositories/auth_repository.dart';
import 'package:nojpos_tablet_ui/features/pos/widgets/pos_dialogs.dart';
import 'package:nojpos_tablet_ui/features/shift/repositories/shift_repository.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  testWidgets(
    'close shift dialog shows server expected cash and payment totals',
    (tester) async {
      final auth = _FakeAuthRepository();
      final shift = _FakeShiftRepository();
      final container = ProviderContainer(
        overrides: [
          authRepositoryProvider.overrideWith((ref) => auth),
          shiftRepositoryProvider.overrideWith((ref) => shift),
          checkoutOutboxStoreProvider.overrideWith(
            (ref) => InMemoryCheckoutOutboxStore(),
          ),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(nojposSessionProvider.notifier);
      await notifier.login(
        email: 'owner@test.local',
        password: 'password',
        deviceUuid: 'device-uuid',
      );
      await notifier.pinSwitch('1234');
      await notifier.refreshCurrentShift();

      tester.view.physicalSize = const Size(1280, 900);
      tester.view.devicePixelRatio = 1;
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
      });

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: MaterialApp(
            home: Consumer(
              builder: (context, ref, child) => Scaffold(
                body: Center(
                  child: TextButton(
                    onPressed: () => showCloseShiftDialog(context, ref),
                    child: const Text('Open close shift'),
                  ),
                ),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('Open close shift'));
      await tester.pumpAndSettle();

      expect(find.text('Expected cash'), findsOneWidget);
      expect(find.text('Rp 125.000'), findsOneWidget);
      expect(find.text('Payment totals'), findsOneWidget);
      expect(find.text('Tunai (cash)'), findsOneWidget);
      expect(find.text('QRIS Statis'), findsOneWidget);

      await tester.enterText(
        find.byKey(const ValueKey('close_shift_actual_cash')),
        '120000',
      );
      await tester.pump();
      await tester.tap(find.text('Tutup Kasir'));
      await tester.pumpAndSettle();

      expect(shift.lastActualCash, 120000);
      expect(find.text('Hasil final dari server'), findsOneWidget);
      expect(find.text('-Rp 5.000'), findsOneWidget);
    },
  );
}

class _FakeAuthRepository implements AuthRepository {
  @override
  Future<AuthSession?> restoreSession() async => null;

  @override
  Future<AuthSession> login({
    required String email,
    required String password,
    required String deviceUuid,
  }) async {
    return const AuthSession(
      user: Employee(id: 'owner-id', name: 'Owner', role: 'owner'),
      businessName: 'NojPOS Demo',
      outlets: [Outlet(id: 'outlet-id', name: 'Outlet Demo', isOnline: true)],
      deviceId: 'device-id',
      deviceUuid: 'device-uuid',
    );
  }

  @override
  Future<AuthSession> me() => login(
    email: 'owner@test.local',
    password: 'password',
    deviceUuid: 'device-uuid',
  );

  @override
  Future<List<Outlet>> listOutlets() async => const [
    Outlet(id: 'outlet-id', name: 'Outlet Demo', isOnline: true),
  ];

  @override
  Future<Employee> pinSwitch({
    required String pin,
    required String deviceId,
    required String outletId,
  }) async {
    return const Employee(id: 'cashier-id', name: 'Cashier', role: 'cashier');
  }

  @override
  Future<void> selectOutlet(Outlet outlet) async {}

  @override
  Future<void> logout() async {}
}

class _FakeShiftRepository implements ShiftRepository {
  int? lastActualCash;

  @override
  Future<ShiftSession> openShift({
    required String outletId,
    required String deviceId,
    required String cashierId,
    required int openingCash,
  }) async {
    return _shift(status: 'open', openingCash: openingCash);
  }

  @override
  Future<ShiftSession?> currentShift({
    required String outletId,
    required String deviceId,
  }) async {
    return _shift(status: 'open');
  }

  @override
  Future<ShiftSession> closeShift({
    required String shiftId,
    required int actualCash,
    required String pin,
    String? varianceReason,
  }) async {
    lastActualCash = actualCash;
    return _shift(status: 'closed', actualCash: actualCash, difference: -5000);
  }

  @override
  Future<void> cashMovement({
    required String shiftId,
    required String type,
    required int amount,
    String? reason,
  }) async {}

  ShiftSession _shift({
    required String status,
    int openingCash = 100000,
    int? actualCash,
    int? difference,
  }) {
    return ShiftSession(
      id: 'shift-id',
      cashier: const Employee(
        id: 'cashier-id',
        name: 'Cashier',
        role: 'cashier',
      ),
      openedAt: DateTime(2026, 6, 18, 8),
      closedAt: status == 'closed' ? DateTime(2026, 6, 18, 20) : null,
      openingCash: openingCash,
      expectedCash: 125000,
      actualCash: actualCash,
      cashDifference: difference,
      status: status,
      paymentTotals: const [
        ShiftPaymentTotal(method: 'Tunai', amount: 25000, isCash: true),
        ShiftPaymentTotal(method: 'QRIS Statis', amount: 60000),
      ],
    );
  }
}
