import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import 'router/app_router.dart';
import 'theme.dart';

class NojposApp extends StatelessWidget {
  const NojposApp({super.key});

  @override
  Widget build(BuildContext context) {
    final textTheme = GoogleFonts.plusJakartaSansTextTheme();

    return MaterialApp.router(
      title: 'NojPOS Tablet',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: const ColorScheme.light(
          primary: NojposColors.primary,
          onPrimary: Colors.white,
          primaryContainer: NojposColors.primarySoft,
          onPrimaryContainer: NojposColors.primaryText,
          secondary: NojposColors.secondary,
          onSecondary: Colors.white,
          secondaryContainer: NojposColors.secondarySoft,
          onSecondaryContainer: NojposColors.secondaryDark,
          error: NojposColors.danger,
          onError: Colors.white,
          errorContainer: NojposColors.dangerSurface,
          onErrorContainer: NojposColors.danger,
          surface: NojposColors.surface,
          onSurface: NojposColors.text,
          outline: NojposColors.border,
          outlineVariant: NojposColors.line,
        ),
        scaffoldBackgroundColor: NojposColors.canvas,
        textTheme: textTheme.apply(
          bodyColor: NojposColors.text,
          displayColor: NojposColors.textStrong,
        ),
        inputDecorationTheme: InputDecorationTheme(
          filled: true,
          fillColor: NojposColors.canvas,
          contentPadding: const EdgeInsets.symmetric(
            horizontal: NojposSpacing.lg,
            vertical: NojposSpacing.md,
          ),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(NojposRadius.lg),
            borderSide: const BorderSide(color: NojposColors.line),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(NojposRadius.lg),
            borderSide: const BorderSide(color: NojposColors.line),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(NojposRadius.lg),
            borderSide: const BorderSide(color: NojposColors.primary, width: 1.6),
          ),
          errorBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(NojposRadius.lg),
            borderSide: const BorderSide(color: NojposColors.danger),
          ),
          focusedErrorBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(NojposRadius.lg),
            borderSide: const BorderSide(color: NojposColors.danger, width: 1.6),
          ),
        ),
        filledButtonTheme: FilledButtonThemeData(
          style: FilledButton.styleFrom(
            minimumSize: const Size(0, 52),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(NojposRadius.lg),
            ),
            textStyle: const TextStyle(fontWeight: FontWeight.w900),
          ),
        ),
        outlinedButtonTheme: OutlinedButtonThemeData(
          style: OutlinedButton.styleFrom(
            minimumSize: const Size(0, 48),
            side: const BorderSide(color: NojposColors.line),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(NojposRadius.lg),
            ),
            textStyle: const TextStyle(fontWeight: FontWeight.w800),
          ),
        ),
        snackBarTheme: SnackBarThemeData(
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(NojposRadius.md),
          ),
        ),
        cardTheme: CardThemeData(
          color: NojposColors.surface,
          surfaceTintColor: Colors.transparent,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(NojposRadius.lg),
            side: const BorderSide(color: NojposColors.line),
          ),
        ),
        dividerTheme: const DividerThemeData(
          color: NojposColors.line,
          thickness: 1,
          space: 1,
        ),
        useMaterial3: true,
      ),
      routerConfig: appRouter,
    );
  }
}
