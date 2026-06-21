import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../../../app/theme.dart';
import '../../../shared/models/nojpos_models.dart';
import '../../attendance/providers/attendance_controller.dart';
import '../../pos/providers/pos_providers.dart';

class PinScreen extends ConsumerStatefulWidget {
  const PinScreen({super.key});

  @override
  ConsumerState<PinScreen> createState() => _PinScreenState();
}

class _PinScreenState extends ConsumerState<PinScreen> {
  String pin = '';
  String? selectedEmployeeId;
  bool _silentRefreshStarted = false;

  @override
  void initState() {
    super.initState();
    Future<void>.microtask(_silentRefreshTerminalData);
  }

  Future<void> _silentRefreshTerminalData() async {
    if (_silentRefreshStarted) return;
    _silentRefreshStarted = true;
    final session = ref.read(nojposSessionProvider);
    if (session.outlet.id.isEmpty) return;

    await Future.wait<void>([
      ref.read(posCatalogProvider.notifier).load().catchError((_) {}),
      ref.read(attendanceControllerProvider.notifier).load().catchError((_) {}),
      ref
          .read(nojposSessionProvider.notifier)
          .refreshCurrentShift()
          .catchError((_) {}),
    ]);
  }

  Future<void> _tap(String value) async {
    if (pin.length >= 4) return;
    setState(() => pin += value);
    if (pin.length == 4) {
      final ok = await ref.read(nojposSessionProvider.notifier).pinSwitch(pin);
      if (!mounted) return;
      if (ok) {
        context.go('/pos');
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
    final size = MediaQuery.sizeOf(context);
    final compact = size.height < 720 || size.width < 420;

    return Scaffold(
      backgroundColor: MokposColors.canvas,
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: EdgeInsets.symmetric(
              horizontal: compact ? 16 : 24,
              vertical: compact ? 10 : 18,
            ),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 430),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  _PinHeader(session: session, compact: compact),
                  SizedBox(height: compact ? 12 : 16),
                  if (showStaffPicker)
                    _StaffPicker(
                      staff: staff,
                      selectedEmployeeId:
                          selectedEmployee?.id ?? selectedEmployeeId,
                      selectedEmployee: selectedEmployee,
                      onSelect: _selectEmployee,
                    )
                  else
                    const _StaffFallbackNotice(),
                  SizedBox(height: compact ? 14 : 18),
                  Text(
                    selectedEmployee == null
                        ? 'Masukkan PIN karyawan'
                        : 'Masukkan PIN ${selectedEmployee.name}',
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      color: MokposColors.text,
                      fontWeight: FontWeight.w900,
                      fontSize: compact ? 13 : 14,
                    ),
                  ),
                  SizedBox(height: compact ? 10 : 14),
                  _PinDots(length: pin.length, compact: compact),
                  if (session.errorMessage != null) ...[
                    const SizedBox(height: 10),
                    Text(
                      session.errorMessage!,
                      textAlign: TextAlign.center,
                      style: const TextStyle(color: MokposColors.danger),
                    ),
                  ],
                  SizedBox(height: compact ? 16 : 22),
                  if (session.isBusy)
                    const Padding(
                      padding: EdgeInsets.symmetric(vertical: 36),
                      child: CircularProgressIndicator(),
                    )
                  else
                    ConstrainedBox(
                      constraints: BoxConstraints(
                        maxWidth: compact ? 330 : 360,
                      ),
                      child: _PinPad(
                        onTap: _tap,
                        onDelete: _delete,
                        compact: compact,
                      ),
                    ),
                  SizedBox(height: compact ? 10 : 14),
                  _PinFooterActions(
                    canChangeOutlet: session.outlets.length > 1,
                    onChangeOutlet: () => context.go('/outlet?change=1'),
                    onLogout: () async {
                      await ref.read(nojposSessionProvider.notifier).logout();
                      if (!context.mounted) return;
                      context.go('/login');
                    },
                  ),
                ],
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

class _PinHeader extends StatelessWidget {
  const _PinHeader({required this.session, required this.compact});

  final NojposSessionState session;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          width: compact ? 48 : 56,
          height: compact ? 48 : 56,
          decoration: BoxDecoration(
            color: MokposColors.primarySoft,
            borderRadius: BorderRadius.circular(compact ? 16 : 18),
          ),
          child: Icon(
            LucideIcons.usersRound,
            color: MokposColors.primary,
            size: compact ? 24 : 28,
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Login Karyawan',
                style: Theme.of(context).textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.w900,
                  color: MokposColors.text,
                  height: 1.05,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                '${session.outlet.name} · ${session.businessName}',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(color: MokposColors.muted, fontSize: 12),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _StaffPicker extends StatelessWidget {
  const _StaffPicker({
    required this.staff,
    required this.selectedEmployeeId,
    required this.selectedEmployee,
    required this.onSelect,
  });

  final List<Employee> staff;
  final String? selectedEmployeeId;
  final Employee? selectedEmployee;
  final ValueChanged<Employee> onSelect;

  @override
  Widget build(BuildContext context) {
    if (staff.length == 1) {
      return _SelectedStaffCard(employee: staff.single);
    }

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(MokposRadius.md),
        boxShadow: NojposShadow.card,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          DropdownButtonFormField<String>(
            initialValue: selectedEmployeeId ?? staff.first.id,
            isExpanded: true,
            decoration: InputDecoration(
              labelText: 'Pilih karyawan',
              prefixIcon: const Icon(LucideIcons.userRound, size: 18),
              filled: true,
              fillColor: MokposColors.canvas,
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 12,
                vertical: 10,
              ),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(MokposRadius.md),
              ),
            ),
            items: [
              for (final employee in staff)
                DropdownMenuItem<String>(
                  value: employee.id,
                  child: Text(
                    employee.name,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(fontWeight: FontWeight.w900),
                  ),
                ),
            ],
            onChanged: (value) {
              if (value == null) return;
              final employee = staff.firstWhere((item) => item.id == value);
              onSelect(employee);
            },
          ),
          if (selectedEmployee != null) ...[
            const SizedBox(height: 8),
            Text(
              selectedEmployee!.role,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: MokposColors.muted,
                fontSize: 12,
                fontWeight: FontWeight.w700,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _SelectedStaffCard extends StatelessWidget {
  const _SelectedStaffCard({required this.employee});

  final Employee employee;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: MokposColors.line),
        borderRadius: BorderRadius.circular(MokposRadius.md),
        boxShadow: NojposShadow.card,
      ),
      child: Row(
        children: [
          CircleAvatar(
            radius: 18,
            backgroundColor: MokposColors.primarySoft,
            foregroundColor: MokposColors.primary,
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
          const Icon(
            LucideIcons.circleCheck,
            color: MokposColors.primary,
            size: 18,
          ),
        ],
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
        'Data staff sedang dimuat. Masukkan PIN kasir yang terdaftar di perangkat ini.',
        textAlign: TextAlign.center,
        style: TextStyle(color: MokposColors.text, fontWeight: FontWeight.w700),
      ),
    );
  }
}

class _PinDots extends StatelessWidget {
  const _PinDots({required this.length, required this.compact});

  final int length;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: List.generate(
        4,
        (index) => AnimatedContainer(
          duration: const Duration(milliseconds: 160),
          width: compact ? 14 : 16,
          height: compact ? 14 : 16,
          margin: const EdgeInsets.symmetric(horizontal: 7),
          decoration: BoxDecoration(
            color: index < length ? MokposColors.primary : Colors.white,
            shape: BoxShape.circle,
            border: Border.all(
              color: index < length ? MokposColors.primary : MokposColors.line,
              width: 2,
            ),
          ),
        ),
      ),
    );
  }
}

class _PinPad extends StatelessWidget {
  const _PinPad({
    required this.onTap,
    required this.onDelete,
    required this.compact,
  });

  final ValueChanged<String> onTap;
  final VoidCallback onDelete;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final keys = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '', '0', 'del'];
    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: keys.length,
      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 3,
        mainAxisSpacing: compact ? 9 : 11,
        crossAxisSpacing: compact ? 9 : 11,
        childAspectRatio: compact ? 1.75 : 1.6,
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
            textStyle: TextStyle(
              fontSize: compact ? 20 : 22,
              fontWeight: FontWeight.w900,
            ),
          ),
          child: key == 'del'
              ? Icon(LucideIcons.delete, size: compact ? 20 : 22)
              : Text(key),
        );
      },
    );
  }
}

class _PinFooterActions extends StatelessWidget {
  const _PinFooterActions({
    required this.canChangeOutlet,
    required this.onChangeOutlet,
    required this.onLogout,
  });

  final bool canChangeOutlet;
  final VoidCallback onChangeOutlet;
  final Future<void> Function() onLogout;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      alignment: WrapAlignment.center,
      spacing: 8,
      runSpacing: 4,
      children: [
        TextButton.icon(
          onPressed: canChangeOutlet ? onChangeOutlet : null,
          icon: const Icon(LucideIcons.store, size: 16),
          label: const Text('Ganti outlet'),
          style: TextButton.styleFrom(
            visualDensity: VisualDensity.compact,
            textStyle: const TextStyle(fontWeight: FontWeight.w800),
          ),
        ),
        TextButton.icon(
          onPressed: onLogout,
          icon: const Icon(LucideIcons.logOut, size: 16),
          label: const Text('Keluar akun'),
          style: TextButton.styleFrom(
            visualDensity: VisualDensity.compact,
            foregroundColor: MokposColors.muted,
            textStyle: const TextStyle(fontWeight: FontWeight.w800),
          ),
        ),
      ],
    );
  }
}
