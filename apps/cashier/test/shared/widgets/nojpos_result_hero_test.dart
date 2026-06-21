import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:nojpos_tablet_ui/shared/widgets/nojpos_result_hero.dart';

void main() {
  testWidgets('NojposResultHero renders success content', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: NojposResultHero.success(
            title: 'Pembayaran berhasil',
            subtitle: 'Transaksi sudah tersimpan.',
            amount: 'Rp 125.000',
          ),
        ),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.text('Pembayaran berhasil'), findsOneWidget);
    expect(find.text('Rp 125.000'), findsOneWidget);
  });

  testWidgets('NojposResultHero falls back when illustration is missing', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: NojposResultHero.failed(
            title: 'Pembayaran gagal',
            subtitle: 'Gunakan metode lain.',
            illustrationAsset: 'assets/illustrations/missing.svg',
            compact: true,
          ),
        ),
      ),
    );

    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    expect(find.text('Pembayaran gagal'), findsOneWidget);
  });
}
