import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/theme.dart';
import '../../../shared/models/nojpos_models.dart';
import '../providers/inventory_operations_provider.dart';
import '../repositories/inventory_repository.dart';

class InventoryActionPanel extends ConsumerStatefulWidget {
  const InventoryActionPanel({
    required this.canManage,
    required this.outletId,
    required this.outlets,
    required this.items,
    required this.onChanged,
    super.key,
  });

  final bool canManage;
  final String outletId;
  final List<Outlet> outlets;
  final List<InventoryStockItem> items;
  final Future<void> Function() onChanged;

  @override
  ConsumerState<InventoryActionPanel> createState() =>
      _InventoryActionPanelState();
}

class _InventoryActionPanelState extends ConsumerState<InventoryActionPanel> {
  final countController = TextEditingController();
  final wasteQuantityController = TextEditingController(text: '1');
  final transferQuantityController = TextEditingController(text: '1');
  final reasonController = TextEditingController();
  String? selectedProductId;
  String? destinationOutletId;

  @override
  void initState() {
    super.initState();
    selectedProductId = widget.items.isNotEmpty
        ? widget.items.first.productId
        : null;
    destinationOutletId = widget.outlets
        .where((outlet) => outlet.id != widget.outletId)
        .map((outlet) => outlet.id)
        .firstOrNull;
  }

  @override
  void didUpdateWidget(covariant InventoryActionPanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (selectedProductId == null ||
        !widget.items.any((item) => item.productId == selectedProductId)) {
      selectedProductId = widget.items.isNotEmpty
          ? widget.items.first.productId
          : null;
    }
    if (destinationOutletId == null ||
        !widget.outlets.any((outlet) => outlet.id == destinationOutletId)) {
      destinationOutletId = widget.outlets
          .where((outlet) => outlet.id != widget.outletId)
          .map((outlet) => outlet.id)
          .firstOrNull;
    }
  }

  @override
  void dispose() {
    countController.dispose();
    wasteQuantityController.dispose();
    transferQuantityController.dispose();
    reasonController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final operations = ref.watch(inventoryOperationsProvider);
    final selectedProduct = widget.items
        .where((item) => item.productId == selectedProductId)
        .firstOrNull;

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(18),
        side: const BorderSide(color: MokposColors.line),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Aksi Inventory',
              style: TextStyle(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 12),
            if (!widget.canManage)
              const _MutedNotice(
                'Mode cashier read-only. Aksi stok hanya untuk owner/admin.',
              ),
            if (widget.canManage && widget.items.isEmpty)
              const _MutedNotice('Tidak ada produk untuk aksi inventory.'),
            if (widget.canManage && widget.items.isNotEmpty) ...[
              DropdownButtonFormField<String>(
                initialValue: selectedProductId,
                decoration: const InputDecoration(labelText: 'Produk'),
                items: [
                  for (final item in widget.items)
                    DropdownMenuItem(
                      value: item.productId,
                      child: Text(item.name),
                    ),
                ],
                onChanged: operations.isBusy
                    ? null
                    : (value) => setState(() => selectedProductId = value),
              ),
              const SizedBox(height: 10),
              Wrap(
                spacing: 10,
                runSpacing: 10,
                children: [
                  SizedBox(
                    width: 180,
                    child: TextField(
                      controller: countController,
                      keyboardType: TextInputType.number,
                      decoration: InputDecoration(
                        labelText: 'Stok opname',
                        hintText: selectedProduct == null
                            ? null
                            : '${selectedProduct.stockOnHand}',
                      ),
                    ),
                  ),
                  SizedBox(
                    width: 220,
                    child: TextField(
                      controller: reasonController,
                      decoration: const InputDecoration(
                        labelText: 'Alasan / catatan',
                      ),
                    ),
                  ),
                  FilledButton(
                    onPressed: operations.isBusy ? null : () => _submitCount(),
                    child: const Text('Simpan Opname'),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              Wrap(
                spacing: 10,
                runSpacing: 10,
                children: [
                  SizedBox(
                    width: 160,
                    child: TextField(
                      controller: wasteQuantityController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(
                        labelText: 'Qty terbuang',
                      ),
                    ),
                  ),
                  FilledButton.tonal(
                    onPressed: operations.isBusy ? null : () => _submitWaste(),
                    child: const Text('Catat Waste'),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              Wrap(
                spacing: 10,
                runSpacing: 10,
                crossAxisAlignment: WrapCrossAlignment.center,
                children: [
                  SizedBox(
                    width: 220,
                    child: DropdownButtonFormField<String>(
                      initialValue: destinationOutletId,
                      decoration: const InputDecoration(
                        labelText: 'Tujuan transfer',
                      ),
                      items: [
                        for (final outlet in widget.outlets.where(
                          (outlet) => outlet.id != widget.outletId,
                        ))
                          DropdownMenuItem(
                            value: outlet.id,
                            child: Text(outlet.name),
                          ),
                      ],
                      onChanged: operations.isBusy
                          ? null
                          : (value) =>
                                setState(() => destinationOutletId = value),
                    ),
                  ),
                  SizedBox(
                    width: 150,
                    child: TextField(
                      controller: transferQuantityController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(
                        labelText: 'Qty transfer',
                      ),
                    ),
                  ),
                  FilledButton.tonal(
                    onPressed: operations.isBusy
                        ? null
                        : () => _submitTransfer(),
                    child: const Text('Buat Transfer'),
                  ),
                ],
              ),
            ],
            if (operations.errorMessage != null) ...[
              const SizedBox(height: 10),
              Text(
                operations.errorMessage!,
                style: const TextStyle(color: MokposColors.danger),
              ),
            ],
            if (operations.successMessage != null) ...[
              const SizedBox(height: 10),
              Text(
                operations.successMessage!,
                style: const TextStyle(color: MokposColors.success),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Future<void> _submitCount() async {
    final productId = selectedProductId;
    final counted = int.tryParse(countController.text.trim());
    if (productId == null || counted == null) return;
    final result = await ref
        .read(inventoryOperationsProvider.notifier)
        .createCount(
          outletId: widget.outletId,
          productId: productId,
          countedQuantity: counted,
          reason: reasonController.text.trim().isEmpty
              ? null
              : reasonController.text.trim(),
        );
    if (result != null) await widget.onChanged();
  }

  Future<void> _submitWaste() async {
    final productId = selectedProductId;
    final quantity = int.tryParse(wasteQuantityController.text.trim());
    final reason = reasonController.text.trim();
    if (productId == null || quantity == null || reason.isEmpty) return;
    final result = await ref
        .read(inventoryOperationsProvider.notifier)
        .createWaste(
          outletId: widget.outletId,
          productId: productId,
          quantity: quantity,
          reason: reason,
        );
    if (result != null) await widget.onChanged();
  }

  Future<void> _submitTransfer() async {
    final productId = selectedProductId;
    final destinationId = destinationOutletId;
    final quantity = int.tryParse(transferQuantityController.text.trim());
    if (productId == null || destinationId == null || quantity == null) return;
    final result = await ref
        .read(inventoryOperationsProvider.notifier)
        .createTransfer(
          sourceOutletId: widget.outletId,
          destinationOutletId: destinationId,
          productId: productId,
          quantity: quantity,
          notes: reasonController.text.trim().isEmpty
              ? null
              : reasonController.text.trim(),
        );
    if (result != null) await widget.onChanged();
  }
}

class InventoryMovementPanel extends StatelessWidget {
  const InventoryMovementPanel({required this.movements, super.key});

  final List<StockMovement> movements;

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(18),
        side: const BorderSide(color: MokposColors.line),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Pergerakan Stok',
              style: TextStyle(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 8),
            if (movements.isEmpty)
              const _MutedNotice('Belum ada pergerakan stok.')
            else
              for (final movement in movements.take(6))
                ListTile(
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  title: Text(movement.productName ?? movement.productId),
                  subtitle: Text(
                    '${movement.type}${movement.reason == null ? '' : ' · ${movement.reason}'}',
                  ),
                  trailing: Text(
                    movement.quantityDelta > 0
                        ? '+${movement.quantityDelta}'
                        : '${movement.quantityDelta}',
                    style: const TextStyle(fontWeight: FontWeight.w800),
                  ),
                ),
          ],
        ),
      ),
    );
  }
}

class InTransitTransfersPanel extends ConsumerWidget {
  const InTransitTransfersPanel({
    required this.canManage,
    required this.transfers,
    required this.onChanged,
    super.key,
  });

  final bool canManage;
  final List<InventoryTransferResult> transfers;
  final Future<void> Function() onChanged;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final operations = ref.watch(inventoryOperationsProvider);
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(18),
        side: const BorderSide(color: MokposColors.line),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Transfer In-Transit',
              style: TextStyle(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 8),
            if (transfers.isEmpty)
              const _MutedNotice('Tidak ada transfer yang sedang dikirim.')
            else
              for (final transfer in transfers)
                ListTile(
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  title: Text(
                    '${transfer.number} · ${transfer.sourceOutletName} → ${transfer.destinationOutletName}',
                  ),
                  subtitle: Text(_transferLineText(transfer)),
                  trailing: canManage
                      ? Wrap(
                          spacing: 8,
                          children: [
                            OutlinedButton(
                              onPressed: operations.isBusy
                                  ? null
                                  : () => _run(
                                      ref
                                          .read(
                                            inventoryOperationsProvider
                                                .notifier,
                                          )
                                          .sendTransfer(transfer.id),
                                    ),
                              child: const Text('Kirim'),
                            ),
                            FilledButton.tonal(
                              onPressed: operations.isBusy
                                  ? null
                                  : () => _run(
                                      ref
                                          .read(
                                            inventoryOperationsProvider
                                                .notifier,
                                          )
                                          .receiveTransfer(transfer.id),
                                    ),
                              child: const Text('Terima'),
                            ),
                          ],
                        )
                      : null,
                ),
          ],
        ),
      ),
    );
  }

  Future<void> _run(Future<InventoryTransferResult?> action) async {
    final result = await action;
    if (result != null) await onChanged();
  }

  String _transferLineText(InventoryTransferResult transfer) {
    if (transfer.lines.isEmpty) return transfer.status;
    final line = transfer.lines.first;
    return '${line.productName}: transit ${line.inTransitQuantity} dari ${line.sentQuantity}';
  }
}

class _MutedNotice extends StatelessWidget {
  const _MutedNotice(this.text);

  final String text;

  @override
  Widget build(BuildContext context) {
    return Text(text, style: const TextStyle(color: MokposColors.muted));
  }
}
