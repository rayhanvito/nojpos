import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/app/providers/nojpos_session_provider.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/features/screen_lock/providers/terminal_lock_controller.dart';
import 'package:nojpos_tablet_ui/features/screen_lock/repositories/terminal_lock_repository.dart';
import 'package:nojpos_tablet_ui/features/screen_lock/widgets/terminal_lock_overlay.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

const _validPin =
    '12'
    '34';
const _invalidPin =
    '99'
    '99';

void main() {
  test('repository parses lock state and unlock response', () async {
    final adapter = _RouteAdapter({
      'GET /terminal/lock-state': _envelope(_lockStateJson(locked: true)),
      'POST /terminal/lock': _envelope(_lockStateJson(locked: true)),
      'POST /terminal/unlock': _envelope(_unlockJson()),
    });
    final repository = ApiTerminalLockRepository(
      apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
    );

    final state = await repository.getLockState(
      outletId: 'outlet-id',
      deviceId: 'device-id',
      staffId: 'cashier-id',
    );
    expect(state.locked, isTrue);
    expect(state.lockReason, 'manual');
    expect(state.failedAttemptsRemaining, 4);
    expect(state.idleTimeoutSeconds, 30);

    await repository.lockTerminal(
      outletId: 'outlet-id',
      deviceId: 'device-id',
      reason: TerminalLockReason.idleTimeout,
      cashierId: 'cashier-id',
      shiftId: 'shift-id',
    );
    expect(adapter.lastRequestData, {
      'outlet_id': 'outlet-id',
      'device_id': 'device-id',
      'reason': 'idle_timeout',
      'cashier_id': 'cashier-id',
      'shift_id': 'shift-id',
    });

    final unlock = await repository.unlockTerminal(
      outletId: 'outlet-id',
      deviceId: 'device-id',
      staffId: 'cashier-id',
      pin: _validPin,
    );
    expect(unlock.locked, isFalse);
    expect(unlock.sameStaff, isTrue);
    expect(unlock.cartPreserved, isTrue);
    expect(unlock.shiftPreserved, isTrue);
    expect(adapter.lastRequestData, {
      'outlet_id': 'outlet-id',
      'device_id': 'device-id',
      'staff_id': 'cashier-id',
      'pin': _validPin,
      'mode': 'resume_current',
    });
  });

  testWidgets('manual lock shows full-screen overlay and hides POS data', (
    tester,
  ) async {
    final fake = _FakeTerminalLockRepository();
    await _pumpGate(tester, fake);
    await tester.pumpAndSettle();

    expect(find.text('Rp 10.000'), findsOneWidget);
    await tester.tap(find.text('Manual lock'));
    await tester.pumpAndSettle();

    expect(find.text('Terminal terkunci'), findsOneWidget);
    expect(find.text('Rp 10.000'), findsNothing);
    expect(fake.lastLockReason, TerminalLockReason.manual);
  });

  testWidgets('valid unlock removes overlay without clearing session or cart', (
    tester,
  ) async {
    final fake = _FakeTerminalLockRepository(initialLocked: true);
    await _pumpGate(tester, fake);
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byKey(const ValueKey('terminal_lock_pin')),
      _validPin,
    );
    await tester.tap(find.byKey(const ValueKey('terminal_unlock_submit')));
    await tester.pumpAndSettle();

    expect(find.text('Terminal terkunci'), findsNothing);
    expect(find.text('Rp 10.000'), findsOneWidget);
    expect(fake.lastUnlockPin, _validPin);
  });

  testWidgets('invalid PIN keeps overlay and shows error', (tester) async {
    final fake = _FakeTerminalLockRepository(
      initialLocked: true,
      unlockError: const ValidationApiException(
        code: 'INVALID_PIN',
        message: 'bad pin',
        details: {'failed_attempts_remaining': 2},
      ),
    );
    await _pumpGate(tester, fake);
    await tester.pumpAndSettle();

    await tester.enterText(
      find.byKey(const ValueKey('terminal_lock_pin')),
      _invalidPin,
    );
    await tester.tap(find.byKey(const ValueKey('terminal_unlock_submit')));
    await tester.pumpAndSettle();

    expect(find.text('Terminal terkunci'), findsOneWidget);
    expect(find.text('PIN tidak valid.'), findsOneWidget);
    expect(find.text('Sisa percobaan PIN: 2'), findsOneWidget);
    expect(find.text('Rp 10.000'), findsNothing);
  });

  testWidgets('lockout state is shown while locked', (tester) async {
    final fake = _FakeTerminalLockRepository(
      initialLocked: true,
      initialLockoutUntil: DateTime(2026, 6, 20, 10, 15),
    );
    await _pumpGate(tester, fake);
    await tester.pumpAndSettle();

    expect(find.textContaining('PIN terkunci sampai'), findsOneWidget);
    expect(find.text('Rp 10.000'), findsNothing);
  });

  testWidgets('idle timer triggers backend lock', (tester) async {
    final fake = _FakeTerminalLockRepository(idleSeconds: 1);
    await _pumpGate(tester, fake);
    await tester.pumpAndSettle();

    await tester.pump(const Duration(seconds: 1));
    await tester.pumpAndSettle();

    expect(fake.lastLockReason, TerminalLockReason.idleTimeout);
    expect(find.text('Terminal terkunci'), findsOneWidget);
  });
}

Future<void> _pumpGate(
  WidgetTester tester,
  _FakeTerminalLockRepository repository,
) async {
  tester.view.physicalSize = const Size(900, 700);
  tester.view.devicePixelRatio = 1;
  addTearDown(() {
    tester.view.resetPhysicalSize();
    tester.view.resetDevicePixelRatio();
  });
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        terminalLockRepositoryProvider.overrideWithValue(repository),
        nojposSessionProvider.overrideWith(() => _LockSessionNotifier()),
      ],
      child: MaterialApp(
        home: TerminalLockGate(
          child: Scaffold(
            body: Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Text('Rp 10.000'),
                  Consumer(
                    builder: (context, ref, child) => ElevatedButton(
                      onPressed: () {
                        // This is the same controller path used by the POS header.
                        ref
                            .read(terminalLockControllerProvider.notifier)
                            .lock();
                      },
                      child: const Text('Manual lock'),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    ),
  );
}

Map<String, Object?> _envelope(Map<String, Object?> data) => {
  'data': data,
  'meta': <String, Object?>{},
};

Map<String, Object?> _lockStateJson({bool locked = false}) => {
  'locked': locked,
  'locked_at': locked ? '2026-06-20T10:00:00Z' : null,
  'lock_reason': locked ? 'manual' : null,
  'outlet_id': 'outlet-id',
  'device_id': 'device-id',
  'cashier': {'id': 'cashier-id', 'name': 'Siti', 'role': 'cashier'},
  'shift': {'id': 'shift-id', 'preserved': true},
  'failed_attempts_remaining': 4,
  'lockout_until': null,
  'idle_timeout_seconds': 30,
  'session_timeout_seconds': 900,
  'server_time': '2026-06-20T10:00:01Z',
};

Map<String, Object?> _unlockJson() => {
  ..._lockStateJson(locked: false),
  'unlocked_by': {'id': 'cashier-id', 'name': 'Siti', 'role': 'cashier'},
  'mode': 'resume_current',
  'same_staff': true,
  'handover_required': false,
  'shift_preserved': true,
  'cart_preserved': true,
};

class _FakeTerminalLockRepository implements TerminalLockRepository {
  _FakeTerminalLockRepository({
    this.initialLocked = false,
    this.unlockError,
    this.initialLockoutUntil,
    this.idleSeconds = 30,
  });

  final bool initialLocked;
  final ApiException? unlockError;
  final DateTime? initialLockoutUntil;
  final int idleSeconds;
  TerminalLockReason? lastLockReason;
  String? lastUnlockPin;
  TerminalLockState? current;

  @override
  Future<TerminalLockState> getLockState({
    required String outletId,
    required String deviceId,
    String? staffId,
  }) async {
    current ??= _state(initialLocked, lockoutUntil: initialLockoutUntil);
    return current!;
  }

  @override
  Future<TerminalLockState> lockTerminal({
    required String outletId,
    required String deviceId,
    required TerminalLockReason reason,
    String? cashierId,
    String? shiftId,
  }) async {
    lastLockReason = reason;
    current = _state(true, reason: reason.value);
    return current!;
  }

  @override
  Future<TerminalUnlockResponse> unlockTerminal({
    required String outletId,
    required String deviceId,
    required String staffId,
    required String pin,
    String mode = 'resume_current',
  }) async {
    lastUnlockPin = pin;
    final failure = unlockError;
    if (failure != null) throw failure;
    final response = TerminalUnlockResponse.fromJson(_unlockJson());
    current = response;
    return response;
  }

  TerminalLockState _state(
    bool locked, {
    String? reason,
    DateTime? lockoutUntil,
  }) {
    return TerminalLockState(
      locked: locked,
      outletId: 'outlet-id',
      deviceId: 'device-id',
      lockedAt: locked ? DateTime(2026, 6, 20, 10) : null,
      lockReason: reason,
      cashier: const TerminalStaffSummary(
        id: 'cashier-id',
        name: 'Siti',
        role: 'cashier',
      ),
      shift: const TerminalShiftSummary(id: 'shift-id', preserved: true),
      failedAttemptsRemaining: 4,
      lockoutUntil: lockoutUntil,
      idleTimeoutSeconds: idleSeconds,
      sessionTimeoutSeconds: 900,
      serverTime: DateTime(2026, 6, 20, 10),
    );
  }
}

class _LockSessionNotifier extends NojposSessionNotifier {
  @override
  NojposSessionState build() {
    const cashier = Employee(id: 'cashier-id', name: 'Siti', role: 'cashier');
    return NojposSessionState.initial().copyWith(
      cashier: cashier,
      account: cashier,
      status: SessionStatus.ready,
      outlet: const Outlet(
        id: 'outlet-id',
        name: 'Outlet Demo',
        isOnline: true,
      ),
      deviceId: 'device-id',
      activeShift: ShiftSession(
        id: 'shift-id',
        cashier: cashier,
        openedAt: DateTime(2026, 6, 20, 9),
      ),
    );
  }
}

class _RouteAdapter implements HttpClientAdapter {
  _RouteAdapter(this.responses);

  final Map<String, Map<String, Object?>> responses;
  Object? lastRequestData;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final key = '${options.method} ${options.path}';
    lastRequestData = options.data;
    final body = responses[key] ?? _envelope(const <String, Object?>{});
    return ResponseBody.fromString(
      jsonEncode(body),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
