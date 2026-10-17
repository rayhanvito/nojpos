// Hooks TanStack Query fitur transactions.
// Aturan transisi (sama seperti fitur promo/pos): onSuccess mutasi refund
// menginvalidasi DUA sistem — cache TanStack Query (queryKey transactions)
// DAN cache lama shared/cache (clearCache, satu-satunya mekanisme
// invalidasi yang diekspor cache lama) — agar halaman lain yang masih
// memakai useCache tetap konsisten setelah refund.
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clearCache } from '../../shared/cache'
import { useDB } from '../../shared/session/store'
import {
  apiGetDashboard, apiListTransactions, apiListUsers, apiRefundTransaction, fetchAll,
  type DashboardAdmin, type DashboardCashier, type Page, type Trx, type User,
} from './api'

export interface TrxFilter {
  q?: string
  method?: string
  status?: string
  cashierId?: string
  from?: string
  to?: string
}

export const transactionsKeys = {
  all: ['transactions'] as const,
  list: (who: string, f: TrxFilter, page: number, pageSize: number) =>
    ['transactions', 'list', who, f.q ?? '', f.method ?? '', f.status ?? '', f.cashierId ?? '', f.from ?? '', f.to ?? '', page, pageSize] as const,
  aggregate: (who: string, f: TrxFilter) =>
    ['transactions', 'aggregate', who, f.q ?? '', f.method ?? '', f.status ?? '', f.cashierId ?? '', f.from ?? '', f.to ?? ''] as const,
  users: (who: string) => ['transactions', 'users', who] as const,
  dashboard: (who: string) => ['transactions', 'dashboard', who] as const,
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

// Bentuk hasil disamakan dengan useCache lama (data/err/loading/reload)
// agar halaman tidak berubah perilaku maupun tampilannya.
export function useTransactionsList(f: TrxFilter, page: number, pageSize: number): {
  data: Page<Trx> | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: transactionsKeys.list(who, f, page, pageSize),
    queryFn: () => apiListTransactions({ ...f, page: page + 1, limit: pageSize }),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat transaksi.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

// Ringkasan mengikuti filter aktif (fetch-all, pola sama seperti Export).
export function useTransactionsAggregate(f: TrxFilter): {
  data: Trx[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: transactionsKeys.aggregate(who, f),
    queryFn: () => fetchAll<Trx>((pg) => apiListTransactions({ ...f, page: pg, limit: 200 })),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat ringkasan.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useTransactionUsers(): {
  data: User[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: transactionsKeys.users(who),
    queryFn: () => apiListUsers(),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat kasir.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useTransactionsDashboard(): {
  data: DashboardAdmin | DashboardCashier | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: transactionsKeys.dashboard(who),
    queryFn: () => apiGetDashboard(),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useRefundTransaction() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, items, reason }: { id: string; items: { productId: string; qty: number }[]; reason: string }) =>
      apiRefundTransaction(id, items, reason),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: transactionsKeys.all })
      clearCache()
    },
  })
}
