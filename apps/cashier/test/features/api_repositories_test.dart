import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/core/storage/token_storage.dart';
import 'package:nojpos_tablet_ui/features/attendance/repositories/attendance_repository.dart';
import 'package:nojpos_tablet_ui/features/auth/repositories/auth_repository.dart';
import 'package:nojpos_tablet_ui/features/customers/repositories/customer_repository.dart';
import 'package:nojpos_tablet_ui/features/inventory/repositories/inventory_repository.dart';
import 'package:nojpos_tablet_ui/features/payment/pages/payment_screen.dart';
import 'package:nojpos_tablet_ui/features/pos/repositories/product_repository.dart';
import 'package:nojpos_tablet_ui/features/reports/repositories/report_repository.dart';
import 'package:nojpos_tablet_ui/features/shift/repositories/shift_repository.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  test('ApiAuthRepository stores token and maps login session', () async {
    final storage = InMemoryTokenStorage();
    final client = ApiClient(
      dio: Dio()
        ..httpClientAdapter = _RouteAdapter({
          'POST /auth/login': {
            'data': {
              'token': 'plain-token',
              'user': {
                'id': 'owner-id',
                'name': 'Owner Demo',
                'email': 'owner@demo.nojpos.test',
                'role': 'owner',
              },
              'business': {'id': 'business-id', 'name': 'NojPOS Demo'},
              'device': {'id': 'device-id', 'device_uuid': 'demo-device'},
              'outlets': [
                {'id': 'outlet-id', 'name': 'Outlet Demo'},
              ],
            },
            'meta': {},
          },
        }),
      tokenReader: storage.readToken,
    );

    final session =
        await ApiAuthRepository(apiClient: client, tokenStorage: storage).login(
          email: 'owner@demo.nojpos.test',
          password: 'password',
          deviceUuid: 'demo-device',
        );

    expect(await storage.readToken(), 'plain-token');
    expect(session.outlets.single.name, 'Outlet Demo');
    expect(session.deviceId, 'device-id');
  });

  test(
    'ApiProductRepository maps products and categories from catalog',
    () async {
      final repository = ApiProductRepository(
        apiClient: ApiClient(
          dio: Dio()
            ..httpClientAdapter = _RouteAdapter({
              'GET /categories': {
                'data': [
                  {'id': 'category-id', 'name': 'Minuman'},
                ],
                'meta': {},
              },
              'GET /products': {
                'data': {
                  'products': [
                    {
                      'id': 'product-id',
                      'product_category_id': 'category-id',
                      'name': 'Kopi Susu Demo',
                      'price': 18000,
                      'barcode': 'DEMO-KOPI',
                      'track_stock': true,
                    },
                  ],
                },
                'meta': {},
              },
            }),
        ),
      );

      final catalog = await repository.getCatalog();

      expect(catalog.categories, ['Semua', 'Favorit', 'Minuman']);
      expect(catalog.products.single.name, 'Kopi Susu Demo');
      expect(catalog.products.single.category, 'Minuman');
      expect(catalog.products.single.price, 18000);
    },
  );

  test('ApiInventoryRepository lists stock and creates purchases', () async {
    final adapter = _RouteAdapter({
      'GET /inventory': {
        'data': {
          'items': [
            {
              'product_id': 'product-id',
              'name': 'Kopi Susu',
              'price': 18000,
              'stock_on_hand': 12,
              'warning': null,
            },
            {
              'product_id': 'negative-product-id',
              'name': 'Roti Bakar',
              'price': 25000,
              'stock_on_hand': -3,
              'warning': 'NEGATIVE_STOCK_ALLOWED',
            },
          ],
          'movements': [
            {
              'id': 'movement-id',
              'product_id': 'product-id',
              'type': 'purchase',
              'quantity_delta': 12,
            },
          ],
        },
        'meta': {},
      },
      'POST /inventory/purchases': {
        'data': {
          'id': 'purchase-id',
          'number': 'INV-001',
          'supplier_name': 'Supplier Kopi',
          'total': 120000,
          'purchased_at': '2026-06-18T10:00:00Z',
          'items': [
            {
              'product_id': 'product-id',
              'quantity': 12,
              'unit_cost': 10000,
              'subtotal': 120000,
            },
          ],
        },
        'meta': {},
      },
    });
    final repository = ApiInventoryRepository(
      apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
    );

    final inventory = await repository.getInventory(
      outletId: 'outlet-id',
      search: 'Kopi',
      type: 'purchase',
    );
    expect(adapter.lastQueryParameters, {
      'outlet_id': 'outlet-id',
      'type': 'purchase',
      'search': 'Kopi',
    });

    final purchase = await repository.createPurchase(
      outletId: 'outlet-id',
      number: 'INV-001',
      supplierName: 'Supplier Kopi',
      items: const [
        PurchaseItemDraft(
          productId: 'product-id',
          quantity: 12,
          unitCost: 10000,
        ),
      ],
    );

    expect(inventory.items.first.stockOnHand, 12);
    expect(inventory.items.last.isNegative, true);
    expect(inventory.items.last.warning, 'NEGATIVE_STOCK_ALLOWED');
    expect(inventory.movements.single.quantityDelta, 12);
    expect(purchase.total, 120000);
    expect(adapter.lastRequestData, {
      'outlet_id': 'outlet-id',
      'number': 'INV-001',
      'supplier_name': 'Supplier Kopi',
      'items': [
        {'product_id': 'product-id', 'quantity': 12, 'unit_cost': 10000},
      ],
    });
  });

  test('ApiProductRepository manages products and categories', () async {
    final adapter = _RouteAdapter({
      'POST /categories': {
        'data': {'id': 'category-id', 'name': 'Snack'},
        'meta': {},
      },
      'PUT /categories/category-id': {
        'data': {'id': 'category-id', 'name': 'Camilan'},
        'meta': {},
      },
      'POST /products': {
        'data': {
          'id': 'product-id',
          'product_category_id': 'category-id',
          'name': 'Keripik Singkong',
          'barcode': '899100000001',
          'price': 15000,
          'track_stock': true,
        },
        'meta': {},
      },
      'PUT /products/product-id': {
        'data': {
          'id': 'product-id',
          'product_category_id': 'category-id',
          'name': 'Keripik Balado',
          'barcode': '899100000002',
          'price': 17000,
          'track_stock': false,
        },
        'meta': {},
      },
    });
    final repository = ApiProductRepository(
      apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
    );

    final category = await repository.createCategory(name: 'Snack');
    final updatedCategory = await repository.updateCategory(
      id: category.id,
      name: 'Camilan',
    );
    final product = await repository.createProduct(
      outletId: 'outlet-id',
      productCategoryId: category.id,
      name: 'Keripik Singkong',
      category: category.name,
      barcode: '899100000001',
      price: 15000,
      trackStock: true,
    );
    final updatedProduct = await repository.updateProduct(
      id: product.id,
      outletId: 'outlet-id',
      productCategoryId: updatedCategory.id,
      name: 'Keripik Balado',
      category: updatedCategory.name,
      barcode: '899100000002',
      price: 17000,
      trackStock: false,
    );

    expect(category.name, 'Snack');
    expect(updatedCategory.name, 'Camilan');
    expect(product.price, 15000);
    expect(product.category, 'Snack');
    expect(updatedProduct.name, 'Keripik Balado');
    expect(updatedProduct.price, 17000);
    expect(updatedProduct.category, 'Camilan');
    expect(
      adapter.requestDataByKey['POST /products'],
      containsPair('product_category_id', 'category-id'),
    );
    expect(
      adapter.requestDataByKey['PUT /products/product-id'],
      containsPair('product_category_id', 'category-id'),
    );
  });

  test('ApiProductRepository surfaces cashier read-only 403', () async {
    final repository = ApiProductRepository(
      apiClient: ApiClient(
        dio: Dio()
          ..httpClientAdapter = _RouteAdapter({
            'POST /products': {
              'status': 403,
              'error': {
                'code': 'FORBIDDEN',
                'message': 'Only owner or admin can manage products.',
                'details': {},
              },
            },
          }),
      ),
    );

    await expectLater(
      repository.createProduct(
        outletId: 'outlet-id',
        productCategoryId: 'category-id',
        name: 'Tidak Boleh',
        price: 1000,
      ),
      throwsA(isA<ForbiddenApiException>()),
    );
  });

  test('ApiShiftRepository opens, reads, and closes current shift', () async {
    final repository = ApiShiftRepository(
      apiClient: ApiClient(
        dio: Dio()
          ..httpClientAdapter = _RouteAdapter({
            'POST /shifts/open': {
              'data': {
                'id': 'shift-id',
                'cashier_id': 'cashier-id',
                'status': 'open',
                'opening_cash': 100000,
              },
              'meta': {},
            },
            'GET /shifts/current': {
              'data': {
                'shift': {
                  'id': 'shift-id',
                  'cashier_id': 'cashier-id',
                  'status': 'open',
                  'opening_cash': 100000,
                },
              },
              'meta': {},
            },
            'POST /shifts/shift-id/close': {
              'data': {
                'id': 'shift-id',
                'cashier_id': 'cashier-id',
                'status': 'closed',
                'opening_cash': 100000,
                'expected_cash': 120000,
                'actual_cash': 119000,
                'cash_difference': -1000,
              },
              'meta': {},
            },
          }),
      ),
    );

    final opened = await repository.openShift(
      outletId: 'outlet-id',
      deviceId: 'device-id',
      cashierId: 'cashier-id',
      openingCash: 100000,
    );
    final current = await repository.currentShift(
      outletId: 'outlet-id',
      deviceId: 'device-id',
    );
    final closed = await repository.closeShift(
      shiftId: 'shift-id',
      actualCash: 119000,
      pin: '1234',
    );

    expect(opened.id, 'shift-id');
    expect(current?.isOpen, true);
    expect(closed.status, 'closed');
    expect(closed.cashDifference, -1000);
  });

  test('ApiCustomerRepository searches and creates customers', () async {
    final adapter = _RouteAdapter({
      'GET /customers': {
        'data': {
          'customers': [
            {
              'id': 'customer-id',
              'name': 'Ani Pelanggan',
              'phone': '0812345678',
              'group': 'VIP',
            },
          ],
        },
        'meta': {},
      },
      'POST /customers': {
        'data': {
          'id': 'new-customer-id',
          'name': 'Budi Baru',
          'phone': '0812999999',
          'group': 'Regular',
        },
        'meta': {},
      },
    });
    final repository = ApiCustomerRepository(
      apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
    );

    final customers = await repository.searchCustomers(
      search: 'Ani',
      group: 'VIP',
    );
    expect(adapter.lastQueryParameters, {'search': 'Ani', 'group': 'VIP'});

    final created = await repository.createCustomer(
      name: 'Budi Baru',
      phone: '0812999999',
      group: 'Regular',
    );

    expect(customers.single.id, 'customer-id');
    expect(created.name, 'Budi Baru');
    expect(adapter.lastRequestData, {
      'name': 'Budi Baru',
      'phone': '0812999999',
      'group': 'Regular',
    });
  });

  test('ApiReportRepository maps current shift sales summary', () async {
    final adapter = _RouteAdapter({
      'GET /reports/sales-summary': {
        'data': {
          'total_sales': 150000,
          'transaction_count': 2,
          'payment_totals': [
            {'method': 'Tunai', 'amount': 100000},
            {'method': 'QRIS Statis', 'amount': 50000},
          ],
          'total_discount': 12000,
          'total_void': 30000,
          'average_transaction_value': 75000,
          'chart': [
            {'label': '2026-06-18 10:00', 'amount': 100000},
            {'label': '2026-06-18 11:00', 'amount': 50000},
          ],
        },
        'meta': {},
      },
    });
    final repository = ApiReportRepository(
      apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
    );

    final summary = await repository.salesSummary(
      date: DateTime(2026, 6, 18),
      shiftId: 'shift-id',
      outletId: 'outlet-id',
    );

    expect(adapter.lastQueryParameters, {
      'date': '2026-06-18',
      'shift_id': 'shift-id',
      'outlet_id': 'outlet-id',
    });
    expect(summary.totalSales, 150000);
    expect(summary.transactionCount, 2);
    expect(
      summary.paymentTotals.singleWhere((p) => p.method == 'Tunai').amount,
      100000,
    );
    expect(summary.totalDiscount, 12000);
    expect(summary.totalVoid, 30000);
    expect(summary.averageTransactionValue, 75000);
    expect(summary.chartPoints.last.amount, 50000);
  });

  test(
    'ApiAttendanceRepository loads staff history and clocks with PIN',
    () async {
      final adapter = _RouteAdapter({
        'GET /staff': {
          'data': {
            'staff': [
              {
                'id': 'cashier-id',
                'name': 'Siti Kasir',
                'role': 'cashier',
                'email': 'siti@example.test',
              },
            ],
          },
          'meta': {},
        },
        'GET /attendance': {
          'data': {
            'attendance': [
              {
                'id': 'attendance-id',
                'staff_id': 'cashier-id',
                'staff': {
                  'id': 'cashier-id',
                  'name': 'Siti Kasir',
                  'role': 'cashier',
                },
                'clock_in_at': '2026-06-18T09:00:00Z',
                'clock_out_at': null,
              },
            ],
          },
          'meta': {},
        },
        'POST /attendance': {
          'data': {
            'id': 'attendance-id',
            'staff_id': 'cashier-id',
            'staff': {
              'id': 'cashier-id',
              'name': 'Siti Kasir',
              'role': 'cashier',
            },
            'clock_in_at': '2026-06-18T09:00:00Z',
            'clock_out_at': '2026-06-18T17:00:00Z',
          },
          'meta': {},
        },
      });
      final repository = ApiAttendanceRepository(
        apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
      );

      final staff = await repository.getStaff(outletId: 'outlet-id');
      expect(adapter.lastQueryParameters, {'outlet_id': 'outlet-id'});

      final history = await repository.getAttendance(
        outletId: 'outlet-id',
        date: DateTime(2026, 6, 18),
        staffId: 'cashier-id',
      );
      expect(adapter.lastQueryParameters, {
        'outlet_id': 'outlet-id',
        'date': '2026-06-18',
        'staff_id': 'cashier-id',
      });

      final record = await repository.clockAttendance(
        employeeId: 'cashier-id',
        pin: '123456',
        action: AttendanceAction.clockOut,
        outletId: 'outlet-id',
      );

      expect(staff.single.name, 'Siti Kasir');
      expect(history.single.isOpen, true);
      expect(record.isOpen, false);
      expect(adapter.lastRequestData, {
        'outlet_id': 'outlet-id',
        'staff_id': 'cashier-id',
        'pin': '123456',
        'action': 'clock_out',
      });
    },
  );

  test('ApiAttendanceRepository rejects wrong PIN clearly', () async {
    final repository = ApiAttendanceRepository(
      apiClient: ApiClient(
        dio: Dio()
          ..httpClientAdapter = _RouteAdapter({
            'POST /attendance': {
              'status': 422,
              'error': {
                'code': 'INVALID_PIN',
                'message': 'PIN absensi salah.',
                'details': {'remaining_attempts': 2},
              },
            },
          }),
      ),
    );

    await expectLater(
      repository.clockAttendance(
        employeeId: 'cashier-id',
        pin: '000000',
        action: AttendanceAction.clockIn,
        outletId: 'outlet-id',
      ),
      throwsA(
        isA<ValidationApiException>().having(
          (error) => error.code,
          'code',
          'INVALID_PIN',
        ),
      ),
    );
  });

  test('CheckoutDraft sends customer discounts notes and served by', () {
    final draft = CheckoutDraft(
      idempotencyKey: 'checkout-key',
      outletId: 'outlet-id',
      deviceId: 'device-id',
      cashierId: 'cashier-id',
      shiftId: 'shift-id',
      customerId: 'customer-id',
      servedBy: 'server-id',
      cartDiscount: 3000,
      notes: 'Meja 7',
      items: const [
        CheckoutItem(
          productId: 'product-id',
          name: 'Nasi Campur',
          quantity: 2,
          unitPrice: 50000,
          discount: 10000,
        ),
      ],
      payments: [CheckoutPayment(method: 'Tunai', amount: 87000, isCash: true)],
      paidAmount: 87000,
    );

    expect(draft.toApiJson(), {
      'outlet_id': 'outlet-id',
      'device_id': 'device-id',
      'cashier_id': 'cashier-id',
      'shift_id': 'shift-id',
      'customer_id': 'customer-id',
      'served_by': 'server-id',
      'cart_discount': 3000,
      'rounding': 0,
      'notes': 'Meja 7',
      'items': [
        {
          'product_id': 'product-id',
          'quantity': 2,
          'unit_price': 50000,
          'discount': 10000,
        },
      ],
      'payments': [
        {'method': 'Tunai', 'amount': 87000},
      ],
    });
  });

  test('CheckoutTransaction maps discount breakdown from API response', () {
    final transaction = CheckoutTransaction.fromJson({
      'id': 'transaction-id',
      'number': 'TRX-001',
      'status': 'paid',
      'grand_total': 87000,
      'item_discount_total': 10000,
      'cart_discount_total': 3000,
    });

    expect(transaction.itemDiscountTotal, 10000);
    expect(transaction.cartDiscountTotal, 3000);
    expect(transaction.grandTotal, 87000);
  });

  test('CheckoutDraft should support split payments and held invoices', () {
    final split = CheckoutDraft(
      idempotencyKey: 'split-key',
      outletId: 'outlet-id',
      deviceId: 'device-id',
      cashierId: 'cashier-id',
      shiftId: 'shift-id',
      items: const [
        CheckoutItem(
          productId: 'product-id',
          name: 'Paket Hemat',
          quantity: 1,
          unitPrice: 100000,
        ),
      ],
      payments: [
        CheckoutPayment(method: 'Tunai', amount: 40000, isCash: true),
        CheckoutPayment(
          method: 'QRIS Statis',
          amount: 60000,
          isCash: false,
          reference: 'QR-123',
        ),
      ],
      paidAmount: 100000,
    );
    final held = CheckoutDraft(
      idempotencyKey: 'held-key',
      outletId: 'outlet-id',
      deviceId: 'device-id',
      cashierId: 'cashier-id',
      shiftId: 'shift-id',
      status: 'held',
      items: const [
        CheckoutItem(
          productId: 'product-id',
          name: 'Paket Hemat',
          quantity: 1,
          unitPrice: 100000,
        ),
      ],
      payments: const [],
      paidAmount: 0,
    );

    expect(split.toApiJson()['payments'], [
      {'method': 'Tunai', 'amount': 40000},
      {'method': 'QRIS Statis', 'amount': 60000, 'reference': 'QR-123'},
    ]);
    expect(held.toApiJson()['status'], 'held');
    expect(held.toApiJson()['payments'], isNull);
  });

  test('single non-cash payment requires intentional manual reference', () {
    const qris = PaymentMethodConfig(method: 'QRIS Statis', isCash: false);
    const cash = PaymentMethodConfig(method: 'Tunai', isCash: true);

    expect(
      () => buildSingleCheckoutPayment(
        method: qris,
        total: 50000,
        paidAmount: 0,
        reference: '',
      ),
      throwsArgumentError,
    );

    final payment = buildSingleCheckoutPayment(
      method: qris,
      total: 50000,
      paidAmount: 0,
      reference: 'QR-20260618-001',
    );
    final cashPayment = buildSingleCheckoutPayment(
      method: cash,
      total: 50000,
      paidAmount: 100000,
      reference: '',
    );

    expect(payment.reference, 'QR-20260618-001');
    expect(payment.amount, 50000);
    expect(cashPayment.reference, isNull);
    expect(cashPayment.amount, 100000);
    expect(
      () => requireNonCashReferences(const [
        CheckoutPayment(method: 'Tunai', amount: 40000, isCash: true),
        CheckoutPayment(method: 'QRIS Statis', amount: 60000, isCash: false),
      ]),
      throwsArgumentError,
    );
    expect(
      () => requireNonCashReferences(const [
        CheckoutPayment(method: 'Tunai', amount: 40000, isCash: true),
        CheckoutPayment(
          method: 'QRIS Statis',
          amount: 60000,
          isCash: false,
          reference: 'QR-SPLIT-001',
        ),
      ]),
      returnsNormally,
    );
  });

  test(
    'ApiTransactionRepository should confirm payments list held and void',
    () async {
      final adapter = _RouteAdapter({
        'POST /payments': {
          'data': {
            'id': 'payment-id',
            'transaction_id': 'transaction-id',
            'method': 'QRIS Statis',
            'amount': 60000,
            'reference': 'QR-123',
            'status': 'confirmed',
            'is_cash': false,
            'confirmed_by': 'cashier-id',
            'confirmed_at': '2026-06-18T10:00:00Z',
            'transaction_status': 'paid',
          },
          'meta': {},
        },
        'GET /transactions': {
          'data': {
            'transactions': [
              {
                'id': 'held-id',
                'number': 'TRX-HELD',
                'status': 'held',
                'grand_total': 100000,
              },
            ],
          },
          'meta': {},
        },
        'POST /voids': {
          'data': {
            'transaction_id': 'transaction-id',
            'status': 'voided',
            'reason': 'Salah input',
          },
          'meta': {},
        },
      });
      final repository = ApiTransactionRepository(
        apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
      );

      final payment = await repository.confirmPayment(
        paymentId: 'payment-id',
        idempotencyKey: 'confirm-key',
      );
      final held = await repository.listTransactions(status: 'held');
      expect(adapter.lastQueryParameters, {'status': 'held'});

      final voided = await repository.voidTransaction(
        transactionId: 'transaction-id',
        shiftId: 'shift-id',
        reason: 'Salah input',
        idempotencyKey: 'void-key',
      );

      expect(payment.status, 'confirmed');
      expect(payment.isCash, false);
      expect(payment.transactionStatus, 'paid');
      expect(held.single.status, 'held');
      expect(voided.status, 'voided');
      expect(adapter.idempotencyKeys, containsAll(['confirm-key', 'void-key']));
    },
  );

  test(
    'ApiTransactionRepository maps parked order revision and lease endpoints',
    () async {
      final adapter = _RouteAdapter({
        'GET /parked-orders': {
          'data': {
            'parked_orders': [
              {
                'id': 'held-id',
                'number': 'PARK-001',
                'status': 'held',
                'revision': 3,
                'grand_total': 45000,
                'lease': {
                  'device_id': 'device-a',
                  'user_id': 'cashier-a',
                  'expires_at': '2026-06-20T05:00:00Z',
                  'remaining_seconds': 80,
                },
              },
            ],
          },
          'meta': {},
        },
        'POST /parked-orders': {
          'data': {
            'id': 'held-id',
            'number': 'PARK-001',
            'status': 'held',
            'revision': 1,
            'grand_total': 45000,
          },
          'meta': {},
        },
        'PUT /parked-orders/held-id': {
          'data': {
            'id': 'held-id',
            'number': 'PARK-001',
            'status': 'held',
            'revision': 4,
            'grand_total': 90000,
          },
          'meta': {},
        },
        'POST /parked-orders/held-id/lease/acquire': {
          'data': {
            'id': 'held-id',
            'number': 'PARK-001',
            'status': 'held',
            'revision': 3,
            'grand_total': 45000,
            'lease': {'device_id': 'device-a', 'user_id': 'cashier-a'},
          },
          'meta': {},
        },
        'POST /parked-orders/held-id/lease/refresh': {
          'data': {
            'id': 'held-id',
            'number': 'PARK-001',
            'status': 'held',
            'revision': 3,
            'grand_total': 45000,
            'lease': {'device_id': 'device-a', 'user_id': 'cashier-a'},
          },
          'meta': {},
        },
        'POST /parked-orders/held-id/lease/release': {
          'data': {
            'id': 'held-id',
            'number': 'PARK-001',
            'status': 'held',
            'revision': 3,
            'grand_total': 45000,
            'lease': {'device_id': null, 'user_id': null},
          },
          'meta': {},
        },
      });
      final repository = ApiTransactionRepository(
        apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
      );
      const item = CheckoutItem(
        productId: 'product-id',
        name: 'Kopi',
        quantity: 1,
        unitPrice: 45000,
      );
      const draft = CheckoutDraft(
        idempotencyKey: 'unused-for-parked',
        outletId: 'outlet-id',
        deviceId: 'device-a',
        cashierId: 'cashier-a',
        shiftId: 'shift-id',
        items: [item],
        payments: [],
        paidAmount: 0,
      );

      final listed = await repository.listParkedOrders();
      final created = await repository.createParkedOrder(draft);
      final acquired = await repository.acquireParkedOrderLease(
        transactionId: 'held-id',
        deviceId: 'device-a',
      );
      final refreshed = await repository.refreshParkedOrderLease(
        transactionId: 'held-id',
        deviceId: 'device-a',
      );
      final updated = await repository.updateParkedOrder(
        transactionId: 'held-id',
        deviceId: 'device-a',
        expectedRevision: 3,
        items: const [item],
        notes: 'Tambah es',
      );
      final released = await repository.releaseParkedOrderLease(
        transactionId: 'held-id',
        deviceId: 'device-a',
      );

      expect(listed.single.revision, 3);
      expect(listed.single.leaseDeviceId, 'device-a');
      expect(listed.single.leaseRemainingSeconds, 80);
      expect(created.revision, 1);
      expect(acquired.leaseUserId, 'cashier-a');
      expect(refreshed.leaseDeviceId, 'device-a');
      expect(updated.revision, 4);
      expect(released.leaseDeviceId, isNull);
      expect(adapter.requestDataByKey['PUT /parked-orders/held-id'], {
        'device_id': 'device-a',
        'expected_revision': 3,
        'cart_discount': 0,
        'notes': 'Tambah es',
        'items': [item.toJson()],
      });
    },
  );

  test(
    'ApiTransactionRepository should map server transaction details',
    () async {
      final adapter = _RouteAdapter({
        'GET /transactions': {
          'data': {
            'transactions': [
              {
                'id': 'transaction-id',
                'number': 'TRX-001',
                'status': 'paid',
                'grand_total': 36000,
                'item_discount_total': 2000,
                'cart_discount_total': 1000,
                'created_at': '2026-06-18T10:00:00Z',
                'notes': 'Tanpa gula',
                'served_by': 'staff-id',
                'customer': {'id': 'customer-id', 'name': 'Ani'},
                'cashier': {
                  'id': 'cashier-id',
                  'name': 'Siti',
                  'role': 'cashier',
                },
                'items': [
                  {
                    'product_id': 'product-id',
                    'name': 'Kopi Susu',
                    'quantity': 2,
                    'unit_price': 18000,
                    'discount': 2000,
                  },
                ],
                'payments': [
                  {
                    'id': 'payment-id',
                    'transaction_id': 'transaction-id',
                    'method': 'Tunai',
                    'amount': 36000,
                    'status': 'confirmed',
                    'is_cash': true,
                  },
                ],
              },
            ],
          },
          'meta': {},
        },
      });
      final repository = ApiTransactionRepository(
        apiClient: ApiClient(dio: Dio()..httpClientAdapter = adapter),
      );

      final transactions = await repository.listTransactions(status: 'paid');

      expect(adapter.lastQueryParameters, {'status': 'paid'});
      expect(transactions.single.items.single.name, 'Kopi Susu');
      expect(transactions.single.items.single.discount, 2000);
      expect(transactions.single.customerId, 'customer-id');
      expect(transactions.single.customerName, 'Ani');
      expect(transactions.single.cashierName, 'Siti');
      expect(transactions.single.notes, 'Tanpa gula');
      expect(
        transactions.single.createdAt,
        DateTime.parse('2026-06-18T10:00:00Z'),
      );
    },
  );

  test('SalesTransaction should calculate change from cash only', () {
    final transaction = SalesTransaction(
      id: 'transaction-id',
      number: 'TRX-001',
      order: SalesOrder(
        id: 'order-id',
        number: 'ORD-001',
        type: OrderType.dineIn,
        status: OrderStatus.paid,
        lines: const [
          OrderLine(
            productId: 'product-id',
            name: 'Paket Hemat',
            quantity: 1,
            unitPrice: 100000,
          ),
        ],
        createdAt: DateTime(2026, 6, 18),
      ),
      payments: [
        PaymentLine(
          method: PaymentMethod.cash,
          methodName: 'Tunai',
          amount: 50000,
          isCash: true,
        ),
        PaymentLine(
          method: PaymentMethod.qris,
          methodName: 'QRIS Statis',
          amount: 60000,
          isCash: false,
          reference: 'QR-123',
          status: 'pending',
        ),
      ],
      cashier: const Employee(id: 'cashier-id', name: 'Kasir', role: 'cashier'),
      createdAt: DateTime(2026, 6, 18),
    );

    expect(transaction.paidAmount, 110000);
    expect(transaction.change, 10000);
  });

  test(
    'ApiTransactionRepository should surface void rejection errors',
    () async {
      final forbidden = ApiTransactionRepository(
        apiClient: ApiClient(
          dio: Dio()
            ..httpClientAdapter = _RouteAdapter({
              'POST /voids': {
                'status': 403,
                'error': {
                  'code': 'FORBIDDEN',
                  'message': 'User is not allowed to void transactions.',
                  'details': {},
                },
              },
            }),
        ),
      );
      final shiftMismatch = ApiTransactionRepository(
        apiClient: ApiClient(
          dio: Dio()
            ..httpClientAdapter = _RouteAdapter({
              'POST /voids': {
                'status': 422,
                'error': {
                  'code': 'VOID_SHIFT_MISMATCH',
                  'message':
                      'Transaction can only be voided from the same shift.',
                  'details': {},
                },
              },
            }),
        ),
      );

      await expectLater(
        forbidden.voidTransaction(
          transactionId: 'transaction-id',
          shiftId: 'shift-id',
          reason: 'Salah input',
          idempotencyKey: 'void-key',
        ),
        throwsA(isA<ForbiddenApiException>()),
      );
      await expectLater(
        shiftMismatch.voidTransaction(
          transactionId: 'transaction-id',
          shiftId: 'other-shift-id',
          reason: 'Salah input',
          idempotencyKey: 'void-key',
        ),
        throwsA(
          isA<ValidationApiException>().having(
            (error) => error.code,
            'code',
            'VOID_SHIFT_MISMATCH',
          ),
        ),
      );
    },
  );
}

class _RouteAdapter implements HttpClientAdapter {
  _RouteAdapter(this.responses);

  final Map<String, Map<String, Object?>> responses;
  Map<String, Object?>? lastQueryParameters;
  Object? lastRequestData;
  final Map<String, Object?> requestDataByKey = {};
  final List<String> requests = [];
  final List<String> idempotencyKeys = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final key = '${options.method} ${options.path}';
    requests.add(key);
    lastQueryParameters = Map<String, Object?>.from(options.queryParameters);
    lastRequestData = options.data;
    requestDataByKey[key] = options.data;
    final idempotencyKey = options.headers['Idempotency-Key'];
    if (idempotencyKey is String) idempotencyKeys.add(idempotencyKey);
    final response = responses[key];
    if (response == null) {
      return ResponseBody.fromString(
        jsonEncode({
          'error': {'code': 'NOT_FOUND', 'message': key, 'details': {}},
        }),
        404,
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
