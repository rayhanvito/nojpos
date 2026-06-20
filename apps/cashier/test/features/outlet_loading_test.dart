import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/core/storage/token_storage.dart';
import 'package:nojpos_tablet_ui/features/auth/repositories/auth_repository.dart';

void main() {
  test(
    'ApiAuthRepository should load outlet picker from GET outlets',
    () async {
      final storage = InMemoryTokenStorage();
      final adapter = _OutletRouteAdapter();

      final session =
          await ApiAuthRepository(
            apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
            tokenStorage: storage,
          ).login(
            email: 'owner@test.local',
            password: 'password',
            deviceUuid: 'demo-device',
          );

      expect(adapter.requests, contains('GET /outlets'));
      expect(session.outlets.map((outlet) => outlet.id), [
        'outlet-api-1',
        'outlet-api-2',
      ]);
    },
  );
}

class _OutletRouteAdapter implements HttpClientAdapter {
  final List<String> requests = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final key = '${options.method} ${options.path}';
    requests.add(key);
    if (key == 'POST /auth/login') {
      return _json({
        'data': {
          'token': 'plain-token',
          'user': {
            'id': 'owner-id',
            'name': 'Owner Demo',
            'email': 'owner@test.local',
            'role': 'owner',
          },
          'business': {'id': 'business-id', 'name': 'NojPOS Demo'},
          'device': {'id': 'device-id', 'device_uuid': 'demo-device'},
          'outlets': [
            {'id': 'fallback-outlet', 'name': 'Fallback Outlet'},
          ],
        },
        'meta': {},
      });
    }
    if (key == 'GET /outlets') {
      return _json({
        'data': [
          {'id': 'outlet-api-1', 'name': 'Outlet API 1'},
          {'id': 'outlet-api-2', 'name': 'Outlet API 2'},
        ],
        'meta': {},
      });
    }
    return _json({
      'error': {'code': 'NOT_FOUND', 'message': key, 'details': {}},
    }, status: 404);
  }

  ResponseBody _json(Map<String, Object?> body, {int status = 200}) {
    return ResponseBody.fromString(
      jsonEncode(body),
      status,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
