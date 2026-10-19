// Hooks TanStack Query fitur report.
// Fitur ini tidak punya mutasi — semua datanya bacaan (bundle laporan,
// pajak, katalog, pergerakan, transaksi per jam, serta shift/pengguna/
// pembanding untuk tab Kasir); ekspor CSV/Excel tetap panggilan
// imperatif dari halaman ke fungsi yang sama persis di ./api. Karena
// tak ada mutasi, tak ada invalidasi di sini (pola sama seperti fitur
// dashboard): perubahan data terjadi lewat fitur lain yang mutasinya
// sudah menginvalidasi cache TanStack-nya sendiri sekaligus cache lama
// (clearCache). Bentuk hasil disamakan dengan useCache lama
// (data/err/loading/reload) agar halaman tidak berubah perilaku maupun
// tampilannya.
import { useQuery } from '@tanstack/react-query'
import { useDB } from '../../shared/session/store'
import { outletQuery } from '../../shared/outlet'
import { taxRowsForYear, type TaxReportRow } from '../../shared/tax-report'
import {
  apiGetReport, apiGetTaxReport, apiListMovements, apiListProducts, apiListShifts, apiListTransactions, apiListUsers, fetchAll,
  type Movement, type Product, type ReportBundle, type Shift, type TaxReport, type Trx, type User,
} from './api'

const reportKeys = {
  all: ['report'] as const,
  main: (who: string, periodKey: string, outletKey: string) =>
    ['report', 'main', who, periodKey, outletKey] as const,
  tax: (who: string, year: number, outletKey: string) =>
    ['report', 'tax', who, year, outletKey] as const,
  prev: (who: string, tab: string, period: string, outletKey: string) =>
    ['report', 'prev', who, tab, period, outletKey] as const,
  catalog: (who: string, tab: string) => ['report', 'catalog', who, tab] as const,
  movements: (who: string, tab: string) => ['report', 'movements', who, tab] as const,
  hourly: (who: string, tab: string, period: string, dayStr: string) =>
    ['report', 'hourly', who, tab, period, dayStr] as const,
  cashierShifts: (who: string, outletKey: string, rangeKey: string) =>
    ['report', 'cashier-shifts', who, outletKey, rangeKey] as const,
  cashierUsers: (who: string) => ['report', 'cashier-users', who] as const,
  cashierPrev: (who: string, outletKey: string, specKey: string) =>
    ['report', 'cashier-prev', who, outletKey, specKey] as const,
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

type Range = { from: string; to: string }

export function useReport(period: string, selectedOutlet: string | undefined, range: Range | null): {
  data: ReportBundle | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: reportKeys.main(who, range ? `range:${range.from}..${range.to}` : period, selectedOutlet ?? 'all'),
    queryFn: () => apiGetReport(period, undefined, selectedOutlet, range ?? undefined),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat laporan.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useTaxReport(year: number, months: string[], selectedOutlet: string | undefined): {
  data: TaxReportRow[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: reportKeys.tax(who, year, selectedOutlet ?? 'all'),
    queryFn: async () => {
      const reports: Partial<Record<string, TaxReport>> = Object.fromEntries(await Promise.all(
        months.map(async (month) => [month, await apiGetTaxReport(month, selectedOutlet)] as const),
      ))
      return taxRowsForYear(months, reports)
    },
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat laporan pajak.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useReportPrev(needPrev: boolean, tab: string, period: string, selectedOutlet: string | undefined): {
  data: ReportBundle | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: reportKeys.prev(who, tab, period, selectedOutlet ?? 'all'),
    queryFn: () => (needPrev ? apiGetReport('yesterday', undefined, selectedOutlet) : Promise.resolve(null)),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useReportCatalog(tab: string): {
  data: Product[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: reportKeys.catalog(who, tab),
    queryFn: () => (tab === 'products' || tab === 'stock' ? fetchAll<Product>((pg) => apiListProducts({ page: pg, limit: 200 })) : Promise.resolve(null)),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useReportMovements(tab: string): {
  data: Movement[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: reportKeys.movements(who, tab),
    queryFn: () => (tab === 'stock' ? apiListMovements({ limit: 5 }).then((r) => r.items) : Promise.resolve([])),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useReportHourly(needHourly: boolean, tab: string, period: string, dayStr: string): {
  data: Trx[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: reportKeys.hourly(who, tab, period, dayStr),
    queryFn: () => (needHourly
      ? fetchAll<Trx>((pg) => apiListTransactions({ date: dayStr, page: pg, limit: 200 }))
      : Promise.resolve(null)),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useCashierShifts(selectedOutlet: string | undefined, from?: string, to?: string): {
  data: Shift[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: reportKeys.cashierShifts(who, selectedOutlet ?? 'all', from && to ? `${from}..${to}` : 'all'),
    queryFn: () => apiListShifts({ ...outletQuery(selectedOutlet), ...(from && to ? { from, to } : {}) }),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat daftar shift.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useCashierUsers(): {
  data: User[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: reportKeys.cashierUsers(who),
    queryFn: () => apiListUsers(),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat daftar kasir.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export type CashierPrevSpec = { kind: 'yesterday' } | { kind: 'range'; from: string; to: string } | null

export function useCashierPrev(spec: CashierPrevSpec, selectedOutlet: string | undefined): {
  data: ReportBundle | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: reportKeys.cashierPrev(who, selectedOutlet ?? 'all', spec ? JSON.stringify(spec) : 'none'),
    queryFn: () => (spec === null
      ? Promise.resolve(null)
      : spec.kind === 'yesterday'
        ? apiGetReport('yesterday', undefined, selectedOutlet)
        : apiGetReport('today', undefined, selectedOutlet, { from: spec.from, to: spec.to })),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat periode pembanding.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}
