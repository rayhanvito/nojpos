// Hooks TanStack Query fitur purchasing.
// Aturan transisi (sama seperti fitur promo/pos): onSuccess setiap mutasi
// menginvalidasi DUA sistem — cache TanStack Query (queryKey purchasing)
// DAN cache lama shared/cache (clearCache, satu-satunya mekanisme
// invalidasi yang diekspor cache lama) — agar halaman lain yang masih
// memakai useCache tetap konsisten setelah PO/supplier berubah.
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clearCache } from '../../shared/cache'
import { useDB } from '../../shared/session/store'
import {
  apiCancelPurchaseOrder, apiCreatePurchaseOrder, apiCreateSupplier, apiDeletePurchaseOrder, apiDeleteSupplier,
  apiListPayables, apiListPurchaseOrders, apiListProducts, apiListSuppliers, apiOrderPurchaseOrder,
  apiReceivePurchaseOrder, apiUpdatePurchaseOrder, apiUpdateSupplier, fetchAll,
  type Payable, type Product, type PurchaseOrder, type PurchaseOrderInput, type PurchaseOrderReceiptInput, type Supplier, type SupplierInput,
} from './api'

export const purchasingKeys = {
  all: ['purchasing'] as const,
  orders: (who: string) => ['purchasing', 'orders', who] as const,
  suppliers: (who: string) => ['purchasing', 'suppliers', who] as const,
  products: (who: string) => ['purchasing', 'products', who] as const,
  payables: (who: string) => ['purchasing', 'payables', who] as const,
}

function useInvalidatePurchasing() {
  const queryClient = useQueryClient()
  return () => {
    void queryClient.invalidateQueries({ queryKey: purchasingKeys.all })
    clearCache()
  }
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

// Bentuk hasil disamakan dengan useCache/useFetch lama
// (data/err/loading/reload) agar halaman tidak berubah perilaku
// maupun tampilannya.
export function usePurchaseOrders(): {
  data: PurchaseOrder[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: purchasingKeys.orders(who),
    queryFn: apiListPurchaseOrders,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat PO.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function usePurchaseSuppliers(): {
  data: Supplier[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: purchasingKeys.suppliers(who),
    queryFn: apiListSuppliers,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat supplier.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function usePurchaseProducts(): {
  data: Product[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: purchasingKeys.products(who),
    queryFn: () => fetchAll<Product>((page) => apiListProducts({ page, limit: 200 })),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat produk.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useSuppliers(): {
  data: Supplier[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: purchasingKeys.suppliers(who),
    queryFn: apiListSuppliers,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Terjadi kesalahan.') : '',
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
    queryKey: purchasingKeys.payables(who),
    queryFn: apiListPayables,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Terjadi kesalahan.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useCreatePurchaseOrder() {
  const invalidate = useInvalidatePurchasing()
  return useMutation({
    mutationFn: (body: PurchaseOrderInput) => apiCreatePurchaseOrder(body),
    onSuccess: invalidate,
  })
}

export function useUpdatePurchaseOrder() {
  const invalidate = useInvalidatePurchasing()
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: PurchaseOrderInput }) => apiUpdatePurchaseOrder(id, body),
    onSuccess: invalidate,
  })
}

export function useDeletePurchaseOrder() {
  const invalidate = useInvalidatePurchasing()
  return useMutation({
    mutationFn: (id: number) => apiDeletePurchaseOrder(id),
    onSuccess: invalidate,
  })
}

export function useOrderPurchaseOrder() {
  const invalidate = useInvalidatePurchasing()
  return useMutation({
    mutationFn: (id: number) => apiOrderPurchaseOrder(id),
    onSuccess: invalidate,
  })
}

export function useReceivePurchaseOrder() {
  const invalidate = useInvalidatePurchasing()
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: PurchaseOrderReceiptInput }) => apiReceivePurchaseOrder(id, body),
    onSuccess: invalidate,
  })
}

export function useCancelPurchaseOrder() {
  const invalidate = useInvalidatePurchasing()
  return useMutation({
    mutationFn: (id: number) => apiCancelPurchaseOrder(id),
    onSuccess: invalidate,
  })
}

export function useCreateSupplier() {
  const invalidate = useInvalidatePurchasing()
  return useMutation({
    mutationFn: (body: SupplierInput) => apiCreateSupplier(body),
    onSuccess: invalidate,
  })
}

export function useUpdateSupplier() {
  const invalidate = useInvalidatePurchasing()
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: SupplierInput }) => apiUpdateSupplier(id, body),
    onSuccess: invalidate,
  })
}

export function useDeleteSupplier() {
  const invalidate = useInvalidatePurchasing()
  return useMutation({
    mutationFn: (id: number) => apiDeleteSupplier(id),
    onSuccess: invalidate,
  })
}
