import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/nojpos_assets.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../../core/printing/receipt_printing_service.dart';
import '../../../core/share/receipt_share_service.dart';
import '../../../shared/widgets/nojpos_result_hero.dart';
import '../../../shared/widgets/nojpos_toast.dart';
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
      backgroundColor: NojposColors.canvas,
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 720),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: Column(
                children: [
                  const _SuccessHeader(),
                  const SizedBox(height: 18),
                  NojposResultHero.success(
                    title: 'Transaksi Berhasil',
                    subtitle:
                        'Pembayaran tercatat. Struk siap dicetak atau dibagikan.',
                    illustrationAsset: NojposAssets.paymentSuccess,
                    amount: rupiah(total),
                    summary: transaction?.number == null
                        ? null
                        : 'No. ${transaction!.number}',
                    compact: true,
                  ),
                  const SizedBox(height: 18),
                  _ReceiptSummaryCard(
                    transactionNumber: transaction?.number ?? '-',
                    customerName:
                        transaction?.order.customer?.name ?? 'Tanpa Pelanggan',
                    total: total,
                    paid: paid,
                    change: change,
                    paymentLabel: paymentLabel,
                    payments: payments,
                    itemDiscountTotal: transaction?.itemDiscountTotal ?? 0,
                    cartDiscountTotal: transaction?.cartDiscountTotal ?? 0,
                    onConfirmPending: (paymentId) async {
                      await ref
                          .read(nojposSessionProvider.notifier)
                          .confirmPaymentLine(paymentId);
                      if (!context.mounted) return;
                      NojposToast.success(
                        context,
                        'Pembayaran diterima',
                        description: 'Line pembayaran sudah ditandai diterima.',
                      );
                    },
                  ),
                  const SizedBox(height: 16),
                  _ReceiptActionBar(
                    receiptText: receiptText,
                    canReprint: canReprint,
                    onShare: receiptText.isEmpty
                        ? null
                        : () async {
                            try {
                              await shareService.shareReceipt(receiptText);
                              if (!context.mounted) return;
                              NojposToast.info(
                                context,
                                'Share sheet dibuka',
                                description:
                                    'Pilih WhatsApp atau aplikasi lain untuk mengirim struk.',
                              );
                            } catch (error) {
                              if (!context.mounted) return;
                              NojposToast.error(
                                context,
                                'Share gagal',
                                description: _messageForUi(error),
                              );
                            }
                          },
                    onPrint: transaction == null
                        ? null
                        : () async {
                            final result = await printingService.printReceipt(
                              transaction,
                              outlet: outlet,
                            );
                            if (!context.mounted) return;
                            setState(() {
                              printStatus = result.message;
                              canReprint = result.canReprint;
                            });
                            if (result.canReprint) {
                              NojposToast.warning(
                                context,
                                'Printer perlu dicek',
                                description: result.message,
                              );
                            } else {
                              NojposToast.success(
                                context,
                                'Struk dikirim ke printer',
                                description: result.message,
                              );
                            }
                          },
                    onCopy: receiptText.isEmpty
                        ? null
                        : () async {
                            await Clipboard.setData(
                              ClipboardData(text: receiptText),
                            );
                            if (!context.mounted) return;
                            NojposToast.success(
                              context,
                              'Struk digital disalin',
                            );
                          },
                  ),
                  if (printStatus != null) ...[
                    const SizedBox(height: 10),
                    Text(
                      printStatus!,
                      textAlign: TextAlign.center,
                      style: const TextStyle(
                        color: NojposColors.muted,
                        fontWeight: FontWeight.w700,
                        fontSize: 12,
                      ),
                    ),
                  ],
                  const SizedBox(height: 8),
                ],
              ),
            ),
          ),
        ),
      ),
      bottomNavigationBar: _SuccessBottomBar(
        onDone: () {
          ref.read(cartProvider.notifier).clear();
          ref.read(checkoutDetailsProvider.notifier).clear();
          context.go('/pos');
        },
      ),
    );
  }
}

class _SuccessBottomBar extends StatelessWidget {
  const _SuccessBottomBar({required this.onDone});

  final VoidCallback onDone;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      top: false,
      child: Container(
        padding: const EdgeInsets.fromLTRB(20, 12, 20, 16),
        decoration: const BoxDecoration(
          color: Colors.white,
          border: Border(top: BorderSide(color: NojposColors.line)),
        ),
        child: Center(
          heightFactor: 1,
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 720),
            child: FilledButton.icon(
              key: const ValueKey('success_done'),
              onPressed: onDone,
              icon: const Icon(LucideIcons.arrowRight),
              label: const Text('Selesai dan kembali ke POS'),
              style: FilledButton.styleFrom(
                backgroundColor: NojposColors.primary,
                foregroundColor: Colors.white,
                minimumSize: const Size.fromHeight(54),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(NojposRadius.lg),
                ),
                textStyle: const TextStyle(fontWeight: FontWeight.w900),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _SuccessHeader extends StatelessWidget {
  const _SuccessHeader();

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Container(
          width: 42,
          height: 42,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(14),
            boxShadow: NojposShadow.soft,
          ),
          child: Padding(
            padding: const EdgeInsets.all(7),
            child: SvgPicture.asset(
              NojposAssets.logoSymbol,
              fit: BoxFit.contain,
            ),
          ),
        ),
        const SizedBox(width: 10),
        SizedBox(
          width: 104,
          height: 30,
          child: SvgPicture.asset(
            NojposAssets.logoFull,
            fit: BoxFit.contain,
            alignment: Alignment.centerLeft,
          ),
        ),
      ],
    );
  }
}

class _ReceiptSummaryCard extends StatelessWidget {
  const _ReceiptSummaryCard({
    required this.transactionNumber,
    required this.customerName,
    required this.total,
    required this.paid,
    required this.change,
    required this.paymentLabel,
    required this.payments,
    required this.itemDiscountTotal,
    required this.cartDiscountTotal,
    required this.onConfirmPending,
  });

  final String transactionNumber;
  final String customerName;
  final int total;
  final int paid;
  final int change;
  final String paymentLabel;
  final List<dynamic> payments;
  final int itemDiscountTotal;
  final int cartDiscountTotal;
  final Future<void> Function(String paymentId) onConfirmPending;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(NojposRadius.xl),
        border: Border.all(color: NojposColors.line),
        boxShadow: NojposShadow.card,
      ),
      child: Column(
        children: [
          const Row(
            children: [
              Icon(LucideIcons.fileText, color: NojposColors.primary, size: 20),
              SizedBox(width: 8),
              Text(
                'Ringkasan Struk',
                style: TextStyle(
                  color: NojposColors.text,
                  fontWeight: FontWeight.w900,
                  fontSize: 16,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          _ReceiptRow(label: 'Nomor', value: transactionNumber),
          const SizedBox(height: 8),
          _ReceiptRow(label: 'Pelanggan', value: customerName),
          if (itemDiscountTotal > 0) ...[
            const SizedBox(height: 12),
            _ReceiptRow(label: 'Diskon Item', value: rupiah(itemDiscountTotal)),
          ],
          if (cartDiscountTotal > 0) ...[
            const SizedBox(height: 12),
            _ReceiptRow(label: 'Diskon Cart', value: rupiah(cartDiscountTotal)),
          ],
          const SizedBox(height: 10),
          const Divider(color: NojposColors.line),
          const SizedBox(height: 10),
          _ReceiptRow(
            label: 'Total Tagihan',
            value: rupiah(total),
            emphasized: true,
          ),
          const SizedBox(height: 12),
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
                  key: const ValueKey('confirm_pending_payment'),
                  onPressed: () => onConfirmPending(payment.paymentId!),
                  icon: const Icon(LucideIcons.check, size: 16),
                  label: const Text('Tandai sudah diterima'),
                ),
              ),
            ],
            const SizedBox(height: 12),
          ],
          if (payments.isEmpty)
            _ReceiptRow(label: paymentLabel, value: rupiah(paid)),
          const SizedBox(height: 8),
          _ReceiptRow(
            label: 'Kembalian',
            value: rupiah(change),
            emphasized: true,
          ),
        ],
      ),
    );
  }
}

class _ReceiptActionBar extends StatelessWidget {
  const _ReceiptActionBar({
    required this.receiptText,
    required this.canReprint,
    required this.onShare,
    required this.onPrint,
    required this.onCopy,
  });

  final String receiptText;
  final bool canReprint;
  final VoidCallback? onShare;
  final VoidCallback? onPrint;
  final VoidCallback? onCopy;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(NojposRadius.xl),
        border: Border.all(color: NojposColors.line),
      ),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final compact = constraints.maxWidth < 420;
          if (compact) {
            return Wrap(
              alignment: WrapAlignment.center,
              spacing: 8,
              runSpacing: 8,
              children: [
                _CompactReceiptAction(
                  onPressed: onShare,
                  icon: LucideIcons.share2,
                  label: 'Bagikan',
                ),
                _CompactReceiptAction(
                  onPressed: onPrint,
                  icon: LucideIcons.printer,
                  label: canReprint ? 'Cetak Ulang' : 'Cetak',
                ),
                _CompactReceiptAction(
                  onPressed: onCopy,
                  icon: LucideIcons.copy,
                  label: 'Salin',
                ),
              ],
            );
          }
          return Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: onShare,
                  icon: const Icon(LucideIcons.share2, size: 18),
                  label: const Text('Bagikan/WhatsApp'),
                  style: _outlineStyle(),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: onPrint,
                  icon: const Icon(LucideIcons.printer, size: 18),
                  label: Text(canReprint ? 'Cetak Ulang' : 'Cetak'),
                  style: _outlineStyle(),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: TextButton.icon(
                  onPressed: onCopy,
                  icon: const Icon(LucideIcons.copy, size: 16),
                  label: const Text('Salin'),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  ButtonStyle _outlineStyle() {
    return OutlinedButton.styleFrom(
      foregroundColor: NojposColors.primary,
      side: const BorderSide(color: NojposColors.primary),
      minimumSize: const Size.fromHeight(46),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(NojposRadius.md),
      ),
      textStyle: const TextStyle(fontWeight: FontWeight.w900),
    );
  }
}

class _CompactReceiptAction extends StatelessWidget {
  const _CompactReceiptAction({
    required this.onPressed,
    required this.icon,
    required this.label,
  });

  final VoidCallback? onPressed;
  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return OutlinedButton.icon(
      onPressed: onPressed,
      icon: Icon(icon, size: 16),
      label: Text(label),
      style: OutlinedButton.styleFrom(
        foregroundColor: NojposColors.primary,
        side: const BorderSide(color: NojposColors.line),
        visualDensity: VisualDensity.compact,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(NojposRadius.md),
        ),
        textStyle: const TextStyle(fontWeight: FontWeight.w900),
      ),
    );
  }
}

String _messageForUi(Object error) {
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}

class _ReceiptRow extends StatelessWidget {
  const _ReceiptRow({
    required this.label,
    required this.value,
    this.emphasized = false,
  });

  final String label;
  final String value;
  final bool emphasized;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: Text(
            label,
            style: TextStyle(
              color: emphasized ? NojposColors.text : NojposColors.muted,
              fontWeight: emphasized ? FontWeight.w900 : FontWeight.w700,
              fontSize: emphasized ? 15 : 14,
            ),
          ),
        ),
        const SizedBox(width: 12),
        Text(
          value,
          textAlign: TextAlign.right,
          style: TextStyle(
            color: emphasized ? NojposColors.primary : NojposColors.text,
            fontWeight: FontWeight.w900,
            fontSize: emphasized ? 16 : 14,
          ),
        ),
      ],
    );
  }
}
