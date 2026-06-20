import 'package:flutter/material.dart';

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
    final method = transaction.payments.isEmpty
        ? ''
        : transaction.payments.first.method.toLowerCase();
    final message = switch (transaction.status) {
      'payment_failed' ||
      'declined' => 'Pembayaran gagal. Pilih metode pembayaran lagi.',
      'expired' => 'Waktu pembayaran habis. Pilih metode pembayaran lagi.',
      _ when method.contains('qris') =>
        'Menunggu konfirmasi pembayaran QRIS...',
      _ when method.contains('edc') => 'Menunggu konfirmasi mesin EDC...',
      _ => 'Menunggu konfirmasi transfer...',
    };
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const CircularProgressIndicator(),
              const SizedBox(height: 20),
              Text(message, textAlign: TextAlign.center),
              if (errorMessage != null)
                Padding(
                  padding: const EdgeInsets.only(top: 12),
                  child: Text(errorMessage!),
                ),
              const SizedBox(height: 20),
              TextButton(
                onPressed: onCancel,
                child: const Text('Kembali ke pembayaran'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
