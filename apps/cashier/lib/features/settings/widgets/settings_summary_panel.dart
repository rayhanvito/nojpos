import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../repositories/settings_repository.dart';

class SettingsSummaryPanel extends StatelessWidget {
  const SettingsSummaryPanel({required this.settings, super.key});

  final SettingsAggregate settings;

  @override
  Widget build(BuildContext context) {
    final outlet = settings.outlets.isNotEmpty ? settings.outlets.first : null;
    final activeMethods = outlet == null
        ? settings.paymentMethods.where((method) => method.active).toList()
        : settings.activePaymentMethodsForOutlet(outlet.id);
    final canUpdate = settings.permissions.canUpdateSettings;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              const Icon(LucideIcons.shieldCheck, color: MokposColors.primary),
              const SizedBox(width: 10),
              const Expanded(
                child: Text(
                  'Pengaturan Server',
                  style: TextStyle(
                    color: MokposColors.text,
                    fontWeight: FontWeight.w900,
                    fontSize: 16,
                  ),
                ),
              ),
              _ModePill(
                label: canUpdate ? 'Owner/Admin' : 'Kasir: hanya lihat',
                active: canUpdate,
              ),
            ],
          ),
          const SizedBox(height: 10),
          Text(
            canUpdate
                ? 'Konfigurasi operasional dimuat dari backend. Perubahan lanjutan tetap dibatasi owner/admin.'
                : 'Mode kasir: hanya lihat. Perubahan pengaturan hanya untuk owner/admin.',
            style: const TextStyle(
              color: MokposColors.muted,
              fontWeight: FontWeight.w700,
              height: 1.35,
            ),
          ),
          const SizedBox(height: 14),
          _SummaryRow(label: 'Bisnis', value: settings.business.name),
          _SummaryRow(
            label: 'Outlet aktif',
            value: outlet == null
                ? 'Belum ada outlet'
                : '${outlet.name} · ${outlet.timezone}',
          ),
          if (outlet != null) ...[
            _SummaryRow(
              label: 'Pajak & service',
              value:
                  'Pajak ${outlet.taxRate}% · Service ${outlet.serviceChargeRate}% · Rounding ${outlet.roundingPolicy}',
            ),
            _SummaryRow(
              label: 'Struk',
              value:
                  '${outlet.receiptPaperWidth}${outlet.receiptHeader.isEmpty ? '' : ' · ${outlet.receiptHeader}'}',
            ),
          ],
          _SummaryRow(
            label: 'Metode bayar aktif',
            value: activeMethods.isEmpty
                ? 'Belum ada metode aktif'
                : activeMethods.map((method) => method.method).join(', '),
          ),
          _SummaryRow(
            label: 'Kebijakan PIN',
            value:
                '${settings.security.pinPolicy.maxAttempts} percobaan · lock ${settings.security.pinPolicy.lockoutMinutes} menit',
          ),
          _SummaryRow(
            label: 'Auto-lock terminal',
            value:
                '${settings.security.terminalPolicy.idleLockTimeoutSeconds} detik',
          ),
          _SummaryRow(
            label: 'Aksi sensitif',
            value:
                'Void ${_pinText(settings.security.requiresPin('void'))} · Refund ${_pinText(settings.security.requiresPin('refund'))} · Tutup shift ${_pinText(settings.security.requiresPin('close_shift'))}',
          ),
        ],
      ),
    );
  }

  String _pinText(bool required) => required ? 'perlu PIN' : 'tanpa PIN';
}

class _ModePill extends StatelessWidget {
  const _ModePill({required this.label, required this.active});

  final String label;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: active ? MokposColors.primarySoft : MokposColors.disabledSurface,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(
          color: active ? MokposColors.primary : MokposColors.line,
        ),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: active ? MokposColors.primaryDark : MokposColors.muted,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _SummaryRow extends StatelessWidget {
  const _SummaryRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 5),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 150,
            child: Text(
              label,
              style: const TextStyle(
                color: MokposColors.muted,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
