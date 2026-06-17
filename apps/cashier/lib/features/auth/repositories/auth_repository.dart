import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/mock/mock_seed_data.dart';
import '../../../shared/models/nojpos_models.dart';

final authRepositoryProvider = Provider<AuthRepository>(
  (ref) => const MockAuthRepository(),
);

abstract interface class AuthRepository {
  Outlet getDefaultOutlet();

  Employee getDefaultCashier();

  List<Employee> getEmployees();

  Employee verifyPin({required String employeeId, required String pin});
}

class MockAuthRepository implements AuthRepository {
  const MockAuthRepository();

  @override
  Outlet getDefaultOutlet() => MockSeedData.outlet;

  @override
  Employee getDefaultCashier() => MockSeedData.employees.first;

  @override
  List<Employee> getEmployees() => MockSeedData.employees;

  @override
  Employee verifyPin({required String employeeId, required String pin}) {
    return MockSeedData.employees.firstWhere(
      (employee) => employee.id == employeeId && employee.pin == pin,
      orElse: () => MockSeedData.employees.first,
    );
  }
}
