import { Banknote, Package, ReceiptText, Sigma, Ticket, TicketPercent } from 'lucide-react'
import type { ReportBundle, Trx } from '../../lib/api'
import { fmtDate, fmtInv, fmtRp, fmtShort } from '../../lib/store'
import { CardGrid, DataCard, StatCard, StatusPill, Td, Th } from '../../lib/ui'
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart'
import { Bar, CartesianGrid, Cell, ComposedChart, LabelList, Line, Pie, PieChart, XAxis, YAxis } from 'recharts'
import { delta, payConfig, salesChartConfig, tipLabel, type dailyFrom, type topProductsFrom } from './derive'
import { ChartCard } from './ChartCard'
import { Link } from 'react-router'

export function SalesTab({ data, daily, topProducts, prev, banding, animate, needHourly, hourly, setTab }: {
  data: ReportBundle
  daily: ReturnType<typeof dailyFrom>
  topProducts: ReturnType<typeof topProductsFrom>
  prev: ReportBundle | null
  banding: string
  animate: boolean
  needHourly: boolean
  hourly: Trx[] | null | undefined
  setTab: (t: 'sales' | 'products' | 'profit' | 'stock' | 'shift' | 'kasir' | 'pajak') => void
}) {
  return (
    <>
      <div className="space-y-5">
        <div className="grid grid-cols-2 gap-4 xl:grid-cols-4">
          <StatCard label="Total Penjualan" value={fmtRp(data.summary.omzet)} sub={`${data.summary.trx_count} transaksi`} compare={delta(data.summary.omzet, prev?.summary.omzet, banding)} icon={Banknote} />
          <StatCard label="Jumlah Transaksi" value={String(data.summary.trx_count)} sub="selesai dalam periode ini" compare={delta(data.summary.trx_count, prev?.summary.trx_count, banding)} icon={ReceiptText} />
          <StatCard
            label="Rata-rata Transaksi"
            value={fmtRp(data.summary.trx_count > 0 ? Math.round(data.summary.omzet / data.summary.trx_count) : 0)}
            sub="per transaksi"
            compare={delta(
              data.summary.trx_count > 0 ? Math.round(data.summary.omzet / data.summary.trx_count) : 0,
              prev && prev.summary.trx_count > 0 ? Math.round(prev.summary.omzet / prev.summary.trx_count) : undefined,
              banding,
            )}
            icon={Sigma}
          />
          <StatCard label="Produk Terjual" value={String(data.summary.items_sold)} sub="satuan produk" compare={delta(data.summary.items_sold, prev?.summary.items_sold, banding)} icon={Package} />
          <StatCard label="Diskon Promo" value={fmtRp(data.summary.promo_discount_total ?? 0)} sub="dari promo otomatis" icon={TicketPercent} />
          <StatCard label="Diskon Voucher" value={fmtRp(data.summary.voucher_discount_total ?? 0)} sub="dari voucher berkode" icon={Ticket} />
        </div>

        <div className="grid gap-4 lg:grid-cols-[2fr_1fr]">
          <ChartCard title="Penjualan & Transaksi" sub={daily.length > 0 ? `Total ${fmtRp(data.summary.omzet)} · per tanggal` : 'Tidak ada data pada periode ini'}>
            {daily.length === 0 ? (
              needHourly && !hourly ? (
                <p className="py-16 text-center text-sm text-muted-foreground">Memuat rincian jam…</p>
              ) : (
                <p className="py-16 text-center text-sm text-muted-foreground">Tidak ada transaksi pada periode ini.</p>
              )
            ) : (
              <>
                <div className="mb-3 flex items-center gap-4 text-xs text-muted-foreground">
                  <span className="flex items-center gap-1.5">
                    <span className="size-2.5 rounded-full" style={{ background: 'var(--chart-omzet)' }} />
                    Penjualan
                  </span>
                  {!needHourly && (
                    <span className="flex items-center gap-1.5">
                      <span className="size-2.5 rounded-full" style={{ background: 'var(--chart-1)' }} />
                      Transaksi
                    </span>
                  )}
                </div>
                <ChartContainer config={salesChartConfig} className="h-64 w-full [&_:focus]:outline-none">
                  <ComposedChart data={daily} margin={{ top: 8, right: 8, bottom: 0, left: 8 }} barCategoryGap="30%">
                    <CartesianGrid vertical={false} strokeDasharray="4 4" stroke="color-mix(in oklch, var(--foreground) 18%, transparent)" />
                    <XAxis dataKey="label" interval="preserveStartEnd" minTickGap={24} tickLine={false} axisLine={false} tickMargin={8} tick={{ fontSize: 11 }} />
                    <YAxis yAxisId="left" tickLine={false} axisLine={false} width={44} domain={[0, 'auto']} tickFormatter={(v: number) => fmtShort(v)} tick={{ fontSize: 11 }} />
                    <YAxis yAxisId="right" orientation="right" hide domain={[0, 'auto']} />
                    <ChartTooltip cursor={{ stroke: 'var(--border)', strokeWidth: 1 }} content={<ChartTooltipContent labelFormatter={tipLabel} formatter={(v, name) => (name === 'Transaksi' ? `${v} transaksi` : fmtRp(Number(v)))} />} />
                    <Bar yAxisId="left" dataKey="omzet" name="Penjualan" fill="var(--color-omzet)" radius={[6, 6, 0, 0]} maxBarSize={40} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out">
                      {daily.length <= 14 && <LabelList dataKey="omzet" position="top" formatter={(v) => fmtShort(Number(v))} fontSize={11} className="fill-muted-foreground" />}
                    </Bar>
                    {!needHourly && (
                      <Line yAxisId="right" type="monotone" dataKey="trx" name="Transaksi" stroke="var(--chart-1)" strokeWidth={2.5} strokeLinecap="round" strokeLinejoin="round" dot={{ r: 3, fill: 'var(--chart-1)', strokeWidth: 0 }} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out" />
                    )}
                  </ComposedChart>
                </ChartContainer>
              </>
            )}
          </ChartCard>

          <ChartCard title="Metode Pembayaran" sub={data.by_method.length > 0 ? `Total ${fmtRp(data.summary.omzet)}` : 'Belum ada data'}>
            {data.by_method.length === 0 ? (
              <p className="py-16 text-center text-sm text-muted-foreground">Tidak ada data.</p>
            ) : (
              <div className="space-y-4">
                <ChartContainer config={{}} className="relative mx-auto h-44 w-full">
                  <PieChart>
                    <ChartTooltip content={<ChartTooltipContent formatter={(v) => fmtRp(Number(v))} hideLabel />} />
                    <Pie data={data.by_method.map((m) => ({ ...m, fill: payConfig[m.method] ?? 'var(--chart-4)' }))} dataKey="total" nameKey="method" innerRadius={52} outerRadius={74} paddingAngle={3} strokeWidth={0} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out">
                      {data.by_method.map((m, i) => (
                        <Cell key={i} fill={payConfig[m.method] ?? 'var(--chart-4)'} />
                      ))}
                    </Pie>
                  </PieChart>
                  <div className="pointer-events-none absolute inset-0 grid place-items-center">
                    <div className="text-center">
                      <p className="text-sm font-semibold tabular-nums tracking-tight">{fmtShort(data.summary.omzet)}</p>
                      <p className="text-[11px] text-muted-foreground">Total Penjualan</p>
                    </div>
                  </div>
                </ChartContainer>
                <div className="space-y-2">
                  {(() => {
                    const payTotal = data.by_method.reduce((n, m) => n + m.total, 0) || 1
                    return data.by_method.map((m) => (
                      <div key={m.method} className="flex items-center justify-between gap-3 text-sm">
                        <span className="flex min-w-0 items-center gap-2">
                          <span className="size-2.5 shrink-0 rounded-full" style={{ background: payConfig[m.method] ?? 'var(--chart-4)' }} />
                          <span className="truncate">{m.method}</span>
                        </span>
                        <span className="shrink-0 tabular-nums">
                          <span className="font-medium">{fmtRp(m.total)}</span>
                          <span className="text-muted-foreground"> · {Math.round((m.total / payTotal) * 100)}%</span>
                        </span>
                      </div>
                    ))
                  })()}
                </div>
              </div>
            )}
          </ChartCard>
        </div>

        <div className="grid gap-4 lg:grid-cols-[1.6fr_1fr]">
          <ChartCard
            title="Transaksi Terbaru"
            sub="10 transaksi terakhir periode ini"
            action={<Link to="/app/transaksi" className="shrink-0 text-[13px] font-medium text-jet hover:underline">Lihat semua →</Link>}
          >
            <div className="hidden overflow-x-auto min-[1280px]:block">
              {data.transactions.length === 0 ? (
                <p className="py-10 text-center text-sm text-muted-foreground">Tidak ada data.</p>
              ) : (
                <table className="w-full border-collapse">
                  <thead>
                    <tr><Th>Invoice</Th><Th>Tanggal</Th><Th>Kasir</Th><Th>Metode</Th><Th right>Total</Th><Th>Status</Th></tr>
                  </thead>
                  <tbody>
                    {data.transactions.slice(0, 10).map((t) => (
                      <tr key={t.id}>
                        <Td mono><span className="font-medium text-fg">{fmtInv(t.id)}</span></Td>
                        <Td mono>{fmtDate(t.date)}</Td>
                        <Td>{t.cashier}</Td>
                        <Td>{t.method}</Td>
                        <Td right><span className="font-medium text-fg">{fmtRp(t.total)}</span></Td>
                        <Td><StatusPill status={t.status} /></Td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
            {data.transactions.length > 0 && (
              <CardGrid>
                {data.transactions.slice(0, 10).map((t) => (
                  <DataCard key={t.id}>
                    <div className="flex items-center justify-between gap-2">
                      <p className="min-w-0 truncate font-mono text-[13px] font-medium text-fg">{fmtInv(t.id)}</p>
                      <StatusPill status={t.status} />
                    </div>
                    <p className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-fg">{fmtRp(t.total)}</p>
                    <p className="mt-1 text-xs leading-snug text-fog">{fmtDate(t.date)} · {t.cashier} · {t.method}</p>
                  </DataCard>
                ))}
              </CardGrid>
            )}
          </ChartCard>

          <ChartCard
            title="Produk Terlaris"
            sub="Top 5 jumlah terjual"
            action={<button onClick={() => setTab('products')} className="shrink-0 text-[13px] font-medium text-jet hover:underline">Lihat semua →</button>}
          >
            {topProducts.length === 0 ? (
              <p className="py-12 text-center text-sm text-muted-foreground">Tidak ada data.</p>
            ) : (
              <div className="space-y-4">
                {topProducts.map((p, i) => (
                  <div key={p.product_id} className="flex items-baseline gap-3">
                    <span className="w-6 shrink-0 font-mono text-xs text-muted-foreground">{String(i + 1).padStart(2, '0')}</span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium">{p.name}</p>
                      <p className="mt-0.5 text-xs tabular-nums text-muted-foreground">{p.qty} terjual</p>
                    </div>
                    <span className="shrink-0 text-sm font-medium tabular-nums">{fmtRp(p.revenue)}</span>
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
