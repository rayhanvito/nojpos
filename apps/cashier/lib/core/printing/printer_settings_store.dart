import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'receipt_printing_service.dart';

final printerSettingsStoreProvider = Provider<PrinterSettingsStore>(
  (ref) => const SecurePrinterSettingsStore(),
);

class PrinterSettings {
  const PrinterSettings({
    this.defaultPrinter,
    this.cashDrawerEnabled = false,
  });

  final BluetoothPrinterDevice? defaultPrinter;
  final bool cashDrawerEnabled;

  PrinterSettings copyWith({
    BluetoothPrinterDevice? defaultPrinter,
    bool clearDefaultPrinter = false,
    bool? cashDrawerEnabled,
  }) {
    return PrinterSettings(
      defaultPrinter: clearDefaultPrinter
          ? null
          : defaultPrinter ?? this.defaultPrinter,
      cashDrawerEnabled: cashDrawerEnabled ?? this.cashDrawerEnabled,
    );
  }
}

abstract interface class PrinterSettingsStore {
  Future<PrinterSettings> readSettings();

  Future<void> saveDefaultPrinter(BluetoothPrinterDevice printer);

  Future<void> clearDefaultPrinter();

  Future<void> setCashDrawerEnabled(bool enabled);
}

class SecurePrinterSettingsStore implements PrinterSettingsStore {
  const SecurePrinterSettingsStore({FlutterSecureStorage? secureStorage})
    : _secureStorage = secureStorage ?? const FlutterSecureStorage();

  static const _printerNameKey = 'nojpos.printer.default_name';
  static const _printerAddressKey = 'nojpos.printer.default_address';
  static const _cashDrawerEnabledKey = 'nojpos.printer.cash_drawer_enabled';

  final FlutterSecureStorage _secureStorage;

  @override
  Future<PrinterSettings> readSettings() async {
    final name = await _secureStorage.read(key: _printerNameKey);
    final address = await _secureStorage.read(key: _printerAddressKey);
    final cashDrawer = await _secureStorage.read(key: _cashDrawerEnabledKey);
    return PrinterSettings(
      defaultPrinter: address == null || address.isEmpty
          ? null
          : BluetoothPrinterDevice(
              name: name == null || name.isEmpty ? 'Printer Bluetooth' : name,
              address: address,
            ),
      cashDrawerEnabled: cashDrawer == 'true',
    );
  }

  @override
  Future<void> saveDefaultPrinter(BluetoothPrinterDevice printer) async {
    await _secureStorage.write(key: _printerNameKey, value: printer.name);
    await _secureStorage.write(key: _printerAddressKey, value: printer.address);
  }

  @override
  Future<void> clearDefaultPrinter() async {
    await _secureStorage.delete(key: _printerNameKey);
    await _secureStorage.delete(key: _printerAddressKey);
  }

  @override
  Future<void> setCashDrawerEnabled(bool enabled) {
    return _secureStorage.write(
      key: _cashDrawerEnabledKey,
      value: enabled ? 'true' : 'false',
    );
  }
}

class InMemoryPrinterSettingsStore implements PrinterSettingsStore {
  PrinterSettings _settings;

  InMemoryPrinterSettingsStore([this._settings = const PrinterSettings()]);

  @override
  Future<PrinterSettings> readSettings() async => _settings;

  @override
  Future<void> saveDefaultPrinter(BluetoothPrinterDevice printer) async {
    _settings = _settings.copyWith(defaultPrinter: printer);
  }

  @override
  Future<void> clearDefaultPrinter() async {
    _settings = _settings.copyWith(clearDefaultPrinter: true);
  }

  @override
  Future<void> setCashDrawerEnabled(bool enabled) async {
    _settings = _settings.copyWith(cashDrawerEnabled: enabled);
  }
}
