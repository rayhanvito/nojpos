import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../core/network/api_client.dart';
import '../../../core/outbox/checkout_outbox.dart';
import '../../../app/theme.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../../../shared/models/nojpos_models.dart';
import '../../../shared/widgets/nojpos_asset_icon.dart';
import '../../../shared/widgets/nojpos_toast.dart';
import '../../connectivity/widgets/connectivity_status_chip.dart';
import '../../pos/formatters.dart';
import '../../pos/models/cart_item.dart';
import '../../pos/providers/parked_order_lease_controller.dart';
import '../../pos/providers/pos_providers.dart';
import '../../pos/widgets/pos_dialogs.dart';
import '../../transactions/repositories/transaction_repository.dart';
import '../providers/pending_payment_controller.dart';
import '../widgets/pending_payment_status_view.dart';
import '../widgets/promotion_quote_panel.dart';

CheckoutPayment buildSingleCheckoutPayment({
  required PaymentMethodConfig method,
  required int total,
  required int paidAmount,
  required String reference,
}) {
  if (method.isCash) {
    return CheckoutPayment(
      method: method.method,
      amount: paidAmount,
      isCash: true,
    );
  }
  final normalizedReference = reference.trim();
  if (normalizedReference.isEmpty) {
    throw ArgumentError('Reference non-tunai wajib diisi.');
  }
  return CheckoutPayment(
    method: method.method,
    amount: total,
    isCash: false,
    reference: normalizedReference,
  );
}

void requireNonCashReferences(List<CheckoutPayment> payments) {
  for (final payment in payments) {
    if (payment.isCash) continue;
    if (payment.reference?.trim().isNotEmpty ?? false) continue;
    throw ArgumentError('Reference non-tunai wajib diisi.');
  }
}

final paymentPromotionCodesProvider =
    NotifierProvider<PaymentPromotionCodesNotifier, List<String>>(
      PaymentPromotionCodesNotifier.new,
    );

class PaymentPromotionCodesNotifier extends Notifier<List<String>> {
  @override
  List<String> build() => const [];

  void add(String code) {
    if (state.contains(code)) return;
    state = [...state, code];
  }

  void remove(String code) {
    state = [
      for (final current in state)
        if (current != code) current,
    ];
  }
}

final checkoutQuoteProvider = FutureProvider.autoDispose<CheckoutQuote>((ref) {
  final items = ref.watch(cartProvider);
  final details = ref.watch(checkoutDetailsProvider);
  final promotionCodes = ref.watch(paymentPromotionCodesProvider);
  return ref
      .watch(nojposSessionProvider.notifier)
      .quoteCheckout(
        cartItems: items,
        cartDiscount: details.cartDiscount,
        promotionCodes: promotionCodes,
        notes: details.notes,
        servedBy: details.servedBy,
      );
});

class PaymentScreen extends ConsumerStatefulWidget {
  const PaymentScreen({super.key});

  @override
  ConsumerState<PaymentScreen> createState() => _PaymentScreenState();
}

class _PaymentScreenState extends ConsumerState<PaymentScreen> {
  String? selectedMethod;
  int paidAmount = 0;
  final paymentLines = <CheckoutPayment>[];
  final referenceController = TextEditingController();
  final referenceFocusNode = FocusNode();
  final promotionCodeController = TextEditingController();
  bool submitting = false;

  @override
  void initState() {
    super.initState();
    referenceController.addListener(_refreshReferenceState);
    ref.listenManual(pendingPaymentControllerProvider, (previous, next) {
      if (!mounted ||
          previous?.transaction?.status == next.transaction?.status) {
        return;
      }
      if (next.transaction?.status == 'paid') {
        context.go('/success');
      }
    });
  }

  void _refreshReferenceState() {
    if (mounted) setState(() {});
  }

  @override
  void dispose() {
    referenceController.removeListener(_refreshReferenceState);
    referenceController.dispose();
    referenceFocusNode.dispose();
    promotionCodeController.dispose();
    ref.read(pendingPaymentControllerProvider.notifier).stop();
    super.dispose();
  }

  Future<void> _pay(
    List<CartItem> items,
    int total,
    CheckoutQuote quote,
  ) async {
    final methods = ref.read(nojposSessionProvider).outlet.paymentMethods;
    final selected = _methodByName(methods, selectedMethod);
    late final List<CheckoutPayment> lines;
    try {
      lines = paymentLines.isNotEmpty
          ? [...paymentLines]
          : [
              buildSingleCheckoutPayment(
                method: selected,
                total: total,
                paidAmount: paidAmount,
                reference: referenceController.text,
              ),
            ];
      requireNonCashReferences(lines);
    } on ArgumentError catch (error) {
      NojposToast.warning(context, error.message.toString());
      return;
    }
    setState(() => submitting = true);
    try {
      final details = ref.read(checkoutDetailsProvider);
      final result = await ref
          .read(nojposSessionProvider.notifier)
          .completeCheckout(
            cartItems: items,
            payments: lines,
            paidAmount: lines.fold(0, (sum, line) => sum + line.amount),
            cartDiscount: details.cartDiscount,
            promotionCodes: ref.read(paymentPromotionCodesProvider),
            notes: details.notes,
            servedBy: details.servedBy,
            quote: quote,
          );
      if (!mounted) return;
      if (result is SentCheckoutResult) {
        await ref
            .read(parkedOrderLeaseControllerProvider.notifier)
            .releaseAfterCheckout();
        if (!mounted) return;
        if (result.transaction.status == 'payment_pending') {
          ref
              .read(pendingPaymentControllerProvider.notifier)
              .start(result.transaction);
        } else {
          context.go('/success');
        }
      } else if (result is QueuedCheckoutResult) {
        NojposToast.warning(
          context,
          'Checkout belum terkonfirmasi',
          description: 'Pulihkan atau kirim ulang checkout yang sama.',
        );
        context.go('/pos');
      }
    } catch (error) {
      if (!mounted) return;
      if (error is ApiException && error.code == 'QUOTE_STALE') {
        ref.invalidate(checkoutQuoteProvider);
      }
      NojposToast.error(
        context,
        'Checkout gagal',
        description: _checkoutErrorMessage(error),
      );
    } finally {
      if (mounted) setState(() => submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final pending = ref.watch(pendingPaymentControllerProvider);
    if (pending.transaction != null) {
      return PendingPaymentStatusView(
        transaction: pending.transaction!,
        errorMessage: pending.errorMessage,
        onCancel: () {
          ref.read(pendingPaymentControllerProvider.notifier).stop();
          context.go('/pos');
        },
      );
    }
    final items = ref.watch(cartProvider);
    final baseTotal = ref.watch(cartTotalProvider);
    final promotionCodes = ref.watch(paymentPromotionCodesProvider);
    final quoteAsync = ref.watch(checkoutQuoteProvider);
    final quote = quoteAsync.asData?.value;
    final isQuoteLoading = quote == null && quoteAsync.isLoading;
    final quoteError = quote == null ? quoteAsync.error : null;
    final total = quote?.grandTotal ?? baseTotal;
    final itemDiscountTotal = ref.watch(itemDiscountTotalProvider);
    final cartDiscount = ref.watch(checkoutDetailsProvider).cartDiscount;
    final session = ref.watch(nojposSessionProvider);
    final hasUnresolvedCheckout = ref
        .watch(checkoutOutboxControllerProvider)
        .any((item) => item.blocksClose);
    final methods = session.outlet.paymentMethods;
    if (selectedMethod == null && methods.isNotEmpty) {
      selectedMethod = methods.first.method;
    }
    final selected = _methodByName(methods, selectedMethod);
    final lineTotal = paymentLines.fold(0, (sum, line) => sum + line.amount);
    final effectivePaid = paymentLines.isEmpty ? paidAmount : lineTotal;
    final remaining = (total - effectivePaid).clamp(0, total).toInt();
    final nonCashTotal = paymentLines
        .where((line) => !line.isCash)
        .fold(0, (sum, line) => sum + line.amount);
    final cashDue = (total - nonCashTotal).clamp(0, total).toInt();
    final cashPaid = paymentLines.isEmpty && selected.isCash
        ? paidAmount
        : paymentLines
              .where((line) => line.isCash)
              .fold(0, (sum, line) => sum + line.amount);
    final change = (cashPaid - cashDue).clamp(0, cashPaid).toInt();

    return Scaffold(
      backgroundColor: Colors.white,
      body: SafeArea(
        child: Column(
          children: [
            _PaymentTopBar(onBack: () => context.go('/pos')),
            Expanded(
              child: LayoutBuilder(
                builder: (context, constraints) {
                  final mobile = constraints.maxWidth < 820;
                  final summary = _PaymentOrderSummary(
                    items: items,
                    total: total,
                    itemDiscountTotal: itemDiscountTotal,
                    cartDiscount: cartDiscount,
                    paymentLines: paymentLines,
                    quote: quote,
                    quoteLoading: isQuoteLoading,
                    quoteError: quoteError,
                    promotionCodes: promotionCodes,
                    promotionCodeController: promotionCodeController,
                    onApplyPromotionCode: _applyPromotionCode,
                    onRemovePromotionCode: _removePromotionCode,
                    canPay:
                        !submitting &&
                        !isQuoteLoading &&
                        quoteError == null &&
                        quote != null &&
                        items.isNotEmpty &&
                        !hasUnresolvedCheckout &&
                        remaining == 0 &&
                        methods.isNotEmpty &&
                        _hasReadyPaymentReference(
                          selected: selected,
                          paymentLines: paymentLines,
                          singleReference: referenceController.text,
                        ),
                    onPay: () => _pay(items, total, quote!),
                  );
                  final paymentContent = Column(
                    children: [
                      _PaymentStats(
                        total: total,
                        remaining: remaining,
                        change: change,
                        showChange:
                            selected.isCash ||
                            paymentLines.any((line) => line.isCash),
                      ),
                      _PaymentActions(
                        onSplit: () async {
                          final result = await showSplitPaymentDialog(
                            context,
                            methods: methods,
                            total: total,
                            existing: paymentLines,
                          );
                          if (result != null && mounted) {
                            setState(() {
                              paymentLines
                                ..clear()
                                ..addAll(result);
                              paidAmount = result.fold(
                                0,
                                (sum, line) => sum + line.amount,
                              );
                            });
                          }
                        },
                        onHeld: () => _saveHeld(items),
                      ),
                      Expanded(
                        child: mobile
                            ? Column(
                                children: [
                                  _MethodRail(
                                    methods: methods,
                                    selected: selected.method,
                                    onSelect: (method) => _selectMethod(
                                      methods: methods,
                                      method: method,
                                      total: total,
                                    ),
                                    horizontal: true,
                                  ),
                                  Expanded(
                                    child: _PaymentAmountPanel(
                                      method: selected,
                                      total: total,
                                      paidAmount: paidAmount,
                                      referenceController: referenceController,
                                      referenceFocusNode: referenceFocusNode,
                                      onSelectAmount: (amount) {
                                        setState(() => paidAmount = amount);
                                      },
                                      onOpenAmountDialog: () async {
                                        final result = await showDialog<int>(
                                          context: context,
                                          barrierColor: Colors.black54,
                                          builder: (context) => AmountDialog(
                                            total: total,
                                            initialAmount: paidAmount,
                                          ),
                                        );
                                        if (result != null && mounted) {
                                          setState(() => paidAmount = result);
                                        }
                                      },
                                    ),
                                  ),
                                ],
                              )
                            : Row(
                                children: [
                                  _MethodRail(
                                    methods: methods,
                                    selected: selected.method,
                                    onSelect: (method) => _selectMethod(
                                      methods: methods,
                                      method: method,
                                      total: total,
                                    ),
                                  ),
                                  Expanded(
                                    child: _PaymentAmountPanel(
                                      method: selected,
                                      total: total,
                                      paidAmount: paidAmount,
                                      referenceController: referenceController,
                                      referenceFocusNode: referenceFocusNode,
                                      onSelectAmount: (amount) {
                                        setState(() => paidAmount = amount);
                                      },
                                      onOpenAmountDialog: () async {
                                        final result = await showDialog<int>(
                                          context: context,
                                          barrierColor: Colors.black54,
                                          builder: (context) => AmountDialog(
                                            total: total,
                                            initialAmount: paidAmount,
                                          ),
                                        );
                                        if (result != null && mounted) {
                                          setState(() => paidAmount = result);
                                        }
                                      },
                                    ),
                                  ),
                                ],
                              ),
                      ),
                    ],
                  );

                  if (mobile) {
                    return Column(
                      children: [
                        Expanded(child: paymentContent),
                        SizedBox(height: 330, child: summary),
                      ],
                    );
                  }

                  return Row(
                    children: [
                      Expanded(child: paymentContent),
                      SizedBox(width: 430, child: summary),
                    ],
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _applyPromotionCode() {
    final code = promotionCodeController.text.trim().toUpperCase();
    if (code.isEmpty) return;
    ref.read(paymentPromotionCodesProvider.notifier).add(code);
    promotionCodeController.clear();
  }

  void _removePromotionCode(String code) {
    ref.read(paymentPromotionCodesProvider.notifier).remove(code);
  }

  void _selectMethod({
    required List<PaymentMethodConfig> methods,
    required String method,
    required int total,
  }) {
    setState(() {
      selectedMethod = method;
      paymentLines.clear();
      final config = _methodByName(methods, method);
      paidAmount = config.isCash ? 0 : total;
      referenceController.clear();
    });
    final config = _methodByName(methods, method);
    if (!config.isCash) {
      WidgetsBinding.instance.addPostFrameCallback(
        (_) => referenceFocusNode.requestFocus(),
      );
    }
  }

  Future<void> _saveHeld(List<CartItem> items) async {
    if (items.isEmpty) return;
    final details = ref.read(checkoutDetailsProvider);
    try {
      await ref
          .read(parkedOrderLeaseControllerProvider.notifier)
          .saveCurrentCartAsParked(
            cartItems: items,
            cartDiscount: details.cartDiscount,
            notes: details.notes,
            servedBy: details.servedBy,
          );
      if (!mounted) return;
      ref.read(cartProvider.notifier).clear();
      context.go('/orders');
    } catch (error) {
      if (!mounted) return;
      NojposToast.error(
        context,
        'Parked order belum bisa dilepas',
        description: parkedOrderLeaseMessage(error),
      );
    }
  }
}

String _checkoutErrorMessage(Object error) {
  if (error is ApiException && error.code == 'QUOTE_STALE') {
    return 'Quote promo sudah berubah. Total diperbarui, coba checkout lagi.';
  }
  if (error is ApiException && error.code == 'TOTAL_MISMATCH') {
    return 'Harga atau pajak berubah. Ulangi checkout untuk mengambil total terbaru.';
  }
  if (error is ApiException && error.code == 'STORE_CLOSED_CHECKOUT_BLOCKED') {
    return 'Toko sedang tutup. Buka toko sebelum checkout.';
  }
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}

bool _hasReadyPaymentReference({
  required PaymentMethodConfig selected,
  required List<CheckoutPayment> paymentLines,
  required String singleReference,
}) {
  if (paymentLines.isNotEmpty) {
    return paymentLines.every(
      (line) => line.isCash || (line.reference?.trim().isNotEmpty ?? false),
    );
  }
  return selected.isCash || singleReference.trim().isNotEmpty;
}

PaymentMethodConfig _methodByName(
  List<PaymentMethodConfig> methods,
  String? name,
) {
  if (methods.isEmpty) {
    return const PaymentMethodConfig(method: 'Tidak tersedia', isCash: false);
  }
  return methods.firstWhere(
    (method) => method.method == name,
    orElse: () => methods.first,
  );
}

class _PaymentTopBar extends StatelessWidget {
  const _PaymentTopBar({required this.onBack});

  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 58,
      color: MokposColors.primary,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      child: Row(
        children: [
          IconButton(
            onPressed: onBack,
            icon: const Icon(LucideIcons.arrowLeft),
            color: Colors.white,
          ),
          const SizedBox(width: 4),
          const Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Pembayaran',
                style: TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.w900,
                  fontSize: 16,
                ),
              ),
              SizedBox(height: 1),
              ConnectivityStatusChip(
                textStyle: TextStyle(
                  color: MokposColors.onPrimaryMuted,
                  fontSize: 11,
                ),
                dotSize: 8,
              ),
            ],
          ),
          const Spacer(),
          const Text(
            'NojPOS',
            style: TextStyle(
              color: Colors.white,
              fontSize: 24,
              fontWeight: FontWeight.w900,
            ),
          ),
          const Spacer(flex: 2),
        ],
      ),
    );
  }
}

class _PaymentStats extends StatelessWidget {
  const _PaymentStats({
    required this.total,
    required this.remaining,
    required this.change,
    required this.showChange,
  });

  final int total;
  final int remaining;
  final int change;
  final bool showChange;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 74,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: Row(
        children: [
          Expanded(
            child: _StatBox(label: 'Total Tagihan', value: rupiah(total)),
          ),
          Expanded(
            child: _StatBox(
              label: 'Sisa Tagihan',
              value: rupiah(remaining),
              valueColor: remaining > 0
                  ? MokposColors.danger
                  : MokposColors.primary,
            ),
          ),
          Expanded(
            child: _StatBox(
              label: 'Kembalian',
              value: showChange ? rupiah(change) : '-',
              valueColor: MokposColors.accent,
            ),
          ),
        ],
      ),
    );
  }
}

class _StatBox extends StatelessWidget {
  const _StatBox({required this.label, required this.value, this.valueColor});

  final String label;
  final String value;
  final Color? valueColor;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Text(
          label,
          style: const TextStyle(
            color: MokposColors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w700,
          ),
        ),
        const SizedBox(height: 5),
        Text(
          value,
          style: TextStyle(
            color: valueColor ?? MokposColors.text,
            fontSize: 17,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}

class _PaymentActions extends ConsumerWidget {
  const _PaymentActions({required this.onSplit, required this.onHeld});

  final VoidCallback onSplit;
  final VoidCallback onHeld;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Container(
      height: 54,
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: Row(
        children: [
          Expanded(
            child: _PaymentAction(
              icon: LucideIcons.creditCard,
              label: 'Pisah Bayar',
              onTap: onSplit,
            ),
          ),
          const VerticalDivider(width: 1, color: MokposColors.line),
          Expanded(
            child: _PaymentAction(
              icon: LucideIcons.fileText,
              label: 'Simpan Order',
              onTap: onHeld,
            ),
          ),
          const VerticalDivider(width: 1, color: MokposColors.line),
          Expanded(
            child: _PaymentAction(
              icon: LucideIcons.pencil,
              label: 'Catatan',
              onTap: () => showNotesDialog(context, ref),
            ),
          ),
        ],
      ),
    );
  }
}

class _PaymentAction extends StatelessWidget {
  const _PaymentAction({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, size: 18, color: MokposColors.primary),
          const SizedBox(width: 8),
          Text(
            label,
            style: const TextStyle(
              color: MokposColors.primary,
              fontWeight: FontWeight.w900,
              fontSize: 13,
            ),
          ),
        ],
      ),
    );
  }
}

class _MethodRail extends StatelessWidget {
  const _MethodRail({
    required this.methods,
    required this.selected,
    required this.onSelect,
    this.horizontal = false,
  });

  final List<PaymentMethodConfig> methods;
  final String selected;
  final ValueChanged<String> onSelect;
  final bool horizontal;

  @override
  Widget build(BuildContext context) {
    final header = Container(
      height: 44,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: const Row(
        children: [
          Expanded(
            child: Text(
              'Metode Pembayaran',
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: MokposColors.muted,
                fontSize: 12,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          Icon(LucideIcons.listFilter, size: 16, color: MokposColors.text),
        ],
      ),
    );

    final empty = const Padding(
      padding: EdgeInsets.all(12),
      child: Text(
        'Metode pembayaran belum dikonfigurasi',
        style: TextStyle(color: MokposColors.danger, fontSize: 12),
      ),
    );

    final tiles = [
      for (final method in methods)
        _MethodTile(
          key: ValueKey('payment_method_${method.method}'),
          text: '${method.method} ${method.isCash ? '(cash)' : '(non-cash)'}',
          active: selected == method.method,
          onTap: () => onSelect(method.method),
          compact: horizontal,
        ),
    ];

    if (horizontal) {
      return Container(
        height: 98,
        decoration: const BoxDecoration(
          color: Colors.white,
          border: Border(bottom: BorderSide(color: MokposColors.line)),
        ),
        child: Column(
          children: [
            header,
            Expanded(
              child: methods.isEmpty
                  ? empty
                  : ListView.separated(
                      scrollDirection: Axis.horizontal,
                      padding: const EdgeInsets.symmetric(horizontal: 8),
                      itemBuilder: (context, index) => SizedBox(
                        width: 156,
                        child: tiles[index],
                      ),
                      separatorBuilder: (context, index) =>
                          const SizedBox(width: 8),
                      itemCount: tiles.length,
                    ),
            ),
          ],
        ),
      );
    }

    return Container(
      width: 174,
      decoration: const BoxDecoration(
        border: Border(right: BorderSide(color: MokposColors.line)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [header, if (methods.isEmpty) empty, ...tiles],
      ),
    );
  }
}

class _MethodTile extends StatelessWidget {
  const _MethodTile({
    required this.text,
    required this.active,
    required this.onTap,
    this.compact = false,
    super.key,
  });

  final String text;
  final bool active;
  final VoidCallback onTap;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Container(
        height: compact ? 44 : 42,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        decoration: BoxDecoration(
          color: active ? MokposColors.primarySoft : Colors.white,
          borderRadius: compact ? BorderRadius.circular(NojposRadius.md) : null,
          border: compact
              ? Border.all(
                  color: active ? NojposColors.primary : NojposColors.line,
                )
              : Border(
                  bottom: const BorderSide(color: MokposColors.line),
                  left: BorderSide(
                    color: active ? MokposColors.primary : Colors.transparent,
                    width: 3,
                  ),
                ),
        ),
        alignment: Alignment.centerLeft,
        child: Row(
          children: [
            Container(
              width: 28,
              height: 28,
              decoration: BoxDecoration(
                color: active ? Colors.white : MokposColors.canvas,
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: MokposColors.line),
              ),
              child: Center(
                child: NojposAssetIcon.named(
                  _paymentMethodAssetFor(text),
                  fallbackIcon: _paymentMethodIconFor(text),
                  size: 18,
                  applyColor: false,
                  semanticLabel: text,
                ),
              ),
            ),
            const SizedBox(width: 9),
            Expanded(
              child: Text(
                text,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  color: active ? MokposColors.text : MokposColors.muted,
                  fontSize: 12,
                  fontWeight: active ? FontWeight.w900 : FontWeight.w700,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

String _paymentMethodAssetFor(String value) {
  final method = value.toLowerCase();
  if (method.contains('qris') || method.contains('qr')) return 'qris';
  if (method.contains('edc') ||
      method.contains('card') ||
      method.contains('kartu')) {
    return 'edc';
  }
  if (method.contains('cash') || method.contains('tunai')) return 'cash';
  return 'receipt';
}

IconData _paymentMethodIconFor(String value) {
  final method = value.toLowerCase();
  if (method.contains('qris') || method.contains('qr')) {
    return LucideIcons.qrCode;
  }
  if (method.contains('edc') ||
      method.contains('card') ||
      method.contains('kartu')) {
    return LucideIcons.creditCard;
  }
  if (method.contains('cash') || method.contains('tunai')) {
    return LucideIcons.banknote;
  }
  return LucideIcons.receiptText;
}

class _PaymentAmountPanel extends StatelessWidget {
  const _PaymentAmountPanel({
    required this.method,
    required this.total,
    required this.paidAmount,
    required this.referenceController,
    required this.referenceFocusNode,
    required this.onSelectAmount,
    required this.onOpenAmountDialog,
  });

  final PaymentMethodConfig method;
  final int total;
  final int paidAmount;
  final TextEditingController referenceController;
  final FocusNode referenceFocusNode;
  final ValueChanged<int> onSelectAmount;
  final VoidCallback onOpenAmountDialog;

  @override
  Widget build(BuildContext context) {
    final cashOptions = [total, 20000, 50000];

    if (!method.isCash) {
      return Padding(
        padding: const EdgeInsets.all(28),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _SelectedPaymentChip(method: method.method, amount: total),
            const SizedBox(height: 24),
            const Text(
              'Non-tunai membutuhkan reference manual dan konfirmasi diterima.',
              style: TextStyle(color: MokposColors.muted),
            ),
            const SizedBox(height: 16),
            TextField(
              key: const ValueKey('payment_reference'),
              controller: referenceController,
              focusNode: referenceFocusNode,
              autofocus: true,
              textInputAction: TextInputAction.done,
              decoration: const InputDecoration(
                labelText: 'Reference manual',
                hintText: 'Nomor referensi QRIS/transfer/e-wallet',
                border: OutlineInputBorder(),
              ),
            ),
          ],
        ),
      );
    }

    return Padding(
      padding: const EdgeInsets.all(14),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (paidAmount > 0) ...[
            _SelectedPaymentChip(method: method.method, amount: paidAmount),
            const SizedBox(height: 14),
          ],
          Wrap(
            spacing: 10,
            runSpacing: 10,
            children: [
              for (final amount in cashOptions)
                _AmountOption(
                  key: ValueKey(
                    amount == total
                        ? 'payment_amount_exact'
                        : 'payment_amount_$amount',
                  ),
                  text: amount == total ? 'Uang Pas' : rupiah(amount),
                  active: paidAmount == amount,
                  onTap: () => onSelectAmount(amount),
                ),
              _AmountOption(
                key: const ValueKey('payment_amount_custom'),
                text: 'Lainnya',
                onTap: onOpenAmountDialog,
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _SelectedPaymentChip extends StatelessWidget {
  const _SelectedPaymentChip({required this.method, required this.amount});

  final String method;
  final int amount;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            '$method   ${rupiah(amount)}',
            style: const TextStyle(
              color: MokposColors.text,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(width: 8),
          const Icon(LucideIcons.circleX, size: 16, color: MokposColors.danger),
        ],
      ),
    );
  }
}

class _AmountOption extends StatelessWidget {
  const _AmountOption({
    required this.text,
    required this.onTap,
    this.active = false,
    super.key,
  });

  final String text;
  final VoidCallback onTap;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 168,
      height: 52,
      child: OutlinedButton(
        onPressed: onTap,
        style: OutlinedButton.styleFrom(
          alignment: Alignment.centerLeft,
          foregroundColor: MokposColors.text,
          backgroundColor: active ? MokposColors.primarySoft : Colors.white,
          side: BorderSide(
            color: active ? MokposColors.primary : MokposColors.line,
          ),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(MokposRadius.sm),
          ),
        ),
        child: Text(
          text,
          style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w800),
        ),
      ),
    );
  }
}

class _PaymentOrderSummary extends StatelessWidget {
  const _PaymentOrderSummary({
    required this.items,
    required this.total,
    required this.itemDiscountTotal,
    required this.cartDiscount,
    required this.paymentLines,
    required this.quote,
    required this.quoteLoading,
    required this.quoteError,
    required this.promotionCodes,
    required this.promotionCodeController,
    required this.onApplyPromotionCode,
    required this.onRemovePromotionCode,
    required this.canPay,
    required this.onPay,
  });

  final List<CartItem> items;
  final int total;
  final int itemDiscountTotal;
  final int cartDiscount;
  final List<CheckoutPayment> paymentLines;
  final CheckoutQuote? quote;
  final bool quoteLoading;
  final Object? quoteError;
  final List<String> promotionCodes;
  final TextEditingController promotionCodeController;
  final VoidCallback onApplyPromotionCode;
  final ValueChanged<String> onRemovePromotionCode;
  final bool canPay;
  final VoidCallback onPay;

  @override
  Widget build(BuildContext context) {
    final itemCount = items.fold(0, (sum, item) => sum + item.quantity);
    final quoteLabel = quote == null
        ? 'Menunggu total server'
        : 'Quote ${quote!.quoteId}';

    return Container(
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(left: BorderSide(color: MokposColors.line)),
      ),
      child: Column(
        children: [
          Container(
            height: 54,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: MokposColors.line)),
            ),
            child: Row(
              children: [
                const Icon(
                  LucideIcons.circleUserRound,
                  size: 18,
                  color: MokposColors.muted,
                ),
                const SizedBox(width: 8),
                const Text(
                  'Tanpa Pelanggan',
                  style: TextStyle(
                    color: MokposColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const Spacer(),
                Text(
                  quoteLabel,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(color: MokposColors.muted, fontSize: 11),
                ),
              ],
            ),
          ),
          Expanded(
            child: ListView.separated(
              itemCount: items.length,
              separatorBuilder: (context, index) =>
                  const Divider(height: 1, color: MokposColors.line),
              itemBuilder: (context, index) {
                final item = items[index];
                return SizedBox(
                  height: 44,
                  child: Row(
                    children: [
                      SizedBox(
                        width: 42,
                        child: Text(
                          '${item.quantity}',
                          textAlign: TextAlign.center,
                          style: const TextStyle(
                            color: MokposColors.text,
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ),
                      Expanded(
                        child: Text(
                          item.product.name,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            color: MokposColors.text,
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ),
                      Padding(
                        padding: const EdgeInsets.only(right: 12),
                        child: Text(
                          rupiah(item.total),
                          style: const TextStyle(
                            color: MokposColors.text,
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),
          Container(
            height: paymentLines.isEmpty ? 290 : 330,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            decoration: const BoxDecoration(
              color: MokposColors.canvas,
              border: Border(top: BorderSide(color: MokposColors.line)),
            ),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                _SummaryLine(label: 'Diskon item', value: itemDiscountTotal),
                _SummaryLine(label: 'Diskon cart', value: cartDiscount),
                PromotionQuotePanel(
                  quote: quote,
                  promotionCodes: promotionCodes,
                  controller: promotionCodeController,
                  onApply: onApplyPromotionCode,
                  onRemove: onRemovePromotionCode,
                ),
                if (quoteLoading)
                  const _SummaryMessage('Mengambil total server...')
                else if (quoteError != null)
                  const _SummaryMessage('Gagal mengambil total server')
                else ...[
                  _SummaryLine(
                    label: 'Diskon promo',
                    value: quote?.promotionDiscountTotal ?? 0,
                  ),
                  _SummaryLine(
                    label: 'Service charge',
                    value: quote?.serviceChargeTotal ?? 0,
                  ),
                  _SummaryLine(label: 'Pajak', value: quote?.taxTotal ?? 0),
                  _SummaryLine(
                    label: 'Rounding',
                    value: quote?.roundingTotal ?? 0,
                  ),
                ],
                for (final line in paymentLines)
                  _SummaryLine(
                    label: '${line.method}${line.isCash ? '' : ' pending'}',
                    value: line.amount,
                  ),
                _SummaryLine(
                  label: 'Grand total $itemCount Produk',
                  value: total,
                  emphasized: true,
                ),
              ],
            ),
          ),
          InkWell(
            key: const ValueKey('process_payment_button'),
            onTap: canPay ? onPay : null,
            child: Container(
              height: 58,
              color: canPay ? MokposColors.primary : MokposColors.line,
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Row(
                children: [
                  Text(
                    'Proses Bayar',
                    style: TextStyle(
                      color: canPay ? Colors.white : MokposColors.muted,
                      fontSize: 15,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const Spacer(),
                  Icon(
                    LucideIcons.chevronRight,
                    color: canPay ? Colors.white : MokposColors.muted,
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

class _SummaryMessage extends StatelessWidget {
  const _SummaryMessage(this.message);

  final String message;

  @override
  Widget build(BuildContext context) {
    return Align(
      alignment: Alignment.centerLeft,
      child: Text(
        message,
        style: const TextStyle(
          color: MokposColors.muted,
          fontWeight: FontWeight.w700,
          fontSize: 12,
        ),
      ),
    );
  }
}

class _SummaryLine extends StatelessWidget {
  const _SummaryLine({
    required this.label,
    required this.value,
    this.emphasized = false,
  });

  final String label;
  final int value;
  final bool emphasized;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Text(
          label,
          style: TextStyle(
            color: emphasized ? MokposColors.text : MokposColors.muted,
            fontWeight: emphasized ? FontWeight.w900 : FontWeight.w700,
            fontSize: 12,
          ),
        ),
        const Spacer(),
        Text(
          rupiah(value),
          style: TextStyle(
            color: emphasized ? MokposColors.text : MokposColors.muted,
            fontWeight: FontWeight.w900,
            fontSize: 12,
          ),
        ),
      ],
    );
  }
}

class AmountDialog extends StatefulWidget {
  const AmountDialog({
    required this.total,
    required this.initialAmount,
    super.key,
  });

  final int total;
  final int initialAmount;

  @override
  State<AmountDialog> createState() => _AmountDialogState();
}

Future<List<CheckoutPayment>?> showSplitPaymentDialog(
  BuildContext context, {
  required List<PaymentMethodConfig> methods,
  required int total,
  required List<CheckoutPayment> existing,
}) {
  return showDialog<List<CheckoutPayment>>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) =>
        _SplitPaymentDialog(methods: methods, total: total, existing: existing),
  );
}

class _SplitPaymentDialog extends StatefulWidget {
  const _SplitPaymentDialog({
    required this.methods,
    required this.total,
    required this.existing,
  });

  final List<PaymentMethodConfig> methods;
  final int total;
  final List<CheckoutPayment> existing;

  @override
  State<_SplitPaymentDialog> createState() => _SplitPaymentDialogState();
}

class _SplitPaymentDialogState extends State<_SplitPaymentDialog> {
  late final lines = [...widget.existing];
  late String method = widget.methods.firstOrNull?.method ?? '';
  final amountController = TextEditingController();
  final referenceController = TextEditingController();
  final amountFocusNode = FocusNode();
  final referenceFocusNode = FocusNode();
  String? lineError;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback(
      (_) => amountFocusNode.requestFocus(),
    );
  }

  @override
  void dispose() {
    amountController.dispose();
    referenceController.dispose();
    amountFocusNode.dispose();
    referenceFocusNode.dispose();
    super.dispose();
  }

  void _addLine() {
    final config = _methodByName(widget.methods, method);
    final amount =
        int.tryParse(amountController.text.replaceAll(RegExp(r'[^0-9]'), '')) ??
        0;
    if (amount <= 0 || config.method.isEmpty) return;
    final reference = referenceController.text.trim();
    if (!config.isCash && reference.isEmpty) {
      setState(() => lineError = 'Reference non-tunai wajib diisi.');
      return;
    }
    setState(() {
      lineError = null;
      lines.add(
        CheckoutPayment(
          method: config.method,
          amount: amount,
          isCash: config.isCash,
          reference: config.isCash ? null : reference,
        ),
      );
      amountController.clear();
      referenceController.clear();
    });
    amountFocusNode.requestFocus();
  }

  @override
  Widget build(BuildContext context) {
    final paid = lines.fold(0, (sum, line) => sum + line.amount);
    final remaining = (widget.total - paid).clamp(0, widget.total);
    final selected = _methodByName(widget.methods, method);
    return AlertDialog(
      backgroundColor: Colors.white,
      insetPadding: const EdgeInsets.symmetric(horizontal: 18, vertical: 24),
      title: const Text('Pisah Bayar'),
      content: ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: 500,
          maxHeight: MediaQuery.sizeOf(context).height * .68,
        ),
        child: SingleChildScrollView(
          keyboardDismissBehavior: ScrollViewKeyboardDismissBehavior.onDrag,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              DropdownButtonFormField<String>(
                initialValue: method.isEmpty ? null : method,
                items: [
                  for (final config in widget.methods)
                    DropdownMenuItem(
                      value: config.method,
                      child: Text(
                        '${config.method} ${config.isCash ? '(cash)' : '(non-cash)'}',
                      ),
                    ),
                ],
                onChanged: (value) {
                  setState(() => method = value ?? method);
                  final config = _methodByName(widget.methods, value ?? method);
                  WidgetsBinding.instance.addPostFrameCallback((_) {
                    if (config.isCash) {
                      amountFocusNode.requestFocus();
                    } else {
                      referenceFocusNode.requestFocus();
                    }
                  });
                },
                decoration: const InputDecoration(
                  labelText: 'Metode',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: amountController,
                focusNode: amountFocusNode,
                autofocus: true,
                keyboardType: TextInputType.number,
                textInputAction: selected.isCash
                    ? TextInputAction.done
                    : TextInputAction.next,
                onSubmitted: (_) {
                  if (selected.isCash) {
                    _addLine();
                  } else {
                    referenceFocusNode.requestFocus();
                  }
                },
                decoration: InputDecoration(
                  labelText: 'Nominal',
                  helperText: 'Sisa tagihan ${rupiah(remaining)}',
                  prefixText: 'Rp ',
                  border: const OutlineInputBorder(),
                ),
              ),
              if (!selected.isCash) ...[
                const SizedBox(height: 12),
                TextField(
                  controller: referenceController,
                  focusNode: referenceFocusNode,
                  textInputAction: TextInputAction.done,
                  onSubmitted: (_) => _addLine(),
                  decoration: const InputDecoration(
                    labelText: 'Reference manual',
                    border: OutlineInputBorder(),
                  ),
                ),
              ],
              if (lineError != null) ...[
                const SizedBox(height: 8),
                Align(
                  alignment: Alignment.centerLeft,
                  child: Text(
                    lineError!,
                    style: const TextStyle(
                      color: MokposColors.danger,
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
              ],
              const SizedBox(height: 12),
              Align(
                alignment: Alignment.centerRight,
                child: OutlinedButton.icon(
                  onPressed: _addLine,
                  icon: const Icon(LucideIcons.plus, size: 16),
                  label: const Text('Tambah Line'),
                ),
              ),
              const Divider(),
              for (final (index, line) in lines.indexed)
                ListTile(
                  dense: true,
                  title: Text(line.method),
                  subtitle: Text(
                    line.reference ?? (line.isCash ? 'cash' : 'pending'),
                  ),
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(rupiah(line.amount)),
                      IconButton(
                        onPressed: () => setState(() => lines.removeAt(index)),
                        icon: const Icon(LucideIcons.x, size: 16),
                      ),
                    ],
                  ),
                ),
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
          onPressed: lines.isEmpty
              ? null
              : () => Navigator.of(context).pop([...lines]),
          child: const Text('Simpan Split'),
        ),
      ],
    );
  }
}

class _AmountDialogState extends State<AmountDialog> {
  late int amount = widget.initialAmount;

  @override
  Widget build(BuildContext context) {
    final amounts = [
      100000,
      50000,
      20000,
      10000,
      5000,
      2000,
      1000,
      500,
      200,
      100,
    ];

    return Dialog(
      insetPadding: const EdgeInsets.symmetric(horizontal: 110, vertical: 34),
      backgroundColor: Colors.white,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      child: Column(
        children: [
          Container(
            height: 58,
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: MokposColors.line)),
            ),
            child: Row(
              children: [
                IconButton(
                  onPressed: () => Navigator.of(context).pop(),
                  icon: const Icon(LucideIcons.x),
                ),
                Text(
                  'Total Tagihan ${rupiah(widget.total)}',
                  style: const TextStyle(
                    color: MokposColors.text,
                    fontWeight: FontWeight.w900,
                    fontSize: 16,
                  ),
                ),
                const Spacer(),
                InkWell(
                  onTap: () => Navigator.of(context).pop(amount),
                  child: Container(
                    width: 96,
                    height: 58,
                    color: MokposColors.primary,
                    alignment: Alignment.center,
                    child: const Text(
                      'Simpan',
                      style: TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 18),
          const Text(
            'Total dibayar',
            style: TextStyle(
              color: MokposColors.muted,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            rupiah(amount),
            style: const TextStyle(
              color: MokposColors.text,
              fontSize: 34,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 24),
          Expanded(
            child: GridView.builder(
              padding: const EdgeInsets.fromLTRB(18, 0, 18, 22),
              itemCount: amounts.length + 2,
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 3,
                mainAxisExtent: 54,
                mainAxisSpacing: 12,
                crossAxisSpacing: 12,
              ),
              itemBuilder: (context, index) {
                if (index < amounts.length) {
                  final value = amounts[index];
                  return _DialogAmountButton(
                    text: rupiah(value),
                    active: amount == value,
                    onTap: () => setState(() => amount = value),
                  );
                }
                if (index == amounts.length) {
                  return _DialogAmountButton(
                    text: 'Uang Pas',
                    active: amount == widget.total,
                    onTap: () => setState(() => amount = widget.total),
                  );
                }
                return _DialogAmountButton(
                  text: 'Clear',
                  danger: true,
                  onTap: () => setState(() => amount = 0),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _DialogAmountButton extends StatelessWidget {
  const _DialogAmountButton({
    required this.text,
    required this.onTap,
    this.active = false,
    this.danger = false,
  });

  final String text;
  final VoidCallback onTap;
  final bool active;
  final bool danger;

  @override
  Widget build(BuildContext context) {
    return OutlinedButton(
      onPressed: onTap,
      style: OutlinedButton.styleFrom(
        alignment: Alignment.centerLeft,
        foregroundColor: danger ? Colors.white : MokposColors.text,
        backgroundColor: danger
            ? MokposColors.danger
            : active
            ? MokposColors.primarySoft
            : Colors.white,
        side: BorderSide(
          color: active ? MokposColors.primary : MokposColors.line,
        ),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(MokposRadius.sm),
        ),
      ),
      child: Text(
        text,
        style: const TextStyle(fontWeight: FontWeight.w900, fontSize: 13),
      ),
    );
  }
}
