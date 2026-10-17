// Hooks TanStack Query fitur settings.
// Aturan transisi (sama seperti fitur promo/pos): onSuccess setiap mutasi
// menginvalidasi DUA sistem — cache TanStack Query (queryKey settings) DAN
// cache lama shared/cache (clearCache, satu-satunya mekanisme invalidasi
// yang diekspor cache lama) — agar halaman lain yang masih memakai
// useCache tetap konsisten setelah pengaturan berubah. Alur imperatif
// yang bukan data pengaturan (kirim OTP, reset kata sandi, keluar) tetap
// panggilan langsung ke ./api dari tab Akun, persis seperti sebelumnya.
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clearCache } from '../../shared/cache'
import { useDB } from '../../shared/session/store'
import {
  apiCreateOutlet, apiGetSettings, apiListTransactions, apiListUsers, apiSetPasscode, apiUpdateOutlet, apiUpdateSettings,
  type Outlet, type Page, type StoreSettings, type Trx, type User,
} from './api'

export const settingsKeys = {
  all: ['settings'] as const,
  main: (who: string) => ['settings', 'main', who] as const,
  users: (who: string) => ['settings', 'users', who] as const,
  transactions: (who: string, today: string) => ['settings', 'transactions', who, today] as const,
}

function useInvalidateSettings() {
  const queryClient = useQueryClient()
  return () => {
    void queryClient.invalidateQueries({ queryKey: settingsKeys.all })
    clearCache()
  }
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

// Bentuk hasil disamakan dengan pola muat lama (data/err/loading/reload)
// agar halaman tidak berubah perilaku maupun tampilannya.
export function useSettings(): {
  data: StoreSettings | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const query = useQuery({
    queryKey: settingsKeys.main(`${session?.id}`),
    queryFn: apiGetSettings,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat pengaturan.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useSettingsUsers(): {
  data: User[] | null
  err: string
  loading: boolean
  reload: () => void
  mutate: (d: User[]) => void
} {
  const { session } = useDB()
  const who = `${session?.id}`
  const queryClient = useQueryClient()
  const query = useQuery({
    queryKey: settingsKeys.users(who),
    queryFn: apiListUsers,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
    // Setara mutate() useCache lama: ganti data lokal seketika (dipakai
    // tab Akun setelah passcode disimpan/dihapus, sebelum refetch latar).
    mutate: (d: User[]) => { queryClient.setQueryData(settingsKeys.users(who), d) },
  }
}

export function useSettingsTransactions(today: string): {
  data: Page<Trx> | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const query = useQuery({
    queryKey: settingsKeys.transactions(`${session?.id}`, today),
    queryFn: () => apiListTransactions({ date: today, limit: 200, page: 1 }),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useUpdateSettings() {
  const invalidate = useInvalidateSettings()
  return useMutation({
    mutationFn: (body: StoreSettings) => apiUpdateSettings(body),
    onSuccess: invalidate,
  })
}

export function useSetPasscode() {
  const invalidate = useInvalidateSettings()
  return useMutation({
    mutationFn: ({ id, passcode, role }: { id: string; passcode: string; role?: string }) => apiSetPasscode(id, passcode, role),
    onSuccess: invalidate,
  })
}

export function useCreateOutlet() {
  const invalidate = useInvalidateSettings()
  return useMutation({
    mutationFn: (body: Pick<Outlet, 'name' | 'address'>) => apiCreateOutlet(body),
    onSuccess: invalidate,
  })
}

export function useUpdateOutlet() {
  const invalidate = useInvalidateSettings()
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: Pick<Outlet, 'name' | 'address'> }) => apiUpdateOutlet(id, body),
    onSuccess: invalidate,
  })
}
