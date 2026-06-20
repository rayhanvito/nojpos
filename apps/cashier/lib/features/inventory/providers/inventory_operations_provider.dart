import 'dart:math';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../repositories/inventory_repository.dart';

final inventoryOperationsProvider =
    NotifierProvider<InventoryOperationsController, InventoryOperationsState>(
      InventoryOperationsController.new,
    );

class InventoryOperationsState {
  const InventoryOperationsState({
    this.isBusy = false,
    this.errorMessage,
    this.successMessage,
    this.movements,
    this.inTransit,
  });

  final bool isBusy;
  final String? errorMessage;
  final String? successMessage;
  final InventoryMovementHistory? movements;
  final InventoryTransferList? inTransit;

  InventoryOperationsState copyWith({
    bool? isBusy,
    String? errorMessage,
    String? successMessage,
    InventoryMovementHistory? movements,
    InventoryTransferList? inTransit,
    bool clearError = false,
    bool clearSuccess = false,
  }) {
    return InventoryOperationsState(
      isBusy: isBusy ?? this.isBusy,
      errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
      successMessage: clearSuccess
          ? null
          : successMessage ?? this.successMessage,
      movements: movements ?? this.movements,
      inTransit: inTransit ?? this.inTransit,
    );
  }
}

class InventoryOperationsController extends Notifier<InventoryOperationsState> {
  @override
  InventoryOperationsState build() => const InventoryOperationsState();

  InventoryRepository get _repository => ref.read(inventoryRepositoryProvider);

  Future<void> loadMovements({
    String? outletId,
    String? productId,
    String? type,
  }) async {
    state = state.copyWith(isBusy: true, clearError: true, clearSuccess: true);
    try {
      final movements = await _repository.getMovements(
        outletId: outletId,
        productId: productId,
        type: type,
      );
      state = state.copyWith(isBusy: false, movements: movements);
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
    }
  }

  Future<void> loadInTransit({String? outletId}) async {
    state = state.copyWith(isBusy: true, clearError: true, clearSuccess: true);
    try {
      final transfers = await _repository.listInTransitTransfers(
        outletId: outletId,
      );
      state = state.copyWith(isBusy: false, inTransit: transfers);
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
    }
  }

  Future<InventoryCountResult?> createCount({
    required String outletId,
    required String productId,
    required int countedQuantity,
    String? reason,
    String? notes,
  }) async {
    state = state.copyWith(isBusy: true, clearError: true, clearSuccess: true);
    try {
      final result = await _repository.createCount(
        outletId: outletId,
        notes: notes,
        idempotencyKey: _newUuid(),
        lines: [
          InventoryCountLineDraft(
            productId: productId,
            countedQuantity: countedQuantity,
            reason: reason,
          ),
        ],
      );
      state = state.copyWith(
        isBusy: false,
        successMessage: 'Stok opname tersimpan.',
      );
      return result;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return null;
    }
  }

  Future<InventoryWasteResult?> createWaste({
    required String outletId,
    required String productId,
    required int quantity,
    required String reason,
  }) async {
    state = state.copyWith(isBusy: true, clearError: true, clearSuccess: true);
    try {
      final result = await _repository.createWaste(
        outletId: outletId,
        productId: productId,
        quantity: quantity,
        reason: reason,
        idempotencyKey: _newUuid(),
      );
      state = state.copyWith(
        isBusy: false,
        successMessage: 'Stok terbuang tersimpan.',
      );
      return result;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return null;
    }
  }

  Future<InventoryTransferResult?> createTransfer({
    required String sourceOutletId,
    required String destinationOutletId,
    required String productId,
    required int quantity,
    String? notes,
  }) async {
    state = state.copyWith(isBusy: true, clearError: true, clearSuccess: true);
    try {
      final result = await _repository.createTransfer(
        sourceOutletId: sourceOutletId,
        destinationOutletId: destinationOutletId,
        notes: notes,
        idempotencyKey: _newUuid(),
        lines: [
          InventoryTransferLineDraft(productId: productId, quantity: quantity),
        ],
      );
      state = state.copyWith(isBusy: false, successMessage: 'Transfer dibuat.');
      return result;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return null;
    }
  }

  Future<InventoryTransferResult?> sendTransfer(String transferId) {
    return _transferAction(
      () => _repository.sendTransfer(
        transferId: transferId,
        idempotencyKey: _newUuid(),
      ),
      'Transfer dikirim.',
    );
  }

  Future<InventoryTransferResult?> receiveTransfer(String transferId) {
    return _transferAction(
      () => _repository.receiveTransfer(
        transferId: transferId,
        idempotencyKey: _newUuid(),
      ),
      'Transfer diterima.',
    );
  }

  Future<InventoryTransferResult?> cancelTransfer(String transferId) {
    return _transferAction(
      () => _repository.cancelTransfer(
        transferId: transferId,
        idempotencyKey: _newUuid(),
      ),
      'Transfer dibatalkan.',
    );
  }

  Future<InventoryTransferResult?> _transferAction(
    Future<InventoryTransferResult> Function() action,
    String successMessage,
  ) async {
    state = state.copyWith(isBusy: true, clearError: true, clearSuccess: true);
    try {
      final result = await action();
      state = state.copyWith(isBusy: false, successMessage: successMessage);
      return result;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return null;
    }
  }
}

String _messageFor(Object error) {
  if (error is ApiException) {
    return switch (error.code) {
      'FORBIDDEN' => 'Akses inventory ditolak.',
      'VALIDATION_ERROR' => 'Periksa kembali data inventory.',
      'STOCK_INSUFFICIENT' => 'Stok tidak mencukupi.',
      'TRANSFER_STATE_INVALID' => 'Status transfer tidak valid.',
      _ => error.message,
    };
  }
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}

String _newUuid() {
  final random = Random.secure();
  final bytes = List<int>.generate(16, (_) => random.nextInt(256));
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  final hex = bytes
      .map((byte) => byte.toRadixString(16).padLeft(2, '0'))
      .join();
  return '${hex.substring(0, 8)}-${hex.substring(8, 12)}-${hex.substring(12, 16)}-${hex.substring(16, 20)}-${hex.substring(20)}';
}
