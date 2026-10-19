// Hooks TanStack Query fitur finance.
// Aturan transisi (sama seperti fitur promo/pos): onSuccess setiap mutasi
// menginvalidasi DUA sistem — cache TanStack Query (queryKey finance)
// DAN cache lama shared/cache (clearCache, satu-satunya mekanisme
// invalidasi yang diekspor cache lama) — agar halaman lain yang masih
// memakai useCache tetap konsisten setelah hutang/piutang/pengeluaran
// berubah.
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clearCache } from '../../shared/cache'
import { useDB } from '../../shared/session/store'
import {
  apiAddPayablePayment, apiAddReceivablePayment, apiCreateExpense, apiDeleteExpense, apiListExpenses,
  apiListPayables, apiListReceivables, apiUpdateExpense, fetchAll,
  type DebtPaymentInput, type Expense, type ExpenseInput, type Payable, type Receivable,
} from './api'

const financeKeys = {
  all: ['finance'] as const,
  receivables: (who: string) => ['finance', 'receivables', who] as const,
  payables: (who: string) => ['finance', 'payables', who] as const,
  expenses: (who: string) => ['finance', 'expenses', who] as const,
}

function useInvalidateFinance() {
  const queryClient = useQueryClient()
  return () => {
    void queryClient.invalidateQueries({ queryKey: financeKeys.all })
    clearCache()
  }
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

// Bentuk hasil disamakan dengan useCache lama (data/err/loading/reload)
// agar halaman tidak berubah perilaku maupun tampilannya.
export function useReceivables(): {
  data: Receivable[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: financeKeys.receivables(who),
    queryFn: apiListReceivables,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat daftar piutang.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function usePayables(): {
  data: Payable[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: financeKeys.payables(who),
    queryFn: apiListPayables,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat daftar hutang.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useExpenses(): {
  data: Expense[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: financeKeys.expenses(who),
    queryFn: () => fetchAll<Expense>((page) => apiListExpenses({ page, limit: 200 })),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat pengeluaran.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useAddReceivablePayment() {
  const invalidate = useInvalidateFinance()
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: DebtPaymentInput }) => apiAddReceivablePayment(id, body),
    onSuccess: invalidate,
  })
}

export function useAddPayablePayment() {
  const invalidate = useInvalidateFinance()
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: DebtPaymentInput }) => apiAddPayablePayment(id, body),
    onSuccess: invalidate,
  })
}

export function useCreateExpense() {
  const invalidate = useInvalidateFinance()
  return useMutation({
    mutationFn: (body: ExpenseInput) => apiCreateExpense(body),
    onSuccess: invalidate,
  })
}

export function useUpdateExpense() {
  const invalidate = useInvalidateFinance()
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: ExpenseInput }) => apiUpdateExpense(id, body),
    onSuccess: invalidate,
  })
}

export function useDeleteExpense() {
  const invalidate = useInvalidateFinance()
  return useMutation({
    mutationFn: (id: number) => apiDeleteExpense(id),
    onSuccess: invalidate,
  })
}
