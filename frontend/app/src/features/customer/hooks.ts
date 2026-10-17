// Hooks TanStack Query fitur customer.
// Aturan transisi (sama seperti fitur promo/pos): onSuccess setiap mutasi
// menginvalidasi DUA sistem — cache TanStack Query (queryKey customer) DAN
// cache lama shared/cache (clearCache, satu-satunya mekanisme invalidasi
// yang diekspor cache lama) — agar halaman lain yang masih memakai
// useCache tetap konsisten setelah data pelanggan berubah.
import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clearCache } from '../../shared/cache'
import { useDB } from '../../shared/session/store'
import {
  apiCreateCustomer, apiDeleteCustomer, apiListCustomers, apiListReceivables, apiUpdateCustomer,
  type Customer, type CustomerInput, type Page, type Receivable,
} from './api'

export const customerKeys = {
  all: ['customer'] as const,
  list: (who: string, search: string, page: number, pageSize: number) =>
    ['customer', 'list', who, search, page, pageSize] as const,
  receivables: (who: string) => ['customer', 'receivables', who] as const,
}

function useInvalidateCustomer() {
  const queryClient = useQueryClient()
  return () => {
    void queryClient.invalidateQueries({ queryKey: customerKeys.all })
    clearCache()
  }
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

// Bentuk hasil disamakan dengan pola muat lama (data/err/loading/reload)
// agar halaman tidak berubah perilaku maupun tampilannya.
export function useCustomersList(search: string, page: number, pageSize: number): {
  data: Page<Customer> | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: customerKeys.list(who, search, page, pageSize),
    queryFn: () => apiListCustomers({ search, page: page + 1, limit: pageSize }),
    placeholderData: keepPreviousData,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat pelanggan.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useReceivables(): {
  data: Receivable[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: customerKeys.receivables(who),
    queryFn: apiListReceivables,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useCreateCustomer() {
  const invalidate = useInvalidateCustomer()
  return useMutation({
    mutationFn: (body: CustomerInput) => apiCreateCustomer(body),
    onSuccess: invalidate,
  })
}

export function useUpdateCustomer() {
  const invalidate = useInvalidateCustomer()
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: CustomerInput }) => apiUpdateCustomer(id, body),
    onSuccess: invalidate,
  })
}

export function useDeleteCustomer() {
  const invalidate = useInvalidateCustomer()
  return useMutation({
    mutationFn: (id: number) => apiDeleteCustomer(id),
    onSuccess: invalidate,
  })
}
