// Endpoint autentikasi, akun, pengguna, dan outlet.

import { BASE, REFRESH_KEY, ACCOUNTS_KEY, getToken, clearTokens, saveTokens, request } from './core.ts'
import type { Role, User, Outlet, AuthResp, UserResponse, OutletResponseItem } from './types.ts'
import { normalizeUser, normalizeAuthResponse, normalizeOutlet, outletIdForWire } from './types.ts'

export function apiLogin(email: string, password: string) {
  // Passcode tidak dikirim saat login — ia hanya dipakai saat ganti akun (switch).
  return request<Omit<AuthResp, 'user'> & { user: UserResponse }>('POST', '/auth/login', {
    email,
    password,
  }, false).then((r) => { saveTokens(r.access_token, r.refresh_token); return normalizeAuthResponse(r) })
}

export function apiRegister(name: string, email: string, password: string, storeName: string) {
  return request<Omit<AuthResp, 'user'> & { user: UserResponse }>('POST', '/auth/register', { name, email, password, storeName }, false)
    .then((r) => { saveTokens(r.access_token, r.refresh_token); return normalizeAuthResponse(r) })
}

export async function apiLogout() {
  const rt = sessionStorage.getItem(REFRESH_KEY)
  const token = getToken()
  try {
    // Manual fetch (bukan request()): logout tak boleh memicu auto-refresh.
    // Authorization ikut dikirim agar backend tahu sesi kasir mana yang
    // harus ditandai offline (kontrak presence), tanpa bocor ke body.
    await fetch(`${BASE}/auth/logout`, {
      method: 'POST',
      headers: {
        ...(rt ? { 'Content-Type': 'application/json' } : {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: rt ? JSON.stringify({ refresh_token: rt }) : undefined,
    })
  } catch {
    // best-effort
  }
  clearTokens()
}

export function apiMe() {
  return request<{ user: UserResponse }>('GET', '/auth/me').then((response) => ({ user: normalizeUser(response.user) }))
}

// Heartbeat presence: dipanggil tiap 30 dtk selagi login.
// Best-effort — 404 di backend lama bukan error.
// ponytail: tanpa retry/tracking; server yang hitung online dari last_seen.
export async function apiHeartbeat(): Promise<void> {
  try {
    await request<{ status: string }>('POST', '/presence/heartbeat')
  } catch {
    // abaikan; presence tak tampil sampai kontrak live
  }
}

// ── ganti akun cepat (admin ⇄ kasir, dalam toko yang sama) ────────────
// Kontrak: README backend §"Alur Pembagian Sesi Akun".

export function apiSwitchAccount(targetUserId: string, passcode?: string, role?: Role) {
  return request<Omit<AuthResp, 'user'> & { user: UserResponse }>(
    'POST',
    '/auth/switch',
    { target_user_id: targetUserId, ...(passcode ? { passcode } : {}), ...(role ? { role } : {}) },
  ).then((r) => { saveTokens(r.access_token, r.refresh_token); return normalizeAuthResponse(r) })
}

export function apiListOutlets(): Promise<Outlet[]> {
  return request<{ items: OutletResponseItem[] | null }>('GET', '/outlets').then((d) => (d.items ?? []).map(normalizeOutlet))
}

type OutletResponse = { outlet: OutletResponseItem } | OutletResponseItem

function outletFromResponse(data: OutletResponse): Outlet {
  return normalizeOutlet('outlet' in data ? data.outlet : data)
}

export function apiCreateOutlet(body: Pick<Outlet, 'name' | 'address'>): Promise<Outlet> {
  return request<OutletResponse>('POST', '/outlets', body).then(outletFromResponse)
}

export function apiUpdateOutlet(id: string, body: Pick<Outlet, 'name' | 'address'>): Promise<Outlet> {
  return request<OutletResponse>('PUT', `/outlets/${id}`, body).then(outletFromResponse)
}

export function apiSwitchOutlet(outletId: string): Promise<AuthResp> {
  return request<Omit<AuthResp, 'user'> & { user: UserResponse }>('POST', '/outlets/switch', { outletId: outletIdForWire(outletId) }).then((response) => {
    const r = normalizeAuthResponse(response)
    saveTokens(r.access_token, r.refresh_token)
    return r
  })
}

// Landing pasca-login: true bila ada kasir aktif untuk dipilih.
export async function apiHasActiveCashiers(): Promise<boolean> {
  try {
    const users = await apiListUsers()
    return users.some((u) => u.role === 'cashier' && u.active)
  } catch {
    return false
  }
}

// Cache akun per-tab: cegah daftar akun admin terbaca sesi lain
// di browser yang sama setelah ganti akun/logout.
export function getCachedAccounts(): User[] {
  try {
    const raw = sessionStorage.getItem(ACCOUNTS_KEY)
    return raw ? (JSON.parse(raw) as User[]) : []
  } catch {
    return []
  }
}

export function setCachedAccounts(users: User[]) {
  try {
    sessionStorage.setItem(ACCOUNTS_KEY, JSON.stringify(users))
  } catch {
    // abaikan — cache hanya akselerator menu ganti akun
  }
  // Sinkronkan menu ganti akun di sidebar tanpa reload halaman.
  window.dispatchEvent(new Event('op:accounts-changed'))
}

// ── verifikasi email OTP ─────────────────────────────────────────────
// Kontrak: docs/API-CONTRACT-EMAIL-OTP.md (diajukan ke backend developer)

// purpose "login": email terdaftar tetap dikirim OTP (kontrak LOGIN-OTP).
// Tanpa purpose: perilaku registrasi lama (email terdaftar tidak dikirim).
export function apiSendOtp(email: string, purpose?: 'login') {
  return request<{ message: string }>('POST', '/auth/otp/send', { email, ...(purpose ? { purpose } : {}) }, false)
}

export function apiVerifyOtp(email: string, code: string) {
  return request<{ verified: boolean; message: string }>('POST', '/auth/otp/verify', { email, code }, false)
}

// ── Google login (GIS ID-token flow) ───────────────────────────────────
// Kontrak: FRONTEND_AUTH.md di repo backend (rayhanvito/nojpos)

// Google login. Akun ber-passcode: backend balas passcode_required →
// pemanggil menampilkan form PIN lalu memanggil ulang dengan passcode.
export function apiGoogleLogin(idToken: string, storeName?: string) {
  return request<Omit<AuthResp, 'user'> & { user: UserResponse }>('POST', '/auth/google', {
    id_token: idToken,
    ...(storeName ? { storeName } : {}),
  }, false).then((r) => { saveTokens(r.access_token, r.refresh_token); return normalizeAuthResponse(r) })
}

// ── users ────────────────────────────────────────────────────────────

export function apiListUsers() {
  return request<{ users: UserResponse[] | null }>('GET', '/users').then((d) => (d.users ?? []).map(normalizeUser))
}

export function apiCreateUser(body: { name: string; outletId?: string }) {
  const requestBody = { ...body, ...(body.outletId ? { outletId: outletIdForWire(body.outletId) } : {}) }
  return request<{ user: UserResponse }>('POST', '/users', requestBody).then((d) => normalizeUser(d.user))
}

export function apiSetUserActive(id: string, active: boolean) {
  return request<{ message: string }>('PATCH', `/users/${id}/active`, { active })
}

// Ganti nama kasir (live backend 10 Sep) — lihat docs/API-CONTRACT-USER-RENAME.md.
export function apiRenameUser(id: string, name: string, outletId?: string) {
  return request<{ message: string }>('PATCH', `/users/${id}`, { name, ...(outletId ? { outletId: outletIdForWire(outletId) } : {}) })
}

export function apiDeleteUser(id: string) {
  return request<{ message: string }>('DELETE', `/users/${id}`)
}

export function apiSetPasscode(id: string, passcode: string, role?: string) {
  return request<{ message: string }>('PUT', `/users/${id}/passcode`, role ? { passcode, role } : { passcode })
}

// ── katalog ──────────────────────────────────────────────────────────

export function apiSendPasswordResetOtp(email: string) {
  return request<{ message: string }>('POST', '/auth/forgot-password/send', { email }, false)
}

export function apiResetPassword(email: string, code: string, newPassword: string) {
  return request<{ message: string }>('POST', '/auth/forgot-password/reset', { email, code, new_password: newPassword }, false)
}

// date = hari lokal 'YYYY-MM-DD' (kontrak docs/API-CONTRACT-DASHBOARD-DATE.md).
// Tanpa date = hari ini zona toko. Backend abaikan date tak valid (jangan 400).
