import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../repositories/settings_repository.dart';

final settingsAggregateProvider = FutureProvider.autoDispose<SettingsAggregate>(
  (ref) {
    return ref.watch(settingsRepositoryProvider).fetchSettings();
  },
);
