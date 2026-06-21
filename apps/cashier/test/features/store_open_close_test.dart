import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/app/providers/nojpos_session_provider.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/features/shift/repositories/shift_repository.dart';
import 'package:nojpos_tablet_ui/features/store/providers/store_controller.dart';
import 'package:nojpos_tablet_ui/features/store/repositories/store_repository.dart';
import 'package:nojpos_tablet_ui/features/store/widgets/store_state_panel.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  test('repository parses store state and sends idempotency keys', () async {
    final adapter = _RouteAdapter({
      'GET /outlets/outlet-id/store-state': _envelope(_storeStateJson()),
      'POST /outlets/outlet-id/store/open': _envelope(_storeStateJson()),
      'POST /outlets/outlet-id/store/close': _envelope(
        _storeStateJson(status: 'closed'),
      ),
    });
    final repository = ApiStoreRepository(
      apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
    );

    final state = await repository.getStoreState('outlet-id');
    expect(state.outletName, 'Outlet Demo');
    expect(state.enabled, isTrue);
    expect(state.pinRequired, isTrue);
    expect(state.blockingShifts.single.cashierName, 'Siti');

    await repository.openStore(
      outletId: 'outlet-id',
      authorizationPin: '1234',
      reason: 'Buka pagi',
      idempotencyKey: 'open-key',
    );
    await repository.closeStore(
      outletId: 'outlet-id',
      authorizationPin: '1234',
      reason: 'Tutup malam',
      idempotencyKey: 'close-key',
    );

    expect(adapter.idempotencyKeys, ['open-key', 'close-key']);
    expect(adapter.lastRequestData, {
      'authorization_pin': '1234',
      'reason': 'Tutup malam',
    });
  });

  test(
    'controller surfaces blocked close shifts and store closed messages',
    () async {
      final fake = _FakeStoreRepository(
        closeError: const ConflictApiException(
          code: 'STORE_CLOSE_BLOCKED_OPEN_SHIFTS',
          message: 'blocked',
          details: {
            'blocking_shifts': [
              {
                'shift_id': 'shift-id',
                'status': 'open',
                'cashier': {'name': 'Siti'},
                'device': {'name': 'Terminal 1'},
              },
            ],
          },
        ),
      );
      final container = ProviderContainer(
        overrides: [storeRepositoryProvider.overrideWithValue(fake)],
      );
      addTearDown(container.dispose);

      final ok = await container
          .read(storeControllerProvider.notifier)
          .closeStore(outletId: 'outlet-id', pin: '1234');

      final state = container.read(storeControllerProvider);
      expect(ok, isFalse);
      expect(
        state.errorMessage,
        'Tutup toko ditolak. Masih ada shift terbuka atau pending close.',
      );
      expect(state.blockingShifts.single.deviceName, 'Terminal 1');

      final checkoutContainer = ProviderContainer(
        overrides: [
          storeRepositoryProvider.overrideWithValue(
            _FakeStoreRepository(
              getError: const ConflictApiException(
                code: 'STORE_CLOSED_CHECKOUT_BLOCKED',
                message: 'closed',
              ),
            ),
          ),
        ],
      );
      addTearDown(checkoutContainer.dispose);

      await checkoutContainer
          .read(storeControllerProvider.notifier)
          .load('outlet-id');
      expect(
        checkoutContainer.read(storeControllerProvider).errorMessage,
        'Toko sedang tutup. Buka toko sebelum checkout.',
      );
    },
  );

  testWidgets(
    'owner admin UI shows store actions and cashier UI is read-only',
    (tester) async {
      await _pumpPanel(
        tester,
        role: 'owner',
        store: StoreState.fromJson(_storeStateJson()),
      );
      await tester.pumpAndSettle();

      final closeButton = tester.widget<OutlinedButton>(
        find.widgetWithText(OutlinedButton, 'Tutup toko'),
      );
      expect(closeButton.onPressed, isNotNull);

      await _pumpPanel(
        tester,
        role: 'cashier',
        store: StoreState.fromJson(_storeStateJson()),
      );
      await tester.pumpAndSettle();

      final cashierCloseButton = tester.widget<OutlinedButton>(
        find.widgetWithText(OutlinedButton, 'Tutup toko'),
      );
      expect(cashierCloseButton.onPressed, isNull);
    },
  );

  testWidgets('blocked close displays blocking shifts', (tester) async {
    final fake = _FakeStoreRepository(
      store: StoreState.fromJson(_storeStateJson()),
      closeError: const ConflictApiException(
        code: 'STORE_CLOSE_BLOCKED_OPEN_SHIFTS',
        message: 'blocked',
        details: {
          'blocking_shifts': [
            {
              'shift_id': 'shift-id',
              'status': 'pending_close',
              'cashier': {'name': 'Siti'},
              'device': {'name': 'Terminal 1'},
            },
          ],
        },
      ),
    );
    await _pumpPanel(tester, role: 'owner', storeRepository: fake);
    await tester.pumpAndSettle();

    await tester.tap(find.widgetWithText(OutlinedButton, 'Tutup toko'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextField).last, '1234');
    await tester.tap(find.widgetWithText(FilledButton, 'Tutup'));
    await tester.pumpAndSettle();

    expect(find.text('Shift yang menghambat tutup toko:'), findsOneWidget);
    expect(find.textContaining('pending_close'), findsOneWidget);
    expect(find.textContaining('Terminal 1'), findsOneWidget);
  });

  test('session open shift displays store closed message', () async {
    final container = ProviderContainer(
      overrides: [
        nojposSessionProvider.overrideWith(() => _StoreSessionNotifier()),
        shiftRepositoryProvider.overrideWithValue(
          _FakeShiftRepository(
            error: const ConflictApiException(
              code: 'STORE_CLOSED_SHIFT_OPEN_BLOCKED',
              message: 'closed',
            ),
          ),
        ),
      ],
    );
    addTearDown(container.dispose);

    final ok = await container
        .read(nojposSessionProvider.notifier)
        .openShift(openingCash: 100000);

    expect(ok, isFalse);
    expect(
      container.read(nojposSessionProvider).errorMessage,
      'Toko sedang tutup. Buka toko sebelum membuka shift.',
    );
  });

  testWidgets(
    'PIN validation error is displayed before store operation submit',
    (tester) async {
      await _pumpPanel(
        tester,
        role: 'owner',
        store: StoreState.fromJson(_storeStateJson()),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(OutlinedButton, 'Tutup toko'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(FilledButton, 'Tutup'));
      await tester.pumpAndSettle();

      expect(find.text('PIN otorisasi wajib diisi.'), findsOneWidget);
    },
  );
}

Future<void> _pumpPanel(
  WidgetTester tester, {
  required String role,
  StoreState? store,
  StoreRepository? storeRepository,
}) async {
  tester.view.physicalSize = const Size(900, 700);
  tester.view.devicePixelRatio = 1;
  addTearDown(() {
    tester.view.resetPhysicalSize();
    tester.view.resetDevicePixelRatio();
  });
  await tester.pumpWidget(
    ProviderScope(
      key: UniqueKey(),
      overrides: [
        nojposSessionProvider.overrideWith(
          () => _StoreSessionNotifier(role: role),
        ),
        storeRepositoryProvider.overrideWithValue(
          storeRepository ?? _FakeStoreRepository(store: store),
        ),
      ],
      child: const MaterialApp(home: Scaffold(body: StoreStatePanel())),
    ),
  );
}

Map<String, Object?> _envelope(Map<String, Object?> data) => {
  'data': data,
  'meta': <String, Object?>{},
};

Map<String, Object?> _storeStateJson({String status = 'open'}) => {
  'outlet_id': 'outlet-id',
  'outlet_name': 'Outlet Demo',
  'status': status,
  'opened_at': '2026-06-20T09:00:00Z',
  'closed_at': status == 'closed' ? '2026-06-20T21:00:00Z' : null,
  'reason': status == 'closed' ? 'Tutup malam' : null,
  'blocking_shifts': [
    {
      'shift_id': 'shift-id',
      'status': 'open',
      'cashier': {'name': 'Siti'},
      'device': {'name': 'Terminal 1'},
      'opened_at': '2026-06-20T08:00:00Z',
    },
  ],
  'settings': {
    'store_open_close_enabled': true,
    'pin_required_store_open_close': true,
  },
  'server_time': '2026-06-20T10:00:00Z',
};

class _FakeStoreRepository implements StoreRepository {
  _FakeStoreRepository({this.store, this.getError, this.closeError});

  final StoreState? store;
  final ApiException? getError;
  final ApiException? closeError;

  @override
  Future<StoreState> getStoreState(String outletId) async {
    final failure = getError;
    if (failure != null) throw failure;
    return store ?? StoreState.fromJson(_storeStateJson());
  }

  @override
  Future<StoreState> openStore({
    required String outletId,
    required String authorizationPin,
    String? reason,
    String? idempotencyKey,
  }) async {
    return StoreState.fromJson(_storeStateJson(status: 'open'));
  }

  @override
  Future<StoreState> closeStore({
    required String outletId,
    required String authorizationPin,
    String? reason,
    String? idempotencyKey,
  }) async {
    final failure = closeError;
    if (failure != null) throw failure;
    return StoreState.fromJson(_storeStateJson(status: 'closed'));
  }
}

class _StoreSessionNotifier extends NojposSessionNotifier {
  _StoreSessionNotifier({this.role = 'owner'});

  final String role;

  @override
  NojposSessionState build() {
    final user = Employee(id: '$role-id', name: role, role: role);
    return NojposSessionState.initial().copyWith(
      account: user,
      cashier: user,
      status: SessionStatus.ready,
      outlet: const Outlet(
        id: 'outlet-id',
        name: 'Outlet Demo',
        isOnline: true,
      ),
      deviceId: 'device-id',
      clearActiveShift: true,
    );
  }
}

class _FakeShiftRepository implements ShiftRepository {
  const _FakeShiftRepository({this.error});

  final ApiException? error;

  @override
  Future<ShiftSession> openShift({
    required String outletId,
    required String deviceId,
    required String cashierId,
    required int openingCash,
  }) async {
    final failure = error;
    if (failure != null) throw failure;
    return ShiftSession(
      id: 'shift-id',
      cashier: Employee(id: cashierId, name: 'Kasir', role: 'cashier'),
      openedAt: DateTime(2026, 6, 20, 9),
    );
  }

  @override
  noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

class _RouteAdapter implements HttpClientAdapter {
  _RouteAdapter(this.responses);

  final Map<String, Map<String, Object?>> responses;
  Object? lastRequestData;
  final List<String> idempotencyKeys = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final key = '${options.method} ${options.path}';
    lastRequestData = options.data;
    final idempotencyKey = options.headers['Idempotency-Key'];
    if (idempotencyKey is String) idempotencyKeys.add(idempotencyKey);
    final response = responses[key];
    if (response == null) {
      return ResponseBody.fromString(
        jsonEncode({
          'error': {'code': 'NOT_FOUND', 'message': key, 'details': {}},
        }),
        404,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }
    return ResponseBody.fromString(
      jsonEncode(response),
      (response['status'] as num?)?.toInt() ?? 200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
