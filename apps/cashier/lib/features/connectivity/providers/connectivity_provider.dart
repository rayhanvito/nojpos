import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../repositories/connectivity_repository.dart';

final connectivityControllerProvider =
    NotifierProvider<ConnectivityController, ConnectivityState>(
      ConnectivityController.new,
    );

enum ConnectivityStatus { checking, online, degraded, offline }

class ConnectivityState {
  const ConnectivityState({
    required this.status,
    this.isChecking = false,
    this.lastCheckedAt,
    this.errorMessage,
  });

  const ConnectivityState.initial()
    : status = ConnectivityStatus.checking,
      isChecking = true,
      lastCheckedAt = null,
      errorMessage = null;

  final ConnectivityStatus status;
  final bool isChecking;
  final DateTime? lastCheckedAt;
  final String? errorMessage;

  bool get isOnline => status == ConnectivityStatus.online;

  String get label {
    return switch (status) {
      ConnectivityStatus.checking => 'Mengecek koneksi',
      ConnectivityStatus.online => 'Online',
      ConnectivityStatus.degraded => 'Koneksi terbatas',
      ConnectivityStatus.offline => 'Offline',
    };
  }

  String get friendlyMessage {
    if (errorMessage != null && errorMessage!.isNotEmpty) {
      return errorMessage!;
    }
    return switch (status) {
      ConnectivityStatus.checking => 'Sedang memeriksa koneksi server.',
      ConnectivityStatus.online => 'Terhubung ke server NOJPOS.',
      ConnectivityStatus.degraded =>
        'Server merespons, tetapi sesi atau layanan perlu diperiksa.',
      ConnectivityStatus.offline =>
        'Tidak dapat menghubungi server. POS tetap bisa dibuka, tetapi data terbaru mungkin belum tersedia.',
    };
  }

  ConnectivityState copyWith({
    ConnectivityStatus? status,
    bool? isChecking,
    DateTime? lastCheckedAt,
    bool clearLastCheckedAt = false,
    String? errorMessage,
    bool clearError = false,
  }) {
    return ConnectivityState(
      status: status ?? this.status,
      isChecking: isChecking ?? this.isChecking,
      lastCheckedAt: clearLastCheckedAt
          ? null
          : lastCheckedAt ?? this.lastCheckedAt,
      errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
    );
  }
}

class ConnectivityController extends Notifier<ConnectivityState> {
  Timer? _timer;

  @override
  ConnectivityState build() {
    ref.onDispose(() => _timer?.cancel());
    Future<void>.microtask(checkNow);
    _timer = Timer.periodic(
      const Duration(seconds: 45),
      (_) => checkNow(silent: true),
    );
    return const ConnectivityState.initial();
  }

  Future<void> checkNow({bool silent = false}) async {
    final previous = state;
    state = previous.copyWith(
      status: previous.lastCheckedAt == null
          ? ConnectivityStatus.checking
          : previous.status,
      isChecking: !silent,
      clearError: true,
    );

    try {
      await ref.read(connectivityRepositoryProvider).heartbeat();
      state = ConnectivityState(
        status: ConnectivityStatus.online,
        lastCheckedAt: DateTime.now(),
      );
    } on ServerApiException catch (error) {
      final status = error.code == 'NETWORK_ERROR'
          ? ConnectivityStatus.offline
          : ConnectivityStatus.degraded;
      state = ConnectivityState(
        status: status,
        lastCheckedAt: DateTime.now(),
        errorMessage: _messageFor(error, status),
      );
    } on ApiException catch (error) {
      state = ConnectivityState(
        status: ConnectivityStatus.degraded,
        lastCheckedAt: DateTime.now(),
        errorMessage: _messageFor(error, ConnectivityStatus.degraded),
      );
    } catch (_) {
      state = ConnectivityState(
        status: ConnectivityStatus.offline,
        lastCheckedAt: DateTime.now(),
        errorMessage: 'Koneksi ke server belum tersedia.',
      );
    }
  }

  String _messageFor(ApiException error, ConnectivityStatus status) {
    if (status == ConnectivityStatus.offline) {
      return 'Server tidak dapat dijangkau. Periksa jaringan atau alamat API.';
    }
    if (error is UnauthorizedApiException) {
      return 'Sesi login perlu diperbarui sebelum heartbeat bisa berhasil.';
    }
    return error.message;
  }
}
