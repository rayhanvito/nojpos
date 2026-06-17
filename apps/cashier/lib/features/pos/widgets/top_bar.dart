import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';

class PosTopBar extends StatelessWidget {
  const PosTopBar({
    required this.onOpenMenu,
    required this.onOpenOrders,
    required this.onOpenMode,
    required this.onShowNotification,
    super.key,
  });

  final VoidCallback onOpenMenu;
  final VoidCallback onOpenOrders;
  final VoidCallback onOpenMode;
  final VoidCallback onShowNotification;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 64,
      color: MokposColors.primary,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      child: Row(
        children: [
          _TopIconButton(icon: LucideIcons.menu, onTap: onOpenMenu),
          const SizedBox(width: 10),
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(14),
            ),
            child: const Icon(
              LucideIcons.store,
              color: MokposColors.primary,
              size: 25,
            ),
          ),
          const SizedBox(width: 12),
          const Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Kedai Nusantara - Kasir',
                style: TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.w900,
                  fontSize: 16,
                ),
              ),
              SizedBox(height: 2),
              Row(
                children: [
                  _OnlineDot(),
                  SizedBox(width: 6),
                  Text(
                    'Status: Online · Shift Aktif',
                    style: TextStyle(color: Color(0xE6FFFFFF), fontSize: 12),
                  ),
                ],
              ),
            ],
          ),
          const Spacer(),
          const Text(
            'NojPOS',
            style: TextStyle(
              color: Colors.white,
              fontSize: 26,
              fontWeight: FontWeight.w900,
            ),
          ),
          const Spacer(),
          _TopIconButton(icon: LucideIcons.bell, onTap: onShowNotification),
          _TopIconButton(icon: LucideIcons.layoutGrid, onTap: onOpenMode),
          _TopIconButton(
            icon: LucideIcons.lockKeyhole,
            onTap: () => context.go('/login'),
          ),
          const SizedBox(width: 10),
          GestureDetector(
            behavior: HitTestBehavior.opaque,
            onTap: onOpenOrders,
            child: Container(
              height: 48,
              width: 210,
              padding: const EdgeInsets.symmetric(horizontal: 18),
              decoration: BoxDecoration(
                color: MokposColors.primaryDark,
                borderRadius: BorderRadius.circular(MokposRadius.sm),
              ),
              child: const Row(
                children: [
                  Expanded(
                    child: Text(
                      'Daftar Order',
                      style: TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.w900,
                        fontSize: 16,
                      ),
                    ),
                  ),
                  Icon(LucideIcons.chevronRight, color: Colors.white, size: 21),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _OnlineDot extends StatelessWidget {
  const _OnlineDot();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 10,
      height: 10,
      decoration: const BoxDecoration(
        color: Color(0xFF70E06D),
        shape: BoxShape.circle,
      ),
    );
  }
}

class _TopIconButton extends StatelessWidget {
  const _TopIconButton({required this.icon, required this.onTap});

  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: 'Aksi cepat',
      child: IconButton(
        onPressed: onTap,
        icon: Icon(icon),
        color: Colors.white,
        iconSize: 23,
        style: IconButton.styleFrom(
          fixedSize: const Size(46, 46),
          backgroundColor: Colors.white.withValues(alpha: .08),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(MokposRadius.sm),
          ),
        ),
      ),
    );
  }
}
