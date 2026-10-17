import type { TaxReport } from './api'

export type TaxReportRow = TaxReport

export function taxMonthsForYear(year: number): string[] {
  return Array.from({ length: 12 }, (_, index) => `${year}-${String(index + 1).padStart(2, '0')}`)
}

export function taxMonthLabel(month: string): string {
  const [year, monthNumber] = month.split('-').map(Number)
  return new Date(year, monthNumber - 1, 1).toLocaleDateString('id-ID', { month: 'long' })
}

export function taxRowsForYear(months: string[], reports: Partial<Record<string, TaxReport>>): TaxReportRow[] {
  return months.map((month) => {
    const report = reports[month]
    return {
      month,
      omzet_bruto: report?.omzet_bruto ?? 0,
      refund_total: report?.refund_total ?? 0,
      omzet_neto: report?.omzet_neto ?? 0,
      dpp: report?.dpp ?? 0,
      ppn_terkumpul: report?.ppn_terkumpul ?? 0,
      trx_count: report?.trx_count ?? 0,
      refund_count: report?.refund_count ?? 0,
      pph_final_estimate: report?.pph_final_estimate ?? 0,
    }
  })
}
