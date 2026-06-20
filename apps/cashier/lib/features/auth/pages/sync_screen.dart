import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../attendance/providers/attendance_controller.dart';
import '../../pos/providers/pos_providers.dart';

class SyncScreen extends ConsumerStatefulWidget {
  const SyncScreen({super.key});

  @override
  ConsumerState<SyncScreen> createState() => _SyncScreenState();
}

class _SyncScreenState extends ConsumerState<SyncScreen> {
  final List<_SyncStep> steps = const [
    _SyncStep('catalog', 'Mengambil produk dan kategori'),
    _SyncStep('attendance', 'Mengambil staff dan absensi'),
    _SyncStep('shift', 'Mengecek shift outlet'),
  ];
  final Set<String> completed = {};
  String? activeStep;
  String? errorMessage;

  @override
  void initState() {
    super.initState();
    Future<void>.microtask(_sync);
  }

  Future<void> _sync() async {
    setState(() {
      completed.clear();
      activeStep = null;
      errorMessage = null;
    });

    final session = ref.read(nojposSessionProvider);
    if (session.outlet.id.isEmpty) {
      if (!mounted) return;
      context.go('/outlet');
      return;
    }

    try {
      await _runStep('catalog', () {
        return ref.read(posCatalogProvider.notifier).load();
      });
      await _runStep('attendance', () {
        return ref.read(attendanceControllerProvider.notifier).load();
      });
      await _runStep('shift', () {
        return ref.read(nojposSessionProvider.notifier).refreshCurrentShift();
      });
    } catch (error) {
      if (!mounted) return;
      setState(() {
        activeStep = null;
        errorMessage = _messageForUi(error);
      });
      return;
    }

    if (!mounted) return;
    final status = ref.read(nojposSessionProvider).status;
    if (status == SessionStatus.unauthenticated) {
      context.go('/login');
      return;
    }
    context.go('/pin');
  }

  Future<void> _runStep(String id, Future<void> Function() action) async {
    if (!mounted) return;
    setState(() => activeStep = id);
    await action();
    if (!mounted) return;
    setState(() {
      completed.add(id);
      activeStep = null;
    });
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 520),
            child: Padding(
              padding: const EdgeInsets.all(28),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const _SyncIcon(),
                  const SizedBox(height: 24),
                  Text(
                    'Menyinkronkan Data Outlet',
                    textAlign: TextAlign.center,
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      color: MokposColors.text,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    session.outlet.id.isEmpty
                        ? 'Menyiapkan outlet kerja.'
                        : '${session.outlet.name} · ${session.businessName}',
                    textAlign: TextAlign.center,
                    style: const TextStyle(color: MokposColors.muted),
                  ),
                  const SizedBox(height: 28),
                  DecoratedBox(
                    decoration: BoxDecoration(
                      color: Colors.white,
                      border: Border.all(color: MokposColors.line),
                      borderRadius: BorderRadius.circular(MokposRadius.md),
                    ),
                    child: Column(
                      children: [
                        for (final step in steps)
                          _SyncStepTile(
                            label: step.label,
                            isActive: activeStep == step.id,
                            isDone: completed.contains(step.id),
                          ),
                      ],
                    ),
                  ),
                  if (errorMessage != null) ...[
                    const SizedBox(height: 18),
                    Text(
                      errorMessage!,
                      textAlign: TextAlign.center,
                      style: const TextStyle(color: MokposColors.danger),
                    ),
                    const SizedBox(height: 14),
                    OutlinedButton.icon(
                      onPressed: _sync,
                      icon: const Icon(LucideIcons.refreshCw, size: 18),
                      label: const Text('Coba Lagi'),
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

class _SyncStep {
  const _SyncStep(this.id, this.label);

  final String id;
  final String label;
}

class _SyncIcon extends StatelessWidget {
  const _SyncIcon();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        width: 76,
        height: 76,
        decoration: BoxDecoration(
          color: MokposColors.primarySoft,
          borderRadius: BorderRadius.circular(24),
        ),
        child: const Icon(
          LucideIcons.cloudCog,
          color: MokposColors.primary,
          size: 36,
        ),
      ),
    );
  }
}

class _SyncStepTile extends StatelessWidget {
  const _SyncStepTile({
    required this.label,
    required this.isActive,
    required this.isDone,
  });

  final String label;
  final bool isActive;
  final bool isDone;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: isDone
          ? const Icon(LucideIcons.circleCheck, color: MokposColors.accent)
          : isActive
          ? const SizedBox.square(
              dimension: 22,
              child: CircularProgressIndicator(strokeWidth: 2),
            )
          : const Icon(LucideIcons.circle, color: MokposColors.muted),
      title: Text(
        label,
        style: const TextStyle(
          color: MokposColors.text,
          fontWeight: FontWeight.w800,
        ),
      ),
      subtitle: Text(
        isDone
            ? 'Selesai'
            : isActive
            ? 'Sedang diproses'
            : 'Menunggu',
        style: const TextStyle(color: MokposColors.muted),
      ),
    );
  }
}

String _messageForUi(Object error) {
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}
