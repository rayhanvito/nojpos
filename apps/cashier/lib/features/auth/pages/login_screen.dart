import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../repositories/auth_repository.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final emailController = TextEditingController(text: 'owner@demo.nojpos.test');
  final passwordController = TextEditingController(text: 'password');
  String deviceUuid = newDeviceUuid();

  @override
  void dispose() {
    emailController.dispose();
    passwordController.dispose();
    super.dispose();
  }

  Future<void> _login() async {
    await ref
        .read(nojposSessionProvider.notifier)
        .login(
          email: emailController.text.trim(),
          password: passwordController.text,
          deviceUuid: deviceUuid,
        );
    if (!mounted) return;
    final status = ref.read(nojposSessionProvider).status;
    switch (status) {
      case SessionStatus.outletRequired:
        context.go('/outlet');
      case SessionStatus.pinRequired:
      case SessionStatus.ready:
        context.go('/sync');
      case SessionStatus.booting:
      case SessionStatus.unauthenticated:
    }
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final size = MediaQuery.sizeOf(context);
    final isWide = size.width >= 900;
    return Scaffold(
      body: SafeArea(
        child: Row(
          children: [
            if (isWide)
              Expanded(
                child: Container(
                  margin: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(MokposRadius.xl),
                    gradient: const LinearGradient(
                      colors: [MokposColors.primary, MokposColors.accent],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                  ),
                  child: const _BrandPanel(),
                ),
              ),
            Expanded(
              child: Center(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.symmetric(vertical: 20),
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 430),
                    child: Padding(
                      padding: const EdgeInsets.all(28),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          const _LogoMark(),
                          const SizedBox(height: 30),
                          Text(
                            'Masuk ke NojPOS',
                            style: Theme.of(context).textTheme.headlineMedium
                                ?.copyWith(
                                  fontWeight: FontWeight.w800,
                                  color: MokposColors.text,
                                ),
                          ),
                          const SizedBox(height: 8),
                          const Text(
                            'Login akun, pilih outlet, lalu lanjut PIN kasir.',
                            style: TextStyle(
                              color: MokposColors.muted,
                              fontSize: 15,
                            ),
                          ),
                          const SizedBox(height: 28),
                          TextField(
                            key: const ValueKey('login_email'),
                            controller: emailController,
                            keyboardType: TextInputType.emailAddress,
                            decoration: const InputDecoration(
                              labelText: 'Email',
                              prefixIcon: Icon(LucideIcons.mail),
                              border: OutlineInputBorder(),
                            ),
                          ),
                          const SizedBox(height: 14),
                          TextField(
                            key: const ValueKey('login_password'),
                            controller: passwordController,
                            obscureText: true,
                            decoration: const InputDecoration(
                              labelText: 'Password',
                              prefixIcon: Icon(LucideIcons.lockKeyhole),
                              border: OutlineInputBorder(),
                            ),
                          ),
                          if (session.errorMessage != null) ...[
                            const SizedBox(height: 14),
                            Text(
                              session.errorMessage!,
                              style: const TextStyle(
                                color: MokposColors.danger,
                              ),
                            ),
                          ],
                          const SizedBox(height: 28),
                          FilledButton.icon(
                            key: const ValueKey('login_submit'),
                            onPressed: session.isBusy ? null : _login,
                            icon: session.isBusy
                                ? const SizedBox.square(
                                    dimension: 18,
                                    child: CircularProgressIndicator(
                                      strokeWidth: 2,
                                    ),
                                  )
                                : const Icon(LucideIcons.arrowRight, size: 20),
                            label: const Text('Masuk'),
                            style: FilledButton.styleFrom(
                              backgroundColor: MokposColors.primary,
                              foregroundColor: Colors.white,
                              minimumSize: const Size.fromHeight(56),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(
                                  MokposRadius.md,
                                ),
                              ),
                              textStyle: const TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                          ),
                          const SizedBox(height: 18),
                          const Row(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(
                                LucideIcons.wifi,
                                color: MokposColors.primary,
                                size: 18,
                              ),
                              SizedBox(width: 8),
                              Text(
                                'Online · Laravel API',
                                style: TextStyle(
                                  color: MokposColors.muted,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _BrandPanel extends StatelessWidget {
  const _BrandPanel();

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(48),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const _LogoMark(inverted: true),
          const Spacer(),
          Text(
            'POS tablet cepat untuk pilih produk, masuk keranjang, lalu bayar.',
            style: Theme.of(context).textTheme.displaySmall?.copyWith(
              color: Colors.white,
              height: 1.08,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 20),
          const Text(
            'Wave 1A tersambung ke Laravel API untuk login, PIN, shift, dan katalog.',
            style: TextStyle(
              color: Color(0xDFFFFFFF),
              fontSize: 18,
              height: 1.5,
            ),
          ),
        ],
      ),
    );
  }
}

class _LogoMark extends StatelessWidget {
  const _LogoMark({this.inverted = false});

  final bool inverted;

  @override
  Widget build(BuildContext context) {
    final color = inverted ? Colors.white : MokposColors.primary;
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 52,
          height: 52,
          decoration: BoxDecoration(
            color: inverted
                ? Colors.white.withValues(alpha: .16)
                : Colors.white,
            borderRadius: BorderRadius.circular(16),
            boxShadow: inverted ? null : MokposShadow.soft,
          ),
          child: Icon(LucideIcons.receiptText, color: color, size: 30),
        ),
        const SizedBox(width: 14),
        Text(
          'NojPOS',
          style: TextStyle(
            color: inverted ? Colors.white : MokposColors.text,
            fontSize: 26,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}
