import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../formatters.dart';
import '../models/cart_item.dart';
import '../providers/parked_order_lease_controller.dart';
import '../providers/pos_providers.dart';
import 'pos_dialogs.dart';

class OrderPanel extends ConsumerWidget {
  const OrderPanel({
    required this.orderType,
    required this.onSelectOrderType,
    required this.onSelectCustomer,
    super.key,
  });

  final String orderType;
  final VoidCallback onSelectOrderType;
  final VoidCallback onSelectCustomer;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final items = ref.watch(cartProvider);
    final total = ref.watch(cartTotalProvider);
    final itemDiscountTotal = ref.watch(itemDiscountTotalProvider);
    final cartDiscount = ref.watch(checkoutDetailsProvider).cartDiscount;
    final selectedCustomer = ref.watch(
      nojposSessionProvider.select((session) => session.selectedCustomer),
    );
    final leaseState = ref.watch(parkedOrderLeaseControllerProvider);

    return Container(
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(left: BorderSide(color: MokposColors.line)),
      ),
      child: Column(
        children: [
          _OrderHeader(
            orderType: orderType,
            customerLabel: selectedCustomer?.name ?? 'Pelanggan',
            onSelectOrderType: onSelectOrderType,
            onSelectCustomer: onSelectCustomer,
          ),
          if (leaseState.active != null || leaseState.errorMessage != null)
            _ParkedOrderLeaseBanner(state: leaseState),
          Expanded(
            child: items.isEmpty
                ? const _EmptyOrder()
                : ListView.separated(
                    padding: EdgeInsets.zero,
                    itemCount: items.length,
                    separatorBuilder: (context, index) =>
                        const Divider(height: 1, color: MokposColors.line),
                    itemBuilder: (context, index) =>
                        _CartItemTile(item: items[index]),
                  ),
          ),
          _OrderFooter(
            total: total,
            itemDiscountTotal: itemDiscountTotal,
            cartDiscount: cartDiscount,
            itemCount: items.fold(0, (sum, item) => sum + item.quantity),
            orderType: orderType,
          ),
        ],
      ),
    );
  }
}

class _OrderHeader extends StatelessWidget {
  const _OrderHeader({
    required this.orderType,
    required this.customerLabel,
    required this.onSelectOrderType,
    required this.onSelectCustomer,
  });

  final String orderType;
  final String customerLabel;
  final VoidCallback onSelectOrderType;
  final VoidCallback onSelectCustomer;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 58,
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: Row(
        children: [
          Expanded(
            child: _HeaderAction(
              icon: LucideIcons.utensils,
              label: orderType,
              trailing: LucideIcons.chevronDown,
              onTap: onSelectOrderType,
            ),
          ),
          const VerticalDivider(width: 1, color: MokposColors.line),
          Expanded(
            child: _HeaderAction(
              icon: LucideIcons.userPlus,
              label: customerLabel,
              accent: true,
              onTap: onSelectCustomer,
            ),
          ),
        ],
      ),
    );
  }
}

class _HeaderAction extends StatelessWidget {
  const _HeaderAction({
    required this.icon,
    required this.label,
    this.trailing,
    this.accent = false,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final IconData? trailing;
  final bool accent;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              icon,
              size: 20,
              color: accent ? MokposColors.primary : MokposColors.text,
            ),
            const SizedBox(width: 8),
            Flexible(
              child: Text(
                label,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  color: accent ? MokposColors.primary : MokposColors.text,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            if (trailing != null) ...[
              const SizedBox(width: 6),
              Icon(trailing, size: 18, color: MokposColors.primary),
            ],
          ],
        ),
      ),
    );
  }
}

class _EmptyOrder extends StatelessWidget {
  const _EmptyOrder();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(30),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text(
              'Silakan masukkan pesanan dari pelanggan',
              textAlign: TextAlign.center,
              style: TextStyle(
                color: MokposColors.muted,
                fontSize: 15,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ParkedOrderLeaseBanner extends StatelessWidget {
  const _ParkedOrderLeaseBanner({required this.state});

  final ParkedOrderLeaseState state;

  @override
  Widget build(BuildContext context) {
    final active = state.active;
    final message =
        state.errorMessage ??
        (active == null
            ? ''
            : 'Mengedit ${active.orderNumber} · rev ${active.revision}');
    if (message.isEmpty) return const SizedBox.shrink();
    final isError = state.errorMessage != null;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      color: isError ? MokposColors.danger.withValues(alpha: 0.08) : null,
      child: Row(
        children: [
          Icon(
            isError ? LucideIcons.triangleAlert : LucideIcons.lockKeyhole,
            size: 15,
            color: isError ? MokposColors.danger : MokposColors.primary,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              message,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: isError ? MokposColors.danger : MokposColors.text,
                fontSize: 12,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _CartItemTile extends ConsumerWidget {
  const _CartItemTile({required this.item});

  final CartItem item;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final notifier = ref.read(cartProvider.notifier);
    return SizedBox(
      height: 50,
      child: Row(
        children: [
          SizedBox(
            width: 42,
            child: Text(
              '${item.quantity}',
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w800,
                fontSize: 13,
              ),
            ),
          ),
          Expanded(
            child: Text(
              item.product.name,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w800,
                fontSize: 13,
              ),
            ),
          ),
          SizedBox(
            width: 96,
            child: Text(
              rupiah(item.total),
              textAlign: TextAlign.right,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w800,
                fontSize: 13,
              ),
            ),
          ),
          IconButton(
            key: ValueKey('item_discount_${item.product.id}'),
            tooltip: 'Diskon item',
            onPressed: () => showItemDiscountDialog(context, ref, item),
            icon: const Icon(LucideIcons.badgePercent, size: 15),
            color: item.discount > 0
                ? MokposColors.warning
                : MokposColors.primary,
            style: IconButton.styleFrom(
              fixedSize: const Size(30, 30),
              minimumSize: const Size(30, 30),
              padding: EdgeInsets.zero,
            ),
          ),
          IconButton(
            tooltip: 'Kurangi item',
            onPressed: () => notifier.decrease(item.product),
            icon: const Icon(LucideIcons.minus, size: 15),
            color: MokposColors.primary,
            style: IconButton.styleFrom(
              fixedSize: const Size(30, 30),
              minimumSize: const Size(30, 30),
              padding: EdgeInsets.zero,
            ),
          ),
          IconButton(
            tooltip: 'Tambah item',
            onPressed: () => notifier.add(item.product),
            icon: const Icon(LucideIcons.plus, size: 15),
            color: MokposColors.primary,
            style: IconButton.styleFrom(
              fixedSize: const Size(30, 30),
              minimumSize: const Size(30, 30),
              padding: EdgeInsets.zero,
            ),
          ),
          IconButton(
            tooltip: 'Hapus item',
            onPressed: () => notifier.remove(item.product),
            icon: const Icon(LucideIcons.circleX, size: 16),
            color: MokposColors.danger,
            style: IconButton.styleFrom(
              fixedSize: const Size(32, 32),
              minimumSize: const Size(32, 32),
              padding: EdgeInsets.zero,
            ),
          ),
          const SizedBox(width: 4),
        ],
      ),
    );
  }
}

class _OrderFooter extends ConsumerWidget {
  const _OrderFooter({
    required this.total,
    required this.itemDiscountTotal,
    required this.cartDiscount,
    required this.itemCount,
    required this.orderType,
  });

  final int total;
  final int itemDiscountTotal;
  final int cartDiscount;
  final int itemCount;
  final String orderType;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    void openPayment() {
      if (itemCount <= 0) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Pilih produk terlebih dahulu')),
        );
        return;
      }
      context.go('/payment');
    }

    Future<void> saveOrder() async {
      final items = ref.read(cartProvider);
      if (items.isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Pilih produk terlebih dahulu')),
        );
        return;
      }
      final details = ref.read(checkoutDetailsProvider);
      try {
        final transaction = await ref
            .read(parkedOrderLeaseControllerProvider.notifier)
            .saveCurrentCartAsParked(
              cartItems: items,
              cartDiscount: details.cartDiscount,
              notes: details.notes,
              servedBy: details.servedBy,
            );
        if (!context.mounted || transaction == null) return;
        ref.read(cartProvider.notifier).clear();
        ref.read(checkoutDetailsProvider.notifier).clear();
        final number = transaction.number.isEmpty
            ? transaction.id
            : transaction.number;
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('$number disimpan')));
      } catch (error) {
        if (!context.mounted) return;
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(parkedOrderLeaseMessage(error))));
      }
    }

    return Container(
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(top: BorderSide(color: MokposColors.line)),
      ),
      child: Column(
        children: [
          SizedBox(
            height: 56,
            child: Row(
              children: [
                _FooterTool(
                  icon: LucideIcons.trash2,
                  onTap: itemCount > 0
                      ? () async {
                          await ref
                              .read(parkedOrderLeaseControllerProvider.notifier)
                              .releaseActive(restoreOrder: true);
                          ref.read(cartProvider.notifier).clear();
                          ref.read(checkoutDetailsProvider.notifier).clear();
                        }
                      : null,
                ),
                const VerticalDivider(width: 1, color: MokposColors.line),
                _FooterTool(
                  key: const ValueKey('cart_discount_button'),
                  icon: LucideIcons.badgePercent,
                  onTap: () => showCartDiscountDialog(context, ref),
                ),
                const VerticalDivider(width: 1, color: MokposColors.line),
                _FooterTool(
                  icon: LucideIcons.download,
                  label: 'Simpan',
                  onTap: saveOrder,
                ),
              ],
            ),
          ),
          if (itemDiscountTotal > 0 || cartDiscount > 0)
            Padding(
              padding: const EdgeInsets.fromLTRB(18, 0, 18, 10),
              child: Column(
                children: [
                  if (itemDiscountTotal > 0)
                    _DiscountLine(
                      label: 'Diskon item',
                      amount: itemDiscountTotal,
                    ),
                  if (cartDiscount > 0)
                    _DiscountLine(label: 'Diskon cart', amount: cartDiscount),
                ],
              ),
            ),
          GestureDetector(
            key: const ValueKey('open_payment_button'),
            behavior: HitTestBehavior.opaque,
            onTap: openPayment,
            child: Container(
              height: 70,
              color: MokposColors.primary,
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 18),
                child: Row(
                  children: [
                    Stack(
                      clipBehavior: Clip.none,
                      children: [
                        const Text(
                          'Bayar',
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 22,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                        Positioned(
                          left: -10,
                          top: -11,
                          child: Container(
                            width: 22,
                            height: 22,
                            decoration: const BoxDecoration(
                              color: MokposColors.warning,
                              shape: BoxShape.circle,
                            ),
                            child: Center(
                              child: Text(
                                '$itemCount',
                                style: const TextStyle(
                                  color: MokposColors.text,
                                  fontSize: 12,
                                  fontWeight: FontWeight.w900,
                                ),
                              ),
                            ),
                          ),
                        ),
                      ],
                    ),
                    const Spacer(),
                    Text(
                      rupiah(total),
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 21,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    const SizedBox(width: 8),
                    const Icon(
                      LucideIcons.chevronRight,
                      color: Colors.white,
                      size: 20,
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _FooterTool extends StatelessWidget {
  const _FooterTool({required this.icon, this.label, this.onTap, super.key});

  final IconData icon;
  final String? label;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: InkWell(
        onTap: onTap,
        child: Center(
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                icon,
                color: onTap == null && label == null
                    ? MokposColors.primary
                    : MokposColors.muted,
                size: 22,
              ),
              if (label != null) ...[
                const SizedBox(width: 8),
                Text(
                  label!,
                  style: const TextStyle(
                    color: MokposColors.muted,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _DiscountLine extends StatelessWidget {
  const _DiscountLine({required this.label, required this.amount});

  final String label;
  final int amount;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Text(
          label,
          style: const TextStyle(
            color: MokposColors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w700,
          ),
        ),
        const Spacer(),
        Text(
          '-${rupiah(amount)}',
          style: const TextStyle(
            color: MokposColors.danger,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}
