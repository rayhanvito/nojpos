import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../../../app/providers/nojpos_session_provider.dart';
import '../../../shared/models/nojpos_models.dart';
import '../formatters.dart';
import '../models/cart_item.dart';
import '../models/product.dart';
import '../providers/pos_providers.dart';

enum OperationsPageType {
  orders,
  sales,
  reports,
  inventory,
  settings,
  attendance,
}

class OperationsScreen extends StatelessWidget {
  const OperationsScreen({required this.type, super.key});

  final OperationsPageType type;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      body: SafeArea(
        child: Column(
          children: [
            _OpsTopBar(config: _config(type)),
            Expanded(child: _bodyForType(type)),
          ],
        ),
      ),
    );
  }

  Widget _bodyForType(OperationsPageType type) {
    return switch (type) {
      OperationsPageType.orders => const _OrdersBody(),
      OperationsPageType.sales => const _SalesBody(),
      OperationsPageType.reports => const _ReportsBody(),
      OperationsPageType.inventory => const _InventoryBody(),
      OperationsPageType.settings => const _SettingsBody(),
      OperationsPageType.attendance => const _AttendanceBody(),
    };
  }

  _OpsConfig _config(OperationsPageType type) {
    return switch (type) {
      OperationsPageType.orders => const _OpsConfig(title: 'Daftar Order'),
      OperationsPageType.sales => const _OpsConfig(title: 'Penjualan'),
      OperationsPageType.reports => const _OpsConfig(
        title: 'Laporan - Ringkasan Penjualan',
      ),
      OperationsPageType.inventory => const _OpsConfig(
        title: 'Inventori - Faktur Pembelian',
        action: 'Tambah Faktur Pembelian',
      ),
      OperationsPageType.settings => const _OpsConfig(title: 'Pengaturan'),
      OperationsPageType.attendance => const _OpsConfig(
        title: 'Absensi',
        action: 'Daftar Kehadiran',
      ),
    };
  }
}

class _OpsConfig {
  const _OpsConfig({required this.title, this.action});

  final String title;
  final String? action;
}

class _OpsTopBar extends ConsumerWidget {
  const _OpsTopBar({required this.config});

  final _OpsConfig config;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Container(
      height: 56,
      color: MokposColors.primary,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      child: Row(
        children: [
          IconButton(
            onPressed: () => context.go('/pos'),
            icon: const Icon(LucideIcons.arrowLeft),
            color: Colors.white,
          ),
          const SizedBox(width: 2),
          Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                config.title,
                style: const TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.w900,
                  fontSize: 15,
                ),
              ),
              const SizedBox(height: 1),
              const Row(
                children: [
                  _OnlineDot(),
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
          if (config.action != null)
            InkWell(
              onTap: () => _handleAction(context, ref),
              borderRadius: BorderRadius.circular(MokposRadius.sm),
              child: Container(
                height: 42,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                decoration: BoxDecoration(
                  color: MokposColors.primaryDark,
                  borderRadius: BorderRadius.circular(MokposRadius.sm),
                ),
                alignment: Alignment.center,
                child: Text(
                  config.action!,
                  style: const TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
          const SizedBox(width: 10),
          const _TopSearch(),
        ],
      ),
    );
  }

  Future<void> _handleAction(BuildContext context, WidgetRef ref) async {
    if (config.title.startsWith('Inventori')) {
      final result = await showDialog<({String supplierName, int total})>(
        context: context,
        builder: (context) => const _PurchaseDialog(),
      );
      if (result == null || !context.mounted) return;
      final purchase = ref
          .read(nojposSessionProvider.notifier)
          .addPurchase(supplierName: result.supplierName, total: result.total);
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text('${purchase.number} ditambahkan')));
      return;
    }

    if (config.title == 'Absensi') {
      await showDialog<void>(
        context: context,
        builder: (context) => const _AttendanceListDialog(),
      );
    }
  }
}

class _OnlineDot extends StatelessWidget {
  const _OnlineDot();

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

class _TopSearch extends StatelessWidget {
  const _TopSearch();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 210,
      height: 40,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: MokposColors.primaryDark,
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      child: const Row(
        children: [
          Icon(LucideIcons.search, color: Colors.white, size: 18),
          SizedBox(width: 8),
          Text(
            'Cari...',
            style: TextStyle(color: Colors.white, fontWeight: FontWeight.w700),
          ),
        ],
      ),
    );
  }
}

class _OrdersBody extends ConsumerWidget {
  const _OrdersBody();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final orders = ref.watch(nojposSessionProvider).savedOrders;
    return Row(
      children: [
        _LeftList(
          title: 'Kategori Order',
          items: [
            'Semua (${orders.length})',
            'Kasir (${orders.length})',
            'Order Online (0)',
            'Faktur Penjualan (0)',
          ],
          activeIndex: 0,
        ),
        Expanded(
          child: Column(
            children: [
              const _TableHeader(
                columns: [
                  'Jenis',
                  'Nama',
                  'No Transaksi',
                  'Tanggal',
                  'Status',
                  'Tagihan',
                ],
              ),
              Expanded(
                child: orders.isEmpty
                    ? const _EmptyState(
                        icon: LucideIcons.receiptText,
                        title: 'Belum Ada Pesanan',
                        subtitle: 'Saatnya tingkatkan promosimu!',
                      )
                    : ListView.separated(
                        itemCount: orders.length,
                        separatorBuilder: (context, index) =>
                            const Divider(height: 1, color: MokposColors.line),
                        itemBuilder: (context, index) {
                          final order = orders[index];
                          return _DataRow(
                            onTap: () {
                              final activated = ref
                                  .read(nojposSessionProvider.notifier)
                                  .activateSavedOrder(order.id);
                              if (activated == null) return;
                              final products = ref.read(productsProvider);
                              ref
                                  .read(cartProvider.notifier)
                                  .replaceWith(
                                    _cartItemsFromOrder(activated, products),
                                  );
                              ScaffoldMessenger.of(context).showSnackBar(
                                SnackBar(
                                  content: Text(
                                    '${activated.number} dibuka ke cart',
                                  ),
                                ),
                              );
                              context.go('/pos');
                            },
                            cells: [
                              order.type.label,
                              order.customer?.name ?? 'Tanpa Pelanggan',
                              order.number,
                              _formatDateTime(order.createdAt),
                              _statusLabel(order.status),
                              rupiah(order.total),
                            ],
                          );
                        },
                      ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _SalesBody extends ConsumerWidget {
  const _SalesBody();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final transactions = ref.watch(nojposSessionProvider).transactions;
    final totalSales = transactions.fold(
      0,
      (sum, transaction) => sum + transaction.order.total,
    );
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              Expanded(
                child: _MetricCard(
                  label: 'Total Penjualan',
                  value: rupiah(totalSales),
                  color: MokposColors.primary,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: _MetricCard(
                  label: 'Total Transaksi',
                  value: '${transactions.length}',
                  color: MokposColors.accent,
                ),
              ),
            ],
          ),
        ),
        const _TableHeader(
          columns: [
            'Jenis',
            'Nama',
            'No Transaksi',
            'Waktu',
            'Kasir',
            'Pembayaran',
          ],
        ),
        Expanded(
          child: transactions.isEmpty
              ? const _EmptyState(
                  icon: LucideIcons.badgeDollarSign,
                  title: 'Belum Ada Penjualan',
                  subtitle:
                      'Transaksi yang berhasil dibayar akan muncul di sini.',
                )
              : ListView.separated(
                  itemCount: transactions.length,
                  separatorBuilder: (context, index) =>
                      const Divider(height: 1, color: MokposColors.line),
                  itemBuilder: (context, index) {
                    final transaction = transactions[index];
                    final order = transaction.order;
                    final payment = transaction.payments.isEmpty
                        ? '-'
                        : transaction.payments.first.method.label;
                    return _DataRow(
                      cells: [
                        order.type.label,
                        order.customer?.name ?? 'Tanpa Pelanggan',
                        transaction.number,
                        _formatDateTime(transaction.createdAt),
                        transaction.cashier.name,
                        '${rupiah(order.total)}\n$payment',
                      ],
                    );
                  },
                ),
        ),
      ],
    );
  }
}

class _ReportsBody extends ConsumerWidget {
  const _ReportsBody();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final transactions = ref.watch(nojposSessionProvider).transactions;
    final totalSales = transactions.fold(
      0,
      (sum, transaction) => sum + transaction.order.total,
    );
    final productCount = transactions.fold(
      0,
      (sum, transaction) =>
          sum +
          transaction.order.lines.fold(
            0,
            (lineSum, line) => lineSum + line.quantity,
          ),
    );
    return Row(
      children: [
        const _LeftList(
          title: 'Kategori Laporan',
          items: [
            'Ringkasan Penjualan',
            '10 Laporan Teratas',
            'Komisi',
            'Void',
            'Kasir',
            'Kas Kasir',
            'Produk Terjual',
            'Jenis Bayar',
            'Kepuasan Pelanggan',
            'Lainnya',
            'Penjualan Deposit',
          ],
          activeIndex: 0,
        ),
        Expanded(
          child: Padding(
            padding: const EdgeInsets.all(18),
            child: Column(
              children: [
                const _FilterRow(),
                const SizedBox(height: 22),
                Expanded(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      SizedBox(
                        width: 450,
                        child: GridView.count(
                          crossAxisCount: 2,
                          mainAxisSpacing: 12,
                          crossAxisSpacing: 12,
                          childAspectRatio: 2.45,
                          children: [
                            _MetricCard(
                              label: 'Total Penjualan',
                              value: rupiah(totalSales),
                              color: MokposColors.primary,
                            ),
                            _MetricCard(
                              label: 'Total Transaksi',
                              value: '${transactions.length}',
                              color: const Color(0xFFFF4FC3),
                            ),
                            _MetricCard(
                              label: 'Laba Kotor',
                              value: rupiah(totalSales),
                              color: const Color(0xFF7C4DFF),
                            ),
                            _MetricCard(
                              label: 'Total Produk',
                              value: '$productCount',
                              color: const Color(0xFFFFD43B),
                            ),
                            _MetricCard(
                              label: 'Terima Pembayaran',
                              value: rupiah(totalSales),
                              color: const Color(0xFF4FC3F7),
                            ),
                            _MetricCard(
                              label: 'Refund',
                              value: 'Rp 0',
                              color: MokposColors.danger,
                            ),
                            _MetricCard(
                              label: 'Order/Transaksi',
                              value: transactions.isEmpty
                                  ? 'Rp 0'
                                  : rupiah(totalSales ~/ transactions.length),
                              color: MokposColors.accent,
                            ),
                            _MetricCard(
                              label: 'Komisi',
                              value: 'Rp 0',
                              color: const Color(0xFFFF7043),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 50),
                      const Expanded(child: _SimpleChart()),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}

class _InventoryBody extends ConsumerWidget {
  const _InventoryBody();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final purchases = ref.watch(nojposSessionProvider).purchases;
    return Row(
      children: [
        const _LeftList(
          title: 'Kategori Inventori',
          items: [
            'Faktur Pembelian',
            'Stok Opname',
            'Terima Mutasi Stok',
            'Stok Terbuang',
          ],
          activeIndex: 0,
        ),
        Expanded(
          child: Column(
            children: [
              const _FilterRow(),
              const _TableHeader(
                columns: ['Supplier', 'No Faktur', 'Tanggal', 'Total'],
              ),
              Expanded(
                child: purchases.isEmpty
                    ? const _EmptyState(
                        icon: LucideIcons.folderSearch,
                        title: 'Faktur Pembelian Tidak Tersedia',
                        subtitle:
                            'Silakan masukkan faktur pembelian yang dimiliki melalui tombol Tambah Faktur Pembelian terlebih dahulu',
                      )
                    : ListView.separated(
                        itemCount: purchases.length,
                        separatorBuilder: (context, index) =>
                            const Divider(height: 1, color: MokposColors.line),
                        itemBuilder: (context, index) {
                          final purchase = purchases[index];
                          return _DataRow(
                            cells: [
                              purchase.supplierName,
                              purchase.number,
                              _formatDateTime(purchase.createdAt),
                              rupiah(purchase.total),
                            ],
                          );
                        },
                      ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _SettingsBody extends ConsumerWidget {
  const _SettingsBody();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final catalog = ref.watch(posCatalogProvider);
    return Row(
      children: [
        const _SettingsMenu(),
        Expanded(
          child: Column(
            children: [
              Container(
                height: 56,
                decoration: const BoxDecoration(
                  border: Border(bottom: BorderSide(color: MokposColors.line)),
                ),
                child: const Row(
                  children: [
                    Expanded(
                      child: Center(
                        child: _TabText(text: 'Produk', active: true),
                      ),
                    ),
                    Expanded(
                      child: Center(child: _TabText(text: 'Kategori')),
                    ),
                  ],
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  children: [
                    const _SearchLine(),
                    const SizedBox(height: 16),
                    const _AddProductButton(),
                    const SizedBox(height: 14),
                    _CategoryStrip(categories: catalog.categories),
                    const SizedBox(height: 14),
                    SizedBox(
                      height: 420,
                      child: ListView.separated(
                        itemCount: catalog.products.length,
                        separatorBuilder: (context, index) =>
                            const Divider(height: 1, color: MokposColors.line),
                        itemBuilder: (context, index) => _ProductSettingRow(
                          product: catalog.products[index],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _AttendanceBody extends ConsumerWidget {
  const _AttendanceBody();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final now = DateTime.now();
    final attendance = ref.watch(nojposSessionProvider).attendance;
    final latest = attendance.isEmpty ? null : attendance.first;
    return Row(
      children: [
        Expanded(
          child: ColoredBox(
            color: const Color(0xFF3D4543),
            child: Padding(
              padding: const EdgeInsets.all(18),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '${now.hour.toString().padLeft(2, '0')}:${now.minute.toString().padLeft(2, '0')}',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 34,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    _formatDateTime(now),
                    style: const TextStyle(color: Colors.white70),
                  ),
                  const Spacer(),
                  Center(
                    child: Text(
                      latest == null
                          ? 'Belum ada aktivitas absensi hari ini'
                          : latest.isOpen
                          ? '${latest.employee.name} sedang clock in'
                          : '${latest.employee.name} sudah clock out',
                      textAlign: TextAlign.center,
                      style: const TextStyle(color: Colors.white70),
                    ),
                  ),
                  const Spacer(),
                ],
              ),
            ),
          ),
        ),
        const Expanded(child: _AttendancePinPanel()),
      ],
    );
  }
}

String _formatDateTime(DateTime value) {
  const months = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'Mei',
    'Jun',
    'Jul',
    'Agu',
    'Sep',
    'Okt',
    'Nov',
    'Des',
  ];
  final day = value.day.toString().padLeft(2, '0');
  final month = months[value.month - 1];
  final hour = value.hour.toString().padLeft(2, '0');
  final minute = value.minute.toString().padLeft(2, '0');
  return '$day $month ${value.year}, $hour:$minute';
}

String _statusLabel(OrderStatus status) {
  return switch (status) {
    OrderStatus.active => 'Aktif',
    OrderStatus.saved => 'Disimpan',
    OrderStatus.paid => 'Lunas',
    OrderStatus.canceled => 'Batal',
  };
}

List<CartItem> _cartItemsFromOrder(SalesOrder order, List<Product> products) {
  return [
    for (final line in order.lines)
      CartItem(
        product: products.firstWhere(
          (product) => product.id == line.productId,
          orElse: () => Product(
            id: line.productId,
            name: line.name,
            category: 'Lainnya',
            price: line.unitPrice,
            imageUrl: '',
          ),
        ),
        quantity: line.quantity,
      ),
  ];
}

class _LeftList extends StatelessWidget {
  const _LeftList({
    required this.title,
    required this.items,
    required this.activeIndex,
  });

  final String title;
  final List<String> items;
  final int activeIndex;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 190,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(right: BorderSide(color: MokposColors.line)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Container(
            height: 52,
            padding: const EdgeInsets.symmetric(horizontal: 10),
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: MokposColors.line)),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    title,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: MokposColors.muted,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ),
                const Icon(LucideIcons.listFilter, size: 16),
              ],
            ),
          ),
          for (final (index, item) in items.indexed)
            Container(
              height: 48,
              padding: const EdgeInsets.symmetric(horizontal: 10),
              alignment: Alignment.centerLeft,
              decoration: BoxDecoration(
                color: index == activeIndex
                    ? MokposColors.primarySoft
                    : Colors.white,
                border: Border(
                  left: BorderSide(
                    color: index == activeIndex
                        ? MokposColors.primary
                        : Colors.transparent,
                    width: 3,
                  ),
                  bottom: const BorderSide(color: MokposColors.line),
                ),
              ),
              child: Text(
                item,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  color: index == activeIndex
                      ? MokposColors.text
                      : MokposColors.muted,
                  fontSize: 12,
                  fontWeight: index == activeIndex
                      ? FontWeight.w900
                      : FontWeight.w700,
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _TableHeader extends StatelessWidget {
  const _TableHeader({required this.columns});

  final List<String> columns;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 48,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: Row(
        children: [
          for (final column in columns)
            Expanded(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                child: Text(
                  column,
                  style: const TextStyle(
                    color: MokposColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _DataRow extends StatelessWidget {
  const _DataRow({required this.cells, this.onTap});

  final List<String> cells;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: SizedBox(
        height: 76,
        child: Row(
          children: [
            for (final cell in cells)
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  child: Text(
                    cell,
                    style: const TextStyle(
                      color: MokposColors.text,
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _EmptyState extends StatelessWidget {
  const _EmptyState({
    required this.icon,
    required this.title,
    required this.subtitle,
  });

  final IconData icon;
  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 92, color: MokposColors.primary),
          const SizedBox(height: 18),
          Text(
            title,
            style: const TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
              fontSize: 16,
            ),
          ),
          const SizedBox(height: 8),
          SizedBox(
            width: 430,
            child: Text(
              subtitle,
              textAlign: TextAlign.center,
              style: const TextStyle(color: MokposColors.muted, height: 1.4),
            ),
          ),
        ],
      ),
    );
  }
}

class _MetricCard extends StatelessWidget {
  const _MetricCard({
    required this.label,
    required this.value,
    required this.color,
  });

  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 70,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: MokposColors.line),
        boxShadow: const [
          BoxShadow(
            color: Color(0x0F000000),
            blurRadius: 12,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: const TextStyle(color: MokposColors.muted, fontSize: 11),
          ),
          const Spacer(),
          Text(
            value,
            style: const TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
              fontSize: 16,
            ),
          ),
          const SizedBox(height: 6),
          Container(height: 3, color: color),
        ],
      ),
    );
  }
}

class _FilterRow extends StatelessWidget {
  const _FilterRow();

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(14),
      child: Row(
        children: [
          _FilterBox(text: 'Hari Ini'),
          const SizedBox(width: 14),
          _FilterBox(text: '16 Jun 2026 - 16 Jun 2026', width: 260),
          const SizedBox(width: 14),
          _FilterBox(text: 'Semua Status', width: 180),
        ],
      ),
    );
  }
}

class _FilterBox extends StatelessWidget {
  const _FilterBox({required this.text, this.width = 160});

  final String text;
  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(6),
      ),
      child: Row(
        children: [
          Expanded(
            child: Text(
              text,
              style: const TextStyle(color: MokposColors.muted, fontSize: 12),
            ),
          ),
          const Icon(
            LucideIcons.chevronDown,
            size: 16,
            color: MokposColors.muted,
          ),
        ],
      ),
    );
  }
}

class _SimpleChart extends StatelessWidget {
  const _SimpleChart();

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 280,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          for (final height in [110.0, 175.0, 112.0])
            Padding(
              padding: const EdgeInsets.only(right: 12),
              child: Container(
                width: 14,
                height: height,
                color: MokposColors.primary,
              ),
            ),
          const Expanded(
            child: Align(
              alignment: Alignment.bottomLeft,
              child: Text(
                '00:00      08:00      16:00      24:00',
                style: TextStyle(color: MokposColors.muted, fontSize: 11),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _SettingsMenu extends StatelessWidget {
  const _SettingsMenu();

  @override
  Widget build(BuildContext context) {
    final items = [
      'Produk & Kategori',
      'Order Online',
      'Struk & Biaya',
      'Pajak',
      'Kasir',
      'Perangkat',
      'Notifikasi Suara',
      'Pembayaran Nontunai',
      'Promo',
      'Refund',
      'Keluar',
    ];
    return Container(
      width: 330,
      decoration: const BoxDecoration(
        border: Border(right: BorderSide(color: MokposColors.line)),
      ),
      child: ListView(
        children: [
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 18, 16, 8),
            child: Text(
              'OPERASIONAL',
              style: TextStyle(
                color: MokposColors.muted,
                fontWeight: FontWeight.w900,
                fontSize: 12,
              ),
            ),
          ),
          for (final (index, item) in items.indexed)
            ListTile(
              dense: true,
              leading: Icon(
                index == 0 ? LucideIcons.boxes : LucideIcons.settings,
                size: 18,
                color: index == 0 ? MokposColors.primary : MokposColors.muted,
              ),
              title: Text(
                item,
                style: TextStyle(
                  color: index == 0 ? MokposColors.primary : MokposColors.text,
                  fontWeight: FontWeight.w800,
                  fontSize: 13,
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _TabText extends StatelessWidget {
  const _TabText({required this.text, this.active = false});

  final String text;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 56,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        border: Border(
          bottom: BorderSide(
            color: active ? MokposColors.warning : Colors.transparent,
            width: 4,
          ),
        ),
      ),
      child: Text(
        text,
        style: const TextStyle(
          color: Colors.white,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _SearchLine extends StatelessWidget {
  const _SearchLine();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: const Row(
        children: [
          Icon(LucideIcons.search, color: MokposColors.muted, size: 18),
          SizedBox(width: 12),
          Text('Cari ...', style: TextStyle(color: MokposColors.muted)),
          Spacer(),
          Icon(LucideIcons.listFilter, color: MokposColors.muted, size: 18),
        ],
      ),
    );
  }
}

class _AddProductButton extends ConsumerWidget {
  const _AddProductButton();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return OutlinedButton(
      onPressed: () async {
        final result =
            await showDialog<({String name, String category, int price})>(
              context: context,
              builder: (context) => const _ProductDialog(),
            );
        if (result == null || !context.mounted) return;
        final product = ref
            .read(posCatalogProvider.notifier)
            .addProduct(
              name: result.name,
              category: result.category,
              price: result.price,
            );
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('${product.name} ditambahkan')));
      },
      style: OutlinedButton.styleFrom(
        minimumSize: const Size.fromHeight(46),
        foregroundColor: MokposColors.primary,
        side: const BorderSide(color: MokposColors.primary),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
      ),
      child: const Text(
        'Tambah Produk',
        style: TextStyle(fontWeight: FontWeight.w900),
      ),
    );
  }
}

class _CategoryStrip extends StatelessWidget {
  const _CategoryStrip({required this.categories});

  final List<String> categories;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 38,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        itemCount: categories.length,
        separatorBuilder: (context, index) => const SizedBox(width: 8),
        itemBuilder: (context, index) {
          final category = categories[index];
          return Container(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: index == 0 ? MokposColors.primarySoft : Colors.white,
              border: Border.all(color: MokposColors.line),
              borderRadius: BorderRadius.circular(6),
            ),
            child: Text(
              category,
              style: TextStyle(
                color: index == 0 ? MokposColors.primary : MokposColors.text,
                fontWeight: FontWeight.w800,
                fontSize: 12,
              ),
            ),
          );
        },
      ),
    );
  }
}

class _ProductSettingRow extends StatelessWidget {
  const _ProductSettingRow({required this.product});

  final Product product;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: CircleAvatar(
        backgroundColor: const Color(0xFFE4EEE7),
        child: Text(product.name.characters.first.toUpperCase()),
      ),
      title: Text(
        product.name,
        style: const TextStyle(fontWeight: FontWeight.w900),
      ),
      subtitle: Text(product.category),
      trailing: Text(
        '${rupiah(product.price)}\n0',
        textAlign: TextAlign.right,
        style: const TextStyle(fontWeight: FontWeight.w800),
      ),
    );
  }
}

class _ProductDialog extends ConsumerStatefulWidget {
  const _ProductDialog();

  @override
  ConsumerState<_ProductDialog> createState() => _ProductDialogState();
}

class _ProductDialogState extends ConsumerState<_ProductDialog> {
  final nameController = TextEditingController(text: 'Produk Baru');
  final priceController = TextEditingController(text: '15000');
  String category = 'Makanan';

  @override
  void dispose() {
    nameController.dispose();
    priceController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final categories = ref
        .watch(posCatalogProvider)
        .categories
        .where((item) => item != 'Semua' && item != 'Favorit')
        .toList();
    if (!categories.contains(category) && categories.isNotEmpty) {
      category = categories.first;
    }

    return AlertDialog(
      backgroundColor: Colors.white,
      title: const Text(
        'Tambah Produk',
        style: TextStyle(fontWeight: FontWeight.w900),
      ),
      content: SizedBox(
        width: 420,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: nameController,
              autofocus: true,
              decoration: const InputDecoration(
                labelText: 'Nama Produk',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              initialValue: category,
              items: [
                for (final item in categories)
                  DropdownMenuItem(value: item, child: Text(item)),
              ],
              onChanged: (value) {
                if (value != null) setState(() => category = value);
              },
              decoration: const InputDecoration(
                labelText: 'Kategori',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: priceController,
              keyboardType: TextInputType.number,
              decoration: const InputDecoration(
                labelText: 'Harga',
                prefixText: 'Rp ',
                border: OutlineInputBorder(),
              ),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Batal'),
        ),
        FilledButton(
          onPressed: () {
            final price =
                int.tryParse(
                  priceController.text.replaceAll(RegExp(r'[^0-9]'), ''),
                ) ??
                0;
            if (nameController.text.trim().isEmpty || price <= 0) return;
            Navigator.of(context).pop((
              name: nameController.text.trim(),
              category: category,
              price: price,
            ));
          },
          child: const Text('Simpan'),
        ),
      ],
    );
  }
}

class _PurchaseDialog extends StatefulWidget {
  const _PurchaseDialog();

  @override
  State<_PurchaseDialog> createState() => _PurchaseDialogState();
}

class _PurchaseDialogState extends State<_PurchaseDialog> {
  final supplierController = TextEditingController(text: 'Supplier Nusantara');
  final totalController = TextEditingController(text: '250000');

  @override
  void dispose() {
    supplierController.dispose();
    totalController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      backgroundColor: Colors.white,
      title: const Text(
        'Tambah Faktur Pembelian',
        style: TextStyle(fontWeight: FontWeight.w900),
      ),
      content: SizedBox(
        width: 420,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: supplierController,
              autofocus: true,
              decoration: const InputDecoration(
                labelText: 'Supplier',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: totalController,
              keyboardType: TextInputType.number,
              decoration: const InputDecoration(
                labelText: 'Total Faktur',
                prefixText: 'Rp ',
                border: OutlineInputBorder(),
              ),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Batal'),
        ),
        FilledButton(
          onPressed: () {
            final total =
                int.tryParse(
                  totalController.text.replaceAll(RegExp(r'[^0-9]'), ''),
                ) ??
                0;
            if (supplierController.text.trim().isEmpty || total <= 0) return;
            Navigator.of(
              context,
            ).pop((supplierName: supplierController.text.trim(), total: total));
          },
          child: const Text('Tambah'),
        ),
      ],
    );
  }
}

class _AttendanceListDialog extends ConsumerWidget {
  const _AttendanceListDialog();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final records = ref.watch(nojposSessionProvider).attendance;
    return AlertDialog(
      backgroundColor: Colors.white,
      title: const Text(
        'Daftar Kehadiran',
        style: TextStyle(fontWeight: FontWeight.w900),
      ),
      content: SizedBox(
        width: 620,
        height: 360,
        child: records.isEmpty
            ? const Center(
                child: Text(
                  'Belum ada data kehadiran',
                  style: TextStyle(color: MokposColors.muted),
                ),
              )
            : ListView.separated(
                itemCount: records.length,
                separatorBuilder: (context, index) =>
                    const Divider(height: 1, color: MokposColors.line),
                itemBuilder: (context, index) {
                  final record = records[index];
                  return ListTile(
                    leading: Icon(
                      record.isOpen
                          ? LucideIcons.clock3
                          : LucideIcons.circleCheck,
                      color: record.isOpen
                          ? MokposColors.primary
                          : MokposColors.accent,
                    ),
                    title: Text(
                      record.employee.name,
                      style: const TextStyle(fontWeight: FontWeight.w900),
                    ),
                    subtitle: Text(
                      'Masuk ${_formatDateTime(record.clockInAt)}'
                      '${record.clockOutAt == null ? '' : '\nKeluar ${_formatDateTime(record.clockOutAt!)}'}',
                    ),
                    trailing: Text(record.isOpen ? 'Aktif' : 'Selesai'),
                  );
                },
              ),
      ),
      actions: [
        FilledButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Tutup'),
        ),
      ],
    );
  }
}

class _AttendancePinPanel extends ConsumerWidget {
  const _AttendancePinPanel();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(nojposSessionProvider);
    final employee = session.cashier;
    final openRecord = session.attendance.where(
      (record) => record.employee.id == employee.id && record.isOpen,
    );
    final isClockedIn = openRecord.isNotEmpty;
    return Padding(
      padding: const EdgeInsets.fromLTRB(72, 34, 72, 34),
      child: Column(
        children: [
          _FilterBox(text: employee.name, width: 450),
          const SizedBox(height: 34),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: List.generate(
              6,
              (_) => Container(
                width: 10,
                height: 10,
                margin: const EdgeInsets.symmetric(horizontal: 10),
                decoration: const BoxDecoration(
                  color: MokposColors.line,
                  shape: BoxShape.circle,
                ),
              ),
            ),
          ),
          const SizedBox(height: 20),
          const Text(
            'Masukkan PIN',
            style: TextStyle(color: MokposColors.muted),
          ),
          const SizedBox(height: 44),
          Expanded(
            child: GridView.count(
              crossAxisCount: 3,
              childAspectRatio: 1.6,
              children: const [
                Center(child: Text('1')),
                Center(child: Text('2')),
                Center(child: Text('3')),
                Center(child: Text('4')),
                Center(child: Text('5')),
                Center(child: Text('6')),
                Center(child: Text('7')),
                Center(child: Text('8')),
                Center(child: Text('9')),
                Center(child: Text('C')),
                Center(child: Text('0')),
                Center(child: Icon(LucideIcons.delete, size: 16)),
              ],
            ),
          ),
          FilledButton(
            onPressed: () {
              final record = ref
                  .read(nojposSessionProvider.notifier)
                  .toggleAttendance(employee);
              ScaffoldMessenger.of(context).showSnackBar(
                SnackBar(
                  content: Text(
                    record.isOpen ? 'Clock in berhasil' : 'Clock out berhasil',
                  ),
                ),
              );
            },
            style: FilledButton.styleFrom(
              minimumSize: const Size.fromHeight(48),
              backgroundColor: MokposColors.primary,
              foregroundColor: Colors.white,
            ),
            child: Text(isClockedIn ? 'Clock Out' : 'Clock In'),
          ),
        ],
      ),
    );
  }
}
