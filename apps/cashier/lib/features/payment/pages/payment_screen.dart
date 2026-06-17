import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../../../shared/models/nojpos_models.dart';
import '../../pos/formatters.dart';
import '../../pos/models/cart_item.dart';
import '../../pos/providers/pos_providers.dart';

class PaymentScreen extends ConsumerStatefulWidget {
  const PaymentScreen({super.key});

  @override
  ConsumerState<PaymentScreen> createState() => _PaymentScreenState();
}

class _PaymentScreenState extends ConsumerState<PaymentScreen> {
  String selectedMethod = 'Tunai';
  int paidAmount = 0;

  @override
  Widget build(BuildContext context) {
    final items = ref.watch(cartProvider);
    final total = ref.watch(cartTotalProvider);
    final remaining = (total - paidAmount).clamp(0, total);
    final change = (paidAmount - total).clamp(0, paidAmount);

    return Scaffold(
      backgroundColor: Colors.white,
      body: SafeArea(
        child: Column(
          children: [
            _PaymentTopBar(onBack: () => context.go('/pos')),
            Expanded(
              child: Row(
                children: [
                  Expanded(
                    child: Column(
                      children: [
                        _PaymentStats(
                          total: total,
                          remaining: remaining,
                          change: change,
                        ),
                        const _PaymentActions(),
                        Expanded(
                          child: Row(
                            children: [
                              _MethodRail(
                                selected: selectedMethod,
                                onSelect: (method) {
                                  setState(() {
                                    selectedMethod = method;
                                    paidAmount = method == 'Tunai' ? 0 : total;
                                  });
                                },
                              ),
                              Expanded(
                                child: _PaymentAmountPanel(
                                  method: selectedMethod,
                                  total: total,
                                  paidAmount: paidAmount,
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
                    ),
                  ),
                  SizedBox(
                    width: 430,
                    child: _PaymentOrderSummary(
                      items: items,
                      total: total,
                      canPay: items.isNotEmpty && remaining == 0,
                      onPay: () {
                        ref
                            .read(nojposSessionProvider.notifier)
                            .completeTransaction(
                              cartItems: items,
                              method: _paymentMethodFromLabel(selectedMethod),
                              paidAmount: paidAmount,
                            );
                        context.go('/success');
                      },
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

PaymentMethod _paymentMethodFromLabel(String label) {
  return switch (label) {
    'Nontunai' => PaymentMethod.cashless,
    'Transfer' => PaymentMethod.transfer,
    'QRIS' => PaymentMethod.qris,
    'Komplimen' => PaymentMethod.compliment,
    'Deposit' => PaymentMethod.deposit,
    _ => PaymentMethod.cash,
  };
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
              Row(
                children: [
                  _TinyDot(),
                  SizedBox(width: 5),
                  Text(
                    'Status: Online',
                    style: TextStyle(color: Color(0xE6FFFFFF), fontSize: 11),
                  ),
                ],
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

class _TinyDot extends StatelessWidget {
  const _TinyDot();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 8,
      height: 8,
      decoration: const BoxDecoration(
        color: Color(0xFF70E06D),
        shape: BoxShape.circle,
      ),
    );
  }
}

class _PaymentStats extends StatelessWidget {
  const _PaymentStats({
    required this.total,
    required this.remaining,
    required this.change,
  });

  final int total;
  final int remaining;
  final int change;

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
              value: rupiah(change),
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

class _PaymentActions extends StatelessWidget {
  const _PaymentActions();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 54,
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: const Row(
        children: [
          Expanded(
            child: _PaymentAction(
              icon: LucideIcons.creditCard,
              label: 'Pisah Bayar',
            ),
          ),
          VerticalDivider(width: 1, color: MokposColors.line),
          Expanded(
            child: _PaymentAction(
              icon: LucideIcons.fileText,
              label: 'Jadikan Invoice',
            ),
          ),
          VerticalDivider(width: 1, color: MokposColors.line),
          Expanded(
            child: _PaymentAction(icon: LucideIcons.pencil, label: 'Catatan'),
          ),
        ],
      ),
    );
  }
}

class _PaymentAction extends StatelessWidget {
  const _PaymentAction({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () {
        if (label == 'Catatan') {
          showDialog<void>(
            context: context,
            builder: (context) => AlertDialog(
              backgroundColor: Colors.white,
              title: const Text('Catatan Order'),
              content: const TextField(
                autofocus: true,
                maxLines: 3,
                decoration: InputDecoration(
                  hintText: 'Tulis catatan transaksi',
                  border: OutlineInputBorder(),
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text('Batal'),
                ),
                FilledButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text('Simpan'),
                ),
              ],
            ),
          );
          return;
        }
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('$label siap disambungkan ke backend')),
        );
      },
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
  const _MethodRail({required this.selected, required this.onSelect});

  final String selected;
  final ValueChanged<String> onSelect;

  @override
  Widget build(BuildContext context) {
    final methods = [
      'Tunai',
      'Nontunai',
      'Transfer',
      'QRIS',
      'Komplimen',
      'Deposit',
    ];
    return Container(
      width: 174,
      decoration: const BoxDecoration(
        border: Border(right: BorderSide(color: MokposColors.line)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Container(
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
                Icon(
                  LucideIcons.listFilter,
                  size: 16,
                  color: MokposColors.text,
                ),
              ],
            ),
          ),
          for (final method in methods)
            _MethodTile(
              text: method,
              active: selected == method,
              onTap: () => onSelect(method),
            ),
        ],
      ),
    );
  }
}

class _MethodTile extends StatelessWidget {
  const _MethodTile({
    required this.text,
    required this.active,
    required this.onTap,
  });

  final String text;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Container(
        height: 42,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        decoration: BoxDecoration(
          color: active ? MokposColors.primarySoft : Colors.white,
          border: Border(
            bottom: const BorderSide(color: MokposColors.line),
            left: BorderSide(
              color: active ? MokposColors.primary : Colors.transparent,
              width: 3,
            ),
          ),
        ),
        alignment: Alignment.centerLeft,
        child: Text(
          text,
          style: TextStyle(
            color: active ? MokposColors.text : MokposColors.muted,
            fontSize: 12,
            fontWeight: active ? FontWeight.w900 : FontWeight.w700,
          ),
        ),
      ),
    );
  }
}

class _PaymentAmountPanel extends StatelessWidget {
  const _PaymentAmountPanel({
    required this.method,
    required this.total,
    required this.paidAmount,
    required this.onSelectAmount,
    required this.onOpenAmountDialog,
  });

  final String method;
  final int total;
  final int paidAmount;
  final ValueChanged<int> onSelectAmount;
  final VoidCallback onOpenAmountDialog;

  @override
  Widget build(BuildContext context) {
    final cashOptions = [total, 20000, 50000];

    if (method != 'Tunai') {
      return Padding(
        padding: const EdgeInsets.all(28),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _SelectedPaymentChip(method: method, amount: total),
            const SizedBox(height: 24),
            const Text(
              'Metode ini akan mencatat pembayaran penuh secara otomatis.',
              style: TextStyle(color: MokposColors.muted),
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
            _SelectedPaymentChip(method: method, amount: paidAmount),
            const SizedBox(height: 14),
          ],
          Wrap(
            spacing: 10,
            runSpacing: 10,
            children: [
              for (final amount in cashOptions)
                _AmountOption(
                  text: amount == total ? 'Uang Pas' : rupiah(amount),
                  active: paidAmount == amount,
                  onTap: () => onSelectAmount(amount),
                ),
              _AmountOption(text: 'Lainnya', onTap: onOpenAmountDialog),
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
    required this.canPay,
    required this.onPay,
  });

  final List<CartItem> items;
  final int total;
  final bool canPay;
  final VoidCallback onPay;

  @override
  Widget build(BuildContext context) {
    final itemCount = items.fold(0, (sum, item) => sum + item.quantity);

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
            child: const Row(
              children: [
                Icon(
                  LucideIcons.circleUserRound,
                  size: 18,
                  color: MokposColors.muted,
                ),
                SizedBox(width: 8),
                Text(
                  'Tanpa Pelanggan',
                  style: TextStyle(
                    color: MokposColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                Spacer(),
                Text(
                  'CS/01/260616/0003',
                  style: TextStyle(color: MokposColors.muted, fontSize: 11),
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
                          rupiah(item.subtotal),
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
            height: 44,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            decoration: const BoxDecoration(
              color: MokposColors.canvas,
              border: Border(top: BorderSide(color: MokposColors.line)),
            ),
            child: Row(
              children: [
                Text(
                  'Total $itemCount Produk',
                  style: const TextStyle(
                    color: MokposColors.text,
                    fontWeight: FontWeight.w900,
                    fontSize: 13,
                  ),
                ),
                const Spacer(),
                Text(
                  rupiah(total),
                  style: const TextStyle(
                    color: MokposColors.text,
                    fontWeight: FontWeight.w900,
                    fontSize: 13,
                  ),
                ),
              ],
            ),
          ),
          InkWell(
            onTap: canPay ? onPay : null,
            child: Container(
              height: 58,
              color: canPay ? MokposColors.primary : const Color(0xFFD8DDDD),
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
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(2)),
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
