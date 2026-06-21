import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/nojpos_assets.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../../shared/widgets/nojpos_state_view.dart';

class ShiftGateScreen extends ConsumerStatefulWidget {
  const ShiftGateScreen({super.key});

  @override
  ConsumerState<ShiftGateScreen> createState() => _ShiftGateScreenState();
}

class _ShiftGateScreenState extends ConsumerState<ShiftGateScreen> {
  final openingCashController = TextEditingController();
  bool checked = false;

  @override
  void initState() {
    super.initState();
    Future<void>.microtask(_refresh);
  }

  @override
  void dispose() {
    openingCashController.dispose();
    super.dispose();
  }

  Future<void> _refresh() async {
    await ref.read(nojposSessionProvider.notifier).refreshOutlets();
    await ref.read(nojposSessionProvider.notifier).refreshCurrentShift();
    if (!mounted) return;
    setState(() => checked = true);
    if (ref.read(nojposSessionProvider).hasOpenShift) {
      context.go('/pos');
    }
  }

  Future<void> _openShift() async {
    final openingCash = int.tryParse(
      openingCashController.text.replaceAll(RegExp(r'[^0-9]'), ''),
    );
    if (openingCash == null) return;
    final ok = await ref
        .read(nojposSessionProvider.notifier)
        .openShift(openingCash: openingCash);
    if (!mounted) return;
    if (ok) context.go('/pos');
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final isBusy = session.isBusy || !checked;

    return Scaffold(
      backgroundColor: NojposColors.canvas,
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 460),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: DecoratedBox(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(NojposRadius.xl),
                  border: Border.all(color: NojposColors.line),
                  boxShadow: NojposShadow.card,
                ),
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      NojposStateView.success(
                        title: session.hasOpenShift
                            ? 'Shift aktif'
                            : 'Buka Shift',
                        subtitle:
                            'Outlet aktif: ${session.outlet.name}\nKasir aktif: ${session.cashier.name}',
                        illustrationAsset: session.hasOpenShift
                            ? NojposAssets.shiftOpen
                            : NojposAssets.shiftClosed,
                        compact: true,
                      ),
                      const SizedBox(height: 22),
                      if (isBusy)
                        const Center(child: CircularProgressIndicator())
                      else ...[
                        TextField(
                          key: const ValueKey('shift_opening_cash'),
                          controller: openingCashController,
                          keyboardType: TextInputType.number,
                          decoration: InputDecoration(
                            labelText:
                                'Modal awal shift untuk ${session.outlet.name}',
                            prefixText: 'Rp ',
                            filled: true,
                            fillColor: NojposColors.canvas,
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(
                                NojposRadius.lg,
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(height: 10),
                        const Text(
                          'Masukkan uang tunai fisik di laci saat shift dibuka. Nilai ini dikirim sebagai opening_cash ke server.',
                          style: TextStyle(color: NojposColors.muted),
                        ),
                        if (session.errorMessage != null) ...[
                          const SizedBox(height: 14),
                          NojposStateView.error(
                            title: 'Shift belum bisa dibuka',
                            subtitle: session.errorMessage!,
                            illustrationAsset: NojposAssets.errorOccurred,
                            compact: true,
                          ),
                        ],
                        const SizedBox(height: 24),
                        FilledButton(
                          key: const ValueKey('shift_open_submit'),
                          onPressed: _openShift,
                          style: FilledButton.styleFrom(
                            minimumSize: const Size.fromHeight(54),
                            backgroundColor: NojposColors.primary,
                            foregroundColor: Colors.white,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(
                                NojposRadius.lg,
                              ),
                            ),
                            textStyle: const TextStyle(
                              fontWeight: FontWeight.w900,
                            ),
                          ),
                          child: const Text('Buka Shift'),
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
