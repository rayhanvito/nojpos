import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';

class ShiftGateScreen extends ConsumerStatefulWidget {
  const ShiftGateScreen({super.key});

  @override
  ConsumerState<ShiftGateScreen> createState() => _ShiftGateScreenState();
}

class _ShiftGateScreenState extends ConsumerState<ShiftGateScreen> {
  final openingCashController = TextEditingController(text: '100000');
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
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 430),
            child: Padding(
              padding: const EdgeInsets.all(28),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    session.hasOpenShift ? 'Shift aktif' : 'Buka Shift',
                    style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      color: MokposColors.text,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Outlet aktif: ${session.outlet.name}\nKasir aktif: ${session.cashier.name}',
                    style: const TextStyle(color: MokposColors.muted),
                  ),
                  const SizedBox(height: 26),
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
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 10),
                    Text(
                      'Masukkan uang tunai fisik di laci saat shift dibuka. Nilai ini dikirim sebagai opening_cash ke server.',
                      style: const TextStyle(color: MokposColors.muted),
                    ),
                    if (session.errorMessage != null) ...[
                      const SizedBox(height: 14),
                      Text(
                        session.errorMessage!,
                        style: const TextStyle(color: MokposColors.danger),
                      ),
                    ],
                    const SizedBox(height: 24),
                    FilledButton(
                      key: const ValueKey('shift_open_submit'),
                      onPressed: _openShift,
                      style: FilledButton.styleFrom(
                        minimumSize: const Size.fromHeight(54),
                        backgroundColor: MokposColors.primary,
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
    );
  }
}
