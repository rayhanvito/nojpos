import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../storage/token_storage.dart';

const defaultApiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://10.0.2.2:8000/api/v1',
);

final tokenStorageProvider = Provider<TokenStorage>(
  (ref) => const SecureTokenStorage(),
);

final apiClientProvider = Provider<ApiClient>((ref) {
  final storage = ref.watch(tokenStorageProvider);
  return ApiClient(tokenReader: storage.readToken);
});

class ApiClient {
  ApiClient({
    Dio? dio,
    Future<String?> Function()? tokenReader,
    String baseUrl = defaultApiBaseUrl,
  }) : _tokenReader = tokenReader,
       _dio =
           dio ??
           Dio(
             BaseOptions(
               baseUrl: baseUrl,
               headers: const {'Accept': 'application/json'},
               responseType: ResponseType.json,
             ),
           ) {
    _dio.options.baseUrl = baseUrl;
    _dio.options.headers['Accept'] = 'application/json';
  }

  final Dio _dio;
  final Future<String?> Function()? _tokenReader;

  Future<ApiEnvelope<T>> get<T>(
    String path, {
    Map<String, Object?>? queryParameters,
  }) {
    return _request<T>(
      () => _dio.get<Object?>(
        path,
        queryParameters: queryParameters,
        options: Options(headers: _authHeaders),
      ),
    );
  }

  Future<ApiEnvelope<T>> post<T>(
    String path, {
    Object? data,
    Map<String, Object?>? queryParameters,
    String? idempotencyKey,
  }) {
    return _request<T>(
      () => _dio.post<Object?>(
        path,
        data: data,
        queryParameters: queryParameters,
        options: Options(
          headers: {
            ..._authHeaders,
            ...?(idempotencyKey == null
                ? null
                : {'Idempotency-Key': idempotencyKey}),
          },
        ),
      ),
    );
  }

  Future<ApiEnvelope<T>> put<T>(
    String path, {
    Object? data,
    Map<String, Object?>? queryParameters,
  }) {
    return _request<T>(
      () => _dio.put<Object?>(
        path,
        data: data,
        queryParameters: queryParameters,
        options: Options(headers: _authHeaders),
      ),
    );
  }

  Map<String, Object?> get _authHeaders => {'Accept': 'application/json'};

  Future<ApiEnvelope<T>> _request<T>(
    Future<Response<Object?>> Function() send,
  ) async {
    final token = await _tokenReader?.call();
    if (token != null && token.isNotEmpty) {
      _dio.options.headers['Authorization'] = 'Bearer $token';
    } else {
      _dio.options.headers.remove('Authorization');
    }

    try {
      final response = await send();
      return _parseEnvelope<T>(response.data);
    } on DioException catch (error) {
      final response = error.response;
      if (response != null) {
        throw _exceptionFor(response.statusCode ?? 500, response.data);
      }
      throw ServerApiException(
        code: 'NETWORK_ERROR',
        message: error.message ?? 'Network request failed.',
      );
    }
  }

  ApiEnvelope<T> _parseEnvelope<T>(Object? body) {
    final json = _asMap(body);
    return ApiEnvelope<T>(
      data: json['data'] as T,
      meta: _asMap(json['meta'] ?? const <String, Object?>{}),
    );
  }

  ApiException _exceptionFor(int statusCode, Object? body) {
    final json = _asMap(body);
    final error = _asMap(json['error'] ?? const <String, Object?>{});
    final code = (error['code'] as String?) ?? 'API_ERROR';
    final message = (error['message'] as String?) ?? 'Request failed.';
    final details = _asMap(error['details'] ?? const <String, Object?>{});

    return switch (statusCode) {
      401 => UnauthorizedApiException(
        code: code,
        message: message,
        details: details,
      ),
      403 => ForbiddenApiException(
        code: code,
        message: message,
        details: details,
      ),
      404 => NotFoundApiException(
        code: code,
        message: message,
        details: details,
      ),
      409 => ConflictApiException(
        code: code,
        message: message,
        details: details,
      ),
      422 => ValidationApiException(
        code: code,
        message: message,
        details: details,
      ),
      _ => ServerApiException(code: code, message: message, details: details),
    };
  }
}

class ApiEnvelope<T> {
  const ApiEnvelope({required this.data, required this.meta});

  final T data;
  final Map<String, Object?> meta;
}

sealed class ApiException implements Exception {
  const ApiException({
    required this.code,
    required this.message,
    this.details = const {},
  });

  final String code;
  final String message;
  final Map<String, Object?> details;

  @override
  String toString() => '$runtimeType($code): $message';
}

class UnauthorizedApiException extends ApiException {
  const UnauthorizedApiException({
    required super.code,
    required super.message,
    super.details,
  });
}

class ForbiddenApiException extends ApiException {
  const ForbiddenApiException({
    required super.code,
    required super.message,
    super.details,
  });
}

class NotFoundApiException extends ApiException {
  const NotFoundApiException({
    required super.code,
    required super.message,
    super.details,
  });
}

class ConflictApiException extends ApiException {
  const ConflictApiException({
    required super.code,
    required super.message,
    super.details,
  });
}

class ValidationApiException extends ApiException {
  const ValidationApiException({
    required super.code,
    required super.message,
    super.details,
  });
}

class ServerApiException extends ApiException {
  const ServerApiException({
    required super.code,
    required super.message,
    super.details,
  });
}

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}
