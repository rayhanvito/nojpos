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
  const Outlet({required this.id, required this.name, required this.isOnline});

  final String id;
  final String name;
  final bool isOnline;
}

class Employee {
  const Employee({
    required this.id,
    required this.name,
    required this.role,
    required this.pin,
  });

  final String id;
  final String name;
  final String role;
  final String pin;
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
    this.note = '',
  });

  final String productId;
  final String name;
  final int quantity;
  final int unitPrice;
  final String note;

  int get subtotal => quantity * unitPrice;
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

  int get total => (subtotal - discount).clamp(0, subtotal);
}

class PaymentLine {
  const PaymentLine({required this.method, required this.amount});

  final PaymentMethod method;
  final int amount;
}

class SalesTransaction {
  const SalesTransaction({
    required this.id,
    required this.number,
    required this.order,
    required this.payments,
    required this.cashier,
    required this.createdAt,
  });

  final String id;
  final String number;
  final SalesOrder order;
  final List<PaymentLine> payments;
  final Employee cashier;
  final DateTime createdAt;

  int get paidAmount =>
      payments.fold(0, (sum, payment) => sum + payment.amount);

  int get change => (paidAmount - order.total).clamp(0, paidAmount);
}

class InventoryPurchase {
  const InventoryPurchase({
    required this.id,
    required this.number,
    required this.supplierName,
    required this.total,
    required this.createdAt,
  });

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

class ShiftSession {
  const ShiftSession({
    required this.id,
    required this.cashier,
    required this.openedAt,
    this.closedAt,
    this.openingCash = 0,
    this.closingCash = 0,
  });

  final String id;
  final Employee cashier;
  final DateTime openedAt;
  final DateTime? closedAt;
  final int openingCash;
  final int closingCash;

  bool get isOpen => closedAt == null;
}
