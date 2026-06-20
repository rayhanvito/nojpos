import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:share_plus/share_plus.dart';

final receiptShareServiceProvider = Provider<ReceiptShareService>(
  (ref) => const SharePlusReceiptShareService(),
);

abstract interface class ReceiptShareService {
  Future<void> shareReceipt(String receiptText);
}

class SharePlusReceiptShareService implements ReceiptShareService {
  const SharePlusReceiptShareService();

  @override
  Future<void> shareReceipt(String receiptText) async {
    if (receiptText.trim().isEmpty) {
      throw ArgumentError('Struk digital belum tersedia.');
    }
    await SharePlus.instance.share(
      ShareParams(text: receiptText, subject: 'Struk NojPOS'),
    );
  }
}
