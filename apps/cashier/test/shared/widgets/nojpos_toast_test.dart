import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:nojpos_tablet_ui/shared/widgets/nojpos_toast.dart';

void main() {
  testWidgets('NojposToast renders success message and action works', (
    tester,
  ) async {
    var tapped = false;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Builder(
            builder: (context) => FilledButton(
              onPressed: () => NojposToast.success(
                context,
                'Tersimpan',
                description: 'Perubahan berhasil disimpan.',
                actionLabel: 'Lihat',
                onAction: () => tapped = true,
              ),
              child: const Text('Show toast'),
            ),
          ),
        ),
      ),
    );

    await tester.tap(find.text('Show toast'));
    await tester.pumpAndSettle();

    expect(find.text('Tersimpan'), findsOneWidget);
    expect(find.text('Lihat'), findsOneWidget);

    final actionButton = find.widgetWithText(TextButton, 'Lihat');
    expect(actionButton, findsOneWidget);
    await tester.tap(actionButton, warnIfMissed: false);
    await tester.pump();
    expect(tapped, isTrue);
  });
}
