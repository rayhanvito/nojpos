import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';

void main() {
  test('ApiClient reads success envelope and sends auth headers', () async {
    final adapter = _FakeAdapter((options) {
      expect(options.headers['Accept'], 'application/json');
      expect(options.headers['Authorization'], 'Bearer token-123');

      return ResponseBody.fromString(
        jsonEncode({
          'data': {'ok': true},
          'meta': {'page': 1},
        }),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    });
    final client = ApiClient(
      dio: Dio()..httpClientAdapter = adapter,
      tokenReader: () async => 'token-123',
    );

    final response = await client.get<Map<String, dynamic>>('/me');

    expect(response.data['ok'], true);
    expect(response.meta['page'], 1);
  });

  test('ApiClient maps validation envelope to domain exception', () async {
    final client = ApiClient(
      dio: Dio()
        ..httpClientAdapter = _FakeAdapter(
          (_) => ResponseBody.fromString(
            jsonEncode({
              'error': {
                'code': 'INVALID_PIN',
                'message': 'PIN is invalid.',
                'details': {
                  'pin': ['Invalid'],
                },
              },
            }),
            422,
            headers: {
              Headers.contentTypeHeader: [Headers.jsonContentType],
            },
          ),
        ),
    );

    expect(
      () => client.post<Map<String, dynamic>>('/auth/pin-switch', data: {}),
      throwsA(
        isA<ValidationApiException>()
            .having((error) => error.code, 'code', 'INVALID_PIN')
            .having((error) => error.message, 'message', 'PIN is invalid.'),
      ),
    );
  });
}

class _FakeAdapter implements HttpClientAdapter {
  _FakeAdapter(this.handler);

  final ResponseBody Function(RequestOptions options) handler;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return handler(options);
  }

  @override
  void close({bool force = false}) {}
}
