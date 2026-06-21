import 'package:flutter/material.dart';

/// Compatibility color tokens kept under the historical Mokpos name.
///
/// New UI should prefer [NojposColors]. Existing screens still reference
/// [MokposColors], so these values intentionally mirror the current NOJPOS
/// blue-first brand palette.
class MokposColors {
  const MokposColors._();

  static const primary = Color(0xFF0060F0);
  static const primaryDark = Color(0xFF0047B8);
  static const primarySoft = Color(0xFFEAF2FF);
  static const accent = Color(0xFF009E82);
  static const accentDark = Color(0xFF007C66);
  static const accentSoft = Color(0xFFE6F8F4);
  static const canvas = Color(0xFFF7F9FC);
  static const surface = Colors.white;
  static const surfaceSoft = Color(0xFFF1F6FF);
  static const line = Color(0xFFE2E8F0);
  static const border = Color(0xFFCBD5E1);
  static const text = Color(0xFF111827);
  static const textStrong = Color(0xFF001040);
  static const muted = Color(0xFF64748B);
  static const subtle = Color(0xFF94A3B8);
  static const warning = Color(0xFFD97706);
  static const danger = Color(0xFFDC2626);
  static const success = Color(0xFF059669);
  static const disabledSurface = Color(0xFFF1F5F9);
  static const productFallback = Color(0xFFD9E7FF);
  static const onPrimaryMuted = Color(0xE6FFFFFF);
}

class MokposRadius {
  const MokposRadius._();

  static const xs = 4.0;
  static const sm = 8.0;
  static const md = 12.0;
  static const lg = 16.0;
  static const xl = 22.0;
}

class MokposSpacing {
  const MokposSpacing._();

  static const xs = 4.0;
  static const sm = 8.0;
  static const md = 12.0;
  static const lg = 16.0;
  static const xl = 24.0;
  static const xxl = 32.0;
}

class MokposShadow {
  const MokposShadow._();

  static const soft = [
    BoxShadow(color: Color(0x14001040), blurRadius: 18, offset: Offset(0, 8)),
  ];
}

class NojposColors {
  const NojposColors._();

  static const primary = MokposColors.primary;
  static const primaryDark = MokposColors.primaryDark;
  static const primarySoft = MokposColors.primarySoft;
  static const primaryBorder = Color(0xFFC9DCFF);
  static const primaryText = Color(0xFF003C8F);

  /// POS operational accent. Use this for online/active/healthy terminal
  /// signals, not for primary commerce actions.
  static const secondary = MokposColors.accent;
  static const secondaryDark = MokposColors.accentDark;
  static const secondarySoft = MokposColors.accentSoft;

  /// Historical alias. Prefer [secondary] in new code.
  static const accent = secondary;
  static const accentDark = secondaryDark;
  static const accentSoft = secondarySoft;

  static const canvas = MokposColors.canvas;
  static const surface = MokposColors.surface;
  static const surfaceSoft = MokposColors.surfaceSoft;
  static const line = MokposColors.line;
  static const border = MokposColors.border;
  static const text = MokposColors.text;
  static const textStrong = MokposColors.textStrong;
  static const muted = MokposColors.muted;
  static const subtle = MokposColors.subtle;
  static const warning = MokposColors.warning;
  static const danger = MokposColors.danger;
  static const success = MokposColors.success;
  static const disabledSurface = MokposColors.disabledSurface;
  static const productFallback = MokposColors.productFallback;
  static const onPrimaryMuted = MokposColors.onPrimaryMuted;

  static const successSurface = Color(0xFFECFDF5);
  static const successBorder = Color(0xFFA7F3D0);
  static const dangerSurface = Color(0xFFFEF2F2);
  static const dangerBorder = Color(0xFFFECACA);
  static const warningSurface = Color(0xFFFFFBEB);
  static const warningBorder = Color(0xFFFDE68A);
  static const infoSurface = primarySoft;
  static const infoBorder = primaryBorder;
  static const glass = Color(0x33FFFFFF);

  static const premiumGradient = LinearGradient(
    colors: [Color(0xFF0060F0), Color(0xFF009E82)],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );

  static const softGradient = LinearGradient(
    colors: [Color(0xFFFFFFFF), Color(0xFFF1F6FF)],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );
}

class NojposRadius {
  const NojposRadius._();

  static const xs = MokposRadius.xs;
  static const sm = MokposRadius.sm;
  static const md = MokposRadius.md;
  static const lg = MokposRadius.lg;
  static const xl = MokposRadius.xl;
  static const xxl = 28.0;
}

class NojposSpacing {
  const NojposSpacing._();

  static const xs = MokposSpacing.xs;
  static const sm = MokposSpacing.sm;
  static const md = MokposSpacing.md;
  static const lg = MokposSpacing.lg;
  static const xl = MokposSpacing.xl;
  static const xxl = MokposSpacing.xxl;
  static const xxxl = 40.0;
}

class NojposShadow {
  const NojposShadow._();

  static const soft = MokposShadow.soft;
  static const card = [
    BoxShadow(color: Color(0x12001040), blurRadius: 24, offset: Offset(0, 12)),
  ];
  static const floating = [
    BoxShadow(color: Color(0x1F001040), blurRadius: 32, offset: Offset(0, 18)),
  ];
}
