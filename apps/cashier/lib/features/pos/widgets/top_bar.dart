import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../../core/outbox/checkout_outbox.dart';
import '../../connectivity/widgets/connectivity_status_chip.dart';
import '../../notifications/widgets/checkout_outbox_center.dart';
import '../../screen_lock/providers/terminal_lock_controller.dart';
import '../../screen_lock/repositories/terminal_lock_repository.dart';

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
    return LayoutBuilder(
      builder: (context, constraints) {
        final compact = constraints.maxWidth < 720;
        final medium = constraints.maxWidth < 980;
        final title = '${session.outlet.name} - ${session.cashier.name}';

        return Container(
          height: compact ? 58 : 64,
          color: NojposColors.primary,
          padding: EdgeInsets.symmetric(horizontal: compact ? 8 : 14),
          child: Row(
            children: [
              _TopIconButton(
                icon: LucideIcons.menu,
                label: 'Menu',
                compact: compact,
                onTap: onOpenMenu,
              ),
              SizedBox(width: compact ? 6 : 10),
              if (!compact) ...[
                Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(NojposRadius.md),
                  ),
                  child: const Icon(
                    LucideIcons.store,
                    color: NojposColors.primary,
                    size: 25,
                  ),
                ),
                const SizedBox(width: 12),
              ],
              Expanded(
                flex: medium ? 2 : 1,
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      compact ? session.outlet.name : title,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.w900,
                        fontSize: compact ? 14 : 16,
                      ),
                    ),
                    const SizedBox(height: 2),
                    ConnectivityStatusChip(
                      suffix: compact
                          ? null
                          : session.hasOpenShift
                          ? 'Shift Aktif'
                          : 'Shift belum dibuka',
                      textStyle: const TextStyle(
                        color: NojposColors.onPrimaryMuted,
                        fontSize: 12,
                      ),
                      dotSize: compact ? 7 : 8,
                    ),
                  ],
                ),
              ),
              if (!medium) ...[
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
              ],
              _TopIconButton(
                icon: LucideIcons.bell,
                label: 'Notifikasi',
                badge: blockingOutboxCount,
                compact: compact,
                onTap: () => showCheckoutOutboxDialog(context, ref),
              ),
              if (!compact)
                _TopIconButton(
                  icon: LucideIcons.layoutGrid,
                  label: 'Mode order',
                  onTap: onOpenMode,
                ),
              _TopIconButton(
                icon: LucideIcons.lockKeyhole,
                label: 'Kunci layar',
                compact: compact,
                onTap: () {
                  ref
                      .read(terminalLockControllerProvider.notifier)
                      .lock(reason: TerminalLockReason.manual);
                },
              ),
              SizedBox(width: compact ? 6 : 10),
              _OrdersButton(compact: compact, onTap: onOpenOrders),
            ],
          ),
        );
      },
    );
  }
}

class _TopIconButton extends StatelessWidget {
  const _TopIconButton({
    required this.icon,
    required this.onTap,
    required this.label,
    this.badge = 0,
    this.compact = false,
  });

  final IconData icon;
  final VoidCallback onTap;
  final String label;
  final int badge;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final size = compact ? 40.0 : 46.0;
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
        iconSize: compact ? 21 : 23,
        style: IconButton.styleFrom(
          fixedSize: Size(size, size),
          backgroundColor: Colors.white.withValues(alpha: .10),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(NojposRadius.sm),
          ),
        ),
      ),
    );
  }
}

class _OrdersButton extends StatelessWidget {
  const _OrdersButton({required this.compact, required this.onTap});

  final bool compact;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: 'Daftar Order',
      child: GestureDetector(
        key: const ValueKey('topbar_orders'),
        behavior: HitTestBehavior.opaque,
        onTap: onTap,
        child: Container(
          height: compact ? 40 : 48,
          width: compact ? 44 : 210,
          padding: EdgeInsets.symmetric(horizontal: compact ? 0 : 18),
          decoration: BoxDecoration(
            color: NojposColors.primaryDark,
            borderRadius: BorderRadius.circular(NojposRadius.sm),
          ),
          child: compact
              ? const Icon(LucideIcons.listOrdered, color: Colors.white, size: 21)
              : const Row(
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
                    Icon(
                      LucideIcons.chevronRight,
                      color: Colors.white,
                      size: 21,
                    ),
                  ],
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
