enum OrderType {
  dineIn('Makan di Tempat'),
  delivery('Pengiriman'),
  online('Ojek Online'),
  quickService('Quick Service'),
  pickup('Ambil Sendiri');

  const OrderType(this.label);

  final String label;
}

enum PaymentMethod {
  cash('Tunai'),
  cashless('Nontunai'),
  transfer('Transfer'),
  qris('QRIS'),
  compliment('Komplimen'),
  deposit('Deposit');

  const PaymentMethod(this.label);

  final String label;
}

enum OrderStatus { active, saved, paid, canceled }

class Outlet {
  const Outlet({
    required this.id,
    required this.name,
    required this.isOnline,
    this.timezone = 'Asia/Jakarta',
    this.paymentMethods = const [],
    this.receiptConfig = const ReceiptConfig(),
  });

  final String id;
  final String name;
  final bool isOnline;
  final String timezone;
  final List<PaymentMethodConfig> paymentMethods;
  final ReceiptConfig receiptConfig;
}

enum ReceiptPaperWidth {
  mm58('58mm'),
  mm80('80mm');

  const ReceiptPaperWidth(this.label);

  final String label;

  static ReceiptPaperWidth fromJson(Object? value) {
    return value == '80mm' ? ReceiptPaperWidth.mm80 : ReceiptPaperWidth.mm58;
  }
}

class ReceiptConfig {
  const ReceiptConfig({
    this.paperWidth = ReceiptPaperWidth.mm58,
    this.header = '',
    this.footer = '',
    this.showLogo = false,
    this.showQrisInfo = false,
  });

  factory ReceiptConfig.fromJson(Object? value) {
    final json = _asMap(value);
    return ReceiptConfig(
      paperWidth: ReceiptPaperWidth.fromJson(json['paper_width']),
      header: (json['header'] ?? json['store_name'] ?? '').toString(),
      footer: (json['footer'] ?? json['footer_note'] ?? '').toString(),
      showLogo: json['show_logo'] as bool? ?? false,
      showQrisInfo: json['show_qris_info'] as bool? ?? false,
    );
  }

  final ReceiptPaperWidth paperWidth;
  final String header;
  final String footer;
  final bool showLogo;
  final bool showQrisInfo;
}

class PaymentMethodConfig {
  const PaymentMethodConfig({required this.method, required this.isCash});

  final String method;
  final bool isCash;
}

class Employee {
  const Employee({
    required this.id,
    required this.name,
    required this.role,
    this.pin = '',
    this.email = '',
  });

  final String id;
  final String name;
  final String role;
  final String pin;
  final String email;
}

class ProductCategory {
  const ProductCategory({required this.id, required this.name});

  final String id;
  final String name;
}

class Customer {
  const Customer({
    required this.id,
    required this.name,
    this.phone = '',
    this.group = 'Tanpa Grup',
  });

  final String id;
  final String name;
  final String phone;
  final String group;
}

class OrderLine {
  const OrderLine({
    required this.productId,
    required this.name,
    required this.quantity,
    required this.unitPrice,
    this.transactionItemId,
    this.discount = 0,
    this.note = '',
  });

  final String productId;
  final String? transactionItemId;
  final String name;
  final int quantity;
  final int unitPrice;
  final int discount;
  final String note;

  int get subtotal => quantity * unitPrice;

  int get total => (subtotal - discount).clamp(0, subtotal);
}

class SalesOrder {
  const SalesOrder({
    required this.id,
    required this.number,
    required this.type,
    required this.status,
    required this.lines,
    required this.createdAt,
    this.customer,
    this.note = '',
    this.discount = 0,
  });

  final String id;
  final String number;
  final OrderType type;
  final OrderStatus status;
  final List<OrderLine> lines;
  final DateTime createdAt;
  final Customer? customer;
  final String note;
  final int discount;

  int get subtotal => lines.fold(0, (sum, line) => sum + line.subtotal);

  int get total => (lines.fold(0, (sum, line) => sum + line.total) - discount)
      .clamp(0, subtotal);
}

class PaymentLine {
  PaymentLine({
    required this.method,
    required this.amount,
    String? methodName,
    bool? isCash,
    this.reference,
    this.status = 'confirmed',
    this.paymentId,
  }) : methodName = methodName ?? method.label,
       isCash = isCash ?? method == PaymentMethod.cash;

  final PaymentMethod method;
  final String methodName;
  final int amount;
  final bool isCash;
  final String? reference;
  final String status;
  final String? paymentId;
}

class SalesTransaction {
  const SalesTransaction({
    required this.id,
    required this.number,
    required this.order,
    required this.payments,
    required this.cashier,
    required this.createdAt,
    this.status = 'paid',
    this.itemDiscountTotal = 0,
    this.cartDiscountTotal = 0,
    this.grandTotal,
  });

  final String id;
  final String number;
  final SalesOrder order;
  final List<PaymentLine> payments;
  final Employee cashier;
  final DateTime createdAt;
  final String status;
  final int itemDiscountTotal;
  final int cartDiscountTotal;
  final int? grandTotal;

  int get paidAmount =>
      payments.fold(0, (sum, payment) => sum + payment.amount);

  int get cashPaidAmount => payments
      .where((payment) => payment.isCash)
      .fold(0, (sum, payment) => sum + payment.amount);

  int get nonCashPaidAmount => payments
      .where((payment) => !payment.isCash)
      .fold(0, (sum, payment) => sum + payment.amount);

  int get total => grandTotal ?? order.total;

  int get change {
    final cashDue = (total - nonCashPaidAmount).clamp(0, total);
    return (cashPaidAmount - cashDue).clamp(0, cashPaidAmount);
  }
}

class InventoryPurchase {
  const InventoryPurchase({
    required this.id,
    required this.number,
    required this.supplierName,
    required this.total,
    required this.createdAt,
  });

  factory InventoryPurchase.fromJson(Object? value) {
    final json = _asMap(value);
    return InventoryPurchase(
      id: (json['id'] as String?) ?? '',
      number: (json['number'] as String?) ?? '',
      supplierName: (json['supplier_name'] as String?) ?? '',
      total: (json['total'] as num?)?.toInt() ?? 0,
      createdAt:
          DateTime.tryParse(
            (json['purchased_at'] ?? json['created_at'] ?? '').toString(),
          ) ??
          DateTime.now(),
    );
  }

  final String id;
  final String number;
  final String supplierName;
  final int total;
  final DateTime createdAt;
}

class AttendanceRecord {
  const AttendanceRecord({
    required this.id,
    required this.employee,
    required this.clockInAt,
    this.clockOutAt,
  });

  final String id;
  final Employee employee;
  final DateTime clockInAt;
  final DateTime? clockOutAt;

  bool get isOpen => clockOutAt == null;
}

class ShiftPaymentTotal {
  const ShiftPaymentTotal({
    required this.method,
    required this.amount,
    this.isCash = false,
  });

  factory ShiftPaymentTotal.fromJson(Object? value) {
    final json = _asMap(value);
    return ShiftPaymentTotal(
      method: (json['method'] as String?) ?? '',
      amount: (json['amount'] as num?)?.toInt() ?? 0,
      isCash: json['is_cash'] as bool? ?? false,
    );
  }

  final String method;
  final int amount;
  final bool isCash;
}

class ShiftSession {
  const ShiftSession({
    required this.id,
    required this.cashier,
    required this.openedAt,
    this.closedAt,
    this.openingCash = 0,
    this.closingCash = 0,
    this.expectedCash,
    this.actualCash,
    this.cashDifference,
    this.paymentTotals = const [],
    this.status = 'open',
  });

  final String id;
  final Employee cashier;
  final DateTime openedAt;
  final DateTime? closedAt;
  final int openingCash;
  final int closingCash;
  final int? expectedCash;
  final int? actualCash;
  final int? cashDifference;
  final List<ShiftPaymentTotal> paymentTotals;
  final String status;

  bool get isOpen => status == 'open' && closedAt == null;
}

enum CashMovementType {
  cashIn('cash_in', 'Kas Masuk'),
  cashOut('cash_out', 'Kas Keluar');

  const CashMovementType(this.apiValue, this.label);

  final String apiValue;
  final String label;
}

class CashMovementRecord {
  const CashMovementRecord({
    required this.id,
    required this.type,
    required this.amount,
    required this.createdAt,
    this.reason = '',
  });

  final String id;
  final CashMovementType type;
  final int amount;
  final DateTime createdAt;
  final String reason;

  int get signedAmount => type == CashMovementType.cashIn ? amount : -amount;
}

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}
