import 'package:flutter/material.dart';
import 'package:skeletonizer/skeletonizer.dart';

import '../../app/theme.dart';

class NojposSkeleton extends StatelessWidget {
  const NojposSkeleton({super.key, required this.enabled, required this.child});

  final bool enabled;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Skeletonizer(
      enabled: enabled,
      effect: const ShimmerEffect(
        baseColor: Color(0xFFE9EEF1),
        highlightColor: Color(0xFFF7FAFB),
        duration: Duration(milliseconds: 1100),
      ),
      child: child,
    );
  }
}

class NojposSkeletonCard extends StatelessWidget {
  const NojposSkeletonCard({super.key, this.height = 120, this.radius = 16});

  final double height;
  final double radius;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: height,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(radius),
        border: Border.all(color: NojposColors.line),
      ),
    );
  }
}
