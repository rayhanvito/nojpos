import 'package:flutter_secure_storage/flutter_secure_storage.dart';

abstract interface class TokenStorage {
  Future<String?> readToken();

  Future<void> writeToken(String token);

  Future<void> clearToken();

  Future<String?> readDeviceUuid();

  Future<void> writeDeviceUuid(String deviceUuid);

  Future<String?> readDeviceId();

  Future<void> writeDeviceId(String deviceId);

  Future<String?> readOutletId();

  Future<void> writeOutletId(String outletId);

  Future<bool> readInitialSyncCompleted({
    required String outletId,
    required String deviceId,
  });

  Future<void> writeInitialSyncCompleted({
    required String outletId,
    required String deviceId,
  });
}

class SecureTokenStorage implements TokenStorage {
  const SecureTokenStorage({FlutterSecureStorage? secureStorage})
    : _secureStorage = secureStorage ?? const FlutterSecureStorage();

  static const _tokenKey = 'nojpos.auth_token';
  static const _deviceUuidKey = 'nojpos.device_uuid';
  static const _deviceIdKey = 'nojpos.device_id';
  static const _outletIdKey = 'nojpos.outlet_id';

  final FlutterSecureStorage _secureStorage;

  @override
  Future<String?> readToken() => _secureStorage.read(key: _tokenKey);

  @override
  Future<void> writeToken(String token) =>
      _secureStorage.write(key: _tokenKey, value: token);

  @override
  Future<void> clearToken() => _secureStorage.delete(key: _tokenKey);

  @override
  Future<String?> readDeviceUuid() => _secureStorage.read(key: _deviceUuidKey);

  @override
  Future<void> writeDeviceUuid(String deviceUuid) =>
      _secureStorage.write(key: _deviceUuidKey, value: deviceUuid);

  @override
  Future<String?> readDeviceId() => _secureStorage.read(key: _deviceIdKey);

  @override
  Future<void> writeDeviceId(String deviceId) =>
      _secureStorage.write(key: _deviceIdKey, value: deviceId);

  @override
  Future<String?> readOutletId() => _secureStorage.read(key: _outletIdKey);

  @override
  Future<void> writeOutletId(String outletId) =>
      _secureStorage.write(key: _outletIdKey, value: outletId);

  @override
  Future<bool> readInitialSyncCompleted({
    required String outletId,
    required String deviceId,
  }) async {
    final value = await _secureStorage.read(
      key: _initialSyncKey(outletId: outletId, deviceId: deviceId),
    );
    return value == '1';
  }

  @override
  Future<void> writeInitialSyncCompleted({
    required String outletId,
    required String deviceId,
  }) {
    return _secureStorage.write(
      key: _initialSyncKey(outletId: outletId, deviceId: deviceId),
      value: '1',
    );
  }

  static String _initialSyncKey({
    required String outletId,
    required String deviceId,
  }) {
    return 'nojpos.initial_sync_completed.$outletId.$deviceId';
  }
}

class InMemoryTokenStorage implements TokenStorage {
  String? _token;
  String? _deviceUuid;
  String? _deviceId;
  String? _outletId;
  final Set<String> _completedInitialSyncKeys = <String>{};

  @override
  Future<String?> readToken() async => _token;

  @override
  Future<void> writeToken(String token) async {
    _token = token;
  }

  @override
  Future<void> clearToken() async {
    _token = null;
  }

  @override
  Future<String?> readDeviceUuid() async => _deviceUuid;

  @override
  Future<void> writeDeviceUuid(String deviceUuid) async {
    _deviceUuid = deviceUuid;
  }

  @override
  Future<String?> readDeviceId() async => _deviceId;

  @override
  Future<void> writeDeviceId(String deviceId) async {
    _deviceId = deviceId;
  }

  @override
  Future<String?> readOutletId() async => _outletId;

  @override
  Future<void> writeOutletId(String outletId) async {
    _outletId = outletId;
  }

  @override
  Future<bool> readInitialSyncCompleted({
    required String outletId,
    required String deviceId,
  }) async {
    return _completedInitialSyncKeys.contains('$outletId::$deviceId');
  }

  @override
  Future<void> writeInitialSyncCompleted({
    required String outletId,
    required String deviceId,
  }) async {
    _completedInitialSyncKeys.add('$outletId::$deviceId');
  }
}
