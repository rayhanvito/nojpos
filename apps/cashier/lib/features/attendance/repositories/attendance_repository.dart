import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/nojpos_models.dart';

final attendanceRepositoryProvider = Provider<AttendanceRepository>(
  (ref) => const MockAttendanceRepository(),
);

abstract interface class AttendanceRepository {
  AttendanceRecord clockIn(Employee employee);

  AttendanceRecord clockOut(AttendanceRecord record);
}

class MockAttendanceRepository implements AttendanceRepository {
  const MockAttendanceRepository();

  @override
  AttendanceRecord clockIn(Employee employee) {
    return AttendanceRecord(
      id: 'att-${DateTime.now().microsecondsSinceEpoch}',
      employee: employee,
      clockInAt: DateTime.now(),
    );
  }

  @override
  AttendanceRecord clockOut(AttendanceRecord record) {
    return AttendanceRecord(
      id: record.id,
      employee: record.employee,
      clockInAt: record.clockInAt,
      clockOutAt: DateTime.now(),
    );
  }
}
