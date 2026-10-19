// Hooks TanStack Query fitur auth.
// Mutasi di sini adalah mutasi sesi (masuk, daftar, ganti akun, keluar,
// simpan passcode): onSuccess setiap mutasi menginvalidasi DUA sistem —
// seluruh cache TanStack Query (pergantian sesi menyangkut semua fitur,
// bukan hanya auth) DAN cache lama shared/cache (clearCache, satu-satunya
// mekanisme invalidasi yang diekspor cache lama) — agar tidak ada data
// akun/toko sebelumnya yang terbawa ke sesi baru. Panggilan baca dan
// langkah OTP/verifikasi yang alurnya imperatif (cek kasir aktif, kirim
// & verifikasi OTP, reset kata sandi, muat awal pemilih akun) tetap
// panggilan langsung ke ./api dari halaman, persis seperti sebelumnya.
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { clearCache } from '../../shared/cache'
import {
  apiGoogleLogin, apiLogin, apiLogout, apiRegister, apiSetPasscode, apiSwitchAccount,
} from './api'


function useInvalidateAuth() {
  const queryClient = useQueryClient()
  return () => {
    void queryClient.invalidateQueries()
    clearCache()
  }
}

export function useLogin() {
  const invalidate = useInvalidateAuth()
  return useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) => apiLogin(email, password),
    onSuccess: invalidate,
  })
}

export function useGoogleLogin() {
  const invalidate = useInvalidateAuth()
  return useMutation({
    mutationFn: (credential: string) => apiGoogleLogin(credential),
    onSuccess: invalidate,
  })
}

export function useRegister() {
  const invalidate = useInvalidateAuth()
  return useMutation({
    mutationFn: ({ name, email, password, storeName }: { name: string; email: string; password: string; storeName: string }) =>
      apiRegister(name, email, password, storeName),
    onSuccess: invalidate,
  })
}

export function useSetPasscode() {
  const invalidate = useInvalidateAuth()
  return useMutation({
    mutationFn: ({ id, passcode, role }: { id: string; passcode: string; role?: string }) => apiSetPasscode(id, passcode, role),
    onSuccess: invalidate,
  })
}

export function useSwitchAccount() {
  const invalidate = useInvalidateAuth()
  return useMutation({
    mutationFn: ({ id, passcode, role }: { id: string; passcode?: string; role?: 'admin' | 'cashier' }) =>
      apiSwitchAccount(id, passcode, role),
    onSuccess: invalidate,
  })
}

export function useLogout() {
  const invalidate = useInvalidateAuth()
  return useMutation({
    mutationFn: () => apiLogout(),
    onSuccess: invalidate,
  })
}
