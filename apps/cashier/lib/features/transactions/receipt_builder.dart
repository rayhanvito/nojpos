import '../../shared/models/nojpos_models.dart';
import '../pos/formatters.dart';

class ReceiptBuilder {
  const ReceiptBuilder();

  String build(SalesTransaction transaction, {Outlet? outlet}) {
    final config = outlet?.receiptConfig ?? const ReceiptConfig();
    final header = config.header.isEmpty
        ? outlet?.name ?? 'NojPOS'
        : config.header;
    final buffer = StringBuffer()
      ..writeln(header)
      ..writeln('Paper: ${config.paperWidth.label}');
    if (config.showLogo) {
      buffer.writeln('Logo: ditampilkan');
    }
    if (config.showQrisInfo) {
      buffer.writeln('QRIS: tersedia di kasir');
    }
    buffer
      ..writeln(transaction.number)
      ..writeln('Kasir: ${transaction.cashier.name}');
    final customer = transaction.order.customer;
    if (customer != null) {
      buffer.writeln('Pelanggan: ${customer.name}');
    }
    buffer.writeln('---');
    for (final line in transaction.order.lines) {
      buffer.writeln('${line.quantity}x ${line.name} ${rupiah(line.subtotal)}');
    }
    buffer
      ..writeln('---')
      ..writeln('Total: ${rupiah(transaction.order.total)}');
    for (final payment in transaction.payments) {
      buffer.writeln('${payment.methodName}: ${rupiah(payment.amount)}');
    }
    buffer.writeln('Kembalian: ${rupiah(transaction.change)}');
    if (config.footer.isNotEmpty) {
      buffer.writeln(config.footer);
    }
    return buffer.toString();
  }
}
