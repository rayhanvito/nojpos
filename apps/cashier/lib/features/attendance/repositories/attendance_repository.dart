import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../shared/models/nojpos_models.dart';

final attendanceRepositoryProvider = Provider<AttendanceRepository>(
  (ref) => ApiAttendanceRepository(apiClient: ref.watch(apiClientProvider)),
);

enum AttendanceAction {
  clockIn('clock_in'),
  clockOut('clock_out');

  const AttendanceAction(this.apiValue);

  final String apiValue;
}

abstract interface class AttendanceRepository {
  Future<List<Employee>> getStaff({String? outletId});

  Future<List<AttendanceRecord>> getAttendance({
    String? outletId,
    DateTime? date,
    String? staffId,
  });

  Future<AttendanceRecord> clockAttendance({
    required String employeeId,
    required String pin,
    required AttendanceAction action,
    String? outletId,
  });
}

class ApiAttendanceRepository implements AttendanceRepository {
  const ApiAttendanceRepository({required ApiClient apiClient})
    : _apiClient = apiClient;

  final ApiClient _apiClient;

  @override
  Future<List<Employee>> getStaff({String? outletId}) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/staff',
      queryParameters: {
        if (outletId != null && outletId.isNotEmpty) 'outlet_id': outletId,
      },
    );
    return [
      for (final item in response.data['staff'] as List? ?? const [])
        _employeeFromJson(item),
    ];
  }

  @override
  Future<List<AttendanceRecord>> getAttendance({
    String? outletId,
    DateTime? date,
    String? staffId,
  }) async {
    final response = await _apiClient.get<Map<String, Object?>>(
      '/attendance',
      queryParameters: {
        if (outletId != null && outletId.isNotEmpty) 'outlet_id': outletId,
        if (date != null) 'date': _formatDate(date),
        if (staffId != null && staffId.isNotEmpty) 'staff_id': staffId,
      },
    );
    return [
      for (final item in response.data['attendance'] as List? ?? const [])
        _attendanceFromJson(item),
    ];
  }

  @override
  Future<AttendanceRecord> clockAttendance({
    required String employeeId,
    required String pin,
    required AttendanceAction action,
    String? outletId,
  }) async {
    final response = await _apiClient.post<Map<String, Object?>>(
      '/attendance',
      data: {
        if (outletId != null && outletId.isNotEmpty) 'outlet_id': outletId,
        'staff_id': employeeId,
        'pin': pin,
        'action': action.apiValue,
      },
    );
    return _attendanceFromJson(response.data);
  }
}

AttendanceRecord _attendanceFromJson(Object? value) {
  final json = _asMap(value);
  final employeeJson = _asMap(json['staff'] ?? json['employee']);
  final staffId = (json['staff_id'] ?? json['employee_id'] ?? '').toString();
  final employee = employeeJson.isEmpty
      ? Employee(id: staffId, name: 'Staff', role: 'cashier')
      : _employeeFromJson(employeeJson);
  return AttendanceRecord(
    id: (json['id'] as String?) ?? '',
    employee: employee,
    clockInAt:
        DateTime.tryParse((json['clock_in_at'] ?? '').toString()) ??
        DateTime.now(),
    clockOutAt: DateTime.tryParse((json['clock_out_at'] ?? '').toString()),
  );
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

String _formatDate(DateTime value) {
  final month = value.month.toString().padLeft(2, '0');
  final day = value.day.toString().padLeft(2, '0');
  return '${value.year}-$month-$day';
}
