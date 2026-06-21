import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'package:uuid/uuid.dart';

import '../../../app/theme.dart';
import '../../../core/network/api_client.dart';
import '../providers/settings_providers.dart';
import '../repositories/settings_repository.dart';

class PaymentMethodSettingsPanel extends ConsumerWidget {
  const PaymentMethodSettingsPanel({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(settingsAggregateProvider);
    return settings.when(
      data: (value) => _PaymentMethodList(settings: value),
      loading: () =>
          const _SettingsLoadingCard(message: 'Memuat metode pembayaran...'),
      error: (error, stackTrace) => _SettingsErrorCard(
        title: 'Gagal memuat metode pembayaran',
        message: _messageForUi(error),
        onRetry: () => ref.invalidate(settingsAggregateProvider),
      ),
    );
  }
}

class _PaymentMethodList extends ConsumerStatefulWidget {
  const _PaymentMethodList({required this.settings});

  final SettingsAggregate settings;

  @override
  ConsumerState<_PaymentMethodList> createState() => _PaymentMethodListState();
}

class _PaymentMethodListState extends ConsumerState<_PaymentMethodList> {
  final Set<String> _saving = <String>{};
  String? _status;

  Future<void> _toggle(PaymentMethodSetting method, bool active) async {
    setState(() {
      _saving.add(method.id);
      _status = null;
    });
    try {
      await ref
          .read(settingsRepositoryProvider)
          .updatePaymentMethod(
            configId: method.id,
            active: active,
            idempotencyKey: const Uuid().v4(),
          );
      ref.invalidate(settingsAggregateProvider);
      if (!mounted) return;
      setState(() => _status = 'Metode ${method.method} disimpan dari server.');
    } catch (error) {
      if (!mounted) return;
      setState(() => _status = 'Gagal menyimpan: ${_messageForUi(error)}');
    } finally {
      if (mounted) {
        setState(() => _saving.remove(method.id));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final canUpdate = widget.settings.permissions.canUpdateSettings;
    final methods = widget.settings.paymentMethods;
    return _SettingsCard(
      icon: LucideIcons.creditCard,
      title: 'Metode Pembayaran',
      subtitle: canUpdate
          ? 'Tersambung ke API /settings/payment-methods. Aktif/nonaktif tersimpan ke backend.'
          : 'Mode kasir: daftar metode pembayaran dari backend, perubahan hanya owner/admin.',
      children: [
        if (methods.isEmpty)
          const _RoadmapNote(
            title: 'Belum ada metode pembayaran',
            message:
                'Tambahkan konfigurasi metode pembayaran dari seed/backend admin bisnis.',
          )
        else
          for (final method in methods) ...[
            SwitchListTile(
              contentPadding: EdgeInsets.zero,
              value: method.active,
              onChanged: !canUpdate || _saving.contains(method.id)
                  ? null
                  : (value) => _toggle(method, value),
              title: Text(
                method.method,
                style: const TextStyle(fontWeight: FontWeight.w900),
              ),
              subtitle: Text(
                '${method.isCash ? 'Tunai' : 'Nontunai'} · ${method.outletId == null ? 'Semua outlet' : 'Outlet khusus'}',
              ),
              secondary: _saving.contains(method.id)
                  ? const SizedBox.square(
                      dimension: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Icon(
                      method.isCash
                          ? LucideIcons.banknote
                          : LucideIcons.walletCards,
                      color: method.active
                          ? MokposColors.primary
                          : MokposColors.muted,
                    ),
            ),
            const Divider(height: 1, color: MokposColors.line),
          ],
        if (_status != null) ...[
          const SizedBox(height: 12),
          _InlineStatus(message: _status!),
        ],
      ],
    );
  }
}

class TaxServiceSettingsPanel extends ConsumerWidget {
  const TaxServiceSettingsPanel({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(settingsAggregateProvider);
    return settings.when(
      data: (value) => _TaxServiceForm(settings: value),
      loading: () => const _SettingsLoadingCard(
        message: 'Memuat pajak dan service charge...',
      ),
      error: (error, stackTrace) => _SettingsErrorCard(
        title: 'Gagal memuat pajak',
        message: _messageForUi(error),
        onRetry: () => ref.invalidate(settingsAggregateProvider),
      ),
    );
  }
}

class _TaxServiceForm extends ConsumerStatefulWidget {
  const _TaxServiceForm({required this.settings});

  final SettingsAggregate settings;

  @override
  ConsumerState<_TaxServiceForm> createState() => _TaxServiceFormState();
}

class _TaxServiceFormState extends ConsumerState<_TaxServiceForm> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _taxController;
  late final TextEditingController _serviceController;
  bool _saving = false;
  String? _status;

  OutletSettings? get _outlet =>
      widget.settings.outlets.isEmpty ? null : widget.settings.outlets.first;

  @override
  void initState() {
    super.initState();
    final outlet = _outlet;
    _taxController = TextEditingController(text: '${outlet?.taxRate ?? 0}');
    _serviceController = TextEditingController(
      text: '${outlet?.serviceChargeRate ?? 0}',
    );
  }

  @override
  void dispose() {
    _taxController.dispose();
    _serviceController.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final outlet = _outlet;
    if (outlet == null || !_formKey.currentState!.validate()) return;
    setState(() {
      _saving = true;
      _status = null;
    });
    try {
      await ref
          .read(settingsRepositoryProvider)
          .updateOutletSettings(
            outletId: outlet.id,
            taxRate: int.parse(_taxController.text),
            serviceChargeRate: int.parse(_serviceController.text),
            idempotencyKey: const Uuid().v4(),
          );
      ref.invalidate(settingsAggregateProvider);
      if (!mounted) return;
      setState(() => _status = 'Pajak dan service charge disimpan di server.');
    } catch (error) {
      if (!mounted) return;
      setState(() => _status = 'Gagal menyimpan: ${_messageForUi(error)}');
    } finally {
      if (mounted) {
        setState(() => _saving = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final outlet = _outlet;
    final canUpdate = widget.settings.permissions.canUpdateSettings;
    return _SettingsCard(
      icon: LucideIcons.percent,
      title: 'Pajak & Service Charge',
      subtitle: canUpdate
          ? 'Tersambung ke API /settings/outlets/{outlet}. Nilai ini dipakai backend saat membuat quote checkout.'
          : 'Mode kasir: nilai pajak/service hanya tampil baca dari backend.',
      children: [
        if (outlet == null)
          const _RoadmapNote(
            title: 'Outlet belum tersedia',
            message: 'Konfigurasi pajak membutuhkan outlet backend.',
          )
        else
          Form(
            key: _formKey,
            child: Column(
              children: [
                _ReadOnlyLine(
                  label: 'Outlet',
                  value: '${outlet.name} · ${outlet.timezone}',
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: TextFormField(
                        controller: _taxController,
                        enabled: canUpdate && !_saving,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                          labelText: 'Pajak (%)',
                          border: OutlineInputBorder(),
                        ),
                        validator: _rateValidator,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: TextFormField(
                        controller: _serviceController,
                        enabled: canUpdate && !_saving,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                          labelText: 'Service charge (%)',
                          border: OutlineInputBorder(),
                        ),
                        validator: _rateValidator,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                _ReadOnlyLine(label: 'Rounding', value: outlet.roundingPolicy),
                const SizedBox(height: 14),
                Align(
                  alignment: Alignment.centerRight,
                  child: FilledButton.icon(
                    onPressed: canUpdate && !_saving ? _save : null,
                    icon: _saving
                        ? const SizedBox.square(
                            dimension: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(LucideIcons.save, size: 16),
                    label: const Text('Simpan Pajak'),
                  ),
                ),
              ],
            ),
          ),
        if (_status != null) ...[
          const SizedBox(height: 12),
          _InlineStatus(message: _status!),
        ],
      ],
    );
  }
}

class RefundSettingsPanel extends ConsumerWidget {
  const RefundSettingsPanel({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(settingsAggregateProvider);
    return settings.when(
      data: (value) {
        final refund = value.security.sensitiveActions['refund'];
        return _SettingsCard(
          icon: LucideIcons.undo2,
          title: 'Kebijakan Refund',
          subtitle:
              'Refund sudah memakai kontrak backend /transactions/{transaction}/refund. Panel ini menampilkan policy dari /settings; write khusus refund policy belum dipisah dari security settings.',
          children: [
            _ReadOnlyLine(
              label: 'PIN refund',
              value: refund?.requiresPin ?? true
                  ? 'Wajib PIN owner/admin'
                  : 'Tidak wajib PIN',
            ),
            _ReadOnlyLine(
              label: 'Role yang boleh refund',
              value: refund?.roles.join(', ') ?? 'admin, owner',
            ),
            const SizedBox(height: 12),
            const _RoadmapNote(
              title: 'Read-only policy',
              message:
                  'Batas nominal, window hari refund, dan policy restock belum punya endpoint settings terpisah. Aksi refund tetap diproses server dari dialog transaksi.',
            ),
          ],
        );
      },
      loading: () =>
          const _SettingsLoadingCard(message: 'Memuat kebijakan refund...'),
      error: (error, stackTrace) => _SettingsErrorCard(
        title: 'Gagal memuat refund settings',
        message: _messageForUi(error),
        onRetry: () => ref.invalidate(settingsAggregateProvider),
      ),
    );
  }
}

class SecuritySettingsPanel extends ConsumerWidget {
  const SecuritySettingsPanel({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(settingsAggregateProvider);
    return settings.when(
      data: (value) => _SecurityForm(settings: value),
      loading: () =>
          const _SettingsLoadingCard(message: 'Memuat security settings...'),
      error: (error, stackTrace) => _SettingsErrorCard(
        title: 'Gagal memuat security settings',
        message: _messageForUi(error),
        onRetry: () => ref.invalidate(settingsAggregateProvider),
      ),
    );
  }
}

class _SecurityForm extends ConsumerStatefulWidget {
  const _SecurityForm({required this.settings});

  final SettingsAggregate settings;

  @override
  ConsumerState<_SecurityForm> createState() => _SecurityFormState();
}

class _SecurityFormState extends ConsumerState<_SecurityForm> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _maxAttemptsController;
  late final TextEditingController _lockoutController;
  late final TextEditingController _idleController;
  late final TextEditingController _sessionController;
  late Map<String, bool> _actions;
  bool _saving = false;
  String? _status;

  @override
  void initState() {
    super.initState();
    final security = widget.settings.security;
    _maxAttemptsController = TextEditingController(
      text: '${security.pinPolicy.maxAttempts}',
    );
    _lockoutController = TextEditingController(
      text: '${security.pinPolicy.lockoutMinutes}',
    );
    _idleController = TextEditingController(
      text: '${security.terminalPolicy.idleLockTimeoutSeconds}',
    );
    _sessionController = TextEditingController(
      text: '${security.terminalPolicy.sessionTimeoutSeconds}',
    );
    _actions = {
      for (final entry in security.sensitiveActions.entries)
        entry.key: entry.value.requiresPin,
    };
  }

  @override
  void dispose() {
    _maxAttemptsController.dispose();
    _lockoutController.dispose();
    _idleController.dispose();
    _sessionController.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() {
      _saving = true;
      _status = null;
    });
    try {
      await ref
          .read(settingsRepositoryProvider)
          .updateSecuritySettings(
            maxAttempts: int.parse(_maxAttemptsController.text),
            lockoutMinutes: int.parse(_lockoutController.text),
            idleLockTimeoutSeconds: int.parse(_idleController.text),
            sessionTimeoutSeconds: int.parse(_sessionController.text),
            sensitiveActionPins: _actions,
            idempotencyKey: const Uuid().v4(),
          );
      ref.invalidate(settingsAggregateProvider);
      if (!mounted) return;
      setState(() => _status = 'Security settings disimpan di server.');
    } catch (error) {
      if (!mounted) return;
      setState(() => _status = 'Gagal menyimpan: ${_messageForUi(error)}');
    } finally {
      if (mounted) {
        setState(() => _saving = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final canUpdate = widget.settings.permissions.canUpdateSecuritySettings;
    return _SettingsCard(
      icon: LucideIcons.shieldCheck,
      title: 'Security & PIN Policy',
      subtitle: canUpdate
          ? 'Tersambung ke API /settings/security. Tidak ada PIN/password yang ditampilkan atau disimpan lokal.'
          : 'Mode kasir: policy security hanya baca, tidak mengekspos data sensitif.',
      children: [
        Form(
          key: _formKey,
          child: Column(
            children: [
              Row(
                children: [
                  Expanded(
                    child: _NumberField(
                      controller: _maxAttemptsController,
                      enabled: canUpdate && !_saving,
                      label: 'Maks. percobaan PIN',
                      validator: (value) => _intRangeValidator(value, 1, 20),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _NumberField(
                      controller: _lockoutController,
                      enabled: canUpdate && !_saving,
                      label: 'Lockout menit',
                      validator: (value) => _intRangeValidator(value, 1, 1440),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: _NumberField(
                      controller: _idleController,
                      enabled: canUpdate && !_saving,
                      label: 'Idle lock detik',
                      validator: (value) =>
                          _intRangeValidator(value, 30, 86400),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _NumberField(
                      controller: _sessionController,
                      enabled: canUpdate && !_saving,
                      label: 'Session timeout detik',
                      validator: (value) =>
                          _intRangeValidator(value, 60, 86400),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 14),
              for (final entry in _actions.entries)
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  value: entry.value,
                  onChanged: canUpdate && !_saving
                      ? (value) => setState(() => _actions[entry.key] = value)
                      : null,
                  title: Text(
                    _actionLabel(entry.key),
                    style: const TextStyle(fontWeight: FontWeight.w900),
                  ),
                  subtitle: Text(
                    'Role: ${widget.settings.security.sensitiveActions[entry.key]?.roles.join(', ') ?? '-'}',
                  ),
                ),
              Align(
                alignment: Alignment.centerRight,
                child: FilledButton.icon(
                  onPressed: canUpdate && !_saving ? _save : null,
                  icon: _saving
                      ? const SizedBox.square(
                          dimension: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Icon(LucideIcons.save, size: 16),
                  label: const Text('Simpan Security'),
                ),
              ),
            ],
          ),
        ),
        if (_status != null) ...[
          const SizedBox(height: 12),
          _InlineStatus(message: _status!),
        ],
      ],
    );
  }
}

class _SettingsCard extends StatelessWidget {
  const _SettingsCard({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.children,
  });

  final IconData icon;
  final String title;
  final String subtitle;
  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(18),
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
              Icon(icon, color: MokposColors.primary),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  title,
                  style: const TextStyle(
                    color: MokposColors.text,
                    fontWeight: FontWeight.w900,
                    fontSize: 18,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            subtitle,
            style: const TextStyle(
              color: MokposColors.muted,
              fontWeight: FontWeight.w700,
              height: 1.35,
            ),
          ),
          const SizedBox(height: 16),
          ...children,
        ],
      ),
    );
  }
}

class _SettingsLoadingCard extends StatelessWidget {
  const _SettingsLoadingCard({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return _SettingsCard(
      icon: LucideIcons.loader,
      title: 'Memuat Settings',
      subtitle: message,
      children: const [LinearProgressIndicator()],
    );
  }
}

class _SettingsErrorCard extends StatelessWidget {
  const _SettingsErrorCard({
    required this.title,
    required this.message,
    required this.onRetry,
  });

  final String title;
  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return _SettingsCard(
      icon: LucideIcons.triangleAlert,
      title: title,
      subtitle: message,
      children: [
        Align(
          alignment: Alignment.centerLeft,
          child: OutlinedButton.icon(
            onPressed: onRetry,
            icon: const Icon(LucideIcons.refreshCcw, size: 16),
            label: const Text('Coba lagi'),
          ),
        ),
      ],
    );
  }
}

class _RoadmapNote extends StatelessWidget {
  const _RoadmapNote({required this.title, required this.message});

  final String title;
  final String message;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: MokposColors.disabledSurface,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(LucideIcons.info, color: MokposColors.muted, size: 18),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(fontWeight: FontWeight.w900),
                ),
                const SizedBox(height: 4),
                Text(
                  message,
                  style: const TextStyle(
                    color: MokposColors.muted,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _InlineStatus extends StatelessWidget {
  const _InlineStatus({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    final isError = message.toLowerCase().startsWith('gagal');
    return Text(
      message,
      style: TextStyle(
        color: isError ? MokposColors.danger : MokposColors.success,
        fontWeight: FontWeight.w800,
      ),
    );
  }
}

class _ReadOnlyLine extends StatelessWidget {
  const _ReadOnlyLine({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 5),
      child: Row(
        children: [
          SizedBox(
            width: 180,
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
              style: const TextStyle(fontWeight: FontWeight.w900),
            ),
          ),
        ],
      ),
    );
  }
}

class _NumberField extends StatelessWidget {
  const _NumberField({
    required this.controller,
    required this.enabled,
    required this.label,
    required this.validator,
  });

  final TextEditingController controller;
  final bool enabled;
  final String label;
  final FormFieldValidator<String> validator;

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      enabled: enabled,
      keyboardType: TextInputType.number,
      decoration: InputDecoration(
        labelText: label,
        border: const OutlineInputBorder(),
      ),
      validator: validator,
    );
  }
}

String? _rateValidator(String? value) => _intRangeValidator(value, 0, 100);

String? _intRangeValidator(String? value, int min, int max) {
  final parsed = int.tryParse((value ?? '').trim());
  if (parsed == null) return 'Harus angka';
  if (parsed < min || parsed > max) return 'Rentang $min-$max';
  return null;
}

String _actionLabel(String action) {
  return switch (action) {
    'void' => 'Void transaksi wajib PIN',
    'refund' => 'Refund wajib PIN',
    'discount_override' => 'Override diskon wajib PIN',
    'cash_out_over_limit' => 'Cash out besar wajib PIN',
    'close_shift' => 'Tutup shift wajib PIN',
    'store_open_close' => 'Buka/tutup toko wajib PIN',
    'settings_change' => 'Ubah settings wajib PIN',
    _ => '$action wajib PIN',
  };
}

String _messageForUi(Object error) {
  if (error is ApiException) return error.message;
  return error.toString();
}
