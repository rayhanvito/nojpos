import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../../shared/models/nojpos_models.dart';
import '../../attendance/providers/attendance_controller.dart';

class PinScreen extends ConsumerStatefulWidget {
  const PinScreen({super.key});

  @override
  ConsumerState<PinScreen> createState() => _PinScreenState();
}

class _PinScreenState extends ConsumerState<PinScreen> {
  String pin = '';
  String? selectedEmployeeId;

  Future<void> _tap(String value) async {
    if (pin.length >= 4) return;
    setState(() => pin += value);
    if (pin.length == 4) {
      final ok = await ref.read(nojposSessionProvider.notifier).pinSwitch(pin);
      if (!mounted) return;
      if (ok) {
        context.go('/shift');
      } else {
        setState(() => pin = '');
      }
    }
  }

  void _delete() {
    if (pin.isEmpty) return;
    setState(() => pin = pin.substring(0, pin.length - 1));
  }

  void _selectEmployee(Employee employee) {
    setState(() {
      selectedEmployeeId = employee.id;
      pin = '';
    });
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(nojposSessionProvider);
    final staff = ref.watch(attendanceControllerProvider).staff;
    final selectedEmployee = _selectedEmployee(staff, selectedEmployeeId);
    final showStaffPicker = staff.isNotEmpty;

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(vertical: 20),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 720),
              child: Padding(
                padding: const EdgeInsets.all(28),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      width: 72,
                      height: 72,
                      decoration: BoxDecoration(
                        color: MokposColors.primarySoft,
                        borderRadius: BorderRadius.circular(22),
                      ),
                      child: const Icon(
                        LucideIcons.usersRound,
                        color: MokposColors.primary,
                        size: 34,
                      ),
                    ),
                    const SizedBox(height: 24),
                    Text(
                      'Login Karyawan',
                      style: Theme.of(context).textTheme.headlineSmall
                          ?.copyWith(
                            fontWeight: FontWeight.w900,
                            color: MokposColors.text,
                          ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      '${session.outlet.name} · ${session.businessName}',
                      textAlign: TextAlign.center,
                      style: const TextStyle(color: MokposColors.muted),
                    ),
                    const SizedBox(height: 24),
                    if (showStaffPicker) ...[
                      Align(
                        alignment: Alignment.centerLeft,
                        child: Text(
                          'Pilih karyawan',
                          style: Theme.of(context).textTheme.titleMedium
                              ?.copyWith(
                                color: MokposColors.text,
                                fontWeight: FontWeight.w900,
                              ),
                        ),
                      ),
                      const SizedBox(height: 10),
                      _StaffPicker(
                        staff: staff,
                        selectedEmployeeId:
                            selectedEmployee?.id ?? selectedEmployeeId,
                        onSelect: _selectEmployee,
                      ),
                      const SizedBox(height: 24),
                    ] else ...[
                      const _StaffFallbackNotice(),
                      const SizedBox(height: 24),
                    ],
                    Text(
                      selectedEmployee == null
                          ? 'Masukkan PIN karyawan'
                          : 'Masukkan PIN ${selectedEmployee.name}',
                      style: const TextStyle(
                        color: MokposColors.text,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    const SizedBox(height: 18),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: List.generate(
                        4,
                        (index) => AnimatedContainer(
                          duration: const Duration(milliseconds: 160),
                          width: 18,
                          height: 18,
                          margin: const EdgeInsets.symmetric(horizontal: 9),
                          decoration: BoxDecoration(
                            color: index < pin.length
                                ? MokposColors.primary
                                : Colors.white,
                            shape: BoxShape.circle,
                            border: Border.all(
                              color: index < pin.length
                                  ? MokposColors.primary
                                  : MokposColors.line,
                              width: 2,
                            ),
                          ),
                        ),
                      ),
                    ),
                    if (session.errorMessage != null) ...[
                      const SizedBox(height: 16),
                      Text(
                        session.errorMessage!,
                        textAlign: TextAlign.center,
                        style: const TextStyle(color: MokposColors.danger),
                      ),
                    ],
                    const SizedBox(height: 32),
                    if (session.isBusy)
                      const CircularProgressIndicator()
                    else
                      ConstrainedBox(
                        constraints: const BoxConstraints(maxWidth: 420),
                        child: _PinPad(onTap: _tap, onDelete: _delete),
                      ),
                    const SizedBox(height: 20),
                    TextButton.icon(
                      onPressed: session.outlets.length <= 1
                          ? null
                          : () => context.go('/outlet?change=1'),
                      icon: const Icon(LucideIcons.store, size: 18),
                      label: const Text('Ganti outlet'),
                    ),
                    TextButton.icon(
                      onPressed: () async {
                        await ref.read(nojposSessionProvider.notifier).logout();
                        if (!context.mounted) return;
                        context.go('/login');
                      },
                      icon: const Icon(LucideIcons.logOut, size: 18),
                      label: const Text('Logout akun/outlet'),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

Employee? _selectedEmployee(List<Employee> staff, String? selectedEmployeeId) {
  if (staff.isEmpty) return null;
  if (selectedEmployeeId == null) return staff.first;
  for (final employee in staff) {
    if (employee.id == selectedEmployeeId) return employee;
  }
  return staff.first;
}

class _StaffPicker extends StatelessWidget {
  const _StaffPicker({
    required this.staff,
    required this.selectedEmployeeId,
    required this.onSelect,
  });

  final List<Employee> staff;
  final String? selectedEmployeeId;
  final ValueChanged<Employee> onSelect;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 10,
      runSpacing: 10,
      children: [
        for (final employee in staff)
          _StaffChip(
            employee: employee,
            selected: employee.id == selectedEmployeeId ||
                (selectedEmployeeId == null && employee.id == staff.first.id),
            onTap: () => onSelect(employee),
          ),
      ],
    );
  }
}

class _StaffChip extends StatelessWidget {
  const _StaffChip({
    required this.employee,
    required this.selected,
    required this.onTap,
  });

  final Employee employee;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      borderRadius: BorderRadius.circular(MokposRadius.md),
      onTap: onTap,
      child: Container(
        width: 214,
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: selected ? MokposColors.primarySoft : Colors.white,
          border: Border.all(
            color: selected ? MokposColors.primary : MokposColors.line,
          ),
          borderRadius: BorderRadius.circular(MokposRadius.md),
        ),
        child: Row(
          children: [
            CircleAvatar(
              backgroundColor: selected ? MokposColors.primary : MokposColors.line,
              foregroundColor: selected ? Colors.white : MokposColors.text,
              child: Text(employee.name.characters.first.toUpperCase()),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    employee.name,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: MokposColors.text,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  Text(
                    employee.role,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: MokposColors.muted,
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ),
            if (selected)
              const Icon(
                LucideIcons.circleCheck,
                color: MokposColors.primary,
                size: 18,
              ),
          ],
        ),
      ),
    );
  }
}

class _StaffFallbackNotice extends StatelessWidget {
  const _StaffFallbackNotice();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: MokposColors.primarySoft,
        borderRadius: BorderRadius.circular(MokposRadius.md),
      ),
      child: const Text(
        'Data staff belum tersedia. Masukkan PIN untuk masuk sebagai kasir yang terdaftar di perangkat ini.',
        textAlign: TextAlign.center,
        style: TextStyle(
          color: MokposColors.text,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

class _PinPad extends StatelessWidget {
  const _PinPad({required this.onTap, required this.onDelete});

  final ValueChanged<String> onTap;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    final keys = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '', '0', 'del'];
    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: keys.length,
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 3,
        mainAxisSpacing: 14,
        crossAxisSpacing: 14,
        childAspectRatio: 1.45,
      ),
      itemBuilder: (context, index) {
        final key = keys[index];
        if (key.isEmpty) return const SizedBox.shrink();
        return FilledButton.tonal(
          key: ValueKey(key == 'del' ? 'pin_delete' : 'pin_$key'),
          onPressed: key == 'del' ? onDelete : () => onTap(key),
          style: FilledButton.styleFrom(
            backgroundColor: Colors.white,
            foregroundColor: MokposColors.text,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(MokposRadius.md),
              side: const BorderSide(color: MokposColors.line),
            ),
            textStyle: const TextStyle(
              fontSize: 24,
              fontWeight: FontWeight.w900,
            ),
          ),
          child: key == 'del'
              ? const Icon(LucideIcons.delete, size: 24)
              : Text(key),
        );
      },
    );
  }
}
