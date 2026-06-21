import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../../shared/models/nojpos_models.dart';

class OutletSelectScreen extends ConsumerStatefulWidget {
  const OutletSelectScreen({this.allowChange = false, super.key});

  final bool allowChange;

  @override
  ConsumerState<OutletSelectScreen> createState() => _OutletSelectScreenState();
}

class _OutletSelectScreenState extends ConsumerState<OutletSelectScreen> {
  bool _started = false;

  @override
  void initState() {
    super.initState();
    Future<void>.microtask(_prepare);
  }

  Future<void> _prepare() async {
    if (_started) return;
    _started = true;
    await ref.read(nojposSessionProvider.notifier).refreshOutlets();
    if (!mounted) return;
    final session = ref.read(nojposSessionProvider);
    if (!widget.allowChange && session.outlet.id.isNotEmpty) {
      await _goToPinOrSync();
      return;
    }
    if (!widget.allowChange && session.outlets.length == 1) {
      await _select(session.outlets.single);
    }
  }

  Future<void> _select(Outlet outlet) async {
    await ref.read(nojposSessionProvider.notifier).selectOutlet(outlet);
    if (!mounted) return;
    await _goToPinOrSync();
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

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 560),
            child: Padding(
              padding: const EdgeInsets.all(28),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const _OutletHeader(),
                  const SizedBox(height: 24),
                  if (session.isBusy)
                    const Center(child: CircularProgressIndicator())
                  else if (session.outlets.isEmpty)
                    _EmptyOutletState(
                      message:
                          session.errorMessage ??
                          'Tidak ada outlet yang tersedia untuk akun ini.',
                      onRetry: () => ref
                          .read(nojposSessionProvider.notifier)
                          .refreshOutlets(),
                    )
                  else
                    _OutletList(outlets: session.outlets, onSelect: _select),
                  if (session.errorMessage != null &&
                      session.outlets.isNotEmpty) ...[
                    const SizedBox(height: 14),
                    Text(
                      session.errorMessage!,
                      style: const TextStyle(color: MokposColors.danger),
                    ),
                  ],
                  const SizedBox(height: 18),
                  TextButton.icon(
                    onPressed: () async {
                      await ref.read(nojposSessionProvider.notifier).logout();
                      if (!context.mounted) return;
                      context.go('/login');
                    },
                    icon: const Icon(LucideIcons.logOut, size: 18),
                    label: const Text('Logout akun'),
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

class _OutletHeader extends StatelessWidget {
  const _OutletHeader();

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          width: 64,
          height: 64,
          decoration: BoxDecoration(
            color: MokposColors.primarySoft,
            borderRadius: BorderRadius.circular(18),
          ),
          child: const Icon(
            LucideIcons.store,
            color: MokposColors.primary,
            size: 32,
          ),
        ),
        const SizedBox(height: 18),
        Text(
          'Pilih Outlet',
          style: Theme.of(context).textTheme.headlineMedium?.copyWith(
            color: MokposColors.text,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 8),
        const Text(
          'Pilih outlet kerja sebelum masuk ke PIN kasir.',
          style: TextStyle(color: MokposColors.muted, fontSize: 15),
        ),
      ],
    );
  }
}

class _OutletList extends StatelessWidget {
  const _OutletList({required this.outlets, required this.onSelect});

  final List<Outlet> outlets;
  final ValueChanged<Outlet> onSelect;

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      shrinkWrap: true,
      itemCount: outlets.length,
      separatorBuilder: (context, index) => const SizedBox(height: 10),
      itemBuilder: (context, index) {
        final outlet = outlets[index];
        return Material(
          color: Colors.white,
          borderRadius: BorderRadius.circular(MokposRadius.md),
          child: InkWell(
            key: ValueKey('outlet_${outlet.id}'),
            onTap: () => onSelect(outlet),
            borderRadius: BorderRadius.circular(MokposRadius.md),
            child: Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                border: Border.all(color: MokposColors.line),
                borderRadius: BorderRadius.circular(MokposRadius.md),
              ),
              child: Row(
                children: [
                  CircleAvatar(
                    backgroundColor: MokposColors.primarySoft,
                    child: Text(
                      outlet.name.characters.first.toUpperCase(),
                      style: const TextStyle(
                        color: MokposColors.primary,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Text(
                      outlet.name,
                      style: const TextStyle(
                        color: MokposColors.text,
                        fontWeight: FontWeight.w900,
                        fontSize: 15,
                      ),
                    ),
                  ),
                  const Icon(
                    LucideIcons.arrowRight,
                    color: MokposColors.muted,
                    size: 20,
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}

class _EmptyOutletState extends StatelessWidget {
  const _EmptyOutletState({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(
          message,
          textAlign: TextAlign.center,
          style: const TextStyle(color: MokposColors.muted),
        ),
        const SizedBox(height: 12),
        OutlinedButton.icon(
          onPressed: onRetry,
          icon: const Icon(LucideIcons.refreshCcw, size: 18),
          label: const Text('Muat ulang outlet'),
        ),
      ],
    );
  }
}
