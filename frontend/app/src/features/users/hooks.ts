// Hooks TanStack Query fitur users.
// Aturan transisi (sama seperti fitur promo/pos): onSuccess setiap mutasi
// menginvalidasi DUA sistem — cache TanStack Query (queryKey users) DAN
// cache lama shared/cache (clearCache, satu-satunya mekanisme invalidasi
// yang diekspor cache lama) — agar halaman lain yang masih memakai
// useCache tetap konsisten setelah data pengguna berubah.
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clearCache } from '../../shared/cache'
import { useDB } from '../../shared/session/store'
import {
  apiCreateUser, apiDeleteUser, apiGetReport, apiListActivity, apiListTransactions, apiListUsers, apiRenameUser, apiSetUserActive,
  type ActivityItem, type Page, type ReportBundle, type Trx, type User,
} from './api'

const usersKeys = {
  all: ['users'] as const,
  list: (who: string) => ['users', 'list', who] as const,
  todayReport: (who: string) => ['users', 'today-report', who] as const,
  allReport: (who: string) => ['users', 'all-report', who] as const,
  activity: (who: string) => ['users', 'activity', who] as const,
  recentTransactions: (who: string) => ['users', 'recent-transactions', who] as const,
}

function useInvalidateUsers() {
  const queryClient = useQueryClient()
  return () => {
    void queryClient.invalidateQueries({ queryKey: usersKeys.all })
    clearCache()
  }
}

function errText(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback
}

// Bentuk hasil disamakan dengan useCache lama (data/err/loading/reload)
// agar halaman tidak berubah perilaku maupun tampilannya.
export function useUsersList(): {
  data: User[] | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const query = useQuery({
    queryKey: usersKeys.list(`${session?.id}`),
    queryFn: apiListUsers,
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat pengguna.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useUsersTodayReport(): {
  data: ReportBundle | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const query = useQuery({
    queryKey: usersKeys.todayReport(`${session?.id}`),
    queryFn: () => apiGetReport('today'),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useUsersAllReport(): {
  data: ReportBundle | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const query = useQuery({
    queryKey: usersKeys.allReport(`${session?.id}`),
    queryFn: () => apiGetReport('all'),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useUsersActivity(): {
  data: Page<ActivityItem> | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const query = useQuery({
    queryKey: usersKeys.activity(`${session?.id}`),
    queryFn: () => apiListActivity({ page: 1, limit: 20 }),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useUsersRecentTransactions(): {
  data: Page<Trx> | null
  err: string
  loading: boolean
  reload: () => void
} {
  const { session } = useDB()
  const query = useQuery({
    queryKey: usersKeys.recentTransactions(`${session?.id}`),
    queryFn: () => apiListTransactions({ page: 1, limit: 200 }),
  })
  return {
    data: query.data ?? null,
    err: query.error ? errText(query.error, 'Gagal memuat data.') : '',
    loading: query.isLoading,
    reload: () => { void query.refetch() },
  }
}

export function useCreateUser() {
  const invalidate = useInvalidateUsers()
  return useMutation({
    mutationFn: (body: Parameters<typeof apiCreateUser>[0]) => apiCreateUser(body),
    onSuccess: invalidate,
  })
}

export function useSetUserActive() {
  const invalidate = useInvalidateUsers()
  return useMutation({
    mutationFn: ({ id, active }: { id: string; active: boolean }) => apiSetUserActive(id, active),
    onSuccess: invalidate,
  })
}

export function useDeleteUser() {
  const invalidate = useInvalidateUsers()
  return useMutation({
    mutationFn: (id: string) => apiDeleteUser(id),
    onSuccess: invalidate,
  })
}

export function useRenameUser() {
  const invalidate = useInvalidateUsers()
  return useMutation({
    mutationFn: ({ id, name, outletId }: { id: string; name: string; outletId?: string }) => apiRenameUser(id, name, outletId),
    onSuccess: invalidate,
  })
}
