import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../app/theme.dart';

class NojposAssetIcon extends StatelessWidget {
  const NojposAssetIcon({
    super.key,
    this.assetName,
    this.fallbackIcon = LucideIcons.circleHelp,
    this.size = 24,
    this.color,
    this.fit = BoxFit.contain,
    this.applyColor = true,
    this.semanticLabel,
  });

  factory NojposAssetIcon.named(
    String name, {
    Key? key,
    IconData fallbackIcon = LucideIcons.circleHelp,
    double size = 24,
    Color? color,
    BoxFit fit = BoxFit.contain,
    bool applyColor = true,
    String? semanticLabel,
  }) {
    return NojposAssetIcon(
      key: key,
      assetName: 'assets/icons/$name.svg',
      fallbackIcon: fallbackIcon,
      size: size,
      color: color,
      fit: fit,
      applyColor: applyColor,
      semanticLabel: semanticLabel,
    );
  }

  final String? assetName;
  final IconData fallbackIcon;
  final double size;
  final Color? color;
  final BoxFit fit;
  final bool applyColor;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final iconColor = color ?? NojposColors.primary;
    final fallback = Icon(
      fallbackIcon,
      size: size,
      color: iconColor,
      semanticLabel: semanticLabel,
    );
    final asset = assetName;
    if (asset == null || asset.isEmpty) return fallback;

    return SvgPicture.asset(
      asset,
      width: size,
      height: size,
      fit: fit,
      semanticsLabel: semanticLabel,
      colorFilter: applyColor
          ? ColorFilter.mode(iconColor, BlendMode.srcIn)
          : null,
      placeholderBuilder: (_) => fallback,
      errorBuilder: (context, error, stackTrace) => fallback,
    );
  }
}
