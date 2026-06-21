import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/features/pos/widgets/pos_dialogs.dart';

void main() {
  testWidgets('NojPOS Care dialog shows production support panel', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1200, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      ProviderScope(
        child: MaterialApp(
          home: Consumer(
            builder: (context, ref, child) {
              return Scaffold(
                body: Center(
                  child: TextButton(
                    onPressed: () => showNojposCareDialog(context, ref),
                    child: const Text('Buka Care'),
                  ),
                ),
              );
            },
          ),
        ),
      ),
    );

    await tester.tap(find.text('Buka Care'));
    await tester.pumpAndSettle();

    expect(find.text('NojPOS Care'), findsOneWidget);
    expect(find.text('Butuh bantuan operasional?'), findsOneWidget);
    expect(find.text('Hubungi administrator NOJPOS Anda'), findsOneWidget);
    expect(find.text('Status aplikasi'), findsOneWidget);
    expect(find.text('Outlet & terminal'), findsOneWidget);
    expect(find.text('Copy diagnostic info'), findsOneWidget);
    expect(find.textContaining('coming soon'), findsNothing);
    expect(find.textContaining('Belum aktif di versi ini'), findsNothing);
  });
}
