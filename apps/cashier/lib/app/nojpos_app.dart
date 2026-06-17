import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import 'router/app_router.dart';
import 'theme.dart';

class NojposApp extends StatelessWidget {
  const NojposApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'NojPOS Tablet',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: MokposColors.primary,
          brightness: Brightness.light,
        ),
        scaffoldBackgroundColor: MokposColors.canvas,
        textTheme: GoogleFonts.plusJakartaSansTextTheme(),
        useMaterial3: true,
      ),
      routerConfig: appRouter,
    );
  }
}
