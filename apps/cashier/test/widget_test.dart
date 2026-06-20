import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nojpos_tablet_ui/app/nojpos_app.dart';
import 'package:nojpos_tablet_ui/core/network/api_client.dart';
import 'package:nojpos_tablet_ui/core/storage/token_storage.dart';

void main() {
  testWidgets('renders login screen', (tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          tokenStorageProvider.overrideWith((ref) => InMemoryTokenStorage()),
        ],
        child: const NojposApp(),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Masuk ke NojPOS'), findsOneWidget);
    expect(find.text('Masuk'), findsOneWidget);
  });
}
