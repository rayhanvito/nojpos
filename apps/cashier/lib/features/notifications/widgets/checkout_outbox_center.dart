import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../../../core/outbox/checkout_outbox.dart';

class CheckoutOutboxCenter extends StatelessWidget {
  const CheckoutOutboxCenter({
    required this.items,
    required this.onRetry,
    super.key,
  });

  final List<CheckoutOutboxItem> items;
  final ValueChanged<CheckoutOutboxItem> onRetry;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return const Center(
        child: Text(
          'Tidak ada checkout yang perlu dipulihkan.',
          style: TextStyle(color: MokposColors.muted),
        ),
      );
    }
    return ListView.separated(
      shrinkWrap: true,
      itemCount: items.length,
      separatorBuilder: (context, index) =>
          const Divider(height: 1, color: MokposColors.line),
      itemBuilder: (context, index) {
        final item = items[index];
        return ListTile(
          contentPadding: EdgeInsets.zero,
          leading: Icon(
            _iconFor(item.status),
            color: item.blocksClose ? MokposColors.danger : MokposColors.accent,
          ),
          title: Text(
            checkoutOutboxStatusLabel(item.status),
            style: const TextStyle(fontWeight: FontWeight.w900),
          ),
          subtitle: Text(
            [
              'Key: ${item.idempotencyKey}',
              'Percobaan kirim ulang: ${item.retryCount}',
              if (item.lastError != null) item.lastError!,
            ].join('\n'),
          ),
          trailing:
              item.status == CheckoutOutboxStatus.needsAction ||
                  item.status == CheckoutOutboxStatus.pending
              ? TextButton.icon(
                  onPressed: () => onRetry(item),
                  icon: const Icon(LucideIcons.refreshCw, size: 16),
                  label: const Text('Kirim ulang'),
                )
              : null,
        );
      },
    );
  }
}

String checkoutOutboxStatusLabel(CheckoutOutboxStatus status) {
  return switch (status) {
    CheckoutOutboxStatus.pending => 'Menunggu konfirmasi server',
    CheckoutOutboxStatus.sending => 'Mengirim ulang checkout',
    CheckoutOutboxStatus.sent => 'Checkout dikonfirmasi',
    CheckoutOutboxStatus.needsAction => 'Perlu tindakan kasir',
  };
}

IconData _iconFor(CheckoutOutboxStatus status) {
  return switch (status) {
    CheckoutOutboxStatus.pending => LucideIcons.clock,
    CheckoutOutboxStatus.sending => LucideIcons.loader,
    CheckoutOutboxStatus.sent => LucideIcons.circleCheck,
    CheckoutOutboxStatus.needsAction => LucideIcons.circleAlert,
  };
}
