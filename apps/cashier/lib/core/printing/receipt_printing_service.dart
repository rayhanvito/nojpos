import 'package:esc_pos_utils/esc_pos_utils.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:print_bluetooth_thermal/print_bluetooth_thermal.dart';

import '../../features/transactions/receipt_builder.dart';
import '../../shared/models/nojpos_models.dart';
import 'printer_settings_store.dart';

final receiptPrintingServiceProvider = Provider<ReceiptPrintingService>(
  (ref) => ReceiptPrintingService(
    adapter: const PrintBluetoothThermalAdapter(),
    receiptBuilder: const ReceiptBuilder(),
    settingsStore: ref.watch(printerSettingsStoreProvider),
  ),
);

abstract interface class BluetoothPrinterAdapter {
  Future<List<BluetoothPrinterDevice>> scan();

  Future<bool> connect(BluetoothPrinterDevice printer);

  Future<bool> writeBytes(List<int> bytes);
}

class BluetoothPrinterDevice {
  const BluetoothPrinterDevice({required this.name, required this.address});

  final String name;
  final String address;
}

class PrintBluetoothThermalAdapter implements BluetoothPrinterAdapter {
  const PrintBluetoothThermalAdapter();

  @override
  Future<List<BluetoothPrinterDevice>> scan() async {
    final printers = await PrintBluetoothThermal.pairedBluetooths;
    return [
      for (final printer in printers)
        BluetoothPrinterDevice(name: printer.name, address: printer.macAdress),
    ];
  }

  @override
  Future<bool> connect(BluetoothPrinterDevice printer) {
    return PrintBluetoothThermal.connect(macPrinterAddress: printer.address);
  }

  @override
  Future<bool> writeBytes(List<int> bytes) {
    return PrintBluetoothThermal.writeBytes(bytes);
  }
}

class ReceiptPrintingService {
  const ReceiptPrintingService({
    required BluetoothPrinterAdapter adapter,
    required ReceiptBuilder receiptBuilder,
    required PrinterSettingsStore settingsStore,
  }) : _adapter = adapter,
       _receiptBuilder = receiptBuilder,
       _settingsStore = settingsStore;

  final BluetoothPrinterAdapter _adapter;
  final ReceiptBuilder _receiptBuilder;
  final PrinterSettingsStore _settingsStore;

  Future<List<BluetoothPrinterDevice>> scanPrinters() => _adapter.scan();

  Future<PrinterSettings> loadSettings() => _settingsStore.readSettings();

  Future<ReceiptPrintResult> saveDefaultPrinter(
    BluetoothPrinterDevice printer,
  ) async {
    await _settingsStore.saveDefaultPrinter(printer);
    try {
      final connected = await _adapter.connect(printer);
      return connected
          ? ReceiptPrintResult.success(
              message: 'Printer default disimpan dan tersambung: ${printer.name}.',
            )
          : ReceiptPrintResult.failure(
              message:
                  'Printer default disimpan, tetapi belum tersambung. Transaksi tetap tidak diblokir.',
            );
    } catch (error) {
      return ReceiptPrintResult.failure(
        message:
            'Printer default disimpan, tetapi koneksi gagal: ${_messageFor(error)}. Transaksi tetap tidak diblokir.',
      );
    }
  }

  Future<void> clearDefaultPrinter() => _settingsStore.clearDefaultPrinter();

  Future<void> setCashDrawerEnabled(bool enabled) =>
      _settingsStore.setCashDrawerEnabled(enabled);

  Future<bool> selectPrinter(BluetoothPrinterDevice printer) {
    return _adapter.connect(printer);
  }

  String buildDigitalReceipt(SalesTransaction transaction, {Outlet? outlet}) {
    return _receiptBuilder.build(transaction, outlet: outlet);
  }

  Future<ReceiptPrintResult> printReceipt(
    SalesTransaction transaction, {
    Outlet? outlet,
  }) async {
    try {
      final settings = await _settingsStore.readSettings();
      final printer = settings.defaultPrinter;
      if (printer == null) {
        return const ReceiptPrintResult.failure(
          message:
              'Printer default belum disimpan. Transaksi tetap tersimpan dan bisa cetak ulang nanti.',
        );
      }
      final connected = await _adapter.connect(printer);
      if (!connected) {
        return ReceiptPrintResult.failure(
          message:
              'Printer ${printer.name} tidak tersambung. Transaksi tetap tersimpan.',
        );
      }
      final bytes = await buildEscPosBytes(
        transaction,
        outlet: outlet,
        kickCashDrawer:
            settings.cashDrawerEnabled && transaction.cashPaidAmount > 0,
      );
      final ok = await _adapter.writeBytes(bytes);
      if (!ok) {
        return const ReceiptPrintResult.failure(
          message: 'Printer tidak merespons. Transaksi tetap tersimpan.',
        );
      }
      return ReceiptPrintResult.success(
        message: settings.cashDrawerEnabled && transaction.cashPaidAmount > 0
            ? 'Struk dicetak dan perintah buka laci kas dikirim.'
            : 'Struk berhasil dicetak.',
      );
    } catch (error) {
      return ReceiptPrintResult.failure(
        message: '${_messageFor(error)}. Transaksi tetap tersimpan.',
      );
    }
  }

  Future<ReceiptPrintResult> reprint(
    SalesTransaction transaction, {
    Outlet? outlet,
  }) {
    return printReceipt(transaction, outlet: outlet);
  }

  Future<ReceiptPrintResult> testPrint() async {
    try {
      final settings = await _settingsStore.readSettings();
      final printer = settings.defaultPrinter;
      if (printer == null) {
        return const ReceiptPrintResult.failure(
          message: 'Printer default belum disimpan.',
        );
      }
      final connected = await _adapter.connect(printer);
      if (!connected) {
        return ReceiptPrintResult.failure(
          message: 'Printer ${printer.name} tidak tersambung.',
        );
      }
      final profile = await CapabilityProfile.load();
      final generator = Generator(PaperSize.mm58, profile);
      final ok = await _adapter.writeBytes([
        ...generator.text('NojPOS test print'),
        ...generator.text(DateTime.now().toIso8601String()),
        ...generator.feed(2),
        ...generator.cut(),
      ]);
      return ok
          ? const ReceiptPrintResult.success(message: 'Test print berhasil.')
          : const ReceiptPrintResult.failure(
              message: 'Printer tidak merespons saat test print.',
            );
    } catch (error) {
      return ReceiptPrintResult.failure(
        message: 'Test print gagal: ${_messageFor(error)}.',
      );
    }
  }

  Future<ReceiptPrintResult> kickCashDrawer() async {
    try {
      final settings = await _settingsStore.readSettings();
      if (!settings.cashDrawerEnabled) {
        return const ReceiptPrintResult.failure(
          message: 'Cash drawer sedang nonaktif di pengaturan.',
        );
      }
      final printer = settings.defaultPrinter;
      if (printer == null) {
        return const ReceiptPrintResult.failure(
          message: 'Printer default belum disimpan untuk membuka laci kas.',
        );
      }
      final connected = await _adapter.connect(printer);
      if (!connected) {
        return ReceiptPrintResult.failure(
          message: 'Printer ${printer.name} tidak tersambung untuk buka laci.',
        );
      }
      final profile = await CapabilityProfile.load();
      final generator = Generator(PaperSize.mm58, profile);
      final ok = await _adapter.writeBytes(generator.drawer());
      return ok
          ? const ReceiptPrintResult.success(
              message: 'Perintah buka laci kas berhasil dikirim.',
            )
          : const ReceiptPrintResult.failure(
              message: 'Printer tidak merespons perintah buka laci kas.',
            );
    } catch (error) {
      return ReceiptPrintResult.failure(
        message: 'Buka laci kas gagal: ${_messageFor(error)}.',
      );
    }
  }

  Future<List<int>> buildEscPosBytes(
    SalesTransaction transaction, {
    Outlet? outlet,
    bool kickCashDrawer = false,
  }) async {
    final config = outlet?.receiptConfig ?? const ReceiptConfig();
    final profile = await CapabilityProfile.load();
    final paperSize = config.paperWidth == ReceiptPaperWidth.mm80
        ? PaperSize.mm80
        : PaperSize.mm58;
    final generator = Generator(paperSize, profile);
    final text = buildDigitalReceipt(transaction, outlet: outlet);
    final bytes = <int>[
      ...generator.text(text),
      ...generator.feed(2),
      if (kickCashDrawer) ...generator.drawer(),
      ...generator.cut(),
    ];
    return bytes;
  }
}

class ReceiptPrintResult {
  const ReceiptPrintResult._({
    required this.success,
    required this.message,
    required this.blocksSale,
    required this.canReprint,
  });

  const ReceiptPrintResult.success({String message = 'Struk berhasil dicetak.'})
    : this._(
        success: true,
        message: message,
        blocksSale: false,
        canReprint: true,
      );

  const ReceiptPrintResult.failure({required String message})
    : this._(
        success: false,
        message: message,
        blocksSale: false,
        canReprint: true,
      );

  final bool success;
  final String message;
  final bool blocksSale;
  final bool canReprint;
}

String _messageFor(Object error) {
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}
