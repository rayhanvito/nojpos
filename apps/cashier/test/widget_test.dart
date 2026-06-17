import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nojpos_tablet_ui/app/nojpos_app.dart';

void main() {
  testWidgets('renders login screen', (tester) async {
    await tester.pumpWidget(const ProviderScope(child: NojposApp()));

    expect(find.text('Masuk ke NojPOS'), findsOneWidget);
    expect(find.text('Lanjut ke PIN'), findsOneWidget);
  });
}
