import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../shared/models/nojpos_models.dart';

final staffRepositoryProvider = Provider<StaffRepository>(
  (ref) => ApiStaffRepository(apiClient: ref.watch(apiClientProvider)),
);

abstract interface class StaffRepository {
  Future<List<Employee>> listStaff({String? outletId, String? search});

  Future<Employee> createStaff({
    required String name,
    required String role,
    required String pin,
    String? email,
    String? password,
  });

  Future<Employee> updateStaff({
    required String id,
    String? name,
    String? role,
    String? pin,
    String? email,
    String? password,
  });

  Future<void> deleteStaff(String id);
}

class ApiStaffRepository implements StaffRepository {
  const ApiStaffRepository({required ApiClient apiClient}) : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<List<Employee>> listStaff({String? outletId, String? search}) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/staff',
      queryParameters: {
        if (outletId != null && outletId.isNotEmpty) 'outlet_id': outletId,
        if (search != null && search.trim().isNotEmpty) 'search': search.trim(),
      },
    );
    return [
      for (final item in response.data['staff'] as List? ?? const [])
        _employeeFromJson(item),
    ];
  }

  @override
  Future<Employee> createStaff({
    required String name,
    required String role,
    required String pin,
    String? email,
    String? password,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/staff',
      data: {
        'name': name,
        'role': role,
        'pin': pin,
        if (email != null && email.trim().isNotEmpty) 'email': email.trim(),
        if (password != null && password.isNotEmpty) 'password': password,
      },
    );
    return _employeeFromJson(response.data['staff']);
  }

  @override
  Future<Employee> updateStaff({
    required String id,
    String? name,
    String? role,
    String? pin,
    String? email,
    String? password,
  }) async {
    final response = await _apiClient.put<Map<String, Object?>>(
      '/staff/$id',
      data: {
        if (name != null && name.trim().isNotEmpty) 'name': name.trim(),
        if (role != null && role.isNotEmpty) 'role': role,
        if (pin != null && pin.isNotEmpty) 'pin': pin,
        if (email != null) 'email': email.trim(),
        if (password != null && password.isNotEmpty) 'password': password,
      },
    );
    return _employeeFromJson(response.data['staff']);
  }

  @override
  Future<void> deleteStaff(String id) async {
    await _apiClient.post<Map<String, Object?>>('/staff/$id/delete');
  }
}

Employee _employeeFromJson(Object? value) {
  final json = _asMap(value);
  return Employee(
    id: (json['id'] as String?) ?? '',
    name: (json['name'] as String?) ?? 'Staff',
    role: (json['role'] as String?) ?? 'cashier',
    email: (json['email'] as String?) ?? '',
  );
}

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}
