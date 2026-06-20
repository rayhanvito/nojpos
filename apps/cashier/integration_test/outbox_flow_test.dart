import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:nojpos_tablet_ui/app/providers/nojpos_session_provider.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/core/outbox/checkout_outbox.dart';
import 'package:nojpos_tablet_ui/features/auth/repositories/auth_repository.dart';
import 'package:nojpos_tablet_ui/features/shift/repositories/shift_repository.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('Checkout outbox runtime hardening', () {
    testWidgets('should prove checkout-only outbox hard scenarios', (
      tester,
    ) async {
      final report = _RuntimeReport();
      const storage = FlutterSecureStorage();
      await storage.deleteAll();

      final repo = _RuntimeTransactionRepository();
      final slept = <Duration>[];

      ProviderContainer container() {
        return ProviderContainer(
          overrides: [
            transactionRepositoryProvider.overrideWith((ref) => repo),
            checkoutOutboxSleeperProvider.overrideWith(
              (ref) =>
                  (duration) async => slept.add(duration),
            ),
          ],
        );
      }

      var current = container();
      addTearDown(current.dispose);

      await report.step('1 API mati/timeout -> pending outbox', () async {
        repo.mode = _RuntimeMode.networkDown;
        final result = await current
            .read(checkoutOutboxControllerProvider.notifier)
            .submitOrQueue(_draft('runtime-key-1'));

        expect(result, isA<QueuedCheckoutResult>());
        final items = await current.read(checkoutOutboxStoreProvider).readAll();
        expect(items.single.status, CheckoutOutboxStatus.pending);
        expect(items.single.draft.quote?.serviceChargeTotal, 3300);
        expect(items.single.draft.quote?.taxTotal, 3993);
        expect(items.single.draft.quote?.grandTotal, 40293);
      });

      await report.step('5 app kill/reopen -> outbox persisted', () async {
        current.dispose();
        current = container();
        final items = await current.read(checkoutOutboxStoreProvider).readAll();
        expect(items.single.idempotencyKey, 'runtime-key-1');
        expect(items.single.status, CheckoutOutboxStatus.pending);
      });

      await report.step('2 reconnect -> auto retry -> sent paid', () async {
        repo.mode = _RuntimeMode.success;
        await current
            .read(checkoutOutboxControllerProvider.notifier)
            .retryPending();

        final items = await current.read(checkoutOutboxStoreProvider).readAll();
        expect(items.single.status, CheckoutOutboxStatus.sent);
        expect(repo.createdTransactions.single.status, 'paid');
      });

      await report.step('4 retry memakai key sama -> satu transaksi', () async {
        expect(
          repo.seenKeys.where((key) => key == 'runtime-key-1'),
          hasLength(2),
        );
        expect(repo.createdTransactions, hasLength(1));
      });

      await report.step('3 foreground resume -> auto retry jalan', () async {
        repo.mode = _RuntimeMode.networkDown;
        await current
            .read(checkoutOutboxControllerProvider.notifier)
            .submitOrQueue(_draft('runtime-key-resume'));
        repo.mode = _RuntimeMode.success;

        current
            .read(checkoutOutboxControllerProvider.notifier)
            .didChangeAppLifecycleState(AppLifecycleState.resumed);
        await tester.pump(const Duration(milliseconds: 100));

        final items = await current.read(checkoutOutboxStoreProvider).readAll();
        final resumed = items.singleWhere(
          (item) => item.idempotencyKey == 'runtime-key-resume',
        );
        expect(resumed.status, CheckoutOutboxStatus.sent);
      });

      await report.step('6 backoff 5s -> 30s -> 2m cap', () async {
        slept.clear();
        repo.mode = _RuntimeMode.networkDown;
        final store = current.read(checkoutOutboxStoreProvider);
        await store.save([
          CheckoutOutboxItem.fromDraft(
            _draft('runtime-backoff-0'),
            now: DateTime(2026, 6, 19, 8),
          ),
          CheckoutOutboxItem.fromDraft(
            _draft('runtime-backoff-1'),
            now: DateTime(2026, 6, 19, 8),
          ).copyWith(retryCount: 1),
          CheckoutOutboxItem.fromDraft(
            _draft('runtime-backoff-2'),
            now: DateTime(2026, 6, 19, 8),
          ).copyWith(retryCount: 2),
        ]);

        await current
            .read(checkoutOutboxControllerProvider.notifier)
            .retryPending();

        expect(slept, [
          const Duration(seconds: 5),
          const Duration(seconds: 30),
          const Duration(minutes: 2),
        ]);
      });

      await report.step('6 needs-action setelah 5x gagal', () async {
        final store = current.read(checkoutOutboxStoreProvider);
        await store.save([
          CheckoutOutboxItem.fromDraft(
            _draft('runtime-needs-action'),
            now: DateTime.now(),
          ).copyWith(retryCount: 4),
        ]);

        await current
            .read(checkoutOutboxProvider)
            .retryNow('runtime-needs-action');

        final item = (await store.readAll()).single;
        expect(item.retryCount, 5);
        expect(item.status, CheckoutOutboxStatus.needsAction);
      });

      await report.step('7 close shift ditolak dengan jumlah outbox', () async {
        final store = current.read(checkoutOutboxStoreProvider);
        await store.save([
          CheckoutOutboxItem.fromDraft(
            _draft('runtime-close-block'),
            now: DateTime.now(),
          ),
        ]);

        final auth = _RuntimeAuthRepository();
        final shift = _RuntimeShiftRepository();
        final uiContainer = ProviderContainer(
          overrides: [
            authRepositoryProvider.overrideWith((ref) => auth),
            shiftRepositoryProvider.overrideWith((ref) => shift),
            checkoutOutboxStoreProvider.overrideWith((ref) => store),
          ],
        );
        addTearDown(uiContainer.dispose);

        final notifier = uiContainer.read(nojposSessionProvider.notifier);
        await notifier.login(
          email: 'owner@test.local',
          password: 'password',
          deviceUuid: 'device-uuid',
        );
        await notifier.pinSwitch('1234');
        await notifier.openShift(openingCash: 100000);
        final closed = await notifier.closeShift(
          actualCash: 100000,
          pin: '1234',
        );

        await tester.pumpWidget(
          UncontrolledProviderScope(
            container: uiContainer,
            child: MaterialApp(
              home: Consumer(
                builder: (context, ref, child) {
                  final message =
                      ref.watch(nojposSessionProvider).errorMessage ?? '';
                  return Text(message, textDirection: TextDirection.ltr);
                },
              ),
            ),
          ),
        );

        expect(closed, false);
        expect(shift.closeCalls, 0);
        expect(find.textContaining('1 checkout tertunda'), findsOneWidget);
      });

      await report.step('8 stale quote retry -> needs-action', () async {
        repo.mode = _RuntimeMode.totalMismatch;
        final staleStore = current.read(checkoutOutboxStoreProvider);
        await staleStore.save([
          CheckoutOutboxItem.fromDraft(
            _draft('runtime-stale-key'),
            now: DateTime.now(),
          ),
        ]);

        await current
            .read(checkoutOutboxProvider)
            .retryNow('runtime-stale-key');

        final item = (await staleStore.readAll()).single;
        expect(item.status, CheckoutOutboxStatus.needsAction);
        expect(item.lastError, contains('server quote'));
        expect(item.draft.idempotencyKey, 'runtime-stale-key');
      });

      await report.step(
        '9 stale quote direct submit -> clear error no queue',
        () async {
          repo.mode = _RuntimeMode.totalMismatch;
          final staleStore = current.read(checkoutOutboxStoreProvider);
          await staleStore.save(const []);

          await expectLater(
            current
                .read(checkoutOutboxControllerProvider.notifier)
                .submitOrQueue(_draft('runtime-direct-stale')),
            throwsA(
              isA<ValidationApiException>().having(
                (error) => error.code,
                'code',
                'TOTAL_MISMATCH',
              ),
            ),
          );
          expect(await staleStore.readAll(), isEmpty);
        },
      );

      report.printSummary();
      expect(report.failures, isEmpty, reason: report.failures.join('\n'));
    });
  });
}

CheckoutDraft _draft(String idempotencyKey) {
  return CheckoutDraft(
    idempotencyKey: idempotencyKey,
    outletId: 'outlet-id',
    deviceId: 'device-id',
    cashierId: 'cashier-id',
    shiftId: 'shift-id',
    items: const [
      CheckoutItem(
        productId: 'product-id',
        name: 'Kopi',
        quantity: 2,
        unitPrice: 18000,
        discount: 2000,
      ),
    ],
    payments: const [
      CheckoutPayment(method: 'Tunai', amount: 40293, isCash: true),
    ],
    paidAmount: 40293,
    cartDiscount: 1000,
    quote: const CheckoutQuote(
      subtotal: 36000,
      itemDiscountTotal: 2000,
      cartDiscountTotal: 1000,
      discountTotal: 3000,
      serviceChargeTotal: 3300,
      taxTotal: 3993,
      roundingTotal: 0,
      grandTotal: 40293,
    ),
  );
}

enum _RuntimeMode { networkDown, success, totalMismatch }

class _RuntimeTransactionRepository implements TransactionRepository {
  _RuntimeMode mode = _RuntimeMode.success;
  final seenKeys = <String>[];
  final createdTransactions = <CheckoutTransaction>[];

  @override
  Future<CheckoutQuote> quoteTransaction(CheckoutDraft draft) async {
    return draft.quote!;
  }

  @override
  Future<CheckoutTransaction> createTransaction(CheckoutDraft draft) async {
    seenKeys.add(draft.idempotencyKey);
    switch (mode) {
      case _RuntimeMode.networkDown:
        throw const ServerApiException(
          code: 'NETWORK_ERROR',
          message: 'Connection timed out.',
        );
      case _RuntimeMode.totalMismatch:
        throw const ValidationApiException(
          code: 'TOTAL_MISMATCH',
          message: 'Checkout total does not match server quote.',
        );
      case _RuntimeMode.success:
        final transaction = CheckoutTransaction(
          id: 'txn-${draft.idempotencyKey}',
          number: 'TRX-${draft.idempotencyKey}',
          status: 'paid',
          grandTotal: draft.quote!.grandTotal,
        );
        createdTransactions.add(transaction);
        return transaction;
    }
  }

  @override
  Future<CheckoutRecovery> recoverCheckout(String idempotencyKey) async {
    final transaction = createdTransactions
        .where((transaction) => transaction.id == 'txn-$idempotencyKey')
        .firstOrNull;
    return transaction == null
        ? const CheckoutRecovery(status: CheckoutRecoveryStatus.missing)
        : CheckoutRecovery(
            status: CheckoutRecoveryStatus.completed,
            transaction: transaction,
          );
  }

  @override
  Future<CheckoutPaymentLine> confirmPayment({
    required String paymentId,
    required String idempotencyKey,
  }) {
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
  Future<List<CheckoutTransaction>> listTransactions({String? status}) {
    throw UnimplementedError();
  }

  @override
  Future<VoidResult> voidTransaction({
    required String transactionId,
    required String shiftId,
    required String reason,
    required String idempotencyKey,
  }) {
    throw UnimplementedError();
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

class _RuntimeAuthRepository implements AuthRepository {
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
  Future<void> logout() async {}

  @override
  Future<AuthSession> me() => login(
    email: 'owner@test.local',
    password: 'password',
    deviceUuid: 'device-uuid',
  );

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
  Future<List<Outlet>> listOutlets() async => const [
    Outlet(id: 'outlet-id', name: 'Outlet Demo', isOnline: true),
  ];
}

class _RuntimeShiftRepository implements ShiftRepository {
  int closeCalls = 0;

  @override
  Future<void> cashMovement({
    required String shiftId,
    required String type,
    required int amount,
    String? reason,
  }) async {}

  @override
  Future<ShiftSession> closeShift({
    required String shiftId,
    required int actualCash,
    required String pin,
    String? varianceReason,
  }) async {
    closeCalls += 1;
    return _shift(status: 'closed');
  }

  @override
  Future<ShiftSession?> currentShift({
    required String outletId,
    required String deviceId,
  }) async {
    return _shift(status: 'open');
  }

  @override
  Future<ShiftSession> openShift({
    required String outletId,
    required String deviceId,
    required String cashierId,
    required int openingCash,
  }) async {
    return _shift(status: 'open');
  }

  ShiftSession _shift({required String status}) {
    return ShiftSession(
      id: 'shift-id',
      cashier: const Employee(
        id: 'cashier-id',
        name: 'Cashier',
        role: 'cashier',
      ),
      openedAt: DateTime(2026, 6, 19, 8),
      openingCash: 100000,
      status: status,
    );
  }
}

class _RuntimeReport {
  final rows = <String>[];
  final failures = <String>[];

  Future<void> step(String name, Future<void> Function() body) async {
    try {
      await body();
      rows.add('WORKING | $name | OK');
    } catch (error, stackTrace) {
      final line = stackTrace.toString().split('\n').first;
      final message = '$name: $error at $line';
      failures.add(message);
      rows.add('BROKEN | $name | $error at $line');
    }
  }

  void printSummary() {
    // ignore: avoid_print
    print(['NOJPOS_OUTBOX_RUNTIME_STATUS', ...rows].join('\n'));
  }
}
