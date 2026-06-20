import 'dart:async';
import 'dart:convert';

import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../network/api_client.dart';
import '../../features/transactions/repositories/transaction_repository.dart';

typedef CheckoutOutboxClock = DateTime Function();
typedef CheckoutOutboxSleeper = Future<void> Function(Duration duration);

final checkoutOutboxStoreProvider = Provider<CheckoutOutboxStore>(
  (ref) => const SecureCheckoutOutboxStore(),
);

final checkoutOutboxProvider = Provider<CheckoutOutbox>((ref) {
  return CheckoutOutbox(
    store: ref.watch(checkoutOutboxStoreProvider),
    repository: ref.watch(transactionRepositoryProvider),
    clock: ref.watch(checkoutOutboxClockProvider),
  );
});

final checkoutOutboxClockProvider = Provider<CheckoutOutboxClock>(
  (ref) => DateTime.now,
);

final checkoutOutboxSleeperProvider = Provider<CheckoutOutboxSleeper>(
  (ref) =>
      (duration) => Future<void>.delayed(duration),
);

final checkoutOutboxControllerProvider =
    NotifierProvider<CheckoutOutboxController, List<CheckoutOutboxItem>>(
      CheckoutOutboxController.new,
    );

enum CheckoutOutboxStatus { pending, sending, sent, needsAction }

sealed class CheckoutSubmitResult {
  const CheckoutSubmitResult();
}

class SentCheckoutResult extends CheckoutSubmitResult {
  const SentCheckoutResult(this.transaction);

  final CheckoutTransaction transaction;
}

class QueuedCheckoutResult extends CheckoutSubmitResult {
  const QueuedCheckoutResult(this.item);

  final CheckoutOutboxItem item;
}

class CheckoutOutboxItem {
  const CheckoutOutboxItem({
    required this.id,
    required this.draft,
    required this.createdAt,
    required this.retryCount,
    required this.status,
    this.lastError,
  });

  factory CheckoutOutboxItem.fromDraft(
    CheckoutDraft draft, {
    required DateTime now,
    String? error,
  }) {
    return CheckoutOutboxItem(
      id: draft.idempotencyKey,
      draft: draft,
      createdAt: now,
      retryCount: 0,
      status: CheckoutOutboxStatus.pending,
      lastError: error,
    );
  }

  factory CheckoutOutboxItem.fromJson(Map<String, Object?> json) {
    return CheckoutOutboxItem(
      id: json['id'] as String,
      draft: CheckoutDraft.fromJson(_asMap(json['draft'])),
      createdAt: DateTime.parse(json['created_at'] as String),
      retryCount: (json['retry_count'] as num).toInt(),
      status: CheckoutOutboxStatus.values.byName(json['status'] as String),
      lastError: json['last_error'] as String?,
    );
  }

  final String id;
  final CheckoutDraft draft;
  final DateTime createdAt;
  final int retryCount;
  final CheckoutOutboxStatus status;
  final String? lastError;

  String get idempotencyKey => draft.idempotencyKey;

  bool get blocksClose =>
      status == CheckoutOutboxStatus.pending ||
      status == CheckoutOutboxStatus.sending ||
      status == CheckoutOutboxStatus.needsAction;

  CheckoutOutboxItem copyWith({
    int? retryCount,
    CheckoutOutboxStatus? status,
    Object? lastError = _keepLastError,
  }) {
    return CheckoutOutboxItem(
      id: id,
      draft: draft,
      createdAt: createdAt,
      retryCount: retryCount ?? this.retryCount,
      status: status ?? this.status,
      lastError: identical(lastError, _keepLastError)
          ? this.lastError
          : lastError as String?,
    );
  }

  Map<String, Object?> toJson() {
    return {
      'id': id,
      'draft': draft.toJson(),
      'created_at': createdAt.toIso8601String(),
      'retry_count': retryCount,
      'status': status.name,
      if (lastError != null) 'last_error': lastError,
    };
  }
}

const Object _keepLastError = Object();

abstract interface class CheckoutOutboxStore {
  Future<List<CheckoutOutboxItem>> readAll();

  Future<void> save(List<CheckoutOutboxItem> items);

  Future<bool> hasBlockingItemsForShift(String shiftId);

  Future<int> blockingItemCountForShift(String shiftId);
}

class SecureCheckoutOutboxStore implements CheckoutOutboxStore {
  const SecureCheckoutOutboxStore({FlutterSecureStorage? secureStorage})
    : _secureStorage = secureStorage ?? const FlutterSecureStorage();

  static const _key = 'nojpos.checkout_outbox';

  final FlutterSecureStorage _secureStorage;

  @override
  Future<List<CheckoutOutboxItem>> readAll() async {
    final raw = await _secureStorage.read(key: _key);
    if (raw == null || raw.isEmpty) return const [];
    final decoded = jsonDecode(raw) as List;
    final items = [
      for (final item in decoded) CheckoutOutboxItem.fromJson(_asMap(item)),
    ];
    if (items.length > 1) {
      final singleItem = [items.first];
      await save(singleItem);
      return singleItem;
    }
    return items;
  }

  @override
  Future<void> save(List<CheckoutOutboxItem> items) async {
    final singleItem = items.take(1);
    await _secureStorage.write(
      key: _key,
      value: jsonEncode([for (final item in singleItem) item.toJson()]),
    );
  }

  @override
  Future<bool> hasBlockingItemsForShift(String shiftId) async {
    return (await blockingItemCountForShift(shiftId)) > 0;
  }

  @override
  Future<int> blockingItemCountForShift(String shiftId) async {
    final items = await readAll();
    return items
        .where((item) => item.draft.shiftId == shiftId && item.blocksClose)
        .length;
  }
}

class InMemoryCheckoutOutboxStore implements CheckoutOutboxStore {
  List<CheckoutOutboxItem> _items = [];

  @override
  Future<List<CheckoutOutboxItem>> readAll() async => [..._items];

  @override
  Future<void> save(List<CheckoutOutboxItem> items) async {
    _items = items.take(1).toList();
  }

  @override
  Future<bool> hasBlockingItemsForShift(String shiftId) async {
    return (await blockingItemCountForShift(shiftId)) > 0;
  }

  @override
  Future<int> blockingItemCountForShift(String shiftId) async {
    return _items
        .where((item) => item.draft.shiftId == shiftId && item.blocksClose)
        .length;
  }
}

class CheckoutOutbox {
  CheckoutOutbox({
    required CheckoutOutboxStore store,
    required TransactionRepository repository,
    CheckoutOutboxClock? clock,
  }) : _store = store,
       _repository = repository,
       _clock = clock ?? DateTime.now;

  final CheckoutOutboxStore _store;
  final TransactionRepository _repository;
  final CheckoutOutboxClock _clock;

  Future<CheckoutSubmitResult> submitOrQueue(CheckoutDraft draft) async {
    final existing = (await _store.readAll())
        .where((item) => item.blocksClose)
        .firstOrNull;
    if (existing != null && existing.idempotencyKey != draft.idempotencyKey) {
      throw StateError(
        'Selesaikan checkout sebelumnya sebelum memulai checkout baru.',
      );
    }

    try {
      final transaction = await _repository.createTransaction(draft);
      return SentCheckoutResult(transaction);
    } on ApiException catch (error) {
      if (error.code != 'NETWORK_ERROR') rethrow;
      final item = CheckoutOutboxItem.fromDraft(
        draft,
        now: _clock(),
        error: error.message,
      );
      await _upsert(item);
      return QueuedCheckoutResult(item);
    }
  }

  Future<CheckoutTransaction?> retryNow(String itemId) async {
    final items = await _store.readAll();
    final index = items.indexWhere((item) => item.id == itemId);
    if (index == -1) return null;

    var item = items[index].copyWith(status: CheckoutOutboxStatus.sending);
    items[index] = item;
    await _store.save(items);

    try {
      final transaction = await _repository.createTransaction(item.draft);
      final latest = await _store.readAll();
      final sentIndex = latest.indexWhere((existing) => existing.id == itemId);
      if (sentIndex != -1) {
        latest[sentIndex] = item.copyWith(status: CheckoutOutboxStatus.sent);
        await _store.save(latest);
      }
      return transaction;
    } on ApiException catch (error) {
      final retryCount = item.retryCount + 1;
      final needsAction =
          _requiresManualAction(error) ||
          retryCount >= 5 ||
          _clock().difference(item.createdAt) >= const Duration(minutes: 15);
      item = item.copyWith(
        retryCount: retryCount,
        status: needsAction
            ? CheckoutOutboxStatus.needsAction
            : CheckoutOutboxStatus.pending,
        lastError: error.message,
      );
      final latest = await _store.readAll();
      final retryIndex = latest.indexWhere((existing) => existing.id == itemId);
      if (retryIndex == -1) return null;
      latest[retryIndex] = item;
      await _store.save(latest);
      return null;
    }
  }

  Future<CheckoutTransaction?> recoverAfterRestart() async {
    final item = (await _store.readAll())
        .where((item) => item.blocksClose)
        .firstOrNull;
    if (item == null) return null;

    late final CheckoutRecovery recovery;
    try {
      recovery = await _repository.recoverCheckout(item.idempotencyKey);
    } on ApiException catch (error) {
      if (error.code == 'NETWORK_ERROR') return null;
      rethrow;
    }
    if (recovery.status != CheckoutRecoveryStatus.completed ||
        recovery.transaction == null) {
      return null;
    }

    await _store.save(const []);
    return recovery.transaction;
  }

  Future<void> _upsert(CheckoutOutboxItem item) async {
    final items = await _store.readAll();
    final index = items.indexWhere((existing) => existing.id == item.id);
    if (index == -1) {
      await _store.save([item]);
      return;
    }
    items[index] = item;
    await _store.save(items);
  }
}

bool _requiresManualAction(ApiException error) {
  return switch (error.code) {
    'TOTAL_MISMATCH' ||
    'IDEMPOTENCY_CONFLICT' ||
    'VALIDATION_ERROR' ||
    'PRODUCT_NOT_FOUND' ||
    'SHIFT_NOT_OPEN' => true,
    _ => false,
  };
}

class CheckoutOutboxController extends Notifier<List<CheckoutOutboxItem>>
    with WidgetsBindingObserver {
  StreamSubscription<List<ConnectivityResult>>? _connectivitySubscription;
  bool _retryInFlight = false;

  @override
  List<CheckoutOutboxItem> build() {
    WidgetsBinding.instance.addObserver(this);
    ref.onDispose(() {
      WidgetsBinding.instance.removeObserver(this);
      _connectivitySubscription?.cancel();
    });
    _connectivitySubscription = Connectivity().onConnectivityChanged.listen((
      results,
    ) {
      if (!results.contains(ConnectivityResult.none)) {
        retryPending();
      }
    });
    Future<void>.microtask(load);
    return const [];
  }

  Future<void> load() async {
    await ref.read(checkoutOutboxProvider).recoverAfterRestart();
    state = await ref.read(checkoutOutboxStoreProvider).readAll();
  }

  Future<CheckoutSubmitResult> submitOrQueue(CheckoutDraft draft) async {
    final result = await ref.read(checkoutOutboxProvider).submitOrQueue(draft);
    await load();
    return result;
  }

  Future<CheckoutTransaction?> retryNow(String itemId) async {
    final transaction = await ref.read(checkoutOutboxProvider).retryNow(itemId);
    await load();
    return transaction;
  }

  Future<void> retryPending() async {
    if (_retryInFlight) return;
    _retryInFlight = true;
    try {
      await ref.read(checkoutOutboxProvider).recoverAfterRestart();
      await load();
    } finally {
      _retryInFlight = false;
    }
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      load();
    }
  }
}

Duration checkoutOutboxBackoffFor(int retryCount) {
  if (retryCount <= 0) return const Duration(seconds: 5);
  if (retryCount == 1) return const Duration(seconds: 30);
  return const Duration(minutes: 2);
}

Map<String, Object?> _asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) return Map<String, Object?>.from(value);
  return const {};
}
