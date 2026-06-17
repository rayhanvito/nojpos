import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../providers/pos_providers.dart';

class PosSidebar extends ConsumerWidget {
  const PosSidebar({
    required this.categoryPanelOpen,
    required this.onToggleCategories,
    required this.onOpenMode,
    required this.onShowFeature,
    super.key,
  });

  final bool categoryPanelOpen;
  final VoidCallback onToggleCategories;
  final VoidCallback onOpenMode;
  final ValueChanged<String> onShowFeature;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final selectedCategory = ref.watch(selectedCategoryProvider);
    final categories = ref.watch(categoriesProvider);
    final expanded = categoryPanelOpen;

    void selectCategory(String category) {
      ref.read(selectedCategoryProvider.notifier).select(category);
    }

    return AnimatedContainer(
      duration: const Duration(milliseconds: 180),
      curve: Curves.easeOutCubic,
      width: expanded ? 220 : 72,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(right: BorderSide(color: MokposColors.line)),
      ),
      child: ClipRect(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const SizedBox(height: 8),
            _RailItem(
              icon: expanded
                  ? LucideIcons.panelLeftClose
                  : LucideIcons.panelLeftOpen,
              label: 'Kategori Produk',
              expanded: expanded,
              active: expanded,
              onTap: onToggleCategories,
            ),
            _RailItem(
              icon: LucideIcons.blocks,
              label: 'Grid',
              expanded: expanded,
              active: selectedCategory == 'Semua',
              onTap: onOpenMode,
            ),
            _RailItem(
              icon: LucideIcons.star,
              label: 'Favorit',
              expanded: expanded,
              active: selectedCategory == 'Favorit',
              onTap: () => selectCategory('Favorit'),
            ),
            _RailItem(
              icon: LucideIcons.package,
              label: 'Produk Paket',
              expanded: expanded,
              active: selectedCategory == 'Paket',
              onTap: () => selectCategory('Paket'),
            ),
            _RailItem(
              icon: LucideIcons.bookOpen,
              label: 'Buku Menu',
              expanded: expanded,
              onTap: () => onShowFeature('Buku Menu'),
            ),
            _RailItem(
              icon: LucideIcons.ticketPercent,
              label: 'Promo',
              expanded: expanded,
              onTap: () => onShowFeature('Promo / Voucher'),
            ),
            _RailItem(
              icon: LucideIcons.walletCards,
              label: 'Kas / Wallet',
              expanded: expanded,
              onTap: () => onShowFeature('Kas / Wallet'),
            ),
            const Divider(height: 20, color: MokposColors.line),
            if (expanded) ...[
              const Padding(
                padding: EdgeInsets.fromLTRB(16, 4, 16, 8),
                child: Text(
                  'Kategori',
                  style: TextStyle(
                    color: MokposColors.muted,
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              for (final category in categories)
                _CategoryTextRow(
                  text: category,
                  active: selectedCategory == category,
                  onTap: () => selectCategory(category),
                ),
            ] else ...[
              _Shortcut(
                text: 'MA',
                active: selectedCategory == 'Makanan',
                onTap: () => selectCategory('Makanan'),
              ),
              _Shortcut(
                text: 'MU',
                active: selectedCategory == 'Minuman',
                onTap: () => selectCategory('Minuman'),
              ),
              _Shortcut(
                text: 'PA',
                active: selectedCategory == 'Paket',
                onTap: () => selectCategory('Paket'),
              ),
              _Shortcut(
                text: 'LA',
                active: selectedCategory == 'Layanan',
                onTap: () => selectCategory('Layanan'),
              ),
              _Shortcut(
                text: 'FA',
                active: selectedCategory == 'Favorit',
                onTap: () => selectCategory('Favorit'),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _RailItem extends StatelessWidget {
  const _RailItem({
    required this.icon,
    required this.label,
    required this.expanded,
    required this.onTap,
    this.active = false,
  });

  final IconData icon;
  final String label;
  final bool expanded;
  final VoidCallback onTap;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: expanded ? '' : label,
      child: InkWell(
        onTap: onTap,
        child: Container(
          height: 54,
          padding: EdgeInsets.only(left: expanded ? 16 : 0, right: 10),
          decoration: BoxDecoration(
            color: active ? MokposColors.primarySoft : Colors.transparent,
            border: Border(
              left: BorderSide(
                color: active ? MokposColors.primary : Colors.transparent,
                width: 3,
              ),
            ),
          ),
          child: Row(
            mainAxisAlignment: expanded
                ? MainAxisAlignment.start
                : MainAxisAlignment.center,
            children: [
              Icon(
                icon,
                color: active ? MokposColors.primary : MokposColors.muted,
                size: 24,
              ),
              if (expanded) ...[
                const SizedBox(width: 14),
                Expanded(
                  child: Text(
                    label,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: active
                          ? MokposColors.primaryDark
                          : MokposColors.text,
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _CategoryTextRow extends StatelessWidget {
  const _CategoryTextRow({
    required this.text,
    required this.active,
    required this.onTap,
  });

  final String text;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Container(
        height: 38,
        padding: const EdgeInsets.only(left: 48, right: 12),
        decoration: BoxDecoration(
          color: active ? MokposColors.primarySoft : Colors.white,
          border: const Border(bottom: BorderSide(color: MokposColors.line)),
        ),
        alignment: Alignment.centerLeft,
        child: Text(
          text,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: TextStyle(
            color: active ? MokposColors.primaryDark : MokposColors.muted,
            fontSize: 12,
            fontWeight: active ? FontWeight.w900 : FontWeight.w700,
          ),
        ),
      ),
    );
  }
}

class _Shortcut extends StatelessWidget {
  const _Shortcut({
    required this.text,
    required this.active,
    required this.onTap,
  });

  final String text;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: SizedBox(
        height: 46,
        child: Center(
          child: Container(
            width: 42,
            height: 34,
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: active ? MokposColors.primarySoft : Colors.transparent,
              borderRadius: BorderRadius.circular(6),
            ),
            child: Text(
              text,
              style: TextStyle(
                color: active ? MokposColors.primary : MokposColors.muted,
                fontWeight: FontWeight.w900,
                fontSize: 12,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
