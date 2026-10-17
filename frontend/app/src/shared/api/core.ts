// Inti klien API: alamat dasar, token sesi, refresh, dan fungsi request bersama.

import { useEffect, useState } from 'react'
import type { Page } from './types.ts'

export const BASE = import.meta.env?.VITE_API_URL ?? '/api/v1'
export const ACCESS_KEY = 'op_access'
export const REFRESH_KEY = 'op_refresh'
export const ACCOUNTS_KEY = 'op_accounts'

// Sesi per-tab (sessionStorage): dua tab/akun di browser yang sama tidak
// saling menimpa token (bug "pindah akun tiap refresh"). Token lama di
// localStorage dimigrasikan sekali lalu dihapus agar tak jadi sesi bayangan.
function migrateLegacySession() {
  try {
    if (sessionStorage.getItem(ACCESS_KEY) || sessionStorage.getItem(REFRESH_KEY)) return
    const access = localStorage.getItem(ACCESS_KEY)
    const refresh = localStorage.getItem(REFRESH_KEY)
    if (access) sessionStorage.setItem(ACCESS_KEY, access)
    if (refresh) sessionStorage.setItem(REFRESH_KEY, refresh)
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
    // Cache akun lama juga dibuang: sudah tak dibaca (kini per-tab)
    // dan jangan menyisakan daftar akun di penyimpanan global.
    localStorage.removeItem(ACCOUNTS_KEY)
  } catch {
    // storage diblokir (mode privat) — sesi tetap jalan tanpa persistensi
  }
}
migrateLegacySession()

export class ApiError extends Error {
  status: number
  code: string
  constructor(status: number, message: string, code = '') {
    super(message)
    this.status = status
    // Backend mengirim code uppercase (PASSCODE_REQUIRED/OTP_REQUIRED),
    // sedangkan pemanggil membandingkan lowercase. Normalkan sekali di sini.
    this.code = code.toLowerCase()
  }
}

export function getToken(): string | null {
  return sessionStorage.getItem(ACCESS_KEY)
}

export function hasToken(): boolean {
  return !!getToken()
}

export function clearTokens() {
  sessionStorage.removeItem(ACCESS_KEY)
  sessionStorage.removeItem(REFRESH_KEY)
  localStorage.removeItem(ACCESS_KEY)
  localStorage.removeItem(REFRESH_KEY)
  // Cache daftar akun ikut dibersihkan: jangan bocor ke sesi berikutnya.
  sessionStorage.removeItem(ACCOUNTS_KEY)
  localStorage.removeItem(ACCOUNTS_KEY)
}

export function saveTokens(access: string, refresh: string) {
  sessionStorage.setItem(ACCESS_KEY, access)
  sessionStorage.setItem(REFRESH_KEY, refresh)
}

let refreshing: Promise<string> | null = null

export async function refreshTokens(): Promise<string> {
  if (!refreshing) {
    refreshing = (async () => {
      const rt = sessionStorage.getItem(REFRESH_KEY)
      if (!rt) throw new ApiError(401, 'Sesi berakhir. Silakan masuk kembali.')
      const res = await fetch(`${BASE}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: rt }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        clearTokens()
        throw new ApiError(401, data.error ?? 'Sesi berakhir. Silakan masuk kembali.')
      }
      saveTokens(data.access_token, data.refresh_token)
      return data.access_token as string
    })()
    try {
      await refreshing
    } finally {
      refreshing = null
    }
  }
  return refreshing!
}

export async function request<T>(method: string, path: string, body?: unknown, auth = true): Promise<T> {
  let token = auth ? getToken() : null
  for (let attempt = 0; attempt < 2; attempt++) {
    const res = await fetch(BASE + path, {
      method,
      headers: {
        ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: body !== undefined ? JSON.stringify(body) : undefined,
    })
    if (res.status === 401 && attempt === 0 && token && !path.startsWith('/auth/')) {
      token = await refreshTokens()
      continue
    }
    const data = await res.json().catch(() => ({}))
    if (!res.ok) throw new ApiError(res.status, data.error ?? 'Terjadi kesalahan.', data.code ?? data.error ?? '')
    return data as T
  }
  throw new ApiError(401, 'Sesi berakhir. Silakan masuk kembali.')
}

export async function requestFile(path: string): Promise<{ blob: Blob; filename: string | null }> {
  let token = getToken()
  for (let attempt = 0; attempt < 2; attempt++) {
    const res = await fetch(BASE + path, { headers: token ? { Authorization: `Bearer ${token}` } : {} })
    if (res.status === 401 && attempt === 0 && token && !path.startsWith('/auth/')) {
      token = await refreshTokens()
      continue
    }
    if (!res.ok) {
      const data = await res.json().catch(() => ({}))
      throw new ApiError(res.status, data.error ?? 'Terjadi kesalahan.', data.code ?? data.error ?? '')
    }
    const disposition = res.headers.get('Content-Disposition')
    const filename = disposition?.match(/filename="?([^";]+)"?/i)?.[1] ?? null
    return { blob: await res.blob(), filename }
  }
  throw new ApiError(401, 'Sesi berakhir. Silakan masuk kembali.')
}

// ── tipe server (snake_case) ─────────────────────────────────────────

export function qs(p: Record<string, string | number | boolean | undefined>) {
  const s = new URLSearchParams()
  for (const [k, v] of Object.entries(p)) if (v !== undefined && v !== '') s.set(k, String(v))
  const str = s.toString()
  return str ? '?' + str : ''
}

export async function fetchAll<T>(pageFn: (page: number) => Promise<Page<T>>): Promise<T[]> {
  const out: T[] = []
  for (let page = 1; ; page++) {
    const d = await pageFn(page)
    out.push(...d.items)
    if (out.length >= d.total) break
  }
  return out
}

export function useFetch<T>(fn: () => Promise<T>, deps: unknown[]): { data: T | null; err: string; loading: boolean; reload: () => void } {
  const [data, setData] = useState<T | null>(null)
  const [err, setErr] = useState('')
  const [loading, setLoading] = useState(true)
  const [tick, setTick] = useState(0)
  useEffect(() => {
    let dead = false
    setLoading(true)
    setErr('')
    fn().then((d) => { if (!dead) setData(d) })
      .catch((e) => { if (!dead) setErr(e instanceof Error ? e.message : 'Terjadi kesalahan.') })
      .finally(() => { if (!dead) setLoading(false) })
    return () => { dead = true }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, tick])
  return { data, err, loading, reload: () => setTick((t) => t + 1) }
}
