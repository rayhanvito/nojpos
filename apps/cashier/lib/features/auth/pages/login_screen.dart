import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';

class LoginScreen extends StatelessWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context) {
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
                            'Pilih kasir aktif dan lanjutkan transaksi toko.',
                            style: TextStyle(
                              color: MokposColors.muted,
                              fontSize: 15,
                            ),
                          ),
                          const SizedBox(height: 28),
                          const _FieldLabel('Outlet'),
                          const _SelectField(
                            icon: LucideIcons.store,
                            text: 'Kedai Nusantara',
                          ),
                          const SizedBox(height: 16),
                          const _FieldLabel('Kasir'),
                          const _SelectField(
                            icon: LucideIcons.userRound,
                            text: 'Rayhan - Kasir Utama',
                          ),
                          const SizedBox(height: 28),
                          FilledButton.icon(
                            onPressed: () => context.go('/pin'),
                            icon: const Icon(LucideIcons.arrowRight, size: 20),
                            label: const Text('Lanjut ke PIN'),
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
                                'Online · Shift aktif',
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
            'UI mock bersih untuk fondasi kasir operasional ala Majoo/Moka.',
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

class _FieldLabel extends StatelessWidget {
  const _FieldLabel(this.text);

  final String text;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(
        text,
        style: const TextStyle(
          color: MokposColors.text,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _SelectField extends StatelessWidget {
  const _SelectField({required this.icon, required this.text});

  final IconData icon;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 56,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(MokposRadius.md),
      ),
      child: Row(
        children: [
          Icon(icon, color: MokposColors.primary, size: 21),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              text,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
          const Icon(
            LucideIcons.chevronDown,
            color: MokposColors.muted,
            size: 20,
          ),
        ],
      ),
    );
  }
}
