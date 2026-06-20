import 'package:flutter/material.dart';

import '../../../app/theme.dart';
import '../../../features/pos/formatters.dart';
import '../repositories/report_repository.dart';

class ReportsFilterChips extends StatelessWidget {
  const ReportsFilterChips({required this.isCashier, super.key});

  final bool isCashier;

  @override
  Widget build(BuildContext context) {
    final labels = isCashier
        ? const ['Hari Ini', 'Shift aktif kasir', 'Outlet saat ini']
        : const ['Hari Ini', 'Outlet aktif', 'Scope dari server'];
    return Wrap(
      spacing: 10,
      runSpacing: 8,
      children: [
        for (final label in labels)
          Chip(
            label: Text(label),
            side: const BorderSide(color: MokposColors.line),
            backgroundColor: Colors.white,
            labelStyle: const TextStyle(
              color: MokposColors.muted,
              fontSize: 12,
              fontWeight: FontWeight.w700,
            ),
          ),
      ],
    );
  }
}

class SoldProductsReportPanel extends StatelessWidget {
  const SoldProductsReportPanel({required this.report, super.key});

  final SoldProductsReport report;

  @override
  Widget build(BuildContext context) {
    return _ReportCard(
      title: 'Produk Terjual',
      emptyTitle: 'Belum ada produk terjual',
      isEmpty: report.rows.isEmpty,
      child: Column(
        children: [
          for (final row in report.rows.take(5))
            _KeyValueRow(
              label: row.productName,
              value: '${row.quantitySold} item • ${rupiah(row.netSales)}',
              subtitle:
                  '${row.categoryName ?? 'Tanpa kategori'} • Diskon ${rupiah(row.discountTotal)}',
            ),
        ],
      ),
    );
  }
}

class PaymentMethodsReportPanel extends StatelessWidget {
  const PaymentMethodsReportPanel({required this.report, super.key});

  final PaymentMethodsReport report;

  @override
  Widget build(BuildContext context) {
    return _ReportCard(
      title: 'Jenis Bayar',
      emptyTitle: 'Belum ada pembayaran terkonfirmasi',
      isEmpty: report.rows.isEmpty,
      child: Column(
        children: [
          for (final row in report.rows)
            _KeyValueRow(
              label: row.method,
              value: rupiah(row.grossAmount),
              subtitle: '${row.transactionCount} transaksi',
            ),
        ],
      ),
    );
  }
}

class CashierShiftsReportPanel extends StatelessWidget {
  const CashierShiftsReportPanel({required this.report, super.key});

  final CashierShiftsReport report;

  @override
  Widget build(BuildContext context) {
    return _ReportCard(
      title: 'Shift Kasir',
      emptyTitle: 'Belum ada shift pada scope ini',
      isEmpty: report.rows.isEmpty,
      child: Column(
        children: [
          for (final row in report.rows.take(4))
            _KeyValueRow(
              label: '${row.cashier.name} • ${row.outlet.name}',
              value: row.status,
              subtitle:
                  'Buka ${rupiah(row.openingCash)} • Ekspektasi ${rupiah(row.expectedCash)} • Selisih ${rupiah(row.variance)}',
            ),
        ],
      ),
    );
  }
}

class VoidRefundAuditReportPanel extends StatelessWidget {
  const VoidRefundAuditReportPanel({required this.report, super.key});

  final VoidRefundAuditReport report;

  @override
  Widget build(BuildContext context) {
    final rows = [...report.rows, ...report.refundRows];
    return _ReportCard(
      title: 'Void / Refund Audit',
      emptyTitle: 'Belum ada void atau refund',
      isEmpty: rows.isEmpty,
      child: Column(
        children: [
          for (final row in rows.take(5))
            _KeyValueRow(
              label: '${row.type.toUpperCase()} • ${row.transactionNumber}',
              value: rupiah(row.refundAmount ?? row.amount),
              subtitle: row.reason ?? 'Tanpa alasan',
            ),
        ],
      ),
    );
  }
}

class TopTenReportPanel extends StatelessWidget {
  const TopTenReportPanel({required this.report, super.key});

  final TopTenReport report;

  @override
  Widget build(BuildContext context) {
    final grossProfitAvailability = report.dataAvailability['gross_profit'];
    return _ReportCard(
      title: 'Top 10',
      emptyTitle: 'Belum ada data top 10',
      isEmpty: report.bestSelling.isEmpty,
      footer: grossProfitAvailability == 'cost_coverage_incomplete'
          ? 'Gross profit belum tersedia: data cost belum lengkap.'
          : null,
      child: Column(
        children: [
          for (final row in report.bestSelling.take(5))
            _KeyValueRow(
              label: row.productName,
              value: '${row.quantitySold} item',
              subtitle: 'Best selling',
            ),
        ],
      ),
    );
  }
}

class _ReportCard extends StatelessWidget {
  const _ReportCard({
    required this.title,
    required this.emptyTitle,
    required this.isEmpty,
    required this.child,
    this.footer,
  });

  final String title;
  final String emptyTitle;
  final bool isEmpty;
  final Widget child;
  final String? footer;

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: MokposColors.line),
      ),
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              title,
              style: const TextStyle(
                color: MokposColors.text,
                fontWeight: FontWeight.w900,
                fontSize: 15,
              ),
            ),
            const SizedBox(height: 10),
            if (isEmpty)
              Text(
                emptyTitle,
                style: const TextStyle(color: MokposColors.muted, fontSize: 12),
              )
            else
              child,
            if (footer != null) ...[
              const SizedBox(height: 10),
              Text(
                footer!,
                style: const TextStyle(
                  color: MokposColors.warning,
                  fontSize: 12,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _KeyValueRow extends StatelessWidget {
  const _KeyValueRow({required this.label, required this.value, this.subtitle});

  final String label;
  final String value;
  final String? subtitle;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 7),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: MokposColors.text,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                if (subtitle != null)
                  Text(
                    subtitle!,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: MokposColors.muted,
                      fontSize: 11,
                    ),
                  ),
              ],
            ),
          ),
          const SizedBox(width: 10),
          Text(
            value,
            textAlign: TextAlign.right,
            style: const TextStyle(
              color: MokposColors.text,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}
