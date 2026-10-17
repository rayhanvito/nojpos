import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/primitives/card'
import { AlertStrip, DataCard, CardGrid, SkeletonRows, Th, Td } from '../../../shared/ui'
import { fmtRp } from '../../../shared/session/store'
import { taxMonthLabel, type TaxReportRow } from '../../../shared/tax-report'

export function TaxSection({ year, currentMonth, selectedMonth, rows, loading, error, exportError, onSelectMonth }: {
  year: number
  currentMonth: string
  selectedMonth: string
  rows: TaxReportRow[] | null
  loading: boolean
  error: string
  exportError: string
  onSelectMonth: (month: string) => void
}) {
  return (
    <Card className="mt-5">
      <CardHeader>
        <CardTitle>Pajak</CardTitle>
        <CardDescription>Rekap pajak Januari–Desember {year}. Pilih bulan untuk ekspor CSV dari menu Export.</CardDescription>
      </CardHeader>
      <CardContent>
        {error && <AlertStrip className="mb-3">{error}</AlertStrip>}
        <div className="hidden overflow-x-auto min-[1280px]:block">
          <table className="w-full min-w-[980px] border-collapse" aria-busy={loading || undefined}>
            <thead>
              <tr>
                <Th>Bulan</Th>
                <Th right>Omzet Bruto</Th>
                <Th right>Refund</Th>
                <Th right>Omzet Neto</Th>
                <Th right>DPP</Th>
                <Th right>PPN</Th>
                <Th right>Jml Transaksi</Th>
                <Th right>Est. PPh Final</Th>
              </tr>
            </thead>
            {rows ? (
              <tbody>
                {rows.map((row) => {
                  const isCurrentMonth = row.month === currentMonth
                  const isSelected = row.month === selectedMonth
                  return (
                    <tr key={row.month} className={isCurrentMonth ? 'bg-surface' : undefined}>
                      <Td>
                        <button
                          type="button"
                          aria-pressed={isSelected}
                          title={`Pilih ${taxMonthLabel(row.month)} untuk ekspor CSV`}
                          onClick={() => onSelectMonth(row.month)}
                          className={`font-medium hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring ${isSelected ? 'text-fg underline' : 'text-muted'}`}
                        >
                          {taxMonthLabel(row.month)}
                        </button>
                        {isCurrentMonth && <span className="ml-2 rounded-full bg-paper px-2 py-0.5 text-[11px] font-medium text-steel">Bulan ini</span>}
                      </Td>
                      <Td right>{fmtRp(row.omzet_bruto)}</Td>
                      <Td right>{fmtRp(row.refund_total)}</Td>
                      <Td right>{fmtRp(row.omzet_neto)}</Td>
                      <Td right>{fmtRp(row.dpp)}</Td>
                      <Td right>{fmtRp(row.ppn_terkumpul)}</Td>
                      <Td right>{row.trx_count}</Td>
                      <Td right>{fmtRp(row.pph_final_estimate)}</Td>
                    </tr>
                  )
                })}
              </tbody>
            ) : !error ? <SkeletonRows cols={8} rows={12} /> : null}
          </table>
        </div>
        {rows && (
          <CardGrid>
            {rows.map((row) => {
              const isCurrentMonth = row.month === currentMonth
              const isSelected = row.month === selectedMonth
              return (
                <DataCard key={row.month} className={isSelected ? 'ring-1 ring-jet' : ''}>
                  <button
                    type="button" aria-pressed={isSelected}
                    title={`Pilih ${taxMonthLabel(row.month)} untuk ekspor CSV`}
                    onClick={() => onSelectMonth(row.month)}
                    className="flex w-full flex-col text-left outline-none"
                  >
                    <span className="flex items-center justify-between gap-2">
                      <span className="font-medium text-fg">{taxMonthLabel(row.month)}{isCurrentMonth && <span className="ml-2 rounded-full bg-surface px-2 py-0.5 text-[11px] font-medium text-steel">Bulan ini</span>}</span>
                      {isSelected && <span className="text-xs font-medium text-jet">Dipilih</span>}
                    </span>
                    <span className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-fg">{fmtRp(row.omzet_neto)}</span>
                    <span className="mt-0.5 text-right text-xs text-fog">Est. PPh Final {fmtRp(row.pph_final_estimate)}</span>
                    <span className="mt-1.5 text-xs leading-snug text-fog">Bruto {fmtRp(row.omzet_bruto)} · Refund {fmtRp(row.refund_total)} · DPP {fmtRp(row.dpp)} · PPN {fmtRp(row.ppn_terkumpul)} · {row.trx_count} transaksi</span>
                  </button>
                </DataCard>
              )
            })}
          </CardGrid>
        )}
        {exportError && <AlertStrip className="mt-3">{exportError}</AlertStrip>}
      </CardContent>
    </Card>
  )
}
