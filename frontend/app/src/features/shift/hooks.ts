// Hooks TanStack Query fitur shift.
// Aturan transisi (sama seperti fitur promo/pos): onSuccess mutasi
// tutup paksa menginvalidasi DUA sistem — cache TanStack Query
// (queryKey shift) DAN cache lama shared/cache (clearCache,
// satu-satunya mekanisme invalidasi yang diekspor cache lama) — agar
// halaman lain yang masih memakai useCache tetap konsisten setelah
// shift berubah.
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clearCache } from '../../shared/cache'
import { useDB } from '../../shared/session/store'
import { outletQuery } from '../../shared/outlet'
import {
  apiForceCloseShift, apiListShifts,
  type Shift,
} from './api'

export const shiftKeys = {
  all: ['shift'] as const,
  list: (who: string, outletKey: string, rangeKey: string) =>
    ['shift', 'list', who, outletKey, rangeKey] as const,
}

function useInvalidateShift() {
  const queryClient = useQueryClient()
  return () => {
    void queryClient.invalidateQueries({ queryKey: shiftKeys.all })
    clearCache()
  }
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

// Bentuk hasil disamakan dengan useCache lama (data/err/loading/reload)
// agar halaman tidak berubah perilaku maupun tampilannya.
export function useShifts(selectedOutlet: string | undefined, from?: string, to?: string): {
  data: Shift[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const query = useQuery({
    queryKey: shiftKeys.list(who, selectedOutlet ?? 'all', from && to ? `${from}..${to}` : 'all'),
    queryFn: () => apiListShifts({ ...outletQuery(selectedOutlet), ...(from && to ? { from, to } : {}) }),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat daftar shift.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useForceCloseShift() {
  const invalidate = useInvalidateShift()
  return useMutation({
    mutationFn: ({ id, body }: { id: number; body: { closingCash?: number } }) => apiForceCloseShift(id, body),
    onSuccess: invalidate,
  })
}
