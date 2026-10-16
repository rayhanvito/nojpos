import { Banknote, BarChart3, ReceiptText, Sigma, TrendingUp, Wallet } from 'lucide-react'
import type { ReportBundle } from '../../lib/api'
import { fmtRp, fmtShort } from '../../lib/store'
import { CardGrid, DataCard, Pager, StatCard, Td, Th } from '../../lib/ui'
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart'
import { Bar, CartesianGrid, ComposedChart, Line, LineChart, XAxis, YAxis } from 'recharts'
import { delta, profitMarginConfig, profitTrendConfig, tipLabel, type profitDailyFrom, type prodProfitFrom } from './derive'
import { ChartCard } from './ChartCard'

export function ProfitTab({ data, profitDaily, profitInsights, prodProfit, profitPages, profitPageSafe, setProfitPage, profitPageSize, onProfitPageSize, profitTotal, prev, banding, animate }: {
  data: ReportBundle
  profitDaily: ReturnType<typeof profitDailyFrom>
  profitInsights: string[]
  prodProfit: ReturnType<typeof prodProfitFrom>
  profitPages: number
  profitPageSafe: number
  profitPageSize: number
  onProfitPageSize: (n: number) => void
  profitTotal: number
  setProfitPage: (v: number) => void
  prev: ReportBundle | null
  banding: string
  animate: boolean
}) {
  return (
    <>
      <div className="space-y-5">
        <div className="grid grid-cols-2 gap-4 xl:grid-cols-4">
          <StatCard label="Profit Kotor" value={fmtRp(data.summary.gross_profit)} sub={`dari omzet ${fmtRp(data.summary.omzet)}`} compare={delta(data.summary.gross_profit, prev?.summary.gross_profit, banding)} icon={TrendingUp} />
          <StatCard
            label="Margin Profit"
            value={data.summary.omzet > 0 ? `${Math.round((data.summary.gross_profit / data.summary.omzet) * 100)}%` : '0%'}
            sub="profit per rupiah omzet"
            compare={(() => {
              if (!prev || prev.summary.omzet <= 0 || data.summary.omzet <= 0) return null
              const d = Math.round((data.summary.gross_profit / data.summary.omzet) * 100) - Math.round((prev.summary.gross_profit / prev.summary.omzet) * 100)
              return d === 0 ? `±0% dari ${banding}` : `${d > 0 ? '↑' : '↓'} ${Math.abs(d)}% dari ${banding}`
            })()}
            icon={BarChart3}
          />
          <StatCard label="Pendapatan" value={fmtRp(data.summary.omzet)} sub="total penjualan" compare={delta(data.summary.omzet, prev?.summary.omzet, banding)} icon={Banknote} />
          <StatCard label="HPP" value={fmtRp(data.transactions.reduce((n, t) => n + t.hpp, 0))} sub="modal barang terjual" compare={delta(data.transactions.reduce((n, t) => n + t.hpp, 0), prev ? prev.transactions.reduce((n, t) => n + t.hpp, 0) : undefined, banding)} invert icon={Wallet} />
          <StatCard label="Pengeluaran Operasional" value={fmtRp(data.summary.expense_total ?? 0)} sub="biaya operasional periode ini" icon={ReceiptText} />
          <StatCard label="Laba Bersih" value={fmtRp(data.summary.gross_profit - (data.summary.expense_total ?? 0))} sub="profit kotor − pengeluaran operasional" icon={Sigma} />
        </div>

        <ChartCard title="Pengeluaran per Kategori" sub="Biaya operasional pada periode ini">
          {(data.expenses ?? []).length === 0 ? (
            <p className="py-10 text-center text-sm text-muted-foreground">Tidak ada pengeluaran tercatat pada periode ini.</p>
          ) : (
            <>
            <div className="divide-y divide-dove min-[1280px]:hidden">
              {(data.expenses ?? []).map((row) => (
                <div key={row.category} className="flex items-center justify-between gap-3 py-2">
                  <p className="min-w-0 text-sm text-fg"><span className="font-medium">{row.category}</span> <span className="text-xs text-fog">· {row.count} catatan</span></p>
                  <p className="shrink-0 text-sm font-semibold tabular-nums text-fg">{fmtRp(row.total)}</p>
                </div>
              ))}
            </div>
            <table className="hidden w-full border-collapse min-[1280px]:table">
              <thead><tr><Th>Kategori</Th><Th right>Catatan</Th><Th right>Total</Th></tr></thead>
              <tbody>
                {(data.expenses ?? []).map((row) => (
                  <tr key={row.category}>
                    <Td><span className="font-medium text-fg">{row.category}</span></Td>
                    <Td right>{row.count}</Td>
                    <Td right>{fmtRp(row.total)}</Td>
                  </tr>
                ))}
              </tbody>
            </table>
            </>
          )}
        </ChartCard>

        <div className="grid gap-4 lg:grid-cols-[2fr_1fr]">
          <ChartCard title="Tren Profit" sub={profitDaily.length > 0 ? 'Pendapatan, HPP, dan profit per tanggal' : 'Tidak ada data pada periode ini'}>
            {profitDaily.length === 0 ? (
              <p className="py-16 text-center text-sm text-muted-foreground">Tidak ada transaksi pada periode ini.</p>
            ) : (
              <>
                <div className="mb-3 flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
                  <span className="flex items-center gap-1.5">
                    <span className="size-2.5 rounded-full" style={{ background: 'var(--chart-1)' }} />
                    Pendapatan
                  </span>
                  <span className="flex items-center gap-1.5">
                    <span className="size-2.5 rounded-full" style={{ background: 'var(--muted)' }} />
                    HPP
                  </span>
                  <span className="flex items-center gap-1.5">
                    <span className="size-2.5 rounded-full" style={{ background: 'var(--chart-omzet)' }} />
                    Profit
                  </span>
                </div>
                <ChartContainer config={profitTrendConfig} className="h-64 w-full [&_:focus]:outline-none">
                  <ComposedChart data={profitDaily} margin={{ top: 8, right: 8, bottom: 0, left: 8 }} barCategoryGap="30%">
                    <CartesianGrid vertical={false} strokeDasharray="4 4" stroke="color-mix(in oklch, var(--foreground) 18%, transparent)" />
                    <XAxis dataKey="label" interval="preserveStartEnd" minTickGap={24} tickLine={false} axisLine={false} tickMargin={8} tick={{ fontSize: 11 }} />
                    <YAxis yAxisId="left" tickLine={false} axisLine={false} width={44} domain={[0, 'auto']} tickFormatter={(v: number) => fmtShort(v)} tick={{ fontSize: 11 }} />
                    <YAxis yAxisId="right" orientation="right" hide domain={[0, 'auto']} />
<ChartTooltip cursor={{ stroke: 'var(--border)', strokeWidth: 1 }} content={<ChartTooltipContent labelFormatter={tipLabel} formatter={(v) => fmtRp(Number(v))} />} />
                    <Bar yAxisId="left" dataKey="revenue" name="Pendapatan" fill="var(--color-revenue)" radius={[6, 6, 0, 0]} maxBarSize={28} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out" />
                    <Bar yAxisId="left" dataKey="hpp" name="HPP" fill="var(--color-hpp)" radius={[6, 6, 0, 0]} maxBarSize={28} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out" />
                    <Line yAxisId="right" type="monotone" dataKey="profit" name="Profit" stroke="var(--chart-omzet)" strokeWidth={2.5} strokeLinecap="round" strokeLinejoin="round" dot={{ r: 3, fill: 'var(--chart-omzet)', strokeWidth: 0 }} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out" />
                  </ComposedChart>
                </ChartContainer>
              </>
            )}
          </ChartCard>

          <ChartCard title="Margin Profit" sub={profitDaily.length > 0 ? 'Margin per tanggal' : 'Belum ada data'}>
            {profitDaily.length === 0 ? (
              <p className="py-16 text-center text-sm text-muted-foreground">Tidak ada data.</p>
            ) : (
              <ChartContainer config={profitMarginConfig} className="h-64 w-full [&_:focus]:outline-none">
                <LineChart data={profitDaily} margin={{ top: 8, right: 8, bottom: 0, left: 8 }}>
                  <CartesianGrid vertical={false} strokeDasharray="4 4" stroke="color-mix(in oklch, var(--foreground) 18%, transparent)" />
                  <XAxis dataKey="label" interval="preserveStartEnd" minTickGap={24} tickLine={false} axisLine={false} tickMargin={8} tick={{ fontSize: 11 }} />
                  <YAxis tickLine={false} axisLine={false} width={40} domain={[0, 'auto']} tickFormatter={(v: number) => `${v}%`} tick={{ fontSize: 11 }} />
                  <ChartTooltip cursor={{ stroke: 'var(--border)', strokeWidth: 1 }} content={<ChartTooltipContent labelFormatter={tipLabel} formatter={(v) => `Margin: ${v}%`} />} />
                  <Line type="monotone" dataKey="margin" name="Margin" stroke="var(--color-margin)" strokeWidth={2.5} strokeLinecap="round" strokeLinejoin="round" dot={{ r: 3, fill: 'var(--color-margin)', strokeWidth: 0 }} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out" />
                </LineChart>
              </ChartContainer>
            )}
          </ChartCard>
        </div>

        <div className="grid gap-4 lg:grid-cols-[1.6fr_1fr]">
          <ChartCard title="Produk Paling Menguntungkan" sub="Urut profit terbesar">
            <div className="hidden overflow-x-auto min-[1280px]:block">
              {prodProfit.length === 0 ? (
                <p className="py-10 text-center text-sm text-muted-foreground">Tidak ada data.</p>
              ) : (
                <table className="w-full border-collapse">
                  <thead>
                    <tr><Th>#</Th><Th>Produk</Th><Th right>Qty Terjual</Th><Th right>Pendapatan</Th><Th right>HPP</Th><Th right>Profit</Th><Th right>Margin</Th></tr>
                  </thead>
                  <tbody>
                    {prodProfit.slice(profitPageSafe * profitPageSize, (profitPageSafe + 1) * profitPageSize).map((p, i) => (
                      <tr key={p.product_id}>
                        <Td mono>{String(profitPageSafe * profitPageSize + i + 1).padStart(2, '0')}</Td>
                        <Td><span className="font-medium text-fg">{p.name}</span></Td>
                        <Td right>{p.qty}</Td>
                        <Td right>{fmtRp(p.revenue)}</Td>
                        <Td right>{fmtRp(p.revenue - p.profit)}</Td>
                        <Td right><span className="font-medium text-fg">{fmtRp(p.profit)}</span></Td>
                        <Td right>{p.revenue > 0 ? `${Math.round((p.profit / p.revenue) * 100)}%` : '—'}</Td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
            {prodProfit.length > 0 && (
              <CardGrid>
                {prodProfit.slice(profitPageSafe * profitPageSize, (profitPageSafe + 1) * profitPageSize).map((p, i) => (
                  <DataCard key={p.product_id}>
                    <div className="flex items-center justify-between gap-2">
                      <p className="min-w-0 truncate font-medium text-fg"><span className="mr-1.5 font-mono text-xs text-fog">{String(profitPageSafe * profitPageSize + i + 1).padStart(2, '0')}</span>{p.name}</p>
                      <p className="shrink-0 text-[15px] font-semibold tabular-nums text-fg">{fmtRp(p.profit)} <span className="text-xs font-normal text-fog">({p.revenue > 0 ? `${Math.round((p.profit / p.revenue) * 100)}%` : '—'})</span></p>
                    </div>
                    <div className="mt-2 grid grid-cols-3 gap-2 text-xs">
                      <p><span className="block text-fog">Pendapatan</span><span className="font-medium tabular-nums text-fg">{fmtRp(p.revenue)}</span></p>
                      <p><span className="block text-fog">HPP</span><span className="font-medium tabular-nums text-fg">{fmtRp(p.revenue - p.profit)}</span></p>
                      <p><span className="block text-fog">Qty</span><span className="font-medium tabular-nums text-fg">{p.qty}</span></p>
                    </div>
                  </DataCard>
                ))}
              </CardGrid>
            )}
            <Pager page={profitPageSafe} total={profitPages} onChange={setProfitPage} totalItems={profitTotal} pageSize={profitPageSize} onPageSizeChange={onProfitPageSize} />
          </ChartCard>

          <ChartCard title="Profit Insight" sub="Dari data periode ini">
            {profitInsights.length === 0 && data.summary.trx_count === 0 ? (
              <p className="py-12 text-center text-sm text-muted-foreground">Tidak ada data.</p>
            ) : (
              <div className="space-y-4">
                <div className="rounded-lg bg-surface p-4">
                  <p className="text-[13px] text-muted-foreground">Profit per Transaksi</p>
                  <p className="mt-1 text-2xl font-semibold tabular-nums tracking-tight">
                    {data.summary.trx_count > 0 ? fmtRp(Math.round(data.summary.gross_profit / data.summary.trx_count)) : 'Rp 0'}
                  </p>
                  {(() => {
                    const c = prev && prev.summary.trx_count > 0 && data.summary.trx_count > 0
                      ? delta(Math.round(data.summary.gross_profit / data.summary.trx_count), Math.round(prev.summary.gross_profit / prev.summary.trx_count))
                      : null
                    return c ? <p className={`mt-1 text-xs tabular-nums ${c.startsWith('↑') ? 'text-sprout' : c.startsWith('↓') ? 'text-ember' : 'text-muted-foreground'}`}>{c}</p> : null
                  })()}
                </div>
                {profitInsights.map((s) => (
                  <div key={s} className="flex items-start gap-2.5">
                    <span className="size-2 mt-1.5 shrink-0 rounded-full" style={{ background: 'var(--chart-2)' }} aria-hidden="true" />
                    <p className="text-sm leading-relaxed">{s}</p>
                  </div>
                ))}
              </div>
            )}
          </ChartCard>
        </div>
      </div>
    </>
  )
}
