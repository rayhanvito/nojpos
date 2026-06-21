import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/theme.dart';
import '../providers/connectivity_provider.dart';

class ConnectivityStatusChip extends ConsumerWidget {
  const ConnectivityStatusChip({
    this.prefix = 'Koneksi',
    this.suffix,
    this.compact = false,
    this.textStyle,
    this.dotSize = 8,
    super.key,
  });

  final String prefix;
  final String? suffix;
  final bool compact;
  final TextStyle? textStyle;
  final double dotSize;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(connectivityControllerProvider);
    final label = compact ? state.label : '$prefix: ${state.label}';
    final message = suffix == null ? label : '$label · $suffix';

    return Tooltip(
      message: state.friendlyMessage,
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          _ConnectivityDot(status: state.status, size: dotSize),
          const SizedBox(width: 6),
          Flexible(
            child: Text(
              message,
              overflow: TextOverflow.ellipsis,
              style:
                  textStyle ??
                  const TextStyle(color: MokposColors.onPrimaryMuted),
            ),
          ),
        ],
      ),
    );
  }
}

class _ConnectivityDot extends StatelessWidget {
  const _ConnectivityDot({required this.status, required this.size});

  final ConnectivityStatus status;
  final double size;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: _colorFor(status),
        shape: BoxShape.circle,
      ),
    );
  }

  Color _colorFor(ConnectivityStatus status) {
    return switch (status) {
      ConnectivityStatus.checking => MokposColors.muted,
      ConnectivityStatus.online => MokposColors.success,
      ConnectivityStatus.degraded => MokposColors.warning,
      ConnectivityStatus.offline => MokposColors.danger,
    };
  }
}
