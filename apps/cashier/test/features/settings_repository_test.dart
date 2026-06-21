import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/features/settings/repositories/settings_repository.dart';
import 'package:nojpos_tablet_ui/features/settings/widgets/settings_summary_panel.dart';

void main() {
  test('ApiSettingsRepository parses settings aggregate', () async {
    final repository = ApiSettingsRepository(
      apiClient: ApiClient(
        dio: Dio()
          ..httpClientAdapter = _Adapter({
            'GET /settings': _settingsEnvelope(canUpdate: true, scope: 'admin'),
          }),
      ),
    );

    final settings = await repository.fetchSettings();

    expect(settings.business.name, 'NojPOS Demo');
    expect(settings.business.currency, 'IDR');
    expect(settings.outlets.single.name, 'Outlet Demo');
    expect(settings.outlets.single.taxRate, 11);
    expect(settings.outlets.single.serviceChargeRate, 5);
    expect(settings.paymentMethods.single.method, 'cash');
    expect(settings.paymentMethods.single.isCash, true);
    expect(settings.security.pinPolicy.maxAttempts, 5);
    expect(settings.security.terminalPolicy.idleLockTimeoutSeconds, 180);
    expect(settings.security.requiresPin('void'), true);
    expect(settings.permissions.canUpdateSettings, true);
    expect(settings.configVersion, 'cfg-1');
    expect(settings.scope, 'admin');
  });

  test(
    'ApiSettingsRepository updates outlet, payment method, and security with idempotency',
    () async {
      final adapter = _Adapter({
        'PATCH /settings/outlets/outlet-id': _settingsEnvelope(
          canUpdate: true,
          scope: 'admin',
        ),
        'PATCH /settings/payment-methods/payment-id': _settingsEnvelope(
          canUpdate: true,
          scope: 'admin',
        ),
        'PATCH /settings/security': _settingsEnvelope(
          canUpdate: true,
          scope: 'admin',
        ),
      });
      final repository = ApiSettingsRepository(
        apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
      );

      await repository.updateOutletSettings(
        outletId: 'outlet-id',
        taxRate: 10,
        serviceChargeRate: 3,
        idempotencyKey: 'outlet-key',
      );
      await repository.updatePaymentMethod(
        configId: 'payment-id',
        active: false,
        idempotencyKey: 'payment-key',
      );
      await repository.updateSecuritySettings(
        maxAttempts: 4,
        lockoutMinutes: 10,
        idleLockTimeoutSeconds: 120,
        sessionTimeoutSeconds: 600,
        sensitiveActionPins: const {'refund': true},
        idempotencyKey: 'security-key',
      );

      expect(adapter.idempotencyKeys, [
        'outlet-key',
        'payment-key',
        'security-key',
      ]);
      expect(adapter.payloads[0]['tax_rate'], 10);
      expect(adapter.payloads[0]['service_charge_rate'], 3);
      expect(adapter.payloads[1]['active'], false);
      final security = adapter.payloads[2];
      expect((security['pin_policy'] as Map)['max_attempts'], 4);
      expect((security['sensitive_actions'] as Map)['refund'], {
        'requires_pin': true,
      });
    },
  );

  testWidgets('cashier settings summary does not expose update action', (
    tester,
  ) async {
    final settings = SettingsAggregate.fromJson(
      _settingsEnvelope(canUpdate: false, scope: 'pos_safe')['data']!
          as Map<String, Object?>,
      _settingsEnvelope(canUpdate: false, scope: 'pos_safe')['meta']!
          as Map<String, Object?>,
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(body: SettingsSummaryPanel(settings: settings)),
      ),
    );

    expect(find.text('Kasir: hanya lihat'), findsOneWidget);
    expect(
      find.text(
        'Mode kasir: hanya lihat. Perubahan pengaturan hanya untuk owner/admin.',
      ),
      findsOneWidget,
    );
    expect(find.text('Simpan perubahan'), findsNothing);
  });
}

Map<String, Object?> _settingsEnvelope({
  required bool canUpdate,
  required String scope,
}) {
  return {
    'data': {
      'business': {
        'id': 'business-id',
        'name': 'NojPOS Demo',
        'timezone': 'Asia/Jakarta',
        'currency': 'IDR',
        'default_outlet_id': 'outlet-id',
      },
      'outlets': [
        {
          'id': 'outlet-id',
          'business_id': 'business-id',
          'name': 'Outlet Demo',
          'timezone': 'Asia/Jakarta',
          'transaction_config': {
            'service_charge_rate': 5,
            'tax_rate': 11,
            'rounding_policy': 'none',
          },
          'receipt_config': {
            'paper_width': '58mm',
            'header_name': 'NojPOS Demo',
            'footer_note': 'Terima kasih',
          },
          'operational_config': {
            'cash_out_limit': 500000,
            'store_open_close_enabled': false,
          },
        },
      ],
      'payment_methods': [
        {
          'id': 'payment-id',
          'business_id': 'business-id',
          'outlet_id': 'outlet-id',
          'method': 'cash',
          'enabled': true,
          'active': true,
          'is_cash': true,
          'config': {},
        },
      ],
      'security': {
        'id': 'security-id',
        'pin_policy': {
          'scope': 'staff_business',
          'max_attempts': 5,
          'lockout_minutes': 15,
          'owner_admin_can_clear': true,
        },
        'terminal_policy': {
          'idle_lock_timeout_seconds': 180,
          'session_timeout_seconds': 900,
        },
        'sensitive_actions': {
          'void': {
            'requires_pin': true,
            'roles': ['cashier_with_approval', 'admin', 'owner'],
          },
          'refund': {
            'requires_pin': true,
            'roles': ['admin', 'owner'],
          },
          'close_shift': {
            'requires_pin': true,
            'roles': ['cashier', 'admin', 'owner'],
          },
        },
      },
      'permissions': {
        'current_user': {
          'id': canUpdate ? 'owner-id' : 'cashier-id',
          'role': canUpdate ? 'owner' : 'cashier',
          'can_view_settings': true,
          'can_update_settings': canUpdate,
          'can_update_security_settings': canUpdate,
        },
      },
    },
    'meta': {
      'config_version': 'cfg-1',
      'server_timestamp': '2026-06-20T00:00:00Z',
      'scope': scope,
    },
  };
}

class _Adapter implements HttpClientAdapter {
  _Adapter(this.responses);

  final Map<String, Map<String, Object?>> responses;
  final List<String> idempotencyKeys = [];
  final List<Map<String, Object?>> payloads = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final idempotencyKey = options.headers['Idempotency-Key'];
    if (idempotencyKey is String) idempotencyKeys.add(idempotencyKey);
    if (requestStream != null) {
      final body = await utf8.decodeStream(requestStream);
      if (body.isNotEmpty) {
        payloads.add(jsonDecode(body) as Map<String, Object?>);
      }
    }
    final key = '${options.method} ${options.path}';
    final response = responses[key];
    if (response == null) {
      return ResponseBody.fromString(
        jsonEncode({
          'error': {'code': 'NOT_FOUND', 'message': key, 'details': {}},
        }),
        404,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }

    return ResponseBody.fromString(
      jsonEncode(response),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
