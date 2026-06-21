import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';

final connectivityRepositoryProvider = Provider<ConnectivityRepository>(
  (ref) => ApiConnectivityRepository(apiClient: ref.watch(apiClientProvider)),
);

abstract interface class ConnectivityRepository {
  Future<void> heartbeat();
}

class ApiConnectivityRepository implements ConnectivityRepository {
  const ApiConnectivityRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<void> heartbeat() async {
    await _apiClient.get<Map<String, Object?>>('/me');
  }
}
