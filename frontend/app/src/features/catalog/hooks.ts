// Hooks TanStack Query fitur catalog.
// Aturan transisi (sama seperti fitur promo/pos): onSuccess setiap mutasi
// menginvalidasi DUA sistem — cache TanStack Query (queryKey catalog) DAN
// cache lama shared/cache (clearCache, satu-satunya mekanisme invalidasi
// yang diekspor cache lama) — agar halaman lain yang masih memakai
// useCache tetap konsisten setelah katalog berubah.
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clearCache } from '../../shared/cache'
import { useDB } from '../../shared/session/store'
import {
  apiCreateCategory, apiCreateProduct, apiCreateProductPriceTier, apiCreateVariant, apiCreateVariantPriceTier,
  apiDeleteCategory, apiDeleteProduct, apiDeleteProductPriceTier, apiDeleteVariant, apiDeleteVariantPriceTier, apiRenameCategory,
  apiGetSettings, apiImportCategories, apiImportProducts, apiListCategories, apiListProducts, apiSetProductActive,
  apiUpdateProduct, apiUpdateProductPriceTier, apiUpdateVariant, apiUpdateVariantPriceTier, fetchAll,
  type Category, type Product, type StoreSettings,
} from './api'

export const catalogKeys = {
  all: ['catalog'] as const,
  products: (who: string, search: string) => ['catalog', 'products', who, search] as const,
  categories: (who: string) => ['catalog', 'categories', who] as const,
  settings: (who: string) => ['catalog', 'settings', who] as const,
}

function useInvalidateCatalog() {
  const queryClient = useQueryClient()
  return () => {
    void queryClient.invalidateQueries({ queryKey: catalogKeys.all })
    clearCache()
  }
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

// Bentuk hasil disamakan dengan useCache lama (data/err/loading/reload)
// agar halaman tidak berubah perilaku maupun tampilannya.
export function useProducts(search: string): {
  data: Product[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: catalogKeys.products(who, search),
    queryFn: () => fetchAll<Product>((page) => apiListProducts({ q: search || undefined, page, limit: 200 })),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat produk.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useCategories(): {
  data: Category[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: catalogKeys.categories(who),
    queryFn: apiListCategories,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useCatalogSettings(): {
  data: StoreSettings | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const query = useQuery({
    queryKey: catalogKeys.settings(`${session?.id}`),
    queryFn: () => apiGetSettings(),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

type CreateProductBody = Parameters<typeof apiCreateProduct>[0]
type UpdateProductBody = Parameters<typeof apiUpdateProduct>[1]
type VariantBody = Parameters<typeof apiCreateVariant>[1]
type PriceTierBody = Parameters<typeof apiCreateProductPriceTier>[1]

export function useCreateProduct() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: (body: CreateProductBody) => apiCreateProduct(body),
    onSuccess: invalidate,
  })
}

export function useUpdateProduct() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: UpdateProductBody }) => apiUpdateProduct(id, body),
    onSuccess: invalidate,
  })
}

export function useDeleteProduct() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: (id: string) => apiDeleteProduct(id),
    onSuccess: invalidate,
  })
}

export function useSetProductActive() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ id, active }: { id: string; active: boolean }) => apiSetProductActive(id, active),
    onSuccess: invalidate,
  })
}

export function useCreateCategory() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: (name: string) => apiCreateCategory(name),
    onSuccess: invalidate,
  })
}

export function useDeleteCategory() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: (id: string) => apiDeleteCategory(id),
    onSuccess: invalidate,
  })
}

export function useRenameCategory() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) => apiRenameCategory(id, name),
    onSuccess: invalidate,
  })
}

export function useImportProducts() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: (file: File) => apiImportProducts(file),
    onSuccess: invalidate,
  })
}

export function useImportCategories() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: (file: File) => apiImportCategories(file),
    onSuccess: invalidate,
  })
}

export function useCreateVariant() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ productId, body }: { productId: string; body: VariantBody }) => apiCreateVariant(productId, body),
    onSuccess: invalidate,
  })
}

export function useUpdateVariant() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ productId, variantId, body }: { productId: string; variantId: number; body: VariantBody }) =>
      apiUpdateVariant(productId, variantId, body),
    onSuccess: invalidate,
  })
}

export function useDeleteVariant() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ productId, variantId }: { productId: string; variantId: number }) => apiDeleteVariant(productId, variantId),
    onSuccess: invalidate,
  })
}

export function useCreateProductPriceTier() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ productId, body }: { productId: string; body: PriceTierBody }) => apiCreateProductPriceTier(productId, body),
    onSuccess: invalidate,
  })
}

export function useUpdateProductPriceTier() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ productId, tierId, body }: { productId: string; tierId: number; body: PriceTierBody }) =>
      apiUpdateProductPriceTier(productId, tierId, body),
    onSuccess: invalidate,
  })
}

export function useDeleteProductPriceTier() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ productId, tierId }: { productId: string; tierId: number }) => apiDeleteProductPriceTier(productId, tierId),
    onSuccess: invalidate,
  })
}

export function useCreateVariantPriceTier() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ productId, variantId, body }: { productId: string; variantId: number; body: PriceTierBody }) =>
      apiCreateVariantPriceTier(productId, variantId, body),
    onSuccess: invalidate,
  })
}

export function useUpdateVariantPriceTier() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ productId, variantId, tierId, body }: { productId: string; variantId: number; tierId: number; body: PriceTierBody }) =>
      apiUpdateVariantPriceTier(productId, variantId, tierId, body),
    onSuccess: invalidate,
  })
}

export function useDeleteVariantPriceTier() {
  const invalidate = useInvalidateCatalog()
  return useMutation({
    mutationFn: ({ productId, variantId, tierId }: { productId: string; variantId: number; tierId: number }) =>
      apiDeleteVariantPriceTier(productId, variantId, tierId),
    onSuccess: invalidate,
  })
}
