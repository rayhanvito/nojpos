import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../core/network/api_client.dart';
import '../repositories/terminal_lock_repository.dart';

final terminalLockControllerProvider =
    NotifierProvider<TerminalLockController, TerminalLockControllerState>(
      TerminalLockController.new,
    );

class TerminalLockControllerState {
  const TerminalLockControllerState({
    this.lockState,
    this.unlockResponse,
    this.isBusy = false,
    this.errorMessage,
  });

  final TerminalLockState? lockState;
  final TerminalUnlockResponse? unlockResponse;
  final bool isBusy;
  final String? errorMessage;

  bool get isLocked => lockState?.locked ?? false;

  TerminalLockControllerState copyWith({
    TerminalLockState? lockState,
    TerminalUnlockResponse? unlockResponse,
    bool? isBusy,
    String? errorMessage,
    bool clearError = false,
    bool clearUnlock = false,
  }) {
    return TerminalLockControllerState(
      lockState: lockState ?? this.lockState,
      unlockResponse: clearUnlock
          ? null
          : unlockResponse ?? this.unlockResponse,
      isBusy: isBusy ?? this.isBusy,
      errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
    );
  }
}

class TerminalLockController extends Notifier<TerminalLockControllerState> {
  @override
  TerminalLockControllerState build() => const TerminalLockControllerState();

  TerminalLockRepository get _repository =>
      ref.read(terminalLockRepositoryProvider);

  Future<void> load() async {
    final context = _contextOrNull();
    if (context == null) return;
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final lockState = await _repository.getLockState(
        outletId: context.outletId,
        deviceId: context.deviceId,
        staffId: context.cashierId,
      );
      state = state.copyWith(lockState: lockState, isBusy: false);
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
    }
  }

  Future<bool> lock({
    TerminalLockReason reason = TerminalLockReason.manual,
  }) async {
    final context = _contextOrNull();
    if (context == null) {
      state = state.copyWith(errorMessage: 'Terminal belum siap dikunci.');
      return false;
    }

    state = state.copyWith(isBusy: true, clearError: true, clearUnlock: true);
    try {
      final lockState = await _repository.lockTerminal(
        outletId: context.outletId,
        deviceId: context.deviceId,
        reason: reason,
        cashierId: context.cashierId,
        shiftId: context.shiftId,
      );
      state = state.copyWith(lockState: lockState, isBusy: false);
      return true;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return false;
    }
  }

  Future<bool> unlock(String pin, {String mode = 'resume_current'}) async {
    final context = _contextOrNull();
    if (context == null) {
      state = state.copyWith(errorMessage: 'Terminal belum siap dibuka.');
      return false;
    }

    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final unlock = await _repository.unlockTerminal(
        outletId: context.outletId,
        deviceId: context.deviceId,
        staffId: context.cashierId,
        pin: pin,
        mode: mode,
      );
      state = state.copyWith(
        lockState: unlock,
        unlockResponse: unlock,
        isBusy: false,
      );
      return true;
    } catch (error) {
      final current = state.lockState;
      state = state.copyWith(
        lockState: _lockStateFromError(current, error),
        isBusy: false,
        errorMessage: _messageFor(error),
      );
      return false;
    }
  }

  _TerminalContext? _contextOrNull() {
    final session = ref.read(nojposSessionProvider);
    final outletId = session.outlet.id;
    final deviceId = session.deviceId;
    final cashierId = session.cashier.id;
    if (outletId.isEmpty ||
        deviceId == null ||
        deviceId.isEmpty ||
        cashierId.isEmpty) {
      return null;
    }
    return _TerminalContext(
      outletId: outletId,
      deviceId: deviceId,
      cashierId: cashierId,
      shiftId: session.activeShift?.id,
    );
  }
}

class _TerminalContext {
  const _TerminalContext({
    required this.outletId,
    required this.deviceId,
    required this.cashierId,
    this.shiftId,
  });

  final String outletId;
  final String deviceId;
  final String cashierId;
  final String? shiftId;
}

TerminalLockState? _lockStateFromError(TerminalLockState? state, Object error) {
  if (state == null || error is! ApiException) return state;
  final remaining = error.details['failed_attempts_remaining'];
  final lockoutUntil = error.details['lockout_until'];
  return state.copyWith(
    failedAttemptsRemaining: remaining is int ? remaining : null,
    lockoutUntil: lockoutUntil is String
        ? DateTime.tryParse(lockoutUntil)
        : null,
  );
}

String _messageFor(Object error) {
  if (error is ApiException) {
    return switch (error.code) {
      'INVALID_PIN' => 'PIN tidak valid.',
      'TERMINAL_LOCKED_OUT' => 'Terlalu banyak percobaan PIN. Coba lagi nanti.',
      'TERMINAL_ALREADY_LOCKED' => 'Terminal sudah terkunci.',
      'TERMINAL_NOT_LOCKED' => 'Terminal belum terkunci.',
      'FORBIDDEN' => 'Staff tidak berwenang membuka terminal ini.',
      _ => error.message,
    };
  }
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}
