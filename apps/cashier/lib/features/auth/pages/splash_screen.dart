import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';

class SplashScreen extends ConsumerStatefulWidget {
  const SplashScreen({super.key});

  @override
  ConsumerState<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends ConsumerState<SplashScreen> {
  @override
  void initState() {
    super.initState();
    Future<void>.microtask(_bootstrap);
  }

  Future<void> _bootstrap() async {
    await ref.read(nojposSessionProvider.notifier).bootstrap();
    if (!mounted) return;
    final status = ref.read(nojposSessionProvider).status;
    switch (status) {
      case SessionStatus.pinRequired:
      case SessionStatus.ready:
        context.go('/sync');
      case SessionStatus.outletRequired:
        context.go('/outlet');
      case SessionStatus.unauthenticated:
      case SessionStatus.booting:
        context.go('/login');
    }
  }

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            CircularProgressIndicator(),
            SizedBox(height: 18),
            Text(
              'Menyiapkan NojPOS',
              style: TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w900,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
