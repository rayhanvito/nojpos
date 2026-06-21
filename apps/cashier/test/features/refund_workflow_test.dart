import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/app/providers/nojpos_session_provider.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/features/pos/screens/operations_screen.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  test(
    'repository sends refund request with idempotency key and parses response',
    () async {
      final adapter = _RouteAdapter({
        'POST /transactions/transaction-id/refund': {
          'data': _refundResponse(status: 'partially_refunded'),
          'meta': {},
        },
      });
      final repository = ApiTransactionRepository(
        apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
      );

      final result = await repository.createRefund(
        transactionId: 'transaction-id',
        request: const RefundRequest(
          idempotencyKey: 'refund-key',
          reason: 'Retur pelanggan',
          refundMethod: 'cash',
          authorizationPin: '1234',
          lines: [RefundLineRequest(transactionItemId: 'item-id', quantity: 1)],
        ),
      );

      expect(adapter.idempotencyKeys, ['refund-key']);
      expect(adapter.lastRequestData, {
        'reason': 'Retur pelanggan',
        'refund_method': 'cash',
        'authorization_pin': '1234',
        'lines': [
          {'transaction_item_id': 'item-id', 'quantity': 1, 'restock': true},
        ],
      });
      expect(result.transactionStatus, 'partially_refunded');
      expect(result.totalRefundAmount, 10000);
      expect(result.refundedLines.single.transactionItemId, 'item-id');
      expect(result.stockMovements.single.type, 'refund_reversal');
      expect(result.cashMovement?.type, 'cash_out');
    },
  );

  test(
    'session updates transaction status from server refund response',
    () async {
      final notifier = _RefundSessionNotifier(role: 'owner');
      final container = ProviderContainer(
        overrides: [nojposSessionProvider.overrideWith(() => notifier)],
      );
      addTearDown(container.dispose);

      final result = await container
          .read(nojposSessionProvider.notifier)
          .createRefund(
            transaction: _transaction(),
            reason: 'Retur pelanggan',
            refundMethod: 'cash',
            authorizationCode: '1234',
            lines: const [
              RefundLineRequest(transactionItemId: 'item-id', quantity: 1),
            ],
          );

      expect(result?.transactionStatus, 'refunded');
      expect(
        container.read(nojposSessionProvider).transactions.single.status,
        'refunded',
      );
    },
  );

  testWidgets('cashier UI hides enabled refund action', (tester) async {
    await _pumpSales(tester, role: 'cashier');
    await tester.tap(find.textContaining('TRX-REFUND'));
    await tester.pumpAndSettle();
    final cashierButton = tester.widget<FilledButton>(
      find.byKey(const ValueKey('transaction_refund_action')),
    );
    expect(cashierButton.onPressed, isNull);
    expect(find.text('Refund owner/admin'), findsOneWidget);
  });

  testWidgets('owner UI shows enabled refund action', (tester) async {
    await _pumpSales(tester, role: 'owner');
    await tester.tap(find.textContaining('TRX-REFUND'));
    await tester.pumpAndSettle();
    expect(find.text('Refund'), findsOneWidget);
    final ownerButton = tester.widget<FilledButton>(
      find.byKey(const ValueKey('transaction_refund_action')),
    );
    expect(ownerButton.onPressed, isNotNull);
  });

  testWidgets('refund dialog shows server validation or conflict error', (
    tester,
  ) async {
    final notifier = _RefundSessionNotifier(
      role: 'admin',
      failMessage: 'Otorisasi wajib atau refund melebihi sisa transaksi.',
    );
    await _pumpSales(tester, role: 'admin', notifier: notifier);
    await tester.tap(find.textContaining('TRX-REFUND'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('transaction_refund_action')));
    await tester.pumpAndSettle();
    await tester.enterText(
      find.byKey(const ValueKey('refund_reason')),
      'Retur pelanggan',
    );
    await tester.tap(find.byKey(const ValueKey('refund_submit')));
    await tester.pumpAndSettle();

    expect(
      find.text('Otorisasi wajib atau refund melebihi sisa transaksi.'),
      findsAtLeastNWidgets(1),
    );
  });

  testWidgets(
    'refund dialog validates non-restock reason and submits server action',
    (tester) async {
      final notifier = _RefundSessionNotifier(role: 'owner');
      await _pumpSales(tester, role: 'owner', notifier: notifier);
      await tester.tap(find.textContaining('TRX-REFUND'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('transaction_refund_action')));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.byKey(const ValueKey('refund_reason')),
        'Retur pelanggan',
      );
      await tester.enterText(
        find.byKey(const ValueKey('refund_authorization_pin')),
        '1234',
      );
      await tester.tap(find.byKey(const ValueKey('refund_restock_toggle')));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('refund_submit')));
      await tester.pumpAndSettle();

      expect(find.text('Alasan tidak restock wajib diisi.'), findsOneWidget);

      await tester.enterText(
        find.byKey(const ValueKey('refund_non_restock_reason')),
        'Barang rusak',
      );
      await tester.tap(find.byKey(const ValueKey('refund_submit')));
      await tester.pumpAndSettle();

      expect(notifier.lastRefundMethod, 'cash');
      expect(notifier.lastRefundLines.single.restock, isFalse);
      expect(notifier.lastRefundLines.single.nonRestockReason, 'Barang rusak');
    },
  );
}

Future<void> _pumpSales(
  WidgetTester tester, {
  required String role,
  _RefundSessionNotifier? notifier,
}) async {
  tester.view.physicalSize = const Size(1280, 800);
  tester.view.devicePixelRatio = 1;
  addTearDown(() {
    tester.view.resetPhysicalSize();
    tester.view.resetDevicePixelRatio();
  });
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        nojposSessionProvider.overrideWith(
          () => notifier ?? _RefundSessionNotifier(role: role),
        ),
      ],
      child: const MaterialApp(
        home: OperationsScreen(type: OperationsPageType.sales),
      ),
    ),
  );
  await tester.pumpAndSettle();
}

Map<String, Object?> _refundResponse({String status = 'refunded'}) {
  return {
    'id': 'refund-id',
    'transaction_id': 'transaction-id',
    'transaction_status': status,
    'status': 'finalized',
    'refund_method': 'cash',
    'total_refund_amount': 10000,
    'remaining_refundable_amount': status == 'refunded' ? 0 : 10000,
    'reason': 'Retur pelanggan',
    'refunded_lines': [
      {
        'id': 'refund-line-id',
        'transaction_item_id': 'item-id',
        'product_id': 'product-id',
        'quantity': 1,
        'amount': 10000,
        'restock': true,
      },
    ],
    'stock_movements': [
      {
        'movement_id': 'stock-movement-id',
        'product_id': 'product-id',
        'quantity_delta': 1,
        'type': 'refund_reversal',
      },
    ],
    'cash_movement': {
      'movement_id': 'cash-movement-id',
      'type': 'cash_out',
      'amount': 10000,
    },
    'server_time': '2026-06-20T10:00:00Z',
  };
}

SalesTransaction _transaction() {
  return SalesTransaction(
    id: 'transaction-id',
    number: 'TRX-REFUND',
    order: SalesOrder(
      id: 'transaction-id',
      number: 'TRX-REFUND',
      type: OrderType.dineIn,
      status: OrderStatus.paid,
      lines: const [
        OrderLine(
          transactionItemId: 'item-id',
          productId: 'product-id',
          name: 'Kopi Susu',
          quantity: 1,
          unitPrice: 10000,
        ),
      ],
      createdAt: DateTime(2026, 6, 20, 10),
    ),
    payments: [PaymentLine(method: PaymentMethod.cash, amount: 10000)],
    cashier: const Employee(id: 'cashier-id', name: 'Kasir', role: 'cashier'),
    createdAt: DateTime(2026, 6, 20, 10),
    status: 'paid',
    grandTotal: 10000,
  );
}

class _RefundSessionNotifier extends NojposSessionNotifier {
  _RefundSessionNotifier({required this.role, this.failMessage});

  final String role;
  final String? failMessage;
  String? lastRefundMethod;
  List<RefundLineRequest> lastRefundLines = const [];

  @override
  NojposSessionState build() {
    final user = Employee(id: '$role-id', name: role, role: role);
    return NojposSessionState.initial().copyWith(
      account: user,
      cashier: user,
      status: SessionStatus.ready,
      transactions: [_transaction()],
      activeShift: ShiftSession(
        id: 'shift-id',
        cashier: user,
        openedAt: DateTime(2026, 6, 20, 9),
      ),
    );
  }

  @override
  Future<void> loadSalesTransactions({String? status}) async {}

  @override
  Future<RefundResult?> createRefund({
    required SalesTransaction transaction,
    required String reason,
    required String refundMethod,
    required List<RefundLineRequest> lines,
    String? authorizationCode,
    String? notes,
  }) async {
    final message = failMessage;
    if (message != null) {
      state = state.copyWith(errorMessage: message);
      return null;
    }
    lastRefundMethod = refundMethod;
    lastRefundLines = lines;
    state = state.copyWith(
      transactions: [
        SalesTransaction(
          id: transaction.id,
          number: transaction.number,
          order: transaction.order,
          payments: transaction.payments,
          cashier: transaction.cashier,
          createdAt: transaction.createdAt,
          status: 'refunded',
          grandTotal: transaction.grandTotal,
        ),
      ],
    );
    return const RefundResult(
      id: 'refund-id',
      transactionId: 'transaction-id',
      transactionStatus: 'refunded',
      status: 'finalized',
      refundMethod: 'cash',
      totalRefundAmount: 10000,
      remainingRefundableAmount: 0,
    );
  }
}

class _RouteAdapter implements HttpClientAdapter {
  _RouteAdapter(this.responses);

  final Map<String, Map<String, Object?>> responses;
  Object? lastRequestData;
  final List<String> idempotencyKeys = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final key = '${options.method} ${options.path}';
    lastRequestData = options.data;
    final idempotencyKey = options.headers['Idempotency-Key'];
    if (idempotencyKey is String) idempotencyKeys.add(idempotencyKey);
    final response = responses[key];
    if (response == null) {
      return ResponseBody.fromString(
        jsonEncode({
          'error': {'code': 'NOT_FOUND', 'message': key, 'details': {}},
        }),
        404,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }
    return ResponseBody.fromString(
      jsonEncode(response),
      (response['status'] as num?)?.toInt() ?? 200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
