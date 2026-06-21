import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../providers/store_controller.dart';
import '../repositories/store_repository.dart';

class StoreStatePanel extends ConsumerStatefulWidget {
  const StoreStatePanel({super.key});

  @override
  ConsumerState<StoreStatePanel> createState() => _StoreStatePanelState();
}

class _StoreStatePanelState extends ConsumerState<StoreStatePanel> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _load());
  }

  void _load() {
    final outletId = ref.read(nojposSessionProvider).outlet.id;
    if (outletId.isNotEmpty) {
      ref.read(storeControllerProvider.notifier).load(outletId);
    }
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final state = ref.watch(storeControllerProvider);
    final store = state.storeState;
    final canManage = session.canManageMasterData;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Expanded(
                child: Text(
                  'Status toko',
                  style: TextStyle(fontWeight: FontWeight.w900, fontSize: 16),
                ),
              ),
              if (state.isBusy)
                const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              else
                OutlinedButton(
                  onPressed: _load,
                  child: const Text('Muat ulang'),
                ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            store == null
                ? 'Status toko belum dimuat.'
                : '${store.outletName}: ${store.isOpen ? 'Buka' : 'Tutup'}',
            style: TextStyle(
              color: store?.isClosed == true
                  ? MokposColors.danger
                  : MokposColors.success,
              fontWeight: FontWeight.w900,
            ),
          ),
          if (store?.enabled == false) ...[
            const SizedBox(height: 8),
            const Text(
              'Buka/tutup toko belum tersedia untuk outlet ini.',
              style: TextStyle(color: MokposColors.muted),
            ),
          ],
          if (!canManage) ...[
            const SizedBox(height: 8),
            const Text(
              'Mode cashier read-only. Buka/tutup toko hanya untuk owner/admin.',
              style: TextStyle(color: MokposColors.danger),
            ),
          ],
          if (state.errorMessage != null) ...[
            const SizedBox(height: 8),
            Text(
              state.errorMessage!,
              style: const TextStyle(
                color: MokposColors.danger,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
          if (state.blockingShifts.isNotEmpty) ...[
            const SizedBox(height: 8),
            _BlockingShiftList(shifts: state.blockingShifts),
          ],
          const SizedBox(height: 12),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              FilledButton(
                onPressed:
                    canManage &&
                        store?.enabled == true &&
                        store?.isClosed == true
                    ? () => _showOperationDialog(open: true)
                    : null,
                child: const Text('Buka toko'),
              ),
              OutlinedButton(
                onPressed:
                    canManage && store?.enabled == true && store?.isOpen == true
                    ? () => _showOperationDialog(open: false)
                    : null,
                child: const Text('Tutup toko'),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Future<void> _showOperationDialog({required bool open}) async {
    final result = await showDialog<_StoreOperationInput>(
      context: context,
      builder: (context) => _StoreOperationDialog(open: open),
    );
    if (result == null || !mounted) return;

    final outletId = ref.read(nojposSessionProvider).outlet.id;
    final notifier = ref.read(storeControllerProvider.notifier);
    final ok = open
        ? await notifier.openStore(
            outletId: outletId,
            pin: result.pin,
            reason: result.reason,
          )
        : await notifier.closeStore(
            outletId: outletId,
            pin: result.pin,
            reason: result.reason,
          );

    if (!mounted) return;
    final state = ref.read(storeControllerProvider);
    final message = ok
        ? (open ? 'Toko berhasil dibuka.' : 'Toko berhasil ditutup.')
        : (state.errorMessage ?? 'Operasi toko gagal.');
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }
}

class _BlockingShiftList extends StatelessWidget {
  const _BlockingShiftList({required this.shifts});

  final List<BlockingShift> shifts;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'Shift yang menghambat tutup toko:',
          style: TextStyle(fontWeight: FontWeight.w900),
        ),
        const SizedBox(height: 4),
        for (final shift in shifts)
          Text(
            '• ${shift.status} — ${shift.cashierName ?? 'Kasir'} / ${shift.deviceName ?? shift.shiftId}',
            style: const TextStyle(color: MokposColors.muted),
          ),
      ],
    );
  }
}

class _StoreOperationInput {
  const _StoreOperationInput({required this.pin, this.reason});

  final String pin;
  final String? reason;
}

class _StoreOperationDialog extends StatefulWidget {
  const _StoreOperationDialog({required this.open});

  final bool open;

  @override
  State<_StoreOperationDialog> createState() => _StoreOperationDialogState();
}

class _StoreOperationDialogState extends State<_StoreOperationDialog> {
  final pinController = TextEditingController();
  final reasonController = TextEditingController();
  String? error;

  @override
  void dispose() {
    pinController.clear();
    reasonController.dispose();
    pinController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(widget.open ? 'Buka toko' : 'Tutup toko'),
      content: SizedBox(
        width: 360,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: reasonController,
              decoration: const InputDecoration(labelText: 'Alasan / catatan'),
              textInputAction: TextInputAction.next,
            ),
            const SizedBox(height: 10),
            TextField(
              controller: pinController,
              decoration: const InputDecoration(labelText: 'PIN otorisasi'),
              obscureText: true,
              keyboardType: TextInputType.number,
            ),
            if (error != null) ...[
              const SizedBox(height: 8),
              Text(error!, style: const TextStyle(color: MokposColors.danger)),
            ],
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Batal'),
        ),
        FilledButton(
          onPressed: () {
            final pin = pinController.text.trim();
            if (pin.isEmpty) {
              setState(() => error = 'PIN otorisasi wajib diisi.');
              return;
            }
            Navigator.of(context).pop(
              _StoreOperationInput(
                pin: pin,
                reason: reasonController.text.trim().isEmpty
                    ? null
                    : reasonController.text.trim(),
              ),
            );
          },
          child: Text(widget.open ? 'Buka' : 'Tutup'),
        ),
      ],
    );
  }
}
