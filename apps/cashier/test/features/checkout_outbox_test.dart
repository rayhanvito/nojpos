import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/core/outbox/checkout_outbox.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  test(
    'ApiTransactionRepository reuses the provided idempotency key',
    () async {
      final seenKeys = <String>[];
      final repository = ApiTransactionRepository(
        apiClient: ApiClient(
          dio: Dio()
            ..httpClientAdapter = _Adapter((options) {
              seenKeys.add(options.headers['Idempotency-Key'] as String);
              return _successTransaction();
            }),
        ),
      );
      final draft = _draft(idempotencyKey: 'fixed-key');

      await repository.createTransaction(draft);
      await repository.createTransaction(draft);

      expect(seenKeys, ['fixed-key', 'fixed-key']);
    },
  );

  test('checkout network failure is stored as pending outbox item', () async {
    final store = InMemoryCheckoutOutboxStore();
    final outbox = CheckoutOutbox(
      store: store,
      repository: ApiTransactionRepository(
        apiClient: ApiClient(
          dio: Dio()
            ..httpClientAdapter = _Adapter((options) {
              throw DioException.connectionTimeout(
                timeout: const Duration(seconds: 5),
                requestOptions: options,
              );
            }),
        ),
      ),
    );

    final result = await outbox.submitOrQueue(
      _draft(idempotencyKey: 'retry-1'),
    );

    expect(result, isA<QueuedCheckoutResult>());
    final items = await store.readAll();
    expect(items.single.idempotencyKey, 'retry-1');
    expect(items.single.status, CheckoutOutboxStatus.pending);
  });

  test('one unresolved checkout blocks a different new checkout', () async {
    final store = InMemoryCheckoutOutboxStore();
    final outbox = CheckoutOutbox(
      store: store,
      repository: _ThrowingRepository(
        const ServerApiException(
          code: 'NETWORK_ERROR',
          message: 'Network unavailable.',
        ),
      ),
    );

    await outbox.submitOrQueue(_draft(idempotencyKey: 'first-key'));

    await expectLater(
      outbox.submitOrQueue(_draft(idempotencyKey: 'second-key')),
      throwsA(isA<StateError>()),
    );

    final items = await store.readAll();
    expect(items, hasLength(1));
    expect(items.single.idempotencyKey, 'first-key');
  });

  test('checkout retry buffer keeps only one persisted record', () async {
    final store = InMemoryCheckoutOutboxStore();
    await store.save([
      CheckoutOutboxItem.fromDraft(
        _draft(idempotencyKey: 'first-key'),
        now: DateTime(2026, 6, 20, 10),
      ),
      CheckoutOutboxItem.fromDraft(
        _draft(idempotencyKey: 'second-key'),
        now: DateTime(2026, 6, 20, 11),
      ),
    ]);

    final items = await store.readAll();

    expect(items, hasLength(1));
    expect(items.single.idempotencyKey, 'first-key');
  });

  test(
    'restart recovery clears the one buffer record from stored server result',
    () async {
      final store = InMemoryCheckoutOutboxStore();
      await store.save([
        CheckoutOutboxItem.fromDraft(
          _draft(idempotencyKey: 'recovery-key'),
          now: DateTime(2026, 6, 20, 10),
        ),
      ]);
      final repository = _RecoveryRepository(
        const CheckoutRecovery(
          status: CheckoutRecoveryStatus.completed,
          transaction: CheckoutTransaction(
            id: 'txn-recovered',
            number: 'TRX-RECOVERED',
            status: 'paid',
            grandTotal: 40000,
          ),
        ),
      );
      final outbox = CheckoutOutbox(store: store, repository: repository);

      final recovered = await outbox.recoverAfterRestart();

      expect(recovered?.id, 'txn-recovered');
      expect(repository.createCalls, 0);
      expect(await store.readAll(), isEmpty);
    },
  );

  test(
    'restart recovery keeps the buffer blocked when backend outcome is unknown',
    () async {
      final store = InMemoryCheckoutOutboxStore();
      await store.save([
        CheckoutOutboxItem.fromDraft(
          _draft(idempotencyKey: 'unknown-key'),
          now: DateTime(2026, 6, 20, 10),
        ),
      ]);
      final repository = _RecoveryRepository(
        const CheckoutRecovery(status: CheckoutRecoveryStatus.missing),
      );
      final outbox = CheckoutOutbox(store: store, repository: repository);

      final recovered = await outbox.recoverAfterRestart();

      expect(recovered, isNull);
      expect(repository.createCalls, 0);
      expect((await store.readAll()).single.idempotencyKey, 'unknown-key');
    },
  );

  test(
    'restart recovery keeps the buffer blocked when lookup is offline',
    () async {
      final store = InMemoryCheckoutOutboxStore();
      await store.save([
        CheckoutOutboxItem.fromDraft(
          _draft(idempotencyKey: 'offline-key'),
          now: DateTime(2026, 6, 20, 10),
        ),
      ]);
      final outbox = CheckoutOutbox(
        store: store,
        repository: _ThrowingRepository(
          const ServerApiException(
            code: 'NETWORK_ERROR',
            message: 'Network unavailable.',
          ),
        ),
      );

      final recovered = await outbox.recoverAfterRestart();

      expect(recovered, isNull);
      expect((await store.readAll()).single.idempotencyKey, 'offline-key');
    },
  );

  test('retry sends the same key and marks item sent after success', () async {
    final store = InMemoryCheckoutOutboxStore();
    final seenKeys = <String>[];
    await store.save([
      CheckoutOutboxItem.fromDraft(
        _draft(idempotencyKey: 'same-key', quote: _quote()),
        now: DateTime(2026, 6, 18, 10),
      ),
    ]);
    final outbox = CheckoutOutbox(
      store: store,
      repository: ApiTransactionRepository(
        apiClient: ApiClient(
          dio: Dio()
            ..httpClientAdapter = _Adapter((options) {
              seenKeys.add(options.headers['Idempotency-Key'] as String);
              return _successTransaction();
            }),
        ),
      ),
    );

    await outbox.retryNow((await store.readAll()).single.id);

    expect(seenKeys, ['same-key']);
    final items = await store.readAll();
    expect(items.single.status, CheckoutOutboxStatus.sent);
    expect(items.single.blocksClose, false);
    expect(items.single.draft.quote?.grandTotal, 40293);
  });

  test('pending outbox item blocks closing its shift', () async {
    final store = InMemoryCheckoutOutboxStore();
    await store.save([
      CheckoutOutboxItem.fromDraft(
        _draft(idempotencyKey: 'pending-key'),
        now: DateTime(2026, 6, 18, 10),
      ),
    ]);

    expect(await store.hasBlockingItemsForShift('shift-id'), true);
    expect(await store.hasBlockingItemsForShift('other-shift'), false);
  });

  test(
    'checkout outbox persists and restores quote breakdown payload',
    () async {
      final store = InMemoryCheckoutOutboxStore();
      final item = CheckoutOutboxItem.fromDraft(
        _draft(idempotencyKey: 'persist-key', quote: _quote()),
        now: DateTime(2026, 6, 18, 10),
        error: 'timeout',
      );

      final encoded = jsonEncode([item.toJson()]);
      final restored = CheckoutOutboxItem.fromJson(
        (jsonDecode(encoded) as List).single as Map<String, Object?>,
      );

      await store.save([restored]);
      final saved = (await store.readAll()).single;
      expect(saved.idempotencyKey, 'persist-key');
      expect(saved.status, CheckoutOutboxStatus.pending);
      expect(saved.draft.quote?.serviceChargeTotal, 3300);
      expect(saved.draft.quote?.taxTotal, 3993);
      expect(saved.draft.quote?.roundingTotal, 0);
      expect(saved.draft.quote?.grandTotal, 40293);
    },
  );

  test('retry backoff is 5s then 30s then 2m capped', () async {
    expect(checkoutOutboxBackoffFor(0), const Duration(seconds: 5));
    expect(checkoutOutboxBackoffFor(1), const Duration(seconds: 30));
    expect(checkoutOutboxBackoffFor(2), const Duration(minutes: 2));
    expect(checkoutOutboxBackoffFor(9), const Duration(minutes: 2));
  });

  test('retry moves to needs action after five failures', () async {
    var now = DateTime(2026, 6, 18, 10);
    final store = InMemoryCheckoutOutboxStore();
    await store.save([
      CheckoutOutboxItem.fromDraft(
        _draft(idempotencyKey: 'retry-five', quote: _quote()),
        now: now,
      ),
    ]);
    final outbox = CheckoutOutbox(
      store: store,
      repository: _ThrowingRepository(
        const ServerApiException(
          code: 'NETWORK_ERROR',
          message: 'Still offline.',
        ),
      ),
      clock: () => now,
    );

    for (var i = 0; i < 5; i++) {
      now = now.add(const Duration(seconds: 1));
      await outbox.retryNow('retry-five');
    }

    final item = (await store.readAll()).single;
    expect(item.retryCount, 5);
    expect(item.status, CheckoutOutboxStatus.needsAction);
    expect(item.lastError, 'Still offline.');
  });

  test('retry moves to needs action after fifteen minutes', () async {
    final createdAt = DateTime(2026, 6, 18, 10);
    final store = InMemoryCheckoutOutboxStore();
    await store.save([
      CheckoutOutboxItem.fromDraft(
        _draft(idempotencyKey: 'retry-age', quote: _quote()),
        now: createdAt,
      ),
    ]);
    final outbox = CheckoutOutbox(
      store: store,
      repository: _ThrowingRepository(
        const ServerApiException(
          code: 'NETWORK_ERROR',
          message: 'Still offline.',
        ),
      ),
      clock: () => createdAt.add(const Duration(minutes: 15)),
    );

    await outbox.retryNow('retry-age');

    final item = (await store.readAll()).single;
    expect(item.retryCount, 1);
    expect(item.status, CheckoutOutboxStatus.needsAction);
  });

  test('stale quote mismatch on retry becomes needs action', () async {
    final store = InMemoryCheckoutOutboxStore();
    await store.save([
      CheckoutOutboxItem.fromDraft(
        _draft(idempotencyKey: 'stale-key', quote: _quote()),
        now: DateTime(2026, 6, 18, 10),
      ),
    ]);
    final outbox = CheckoutOutbox(
      store: store,
      repository: _ThrowingRepository(
        const ValidationApiException(
          code: 'TOTAL_MISMATCH',
          message: 'Checkout total does not match server quote.',
        ),
      ),
    );

    await outbox.retryNow('stale-key');

    final item = (await store.readAll()).single;
    expect(item.status, CheckoutOutboxStatus.needsAction);
    expect(item.retryCount, 1);
    expect(item.lastError, 'Checkout total does not match server quote.');
    expect(item.draft.idempotencyKey, 'stale-key');
  });

  test('direct stale quote submit is not queued silently', () async {
    final store = InMemoryCheckoutOutboxStore();
    final outbox = CheckoutOutbox(
      store: store,
      repository: _ThrowingRepository(
        const ValidationApiException(
          code: 'TOTAL_MISMATCH',
          message: 'Checkout total does not match server quote.',
        ),
      ),
    );

    expect(
      () => outbox.submitOrQueue(_draft(idempotencyKey: 'direct-stale')),
      throwsA(
        isA<ValidationApiException>().having(
          (error) => error.code,
          'code',
          'TOTAL_MISMATCH',
        ),
      ),
    );
    expect(await store.readAll(), isEmpty);
  });
}

CheckoutDraft _draft({required String idempotencyKey, CheckoutQuote? quote}) {
  return CheckoutDraft(
    idempotencyKey: idempotencyKey,
    outletId: 'outlet-id',
    deviceId: 'device-id',
    cashierId: 'cashier-id',
    shiftId: 'shift-id',
    items: const [
      CheckoutItem(
        productId: 'product-id',
        name: 'Kopi',
        quantity: 2,
        unitPrice: 18000,
      ),
    ],
    payments: const [
      CheckoutPayment(method: 'Tunai', amount: 40000, isCash: true),
    ],
    paidAmount: 40000,
    quote: quote,
  );
}

CheckoutQuote _quote() {
  return const CheckoutQuote(
    subtotal: 36000,
    itemDiscountTotal: 2000,
    cartDiscountTotal: 1000,
    discountTotal: 3000,
    serviceChargeTotal: 3300,
    taxTotal: 3993,
    roundingTotal: 0,
    grandTotal: 40293,
  );
}

ResponseBody _successTransaction() {
  return ResponseBody.fromString(
    jsonEncode({
      'data': {
        'id': 'txn-id',
        'number': 'TRX-001',
        'status': 'paid',
        'grand_total': 36000,
      },
      'meta': {},
    }),
    201,
    headers: {
      Headers.contentTypeHeader: [Headers.jsonContentType],
    },
  );
}

class _Adapter implements HttpClientAdapter {
  _Adapter(this.handler);

  final ResponseBody Function(RequestOptions options) handler;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return handler(options);
  }

  @override
  void close({bool force = false}) {}
}

class _ThrowingRepository implements TransactionRepository {
  const _ThrowingRepository(this.error);

  final ApiException error;

  @override
  Future<CheckoutQuote> quoteTransaction(CheckoutDraft draft) async => _quote();

  @override
  Future<CheckoutTransaction> createTransaction(CheckoutDraft draft) async {
    throw error;
  }

  @override
  Future<CheckoutRecovery> recoverCheckout(String idempotencyKey) async {
    throw error;
  }

  @override
  Future<CheckoutPaymentLine> confirmPayment({
    required String paymentId,
    required String idempotencyKey,
  }) {
    throw UnimplementedError();
  }

  @override
  SalesTransaction createLocalTransaction({
    required SalesOrder order,
    required List<PaymentLine> payments,
    required Employee cashier,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<List<CheckoutTransaction>> listTransactions({String? status}) {
    throw UnimplementedError();
  }

  @override
  Future<VoidResult> voidTransaction({
    required String transactionId,
    required String shiftId,
    required String reason,
    required String idempotencyKey,
  }) {
    throw UnimplementedError();
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}

class _RecoveryRepository implements TransactionRepository {
  _RecoveryRepository(this.recovery);

  final CheckoutRecovery recovery;
  int createCalls = 0;

  @override
  Future<CheckoutQuote> quoteTransaction(CheckoutDraft draft) async => _quote();

  @override
  Future<CheckoutTransaction> createTransaction(CheckoutDraft draft) async {
    createCalls += 1;
    throw UnimplementedError();
  }

  @override
  Future<CheckoutRecovery> recoverCheckout(String idempotencyKey) async {
    return recovery;
  }

  @override
  Future<CheckoutPaymentLine> confirmPayment({
    required String paymentId,
    required String idempotencyKey,
  }) {
    throw UnimplementedError();
  }

  @override
  SalesTransaction createLocalTransaction({
    required SalesOrder order,
    required List<PaymentLine> payments,
    required Employee cashier,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<List<CheckoutTransaction>> listTransactions({String? status}) {
    throw UnimplementedError();
  }

  @override
  Future<VoidResult> voidTransaction({
    required String transactionId,
    required String shiftId,
    required String reason,
    required String idempotencyKey,
  }) {
    throw UnimplementedError();
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}
