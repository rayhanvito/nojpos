import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/app/providers/nojpos_session_provider.dart';
import 'package:nojpos_tablet_ui/core/outbox/checkout_outbox.dart';
import 'package:nojpos_tablet_ui/features/auth/repositories/auth_repository.dart';
import 'package:nojpos_tablet_ui/features/shift/repositories/shift_repository.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  test(
    'lock returns to PIN while keeping token session and open shift',
    () async {
      final auth = _FakeAuthRepository();
      final shift = _FakeShiftRepository();
      final container = ProviderContainer(
        overrides: [
          authRepositoryProvider.overrideWith((ref) => auth),
          shiftRepositoryProvider.overrideWith((ref) => shift),
        ],
      );
      addTearDown(container.dispose);

      await container
          .read(nojposSessionProvider.notifier)
          .login(
            email: 'owner@demo.nojpos.test',
            password: 'password',
            deviceUuid: 'device-uuid',
          );
      await container.read(nojposSessionProvider.notifier).pinSwitch('1234');
      await container
          .read(nojposSessionProvider.notifier)
          .openShift(openingCash: 100000);

      container.read(nojposSessionProvider.notifier).lock();

      final state = container.read(nojposSessionProvider);
      expect(state.status, SessionStatus.pinRequired);
      expect(state.activeShift?.id, 'shift-id');
      expect(auth.logoutCalls, 0);
      expect(shift.closeCalls, 0);
    },
  );

  test(
    'logout revokes token locally without closing the server shift',
    () async {
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

      await container
          .read(nojposSessionProvider.notifier)
          .login(
            email: 'owner@demo.nojpos.test',
            password: 'password',
            deviceUuid: 'device-uuid',
          );
      await container.read(nojposSessionProvider.notifier).pinSwitch('1234');
      await container
          .read(nojposSessionProvider.notifier)
          .openShift(openingCash: 100000);

      await container.read(nojposSessionProvider.notifier).logout();

      final state = container.read(nojposSessionProvider);
      expect(auth.logoutCalls, 1);
      expect(shift.closeCalls, 0);
      expect(state.status, SessionStatus.unauthenticated);
    },
  );

  test(
    'successful close clears active shift so POS routes back to shift gate',
    () async {
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
        email: 'owner@demo.nojpos.test',
        password: 'password',
        deviceUuid: 'device-uuid',
      );
      await notifier.pinSwitch('1234');
      await notifier.openShift(openingCash: 100000);

      final closed = await notifier.closeShift(actualCash: 100000, pin: '1234');

      expect(closed, true);
      expect(shift.closeCalls, 1);
      expect(container.read(nojposSessionProvider).activeShift, null);
      expect(container.read(nojposSessionProvider).hasOpenShift, false);
    },
  );

  test('cash wallet records real cash movement and updates summary', () async {
    final auth = _FakeAuthRepository();
    final shift = _FakeShiftRepository();
    final container = ProviderContainer(
      overrides: [
        authRepositoryProvider.overrideWith((ref) => auth),
        shiftRepositoryProvider.overrideWith((ref) => shift),
      ],
    );
    addTearDown(container.dispose);

    final notifier = container.read(nojposSessionProvider.notifier);
    await notifier.login(
      email: 'owner@demo.nojpos.test',
      password: 'password',
      deviceUuid: 'device-uuid',
    );
    await notifier.pinSwitch('1234');
    await notifier.openShift(openingCash: 100000);

    final ok = await notifier.addCashMovement(
      type: CashMovementType.cashIn,
      amount: 25000,
      reason: 'Modal tambahan',
    );

    final state = container.read(nojposSessionProvider);
    expect(ok, true);
    expect(shift.cashMovementCalls, 1);
    expect(shift.lastCashMovementType, 'cash_in');
    expect(state.cashMovements.single.amount, 25000);
    expect(state.cashSummary, 125000);
  });

  test('sales history is loaded from API transactions', () async {
    final transactions = _FakeTransactionRepository();
    final container = ProviderContainer(
      overrides: [
        transactionRepositoryProvider.overrideWith((ref) => transactions),
      ],
    );
    addTearDown(container.dispose);

    await container
        .read(nojposSessionProvider.notifier)
        .loadSalesTransactions(status: 'paid');

    final state = container.read(nojposSessionProvider);
    expect(transactions.lastListStatus, 'paid');
    expect(state.transactions.single.id, 'transaction-paid');
    expect(state.transactions.single.order.lines.single.name, 'Kopi Susu');
    expect(state.transactions.single.payments.single.methodName, 'Tunai');
  });

  test('held list is loaded from API and can reopen into cart data', () async {
    final auth = _FakeAuthRepository();
    final transactions = _FakeTransactionRepository();
    final container = ProviderContainer(
      overrides: [
        authRepositoryProvider.overrideWith((ref) => auth),
        transactionRepositoryProvider.overrideWith((ref) => transactions),
      ],
    );
    addTearDown(container.dispose);

    final notifier = container.read(nojposSessionProvider.notifier);
    await notifier.login(
      email: 'owner@demo.nojpos.test',
      password: 'password',
      deviceUuid: 'device-uuid',
    );
    await notifier.loadHeldTransactions();
    final order = container.read(nojposSessionProvider).savedOrders.single;
    final activated = await notifier.activateSavedOrder(order.id);

    expect(transactions.parkedOrdersListed, true);
    expect(transactions.lastAcquiredLeaseTransactionId, order.id);
    expect(order.id, 'transaction-held');
    expect(order.number, 'TRX-HELD');
    expect(activated?.lines.single.name, 'Nasi Goreng');
  });

  test('void uses the server transaction id from sales history', () async {
    final auth = _FakeAuthRepository();
    final shift = _FakeShiftRepository();
    final transactions = _FakeTransactionRepository();
    final container = ProviderContainer(
      overrides: [
        authRepositoryProvider.overrideWith((ref) => auth),
        shiftRepositoryProvider.overrideWith((ref) => shift),
        transactionRepositoryProvider.overrideWith((ref) => transactions),
      ],
    );
    addTearDown(container.dispose);

    final notifier = container.read(nojposSessionProvider.notifier);
    await notifier.login(
      email: 'owner@demo.nojpos.test',
      password: 'password',
      deviceUuid: 'device-uuid',
    );
    await notifier.pinSwitch('1234');
    await notifier.openShift(openingCash: 100000);
    await notifier.loadSalesTransactions(status: 'paid');

    final ok = await notifier.voidTransaction(
      transaction: container.read(nojposSessionProvider).transactions.single,
      reason: 'Salah input',
    );

    expect(ok, true);
    expect(transactions.lastVoidedTransactionId, 'transaction-paid');
    expect(
      container.read(nojposSessionProvider).transactions.single.status,
      'voided',
    );
  });

  test(
    'pending payments list from API and confirm reflects server status',
    () async {
      final transactions = _FakeTransactionRepository();
      transactions.confirmedTransactionStatus = 'partial';
      final container = ProviderContainer(
        overrides: [
          transactionRepositoryProvider.overrideWith((ref) => transactions),
        ],
      );
      addTearDown(container.dispose);

      final notifier = container.read(nojposSessionProvider.notifier);
      await notifier.loadSalesTransactions(status: 'pending');

      final pending = container.read(nojposSessionProvider).transactions.single;
      expect(transactions.lastListStatus, 'pending');
      expect(pending.status, 'partial');
      expect(pending.payments.single.status, 'pending');

      await notifier.confirmPaymentLine('payment-pending');

      final updated = container.read(nojposSessionProvider).transactions.single;
      expect(transactions.lastConfirmedPaymentId, 'payment-pending');
      expect(updated.status, 'partial');
      expect(updated.payments.single.status, 'confirmed');
    },
  );
}

class _FakeAuthRepository implements AuthRepository {
  int logoutCalls = 0;

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
  Future<AuthSession> me() async => login(
    email: 'owner@demo.nojpos.test',
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
  Future<void> logout() async {
    logoutCalls += 1;
  }
}

class _FakeShiftRepository implements ShiftRepository {
  int closeCalls = 0;
  int cashMovementCalls = 0;
  String? lastCashMovementType;

  @override
  Future<ShiftSession> openShift({
    required String outletId,
    required String deviceId,
    required String cashierId,
    required int openingCash,
  }) async {
    return ShiftSession(
      id: 'shift-id',
      cashier: const Employee(
        id: 'cashier-id',
        name: 'Cashier',
        role: 'cashier',
      ),
      openedAt: DateTime(2026, 6, 18, 9),
      openingCash: 100000,
    );
  }

  @override
  Future<ShiftSession?> currentShift({
    required String outletId,
    required String deviceId,
  }) async {
    return openShift(
      outletId: outletId,
      deviceId: deviceId,
      cashierId: 'cashier-id',
      openingCash: 100000,
    );
  }

  @override
  Future<ShiftSession> closeShift({
    required String shiftId,
    required int actualCash,
    required String pin,
    String? varianceReason,
  }) async {
    closeCalls += 1;
    return ShiftSession(
      id: shiftId,
      cashier: const Employee(
        id: 'cashier-id',
        name: 'Cashier',
        role: 'cashier',
      ),
      openedAt: DateTime(2026, 6, 18),
      closedAt: DateTime(2026, 6, 18, 12),
      status: 'closed',
    );
  }

  @override
  Future<void> cashMovement({
    required String shiftId,
    required String type,
    required int amount,
    String? reason,
  }) async {
    cashMovementCalls += 1;
    lastCashMovementType = type;
  }
}

class _FakeTransactionRepository implements TransactionRepository {
  String? lastListStatus;
  bool parkedOrdersListed = false;
  String? lastAcquiredLeaseTransactionId;
  String? lastVoidedTransactionId;
  String? lastConfirmedPaymentId;
  String confirmedTransactionStatus = 'paid';

  @override
  Future<CheckoutQuote> quoteTransaction(CheckoutDraft draft) async {
    return CheckoutQuote(
      subtotal: draft.subtotal,
      itemDiscountTotal: draft.itemDiscountTotal,
      cartDiscountTotal: draft.cartDiscount,
      discountTotal: draft.discountTotal,
      serviceChargeTotal: 0,
      taxTotal: 0,
      roundingTotal: draft.rounding,
      grandTotal: draft.grandTotal,
    );
  }

  @override
  Future<CheckoutTransaction> createTransaction(CheckoutDraft draft) async {
    return const CheckoutTransaction(
      id: 'transaction-created',
      number: 'TRX-CREATED',
      status: 'paid',
      grandTotal: 50000,
    );
  }

  @override
  Future<CheckoutTransaction> createParkedOrder(CheckoutDraft draft) async {
    return CheckoutTransaction(
      id: 'transaction-held',
      number: 'TRX-HELD',
      status: 'held',
      grandTotal: draft.grandTotal,
      revision: 1,
      items: draft.items,
    );
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
      transactionId: 'transaction-paid',
      method: 'QRIS Statis',
      amount: 10000,
      status: 'confirmed',
      isCash: false,
      confirmedBy: 'cashier-id',
      confirmedAt: '2026-06-18T10:00:00Z',
      transactionStatus: confirmedTransactionStatus,
    );
  }

  @override
  SalesTransaction createLocalTransaction({
    required SalesOrder order,
    required List<PaymentLine> payments,
    required Employee cashier,
  }) {
    return SalesTransaction(
      id: 'local-transaction',
      number: 'TRX-LOCAL',
      order: order,
      payments: payments,
      cashier: cashier,
      createdAt: DateTime(2026, 6, 18, 10),
    );
  }

  @override
  Future<List<CheckoutTransaction>> listTransactions({String? status}) async {
    lastListStatus = status;
    if (status == 'held') {
      return [
        CheckoutTransaction(
          id: 'transaction-held',
          number: 'TRX-HELD',
          status: 'held',
          grandTotal: 25000,
          createdAt: DateTime(2026, 6, 18, 11),
          customerId: 'customer-id',
          customerName: 'Ani',
          cashierId: 'cashier-id',
          cashierName: 'Cashier',
          items: const [
            CheckoutItem(
              productId: 'product-held',
              name: 'Nasi Goreng',
              quantity: 1,
              unitPrice: 25000,
            ),
          ],
        ),
      ];
    }
    if (status == 'pending') {
      return [
        CheckoutTransaction(
          id: 'transaction-paid',
          number: 'TRX-PENDING',
          status: 'partial',
          grandTotal: 50000,
          createdAt: DateTime(2026, 6, 18, 12),
          cashierId: 'cashier-id',
          cashierName: 'Cashier',
          items: const [
            CheckoutItem(
              productId: 'product-pending',
              name: 'Bakmi',
              quantity: 1,
              unitPrice: 50000,
            ),
          ],
          payments: const [
            CheckoutPaymentLine(
              id: 'payment-pending',
              transactionId: 'transaction-paid',
              method: 'QRIS Statis',
              amount: 25000,
              reference: 'QR-001',
              status: 'pending',
              isCash: false,
            ),
          ],
        ),
      ];
    }
    return [
      CheckoutTransaction(
        id: 'transaction-paid',
        number: 'TRX-PAID',
        status: 'paid',
        grandTotal: 36000,
        createdAt: DateTime(2026, 6, 18, 10),
        cashierId: 'cashier-id',
        cashierName: 'Cashier',
        items: const [
          CheckoutItem(
            productId: 'product-paid',
            name: 'Kopi Susu',
            quantity: 2,
            unitPrice: 18000,
          ),
        ],
        payments: const [
          CheckoutPaymentLine(
            id: 'payment-id',
            transactionId: 'transaction-paid',
            method: 'Tunai',
            amount: 36000,
            status: 'confirmed',
            isCash: true,
          ),
        ],
      ),
    ];
  }

  @override
  Future<List<CheckoutTransaction>> listParkedOrders() async {
    parkedOrdersListed = true;
    return listTransactions(status: 'held');
  }

  @override
  Future<CheckoutTransaction> acquireParkedOrderLease({
    required String transactionId,
    required String deviceId,
  }) async {
    lastAcquiredLeaseTransactionId = transactionId;
    return const CheckoutTransaction(
      id: 'transaction-held',
      number: 'TRX-HELD',
      status: 'held',
      grandTotal: 25000,
      revision: 1,
      leaseDeviceId: 'device-id',
      leaseUserId: 'cashier-id',
      items: [
        CheckoutItem(
          productId: 'product-held',
          name: 'Nasi Goreng',
          quantity: 1,
          unitPrice: 25000,
        ),
      ],
    );
  }

  @override
  Future<CheckoutTransaction> refreshParkedOrderLease({
    required String transactionId,
    required String deviceId,
  }) async {
    return acquireParkedOrderLease(
      transactionId: transactionId,
      deviceId: deviceId,
    );
  }

  @override
  Future<CheckoutTransaction> releaseParkedOrderLease({
    required String transactionId,
    required String deviceId,
  }) async {
    return const CheckoutTransaction(
      id: 'transaction-held',
      number: 'TRX-HELD',
      status: 'held',
      grandTotal: 25000,
      revision: 1,
    );
  }

  @override
  Future<CheckoutTransaction> updateParkedOrder({
    required String transactionId,
    required String deviceId,
    required int expectedRevision,
    required List<CheckoutItem> items,
    String? customerId,
    String? servedBy,
    int cartDiscount = 0,
    String? notes,
  }) async {
    return CheckoutTransaction(
      id: transactionId,
      number: 'TRX-HELD',
      status: 'held',
      grandTotal: items.fold(0, (sum, item) => sum + item.subtotal),
      revision: expectedRevision + 1,
      items: items,
    );
  }

  @override
  Future<VoidResult> voidTransaction({
    required String transactionId,
    required String shiftId,
    required String reason,
    required String idempotencyKey,
  }) async {
    lastVoidedTransactionId = transactionId;
    return VoidResult(
      transactionId: transactionId,
      status: 'voided',
      reason: reason,
    );
  }
}
