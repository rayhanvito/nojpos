import type { Movement, Product, ReportBundle, Trx } from '../../lib/api'
import { fmtRp } from '../../lib/store'

export type Period = 'today' | 'yesterday' | 'week' | 'month' | 'all'

export const PERIODS: { id: Period; label: string }[] = [
  { id: 'today', label: 'Hari ini' },
  { id: 'yesterday', label: 'Kemarin' },
  { id: 'week', label: 'Minggu ini' },
  { id: 'month', label: 'Bulan ini' },
  { id: 'all', label: 'Semua' },
]

export const payConfig: Record<string, string> = {
  Cash: 'var(--chart-1)',
  'Bank Transfer': 'var(--chart-2)',
  QRIS: 'var(--chart-3)',
  'E-Wallet': 'var(--chart-4)',
  Card: 'var(--chart-5)',
  Credit: 'var(--chart-6)',
}

export const salesChartConfig = {
  omzet: { label: 'Penjualan', color: 'var(--chart-omzet)' },
  trx: { label: 'Transaksi', color: 'var(--chart-1)' },
} as const
export const profitTrendConfig = {
  revenue: { label: 'Pendapatan', color: 'var(--chart-1)' },
  hpp: { label: 'HPP', color: 'var(--muted)' },
  profit: { label: 'Profit', color: 'var(--chart-omzet)' },
} as const
export const profitMarginConfig = {
  margin: { label: 'Margin', color: 'var(--chart-2)' },
} as const
export const stockConfig = { nilai: { label: 'Nilai stok', color: 'var(--chart-3)' } } as const
export const CAT_COLORS = ['var(--chart-1)', 'var(--chart-2)', 'var(--chart-3)', 'var(--chart-4)', 'var(--chart-5)']

export const MOV_LABEL: Record<Movement['type'], string> = {
  sale: 'Penjualan', refund: 'Refund', adjust: 'Penyesuaian', initial: 'Stok awal',
}



// Slot jam untuk periode 1 hari (Hari ini/Kemarin). Range saya yang tentukan:
// Pagi 05–11, Siang 11–15, Sore 15–18, Malam = sisanya.
export const TIME_SLOTS = [
  { label: 'Pagi', range: '05.00–10.59' },
  { label: 'Siang', range: '11.00–14.59' },
  { label: 'Sore', range: '15.00–18.59' },
  { label: 'Malam', range: '19.00–04.59' },
] as const

export function slotOf(hour: number): number {
  if (hour >= 5 && hour < 11) return 0
  if (hour >= 11 && hour < 15) return 1
  if (hour >= 15 && hour < 19) return 2
  return 3
}

export function localDayISO(offsetDays: number): string {
  const d = new Date()
  d.setDate(d.getDate() - offsetDays)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// Label tooltip: range jam bila datum slot waktu, kalau tidak ya label sumbu apa adanya.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function tipLabel(label: any, payload?: any): any {
  return payload?.[0]?.payload?.range ?? label
}

// Selisih % vs periode pembanding. Null bila tak terdefinisi (prev<=0) → fallback sub biasa.
export function delta(cur: number, prev: number | undefined, ref = 'kemarin'): string | null {
  if (prev === undefined || prev <= 0 || !Number.isFinite(cur)) return null
  const pct = Math.round(((cur - prev) / prev) * 100)
  if (pct === 0) return `±0% dari ${ref}`
  return `${pct > 0 ? '↑' : '↓'} ${Math.abs(pct)}% dari ${ref}`
}

export function catMapFrom(catData: Product[] | null | undefined) {
  const m = new Map<string, { category: string; active: boolean }>()
  for (const p of catData ?? []) m.set(p.id, { category: p.category_name ?? 'Tanpa kategori', active: p.active })
  return m
}

export function dailyFrom(data: ReportBundle | null | undefined, needHourly: boolean, hourly: Trx[] | null | undefined) {
  if (!data) return []
  if (needHourly) {
    if (!hourly) return []
    const slots = TIME_SLOTS.map((s) => ({ label: s.label, range: s.range, omzet: 0, trx: 0 }))
    for (const t of hourly) {
      const s = slots[slotOf(new Date(t.created_at).getHours())]
      s.omzet += t.total
      s.trx += 1
    }
    return slots
  }
  const map = new Map<string, { omzet: number; trx: number }>()
  for (const t of data.transactions) {
    const d = t.date.slice(0, 10)
    const e = map.get(d) ?? { omzet: 0, trx: 0 }
    e.omzet += t.total
    e.trx += 1
    map.set(d, e)
  }
  return [...map.entries()]
    .sort((a, b) => a[0].localeCompare(b[0]))
    .map(([date, v]) => ({ label: `${date.slice(8, 10)}/${date.slice(5, 7)}`, omzet: v.omzet, trx: v.trx }))
}

export function topProductsFrom(data: ReportBundle | null | undefined) {
  return data ? [...data.products].sort((a, b) => b.qty - a.qty).slice(0, 5) : []
}

export function prodSortedFrom(data: ReportBundle | null | undefined) {
  return data ? [...data.products].sort((a, b) => b.qty - a.qty) : []
}

export function lowProductsFrom(data: ReportBundle | null | undefined) {
  return data ? [...data.products].sort((a, b) => a.qty - b.qty).slice(0, 5) : []
}

export function prodProfitFrom(data: ReportBundle | null | undefined) {
  return data ? [...data.products].sort((a, b) => b.profit - a.profit) : []
}

export function prodPageData(prodSorted: ReportBundle['products'], prodAll: boolean, prodPage: number, pageSize = 10) {
  const prodPages = Math.max(1, Math.ceil(prodSorted.length / pageSize))
  const prodPageSafe = Math.min(prodPage, prodPages - 1)
  const prodShown = prodAll ? prodSorted.slice(prodPageSafe * pageSize, prodPageSafe * pageSize + pageSize) : prodSorted.slice(0, 5)
  return { prodPages, prodPageSafe, prodShown }
}

export function profitPageData(prodProfit: ReportBundle['products'], profitPage: number, pageSize = 10) {
  const profitPages = Math.max(1, Math.ceil(prodProfit.length / pageSize))
  const profitPageSafe = Math.min(profitPage, profitPages - 1)
  return { profitPages, profitPageSafe }
}

export function profitDailyFrom(data: ReportBundle | null | undefined, needHourly: boolean, hourly: Trx[] | null | undefined) {
  if (!data) return []
  if (needHourly) {
    if (!hourly) return []
    const slots = TIME_SLOTS.map((s) => ({ label: s.label, range: s.range, revenue: 0, hpp: 0, profit: 0, margin: 0 }))
    for (const t of hourly) {
      const s = slots[slotOf(new Date(t.created_at).getHours())]
      const h = t.items.reduce((n, i) => n + i.buy_price * i.qty, 0)
      s.revenue += t.total
      s.hpp += h
      s.profit += t.total - h
    }
    for (const s of slots) s.margin = s.revenue > 0 ? Math.round((s.profit / s.revenue) * 100) : 0
    return slots
  }
  const map = new Map<string, { revenue: number; hpp: number; profit: number }>()
  for (const t of data.transactions) {
    const d = t.date.slice(0, 10)
    const e = map.get(d) ?? { revenue: 0, hpp: 0, profit: 0 }
    e.revenue += t.total
    e.hpp += t.hpp
    e.profit += t.profit
    map.set(d, e)
  }
  return [...map.entries()]
    .sort((a, b) => a[0].localeCompare(b[0]))
    .map(([date, v]) => ({
      label: `${date.slice(8, 10)}/${date.slice(5, 7)}`,
      ...v,
      margin: v.revenue > 0 ? Math.round((v.profit / v.revenue) * 100) : 0,
    }))
}

// Insight valid dari data existing (maks 5, tanpa klaim buatan).
export function profitInsightsFrom(data: ReportBundle | null | undefined, period: Period, prev: ReportBundle | null | undefined, catMap: Map<string, { category: string; active: boolean }>, prodProfit: ReportBundle['products'], banding: string): string[] {
  if (!data || data.transactions.length === 0) return []
  const out: string[] = []
  const omzet = data.summary.omzet
  if (period === 'today' && prev && prev.summary.omzet > 0 && omzet > 0) {
    const cur = Math.round((data.summary.gross_profit / omzet) * 100)
    const pr = Math.round((prev.summary.gross_profit / prev.summary.omzet) * 100)
    const d = cur - pr
    out.push(`Margin profit hari ini ${cur}%, ${d === 0 ? `sama seperti ${banding}` : d > 0 ? `naik ${d}% dibanding ${banding}` : `turun ${Math.abs(d)}% dibanding ${banding}`}.`)
  }
  if (catMap.size > 0 && data.summary.gross_profit > 0) {
    const byCat = new Map<string, number>()
    for (const p of data.products) {
      const c = catMap.get(p.product_id)?.category ?? 'Tanpa kategori'
      byCat.set(c, (byCat.get(c) ?? 0) + p.profit)
    }
    const top = [...byCat.entries()].sort((a, b) => b[1] - a[1])[0]
    if (top) out.push(`Kategori ${top[0]} menyumbang ${Math.round((top[1] / data.summary.gross_profit) * 100)}% dari total profit.`)
  }
  const star = prodProfit[0]
  if (star && star.profit > 0) out.push(`${star.name} menjadi produk dengan kontribusi profit terbesar (${fmtRp(star.profit)}).`)
  const hppTotal = data.transactions.reduce((n, t) => n + t.hpp, 0)
  if (omzet > 0 && hppTotal > 0) out.push(`HPP periode ini ${fmtRp(hppTotal)} (${Math.round((hppTotal / omzet) * 100)}% dari omzet).`)
  const fat = [...data.products].filter((p) => p.revenue > 0).sort((a, b) => (b.profit / b.revenue) - (a.profit / a.revenue))[0]
  if (fat && fat.profit > 0) out.push(`${fat.name} margin tertinggi ${Math.round((fat.profit / fat.revenue) * 100)}% (${fmtRp(fat.profit)} profit).`)
  return out.slice(0, 5)
}

// Donut kategori: gabung agregat laporan + kategori katalog; 4 teratas + Lainnya.
export function catDonutFrom(data: ReportBundle | null | undefined, catMap: Map<string, { category: string; active: boolean }>) {
  if (!data) return []
  const map = new Map<string, number>()
  for (const p of data.products) {
    const c = catMap.get(p.product_id)?.category ?? 'Tanpa kategori'
    map.set(c, (map.get(c) ?? 0) + p.qty)
  }
  const rows = [...map.entries()].map(([name, qty]) => ({ name, qty })).sort((a, b) => b.qty - a.qty)
  if (rows.length <= 5) return rows
  const top = rows.slice(0, 4)
  const rest = rows.slice(4).reduce((n, r) => n + r.qty, 0)
  return [...top, { name: 'Lainnya', qty: rest }]
}

// Baris stok + kategori katalog (cocok SKU lalu nama).
export function stockRowsFrom(data: ReportBundle | null | undefined, catData: Product[] | null | undefined) {
  if (!data) return []
  const bySku = new Map((catData ?? []).map((p) => [p.sku, p.category_name ?? 'Tanpa kategori']))
  const byName = new Map((catData ?? []).map((p) => [p.name, p.category_name ?? 'Tanpa kategori']))
  return data.stock.map((s) => ({ ...s, category: bySku.get(s.sku) ?? byName.get(s.name) ?? 'Tanpa kategori' }))
}

export function stockStatusFrom(stockRows: ReturnType<typeof stockRowsFrom>) {
  let aman = 0, menipis = 0, habis = 0
  for (const s of stockRows) {
    if (s.stock === 0) habis++
    else if (s.stock <= 5) menipis++
    else aman++
  }
  return { aman, menipis, habis }
}

export function lowStockFrom(stockRows: ReturnType<typeof stockRowsFrom>) {
  return stockRows.filter((s) => s.stock <= 5).sort((a, b) => a.stock - b.stock)
}

export function catValueFrom(stockRows: ReturnType<typeof stockRowsFrom>) {
  const map = new Map<string, number>()
  for (const s of stockRows) map.set(s.category, (map.get(s.category) ?? 0) + s.stock_value)
  const rows = [...map.entries()].map(([name, nilai]) => ({ name, nilai })).sort((a, b) => b.nilai - a.nilai)
  if (rows.length <= 5) return rows
  const top = rows.slice(0, 4)
  const rest = rows.slice(4).reduce((n, r) => n + r.nilai, 0)
  return [...top, { name: 'Lainnya', nilai: rest }]
}

export function rowsSalesFrom(data: ReportBundle | null | undefined): string[][] {
  if (!data) return []
  return [
    ['tanggal', 'id', 'kasir', 'metode', 'total'],
    ...data.transactions.map((t) => [t.date, t.id, t.cashier, t.method, String(t.total)]),
  ]
}

export function rowsProductsFrom(data: ReportBundle | null | undefined): string[][] {
  if (!data) return []
  return [
    ['produk', 'sku', 'qty_terjual', 'pendapatan', 'profit'],
    ...data.products.map((p) => [p.name, p.sku, String(p.qty), String(p.revenue), String(p.profit)]),
  ]
}

export function rowsStockFrom(data: ReportBundle | null | undefined): string[][] {
  if (!data) return []
  return [
    ['produk', 'sku', 'stok', 'harga_beli', 'harga_jual', 'nilai_stok'],
    ...data.stock.map((s) => [s.name, s.sku, String(s.stock), String(s.buy_price), String(s.sell_price), String(s.stock_value)]),
  ]
}

export function rowsProfitFrom(data: ReportBundle | null | undefined): string[][] {
  if (!data) return []
  return [
    ['tanggal', 'id', 'kasir', 'total', 'hpp', 'profit'],
    ...data.transactions.map((t) => [t.date, t.id, t.cashier, String(t.total), String(t.hpp), String(t.profit)]),
  ]
}
