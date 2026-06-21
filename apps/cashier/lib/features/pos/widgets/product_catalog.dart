import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/nojpos_assets.dart';
import '../../../app/theme.dart';
import '../../../shared/widgets/nojpos_skeleton.dart';
import '../../../shared/widgets/nojpos_state_view.dart';
import '../../../shared/widgets/nojpos_toast.dart';
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
                ? _CatalogSkeleton(compact: compact)
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

class _CatalogSkeleton extends StatelessWidget {
  const _CatalogSkeleton({required this.compact});

  final bool compact;

  @override
  Widget build(BuildContext context) {
    return NojposSkeleton(
      enabled: true,
      child: GridView.builder(
        itemCount: compact ? 8 : 12,
        gridDelegate: SliverGridDelegateWithMaxCrossAxisExtent(
          maxCrossAxisExtent: compact ? 142 : 158,
          mainAxisSpacing: 8,
          crossAxisSpacing: 8,
          mainAxisExtent: compact ? 178 : 192,
        ),
        itemBuilder: (context, index) => Container(
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(NojposRadius.md),
            border: Border.all(color: NojposColors.line),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Container(
                  decoration: const BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.vertical(
                      top: Radius.circular(NojposRadius.md),
                    ),
                  ),
                ),
              ),
              const Padding(
                padding: EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Memuat produk'),
                    SizedBox(height: 8),
                    Text('Harga dari server'),
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

class _CatalogError extends ConsumerWidget {
  const _CatalogError({required this.message});

  final String message;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return NojposStateView.error(
      title: 'Produk gagal dimuat',
      subtitle: message,
      illustrationAsset: NojposAssets.noInternetCashier,
      actionLabel: 'Coba lagi',
      onAction: () => ref.read(posCatalogProvider.notifier).load(),
    );
  }
}

class _NoProductsFound extends StatelessWidget {
  const _NoProductsFound();

  @override
  Widget build(BuildContext context) {
    return NojposStateView.empty(
      title: 'Produk tidak ditemukan',
      subtitle:
          'Coba kata kunci lain atau pilih kategori berbeda. Pencarian mendukung nama, SKU, dan barcode.',
      illustrationAsset: NojposAssets.emptyProducts,
      compact: false,
    );
  }
}

class _CatalogToolbar extends ConsumerWidget {
  const _CatalogToolbar({required this.compact});

  final bool compact;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Container(
      height: 62,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(bottom: BorderSide(color: NojposColors.line)),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      child: Row(
        children: [
          Expanded(
            child: DecoratedBox(
              decoration: BoxDecoration(
                color: NojposColors.canvas,
                borderRadius: BorderRadius.circular(NojposRadius.lg),
                border: Border.all(color: NojposColors.line),
              ),
              child: TextField(
                onChanged: (value) =>
                    ref.read(productSearchQueryProvider.notifier).update(value),
                decoration: const InputDecoration(
                  prefixIcon: Icon(
                    LucideIcons.search,
                    color: NojposColors.muted,
                    size: 22,
                  ),
                  hintText: 'Cari produk, SKU, atau barcode',
                  hintStyle: TextStyle(color: NojposColors.muted, fontSize: 15),
                  border: InputBorder.none,
                  contentPadding: EdgeInsets.symmetric(vertical: 13),
                ),
              ),
            ),
          ),
          if (!compact) ...[
            const SizedBox(width: 12),
            const _IconAction(
              icon: LucideIcons.scanBarcode,
              tooltip:
                  'Scanner kamera nonaktif di preview produksi. Gunakan kolom pencarian untuk SKU atau barcode.',
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
        onTap: enabled
            ? () {}
            : () => NojposToast.info(
                context,
                'Scanner kamera nonaktif',
                description:
                    'Gunakan kolom pencarian untuk mengetik SKU atau barcode produk.',
              ),
        borderRadius: BorderRadius.circular(NojposRadius.md),
        child: Container(
          width: 46,
          height: 46,
          decoration: BoxDecoration(
            color: enabled ? Colors.white : NojposColors.disabledSurface,
            border: Border.all(color: NojposColors.line),
            borderRadius: BorderRadius.circular(NojposRadius.md),
          ),
          child: Icon(
            icon,
            color: enabled ? NojposColors.text : NojposColors.muted,
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
      borderRadius: BorderRadius.circular(NojposRadius.md),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(NojposRadius.md),
        child: Ink(
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(NojposRadius.md),
            border: Border.all(color: NojposColors.line),
            boxShadow: NojposShadow.soft,
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
                          top: Radius.circular(NojposRadius.md),
                        ),
                        child: product.imageUrl.trim().isEmpty
                            ? _ProductImageFallback(product: product)
                            : CachedNetworkImage(
                                imageUrl: product.imageUrl,
                                fit: BoxFit.cover,
                                placeholder: (context, url) =>
                                    const NojposSkeleton(
                                      enabled: true,
                                      child: NojposSkeletonCard(
                                        height: double.infinity,
                                        radius: NojposRadius.md,
                                      ),
                                    ),
                                errorWidget: (context, url, error) =>
                                    _ProductImageFallback(product: product),
                              ),
                      ),
                    ),
                    Positioned.fill(
                      child: DecoratedBox(
                        decoration: BoxDecoration(
                          gradient: LinearGradient(
                            colors: [
                              Colors.black.withValues(alpha: .10),
                              Colors.transparent,
                            ],
                            begin: Alignment.bottomCenter,
                            end: Alignment.center,
                          ),
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
                            color: Colors.white.withValues(alpha: .92),
                            borderRadius: BorderRadius.circular(
                              NojposRadius.sm,
                            ),
                            boxShadow: NojposShadow.soft,
                          ),
                          child: Text(
                            product.badge!,
                            style: const TextStyle(
                              color: NojposColors.muted,
                              fontSize: 11,
                              fontWeight: FontWeight.w900,
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
                        color: NojposColors.text,
                        fontWeight: FontWeight.w900,
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
                          color: NojposColors.primaryDark,
                          fontWeight: FontWeight.w900,
                          fontSize: 13,
                        ),
                      ),
                    ),
                    if (product.isFavorite)
                      Container(
                        width: 26,
                        height: 26,
                        decoration: BoxDecoration(
                          color: NojposColors.warningSurface,
                          borderRadius: BorderRadius.circular(999),
                        ),
                        child: const Icon(
                          LucideIcons.star,
                          color: Color(0xFFB58200),
                          size: 16,
                        ),
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

class _ProductImageFallback extends StatelessWidget {
  const _ProductImageFallback({required this.product});

  final Product product;

  @override
  Widget build(BuildContext context) {
    final initial = product.name.trim().isEmpty
        ? '?'
        : product.name.characters.first.toUpperCase();
    return Container(
      decoration: const BoxDecoration(gradient: NojposColors.softGradient),
      child: Center(
        child: Container(
          width: 62,
          height: 62,
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: .78),
            borderRadius: BorderRadius.circular(22),
            border: Border.all(color: NojposColors.successBorder),
          ),
          child: Center(
            child: Text(
              initial,
              style: const TextStyle(
                color: NojposColors.primaryDark,
                fontSize: 30,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
