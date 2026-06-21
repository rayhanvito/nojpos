import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/nojpos_assets.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../../shared/widgets/nojpos_state_view.dart';
import '../../../shared/widgets/nojpos_toast.dart';
import '../repositories/auth_repository.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final emailController = TextEditingController();
  final passwordController = TextEditingController();
  final emailFocusNode = FocusNode();
  final passwordFocusNode = FocusNode();
  String deviceUuid = newDeviceUuid();
  String? _lastShownError;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      if (emailController.text.trim().isEmpty) {
        emailFocusNode.requestFocus();
      } else {
        passwordFocusNode.requestFocus();
      }
    });
  }

  @override
  void dispose() {
    emailController.dispose();
    passwordController.dispose();
    emailFocusNode.dispose();
    passwordFocusNode.dispose();
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
        await _goToPinOrSync();
      case SessionStatus.booting:
      case SessionStatus.unauthenticated:
    }
  }

  Future<void> _goToPinOrSync() async {
    final completed = await ref
        .read(nojposSessionProvider.notifier)
        .hasCompletedInitialSync();
    if (!mounted) return;
    context.go(completed ? '/pin' : '/sync');
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final error = session.errorMessage;
    if (error != null && error != _lastShownError) {
      _lastShownError = error;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) {
          NojposToast.error(
            context,
            'Login belum berhasil',
            description: error,
          );
        }
      });
    }

    final size = MediaQuery.sizeOf(context);
    final isWide = size.width >= 900;
    return Scaffold(
      backgroundColor: NojposColors.canvas,
      body: SafeArea(
        child: Row(
          children: [
            if (isWide)
              Expanded(
                child: Container(
                  margin: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(NojposRadius.xl),
                    gradient: NojposColors.premiumGradient,
                    boxShadow: NojposShadow.floating,
                  ),
                  child: const _BrandPanel(),
                ).animate().fadeIn(duration: 300.ms).slideX(begin: -.025),
              ),
            Expanded(
              child: Center(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.symmetric(vertical: 20),
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 460),
                    child: Padding(
                      padding: EdgeInsets.all(isWide ? 28 : 20),
                      child: _LoginCard(
                        session: session,
                        emailController: emailController,
                        passwordController: passwordController,
                        emailFocusNode: emailFocusNode,
                        passwordFocusNode: passwordFocusNode,
                        onLogin: _login,
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

class _LoginCard extends StatelessWidget {
  const _LoginCard({
    required this.session,
    required this.emailController,
    required this.passwordController,
    required this.emailFocusNode,
    required this.passwordFocusNode,
    required this.onLogin,
  });

  final NojposSessionState session;
  final TextEditingController emailController;
  final TextEditingController passwordController;
  final FocusNode emailFocusNode;
  final FocusNode passwordFocusNode;
  final Future<void> Function() onLogin;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(NojposRadius.xl),
        border: Border.all(color: NojposColors.line),
        boxShadow: NojposShadow.card,
      ),
      child: Padding(
        padding: const EdgeInsets.all(28),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const _LogoMark(),
            const SizedBox(height: 28),
            Text(
              'Masuk ke NojPOS',
              style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                fontWeight: FontWeight.w900,
                color: NojposColors.text,
                height: 1.05,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              'Login akun perangkat, pilih outlet, lalu lanjut PIN kasir.',
              style: TextStyle(color: NojposColors.muted, fontSize: 15),
            ),
            const SizedBox(height: 24),
            _PremiumTextField(
              key: const ValueKey('login_email'),
              controller: emailController,
              label: 'Email',
              icon: LucideIcons.mail,
              keyboardType: TextInputType.emailAddress,
              focusNode: emailFocusNode,
              textInputAction: TextInputAction.next,
              onSubmitted: (_) => passwordFocusNode.requestFocus(),
            ),
            const SizedBox(height: 14),
            _PremiumTextField(
              key: const ValueKey('login_password'),
              controller: passwordController,
              label: 'Password',
              icon: LucideIcons.lockKeyhole,
              obscureText: true,
              focusNode: passwordFocusNode,
              textInputAction: TextInputAction.done,
              onSubmitted: (_) {
                if (!session.isBusy) onLogin();
              },
            ),
            if (session.errorMessage != null) ...[
              const SizedBox(height: 16),
              NojposStateView.error(
                title: 'Login belum berhasil',
                subtitle: session.errorMessage!,
                compact: true,
              ),
            ],
            const SizedBox(height: 24),
            FilledButton.icon(
              key: const ValueKey('login_submit'),
              onPressed: session.isBusy ? null : onLogin,
              icon: session.isBusy
                  ? const SizedBox.square(
                      dimension: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(LucideIcons.arrowRight, size: 20),
              label: Text(session.isBusy ? 'Memverifikasi...' : 'Masuk'),
              style: FilledButton.styleFrom(
                backgroundColor: NojposColors.primary,
                foregroundColor: Colors.white,
                minimumSize: const Size.fromHeight(56),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(NojposRadius.lg),
                ),
                textStyle: const TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            const SizedBox(height: 18),
            const _TerminalReadyPill(),
          ],
        ),
      ),
    ).animate().fadeIn(duration: 260.ms).slideY(begin: .03, end: 0);
  }
}

class _PremiumTextField extends StatelessWidget {
  const _PremiumTextField({
    super.key,
    required this.controller,
    required this.label,
    required this.icon,
    this.keyboardType,
    this.obscureText = false,
    this.focusNode,
    this.textInputAction,
    this.onSubmitted,
  });

  final TextEditingController controller;
  final String label;
  final IconData icon;
  final TextInputType? keyboardType;
  final bool obscureText;
  final FocusNode? focusNode;
  final TextInputAction? textInputAction;
  final ValueChanged<String>? onSubmitted;

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      focusNode: focusNode,
      keyboardType: keyboardType,
      obscureText: obscureText,
      textInputAction: textInputAction,
      onSubmitted: onSubmitted,
      decoration: InputDecoration(
        labelText: label,
        prefixIcon: Icon(icon, color: NojposColors.muted),
        filled: true,
        fillColor: NojposColors.canvas,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(NojposRadius.lg),
          borderSide: const BorderSide(color: NojposColors.line),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(NojposRadius.lg),
          borderSide: const BorderSide(color: NojposColors.line),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(NojposRadius.lg),
          borderSide: const BorderSide(color: NojposColors.primary, width: 1.6),
        ),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 16,
          vertical: 18,
        ),
      ),
    );
  }
}

class _BrandPanel extends StatelessWidget {
  const _BrandPanel();

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        Positioned(
          right: -70,
          top: -70,
          child: _GlowBlob(size: 220, opacity: .20),
        ),
        Positioned(
          left: -90,
          bottom: -90,
          child: _GlowBlob(size: 260, opacity: .14),
        ),
        Padding(
          padding: const EdgeInsets.all(48),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const _LogoMark(inverted: true),
              const Spacer(),
              const _HeroTerminalCard(),
              const SizedBox(height: 28),
              Text(
                'Kasir cepat, aman, dan siap operasional harian.',
                style: Theme.of(context).textTheme.displaySmall?.copyWith(
                  color: Colors.white,
                  height: 1.08,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 18),
              const Text(
                'Satu terminal untuk login outlet, PIN kasir, katalog produk, pembayaran, dan struk digital.',
                style: TextStyle(
                  color: Color(0xE8FFFFFF),
                  fontSize: 17,
                  height: 1.5,
                ),
              ),
              const SizedBox(height: 24),
              const Wrap(
                spacing: 12,
                runSpacing: 12,
                children: [
                  _FloatingFeature(
                    icon: LucideIcons.zap,
                    label: 'Fast Checkout',
                  ),
                  _FloatingFeature(
                    icon: LucideIcons.shieldCheck,
                    label: 'Secure Payment',
                  ),
                  _FloatingFeature(
                    icon: LucideIcons.chartNoAxesCombined,
                    label: 'Live Sales',
                  ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _HeroTerminalCard extends StatelessWidget {
  const _HeroTerminalCard();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 360,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: .16),
        borderRadius: BorderRadius.circular(28),
        border: Border.all(color: Colors.white.withValues(alpha: .22)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 46,
                height: 46,
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: .22),
                  borderRadius: BorderRadius.circular(16),
                ),
                child: const Icon(LucideIcons.store, color: Colors.white),
              ),
              const SizedBox(width: 12),
              const Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'NOJPOS Terminal',
                      style: TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.w900,
                        fontSize: 16,
                      ),
                    ),
                    SizedBox(height: 3),
                    Text(
                      'Cashier terminal ready',
                      style: TextStyle(color: Color(0xCCFFFFFF), fontSize: 12),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 18),
          SizedBox(
            height: 116,
            child: SvgPicture.asset(
              'assets/illustrations/login_pos_hero.svg',
              fit: BoxFit.contain,
              placeholderBuilder: (_) => const _HeroIllustrationFallback(),
              errorBuilder: (context, error, stackTrace) =>
                  const _HeroIllustrationFallback(),
            ),
          ),
          const SizedBox(height: 18),
          Container(
            height: 78,
            decoration: BoxDecoration(
              color: Colors.white.withValues(alpha: .18),
              borderRadius: BorderRadius.circular(20),
            ),
            child: const Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: [
                _Metric(value: 'Server', label: 'Total resmi'),
                _Metric(value: 'Ready', label: 'Terminal'),
                _Metric(value: 'PIN', label: 'Kasir login'),
              ],
            ),
          ),
        ],
      ),
    ).animate().fadeIn(duration: 420.ms).slideY(begin: .05, end: 0);
  }
}

class _HeroIllustrationFallback extends StatelessWidget {
  const _HeroIllustrationFallback();

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: .18),
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: Colors.white.withValues(alpha: .20)),
      ),
      child: const Center(
        child: Icon(LucideIcons.store, color: Colors.white, size: 46),
      ),
    );
  }
}

class _Metric extends StatelessWidget {
  const _Metric({required this.value, required this.label});

  final String value;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          value,
          style: const TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.w900,
            fontSize: 18,
          ),
        ),
        const SizedBox(height: 3),
        Text(
          label,
          style: const TextStyle(color: Color(0xCCFFFFFF), fontSize: 11),
        ),
      ],
    );
  }
}

class _FloatingFeature extends StatelessWidget {
  const _FloatingFeature({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 13, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: .16),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: Colors.white.withValues(alpha: .22)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, color: Colors.white, size: 16),
          const SizedBox(width: 8),
          Text(
            label,
            style: const TextStyle(
              color: Colors.white,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _TerminalReadyPill extends StatelessWidget {
  const _TerminalReadyPill();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 9),
        decoration: BoxDecoration(
          color: NojposColors.primarySoft,
          borderRadius: BorderRadius.circular(999),
          border: Border.all(color: NojposColors.successBorder),
        ),
        child: const Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              LucideIcons.monitorCheck,
              color: NojposColors.primary,
              size: 18,
            ),
            SizedBox(width: 8),
            Flexible(
              child: Text(
                'Terminal kasir siap',
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  color: NojposColors.muted,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _GlowBlob extends StatelessWidget {
  const _GlowBlob({required this.size, required this.opacity});

  final double size;
  final double opacity;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: opacity),
        shape: BoxShape.circle,
      ),
    );
  }
}

class _LogoMark extends StatelessWidget {
  const _LogoMark({this.inverted = false});

  final bool inverted;

  @override
  Widget build(BuildContext context) {
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
            boxShadow: inverted ? null : NojposShadow.soft,
          ),
          child: Padding(
            padding: const EdgeInsets.all(8),
            child: SvgPicture.asset(
              NojposAssets.logoSymbol,
              fit: BoxFit.contain,
            ),
          ),
        ),
        const SizedBox(width: 14),
        if (inverted)
          const Text(
            'NOJPOS',
            style: TextStyle(
              color: Colors.white,
              fontSize: 26,
              fontWeight: FontWeight.w900,
            ),
          )
        else
          SizedBox(
            width: 138,
            height: 38,
            child: SvgPicture.asset(
              NojposAssets.logoFull,
              fit: BoxFit.contain,
              alignment: Alignment.centerLeft,
            ),
          ),
      ],
    );
  }
}
