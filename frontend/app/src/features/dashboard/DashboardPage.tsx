// Dashboard — Operate surface. Compact POS overview: KPI → analytics → activity.
// Same Card/Button/StatusPill/Chart/Td/Th tokens as the rest of the app; no new identity.
import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import { BarChart3, Package, ReceiptText, ShoppingBag, TriangleAlert } from 'lucide-react'
import { fmtDate, fmtRp, fmtShort, fmtTime, useDB } from '../../shared/session/store'
import { useOutlets } from '../../hooks/use-outlets'
import { outletQuery } from '../../shared/outlet'
import { OutletFilter } from '../../components/OutletFilter'
import type { DashboardAdmin, Trx } from './api'
import { useCashierCharts, useDashboard, useRecentTransactions } from './hooks'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/primitives/card'
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/shared/ui/primitives/chart'
import { Bar, BarChart, CartesianGrid, Cell, LabelList, Pie, PieChart, XAxis, YAxis } from 'recharts'
import { Button } from '@/shared/ui/primitives/button'
import { Empty, EmptyContent, EmptyDescription, EmptyTitle } from '@/shared/ui/primitives/empty'
import { Skeleton } from '@/shared/ui/primitives/skeleton'
import { AlertStrip, CardGrid, DataCard, PeriodPicker, StatCard, StatusPill, Td, Th } from '../../shared/ui'

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']

// Preset periode standar — nilai sama persis dengan PERIODS di
// pages/reports/derive (dipakai Laporan). Fitur tidak boleh mengimpor
// pages (aturan dependency-cruiser), jadi konstanta kecil ini lokal.
const PERIODS = [
  { id: 'today', label: 'Hari ini' },
  { id: 'yesterday', label: 'Kemarin' },
  { id: 'week', label: 'Minggu ini' },
  { id: 'month', label: 'Bulan ini' },
  { id: 'all', label: 'Semua' },
]

const salesConfig = {
  omzet: { label: 'Penjualan', color: 'var(--chart-1)' },
} as const

const payConfig = {
  Cash: { label: 'Cash', color: 'var(--chart-1)' },
  'Bank Transfer': { label: 'Bank Transfer', color: 'var(--chart-2)' },
  QRIS: { label: 'QRIS', color: 'var(--chart-3)' },
  'E-Wallet': { label: 'E-Wallet', color: 'var(--chart-4)' },
  Card: { label: 'Card', color: 'var(--chart-5)' },
  Credit: { label: 'Kredit', color: 'var(--chart-6)' },
} as const

function dayLabel(iso: string): string {
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(iso)
  if (!m) return iso
  return `${Number(m[3])} ${MONTHS[Number(m[2]) - 1]}`
}

function isoLocal(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// Preset PeriodPicker standar -> rentang tanggal lokal. 'all' = rentang lebar
// eksplisit, karena /dashboard tanpa from/to berarti hari ini saja.
function presetToRangeDash(id: string): { from: string; to: string } {
  const now = new Date()
  const today = isoLocal(now)
  if (id === 'yesterday') {
    const y = new Date(now)
    y.setDate(y.getDate() - 1)
    const ys = isoLocal(y)
    return { from: ys, to: ys }
  }
  if (id === 'week') {
    const m = new Date(now)
    m.setDate(m.getDate() - ((m.getDay() + 6) % 7))
    return { from: isoLocal(m), to: today }
  }
  if (id === 'month') return { from: `${today.slice(0, 8)}01`, to: today }
  if (id === 'year') return { from: `${today.slice(0, 4)}-01-01`, to: `${today.slice(0, 4)}-12-31` }
  if (id === 'all') return { from: '2019-01-01', to: today }
  return { from: today, to: today }
}

export default function Dashboard() {
  const db = useDB()
  const s = db.session!
  // Animasi chart hidup hanya saat mount; refresh data tak me-restartnya.
  const [animate, setAnimate] = useState(true)
  useEffect(() => {
    const t = setTimeout(() => setAnimate(false), 900)
    return () => clearTimeout(t)
  }, [])
  // Periode standar (sama seperti Transaksi/Laporan): preset + rentang kustom.
  const [period, setPeriod] = useState('today')
  const [range, setRange] = useState<{ from: string; to: string } | null>(null)
  const eff = range ?? presetToRangeDash(period)
  const pickPreset = (id: string) => {
    if (id === 'year') {
      const y = new Date().getFullYear()
      setRange({ from: `${y}-01-01`, to: `${y}-12-31` })
      setPeriod('custom')
    } else {
      setPeriod(id)
      setRange(null)
    }
  }
  const pickRange = (from: string, to: string) => { setRange({ from, to }); setPeriod('custom') }
  const [outletId, setOutletId] = useState('')
  const { outlets } = useOutlets()
  const selectedOutlet = outletQuery(outletId).outletId
  const dash = useDashboard(eff, selectedOutlet)
  const recent = useRecentTransactions(eff)
  const data = dash.data
  const err = dash.err
  const recentTrx = recent.data ? recent.data.items : recent.err ? [] : null
  // Analitik backend kadang tak dikirim untuk kasir (sales7/methods/top).
  // sales7 7 entri = backend implementasi; selain itu hitung di frontend
  // dari transaksi sendiri (kasir di-scope server; admin = se-toko).
  const adminPre = data as DashboardAdmin | undefined
  const chartsMode: 'loading' | 'be' | 'fb' =
    !data ? 'loading' : (adminPre?.sales7?.length === 7 ? 'be' : 'fb')
  const cashierCharts = useCashierCharts(eff, chartsMode)

  const isAdmin = s.role === 'admin'
  if (err && !data) return <AlertStrip>{err}</AlertStrip>
  if (!data || data.role !== s.role) return <DashboardSkeleton />

  const today = data.today
  const admin = data as DashboardAdmin
  const phrase = range ? 'periode ini' : period === 'yesterday' ? 'kemarin' : period === 'week' ? 'minggu ini' : period === 'month' ? 'bulan ini' : period === 'all' ? 'semua waktu' : 'hari ini'
  const cap = (s: string) => s.replace(/\b\w/g, (c) => c.toUpperCase())
  const isTodayEff = eff.from === eff.to && eff.to === isoLocal(new Date())
  // sales7 backend = 7 tanggal berjalan; label dari tanggal asli (bukan index).
  // Chart pakai backend bila lengkap; bila tidak, fallback buildOwnCharts
  // dihitung dari transaksi sendiri.
  const fb = chartsMode === 'fb' ? cashierCharts.data : null
  const sales7src = chartsMode === 'be' ? admin.sales7 : (fb?.sales7 ?? [])
  const methodSrc = chartsMode === 'be' ? (admin.methods ?? []) : (fb?.methods ?? [])
  const topSrc = chartsMode === 'be' ? (admin.top_products ?? []) : (fb?.top_products ?? [])
  const sales7 = sales7src.map((d) => ({ label: dayLabel(d.date), omzet: d.omzet }))
  const payData = (Object.entries(payConfig) as [keyof typeof payConfig, (typeof payConfig)[keyof typeof payConfig]][])
        .map(([name, cfg]) => ({ name, total: methodSrc.find((m) => m.method === name)?.total ?? 0, fill: cfg.color }))
        .filter((d) => d.total > 0)
  const payTotal = payData.reduce((n, d) => n + d.total, 0)
  const topProducts = topSrc.slice(0, 5)

  const avgTrx = today.trx_count > 0 ? Math.round(today.omzet / today.trx_count) : 0
  const kpis: { label: string; value: string; icon: React.ComponentType<{ className?: string }>; href?: string }[] = isAdmin
    ? [
        { label: `Omzet ${cap(phrase)}`, value: fmtRp(today.omzet), icon: ShoppingBag, href: '/app/reports' },
        { label: `Transaksi ${cap(phrase)}`, value: String(today.trx_count), icon: ReceiptText, href: '/app/transactions' },
        { label: 'Produk Terjual', value: String(today.items_sold), icon: Package, href: '/app/reports' },
        {
          label: 'Stok Menipis',
          value: String(admin.today.low_stock ?? 0),
          icon: TriangleAlert,
          href: '/app/stock',
        },
      ]
    : [
        { label: `Total Penjualan ${cap(phrase)}`, value: fmtRp(today.omzet), icon: ShoppingBag },
        { label: `Transaksi ${cap(phrase)}`, value: String(today.trx_count), icon: ReceiptText },
        { label: 'Produk Terjual', value: String(today.items_sold), icon: Package },
        { label: 'Rata-rata Transaksi', value: fmtRp(avgTrx), icon: BarChart3 },
      ]

  if (!isAdmin) {
    return (
      <div className="space-y-4">
        <div className="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">Halo, {s.name}</h1>
            <p className="mt-1.5 text-sm text-muted-foreground">
              {today.trx_count > 0
                ? `${today.trx_count} transaksi ${phrase} dengan total penjualan ${fmtRp(today.omzet)}.`
                : `Belum ada transaksi ${phrase}.`}
            </p>
          </div>
          <div className="w-full sm:w-72">
            <PeriodPicker options={[...PERIODS, { id: 'year', label: 'Tahun ini' }]} period={period} range={range} onPreset={pickPreset} onRange={pickRange} label="Pilih periode dashboard" />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4 xl:grid-cols-4">
          {kpis.map((k) => (
            <StatCard key={k.label} label={k.label} value={k.value} icon={k.icon} />
          ))}
        </div>

        <div className="grid gap-4 lg:grid-cols-[2fr_1fr]">
          <Card>
            <CardHeader className="border-b border-border/70 pb-4">
              <div className="opc-panel-head">
                <span className="opc-panel-ico" aria-hidden="true"><BarChart3 /></span>
                <div>
                  <CardTitle>Penjualan</CardTitle>
                  <CardDescription>7 hari terakhir</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="pt-4">
              <ChartContainer config={salesConfig} className="h-60 w-full [&_:focus]:outline-none">
                <BarChart data={sales7} margin={{ top: 20, right: 12, bottom: 0, left: 0 }} barCategoryGap="25%">
                  <CartesianGrid vertical={false} strokeDasharray="4 4" stroke="color-mix(in oklch, var(--foreground) 24%, transparent)" />
                  <XAxis dataKey="label" interval="preserveStartEnd" minTickGap={16} tickLine={false} axisLine={false} tickMargin={8} tick={{ fontSize: 12 }} />
                  <YAxis tickLine={false} axisLine={false} width={48} domain={[0, 'auto']} tickFormatter={(v: number) => fmtShort(v)} tick={{ fontSize: 12 }} className="fill-muted-foreground" />
                  <ChartTooltip cursor={false} content={<ChartTooltipContent formatter={(v) => fmtRp(Number(v))} />} />
                  <Bar dataKey="omzet" fill="var(--color-omzet)" radius={[6, 6, 0, 0]} maxBarSize={46} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out">
                    <LabelList dataKey="omzet" position="top" formatter={(v) => fmtShort(Number(v))} fontSize={12} className="fill-foreground font-medium" />
                  </Bar>
                </BarChart>
              </ChartContainer>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <div className="opc-panel-head">
                <span className="opc-panel-ico" aria-hidden="true"><Package /></span>
                <div>
                  <CardTitle>Metode Pembayaran</CardTitle>
                  <CardDescription>{`Total pembayaran ${phrase}`}</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {payData.length === 0 ? (
                <p className="py-12 text-center text-sm text-muted-foreground">{`Belum ada transaksi ${phrase}.`}</p>
              ) : (
                <div className="opc-pay">
                  <ChartContainer config={{}} className="relative h-44 w-full max-w-48 shrink-0">
                    <PieChart>
                      <ChartTooltip content={<ChartTooltipContent formatter={(v) => fmtRp(Number(v))} hideLabel />} />
                      <Pie data={payData} dataKey="total" nameKey="name" innerRadius={52} outerRadius={74} paddingAngle={3} strokeWidth={0} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out">
                        {payData.map((d) => (
                          <Cell key={d.name} fill={d.fill} />
                        ))}
                      </Pie>
                    </PieChart>
                    <div className="pointer-events-none absolute inset-0 grid place-items-center">
                      <div className="text-center">
                        <p className="text-sm font-semibold tabular-nums tracking-tight">{fmtShort(payTotal)}</p>
                        <p className="text-[11px] text-muted-foreground">Total Penjualan</p>
                      </div>
                    </div>
                  </ChartContainer>
                  <div className="w-full min-w-0 flex-1 space-y-2.5 text-[13px]">
                    {payData.map((d) => (
                      <div key={d.name} className="flex items-center justify-between gap-2">
                        <span className="flex min-w-0 items-center gap-2">
                          <span className="size-2.5 shrink-0 rounded-full" style={{ background: d.fill }} />
                          <span className="leading-snug">{d.name}</span>
                        </span>
                        <span className="shrink-0 font-medium tabular-nums">
                          {payTotal > 0 ? Math.round((d.total / payTotal) * 100) : 0}%
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        <div className="grid gap-4 lg:grid-cols-[1.6fr_1fr]">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <div className="opc-panel-head">
                <span className="opc-panel-ico" aria-hidden="true"><ReceiptText /></span>
                <div>
                  <CardTitle>Transaksi Saya</CardTitle>
                  <CardDescription>{`Transaksi terbaru ${phrase}`}</CardDescription>
                </div>
              </div>
              <Link to="/app/transactions" className="opc-btn-sm">
                Lihat semua
              </Link>
            </CardHeader>
            <CardContent>
              {recentTrx === null ? (
                <RecentSkeleton rows={5} />
              ) : recentTrx.length === 0 ? (
                <div className="py-8 text-center">
                  <p className="text-sm text-muted-foreground">Belum ada transaksi. Mulai dari POS Kasir.</p>
                  <Button className="mt-4" render={<Link to="/app/pos" />}>Buka POS</Button>
                </div>
              ) : (
                <RecentTable items={recentTrx} showCashier={false} />
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="space-y-0">
              <div className="opc-panel-head">
                <span className="opc-panel-ico" aria-hidden="true"><Package /></span>
                <div>
                  <CardTitle>Produk Terlaris</CardTitle>
                  <CardDescription>{`Penjualan tertinggi ${phrase}`}</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {topProducts.length === 0 ? (
                <p className="py-12 text-center text-sm text-muted-foreground">{`Belum ada penjualan ${phrase}.`}</p>
              ) : (
                <div>
                  {topProducts.map((p, i) => (
                    <div key={p.product_id} className="opc-rank">
                      <span className="opc-rank-num">{String(i + 1).padStart(2, '0')}</span>
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium">{p.name}</p>
                        <p className="mt-0.5 text-xs tabular-nums text-muted-foreground">{p.qty} terjual</p>
                      </div>
                      <span className="shrink-0 text-sm font-medium tabular-nums">{fmtRp(p.revenue)}</span>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div>
        <nav className="opc-crumb" aria-label="Breadcrumb">
          <span>Home</span>
          <span aria-hidden="true">›</span>
          <span aria-current="page" className="text-foreground">Beranda</span>
        </nav>
        <div className="mt-1.5 flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between">
        <div>
          <h1>Beranda</h1>
          {!isTodayEff && <p className="mt-1 text-sm text-muted-foreground">{`Menampilkan data ${eff.from === eff.to ? fmtDate(eff.from) : `${fmtDate(eff.from)} – ${fmtDate(eff.to)}`}`}</p>}
        </div>
        <div className="flex w-full flex-col items-start gap-2 sm:w-auto sm:items-end">
          <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row sm:items-center">
            <OutletFilter outlets={outlets} value={outletId} onChange={setOutletId} />
            <div className="w-full sm:w-72">
            <PeriodPicker options={[...PERIODS, { id: 'year', label: 'Tahun ini' }]} period={period} range={range} onPreset={pickPreset} onRange={pickRange} label="Pilih periode dashboard" />
            </div>
          </div>
        </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 xl:grid-cols-4">
        {kpis.map((k) => (
          <StatCard key={k.label} label={k.label} value={k.value} icon={k.icon} href={k.href} />
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-[2fr_1fr]">
        <Card>
          <CardHeader className="border-b border-border/70 pb-4">
            <div className="opc-panel-head">
              <span className="opc-panel-ico" aria-hidden="true"><BarChart3 /></span>
              <div>
                <CardTitle>Penjualan</CardTitle>
                <CardDescription>7 hari terakhir</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="pt-4">
            <ChartContainer config={salesConfig} className="h-60 w-full [&_:focus]:outline-none">
              <BarChart data={sales7} margin={{ top: 20, right: 12, bottom: 0, left: 0 }} barCategoryGap="25%">
                <CartesianGrid vertical={false} strokeDasharray="4 4" stroke="color-mix(in oklch, var(--foreground) 24%, transparent)" />
                <XAxis dataKey="label" interval="preserveStartEnd" minTickGap={16} tickLine={false} axisLine={false} tickMargin={8} tick={{ fontSize: 12 }} />
                <YAxis tickLine={false} axisLine={false} width={48} domain={[0, 'auto']} tickFormatter={(v: number) => fmtShort(v)} tick={{ fontSize: 12 }} className="fill-muted-foreground" />
                <ChartTooltip cursor={false} content={<ChartTooltipContent formatter={(v) => fmtRp(Number(v))} />} />
                <Bar dataKey="omzet" fill="var(--color-omzet)" radius={[6, 6, 0, 0]} maxBarSize={46} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out">
                  <LabelList dataKey="omzet" position="top" formatter={(v) => fmtShort(Number(v))} fontSize={12} className="fill-foreground font-medium" />
                </Bar>
              </BarChart>
            </ChartContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="opc-panel-head">
              <span className="opc-panel-ico" aria-hidden="true"><Package /></span>
              <div>
                <CardTitle>Metode Pembayaran</CardTitle>
                <CardDescription>{`Total pembayaran ${phrase}`}</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {payData.length === 0 ? (
              <p className="py-12 text-center text-sm text-muted-foreground">{`Belum ada transaksi ${phrase}.`}</p>
            ) : (
              <div className="opc-pay">
                <ChartContainer config={{}} className="relative h-44 w-full max-w-48 shrink-0">
                  <PieChart>
                    <ChartTooltip content={<ChartTooltipContent formatter={(v) => fmtRp(Number(v))} hideLabel />} />
                    <Pie data={payData} dataKey="total" nameKey="name" innerRadius={52} outerRadius={74} paddingAngle={3} strokeWidth={0} isAnimationActive={animate} animationDuration={650} animationEasing="ease-out">
                      {payData.map((d) => (
                        <Cell key={d.name} fill={d.fill} />
                      ))}
                    </Pie>
                  </PieChart>
                  <div className="pointer-events-none absolute inset-0 grid place-items-center">
                    <div className="text-center">
                      <p className="text-sm font-semibold tabular-nums tracking-tight">{fmtShort(payTotal)}</p>
                      <p className="text-[11px] text-muted-foreground">Total Omzet</p>
                    </div>
                  </div>
                </ChartContainer>
                <div className="w-full min-w-0 flex-1 space-y-2.5 text-[13px]">
                  {payData.map((d) => (
                    <div key={d.name} className="flex items-center justify-between gap-2">
                      <span className="flex min-w-0 items-center gap-2">
                        <span className="size-2.5 shrink-0 rounded-full" style={{ background: d.fill }} />
                        <span className="leading-snug">{d.name}</span>
                      </span>
                      <span className="shrink-0 font-medium tabular-nums">
                        {payTotal > 0 ? Math.round((d.total / payTotal) * 100) : 0}%
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-[1.6fr_1fr]">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0">
            <div className="opc-panel-head">
              <span className="opc-panel-ico" aria-hidden="true"><ReceiptText /></span>
              <div>
                <CardTitle>Transaksi Terbaru</CardTitle>
                <CardDescription>{`Transaksi terbaru ${phrase}`}</CardDescription>
              </div>
            </div>
            <Link to="/app/transactions" className="opc-btn-sm">
              Lihat semua
            </Link>
          </CardHeader>
          <CardContent>
            {recentTrx === null ? (
              <RecentSkeleton rows={5} />
            ) : recentTrx.length === 0 ? (
              <p className="py-12 text-center text-sm text-muted-foreground">
                {`Belum ada transaksi. Transaksi yang masuk ${phrase} akan muncul di sini.`}
              </p>
            ) : (
              <RecentTable items={recentTrx} showCashier />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0">
            <div className="opc-panel-head">
              <span className="opc-panel-ico" aria-hidden="true"><Package /></span>
              <div>
                <CardTitle>Produk Terlaris</CardTitle>
                <CardDescription>{`Penjualan tertinggi ${phrase}`}</CardDescription>
              </div>
            </div>
            <Link to="/app/reports" className="opc-btn-sm">
              Lihat semua
            </Link>
          </CardHeader>
          <CardContent>
            {topProducts.length === 0 ? (
              <p className="py-12 text-center text-sm text-muted-foreground">{`Belum ada penjualan ${phrase}.`}</p>
            ) : (
              <div>
                {topProducts.map((p, i) => (
                  <div key={p.product_id} className="opc-rank">
                    <span className="opc-rank-num">{String(i + 1).padStart(2, '0')}</span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium">{p.name}</p>
                      <p className="mt-0.5 text-xs tabular-nums text-muted-foreground">{p.qty} terjual</p>
                    </div>
                    <span className="shrink-0 text-sm font-medium tabular-nums">{fmtRp(p.revenue)}</span>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {admin.today.trx_count === 0 && (
        <Empty>
          <EmptyContent>
            <EmptyTitle>Belum ada transaksi hari ini</EmptyTitle>
            <EmptyDescription>
              Buka menu POS Kasir untuk memulai transaksi pertama, atau cek produk Anda sudah siap dijual.
            </EmptyDescription>
            <Button render={<Link to="/app/pos" />}>Buka POS Kasir</Button>
          </EmptyContent>
        </Empty>
      )}
    </div>
  )
}

function RecentTable({ items, showCashier }: { items: Trx[]; showCashier: boolean }) {
  return (
    <>
    <div className="hidden overflow-x-auto min-[1280px]:block">
      <table className="w-full border-collapse">
        <thead>
          <tr>
            <Th>Waktu</Th><Th>No. Invoice</Th>{showCashier && <Th>Kasir</Th>}<Th>Metode</Th><Th right>Total</Th><Th>Status</Th>
          </tr>
        </thead>
        <tbody>
          {items.map((t) => (
            <tr key={t.id}>
              <Td mono>{fmtDate(t.created_at)} {fmtTime(t.created_at)}</Td>
              <Td mono>#TRX-{String(t.id).padStart(5, '0')}</Td>
              {showCashier && <Td>{t.cashier_name}</Td>}
              <Td>{t.method}</Td>
              <Td right><span className="font-medium text-fg">{fmtRp(t.total)}</span></Td>
              <Td><StatusPill status={t.status} /></Td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
    <CardGrid>
      {items.map((t) => (
        <DataCard key={t.id}>
          <div className="flex items-center justify-between gap-2">
            <p className="min-w-0 truncate font-mono text-[13px] font-medium text-fg">#TRX-{String(t.id).padStart(5, '0')}</p>
            <StatusPill status={t.status} />
          </div>
          <p className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-fg">{fmtRp(t.total)}</p>
          <p className="mt-1 text-xs leading-snug text-fog">
            {fmtDate(t.created_at)} {fmtTime(t.created_at)} · {t.method}{showCashier ? ` · ${t.cashier_name}` : ''}
          </p>
        </DataCard>
      ))}
    </CardGrid>
    </>
  )
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6" aria-busy="true" aria-label="Memuat dashboard">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div className="space-y-2">
          <Skeleton className="h-8 w-56" />
          <Skeleton className="h-4 w-72" />
        </div>
        <Skeleton className="h-4 w-24" />
      </div>

      <div className="grid grid-cols-2 gap-4 xl:grid-cols-4">
        {[0, 1, 2, 3].map((i) => (
          <Card key={i}>
            <CardContent className="space-y-3 p-5">
              <div className="flex items-center justify-between">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="size-9 rounded-lg" />
              </div>
              <Skeleton className="h-7 w-32" />
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-[2fr_1fr]">
          <Card>
            <CardHeader>
              <Skeleton className="h-5 w-32" />
              <Skeleton className="h-4 w-28" />
            </CardHeader>
            <CardContent>
              <div className="flex h-56 items-end gap-3" aria-hidden="true">
                {[42, 68, 52, 84, 60, 76, 48].map((h, i) => (
                  <Skeleton key={i} className="flex-1 rounded-b-none rounded-t-md" style={{ height: `${h}%` }} />
                ))}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <Skeleton className="h-5 w-40" />
              <Skeleton className="h-4 w-28" />
            </CardHeader>
            <CardContent>
              <Skeleton className="mx-auto size-36 rounded-full" />
              <div className="mt-4 space-y-2">
                {[0, 1, 2].map((i) => (
                  <div key={i} className="flex items-center justify-between">
                    <Skeleton className="h-4 w-24" />
                    <Skeleton className="h-4 w-10" />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

      <div className="grid gap-4 lg:grid-cols-[1.6fr_1fr]">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <div className="space-y-2">
                <Skeleton className="h-5 w-36" />
                <Skeleton className="h-4 w-28" />
              </div>
              <Skeleton className="h-8 w-24 rounded-md" />
            </CardHeader>
            <CardContent>
              <div className="space-y-2" aria-hidden="true">
                {[0, 1, 2, 3, 4].map((i) => (
                  <Skeleton key={i} className="h-11 w-full rounded-md" />
                ))}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <div className="space-y-2">
                <Skeleton className="h-5 w-32" />
                <Skeleton className="h-4 w-28" />
              </div>
              <Skeleton className="h-8 w-24 rounded-md" />
            </CardHeader>
            <CardContent>
              <div className="space-y-4" aria-hidden="true">
                {[0, 1, 2, 3].map((i) => (
                  <div key={i} className="flex items-baseline gap-3">
                    <Skeleton className="h-4 w-6" />
                    <div className="flex-1 space-y-1.5">
                      <Skeleton className="h-4 w-2/3" />
                      <Skeleton className="h-3 w-16" />
                    </div>
                    <Skeleton className="h-4 w-16" />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
    </div>
  )
}

function RecentSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-2" aria-hidden="true">
      {Array.from({ length: rows }).map((_, i) => (
        <Skeleton key={i} className="h-11 w-full rounded-md" />
      ))}
    </div>
  )
}

