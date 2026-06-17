import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../../pos/formatters.dart';
import '../../pos/providers/pos_providers.dart';

class SuccessScreen extends ConsumerWidget {
  const SuccessScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final transaction = ref.watch(nojposSessionProvider).lastTransaction;
    final fallbackTotal = ref.watch(cartTotalProvider);
    final total = transaction?.order.total ?? fallbackTotal;
    final paid = transaction?.paidAmount ?? total;
    final change = transaction?.change ?? 0;
    final payments = transaction?.payments ?? const [];
    final paymentLabel = payments.isEmpty
        ? 'Tunai'
        : payments.first.method.label;

    return Scaffold(
      backgroundColor: MokposColors.primary,
      body: SafeArea(
        child: Center(
          child: Container(
            width: 520,
            constraints: const BoxConstraints(maxHeight: 620),
            color: Colors.white,
            child: Column(
              children: [
                Container(
                  height: 56,
                  color: MokposColors.primary.withValues(alpha: .82),
                  alignment: Alignment.center,
                  child: const Text(
                    'NojPOS',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 22,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.fromLTRB(34, 22, 34, 22),
                    child: Column(
                      children: [
                        Container(
                          width: 22,
                          height: 22,
                          decoration: const BoxDecoration(
                            color: Color(0xFF67D875),
                            shape: BoxShape.circle,
                          ),
                          child: const Icon(
                            LucideIcons.check,
                            color: Colors.white,
                            size: 15,
                          ),
                        ),
                        const SizedBox(height: 10),
                        const Text(
                          'Sukses!',
                          style: TextStyle(
                            color: MokposColors.text,
                            fontSize: 20,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                        const SizedBox(height: 28),
                        _ReceiptRow(
                          label: 'Total Tagihan',
                          value: rupiah(total),
                        ),
                        const SizedBox(height: 16),
                        _ReceiptRow(label: paymentLabel, value: rupiah(paid)),
                        const SizedBox(height: 28),
                        const Divider(color: MokposColors.line),
                        const SizedBox(height: 16),
                        _ReceiptRow(label: 'Kembalian', value: rupiah(change)),
                        const Spacer(),
                        Row(
                          children: [
                            Expanded(
                              child: OutlinedButton.icon(
                                onPressed: () {
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    const SnackBar(
                                      content: Text('Struk siap dibagikan'),
                                    ),
                                  );
                                },
                                icon: const Icon(LucideIcons.share2, size: 18),
                                label: const Text('Bagikan Struk'),
                                style: _outlineStyle(),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: OutlinedButton.icon(
                                onPressed: () {
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    const SnackBar(
                                      content: Text('Struk dikirim ke printer'),
                                    ),
                                  );
                                },
                                icon: const Icon(LucideIcons.printer, size: 18),
                                label: const Text('Cetak'),
                                style: _outlineStyle(),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        FilledButton(
                          onPressed: () {
                            ref.read(cartProvider.notifier).clear();
                            context.go('/pos');
                          },
                          style: FilledButton.styleFrom(
                            backgroundColor: MokposColors.primary,
                            foregroundColor: Colors.white,
                            minimumSize: const Size.fromHeight(46),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(
                                MokposRadius.sm,
                              ),
                            ),
                          ),
                          child: const Text(
                            'Selesai',
                            style: TextStyle(fontWeight: FontWeight.w900),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  ButtonStyle _outlineStyle() {
    return OutlinedButton.styleFrom(
      foregroundColor: MokposColors.primary,
      side: const BorderSide(color: MokposColors.primary),
      minimumSize: const Size.fromHeight(46),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      textStyle: const TextStyle(fontWeight: FontWeight.w900),
    );
  }
}

class _ReceiptRow extends StatelessWidget {
  const _ReceiptRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Text(
          label,
          style: const TextStyle(
            color: MokposColors.text,
            fontWeight: FontWeight.w800,
            fontSize: 15,
          ),
        ),
        const Spacer(),
        Text(
          value,
          style: const TextStyle(
            color: MokposColors.text,
            fontWeight: FontWeight.w900,
            fontSize: 15,
          ),
        ),
      ],
    );
  }
}
