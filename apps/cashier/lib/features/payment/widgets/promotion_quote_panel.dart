import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';

import '../../../app/theme.dart';
import '../../pos/formatters.dart';
import '../../transactions/repositories/transaction_repository.dart';

class PromotionQuotePanel extends StatelessWidget {
  const PromotionQuotePanel({
    super.key,
    required this.quote,
    required this.promotionCodes,
    required this.controller,
    required this.onApply,
    required this.onRemove,
  });

  final CheckoutQuote? quote;
  final List<String> promotionCodes;
  final TextEditingController controller;
  final VoidCallback onApply;
  final ValueChanged<String> onRemove;

  @override
  Widget build(BuildContext context) {
    final applied = quote?.appliedPromotions ?? const <AppliedPromotion>[];
    final rejected = quote?.rejectedPromotions ?? const <RejectedPromotion>[];

    return Container(
      key: const ValueKey('promotion_quote_panel'),
      padding: const EdgeInsets.symmetric(vertical: 6),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: MokposColors.line)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(
                LucideIcons.ticketPercent,
                size: 14,
                color: MokposColors.muted,
              ),
              const SizedBox(width: 6),
              const Text(
                'Promo',
                style: TextStyle(fontSize: 12, fontWeight: FontWeight.w900),
              ),
              const Spacer(),
              if ((quote?.promotionDiscountTotal ?? 0) > 0)
                Text(
                  '-${rupiah(quote!.promotionDiscountTotal)}',
                  key: const ValueKey('promotion_discount_total'),
                  style: const TextStyle(
                    color: MokposColors.success,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
            ],
          ),
          const SizedBox(height: 6),
          Row(
            children: [
              Expanded(
                child: SizedBox(
                  height: 34,
                  child: TextField(
                    key: const ValueKey('promotion_code_input'),
                    controller: controller,
                    textCapitalization: TextCapitalization.characters,
                    decoration: const InputDecoration(
                      isDense: true,
                      hintText: 'Kode promo',
                      border: OutlineInputBorder(),
                      contentPadding: EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 8,
                      ),
                    ),
                    onSubmitted: (_) => onApply(),
                  ),
                ),
              ),
              const SizedBox(width: 6),
              SizedBox(
                height: 34,
                child: OutlinedButton(
                  key: const ValueKey('apply_promotion_code_button'),
                  onPressed: onApply,
                  child: const Text('Apply'),
                ),
              ),
            ],
          ),
          if (promotionCodes.isNotEmpty) ...[
            const SizedBox(height: 6),
            Wrap(
              spacing: 6,
              runSpacing: 4,
              children: [
                for (final code in promotionCodes)
                  InputChip(
                    key: ValueKey('promotion_code_$code'),
                    label: Text(code),
                    onDeleted: () => onRemove(code),
                  ),
              ],
            ),
          ],
          if (applied.isEmpty && rejected.isEmpty) ...[
            const SizedBox(height: 6),
            const Text(
              'Belum ada promo aktif.',
              key: ValueKey('promotion_empty_state'),
              style: TextStyle(color: MokposColors.muted, fontSize: 11),
            ),
          ],
          for (final promotion in applied) ...[
            const SizedBox(height: 6),
            Text(
              '${promotion.name}: -${rupiah(promotion.discountAmount)}',
              key: ValueKey('applied_promotion_${promotion.promotionId}'),
              style: const TextStyle(
                color: MokposColors.success,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
          for (final promotion in rejected) ...[
            const SizedBox(height: 6),
            Text(
              promotion.message.isEmpty
                  ? 'Promo ditolak: ${promotion.reasonCode}'
                  : promotion.message,
              key: ValueKey(
                'rejected_promotion_${promotion.code ?? promotion.reasonCode}',
              ),
              style: const TextStyle(
                color: MokposColors.danger,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ],
      ),
    );
  }
}
