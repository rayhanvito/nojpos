import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../providers/terminal_lock_controller.dart';
import '../repositories/terminal_lock_repository.dart';

class TerminalLockGate extends ConsumerStatefulWidget {
  const TerminalLockGate({required this.child, super.key});

  final Widget child;

  @override
  ConsumerState<TerminalLockGate> createState() => _TerminalLockGateState();
}

class _TerminalLockGateState extends ConsumerState<TerminalLockGate> {
  Timer? _idleTimer;

  @override
  void initState() {
    super.initState();
    Future<void>.microtask(() {
      if (!mounted) return;
      ref.read(terminalLockControllerProvider.notifier).load();
      _resetIdleTimer();
    });
  }

  @override
  void dispose() {
    _idleTimer?.cancel();
    super.dispose();
  }

  void _resetIdleTimer() {
    _idleTimer?.cancel();
    final lockState = ref.read(terminalLockControllerProvider).lockState;
    if (lockState?.locked ?? false) return;
    final timeout = lockState?.idleTimeoutSeconds ?? 180;
    if (timeout <= 0) return;
    _idleTimer = Timer(Duration(seconds: timeout), () {
      if (!mounted) return;
      ref
          .read(terminalLockControllerProvider.notifier)
          .lock(reason: TerminalLockReason.idleTimeout);
    });
  }

  @override
  Widget build(BuildContext context) {
    final lockState = ref.watch(terminalLockControllerProvider);
    ref.listen<TerminalLockControllerState>(terminalLockControllerProvider, (
      previous,
      next,
    ) {
      final previousLock = previous?.lockState;
      final nextLock = next.lockState;
      if ((previousLock?.locked ?? false) != (nextLock?.locked ?? false) ||
          previousLock?.idleTimeoutSeconds != nextLock?.idleTimeoutSeconds) {
        _resetIdleTimer();
      }
    });

    if (lockState.isLocked) {
      _idleTimer?.cancel();
      return _TerminalLockedScreen(state: lockState);
    }

    return Listener(
      behavior: HitTestBehavior.translucent,
      onPointerDown: (_) => _resetIdleTimer(),
      onPointerMove: (_) => _resetIdleTimer(),
      child: widget.child,
    );
  }
}

class _TerminalLockedScreen extends ConsumerStatefulWidget {
  const _TerminalLockedScreen({required this.state});

  final TerminalLockControllerState state;

  @override
  ConsumerState<_TerminalLockedScreen> createState() =>
      _TerminalLockedScreenState();
}

class _TerminalLockedScreenState extends ConsumerState<_TerminalLockedScreen> {
  final TextEditingController pinController = TextEditingController();

  @override
  void dispose() {
    pinController.clear();
    pinController.dispose();
    super.dispose();
  }

  Future<void> _unlock() async {
    final pin = pinController.text.trim();
    pinController.clear();
    if (pin.isEmpty) return;
    await ref.read(terminalLockControllerProvider.notifier).unlock(pin);
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final state = widget.state;
    final lockState = state.lockState;
    final lockout = lockState?.lockoutUntil;
    return Scaffold(
      backgroundColor: MokposColors.primaryDark,
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 460),
            child: Card(
              elevation: 0,
              color: Colors.white,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(MokposRadius.lg),
              ),
              child: Padding(
                padding: const EdgeInsets.all(28),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    const Icon(
                      LucideIcons.lockKeyhole,
                      color: MokposColors.primary,
                      size: 54,
                    ),
                    const SizedBox(height: 16),
                    const Text(
                      'Terminal terkunci',
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontSize: 24,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      session.outlet.name,
                      textAlign: TextAlign.center,
                      style: const TextStyle(color: MokposColors.muted),
                    ),
                    const SizedBox(height: 20),
                    if (state.errorMessage != null) ...[
                      _LockMessage(text: state.errorMessage!, isError: true),
                      const SizedBox(height: 12),
                    ],
                    if (lockout != null) ...[
                      _LockMessage(
                        text: 'PIN terkunci sampai ${_time(lockout)}.',
                        isError: true,
                      ),
                      const SizedBox(height: 12),
                    ] else if (lockState != null) ...[
                      _LockMessage(
                        text:
                            'Sisa percobaan PIN: ${lockState.failedAttemptsRemaining}',
                      ),
                      const SizedBox(height: 12),
                    ],
                    TextField(
                      key: const ValueKey('terminal_lock_pin'),
                      controller: pinController,
                      keyboardType: TextInputType.number,
                      obscureText: true,
                      decoration: const InputDecoration(
                        labelText: 'PIN staff',
                        border: OutlineInputBorder(),
                      ),
                      onSubmitted: (_) => _unlock(),
                    ),
                    const SizedBox(height: 14),
                    FilledButton.icon(
                      key: const ValueKey('terminal_unlock_submit'),
                      onPressed: state.isBusy ? null : _unlock,
                      icon: state.isBusy
                          ? const SizedBox(
                              width: 16,
                              height: 16,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(LucideIcons.unlockKeyhole),
                      label: const Text('Buka terminal'),
                    ),
                    if (state.unlockResponse?.handoverRequired ?? false) ...[
                      const SizedBox(height: 14),
                      const _LockMessage(
                        text:
                            'Staff berbeda terdeteksi. Lanjutkan sebagai kasir saat ini atau lakukan handover sesuai policy.',
                      ),
                    ],
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _LockMessage extends StatelessWidget {
  const _LockMessage({required this.text, this.isError = false});

  final String text;
  final bool isError;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: isError
            ? MokposColors.danger.withValues(alpha: .08)
            : MokposColors.surface,
        borderRadius: BorderRadius.circular(MokposRadius.sm),
        border: Border.all(
          color: isError ? MokposColors.danger : MokposColors.line,
        ),
      ),
      child: Text(
        text,
        textAlign: TextAlign.center,
        style: TextStyle(
          color: isError ? MokposColors.danger : MokposColors.text,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

String _time(DateTime value) {
  final local = value.toLocal();
  return '${local.hour.toString().padLeft(2, '0')}:${local.minute.toString().padLeft(2, '0')}';
}
