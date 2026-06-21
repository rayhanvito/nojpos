import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';

final terminalLockRepositoryProvider = Provider<TerminalLockRepository>(
  (ref) => ApiTerminalLockRepository(apiClient: ref.watch(apiClientProvider)),
);

abstract interface class TerminalLockRepository {
  Future<TerminalLockState> getLockState({
    required String outletId,
    required String deviceId,
    String? staffId,
  });

  Future<TerminalLockState> lockTerminal({
    required String outletId,
    required String deviceId,
    required TerminalLockReason reason,
    String? cashierId,
    String? shiftId,
  });

  Future<TerminalUnlockResponse> unlockTerminal({
    required String outletId,
    required String deviceId,
    required String staffId,
    required String pin,
    String mode,
  });
}

class ApiTerminalLockRepository implements TerminalLockRepository {
  const ApiTerminalLockRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<TerminalLockState> getLockState({
    required String outletId,
    required String deviceId,
    String? staffId,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/terminal/lock-state',
      queryParameters: {
        'outlet_id': outletId,
        'device_id': deviceId,
        if (staffId != null && staffId.isNotEmpty) 'staff_id': staffId,
      },
    );
    return TerminalLockState.fromJson(response.data);
  }

  @override
  Future<TerminalLockState> lockTerminal({
    required String outletId,
    required String deviceId,
    required TerminalLockReason reason,
    String? cashierId,
    String? shiftId,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/terminal/lock',
      data: {
        'outlet_id': outletId,
        'device_id': deviceId,
        'reason': reason.value,
        if (cashierId != null && cashierId.isNotEmpty) 'cashier_id': cashierId,
        if (shiftId != null && shiftId.isNotEmpty) 'shift_id': shiftId,
      },
    );
    return TerminalLockState.fromJson(response.data);
  }

  @override
  Future<TerminalUnlockResponse> unlockTerminal({
    required String outletId,
    required String deviceId,
    required String staffId,
    required String pin,
    String mode = 'resume_current',
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/terminal/unlock',
      data: {
        'outlet_id': outletId,
        'device_id': deviceId,
        'staff_id': staffId,
        'pin': pin,
        'mode': mode,
      },
    );
    return TerminalUnlockResponse.fromJson(response.data);
  }
}

enum TerminalLockReason {
  manual('manual'),
  idleTimeout('idle_timeout'),
  sessionTimeout('session_timeout');

  const TerminalLockReason(this.value);
  final String value;
}

class TerminalLockState {
  const TerminalLockState({
    required this.locked,
    required this.outletId,
    required this.deviceId,
    required this.failedAttemptsRemaining,
    required this.idleTimeoutSeconds,
    required this.sessionTimeoutSeconds,
    this.lockedAt,
    this.lockReason,
    this.cashier,
    this.shift,
    this.lockoutUntil,
    this.unlockedAt,
    this.unlockedBy,
    this.serverTime,
  });

  factory TerminalLockState.fromJson(Object? value) {
    final json = _asMap(value);
    return TerminalLockState(
      locked: json['locked'] as bool? ?? false,
      outletId: json['outlet_id'] as String? ?? '',
      deviceId: json['device_id'] as String? ?? '',
      lockedAt: _date(json['locked_at']),
      lockReason: json['lock_reason'] as String?,
      cashier: TerminalStaffSummary.fromJsonOrNull(json['cashier']),
      shift: TerminalShiftSummary.fromJsonOrNull(json['shift']),
      failedAttemptsRemaining: _int(json['failed_attempts_remaining'], 5),
      lockoutUntil: _date(json['lockout_until']),
      idleTimeoutSeconds: _int(json['idle_timeout_seconds'], 180),
      sessionTimeoutSeconds: _int(json['session_timeout_seconds'], 900),
      unlockedAt: _date(json['unlocked_at']),
      unlockedBy: TerminalStaffSummary.fromJsonOrNull(json['unlocked_by']),
      serverTime: _date(json['server_time']),
    );
  }

  final bool locked;
  final String outletId;
  final String deviceId;
  final DateTime? lockedAt;
  final String? lockReason;
  final TerminalStaffSummary? cashier;
  final TerminalShiftSummary? shift;
  final int failedAttemptsRemaining;
  final DateTime? lockoutUntil;
  final int idleTimeoutSeconds;
  final int sessionTimeoutSeconds;
  final DateTime? unlockedAt;
  final TerminalStaffSummary? unlockedBy;
  final DateTime? serverTime;

  TerminalLockState copyWith({
    bool? locked,
    DateTime? lockedAt,
    String? lockReason,
    int? failedAttemptsRemaining,
    DateTime? lockoutUntil,
  }) {
    return TerminalLockState(
      locked: locked ?? this.locked,
      outletId: outletId,
      deviceId: deviceId,
      lockedAt: lockedAt ?? this.lockedAt,
      lockReason: lockReason ?? this.lockReason,
      cashier: cashier,
      shift: shift,
      failedAttemptsRemaining:
          failedAttemptsRemaining ?? this.failedAttemptsRemaining,
      lockoutUntil: lockoutUntil ?? this.lockoutUntil,
      idleTimeoutSeconds: idleTimeoutSeconds,
      sessionTimeoutSeconds: sessionTimeoutSeconds,
      unlockedAt: unlockedAt,
      unlockedBy: unlockedBy,
      serverTime: serverTime,
    );
  }
}

class TerminalUnlockResponse extends TerminalLockState {
  const TerminalUnlockResponse({
    required super.locked,
    required super.outletId,
    required super.deviceId,
    required super.failedAttemptsRemaining,
    required super.idleTimeoutSeconds,
    required super.sessionTimeoutSeconds,
    required this.mode,
    required this.sameStaff,
    required this.handoverRequired,
    required this.shiftPreserved,
    required this.cartPreserved,
    super.lockedAt,
    super.lockReason,
    super.cashier,
    super.shift,
    super.lockoutUntil,
    super.unlockedAt,
    super.unlockedBy,
    super.serverTime,
  });

  factory TerminalUnlockResponse.fromJson(Object? value) {
    final base = TerminalLockState.fromJson(value);
    final json = _asMap(value);
    return TerminalUnlockResponse(
      locked: base.locked,
      outletId: base.outletId,
      deviceId: base.deviceId,
      lockedAt: base.lockedAt,
      lockReason: base.lockReason,
      cashier: base.cashier,
      shift: base.shift,
      failedAttemptsRemaining: base.failedAttemptsRemaining,
      lockoutUntil: base.lockoutUntil,
      idleTimeoutSeconds: base.idleTimeoutSeconds,
      sessionTimeoutSeconds: base.sessionTimeoutSeconds,
      unlockedAt: base.unlockedAt,
      unlockedBy: base.unlockedBy,
      serverTime: base.serverTime,
      mode: json['mode'] as String? ?? 'resume_current',
      sameStaff: json['same_staff'] as bool? ?? true,
      handoverRequired: json['handover_required'] as bool? ?? false,
      shiftPreserved: json['shift_preserved'] as bool? ?? true,
      cartPreserved: json['cart_preserved'] as bool? ?? true,
    );
  }

  final String mode;
  final bool sameStaff;
  final bool handoverRequired;
  final bool shiftPreserved;
  final bool cartPreserved;
}

class TerminalStaffSummary {
  const TerminalStaffSummary({
    required this.id,
    required this.name,
    required this.role,
  });

  static TerminalStaffSummary? fromJsonOrNull(Object? value) {
    final json = _asMap(value);
    if (json.isEmpty) return null;
    return TerminalStaffSummary(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      role: json['role'] as String? ?? '',
    );
  }

  final String id;
  final String name;
  final String role;
}

class TerminalShiftSummary {
  const TerminalShiftSummary({required this.id, required this.preserved});

  static TerminalShiftSummary? fromJsonOrNull(Object? value) {
    final json = _asMap(value);
    if (json.isEmpty) return null;
    return TerminalShiftSummary(
      id: json['id'] as String? ?? '',
      preserved: json['preserved'] as bool? ?? false,
    );
  }

  final String id;
  final bool preserved;
}

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}

DateTime? _date(Object? value) {
  if (value is! String || value.isEmpty) return null;
  return DateTime.tryParse(value);
}

int _int(Object? value, int fallback) {
  if (value is int) return value;
  if (value is num) return value.toInt();
  return int.tryParse(value?.toString() ?? '') ?? fallback;
}
