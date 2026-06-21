import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/nojpos_assets.dart';
import '../../../app/theme.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../../../core/network/api_client.dart';
import '../../../core/printing/printer_settings_store.dart';
import '../../../core/printing/receipt_printing_service.dart';
import '../../../features/attendance/providers/attendance_controller.dart';
import '../../../features/connectivity/widgets/connectivity_status_chip.dart';
import '../../../features/inventory/providers/inventory_operations_provider.dart';
import '../../../features/inventory/repositories/inventory_repository.dart';
import '../../../features/inventory/widgets/inventory_operation_panels.dart';
import '../../../features/reports/providers/report_providers.dart';
import '../../../features/reports/repositories/report_repository.dart';
import '../../../features/reports/widgets/report_panels.dart';
import '../../../features/screen_lock/providers/terminal_lock_controller.dart';
import '../../../features/settings/providers/settings_providers.dart';
import '../../../features/settings/widgets/settings_detail_panels.dart';
import '../../../features/settings/widgets/settings_summary_panel.dart';
import '../../../features/staff/widgets/staff_settings_panel.dart';
import '../../../features/store/widgets/store_state_panel.dart';
import '../../../features/transactions/repositories/transaction_repository.dart';
import '../../../shared/models/nojpos_models.dart';
import '../../../shared/widgets/nojpos_dialog.dart';
import '../../../shared/widgets/nojpos_state_view.dart';
import '../../../shared/widgets/nojpos_toast.dart';
import '../formatters.dart';
import '../models/cart_item.dart';
import '../models/product.dart';
import '../providers/parked_order_lease_controller.dart';
import '../providers/pos_providers.dart';

enum OperationsPageType {
  orders,
  sales,
  reports,
  inventory,
  settings,
  attendance,
}

class OperationsScreen extends StatelessWidget {
  const OperationsScreen({required this.type, super.key});

  final OperationsPageType type;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      body: SafeArea(
        child: Column(
          children: [
            _OpsTopBar(config: _config(type)),
            Expanded(child: _bodyForType(type)),
          ],
        ),
      ),
    );
  }

  Widget _bodyForType(OperationsPageType type) {
    return switch (type) {
      OperationsPageType.orders => const _OrdersBody(),
      OperationsPageType.sales => const _SalesBody(),
      OperationsPageType.reports => const _ReportsBody(),
      OperationsPageType.inventory => const _InventoryBody(),
      OperationsPageType.settings => const _SettingsBody(),
      OperationsPageType.attendance => const _AttendanceBody(),
    };
  }

  _OpsConfig _config(OperationsPageType type) {
    return switch (type) {
      OperationsPageType.orders => const _OpsConfig(title: 'Daftar Order'),
      OperationsPageType.sales => const _OpsConfig(title: 'Penjualan'),
      OperationsPageType.reports => const _OpsConfig(
        title: 'Laporan - Ringkasan Penjualan',
      ),
      OperationsPageType.inventory => const _OpsConfig(
        title: 'Inventori - Faktur Pembelian',
        action: 'Tambah Faktur Pembelian',
      ),
      OperationsPageType.settings => const _OpsConfig(title: 'Pengaturan'),
      OperationsPageType.attendance => const _OpsConfig(
        title: 'Absensi',
        action: 'Daftar Kehadiran',
      ),
    };
  }
}

class _OpsConfig {
  const _OpsConfig({required this.title, this.action});

  final String title;
  final String? action;
}

class _OpsTopBar extends ConsumerWidget {
  const _OpsTopBar({required this.config});

  final _OpsConfig config;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Container(
      height: 56,
      color: MokposColors.primary,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      child: Row(
        children: [
          IconButton(
            key: const ValueKey('ops_back'),
            onPressed: () => context.go('/pos'),
            icon: const Icon(LucideIcons.arrowLeft),
            color: Colors.white,
          ),
          const SizedBox(width: 2),
          Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                config.title,
                style: const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.w900,
                  fontSize: 15,
                ),
              ),
              const SizedBox(height: 1),
              const ConnectivityStatusChip(
                textStyle: TextStyle(color: Color(0xE6FFFFFF), fontSize: 11),
                dotSize: 8,
              ),
            ],
          ),
          const Spacer(),
          if (config.action != null)
            InkWell(
              onTap: () => _handleAction(context, ref),
              borderRadius: BorderRadius.circular(MokposRadius.sm),
              child: Container(
                height: 42,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                decoration: BoxDecoration(
                  color: MokposColors.primaryDark,
                  borderRadius: BorderRadius.circular(MokposRadius.sm),
                ),
                alignment: Alignment.center,
                child: Text(
                  config.action!,
                  style: const TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
          const SizedBox(width: 10),
          const _TopSearch(),
        ],
      ),
    );
  }

  Future<void> _handleAction(BuildContext context, WidgetRef ref) async {
    if (config.title.startsWith('Inventori')) {
      final result =
          await showDialog<
            ({
              String supplierName,
              String productId,
              int quantity,
              int unitCost,
            })
          >(context: context, builder: (context) => const _PurchaseDialog());
      if (result == null || !context.mounted) return;
      final purchase = await ref
          .read(nojposSessionProvider.notifier)
          .addPurchase(
            supplierName: result.supplierName,
            productId: result.productId,
            quantity: result.quantity,
            unitCost: result.unitCost,
          );
      if (!context.mounted || purchase == null) return;
      NojposToast.success(
        context,
        '${purchase.number} ditambahkan',
        description: 'Faktur pembelian berhasil dicatat.',
      );
      return;
    }

    if (config.title == 'Absensi') {
      await showDialog<void>(
        context: context,
        builder: (context) => const _AttendanceListDialog(),
      );
    }
  }
}

class _TopSearch extends StatelessWidget {
  const _TopSearch();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 210,
      height: 40,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: MokposColors.primaryDark,
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      child: const Row(
        children: [
          Icon(LucideIcons.search, color: Colors.white, size: 18),
          SizedBox(width: 8),
          Text(
            'Cari...',
            style: TextStyle(color: Colors.white, fontWeight: FontWeight.w700),
          ),
        ],
      ),
    );
  }
}

class _OrdersBody extends ConsumerStatefulWidget {
  const _OrdersBody();

  @override
  ConsumerState<_OrdersBody> createState() => _OrdersBodyState();
}

class _OrdersBodyState extends ConsumerState<_OrdersBody> {
  @override
  void initState() {
    super.initState();
    Future.microtask(
      () => ref.read(nojposSessionProvider.notifier).loadHeldTransactions(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final orders = session.savedOrders;
    final isBusy = session.isBusy;
    final error = session.errorMessage;
    return Row(
      children: [
        _LeftList(
          title: 'Kategori Order',
          items: ['Semua (${orders.length})', 'Kasir (${orders.length})'],
          activeIndex: 0,
        ),
        Expanded(
          child: Column(
            children: [
              const _TableHeader(
                columns: [
                  'Jenis',
                  'Nama',
                  'No Transaksi',
                  'Tanggal',
                  'Status',
                  'Tagihan',
                ],
              ),
              const _OrdersScopeNotice(),
              Expanded(
                child: isBusy
                    ? const Center(child: CircularProgressIndicator())
                    : error != null
                    ? _LoadErrorState(
                        message: error,
                        onRetry: () => ref
                            .read(nojposSessionProvider.notifier)
                            .loadHeldTransactions(),
                      )
                    : orders.isEmpty
                    ? const _EmptyState(
                        icon: LucideIcons.receiptText,
                        title: 'Belum Ada Pesanan',
                        subtitle: 'Saatnya tingkatkan promosimu!',
                      )
                    : ListView.separated(
                        itemCount: orders.length,
                        separatorBuilder: (context, index) =>
                            const Divider(height: 1, color: MokposColors.line),
                        itemBuilder: (context, index) {
                          final order = orders[index];
                          return _DataRow(
                            onTap: () async {
                              final activated = await ref
                                  .read(
                                    parkedOrderLeaseControllerProvider.notifier,
                                  )
                                  .acquireForEdit(order.id);
                              if (!context.mounted) return;
                              if (activated == null) {
                                final message =
                                    ref
                                        .read(
                                          parkedOrderLeaseControllerProvider,
                                        )
                                        .errorMessage ??
                                    'Order tidak bisa dibuka.';
                                NojposToast.error(context, message);
                                return;
                              }
                              final products = ref.read(productsProvider);
                              ref
                                  .read(cartProvider.notifier)
                                  .replaceWith(
                                    _cartItemsFromOrder(activated, products),
                                  );
                              NojposToast.success(
                                context,
                                '${activated.number} dibuka ke cart',
                              );
                              context.go('/pos');
                            },
                            cells: [
                              order.type.label,
                              order.customer?.name ?? 'Tanpa Pelanggan',
                              order.number,
                              _formatDateTime(order.createdAt),
                              _statusLabel(order.status),
                              rupiah(order.total),
                            ],
                          );
                        },
                      ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _OrdersScopeNotice extends StatelessWidget {
  const _OrdersScopeNotice();

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.fromLTRB(12, 12, 12, 0),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAF9),
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(8),
      ),
      child: const Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(LucideIcons.info, size: 18, color: MokposColors.primary),
          SizedBox(width: 10),
          Expanded(
            child: Text(
              'Daftar ini menampilkan order tersimpan kasir dari server. Kanal order lain nonaktif di terminal kasir preview.',
              style: TextStyle(
                color: MokposColors.muted,
                fontSize: 12,
                fontWeight: FontWeight.w700,
                height: 1.35,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _SalesBody extends ConsumerStatefulWidget {
  const _SalesBody();

  @override
  ConsumerState<_SalesBody> createState() => _SalesBodyState();
}

class _SalesBodyState extends ConsumerState<_SalesBody> {
  String? statusFilter;

  @override
  void initState() {
    super.initState();
    Future.microtask(() => _loadSales(null));
  }

  Future<void> _loadSales(String? status) async {
    if (mounted) setState(() => statusFilter = status);
    await ref
        .read(nojposSessionProvider.notifier)
        .loadSalesTransactions(status: status);
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final transactions = session.transactions;
    final isBusy = session.isBusy;
    final error = session.errorMessage;
    final pendingPaymentRows = [
      for (final transaction in transactions)
        for (final payment in transaction.payments)
          if (!payment.isCash &&
              payment.status == 'pending' &&
              payment.paymentId != null)
            (transaction: transaction, payment: payment),
    ];
    final totalSales = transactions.fold(
      0,
      (sum, transaction) => sum + transaction.total,
    );
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              Expanded(
                child: _MetricCard(
                  label: 'Total Penjualan',
                  value: rupiah(totalSales),
                  color: MokposColors.primary,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: _MetricCard(
                  label: 'Total Transaksi',
                  value: '${transactions.length}',
                  color: MokposColors.accent,
                ),
              ),
            ],
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(12, 0, 12, 12),
          child: Row(
            children: [
              OutlinedButton.icon(
                key: const ValueKey('sales_filter_all'),
                onPressed: statusFilter == null ? null : () => _loadSales(null),
                icon: const Icon(LucideIcons.receiptText, size: 16),
                label: const Text('Semua Transaksi'),
              ),
              const SizedBox(width: 8),
              OutlinedButton.icon(
                key: const ValueKey('sales_filter_pending'),
                onPressed: statusFilter == 'pending'
                    ? null
                    : () => _loadSales('pending'),
                icon: const Icon(LucideIcons.clock3, size: 16),
                label: const Text('Pembayaran Pending'),
              ),
            ],
          ),
        ),
        _TableHeader(
          columns: statusFilter == 'pending'
              ? const [
                  'No Transaksi',
                  'Pelanggan',
                  'Waktu',
                  'Metode',
                  'Reference',
                  'Aksi',
                ]
              : const [
                  'Jenis',
                  'Nama',
                  'No Transaksi',
                  'Waktu',
                  'Kasir',
                  'Pembayaran',
                ],
        ),
        Expanded(
          child: isBusy
              ? const Center(child: CircularProgressIndicator())
              : error != null
              ? _LoadErrorState(
                  message: error,
                  onRetry: () => ref
                      .read(nojposSessionProvider.notifier)
                      .loadSalesTransactions(status: statusFilter),
                )
              : statusFilter == 'pending'
              ? pendingPaymentRows.isEmpty
                    ? const _EmptyState(
                        icon: LucideIcons.clock3,
                        title: 'Tidak Ada Pembayaran Pending',
                        subtitle:
                            'Pembayaran non-tunai pending dari server akan muncul di sini.',
                      )
                    : ListView.separated(
                        itemCount: pendingPaymentRows.length,
                        separatorBuilder: (context, index) =>
                            const Divider(height: 1, color: MokposColors.line),
                        itemBuilder: (context, index) {
                          final row = pendingPaymentRows[index];
                          final transaction = row.transaction;
                          final payment = row.payment;
                          return _DataRow(
                            key: ValueKey(
                              'pending_payment_${payment.paymentId}',
                            ),
                            cells: [
                              '${transaction.number}\n${transaction.status}',
                              transaction.order.customer?.name ??
                                  'Tanpa Pelanggan',
                              _formatDateTime(transaction.createdAt),
                              '${payment.methodName}\n${rupiah(payment.amount)}',
                              payment.reference ?? '-',
                              'Tandai sudah diterima',
                            ],
                            onTap: () => _showConfirmPaymentDialog(
                              context,
                              ref,
                              transaction,
                              payment,
                            ),
                          );
                        },
                      )
              : transactions.isEmpty
              ? const _EmptyState(
                  icon: LucideIcons.badgeDollarSign,
                  title: 'Belum Ada Penjualan',
                  subtitle:
                      'Transaksi yang berhasil dibayar akan muncul di sini.',
                )
              : ListView.separated(
                  itemCount: transactions.length,
                  separatorBuilder: (context, index) =>
                      const Divider(height: 1, color: MokposColors.line),
                  itemBuilder: (context, index) {
                    final transaction = transactions[index];
                    final order = transaction.order;
                    final payment = transaction.payments.isEmpty
                        ? '-'
                        : transaction.payments.first.methodName;
                    return _DataRow(
                      key: ValueKey('sales_transaction_${transaction.id}'),
                      cells: [
                        order.type.label,
                        order.customer?.name ?? 'Tanpa Pelanggan',
                        '${transaction.number}\n${transaction.status}',
                        _formatDateTime(transaction.createdAt),
                        transaction.cashier.name,
                        '${rupiah(transaction.total)}\n$payment',
                      ],
                      onTap: transaction.status == 'voided'
                          ? null
                          : () => _showTransactionActionsDialog(
                              context,
                              ref,
                              transaction,
                            ),
                    );
                  },
                ),
        ),
      ],
    );
  }
}

class _DialogInfoRow extends StatelessWidget {
  const _DialogInfoRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 96,
            child: Text(
              label,
              style: const TextStyle(
                color: MokposColors.muted,
                fontSize: 12,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(
                color: MokposColors.text,
                fontSize: 13,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

Future<void> _showConfirmPaymentDialog(
  BuildContext context,
  WidgetRef ref,
  SalesTransaction transaction,
  PaymentLine payment,
) {
  final paymentId = payment.paymentId;
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => NojposWarningDialog(
      title: 'Konfirmasi Pembayaran',
      subtitle:
          'Pastikan dana benar-benar sudah diterima sebelum menandai pembayaran.',
      confirmLabel: 'Tandai diterima',
      cancelLabel: 'Batal',
      confirmEnabled: paymentId != null,
      content: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          _DialogInfoRow(label: 'Transaksi', value: transaction.number),
          _DialogInfoRow(label: 'Metode', value: payment.methodName),
          _DialogInfoRow(label: 'Nominal', value: rupiah(payment.amount)),
          _DialogInfoRow(label: 'Reference', value: payment.reference ?? '-'),
        ],
      ),
      onConfirm: () async {
        await ref
            .read(nojposSessionProvider.notifier)
            .confirmPaymentLine(paymentId!);
        if (!context.mounted) return;
        Navigator.of(context).pop();
        final updated = ref
            .read(nojposSessionProvider)
            .transactions
            .where((item) => item.id == transaction.id)
            .firstOrNull;
        NojposToast.success(
          context,
          'Pembayaran dikonfirmasi',
          description: 'Status transaksi: ${updated?.status ?? payment.status}',
        );
      },
    ),
  );
}

Future<void> _showTransactionActionsDialog(
  BuildContext context,
  WidgetRef ref,
  SalesTransaction transaction,
) {
  final canManage = ref.read(nojposSessionProvider).canManageMasterData;
  final canRefund = canManage && _canRefundTransaction(transaction);
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (dialogContext) => AlertDialog(
      backgroundColor: Colors.white,
      title: Text('Aksi ${transaction.number}'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Status: ${transaction.status}'),
          Text('Total server: ${rupiah(transaction.total)}'),
          const SizedBox(height: 8),
          const Text(
            'Refund hanya untuk owner/admin dan tetap dihitung oleh server.',
            style: TextStyle(color: MokposColors.muted),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(dialogContext).pop(),
          child: const Text('Tutup'),
        ),
        OutlinedButton(
          key: const ValueKey('transaction_void_action'),
          onPressed: () {
            Navigator.of(dialogContext).pop();
            _showVoidDialog(context, ref, transaction);
          },
          child: const Text('Void'),
        ),
        FilledButton(
          key: const ValueKey('transaction_refund_action'),
          onPressed: canRefund
              ? () {
                  Navigator.of(dialogContext).pop();
                  _showRefundDialog(context, ref, transaction);
                }
              : null,
          child: Text(canManage ? 'Refund' : 'Refund owner/admin'),
        ),
      ],
    ),
  );
}

bool _canRefundTransaction(SalesTransaction transaction) {
  return transaction.status == 'paid' ||
      transaction.status == 'partially_refunded';
}

Future<void> _showRefundDialog(
  BuildContext context,
  WidgetRef ref,
  SalesTransaction transaction,
) {
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (dialogContext) => _RefundTransactionDialog(
      parentContext: context,
      ref: ref,
      transaction: transaction,
    ),
  );
}

class _RefundTransactionDialog extends StatefulWidget {
  const _RefundTransactionDialog({
    required this.parentContext,
    required this.ref,
    required this.transaction,
  });

  final BuildContext parentContext;
  final WidgetRef ref;
  final SalesTransaction transaction;

  @override
  State<_RefundTransactionDialog> createState() =>
      _RefundTransactionDialogState();
}

class _RefundTransactionDialogState extends State<_RefundTransactionDialog> {
  final reasonController = TextEditingController();
  final authController = TextEditingController();
  final nonRestockReasonController = TextEditingController();
  String method = 'cash';
  bool restock = true;

  @override
  void dispose() {
    reasonController.dispose();
    authController.dispose();
    nonRestockReasonController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final hasRefundableLines = widget.transaction.order.lines.any(
      (line) => line.transactionItemId != null,
    );
    return AlertDialog(
      backgroundColor: Colors.white,
      title: Text('Refund ${widget.transaction.number}'),
      content: SizedBox(
        width: 420,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Total server: ${rupiah(widget.transaction.total)}'),
              const SizedBox(height: 8),
              TextField(
                key: const ValueKey('refund_reason'),
                controller: reasonController,
                maxLines: 2,
                decoration: const InputDecoration(
                  labelText: 'Alasan refund',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 10),
              DropdownButtonFormField<String>(
                key: const ValueKey('refund_method'),
                initialValue: method,
                items: const [
                  DropdownMenuItem(value: 'cash', child: Text('Cash drawer')),
                  DropdownMenuItem(
                    value: 'original_method',
                    child: Text('Metode asal'),
                  ),
                ],
                onChanged: (value) => setState(() => method = value ?? 'cash'),
                decoration: const InputDecoration(
                  labelText: 'Metode refund',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 10),
              TextField(
                key: const ValueKey('refund_authorization_pin'),
                controller: authController,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'PIN otorisasi',
                  border: OutlineInputBorder(),
                ),
              ),
              SwitchListTile(
                key: const ValueKey('refund_restock_toggle'),
                contentPadding: EdgeInsets.zero,
                title: const Text('Restock item yang dikembalikan'),
                value: restock,
                onChanged: (value) => setState(() => restock = value),
              ),
              if (!restock)
                TextField(
                  key: const ValueKey('refund_non_restock_reason'),
                  controller: nonRestockReasonController,
                  decoration: const InputDecoration(
                    labelText: 'Alasan tidak restock',
                    border: OutlineInputBorder(),
                  ),
                ),
              if (!hasRefundableLines) ...[
                const SizedBox(height: 10),
                const Text(
                  'Transaksi ini belum memuat ID line server. Muat ulang daftar transaksi sebelum refund.',
                  style: TextStyle(color: MokposColors.danger),
                ),
              ],
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Batal'),
        ),
        FilledButton(
          key: const ValueKey('refund_submit'),
          onPressed: hasRefundableLines ? _submit : null,
          child: const Text('Konfirmasi Refund'),
        ),
      ],
    );
  }

  Future<void> _submit() async {
    final reason = reasonController.text.trim();
    final nonRestockReason = nonRestockReasonController.text.trim();
    if (reason.isEmpty) return;
    if (!restock && nonRestockReason.isEmpty) {
      NojposToast.warning(
        widget.parentContext,
        'Alasan tidak restock wajib diisi.',
      );
      return;
    }
    final lines = [
      for (final line in widget.transaction.order.lines)
        if (line.transactionItemId != null)
          RefundLineRequest(
            transactionItemId: line.transactionItemId!,
            quantity: line.quantity,
            restock: restock,
            nonRestockReason: restock ? null : nonRestockReason,
          ),
    ];
    final result = await widget.ref
        .read(nojposSessionProvider.notifier)
        .createRefund(
          transaction: widget.transaction,
          reason: reason,
          refundMethod: method,
          authorizationCode: authController.text.trim(),
          lines: lines,
        );
    authController.clear();
    if (!mounted) return;
    Navigator.of(context).pop();
    final message = result == null
        ? widget.ref.read(nojposSessionProvider).errorMessage ??
              'Refund ditolak'
        : 'Refund berhasil. Status: ${result.transactionStatus}';
    if (!widget.parentContext.mounted) return;
    if (result == null) {
      NojposToast.error(widget.parentContext, message);
    } else {
      NojposToast.success(
        widget.parentContext,
        'Refund berhasil',
        description: message,
      );
    }
  }
}

Future<void> _showVoidDialog(
  BuildContext context,
  WidgetRef ref,
  SalesTransaction transaction,
) {
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (dialogContext) => _VoidTransactionDialog(
      parentContext: context,
      ref: ref,
      transaction: transaction,
    ),
  );
}

class _VoidTransactionDialog extends StatefulWidget {
  const _VoidTransactionDialog({
    required this.parentContext,
    required this.ref,
    required this.transaction,
  });

  final BuildContext parentContext;
  final WidgetRef ref;
  final SalesTransaction transaction;

  @override
  State<_VoidTransactionDialog> createState() => _VoidTransactionDialogState();
}

class _VoidTransactionDialogState extends State<_VoidTransactionDialog> {
  final controller = TextEditingController();

  @override
  void dispose() {
    controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return NojposDangerDialog(
      title: 'Void ${widget.transaction.number}',
      subtitle:
          'Tindakan ini membatalkan transaksi penuh. Masukkan alasan yang jelas untuk audit.',
      confirmLabel: 'Void Penuh',
      cancelLabel: 'Batal',
      content: TextField(
        key: const ValueKey('void_reason'),
        controller: controller,
        autofocus: true,
        maxLines: 3,
        decoration: const InputDecoration(
          labelText: 'Reason',
          border: OutlineInputBorder(),
        ),
      ),
      onConfirm: () async {
        final reason = controller.text.trim();
        if (reason.isEmpty) return;
        final ok = await widget.ref
            .read(nojposSessionProvider.notifier)
            .voidTransaction(transaction: widget.transaction, reason: reason);
        if (!context.mounted) return;
        Navigator.of(context).pop();
        final error = widget.ref.read(nojposSessionProvider).errorMessage;
        if (!widget.parentContext.mounted) return;
        if (ok) {
          NojposToast.success(
            widget.parentContext,
            'Transaksi berhasil di-void',
          );
        } else {
          NojposToast.error(widget.parentContext, error ?? 'Void ditolak');
        }
      },
    );
  }
}

class _ReportsBody extends ConsumerStatefulWidget {
  const _ReportsBody();

  @override
  ConsumerState<_ReportsBody> createState() => _ReportsBodyState();
}

class _ReportsBodyState extends ConsumerState<_ReportsBody> {
  int _activeSection = 0;

  @override
  Widget build(BuildContext context) {
    final dashboard = ref.watch(reportsDashboardProvider);
    final session = ref.watch(nojposSessionProvider);
    final isCashier = !session.canManageMasterData;
    return Row(
      children: [
        _LeftList(
          title: 'Kategori Laporan',
          items: const [
            'Ringkasan Penjualan',
            'Produk Terjual',
            'Jenis Bayar',
            'Shift Kasir',
            'Void / Refund',
            'Top 10',
          ],
          activeIndex: _activeSection,
          onSelected: (index) => setState(() => _activeSection = index),
        ),
        Expanded(
          child: Padding(
            padding: const EdgeInsets.all(18),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                ReportsFilterChips(isCashier: isCashier),
                const SizedBox(height: 16),
                Expanded(
                  child: dashboard.when(
                    loading: () =>
                        const Center(child: CircularProgressIndicator()),
                    error: (error, _) => _ReportMessage(
                      title: 'Laporan gagal dimuat',
                      message: error.toString(),
                      onRetry: () => ref.invalidate(reportsDashboardProvider),
                    ),
                    data: (dashboard) => _ReportsSectionContent(
                      activeSection: _activeSection,
                      dashboard: dashboard,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}

class _ReportsSectionContent extends StatelessWidget {
  const _ReportsSectionContent({
    required this.activeSection,
    required this.dashboard,
  });

  final int activeSection;
  final ReportsDashboard dashboard;

  @override
  Widget build(BuildContext context) {
    return switch (activeSection) {
      0 => _ReportsDashboardView(dashboard: dashboard),
      1 => _SingleReportPanel(
        title: 'Produk Terjual',
        subtitle: 'Daftar produk yang terjual pada scope laporan aktif.',
        child: SoldProductsReportPanel(report: dashboard.soldProducts),
      ),
      2 => _SingleReportPanel(
        title: 'Jenis Bayar',
        subtitle: 'Ringkasan metode pembayaran dari transaksi terkonfirmasi.',
        child: PaymentMethodsReportPanel(report: dashboard.paymentMethods),
      ),
      3 => _SingleReportPanel(
        title: 'Shift Kasir',
        subtitle: 'Ringkasan shift kasir, kas tunai, dan selisih kas.',
        child: CashierShiftsReportPanel(report: dashboard.cashierShifts),
      ),
      4 => _SingleReportPanel(
        title: 'Void / Refund',
        subtitle: 'Audit void dan refund yang tercatat di server.',
        child: VoidRefundAuditReportPanel(report: dashboard.voidRefundAudit),
      ),
      5 => _SingleReportPanel(
        title: 'Top 10',
        subtitle: 'Peringkat produk terlaris dari data laporan server.',
        child: TopTenReportPanel(report: dashboard.topTen),
      ),
      _ => const SizedBox.shrink(),
    };
  }
}

class _SingleReportPanel extends StatelessWidget {
  const _SingleReportPanel({
    required this.title,
    required this.subtitle,
    required this.child,
  });

  final String title;
  final String subtitle;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      child: Align(
        alignment: Alignment.topLeft,
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 760),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(
                title,
                style: const TextStyle(
                  color: MokposColors.text,
                  fontWeight: FontWeight.w900,
                  fontSize: 22,
                ),
              ),
              const SizedBox(height: 6),
              Text(
                subtitle,
                style: const TextStyle(
                  color: MokposColors.muted,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 16),
              child,
            ],
          ),
        ),
      ),
    );
  }
}

class _ReportsDashboardView extends StatelessWidget {
  const _ReportsDashboardView({required this.dashboard});

  final ReportsDashboard dashboard;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        SizedBox(
          height: 250,
          child: _ReportSummary(summary: dashboard.salesSummary),
        ),
        const SizedBox(height: 12),
        Expanded(
          child: GridView(
            gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
              maxCrossAxisExtent: 420,
              mainAxisExtent: 240,
              crossAxisSpacing: 12,
              mainAxisSpacing: 12,
            ),
            children: [
              SoldProductsReportPanel(report: dashboard.soldProducts),
              PaymentMethodsReportPanel(report: dashboard.paymentMethods),
              CashierShiftsReportPanel(report: dashboard.cashierShifts),
              VoidRefundAuditReportPanel(report: dashboard.voidRefundAudit),
              TopTenReportPanel(report: dashboard.topTen),
            ],
          ),
        ),
      ],
    );
  }
}

class _ReportSummary extends StatelessWidget {
  const _ReportSummary({required this.summary});

  final SalesSummary summary;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          width: 450,
          child: GridView.count(
            crossAxisCount: 2,
            mainAxisSpacing: 12,
            crossAxisSpacing: 12,
            childAspectRatio: 2.45,
            children: [
              _MetricCard(
                label: 'Total Penjualan',
                value: rupiah(summary.totalSales),
                color: MokposColors.primary,
              ),
              _MetricCard(
                label: 'Total Transaksi',
                value: '${summary.transactionCount}',
                color: const Color(0xFFFF4FC3),
              ),
              _MetricCard(
                label: 'Total Diskon',
                value: rupiah(summary.totalDiscount),
                color: const Color(0xFF7C4DFF),
              ),
              _MetricCard(
                label: 'Total Void',
                value: rupiah(summary.totalVoid),
                color: MokposColors.danger,
              ),
              _MetricCard(
                label: 'Rata-rata Transaksi',
                value: rupiah(summary.averageTransactionValue),
                color: MokposColors.accent,
              ),
              for (final total in summary.paymentTotals.take(3))
                _MetricCard(
                  label: total.method,
                  value: rupiah(total.amount),
                  color: const Color(0xFF4FC3F7),
                ),
            ],
          ),
        ),
        const SizedBox(width: 50),
        Expanded(child: _SimpleChart(points: summary.chartPoints)),
      ],
    );
  }
}

class _ReportMessage extends StatelessWidget {
  const _ReportMessage({
    required this.title,
    required this.message,
    required this.onRetry,
  });

  final String title;
  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            title,
            style: const TextStyle(
              fontWeight: FontWeight.w900,
              fontSize: 18,
              color: MokposColors.text,
            ),
          ),
          const SizedBox(height: 8),
          Text(message, style: const TextStyle(color: MokposColors.muted)),
          const SizedBox(height: 16),
          OutlinedButton(onPressed: onRetry, child: const Text('Muat ulang')),
        ],
      ),
    );
  }
}

class _InventoryBody extends ConsumerStatefulWidget {
  const _InventoryBody();

  @override
  ConsumerState<_InventoryBody> createState() => _InventoryBodyState();
}

class _InventoryBodyState extends ConsumerState<_InventoryBody> {
  int _activeSection = 0;

  @override
  void initState() {
    super.initState();
    Future<void>.microtask(_reloadInventoryFeature);
  }

  Future<void> _reloadInventoryFeature() async {
    await ref.read(nojposSessionProvider.notifier).loadInventory();
    final outletId = ref.read(nojposSessionProvider).outlet.id;
    if (outletId.isEmpty) return;
    await ref
        .read(inventoryOperationsProvider.notifier)
        .loadMovements(outletId: outletId);
    await ref
        .read(inventoryOperationsProvider.notifier)
        .loadInTransit(outletId: outletId);
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final operations = ref.watch(inventoryOperationsProvider);
    final inventory = session.inventory;
    final movements =
        operations.movements?.items ?? inventory?.movements ?? const [];
    final inTransit = operations.inTransit?.items ?? const [];
    return Row(
      children: [
        _LeftList(
          title: 'Kategori Inventori',
          items: const [
            'Faktur Pembelian',
            'Stok Opname',
            'Transfer In-Transit',
            'Stok Terbuang',
          ],
          activeIndex: _activeSection,
          onSelected: (index) => setState(() => _activeSection = index),
        ),
        Expanded(
          child: Column(
            children: [
              const _FilterRow(),
              const _TableHeader(
                columns: ['Produk', 'Stok', 'Tersedia', 'Status'],
              ),
              SizedBox(
                height: 260,
                child: session.isBusy && inventory == null
                    ? const Center(child: CircularProgressIndicator())
                    : inventory == null || inventory.items.isEmpty
                    ? const _EmptyState(
                        icon: LucideIcons.folderSearch,
                        title: 'Stok Tidak Tersedia',
                        subtitle:
                            'Inventory akan muncul setelah data API dimuat',
                      )
                    : ListView.separated(
                        itemCount: inventory.items.length,
                        separatorBuilder: (context, index) =>
                            const Divider(height: 1, color: MokposColors.line),
                        itemBuilder: (context, index) {
                          final item = inventory.items[index];
                          final transit = item.inTransitOutQuantity > 0
                              ? 'Transit keluar ${item.inTransitOutQuantity}'
                              : item.inTransitInQuantity > 0
                              ? 'Transit masuk ${item.inTransitInQuantity}'
                              : item.isNegative
                              ? 'Stok negatif diizinkan'
                              : 'Normal';
                          return _DataRow(
                            cells: [
                              item.name,
                              '${item.stockOnHand}',
                              '${item.availableQuantity}',
                              transit,
                            ],
                          );
                        },
                      ),
              ),
              Expanded(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  child: _InventorySectionContent(
                    activeSection: _activeSection,
                    canManage: session.canManageMasterData,
                    outletId: session.outlet.id,
                    outlets: session.outlets,
                    items: inventory?.items ?? const [],
                    movements: movements,
                    inTransit: inTransit,
                    onChanged: _reloadInventoryFeature,
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _InventorySectionContent extends StatelessWidget {
  const _InventorySectionContent({
    required this.activeSection,
    required this.canManage,
    required this.outletId,
    required this.outlets,
    required this.items,
    required this.movements,
    required this.inTransit,
    required this.onChanged,
  });

  final int activeSection;
  final bool canManage;
  final String outletId;
  final List<Outlet> outlets;
  final List<InventoryStockItem> items;
  final List<StockMovement> movements;
  final List<InventoryTransferResult> inTransit;
  final Future<void> Function() onChanged;

  @override
  Widget build(BuildContext context) {
    return switch (activeSection) {
      0 => Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InventoryActionPanel(
            canManage: canManage,
            outletId: outletId,
            outlets: outlets,
            items: items,
            onChanged: onChanged,
          ),
          const SizedBox(height: 12),
          InventoryMovementPanel(movements: movements),
        ],
      ),
      1 => Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          InventoryActionPanel(
            canManage: canManage,
            outletId: outletId,
            outlets: outlets,
            items: items,
            onChanged: onChanged,
          ),
          const SizedBox(height: 12),
          InventoryMovementPanel(
            movements: movements
                .where((movement) => movement.type.contains('opname'))
                .toList(),
          ),
        ],
      ),
      2 => InTransitTransfersPanel(
        canManage: canManage,
        transfers: inTransit,
        onChanged: onChanged,
      ),
      3 => InventoryMovementPanel(
        movements: movements
            .where(
              (movement) =>
                  movement.type.contains('waste') ||
                  movement.type.contains('damage') ||
                  movement.type.contains('discard'),
            )
            .toList(),
      ),
      _ => const SizedBox.shrink(),
    };
  }
}

class _SettingsBody extends ConsumerStatefulWidget {
  const _SettingsBody();

  @override
  ConsumerState<_SettingsBody> createState() => _SettingsBodyState();
}

class _SettingsBodyState extends ConsumerState<_SettingsBody> {
  int _activeSection = 0;

  @override
  Widget build(BuildContext context) {
    final catalog = ref.watch(posCatalogProvider);
    final session = ref.watch(nojposSessionProvider);
    final canManage = session.canManageMasterData;
    return Row(
      children: [
        _SettingsMenu(
          activeIndex: _activeSection,
          onSelected: (index) => setState(() => _activeSection = index),
        ),
        Expanded(
          child: _SettingsSectionContent(
            activeSection: _activeSection,
            catalog: catalog,
            canManage: canManage,
            onSearchProducts: (value) =>
                ref.read(posCatalogProvider.notifier).load(search: value),
          ),
        ),
      ],
    );
  }
}

class _SettingsSectionContent extends StatelessWidget {
  const _SettingsSectionContent({
    required this.activeSection,
    required this.catalog,
    required this.canManage,
    required this.onSearchProducts,
  });

  final int activeSection;
  final PosCatalogState catalog;
  final bool canManage;
  final ValueChanged<String> onSearchProducts;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        _SettingsHeader(activeSection: activeSection),
        Expanded(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: switch (activeSection) {
              0 => _ProductSettingsSection(
                catalog: catalog,
                canManage: canManage,
                onSearchProducts: onSearchProducts,
              ),
              1 => const _ReceiptAndFeeSettingsSection(),
              2 => const TaxServiceSettingsPanel(),
              3 => const StaffSettingsPanel(),
              4 => const _DeviceSettingsSection(),
              5 => const PaymentMethodSettingsPanel(),
              6 => const RefundSettingsPanel(),
              7 => const _LogoutSettingsSection(),
              _ => _RoadmapStatePanel(
                title: _settingsItems[activeSection],
                subtitle:
                    'Pengaturan ini nonaktif di terminal kasir. Gunakan panel owner/admin yang berwenang untuk mengelola konfigurasi tersebut.',
              ),
            },
          ),
        ),
      ],
    );
  }
}

class _SettingsHeader extends StatelessWidget {
  const _SettingsHeader({required this.activeSection});

  final int activeSection;

  @override
  Widget build(BuildContext context) {
    final title = _settingsItems[activeSection];
    return Container(
      height: 56,
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: Row(
        children: [
          const SizedBox(width: 24),
          Text(
            title,
            style: const TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
              fontSize: 16,
            ),
          ),
          const Spacer(),
          if (activeSection == 0) ...[
            const _TabText(text: 'Produk', active: true),
            const SizedBox(width: 28),
            const _TabText(text: 'Kategori'),
            const SizedBox(width: 24),
          ],
        ],
      ),
    );
  }
}

class _ProductSettingsSection extends StatelessWidget {
  const _ProductSettingsSection({
    required this.catalog,
    required this.canManage,
    required this.onSearchProducts,
  });

  final PosCatalogState catalog;
  final bool canManage;
  final ValueChanged<String> onSearchProducts;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        const _BackendSettingsSummary(),
        const SizedBox(height: 24),
        _SearchLine(onChanged: onSearchProducts),
        const SizedBox(height: 16),
        if (canManage) const _AddProductButton(),
        if (!canManage) const _ReadOnlyNotice(),
        const SizedBox(height: 14),
        _CategoryStrip(categories: catalog.categoryItems, canManage: canManage),
        const SizedBox(height: 14),
        SizedBox(
          height: 360,
          child: ListView.separated(
            itemCount: catalog.products.length,
            separatorBuilder: (context, index) =>
                const Divider(height: 1, color: MokposColors.line),
            itemBuilder: (context, index) => _ProductSettingRow(
              product: catalog.products[index],
              canManage: canManage,
            ),
          ),
        ),
      ],
    );
  }
}

class _ReceiptAndFeeSettingsSection extends StatelessWidget {
  const _ReceiptAndFeeSettingsSection();

  @override
  Widget build(BuildContext context) {
    return const Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _BackendSettingsSummary(),
        SizedBox(height: 24),
        _PrinterSettingsPanel(),
      ],
    );
  }
}

class _DeviceSettingsSection extends StatelessWidget {
  const _DeviceSettingsSection();

  @override
  Widget build(BuildContext context) {
    return const Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _TerminalRegisteredPanel(),
        SizedBox(height: 24),
        SecuritySettingsPanel(),
        SizedBox(height: 24),
        StoreStatePanel(),
      ],
    );
  }
}

class _TerminalRegisteredPanel extends ConsumerStatefulWidget {
  const _TerminalRegisteredPanel();

  @override
  ConsumerState<_TerminalRegisteredPanel> createState() =>
      _TerminalRegisteredPanelState();
}

class _TerminalRegisteredPanelState
    extends ConsumerState<_TerminalRegisteredPanel> {
  @override
  void initState() {
    super.initState();
    Future<void>.microtask(
      () => ref.read(terminalLockControllerProvider.notifier).load(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final lockState = ref.watch(terminalLockControllerProvider);
    final deviceId = session.deviceId;
    final deviceUuid = session.deviceUuid;
    final hasRegisteredTerminal = deviceId != null && deviceId.isNotEmpty;
    final activeShift = session.activeShift;
    final backendLock = lockState.lockState;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Row(
            children: [
              Icon(LucideIcons.tabletSmartphone, color: MokposColors.primary),
              SizedBox(width: 10),
              Expanded(
                child: Text(
                  'Terminal terdaftar',
                  style: TextStyle(
                    color: MokposColors.text,
                    fontWeight: FontWeight.w900,
                    fontSize: 16,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            hasRegisteredTerminal
                ? 'Informasi terminal diverifikasi server. Panel ini read-only di aplikasi kasir agar enrollment perangkat tetap dikendalikan owner/admin.'
                : 'Terminal belum terdaftar untuk session ini. Login ulang atau minta owner/admin mengecek enrollment perangkat.',
            style: const TextStyle(
              color: MokposColors.muted,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 14),
          _PrinterStatusLine(
            label: 'Status enrollment',
            value: hasRegisteredTerminal
                ? 'Terdaftar di backend'
                : 'Belum ada device_id backend',
          ),
          _PrinterStatusLine(
            label: 'Device ID',
            value: hasRegisteredTerminal ? deviceId : 'Tidak tersedia',
          ),
          _PrinterStatusLine(
            label: 'Device UUID lokal',
            value: (deviceUuid == null || deviceUuid.isEmpty)
                ? 'Tidak tersedia'
                : deviceUuid,
          ),
          _PrinterStatusLine(
            label: 'Outlet session',
            value: session.outlet.id.isEmpty
                ? 'Belum memilih outlet'
                : session.outlet.name,
          ),
          _PrinterStatusLine(
            label: 'Kasir aktif',
            value: session.cashier.id.isEmpty
                ? 'Belum PIN kasir'
                : '${session.cashier.name} (${session.cashier.role})',
          ),
          _PrinterStatusLine(
            label: 'Shift',
            value: activeShift == null
                ? 'Belum ada shift aktif'
                : '${activeShift.status} — ${activeShift.id}',
          ),
          const Divider(height: 24, color: MokposColors.line),
          if (lockState.isBusy && backendLock == null)
            const Row(
              children: [
                SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
                SizedBox(width: 10),
                Text(
                  'Memuat status lock terminal...',
                  style: TextStyle(
                    color: MokposColors.muted,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            )
          else ...[
            _PrinterStatusLine(
              label: 'Lock terminal',
              value: backendLock == null
                  ? 'Status lock belum dimuat'
                  : backendLock.locked
                  ? 'Terkunci${backendLock.lockReason == null ? '' : ' (${backendLock.lockReason})'}'
                  : 'Tidak terkunci',
            ),
            if (backendLock?.cashier != null)
              _PrinterStatusLine(
                label: 'Lock oleh',
                value:
                    '${backendLock!.cashier!.name} (${backendLock.cashier!.role})',
              ),
            if (backendLock?.unlockedBy != null)
              _PrinterStatusLine(
                label: 'Terakhir dibuka',
                value:
                    '${backendLock!.unlockedBy!.name} (${backendLock.unlockedBy!.role})',
              ),
            _PrinterStatusLine(
              label: 'Idle/session policy',
              value: backendLock == null
                  ? 'Mengikuti pengaturan keamanan saat tersedia'
                  : '${backendLock.idleTimeoutSeconds}s / ${backendLock.sessionTimeoutSeconds}s',
            ),
          ],
          if (lockState.errorMessage != null) ...[
            const SizedBox(height: 12),
            Text(
              'Status lock gagal dimuat: ${lockState.errorMessage}',
              style: const TextStyle(
                color: MokposColors.danger,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _LogoutSettingsSection extends ConsumerWidget {
  const _LogoutSettingsSection();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final canLogoutApp = ref.watch(
      nojposSessionProvider.select((session) => session.canManageMasterData),
    );
    return _RoadmapStatePanel(
      title: 'Keluar',
      subtitle: canLogoutApp
          ? 'Keluar akun aplikasi dari perangkat ini.'
          : 'Keluar akun hanya bisa dilakukan owner/admin.',
      action: FilledButton.icon(
        onPressed: canLogoutApp ? () => _confirmAppLogout(context, ref) : null,
        icon: const Icon(LucideIcons.logOut, size: 16),
        label: const Text('Keluar Akun'),
      ),
    );
  }
}

class _RoadmapStatePanel extends StatelessWidget {
  const _RoadmapStatePanel({
    required this.title,
    required this.subtitle,
    this.action,
  });

  final String title;
  final String subtitle;
  final Widget? action;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: const TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
              fontSize: 18,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            subtitle,
            style: const TextStyle(
              color: MokposColors.muted,
              fontWeight: FontWeight.w700,
            ),
          ),
          if (action != null) ...[const SizedBox(height: 16), action!],
        ],
      ),
    );
  }
}

class _BackendSettingsSummary extends ConsumerWidget {
  const _BackendSettingsSummary();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(settingsAggregateProvider);

    return settings.when(
      data: (value) => SettingsSummaryPanel(settings: value),
      loading: () => Container(
        width: double.infinity,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          border: Border.all(color: MokposColors.line),
          borderRadius: BorderRadius.circular(8),
        ),
        child: const Row(
          children: [
            SizedBox(
              width: 18,
              height: 18,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
            SizedBox(width: 10),
            Text(
              'Memuat pengaturan server...',
              style: TextStyle(
                color: MokposColors.muted,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
      ),
      error: (error, stackTrace) => Container(
        width: double.infinity,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          border: Border.all(color: MokposColors.line),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Text(
          'Gagal memuat pengaturan server: ${_messageForUi(error)}',
          style: const TextStyle(
            color: MokposColors.danger,
            fontWeight: FontWeight.w800,
          ),
        ),
      ),
    );
  }
}

class _PrinterSettingsPanel extends ConsumerStatefulWidget {
  const _PrinterSettingsPanel();

  @override
  ConsumerState<_PrinterSettingsPanel> createState() =>
      _PrinterSettingsPanelState();
}

class _PrinterSettingsPanelState extends ConsumerState<_PrinterSettingsPanel> {
  PrinterSettings settings = const PrinterSettings();
  List<BluetoothPrinterDevice> devices = const [];
  bool loading = true;
  String? statusMessage;

  @override
  void initState() {
    super.initState();
    Future<void>.microtask(_loadSettings);
  }

  Future<void> _loadSettings() async {
    final loaded = await ref
        .read(receiptPrintingServiceProvider)
        .loadSettings();
    if (!mounted) return;
    setState(() {
      settings = loaded;
      loading = false;
    });
  }

  Future<void> _scan() async {
    setState(() {
      loading = true;
      statusMessage =
          'Mencari printer Bluetooth yang sudah paired di perangkat ini...';
    });
    try {
      final result = await ref
          .read(receiptPrintingServiceProvider)
          .scanPrinters();
      if (!mounted) return;
      setState(() {
        devices = result;
        statusMessage = result.isEmpty
            ? 'Belum ada printer paired di perangkat ini. Pairing dari Settings Android dulu, lalu scan ulang.'
            : 'Ditemukan ${result.length} perangkat paired. Pilih salah satu untuk disimpan sebagai printer lokal perangkat ini.';
      });
    } catch (error) {
      if (!mounted) return;
      setState(
        () => statusMessage = 'Scan printer gagal: ${_messageForUi(error)}',
      );
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> _save(BluetoothPrinterDevice device) async {
    setState(() => loading = true);
    final result = await ref
        .read(receiptPrintingServiceProvider)
        .saveDefaultPrinter(device);
    final loaded = await ref
        .read(receiptPrintingServiceProvider)
        .loadSettings();
    if (!mounted) return;
    setState(() {
      settings = loaded;
      statusMessage = result.message;
      loading = false;
    });
  }

  Future<void> _toggleCashDrawer(bool enabled) async {
    await ref
        .read(receiptPrintingServiceProvider)
        .setCashDrawerEnabled(enabled);
    await _loadSettings();
    if (!mounted) return;
    setState(() {
      statusMessage = enabled
          ? 'Cash drawer aktif. Perintah buka laci akan dikirim saat cetak transaksi tunai.'
          : 'Cash drawer nonaktif.';
    });
  }

  Future<void> _testPrint() async {
    setState(() => loading = true);
    final result = await ref.read(receiptPrintingServiceProvider).testPrint();
    if (!mounted) return;
    setState(() {
      statusMessage = result.message;
      loading = false;
    });
  }

  Future<void> _kickDrawer() async {
    setState(() => loading = true);
    final result = await ref
        .read(receiptPrintingServiceProvider)
        .kickCashDrawer();
    if (!mounted) return;
    setState(() {
      statusMessage = result.message;
      loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    final printer = settings.defaultPrinter;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              const Icon(LucideIcons.printer, color: MokposColors.primary),
              const SizedBox(width: 10),
              const Expanded(
                child: Text(
                  'Printer lokal perangkat ini',
                  style: TextStyle(
                    color: MokposColors.text,
                    fontWeight: FontWeight.w900,
                    fontSize: 16,
                  ),
                ),
              ),
              OutlinedButton.icon(
                onPressed: loading ? null : _scan,
                icon: const Icon(LucideIcons.bluetooth, size: 16),
                label: const Text('Scan Bluetooth'),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            'Pengaturan ini hanya menyimpan pilihan printer Bluetooth di perangkat kasir ini. Ini terpisah dari terminal terdaftar di backend dan tidak berarti printer sedang connected.',
            style: const TextStyle(
              color: MokposColors.muted,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 12),
          _PrinterStatusLine(
            label: 'Printer lokal',
            value: printer == null
                ? 'Belum ada printer disimpan'
                : '${printer.name} (${printer.address})',
          ),
          _PrinterStatusLine(
            label: 'Status koneksi',
            value: printer == null
                ? 'Tidak ada printer lokal. Cetak tetap best-effort dan transaksi tidak diblokir.'
                : 'Belum diverifikasi connected. Aplikasi akan mencoba connect ulang saat test/cetak.',
          ),
          if (!loading && printer == null && devices.isEmpty) ...[
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: MokposColors.primarySoft,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: MokposColors.line),
              ),
              child: const Text(
                'Belum ada device paired yang ditampilkan. Gunakan Scan Bluetooth setelah printer dipairing dari Settings Android.',
                style: TextStyle(
                  color: MokposColors.text,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          ],
          const SizedBox(height: 12),
          if (devices.isNotEmpty) ...[
            const Text(
              'Printer paired ditemukan',
              style: TextStyle(
                color: MokposColors.muted,
                fontWeight: FontWeight.w900,
                fontSize: 12,
              ),
            ),
            const SizedBox(height: 6),
            for (final device in devices)
              ListTile(
                dense: true,
                contentPadding: EdgeInsets.zero,
                title: Text(device.name),
                subtitle: Text(device.address),
                trailing: FilledButton(
                  onPressed: loading ? null : () => _save(device),
                  child: const Text('Simpan Default'),
                ),
              ),
          ],
          const Divider(height: 24, color: MokposColors.line),
          SwitchListTile(
            value: settings.cashDrawerEnabled,
            onChanged: loading ? null : _toggleCashDrawer,
            contentPadding: EdgeInsets.zero,
            title: const Text(
              'Cash drawer kick',
              style: TextStyle(fontWeight: FontWeight.w900),
            ),
            subtitle: const Text(
              'Best-effort dan tergantung dukungan printer. Penjualan tidak diblokir.',
            ),
          ),
          Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: loading ? null : _testPrint,
                  icon: const Icon(LucideIcons.receiptText, size: 16),
                  label: const Text('Test Print'),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: loading ? null : _kickDrawer,
                  icon: const Icon(LucideIcons.archive, size: 16),
                  label: const Text('Test Buka Laci'),
                ),
              ),
            ],
          ),
          if (statusMessage != null) ...[
            const SizedBox(height: 12),
            Text(
              statusMessage!,
              style: const TextStyle(
                color: MokposColors.muted,
                fontWeight: FontWeight.w700,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _PrinterStatusLine extends StatelessWidget {
  const _PrinterStatusLine({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 132,
            child: Text(
              label,
              style: const TextStyle(
                color: MokposColors.muted,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _AttendanceBody extends ConsumerStatefulWidget {
  const _AttendanceBody();

  @override
  ConsumerState<_AttendanceBody> createState() => _AttendanceBodyState();
}

class _AttendanceBodyState extends ConsumerState<_AttendanceBody> {
  @override
  void initState() {
    super.initState();
    Future<void>.microtask(
      () => ref.read(attendanceControllerProvider.notifier).load(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final now = DateTime.now();
    final attendanceState = ref.watch(attendanceControllerProvider);
    final attendance = attendanceState.records;
    final latest = attendance.isEmpty ? null : attendance.first;
    return Row(
      children: [
        Expanded(
          child: ColoredBox(
            color: const Color(0xFF3D4543),
            child: Padding(
              padding: const EdgeInsets.all(18),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '${now.hour.toString().padLeft(2, '0')}:${now.minute.toString().padLeft(2, '0')}',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 34,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    _formatDateTime(now),
                    style: const TextStyle(color: Colors.white70),
                  ),
                  const Spacer(),
                  Center(
                    child: attendanceState.isLoading && attendance.isEmpty
                        ? const CircularProgressIndicator(color: Colors.white)
                        : Text(
                            attendanceState.errorMessage ??
                                (latest == null
                                    ? 'Belum ada aktivitas absensi hari ini'
                                    : latest.isOpen
                                    ? '${latest.employee.name} sedang clock in'
                                    : '${latest.employee.name} sudah clock out'),
                            textAlign: TextAlign.center,
                            style: TextStyle(
                              color: attendanceState.errorMessage == null
                                  ? Colors.white70
                                  : const Color(0xFFFFD7D7),
                            ),
                          ),
                  ),
                  const Spacer(),
                ],
              ),
            ),
          ),
        ),
        const Expanded(child: _AttendancePinPanel()),
      ],
    );
  }
}

String _formatDateTime(DateTime value) {
  const months = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'Mei',
    'Jun',
    'Jul',
    'Agu',
    'Sep',
    'Okt',
    'Nov',
    'Des',
  ];
  final day = value.day.toString().padLeft(2, '0');
  final month = months[value.month - 1];
  final hour = value.hour.toString().padLeft(2, '0');
  final minute = value.minute.toString().padLeft(2, '0');
  return '$day $month ${value.year}, $hour:$minute';
}

String _statusLabel(OrderStatus status) {
  return switch (status) {
    OrderStatus.active => 'Aktif',
    OrderStatus.saved => 'Disimpan',
    OrderStatus.paid => 'Lunas',
    OrderStatus.canceled => 'Batal',
  };
}

List<CartItem> _cartItemsFromOrder(SalesOrder order, List<Product> products) {
  return [
    for (final line in order.lines)
      CartItem(
        product: products.firstWhere(
          (product) => product.id == line.productId,
          orElse: () => Product(
            id: line.productId,
            name: line.name,
            category: 'Lainnya',
            price: line.unitPrice,
            imageUrl: '',
          ),
        ),
        quantity: line.quantity,
        discount: line.discount,
      ),
  ];
}

class _LeftList extends StatelessWidget {
  const _LeftList({
    required this.title,
    required this.items,
    required this.activeIndex,
    this.onSelected,
  });

  final String title;
  final List<String> items;
  final int activeIndex;
  final ValueChanged<int>? onSelected;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 190,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(right: BorderSide(color: MokposColors.line)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Container(
            height: 52,
            padding: const EdgeInsets.symmetric(horizontal: 10),
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: MokposColors.line)),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    title,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: MokposColors.muted,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ),
                const Icon(LucideIcons.listFilter, size: 16),
              ],
            ),
          ),
          for (final (index, item) in items.indexed)
            InkWell(
              onTap: () => onSelected?.call(index),
              child: Container(
                height: 48,
                padding: const EdgeInsets.symmetric(horizontal: 10),
                alignment: Alignment.centerLeft,
                decoration: BoxDecoration(
                  color: index == activeIndex
                      ? MokposColors.primarySoft
                      : Colors.white,
                  border: Border(
                    left: BorderSide(
                      color: index == activeIndex
                          ? MokposColors.primary
                          : Colors.transparent,
                      width: 3,
                    ),
                    bottom: const BorderSide(color: MokposColors.line),
                  ),
                ),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        item,
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(
                          color: index == activeIndex
                              ? MokposColors.text
                              : MokposColors.muted,
                          fontSize: 12,
                          fontWeight: index == activeIndex
                              ? FontWeight.w900
                              : FontWeight.w700,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _TableHeader extends StatelessWidget {
  const _TableHeader({required this.columns});

  final List<String> columns;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 48,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: Row(
        children: [
          for (final column in columns)
            Expanded(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                child: Text(
                  column,
                  style: const TextStyle(
                    color: MokposColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _DataRow extends StatelessWidget {
  const _DataRow({required this.cells, this.onTap, super.key});

  final List<String> cells;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: SizedBox(
        height: 76,
        child: Row(
          children: [
            for (final cell in cells)
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  child: Text(
                    cell,
                    style: const TextStyle(
                      color: MokposColors.text,
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
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

class _LoadErrorState extends StatelessWidget {
  const _LoadErrorState({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(
              LucideIcons.circleAlert,
              size: 44,
              color: MokposColors.danger,
            ),
            const SizedBox(height: 12),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 16),
            FilledButton.icon(
              onPressed: onRetry,
              icon: const Icon(LucideIcons.refreshCw, size: 18),
              label: const Text('Muat ulang'),
            ),
          ],
        ),
      ),
    );
  }
}

class _EmptyState extends StatelessWidget {
  const _EmptyState({
    required this.icon,
    required this.title,
    required this.subtitle,
  });

  final IconData icon;
  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 92, color: MokposColors.primary),
          const SizedBox(height: 18),
          Text(
            title,
            style: const TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
              fontSize: 16,
            ),
          ),
          const SizedBox(height: 8),
          SizedBox(
            width: 430,
            child: Text(
              subtitle,
              textAlign: TextAlign.center,
              style: const TextStyle(color: MokposColors.muted, height: 1.4),
            ),
          ),
        ],
      ),
    );
  }
}

class _MetricCard extends StatelessWidget {
  const _MetricCard({
    required this.label,
    required this.value,
    required this.color,
  });

  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 86,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: MokposColors.line),
        boxShadow: const [
          BoxShadow(
            color: Color(0x0F000000),
            blurRadius: 12,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: const TextStyle(color: MokposColors.muted, fontSize: 11),
          ),
          const Spacer(),
          Text(
            value,
            style: const TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
              fontSize: 16,
            ),
          ),
          const SizedBox(height: 6),
          Container(height: 3, color: color),
        ],
      ),
    );
  }
}

class _FilterRow extends StatelessWidget {
  const _FilterRow();

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(14),
      child: Row(
        children: [
          _FilterBox(text: 'Hari Ini'),
          const SizedBox(width: 14),
          _FilterBox(text: '16 Jun 2026 - 16 Jun 2026', width: 260),
          const SizedBox(width: 14),
          _FilterBox(text: 'Semua Status', width: 180),
        ],
      ),
    );
  }
}

class _FilterBox extends StatelessWidget {
  const _FilterBox({required this.text, this.width = 160});

  final String text;
  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(6),
      ),
      child: Row(
        children: [
          Expanded(
            child: Text(
              text,
              style: const TextStyle(color: MokposColors.muted, fontSize: 12),
            ),
          ),
          const Icon(
            LucideIcons.chevronDown,
            size: 16,
            color: MokposColors.muted,
          ),
        ],
      ),
    );
  }
}

class _SimpleChart extends StatelessWidget {
  const _SimpleChart({required this.points});

  final List<SalesChartPoint> points;

  @override
  Widget build(BuildContext context) {
    if (points.isEmpty) {
      return SizedBox(
        height: 280,
        child: NojposStateView.empty(
          title: 'Belum ada data chart',
          subtitle:
              'Laporan akan muncul setelah ada transaksi pada periode ini.',
          illustrationAsset: NojposAssets.emptyReports,
          compact: true,
        ),
      );
    }
    final maxAmount = points.fold<int>(
      0,
      (max, point) => point.amount > max ? point.amount : max,
    );
    return SizedBox(
      height: 280,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          for (final point in points)
            Padding(
              padding: const EdgeInsets.only(right: 12),
              child: Container(
                width: 14,
                height: maxAmount == 0 ? 4 : 220 * point.amount / maxAmount,
                color: MokposColors.primary,
              ),
            ),
          Expanded(
            child: Align(
              alignment: Alignment.bottomLeft,
              child: Text(
                points.map((point) => point.label).take(4).join('      '),
                style: const TextStyle(color: MokposColors.muted, fontSize: 11),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

const _settingsItems = [
  'Produk & Kategori',
  'Struk & Biaya',
  'Pajak',
  'Kasir',
  'Perangkat',
  'Pembayaran Nontunai',
  'Refund',
  'Keluar',
];

class _SettingsMenu extends ConsumerWidget {
  const _SettingsMenu({required this.activeIndex, required this.onSelected});

  final int activeIndex;
  final ValueChanged<int> onSelected;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(nojposSessionProvider);
    final canLogoutApp = session.canManageMasterData;
    return Container(
      width: 330,
      decoration: const BoxDecoration(
        border: Border(right: BorderSide(color: MokposColors.line)),
      ),
      child: ListView(
        children: [
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 18, 16, 8),
            child: Text(
              'OPERASIONAL',
              style: TextStyle(
                color: MokposColors.muted,
                fontWeight: FontWeight.w900,
                fontSize: 12,
              ),
            ),
          ),
          for (final (index, item) in _settingsItems.indexed)
            ListTile(
              dense: true,
              selected: index == activeIndex,
              enabled: item != 'Keluar' || canLogoutApp,
              onTap: () {
                if (item == 'Keluar' && !canLogoutApp) {
                  NojposToast.info(
                    context,
                    'Keluar akun hanya untuk owner/admin',
                    description:
                        'Minta owner atau admin untuk mengeluarkan akun aplikasi dari perangkat ini.',
                  );
                  return;
                }
                onSelected(index);
              },
              leading: Icon(
                index == 0
                    ? LucideIcons.boxes
                    : item == 'Keluar'
                    ? LucideIcons.logOut
                    : LucideIcons.settings,
                size: 18,
                color: index == activeIndex
                    ? MokposColors.primary
                    : MokposColors.muted,
              ),
              title: Text(
                item,
                style: TextStyle(
                  color: index == activeIndex
                      ? MokposColors.primary
                      : MokposColors.text,
                  fontWeight: index == activeIndex
                      ? FontWeight.w900
                      : FontWeight.w800,
                  fontSize: 13,
                ),
              ),
              subtitle: item == 'Keluar' && !canLogoutApp
                  ? const Text('Hanya owner/admin')
                  : null,
            ),
        ],
      ),
    );
  }
}

Future<void> _confirmAppLogout(BuildContext context, WidgetRef ref) async {
  final confirmed =
      await showNojposDangerDialog(
        context,
        title: 'Keluar Akun?',
        subtitle:
            'Akun aplikasi POS akan dikeluarkan dari perangkat ini. Setelah keluar, Anda perlu login ulang memakai email dan password.',
        confirmLabel: 'Keluar Akun',
        cancelLabel: 'Batal',
      ) ??
      false;
  if (!context.mounted || !confirmed) return;
  await ref.read(nojposSessionProvider.notifier).logout();
  if (!context.mounted) return;
  context.go('/login');
}

class _TabText extends StatelessWidget {
  const _TabText({required this.text, this.active = false});

  final String text;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 56,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        border: Border(
          bottom: BorderSide(
            color: active ? MokposColors.warning : Colors.transparent,
            width: 4,
          ),
        ),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: active ? MokposColors.primary : MokposColors.muted,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _SearchLine extends StatelessWidget {
  const _SearchLine({this.onChanged});

  final ValueChanged<String>? onChanged;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: Row(
        children: [
          const Icon(LucideIcons.search, color: MokposColors.muted, size: 18),
          const SizedBox(width: 12),
          Expanded(
            child: TextField(
              onChanged: onChanged,
              decoration: const InputDecoration(
                hintText: 'Cari produk...',
                border: InputBorder.none,
              ),
            ),
          ),
          const Icon(
            LucideIcons.listFilter,
            color: MokposColors.muted,
            size: 18,
          ),
        ],
      ),
    );
  }
}

class _ReadOnlyNotice extends StatelessWidget {
  const _ReadOnlyNotice();

  @override
  Widget build(BuildContext context) {
    return const Align(
      alignment: Alignment.centerLeft,
      child: Text(
        'Mode cashier read-only. Edit produk dan kategori hanya untuk owner/admin.',
        style: TextStyle(
          color: MokposColors.muted,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

class _AddProductButton extends ConsumerWidget {
  const _AddProductButton();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return OutlinedButton(
      onPressed: () async {
        final result = await showDialog<_ProductFormResult>(
          context: context,
          builder: (context) => const _ProductDialog(),
        );
        if (result == null || !context.mounted) return;
        try {
          final product = await ref
              .read(posCatalogProvider.notifier)
              .addProduct(
                outletId: ref.read(nojposSessionProvider).outlet.id,
                name: result.name,
                category: result.category,
                price: result.price,
              );
          if (!context.mounted) return;
          NojposToast.success(context, '${product.name} ditambahkan');
        } catch (error) {
          if (!context.mounted) return;
          _showErrorSnackBar(context, _messageForUi(error));
        }
      },
      style: OutlinedButton.styleFrom(
        minimumSize: const Size.fromHeight(46),
        foregroundColor: MokposColors.primary,
        side: const BorderSide(color: MokposColors.primary),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
      ),
      child: const Text(
        'Tambah Produk',
        style: TextStyle(fontWeight: FontWeight.w900),
      ),
    );
  }
}

class _CategoryStrip extends ConsumerWidget {
  const _CategoryStrip({required this.categories, required this.canManage});

  final List<ProductCategory> categories;
  final bool canManage;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return SizedBox(
      height: 38,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        itemCount: categories.length + (canManage ? 1 : 0),
        separatorBuilder: (context, index) => const SizedBox(width: 8),
        itemBuilder: (context, index) {
          if (index >= categories.length) {
            return OutlinedButton.icon(
              onPressed: () async {
                final name = await _promptText(
                  context,
                  title: 'Tambah Kategori',
                  label: 'Nama kategori',
                );
                if (name == null || !context.mounted) return;
                try {
                  await ref.read(posCatalogProvider.notifier).addCategory(name);
                } catch (error) {
                  if (!context.mounted) return;
                  _showErrorSnackBar(context, _messageForUi(error));
                }
              },
              icon: const Icon(LucideIcons.plus, size: 16),
              label: const Text('Kategori'),
            );
          }
          final category = categories[index];
          final isSystemCategory = category.id.startsWith('__system_');
          final chip = Container(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: index == 0 ? MokposColors.primarySoft : Colors.white,
              border: Border.all(color: MokposColors.line),
              borderRadius: BorderRadius.circular(6),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  category.name,
                  style: TextStyle(
                    color: index == 0
                        ? MokposColors.primary
                        : MokposColors.text,
                    fontWeight: FontWeight.w800,
                    fontSize: 12,
                  ),
                ),
                if (canManage && !isSystemCategory) ...[
                  const SizedBox(width: 6),
                  const Icon(LucideIcons.pencil, size: 12),
                ],
              ],
            ),
          );
          if (!canManage || isSystemCategory) return chip;
          return InkWell(
            borderRadius: BorderRadius.circular(6),
            onTap: () async {
              final name = await _promptText(
                context,
                title: 'Edit Kategori',
                label: 'Nama kategori',
                initialValue: category.name,
              );
              if (name == null || !context.mounted) return;
              try {
                await ref
                    .read(posCatalogProvider.notifier)
                    .updateCategory(
                      id: category.id,
                      oldName: category.name,
                      name: name.trim(),
                    );
                if (!context.mounted) return;
                NojposToast.success(context, '${name.trim()} disimpan');
              } catch (error) {
                if (!context.mounted) return;
                _showErrorSnackBar(context, _messageForUi(error));
              }
            },
            child: chip,
          );
        },
      ),
    );
  }
}

class _ProductSettingRow extends StatelessWidget {
  const _ProductSettingRow({required this.product, required this.canManage});

  final Product product;
  final bool canManage;

  @override
  Widget build(BuildContext context) {
    return Consumer(
      builder: (context, ref, child) => ListTile(
        enabled: true,
        onTap: canManage
            ? () async {
                final result = await showDialog<_ProductFormResult>(
                  context: context,
                  builder: (context) => _ProductDialog(product: product),
                );
                if (result == null || !context.mounted) return;
                try {
                  await ref
                      .read(posCatalogProvider.notifier)
                      .updateProduct(
                        product: product,
                        outletId: ref.read(nojposSessionProvider).outlet.id,
                        name: result.name,
                        category: result.category,
                        price: result.price,
                      );
                  if (!context.mounted) return;
                  NojposToast.success(context, '${result.name} diperbarui');
                } catch (error) {
                  if (!context.mounted) return;
                  _showErrorSnackBar(context, _messageForUi(error));
                }
              }
            : null,
        leading: CircleAvatar(
          backgroundColor: const Color(0xFFE4EEE7),
          child: Text(product.name.characters.first.toUpperCase()),
        ),
        title: Text(
          product.name,
          style: const TextStyle(fontWeight: FontWeight.w900),
        ),
        subtitle: Text(product.category),
        trailing: Text(
          '${rupiah(product.price)}\n${canManage ? 'Edit' : 'Read-only'}',
          textAlign: TextAlign.right,
          style: const TextStyle(fontWeight: FontWeight.w800),
        ),
      ),
    );
  }
}

class _ProductDialog extends ConsumerStatefulWidget {
  const _ProductDialog({this.product});

  final Product? product;

  @override
  ConsumerState<_ProductDialog> createState() => _ProductDialogState();
}

class _ProductFormResult {
  const _ProductFormResult({
    required this.name,
    required this.category,
    required this.price,
  });

  final String name;
  final ProductCategory category;
  final int price;
}

Future<String?> _promptText(
  BuildContext context, {
  required String title,
  required String label,
  String initialValue = '',
}) {
  final controller = TextEditingController(text: initialValue);
  return showDialog<String>(
    context: context,
    builder: (context) => AlertDialog(
      backgroundColor: Colors.white,
      title: Text(title, style: const TextStyle(fontWeight: FontWeight.w900)),
      content: TextField(
        controller: controller,
        autofocus: true,
        decoration: InputDecoration(
          labelText: label,
          border: const OutlineInputBorder(),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Batal'),
        ),
        FilledButton(
          onPressed: () {
            final text = controller.text.trim();
            if (text.isEmpty) return;
            Navigator.of(context).pop(text);
          },
          child: const Text('Simpan'),
        ),
      ],
    ),
  ).whenComplete(controller.dispose);
}

void _showErrorSnackBar(BuildContext context, String message) {
  NojposToast.error(context, 'Operasi belum berhasil', description: message);
}

String _messageForUi(Object error) {
  if (error is ForbiddenApiException) {
    return error.message.isEmpty
        ? 'Akses ditolak. Hanya owner/admin yang bisa mengubah produk dan kategori.'
        : error.message;
  }
  if (error is ApiException) return error.message;
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}

class _ProductDialogState extends ConsumerState<_ProductDialog> {
  late final nameController = TextEditingController(
    text: widget.product?.name ?? 'Produk Baru',
  );
  late final priceController = TextEditingController(
    text: (widget.product?.price ?? 15000).toString(),
  );
  ProductCategory? category;

  @override
  void dispose() {
    nameController.dispose();
    priceController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final categories = ref
        .watch(posCatalogProvider)
        .categoryItems
        .where((item) => !item.id.startsWith('__system_'))
        .toList();
    final product = widget.product;
    if (category == null && categories.isNotEmpty) {
      category = categories.firstWhere(
        (item) =>
            item.id == product?.categoryId || item.name == product?.category,
        orElse: () => categories.first,
      );
    }

    return AlertDialog(
      backgroundColor: Colors.white,
      title: const Text(
        'Tambah Produk',
        style: TextStyle(fontWeight: FontWeight.w900),
      ),
      content: SizedBox(
        width: 420,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: nameController,
              autofocus: true,
              decoration: const InputDecoration(
                labelText: 'Nama Produk',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<ProductCategory>(
              initialValue: category,
              items: [
                for (final item in categories)
                  DropdownMenuItem(value: item, child: Text(item.name)),
              ],
              onChanged: (value) {
                if (value != null) setState(() => category = value);
              },
              decoration: const InputDecoration(
                labelText: 'Kategori',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: priceController,
              keyboardType: TextInputType.number,
              decoration: const InputDecoration(
                labelText: 'Harga',
                prefixText: 'Rp ',
                border: OutlineInputBorder(),
              ),
            ),
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
            final price =
                int.tryParse(
                  priceController.text.replaceAll(RegExp(r'[^0-9]'), ''),
                ) ??
                0;
            final selectedCategory = category;
            if (nameController.text.trim().isEmpty ||
                price <= 0 ||
                selectedCategory == null) {
              return;
            }
            Navigator.of(context).pop(
              _ProductFormResult(
                name: nameController.text.trim(),
                category: selectedCategory,
                price: price,
              ),
            );
          },
          child: const Text('Simpan'),
        ),
      ],
    );
  }
}

class _PurchaseDialog extends ConsumerStatefulWidget {
  const _PurchaseDialog();

  @override
  ConsumerState<_PurchaseDialog> createState() => _PurchaseDialogState();
}

class _PurchaseDialogState extends ConsumerState<_PurchaseDialog> {
  final supplierController = TextEditingController(text: 'Supplier Nusantara');
  final quantityController = TextEditingController(text: '12');
  final unitCostController = TextEditingController(text: '10000');
  String? productId;

  @override
  void dispose() {
    supplierController.dispose();
    quantityController.dispose();
    unitCostController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final products = ref.watch(posCatalogProvider).products;
    productId ??= products.firstOrNull?.id;
    return AlertDialog(
      backgroundColor: Colors.white,
      title: const Text(
        'Tambah Faktur Pembelian',
        style: TextStyle(fontWeight: FontWeight.w900),
      ),
      content: SizedBox(
        width: 420,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: supplierController,
              autofocus: true,
              decoration: const InputDecoration(
                labelText: 'Supplier',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              initialValue: productId,
              items: [
                for (final product in products)
                  DropdownMenuItem(
                    value: product.id,
                    child: Text(product.name),
                  ),
              ],
              onChanged: (value) => setState(() => productId = value),
              decoration: const InputDecoration(
                labelText: 'Produk',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: quantityController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(
                      labelText: 'Qty',
                      border: OutlineInputBorder(),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: TextField(
                    controller: unitCostController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(
                      labelText: 'Harga beli',
                      prefixText: 'Rp ',
                      border: OutlineInputBorder(),
                    ),
                  ),
                ),
              ],
            ),
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
            final quantity =
                int.tryParse(
                  quantityController.text.replaceAll(RegExp(r'[^0-9]'), ''),
                ) ??
                0;
            final unitCost =
                int.tryParse(
                  unitCostController.text.replaceAll(RegExp(r'[^0-9]'), ''),
                ) ??
                0;
            final selectedProductId = productId;
            if (supplierController.text.trim().isEmpty ||
                selectedProductId == null ||
                quantity <= 0 ||
                unitCost <= 0) {
              return;
            }
            Navigator.of(context).pop((
              supplierName: supplierController.text.trim(),
              productId: selectedProductId,
              quantity: quantity,
              unitCost: unitCost,
            ));
          },
          child: const Text('Tambah'),
        ),
      ],
    );
  }
}

class _AttendanceListDialog extends ConsumerWidget {
  const _AttendanceListDialog();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final records = ref.watch(attendanceControllerProvider).records;
    return AlertDialog(
      backgroundColor: Colors.white,
      title: const Text(
        'Daftar Kehadiran',
        style: TextStyle(fontWeight: FontWeight.w900),
      ),
      content: SizedBox(
        width: 620,
        height: 360,
        child: records.isEmpty
            ? const Center(
                child: Text(
                  'Belum ada data kehadiran',
                  style: TextStyle(color: MokposColors.muted),
                ),
              )
            : ListView.separated(
                itemCount: records.length,
                separatorBuilder: (context, index) =>
                    const Divider(height: 1, color: MokposColors.line),
                itemBuilder: (context, index) {
                  final record = records[index];
                  return ListTile(
                    leading: Icon(
                      record.isOpen
                          ? LucideIcons.clock3
                          : LucideIcons.circleCheck,
                      color: record.isOpen
                          ? MokposColors.primary
                          : MokposColors.accent,
                    ),
                    title: Text(
                      record.employee.name,
                      style: const TextStyle(fontWeight: FontWeight.w900),
                    ),
                    subtitle: Text(
                      'Masuk ${_formatDateTime(record.clockInAt)}'
                      '${record.clockOutAt == null ? '' : '\nKeluar ${_formatDateTime(record.clockOutAt!)}'}',
                    ),
                    trailing: Text(record.isOpen ? 'Aktif' : 'Selesai'),
                  );
                },
              ),
      ),
      actions: [
        FilledButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Tutup'),
        ),
      ],
    );
  }
}

class _AttendancePinPanel extends ConsumerStatefulWidget {
  const _AttendancePinPanel();

  @override
  ConsumerState<_AttendancePinPanel> createState() =>
      _AttendancePinPanelState();
}

class _AttendancePinPanelState extends ConsumerState<_AttendancePinPanel> {
  String? selectedEmployeeId;
  String pin = '';

  void _appendPin(String value) {
    if (pin.length >= 6) return;
    setState(() => pin += value);
  }

  void _backspacePin() {
    if (pin.isEmpty) return;
    setState(() => pin = pin.substring(0, pin.length - 1));
  }

  void _clearPin() {
    if (pin.isEmpty) return;
    setState(() => pin = '');
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final attendanceState = ref.watch(attendanceControllerProvider);
    final staff = attendanceState.staff.isEmpty
        ? [session.cashier]
        : attendanceState.staff;
    selectedEmployeeId ??= staff.firstOrNull?.id;
    final employee = staff.firstWhere(
      (employee) => employee.id == selectedEmployeeId,
      orElse: () => staff.first,
    );
    final openRecord = attendanceState.records.where(
      (record) => record.employee.id == employee.id && record.isOpen,
    );
    final isClockedIn = openRecord.isNotEmpty;
    return Padding(
      padding: const EdgeInsets.fromLTRB(72, 34, 72, 34),
      child: Column(
        children: [
          DropdownButtonFormField<String>(
            initialValue: employee.id,
            items: [
              for (final item in staff)
                DropdownMenuItem(value: item.id, child: Text(item.name)),
            ],
            onChanged: attendanceState.isLoading
                ? null
                : (value) => setState(() => selectedEmployeeId = value),
            decoration: const InputDecoration(
              labelText: 'Staff',
              border: OutlineInputBorder(),
            ),
          ),
          const SizedBox(height: 34),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: List.generate(
              6,
              (index) => Container(
                width: 10,
                height: 10,
                margin: const EdgeInsets.symmetric(horizontal: 10),
                decoration: BoxDecoration(
                  color: index < pin.length
                      ? MokposColors.primary
                      : MokposColors.line,
                  shape: BoxShape.circle,
                ),
              ),
            ),
          ),
          const SizedBox(height: 20),
          const Text(
            'Masukkan PIN',
            style: TextStyle(color: MokposColors.muted),
          ),
          const SizedBox(height: 44),
          Expanded(
            child: GridView.count(
              crossAxisCount: 3,
              childAspectRatio: 1.6,
              children: [
                for (final key in const [
                  '1',
                  '2',
                  '3',
                  '4',
                  '5',
                  '6',
                  '7',
                  '8',
                  '9',
                  'C',
                  '0',
                  '<',
                ])
                  InkWell(
                    onTap: attendanceState.isLoading
                        ? null
                        : () {
                            if (key == 'C') return _clearPin();
                            if (key == '<') return _backspacePin();
                            _appendPin(key);
                          },
                    child: Center(
                      child: key == '<'
                          ? const Icon(LucideIcons.delete, size: 16)
                          : Text(key),
                    ),
                  ),
              ],
            ),
          ),
          if (attendanceState.errorMessage != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 12),
              child: Text(
                attendanceState.errorMessage!,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: MokposColors.danger,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          FilledButton(
            onPressed: attendanceState.isLoading || pin.isEmpty
                ? null
                : () async {
                    final record = await ref
                        .read(attendanceControllerProvider.notifier)
                        .clockAttendance(employee: employee, pin: pin);
                    if (!context.mounted || record == null) return;
                    setState(() => pin = '');
                    NojposToast.success(
                      context,
                      record.isOpen
                          ? 'Clock in berhasil'
                          : 'Clock out berhasil',
                    );
                  },
            style: FilledButton.styleFrom(
              minimumSize: const Size.fromHeight(48),
              backgroundColor: MokposColors.primary,
              foregroundColor: Colors.white,
            ),
            child: attendanceState.isLoading
                ? const SizedBox.square(
                    dimension: 18,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: Colors.white,
                    ),
                  )
                : Text(isClockedIn ? 'Clock Out' : 'Clock In'),
          ),
        ],
      ),
    );
  }
}
