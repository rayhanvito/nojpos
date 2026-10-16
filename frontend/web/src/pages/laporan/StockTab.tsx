import { OctagonX, Package, TriangleAlert, Wallet } from 'lucide-react'
import type { Movement, Product, ReportBundle } from '../../lib/api'
import { fmtDate, fmtRp, fmtShort, fmtTime } from '../../lib/store'
import { CardGrid, DataCard, Pill, StatCard, Td, Th } from '../../lib/ui'
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart'
import { Bar, BarChart, CartesianGrid, Cell, LabelList, Pie, PieChart, XAxis, YAxis } from 'recharts'
import { MOV_LABEL, stockConfig, tipLabel, type stockRowsFrom, type stockStatusFrom, type lowStockFrom, type catValueFrom, type catMapFrom } from './derive'
import { ChartCard } from './ChartCard'
import { Link } from 'react-router'

export function StockTab({ data, stockRows, stockStatus, lowStock, catValue, movements, catMap, movLoading, animate, catData }: {
  data: ReportBundle
  stockRows: ReturnType<typeof stockRowsFrom>
  stockStatus: ReturnType<typeof stockStatusFrom>
  lowStock: ReturnType<typeof lowStockFrom>
  catValue: ReturnType<typeof catValueFrom>
  movements: Movement[] | null | undefined
  catMap: ReturnType<typeof catMapFrom>
  movLoading: boolean
  animate: boolean
  catData: Product[] | null | undefined
}) {
  return (
    <>
      <div className="space-y-5">
        <div className="grid grid-cols-2 gap-4 xl:grid-cols-4">
          <StatCard label="Nilai Stok" value={fmtRp(data.stock.reduce((n, s) => n + s.stock_value, 0))} sub="total modal di gudang" icon={Wallet} />
          <StatCard
            label="Total Produk"
            value={catData ? String([...catMap.values()].filter((c) => c.active).length) : '…'}
            sub={catData ? `dari ${catMap.size} produk` : 'memuat katalog…'}
            icon={Package}
          />
          <StatCard label="Stok Menipis" value={String(stockStatus.menipis)} sub="Perlu perhatian" icon={TriangleAlert} />
          <StatCard label="Stok Habis" value={String(stockStatus.habis)} sub="Perlu segera restock" icon={OctagonX} />
        </div>

        <div className="grid gap-4 lg:grid-cols-[2fr_1fr]">
          <ChartCard title="Nilai Stok per Kategori" sub={catValue.length > 0 ? 'Kategori penyimpan nilai terbesar' : 'Belum ada data'}>
            {catValue.length === 0 ? (
              <p className="py-14 text-center text-sm text-muted-foreground">Tidak ada data.</p>
            ) : (
              <ChartContainer config={stockConfig} className="h-64 w-full [&_:focus]:outline-none">
                <BarChart data={catValue.map((c) => ({ name: c.name.length > 16 ? `${c.name.slice(0, 16)}…` : c.name, full: c.name, nilai: c.nilai }))} layout="vertical" margin={{ top: 0, right: 48, bottom: 0, left: 8 }}>
                  <CartesianGrid horizontal={false} strokeDasharray="4 4" stroke="color-mix(in oklch, var(--foreground) 18%, transparent)" />
                  <XAxis type="number" hide domain={[0, 'auto']} />
                  <YAxis type="category" dataKey="name" tickLine={false} axisLine={false} width={128} tick={{ fontSize: 11 }} />
                  <ChartTooltip cursor={false} content={<ChartTooltipContent labelFormatter={tipLabel} formatter={(v) => fmtRp(Number(v))} />} />
                  <Bar dataKey="nilai" fill="var(--color-nilai)" radius={[0, 6, 6, 0]} barSize={22} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out">
                    <LabelList dataKey="nilai" position="right" formatter={(v) => fmtShort(Number(v))} fontSize={11} className="fill-muted-foreground" />
                  </Bar>
                </BarChart>
              </ChartContainer>
            )}
          </ChartCard>

          <ChartCard title="Distribusi Status Stok" sub={stockRows.length > 0 ? 'Kondisi inventory' : 'Belum ada data'}>
            {stockRows.length === 0 ? (
              <p className="py-14 text-center text-sm text-muted-foreground">Tidak ada data.</p>
            ) : (
              <div className="space-y-4">
                <ChartContainer config={{}} className="relative mx-auto h-44 w-full">
                  <PieChart>
                    <ChartTooltip content={<ChartTooltipContent formatter={(v) => `${v} produk`} hideLabel />} />
                    <Pie
                      data={[
                        { name: 'Aman', total: stockStatus.aman, fill: 'var(--chart-2)' },
                        { name: 'Menipis', total: stockStatus.menipis, fill: 'var(--t-sunbeam)' },
                        { name: 'Habis', total: stockStatus.habis, fill: 'var(--chart-5)' },
                      ].filter((d) => d.total > 0)}
                      dataKey="total" nameKey="name" innerRadius={52} outerRadius={74} paddingAngle={3} strokeWidth={0}
                      isAnimationActive={animate} animationDuration={650} animationEasing="ease-out"
                    >
                      <Cell fill="var(--chart-2)" />
                      <Cell fill="var(--t-sunbeam)" />
                      <Cell fill="var(--chart-5)" />
                    </Pie>
                  </PieChart>
                  <div className="pointer-events-none absolute inset-0 grid place-items-center">
                    <div className="text-center">
                      <p className="text-sm font-semibold tabular-nums tracking-tight">{stockRows.length}</p>
                      <p className="text-[11px] text-muted-foreground">Total Produk</p>
                    </div>
                  </div>
                </ChartContainer>
                <div className="space-y-2">
                  {(() => {
                    const n = stockRows.length || 1
                    return [
                      { name: 'Aman', total: stockStatus.aman, fill: 'var(--chart-2)' },
                      { name: 'Menipis', total: stockStatus.menipis, fill: 'var(--t-sunbeam)' },
                      { name: 'Habis', total: stockStatus.habis, fill: 'var(--chart-5)' },
                    ].map((d) => (
                      <div key={d.name} className="flex items-center justify-between gap-3 text-sm">
                        <span className="flex min-w-0 items-center gap-2">
                          <span className="size-2.5 shrink-0 rounded-full" style={{ background: d.fill }} />
                          <span className="truncate">{d.name}</span>
                        </span>
                        <span className="shrink-0 tabular-nums">
                          <span className="font-medium">{d.total > 0 ? Math.round((d.total / n) * 100) : 0}%</span>
                          <span className="text-muted-foreground"> ({d.total})</span>
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
            title="Produk Stok Menipis"
            sub="Habis dulu, lalu stok terkecil"
            action={<Link to="/app/stok" className="shrink-0 text-[13px] font-medium text-jet hover:underline">Lihat semua →</Link>}
          >
            <div className="hidden overflow-x-auto min-[1280px]:block">
              {lowStock.length === 0 ? (
                <p className="py-10 text-center text-sm text-muted-foreground">Semua stok aman.</p>
              ) : (
                <table className="w-full border-collapse">
                  <thead>
                    <tr><Th>#</Th><Th>Produk</Th><Th>SKU</Th><Th right>Stok</Th><Th>Kategori</Th><Th>Status</Th></tr>
                  </thead>
                  <tbody>
                    {lowStock.map((s, i) => (
                      <tr key={`${s.sku}-${s.name}`}>
                        <Td mono>{String(i + 1).padStart(2, '0')}</Td>
                        <Td><span className="font-medium text-fg">{s.name}</span></Td>
                        <Td mono>{s.sku}</Td>
                        <Td right><span className="font-medium tabular-nums text-ember">{s.stock}</span></Td>
                        <Td>{s.category}</Td>
                        <Td><Pill tone="warn">{s.stock === 0 ? 'Habis' : 'Menipis'}</Pill></Td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
            {lowStock.length > 0 && (
              <CardGrid>
                {lowStock.map((s) => (
                  <DataCard key={`${s.sku}-${s.name}`}>
                    <div className="flex items-center justify-between gap-2">
                      <p className="min-w-0 truncate font-medium text-fg">{s.name}</p>
                      <Pill tone="warn">{s.stock === 0 ? 'Habis' : 'Menipis'}</Pill>
                    </div>
                    <p className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-ember">{s.stock} <span className="font-sans text-xs font-normal text-fog">stok tersisa</span></p>
                    <p className="mt-1 text-xs leading-snug text-fog">{s.sku} · {s.category}</p>
                  </DataCard>
                ))}
              </CardGrid>
            )}
          </ChartCard>

          <ChartCard
            title="Pergerakan Stok Terbaru"
            sub="5 terbaru · semua periode"
            action={<Link to="/app/stok?tab=riwayat" className="shrink-0 text-[13px] font-medium text-jet hover:underline">Lihat semua →</Link>}
          >
            {(movements ?? []).length === 0 ? (
              <p className="py-10 text-center text-sm text-fog">{!movLoading && movements ? 'Belum ada pergerakan stok.' : 'Memuat…'}</p>
            ) : (
              <div className="space-y-3">
                {(movements ?? []).slice(0, 5).map((m) => (
                  <div key={m.id} className="flex items-baseline gap-3">
                    <span className={`w-12 shrink-0 text-center font-mono text-sm font-medium tabular-nums ${m.qty > 0 ? 'text-sprout' : 'text-ember'}`}>
                      {m.qty > 0 ? `+${m.qty}` : m.qty}
                    </span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium">{m.product_name ?? '—'}</p>
                      <p className="mt-0.5 text-xs text-muted-foreground">
                        {fmtDate(m.created_at)} {fmtTime(m.created_at)} · {MOV_LABEL[m.type]} · {m.actor}
                      </p>
                    </div>
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
