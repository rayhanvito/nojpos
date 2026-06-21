import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/features/connectivity/providers/connectivity_provider.dart';
import 'package:nojpos_tablet_ui/features/connectivity/repositories/connectivity_repository.dart';

void main() {
  test('heartbeat success marks terminal online', () async {
    final repository = _FakeConnectivityRepository();
    final container = ProviderContainer(
      overrides: [connectivityRepositoryProvider.overrideWithValue(repository)],
    );
    addTearDown(container.dispose);

    expect(
      container.read(connectivityControllerProvider).status,
      ConnectivityStatus.checking,
    );

    await container.read(connectivityControllerProvider.notifier).checkNow();

    final state = container.read(connectivityControllerProvider);
    expect(state.status, ConnectivityStatus.online);
    expect(state.lastCheckedAt, isNotNull);
    expect(state.errorMessage, isNull);
  });

  test('network failure marks terminal offline without throwing', () async {
    final repository = _FakeConnectivityRepository(
      error: const ServerApiException(
        code: 'NETWORK_ERROR',
        message: 'Network request failed.',
      ),
    );
    final container = ProviderContainer(
      overrides: [connectivityRepositoryProvider.overrideWithValue(repository)],
    );
    addTearDown(container.dispose);

    await container.read(connectivityControllerProvider.notifier).checkNow();

    final state = container.read(connectivityControllerProvider);
    expect(state.status, ConnectivityStatus.offline);
    expect(state.errorMessage, contains('Server tidak dapat dijangkau'));
  });

  test(
    'authenticated API errors are degraded instead of fake online',
    () async {
      final repository = _FakeConnectivityRepository(
        error: const UnauthorizedApiException(
          code: 'UNAUTHENTICATED',
          message: 'Unauthenticated.',
        ),
      );
      final container = ProviderContainer(
        overrides: [
          connectivityRepositoryProvider.overrideWithValue(repository),
        ],
      );
      addTearDown(container.dispose);

      await container.read(connectivityControllerProvider.notifier).checkNow();

      final state = container.read(connectivityControllerProvider);
      expect(state.status, ConnectivityStatus.degraded);
      expect(state.errorMessage, contains('Sesi login'));
    },
  );
}

class _FakeConnectivityRepository implements ConnectivityRepository {
  const _FakeConnectivityRepository({this.error});

  final Object? error;

  @override
  Future<void> heartbeat() async {
    final failure = error;
    if (failure != null) {
      throw failure;
    }
  }
}
