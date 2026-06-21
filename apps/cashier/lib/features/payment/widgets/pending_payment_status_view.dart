import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/nojpos_assets.dart';
import '../../../app/theme.dart';
import '../../../shared/widgets/nojpos_asset_icon.dart';
import '../../../shared/widgets/nojpos_result_hero.dart';
import '../../pos/formatters.dart';
import '../../transactions/repositories/transaction_repository.dart';

class PendingPaymentStatusView extends StatelessWidget {
  const PendingPaymentStatusView({
    super.key,
    required this.transaction,
    this.errorMessage,
    required this.onCancel,
  });

  final CheckoutTransaction transaction;
  final String? errorMessage;
  final VoidCallback onCancel;

  @override
  Widget build(BuildContext context) {
    final payment = transaction.payments.isEmpty
        ? null
        : transaction.payments.first;
    final method = payment?.method.toLowerCase() ?? '';
    final state = _statusFor(transaction.status);
    final methodLabel = _methodLabel(method);
    final message = switch (transaction.status) {
      'payment_failed' ||
      'declined' => 'Pembayaran gagal. Pilih metode pembayaran lagi.',
      'expired' => 'Waktu pembayaran habis. Pilih metode pembayaran lagi.',
      _ when method.contains('qris') =>
        'Menunggu konfirmasi pembayaran QRIS...',
      _ when method.contains('edc') => 'Menunggu konfirmasi mesin EDC...',
      _ when method.contains('transfer') => 'Menunggu konfirmasi transfer...',
      _ => 'Menunggu konfirmasi pembayaran dari sistem.',
    };

    return Scaffold(
      backgroundColor: NojposColors.canvas,
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 620),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: Column(
                children: [
                  _heroFor(
                    state,
                    title: state.title,
                    subtitle: message,
                    amount: rupiah(transaction.grandTotal),
                  ),
                  const SizedBox(height: 18),
                  _PaymentMethodCard(
                    methodLabel: methodLabel,
                    method: method,
                    amount: payment?.amount ?? transaction.grandTotal,
                    reference: payment?.reference,
                    status: transaction.status,
                  ),
                  if (errorMessage != null) ...[
                    const SizedBox(height: 14),
                    _InlineError(message: errorMessage!),
                  ],
                  const SizedBox(height: 8),
                ],
              ),
            ),
          ),
        ),
      ),
      bottomNavigationBar: _PendingPaymentBottomBar(onCancel: onCancel),
    );
  }

  NojposResultHero _heroFor(
    _PendingVisualState state, {
    required String title,
    required String subtitle,
    required String amount,
  }) {
    return switch (state.type) {
      _PendingType.failed => NojposResultHero.failed(
        title: title,
        subtitle: subtitle,
        illustrationAsset: NojposAssets.paymentFailed,
        amount: amount,
        compact: true,
      ),
      _PendingType.warning => NojposResultHero.warning(
        title: title,
        subtitle: subtitle,
        illustrationAsset: NojposAssets.paymentFailed,
        amount: amount,
        compact: true,
      ),
      _PendingType.pending => NojposResultHero.pending(
        title: title,
        subtitle: subtitle,
        illustrationAsset: NojposAssets.loadingData,
        amount: amount,
        compact: true,
      ),
    };
  }
}

class _PendingPaymentBottomBar extends StatelessWidget {
  const _PendingPaymentBottomBar({required this.onCancel});

  final VoidCallback onCancel;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      top: false,
      child: Container(
        padding: const EdgeInsets.fromLTRB(20, 12, 20, 16),
        decoration: const BoxDecoration(
          color: Colors.white,
          border: Border(top: BorderSide(color: NojposColors.line)),
        ),
        child: Center(
          heightFactor: 1,
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 620),
            child: OutlinedButton.icon(
              onPressed: onCancel,
              icon: const Icon(LucideIcons.arrowLeft, size: 18),
              label: const Text('Kembali ke pembayaran'),
              style: OutlinedButton.styleFrom(
                minimumSize: const Size.fromHeight(52),
                foregroundColor: NojposColors.text,
                side: const BorderSide(color: NojposColors.line),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(NojposRadius.lg),
                ),
                textStyle: const TextStyle(fontWeight: FontWeight.w900),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _PaymentMethodCard extends StatelessWidget {
  const _PaymentMethodCard({
    required this.methodLabel,
    required this.method,
    required this.amount,
    required this.status,
    this.reference,
  });

  final String methodLabel;
  final String method;
  final int amount;
  final String status;
  final String? reference;

  @override
  Widget build(BuildContext context) {
    final icon = _iconFor(method);
    final assetName = _assetNameFor(method);
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(NojposRadius.xl),
        border: Border.all(color: NojposColors.line),
        boxShadow: NojposShadow.card,
      ),
      child: Row(
        children: [
          Container(
            width: 58,
            height: 58,
            decoration: BoxDecoration(
              color: NojposColors.primarySoft,
              borderRadius: BorderRadius.circular(NojposRadius.lg),
            ),
            child: NojposAssetIcon.named(
              assetName,
              fallbackIcon: icon,
              size: 30,
              applyColor: false,
              semanticLabel: methodLabel,
            ),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  methodLabel,
                  style: const TextStyle(
                    color: NojposColors.text,
                    fontWeight: FontWeight.w900,
                    fontSize: 16,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  reference?.trim().isNotEmpty ?? false
                      ? 'Ref: $reference'
                      : 'Status: ${_statusLabel(status)}',
                  style: const TextStyle(
                    color: NojposColors.muted,
                    fontSize: 13,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 12),
          Text(
            rupiah(amount),
            style: const TextStyle(
              color: NojposColors.text,
              fontWeight: FontWeight.w900,
              fontSize: 16,
            ),
          ),
        ],
      ),
    );
  }
}

class _InlineError extends StatelessWidget {
  const _InlineError({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: NojposColors.dangerSurface,
        borderRadius: BorderRadius.circular(NojposRadius.lg),
        border: Border.all(color: NojposColors.dangerBorder),
      ),
      child: Row(
        children: [
          const Icon(LucideIcons.triangleAlert, color: NojposColors.danger),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              message,
              style: const TextStyle(
                color: NojposColors.danger,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

enum _PendingType { pending, failed, warning }

class _PendingVisualState {
  const _PendingVisualState({required this.type, required this.title});

  final _PendingType type;
  final String title;
}

_PendingVisualState _statusFor(String status) => switch (status) {
  'payment_failed' || 'declined' => const _PendingVisualState(
    type: _PendingType.failed,
    title: 'Pembayaran Gagal',
  ),
  'expired' => const _PendingVisualState(
    type: _PendingType.warning,
    title: 'Pembayaran Kedaluwarsa',
  ),
  _ => const _PendingVisualState(
    type: _PendingType.pending,
    title: 'Menunggu Pembayaran',
  ),
};

IconData _iconFor(String method) {
  if (method.contains('qris') || method.contains('qr')) {
    return LucideIcons.qrCode;
  }
  if (method.contains('edc') || method.contains('card')) {
    return LucideIcons.creditCard;
  }
  if (method.contains('cash') || method.contains('tunai')) {
    return LucideIcons.banknote;
  }
  return LucideIcons.landmark;
}

String _assetNameFor(String method) {
  if (method.contains('qris') || method.contains('qr')) return 'qris';
  if (method.contains('edc') || method.contains('card')) return 'edc';
  if (method.contains('cash') || method.contains('tunai')) return 'cash';
  return 'receipt';
}

String _methodLabel(String method) {
  if (method.contains('qris') || method.contains('qr')) return 'QRIS';
  if (method.contains('edc') || method.contains('card')) return 'EDC / Kartu';
  if (method.contains('cash') || method.contains('tunai')) return 'Tunai';
  return 'Transfer';
}

String _statusLabel(String status) => switch (status) {
  'payment_pending' => 'Menunggu konfirmasi',
  'payment_failed' => 'Gagal',
  'declined' => 'Ditolak',
  'expired' => 'Kedaluwarsa',
  _ => status,
};
