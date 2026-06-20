import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../core/outbox/checkout_outbox.dart';
import '../../features/auth/repositories/auth_repository.dart';
import '../../features/customers/repositories/customer_repository.dart';
import '../../features/inventory/repositories/inventory_repository.dart';
import '../../features/orders/repositories/order_repository.dart';
import '../../features/pos/models/cart_item.dart';
import '../../features/shift/repositories/shift_repository.dart';
import '../../features/transactions/repositories/transaction_repository.dart';
import '../../shared/models/nojpos_models.dart';

final nojposSessionProvider =
    NotifierProvider<NojposSessionNotifier, NojposSessionState>(
      NojposSessionNotifier.new,
    );

enum SessionStatus {
  booting,
  unauthenticated,
  outletRequired,
  pinRequired,
  ready,
}

class NojposSessionState {
  const NojposSessionState({
    required this.outlet,
    required this.cashier,
    this.account,
    required this.customers,
    required this.savedOrders,
    required this.transactions,
    required this.purchases,
    this.cashMovements = const [],
    this.inventory,
    this.outlets = const [],
    this.businessName = '',
    this.deviceId,
    this.deviceUuid,
    this.activeShift,
    this.lastClosedShift,
    this.status = SessionStatus.booting,
    this.activeOrderType = OrderType.dineIn,
    this.selectedCustomer,
    this.lastTransaction,
    this.isBusy = false,
    this.errorMessage,
  });

  factory NojposSessionState.initial() {
    return const NojposSessionState(
      outlet: Outlet(id: '', name: 'Pilih outlet', isOnline: false),
      cashier: Employee(id: '', name: 'Belum pilih kasir', role: 'cashier'),
      customers: [],
      savedOrders: [],
      transactions: [],
      purchases: [],
      cashMovements: [],
    );
  }

  final Outlet outlet;
  final Employee cashier;
  final Employee? account;
  final List<Outlet> outlets;
  final String businessName;
  final String? deviceId;
  final String? deviceUuid;
  final ShiftSession? activeShift;
  final ShiftSession? lastClosedShift;
  final List<Customer> customers;
  final List<SalesOrder> savedOrders;
  final List<SalesTransaction> transactions;
  final List<InventoryPurchase> purchases;
  final List<CashMovementRecord> cashMovements;
  final InventorySnapshot? inventory;
  final SessionStatus status;
  final OrderType activeOrderType;
  final Customer? selectedCustomer;
  final SalesTransaction? lastTransaction;
  final bool isBusy;
  final String? errorMessage;

  bool get hasOpenShift => activeShift?.isOpen ?? false;

  bool get canManageMasterData {
    final role = (account ?? cashier).role;
    return role == 'owner' || role == 'admin';
  }

  int get cashSummary {
    final openingCash = activeShift?.openingCash ?? 0;
    return openingCash +
        cashMovements.fold(0, (sum, movement) => sum + movement.signedAmount);
  }

  NojposSessionState copyWith({
    Outlet? outlet,
    Employee? cashier,
    Employee? account,
    bool clearAccount = false,
    List<Outlet>? outlets,
    String? businessName,
    String? deviceId,
    String? deviceUuid,
    ShiftSession? activeShift,
    bool clearActiveShift = false,
    ShiftSession? lastClosedShift,
    bool clearLastClosedShift = false,
    List<Customer>? customers,
    List<SalesOrder>? savedOrders,
    List<SalesTransaction>? transactions,
    List<InventoryPurchase>? purchases,
    List<CashMovementRecord>? cashMovements,
    InventorySnapshot? inventory,
    SessionStatus? status,
    OrderType? activeOrderType,
    Customer? selectedCustomer,
    bool clearSelectedCustomer = false,
    SalesTransaction? lastTransaction,
    bool? isBusy,
    String? errorMessage,
    bool clearError = false,
  }) {
    return NojposSessionState(
      outlet: outlet ?? this.outlet,
      cashier: cashier ?? this.cashier,
      account: clearAccount ? null : account ?? this.account,
      outlets: outlets ?? this.outlets,
      businessName: businessName ?? this.businessName,
      deviceId: deviceId ?? this.deviceId,
      deviceUuid: deviceUuid ?? this.deviceUuid,
      activeShift: clearActiveShift ? null : activeShift ?? this.activeShift,
      lastClosedShift: clearLastClosedShift
          ? null
          : lastClosedShift ?? this.lastClosedShift,
      customers: customers ?? this.customers,
      savedOrders: savedOrders ?? this.savedOrders,
      transactions: transactions ?? this.transactions,
      purchases: purchases ?? this.purchases,
      cashMovements: cashMovements ?? this.cashMovements,
      inventory: inventory ?? this.inventory,
      status: status ?? this.status,
      activeOrderType: activeOrderType ?? this.activeOrderType,
      selectedCustomer: clearSelectedCustomer
          ? null
          : selectedCustomer ?? this.selectedCustomer,
      lastTransaction: lastTransaction ?? this.lastTransaction,
      isBusy: isBusy ?? this.isBusy,
      errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
    );
  }
}

class NojposSessionNotifier extends Notifier<NojposSessionState> {
  @override
  NojposSessionState build() {
    return NojposSessionState.initial();
  }

  Future<void> bootstrap() async {
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final session = await ref.read(authRepositoryProvider).restoreSession();
      if (session == null) {
        state = state.copyWith(
          status: SessionStatus.unauthenticated,
          isBusy: false,
        );
        return;
      }
      _applyAuthSession(session);
      state = state.copyWith(
        status: state.outlet.id.isEmpty
            ? SessionStatus.outletRequired
            : SessionStatus.pinRequired,
        isBusy: false,
      );
    } catch (error) {
      state = state.copyWith(
        status: SessionStatus.unauthenticated,
        isBusy: false,
        errorMessage: _messageFor(error),
      );
    }
  }

  Future<void> login({
    required String email,
    required String password,
    required String deviceUuid,
  }) async {
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final session = await ref
          .read(authRepositoryProvider)
          .login(email: email, password: password, deviceUuid: deviceUuid);
      _applyAuthSession(session);
      if (session.outlets.length == 1) {
        await selectOutlet(session.outlets.single);
        state = state.copyWith(isBusy: false);
        return;
      }
      state = state.copyWith(
        status: state.outlet.id.isEmpty
            ? SessionStatus.outletRequired
            : SessionStatus.pinRequired,
        isBusy: false,
      );
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
    }
  }

  Future<void> selectOutlet(Outlet outlet) async {
    await ref.read(authRepositoryProvider).selectOutlet(outlet);
    state = state.copyWith(outlet: outlet, status: SessionStatus.pinRequired);
  }

  Future<void> refreshOutlets() async {
    try {
      final outlets = await ref.read(authRepositoryProvider).listOutlets();
      if (outlets.isEmpty) return;
      final selectedOutlet = outlets.where((outlet) {
        return outlet.id == state.outlet.id;
      }).firstOrNull;
      if (selectedOutlet != null) {
        state = state.copyWith(outlets: outlets, outlet: selectedOutlet);
        return;
      }
      state = state.copyWith(
        outlets: outlets,
        status: SessionStatus.outletRequired,
      );
    } catch (error) {
      if (state.outlets.isEmpty) {
        state = state.copyWith(errorMessage: _messageFor(error));
      }
    }
  }

  Future<bool> pinSwitch(String pin) async {
    final deviceId = state.deviceId;
    if (deviceId == null || deviceId.isEmpty || state.outlet.id.isEmpty) {
      state = state.copyWith(errorMessage: 'Device atau outlet belum siap.');
      return false;
    }

    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final cashier = await ref
          .read(authRepositoryProvider)
          .pinSwitch(pin: pin, deviceId: deviceId, outletId: state.outlet.id);
      state = state.copyWith(
        cashier: cashier,
        status: SessionStatus.ready,
        isBusy: false,
      );
      return true;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return false;
    }
  }

  Future<void> refreshCurrentShift() async {
    final deviceId = state.deviceId;
    if (deviceId == null || state.outlet.id.isEmpty) return;
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final shift = await ref
          .read(shiftRepositoryProvider)
          .currentShift(outletId: state.outlet.id, deviceId: deviceId);
      state = state.copyWith(
        activeShift: shift,
        clearActiveShift: shift == null,
        isBusy: false,
      );
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
    }
  }

  Future<bool> openShift({required int openingCash}) async {
    final deviceId = state.deviceId;
    if (deviceId == null ||
        state.outlet.id.isEmpty ||
        state.cashier.id.isEmpty) {
      state = state.copyWith(
        errorMessage: 'Kasir, outlet, atau device belum siap.',
      );
      return false;
    }
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final shift = await ref
          .read(shiftRepositoryProvider)
          .openShift(
            outletId: state.outlet.id,
            deviceId: deviceId,
            cashierId: state.cashier.id,
            openingCash: openingCash,
          );
      state = state.copyWith(
        activeShift: shift,
        clearLastClosedShift: true,
        isBusy: false,
      );
      return true;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return false;
    }
  }

  Future<bool> closeShift({
    required int actualCash,
    required String pin,
    String? varianceReason,
  }) async {
    final shiftId = state.activeShift?.id;
    if (shiftId == null) return false;
    final blockingOutboxCount = await ref
        .read(checkoutOutboxStoreProvider)
        .blockingItemCountForShift(shiftId);
    if (blockingOutboxCount > 0) {
      state = state.copyWith(
        errorMessage:
            'Tutup shift ditolak karena ada $blockingOutboxCount checkout yang belum terselesaikan.',
      );
      return false;
    }
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final shift = await ref
          .read(shiftRepositoryProvider)
          .closeShift(
            shiftId: shiftId,
            actualCash: actualCash,
            pin: pin,
            varianceReason: varianceReason,
          );
      state = state.copyWith(
        lastClosedShift: shift,
        clearActiveShift: true,
        isBusy: false,
      );
      return true;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return false;
    }
  }

  Future<bool> addCashMovement({
    required CashMovementType type,
    required int amount,
    String? reason,
  }) async {
    final shiftId = state.activeShift?.id;
    if (shiftId == null || amount <= 0) {
      state = state.copyWith(
        errorMessage: 'Shift aktif dan nominal kas wajib diisi.',
      );
      return false;
    }
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      await ref
          .read(shiftRepositoryProvider)
          .cashMovement(
            shiftId: shiftId,
            type: type.apiValue,
            amount: amount,
            reason: reason,
          );
      final movement = CashMovementRecord(
        id: 'cash-${DateTime.now().microsecondsSinceEpoch}',
        type: type,
        amount: amount,
        reason: reason?.trim() ?? '',
        createdAt: DateTime.now(),
      );
      state = state.copyWith(
        cashMovements: [movement, ...state.cashMovements],
        isBusy: false,
      );
      return true;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return false;
    }
  }

  Future<void> retryCheckoutOutboxItem(CheckoutOutboxItem item) async {
    final transaction = await ref
        .read(checkoutOutboxControllerProvider.notifier)
        .retryNow(item.id);
    if (transaction == null) return;
    final order = SalesOrder(
      id: 'retry-${item.id}',
      number: transaction.number,
      type: state.activeOrderType,
      status: OrderStatus.paid,
      lines: [
        for (final draftItem in item.draft.items)
          OrderLine(
            productId: draftItem.productId,
            name: draftItem.name,
            quantity: draftItem.quantity,
            unitPrice: draftItem.unitPrice,
          ),
      ],
      createdAt: DateTime.now(),
      customer: state.selectedCustomer,
    );
    final payments = [
      for (final payment in item.draft.payments)
        PaymentLine(
          method: payment.isCash ? PaymentMethod.cash : PaymentMethod.cashless,
          amount: payment.amount,
        ),
    ];
    final salesTransaction = SalesTransaction(
      id: transaction.id,
      number: transaction.number,
      order: order,
      payments: payments,
      cashier: state.cashier,
      createdAt: DateTime.now(),
    );
    state = state.copyWith(
      transactions: [salesTransaction, ...state.transactions],
      lastTransaction: salesTransaction,
    );
  }

  Future<void> logout() async {
    await ref.read(authRepositoryProvider).logout();
    state = NojposSessionState.initial().copyWith(
      status: SessionStatus.unauthenticated,
      customers: state.customers,
      clearAccount: true,
    );
  }

  void lock() {
    state = state.copyWith(
      cashier: const Employee(
        id: '',
        name: 'Belum pilih kasir',
        role: 'cashier',
      ),
      status: SessionStatus.pinRequired,
      clearError: true,
    );
  }

  void selectCustomer(Customer? customer) {
    state = state.copyWith(
      selectedCustomer: customer,
      clearSelectedCustomer: customer == null,
    );
  }

  Future<void> searchCustomers({String? search, String? group}) async {
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final customers = await ref
          .read(customerRepositoryProvider)
          .searchCustomers(search: search, group: group);
      state = state.copyWith(customers: customers, isBusy: false);
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
    }
  }

  Future<Customer?> createCustomer({
    required String name,
    String? phone,
    String? group,
  }) async {
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final customer = await ref
          .read(customerRepositoryProvider)
          .createCustomer(name: name, phone: phone, group: group);
      state = state.copyWith(
        customers: [customer, ...state.customers],
        selectedCustomer: customer,
        isBusy: false,
      );
      return customer;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return null;
    }
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

  Future<CheckoutSubmitResult> completeTransaction({
    required List<CartItem> cartItems,
    required PaymentMethod method,
    required int paidAmount,
    int cartDiscount = 0,
    String notes = '',
    Employee? servedBy,
  }) async {
    return completeCheckout(
      cartItems: cartItems,
      payments: [
        CheckoutPayment(
          method: method == PaymentMethod.cash ? 'Tunai' : method.label,
          amount: paidAmount,
          isCash: method == PaymentMethod.cash,
        ),
      ],
      paidAmount: paidAmount,
      cartDiscount: cartDiscount,
      notes: notes,
      servedBy: servedBy,
    );
  }

  Future<CheckoutQuote> quoteCheckout({
    required List<CartItem> cartItems,
    int cartDiscount = 0,
    List<String> promotionCodes = const [],
    String notes = '',
    Employee? servedBy,
  }) async {
    final shiftId = state.activeShift?.id;
    final deviceId = state.deviceId;
    if (shiftId == null || deviceId == null) {
      throw StateError('Shift belum dibuka.');
    }

    final draft = CheckoutDraft(
      idempotencyKey: 'quote',
      outletId: state.outlet.id,
      deviceId: deviceId,
      cashierId: state.cashier.id,
      shiftId: shiftId,
      items: [
        for (final item in cartItems)
          CheckoutItem(
            productId: item.product.id,
            name: item.product.name,
            quantity: item.quantity,
            unitPrice: item.product.price,
            discount: item.discount,
          ),
      ],
      payments: const [],
      paidAmount: 0,
      customerId: state.selectedCustomer?.id,
      servedBy: servedBy?.id,
      cartDiscount: cartDiscount,
      promotionCodes: promotionCodes,
      notes: notes,
    );

    return ref.read(transactionRepositoryProvider).quoteTransaction(draft);
  }

  Future<CheckoutSubmitResult> completeCheckout({
    required List<CartItem> cartItems,
    required List<CheckoutPayment> payments,
    required int paidAmount,
    int cartDiscount = 0,
    List<String> promotionCodes = const [],
    String notes = '',
    Employee? servedBy,
    CheckoutQuote? quote,
    String? status,
  }) async {
    final shiftId = state.activeShift?.id;
    final deviceId = state.deviceId;
    if (shiftId == null || deviceId == null) {
      throw StateError('Shift belum dibuka.');
    }
    final draft = CheckoutDraft(
      idempotencyKey: quote?.checkoutIdempotencyKey.isNotEmpty == true
          ? quote!.checkoutIdempotencyKey
          : const Uuid().v4(),
      outletId: state.outlet.id,
      deviceId: deviceId,
      cashierId: state.cashier.id,
      shiftId: shiftId,
      items: [
        for (final item in cartItems)
          CheckoutItem(
            productId: item.product.id,
            name: item.product.name,
            quantity: item.quantity,
            unitPrice: item.product.price,
            discount: item.discount,
          ),
      ],
      payments: payments,
      paidAmount: paidAmount,
      customerId: state.selectedCustomer?.id,
      servedBy: servedBy?.id,
      cartDiscount: cartDiscount,
      promotionCodes: promotionCodes,
      quote: quote,
      notes: notes,
      status: status,
    );
    final result = await ref
        .read(checkoutOutboxControllerProvider.notifier)
        .submitOrQueue(draft);
    if (result is SentCheckoutResult) {
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
      final paidOrder = SalesOrder(
        id: order.id,
        number: result.transaction.number.isEmpty
            ? order.number
            : result.transaction.number,
        type: order.type,
        status: OrderStatus.paid,
        lines: order.lines,
        createdAt: order.createdAt,
        customer: order.customer,
        note: order.note,
        discount:
            result.transaction.itemDiscountTotal +
            result.transaction.cartDiscountTotal,
      );
      final localPayments = result.transaction.payments.isNotEmpty
          ? [
              for (final payment in result.transaction.payments)
                PaymentLine(
                  method: payment.isCash
                      ? PaymentMethod.cash
                      : PaymentMethod.cashless,
                  methodName: payment.method,
                  amount: payment.amount,
                  isCash: payment.isCash,
                  reference: payment.reference,
                  status: payment.status,
                  paymentId: payment.id,
                ),
            ]
          : [
              for (final payment in payments)
                PaymentLine(
                  method: payment.isCash
                      ? PaymentMethod.cash
                      : PaymentMethod.cashless,
                  methodName: payment.method,
                  amount: payment.amount,
                  isCash: payment.isCash,
                  reference: payment.reference,
                  status: payment.isCash ? 'confirmed' : 'pending',
                ),
            ];
      final transaction = ref
          .read(transactionRepositoryProvider)
          .createLocalTransaction(
            order: paidOrder,
            payments: localPayments,
            cashier: state.cashier,
          );
      final apiTransaction = SalesTransaction(
        id: result.transaction.id,
        number: result.transaction.number,
        order: paidOrder,
        payments: localPayments,
        cashier: state.cashier,
        createdAt: transaction.createdAt,
        status: result.transaction.status,
        itemDiscountTotal: result.transaction.itemDiscountTotal,
        cartDiscountTotal: result.transaction.cartDiscountTotal,
        grandTotal: result.transaction.grandTotal,
      );
      state = state.copyWith(
        transactions: [apiTransaction, ...state.transactions],
        lastTransaction: apiTransaction,
      );
    }
    return result;
  }

  Future<CheckoutTransaction> saveHeldTransaction({
    required List<CartItem> cartItems,
    int cartDiscount = 0,
    String notes = '',
    Employee? servedBy,
  }) async {
    final shiftId = state.activeShift?.id;
    final deviceId = state.deviceId;
    if (shiftId == null || deviceId == null) {
      throw StateError('Shift belum dibuka.');
    }
    final draft = CheckoutDraft(
      idempotencyKey: const Uuid().v4(),
      outletId: state.outlet.id,
      deviceId: deviceId,
      cashierId: state.cashier.id,
      shiftId: shiftId,
      items: [
        for (final item in cartItems)
          CheckoutItem(
            productId: item.product.id,
            name: item.product.name,
            quantity: item.quantity,
            unitPrice: item.product.price,
            discount: item.discount,
          ),
      ],
      payments: const [],
      paidAmount: 0,
      customerId: state.selectedCustomer?.id,
      servedBy: servedBy?.id,
      cartDiscount: cartDiscount,
      notes: notes,
      status: 'held',
    );
    final transaction = await ref
        .read(transactionRepositoryProvider)
        .createParkedOrder(draft);
    final order = _orderFromServerTransaction(
      transaction,
      fallbackItems: draft.items,
      fallbackCustomer: state.selectedCustomer,
      status: OrderStatus.saved,
    );
    state = state.copyWith(savedOrders: [order, ...state.savedOrders]);
    return transaction;
  }

  Future<void> loadHeldTransactions() async {
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final transactions = await ref
          .read(transactionRepositoryProvider)
          .listParkedOrders();
      state = state.copyWith(
        savedOrders: [
          for (final transaction in transactions)
            _orderFromServerTransaction(transaction, status: OrderStatus.saved),
        ],
        isBusy: false,
      );
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
    }
  }

  Future<void> loadSalesTransactions({String? status}) async {
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final transactions = await ref
          .read(transactionRepositoryProvider)
          .listTransactions(status: status);
      state = state.copyWith(
        transactions: [
          for (final transaction in transactions)
            _salesTransactionFromServer(transaction),
        ],
        isBusy: false,
      );
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
    }
  }

  SalesOrder? consumeParkedOrderForEdit(CheckoutTransaction transaction) {
    final existing = state.savedOrders.where((order) {
      return order.id == transaction.id;
    }).firstOrNull;
    if (existing == null) return null;
    final serverOrder = _orderFromServerTransaction(
      transaction,
      fallbackItems: [
        for (final line in existing.lines)
          CheckoutItem(
            productId: line.productId,
            name: line.name,
            quantity: line.quantity,
            unitPrice: line.unitPrice,
            discount: line.discount,
          ),
      ],
      fallbackCustomer: existing.customer,
      status: OrderStatus.saved,
    );
    final order = SalesOrder(
      id: serverOrder.id,
      number: serverOrder.number,
      type: existing.type,
      status: serverOrder.status,
      lines: serverOrder.lines,
      createdAt: serverOrder.createdAt,
      customer: serverOrder.customer,
      note: serverOrder.note,
      discount: serverOrder.discount,
    );
    state = state.copyWith(
      savedOrders: [
        for (final savedOrder in state.savedOrders)
          if (savedOrder.id != transaction.id) savedOrder,
      ],
      activeOrderType: order.type,
      selectedCustomer: order.customer,
      clearSelectedCustomer: order.customer == null,
    );
    return order;
  }

  SalesOrder orderFromParkedTransaction(CheckoutTransaction transaction) {
    return _orderFromServerTransaction(transaction, status: OrderStatus.saved);
  }

  void restoreSavedOrder(SalesOrder order) {
    if (state.savedOrders.any((savedOrder) => savedOrder.id == order.id)) {
      return;
    }
    state = state.copyWith(savedOrders: [order, ...state.savedOrders]);
  }

  void upsertSavedOrderFromParkedTransaction(CheckoutTransaction transaction) {
    final order = orderFromParkedTransaction(transaction);
    state = state.copyWith(
      savedOrders: [
        order,
        for (final savedOrder in state.savedOrders)
          if (savedOrder.id != order.id) savedOrder,
      ],
    );
  }

  SalesOrder _orderFromServerTransaction(
    CheckoutTransaction transaction, {
    List<CheckoutItem> fallbackItems = const [],
    Customer? fallbackCustomer,
    OrderStatus status = OrderStatus.paid,
  }) {
    final items = transaction.items.isNotEmpty
        ? transaction.items
        : fallbackItems;
    final lines = [
      for (final item in items)
        OrderLine(
          productId: item.productId,
          name: item.name,
          quantity: item.quantity,
          unitPrice: item.unitPrice,
          discount: item.discount,
        ),
    ];
    return SalesOrder(
      id: transaction.id,
      number: transaction.number.isEmpty ? transaction.id : transaction.number,
      type: state.activeOrderType,
      status: status,
      lines: lines.isEmpty
          ? [
              OrderLine(
                productId: transaction.id,
                name: 'Transaksi Server',
                quantity: 1,
                unitPrice: transaction.grandTotal,
              ),
            ]
          : lines,
      createdAt: transaction.createdAt ?? DateTime.now(),
      customer: transaction.customerName == null && fallbackCustomer == null
          ? null
          : Customer(
              id: transaction.customerId ?? fallbackCustomer?.id ?? '',
              name: transaction.customerName ?? fallbackCustomer?.name ?? '',
              phone: fallbackCustomer?.phone ?? '',
              group: fallbackCustomer?.group ?? 'Tanpa Grup',
            ),
      note: transaction.notes ?? '',
      discount: transaction.cartDiscountTotal,
    );
  }

  SalesTransaction _salesTransactionFromServer(
    CheckoutTransaction transaction,
  ) {
    final cashier =
        transaction.cashierId == null && transaction.cashierName == null
        ? state.cashier
        : Employee(
            id: transaction.cashierId ?? state.cashier.id,
            name: transaction.cashierName ?? state.cashier.name,
            role: transaction.cashierRole ?? state.cashier.role,
          );
    return SalesTransaction(
      id: transaction.id,
      number: transaction.number.isEmpty ? transaction.id : transaction.number,
      order: _orderFromServerTransaction(
        transaction,
        status: transaction.status == 'held'
            ? OrderStatus.saved
            : transaction.status == 'voided'
            ? OrderStatus.canceled
            : OrderStatus.paid,
      ),
      payments: [
        for (final payment in transaction.payments)
          PaymentLine(
            method: payment.isCash
                ? PaymentMethod.cash
                : PaymentMethod.cashless,
            methodName: payment.method,
            amount: payment.amount,
            isCash: payment.isCash,
            reference: payment.reference,
            status: payment.status,
            paymentId: payment.id,
          ),
      ],
      cashier: cashier,
      createdAt: transaction.createdAt ?? DateTime.now(),
      status: transaction.status,
      itemDiscountTotal: transaction.itemDiscountTotal,
      cartDiscountTotal: transaction.cartDiscountTotal,
      grandTotal: transaction.grandTotal,
    );
  }

  Future<bool> voidTransaction({
    required SalesTransaction transaction,
    required String reason,
  }) async {
    final shiftId = state.activeShift?.id;
    if (shiftId == null) return false;
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      await ref
          .read(transactionRepositoryProvider)
          .voidTransaction(
            transactionId: transaction.id,
            shiftId: shiftId,
            reason: reason,
            idempotencyKey: const Uuid().v4(),
          );
      state = state.copyWith(
        isBusy: false,
        transactions: [
          for (final existing in state.transactions)
            if (existing.id == transaction.id)
              SalesTransaction(
                id: existing.id,
                number: existing.number,
                order: existing.order,
                payments: existing.payments,
                cashier: existing.cashier,
                createdAt: existing.createdAt,
                status: 'voided',
                itemDiscountTotal: existing.itemDiscountTotal,
                cartDiscountTotal: existing.cartDiscountTotal,
                grandTotal: existing.grandTotal,
              )
            else
              existing,
        ],
      );
      return true;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return false;
    }
  }

  Future<void> confirmPaymentLine(String paymentId) async {
    final confirmed = await ref
        .read(transactionRepositoryProvider)
        .confirmPayment(
          paymentId: paymentId,
          idempotencyKey: const Uuid().v4(),
        );
    state = state.copyWith(
      transactions: [
        for (final transaction in state.transactions)
          if (transaction.id == confirmed.transactionId)
            SalesTransaction(
              id: transaction.id,
              number: transaction.number,
              order: transaction.order,
              payments: [
                for (final payment in transaction.payments)
                  if (payment.paymentId == paymentId)
                    PaymentLine(
                      method: payment.method,
                      methodName: payment.methodName,
                      amount: payment.amount,
                      isCash: payment.isCash,
                      reference: payment.reference,
                      status: 'confirmed',
                      paymentId: payment.paymentId,
                    )
                  else
                    payment,
              ],
              cashier: transaction.cashier,
              createdAt: transaction.createdAt,
              status: confirmed.transactionStatus ?? transaction.status,
              itemDiscountTotal: transaction.itemDiscountTotal,
              cartDiscountTotal: transaction.cartDiscountTotal,
              grandTotal: transaction.grandTotal,
            )
          else
            transaction,
      ],
    );
  }

  Future<SalesOrder?> activateSavedOrder(String orderId) async {
    final index = state.savedOrders.indexWhere((order) => order.id == orderId);
    if (index == -1) return null;
    final deviceId = state.deviceId;
    if (deviceId == null || deviceId.isEmpty) {
      state = state.copyWith(errorMessage: 'Perangkat belum terdaftar.');
      return null;
    }

    try {
      await ref
          .read(transactionRepositoryProvider)
          .acquireParkedOrderLease(transactionId: orderId, deviceId: deviceId);
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
    } catch (error) {
      state = state.copyWith(errorMessage: _messageFor(error));
      return null;
    }
  }

  Future<void> loadInventory({String? search}) async {
    if (state.outlet.id.isEmpty) return;
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final snapshot = await ref
          .read(inventoryRepositoryProvider)
          .getInventory(outletId: state.outlet.id, search: search);
      state = state.copyWith(inventory: snapshot, isBusy: false);
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
    }
  }

  Future<InventoryPurchase?> addPurchase({
    required String supplierName,
    required String productId,
    required int quantity,
    required int unitCost,
  }) async {
    if (state.outlet.id.isEmpty) return null;
    state = state.copyWith(isBusy: true, clearError: true);
    try {
      final purchase = await ref
          .read(inventoryRepositoryProvider)
          .createPurchase(
            outletId: state.outlet.id,
            supplierName: supplierName,
            items: [
              PurchaseItemDraft(
                productId: productId,
                quantity: quantity,
                unitCost: unitCost,
              ),
            ],
          );
      final snapshot = await ref
          .read(inventoryRepositoryProvider)
          .getInventory(outletId: state.outlet.id);
      state = state.copyWith(
        purchases: [purchase, ...state.purchases],
        inventory: snapshot,
        isBusy: false,
      );
      return purchase;
    } catch (error) {
      state = state.copyWith(isBusy: false, errorMessage: _messageFor(error));
      return null;
    }
  }

  void _applyAuthSession(AuthSession session) {
    final selectedOutlet = session.outlets.where((outlet) {
      return outlet.id == session.selectedOutletId;
    }).firstOrNull;
    state = state.copyWith(
      cashier: session.user,
      account: session.user,
      businessName: session.businessName,
      outlets: session.outlets,
      outlet: selectedOutlet ?? NojposSessionState.initial().outlet,
      deviceId: session.deviceId,
      deviceUuid: session.deviceUuid,
      clearActiveShift: true,
    );
  }
}

String _messageFor(Object error) {
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}
