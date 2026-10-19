// Hooks TanStack Query fitur promo.
// Aturan transisi: onSuccess setiap mutasi menginvalidasi DUA sistem —
// cache TanStack Query (queryKey promo) DAN cache lama shared/cache
// (clearCache, satu-satunya mekanisme invalidasi yang diekspor cache
// lama) — agar halaman lain yang masih memakai useCache tetap konsisten
// setelah promo berubah.
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clearCache } from '../../shared/cache'
import { useDB } from '../../shared/session/store'
import {
  apiCreatePromo, apiDeletePromo, apiListPromos, apiSetPromoActive, apiUpdatePromo,
  type Promo, type PromoInput,
} from './api'

const promoKeys = {
  all: ['promo'] as const,
  list: (who: string, storeId?: string, outletId?: string) =>
    ['promo', 'list', who, storeId ?? '', outletId ?? ''] as const,
}

function useInvalidatePromo() {
  const queryClient = useQueryClient()
  return () => {
    void queryClient.invalidateQueries({ queryKey: promoKeys.all })
    clearCache()
  }
}

// Bentuk hasil disamakan dengan useCache lama (data/err/loading/reload)
// agar halaman tidak berubah perilaku maupun tampilannya.
export function usePromos(): {
  data: Promo[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: promoKeys.list(who, session?.storeId, session?.outletId),
    queryFn: apiListPromos,
  })
  return {
    data: query.data ?? null,
    err: query.error ? (query.error instanceof Error ? query.error.message : 'Gagal memuat promo.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useCreatePromo() {
  const invalidate = useInvalidatePromo()
  return useMutation({
    mutationFn: (body: PromoInput) => apiCreatePromo(body),
    onSuccess: invalidate,
  })
}

export function useUpdatePromo() {
  const invalidate = useInvalidatePromo()
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: PromoInput }) => apiUpdatePromo(id, body),
    onSuccess: invalidate,
  })
}

export function useDeletePromo() {
  const invalidate = useInvalidatePromo()
  return useMutation({
    mutationFn: (id: number) => apiDeletePromo(id),
    onSuccess: invalidate,
  })
}

export function useSetPromoActive() {
  const invalidate = useInvalidatePromo()
  return useMutation({
    mutationFn: ({ id, active }: { id: number; active: boolean }) => apiSetPromoActive(id, active),
    onSuccess: invalidate,
  })
}
