import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../shared/models/nojpos_models.dart';

final customerRepositoryProvider = Provider<CustomerRepository>(
  (ref) => ApiCustomerRepository(apiClient: ref.watch(apiClientProvider)),
);

abstract interface class CustomerRepository {
  Future<List<Customer>> searchCustomers({String? search, String? group});

  Future<Customer> createCustomer({
    required String name,
    String? phone,
    String? group,
  });
}

class ApiCustomerRepository implements CustomerRepository {
  const ApiCustomerRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<List<Customer>> searchCustomers({
    String? search,
    String? group,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/customers',
      queryParameters: {
        if (search != null && search.trim().isNotEmpty) 'search': search.trim(),
        if (group != null && group.trim().isNotEmpty && group != 'Semua')
          'group': group.trim(),
      },
    );
    final customers = response.data['customers'] as List? ?? const [];
    return [
      for (final customer in customers) _customerFromJson(_asMap(customer)),
    ];
  }

  @override
  Future<Customer> createCustomer({
    required String name,
    String? phone,
    String? group,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/customers',
      data: {
        'name': name,
        if (phone != null && phone.trim().isNotEmpty) 'phone': phone.trim(),
        if (group != null && group.trim().isNotEmpty) 'group': group.trim(),
      },
    );
    return _customerFromJson(response.data);
  }
}

Customer _customerFromJson(Map<String, Object?> json) {
  return Customer(
    id: (json['id'] as String?) ?? '',
    name: (json['name'] as String?) ?? '',
    phone: (json['phone'] as String?) ?? '',
    group: (json['group'] as String?) ?? 'Tanpa Grup',
  );
}

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}
