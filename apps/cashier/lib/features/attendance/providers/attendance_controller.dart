import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../shared/models/nojpos_models.dart';
import '../repositories/attendance_repository.dart';

final attendanceOutletIdProvider = Provider<String>((ref) {
  return ref.watch(nojposSessionProvider.select((state) => state.outlet.id));
});

final attendanceControllerProvider =
    NotifierProvider<AttendanceController, AttendanceState>(
      AttendanceController.new,
    );

class AttendanceState {
  const AttendanceState({
    this.staff = const [],
    this.records = const [],
    this.isLoading = false,
    this.errorMessage,
  });

  final List<Employee> staff;
  final List<AttendanceRecord> records;
  final bool isLoading;
  final String? errorMessage;

  AttendanceState copyWith({
    List<Employee>? staff,
    List<AttendanceRecord>? records,
    bool? isLoading,
    String? errorMessage,
    bool clearError = false,
  }) {
    return AttendanceState(
      staff: staff ?? this.staff,
      records: records ?? this.records,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
    );
  }
}

class AttendanceController extends Notifier<AttendanceState> {
  @override
  AttendanceState build() => const AttendanceState();

  Future<void> load({DateTime? date}) async {
    final outletId = ref.read(attendanceOutletIdProvider);
    if (outletId.isEmpty) return;

    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final repository = ref.read(attendanceRepositoryProvider);
      final staff = await repository.getStaff(outletId: outletId);
      final records = await repository.getAttendance(
        outletId: outletId,
        date: date ?? DateTime.now(),
      );
      state = state.copyWith(staff: staff, records: records, isLoading: false);
    } catch (error) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: _messageFor(error),
      );
    }
  }

  Future<AttendanceRecord?> clockAttendance({
    required Employee employee,
    required String pin,
  }) async {
    final outletId = ref.read(attendanceOutletIdProvider);
    if (outletId.isEmpty) return null;

    final isClockedIn = state.records.any(
      (record) => record.employee.id == employee.id && record.isOpen,
    );
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final record = await ref
          .read(attendanceRepositoryProvider)
          .clockAttendance(
            employeeId: employee.id,
            pin: pin,
            action: isClockedIn
                ? AttendanceAction.clockOut
                : AttendanceAction.clockIn,
            outletId: outletId,
          );
      state = state.copyWith(
        records: [
          record,
          for (final existing in state.records)
            if (existing.id != record.id) existing,
        ],
        isLoading: false,
      );
      return record;
    } catch (error) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: _messageFor(error),
      );
      return null;
    }
  }
}

String _messageFor(Object error) {
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}
