import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../transactions/repositories/transaction_repository.dart';
import '../repositories/pending_payment_repository.dart';

final pendingPaymentRepositoryProvider = Provider<PendingPaymentRepository>(
  (ref) => ApiPendingPaymentRepository(apiClient: ref.watch(apiClientProvider)),
);

final pendingPaymentPollIntervalProvider = Provider<Duration>(
  (ref) => const Duration(seconds: 3),
);

final pendingPaymentControllerProvider =
    NotifierProvider<PendingPaymentController, PendingPaymentState>(
      PendingPaymentController.new,
    );

class PendingPaymentState {
  const PendingPaymentState({
    this.transaction,
    this.errorMessage,
    this.active = false,
  });

  final CheckoutTransaction? transaction;
  final String? errorMessage;
  final bool active;

  bool get isTerminal => switch (transaction?.status) {
    'paid' || 'payment_failed' || 'expired' || 'declined' => true,
    _ => false,
  };

  PendingPaymentState copyWith({
    CheckoutTransaction? transaction,
    String? errorMessage,
    bool? active,
    bool clearError = false,
  }) => PendingPaymentState(
    transaction: transaction ?? this.transaction,
    errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
    active: active ?? this.active,
  );
}

class PendingPaymentController extends Notifier<PendingPaymentState> {
  Timer? _timer;
  bool _disposed = false;

  @override
  PendingPaymentState build() {
    ref.onDispose(() {
      _disposed = true;
      stop();
    });
    return const PendingPaymentState();
  }

  void start(CheckoutTransaction transaction) {
    stop();
    state = PendingPaymentState(transaction: transaction, active: true);
    _poll();
    _timer = Timer.periodic(
      ref.read(pendingPaymentPollIntervalProvider),
      (_) => _poll(),
    );
  }

  Future<void> _poll() async {
    final current = state.transaction;
    if (_disposed || current == null || state.isTerminal) {
      stop();
      return;
    }
    try {
      final next = await ref
          .read(pendingPaymentRepositoryProvider)
          .getTransaction(current.id);
      if (_disposed || state.transaction?.id != current.id) return;
      state = state.copyWith(transaction: next, clearError: true);
      if (state.isTerminal) stop();
    } catch (error) {
      if (!_disposed) {
        state = state.copyWith(
          errorMessage: error is ApiException
              ? error.message
              : 'Gagal memperbarui status pembayaran.',
        );
      }
    }
  }

  void stop() {
    _timer?.cancel();
    _timer = null;
    if (!_disposed && state.active) {
      state = state.copyWith(active: false);
    }
  }
}
