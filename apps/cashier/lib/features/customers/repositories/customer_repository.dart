import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/mock/mock_seed_data.dart';
import '../../../shared/models/nojpos_models.dart';

final customerRepositoryProvider = Provider<CustomerRepository>(
  (ref) => const MockCustomerRepository(),
);

abstract interface class CustomerRepository {
  List<Customer> getCustomers();
}

class MockCustomerRepository implements CustomerRepository {
  const MockCustomerRepository();

  @override
  List<Customer> getCustomers() => MockSeedData.customers;
}
