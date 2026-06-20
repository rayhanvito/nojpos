import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../core/network/api_client.dart';
import '../../../shared/models/nojpos_models.dart';
import '../../transactions/repositories/transaction_repository.dart';
import '../models/cart_item.dart';

final parkedOrderLeaseHeartbeatIntervalProvider = Provider<Duration>(
  (ref) => const Duration(seconds: 30),
);

final parkedOrderLeaseControllerProvider =
    NotifierProvider<ParkedOrderLeaseController, ParkedOrderLeaseState>(
      ParkedOrderLeaseController.new,
    );

class ParkedOrderLeaseState {
  const ParkedOrderLeaseState({
    this.active,
    this.isBusy = false,
    this.errorMessage,
  });

  final ParkedOrderLeaseContext? active;
  final bool isBusy;
  final String? errorMessage;

  bool get hasActiveLease => active != null;

  ParkedOrderLeaseState copyWith({
    ParkedOrderLeaseContext? active,
    bool clearActive = false,
    bool? isBusy,
    String? errorMessage,
    bool clearError = false,
  }) {
    return ParkedOrderLeaseState(
      active: clearActive ? null : active ?? this.active,
      isBusy: isBusy ?? this.isBusy,
      errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
    );
  }
}

class ParkedOrderLeaseContext {
  const ParkedOrderLeaseContext({
    required this.transactionId,
    required this.orderNumber,
    required this.revision,
    required this.deviceId,
    required this.order,
    this.leaseDeviceId,
    this.leaseUserId,
    this.leaseExpiresAt,
    this.leaseRemainingSeconds = 0,
  });

  final String transactionId;
  final String orderNumber;
  final int revision;
  final String deviceId;
  final SalesOrder order;
  final String? leaseDeviceId;
  final String? leaseUserId;
  final String? leaseExpiresAt;
  final int leaseRemainingSeconds;

  ParkedOrderLeaseContext copyWith({
    int? revision,
    SalesOrder? order,
    String? leaseDeviceId,
    String? leaseUserId,
    String? leaseExpiresAt,
    int? leaseRemainingSeconds,
    bool clearLease = false,
  }) {
    return ParkedOrderLeaseContext(
      transactionId: transactionId,
      orderNumber: orderNumber,
      revision: revision ?? this.revision,
      deviceId: deviceId,
      order: order ?? this.order,
      leaseDeviceId: clearLease ? null : leaseDeviceId ?? this.leaseDeviceId,
      leaseUserId: clearLease ? null : leaseUserId ?? this.leaseUserId,
      leaseExpiresAt: clearLease ? null : leaseExpiresAt ?? this.leaseExpiresAt,
      leaseRemainingSeconds: clearLease
          ? 0
          : leaseRemainingSeconds ?? this.leaseRemainingSeconds,
    );
  }
}

class ParkedOrderLeaseController extends Notifier<ParkedOrderLeaseState> {
  Timer? _heartbeatTimer;
  bool _disposed = false;

  @override
  ParkedOrderLeaseState build() {
    _disposed = false;
    ref.onDispose(() {
      _disposed = true;
      _stopHeartbeat();
    });
    ref.listen<String>(
      nojposSessionProvider.select(
        (session) =>
            '${session.status.name}:${session.deviceId ?? ''}:${session.outlet.id}:${session.cashier.id}',
      ),
      (previous, next) {
        if (previous == null || previous == next || state.active == null) {
          return;
        }
        unawaited(releaseActive(restoreOrder: true));
      },
    );
    return const ParkedOrderLeaseState();
  }

  Future<SalesOrder?> acquireForEdit(String transactionId) async {
    final session = ref.read(nojposSessionProvider);
    final deviceId = session.deviceId;
    if (deviceId == null || deviceId.isEmpty) {
      state = state.copyWith(errorMessage: 'Perangkat belum terdaftar.');
      return null;
    }

    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final transaction = await ref
          .read(transactionRepositoryProvider)
          .acquireParkedOrderLease(
            transactionId: transactionId,
            deviceId: deviceId,
          );
      final order = ref
          .read(nojposSessionProvider.notifier)
          .consumeParkedOrderForEdit(transaction);
      if (order == null) {
        state = state.copyWith(
          isBusy: false,
          errorMessage: 'Order tidak ditemukan. Muat ulang daftar order.',
        );
        await _releaseTransactionBestEffort(
          transactionId: transaction.id,
          deviceId: deviceId,
        );
        return null;
      }
      _stopHeartbeat();
      state = state.copyWith(
        active: ParkedOrderLeaseContext(
          transactionId: transaction.id,
          orderNumber: transaction.number.isEmpty
              ? transaction.id
              : transaction.number,
          revision: transaction.revision,
          deviceId: deviceId,
          order: order,
          leaseDeviceId: transaction.leaseDeviceId,
          leaseUserId: transaction.leaseUserId,
          leaseExpiresAt: transaction.leaseExpiresAt,
          leaseRemainingSeconds: transaction.leaseRemainingSeconds,
        ),
        isBusy: false,
        clearError: true,
      );
      _startHeartbeat();
      return order;
    } catch (error) {
      state = state.copyWith(
        isBusy: false,
        errorMessage: parkedOrderLeaseMessage(error),
      );
      return null;
    }
  }

  Future<void> refreshActiveLease() async {
    final active = state.active;
    if (_disposed || active == null) return;
    try {
      final transaction = await ref
          .read(transactionRepositoryProvider)
          .refreshParkedOrderLease(
            transactionId: active.transactionId,
            deviceId: active.deviceId,
          );
      if (_disposed || state.active?.transactionId != active.transactionId) {
        return;
      }
      state = state.copyWith(
        active: active.copyWith(
          revision: transaction.revision,
          leaseDeviceId: transaction.leaseDeviceId,
          leaseUserId: transaction.leaseUserId,
          leaseExpiresAt: transaction.leaseExpiresAt,
          leaseRemainingSeconds: transaction.leaseRemainingSeconds,
        ),
        clearError: true,
      );
    } catch (error) {
      if (_disposed || state.active?.transactionId != active.transactionId) {
        return;
      }
      _stopHeartbeat();
      state = state.copyWith(errorMessage: parkedOrderLeaseMessage(error));
    }
  }

  Future<CheckoutTransaction?> saveCurrentCartAsParked({
    required List<CartItem> cartItems,
    int cartDiscount = 0,
    String notes = '',
    Employee? servedBy,
  }) async {
    final active = state.active;
    if (active == null) {
      return ref
          .read(nojposSessionProvider.notifier)
          .saveHeldTransaction(
            cartItems: cartItems,
            cartDiscount: cartDiscount,
            notes: notes,
            servedBy: servedBy,
          );
    }
    if (cartItems.isEmpty) return null;

    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final session = ref.read(nojposSessionProvider);
      final transaction = await ref
          .read(transactionRepositoryProvider)
          .updateParkedOrder(
            transactionId: active.transactionId,
            deviceId: active.deviceId,
            expectedRevision: active.revision,
            items: _checkoutItemsFromCart(cartItems),
            customerId: session.selectedCustomer?.id,
            servedBy: servedBy?.id,
            cartDiscount: cartDiscount,
            notes: notes,
          );
      ref
          .read(nojposSessionProvider.notifier)
          .upsertSavedOrderFromParkedTransaction(transaction);
      state = state.copyWith(
        active: active.copyWith(
          revision: transaction.revision,
          order: ref
              .read(nojposSessionProvider.notifier)
              .orderFromParkedTransaction(transaction),
        ),
        isBusy: false,
        clearError: true,
      );
      await releaseActive(restoreOrder: false);
      return transaction;
    } catch (error) {
      state = state.copyWith(
        isBusy: false,
        errorMessage: parkedOrderLeaseMessage(error),
      );
      rethrow;
    }
  }

  Future<void> releaseActive({bool restoreOrder = false}) async {
    final active = state.active;
    if (active == null) return;
    _stopHeartbeat();
    state = state.copyWith(isBusy: false, clearError: true);
    if (restoreOrder) {
      ref.read(nojposSessionProvider.notifier).restoreSavedOrder(active.order);
    }
    await _releaseTransactionBestEffort(
      transactionId: active.transactionId,
      deviceId: active.deviceId,
    );
    if (!_disposed && state.active?.transactionId == active.transactionId) {
      state = state.copyWith(
        clearActive: true,
        isBusy: false,
        clearError: true,
      );
    }
  }

  Future<void> releaseAfterCheckout() {
    return releaseActive(restoreOrder: false);
  }

  void _startHeartbeat() {
    _stopHeartbeat();
    _heartbeatTimer = Timer.periodic(
      ref.read(parkedOrderLeaseHeartbeatIntervalProvider),
      (_) => unawaited(refreshActiveLease()),
    );
  }

  void _stopHeartbeat() {
    _heartbeatTimer?.cancel();
    _heartbeatTimer = null;
  }

  Future<void> _releaseTransactionBestEffort({
    required String transactionId,
    required String deviceId,
  }) async {
    try {
      await ref
          .read(transactionRepositoryProvider)
          .releaseParkedOrderLease(
            transactionId: transactionId,
            deviceId: deviceId,
          );
    } catch (_) {
      // Best-effort cleanup only. Save/update/checkout errors are handled by callers.
    }
  }
}

String parkedOrderLeaseMessage(Object error) {
  if (error is ApiException) {
    return switch (error.code) {
      'ORDER_LOCKED' => 'Order sedang dibuka terminal/kasir lain.',
      'CONFLICT_REVISION' =>
        'Order sudah berubah. Muat ulang sebelum menyimpan.',
      'ORDER_LOCK_REQUIRED' => 'Order belum terkunci di terminal ini.',
      _ => error.message,
    };
  }
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}

List<CheckoutItem> _checkoutItemsFromCart(List<CartItem> cartItems) {
  return [
    for (final item in cartItems)
      CheckoutItem(
        productId: item.product.id,
        name: item.product.name,
        quantity: item.quantity,
        unitPrice: item.product.price,
        discount: item.discount,
      ),
  ];
}
