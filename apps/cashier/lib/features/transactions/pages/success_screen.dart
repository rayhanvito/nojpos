import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../../../core/printing/receipt_printing_service.dart';
import '../../../core/share/receipt_share_service.dart';
import '../../pos/formatters.dart';
import '../../pos/providers/pos_providers.dart';

class SuccessScreen extends ConsumerStatefulWidget {
  const SuccessScreen({super.key});

  @override
  ConsumerState<SuccessScreen> createState() => _SuccessScreenState();
}

class _SuccessScreenState extends ConsumerState<SuccessScreen> {
  String? printStatus;
  bool canReprint = false;

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final transaction = session.lastTransaction;
    final outlet = session.outlet;
    final fallbackTotal = ref.watch(cartTotalProvider);
    final total = transaction?.total ?? fallbackTotal;
    final paid = transaction?.paidAmount ?? total;
    final change = transaction?.change ?? 0;
    final payments = transaction?.payments ?? const [];
    final paymentLabel = payments.isEmpty
        ? 'Tunai'
        : payments.first.method.label;
    final printingService = ref.watch(receiptPrintingServiceProvider);
    final shareService = ref.watch(receiptShareServiceProvider);
    final receiptText = transaction == null
        ? ''
        : printingService.buildDigitalReceipt(transaction, outlet: outlet);

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
                  child: SingleChildScrollView(
                    child: Padding(
                      padding: const EdgeInsets.fromLTRB(34, 22, 34, 22),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Container(
                            width: 22,
                            height: 22,
                            decoration: const BoxDecoration(
                              color: MokposColors.success,
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
                            label: 'Nomor',
                            value: transaction?.number ?? '-',
                          ),
                          const SizedBox(height: 12),
                          _ReceiptRow(
                            label: 'Pelanggan',
                            value:
                                transaction?.order.customer?.name ??
                                'Tanpa Pelanggan',
                          ),
                          if ((transaction?.itemDiscountTotal ?? 0) > 0) ...[
                            const SizedBox(height: 12),
                            _ReceiptRow(
                              label: 'Diskon Item',
                              value: rupiah(transaction!.itemDiscountTotal),
                            ),
                          ],
                          if ((transaction?.cartDiscountTotal ?? 0) > 0) ...[
                            const SizedBox(height: 12),
                            _ReceiptRow(
                              label: 'Diskon Cart',
                              value: rupiah(transaction!.cartDiscountTotal),
                            ),
                          ],
                          const SizedBox(height: 16),
                          _ReceiptRow(
                            label: 'Total Tagihan',
                            value: rupiah(total),
                          ),
                          const SizedBox(height: 16),
                          for (final payment in payments) ...[
                            _ReceiptRow(
                              label: payment.methodName,
                              value: rupiah(payment.amount),
                            ),
                            if (!payment.isCash &&
                                payment.status == 'pending' &&
                                payment.paymentId != null) ...[
                              const SizedBox(height: 8),
                              SizedBox(
                                width: double.infinity,
                                child: FilledButton.icon(
                                  key: const ValueKey(
                                    'confirm_pending_payment',
                                  ),
                                  onPressed: () async {
                                    await ref
                                        .read(nojposSessionProvider.notifier)
                                        .confirmPaymentLine(payment.paymentId!);
                                    if (!context.mounted) return;
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      const SnackBar(
                                        content: Text(
                                          'Pembayaran ditandai sudah diterima',
                                        ),
                                      ),
                                    );
                                  },
                                  icon: const Icon(LucideIcons.check, size: 16),
                                  label: const Text('Tandai sudah diterima'),
                                ),
                              ),
                            ],
                            const SizedBox(height: 12),
                          ],
                          if (payments.isEmpty)
                            _ReceiptRow(
                              label: paymentLabel,
                              value: rupiah(paid),
                            ),
                          const SizedBox(height: 28),
                          const Divider(color: MokposColors.line),
                          const SizedBox(height: 16),
                          _ReceiptRow(
                            label: 'Kembalian',
                            value: rupiah(change),
                          ),
                          const SizedBox(height: 18),
                          Row(
                            children: [
                              Expanded(
                                child: OutlinedButton.icon(
                                  onPressed: receiptText.isEmpty
                                      ? null
                                      : () async {
                                          try {
                                            await shareService.shareReceipt(
                                              receiptText,
                                            );
                                            if (!context.mounted) return;
                                            ScaffoldMessenger.of(
                                              context,
                                            ).showSnackBar(
                                              const SnackBar(
                                                content: Text(
                                                  'Share sheet dibuka. Pilih WhatsApp atau aplikasi lain.',
                                                ),
                                              ),
                                            );
                                          } catch (error) {
                                            if (!context.mounted) return;
                                            ScaffoldMessenger.of(
                                              context,
                                            ).showSnackBar(
                                              SnackBar(
                                                content: Text(
                                                  'Share gagal: ${_messageForUi(error)}',
                                                ),
                                              ),
                                            );
                                          }
                                        },
                                  icon: const Icon(
                                    LucideIcons.share2,
                                    size: 18,
                                  ),
                                  label: const Text('Bagikan/WhatsApp'),
                                  style: _outlineStyle(),
                                ),
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: OutlinedButton.icon(
                                  onPressed: transaction == null
                                      ? null
                                      : () async {
                                          final result = await printingService
                                              .printReceipt(
                                                transaction,
                                                outlet: outlet,
                                              );
                                          if (!context.mounted) return;
                                          setState(() {
                                            printStatus = result.message;
                                            canReprint = result.canReprint;
                                          });
                                          ScaffoldMessenger.of(
                                            context,
                                          ).showSnackBar(
                                            SnackBar(
                                              content: Text(result.message),
                                            ),
                                          );
                                        },
                                  icon: const Icon(
                                    LucideIcons.printer,
                                    size: 18,
                                  ),
                                  label: Text(
                                    canReprint ? 'Cetak Ulang' : 'Cetak',
                                  ),
                                  style: _outlineStyle(),
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 8),
                          TextButton.icon(
                            onPressed: receiptText.isEmpty
                                ? null
                                : () async {
                                    await Clipboard.setData(
                                      ClipboardData(text: receiptText),
                                    );
                                    if (!context.mounted) return;
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      const SnackBar(
                                        content: Text('Struk digital disalin.'),
                                      ),
                                    );
                                  },
                            icon: const Icon(LucideIcons.copy, size: 16),
                            label: const Text('Salin struk digital'),
                          ),
                          if (printStatus != null) ...[
                            const SizedBox(height: 8),
                            Text(
                              printStatus!,
                              textAlign: TextAlign.center,
                              style: const TextStyle(
                                color: MokposColors.muted,
                                fontWeight: FontWeight.w700,
                                fontSize: 12,
                              ),
                            ),
                          ],
                          const SizedBox(height: 12),
                          FilledButton(
                            key: const ValueKey('success_done'),
                            onPressed: () {
                              ref.read(cartProvider.notifier).clear();
                              ref
                                  .read(checkoutDetailsProvider.notifier)
                                  .clear();
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

String _messageForUi(Object error) {
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
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
