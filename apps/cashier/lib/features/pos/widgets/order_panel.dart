import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../formatters.dart';
import '../models/cart_item.dart';
import '../providers/pos_providers.dart';

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
    final selectedCustomer = ref.watch(
      nojposSessionProvider.select((session) => session.selectedCustomer),
    );

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
              rupiah(item.subtotal),
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
    required this.itemCount,
    required this.orderType,
  });

  final int total;
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

    void saveOrder() {
      final items = ref.read(cartProvider);
      if (items.isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Pilih produk terlebih dahulu')),
        );
        return;
      }
      final order = ref
          .read(nojposSessionProvider.notifier)
          .saveOrder(cartItems: items);
      ref.read(cartProvider.notifier).clear();
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text('${order.number} disimpan')));
    }

    void showDiscountPlaceholder() {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Diskon siap dihubungkan ke backend')),
      );
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
                      ? () => ref.read(cartProvider.notifier).clear()
                      : null,
                ),
                const VerticalDivider(width: 1, color: MokposColors.line),
                _FooterTool(
                  icon: LucideIcons.badgePercent,
                  onTap: showDiscountPlaceholder,
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
          GestureDetector(
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
  const _FooterTool({required this.icon, this.label, this.onTap});

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
