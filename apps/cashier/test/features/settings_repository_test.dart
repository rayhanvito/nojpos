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

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
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
