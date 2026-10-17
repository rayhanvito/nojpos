import type { Dispatch, SetStateAction } from 'react'
import { Banknote, Boxes, ChevronDown, ChevronUp, CircleCheck, Package } from 'lucide-react'
import type { Product, ReportBundle } from '../../../shared/api'
import { fmtRp, fmtShort } from '../../../shared/session/store'
import { CardGrid, DataCard, Pager, StatCard, Td, Th } from '../../../shared/ui'
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/shared/ui/primitives/chart'
import { Bar, BarChart, CartesianGrid, Cell, LabelList, Pie, PieChart, XAxis, YAxis } from 'recharts'
import { CAT_COLORS, type catDonutFrom, type prodSortedFrom, type topProductsFrom, type lowProductsFrom, type catMapFrom } from './derive'
import { ChartCard } from './ChartCard'

export function ProductsTab({ data, prodSorted, prodShown, prodAll, setProdAll, setProdPage, prodPages, prodPageSafe, prodPageSize, onProdPageSize, prodTotal, topProducts, lowProducts, catMap, animate, catDonut, catData }: {
  data: ReportBundle
  prodSorted: ReturnType<typeof prodSortedFrom>
  prodShown: ReturnType<typeof prodSortedFrom>
  prodAll: boolean
  setProdAll: Dispatch<SetStateAction<boolean>>
  setProdPage: Dispatch<SetStateAction<number>>
  prodPages: number
  prodPageSafe: number
  prodPageSize: number
  onProdPageSize: (n: number) => void
  prodTotal: number
  topProducts: ReturnType<typeof topProductsFrom>
  lowProducts: ReturnType<typeof lowProductsFrom>
  catMap: ReturnType<typeof catMapFrom>
  animate: boolean
  catDonut: ReturnType<typeof catDonutFrom>
  catData: Product[] | null | undefined
}) {
  return (
    <>
      <div className="space-y-5">
        <div className="grid grid-cols-2 gap-4 xl:grid-cols-4">
          <StatCard label="Qty Terjual" value={String(data.products.reduce((n, p) => n + p.qty, 0))} sub="total satuan" icon={Package} />
          <StatCard label="Produk Terjual" value={String(data.products.length)} sub="jenis berbeda terjual" icon={Boxes} />
          <StatCard label="Pendapatan Produk" value={fmtRp(data.products.reduce((n, p) => n + p.revenue, 0))} sub="dari semua produk" icon={Banknote} />
          <StatCard
            label="Produk Aktif"
            value={catData ? String([...catMap.values()].filter((c) => c.active).length) : '…'}
            sub={catData ? `dari ${catMap.size} produk` : 'memuat katalog…'}
            icon={CircleCheck}
          />
        </div>

        <div className="grid gap-4 lg:grid-cols-[2fr_1fr]">
          <ChartCard title="Penjualan Produk Teratas" sub="Top 5 berdasarkan jumlah terjual">
            {topProducts.length === 0 ? (
              <p className="py-12 text-center text-sm text-muted-foreground">Tidak ada data.</p>
            ) : (
              <ChartContainer config={{}} className="h-64 w-full [&_:focus]:outline-none">
                <BarChart data={topProducts.map((p) => ({ name: p.name.length > 16 ? `${p.name.slice(0, 16)}…` : p.name, full: p.name, qty: p.qty }))} layout="vertical" margin={{ top: 0, right: 48, bottom: 0, left: 8 }}>
                  <CartesianGrid horizontal={false} strokeDasharray="4 4" stroke="color-mix(in oklch, var(--foreground) 18%, transparent)" />
                  <XAxis type="number" hide domain={[0, 'auto']} />
                  <YAxis type="category" dataKey="name" tickLine={false} axisLine={false} width={128} tick={{ fontSize: 11 }} />
                  <ChartTooltip cursor={false} content={<ChartTooltipContent labelFormatter={(_, payload) => (payload?.[0]?.payload as { full?: string } | undefined)?.full ?? ''} formatter={(v) => `${v} terjual`} />} />
                  <Bar dataKey="qty" fill="var(--chart-2)" radius={[0, 6, 6, 0]} barSize={22} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out">
                    <LabelList dataKey="qty" position="right" fontSize={11} className="fill-muted-foreground" />
                  </Bar>
                </BarChart>
              </ChartContainer>
            )}
          </ChartCard>

          <ChartCard title="Kategori Terlaris" sub={catDonut.length > 0 ? 'Kontribusi unit per kategori' : 'Belum ada data'}>
            {catDonut.length === 0 ? (
              <p className="py-12 text-center text-sm text-muted-foreground">Tidak ada data.</p>
            ) : (
              <div className="space-y-4">
                <ChartContainer config={{}} className="relative mx-auto h-44 w-full">
                  <PieChart>
                    <ChartTooltip content={<ChartTooltipContent formatter={(v) => `${v} terjual`} hideLabel />} />
                    <Pie data={catDonut.map((c, i) => ({ ...c, fill: CAT_COLORS[i % CAT_COLORS.length] }))} dataKey="qty" nameKey="name" innerRadius={52} outerRadius={74} paddingAngle={3} strokeWidth={0} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out">
                      {catDonut.map((c, i) => (
                        <Cell key={c.name} fill={CAT_COLORS[i % CAT_COLORS.length]} />
                      ))}
                    </Pie>
                  </PieChart>
                  <div className="pointer-events-none absolute inset-0 grid place-items-center">
                    <div className="text-center">
                      <p className="text-sm font-semibold tabular-nums tracking-tight">{fmtShort(catDonut.reduce((n, c) => n + c.qty, 0))}</p>
                      <p className="text-[11px] text-muted-foreground">Unit Terjual</p>
                    </div>
                  </div>
                </ChartContainer>
                <div className="space-y-2">
                  {(() => {
                    const catTotal = catDonut.reduce((n, c) => n + c.qty, 0) || 1
                    return catDonut.map((c) => (
                      <div key={c.name} className="flex items-center justify-between gap-3 text-sm">
                        <span className="flex min-w-0 items-center gap-2">
                          <span className="size-2.5 shrink-0 rounded-full" style={{ background: CAT_COLORS[catDonut.indexOf(c) % CAT_COLORS.length] }} />
                          <span className="truncate">{c.name}</span>
                        </span>
                        <span className="shrink-0 tabular-nums">
                          <span className="font-medium">{c.qty}</span>
                          <span className="text-muted-foreground"> · {Math.round((c.qty / catTotal) * 100)}%</span>
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
            title="Produk Terlaris"
            sub={prodAll ? `Semua ${prodSorted.length} produk` : 'Top 5 jumlah terjual'}
          >
            <div className="hidden overflow-x-auto min-[1280px]:block">
              {prodSorted.length === 0 ? (
                <p className="py-10 text-center text-sm text-muted-foreground">Tidak ada data.</p>
              ) : (
                <table className="w-full border-collapse">
                  <thead>
                    <tr><Th>#</Th><Th>Produk</Th><Th>Kategori</Th><Th right>Qty Terjual</Th><Th right>Pendapatan</Th></tr>
                  </thead>
                  <tbody>
                    {prodShown.map((p, i) => (
                      <tr key={p.product_id}>
                        <Td mono>{String((prodAll ? prodPageSafe * prodPageSize : 0) + i + 1).padStart(2, '0')}</Td>
                        <Td><span className="font-medium text-fg">{p.name}</span></Td>
                        <Td>{catMap.get(p.product_id)?.category ?? 'Tanpa kategori'}</Td>
                        <Td right>{p.qty}</Td>
                        <Td right><span className="font-medium text-fg">{fmtRp(p.revenue)}</span></Td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
            {prodSorted.length > 0 && (
              <CardGrid>
                {prodShown.map((p, i) => (
                  <DataCard key={p.product_id}>
                    <div className="flex items-center justify-between gap-2">
                      <p className="min-w-0 truncate font-medium text-fg"><span className="mr-1.5 font-mono text-xs text-fog">{String((prodAll ? prodPageSafe * prodPageSize : 0) + i + 1).padStart(2, '0')}</span>{p.name}</p>
                    </div>
                    <p className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-fg">{fmtRp(p.revenue)}</p>
                    <p className="mt-1 text-xs leading-snug text-fog">{catMap.get(p.product_id)?.category ?? 'Tanpa kategori'} · Terjual {p.qty}</p>
                  </DataCard>
                ))}
              </CardGrid>
            )}
            {prodAll && <Pager page={prodPageSafe} total={prodPages} onChange={setProdPage} totalItems={prodTotal} pageSize={prodPageSize} onPageSizeChange={onProdPageSize} />}
            {prodSorted.length > 5 && (
              <button
                onClick={() => { setProdAll((v) => !v); setProdPage(0) }}
                aria-expanded={prodAll}
                className="mt-3 flex w-full items-center justify-center gap-1.5 rounded-lg border border-dove bg-paper py-2 text-[13px] font-medium transition hover:border-jet"
              >
                {prodAll ? 'Ringkas' : 'Selengkapnya'}
                {prodAll ? <ChevronUp className="size-4" /> : <ChevronDown className="size-4" />}
              </button>
            )}
          </ChartCard>

          <ChartCard title="Perlu Perhatian" sub="Penjualan terendah periode ini">
            {lowProducts.length === 0 ? (
              <p className="py-12 text-center text-sm text-muted-foreground">Tidak ada data.</p>
            ) : (
              <div className="space-y-4">
                {lowProducts.map((p) => (
                  <div key={p.product_id} className="flex items-baseline gap-3">
                    <span className="size-2 mt-1.5 shrink-0 rounded-full" style={{ background: 'var(--t-sunbeam)' }} aria-hidden="true" />
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium">{p.name}</p>
                      <p className="mt-0.5 text-xs tabular-nums text-muted-foreground">{p.qty} terjual · {fmtRp(p.revenue)}</p>
                    </div>
                  </div>
                ))}
                <p className="text-xs text-muted-foreground">Cek stok, harga, atau promosikan kembali produk ini.</p>
              </div>
            )}
          </ChartCard>
        </div>
      </div>
    </>
  )
}
