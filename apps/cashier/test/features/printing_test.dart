import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/core/printing/printer_settings_store.dart';
import 'package:nojpos_tablet_ui/core/printing/receipt_printing_service.dart';
import 'package:nojpos_tablet_ui/features/transactions/receipt_builder.dart';
import 'package:nojpos_tablet_ui/shared/models/nojpos_models.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test(
    'receipt builder uses paper width header and footer from outlet config',
    () {
      final transaction = _transaction();
      const outlet = Outlet(
        id: 'outlet-id',
        name: 'Outlet Demo',
        isOnline: true,
        receiptConfig: ReceiptConfig(
          paperWidth: ReceiptPaperWidth.mm80,
          header: 'Kedai Demo',
          footer: 'Terima kasih',
          showLogo: true,
          showQrisInfo: true,
        ),
      );

      final receipt = const ReceiptBuilder().build(transaction, outlet: outlet);

      expect(receipt, contains('Kedai Demo'));
      expect(receipt, contains('Paper: 80mm'));
      expect(receipt, contains('Logo: ditampilkan'));
      expect(receipt, contains('QRIS: tersedia di kasir'));
      expect(receipt, contains('Terima kasih'));
    },
  );

  test('printer default is persisted and used for printing', () async {
    final adapter = _RecordingPrinterAdapter();
    final store = InMemoryPrinterSettingsStore();
    final service = ReceiptPrintingService(
      adapter: adapter,
      receiptBuilder: const ReceiptBuilder(),
      settingsStore: store,
    );
    const printer = BluetoothPrinterDevice(
      name: 'Printer Kasir',
      address: 'AA:BB:CC',
    );

    final saveResult = await service.saveDefaultPrinter(printer);
    final result = await service.printReceipt(_transaction());

    expect(saveResult.success, true);
    expect(result.success, true);
    expect((await store.readSettings()).defaultPrinter?.address, 'AA:BB:CC');
    expect(adapter.connectedPrinters.map((item) => item.address), [
      'AA:BB:CC',
      'AA:BB:CC',
    ]);
    expect(adapter.writeCount, 1);
  });

  test(
    'printing without default printer is best effort and never blocks sale',
    () async {
      final service = ReceiptPrintingService(
        adapter: _RecordingPrinterAdapter(),
        receiptBuilder: const ReceiptBuilder(),
        settingsStore: InMemoryPrinterSettingsStore(),
      );

      final result = await service.printReceipt(_transaction());

      expect(result.success, false);
      expect(result.blocksSale, false);
      expect(result.canReprint, true);
      expect(result.message, contains('Printer default belum disimpan'));
    },
  );

  test(
    'printing failure is best effort and reprint remains available',
    () async {
      final service = ReceiptPrintingService(
        adapter: _FailingPrinterAdapter(),
        receiptBuilder: const ReceiptBuilder(),
        settingsStore: InMemoryPrinterSettingsStore(
          const PrinterSettings(
            defaultPrinter: BluetoothPrinterDevice(
              name: 'Printer Demo',
              address: 'AA:BB:CC',
            ),
          ),
        ),
      );

      final result = await service.printReceipt(
        _transaction(),
        outlet: const Outlet(
          id: 'outlet-id',
          name: 'Outlet Demo',
          isOnline: true,
          receiptConfig: ReceiptConfig(header: 'Outlet Demo'),
        ),
      );

      expect(result.success, false);
      expect(result.blocksSale, false);
      expect(result.canReprint, true);
      expect(result.message, contains('Printer offline'));
    },
  );
}

SalesTransaction _transaction() {
  return SalesTransaction(
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
          name: 'Kopi Susu',
          quantity: 2,
          unitPrice: 18000,
        ),
      ],
      createdAt: DateTime(2026, 6, 18, 10),
    ),
    payments: [PaymentLine(method: PaymentMethod.cash, amount: 40000)],
    cashier: const Employee(id: 'cashier-id', name: 'Siti', role: 'cashier'),
    createdAt: DateTime(2026, 6, 18, 10),
  );
}

class _RecordingPrinterAdapter implements BluetoothPrinterAdapter {
  final connectedPrinters = <BluetoothPrinterDevice>[];
  int writeCount = 0;

  @override
  Future<List<BluetoothPrinterDevice>> scan() async => const [
    BluetoothPrinterDevice(name: 'Printer Kasir', address: 'AA:BB:CC'),
  ];

  @override
  Future<bool> connect(BluetoothPrinterDevice printer) async {
    connectedPrinters.add(printer);
    return true;
  }

  @override
  Future<bool> writeBytes(List<int> bytes) async {
    writeCount += 1;
    return true;
  }
}

class _FailingPrinterAdapter implements BluetoothPrinterAdapter {
  @override
  Future<List<BluetoothPrinterDevice>> scan() async => const [];

  @override
  Future<bool> connect(BluetoothPrinterDevice printer) async => true;

  @override
  Future<bool> writeBytes(List<int> bytes) async {
    throw Exception('Printer offline');
  }
}
