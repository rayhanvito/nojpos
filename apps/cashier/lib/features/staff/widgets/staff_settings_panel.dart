import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../../shared/models/nojpos_models.dart';
import '../../attendance/providers/attendance_controller.dart';
import '../repositories/staff_repository.dart';

class StaffSettingsPanel extends ConsumerStatefulWidget {
  const StaffSettingsPanel({super.key});

  @override
  ConsumerState<StaffSettingsPanel> createState() => _StaffSettingsPanelState();
}

class _StaffSettingsPanelState extends ConsumerState<StaffSettingsPanel> {
  List<Employee> staff = const [];
  bool loading = true;
  String? message;

  @override
  void initState() {
    super.initState();
    Future<void>.microtask(_load);
  }

  Future<void> _load() async {
    setState(() {
      loading = true;
      message = null;
    });
    try {
      final outletId = ref.read(nojposSessionProvider).outlet.id;
      final result = await ref.read(staffRepositoryProvider).listStaff(outletId: outletId);
      if (!mounted) return;
      setState(() {
        staff = result;
        loading = false;
      });
      await ref.read(attendanceControllerProvider.notifier).load();
    } catch (error) {
      if (!mounted) return;
      setState(() {
        loading = false;
        message = _message(error);
      });
    }
  }

  Future<void> _openForm({Employee? employee}) async {
    final saved = await showDialog<bool>(
      context: context,
      builder: (context) => _StaffFormDialog(employee: employee),
    );
    if (saved == true) await _load();
  }

  Future<void> _remove(Employee employee) async {
    final confirmed = await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: const Text('Hapus Staff?'),
            content: Text('${employee.name} akan dinonaktifkan dari bisnis ini.'),
            actions: [
              TextButton(onPressed: () => Navigator.of(context).pop(false), child: const Text('Batal')),
              FilledButton(onPressed: () => Navigator.of(context).pop(true), child: const Text('Hapus')),
            ],
          ),
        ) ??
        false;
    if (!confirmed) return;
    try {
      await ref.read(staffRepositoryProvider).deleteStaff(employee.id);
      await _load();
    } catch (error) {
      if (!mounted) return;
      setState(() => message = _message(error));
    }
  }

  @override
  Widget build(BuildContext context) {
    final canManage = ref.watch(nojposSessionProvider).canManageMasterData;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(MokposRadius.md),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(LucideIcons.usersRound, color: MokposColors.primary),
              const SizedBox(width: 10),
              const Expanded(child: Text('Kasir & Staff', style: TextStyle(fontWeight: FontWeight.w900, fontSize: 16))),
              OutlinedButton.icon(onPressed: loading ? null : _load, icon: const Icon(LucideIcons.refreshCw, size: 16), label: const Text('Refresh')),
              const SizedBox(width: 8),
              FilledButton.icon(onPressed: canManage ? () => _openForm() : null, icon: const Icon(LucideIcons.plus, size: 16), label: const Text('Tambah Staff')),
            ],
          ),
          const SizedBox(height: 8),
          const Text('Staff CRUD tersambung ke /api/v1/staff. PIN dipakai untuk login karyawan.', style: TextStyle(color: MokposColors.muted)),
          if (!canManage) const Padding(padding: EdgeInsets.only(top: 8), child: Text('Hanya owner/admin yang boleh mengelola staff.', style: TextStyle(color: MokposColors.danger))),
          if (message != null) Padding(padding: const EdgeInsets.only(top: 8), child: Text(message!, style: const TextStyle(color: MokposColors.danger))),
          const SizedBox(height: 12),
          if (loading)
            const Center(child: Padding(padding: EdgeInsets.all(20), child: CircularProgressIndicator()))
          else if (staff.isEmpty)
            const Text('Belum ada staff.', style: TextStyle(color: MokposColors.muted))
          else
            ListView.separated(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: staff.length,
              separatorBuilder: (context, index) => const Divider(height: 1, color: MokposColors.line),
              itemBuilder: (context, index) {
                final employee = staff[index];
                return ListTile(
                  contentPadding: EdgeInsets.zero,
                  leading: CircleAvatar(child: Text(employee.name.characters.first.toUpperCase())),
                  title: Text(employee.name, style: const TextStyle(fontWeight: FontWeight.w900)),
                  subtitle: Text('${employee.role}${employee.email.isEmpty ? '' : ' · ${employee.email}'}'),
                  trailing: Wrap(
                    spacing: 8,
                    children: [
                      OutlinedButton(onPressed: canManage ? () => _openForm(employee: employee) : null, child: const Text('Edit')),
                      OutlinedButton(onPressed: canManage && employee.role != 'owner' ? () => _remove(employee) : null, child: const Text('Hapus')),
                    ],
                  ),
                );
              },
            ),
        ],
      ),
    );
  }
}

class _StaffFormDialog extends ConsumerStatefulWidget {
  const _StaffFormDialog({this.employee});

  final Employee? employee;

  @override
  ConsumerState<_StaffFormDialog> createState() => _StaffFormDialogState();
}

class _StaffFormDialogState extends ConsumerState<_StaffFormDialog> {
  late final TextEditingController nameController;
  late final TextEditingController emailController;
  late final TextEditingController pinController;
  late String role;
  String? error;

  @override
  void initState() {
    super.initState();
    nameController = TextEditingController(text: widget.employee?.name ?? '');
    emailController = TextEditingController(text: widget.employee?.email ?? '');
    pinController = TextEditingController();
    role = widget.employee?.role == 'admin' ? 'admin' : 'cashier';
  }

  @override
  void dispose() {
    nameController.dispose();
    emailController.dispose();
    pinController.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final name = nameController.text.trim();
    final pin = pinController.text.trim();
    if (name.isEmpty || (widget.employee == null && pin.length < 4)) {
      setState(() => error = 'Nama wajib diisi dan PIN staff baru minimal 4 digit.');
      return;
    }
    try {
      final repo = ref.read(staffRepositoryProvider);
      if (widget.employee == null) {
        await repo.createStaff(name: name, role: role, pin: pin, email: emailController.text);
      } else {
        await repo.updateStaff(id: widget.employee!.id, name: name, role: role, pin: pin.isEmpty ? null : pin, email: emailController.text);
      }
      if (!mounted) return;
      Navigator.of(context).pop(true);
    } catch (error) {
      if (!mounted) return;
      setState(() => this.error = _message(error));
    }
  }

  @override
  Widget build(BuildContext context) {
    final editing = widget.employee != null;
    return AlertDialog(
      title: Text(editing ? 'Edit Staff' : 'Tambah Staff'),
      content: SizedBox(
        width: 420,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: nameController, decoration: const InputDecoration(labelText: 'Nama', border: OutlineInputBorder())),
            const SizedBox(height: 12),
            TextField(controller: emailController, keyboardType: TextInputType.emailAddress, decoration: const InputDecoration(labelText: 'Email opsional', border: OutlineInputBorder())),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              initialValue: role,
              items: const [DropdownMenuItem(value: 'cashier', child: Text('Cashier')), DropdownMenuItem(value: 'admin', child: Text('Admin'))],
              onChanged: widget.employee?.role == 'owner' ? null : (value) => setState(() => role = value ?? 'cashier'),
              decoration: const InputDecoration(labelText: 'Role', border: OutlineInputBorder()),
            ),
            const SizedBox(height: 12),
            TextField(controller: pinController, keyboardType: TextInputType.number, obscureText: true, decoration: InputDecoration(labelText: editing ? 'PIN baru opsional' : 'PIN', border: const OutlineInputBorder())),
            if (error != null) Padding(padding: const EdgeInsets.only(top: 12), child: Text(error!, style: const TextStyle(color: MokposColors.danger))),
          ],
        ),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.of(context).pop(false), child: const Text('Batal')),
        FilledButton(onPressed: _save, child: const Text('Simpan')),
      ],
    );
  }
}

String _message(Object error) {
  final message = error.toString();
  final separator = message.indexOf(': ');
  return separator == -1 ? message : message.substring(separator + 2);
}
