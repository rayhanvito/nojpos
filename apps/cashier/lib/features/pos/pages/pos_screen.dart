import 'dart:ui';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/theme.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../../../shared/models/nojpos_models.dart';
import '../../screen_lock/widgets/terminal_lock_overlay.dart';
import '../providers/parked_order_lease_controller.dart';
import '../providers/pos_providers.dart';
import '../widgets/order_panel.dart';
import '../widgets/pos_dialogs.dart';
import '../widgets/pos_sidebar.dart';
import '../widgets/product_catalog.dart';
import '../widgets/top_bar.dart';

class PosScreen extends ConsumerStatefulWidget {
  const PosScreen({super.key});

  @override
  ConsumerState<PosScreen> createState() => _PosScreenState();
}

class _PosScreenState extends ConsumerState<PosScreen> {
  bool categoryPanelOpen = false;
  bool _openingCashDialogVisible = false;

  @override
  void initState() {
    super.initState();
    Future<void>.microtask(() async {
      await ref.read(nojposSessionProvider.notifier).refreshCurrentShift();
      _showOpeningCashDialogIfNeeded();
    });
  }

  void _showOpeningCashDialogIfNeeded() {
    if (!mounted || _openingCashDialogVisible) return;
    final session = ref.read(nojposSessionProvider);
    if (session.isBusy || session.hasOpenShift || session.cashier.id.isEmpty) {
      return;
    }

    _openingCashDialogVisible = true;
    showOpenShiftDialog(context, ref).whenComplete(() {
      _openingCashDialogVisible = false;
      if (mounted) {
        _showOpeningCashDialogIfNeeded();
      }
    });
  }

  void _showFeature(String label) {
    showOperationalFeatureDialog(context, ref, label);
  }

  Future<void> _showMobileOrderSheet(Widget orderPanel) {
    return showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => FractionallySizedBox(
        heightFactor: .88,
        child: DecoratedBox(
          decoration: const BoxDecoration(
            color: NojposColors.surface,
            borderRadius: BorderRadius.vertical(
              top: Radius.circular(NojposRadius.xl),
            ),
            boxShadow: NojposShadow.floating,
          ),
          child: ClipRRect(
            borderRadius: const BorderRadius.vertical(
              top: Radius.circular(NojposRadius.xl),
            ),
            child: SafeArea(top: false, child: orderPanel),
          ),
        ),
      ),
    );
  }

  Future<void> _selectOrderType() async {
    final orderType = ref.read(nojposSessionProvider).activeOrderType.label;
    final selected = await showOrderTypeMenu(context, orderType);
    if (selected == null || !mounted) return;
    if (selected == 'Dilayani Oleh') {
      await showServedByDialog(context, ref);
      return;
    }
    ref
        .read(nojposSessionProvider.notifier)
        .selectOrderType(_orderTypeFromLabel(selected));
  }

  Future<void> _selectCustomer() async {
    await ref.read(nojposSessionProvider.notifier).searchCustomers();
    if (!mounted) return;
    final selected = await showCustomerPicker(context, ref: ref);
    if (!mounted) return;
    ref.read(nojposSessionProvider.notifier).selectCustomer(selected);
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    if (!session.hasOpenShift) {
      WidgetsBinding.instance.addPostFrameCallback(
        (_) => _showOpeningCashDialogIfNeeded(),
      );
    }
    final orderType = ref.watch(
      nojposSessionProvider.select((session) => session.activeOrderType.label),
    );
    final needsOpenCashier =
        !session.hasOpenShift && session.cashier.id.isNotEmpty;
    return TerminalLockGate(
      child: Scaffold(
        backgroundColor: MokposColors.canvas,
        body: SafeArea(
          child: LayoutBuilder(
            builder: (context, constraints) {
              final compact = constraints.maxWidth < 980;
              final mobile = constraints.maxWidth < 720;
              final orderPanel = OrderPanel(
                orderType: orderType,
                onSelectOrderType: _selectOrderType,
                onSelectCustomer: _selectCustomer,
              );

              return Stack(
                children: [
                  Column(
                    children: [
                      PosTopBar(
                        onOpenMenu: () => showCashierDrawer(context, ref),
                        onOpenOrders: () async {
                          await ref
                              .read(parkedOrderLeaseControllerProvider.notifier)
                              .releaseActive(restoreOrder: true);
                          if (!context.mounted) return;
                          context.go('/orders');
                        },
                        onOpenMode: () => showModeDialog(context, ref),
                        onShowNotification: () => _showFeature('Notifikasi'),
                      ),
                      Expanded(
                        child: mobile
                            ? ProductCatalog(compact: true)
                            : Row(
                                children: [
                                  PosSidebar(
                                    categoryPanelOpen: categoryPanelOpen,
                                    onToggleCategories: () {
                                      setState(() {
                                        categoryPanelOpen = !categoryPanelOpen;
                                      });
                                    },
                                    onOpenMode: () => showModeDialog(context, ref),
                                    onShowFeature: _showFeature,
                                  ),
                                  Expanded(child: ProductCatalog(compact: compact)),
                                  SizedBox(
                                    width: compact ? 360 : 430,
                                    child: orderPanel,
                                  ),
                                ],
                              ),
                      ),
                    ],
                  ),
                  if (mobile)
                    _MobileOrderButton(
                      itemCount: ref
                          .watch(cartProvider)
                          .fold(0, (sum, item) => sum + item.quantity),
                      onPressed: () => _showMobileOrderSheet(orderPanel),
                    ),
                  if (needsOpenCashier) const _CashierLockedBackdrop(),
                ],
              );
            },
          ),
        ),
      ),
    );
  }
}

class _MobileOrderButton extends StatelessWidget {
  const _MobileOrderButton({required this.itemCount, required this.onPressed});

  final int itemCount;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return Positioned(
      left: NojposSpacing.lg,
      right: NojposSpacing.lg,
      bottom: NojposSpacing.lg,
      child: SafeArea(
        minimum: const EdgeInsets.only(bottom: NojposSpacing.sm),
        child: FilledButton.icon(
          key: const ValueKey('mobile_order_button'),
          onPressed: onPressed,
          icon: Badge(
            isLabelVisible: itemCount > 0,
            label: Text('$itemCount'),
            child: const Icon(Icons.shopping_cart_checkout_rounded),
          ),
          label: Text(
            itemCount > 0 ? 'Lihat Keranjang' : 'Buka Keranjang',
          ),
          style: FilledButton.styleFrom(
            backgroundColor: NojposColors.primary,
            foregroundColor: Colors.white,
            minimumSize: const Size.fromHeight(56),
            elevation: 8,
            shadowColor: NojposColors.primary.withValues(alpha: .28),
          ),
        ),
      ),
    );
  }
}

class _CashierLockedBackdrop extends StatelessWidget {
  const _CashierLockedBackdrop();

  @override
  Widget build(BuildContext context) {
    return Positioned.fill(
      child: AbsorbPointer(
        child: BackdropFilter(
          filter: ImageFilter.blur(sigmaX: 3, sigmaY: 3),
          child: Container(
            color: Colors.white.withValues(alpha: 0.68),
            alignment: Alignment.center,
            child: Container(
              width: 380,
              padding: const EdgeInsets.all(22),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(MokposRadius.lg),
                border: Border.all(color: MokposColors.line),
                boxShadow: MokposShadow.soft,
              ),
              child: const Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    Icons.lock_outline_rounded,
                    size: 36,
                    color: MokposColors.primary,
                  ),
                  SizedBox(height: 12),
                  Text(
                    'Kasir belum dibuka',
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      color: MokposColors.text,
                      fontSize: 20,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  SizedBox(height: 8),
                  Text(
                    'Isi modal awal pada popup Buka Kasir untuk mulai transaksi.',
                    textAlign: TextAlign.center,
                    style: TextStyle(color: MokposColors.muted, height: 1.35),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

OrderType _orderTypeFromLabel(String label) {
  return switch (label) {
    'Pengiriman' => OrderType.delivery,
    'Ojek Online' => OrderType.online,
    'Quick Service' => OrderType.quickService,
    'Ambil Sendiri' => OrderType.pickup,
    _ => OrderType.dineIn,
  };
}
