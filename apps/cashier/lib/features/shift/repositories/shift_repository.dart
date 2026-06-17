import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../shared/models/nojpos_models.dart';

final shiftRepositoryProvider = Provider<ShiftRepository>(
  (ref) => const MockShiftRepository(),
);

abstract interface class ShiftRepository {
  ShiftSession createOpenShift({required Employee cashier});
}

class MockShiftRepository implements ShiftRepository {
  const MockShiftRepository();

  @override
  ShiftSession createOpenShift({required Employee cashier}) {
    return ShiftSession(
      id: 'shift-${DateTime.now().microsecondsSinceEpoch}',
      cashier: cashier,
      openedAt: DateTime.now(),
    );
  }
}
