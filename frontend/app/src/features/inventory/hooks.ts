// Hooks TanStack Query fitur inventory.
// Aturan transisi (sama seperti fitur promo/pos): onSuccess setiap mutasi
// menginvalidasi DUA sistem — cache TanStack Query (queryKey inventory)
// DAN cache lama shared/cache (clearCache, satu-satunya mekanisme
// invalidasi yang diekspor cache lama) — agar halaman lain yang masih
// memakai useCache tetap konsisten setelah stok/opname berubah.
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clearCache } from '../../shared/cache'
import { useDB } from '../../shared/session/store'
import {
  apiAdjustStock, apiCancelStockOpname, apiCreateStockOpname, apiCreateTransfer, apiListCategories, apiListMovements, apiListTransfers,
  apiListProducts, apiListStockOpnames, apiPostStockOpname, apiUpdateStockOpnameItems, fetchAll,
  type Category, type Movement, type Page, type Product, type StockOpnameCountsInput, type StockOpnameSession, type StockTransfer, type StockTransferInput,
} from './api'

export const inventoryKeys = {
  all: ['inventory'] as const,
  products: (who: string) => ['inventory', 'products', who] as const,
  movements: (who: string, page: number, pageSize: number, type: string, q: string) =>
    ['inventory', 'movements', who, page, pageSize, type, q] as const,
  categories: (who: string) => ['inventory', 'categories', who] as const,
  opnames: (who: string) => ['inventory', 'opnames', who] as const,
  transfers: (who: string, page: number, pageSize: number) => ['inventory', 'transfers', who, page, pageSize] as const,
}

function useInvalidateInventory() {
  const queryClient = useQueryClient()
  return () => {
    void queryClient.invalidateQueries({ queryKey: inventoryKeys.all })
    clearCache()
  }
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

// Bentuk hasil disamakan dengan useCache lama (data/err/loading/reload)
// agar halaman tidak berubah perilaku maupun tampilannya.
export function useStockProducts(): {
  data: Product[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: inventoryKeys.products(who),
    queryFn: () => fetchAll<Product>((page) => apiListProducts({ page, limit: 200 })),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat stok.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useMovements(page: number, pageSize: number, type: string, q: string): {
  data: Page<Movement> | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: inventoryKeys.movements(who, page, pageSize, type, q),
    queryFn: () => apiListMovements({ page: page + 1, limit: pageSize, type: type || undefined, q: q.trim() || undefined }),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat pergerakan stok.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useStockCategories(): {
  data: Category[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: inventoryKeys.categories(who),
    queryFn: apiListCategories,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useStockOpnames(): {
  data: StockOpnameSession[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: inventoryKeys.opnames(who),
    queryFn: apiListStockOpnames,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat sesi opname.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useTransfers(page: number, pageSize: number): {
  data: Page<StockTransfer> | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: inventoryKeys.transfers(who, page, pageSize),
    queryFn: () => apiListTransfers({ page: page + 1, limit: pageSize }),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat transfer stok.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useAdjustStock() {
  const invalidate = useInvalidateInventory()
  return useMutation({
    mutationFn: ({ productId, direction, qty, reason }: { productId: string; direction: 'plus' | 'minus'; qty: number; reason: string }) =>
      apiAdjustStock(productId, direction, qty, reason),
    onSuccess: invalidate,
  })
}

export function useCreateStockOpname() {
  const invalidate = useInvalidateInventory()
  return useMutation({
    mutationFn: (note?: string) => apiCreateStockOpname(note),
    onSuccess: invalidate,
  })
}

export function useUpdateStockOpnameItems() {
  const invalidate = useInvalidateInventory()
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: StockOpnameCountsInput }) => apiUpdateStockOpnameItems(id, body),
    onSuccess: invalidate,
  })
}

export function usePostStockOpname() {
  const invalidate = useInvalidateInventory()
  return useMutation({
    mutationFn: (id: number) => apiPostStockOpname(id),
    onSuccess: invalidate,
  })
}

export function useCancelStockOpname() {
  const invalidate = useInvalidateInventory()
  return useMutation({
    mutationFn: (id: number) => apiCancelStockOpname(id),
    onSuccess: invalidate,
  })
}

export function useCreateTransfer() {
  const invalidate = useInvalidateInventory()
  return useMutation({
    mutationFn: (body: StockTransferInput) => apiCreateTransfer(body),
    onSuccess: invalidate,
  })
}
