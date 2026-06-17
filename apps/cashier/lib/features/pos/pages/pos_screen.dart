import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/theme.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../../../shared/models/nojpos_models.dart';
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

  void _showFeature(String label) {
    showFeatureSnack(context, label);
  }

  Future<void> _selectOrderType() async {
    final orderType = ref.read(nojposSessionProvider).activeOrderType.label;
    final selected = await showOrderTypeMenu(context, orderType);
    if (selected == null || !mounted) return;
    ref
        .read(nojposSessionProvider.notifier)
        .selectOrderType(_orderTypeFromLabel(selected));
  }

  Future<void> _selectCustomer() async {
    final session = ref.read(nojposSessionProvider);
    final selected = await showCustomerPicker(
      context,
      customers: session.customers,
      selected: session.selectedCustomer,
    );
    if (!mounted) return;
    ref.read(nojposSessionProvider.notifier).selectCustomer(selected);
  }

  @override
  Widget build(BuildContext context) {
    final orderType = ref.watch(
      nojposSessionProvider.select((session) => session.activeOrderType.label),
    );
    return Scaffold(
      backgroundColor: MokposColors.canvas,
      body: SafeArea(
        child: LayoutBuilder(
          builder: (context, constraints) {
            final compact = constraints.maxWidth < 980;
            return Column(
              children: [
                PosTopBar(
                  onOpenMenu: () => showCashierDrawer(context),
                  onOpenOrders: () => context.go('/orders'),
                  onOpenMode: () => showModeDialog(context),
                  onShowNotification: () => _showFeature('Notifikasi'),
                ),
                Expanded(
                  child: Row(
                    children: [
                      PosSidebar(
                        categoryPanelOpen: categoryPanelOpen,
                        onToggleCategories: () {
                          setState(() {
                            categoryPanelOpen = !categoryPanelOpen;
                          });
                        },
                        onOpenMode: () => showModeDialog(context),
                        onShowFeature: _showFeature,
                      ),
                      Expanded(child: ProductCatalog(compact: compact)),
                      SizedBox(
                        width: compact ? 360 : 430,
                        child: OrderPanel(
                          orderType: orderType,
                          onSelectOrderType: _selectOrderType,
                          onSelectCustomer: _selectCustomer,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            );
          },
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
