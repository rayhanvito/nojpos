import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';

class PinScreen extends StatefulWidget {
  const PinScreen({super.key});

  @override
  State<PinScreen> createState() => _PinScreenState();
}

class _PinScreenState extends State<PinScreen> {
  String pin = '';

  void _tap(String value) {
    if (pin.length >= 4) return;
    setState(() => pin += value);
    if (pin.length == 4) {
      Future<void>.delayed(const Duration(milliseconds: 180), () {
        if (mounted) context.go('/pos');
      });
    }
  }

  void _delete() {
    if (pin.isEmpty) return;
    setState(() => pin = pin.substring(0, pin.length - 1));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 420),
            child: Padding(
              padding: const EdgeInsets.all(28),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Container(
                    width: 72,
                    height: 72,
                    decoration: BoxDecoration(
                      color: MokposColors.primarySoft,
                      borderRadius: BorderRadius.circular(22),
                    ),
                    child: const Icon(
                      LucideIcons.lockKeyhole,
                      color: MokposColors.primary,
                      size: 34,
                    ),
                  ),
                  const SizedBox(height: 24),
                  Text(
                    'Masukkan PIN Kasir',
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.w900,
                      color: MokposColors.text,
                    ),
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'PIN demo apa saja, 4 digit langsung masuk.',
                    style: TextStyle(color: MokposColors.muted),
                  ),
                  const SizedBox(height: 28),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: List.generate(
                      4,
                      (index) => AnimatedContainer(
                        duration: const Duration(milliseconds: 160),
                        width: 18,
                        height: 18,
                        margin: const EdgeInsets.symmetric(horizontal: 9),
                        decoration: BoxDecoration(
                          color: index < pin.length
                              ? MokposColors.primary
                              : Colors.white,
                          shape: BoxShape.circle,
                          border: Border.all(
                            color: index < pin.length
                                ? MokposColors.primary
                                : MokposColors.line,
                            width: 2,
                          ),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 32),
                  _PinPad(onTap: _tap, onDelete: _delete),
                  const SizedBox(height: 20),
                  TextButton.icon(
                    onPressed: () => context.go('/login'),
                    icon: const Icon(LucideIcons.arrowLeft, size: 18),
                    label: const Text('Ganti kasir'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _PinPad extends StatelessWidget {
  const _PinPad({required this.onTap, required this.onDelete});

  final ValueChanged<String> onTap;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    final keys = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '', '0', 'del'];
    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: keys.length,
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 3,
        mainAxisSpacing: 14,
        crossAxisSpacing: 14,
        childAspectRatio: 1.45,
      ),
      itemBuilder: (context, index) {
        final key = keys[index];
        if (key.isEmpty) return const SizedBox.shrink();
        return FilledButton.tonal(
          onPressed: key == 'del' ? onDelete : () => onTap(key),
          style: FilledButton.styleFrom(
            backgroundColor: Colors.white,
            foregroundColor: MokposColors.text,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(MokposRadius.md),
              side: const BorderSide(color: MokposColors.line),
            ),
            textStyle: const TextStyle(
              fontSize: 24,
              fontWeight: FontWeight.w900,
            ),
          ),
          child: key == 'del'
              ? const Icon(LucideIcons.delete, size: 24)
              : Text(key),
        );
      },
    );
  }
}
