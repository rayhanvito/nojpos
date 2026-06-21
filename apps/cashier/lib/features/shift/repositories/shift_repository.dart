import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../../core/network/api_client.dart';
import '../../../shared/models/nojpos_models.dart';

final shiftRepositoryProvider = Provider<ShiftRepository>(
  (ref) => ApiShiftRepository(apiClient: ref.watch(apiClientProvider)),
);

abstract interface class ShiftRepository {
  Future<ShiftSession> openShift({
    required String outletId,
    required String deviceId,
    required String cashierId,
    required int openingCash,
  });

  Future<ShiftSession?> currentShift({
    required String outletId,
    required String deviceId,
  });

  Future<ShiftSession> closeShift({
    required String shiftId,
    required int actualCash,
    required String pin,
    String? varianceReason,
  });

  Future<void> cashMovement({
    required String shiftId,
    required String type,
    required int amount,
    String? reason,
  });
}

class ApiShiftRepository implements ShiftRepository {
  const ApiShiftRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<ShiftSession> openShift({
    required String outletId,
    required String deviceId,
    required String cashierId,
    required int openingCash,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/shifts/open',
      data: {
        'outlet_id': outletId,
        'device_id': deviceId,
        'cashier_id': cashierId,
        'opening_cash': openingCash,
      },
      idempotencyKey: const Uuid().v4(),
    );
    return _shiftFromJson(response.data);
  }

  @override
  Future<ShiftSession?> currentShift({
    required String outletId,
    required String deviceId,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/shifts/current',
      queryParameters: {'outlet_id': outletId, 'device_id': deviceId},
    );
    final shift = response.data['shift'];
    return shift == null ? null : _shiftFromJson(shift);
  }

  @override
  Future<ShiftSession> closeShift({
    required String shiftId,
    required int actualCash,
    required String pin,
    String? varianceReason,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/shifts/$shiftId/close',
      data: {
        'actual_cash': actualCash,
        'pin': pin,
        if (varianceReason != null && varianceReason.trim().isNotEmpty)
          'variance_reason': varianceReason.trim(),
      },
      idempotencyKey: const Uuid().v4(),
    );
    return _shiftFromJson(response.data);
  }

  @override
  Future<void> cashMovement({
    required String shiftId,
    required String type,
    required int amount,
    String? reason,
  }) async {
    await _apiClient.post<Map<String, Object?>>(
      '/shifts/$shiftId/cash-movements',
      data: {
        'type': type,
        'amount': amount,
        ...?(reason == null ? null : {'reason': reason}),
      },
      idempotencyKey: const Uuid().v4(),
    );
  }
}

ShiftSession _shiftFromJson(Object? value) {
  final json = _asMap(value);
  final status = (json['status'] as String?) ?? 'open';
  return ShiftSession(
    id: (json['id'] as String?) ?? '',
    cashier: Employee(
      id: (json['cashier_id'] as String?) ?? '',
      name: 'Kasir',
      role: 'cashier',
    ),
    openedAt: DateTime.now(),
    closedAt: status == 'closed' ? DateTime.now() : null,
    openingCash: (json['opening_cash'] as num?)?.toInt() ?? 0,
    closingCash: (json['actual_cash'] as num?)?.toInt() ?? 0,
    expectedCash: (json['expected_cash'] as num?)?.toInt(),
    actualCash: (json['actual_cash'] as num?)?.toInt(),
    cashDifference: (json['cash_difference'] as num?)?.toInt(),
    paymentTotals: [
      for (final total in json['payment_totals'] as List? ?? const [])
        ShiftPaymentTotal.fromJson(total),
    ],
    status: status,
  );
}

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}
