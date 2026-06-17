import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../features/pos/models/cart_item.dart';
import '../../features/attendance/repositories/attendance_repository.dart';
import '../../features/auth/repositories/auth_repository.dart';
import '../../features/customers/repositories/customer_repository.dart';
import '../../features/inventory/repositories/inventory_repository.dart';
import '../../features/orders/repositories/order_repository.dart';
import '../../features/transactions/repositories/transaction_repository.dart';
import '../../shared/models/nojpos_models.dart';

final nojposSessionProvider =
    NotifierProvider<NojposSessionNotifier, NojposSessionState>(
      NojposSessionNotifier.new,
    );

class NojposSessionState {
  const NojposSessionState({
    required this.outlet,
    required this.cashier,
    required this.customers,
    required this.savedOrders,
    required this.transactions,
    required this.purchases,
    required this.attendance,
    this.activeOrderType = OrderType.dineIn,
    this.selectedCustomer,
    this.lastTransaction,
  });

  final Outlet outlet;
  final Employee cashier;
  final List<Customer> customers;
  final List<SalesOrder> savedOrders;
  final List<SalesTransaction> transactions;
  final List<InventoryPurchase> purchases;
  final List<AttendanceRecord> attendance;
  final OrderType activeOrderType;
  final Customer? selectedCustomer;
  final SalesTransaction? lastTransaction;

  NojposSessionState copyWith({
    Outlet? outlet,
    Employee? cashier,
    List<Customer>? customers,
    List<SalesOrder>? savedOrders,
    List<SalesTransaction>? transactions,
    List<InventoryPurchase>? purchases,
    List<AttendanceRecord>? attendance,
    OrderType? activeOrderType,
    Customer? selectedCustomer,
    bool clearSelectedCustomer = false,
    SalesTransaction? lastTransaction,
  }) {
    return NojposSessionState(
      outlet: outlet ?? this.outlet,
      cashier: cashier ?? this.cashier,
      customers: customers ?? this.customers,
      savedOrders: savedOrders ?? this.savedOrders,
      transactions: transactions ?? this.transactions,
      purchases: purchases ?? this.purchases,
      attendance: attendance ?? this.attendance,
      activeOrderType: activeOrderType ?? this.activeOrderType,
      selectedCustomer: clearSelectedCustomer
          ? null
          : selectedCustomer ?? this.selectedCustomer,
      lastTransaction: lastTransaction ?? this.lastTransaction,
    );
  }
}

class NojposSessionNotifier extends Notifier<NojposSessionState> {
  @override
  NojposSessionState build() {
    final authRepository = ref.read(authRepositoryProvider);
    final customerRepository = ref.read(customerRepositoryProvider);
    return NojposSessionState(
      outlet: authRepository.getDefaultOutlet(),
      cashier: authRepository.getDefaultCashier(),
      customers: customerRepository.getCustomers(),
      savedOrders: const [],
      transactions: const [],
      purchases: const [],
      attendance: const [],
    );
  }

  void selectCustomer(Customer? customer) {
    state = state.copyWith(
      selectedCustomer: customer,
      clearSelectedCustomer: customer == null,
    );
  }

  void selectOrderType(OrderType type) {
    state = state.copyWith(activeOrderType: type);
  }

  SalesOrder saveOrder({required List<CartItem> cartItems}) {
    final order = ref
        .read(orderRepositoryProvider)
        .createOrder(
          outletId: state.outlet.id,
          cartItems: cartItems,
          type: state.activeOrderType,
          status: OrderStatus.saved,
          sequence: state.savedOrders.length + 1,
          customer: state.selectedCustomer,
        );
    state = state.copyWith(savedOrders: [order, ...state.savedOrders]);
    return order;
  }

  SalesTransaction completeTransaction({
    required List<CartItem> cartItems,
    required PaymentMethod method,
    required int paidAmount,
  }) {
    final order = ref
        .read(orderRepositoryProvider)
        .createOrder(
          outletId: state.outlet.id,
          cartItems: cartItems,
          type: state.activeOrderType,
          status: OrderStatus.paid,
          sequence: state.transactions.length + 1,
          customer: state.selectedCustomer,
        );
    final payment = PaymentLine(
      method: method,
      amount: method == PaymentMethod.cash ? paidAmount : order.total,
    );
    final transaction = ref
        .read(transactionRepositoryProvider)
        .createTransaction(
          order: order,
          payments: [payment],
          cashier: state.cashier,
        );
    state = state.copyWith(
      transactions: [transaction, ...state.transactions],
      lastTransaction: transaction,
    );
    return transaction;
  }

  SalesOrder? activateSavedOrder(String orderId) {
    final index = state.savedOrders.indexWhere((order) => order.id == orderId);
    if (index == -1) return null;
    final order = state.savedOrders[index];
    state = state.copyWith(
      savedOrders: [
        for (final savedOrder in state.savedOrders)
          if (savedOrder.id != orderId) savedOrder,
      ],
      activeOrderType: order.type,
      selectedCustomer: order.customer,
      clearSelectedCustomer: order.customer == null,
    );
    return order;
  }

  InventoryPurchase addPurchase({
    required String supplierName,
    required int total,
  }) {
    final purchase = ref
        .read(inventoryRepositoryProvider)
        .createPurchase(
          sequence: state.purchases.length + 1,
          supplierName: supplierName,
          total: total,
        );
    state = state.copyWith(purchases: [purchase, ...state.purchases]);
    return purchase;
  }

  AttendanceRecord toggleAttendance(Employee employee) {
    final openIndex = state.attendance.indexWhere(
      (record) => record.employee.id == employee.id && record.isOpen,
    );
    if (openIndex == -1) {
      final record = ref.read(attendanceRepositoryProvider).clockIn(employee);
      state = state.copyWith(attendance: [record, ...state.attendance]);
      return record;
    }

    final current = state.attendance[openIndex];
    final closed = ref.read(attendanceRepositoryProvider).clockOut(current);
    state = state.copyWith(
      attendance: [
        for (final (index, record) in state.attendance.indexed)
          index == openIndex ? closed : record,
      ],
    );
    return closed;
  }
}
