import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/features/attendance/providers/attendance_controller.dart';
import 'package:nojpos_tablet_ui/features/attendance/repositories/attendance_repository.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  group('AttendanceController', () {
    test('should replace an open record after clocking out', () async {
      final repository = _AttendanceRepository();
      final container = ProviderContainer(
        overrides: [
          attendanceRepositoryProvider.overrideWith((ref) => repository),
          attendanceOutletIdProvider.overrideWithValue('outlet-id'),
        ],
      );
      addTearDown(container.dispose);

      final controller = container.read(attendanceControllerProvider.notifier);
      await controller.load();
      final record = await controller.clockAttendance(
        employee: _employee,
        pin: '123456',
      );

      expect(record?.isOpen, isFalse);
      expect(repository.lastAction, AttendanceAction.clockOut);
      expect(
        container.read(attendanceControllerProvider).records.single.isOpen,
        isFalse,
      );
    });

    test('should expose a clock failure without replacing records', () async {
      final repository = _AttendanceRepository(shouldFailClock: true);
      final container = ProviderContainer(
        overrides: [
          attendanceRepositoryProvider.overrideWith((ref) => repository),
          attendanceOutletIdProvider.overrideWithValue('outlet-id'),
        ],
      );
      addTearDown(container.dispose);

      await container.read(attendanceControllerProvider.notifier).load();
      final record = await container
          .read(attendanceControllerProvider.notifier)
          .clockAttendance(employee: _employee, pin: '000000');

      expect(record, isNull);
      expect(
        container.read(attendanceControllerProvider).records.single.isOpen,
        isTrue,
      );
      expect(
        container.read(attendanceControllerProvider).errorMessage,
        'PIN absensi salah.',
      );
    });
  });
}

const _employee = Employee(id: 'cashier-id', name: 'Cashier', role: 'cashier');

class _AttendanceRepository implements AttendanceRepository {
  _AttendanceRepository({this.shouldFailClock = false});

  final bool shouldFailClock;
  AttendanceAction? lastAction;

  @override
  Future<AttendanceRecord> clockAttendance({
    required String employeeId,
    required String pin,
    required AttendanceAction action,
    String? outletId,
  }) async {
    lastAction = action;
    if (shouldFailClock) throw StateError('PIN absensi salah.');
    return AttendanceRecord(
      id: 'attendance-id',
      employee: _employee,
      clockInAt: DateTime(2026, 6, 18, 9),
      clockOutAt: DateTime(2026, 6, 18, 17),
    );
  }

  @override
  Future<List<AttendanceRecord>> getAttendance({
    String? outletId,
    DateTime? date,
    String? staffId,
  }) async {
    return [
      AttendanceRecord(
        id: 'attendance-id',
        employee: _employee,
        clockInAt: DateTime(2026, 6, 18, 9),
      ),
    ];
  }

  @override
  Future<List<Employee>> getStaff({String? outletId}) async {
    return [_employee];
  }
}
