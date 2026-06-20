import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nojpos_tablet_ui/features/payment/pages/payment_screen.dart';
import 'package:nojpos_tablet_ui/features/payment/widgets/promotion_quote_panel.dart';
import 'package:nojpos_tablet_ui/features/transactions/repositories/transaction_repository.dart';

void main() {
  test('repository DTO parses quote with applied promotions', () {
    final quote = CheckoutQuote.fromJson({
      'quote_id': 'quote-id',
      'quote_hash': 'hash-id',
      'quote_revision': 'rev-1',
      'checkout_idempotency_key': 'idem-key',
      'checkout_token': 'checkout-token',
      'server_time': '2026-06-20T10:00:00+07:00',
      'subtotal': 100000,
      'item_discount_total': 5000,
      'cart_discount_total': 3000,
      'promotion_discount_total': 12000,
      'manual_discount_total': 8000,
      'discount_total': 20000,
      'service_charge_total': 4000,
      'tax_total': 4400,
      'rounding_total': 0,
      'grand_total': 88400,
      'applied_promotions': [
        {
          'promotion_id': 'promo-id',
          'name': 'Diskon Hemat',
          'code': 'HEMAT',
          'type': 'percent',
          'discount_amount': 12000,
          'affected_items': [
            {'product_id': 'product-id', 'discount_amount': 12000},
          ],
        },
      ],
    });

    expect(quote.quoteId, 'quote-id');
    expect(quote.quoteHash, 'hash-id');
    expect(quote.quoteRevision, 'rev-1');
    expect(quote.promotionDiscountTotal, 12000);
    expect(quote.manualDiscountTotal, 8000);
    expect(quote.grandTotal, 88400);
    expect(quote.appliedPromotions.single.name, 'Diskon Hemat');
    expect(
      quote.appliedPromotions.single.affectedItems.single.productId,
      'product-id',
    );
  });

  test('repository DTO parses rejected promotions', () {
    final quote = CheckoutQuote.fromJson({
      'subtotal': 50000,
      'discount_total': 0,
      'grand_total': 50000,
      'rejected_promotions': [
        {
          'code': 'EXPIRED',
          'reason_code': 'INACTIVE_OR_EXPIRED',
          'message': 'Promo tidak aktif atau sudah kedaluwarsa.',
        },
      ],
    });

    expect(quote.rejectedPromotions.single.code, 'EXPIRED');
    expect(quote.rejectedPromotions.single.reasonCode, 'INACTIVE_OR_EXPIRED');
    expect(
      quote.rejectedPromotions.single.message,
      'Promo tidak aktif atau sudah kedaluwarsa.',
    );
  });

  test(
    'checkout draft sends promotion codes and quote token hash revision',
    () {
      final quote = CheckoutQuote.fromJson({
        'quote_id': 'quote-id',
        'quote_hash': 'hash-id',
        'quote_revision': 'rev-1',
        'checkout_token': 'checkout-token',
        'grand_total': 42000,
      });
      final draft = CheckoutDraft(
        idempotencyKey: 'idem-key',
        outletId: 'outlet-id',
        deviceId: 'device-id',
        cashierId: 'cashier-id',
        shiftId: 'shift-id',
        promotionCodes: const ['HEMAT'],
        quote: quote,
        items: const [
          CheckoutItem(
            productId: 'product-id',
            name: 'Produk',
            quantity: 1,
            unitPrice: 42000,
          ),
        ],
        payments: const [
          CheckoutPayment(method: 'cash', amount: 42000, isCash: true),
        ],
        paidAmount: 42000,
      );

      final payload = draft.toApiJson();
      expect(payload['promotion_codes'], ['HEMAT']);
      expect(payload['quote_id'], 'quote-id');
      expect(payload['checkout_token'], 'checkout-token');
      expect(payload['quote_token'], 'checkout-token');
      expect(payload['quote_hash'], 'hash-id');
      expect(payload['quote_revision'], 'rev-1');
    },
  );

  testWidgets('promo panel shows applied and rejected promotions', (
    tester,
  ) async {
    final controller = TextEditingController(text: 'HEMAT');
    addTearDown(controller.dispose);
    final quote = CheckoutQuote.fromJson({
      'promotion_discount_total': 12000,
      'applied_promotions': [
        {
          'promotion_id': 'promo-id',
          'name': 'Diskon Hemat',
          'discount_amount': 12000,
        },
      ],
      'rejected_promotions': [
        {
          'code': 'OLD',
          'reason_code': 'INACTIVE_OR_EXPIRED',
          'message': 'Promo tidak aktif atau sudah kedaluwarsa.',
        },
      ],
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PromotionQuotePanel(
            quote: quote,
            promotionCodes: const ['HEMAT'],
            controller: controller,
            onApply: () {},
            onRemove: (_) {},
          ),
        ),
      ),
    );

    expect(
      find.byKey(const ValueKey('applied_promotion_promo-id')),
      findsOneWidget,
    );
    expect(
      find.text('Promo tidak aktif atau sudah kedaluwarsa.'),
      findsOneWidget,
    );
    expect(
      find.byKey(const ValueKey('promotion_discount_total')),
      findsOneWidget,
    );
  });

  testWidgets('promo panel shows empty state', (tester) async {
    final controller = TextEditingController();
    addTearDown(controller.dispose);

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PromotionQuotePanel(
            quote: null,
            promotionCodes: const [],
            controller: controller,
            onApply: () {},
            onRemove: (_) {},
          ),
        ),
      ),
    );

    expect(find.byKey(const ValueKey('promotion_empty_state')), findsOneWidget);
  });

  test('promotion codes provider updates code list', () {
    final container = ProviderContainer();
    addTearDown(container.dispose);

    container.read(paymentPromotionCodesProvider.notifier).add('HEMAT');
    container.read(paymentPromotionCodesProvider.notifier).add('HEMAT');
    container.read(paymentPromotionCodesProvider.notifier).add('WEEKEND');
    expect(container.read(paymentPromotionCodesProvider), ['HEMAT', 'WEEKEND']);

    container.read(paymentPromotionCodesProvider.notifier).remove('HEMAT');
    expect(container.read(paymentPromotionCodesProvider), ['WEEKEND']);
  });
}
