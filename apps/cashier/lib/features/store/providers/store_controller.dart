import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../repositories/store_repository.dart';

final storeControllerProvider =
    NotifierProvider<StoreController, StoreControllerState>(
      StoreController.new,
    );

class StoreControllerState {
  const StoreControllerState({
    this.storeState,
    this.isBusy = false,
    this.errorMessage,
    this.blockingShifts = const [],
  });

  final StoreState? storeState;
  final bool isBusy;
  final String? errorMessage;
  final List<BlockingShift> blockingShifts;

  StoreControllerState copyWith({
    StoreState? storeState,
    bool? isBusy,
    String? errorMessage,
    bool clearError = false,
    List<BlockingShift>? blockingShifts,
  }) {
    return StoreControllerState(
      storeState: storeState ?? this.storeState,
      isBusy: isBusy ?? this.isBusy,
      errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
      blockingShifts: blockingShifts ?? this.blockingShifts,
    );
  }
}

class StoreController extends Notifier<StoreControllerState> {
  @override
  StoreControllerState build() => const StoreControllerState();

  StoreRepository get _repository => ref.read(storeRepositoryProvider);

  Future<void> load(String outletId) async {
    if (outletId.isEmpty) return;
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final storeState = await _repository.getStoreState(outletId);
      state = state.copyWith(
        storeState: storeState,
        blockingShifts: storeState.blockingShifts,
        isBusy: false,
      );
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
    }
  }

  Future<bool> openStore({
    required String outletId,
    required String pin,
    String? reason,
  }) async {
    return _operate(
      () => _repository.openStore(
        outletId: outletId,
        authorizationPin: pin,
        reason: reason,
      ),
    );
  }

  Future<bool> closeStore({
    required String outletId,
    required String pin,
    String? reason,
  }) async {
    return _operate(
      () => _repository.closeStore(
        outletId: outletId,
        authorizationPin: pin,
        reason: reason,
      ),
    );
  }

  Future<bool> _operate(Future<StoreState> Function() command) async {
    state = state.copyWith(
      isBusy: true,
      clearError: true,
      blockingShifts: const [],
    );
    try {
      final storeState = await command();
      state = state.copyWith(storeState: storeState, isBusy: false);
      return true;
    } catch (error) {
      state = state.copyWith(
        isBusy: false,
        errorMessage: _messageFor(error),
        blockingShifts: _blockingShiftsFrom(error),
      );
      return false;
    }
  }
}

String _messageFor(Object error) {
  if (error is ApiException) {
    return switch (error.code) {
      'STORE_CLOSE_BLOCKED_OPEN_SHIFTS' =>
        'Tutup toko ditolak. Masih ada shift terbuka atau pending close.',
      'STORE_CLOSED_CHECKOUT_BLOCKED' =>
        'Toko sedang tutup. Buka toko sebelum checkout.',
      'STORE_CLOSED_SHIFT_OPEN_BLOCKED' =>
        'Toko sedang tutup. Buka toko sebelum membuka shift.',
      'STORE_ALREADY_OPEN' => 'Toko sudah buka.',
      'STORE_ALREADY_CLOSED' => 'Toko sudah tutup.',
      'PIN_REQUIRED' => 'PIN otorisasi wajib diisi.',
      'INVALID_PIN' => 'PIN otorisasi tidak valid.',
      'STORE_OPEN_CLOSE_DISABLED' =>
        'Buka/tutup toko belum tersedia untuk outlet ini.',
      _ => error.message,
    };
  }
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}

List<BlockingShift> _blockingShiftsFrom(Object error) {
  if (error is! ApiException) return const [];
  final values = error.details['blocking_shifts'];
  if (values is! List) return const [];
  return [for (final value in values) BlockingShift.fromJson(value)];
}
