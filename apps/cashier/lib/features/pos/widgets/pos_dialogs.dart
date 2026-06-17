import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../../../shared/models/nojpos_models.dart';
import '../formatters.dart';
import '../models/cart_item.dart';

Future<void> showCashierDrawer(BuildContext context) {
  return showGeneralDialog<void>(
    context: context,
    barrierDismissible: true,
    barrierLabel: 'Menu kasir',
    barrierColor: Colors.black54,
    transitionDuration: const Duration(milliseconds: 180),
    pageBuilder: (context, animation, secondaryAnimation) {
      return Align(
        alignment: Alignment.centerLeft,
        child: Material(
          color: Colors.white,
          child: SizedBox(
            width: 330,
            height: double.infinity,
            child: SafeArea(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const _CashierHeader(),
                  const Divider(height: 1, color: MokposColors.line),
                  const _DrawerItem(
                    icon: LucideIcons.receiptText,
                    label: 'Kasir',
                    active: true,
                    route: '/pos',
                  ),
                  const _DrawerItem(
                    icon: LucideIcons.store,
                    label: 'Penjualan',
                    route: '/sales',
                  ),
                  const _DrawerItem(
                    icon: LucideIcons.clipboardList,
                    label: 'Laporan',
                    route: '/reports',
                  ),
                  const _DrawerItem(
                    icon: LucideIcons.calendarCheck,
                    label: 'Absensi',
                    route: '/attendance',
                  ),
                  const _DrawerItem(
                    icon: LucideIcons.packageOpen,
                    label: 'Inventori',
                    route: '/inventory',
                  ),
                  const _DrawerItem(
                    icon: LucideIcons.slidersHorizontal,
                    label: 'Pengaturan',
                    route: '/settings',
                  ),
                  const _DrawerItem(
                    icon: LucideIcons.lifeBuoy,
                    label: 'NojPOS Care',
                  ),
                  const Divider(height: 1, color: MokposColors.line),
                  const _DrawerItem(
                    icon: LucideIcons.badgeDollarSign,
                    label: 'Tutup Kasir',
                    closeShift: true,
                  ),
                  const Spacer(),
                  const Divider(height: 1, color: MokposColors.line),
                  _DrawerItem(
                    icon: LucideIcons.logOut,
                    label: 'Keluar Kasir',
                    onTap: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
            ),
          ),
        ),
      );
    },
    transitionBuilder: (context, animation, secondaryAnimation, child) {
      return SlideTransition(
        position: Tween<Offset>(begin: const Offset(-1, 0), end: Offset.zero)
            .animate(
              CurvedAnimation(parent: animation, curve: Curves.easeOutCubic),
            ),
        child: child,
      );
    },
  );
}

Future<void> showModeDialog(BuildContext context) {
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => Dialog(
      backgroundColor: Colors.white,
      insetPadding: const EdgeInsets.symmetric(horizontal: 170),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
      child: SizedBox(
        width: 760,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(34, 28, 34, 34),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Pilih Mode',
                style: TextStyle(
                  color: MokposColors.text,
                  fontSize: 20,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 24),
              GridView.count(
                shrinkWrap: true,
                crossAxisCount: 2,
                mainAxisSpacing: 14,
                crossAxisSpacing: 14,
                childAspectRatio: 6.2,
                children: const [
                  _ModeTile(
                    icon: LucideIcons.grid3X3,
                    label: 'Grid',
                    active: true,
                  ),
                  _ModeTile(icon: LucideIcons.list, label: 'SKU'),
                  _ModeTile(icon: LucideIcons.armchair, label: 'Meja'),
                  _ModeTile(icon: LucideIcons.handHeart, label: 'Jasa'),
                  _ModeTile(icon: LucideIcons.store, label: 'E-Commerce'),
                  _ModeTile(icon: LucideIcons.bookOpen, label: 'Buku Menu'),
                ],
              ),
            ],
          ),
        ),
      ),
    ),
  );
}

Future<String?> showOrderTypeMenu(BuildContext context, String selected) {
  return showDialog<String>(
    context: context,
    barrierColor: Colors.black26,
    builder: (context) {
      final types = [
        ('Makan di Tempat', LucideIcons.utensils),
        ('Pengiriman', LucideIcons.truck),
        ('Ojek Online', LucideIcons.bike),
        ('Quick Service', LucideIcons.timer),
        ('Dilayani Oleh', LucideIcons.userCog),
        ('Reset', LucideIcons.rotateCcw),
      ];

      return Align(
        alignment: const Alignment(.36, -.55),
        child: Material(
          color: Colors.white,
          elevation: 8,
          borderRadius: BorderRadius.circular(4),
          child: SizedBox(
            width: 260,
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: 8),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  for (final type in types)
                    ListTile(
                      dense: true,
                      leading: Icon(
                        type.$2,
                        size: 18,
                        color: type.$1 == selected
                            ? MokposColors.primary
                            : MokposColors.muted,
                      ),
                      title: Text(
                        type.$1,
                        style: TextStyle(
                          color: type.$1 == selected
                              ? MokposColors.primary
                              : MokposColors.text,
                          fontWeight: FontWeight.w800,
                          fontSize: 13,
                        ),
                      ),
                      onTap: () => Navigator.of(
                        context,
                      ).pop(type.$1 == 'Reset' ? 'Makan di Tempat' : type.$1),
                    ),
                ],
              ),
            ),
          ),
        ),
      );
    },
  );
}

Future<Customer?> showCustomerPicker(
  BuildContext context, {
  required List<Customer> customers,
  Customer? selected,
}) {
  return showDialog<Customer>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => Dialog(
      insetPadding: const EdgeInsets.symmetric(horizontal: 150, vertical: 70),
      backgroundColor: Colors.white,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(2)),
      child: Column(
        children: [
          Container(
            height: 58,
            padding: const EdgeInsets.symmetric(horizontal: 14),
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: MokposColors.line)),
            ),
            child: Row(
              children: [
                IconButton(
                  onPressed: () => Navigator.of(context).pop(selected),
                  icon: const Icon(LucideIcons.x),
                ),
                const Text(
                  'Pilih Pelanggan',
                  style: TextStyle(
                    color: MokposColors.text,
                    fontWeight: FontWeight.w900,
                    fontSize: 18,
                  ),
                ),
                const SizedBox(width: 12),
                const Text(
                  'Terakhir diperbarui : -',
                  style: TextStyle(color: MokposColors.muted, fontSize: 11),
                ),
                const Spacer(),
                TextButton.icon(
                  onPressed: () {
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(
                        content: Text('Form pelanggan baru siap disambungkan'),
                      ),
                    );
                  },
                  icon: const Icon(LucideIcons.plus, size: 18),
                  label: const Text('Pelanggan Baru'),
                ),
              ],
            ),
          ),
          const _CustomerTabs(),
          const Padding(
            padding: EdgeInsets.symmetric(horizontal: 28, vertical: 16),
            child: _CustomerSearchBox(),
          ),
          Expanded(
            child: customers.isEmpty
                ? const _EmptyCustomers()
                : ListView.separated(
                    padding: const EdgeInsets.fromLTRB(28, 0, 28, 24),
                    itemCount: customers.length + 1,
                    separatorBuilder: (context, index) =>
                        const Divider(height: 1, color: MokposColors.line),
                    itemBuilder: (context, index) {
                      if (index == 0) {
                        return _CustomerTile(
                          name: 'Tanpa Pelanggan',
                          subtitle: 'Transaksi walk-in',
                          active: selected == null,
                          onTap: () => Navigator.of(context).pop(),
                        );
                      }
                      final customer = customers[index - 1];
                      return _CustomerTile(
                        name: customer.name,
                        subtitle: [
                          if (customer.phone.isNotEmpty) customer.phone,
                          customer.group,
                        ].join(' · '),
                        active: selected?.id == customer.id,
                        onTap: () => Navigator.of(context).pop(customer),
                      );
                    },
                  ),
          ),
        ],
      ),
    ),
  );
}

class _CustomerSearchBox extends StatelessWidget {
  const _CustomerSearchBox();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 44,
      decoration: BoxDecoration(
        border: Border.all(color: MokposColors.primary),
        borderRadius: BorderRadius.circular(4),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 12),
      child: const Row(
        children: [
          Icon(LucideIcons.search, size: 18, color: MokposColors.muted),
          SizedBox(width: 10),
          Text('Cari ...', style: TextStyle(color: MokposColors.muted)),
        ],
      ),
    );
  }
}

class _CustomerTile extends StatelessWidget {
  const _CustomerTile({
    required this.name,
    required this.subtitle,
    required this.active,
    required this.onTap,
  });

  final String name;
  final String subtitle;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      onTap: onTap,
      contentPadding: const EdgeInsets.symmetric(horizontal: 6, vertical: 8),
      leading: CircleAvatar(
        backgroundColor: active ? MokposColors.primary : MokposColors.canvas,
        child: Icon(
          LucideIcons.userRound,
          color: active ? Colors.white : MokposColors.muted,
          size: 20,
        ),
      ),
      title: Text(
        name,
        style: const TextStyle(
          color: MokposColors.text,
          fontWeight: FontWeight.w900,
        ),
      ),
      subtitle: Text(subtitle),
      trailing: active
          ? const Icon(LucideIcons.check, color: MokposColors.primary)
          : null,
    );
  }
}

class _EmptyCustomers extends StatelessWidget {
  const _EmptyCustomers();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            LucideIcons.clipboardPlus,
            size: 92,
            color: MokposColors.primary,
          ),
          SizedBox(height: 18),
          Text(
            'Data Pelanggan Kosong',
            style: TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
              fontSize: 16,
            ),
          ),
          SizedBox(height: 8),
          Text(
            'Coba lagi setelah kamu tambahkan\npelanggan, ya!',
            textAlign: TextAlign.center,
            style: TextStyle(color: MokposColors.muted, height: 1.5),
          ),
        ],
      ),
    );
  }
}

Future<void> showCloseShiftDialog(BuildContext context) {
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => Dialog(
      insetPadding: const EdgeInsets.symmetric(horizontal: 190, vertical: 38),
      backgroundColor: Colors.white,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(2)),
      child: Column(
        children: [
          Container(
            height: 58,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: MokposColors.line)),
            ),
            child: Row(
              children: [
                IconButton(
                  onPressed: () => Navigator.of(context).pop(),
                  icon: const Icon(LucideIcons.x),
                ),
                const Text(
                  'Tutup Kasir · Rayhan Vito',
                  style: TextStyle(
                    color: MokposColors.text,
                    fontWeight: FontWeight.w900,
                    fontSize: 18,
                  ),
                ),
              ],
            ),
          ),
          Container(
            height: 58,
            color: MokposColors.canvas,
            padding: const EdgeInsets.symmetric(horizontal: 56),
            child: const Row(
              children: [
                Text(
                  'Buka Kasir\n16 Jun 2026, 03:21',
                  style: TextStyle(color: MokposColors.text, height: 1.5),
                ),
                Spacer(),
                Text(
                  'Tutup Kasir\n16 Jun 2026, 06:41',
                  textAlign: TextAlign.right,
                  style: TextStyle(color: MokposColors.text, height: 1.5),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: List.generate(
              6,
              (_) => Container(
                width: 11,
                height: 11,
                margin: const EdgeInsets.symmetric(horizontal: 12),
                decoration: const BoxDecoration(
                  color: MokposColors.line,
                  shape: BoxShape.circle,
                ),
              ),
            ),
          ),
          const SizedBox(height: 18),
          const Text(
            'Masukkan PIN untuk Verifikasi',
            style: TextStyle(color: MokposColors.muted),
          ),
          const SizedBox(height: 26),
          SizedBox(
            width: 430,
            height: 300,
            child: GridView.count(
              crossAxisCount: 3,
              childAspectRatio: 1.7,
              children: const [
                Center(child: Text('1', style: _pinStyle)),
                Center(child: Text('2', style: _pinStyle)),
                Center(child: Text('3', style: _pinStyle)),
                Center(child: Text('4', style: _pinStyle)),
                Center(child: Text('5', style: _pinStyle)),
                Center(child: Text('6', style: _pinStyle)),
                Center(child: Text('7', style: _pinStyle)),
                Center(child: Text('8', style: _pinStyle)),
                Center(child: Text('9', style: _pinStyle)),
                Center(child: Text('C', style: _pinStyle)),
                Center(child: Text('0', style: _pinStyle)),
                Center(child: Icon(LucideIcons.delete, size: 18)),
              ],
            ),
          ),
        ],
      ),
    ),
  );
}

const _pinStyle = TextStyle(
  color: MokposColors.text,
  fontSize: 22,
  fontWeight: FontWeight.w800,
);

Future<void> showOrdersPanel(BuildContext context, List<CartItem> items) {
  return showGeneralDialog<void>(
    context: context,
    barrierDismissible: true,
    barrierLabel: 'Daftar order',
    barrierColor: Colors.black38,
    transitionDuration: const Duration(milliseconds: 180),
    pageBuilder: (context, animation, secondaryAnimation) {
      return Align(
        alignment: Alignment.centerRight,
        child: Material(
          color: Colors.white,
          child: SizedBox(
            width: 430,
            height: double.infinity,
            child: SafeArea(
              child: Column(
                children: [
                  Container(
                    height: 64,
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    color: MokposColors.primary,
                    child: Row(
                      children: [
                        const Expanded(
                          child: Text(
                            'Daftar Order',
                            style: TextStyle(
                              color: Colors.white,
                              fontWeight: FontWeight.w900,
                              fontSize: 18,
                            ),
                          ),
                        ),
                        IconButton(
                          onPressed: () => Navigator.of(context).pop(),
                          icon: const Icon(LucideIcons.x),
                          color: Colors.white,
                        ),
                      ],
                    ),
                  ),
                  Expanded(
                    child: items.isEmpty
                        ? const Center(
                            child: Text(
                              'Belum ada order tersimpan',
                              style: TextStyle(color: MokposColors.muted),
                            ),
                          )
                        : ListView(
                            children: [
                              _SavedOrderTile(
                                title: 'Order Aktif',
                                subtitle:
                                    '${items.length} item dalam keranjang',
                                total: items.fold(
                                  0,
                                  (sum, item) => sum + item.subtotal,
                                ),
                              ),
                              const _SavedOrderTile(
                                title: 'Meja 04',
                                subtitle: 'Dine in · disimpan 5 menit lalu',
                                total: 42000,
                              ),
                              const _SavedOrderTile(
                                title: 'Take Away - Adit',
                                subtitle: 'Belum dibayar',
                                total: 27000,
                              ),
                            ],
                          ),
                  ),
                ],
              ),
            ),
          ),
        ),
      );
    },
    transitionBuilder: (context, animation, secondaryAnimation, child) {
      return SlideTransition(
        position: Tween<Offset>(begin: const Offset(1, 0), end: Offset.zero)
            .animate(
              CurvedAnimation(parent: animation, curve: Curves.easeOutCubic),
            ),
        child: child,
      );
    },
  );
}

void showFeatureSnack(BuildContext context, String label) {
  ScaffoldMessenger.of(
    context,
  ).showSnackBar(SnackBar(content: Text('$label aktif sebagai mock UI')));
}

class _CashierHeader extends StatelessWidget {
  const _CashierHeader();

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.fromLTRB(24, 24, 24, 18),
      child: Row(
        children: [
          CircleAvatar(
            backgroundColor: Color(0xFFFF897D),
            child: Text('R', style: TextStyle(color: Colors.white)),
          ),
          SizedBox(width: 14),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Kedai Nusantara',
                style: TextStyle(
                  color: MokposColors.text,
                  fontWeight: FontWeight.w900,
                ),
              ),
              SizedBox(height: 4),
              Text(
                'Rayhan · Kasir',
                style: TextStyle(color: MokposColors.muted, fontSize: 12),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _DrawerItem extends StatelessWidget {
  const _DrawerItem({
    required this.icon,
    required this.label,
    this.active = false,
    this.onTap,
    this.route,
    this.closeShift = false,
  });

  final IconData icon;
  final String label;
  final bool active;
  final VoidCallback? onTap;
  final String? route;
  final bool closeShift;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      onTap:
          onTap ??
          () {
            Navigator.of(context).pop();
            if (closeShift) {
              showCloseShiftDialog(context);
              return;
            }
            if (route != null) {
              context.go(route!);
              return;
            }
            showFeatureSnack(context, label);
          },
      leading: Icon(
        icon,
        color: active ? MokposColors.primary : MokposColors.text,
        size: 22,
      ),
      title: Text(
        label,
        style: TextStyle(
          color: active ? MokposColors.primary : MokposColors.text,
          fontWeight: FontWeight.w900,
          fontSize: 14,
        ),
      ),
    );
  }
}

class _ModeTile extends StatelessWidget {
  const _ModeTile({
    required this.icon,
    required this.label,
    this.active = false,
  });

  final IconData icon;
  final String label;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return OutlinedButton(
      onPressed: () => Navigator.of(context).pop(),
      style: OutlinedButton.styleFrom(
        alignment: Alignment.centerLeft,
        foregroundColor: MokposColors.text,
        backgroundColor: active ? MokposColors.primarySoft : Colors.white,
        side: BorderSide(
          color: active ? MokposColors.primary : MokposColors.line,
        ),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      ),
      child: Row(
        children: [
          Expanded(
            child: Text(
              label,
              style: const TextStyle(fontWeight: FontWeight.w900),
            ),
          ),
          Icon(icon, size: 18),
        ],
      ),
    );
  }
}

class _CustomerTabs extends StatelessWidget {
  const _CustomerTabs();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 52,
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: const Row(
        children: [
          SizedBox(width: 82),
          _CustomerTab(text: 'Semua', active: true),
          _CustomerTab(text: 'Tanpa Grup'),
        ],
      ),
    );
  }
}

class _CustomerTab extends StatelessWidget {
  const _CustomerTab({required this.text, this.active = false});

  final String text;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 52,
      padding: const EdgeInsets.symmetric(horizontal: 22),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        border: Border(
          bottom: BorderSide(
            color: active ? MokposColors.primary : Colors.transparent,
            width: 3,
          ),
        ),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: active ? MokposColors.primary : MokposColors.muted,
          fontWeight: FontWeight.w900,
          fontSize: 13,
        ),
      ),
    );
  }
}

class _SavedOrderTile extends StatelessWidget {
  const _SavedOrderTile({
    required this.title,
    required this.subtitle,
    required this.total,
  });

  final String title;
  final String subtitle;
  final int total;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 18, vertical: 8),
      title: Text(
        title,
        style: const TextStyle(
          color: MokposColors.text,
          fontWeight: FontWeight.w900,
        ),
      ),
      subtitle: Text(subtitle),
      trailing: Text(
        rupiah(total),
        style: const TextStyle(
          color: MokposColors.text,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}
