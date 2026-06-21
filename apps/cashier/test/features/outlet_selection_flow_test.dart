import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/core/storage/token_storage.dart';
import 'package:nojpos_tablet_ui/features/auth/pages/login_screen.dart';
import 'package:nojpos_tablet_ui/features/auth/pages/outlet_select_screen.dart';
import 'package:nojpos_tablet_ui/features/auth/pages/pin_screen.dart';
import 'package:nojpos_tablet_ui/features/auth/pages/splash_screen.dart';
import 'package:nojpos_tablet_ui/features/auth/repositories/auth_repository.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  group('Outlet selection flow', () {
    testWidgets('should show outlet select when account has multiple outlets', (
      tester,
    ) async {
      final auth = _FlowAuthRepository(outlets: _outlets);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authRepositoryProvider.overrideWith((ref) => auth),
            tokenStorageProvider.overrideWith((ref) => InMemoryTokenStorage()),
          ],
          child: const _FlowTestApp(initialLocation: '/login'),
        ),
      );
      await _submitLogin(tester);

      expect(find.text('Pilih Outlet'), findsOneWidget);
      expect(find.text('Outlet Utama'), findsOneWidget);
      expect(find.text('Outlet Kedua'), findsOneWidget);
      expect(auth.selectedOutletId, isNull);
    });

    testWidgets('should auto-select one outlet and skip to PIN', (
      tester,
    ) async {
      final auth = _FlowAuthRepository(outlets: [_outlets.first]);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authRepositoryProvider.overrideWith((ref) => auth),
            tokenStorageProvider.overrideWith((ref) => InMemoryTokenStorage()),
          ],
          child: const _FlowTestApp(initialLocation: '/login'),
        ),
      );
      await _submitLogin(tester);

      expect(find.byType(PinScreen), findsOneWidget);
      expect(find.text('Outlet Utama · NojPOS Demo'), findsOneWidget);
      expect(auth.selectedOutletId, 'outlet-1');
    });

    test('should persist selected outlet in token storage', () async {
      final storage = InMemoryTokenStorage();
      final repository = ApiAuthRepository(
        apiClient: ApiClient(),
        tokenStorage: storage,
      );

      await repository.selectOutlet(_outlets.last);

      expect(await storage.readOutletId(), 'outlet-2');
    });

    testWidgets('should route restored token without outlet to outlet screen', (
      tester,
    ) async {
      final auth = _FlowAuthRepository(outlets: _outlets);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            authRepositoryProvider.overrideWith((ref) => auth),
            tokenStorageProvider.overrideWith((ref) => InMemoryTokenStorage()),
          ],
          child: const _FlowTestApp(initialLocation: '/'),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Pilih Outlet'), findsOneWidget);
      expect(find.text('Outlet Utama'), findsOneWidget);
      expect(find.text('Outlet Kedua'), findsOneWidget);
    });
  });
}

Future<void> _submitLogin(WidgetTester tester) async {
  await tester.enterText(
    find.byKey(const ValueKey('login_email')),
    'operator@example.test',
  );
  await tester.enterText(
    find.byKey(const ValueKey('login_password')),
    'secure-password',
  );
  await tester.tap(find.text('Masuk'));
  await tester.pumpAndSettle();
}

const _outlets = [
  Outlet(id: 'outlet-1', name: 'Outlet Utama', isOnline: true),
  Outlet(id: 'outlet-2', name: 'Outlet Kedua', isOnline: true),
];

class _FlowTestApp extends StatelessWidget {
  const _FlowTestApp({required this.initialLocation});

  final String initialLocation;

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      routerConfig: GoRouter(
        initialLocation: initialLocation,
        routes: [
          GoRoute(path: '/', builder: (context, state) => const SplashScreen()),
          GoRoute(
            path: '/login',
            builder: (context, state) => const LoginScreen(),
          ),
          GoRoute(
            path: '/outlet',
            builder: (context, state) => const OutletSelectScreen(),
          ),
          GoRoute(path: '/pin', builder: (context, state) => const PinScreen()),
          GoRoute(
            path: '/sync',
            builder: (context, state) => const PinScreen(),
          ),
        ],
      ),
    );
  }
}

class _FlowAuthRepository implements AuthRepository {
  _FlowAuthRepository({required this.outlets});

  final List<Outlet> outlets;
  String? selectedOutletId;

  @override
  Future<AuthSession?> restoreSession() async {
    return AuthSession(
      user: const Employee(id: 'owner-id', name: 'Owner', role: 'owner'),
      businessName: 'NojPOS Demo',
      outlets: outlets,
      deviceId: 'device-id',
      deviceUuid: 'device-uuid',
    );
  }

  @override
  Future<AuthSession> login({
    required String email,
    required String password,
    required String deviceUuid,
  }) async {
    return AuthSession(
      user: const Employee(id: 'owner-id', name: 'Owner', role: 'owner'),
      businessName: 'NojPOS Demo',
      outlets: outlets,
      deviceId: 'device-id',
      deviceUuid: deviceUuid,
    );
  }

  @override
  Future<AuthSession> me() async => (await restoreSession())!;

  @override
  Future<List<Outlet>> listOutlets() async => outlets;

  @override
  Future<Employee> pinSwitch({
    required String pin,
    required String deviceId,
    required String outletId,
  }) async {
    return const Employee(id: 'cashier-id', name: 'Cashier', role: 'cashier');
  }

  @override
  Future<void> selectOutlet(Outlet outlet) async {
    selectedOutletId = outlet.id;
  }

  @override
  Future<void> logout() async {}
}
