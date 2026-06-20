import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../../core/outbox/checkout_outbox.dart';
import '../../notifications/widgets/checkout_outbox_center.dart';

class PosTopBar extends ConsumerWidget {
  const PosTopBar({
    required this.onOpenMenu,
    required this.onOpenOrders,
    required this.onOpenMode,
    required this.onShowNotification,
    super.key,
  });

  final VoidCallback onOpenMenu;
  final VoidCallback onOpenOrders;
  final VoidCallback onOpenMode;
  final VoidCallback onShowNotification;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(nojposSessionProvider);
    final outboxItems = ref.watch(checkoutOutboxControllerProvider);
    final blockingOutboxCount = outboxItems
        .where((item) => item.blocksClose)
        .length;
    return Container(
      height: 64,
      color: MokposColors.primary,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      child: Row(
        children: [
          _TopIconButton(
            icon: LucideIcons.menu,
            label: 'Menu',
            onTap: onOpenMenu,
          ),
          const SizedBox(width: 10),
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(MokposRadius.md),
            ),
            child: const Icon(
              LucideIcons.store,
              color: MokposColors.primary,
              size: 25,
            ),
          ),
          const SizedBox(width: 12),
          Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                '${session.outlet.name} - ${session.cashier.name}',
                style: const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.w900,
                  fontSize: 16,
                ),
              ),
              const SizedBox(height: 2),
              Row(
                children: [
                  const _OnlineDot(),
                  const SizedBox(width: 6),
                  Text(
                    session.hasOpenShift
                        ? 'Status: Online · Shift Aktif'
                        : 'Status: Online · Shift belum dibuka',
                    style: const TextStyle(
                      color: MokposColors.onPrimaryMuted,
                      fontSize: 12,
                    ),
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
              fontSize: 26,
              fontWeight: FontWeight.w900,
            ),
          ),
          const Spacer(),
          _TopIconButton(
            icon: LucideIcons.bell,
            label: 'Notifikasi',
            badge: blockingOutboxCount,
            onTap: () => showCheckoutOutboxDialog(context, ref),
          ),
          _TopIconButton(
            icon: LucideIcons.layoutGrid,
            label: 'Mode order',
            onTap: onOpenMode,
          ),
          _TopIconButton(
            icon: LucideIcons.lockKeyhole,
            label: 'Lock: kembali ke PIN',
            onTap: () {
              ref.read(nojposSessionProvider.notifier).lock();
              context.go('/pin');
            },
          ),
          const SizedBox(width: 10),
          GestureDetector(
            key: const ValueKey('topbar_orders'),
            behavior: HitTestBehavior.opaque,
            onTap: onOpenOrders,
            child: Container(
              height: 48,
              width: 210,
              padding: const EdgeInsets.symmetric(horizontal: 18),
              decoration: BoxDecoration(
                color: MokposColors.primaryDark,
                borderRadius: BorderRadius.circular(MokposRadius.sm),
              ),
              child: const Row(
                children: [
                  Expanded(
                    child: Text(
                      'Daftar Order',
                      style: TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.w900,
                        fontSize: 16,
                      ),
                    ),
                  ),
                  Icon(LucideIcons.chevronRight, color: Colors.white, size: 21),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _OnlineDot extends StatelessWidget {
  const _OnlineDot();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 10,
      height: 10,
      decoration: const BoxDecoration(
        color: MokposColors.success,
        shape: BoxShape.circle,
      ),
    );
  }
}

class _TopIconButton extends StatelessWidget {
  const _TopIconButton({
    required this.icon,
    required this.onTap,
    required this.label,
    this.badge = 0,
  });

  final IconData icon;
  final VoidCallback onTap;
  final String label;
  final int badge;

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: label,
      child: IconButton(
        key: ValueKey(
          'topbar_${label.toLowerCase().replaceAll(RegExp(r'[^a-z0-9]+'), '_')}',
        ),
        onPressed: onTap,
        icon: Badge(
          isLabelVisible: badge > 0,
          label: Text('$badge'),
          child: Icon(icon),
        ),
        color: Colors.white,
        iconSize: 23,
        style: IconButton.styleFrom(
          fixedSize: const Size(46, 46),
          backgroundColor: Colors.white.withValues(alpha: .08),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(MokposRadius.sm),
          ),
        ),
      ),
    );
  }
}

Future<void> showCheckoutOutboxDialog(BuildContext context, WidgetRef ref) {
  return showDialog<void>(
    context: context,
    builder: (context) {
      final items = ref.watch(checkoutOutboxControllerProvider);
      return AlertDialog(
        backgroundColor: Colors.white,
        title: const Text('Pemulihan Checkout'),
        content: SizedBox(
          width: 460,
          child: CheckoutOutboxCenter(
            items: items,
            onRetry: (item) => ref
                .read(nojposSessionProvider.notifier)
                .retryCheckoutOutboxItem(item),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Tutup'),
          ),
        ],
      );
    },
  );
}
