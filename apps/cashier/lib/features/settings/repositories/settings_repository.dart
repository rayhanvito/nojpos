import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';

final settingsRepositoryProvider = Provider<SettingsRepository>(
  (ref) => ApiSettingsRepository(apiClient: ref.watch(apiClientProvider)),
);

abstract interface class SettingsRepository {
  Future<SettingsAggregate> fetchSettings();

  Future<SettingsAggregate> updateOutletSettings({
    required String outletId,
    required int taxRate,
    required int serviceChargeRate,
    required String idempotencyKey,
  });

  Future<SettingsAggregate> updatePaymentMethod({
    required String configId,
    required bool active,
    required String idempotencyKey,
  });

  Future<SettingsAggregate> updateSecuritySettings({
    required int maxAttempts,
    required int lockoutMinutes,
    required int idleLockTimeoutSeconds,
    required int sessionTimeoutSeconds,
    required Map<String, bool> sensitiveActionPins,
    required String idempotencyKey,
  });
}

class ApiSettingsRepository implements SettingsRepository {
  const ApiSettingsRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<SettingsAggregate> fetchSettings() async {
    final response = await _apiClient.get<Map<String, Object?>>('/settings');
    return SettingsAggregate.fromJson(response.data, response.meta);
  }

  @override
  Future<SettingsAggregate> updateOutletSettings({
    required String outletId,
    required int taxRate,
    required int serviceChargeRate,
    required String idempotencyKey,
  }) async {
    final response = await _apiClient.patch<Map<String, Object?>>(
      '/settings/outlets/$outletId',
      data: {'tax_rate': taxRate, 'service_charge_rate': serviceChargeRate},
      idempotencyKey: idempotencyKey,
    );
    return SettingsAggregate.fromJson(response.data, response.meta);
  }

  @override
  Future<SettingsAggregate> updatePaymentMethod({
    required String configId,
    required bool active,
    required String idempotencyKey,
  }) async {
    final response = await _apiClient.patch<Map<String, Object?>>(
      '/settings/payment-methods/$configId',
      data: {'active': active},
      idempotencyKey: idempotencyKey,
    );
    return SettingsAggregate.fromJson(response.data, response.meta);
  }

  @override
  Future<SettingsAggregate> updateSecuritySettings({
    required int maxAttempts,
    required int lockoutMinutes,
    required int idleLockTimeoutSeconds,
    required int sessionTimeoutSeconds,
    required Map<String, bool> sensitiveActionPins,
    required String idempotencyKey,
  }) async {
    final response = await _apiClient.patch<Map<String, Object?>>(
      '/settings/security',
      data: {
        'pin_policy': {
          'max_attempts': maxAttempts,
          'lockout_minutes': lockoutMinutes,
        },
        'terminal_policy': {
          'idle_lock_timeout_seconds': idleLockTimeoutSeconds,
          'session_timeout_seconds': sessionTimeoutSeconds,
        },
        'sensitive_actions': {
          for (final entry in sensitiveActionPins.entries)
            entry.key: {'requires_pin': entry.value},
        },
      },
      idempotencyKey: idempotencyKey,
    );
    return SettingsAggregate.fromJson(response.data, response.meta);
  }
}

class SettingsAggregate {
  const SettingsAggregate({
    required this.business,
    required this.outlets,
    required this.paymentMethods,
    required this.security,
    required this.permissions,
    required this.configVersion,
    required this.scope,
    this.serverTimestamp,
  });

  factory SettingsAggregate.fromJson(
    Map<String, Object?> json,
    Map<String, Object?> meta,
  ) {
    return SettingsAggregate(
      business: BusinessSettings.fromJson(_asMap(json['business'])),
      outlets: [
        for (final value in json['outlets'] as List? ?? const [])
          OutletSettings.fromJson(_asMap(value)),
      ],
      paymentMethods: [
        for (final value in json['payment_methods'] as List? ?? const [])
          PaymentMethodSetting.fromJson(_asMap(value)),
      ],
      security: SecuritySettings.fromJson(_asMap(json['security'])),
      permissions: SettingsPermissions.fromJson(_asMap(json['permissions'])),
      configVersion: (meta['config_version'] as String?) ?? '',
      serverTimestamp: DateTime.tryParse(
        (meta['server_timestamp'] as String?) ?? '',
      ),
      scope: (meta['scope'] as String?) ?? 'pos_safe',
    );
  }

  final BusinessSettings business;
  final List<OutletSettings> outlets;
  final List<PaymentMethodSetting> paymentMethods;
  final SecuritySettings security;
  final SettingsPermissions permissions;
  final String configVersion;
  final String scope;
  final DateTime? serverTimestamp;

  List<PaymentMethodSetting> activePaymentMethodsForOutlet(String outletId) {
    return paymentMethods
        .where((method) => method.active)
        .where(
          (method) => method.outletId == null || method.outletId == outletId,
        )
        .toList(growable: false);
  }
}

class BusinessSettings {
  const BusinessSettings({
    required this.id,
    required this.name,
    required this.timezone,
    required this.currency,
    this.defaultOutletId,
  });

  factory BusinessSettings.fromJson(Map<String, Object?> json) {
    return BusinessSettings(
      id: (json['id'] as String?) ?? '',
      name: (json['name'] as String?) ?? '',
      timezone: (json['timezone'] as String?) ?? 'Asia/Jakarta',
      currency: (json['currency'] as String?) ?? 'IDR',
      defaultOutletId: json['default_outlet_id'] as String?,
    );
  }

  final String id;
  final String name;
  final String timezone;
  final String currency;
  final String? defaultOutletId;
}

class OutletSettings {
  const OutletSettings({
    required this.id,
    required this.name,
    required this.timezone,
    required this.serviceChargeRate,
    required this.taxRate,
    required this.roundingPolicy,
    required this.receiptPaperWidth,
    required this.receiptHeader,
    required this.receiptFooter,
    required this.cashOutLimit,
    required this.storeOpenCloseEnabled,
  });

  factory OutletSettings.fromJson(Map<String, Object?> json) {
    final transaction = _asMap(json['transaction_config']);
    final receipt = _asMap(json['receipt_config']);
    final operational = _asMap(json['operational_config']);
    return OutletSettings(
      id: (json['id'] as String?) ?? '',
      name: (json['name'] as String?) ?? '',
      timezone: (json['timezone'] as String?) ?? 'Asia/Jakarta',
      serviceChargeRate: _int(transaction['service_charge_rate']),
      taxRate: _int(transaction['tax_rate']),
      roundingPolicy: (transaction['rounding_policy'] as String?) ?? 'none',
      receiptPaperWidth: (receipt['paper_width'] as String?) ?? '58mm',
      receiptHeader:
          (receipt['header_name'] as String?) ??
          (receipt['header'] as String?) ??
          '',
      receiptFooter:
          (receipt['footer_note'] as String?) ??
          (receipt['footer'] as String?) ??
          '',
      cashOutLimit: _int(operational['cash_out_limit']),
      storeOpenCloseEnabled:
          operational['store_open_close_enabled'] as bool? ?? false,
    );
  }

  final String id;
  final String name;
  final String timezone;
  final int serviceChargeRate;
  final int taxRate;
  final String roundingPolicy;
  final String receiptPaperWidth;
  final String receiptHeader;
  final String receiptFooter;
  final int cashOutLimit;
  final bool storeOpenCloseEnabled;
}

class PaymentMethodSetting {
  const PaymentMethodSetting({
    required this.id,
    required this.method,
    required this.active,
    required this.isCash,
    this.outletId,
  });

  factory PaymentMethodSetting.fromJson(Map<String, Object?> json) {
    return PaymentMethodSetting(
      id: (json['id'] as String?) ?? '',
      method: (json['method'] as String?) ?? '',
      active: (json['active'] as bool?) ?? (json['enabled'] as bool? ?? false),
      isCash: json['is_cash'] as bool? ?? false,
      outletId: json['outlet_id'] as String?,
    );
  }

  final String id;
  final String method;
  final bool active;
  final bool isCash;
  final String? outletId;
}

class SecuritySettings {
  const SecuritySettings({
    required this.pinPolicy,
    required this.terminalPolicy,
    required this.sensitiveActions,
  });

  factory SecuritySettings.fromJson(Map<String, Object?> json) {
    final actions = _asMap(json['sensitive_actions']);
    return SecuritySettings(
      pinPolicy: PinPolicy.fromJson(_asMap(json['pin_policy'])),
      terminalPolicy: TerminalPolicy.fromJson(_asMap(json['terminal_policy'])),
      sensitiveActions: {
        for (final entry in actions.entries)
          entry.key: SensitiveActionPolicy.fromJson(_asMap(entry.value)),
      },
    );
  }

  final PinPolicy pinPolicy;
  final TerminalPolicy terminalPolicy;
  final Map<String, SensitiveActionPolicy> sensitiveActions;

  bool requiresPin(String action) {
    return sensitiveActions[action]?.requiresPin ?? true;
  }
}

class PinPolicy {
  const PinPolicy({
    required this.scope,
    required this.maxAttempts,
    required this.lockoutMinutes,
    required this.ownerAdminCanClear,
  });

  factory PinPolicy.fromJson(Map<String, Object?> json) {
    return PinPolicy(
      scope: (json['scope'] as String?) ?? 'staff_business',
      maxAttempts: _int(json['max_attempts'], fallback: 5),
      lockoutMinutes: _int(json['lockout_minutes'], fallback: 15),
      ownerAdminCanClear: json['owner_admin_can_clear'] as bool? ?? true,
    );
  }

  final String scope;
  final int maxAttempts;
  final int lockoutMinutes;
  final bool ownerAdminCanClear;
}

class TerminalPolicy {
  const TerminalPolicy({
    required this.idleLockTimeoutSeconds,
    required this.sessionTimeoutSeconds,
  });

  factory TerminalPolicy.fromJson(Map<String, Object?> json) {
    return TerminalPolicy(
      idleLockTimeoutSeconds: _int(
        json['idle_lock_timeout_seconds'],
        fallback: 180,
      ),
      sessionTimeoutSeconds: _int(
        json['session_timeout_seconds'],
        fallback: 900,
      ),
    );
  }

  final int idleLockTimeoutSeconds;
  final int sessionTimeoutSeconds;
}

class SensitiveActionPolicy {
  const SensitiveActionPolicy({required this.requiresPin, required this.roles});

  factory SensitiveActionPolicy.fromJson(Map<String, Object?> json) {
    return SensitiveActionPolicy(
      requiresPin: json['requires_pin'] as bool? ?? true,
      roles: [
        for (final value in json['roles'] as List? ?? const [])
          value.toString(),
      ],
    );
  }

  final bool requiresPin;
  final List<String> roles;
}

class SettingsPermissions {
  const SettingsPermissions({
    required this.currentUserRole,
    required this.canUpdateSettings,
    required this.canUpdateSecuritySettings,
  });

  factory SettingsPermissions.fromJson(Map<String, Object?> json) {
    final currentUser = _asMap(json['current_user']);
    return SettingsPermissions(
      currentUserRole: (currentUser['role'] as String?) ?? 'cashier',
      canUpdateSettings: currentUser['can_update_settings'] as bool? ?? false,
      canUpdateSecuritySettings:
          currentUser['can_update_security_settings'] as bool? ?? false,
    );
  }

  final String currentUserRole;
  final bool canUpdateSettings;
  final bool canUpdateSecuritySettings;
}

int _int(Object? value, {int fallback = 0}) =>
    (value as num?)?.toInt() ?? fallback;

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}
