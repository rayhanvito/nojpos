import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../../core/network/api_client.dart';

final storeRepositoryProvider = Provider<StoreRepository>(
  (ref) => ApiStoreRepository(apiClient: ref.watch(apiClientProvider)),
);

abstract interface class StoreRepository {
  Future<StoreState> getStoreState(String outletId);

  Future<StoreState> openStore({
    required String outletId,
    required String authorizationPin,
    String? reason,
    String? idempotencyKey,
  });

  Future<StoreState> closeStore({
    required String outletId,
    required String authorizationPin,
    String? reason,
    String? idempotencyKey,
  });
}

class ApiStoreRepository implements StoreRepository {
  const ApiStoreRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<StoreState> getStoreState(String outletId) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/outlets/$outletId/store-state',
    );
    return StoreState.fromJson(response.data);
  }

  @override
  Future<StoreState> openStore({
    required String outletId,
    required String authorizationPin,
    String? reason,
    String? idempotencyKey,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/outlets/$outletId/store/open',
      data: _payload(authorizationPin: authorizationPin, reason: reason),
      idempotencyKey: idempotencyKey ?? const Uuid().v4(),
    );
    return StoreState.fromJson(response.data);
  }

  @override
  Future<StoreState> closeStore({
    required String outletId,
    required String authorizationPin,
    String? reason,
    String? idempotencyKey,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/outlets/$outletId/store/close',
      data: _payload(authorizationPin: authorizationPin, reason: reason),
      idempotencyKey: idempotencyKey ?? const Uuid().v4(),
    );
    return StoreState.fromJson(response.data);
  }

  Map<String, Object?> _payload({
    required String authorizationPin,
    String? reason,
  }) {
    return {
      'authorization_pin': authorizationPin,
      if (reason != null && reason.trim().isNotEmpty) 'reason': reason.trim(),
    };
  }
}

class StoreState {
  const StoreState({
    required this.outletId,
    required this.outletName,
    required this.status,
    required this.enabled,
    required this.pinRequired,
    required this.blockingShifts,
    this.openedAt,
    this.closedAt,
    this.reason,
    this.serverTime,
  });

  factory StoreState.fromJson(Object? value) {
    final json = _asMap(value);
    final settings = _asMap(json['settings']);
    return StoreState(
      outletId: (json['outlet_id'] as String?) ?? '',
      outletName: (json['outlet_name'] as String?) ?? '',
      status: (json['status'] as String?) ?? 'open',
      openedAt: _parseDate(json['opened_at']),
      closedAt: _parseDate(json['closed_at']),
      reason: json['reason'] as String?,
      enabled: settings['store_open_close_enabled'] as bool? ?? false,
      pinRequired: settings['pin_required_store_open_close'] as bool? ?? true,
      blockingShifts: [
        for (final item in json['blocking_shifts'] as List? ?? const [])
          BlockingShift.fromJson(item),
      ],
      serverTime: _parseDate(json['server_time']),
    );
  }

  final String outletId;
  final String outletName;
  final String status;
  final bool enabled;
  final bool pinRequired;
  final List<BlockingShift> blockingShifts;
  final DateTime? openedAt;
  final DateTime? closedAt;
  final String? reason;
  final DateTime? serverTime;

  bool get isOpen => status == 'open';
  bool get isClosed => status == 'closed';
}

class BlockingShift {
  const BlockingShift({
    required this.shiftId,
    required this.status,
    this.cashierName,
    this.deviceName,
    this.openedAt,
  });

  factory BlockingShift.fromJson(Object? value) {
    final json = _asMap(value);
    final cashier = _asMap(json['cashier']);
    final device = _asMap(json['device']);
    return BlockingShift(
      shiftId: (json['shift_id'] as String?) ?? '',
      status: (json['status'] as String?) ?? 'open',
      cashierName: cashier['name'] as String?,
      deviceName: device['name'] as String?,
      openedAt: _parseDate(json['opened_at']),
    );
  }

  final String shiftId;
  final String status;
  final String? cashierName;
  final String? deviceName;
  final DateTime? openedAt;
}

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}

DateTime? _parseDate(Object? value) {
  if (value is! String || value.isEmpty) return null;
  return DateTime.tryParse(value);
}
