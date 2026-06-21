import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:nojpos_tablet_ui/shared/widgets/nojpos_state_view.dart';

void main() {
  testWidgets('NojposStateView renders empty state and handles action tap', (
    tester,
  ) async {
    var tapped = false;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: NojposStateView.empty(
            title: 'Produk kosong',
            subtitle: 'Tambahkan produk pertama untuk mulai berjualan.',
            actionLabel: 'Tambah Produk',
            onAction: () => tapped = true,
          ),
        ),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.text('Produk kosong'), findsOneWidget);
    expect(find.text('Tambah Produk'), findsOneWidget);

    await tester.tap(find.text('Tambah Produk'));
    expect(tapped, isTrue);
  });

  testWidgets(
    'NojposStateView does not throw when illustration asset is missing',
    (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: NojposStateView.error(
              title: 'Gagal memuat',
              subtitle: 'Coba ulang beberapa saat lagi.',
              illustrationAsset: 'assets/illustrations/missing.svg',
              compact: true,
            ),
          ),
        ),
      );

      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      expect(find.text('Gagal memuat'), findsOneWidget);
    },
  );
}
