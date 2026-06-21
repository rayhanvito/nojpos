import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../app/theme.dart';

class NojposConfirmDialog extends StatelessWidget {
  const NojposConfirmDialog({
    super.key,
    required this.title,
    required this.subtitle,
    this.content,
    this.confirmLabel = 'Lanjutkan',
    this.cancelLabel = 'Batal',
    this.icon = LucideIcons.circleCheck,
    this.onConfirm,
  });

  final String title;
  final String subtitle;
  final Widget? content;
  final String confirmLabel;
  final String cancelLabel;
  final IconData icon;
  final VoidCallback? onConfirm;

  @override
  Widget build(BuildContext context) {
    return _NojposBaseDialog(
      title: title,
      subtitle: subtitle,
      content: content,
      icon: icon,
      iconColor: NojposColors.primary,
      iconSurface: NojposColors.primarySoft,
      confirmLabel: confirmLabel,
      cancelLabel: cancelLabel,
      confirmColor: NojposColors.primary,
      onConfirm: onConfirm ?? () => Navigator.of(context).pop(true),
    );
  }
}

class NojposDangerDialog extends StatelessWidget {
  const NojposDangerDialog({
    super.key,
    required this.title,
    required this.subtitle,
    this.content,
    this.confirmLabel = 'Ya, lanjutkan',
    this.cancelLabel = 'Batal',
    this.icon = LucideIcons.triangleAlert,
    this.onConfirm,
  });

  final String title;
  final String subtitle;
  final Widget? content;
  final String confirmLabel;
  final String cancelLabel;
  final IconData icon;
  final VoidCallback? onConfirm;

  @override
  Widget build(BuildContext context) {
    return _NojposBaseDialog(
      title: title,
      subtitle: subtitle,
      content: content,
      icon: icon,
      iconColor: NojposColors.danger,
      iconSurface: NojposColors.dangerSurface,
      confirmLabel: confirmLabel,
      cancelLabel: cancelLabel,
      confirmColor: NojposColors.danger,
      onConfirm: onConfirm ?? () => Navigator.of(context).pop(true),
    );
  }
}

class NojposWarningDialog extends StatelessWidget {
  const NojposWarningDialog({
    super.key,
    required this.title,
    required this.subtitle,
    this.content,
    this.confirmLabel = 'Lanjutkan',
    this.cancelLabel = 'Batal',
    this.icon = LucideIcons.triangleAlert,
    this.confirmEnabled = true,
    this.onConfirm,
  });

  final String title;
  final String subtitle;
  final Widget? content;
  final String confirmLabel;
  final String? cancelLabel;
  final IconData icon;
  final bool confirmEnabled;
  final VoidCallback? onConfirm;

  @override
  Widget build(BuildContext context) {
    return _NojposBaseDialog(
      title: title,
      subtitle: subtitle,
      content: content,
      icon: icon,
      iconColor: const Color(0xFF9A6B00),
      iconSurface: NojposColors.warningSurface,
      confirmLabel: confirmLabel,
      cancelLabel: cancelLabel,
      confirmColor: const Color(0xFF9A6B00),
      confirmEnabled: confirmEnabled,
      onConfirm: onConfirm ?? () => Navigator.of(context).pop(true),
    );
  }
}

class NojposInfoDialog extends StatelessWidget {
  const NojposInfoDialog({
    super.key,
    required this.title,
    required this.subtitle,
    this.content,
    this.confirmLabel = 'Mengerti',
    this.cancelLabel,
    this.icon = LucideIcons.info,
    this.onConfirm,
  });

  final String title;
  final String subtitle;
  final Widget? content;
  final String confirmLabel;
  final String? cancelLabel;
  final IconData icon;
  final VoidCallback? onConfirm;

  @override
  Widget build(BuildContext context) {
    return _NojposBaseDialog(
      title: title,
      subtitle: subtitle,
      content: content,
      icon: icon,
      iconColor: NojposColors.accent,
      iconSurface: NojposColors.infoSurface,
      confirmLabel: confirmLabel,
      cancelLabel: cancelLabel,
      confirmColor: NojposColors.accent,
      onConfirm: onConfirm ?? () => Navigator.of(context).pop(true),
    );
  }
}

Future<bool?> showNojposConfirmDialog(
  BuildContext context, {
  required String title,
  required String subtitle,
  Widget? content,
  String confirmLabel = 'Lanjutkan',
  String cancelLabel = 'Batal',
  bool barrierDismissible = true,
}) {
  return showDialog<bool>(
    context: context,
    barrierDismissible: barrierDismissible,
    builder: (_) => NojposConfirmDialog(
      title: title,
      subtitle: subtitle,
      content: content,
      confirmLabel: confirmLabel,
      cancelLabel: cancelLabel,
    ),
  );
}

Future<bool?> showNojposDangerDialog(
  BuildContext context, {
  required String title,
  required String subtitle,
  Widget? content,
  String confirmLabel = 'Ya, lanjutkan',
  String cancelLabel = 'Batal',
  bool barrierDismissible = true,
}) {
  return showDialog<bool>(
    context: context,
    barrierDismissible: barrierDismissible,
    builder: (_) => NojposDangerDialog(
      title: title,
      subtitle: subtitle,
      content: content,
      confirmLabel: confirmLabel,
      cancelLabel: cancelLabel,
    ),
  );
}

Future<bool?> showNojposWarningDialog(
  BuildContext context, {
  required String title,
  required String subtitle,
  Widget? content,
  String confirmLabel = 'Lanjutkan',
  String? cancelLabel = 'Batal',
  bool barrierDismissible = true,
}) {
  return showDialog<bool>(
    context: context,
    barrierDismissible: barrierDismissible,
    builder: (_) => NojposWarningDialog(
      title: title,
      subtitle: subtitle,
      content: content,
      confirmLabel: confirmLabel,
      cancelLabel: cancelLabel,
    ),
  );
}

Future<bool?> showNojposInfoDialog(
  BuildContext context, {
  required String title,
  required String subtitle,
  Widget? content,
  String confirmLabel = 'Mengerti',
  String? cancelLabel,
  bool barrierDismissible = true,
}) {
  return showDialog<bool>(
    context: context,
    barrierDismissible: barrierDismissible,
    builder: (_) => NojposInfoDialog(
      title: title,
      subtitle: subtitle,
      content: content,
      confirmLabel: confirmLabel,
      cancelLabel: cancelLabel,
    ),
  );
}

class _NojposBaseDialog extends StatelessWidget {
  const _NojposBaseDialog({
    required this.title,
    required this.subtitle,
    required this.icon,
    required this.iconColor,
    required this.iconSurface,
    required this.confirmLabel,
    required this.confirmColor,
    required this.onConfirm,
    this.cancelLabel,
    this.content,
    this.confirmEnabled = true,
  });

  final String title;
  final String subtitle;
  final IconData icon;
  final Color iconColor;
  final Color iconSurface;
  final String confirmLabel;
  final Color confirmColor;
  final VoidCallback onConfirm;
  final String? cancelLabel;
  final Widget? content;
  final bool confirmEnabled;

  @override
  Widget build(BuildContext context) {
    final size = MediaQuery.sizeOf(context);
    final isCompact = size.width < 420;
    final horizontalInset = isCompact ? 16.0 : 24.0;
    final dialogPadding = isCompact ? 18.0 : 24.0;

    return Dialog(
      insetPadding: EdgeInsets.symmetric(
        horizontal: horizontalInset,
        vertical: 24,
      ),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(NojposRadius.xl),
      ),
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: 440, maxHeight: size.height - 48),
        child: SingleChildScrollView(
          padding: EdgeInsets.all(dialogPadding),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    width: isCompact ? 48 : 54,
                    height: isCompact ? 48 : 54,
                    decoration: BoxDecoration(
                      color: iconSurface,
                      borderRadius: BorderRadius.circular(NojposRadius.lg),
                    ),
                    child: Icon(
                      icon,
                      color: iconColor,
                      size: isCompact ? 24 : 28,
                    ),
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          title,
                          style: TextStyle(
                            color: NojposColors.text,
                            fontSize: isCompact ? 18 : 20,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                        const SizedBox(height: 6),
                        Text(
                          subtitle,
                          style: const TextStyle(
                            color: NojposColors.muted,
                            height: 1.45,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              if (content != null) ...[const SizedBox(height: 20), content!],
              const SizedBox(height: 24),
              Wrap(
                spacing: 12,
                runSpacing: 10,
                alignment: WrapAlignment.end,
                children: [
                  if (cancelLabel != null)
                    SizedBox(
                      width: isCompact ? double.infinity : 160,
                      child: OutlinedButton(
                        onPressed: () => Navigator.of(context).pop(false),
                        child: Text(cancelLabel!),
                      ),
                    ),
                  SizedBox(
                    width: isCompact ? double.infinity : 180,
                    child: FilledButton(
                      onPressed: confirmEnabled ? onConfirm : null,
                      style: FilledButton.styleFrom(
                        backgroundColor: confirmColor,
                        foregroundColor: Colors.white,
                        minimumSize: const Size.fromHeight(48),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(NojposRadius.md),
                        ),
                      ),
                      child: Text(
                        confirmLabel,
                        textAlign: TextAlign.center,
                        style: const TextStyle(fontWeight: FontWeight.w900),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
