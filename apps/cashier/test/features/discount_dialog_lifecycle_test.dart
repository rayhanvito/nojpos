import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/features/pos/models/cart_item.dart';
import 'package:nojpos_tablet_ui/features/pos/models/product.dart';
import 'package:nojpos_tablet_ui/features/pos/providers/pos_providers.dart';
import 'package:nojpos_tablet_ui/features/pos/widgets/pos_dialogs.dart';

void main() {
  group('Discount dialogs', () {
    testWidgets(
      'should open submit and close repeatedly without lifecycle exceptions',
      (tester) async {
        await tester.pumpWidget(
          const ProviderScope(
            child: MaterialApp(home: Scaffold(body: _DiscountDialogHarness())),
          ),
        );

        for (var index = 0; index < 3; index += 1) {
          await tester.tap(find.text('Diskon item'));
          await tester.pumpAndSettle();
          await tester.enterText(find.byType(TextField), '1000');
          await tester.tap(find.text('Simpan'));
          await tester.pumpAndSettle();
          expect(tester.takeException(), isNull);
        }

        for (var index = 0; index < 3; index += 1) {
          await tester.tap(find.text('Diskon cart'));
          await tester.pumpAndSettle();
          await tester.enterText(find.byType(TextField), '500');
          await tester.tap(find.text('Simpan'));
          await tester.pumpAndSettle();
          expect(tester.takeException(), isNull);
        }
      },
    );
  });
}

const _product = Product(
  id: 'product-discount',
  name: 'Kopi Susu',
  category: 'Minuman',
  price: 18000,
  imageUrl: '',
);

class _DiscountDialogHarness extends ConsumerWidget {
  const _DiscountDialogHarness();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    const item = CartItem(product: _product, quantity: 2);
    return Column(
      children: [
        ElevatedButton(
          onPressed: () {
            ref.read(cartProvider.notifier).replaceWith([item]);
            showItemDiscountDialog(context, ref, item);
          },
          child: const Text('Diskon item'),
        ),
        ElevatedButton(
          onPressed: () {
            ref.read(cartProvider.notifier).replaceWith([item]);
            showCartDiscountDialog(context, ref);
          },
          child: const Text('Diskon cart'),
        ),
      ],
    );
  }
}
