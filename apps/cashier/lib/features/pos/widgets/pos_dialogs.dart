import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../../shared/models/nojpos_models.dart';
import '../../../shared/widgets/nojpos_dialog.dart';
import '../../../shared/widgets/nojpos_toast.dart';
import '../formatters.dart';
import '../models/cart_item.dart';
import '../models/product.dart';
import '../providers/pos_providers.dart';
import '../providers/sku_entry_provider.dart';

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
                  _DrawerItem(
                    icon: LucideIcons.lifeBuoy,
                    label: 'NojPOS Care',
                    onTap: () {
                      Navigator.of(context).pop();
                      showNojposCareDialog(context, ref);
                    },
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
        builder: (context) => const NojposConfirmDialog(
          title: 'Keluar Kasir?',
          subtitle:
              'Sesi kasir akan dikunci dan aplikasi kembali ke pilih karyawan + PIN. Kasir yang sedang terbuka tidak ikut ditutup, akun aplikasi dan outlet tetap tersimpan.',
          confirmLabel: 'Keluar Kasir',
          cancelLabel: 'Batal',
          icon: LucideIcons.logOut,
        ),
      ) ??
      false;
}

Future<void> showModeDialog(BuildContext context, WidgetRef ref) {
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
                childAspectRatio: 3.9,
                children: [
                  const _ModeTile(
                    icon: LucideIcons.grid3X3,
                    label: 'Grid',
                    active: true,
                  ),
                  _ModeTile(
                    icon: LucideIcons.list,
                    label: 'SKU',
                    onPressed: () {
                      Navigator.of(context).pop();
                      showSkuEntryDialog(context, ref);
                    },
                  ),
                  const _ModeTile(
                    icon: LucideIcons.armchair,
                    label: 'Meja',
                    roadmapState: true,
                    roadmapCopy:
                        'Mode meja nonaktif di terminal kasir preview. Gunakan Quick Service atau Ambil Sendiri.',
                  ),
                  const _ModeTile(
                    icon: LucideIcons.handHeart,
                    label: 'Jasa',
                    roadmapState: true,
                    roadmapCopy:
                        'Produk jasa nonaktif di terminal kasir preview. Gunakan katalog produk aktif.',
                  ),
                  const _ModeTile(
                    icon: LucideIcons.store,
                    label: 'E-Commerce',
                    roadmapState: true,
                    roadmapCopy:
                        'Order online nonaktif di terminal kasir preview. Gunakan order tersimpan untuk operasional outlet.',
                  ),
                  _ModeTile(
                    icon: LucideIcons.bookOpen,
                    label: 'Buku Menu',
                    onPressed: () {
                      Navigator.of(context).pop();
                      showMenuBookDialog(context, ref);
                    },
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

Future<void> showSkuEntryDialog(BuildContext context, WidgetRef ref) {
  ref.read(skuEntryProvider.notifier).clear();
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => const _SkuEntryDialog(),
  );
}

class _SkuEntryDialog extends ConsumerStatefulWidget {
  const _SkuEntryDialog();

  @override
  ConsumerState<_SkuEntryDialog> createState() => _SkuEntryDialogState();
}

class _SkuEntryDialogState extends ConsumerState<_SkuEntryDialog> {
  final _searchController = TextEditingController();
  int _quantity = 1;

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _search() async {
    await ref.read(skuEntryProvider.notifier).search(_searchController.text);
  }

  void _addToCart(Product product) {
    final cart = ref.read(cartProvider.notifier);
    for (var index = 0; index < _quantity; index += 1) {
      cart.add(product);
    }
    NojposToast.success(
      context,
      '${product.name} x$_quantity ditambahkan ke cart',
    );
    Navigator.of(context).pop();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(skuEntryProvider);
    return Dialog(
      backgroundColor: Colors.white,
      insetPadding: const EdgeInsets.symmetric(horizontal: 120, vertical: 48),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 820, maxHeight: 680),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(28, 24, 28, 18),
              child: Row(
                children: [
                  Container(
                    width: 42,
                    height: 42,
                    decoration: BoxDecoration(
                      color: MokposColors.primarySoft,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(
                      LucideIcons.listChecks,
                      color: MokposColors.primary,
                      size: 22,
                    ),
                  ),
                  const SizedBox(width: 14),
                  const Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Mode SKU',
                          style: TextStyle(
                            color: MokposColors.text,
                            fontSize: 20,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                        SizedBox(height: 4),
                        Text(
                          'Cari produk berdasarkan nama, SKU/barcode, lalu tambahkan ke cart.',
                          style: TextStyle(
                            color: MokposColors.muted,
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                      ],
                    ),
                  ),
                  IconButton(
                    tooltip: 'Tutup',
                    onPressed: () => Navigator.of(context).pop(),
                    icon: const Icon(LucideIcons.x, size: 20),
                  ),
                ],
              ),
            ),
            const Divider(height: 1, color: MokposColors.line),
            Padding(
              padding: const EdgeInsets.fromLTRB(28, 18, 28, 12),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      key: const ValueKey('sku_mode_search_field'),
                      controller: _searchController,
                      autofocus: true,
                      textInputAction: TextInputAction.search,
                      onSubmitted: (_) => _search(),
                      decoration: InputDecoration(
                        prefixIcon: const Icon(
                          LucideIcons.search,
                          color: MokposColors.muted,
                        ),
                        hintText: 'Ketik SKU, barcode, atau nama produk',
                        filled: true,
                        fillColor: MokposColors.canvas,
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(MokposRadius.sm),
                          borderSide: const BorderSide(
                            color: MokposColors.line,
                          ),
                        ),
                        enabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(MokposRadius.sm),
                          borderSide: const BorderSide(
                            color: MokposColors.line,
                          ),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  _QuantityStepper(
                    quantity: _quantity,
                    onChanged: (value) => setState(() => _quantity = value),
                  ),
                  const SizedBox(width: 12),
                  FilledButton.icon(
                    key: const ValueKey('sku_mode_search_button'),
                    onPressed: state.isLoading ? null : _search,
                    icon: state.isLoading
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(LucideIcons.search, size: 16),
                    label: const Text('Cari'),
                  ),
                ],
              ),
            ),
            Expanded(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(28, 6, 28, 24),
                child: _SkuEntryResults(
                  state: state,
                  quantity: _quantity,
                  onAdd: _addToCart,
                  onRetry: _search,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _QuantityStepper extends StatelessWidget {
  const _QuantityStepper({required this.quantity, required this.onChanged});

  final int quantity;
  final ValueChanged<int> onChanged;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 52,
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          IconButton(
            tooltip: 'Kurangi jumlah',
            onPressed: quantity <= 1 ? null : () => onChanged(quantity - 1),
            icon: const Icon(LucideIcons.minus, size: 16),
          ),
          SizedBox(
            width: 34,
            child: Text(
              '$quantity',
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          IconButton(
            tooltip: 'Tambah jumlah',
            onPressed: quantity >= 99 ? null : () => onChanged(quantity + 1),
            icon: const Icon(LucideIcons.plus, size: 16),
          ),
        ],
      ),
    );
  }
}

class _SkuEntryResults extends StatelessWidget {
  const _SkuEntryResults({
    required this.state,
    required this.quantity,
    required this.onAdd,
    required this.onRetry,
  });

  final SkuEntryState state;
  final int quantity;
  final ValueChanged<Product> onAdd;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    if (state.isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state.errorMessage != null) {
      return _SkuStateCard(
        icon: LucideIcons.wifiOff,
        title: 'Pencarian produk gagal',
        subtitle: state.errorMessage!,
        action: OutlinedButton.icon(
          onPressed: onRetry,
          icon: const Icon(LucideIcons.refreshCcw, size: 16),
          label: const Text('Coba Lagi'),
        ),
      );
    }
    if (!state.hasSearched) {
      return const _SkuStateCard(
        icon: LucideIcons.keyboard,
        title: 'Mulai dari SKU atau barcode',
        subtitle:
            'Gunakan keyboard scanner atau ketik sebagian nama produk. Harga tetap berasal dari katalog backend.',
      );
    }
    if (state.results.isEmpty) {
      return _SkuStateCard(
        icon: LucideIcons.packageSearch,
        title: 'Produk tidak ditemukan',
        subtitle:
            'Tidak ada hasil untuk "${state.query}". Periksa SKU/barcode atau cari dengan nama produk.',
      );
    }

    return ListView.separated(
      itemCount: state.results.length,
      separatorBuilder: (context, index) =>
          const Divider(height: 1, color: MokposColors.line),
      itemBuilder: (context, index) {
        final product = state.results[index];
        return ListTile(
          key: ValueKey('sku_result_${product.id}'),
          contentPadding: const EdgeInsets.symmetric(
            horizontal: 6,
            vertical: 8,
          ),
          leading: CircleAvatar(
            backgroundColor: MokposColors.primarySoft,
            foregroundColor: MokposColors.primaryDark,
            child: Text(
              product.name.trim().isEmpty
                  ? '?'
                  : product.name.characters.first.toUpperCase(),
              style: const TextStyle(fontWeight: FontWeight.w900),
            ),
          ),
          title: Text(
            product.name,
            style: const TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
            ),
          ),
          subtitle: Text(
            '${product.category} • ${product.badge ?? 'Tanpa barcode'}',
            style: const TextStyle(
              color: MokposColors.muted,
              fontWeight: FontWeight.w700,
            ),
          ),
          trailing: FilledButton.icon(
            key: ValueKey('sku_add_${product.id}'),
            onPressed: () => onAdd(product),
            icon: const Icon(LucideIcons.plus, size: 16),
            label: Text('Tambah x$quantity • ${rupiah(product.price)}'),
          ),
        );
      },
    );
  }
}

class _SkuStateCard extends StatelessWidget {
  const _SkuStateCard({
    required this.icon,
    required this.title,
    required this.subtitle,
    this.action,
  });

  final IconData icon;
  final String title;
  final String subtitle;
  final Widget? action;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        constraints: const BoxConstraints(maxWidth: 460),
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          color: MokposColors.canvas,
          border: Border.all(color: MokposColors.line),
          borderRadius: BorderRadius.circular(MokposRadius.sm),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 34, color: MokposColors.primary),
            const SizedBox(height: 12),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w900,
                fontSize: 17,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              subtitle,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: MokposColors.muted,
                fontWeight: FontWeight.w700,
              ),
            ),
            if (action != null) ...[const SizedBox(height: 14), action!],
          ],
        ),
      ),
    );
  }
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

Future<void> showOpenShiftDialog(BuildContext context, WidgetRef ref) {
  return showDialog<void>(
    context: context,
    barrierDismissible: false,
    barrierColor: Colors.black.withValues(alpha: 0.62),
    builder: (context) => const _OpenShiftDialog(),
  );
}

class _OpenShiftDialog extends ConsumerStatefulWidget {
  const _OpenShiftDialog();

  @override
  ConsumerState<_OpenShiftDialog> createState() => _OpenShiftDialogState();
}

class _OpenShiftDialogState extends ConsumerState<_OpenShiftDialog> {
  final openingCashController = TextEditingController();
  final noteController = TextEditingController();
  final openedAt = DateTime.now();

  @override
  void dispose() {
    openingCashController.dispose();
    noteController.dispose();
    super.dispose();
  }

  int? get _openingCash => int.tryParse(
    openingCashController.text.replaceAll(RegExp(r'[^0-9]'), ''),
  );

  Future<void> _submit() async {
    final openingCash = _openingCash;
    if (openingCash == null) return;
    final ok = await ref
        .read(nojposSessionProvider.notifier)
        .openShift(openingCash: openingCash);
    if (!mounted) return;
    if (ok) {
      Navigator.of(context).pop();
      return;
    }
    final error = ref.read(nojposSessionProvider).errorMessage;
    NojposToast.error(
      context,
      'Buka kasir ditolak',
      description: error ?? 'Periksa modal awal, outlet, dan PIN kasir.',
    );
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final openingCash = _openingCash;
    return Dialog(
      backgroundColor: Colors.white,
      insetPadding: const EdgeInsets.symmetric(horizontal: 120, vertical: 32),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(MokposRadius.lg),
      ),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 560),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Container(
              padding: const EdgeInsets.fromLTRB(28, 24, 28, 22),
              decoration: const BoxDecoration(
                color: MokposColors.primary,
                borderRadius: BorderRadius.vertical(
                  top: Radius.circular(MokposRadius.lg),
                ),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Container(
                        width: 48,
                        height: 48,
                        decoration: BoxDecoration(
                          color: Colors.white.withValues(alpha: 0.18),
                          borderRadius: BorderRadius.circular(MokposRadius.md),
                        ),
                        child: const Icon(
                          LucideIcons.badgeDollarSign,
                          color: Colors.white,
                        ),
                      ),
                      const SizedBox(width: 14),
                      const Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Buka Kasir',
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 24,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                            SizedBox(height: 4),
                            Text(
                              'Masukkan modal awal sebelum transaksi dimulai.',
                              style: TextStyle(
                                color: MokposColors.onPrimaryMuted,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(28, 24, 28, 12),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: MokposColors.canvas,
                      borderRadius: BorderRadius.circular(MokposRadius.md),
                      border: Border.all(color: MokposColors.line),
                    ),
                    child: Column(
                      children: [
                        _OpenShiftInfoLine(
                          icon: LucideIcons.userRound,
                          label: 'Kasir',
                          value: session.cashier.name,
                        ),
                        const Divider(height: 18, color: MokposColors.line),
                        _OpenShiftInfoLine(
                          icon: LucideIcons.store,
                          label: 'Outlet',
                          value: session.outlet.name,
                        ),
                        const Divider(height: 18, color: MokposColors.line),
                        _OpenShiftInfoLine(
                          icon: LucideIcons.clock3,
                          label: 'Jam buka',
                          value: _formatDialogTime(openedAt),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 18),
                  TextField(
                    key: const ValueKey('shift_opening_cash'),
                    controller: openingCashController,
                    keyboardType: TextInputType.number,
                    autofocus: true,
                    onChanged: (_) => setState(() {}),
                    style: const TextStyle(
                      fontSize: 22,
                      fontWeight: FontWeight.w900,
                    ),
                    decoration: InputDecoration(
                      labelText: 'Modal awal',
                      prefixText: 'Rp ',
                      helperText: openingCash == null
                          ? 'Isi uang tunai fisik yang ada di laci kasir.'
                          : 'Modal awal: ${rupiah(openingCash)}',
                      border: const OutlineInputBorder(),
                    ),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    key: const ValueKey('shift_open_note'),
                    controller: noteController,
                    maxLines: 2,
                    decoration: const InputDecoration(
                      labelText: 'Catatan opsional',
                      hintText: 'Contoh: modal dari brankas pagi',
                      border: OutlineInputBorder(),
                    ),
                  ),
                  if (session.errorMessage != null) ...[
                    const SizedBox(height: 14),
                    _InlineError(message: session.errorMessage!),
                  ],
                ],
              ),
            ),
            const Divider(height: 1, color: MokposColors.line),
            Padding(
              padding: const EdgeInsets.fromLTRB(28, 16, 28, 24),
              child: Row(
                children: [
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: session.isBusy
                          ? null
                          : () {
                              ref.read(nojposSessionProvider.notifier).lock();
                              Navigator.of(context).pop();
                              context.go('/pin');
                            },
                      icon: const Icon(LucideIcons.logOut),
                      label: const Text('Keluar Kasir'),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: FilledButton.icon(
                      key: const ValueKey('shift_open_submit'),
                      onPressed: session.isBusy || openingCash == null
                          ? null
                          : _submit,
                      style: FilledButton.styleFrom(
                        backgroundColor: MokposColors.primary,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(vertical: 14),
                      ),
                      icon: session.isBusy
                          ? const SizedBox.square(
                              dimension: 18,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(LucideIcons.lockOpen),
                      label: const Text('Buka Kasir'),
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

class _OpenShiftInfoLine extends StatelessWidget {
  const _OpenShiftInfoLine({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Icon(icon, size: 18, color: MokposColors.primary),
        const SizedBox(width: 10),
        Text(label, style: const TextStyle(color: MokposColors.muted)),
        const Spacer(),
        Flexible(
          child: Text(
            value.isEmpty ? '-' : value,
            textAlign: TextAlign.end,
            style: const TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ],
    );
  }
}

class _InlineError extends StatelessWidget {
  const _InlineError({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: MokposColors.danger.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(MokposRadius.sm),
        border: Border.all(color: MokposColors.danger.withValues(alpha: 0.18)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(LucideIcons.circleAlert, color: MokposColors.danger),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              message,
              style: const TextStyle(
                color: MokposColors.danger,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

String _formatDialogTime(DateTime time) {
  final local = time.toLocal();
  String two(int value) => value.toString().padLeft(2, '0');
  return '${two(local.day)}/${two(local.month)}/${local.year} ${two(local.hour)}:${two(local.minute)}';
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
            ? 'Laporan Tutup Kasir'
            : 'Tutup Kasir - ${session.cashier.name}',
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
              context.go('/pin');
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
                      NojposToast.error(
                        context,
                        'Tutup kasir ditolak',
                        description:
                            error ?? 'Periksa PIN dan nominal kas aktual.',
                      );
                    }
                  },
            child: session.isBusy
                ? const SizedBox.square(
                    dimension: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Tutup Kasir'),
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
  NojposToast.info(
    context,
    '$label nonaktif di terminal kasir',
    description:
        'Fitur ini tidak tersedia untuk workflow kasir preview. Gunakan menu operasional aktif.',
  );
}

Future<void> showOperationalFeatureDialog(
  BuildContext context,
  WidgetRef ref,
  String label,
) {
  return switch (label) {
    'Kas / Wallet' => showCashWalletDialog(context, ref),
    'Buku Menu' => showMenuBookDialog(context, ref),
    'Promo / Voucher' => showPromotionHelpDialog(context),
    'NojPOS Care' => showNojposCareDialog(context, ref),
    _ => showFutureFeatureDialog(context, label),
  };
}

Future<void> showFutureFeatureDialog(BuildContext context, String label) {
  return showNojposInfoDialog(
    context,
    title: label,
    subtitle:
        'Fitur ini nonaktif di terminal kasir preview. Tidak ada data contoh atau status palsu yang ditampilkan.',
    confirmLabel: 'Tutup',
  );
}

Future<void> showPromotionHelpDialog(BuildContext context) {
  return showNojposInfoDialog(
    context,
    title: 'Promo / Voucher',
    subtitle:
        'Kode promo digunakan dari layar pembayaran setelah cart siap checkout. Daftar voucher tidak ditampilkan di terminal kasir preview.',
    confirmLabel: 'Mengerti',
  );
}

Future<void> showNojposCareDialog(BuildContext context, WidgetRef ref) {
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => const _NojposCareDialog(),
  );
}

class _NojposCareDialog extends ConsumerWidget {
  const _NojposCareDialog();

  static const _appVersion = String.fromEnvironment(
    'APP_VERSION',
    defaultValue: 'Tidak tersedia',
  );
  static const _buildNumber = String.fromEnvironment(
    'BUILD_NUMBER',
    defaultValue: 'Tidak tersedia',
  );

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(nojposSessionProvider);
    final outletName = _displayValue(
      session.outlet.name,
      fallback: 'Belum dipilih',
    );
    final outletId = _displayValue(session.outlet.id);
    final deviceId = _displayValue(session.deviceId ?? session.deviceUuid);
    final cashierName = _displayValue(
      session.cashier.name,
      fallback: 'Belum pilih kasir',
    );
    final shiftStatus = session.activeShift?.isOpen == true
        ? 'Shift terbuka'
        : 'Tidak ada shift terbuka';
    final appStatus = session.status == SessionStatus.ready
        ? 'Siap operasional'
        : _sessionStatusLabel(session.status);

    return Dialog(
      backgroundColor: Colors.white,
      insetPadding: const EdgeInsets.symmetric(horizontal: 96, vertical: 40),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 840, maxHeight: 720),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: SingleChildScrollView(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  children: [
                    Container(
                      width: 48,
                      height: 48,
                      decoration: BoxDecoration(
                        color: MokposColors.primary.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(MokposRadius.xs),
                      ),
                      child: const Icon(
                        LucideIcons.lifeBuoy,
                        color: MokposColors.primary,
                      ),
                    ),
                    const SizedBox(width: 14),
                    const Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'NojPOS Care',
                            style: TextStyle(
                              fontSize: 22,
                              fontWeight: FontWeight.w900,
                            ),
                          ),
                          SizedBox(height: 4),
                          Text(
                            'Butuh bantuan operasional?',
                            style: TextStyle(color: MokposColors.muted),
                          ),
                        ],
                      ),
                    ),
                    IconButton(
                      tooltip: 'Tutup',
                      onPressed: () => Navigator.of(context).pop(),
                      icon: const Icon(LucideIcons.x),
                    ),
                  ],
                ),
                const SizedBox(height: 20),
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(18),
                  decoration: BoxDecoration(
                    color: MokposColors.surface,
                    borderRadius: BorderRadius.circular(MokposRadius.sm),
                    border: Border.all(color: MokposColors.line),
                  ),
                  child: const Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Hubungi administrator NOJPOS Anda',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w900,
                          color: MokposColors.text,
                        ),
                      ),
                      SizedBox(height: 8),
                      Text(
                        'Untuk bantuan transaksi, shift, perangkat, dan akses outlet, hubungi administrator internal bisnis Anda terlebih dahulu.',
                        style: TextStyle(
                          color: MokposColors.muted,
                          height: 1.45,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 18),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: _CareCard(
                        icon: LucideIcons.monitorCog,
                        title: 'Status aplikasi',
                        child: Column(
                          children: [
                            _CareInfoRow(
                              label: 'Status sesi',
                              value: appStatus,
                            ),
                            _CareInfoRow(label: 'Shift', value: shiftStatus),
                            const _CareInfoRow(
                              label: 'Versi aplikasi',
                              value: _appVersion,
                            ),
                            const _CareInfoRow(
                              label: 'Build',
                              value: _buildNumber,
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(width: 14),
                    Expanded(
                      child: _CareCard(
                        icon: LucideIcons.store,
                        title: 'Outlet & terminal',
                        child: Column(
                          children: [
                            _CareInfoRow(label: 'Outlet', value: outletName),
                            _CareInfoRow(label: 'Outlet ID', value: outletId),
                            _CareInfoRow(label: 'Device ID', value: deviceId),
                            _CareInfoRow(
                              label: 'Kasir aktif',
                              value: cashierName,
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 14),
                _CareCard(
                  icon: LucideIcons.clipboardCheck,
                  title: 'Diagnostic info',
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Salin ringkasan aman untuk dikirim ke administrator saat meminta bantuan. Tidak berisi token, PIN, password, atau detail pembayaran pelanggan.',
                        style: TextStyle(
                          color: MokposColors.muted,
                          height: 1.45,
                        ),
                      ),
                      const SizedBox(height: 12),
                      Align(
                        alignment: Alignment.centerLeft,
                        child: FilledButton.icon(
                          onPressed: () async {
                            await Clipboard.setData(
                              ClipboardData(text: _diagnosticText(session)),
                            );
                            if (!context.mounted) return;
                            NojposToast.success(
                              context,
                              'Diagnostic info disalin',
                              description:
                                  'Kirim ke administrator NOJPOS Anda bila diminta.',
                            );
                          },
                          icon: const Icon(LucideIcons.copy, size: 18),
                          label: const Text('Copy diagnostic info'),
                        ),
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
  }

  static String _diagnosticText(NojposSessionState session) {
    final shift = session.activeShift;
    return [
      'NOJPOS Care Diagnostic',
      'status=${_sessionStatusLabel(session.status)}',
      'app_version=$_appVersion',
      'build=$_buildNumber',
      'business=${_displayValue(session.businessName)}',
      'outlet_name=${_displayValue(session.outlet.name)}',
      'outlet_id=${_displayValue(session.outlet.id)}',
      'outlet_timezone=${_displayValue(session.outlet.timezone)}',
      'device_id=${_displayValue(session.deviceId)}',
      'device_uuid=${_displayValue(session.deviceUuid)}',
      'cashier_role=${_displayValue(session.cashier.role)}',
      'shift_id=${_displayValue(shift?.id)}',
      'shift_status=${shift?.isOpen == true ? 'open' : 'not_open'}',
    ].join('\n');
  }

  static String _sessionStatusLabel(SessionStatus status) => switch (status) {
    SessionStatus.booting => 'Memuat sesi',
    SessionStatus.unauthenticated => 'Belum login',
    SessionStatus.outletRequired => 'Perlu pilih outlet',
    SessionStatus.pinRequired => 'Perlu PIN kasir',
    SessionStatus.ready => 'Siap operasional',
  };

  static String _displayValue(
    String? value, {
    String fallback = 'Tidak tersedia',
  }) {
    final trimmed = value?.trim() ?? '';
    return trimmed.isEmpty ? fallback : trimmed;
  }
}

class _CareCard extends StatelessWidget {
  const _CareCard({
    required this.icon,
    required this.title,
    required this.child,
  });

  final IconData icon;
  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(MokposRadius.sm),
        border: Border.all(color: MokposColors.line),
        boxShadow: const [
          BoxShadow(
            color: Color(0x08000000),
            blurRadius: 12,
            offset: Offset(0, 6),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, size: 18, color: MokposColors.primary),
              const SizedBox(width: 8),
              Text(
                title,
                style: const TextStyle(
                  color: MokposColors.text,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          child,
        ],
      ),
    );
  }
}

class _CareInfoRow extends StatelessWidget {
  const _CareInfoRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 116,
            child: Text(
              label,
              style: const TextStyle(
                color: MokposColors.muted,
                fontSize: 12,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              value,
              textAlign: TextAlign.right,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

Future<void> showMenuBookDialog(BuildContext context, WidgetRef ref) {
  return showDialog<void>(
    context: context,
    barrierColor: Colors.black54,
    builder: (context) => const _MenuBookDialog(),
  );
}

class _MenuBookDialog extends ConsumerStatefulWidget {
  const _MenuBookDialog();

  @override
  ConsumerState<_MenuBookDialog> createState() => _MenuBookDialogState();
}

class _MenuBookDialogState extends ConsumerState<_MenuBookDialog> {
  final _searchController = TextEditingController();
  String _selectedCategory = 'Semua';
  String _query = '';

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _addToCart(Product product) {
    ref.read(cartProvider.notifier).add(product);
    NojposToast.success(context, '${product.name} ditambahkan ke cart');
  }

  void _showProductDetail(Product product) {
    showNojposInfoDialog(
      context,
      title: product.name,
      subtitle:
          '${product.category}\n${rupiah(product.price)}\n${product.badge ?? 'SKU/barcode belum tersedia'}\nStatus: Tersedia di katalog outlet',
      confirmLabel: 'Tutup',
    );
  }

  @override
  Widget build(BuildContext context) {
    final catalog = ref.watch(posCatalogProvider);
    final categories = _menuBookCategories(catalog.categoryItems);
    final products = _menuBookFilteredProducts(
      products: catalog.products,
      selectedCategory: _selectedCategory,
      query: _query,
    );

    return Dialog(
      backgroundColor: Colors.white,
      insetPadding: const EdgeInsets.symmetric(horizontal: 96, vertical: 40),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 920, maxHeight: 700),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    width: 44,
                    height: 44,
                    decoration: BoxDecoration(
                      color: MokposColors.primary.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(MokposRadius.xs),
                    ),
                    child: const Icon(
                      LucideIcons.bookOpen,
                      color: MokposColors.primary,
                    ),
                  ),
                  const SizedBox(width: 14),
                  const Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Buku Menu',
                          style: TextStyle(
                            fontSize: 22,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                        SizedBox(height: 4),
                        Text(
                          'Lihat katalog produk outlet, cek harga, lalu tambahkan item bila diperlukan.',
                          style: TextStyle(color: MokposColors.muted),
                        ),
                      ],
                    ),
                  ),
                  IconButton(
                    tooltip: 'Tutup',
                    onPressed: () => Navigator.of(context).pop(),
                    icon: const Icon(LucideIcons.x),
                  ),
                ],
              ),
              const SizedBox(height: 18),
              Row(
                children: [
                  Expanded(
                    child: TextField(
                      key: const ValueKey('menu_book_search_field'),
                      controller: _searchController,
                      textInputAction: TextInputAction.search,
                      onChanged: (value) => setState(() => _query = value),
                      decoration: InputDecoration(
                        prefixIcon: const Icon(LucideIcons.search),
                        hintText:
                            'Cari nama produk, kategori, SKU, atau barcode',
                        filled: true,
                        fillColor: MokposColors.surface,
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(MokposRadius.xs),
                          borderSide: const BorderSide(
                            color: MokposColors.line,
                          ),
                        ),
                        enabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(MokposRadius.xs),
                          borderSide: const BorderSide(
                            color: MokposColors.line,
                          ),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  OutlinedButton.icon(
                    key: const ValueKey('menu_book_refresh_button'),
                    onPressed: catalog.isLoading
                        ? null
                        : () => ref.read(posCatalogProvider.notifier).load(),
                    icon: const Icon(LucideIcons.refreshCw, size: 18),
                    label: const Text('Muat ulang'),
                  ),
                ],
              ),
              const SizedBox(height: 14),
              if (!catalog.isLoading && catalog.errorMessage == null)
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      for (final category in categories) ...[
                        ChoiceChip(
                          key: ValueKey('menu_book_category_$category'),
                          label: Text(category),
                          selected: _selectedCategory == category,
                          onSelected: (_) {
                            setState(() => _selectedCategory = category);
                          },
                        ),
                        const SizedBox(width: 8),
                      ],
                    ],
                  ),
                ),
              const SizedBox(height: 14),
              Expanded(
                child: catalog.isLoading
                    ? const _MenuBookLoadingState()
                    : catalog.errorMessage != null
                    ? _MenuBookErrorState(
                        message: catalog.errorMessage!,
                        onRetry: () =>
                            ref.read(posCatalogProvider.notifier).load(),
                      )
                    : products.isEmpty
                    ? _MenuBookEmptyState(
                        hasCatalog: catalog.products.isNotEmpty,
                        onClearSearch: () {
                          _searchController.clear();
                          setState(() {
                            _query = '';
                            _selectedCategory = 'Semua';
                          });
                        },
                      )
                    : ListView.separated(
                        key: const ValueKey('menu_book_product_list'),
                        itemCount: products.length,
                        separatorBuilder: (context, index) =>
                            const SizedBox(height: 10),
                        itemBuilder: (context, index) {
                          final product = products[index];
                          return _MenuBookProductTile(
                            product: product,
                            onAddToCart: () => _addToCart(product),
                            onViewDetail: () => _showProductDetail(product),
                          );
                        },
                      ),
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Text(
                    catalog.errorMessage == null
                        ? '${products.length} produk ditampilkan'
                        : 'Katalog belum tersedia',
                    style: const TextStyle(color: MokposColors.muted),
                  ),
                  const Spacer(),
                  TextButton(
                    onPressed: () => Navigator.of(context).pop(),
                    child: const Text('Tutup'),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

List<String> _menuBookCategories(List<ProductCategory> categories) {
  final values = <String>[];
  for (final category in categories) {
    if (category.name == 'Favorit') continue;
    if (!values.contains(category.name)) values.add(category.name);
  }
  if (!values.contains('Semua')) values.insert(0, 'Semua');
  return values;
}

List<Product> _menuBookFilteredProducts({
  required List<Product> products,
  required String selectedCategory,
  required String query,
}) {
  final normalizedQuery = query.trim().toLowerCase();
  return products.where((product) {
    final matchesCategory =
        selectedCategory == 'Semua' || product.category == selectedCategory;
    final matchesQuery =
        normalizedQuery.isEmpty ||
        product.name.toLowerCase().contains(normalizedQuery) ||
        product.category.toLowerCase().contains(normalizedQuery) ||
        product.id.toLowerCase().contains(normalizedQuery) ||
        (product.badge?.toLowerCase().contains(normalizedQuery) ?? false);
    return matchesCategory && matchesQuery;
  }).toList();
}

class _MenuBookProductTile extends StatelessWidget {
  const _MenuBookProductTile({
    required this.product,
    required this.onAddToCart,
    required this.onViewDetail,
  });

  final Product product;
  final VoidCallback onAddToCart;
  final VoidCallback onViewDetail;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(MokposRadius.xs),
      ),
      child: Row(
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              color: MokposColors.surface,
              borderRadius: BorderRadius.circular(MokposRadius.xs),
            ),
            child: const Icon(LucideIcons.utensils, color: MokposColors.muted),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        product.name,
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                    _MenuBookStatusBadge(available: product.price >= 0),
                  ],
                ),
                const SizedBox(height: 6),
                Wrap(
                  spacing: 8,
                  runSpacing: 6,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [
                    Text(
                      product.category,
                      style: const TextStyle(color: MokposColors.muted),
                    ),
                    if (product.badge != null)
                      Text(
                        product.badge!,
                        style: const TextStyle(color: MokposColors.muted),
                      ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(width: 12),
          Text(
            rupiah(product.price),
            style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w900),
          ),
          const SizedBox(width: 12),
          OutlinedButton(
            key: ValueKey('menu_book_detail_${product.id}'),
            onPressed: onViewDetail,
            child: const Text('Detail'),
          ),
          const SizedBox(width: 8),
          FilledButton.icon(
            key: ValueKey('menu_book_add_${product.id}'),
            onPressed: onAddToCart,
            icon: const Icon(LucideIcons.plus, size: 18),
            label: const Text('Tambah'),
          ),
        ],
      ),
    );
  }
}

class _MenuBookStatusBadge extends StatelessWidget {
  const _MenuBookStatusBadge({required this.available});

  final bool available;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: available
            ? MokposColors.success.withValues(alpha: 0.1)
            : MokposColors.danger.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        available ? 'Tersedia' : 'Tidak tersedia',
        style: TextStyle(
          color: available ? MokposColors.success : MokposColors.danger,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _MenuBookLoadingState extends StatelessWidget {
  const _MenuBookLoadingState();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          CircularProgressIndicator(),
          SizedBox(height: 12),
          Text('Mengambil katalog produk outlet...'),
        ],
      ),
    );
  }
}

class _MenuBookErrorState extends StatelessWidget {
  const _MenuBookErrorState({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(LucideIcons.wifiOff, color: MokposColors.danger, size: 36),
          const SizedBox(height: 12),
          const Text(
            'Katalog belum bisa dimuat',
            style: TextStyle(fontWeight: FontWeight.w900),
          ),
          const SizedBox(height: 6),
          Text(
            message,
            textAlign: TextAlign.center,
            style: const TextStyle(color: MokposColors.muted),
          ),
          const SizedBox(height: 14),
          FilledButton.icon(
            onPressed: onRetry,
            icon: const Icon(LucideIcons.refreshCw, size: 18),
            label: const Text('Coba lagi'),
          ),
        ],
      ),
    );
  }
}

class _MenuBookEmptyState extends StatelessWidget {
  const _MenuBookEmptyState({
    required this.hasCatalog,
    required this.onClearSearch,
  });

  final bool hasCatalog;
  final VoidCallback onClearSearch;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(LucideIcons.bookOpen, color: MokposColors.muted, size: 38),
          const SizedBox(height: 12),
          Text(
            hasCatalog
                ? 'Produk tidak ditemukan'
                : 'Katalog produk masih kosong',
            style: const TextStyle(fontWeight: FontWeight.w900),
          ),
          const SizedBox(height: 6),
          Text(
            hasCatalog
                ? 'Coba kata kunci lain atau pilih kategori Semua.'
                : 'Produk yang dibuat di backend akan muncul di Buku Menu.',
            textAlign: TextAlign.center,
            style: const TextStyle(color: MokposColors.muted),
          ),
          if (hasCatalog) ...[
            const SizedBox(height: 14),
            OutlinedButton(
              onPressed: onClearSearch,
              child: const Text('Reset filter'),
            ),
          ],
        ],
      ),
    );
  }
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
  NojposToast.success(
    context,
    '${type.label} tersimpan',
    description: 'Pengaturan diskon berhasil diterapkan ke keranjang.',
  );
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
  });

  final IconData icon;
  final String label;
  final bool active;
  final VoidCallback? onTap;
  final String? route;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      key: ValueKey(
        'drawer_${label.toLowerCase().replaceAll(RegExp(r'[^a-z0-9]+'), '_')}',
      ),
      onTap:
          onTap ??
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
        'Roadmap',
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
    this.roadmapState = false,
    this.onPressed,
    this.roadmapCopy,
  });

  final IconData icon;
  final String label;
  final bool active;
  final bool roadmapState;
  final VoidCallback? onPressed;
  final String? roadmapCopy;

  @override
  Widget build(BuildContext context) {
    final button = OutlinedButton(
      onPressed: roadmapState
          ? () => NojposToast.info(
              context,
              '$label nonaktif di terminal kasir',
              description:
                  roadmapCopy ??
                  'Mode ini tidak aktif untuk workflow kasir preview.',
            )
          : onPressed ?? () => Navigator.of(context).pop(),
      style: OutlinedButton.styleFrom(
        alignment: Alignment.centerLeft,
        foregroundColor: MokposColors.text,
        disabledForegroundColor: MokposColors.muted,
        backgroundColor: active
            ? MokposColors.primarySoft
            : roadmapState
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
          Icon(icon, size: 18),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: const TextStyle(fontWeight: FontWeight.w900),
                ),
                if (roadmapState && roadmapCopy != null) ...[
                  const SizedBox(height: 4),
                  Text(
                    roadmapCopy!,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: MokposColors.muted,
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ],
              ],
            ),
          ),
          if (roadmapState) ...[const SizedBox(width: 8), const _SoonBadge()],
        ],
      ),
    );
    if (!roadmapState) return button;
    return Tooltip(
      message: roadmapCopy ?? '$label nonaktif di terminal kasir',
      child: button,
    );
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
