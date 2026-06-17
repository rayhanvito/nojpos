import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../formatters.dart';
import '../models/product.dart';
import '../providers/pos_providers.dart';

class ProductCatalog extends ConsumerWidget {
  const ProductCatalog({required this.compact, super.key});

  final bool compact;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final products = ref.watch(filteredProductsProvider);

    return Column(
      children: [
        _CatalogToolbar(compact: compact),
        Expanded(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(8, 8, 8, 8),
            child: products.isEmpty
                ? const _NoProductsFound()
                : GridView.builder(
                    itemCount: products.length,
                    gridDelegate: SliverGridDelegateWithMaxCrossAxisExtent(
                      maxCrossAxisExtent: compact ? 142 : 158,
                      mainAxisSpacing: 8,
                      crossAxisSpacing: 8,
                      childAspectRatio: compact ? .8 : .82,
                    ),
                    itemBuilder: (context, index) => ProductCard(
                      product: products[index],
                      onTap: () =>
                          ref.read(cartProvider.notifier).add(products[index]),
                    ),
                  ),
          ),
        ),
      ],
    );
  }
}

class _NoProductsFound extends StatelessWidget {
  const _NoProductsFound();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(LucideIcons.searchX, color: MokposColors.primary, size: 72),
          SizedBox(height: 14),
          Text(
            'Produk tidak ditemukan',
            style: TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
              fontSize: 16,
            ),
          ),
        ],
      ),
    );
  }
}

class _CatalogToolbar extends ConsumerWidget {
  const _CatalogToolbar({required this.compact});

  final bool compact;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Container(
      height: 54,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 14),
      child: Row(
        children: [
          Expanded(
            child: SizedBox(
              height: 44,
              child: TextField(
                onChanged: (value) =>
                    ref.read(productSearchQueryProvider.notifier).update(value),
                decoration: const InputDecoration(
                  prefixIcon: Icon(
                    LucideIcons.search,
                    color: MokposColors.muted,
                    size: 22,
                  ),
                  hintText: 'Cari produk, SKU, atau barcode',
                  hintStyle: TextStyle(color: MokposColors.muted, fontSize: 15),
                  border: InputBorder.none,
                  contentPadding: EdgeInsets.symmetric(vertical: 12),
                ),
              ),
            ),
          ),
          if (!compact) ...[
            const SizedBox(width: 14),
            _ToolbarAction(
              icon: LucideIcons.circlePlus,
              text: 'Custom Amount',
              onTap: () => _showCustomAmountDialog(context, ref),
            ),
            const SizedBox(width: 10),
            _IconAction(
              icon: LucideIcons.scanBarcode,
              onTap: () {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('Scanner barcode aktif sebagai mock UI'),
                  ),
                );
              },
            ),
          ],
        ],
      ),
    );
  }

  Future<void> _showCustomAmountDialog(
    BuildContext context,
    WidgetRef ref,
  ) async {
    final result = await showDialog<int>(
      context: context,
      builder: (context) => const _CustomAmountDialog(),
    );
    if (result == null || result <= 0) return;
    final product = Product(
      id: 'custom-${DateTime.now().microsecondsSinceEpoch}',
      name: 'Custom Amount',
      category: 'Lainnya',
      price: result,
      imageUrl: '',
      badge: 'Custom',
    );
    ref.read(cartProvider.notifier).add(product);
  }
}

class _ToolbarAction extends StatelessWidget {
  const _ToolbarAction({required this.icon, required this.text, this.onTap});

  final IconData icon;
  final String text;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(MokposRadius.sm),
      child: Container(
        height: 44,
        padding: const EdgeInsets.symmetric(horizontal: 14),
        decoration: BoxDecoration(
          border: Border.all(color: MokposColors.line),
          borderRadius: BorderRadius.circular(MokposRadius.sm),
        ),
        child: Row(
          children: [
            Icon(icon, color: MokposColors.text, size: 20),
            const SizedBox(width: 8),
            Text(
              text,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _IconAction extends StatelessWidget {
  const _IconAction({required this.icon, this.onTap});

  final IconData icon;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(MokposRadius.sm),
      child: Container(
        width: 44,
        height: 44,
        decoration: BoxDecoration(
          border: Border.all(color: MokposColors.line),
          borderRadius: BorderRadius.circular(MokposRadius.sm),
        ),
        child: Icon(icon, color: MokposColors.text, size: 21),
      ),
    );
  }
}

class _CustomAmountDialog extends StatefulWidget {
  const _CustomAmountDialog();

  @override
  State<_CustomAmountDialog> createState() => _CustomAmountDialogState();
}

class _CustomAmountDialogState extends State<_CustomAmountDialog> {
  final controller = TextEditingController(text: '10000');

  @override
  void dispose() {
    controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      backgroundColor: Colors.white,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(MokposRadius.sm),
      ),
      title: const Text(
        'Custom Amount',
        style: TextStyle(fontWeight: FontWeight.w900),
      ),
      content: TextField(
        controller: controller,
        autofocus: true,
        keyboardType: TextInputType.number,
        decoration: const InputDecoration(
          labelText: 'Nominal',
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
            final amount = int.tryParse(
              controller.text.replaceAll(RegExp(r'[^0-9]'), ''),
            );
            Navigator.of(context).pop(amount);
          },
          child: const Text('Tambah'),
        ),
      ],
    );
  }
}

class ProductCard extends StatelessWidget {
  const ProductCard({required this.product, required this.onTap, super.key});

  final Product product;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white,
      borderRadius: BorderRadius.circular(MokposRadius.sm),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(MokposRadius.sm),
        child: Ink(
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(MokposRadius.sm),
            border: Border.all(color: MokposColors.line),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Stack(
                  children: [
                    Positioned.fill(
                      child: ClipRRect(
                        borderRadius: const BorderRadius.vertical(
                          top: Radius.circular(MokposRadius.sm),
                        ),
                        child: Image.network(
                          product.imageUrl,
                          fit: BoxFit.cover,
                          errorBuilder: (context, error, stackTrace) {
                            return Container(
                              color: const Color(0xFFDDE7DD),
                              child: Center(
                                child: Text(
                                  product.name.characters.first.toUpperCase(),
                                  style: const TextStyle(
                                    color: MokposColors.muted,
                                    fontSize: 34,
                                    fontWeight: FontWeight.w800,
                                  ),
                                ),
                              ),
                            );
                          },
                        ),
                      ),
                    ),
                    if (product.badge != null)
                      Positioned(
                        left: 8,
                        top: 8,
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: .9),
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            product.badge!,
                            style: const TextStyle(
                              color: MokposColors.muted,
                              fontSize: 11,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                        ),
                      ),
                  ],
                ),
              ),
              Padding(
                padding: const EdgeInsets.fromLTRB(12, 10, 12, 4),
                child: Text(
                  product.name,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: MokposColors.text,
                    fontWeight: FontWeight.w800,
                    height: 1.15,
                    fontSize: 13,
                  ),
                ),
              ),
              Padding(
                padding: const EdgeInsets.fromLTRB(12, 0, 12, 12),
                child: Row(
                  children: [
                    Expanded(
                      child: Text(
                        rupiah(product.price),
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          color: MokposColors.muted,
                          fontWeight: FontWeight.w800,
                          fontSize: 13,
                        ),
                      ),
                    ),
                    if (product.isFavorite)
                      const Icon(
                        LucideIcons.star,
                        color: MokposColors.warning,
                        size: 18,
                      ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
