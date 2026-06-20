import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/providers/nojpos_session_provider.dart';
import '../repositories/report_repository.dart';

final salesSummaryProvider = FutureProvider.autoDispose<SalesSummary>((ref) {
  final session = ref.watch(nojposSessionProvider);
  return ref
      .watch(reportRepositoryProvider)
      .salesSummary(
        date: DateTime.now(),
        shiftId: session.activeShift?.id,
        outletId: session.outlet.id,
      );
});

final reportsDashboardProvider = FutureProvider.autoDispose<ReportsDashboard>((
  ref,
) async {
  final session = ref.watch(nojposSessionProvider);
  final repository = ref.watch(reportRepositoryProvider);
  final date = DateTime.now();
  final shiftId = session.activeShift?.id;
  final outletId = session.outlet.id;

  final results = await Future.wait<Object>([
    repository.salesSummary(date: date, shiftId: shiftId, outletId: outletId),
    repository.soldProducts(date: date, shiftId: shiftId, outletId: outletId),
    repository.paymentMethods(date: date, shiftId: shiftId, outletId: outletId),
    repository.cashierShifts(date: date, shiftId: shiftId, outletId: outletId),
    repository.voidRefundAudit(
      date: date,
      shiftId: shiftId,
      outletId: outletId,
    ),
    repository.topTen(date: date, shiftId: shiftId, outletId: outletId),
  ]);

  return ReportsDashboard(
    salesSummary: results[0] as SalesSummary,
    soldProducts: results[1] as SoldProductsReport,
    paymentMethods: results[2] as PaymentMethodsReport,
    cashierShifts: results[3] as CashierShiftsReport,
    voidRefundAudit: results[4] as VoidRefundAuditReport,
    topTen: results[5] as TopTenReport,
  );
});
