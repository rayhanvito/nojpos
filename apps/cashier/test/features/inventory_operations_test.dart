import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/features/inventory/providers/inventory_operations_provider.dart';
import 'package:nojpos_tablet_ui/features/inventory/repositories/inventory_repository.dart';
import 'package:nojpos_tablet_ui/features/inventory/widgets/inventory_operation_panels.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  test('repository parses inventory overview', () async {
    final adapter = _RouteAdapter({
      'GET /inventory': {
        'data': {
          'items': [
            {
              'product_id': 'product-id',
              'product_name': 'Kopi Susu',
              'sku': 'SKU-001',
              'category_name': 'Minuman',
              'outlet_id': 'outlet-id',
              'outlet_name': 'Outlet Utama',
              'on_hand_quantity': 10,
              'available_quantity': 10,
              'in_transit_out_quantity': 2,
              'in_transit_in_quantity': 0,
            },
          ],
          'movements': [],
        },
        'meta': {},
      },
    });
    final repository = _repository(adapter);

    final snapshot = await repository.getInventory(outletId: 'outlet-id');

    expect(snapshot.items.single.name, 'Kopi Susu');
    expect(snapshot.items.single.sku, 'SKU-001');
    expect(snapshot.items.single.availableQuantity, 10);
    expect(snapshot.items.single.inTransitOutQuantity, 2);
    expect(adapter.lastQueryParameters, {'outlet_id': 'outlet-id'});
  });

  test('repository parses stock movements', () async {
    final adapter = _RouteAdapter({
      'GET /inventory/movements': {
        'data': {
          'items': [
            {
              'movement_id': 'movement-id',
              'product_id': 'product-id',
              'product_name': 'Kopi Susu',
              'outlet_id': 'outlet-id',
              'outlet_name': 'Outlet Utama',
              'type': 'adjustment',
              'quantity_delta': -3,
              'before_quantity': 10,
              'after_quantity': 7,
              'reference_type': 'inventory_count',
              'reference_id': 'count-id',
              'reason': 'Selisih opname',
              'created_at': '2026-06-20T10:00:00Z',
            },
          ],
        },
        'meta': {},
      },
    });

    final movements = await _repository(
      adapter,
    ).getMovements(outletId: 'outlet-id', type: 'adjustment');

    expect(movements.items.single.id, 'movement-id');
    expect(movements.items.single.productName, 'Kopi Susu');
    expect(movements.items.single.beforeQuantity, 10);
    expect(movements.items.single.afterQuantity, 7);
    expect(movements.items.single.referenceType, 'inventory_count');
    expect(movements.items.single.reason, 'Selisih opname');
  });

  test('repository parses count and waste responses', () async {
    final adapter = _RouteAdapter({
      'POST /inventory/counts': {
        'data': {
          'id': 'count-id',
          'number': 'CNT-001',
          'status': 'completed',
          'lines': [
            {
              'product_id': 'product-id',
              'system_quantity': 10,
              'counted_quantity': 7,
              'delta_quantity': -3,
              'reason': 'Selisih opname',
            },
          ],
        },
        'meta': {},
      },
      'POST /inventory/waste': {
        'data': {
          'id': 'waste-id',
          'product_id': 'product-id',
          'quantity': 2,
          'reason': 'Kadaluarsa',
          'status': 'completed',
        },
        'meta': {},
      },
    });
    final repository = _repository(adapter);

    final count = await repository.createCount(
      outletId: 'outlet-id',
      idempotencyKey: 'count-key',
      lines: const [
        InventoryCountLineDraft(
          productId: 'product-id',
          countedQuantity: 7,
          reason: 'Selisih opname',
        ),
      ],
    );
    final waste = await repository.createWaste(
      outletId: 'outlet-id',
      productId: 'product-id',
      quantity: 2,
      reason: 'Kadaluarsa',
      idempotencyKey: 'waste-key',
    );

    expect(count.lines.single.deltaQuantity, -3);
    expect(waste.reason, 'Kadaluarsa');
    expect(adapter.idempotencyKeys, ['count-key', 'waste-key']);
  });

  test('repository parses transfer lifecycle and in-transit list', () async {
    final transferPayload = {
      'id': 'transfer-id',
      'number': 'TRF-001',
      'source_outlet': {'id': 'source-id', 'name': 'Source'},
      'destination_outlet': {'id': 'destination-id', 'name': 'Destination'},
      'status': 'in_transit',
      'sent_at': '2026-06-20T10:00:00Z',
      'age_seconds': 30,
      'lines': [
        {
          'product_id': 'product-id',
          'product_name': 'Kopi Susu',
          'requested_quantity': 4,
          'sent_quantity': 4,
          'received_quantity': 0,
          'in_transit_quantity': 4,
        },
      ],
    };
    final adapter = _RouteAdapter({
      'POST /inventory/transfers': {'data': transferPayload, 'meta': {}},
      'POST /inventory/transfers/transfer-id/send': {
        'data': transferPayload,
        'meta': {},
      },
      'POST /inventory/transfers/transfer-id/receive': {
        'data': {...transferPayload, 'status': 'received'},
        'meta': {},
      },
      'POST /inventory/transfers/transfer-id/cancel': {
        'data': {...transferPayload, 'status': 'cancelled'},
        'meta': {},
      },
      'GET /inventory/transfers/in-transit': {
        'data': {
          'items': [transferPayload],
        },
        'meta': {},
      },
    });
    final repository = _repository(adapter);

    final created = await repository.createTransfer(
      sourceOutletId: 'source-id',
      destinationOutletId: 'destination-id',
      idempotencyKey: 'create-key',
      lines: const [
        InventoryTransferLineDraft(productId: 'product-id', quantity: 4),
      ],
    );
    final sent = await repository.sendTransfer(
      transferId: 'transfer-id',
      idempotencyKey: 'send-key',
    );
    final received = await repository.receiveTransfer(
      transferId: 'transfer-id',
      idempotencyKey: 'receive-key',
    );
    final cancelled = await repository.cancelTransfer(
      transferId: 'transfer-id',
      idempotencyKey: 'cancel-key',
    );
    final inTransit = await repository.listInTransitTransfers(
      outletId: 'source-id',
    );

    expect(created.lines.single.inTransitQuantity, 4);
    expect(sent.isInTransit, true);
    expect(received.status, 'received');
    expect(cancelled.status, 'cancelled');
    expect(inTransit.items.single.sourceOutletName, 'Source');
    expect(adapter.idempotencyKeys, [
      'create-key',
      'send-key',
      'receive-key',
      'cancel-key',
    ]);
  });

  test('provider handles forbidden and conflict responses', () async {
    final repository = _FakeInventoryRepository()
      ..wasteError = const ForbiddenApiException(
        code: 'FORBIDDEN',
        message: 'Forbidden.',
      )
      ..countError = const ConflictApiException(
        code: 'STOCK_INSUFFICIENT',
        message: 'Insufficient stock.',
      );
    final container = ProviderContainer(
      overrides: [
        inventoryRepositoryProvider.overrideWith((ref) => repository),
      ],
    );
    addTearDown(container.dispose);

    final waste = await container
        .read(inventoryOperationsProvider.notifier)
        .createWaste(
          outletId: 'outlet-id',
          productId: 'product-id',
          quantity: 1,
          reason: 'Rusak',
        );
    expect(waste, isNull);
    expect(
      container.read(inventoryOperationsProvider).errorMessage,
      'Akses inventory ditolak.',
    );

    final count = await container
        .read(inventoryOperationsProvider.notifier)
        .createCount(
          outletId: 'outlet-id',
          productId: 'product-id',
          countedQuantity: 0,
          reason: 'Selisih',
        );
    expect(count, isNull);
    expect(
      container.read(inventoryOperationsProvider).errorMessage,
      'Stok tidak mencukupi.',
    );
  });

  testWidgets('inventory empty states and in-transit state render', (
    tester,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        child: MaterialApp(
          home: Scaffold(
            body: Column(
              children: [
                const InventoryMovementPanel(movements: []),
                InTransitTransfersPanel(
                  canManage: true,
                  onChanged: () async {},
                  transfers: [
                    InventoryTransferResult(
                      id: 'transfer-id',
                      number: 'TRF-001',
                      status: 'in_transit',
                      sourceOutletName: 'Source',
                      destinationOutletName: 'Destination',
                      lines: const [
                        InventoryTransferLineResult(
                          productId: 'product-id',
                          productName: 'Kopi Susu',
                          requestedQuantity: 4,
                          sentQuantity: 4,
                          receivedQuantity: 0,
                          inTransitQuantity: 4,
                        ),
                      ],
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );

    expect(find.text('Belum ada pergerakan stok.'), findsOneWidget);
    expect(find.textContaining('TRF-001'), findsOneWidget);
    expect(find.textContaining('transit 4 dari 4'), findsOneWidget);
  });

  testWidgets('cashier UI does not expose privileged inventory actions', (
    tester,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        child: MaterialApp(
          home: Scaffold(
            body: InventoryActionPanel(
              canManage: false,
              outletId: 'outlet-id',
              outlets: const [
                Outlet(id: 'outlet-id', name: 'Outlet', isOnline: true),
              ],
              items: const [
                InventoryStockItem(
                  productId: 'product-id',
                  name: 'Kopi Susu',
                  price: 18000,
                  stockOnHand: 10,
                  availableQuantity: 10,
                ),
              ],
              onChanged: () async {},
            ),
          ),
        ),
      ),
    );

    expect(
      find.text('Mode cashier read-only. Aksi stok hanya untuk owner/admin.'),
      findsOneWidget,
    );
    expect(find.text('Simpan Opname'), findsNothing);
    expect(find.text('Catat Waste'), findsNothing);
    expect(find.text('Buat Transfer'), findsNothing);
  });
}

ApiInventoryRepository _repository(_RouteAdapter adapter) {
  return ApiInventoryRepository(
    apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
  );
}

class _FakeInventoryRepository implements InventoryRepository {
  Object? countError;
  Object? wasteError;

  @override
  Future<InventoryCountResult> createCount({
    required String outletId,
    String? notes,
    DateTime? countedAt,
    required List<InventoryCountLineDraft> lines,
    required String idempotencyKey,
  }) async {
    final error = countError;
    if (error != null) throw error;
    return const InventoryCountResult(
      id: 'count-id',
      number: 'CNT-001',
      status: 'completed',
      lines: [],
    );
  }

  @override
  Future<InventoryWasteResult> createWaste({
    required String outletId,
    required String productId,
    required int quantity,
    required String reason,
    DateTime? occurredAt,
    required String idempotencyKey,
  }) async {
    final error = wasteError;
    if (error != null) throw error;
    return InventoryWasteResult(
      id: 'waste-id',
      productId: productId,
      quantity: quantity,
      reason: reason,
      status: 'completed',
    );
  }

  @override
  Future<InventoryPurchase> createPurchase({
    required String outletId,
    String? number,
    required String supplierName,
    String? notes,
    required List<PurchaseItemDraft> items,
  }) async {
    throw UnimplementedError();
  }

  @override
  Future<InventoryTransferResult> createTransfer({
    required String sourceOutletId,
    required String destinationOutletId,
    String? notes,
    required List<InventoryTransferLineDraft> lines,
    required String idempotencyKey,
  }) async {
    throw UnimplementedError();
  }

  @override
  Future<InventorySnapshot> getInventory({
    String? outletId,
    String? productId,
    String? type,
    String? search,
  }) async {
    throw UnimplementedError();
  }

  @override
  Future<InventoryMovementHistory> getMovements({
    String? outletId,
    String? productId,
    String? type,
  }) async {
    return const InventoryMovementHistory(items: []);
  }

  @override
  Future<InventoryTransferList> listInTransitTransfers({
    String? outletId,
  }) async {
    return const InventoryTransferList(items: []);
  }

  @override
  Future<InventoryTransferList> listTransfers({
    String? outletId,
    String? status,
  }) async {
    return const InventoryTransferList(items: []);
  }

  @override
  Future<InventoryTransferResult> cancelTransfer({
    required String transferId,
    required String idempotencyKey,
  }) async {
    throw UnimplementedError();
  }

  @override
  Future<InventoryTransferResult> receiveTransfer({
    required String transferId,
    required String idempotencyKey,
  }) async {
    throw UnimplementedError();
  }

  @override
  Future<InventoryTransferResult> sendTransfer({
    required String transferId,
    required String idempotencyKey,
  }) async {
    throw UnimplementedError();
  }
}

class _RouteAdapter implements HttpClientAdapter {
  _RouteAdapter(this.responses);

  final Map<String, Map<String, Object?>> responses;
  Map<String, Object?>? lastQueryParameters;
  final List<String> idempotencyKeys = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final key = '${options.method} ${options.path}';
    lastQueryParameters = Map<String, Object?>.from(options.queryParameters);
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
    final status = (response['status'] as num?)?.toInt() ?? 200;
    if (status >= 400) {
      return ResponseBody.fromString(
        jsonEncode({'error': response['error']}),
        status,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }
    return ResponseBody.fromString(
      jsonEncode(response),
      status,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
