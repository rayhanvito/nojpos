import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../app/theme.dart';
import 'nojpos_asset_icon.dart';

enum NojposStateType { empty, error, warning, success, loading }

class NojposStateView extends StatelessWidget {
  const NojposStateView._({
    required this.type,
    required this.title,
    required this.subtitle,
    this.illustrationAsset,
    this.actionLabel,
    this.onAction,
    this.secondaryActionLabel,
    this.onSecondaryAction,
    this.footer,
    this.compact = false,
    this.maxWidth,
  });

  factory NojposStateView.empty({
    required String title,
    String? subtitle,
    String? illustrationAsset,
    String? actionLabel,
    VoidCallback? onAction,
    String? secondaryActionLabel,
    VoidCallback? onSecondaryAction,
    Widget? footer,
    bool compact = false,
    double? maxWidth,
  }) => NojposStateView._(
    type: NojposStateType.empty,
    title: title,
    subtitle: subtitle ?? 'Belum ada data untuk ditampilkan.',
    illustrationAsset: illustrationAsset,
    actionLabel: actionLabel,
    onAction: onAction,
    secondaryActionLabel: secondaryActionLabel,
    onSecondaryAction: onSecondaryAction,
    footer: footer,
    compact: compact,
    maxWidth: maxWidth,
  );

  factory NojposStateView.error({
    required String title,
    required String subtitle,
    String? illustrationAsset,
    String? actionLabel,
    VoidCallback? onAction,
    String? secondaryActionLabel,
    VoidCallback? onSecondaryAction,
    Widget? footer,
    bool compact = false,
    double? maxWidth,
  }) => NojposStateView._(
    type: NojposStateType.error,
    title: title,
    subtitle: subtitle,
    illustrationAsset: illustrationAsset,
    actionLabel: actionLabel,
    onAction: onAction,
    secondaryActionLabel: secondaryActionLabel,
    onSecondaryAction: onSecondaryAction,
    footer: footer,
    compact: compact,
    maxWidth: maxWidth,
  );

  factory NojposStateView.warning({
    required String title,
    required String subtitle,
    String? illustrationAsset,
    String? actionLabel,
    VoidCallback? onAction,
    String? secondaryActionLabel,
    VoidCallback? onSecondaryAction,
    Widget? footer,
    bool compact = false,
    double? maxWidth,
  }) => NojposStateView._(
    type: NojposStateType.warning,
    title: title,
    subtitle: subtitle,
    illustrationAsset: illustrationAsset,
    actionLabel: actionLabel,
    onAction: onAction,
    secondaryActionLabel: secondaryActionLabel,
    onSecondaryAction: onSecondaryAction,
    footer: footer,
    compact: compact,
    maxWidth: maxWidth,
  );

  factory NojposStateView.success({
    required String title,
    required String subtitle,
    String? illustrationAsset,
    String? actionLabel,
    VoidCallback? onAction,
    String? secondaryActionLabel,
    VoidCallback? onSecondaryAction,
    Widget? footer,
    bool compact = false,
    double? maxWidth,
  }) => NojposStateView._(
    type: NojposStateType.success,
    title: title,
    subtitle: subtitle,
    illustrationAsset: illustrationAsset,
    actionLabel: actionLabel,
    onAction: onAction,
    secondaryActionLabel: secondaryActionLabel,
    onSecondaryAction: onSecondaryAction,
    footer: footer,
    compact: compact,
    maxWidth: maxWidth,
  );

  factory NojposStateView.loading({
    String title = 'Memuat data',
    String subtitle = 'Sebentar, data sedang disiapkan.',
    bool compact = false,
    double? maxWidth,
  }) => NojposStateView._(
    type: NojposStateType.loading,
    title: title,
    subtitle: subtitle,
    compact: compact,
    maxWidth: maxWidth,
  );

  final NojposStateType type;
  final String title;
  final String subtitle;
  final String? illustrationAsset;
  final String? actionLabel;
  final VoidCallback? onAction;
  final String? secondaryActionLabel;
  final VoidCallback? onSecondaryAction;
  final Widget? footer;
  final bool compact;
  final double? maxWidth;

  @override
  Widget build(BuildContext context) {
    final style = _styleFor(type);
    final width = MediaQuery.sizeOf(context).width;
    final resolvedMaxWidth = maxWidth ?? (width < 520 ? width - 32 : 460.0);

    return Center(
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: resolvedMaxWidth),
        child: Padding(
          padding: EdgeInsets.all(
            compact ? NojposSpacing.md : NojposSpacing.xl,
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              _StateVisual(
                type: type,
                style: style,
                illustrationAsset: illustrationAsset,
                compact: compact,
              ),
              SizedBox(height: compact ? 14 : 20),
              Text(
                title,
                textAlign: TextAlign.center,
                style: TextStyle(
                  color: NojposColors.text,
                  fontSize: compact ? 16 : 20,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                subtitle,
                textAlign: TextAlign.center,
                style: TextStyle(
                  color: NojposColors.muted,
                  height: 1.45,
                  fontSize: compact ? 13 : 14,
                ),
              ),
              if (actionLabel != null || secondaryActionLabel != null) ...[
                const SizedBox(height: 20),
                Wrap(
                  alignment: WrapAlignment.center,
                  spacing: 10,
                  runSpacing: 10,
                  children: [
                    if (secondaryActionLabel != null)
                      OutlinedButton(
                        onPressed: onSecondaryAction,
                        child: Text(secondaryActionLabel!),
                      ),
                    if (actionLabel != null)
                      FilledButton(
                        onPressed: onAction,
                        child: Text(actionLabel!),
                      ),
                  ],
                ),
              ],
              if (footer != null) ...[const SizedBox(height: 18), footer!],
            ],
          ).animate().fadeIn(duration: 260.ms).slideY(begin: .03, end: 0),
        ),
      ),
    );
  }
}

class _StateVisual extends StatelessWidget {
  const _StateVisual({
    required this.type,
    required this.style,
    required this.compact,
    this.illustrationAsset,
  });

  final NojposStateType type;
  final _StateStyle style;
  final bool compact;
  final String? illustrationAsset;

  @override
  Widget build(BuildContext context) {
    final asset = illustrationAsset;
    final visualSize = compact ? 74.0 : 132.0;
    if (asset != null && asset.isNotEmpty && type != NojposStateType.loading) {
      return SizedBox(
            width: visualSize,
            height: visualSize,
            child: _StateAssetVisual(
              assetName: asset,
              fallback: _IconBadge(type: type, style: style, compact: compact),
            ),
          )
          .animate()
          .fadeIn(duration: 220.ms)
          .scale(begin: const Offset(.96, .96), duration: 220.ms);
    }

    return _IconBadge(type: type, style: style, compact: compact)
        .animate()
        .fadeIn(duration: 220.ms)
        .scale(begin: const Offset(.96, .96), duration: 220.ms);
  }
}

class _StateAssetVisual extends StatelessWidget {
  const _StateAssetVisual({required this.assetName, required this.fallback});

  final String assetName;
  final Widget fallback;

  @override
  Widget build(BuildContext context) {
    if (assetName.toLowerCase().endsWith('.svg')) {
      return SvgPicture.asset(
        assetName,
        fit: BoxFit.contain,
        placeholderBuilder: (_) => fallback,
        errorBuilder: (context, error, stackTrace) => fallback,
      );
    }

    return Image.asset(
      assetName,
      fit: BoxFit.contain,
      filterQuality: FilterQuality.medium,
      errorBuilder: (context, error, stackTrace) => fallback,
    );
  }
}

class _IconBadge extends StatelessWidget {
  const _IconBadge({
    required this.type,
    required this.style,
    required this.compact,
  });

  final NojposStateType type;
  final _StateStyle style;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: compact ? 68 : 88,
      height: compact ? 68 : 88,
      decoration: BoxDecoration(
        color: style.surface,
        shape: BoxShape.circle,
        border: Border.all(color: style.border),
        boxShadow: NojposShadow.soft,
      ),
      child: type == NojposStateType.loading
          ? Padding(
              padding: EdgeInsets.all(compact ? 20 : 27),
              child: CircularProgressIndicator(
                strokeWidth: 3,
                valueColor: AlwaysStoppedAnimation(style.color),
              ),
            )
          : Center(
              child: NojposAssetIcon(
                fallbackIcon: style.icon,
                size: compact ? 30 : 40,
                color: style.color,
              ),
            ),
    );
  }
}

class _StateStyle {
  const _StateStyle({
    required this.surface,
    required this.border,
    required this.color,
    required this.icon,
  });

  final Color surface;
  final Color border;
  final Color color;
  final IconData icon;
}

_StateStyle _styleFor(NojposStateType type) => switch (type) {
  NojposStateType.empty => const _StateStyle(
    surface: NojposColors.infoSurface,
    border: NojposColors.infoBorder,
    color: NojposColors.accent,
    icon: LucideIcons.packageOpen,
  ),
  NojposStateType.error => const _StateStyle(
    surface: NojposColors.dangerSurface,
    border: NojposColors.dangerBorder,
    color: NojposColors.danger,
    icon: LucideIcons.cloudOff,
  ),
  NojposStateType.warning => const _StateStyle(
    surface: NojposColors.warningSurface,
    border: NojposColors.warningBorder,
    color: Color(0xFF9A6B00),
    icon: LucideIcons.triangleAlert,
  ),
  NojposStateType.success => const _StateStyle(
    surface: NojposColors.successSurface,
    border: NojposColors.successBorder,
    color: Color(0xFF118A4A),
    icon: LucideIcons.circleCheck,
  ),
  NojposStateType.loading => const _StateStyle(
    surface: NojposColors.primarySoft,
    border: NojposColors.successBorder,
    color: NojposColors.primary,
    icon: LucideIcons.loaderCircle,
  ),
};
