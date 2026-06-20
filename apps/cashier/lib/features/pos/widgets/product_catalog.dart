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
    final catalog = ref.watch(posCatalogProvider);
    final products = ref.watch(filteredProductsProvider);

    return Column(
      children: [
        _CatalogToolbar(compact: compact),
        Expanded(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(8, 8, 8, 8),
            child: catalog.isLoading
                ? const Center(child: CircularProgressIndicator())
                : catalog.errorMessage != null
                ? _CatalogError(message: catalog.errorMessage!)
                : products.isEmpty
                ? const _NoProductsFound()
                : GridView.builder(
                    itemCount: products.length,
                    gridDelegate: SliverGridDelegateWithMaxCrossAxisExtent(
                      maxCrossAxisExtent: compact ? 142 : 158,
                      mainAxisSpacing: 8,
                      crossAxisSpacing: 8,
                      mainAxisExtent: compact ? 178 : 192,
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

class _CatalogError extends ConsumerWidget {
  const _CatalogError({required this.message});

  final String message;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(
            LucideIcons.cloudOff,
            color: MokposColors.primary,
            size: 72,
          ),
          const SizedBox(height: 14),
          const Text(
            'Produk gagal dimuat',
            style: TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
              fontSize: 16,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            message,
            textAlign: TextAlign.center,
            style: const TextStyle(color: MokposColors.muted),
          ),
          const SizedBox(height: 16),
          FilledButton(
            onPressed: () => ref.read(posCatalogProvider.notifier).load(),
            child: const Text('Coba lagi'),
          ),
        ],
      ),
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
            const _IconAction(
              icon: LucideIcons.scanBarcode,
              tooltip: 'Segera hadir (keyboard-wedge)',
              enabled: false,
            ),
          ],
        ],
      ),
    );
  }
}

class _IconAction extends StatelessWidget {
  const _IconAction({required this.icon, this.tooltip, this.enabled = true});

  final IconData icon;
  final String? tooltip;
  final bool enabled;

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: tooltip ?? '',
      child: InkWell(
        onTap: null,
        borderRadius: BorderRadius.circular(MokposRadius.sm),
        child: Container(
          width: 44,
          height: 44,
          decoration: BoxDecoration(
            color: enabled ? Colors.white : MokposColors.disabledSurface,
            border: Border.all(color: MokposColors.line),
            borderRadius: BorderRadius.circular(MokposRadius.sm),
          ),
          child: Icon(
            icon,
            color: enabled ? MokposColors.text : MokposColors.muted,
            size: 21,
          ),
        ),
      ),
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
      key: ValueKey('product_${product.id}'),
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
                              color: MokposColors.productFallback,
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
                            borderRadius: BorderRadius.circular(
                              MokposRadius.xs,
                            ),
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
              SizedBox(
                height: 46,
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(12, 10, 12, 4),
                  child: Align(
                    alignment: Alignment.topLeft,
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
