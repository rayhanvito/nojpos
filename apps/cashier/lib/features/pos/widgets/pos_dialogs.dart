import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../../shared/models/nojpos_models.dart';
import '../formatters.dart';
import '../models/cart_item.dart';
import '../providers/pos_providers.dart';

Future<void> showCashierDrawer(BuildContext context, WidgetRef ref) {
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
                    comingSoon: true,
                  ),
                  const Divider(height: 1, color: MokposColors.line),
                  _DrawerItem(
                    icon: LucideIcons.badgeDollarSign,
                    label: 'Tutup Kasir',
                    onTap: () {
                      Navigator.of(context).pop();
                      showCloseShiftDialog(context, ref);
                    },
                  ),
                  const Spacer(),
                  const Divider(height: 1, color: MokposColors.line),
                  _DrawerItem(
                    icon: LucideIcons.logOut,
                    label: 'Keluar Kasir',
                    onTap: () async {
                      Navigator.of(context).pop();
                      final confirmed = await showCashierLogoutDialog(context);
                      if (!context.mounted || !confirmed) return;
                      ref.read(nojposSessionProvider.notifier).lock();
                      context.go('/pin');
                    },
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

Future<bool> showCashierLogoutDialog(BuildContext context) async {
  return await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          backgroundColor: Colors.white,
          title: const Text(
            'Logout Kasir?',
            style: TextStyle(fontWeight: FontWeight.w900),
          ),
          content: const Text(
            'Kasir aktif akan dikunci dan aplikasi kembali ke halaman login karyawan. Akun aplikasi dan outlet tetap tersimpan.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(false),
              child: const Text('Batal'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(context).pop(true),
              style: FilledButton.styleFrom(
                backgroundColor: MokposColors.primary,
                foregroundColor: Colors.white,
              ),
              child: const Text('Logout Kasir'),
            ),
          ],
        ),
      ) ??
      false;
}

Future<void> showModeDialog(BuildContext context) {
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => Dialog(
      backgroundColor: Colors.white,
      insetPadding: const EdgeInsets.symmetric(horizontal: 170),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
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
                  _ModeTile(
                    icon: LucideIcons.list,
                    label: 'SKU',
                    comingSoon: true,
                  ),
                  _ModeTile(
                    icon: LucideIcons.armchair,
                    label: 'Meja',
                    comingSoon: true,
                  ),
                  _ModeTile(
                    icon: LucideIcons.handHeart,
                    label: 'Jasa',
                    comingSoon: true,
                  ),
                  _ModeTile(
                    icon: LucideIcons.store,
                    label: 'E-Commerce',
                    comingSoon: true,
                  ),
                  _ModeTile(
                    icon: LucideIcons.bookOpen,
                    label: 'Buku Menu',
                    comingSoon: true,
                  ),
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
          borderRadius: BorderRadius.circular(MokposRadius.sm),
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
  required WidgetRef ref,
}) {
  return showDialog<Customer>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => _CustomerPickerDialog(ref: ref),
  );
}

class _CustomerPickerDialog extends StatefulWidget {
  const _CustomerPickerDialog({required this.ref});

  final WidgetRef ref;

  @override
  State<_CustomerPickerDialog> createState() => _CustomerPickerDialogState();
}

class _CustomerPickerDialogState extends State<_CustomerPickerDialog> {
  final searchController = TextEditingController();
  String selectedGroup = 'Semua';

  @override
  void dispose() {
    searchController.dispose();
    super.dispose();
  }

  Future<void> _search() {
    return widget.ref
        .read(nojposSessionProvider.notifier)
        .searchCustomers(search: searchController.text, group: selectedGroup);
  }

  Future<void> _createCustomer() async {
    final created = await showDialog<Customer>(
      context: context,
      barrierColor: Colors.black54,
      builder: (context) => _NewCustomerDialog(ref: widget.ref),
    );
    if (!mounted || created == null) return;
    Navigator.of(context).pop(created);
  }

  @override
  Widget build(BuildContext context) {
    final session = widget.ref.watch(nojposSessionProvider);
    final customers = session.customers;
    final selected = session.selectedCustomer;

    return Dialog(
      insetPadding: const EdgeInsets.symmetric(horizontal: 150, vertical: 70),
      backgroundColor: Colors.white,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
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
                  onPressed: _createCustomer,
                  icon: const Icon(LucideIcons.plus, size: 18),
                  label: const Text('Pelanggan Baru'),
                ),
              ],
            ),
          ),
          _CustomerTabs(
            selected: selectedGroup,
            groups: [
              'Semua',
              'Tanpa Grup',
              ...{
                for (final customer in customers)
                  if (customer.group.isNotEmpty) customer.group,
              },
            ],
            onSelect: (group) {
              setState(() => selectedGroup = group);
              _search();
            },
          ),
          Padding(
            padding: EdgeInsets.symmetric(horizontal: 28, vertical: 16),
            child: _CustomerSearchBox(
              controller: searchController,
              onSearch: _search,
            ),
          ),
          Expanded(
            child: session.isBusy
                ? const Center(child: CircularProgressIndicator())
                : session.errorMessage != null
                ? _CustomerError(
                    message: session.errorMessage!,
                    onRetry: _search,
                  )
                : customers.isEmpty
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
    );
  }
}

class _CustomerSearchBox extends StatelessWidget {
  const _CustomerSearchBox({required this.controller, required this.onSearch});

  final TextEditingController controller;
  final VoidCallback onSearch;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 44,
      decoration: BoxDecoration(
        border: Border.all(color: MokposColors.primary),
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 12),
      child: Row(
        children: [
          const Icon(LucideIcons.search, size: 18, color: MokposColors.muted),
          const SizedBox(width: 10),
          Expanded(
            child: TextField(
              controller: controller,
              onSubmitted: (_) => onSearch(),
              decoration: const InputDecoration(
                border: InputBorder.none,
                hintText: 'Search pelanggan',
              ),
            ),
          ),
          IconButton(
            tooltip: 'Cari pelanggan',
            onPressed: onSearch,
            icon: const Icon(LucideIcons.arrowRight, size: 18),
          ),
        ],
      ),
    );
  }
}

class _CustomerError extends StatelessWidget {
  const _CustomerError({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(28),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(
              LucideIcons.wifiOff,
              size: 56,
              color: MokposColors.danger,
            ),
            const SizedBox(height: 12),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(color: MokposColors.text),
            ),
            const SizedBox(height: 16),
            OutlinedButton.icon(
              onPressed: onRetry,
              icon: const Icon(LucideIcons.refreshCcw, size: 16),
              label: const Text('Coba Lagi'),
            ),
          ],
        ),
      ),
    );
  }
}

class _NewCustomerDialog extends StatefulWidget {
  const _NewCustomerDialog({required this.ref});

  final WidgetRef ref;

  @override
  State<_NewCustomerDialog> createState() => _NewCustomerDialogState();
}

class _NewCustomerDialogState extends State<_NewCustomerDialog> {
  final nameController = TextEditingController();
  final phoneController = TextEditingController();
  final groupController = TextEditingController(text: 'Tanpa Grup');

  @override
  void dispose() {
    nameController.dispose();
    phoneController.dispose();
    groupController.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final name = nameController.text.trim();
    if (name.isEmpty) return;
    final customer = await widget.ref
        .read(nojposSessionProvider.notifier)
        .createCustomer(
          name: name,
          phone: phoneController.text,
          group: groupController.text,
        );
    if (!mounted || customer == null) return;
    Navigator.of(context).pop(customer);
  }

  @override
  Widget build(BuildContext context) {
    final busy = widget.ref.watch(nojposSessionProvider).isBusy;
    return AlertDialog(
      backgroundColor: Colors.white,
      title: const Text('Pelanggan Baru'),
      content: SizedBox(
        width: 360,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: nameController,
              autofocus: true,
              decoration: const InputDecoration(
                labelText: 'Nama',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: phoneController,
              keyboardType: TextInputType.phone,
              decoration: const InputDecoration(
                labelText: 'Nomor HP',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: groupController,
              decoration: const InputDecoration(
                labelText: 'Group',
                border: OutlineInputBorder(),
              ),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: busy ? null : () => Navigator.of(context).pop(),
          child: const Text('Batal'),
        ),
        FilledButton(
          onPressed: busy ? null : _save,
          child: Text(busy ? 'Menyimpan...' : 'Simpan'),
        ),
      ],
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

Future<void> showCartDiscountDialog(BuildContext context, WidgetRef ref) {
  final subtotalAfterItemDiscount = ref
      .read(cartProvider)
      .fold(0, (sum, item) => sum + item.total);
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => _CartDiscountDialog(
      initialValue: ref.read(checkoutDetailsProvider).cartDiscount,
      onSubmit: (value) {
        ref
            .read(checkoutDetailsProvider.notifier)
            .setCartDiscount(value, subtotalAfterItemDiscount);
      },
    ),
  );
}

Future<void> showItemDiscountDialog(
  BuildContext context,
  WidgetRef ref,
  CartItem item,
) {
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => _ItemDiscountDialog(
      initialValue: item.discount,
      productName: item.product.name,
      onSubmit: (value) {
        ref.read(cartProvider.notifier).setDiscount(item.product, value);
      },
    ),
  );
}

class _CartDiscountDialog extends StatefulWidget {
  const _CartDiscountDialog({
    required this.initialValue,
    required this.onSubmit,
  });

  final int initialValue;
  final ValueChanged<int> onSubmit;

  @override
  State<_CartDiscountDialog> createState() => _CartDiscountDialogState();
}

class _CartDiscountDialogState extends State<_CartDiscountDialog> {
  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.initialValue.toString());
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      backgroundColor: Colors.white,
      title: const Text('Diskon Cart'),
      content: TextField(
        controller: _controller,
        autofocus: true,
        keyboardType: TextInputType.number,
        decoration: const InputDecoration(
          labelText: 'Nominal diskon',
          prefixText: 'Rp ',
          border: OutlineInputBorder(),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Batal'),
        ),
        FilledButton(
          onPressed: () {
            widget.onSubmit(_parseRupiah(_controller.text));
            Navigator.of(context).pop();
          },
          child: const Text('Simpan'),
        ),
      ],
    );
  }
}

class _ItemDiscountDialog extends StatefulWidget {
  const _ItemDiscountDialog({
    required this.initialValue,
    required this.productName,
    required this.onSubmit,
  });

  final int initialValue;
  final String productName;
  final ValueChanged<int> onSubmit;

  @override
  State<_ItemDiscountDialog> createState() => _ItemDiscountDialogState();
}

class _ItemDiscountDialogState extends State<_ItemDiscountDialog> {
  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.initialValue.toString());
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      backgroundColor: Colors.white,
      title: Text('Diskon ${widget.productName}'),
      content: TextField(
        controller: _controller,
        autofocus: true,
        keyboardType: TextInputType.number,
        decoration: const InputDecoration(
          labelText: 'Nominal diskon item',
          prefixText: 'Rp ',
          border: OutlineInputBorder(),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Batal'),
        ),
        FilledButton(
          onPressed: () {
            widget.onSubmit(_parseRupiah(_controller.text));
            Navigator.of(context).pop();
          },
          child: const Text('Simpan'),
        ),
      ],
    );
  }
}

int _parseRupiah(String value) {
  return int.tryParse(value.replaceAll(RegExp(r'[^0-9]'), '')) ?? 0;
}

Future<void> showNotesDialog(BuildContext context, WidgetRef ref) {
  final controller = TextEditingController(
    text: ref.read(checkoutDetailsProvider).notes,
  );
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => AlertDialog(
      backgroundColor: Colors.white,
      title: const Text('Catatan Order'),
      content: TextField(
        controller: controller,
        autofocus: true,
        maxLines: 3,
        decoration: const InputDecoration(
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
          onPressed: () {
            ref
                .read(checkoutDetailsProvider.notifier)
                .setNotes(controller.text);
            Navigator.of(context).pop();
          },
          child: const Text('Simpan'),
        ),
      ],
    ),
  ).whenComplete(controller.dispose);
}

Future<void> showServedByDialog(BuildContext context, WidgetRef ref) {
  final cashier = ref.read(nojposSessionProvider).cashier;
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => AlertDialog(
      backgroundColor: Colors.white,
      title: const Text('Dilayani Oleh'),
      content: SizedBox(
        width: 340,
        child: ListTile(
          leading: const Icon(LucideIcons.userCog),
          title: Text(cashier.name),
          subtitle: Text(cashier.role),
          onTap: () {
            ref.read(checkoutDetailsProvider.notifier).setServedBy(cashier);
            Navigator.of(context).pop();
          },
        ),
      ),
      actions: [
        TextButton(
          onPressed: () {
            ref.read(checkoutDetailsProvider.notifier).setServedBy(null);
            Navigator.of(context).pop();
          },
          child: const Text('Kosongkan'),
        ),
      ],
    ),
  );
}

Future<void> showCloseShiftDialog(BuildContext context, WidgetRef ref) {
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => const _CloseShiftDialog(),
  );
}

class _CloseShiftDialog extends ConsumerStatefulWidget {
  const _CloseShiftDialog();

  @override
  ConsumerState<_CloseShiftDialog> createState() => _CloseShiftDialogState();
}

class _CloseShiftDialogState extends ConsumerState<_CloseShiftDialog> {
  final controller = TextEditingController();
  final pinController = TextEditingController();
  final varianceReasonController = TextEditingController();
  ShiftSession? closedShift;
  bool printReceipt = true;
  bool printSoldProducts = false;

  @override
  void initState() {
    super.initState();
    Future<void>.microtask(
      () => ref.read(nojposSessionProvider.notifier).refreshCurrentShift(),
    );
    controller.addListener(_onActualCashChanged);
  }

  void _onActualCashChanged() {
    if (mounted) setState(() {});
  }

  @override
  void dispose() {
    controller.removeListener(_onActualCashChanged);
    controller.dispose();
    pinController.dispose();
    varianceReasonController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final previewShift = closedShift ?? session.activeShift;
    final actualCash = int.tryParse(
      controller.text.replaceAll(RegExp(r'[^0-9]'), ''),
    );
    final expectedCash = previewShift?.expectedCash ?? session.cashSummary;
    final diff = actualCash == null ? null : actualCash - expectedCash;
    final isClosed = closedShift != null;

    return AlertDialog(
      backgroundColor: Colors.white,
      title: Text(
        isClosed
            ? 'Ringkasan Tutup Shift'
            : 'Tutup Shift - ${session.cashier.name}',
      ),
      content: SizedBox(
        width: 460,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              'Outlet: ${session.outlet.name}',
              style: const TextStyle(color: MokposColors.muted),
            ),
            if (previewShift != null) ...[
              const SizedBox(height: 12),
              _ShiftSummaryBox(
                title: isClosed
                    ? 'Hasil final dari server'
                    : 'Preview dari server saat ini',
                shift: previewShift,
                actualCash: closedShift?.actualCash ?? actualCash,
                fallbackExpectedCash: expectedCash,
                fallbackDifference: closedShift?.cashDifference ?? diff,
              ),
            ],
            if (!isClosed) ...[
              const SizedBox(height: 18),
              TextField(
                key: const ValueKey('close_shift_actual_cash'),
                controller: controller,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(
                  labelText: 'Kas aktual di laci',
                  helperText:
                      'Bandingkan tunai fisik dengan expected_cash dari server.',
                  prefixText: 'Rp ',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                key: const ValueKey('close_shift_pin'),
                controller: pinController,
                keyboardType: TextInputType.number,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'PIN otorisasi kasir',
                  helperText:
                      'Sesuai flow Tutup Kasir: verifikasi PIN sebelum laporan final.',
                  border: OutlineInputBorder(),
                ),
              ),
              if (diff != null && diff != 0) ...[
                const SizedBox(height: 12),
                TextField(
                  key: const ValueKey('close_shift_variance_reason'),
                  controller: varianceReasonController,
                  decoration: const InputDecoration(
                    labelText: 'Alasan selisih kas',
                    helperText:
                        'Wajib diisi ketika kas aktual berbeda dari expected_cash.',
                    border: OutlineInputBorder(),
                  ),
                ),
              ],
              CheckboxListTile(
                value: printReceipt,
                onChanged: (value) =>
                    setState(() => printReceipt = value ?? false),
                dense: true,
                contentPadding: EdgeInsets.zero,
                title: const Text('Cetak laporan tutup kasir'),
              ),
              CheckboxListTile(
                value: printSoldProducts,
                onChanged: (value) =>
                    setState(() => printSoldProducts = value ?? false),
                dense: true,
                contentPadding: EdgeInsets.zero,
                title: const Text('Cetak produk terjual'),
              ),
            ],
            if (session.errorMessage != null && !isClosed) ...[
              const SizedBox(height: 14),
              Text(
                session.errorMessage!,
                style: const TextStyle(color: MokposColors.danger),
              ),
            ],
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () {
            Navigator.of(context).pop();
            if (isClosed) {
              context.go('/shift');
            }
          },
          child: Text(isClosed ? 'Selesai' : 'Batal'),
        ),
        if (!isClosed)
          FilledButton(
            key: const ValueKey('close_shift_submit'),
            onPressed: session.isBusy
                ? null
                : () async {
                    if (actualCash == null) return;
                    final ok = await ref
                        .read(nojposSessionProvider.notifier)
                        .closeShift(
                          actualCash: actualCash,
                          pin: pinController.text,
                          varianceReason: varianceReasonController.text,
                        );
                    if (!context.mounted) return;
                    if (ok) {
                      setState(() {
                        closedShift = ref
                            .read(nojposSessionProvider)
                            .lastClosedShift;
                      });
                    } else {
                      final error = ref
                          .read(nojposSessionProvider)
                          .errorMessage;
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(content: Text(error ?? 'Tutup shift ditolak')),
                      );
                    }
                  },
            child: session.isBusy
                ? const SizedBox.square(
                    dimension: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Tutup Shift'),
          ),
      ],
    );
  }
}

class _ShiftSummaryBox extends StatelessWidget {
  const _ShiftSummaryBox({
    required this.title,
    required this.shift,
    required this.fallbackExpectedCash,
    this.actualCash,
    this.fallbackDifference,
  });

  final String title;
  final ShiftSession shift;
  final int fallbackExpectedCash;
  final int? actualCash;
  final int? fallbackDifference;

  @override
  Widget build(BuildContext context) {
    final expectedCash = shift.expectedCash ?? fallbackExpectedCash;
    final actual = shift.actualCash ?? actualCash;
    final difference = shift.cashDifference ?? fallbackDifference;
    final paymentTotals = shift.paymentTotals;
    return Container(
      key: const ValueKey('shift_summary_box'),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: MokposColors.canvas,
        borderRadius: BorderRadius.circular(MokposRadius.sm),
        border: Border.all(color: MokposColors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            title,
            style: const TextStyle(
              fontWeight: FontWeight.w900,
              color: MokposColors.text,
            ),
          ),
          const SizedBox(height: 10),
          _ShiftSummaryLine(
            label: 'Opening cash',
            value: rupiah(shift.openingCash),
          ),
          _ShiftSummaryLine(
            label: 'Expected cash',
            value: rupiah(expectedCash),
          ),
          _ShiftSummaryLine(
            label: 'Kas aktual',
            value: actual == null ? '-' : rupiah(actual),
          ),
          _ShiftSummaryLine(
            label: 'Selisih kas',
            value: difference == null ? '-' : rupiah(difference),
            isDanger: (difference ?? 0) != 0,
          ),
          const Divider(height: 18, color: MokposColors.line),
          const Text(
            'Payment totals',
            style: TextStyle(
              fontWeight: FontWeight.w900,
              color: MokposColors.text,
              fontSize: 12,
            ),
          ),
          if (paymentTotals.isEmpty)
            const Padding(
              padding: EdgeInsets.only(top: 6),
              child: Text(
                'Belum ada ringkasan payment_totals dari server.',
                style: TextStyle(color: MokposColors.muted, fontSize: 12),
              ),
            )
          else
            for (final total in paymentTotals)
              _ShiftSummaryLine(
                label: '${total.method}${total.isCash ? ' (cash)' : ''}',
                value: rupiah(total.amount),
              ),
        ],
      ),
    );
  }
}

class _ShiftSummaryLine extends StatelessWidget {
  const _ShiftSummaryLine({
    required this.label,
    required this.value,
    this.isDanger = false,
  });

  final String label;
  final String value;
  final bool isDanger;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3),
      child: Row(
        children: [
          Text(label, style: const TextStyle(color: MokposColors.muted)),
          const Spacer(),
          Text(
            value,
            style: TextStyle(
              color: isDanger ? MokposColors.danger : MokposColors.text,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

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
  ).showSnackBar(SnackBar(content: Text('$label segera hadir')));
}

Future<void> showOperationalFeatureDialog(
  BuildContext context,
  WidgetRef ref,
  String label,
) {
  return switch (label) {
    'Kas / Wallet' => showCashWalletDialog(context, ref),
    'Buku Menu' => showMenuBookDialog(context, ref),
    'Promo / Voucher' ||
    'NojPOS Care' => showFutureFeatureDialog(context, label),
    _ => showFutureFeatureDialog(context, label),
  };
}

Future<void> showFutureFeatureDialog(BuildContext context, String label) {
  return showDialog<void>(
    context: context,
    builder: (context) => AlertDialog(
      backgroundColor: Colors.white,
      title: Text(label, style: const TextStyle(fontWeight: FontWeight.w900)),
      content: const Text(
        'Segera hadir. Fitur ini belum memiliki backend MVP dan tidak memakai logika palsu.',
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Tutup'),
        ),
      ],
    ),
  );
}

Future<void> showMenuBookDialog(BuildContext context, WidgetRef ref) {
  return showDialog<void>(
    context: context,
    builder: (context) {
      final catalog = ref.watch(posCatalogProvider);
      return AlertDialog(
        backgroundColor: Colors.white,
        title: const Text(
          'Buku Menu',
          style: TextStyle(fontWeight: FontWeight.w900),
        ),
        content: SizedBox(
          width: 560,
          height: 420,
          child: catalog.isLoading
              ? const Center(child: CircularProgressIndicator())
              : catalog.errorMessage != null
              ? Text(catalog.errorMessage!)
              : Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        for (final category in catalog.categories)
                          Chip(label: Text(category)),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Expanded(
                      child: catalog.products.isEmpty
                          ? const Center(
                              child: Text(
                                'Belum ada produk dari API.',
                                style: TextStyle(color: MokposColors.muted),
                              ),
                            )
                          : ListView.separated(
                              itemCount: catalog.products.length,
                              separatorBuilder: (context, index) =>
                                  const Divider(
                                    height: 1,
                                    color: MokposColors.line,
                                  ),
                              itemBuilder: (context, index) {
                                final product = catalog.products[index];
                                return ListTile(
                                  contentPadding: EdgeInsets.zero,
                                  title: Text(
                                    product.name,
                                    style: const TextStyle(
                                      fontWeight: FontWeight.w900,
                                    ),
                                  ),
                                  subtitle: Text(product.category),
                                  trailing: Text(rupiah(product.price)),
                                );
                              },
                            ),
                    ),
                  ],
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

Future<void> showCashWalletDialog(BuildContext context, WidgetRef ref) {
  final amountController = TextEditingController();
  final reasonController = TextEditingController();
  return showDialog<void>(
    context: context,
    builder: (context) {
      final session = ref.watch(nojposSessionProvider);
      final cashIn = session.cashMovements
          .where((movement) => movement.type == CashMovementType.cashIn)
          .fold(0, (sum, movement) => sum + movement.amount);
      final cashOut = session.cashMovements
          .where((movement) => movement.type == CashMovementType.cashOut)
          .fold(0, (sum, movement) => sum + movement.amount);
      return AlertDialog(
        backgroundColor: Colors.white,
        title: const Text(
          'Kas / Wallet',
          style: TextStyle(fontWeight: FontWeight.w900),
        ),
        content: SizedBox(
          width: 560,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                children: [
                  Expanded(
                    child: _CashMetric(
                      label: 'Saldo Kas',
                      value: rupiah(session.cashSummary),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: _CashMetric(
                      label: 'Kas Masuk',
                      value: rupiah(cashIn),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: _CashMetric(
                      label: 'Kas Keluar',
                      value: rupiah(cashOut),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: amountController,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(
                        labelText: 'Nominal',
                        prefixText: 'Rp ',
                        border: OutlineInputBorder(),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: TextField(
                      controller: reasonController,
                      decoration: const InputDecoration(
                        labelText: 'Catatan',
                        border: OutlineInputBorder(),
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 14),
              if (session.cashMovements.isEmpty)
                const Text(
                  'Belum ada cash movement pada shift ini.',
                  style: TextStyle(color: MokposColors.muted),
                )
              else
                SizedBox(
                  height: 180,
                  child: ListView.separated(
                    itemCount: session.cashMovements.length,
                    separatorBuilder: (context, index) =>
                        const Divider(height: 1, color: MokposColors.line),
                    itemBuilder: (context, index) {
                      final movement = session.cashMovements[index];
                      return ListTile(
                        contentPadding: EdgeInsets.zero,
                        title: Text(movement.type.label),
                        subtitle: Text(movement.reason),
                        trailing: Text(rupiah(movement.signedAmount)),
                      );
                    },
                  ),
                ),
              if (session.errorMessage != null) ...[
                const SizedBox(height: 8),
                Text(
                  session.errorMessage!,
                  style: const TextStyle(color: MokposColors.danger),
                ),
              ],
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Tutup'),
          ),
          OutlinedButton(
            onPressed: session.isBusy
                ? null
                : () => _submitCashMovement(
                    context,
                    ref,
                    amountController,
                    reasonController,
                    CashMovementType.cashOut,
                  ),
            child: const Text('Kas Keluar'),
          ),
          FilledButton(
            onPressed: session.isBusy
                ? null
                : () => _submitCashMovement(
                    context,
                    ref,
                    amountController,
                    reasonController,
                    CashMovementType.cashIn,
                  ),
            child: const Text('Kas Masuk'),
          ),
        ],
      );
    },
  ).whenComplete(() {
    amountController.dispose();
    reasonController.dispose();
  });
}

Future<void> _submitCashMovement(
  BuildContext context,
  WidgetRef ref,
  TextEditingController amountController,
  TextEditingController reasonController,
  CashMovementType type,
) async {
  final amount =
      int.tryParse(amountController.text.replaceAll(RegExp(r'[^0-9]'), '')) ??
      0;
  final ok = await ref
      .read(nojposSessionProvider.notifier)
      .addCashMovement(
        type: type,
        amount: amount,
        reason: reasonController.text,
      );
  if (!context.mounted || !ok) return;
  amountController.clear();
  reasonController.clear();
  ScaffoldMessenger.of(
    context,
  ).showSnackBar(SnackBar(content: Text('${type.label} tersimpan')));
}

class _CashMetric extends StatelessWidget {
  const _CashMetric({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: const TextStyle(
              color: MokposColors.muted,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 4),
          Text(value, style: const TextStyle(fontWeight: FontWeight.w900)),
        ],
      ),
    );
  }
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
            backgroundColor: MokposColors.danger,
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
    this.comingSoon = false,
  });

  final IconData icon;
  final String label;
  final bool active;
  final VoidCallback? onTap;
  final String? route;
  final bool comingSoon;

  @override
  Widget build(BuildContext context) {
    final content = ListTile(
      key: ValueKey(
        'drawer_${label.toLowerCase().replaceAll(RegExp(r'[^a-z0-9]+'), '_')}',
      ),
      enabled: !comingSoon,
      onTap: comingSoon
          ? null
          : onTap ??
                () {
                  Navigator.of(context).pop();
                  if (route != null) {
                    context.go(route!);
                    return;
                  }
                  showFeatureSnack(context, label);
                },
      leading: Icon(
        icon,
        color: active
            ? MokposColors.primary
            : comingSoon
            ? MokposColors.muted
            : MokposColors.text,
        size: 22,
      ),
      title: Row(
        children: [
          Expanded(
            child: Text(
              label,
              style: TextStyle(
                color: active
                    ? MokposColors.primary
                    : comingSoon
                    ? MokposColors.muted
                    : MokposColors.text,
                fontWeight: FontWeight.w900,
                fontSize: 14,
              ),
            ),
          ),
          if (comingSoon) const _SoonBadge(),
        ],
      ),
    );
    if (!comingSoon) return content;
    return Tooltip(message: '$label segera hadir', child: content);
  }
}

class _SoonBadge extends StatelessWidget {
  const _SoonBadge();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
      decoration: BoxDecoration(
        color: MokposColors.disabledSurface,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(MokposRadius.xs),
      ),
      child: const Text(
        'Segera hadir',
        style: TextStyle(
          color: MokposColors.muted,
          fontSize: 10,
          fontWeight: FontWeight.w900,
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
    this.comingSoon = false,
  });

  final IconData icon;
  final String label;
  final bool active;
  final bool comingSoon;

  @override
  Widget build(BuildContext context) {
    final button = OutlinedButton(
      onPressed: comingSoon ? null : () => Navigator.of(context).pop(),
      style: OutlinedButton.styleFrom(
        alignment: Alignment.centerLeft,
        foregroundColor: MokposColors.text,
        disabledForegroundColor: MokposColors.muted,
        backgroundColor: active
            ? MokposColors.primarySoft
            : comingSoon
            ? MokposColors.disabledSurface
            : Colors.white,
        side: BorderSide(
          color: active ? MokposColors.primary : MokposColors.line,
        ),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(MokposRadius.sm),
        ),
      ),
      child: Row(
        children: [
          Expanded(
            child: Text(
              label,
              style: const TextStyle(fontWeight: FontWeight.w900),
            ),
          ),
          if (comingSoon) ...[const _SoonBadge(), const SizedBox(width: 8)],
          Icon(icon, size: 18),
        ],
      ),
    );
    if (!comingSoon) return button;
    return Tooltip(message: '$label segera hadir', child: button);
  }
}

class _CustomerTabs extends StatelessWidget {
  const _CustomerTabs({
    required this.selected,
    required this.groups,
    required this.onSelect,
  });

  final String selected;
  final List<String> groups;
  final ValueChanged<String> onSelect;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 52,
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: [
            const SizedBox(width: 82),
            for (final group in groups)
              _CustomerTab(
                text: group,
                active: selected == group,
                onTap: () => onSelect(group),
              ),
          ],
        ),
      ),
    );
  }
}

class _CustomerTab extends StatelessWidget {
  const _CustomerTab({
    required this.text,
    required this.onTap,
    this.active = false,
  });

  final String text;
  final VoidCallback onTap;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Container(
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
