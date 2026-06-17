import 'package:flutter/material.dart';

class MokposColors {
  const MokposColors._();

  static const primary = Color(0xFF00BFA5);
  static const primaryDark = Color(0xFF009D88);
  static const primarySoft = Color(0xFFE5FAF6);
  static const accent = Color(0xFF1E88E5);
  static const canvas = Color(0xFFF5F7F8);
  static const surface = Colors.white;
  static const line = Color(0xFFE4E8EC);
  static const text = Color(0xFF18222D);
  static const muted = Color(0xFF707A86);
  static const warning = Color(0xFFFFD43B);
  static const danger = Color(0xFFE53935);
}

class MokposRadius {
  const MokposRadius._();

  static const sm = 8.0;
  static const md = 12.0;
  static const lg = 16.0;
  static const xl = 22.0;
}

class MokposShadow {
  const MokposShadow._();

  static const soft = [
    BoxShadow(color: Color(0x140B1F2A), blurRadius: 18, offset: Offset(0, 8)),
  ];
}
