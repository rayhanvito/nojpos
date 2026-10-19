// Hooks TanStack Query fitur pos.
// Aturan transisi (sama seperti fitur promo): onSuccess mutasi checkout
// menginvalidasi DUA sistem — cache TanStack Query (queryKey pos) DAN
// cache lama shared/cache (clearCache, satu-satunya mekanisme invalidasi
// yang diekspor cache lama) — agar halaman lain yang masih memakai
// useCache tetap konsisten setelah stok/transaksi berubah.
//
// Kontrak offline TIDAK disentuh hooks ini: checkout lewat mutasi hanya
// untuk jalur online; kegagalan jaringan tetap ditangani halaman dengan
// enqueueOfflineCheckout + syncOfflineQueue dari shared/offline-queue
// persis seperti sebelumnya.
import { useEffect, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ApiError, type Customer, type Product, type PromoPreview, type StoreSettings } from '../../shared/api'
import { clearCache } from '../../shared/cache'
import { isNetworkFailure } from '../../shared/offline-queue'
import { useDB } from '../../shared/session/store'
import {
  apiCheckout, apiGetCashierShift, apiGetSettings, apiListCustomers, apiListProducts, apiPreviewPromo, fetchAll,
  type CheckoutBody,
} from './api'
import { cartKey as buildCartKey, checkoutItems, type CartLine } from './logic'

const posKeys = {
  all: ['pos'] as const,
  products: (who: string) => ['pos', 'products', who] as const,
  settings: (who: string) => ['pos', 'settings', who] as const,
  shift: (who: string) => ['pos', 'shift', who] as const,
  customers: (search: string) => ['pos', 'customers', search] as const,
  preview: (key: string, voucherCode: string | null, revision: number) =>
    ['pos', 'preview', key, voucherCode ?? '', revision] as const,
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

// Bentuk hasil disamakan dengan useCache lama (data/err/loading/reload)
// agar halaman tidak berubah perilaku maupun tampilannya.
export function usePosProducts(): {
  data: Product[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const query = useQuery({
    queryKey: posKeys.products(`${session?.id}`),
    queryFn: () => fetchAll<Product>((page) => apiListProducts({ active: true, page, limit: 200 })),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function usePosSettings(): {
  data: StoreSettings | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const query = useQuery({
    queryKey: posKeys.settings(`${session?.id}`),
    queryFn: () => apiGetSettings(),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

// Gerbang shift (2026-10-16): backend menolak checkout tanpa shift aktif.
// FE memeriksa lebih dulu agar kasir diarahkan membuka shift, bukan gagal
// di tengah pembayaran. 404 = pasti belum ada shift; kegagalan jaringan
// = status tak diketahui (data terakhir dipertahankan, null bila belum
// pernah diketahui) dan backend tetap penegak terakhir.
export function useShiftStatus(): {
  hasShift: boolean | null
  refreshShift: () => Promise<boolean | null>
} {
  const { session } = useDB()
  const query = useQuery({
    queryKey: posKeys.shift(`${session?.id}`),
    queryFn: async () => {
      try {
        const { shift } = await apiGetCashierShift()
        return !!shift
      } catch (e) {
        if (e instanceof ApiError && e.status === 404) return false
        throw e
      }
    },
    retry: false,
    // Khusus query ini: 'always' agar refetch saat offline langsung
    // dicoba dan gagal cepat, BUKAN tertahan oleh jeda offline
    // networkMode bawaan 'online'. Data cache terakhir tetap
    // dipertahankan TanStack Query saat fetch gagal, jadi kegagalan
    // cepat ini tidak menghapus status shift yang sudah diketahui.
    // Sengaja TIDAK global (query-client.ts tidak diubah): query lain
    // tidak punya jalur cache/offline seperti gerbang shift ini,
    // jadi mengubahnya global berisiko mengubah perilaku jeda/
    // lanjut-otomatis mereka tanpa kebutuhan yang terbukti.
    networkMode: 'always',
  })
  const refetchRef = useRef(query.refetch)
  refetchRef.current = query.refetch
  // Status bisa berubah di luar halaman ini (shift dibuka/ditutup lewat
  // popup Status di header) — segarkan saat jendela kembali fokus dan
  // berkala, agar pemberitahuan tidak basi.
  useEffect(() => {
    const onFocus = () => { void refetchRef.current() }
    window.addEventListener('pos:shift-changed', onFocus)
    window.addEventListener('focus', onFocus)
    document.addEventListener('visibilitychange', onFocus)
    const timer = setInterval(onFocus, 30000)
    return () => {
      window.removeEventListener('pos:shift-changed', onFocus)
      window.removeEventListener('focus', onFocus)
      document.removeEventListener('visibilitychange', onFocus)
      clearInterval(timer)
    }
  }, [])
  const refreshShift = async (): Promise<boolean | null> => {
    // Offline: jangan refetch sama sekali — kembalikan status dari
    // cache query. Memanggil refetch di sini adalah akar bug
    // checkout offline-penuh (dialog menggantung sampai koneksi
    // pulih, lalu transaksi terkirim sebagai penjualan online).
    if (typeof navigator !== 'undefined' && !navigator.onLine) {
      return query.data ?? null
    }
    const result = await query.refetch()
    if (result.data !== undefined) return result.data
    return query.data ?? null
  }
  return { hasShift: query.data ?? null, refreshShift }
}

// Pencarian pelanggan di modal pembayaran: debounce 250 ms persis
// seperti sebelumnya; hasil hanya berlaku untuk kata kunci terkini.
export function useCustomerLookup(searchRaw: string, active: boolean): {
  results: Customer[]
  error: string
  busy: boolean
} {
  const trimmed = searchRaw.trim()
  const [debounced, setDebounced] = useState('')
  useEffect(() => {
    if (!active || !trimmed) {
      setDebounced('')
      return
    }
    const timer = setTimeout(() => setDebounced(trimmed), 250)
    return () => { clearTimeout(timer) }
  }, [active, trimmed])
  const query = useQuery({
    queryKey: posKeys.customers(debounced),
    queryFn: () => apiListCustomers({ search: debounced, page: 1 }).then((res) => res.items ?? []),
    enabled: active && debounced.length > 0,
    retry: false,
  })
  const current = active && trimmed.length > 0 && debounced === trimmed
  return {
    results: current ? query.data ?? [] : [],
    error: current && query.error ? errText(query.error, 'Gagal mencari pelanggan.') : '',
    busy: active && trimmed.length > 0 && (!current || query.isFetching),
  }
}

export interface PromoPreviewState {
  promo: PromoPreview | null
  error: string
  loading: boolean
  networkFailure: boolean
}

// Pratinjau promo/voucher dari server untuk isi keranjang saat ini.
// Nilai potongan berasal dari server; POS hanya memakainya pada rincian.
export function usePromoPreview(cart: CartLine[], appliedVoucherCode: string | null, previewRevision: number): PromoPreviewState {
  const query = useQuery({
    queryKey: posKeys.preview(buildCartKey(cart), appliedVoucherCode, previewRevision),
    queryFn: () => apiPreviewPromo({
      items: checkoutItems(cart),
      ...(appliedVoucherCode ? { voucherCode: appliedVoucherCode } : {}),
    }),
    enabled: cart.length > 0,
    retry: false,
    refetchOnWindowFocus: false,
  })
  if (cart.length === 0) return { promo: null, error: '', loading: false, networkFailure: false }
  return {
    promo: query.data ?? null,
    error: query.error ? errText(query.error, 'Gagal memeriksa promo.') : '',
    loading: query.isLoading || query.isFetching,
    networkFailure: query.error ? isNetworkFailure(query.error, navigator.onLine) : false,
  }
}

export function useCheckout() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: CheckoutBody & { clientTxId: string }) => apiCheckout(body),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: posKeys.all })
      clearCache()
    },
  })
}
