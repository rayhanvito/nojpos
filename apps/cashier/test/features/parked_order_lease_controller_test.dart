import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/app/providers/nojpos_session_provider.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/features/auth/repositories/auth_repository.dart';
import 'package:nojpos_tablet_ui/features/pos/models/cart_item.dart';
import 'package:nojpos_tablet_ui/features/pos/models/product.dart';
import 'package:nojpos_tablet_ui/features/pos/providers/parked_order_lease_controller.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  test(
    'reopen acquire lease success stores context and removes saved order',
    () async {
      final transactions = FakeTransactionRepository();
      final container = leaseContainer(transactions);
      addTearDown(container.dispose);
      await readySession(container);
      await container
          .read(nojposSessionProvider.notifier)
          .loadHeldTransactions();

      final activated = await container
          .read(parkedOrderLeaseControllerProvider.notifier)
          .acquireForEdit('held-id');

      final lease = container.read(parkedOrderLeaseControllerProvider).active;
      expect(activated?.id, 'held-id');
      expect(transactions.acquireCount, 1);
      expect(lease?.revision, 7);
      expect(lease?.leaseDeviceId, 'device-id');
      expect(container.read(nojposSessionProvider).savedOrders, isEmpty);
    },
  );

  test('reopen ORDER_LOCKED does not move order out of saved list', () async {
    final transactions = FakeTransactionRepository()
      ..acquireError = const ConflictApiException(
        code: 'ORDER_LOCKED',
        message: 'Locked.',
      );
    final container = leaseContainer(transactions);
    addTearDown(container.dispose);
    await readySession(container);
    await container.read(nojposSessionProvider.notifier).loadHeldTransactions();

    final activated = await container
        .read(parkedOrderLeaseControllerProvider.notifier)
        .acquireForEdit('held-id');

    final leaseState = container.read(parkedOrderLeaseControllerProvider);
    expect(activated, isNull);
    expect(leaseState.active, isNull);
    expect(leaseState.errorMessage, 'Order sedang dibuka terminal/kasir lain.');
    expect(
      container.read(nojposSessionProvider).savedOrders.single.id,
      'held-id',
    );
  });

  test('heartbeat refreshes lease and stops after cancel/release', () async {
    final transactions = FakeTransactionRepository();
    final container = leaseContainer(
      transactions,
      heartbeatInterval: const Duration(milliseconds: 10),
    );
    addTearDown(container.dispose);
    await readySession(container);
    await container.read(nojposSessionProvider.notifier).loadHeldTransactions();
    await container
        .read(parkedOrderLeaseControllerProvider.notifier)
        .acquireForEdit('held-id');
    await Future<void>.delayed(const Duration(milliseconds: 35));

    expect(transactions.refreshCount, greaterThanOrEqualTo(2));

    await container
        .read(parkedOrderLeaseControllerProvider.notifier)
        .releaseActive(restoreOrder: true);
    final refreshCountAfterRelease = transactions.refreshCount;
    await Future<void>.delayed(const Duration(milliseconds: 30));

    expect(transactions.releaseCount, 1);
    expect(transactions.refreshCount, refreshCountAfterRelease);
    expect(container.read(parkedOrderLeaseControllerProvider).active, isNull);
    expect(
      container.read(nojposSessionProvider).savedOrders.single.id,
      'held-id',
    );
  });

  test('save active parked order sends revision and releases lease', () async {
    final transactions = FakeTransactionRepository();
    final container = leaseContainer(transactions);
    addTearDown(container.dispose);
    await readySession(container);
    await container.read(nojposSessionProvider.notifier).loadHeldTransactions();
    await container
        .read(parkedOrderLeaseControllerProvider.notifier)
        .acquireForEdit('held-id');

    final updated = await container
        .read(parkedOrderLeaseControllerProvider.notifier)
        .saveCurrentCartAsParked(cartItems: [cartItem()]);

    expect(updated?.revision, 8);
    expect(transactions.lastExpectedRevision, 7);
    expect(transactions.lastUpdatedItems.single.name, 'Kopi Susu');
    expect(transactions.releaseCount, 1);
    expect(container.read(parkedOrderLeaseControllerProvider).active, isNull);
  });

  test(
    'stale revision conflict keeps lease and exposes reload message',
    () async {
      final transactions = FakeTransactionRepository()
        ..updateError = const ConflictApiException(
          code: 'CONFLICT_REVISION',
          message: 'Stale.',
        );
      final container = leaseContainer(transactions);
      addTearDown(container.dispose);
      await readySession(container);
      await container
          .read(nojposSessionProvider.notifier)
          .loadHeldTransactions();
      await container
          .read(parkedOrderLeaseControllerProvider.notifier)
          .acquireForEdit('held-id');

      await expectLater(
        () => container
            .read(parkedOrderLeaseControllerProvider.notifier)
            .saveCurrentCartAsParked(cartItems: [cartItem()]),
        throwsA(isA<ConflictApiException>()),
      );

      final state = container.read(parkedOrderLeaseControllerProvider);
      expect(state.active?.transactionId, 'held-id');
      expect(
        state.errorMessage,
        'Order sudah berubah. Muat ulang sebelum menyimpan.',
      );
      expect(transactions.releaseCount, 0);
    },
  );
}

ProviderContainer leaseContainer(
  FakeTransactionRepository transactions, {
  Duration heartbeatInterval = const Duration(hours: 1),
}) {
  return ProviderContainer(
    overrides: [
      authRepositoryProvider.overrideWith((ref) => FakeAuthRepository()),
      transactionRepositoryProvider.overrideWith((ref) => transactions),
      parkedOrderLeaseHeartbeatIntervalProvider.overrideWithValue(
        heartbeatInterval,
      ),
    ],
  );
}

Future<void> readySession(ProviderContainer container) async {
  await container
      .read(nojposSessionProvider.notifier)
      .login(
        email: 'demo@example.test',
        password: 'secret',
        deviceUuid: 'uuid',
      );
}

CartItem cartItem() {
  return const CartItem(
    product: Product(
      id: 'product-id',
      name: 'Kopi Susu',
      category: 'Minuman',
      price: 18000,
      imageUrl: '',
    ),
    quantity: 2,
  );
}

class FakeAuthRepository implements AuthRepository {
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
  Future<AuthSession> me() {
    return login(
      email: 'demo@example.test',
      password: 'secret',
      deviceUuid: 'uuid',
    );
  }

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

class FakeTransactionRepository implements TransactionRepository {
  int acquireCount = 0;
  int refreshCount = 0;
  int releaseCount = 0;
  int? lastExpectedRevision;
  List<CheckoutItem> lastUpdatedItems = const [];
  ConflictApiException? acquireError;
  ConflictApiException? updateError;

  @override
  Future<List<CheckoutTransaction>> listParkedOrders() async => [
    heldTransaction(revision: 7),
  ];

  @override
  Future<CheckoutTransaction> acquireParkedOrderLease({
    required String transactionId,
    required String deviceId,
  }) async {
    acquireCount += 1;
    final error = acquireError;
    if (error != null) throw error;
    return heldTransaction(revision: 7);
  }

  @override
  Future<CheckoutTransaction> refreshParkedOrderLease({
    required String transactionId,
    required String deviceId,
  }) async {
    refreshCount += 1;
    return heldTransaction(revision: 7, leaseRemainingSeconds: 90);
  }

  @override
  Future<CheckoutTransaction> releaseParkedOrderLease({
    required String transactionId,
    required String deviceId,
  }) async {
    releaseCount += 1;
    return heldTransaction(revision: 7, leaseDeviceId: null, leaseUserId: null);
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
    final error = updateError;
    if (error != null) throw error;
    lastExpectedRevision = expectedRevision;
    lastUpdatedItems = items;
    return heldTransaction(revision: expectedRevision + 1, items: items);
  }

  CheckoutTransaction heldTransaction({
    required int revision,
    int leaseRemainingSeconds = 80,
    String? leaseDeviceId = 'device-id',
    String? leaseUserId = 'cashier-id',
    List<CheckoutItem> items = const [
      CheckoutItem(
        productId: 'product-id',
        name: 'Nasi Goreng',
        quantity: 1,
        unitPrice: 25000,
      ),
    ],
  }) {
    return CheckoutTransaction(
      id: 'held-id',
      number: 'TRX-HELD',
      status: 'held',
      grandTotal: items.fold(0, (sum, item) => sum + item.subtotal),
      revision: revision,
      leaseDeviceId: leaseDeviceId,
      leaseUserId: leaseUserId,
      leaseExpiresAt: '2026-06-20T04:00:00Z',
      leaseRemainingSeconds: leaseRemainingSeconds,
      createdAt: DateTime(2026, 6, 20, 10),
      customerId: 'customer-id',
      customerName: 'Ani',
      items: items,
    );
  }

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
    return heldTransaction(revision: 1, items: draft.items);
  }

  @override
  Future<CheckoutTransaction> createParkedOrder(CheckoutDraft draft) async {
    return heldTransaction(revision: 1, items: draft.items);
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
    return CheckoutPaymentLine(
      id: paymentId,
      transactionId: 'held-id',
      method: 'Tunai',
      amount: 1000,
      status: 'confirmed',
      isCash: true,
    );
  }

  @override
  Future<List<CheckoutTransaction>> listTransactions({String? status}) async {
    return [heldTransaction(revision: 7)];
  }

  @override
  Future<VoidResult> voidTransaction({
    required String transactionId,
    required String shiftId,
    required String reason,
    required String idempotencyKey,
  }) async {
    return VoidResult(
      transactionId: transactionId,
      status: 'voided',
      reason: reason,
    );
  }

  @override
  Future<RefundResult> createRefund({
    required String transactionId,
    required RefundRequest request,
  }) async {
    return RefundResult(
      id: 'refund-id',
      transactionId: transactionId,
      transactionStatus: 'refunded',
      status: 'finalized',
      refundMethod: request.refundMethod,
      totalRefundAmount: 1000,
      remainingRefundableAmount: 0,
    );
  }

  @override
  SalesTransaction createLocalTransaction({
    required SalesOrder order,
    required List<PaymentLine> payments,
    required Employee cashier,
  }) {
    return SalesTransaction(
      id: 'local-id',
      number: order.number,
      order: order,
      payments: payments,
      cashier: cashier,
      createdAt: DateTime(2026, 6, 20, 10),
    );
  }
}
