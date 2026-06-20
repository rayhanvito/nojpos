import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../../core/network/api_client.dart';
import '../../../core/storage/token_storage.dart';
import '../../../shared/models/nojpos_models.dart';

final authRepositoryProvider = Provider<AuthRepository>(
  (ref) => ApiAuthRepository(
    apiClient: ref.watch(apiClientProvider),
    tokenStorage: ref.watch(tokenStorageProvider),
  ),
);

abstract interface class AuthRepository {
  Future<AuthSession?> restoreSession();

  Future<AuthSession> login({
    required String email,
    required String password,
    required String deviceUuid,
  });

  Future<AuthSession> me();

  Future<List<Outlet>> listOutlets();

  Future<Employee> pinSwitch({
    required String pin,
    required String deviceId,
    required String outletId,
  });

  Future<void> selectOutlet(Outlet outlet);

  Future<void> logout();
}

class AuthSession {
  const AuthSession({
    required this.user,
    required this.businessName,
    required this.outlets,
    this.deviceId,
    this.deviceUuid,
    this.selectedOutletId,
  });

  final Employee user;
  final String businessName;
  final List<Outlet> outlets;
  final String? deviceId;
  final String? deviceUuid;
  final String? selectedOutletId;

  AuthSession copyWith({List<Outlet>? outlets, String? selectedOutletId}) {
    return AuthSession(
      user: user,
      businessName: businessName,
      outlets: outlets ?? this.outlets,
      deviceId: deviceId,
      deviceUuid: deviceUuid,
      selectedOutletId: selectedOutletId ?? this.selectedOutletId,
    );
  }
}

class ApiAuthRepository implements AuthRepository {
  const ApiAuthRepository({
    required ApiClient apiClient,
    required TokenStorage tokenStorage,
  }) : _apiClient = apiClient,
       _tokenStorage = tokenStorage;

  final ApiClient _apiClient;
  final TokenStorage _tokenStorage;

  @override
  Future<AuthSession?> restoreSession() async {
    final token = await _tokenStorage.readToken();
    if (token == null || token.isEmpty) return null;
    return me();
  }

  @override
  Future<AuthSession> login({
    required String email,
    required String password,
    required String deviceUuid,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/auth/login',
      data: {'email': email, 'password': password, 'device_uuid': deviceUuid},
    );
    final token = response.data['token'] as String;
    await _tokenStorage.writeToken(token);
    await _tokenStorage.writeDeviceUuid(deviceUuid);
    final fallbackSession = _sessionFromData(response.data);
    final outlets = await listOutletsWithFallback(fallbackSession.outlets);
    final session = fallbackSession.copyWith(outlets: outlets);
    if (session.deviceId != null) {
      await _tokenStorage.writeDeviceId(session.deviceId!);
    }
    return session;
  }

  @override
  Future<AuthSession> me() async {
    final response = await _apiClient.get<Map<String, Object?>>('/me');
    final deviceId = await _tokenStorage.readDeviceId();
    final deviceUuid = await _tokenStorage.readDeviceUuid();
    final outletId = await _tokenStorage.readOutletId();
    final fallbackSession = _sessionFromData(
      response.data,
      deviceId: deviceId,
      deviceUuid: deviceUuid,
    );
    final outlets = await listOutletsWithFallback(fallbackSession.outlets);
    return fallbackSession.copyWith(
      outlets: outlets,
      selectedOutletId: outletId,
    );
  }

  Future<List<Outlet>> listOutletsWithFallback(List<Outlet> fallback) async {
    try {
      final outlets = await listOutlets();
      return outlets.isEmpty ? fallback : outlets;
    } catch (_) {
      return fallback;
    }
  }

  @override
  Future<List<Outlet>> listOutlets() async {
    final response = await _apiClient.get<Object?>('/outlets');
    return [
      for (final outlet in _listFromPayload(response.data, 'outlets'))
        _outletFromJson(outlet),
    ];
  }

  @override
  Future<Employee> pinSwitch({
    required String pin,
    required String deviceId,
    required String outletId,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/auth/pin-switch',
      data: {'pin': pin, 'device_id': deviceId, 'outlet_id': outletId},
    );
    return _employeeFromJson(response.data['cashier']);
  }

  @override
  Future<void> selectOutlet(Outlet outlet) async {
    await _tokenStorage.writeOutletId(outlet.id);
  }

  @override
  Future<void> logout() async {
    try {
      await _apiClient.post<Map<String, Object?>>('/auth/logout');
    } finally {
      await _tokenStorage.clearToken();
    }
  }

  AuthSession _sessionFromData(
    Map<String, Object?> data, {
    String? deviceId,
    String? deviceUuid,
  }) {
    final device = _asMap(data['device']);
    return AuthSession(
      user: _employeeFromJson(data['user']),
      businessName: (_asMap(data['business'])['name'] as String?) ?? '',
      outlets: [
        for (final outlet in (data['outlets'] as List? ?? const []))
          _outletFromJson(outlet),
      ],
      deviceId: (device['id'] as String?) ?? deviceId,
      deviceUuid: (device['device_uuid'] as String?) ?? deviceUuid,
    );
  }
}

String newDeviceUuid() => const Uuid().v4();

Employee _employeeFromJson(Object? value) {
  final json = _asMap(value);
  return Employee(
    id: (json['id'] as String?) ?? '',
    name: (json['name'] as String?) ?? 'Kasir',
    email: (json['email'] as String?) ?? '',
    role: (json['role'] as String?) ?? 'cashier',
  );
}

Outlet _outletFromJson(Object? value) {
  final json = _asMap(value);
  return Outlet(
    id: (json['id'] as String?) ?? '',
    name: (json['name'] as String?) ?? 'Outlet',
    isOnline: true,
    timezone: (json['timezone'] as String?) ?? 'Asia/Jakarta',
    paymentMethods: [
      for (final method in json['payment_methods'] as List? ?? const [])
        _paymentMethodFromJson(method),
    ],
    receiptConfig: ReceiptConfig.fromJson(json['receipt_config']),
  );
}

PaymentMethodConfig _paymentMethodFromJson(Object? value) {
  final json = _asMap(value);
  return PaymentMethodConfig(
    method: (json['method'] as String?) ?? '',
    isCash: json['is_cash'] as bool? ?? false,
  );
}

List<Object?> _listFromPayload(Object? payload, String key) {
  if (payload is List) return List<Object?>.from(payload);
  final json = _asMap(payload);
  final value = json[key] ?? json['items'];
  if (value is List) return List<Object?>.from(value);
  return const [];
}

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}
