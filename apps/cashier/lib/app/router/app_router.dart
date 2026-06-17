import 'package:go_router/go_router.dart';

import '../../features/auth/pages/login_screen.dart';
import '../../features/auth/pages/pin_screen.dart';
import '../../features/pos/screens/operations_screen.dart';
import '../../features/payment/pages/payment_screen.dart';
import '../../features/pos/pages/pos_screen.dart';
import '../../features/transactions/pages/success_screen.dart';

final appRouter = GoRouter(
  initialLocation: '/login',
  routes: [
    GoRoute(path: '/login', builder: (context, state) => const LoginScreen()),
    GoRoute(path: '/pin', builder: (context, state) => const PinScreen()),
    GoRoute(path: '/pos', builder: (context, state) => const PosScreen()),
    GoRoute(
      path: '/orders',
      builder: (context, state) =>
          const OperationsScreen(type: OperationsPageType.orders),
    ),
    GoRoute(
      path: '/sales',
      builder: (context, state) =>
          const OperationsScreen(type: OperationsPageType.sales),
    ),
    GoRoute(
      path: '/reports',
      builder: (context, state) =>
          const OperationsScreen(type: OperationsPageType.reports),
    ),
    GoRoute(
      path: '/inventory',
      builder: (context, state) =>
          const OperationsScreen(type: OperationsPageType.inventory),
    ),
    GoRoute(
      path: '/settings',
      builder: (context, state) =>
          const OperationsScreen(type: OperationsPageType.settings),
    ),
    GoRoute(
      path: '/attendance',
      builder: (context, state) =>
          const OperationsScreen(type: OperationsPageType.attendance),
    ),
    GoRoute(
      path: '/payment',
      builder: (context, state) => const PaymentScreen(),
    ),
    GoRoute(
      path: '/success',
      builder: (context, state) => const SuccessScreen(),
    ),
  ],
);
