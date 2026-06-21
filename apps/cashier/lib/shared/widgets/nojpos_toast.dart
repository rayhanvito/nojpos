import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../app/theme.dart';

enum NojposToastType { success, error, warning, info }

class NojposToast {
  const NojposToast._();

  static void success(
    BuildContext context,
    String message, {
    String? description,
    String? actionLabel,
    VoidCallback? onAction,
  }) => _show(
    context,
    type: NojposToastType.success,
    message: message,
    description: description,
    actionLabel: actionLabel,
    onAction: onAction,
  );

  static void error(
    BuildContext context,
    String message, {
    String? description,
    String? actionLabel,
    VoidCallback? onAction,
  }) => _show(
    context,
    type: NojposToastType.error,
    message: message,
    description: description,
    actionLabel: actionLabel,
    onAction: onAction,
  );

  static void warning(
    BuildContext context,
    String message, {
    String? description,
    String? actionLabel,
    VoidCallback? onAction,
  }) => _show(
    context,
    type: NojposToastType.warning,
    message: message,
    description: description,
    actionLabel: actionLabel,
    onAction: onAction,
  );

  static void info(
    BuildContext context,
    String message, {
    String? description,
    String? actionLabel,
    VoidCallback? onAction,
  }) => _show(
    context,
    type: NojposToastType.info,
    message: message,
    description: description,
    actionLabel: actionLabel,
    onAction: onAction,
  );

  static void _show(
    BuildContext context, {
    required NojposToastType type,
    required String message,
    String? description,
    String? actionLabel,
    VoidCallback? onAction,
  }) {
    final messenger = ScaffoldMessenger.of(context);
    messenger.hideCurrentSnackBar();
    messenger.showSnackBar(
      SnackBar(
        elevation: 0,
        behavior: SnackBarBehavior.floating,
        backgroundColor: Colors.transparent,
        margin: const EdgeInsets.fromLTRB(18, 0, 18, 18),
        padding: EdgeInsets.zero,
        duration: const Duration(seconds: 4),
        content: _NojposToastCard(
          type: type,
          message: message,
          description: description,
          actionLabel: actionLabel,
          onAction: onAction == null
              ? null
              : () {
                  messenger.hideCurrentSnackBar();
                  onAction();
                },
        ),
      ),
    );
  }
}

class _NojposToastCard extends StatelessWidget {
  const _NojposToastCard({
    required this.type,
    required this.message,
    this.description,
    this.actionLabel,
    this.onAction,
  });

  final NojposToastType type;
  final String message;
  final String? description;
  final String? actionLabel;
  final VoidCallback? onAction;

  @override
  Widget build(BuildContext context) {
    final style = _styleFor(type);
    return DecoratedBox(
      decoration: BoxDecoration(
        color: style.surface,
        borderRadius: BorderRadius.circular(NojposRadius.lg),
        border: Border.all(color: style.border),
        boxShadow: NojposShadow.floating,
      ),
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 38,
              height: 38,
              decoration: BoxDecoration(
                color: Colors.white.withValues(alpha: .84),
                borderRadius: BorderRadius.circular(13),
              ),
              child: Icon(style.icon, color: style.color, size: 20),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    message,
                    style: const TextStyle(
                      color: NojposColors.text,
                      fontSize: 14,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  if (description != null) ...[
                    const SizedBox(height: 3),
                    Text(
                      description!,
                      style: const TextStyle(
                        color: NojposColors.muted,
                        fontSize: 12.5,
                        height: 1.35,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            if (actionLabel != null && onAction != null) ...[
              const SizedBox(width: 8),
              TextButton(onPressed: onAction, child: Text(actionLabel!)),
            ],
          ],
        ),
      ),
    );
  }
}

class _ToastStyle {
  const _ToastStyle({
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

_ToastStyle _styleFor(NojposToastType type) => switch (type) {
  NojposToastType.success => const _ToastStyle(
    surface: NojposColors.successSurface,
    border: NojposColors.successBorder,
    color: Color(0xFF118A4A),
    icon: LucideIcons.circleCheck,
  ),
  NojposToastType.error => const _ToastStyle(
    surface: NojposColors.dangerSurface,
    border: NojposColors.dangerBorder,
    color: NojposColors.danger,
    icon: LucideIcons.circleX,
  ),
  NojposToastType.warning => const _ToastStyle(
    surface: NojposColors.warningSurface,
    border: NojposColors.warningBorder,
    color: Color(0xFF9A6B00),
    icon: LucideIcons.triangleAlert,
  ),
  NojposToastType.info => const _ToastStyle(
    surface: NojposColors.infoSurface,
    border: NojposColors.infoBorder,
    color: NojposColors.accent,
    icon: LucideIcons.info,
  ),
};
