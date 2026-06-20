import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:nojpos_tablet_ui/app/nojpos_app.dart';

const _apiBaseUrl = 'http://10.0.2.2:8000/api/v1';
const _ownerEmail = 'owner@demo.nojpos.test';
const _password = 'password';
const _outletId = '22222222-2222-4222-8222-222222222222';
const _productAirId = '88888888-8888-4888-8888-888888888888';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('Cashier core flow E2E', () {
    testWidgets('should verify seeded backend runtime flow', (tester) async {
      final report = _RuntimeReport();
      final api = _ApiProbe();

      await const FlutterSecureStorage().deleteAll();
      await tester.pumpWidget(const ProviderScope(child: NojposApp()));

      await report.step(
        'login akun -> outlet -> PIN -> shift -> POS',
        () async {
          await _waitForText(tester, 'Masuk ke NojPOS');
          await _replaceText(
            tester,
            const ValueKey('login_email'),
            _ownerEmail,
          );
          await _replaceText(
            tester,
            const ValueKey('login_password'),
            _password,
          );
          await _tap(tester, const ValueKey('login_submit'));

          await _waitForAnyText(tester, ['Pilih Outlet', 'Masukkan PIN Kasir']);
          if (find.text('Pilih Outlet').evaluate().isNotEmpty) {
            await _tap(tester, const ValueKey('outlet_$_outletId'));
          }

          await _waitForText(tester, 'Masukkan PIN Kasir');
          for (final digit in ['1', '2', '3', '4']) {
            await _tap(tester, ValueKey('pin_$digit'));
          }

          await _waitForAnyText(tester, ['Buka Shift', 'NojPOS']);
          if (find.text('Buka Shift').evaluate().isNotEmpty) {
            await _replaceText(
              tester,
              const ValueKey('shift_opening_cash'),
              '100000',
            );
            await _tap(tester, const ValueKey('shift_open_submit'));
          }

          await _waitForText(tester, 'Air Mineral Demo');
        },
      );

      await report.step(
        'load produk API -> tambah item -> diskon item/cart',
        () async {
          await _tap(tester, const ValueKey('product_$_productAirId'));
          await _tap(tester, const ValueKey('item_discount_$_productAirId'));
          await _waitForText(tester, 'Diskon Air Mineral Demo');
          await _replaceOnlyTextField(tester, '1000');
          await tester.tap(find.text('Simpan').last);
          await tester.pump(const Duration(seconds: 1));

          await _tap(tester, const ValueKey('cart_discount_button'));
          await _waitForText(tester, 'Diskon Cart');
          await _replaceOnlyTextField(tester, '500');
          await tester.tap(find.text('Simpan').last);
          await tester.pump(const Duration(seconds: 1));
          await _waitForText(tester, 'Rp 4.500');
        },
      );

      var cashTransaction = <String, Object?>{};
      await report.step('checkout TUNAI -> success/struk', () async {
        await _tap(tester, const ValueKey('open_payment_button'));
        await _waitForText(tester, 'Pembayaran');
        await _waitForText(tester, 'Service charge');
        await _waitForText(tester, 'Pajak');
        await _waitForText(tester, 'Rp 5.244');
        await _tap(tester, const ValueKey('payment_amount_exact'));
        await _waitForText(tester, 'Tunai   Rp 5.244');
        await _tap(tester, const ValueKey('process_payment_button'));
        await _waitForText(tester, 'Sukses!');
        cashTransaction = await api.latestTransaction();
        final status = cashTransaction['status'];
        final serverTotal = cashTransaction['grand_total'];
        if (status != 'paid') {
          throw StateError(
            'Expected cash checkout status paid, got $status. '
            'Payment screen total was Rp 5.244, server grand_total was ${_rupiah(serverTotal)}.',
          );
        }
        await _waitForText(tester, _rupiah(serverTotal));
      });

      await report.step(
        'GET /transactions sales history contains cash transaction',
        () async {
          final transactions = await api.transactions();
          final exists = transactions.any(
            (row) => row['id'] == cashTransaction['id'],
          );
          if (!exists) {
            throw StateError(
              'Cash transaction ${cashTransaction['id']} not found in GET /transactions.',
            );
          }
        },
      );

      await report.step('checkout NON-TUNAI -> confirm pending -> paid', () async {
        await _tap(tester, const ValueKey('success_done'));
        await _waitForText(tester, 'Air Mineral Demo');
        await _tap(tester, const ValueKey('product_$_productAirId'));
        await _tap(tester, const ValueKey('open_payment_button'));
        await _waitForText(tester, 'Pembayaran');
        await _waitForText(tester, 'Rp 6.993');
        await _tap(tester, const ValueKey('payment_method_QRIS Statis'));
        await _replaceText(
          tester,
          const ValueKey('payment_reference'),
          'QRIS-E2E-001',
        );
        await _tap(tester, const ValueKey('process_payment_button'));
        await _waitForText(tester, 'Sukses!');
        final nonCash = await api.latestTransaction();
        final payments = _list(nonCash['payments']);
        final pending = payments
            .where((payment) => payment['status'] == 'pending')
            .toList();
        if (pending.isEmpty) {
          throw StateError(
            'No pending non-cash payment returned from /transactions response.',
          );
        }

        await _tap(tester, const ValueKey('confirm_pending_payment'));
        await tester.pump(const Duration(seconds: 2));
        final updated = await api.transaction(nonCash['id'] as String);
        if (updated['status'] != 'paid') {
          throw StateError(
            'Expected non-cash transaction paid after confirm, got ${updated['status']}. '
            'Server grand_total=${_rupiah(updated['grand_total'])}, payment amount=${_rupiah(pending.first['amount'])}.',
          );
        }
      });

      await report.step('void same-shift -> status voided', () async {
        final paid = await api.latestPaidTransaction();
        await _tap(tester, const ValueKey('success_done'));
        await _waitForText(tester, 'Air Mineral Demo');
        await _tap(tester, const ValueKey('topbar_menu'));
        await _tap(tester, const ValueKey('drawer_penjualan'));
        await _waitForText(tester, 'Penjualan');
        await _tap(tester, ValueKey('sales_transaction_${paid['id']}'));
        await _waitForText(tester, 'Reason');
        await _replaceText(
          tester,
          const ValueKey('void_reason'),
          'E2E void same shift',
        );
        await _tap(tester, const ValueKey('void_submit'));
        await tester.pump(const Duration(seconds: 2));
        final voided = await api.transaction(paid['id'] as String);
        if (voided['status'] != 'voided') {
          throw StateError('Expected status voided, got ${voided['status']}.');
        }
      });

      await report.step(
        'close shift -> expected_cash & payment_totals tampil',
        () async {
          await _tap(tester, const ValueKey('ops_back'));
          await _waitForText(tester, 'Air Mineral Demo');
          await _tap(tester, const ValueKey('topbar_menu'));
          await _tap(tester, const ValueKey('drawer_tutup_kasir'));
          await _waitForText(tester, 'Expected cash');
          await _replaceText(
            tester,
            const ValueKey('close_shift_actual_cash'),
            '105244',
          );
          await tester.pump(const Duration(seconds: 1));
          await _tap(tester, const ValueKey('close_shift_submit'));
          await _waitForText(tester, 'Ringkasan Tutup Shift');
          await _waitForText(tester, 'Payment totals');
          final closed = await api.currentShift();
          if (closed != null) {
            throw StateError(
              'Expected no open current shift after close, but server still returned ${closed['id']}.',
            );
          }
        },
      );

      report.printSummary();
      expect(report.failures, isEmpty, reason: report.failureSummary);
    });
  });
}

Future<void> _tap(WidgetTester tester, Key key) async {
  final finder = find.byKey(key);
  await _waitForFinder(tester, finder);
  await tester.ensureVisible(finder);
  await tester.tap(finder);
  await tester.pump(const Duration(milliseconds: 500));
}

Future<void> _replaceText(WidgetTester tester, Key key, String text) async {
  final finder = find.byKey(key);
  await _waitForFinder(tester, finder);
  await tester.ensureVisible(finder);
  await tester.tap(finder);
  await tester.enterText(finder, text);
  await tester.pump(const Duration(milliseconds: 300));
}

Future<void> _replaceOnlyTextField(WidgetTester tester, String text) async {
  final finder = find.byType(TextField).last;
  await _waitForFinder(tester, finder);
  await tester.enterText(finder, text);
  await tester.pump(const Duration(milliseconds: 300));
}

Future<void> _waitForText(WidgetTester tester, String text) {
  return _waitForFinder(tester, find.text(text));
}

Future<void> _waitForAnyText(WidgetTester tester, List<String> labels) async {
  final end = DateTime.now().add(const Duration(seconds: 30));
  while (DateTime.now().isBefore(end)) {
    if (labels.any((label) => find.text(label).evaluate().isNotEmpty)) {
      return;
    }
    await tester.pump(const Duration(milliseconds: 250));
  }
  throw StateError('Timed out waiting for any text: ${labels.join(', ')}');
}

Future<void> _waitForFinder(WidgetTester tester, Finder finder) async {
  final end = DateTime.now().add(const Duration(seconds: 30));
  while (DateTime.now().isBefore(end)) {
    if (finder.evaluate().isNotEmpty) return;
    await tester.pump(const Duration(milliseconds: 250));
  }
  throw StateError('Timed out waiting for $finder');
}

List<Map<String, Object?>> _list(Object? value) {
  if (value is! List) return const [];
  return [
    for (final item in value)
      if (item is Map) Map<String, Object?>.from(item),
  ];
}

String _rupiah(Object? value) {
  final amount = (value as num?)?.toInt() ?? 0;
  final chars = amount.toString().split('').reversed.toList();
  final buffer = StringBuffer('Rp ');
  for (var index = chars.length - 1; index >= 0; index -= 1) {
    buffer.write(chars[index]);
    if (index > 0 && index % 3 == 0) {
      buffer.write('.');
    }
  }
  return buffer.toString();
}

class _RuntimeReport {
  final rows = <({String label, String status, String detail})>[];

  List<String> get failures => [
    for (final row in rows)
      if (row.status == 'BROKEN') '${row.label}: ${row.detail}',
  ];

  String get failureSummary => failures.join('\n');

  Future<void> step(String label, Future<void> Function() body) async {
    try {
      await body();
      rows.add((label: label, status: 'WORKING', detail: 'OK'));
    } catch (error, stackTrace) {
      final location = stackTrace.toString().split('\n').firstOrNull ?? '';
      rows.add((
        label: label,
        status: 'BROKEN',
        detail: '$error ${location.isEmpty ? '' : 'at $location'}',
      ));
    }
  }

  void printSummary() {
    // ignore: avoid_print
    print('NOJPOS_E2E_RUNTIME_STATUS');
    for (final row in rows) {
      // ignore: avoid_print
      print('${row.status} | ${row.label} | ${row.detail}');
    }
  }
}

class _ApiProbe {
  _ApiProbe()
    : _dio = Dio(
        BaseOptions(
          baseUrl: _apiBaseUrl,
          headers: const {'Accept': 'application/json'},
          responseType: ResponseType.json,
        ),
      );

  final Dio _dio;
  String? _token;

  Future<String> get token async {
    final existing = _token;
    if (existing != null) return existing;
    final response = await _dio.post<Object?>(
      '/auth/login',
      data: const {
        'email': _ownerEmail,
        'password': _password,
        'device_uuid': 'core-flow-probe',
      },
    );
    final data = _map(response.data)['data'];
    _token = _map(data)['token'] as String;
    return _token!;
  }

  Future<List<Map<String, Object?>>> transactions() async {
    final response = await _dio.get<Object?>(
      '/transactions',
      options: Options(headers: {'Authorization': 'Bearer ${await token}'}),
    );
    return _list(_map(_map(response.data)['data'])['transactions']);
  }

  Future<Map<String, Object?>> latestTransaction() async {
    final rows = await transactions();
    if (rows.isEmpty) {
      throw StateError('GET /transactions returned empty list.');
    }
    return rows.first;
  }

  Future<Map<String, Object?>> latestPaidTransaction() async {
    final rows = await transactions();
    return rows.firstWhere(
      (row) => row['status'] == 'paid',
      orElse: () =>
          throw StateError('GET /transactions has no paid transaction.'),
    );
  }

  Future<Map<String, Object?>> transaction(String id) async {
    final rows = await transactions();
    return rows.firstWhere(
      (row) => row['id'] == id,
      orElse: () => throw StateError('Transaction $id not found.'),
    );
  }

  Future<Map<String, Object?>?> currentShift() async {
    final response = await _dio.get<Object?>(
      '/shifts/current',
      queryParameters: const {'outlet_id': _outletId},
      options: Options(headers: {'Authorization': 'Bearer ${await token}'}),
    );
    final shift = _map(_map(response.data)['data'])['shift'];
    if (shift == null) return null;
    return _map(shift);
  }

  Map<String, Object?> _map(Object? value) {
    if (value is Map<String, Object?>) return value;
    if (value is Map) return Map<String, Object?>.from(value);
    return const {};
  }
}
