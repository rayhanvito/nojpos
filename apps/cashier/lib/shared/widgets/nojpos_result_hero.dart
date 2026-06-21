import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:lottie/lottie.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../app/theme.dart';
import 'nojpos_asset_icon.dart';

enum NojposResultType { success, failed, pending, warning }

class NojposResultHero extends StatelessWidget {
  const NojposResultHero._({
    required this.type,
    required this.title,
    required this.subtitle,
    this.amount,
    this.summary,
    this.metadata,
    this.animationAsset,
    this.illustrationAsset,
    this.icon,
    this.trailing,
    this.compact = false,
  });

  factory NojposResultHero.success({
    required String title,
    required String subtitle,
    String? amount,
    String? summary,
    Widget? metadata,
    String? animationAsset,
    String? illustrationAsset,
    IconData? icon,
    Widget? trailing,
    bool compact = false,
  }) => NojposResultHero._(
    type: NojposResultType.success,
    title: title,
    subtitle: subtitle,
    amount: amount,
    summary: summary,
    metadata: metadata,
    animationAsset: animationAsset,
    illustrationAsset: illustrationAsset,
    icon: icon,
    trailing: trailing,
    compact: compact,
  );

  factory NojposResultHero.failed({
    required String title,
    required String subtitle,
    String? amount,
    String? summary,
    Widget? metadata,
    String? animationAsset,
    String? illustrationAsset,
    IconData? icon,
    Widget? trailing,
    bool compact = false,
  }) => NojposResultHero._(
    type: NojposResultType.failed,
    title: title,
    subtitle: subtitle,
    amount: amount,
    summary: summary,
    metadata: metadata,
    animationAsset: animationAsset,
    illustrationAsset: illustrationAsset,
    icon: icon,
    trailing: trailing,
    compact: compact,
  );

  factory NojposResultHero.pending({
    required String title,
    required String subtitle,
    String? amount,
    String? summary,
    Widget? metadata,
    String? animationAsset,
    String? illustrationAsset,
    IconData? icon,
    Widget? trailing,
    bool compact = false,
  }) => NojposResultHero._(
    type: NojposResultType.pending,
    title: title,
    subtitle: subtitle,
    amount: amount,
    summary: summary,
    metadata: metadata,
    animationAsset: animationAsset,
    illustrationAsset: illustrationAsset,
    icon: icon,
    trailing: trailing,
    compact: compact,
  );

  factory NojposResultHero.warning({
    required String title,
    required String subtitle,
    String? amount,
    String? summary,
    Widget? metadata,
    String? animationAsset,
    String? illustrationAsset,
    IconData? icon,
    Widget? trailing,
    bool compact = false,
  }) => NojposResultHero._(
    type: NojposResultType.warning,
    title: title,
    subtitle: subtitle,
    amount: amount,
    summary: summary,
    metadata: metadata,
    animationAsset: animationAsset,
    illustrationAsset: illustrationAsset,
    icon: icon,
    trailing: trailing,
    compact: compact,
  );

  final NojposResultType type;
  final String title;
  final String subtitle;
  final String? amount;
  final String? summary;
  final Widget? metadata;
  final String? animationAsset;
  final String? illustrationAsset;
  final IconData? icon;
  final Widget? trailing;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final style = _styleFor(type);
    final mediaWidth = MediaQuery.sizeOf(context).width;
    final visualSize = compact ? 90.0 : (mediaWidth < 420 ? 120.0 : 146.0);
    return Container(
      width: double.infinity,
      padding: EdgeInsets.all(compact ? 20 : 28),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [Colors.white, style.surface],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(NojposRadius.xl),
        border: Border.all(color: style.border),
        boxShadow: NojposShadow.card,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          _ResultVisual(
            size: visualSize,
            style: style,
            fallbackIcon: icon ?? style.icon,
            animationAsset: animationAsset,
            illustrationAsset: illustrationAsset,
            compact: compact,
          ),
          SizedBox(height: compact ? 16 : 20),
          Text(
            title,
            textAlign: TextAlign.center,
            style: TextStyle(
              color: NojposColors.text,
              fontSize: compact ? 22 : 28,
              height: 1.05,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            subtitle,
            textAlign: TextAlign.center,
            style: const TextStyle(
              color: NojposColors.muted,
              fontSize: 14,
              height: 1.45,
              fontWeight: FontWeight.w600,
            ),
          ),
          if (amount != null) ...[
            const SizedBox(height: 18),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
              decoration: BoxDecoration(
                color: Colors.white.withValues(alpha: .82),
                borderRadius: BorderRadius.circular(NojposRadius.lg),
                border: Border.all(color: style.border),
              ),
              child: Text(
                amount!,
                style: TextStyle(
                  color: style.color,
                  fontSize: compact ? 22 : 28,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
          ],
          if (metadata != null) ...[const SizedBox(height: 14), metadata!],
          if (summary != null) ...[
            const SizedBox(height: 12),
            Text(
              summary!,
              textAlign: TextAlign.center,
              style: const TextStyle(color: NojposColors.muted, fontSize: 13),
            ),
          ],
          if (trailing != null) ...[const SizedBox(height: 18), trailing!],
        ],
      ),
    ).animate().fadeIn(duration: 260.ms).slideY(begin: .04, end: 0);
  }
}

class _ResultVisual extends StatelessWidget {
  const _ResultVisual({
    required this.size,
    required this.style,
    required this.fallbackIcon,
    required this.compact,
    this.animationAsset,
    this.illustrationAsset,
  });

  final double size;
  final _ResultStyle style;
  final IconData fallbackIcon;
  final bool compact;
  final String? animationAsset;
  final String? illustrationAsset;

  @override
  Widget build(BuildContext context) {
    final lottieAsset = animationAsset;
    final visualAsset = illustrationAsset;
    final badgeSize = compact ? size : size * .64;
    Widget child;
    if (lottieAsset != null && lottieAsset.isNotEmpty) {
      child = Lottie.asset(
        lottieAsset,
        width: size,
        height: size,
        fit: BoxFit.contain,
        errorBuilder: (context, error, stackTrace) => _IllustrationOrBadge(
          assetName: visualAsset,
          size: size,
          badgeSize: badgeSize,
          style: style,
          fallbackIcon: fallbackIcon,
        ),
      );
    } else if (visualAsset != null && visualAsset.isNotEmpty) {
      child = _AssetVisual(
        assetName: visualAsset,
        width: size,
        height: size,
        fit: BoxFit.contain,
        fallback: _Badge(
          size: badgeSize,
          style: style,
          fallbackIcon: fallbackIcon,
        ),
      );
    } else {
      child = _Badge(size: badgeSize, style: style, fallbackIcon: fallbackIcon);
    }

    return SizedBox(
          width: size,
          height: size,
          child: Center(child: child),
        )
        .animate(
          onPlay: (controller) {
            if (style.shouldPulse) controller.repeat(reverse: true);
          },
        )
        .scale(
          begin: const Offset(1, 1),
          end: Offset(style.pulseScale, style.pulseScale),
          duration: style.pulseDuration,
          curve: Curves.easeInOut,
        );
  }
}

class _IllustrationOrBadge extends StatelessWidget {
  const _IllustrationOrBadge({
    required this.assetName,
    required this.size,
    required this.badgeSize,
    required this.style,
    required this.fallbackIcon,
  });

  final String? assetName;
  final double size;
  final double badgeSize;
  final _ResultStyle style;
  final IconData fallbackIcon;

  @override
  Widget build(BuildContext context) {
    final asset = assetName;
    if (asset == null || asset.isEmpty) {
      return _Badge(size: badgeSize, style: style, fallbackIcon: fallbackIcon);
    }
    return _AssetVisual(
      assetName: asset,
      width: size,
      height: size,
      fit: BoxFit.contain,
      fallback: _Badge(
        size: badgeSize,
        style: style,
        fallbackIcon: fallbackIcon,
      ),
    );
  }
}

class _AssetVisual extends StatelessWidget {
  const _AssetVisual({
    required this.assetName,
    required this.width,
    required this.height,
    required this.fit,
    required this.fallback,
  });

  final String assetName;
  final double width;
  final double height;
  final BoxFit fit;
  final Widget fallback;

  @override
  Widget build(BuildContext context) {
    if (assetName.toLowerCase().endsWith('.svg')) {
      return SvgPicture.asset(
        assetName,
        width: width,
        height: height,
        fit: fit,
        placeholderBuilder: (_) => fallback,
        errorBuilder: (context, error, stackTrace) => fallback,
      );
    }

    return Image.asset(
      assetName,
      width: width,
      height: height,
      fit: fit,
      filterQuality: FilterQuality.medium,
      errorBuilder: (context, error, stackTrace) => fallback,
    );
  }
}

class _Badge extends StatelessWidget {
  const _Badge({
    required this.size,
    required this.style,
    required this.fallbackIcon,
  });

  final double size;
  final _ResultStyle style;
  final IconData fallbackIcon;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: style.surface,
        border: Border.all(color: style.border, width: 1.4),
        boxShadow: NojposShadow.soft,
      ),
      child: Center(
        child: NojposAssetIcon(
          fallbackIcon: fallbackIcon,
          size: size * .46,
          color: style.color,
        ),
      ),
    );
  }
}

class _ResultStyle {
  const _ResultStyle({
    required this.surface,
    required this.border,
    required this.color,
    required this.icon,
    this.pulseScale = 1.02,
    this.pulseDuration = const Duration(milliseconds: 1500),
    this.shouldPulse = true,
  });

  final Color surface;
  final Color border;
  final Color color;
  final IconData icon;
  final double pulseScale;
  final Duration pulseDuration;
  final bool shouldPulse;
}

_ResultStyle _styleFor(NojposResultType type) => switch (type) {
  NojposResultType.success => const _ResultStyle(
    surface: NojposColors.successSurface,
    border: NojposColors.successBorder,
    color: Color(0xFF118A4A),
    icon: LucideIcons.badgeCheck,
    shouldPulse: false,
  ),
  NojposResultType.failed => const _ResultStyle(
    surface: NojposColors.dangerSurface,
    border: NojposColors.dangerBorder,
    color: NojposColors.danger,
    icon: LucideIcons.circleX,
    shouldPulse: false,
  ),
  NojposResultType.pending => const _ResultStyle(
    surface: NojposColors.infoSurface,
    border: NojposColors.infoBorder,
    color: NojposColors.accent,
    icon: LucideIcons.loaderCircle,
    pulseScale: 1.04,
    pulseDuration: Duration(milliseconds: 1100),
  ),
  NojposResultType.warning => const _ResultStyle(
    surface: NojposColors.warningSurface,
    border: NojposColors.warningBorder,
    color: Color(0xFF9A6B00),
    icon: LucideIcons.triangleAlert,
    shouldPulse: false,
  ),
};
