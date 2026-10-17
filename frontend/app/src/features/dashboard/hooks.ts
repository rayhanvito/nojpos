// Hooks TanStack Query fitur dashboard.
// Aturan transisi (sama seperti fitur promo/pos): bentuk hasil disamakan
// dengan useCache lama (data/err/loading/reload) agar halaman tidak
// berubah perilaku maupun tampilannya. Fitur ini tidak punya mutasi —
// data Beranda berubah lewat fitur lain, yang mutasinya sudah
// menginvalidasi cache lama (clearCache) sekaligus query TanStack-nya.
import { useQuery } from '@tanstack/react-query'
import { useDB } from '../../shared/session/store'
import {
  apiGetDashboard, apiListTransactions,
  type DashboardAdmin, type DashboardCashier, type Page, type Trx,
} from './api'

export const dashboardKeys = {
  all: ['dashboard'] as const,
  main: (who: string, from: string, to: string, outletId?: string) =>
    ['dashboard', 'main', who, from, to, outletId ?? ''] as const,
  recent: (who: string, from: string, to: string) =>
    ['dashboard', 'recent', who, from, to] as const,
  charts: (who: string, from: string, to: string, mode: string) =>
    ['dashboard', 'charts', who, from, to, mode] as const,
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

type Range = { from: string; to: string }

// Kunci sesi pemilik data: queryKey memisahkan admin vs kasir.
export function useDashboard(range: Range, outletId?: string): {
  data: DashboardAdmin | DashboardCashier | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: dashboardKeys.main(who, range.from, range.to, outletId),
    queryFn: () => apiGetDashboard(undefined, outletId, range),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat dashboard.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useRecentTransactions(range: Range): {
  data: Page<Trx> | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: dashboardKeys.recent(who, range.from, range.to),
    queryFn: () => apiListTransactions({ limit: 5, from: range.from, to: range.to }),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export type CashierCharts = Awaited<ReturnType<typeof buildOwnCharts>>

export function useCashierCharts(range: Range, mode: 'loading' | 'be' | 'fb'): {
  data: CashierCharts | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: dashboardKeys.charts(who, range.from, range.to, mode),
    queryFn: () => (mode === 'fb' ? buildOwnCharts(range.from, range.to) : Promise.resolve(null)),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat grafik.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

function localDay(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// Grafik kasir dari transaksi sendiri (fallback bila backend tak kirim
// analitik). Batas 5 halaman x 200 = 1000 trx terakhir; cukup untuk UMKM.
async function buildOwnCharts(from: string, to: string) {
  const items: Trx[] = []
  for (let pg = 1; pg <= 5; pg++) {
    const d = await apiListTransactions({ page: pg, limit: 200 })
    items.push(...d.items)
    if (items.length >= d.total) break
  }
  const done = items.filter((t) => t.status === 'completed' || t.status === 'completed_offline_shortage')
  const base = new Date(`${to}T12:00:00`)
  const sales7 = Array.from({ length: 7 }, (_, i) => {
    const dd = new Date(base)
    dd.setDate(base.getDate() - (6 - i))
    const key = localDay(dd)
    return { date: key, omzet: done.filter((t) => localDay(new Date(t.created_at)) === key).reduce((n, t) => n + t.total, 0) }
  })
  const inRange = done.filter((t) => { const d = localDay(new Date(t.created_at)); return d >= from && d <= to })
  const byMethod = new Map<string, number>()
  for (const t of inRange) {
    byMethod.set(t.method, (byMethod.get(t.method) ?? 0) + t.total)
  }
  const top = new Map<string, { name: string; qty: number; revenue: number }>()
  for (const t of inRange) {
    for (const it of t.items) {
      const cur = top.get(it.product_id) ?? { name: it.name, qty: 0, revenue: 0 }
      cur.qty += it.qty
      cur.revenue += it.qty * it.price
      top.set(it.product_id, cur)
    }
  }
  return {
    sales7,
    methods: [...byMethod].map(([method, total]) => ({ method, total })),
    top_products: [...top].map(([product_id, v]) => ({ product_id, ...v })).sort((a, b) => b.qty - a.qty).slice(0, 5),
  }
}
