import { useEffect, useState } from 'react'
import { Link, Navigate, Outlet, useLocation, useNavigate } from 'react-router'
import {
  LayoutDashboard, Store, Package, Boxes, ReceiptText, BarChart3, Users, Settings, TicketPercent, UserRound, Truck, ClipboardList, CreditCard,
  LogOut, Check, PanelLeftClose, PanelLeftOpen, ChevronDown,
} from 'lucide-react'
import { ApiError, apiCheckout, apiHeartbeat, apiListUsers, apiLogout, apiMe, apiSwitchAccount, apiSwitchOutlet, getCachedAccounts, hasToken, setCachedAccounts, type User } from '../shared/api'
import { clearCache } from '../shared/cache'
import { queryClient } from '../shared/api/query-client'
import { setSession, toSession, useDB } from '../shared/session/store'
import { NotifBell } from '../shared/notifications'
import ShiftControl from './ShiftControl'
import { ModalActions, AlertStrip, Button, Logo, Modal, Select } from '../shared/ui'
import { outletRouteKey, shouldShowCashierOutletLabel, shouldShowOutletSwitcher } from '../shared/outlet'
import { useOutlets } from '../hooks/use-outlets'
import {
  dismissOfflineQueueItem,
  listOfflineQueue,
  queueItemsForStore,
  retryOfflineQueueItemInStore,
  syncOfflineQueue,
  type OfflineQueueItem,
} from '../shared/offline-queue'

const MENU: { label: string; to: string; icon: React.ComponentType<{ className?: string }>; adminOnly?: boolean }[] = [
  { label: 'Beranda', to: '/app', icon: LayoutDashboard },
  { label: 'Kasir', to: '/app/pos', icon: Store },
  { label: 'Produk & Kategori', to: '/app/products', icon: Package, adminOnly: true },
  { label: 'Promo & Voucher', to: '/app/promo', icon: TicketPercent, adminOnly: true },
  { label: 'Stok', to: '/app/stock', icon: Boxes, adminOnly: true },
  { label: 'Transaksi', to: '/app/transactions', icon: ReceiptText },
  { label: 'Pelanggan', to: '/app/customers', icon: UserRound },
  { label: 'Supplier', to: '/app/supplier', icon: Truck, adminOnly: true },
  { label: 'Pembelian', to: '/app/purchases', icon: ClipboardList, adminOnly: true },
  { label: 'Hutang Piutang', to: '/app/debts', icon: CreditCard, adminOnly: true },
  { label: 'Laporan', to: '/app/reports', icon: BarChart3, adminOnly: true },
  { label: 'Pengguna', to: '/app/users', icon: Users, adminOnly: true },
  { label: 'Pengaturan', to: '/app/settings', icon: Settings, adminOnly: true },
]

// Pengelompokan visual sidebar (admin). Murni tampilan — rute & permission tak berubah.
const GROUP_OF: Record<string, 'UTAMA' | 'MANAJEMEN' | 'PENGATURAN'> = {
  '/app': 'UTAMA',
  '/app/pos': 'UTAMA',
  '/app/products': 'MANAJEMEN',
  '/app/promo': 'MANAJEMEN',
  '/app/stock': 'MANAJEMEN',
  '/app/transactions': 'MANAJEMEN',
  '/app/customers': 'MANAJEMEN',
  '/app/supplier': 'MANAJEMEN',
  '/app/purchases': 'MANAJEMEN',
  '/app/debts': 'MANAJEMEN',
  '/app/reports': 'MANAJEMEN',
  '/app/users': 'PENGATURAN',
  '/app/settings': 'PENGATURAN',
}
const GROUP_ORDER = ['UTAMA', 'MANAJEMEN', 'PENGATURAN'] as const

function isMenuActive(pathname: string, to: string) {
  return pathname === to || (to !== '/app' && pathname.startsWith(to))
}

export default function AppShell() {
  const db = useDB()
  const loc = useLocation()
  const [navOpen, setNavOpen] = useState(false)
  const [collapsed, setCollapsed] = useState(false)
  const [boot, setBoot] = useState(!hasToken())
  const [online, setOnline] = useState(() => navigator.onLine)
  const [queueItems, setQueueItems] = useState<OfflineQueueItem[]>([])
  const [queueError, setQueueError] = useState('')
  const [queueAction, setQueueAction] = useState<string | null>(null)
  const [outletOpen, setOutletOpen] = useState(false)
  const [outletChoice, setOutletChoice] = useState('')
  const [outletBusy, setOutletBusy] = useState(false)
  const [outletError, setOutletError] = useState('')
  const { outlets } = useOutlets()
  const s = db.session

  async function refreshQueue() {
    try {
      setQueueItems(queueItemsForStore(await listOfflineQueue(), s?.storeId ?? ''))
      setQueueError('')
    } catch (error) {
      setQueueError(error instanceof Error ? error.message : 'Gagal membaca antrean transaksi.')
    }
  }

  // Boot sesi dari token (guard masuk bila tanpa sesi valid).
  useEffect(() => {
    if (!hasToken() || s) return
    let dead = false
    apiMe()
      .then((r) => { if (!dead) setSession(toSession(r.user)) })
      .catch(() => { if (!dead) { apiLogout(); setSession(null) } })
      .finally(() => { if (!dead) setBoot(true) })
    return () => { dead = true }
  }, [s])

  // Heartbeat presence tiap 30 dtk selagi sesi ada (kontrak presence).
  useEffect(() => {
    if (!s) return
    apiHeartbeat()
    const timer = setInterval(apiHeartbeat, 30_000)
    return () => clearInterval(timer)
  }, [s?.id, s?.role])

  useEffect(() => {
    let dead = false
    const refresh = () => {
      void listOfflineQueue()
        .then((items) => { if (!dead) { setQueueItems(queueItemsForStore(items, s?.storeId ?? '')); setQueueError('') } })
        .catch((error: unknown) => { if (!dead) setQueueError(error instanceof Error ? error.message : 'Gagal membaca antrean transaksi.') })
    }
    const onOnline = () => {
      setOnline(true)
      if (s) void syncOfflineQueue(apiCheckout, s.storeId).catch((error: unknown) => { if (!dead) setQueueError(error instanceof Error ? error.message : 'Gagal menyinkronkan antrean.') }).finally(refresh)
      else refresh()
    }
    const onOffline = () => setOnline(false)
    const onQueueChange = () => refresh()
    window.addEventListener('online', onOnline)
    window.addEventListener('offline', onOffline)
    window.addEventListener('op:offline-queue-change', onQueueChange)
    refresh()
    if (s && navigator.onLine) {
      void syncOfflineQueue(apiCheckout, s.storeId).catch((error: unknown) => { if (!dead) setQueueError(error instanceof Error ? error.message : 'Gagal menyinkronkan antrean.') }).finally(refresh)
    }
    return () => {
      dead = true
      window.removeEventListener('online', onOnline)
      window.removeEventListener('offline', onOffline)
      window.removeEventListener('op:offline-queue-change', onQueueChange)
    }
  }, [s?.id, s?.storeId])

  // Tutup drawer tiap pindah halaman (mobile).
  useEffect(() => {
    setNavOpen(false)
  }, [loc.pathname])

  if (!s) {
    if (hasToken() && !boot) return null
    return <Navigate to="/login" replace />
  }

  const menu = MENU.filter((m) => !m.adminOnly || s.role === 'admin')
  const activeStoreId = s.storeId
  const currentOutletId = s.outletId
  // Ikon toggle mengikuti status: tertutup→buka, terbuka→tutup.
  const narrow = typeof window !== 'undefined' && window.innerWidth < 768
  const navExpanded = narrow ? navOpen : !collapsed
  const waitingCount = queueItems.filter((item) => item.status === 'menunggu').length
  const showOutletSwitcher = shouldShowOutletSwitcher(s.role, outlets.length)
  const showCashierOutletLabel = shouldShowCashierOutletLabel(s.role, outlets.length)
  const activeOutletName = s.outletName || outlets.find((outlet) => outlet.id === s.outletId)?.name || ''

  function openOutletPicker() {
    setOutletChoice(currentOutletId ?? outlets[0]?.id ?? '')
    setOutletError('')
    setOutletOpen(true)
  }

  async function switchOutlet() {
    if (!outletChoice) return setOutletError('Pilih outlet terlebih dahulu.')
    setOutletBusy(true)
    setOutletError('')
    try {
      const response = await apiSwitchOutlet(outletChoice)
      setSession(toSession(response.user))
      clearCache()
      queryClient.clear()
      setOutletOpen(false)
      setNavOpen(false)
    } catch (error) {
      setOutletError(error instanceof Error ? error.message : 'Gagal berpindah outlet. Coba lagi.')
    } finally {
      setOutletBusy(false)
    }
  }

  async function retryQueueItem(clientTxId: string) {
    setQueueAction(clientTxId)
    setQueueError('')
    try {
      const item = await retryOfflineQueueItemInStore(activeStoreId ?? '', clientTxId)
      if (item && navigator.onLine) await syncOfflineQueue(apiCheckout, activeStoreId)
      await refreshQueue()
    } catch (error) {
      setQueueError(error instanceof Error ? error.message : 'Gagal mencoba ulang sinkronisasi.')
    } finally {
      setQueueAction(null)
    }
  }

  async function dismissQueueItem(clientTxId: string) {
    setQueueAction(clientTxId)
    setQueueError('')
    try {
      const dismissed = await dismissOfflineQueueItem(activeStoreId ?? '', clientTxId)
      if (!dismissed) setQueueError('Transaksi yang masih menunggu tidak dapat dihapus dari antrean.')
      await refreshQueue()
    } catch (error) {
      setQueueError(error instanceof Error ? error.message : 'Gagal menutup transaksi antrean.')
    } finally {
      setQueueAction(null)
    }
  }

  return (
    <div className={`opc-shell op-classic${navOpen ? ' nav-open' : ''}${collapsed ? ' nav-collapsed' : ''}`}>
      <div className="opc-toprow">
        <Link to="/app" className="opc-brand" aria-label="Dashboard">
          <Logo tone="dark" className="h-[26px] w-auto shrink-0" />
          <span>{s.store || 'Toko Anda'}</span>
        </Link>
        <header className="opc-topbar">
          <button
            className="opc-iconbtn"
            onClick={() => {
              if (window.innerWidth < 768) setNavOpen((v) => !v)
              else setCollapsed((v) => !v)
            }}
            aria-label="Buka/tutup navigasi"
            aria-expanded={navExpanded}
          >
            <span key={navExpanded ? 'open' : 'closed'} className="opc-icopop">
              {navExpanded ? <PanelLeftClose className="size-5" strokeWidth={1.5} /> : <PanelLeftOpen className="size-5" strokeWidth={1.5} />}
            </span>
          </button>
          <div className="ml-auto flex items-center gap-1 min-[640px]:gap-2">
            <ShiftControl online={online} waitingCount={waitingCount} queueItems={queueItems} queueError={queueError} queueAction={queueAction} onRetryItem={retryQueueItem} onDismissItem={dismissQueueItem} onRefreshQueue={refreshQueue} />
            <span className="opc-bell">
              <NotifBell />
            </span>
            <UserMenu />
          </div>
        </header>
      </div>

      <div className="opc-body">
        <aside className="opc-sidebar" aria-label="Navigasi utama">
          {(showOutletSwitcher || showCashierOutletLabel) && (
            <div className="opc-outletbox border-b border-white/10 px-3 py-3">
              {showOutletSwitcher ? (
                <button
                  type="button"
                  onClick={openOutletPicker}
                  aria-label={`Pindah outlet. Outlet aktif: ${activeOutletName || 'tidak diketahui'}`}
                  title="Pindah outlet aktif"
                  className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-[13px] font-medium text-white transition hover:bg-white/10"
                >
                  <Store className="size-4 shrink-0" aria-hidden="true" />
                  <span className="min-w-0 flex-1 truncate">Outlet: {activeOutletName || 'Pilih'}</span>
                  <ChevronDown className="size-4 shrink-0 opacity-70" aria-hidden="true" />
                </button>
              ) : (
                <p className="flex items-center gap-2 px-3 py-2 text-[13px] font-medium text-white/80" aria-label={`Outlet terikat: ${activeOutletName || 'tidak diketahui'}`}>
                  <Store className="size-4 shrink-0" aria-hidden="true" />
                  <span className="min-w-0 flex-1 truncate">Outlet: {activeOutletName || '—'}</span>
                </p>
              )}
            </div>
          )}
          <nav className="opc-nav">
            {GROUP_ORDER.map((g) => {
              const items = menu.filter((m) => GROUP_OF[m.to] === g)
              if (items.length === 0) return null
              return (
                <div key={g}>
                  <ul>
                    {items.map((m) => {
                      const active = isMenuActive(loc.pathname, m.to)
                      return (
                        <li key={m.to}>
                          <Link to={m.to} title={collapsed ? m.label : undefined} aria-label={m.label} aria-current={active ? 'page' : undefined} className={`opc-link${active ? ' active' : ''}`}>
                            <m.icon className="" />
                            <span>{m.label}</span>
                          </Link>
                        </li>
                      )
                    })}
                  </ul>
                </div>
              )
            })}
          </nav>
        </aside>
        <div className="opc-scrim" onClick={() => setNavOpen(false)} aria-hidden="true" />

        <main className="opc-main">
          <Outlet key={outletRouteKey(s.outletId)} />
        </main>
      </div>

      <Modal open={outletOpen} title="Pindah Outlet" onClose={() => { if (!outletBusy) setOutletOpen(false) }}>
        <div className="space-y-4">
          <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
            Pilih outlet
            <Select
              label="Pilih outlet"
              value={outletChoice}
              onChange={setOutletChoice}
              disabled={outletBusy}
              className="w-full"
              options={outlets.map((outlet) => ({ value: String(outlet.id), label: outlet.name }))}
            />
          </label>
          {outletError && <AlertStrip>{outletError}</AlertStrip>}
          <ModalActions>
            <Button variant="ghost" disabled={outletBusy} onClick={() => setOutletOpen(false)}>Batal</Button>
            <Button disabled={outletBusy || !outletChoice} onClick={() => void switchOutlet()}>{outletBusy ? 'Memindahkan…' : 'Pindah Outlet'}</Button>
          </ModalActions>
        </div>
      </Modal>

    </div>
  )
}

function initials(name: string) {
  const parts = name.trim().split(/\s+/)
  return ((parts[0]?.[0] ?? '') + (parts[1]?.[0] ?? '')).toUpperCase() || '?'
}

function UserMenu() {
  const db = useDB()
  const nav = useNavigate()
  const s = db.session!
  const [open, setOpen] = useState(false)
  const [busy, setBusy] = useState(false)
  const [accounts, setAccounts] = useState<User[]>(() => getCachedAccounts())
  const [pending, setPending] = useState<User | null>(null)
  const [passcode, setPasscode] = useState('')
  const [err, setErr] = useState('')

  useEffect(() => {
    if (s.role !== 'admin') return
    apiListUsers()
      .then((users) => { setAccounts(users); setCachedAccounts(users) })
      .catch(() => {})
  }, [s.role, s.id])

  // Ikut cache terbaru (tambah/hapus kasir di halaman Users) tanpa reload.
  useEffect(() => {
    const sync = () => setAccounts(getCachedAccounts())
    window.addEventListener('op:accounts-changed', sync)
    return () => window.removeEventListener('op:accounts-changed', sync)
  }, [])

  function resetMenu() {
    setOpen(false)
    setPending(null)
    setPasscode('')
    setErr('')
  }

  async function keluar() {
    setBusy(true)
    await apiLogout()
    setSession(null)
    // Keluar dari demo = kembali ke landing page awal.
    nav('/login', { replace: true })
  }

  async function pick(u: User) {
    if (u.id === s.id && u.role === s.role) return resetMenu()
    setErr('')
    setBusy(true)
    try {
      const r = await apiSwitchAccount(u.id, undefined, u.role)
      setSession(toSession(r.user))
      resetMenu()
      nav('/app')
    } catch (x) {
      if (x instanceof ApiError && x.code === 'passcode_required') {
        setPending(u)
        setPasscode('')
      } else {
        setErr(x instanceof Error ? x.message : 'Gagal ganti akun.')
      }
    } finally {
      setBusy(false)
    }
  }

  async function submitPasscode() {
    if (!pending) return
    setErr('')
    setBusy(true)
    try {
      const r = await apiSwitchAccount(pending.id, passcode, pending.role)
      setSession(toSession(r.user))
      resetMenu()
      nav('/app')
    } catch (x) {
      setErr(x instanceof Error ? x.message : 'Passcode salah. Coba lagi.')
      setPasscode('')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="relative">
      <button
        onClick={() => { setOpen(!open); setPending(null); setErr('') }}
        aria-expanded={open}
        aria-haspopup="menu"
        className="opc-sideuser"
      >
        <span className="opc-avatar opc-avatar-initials" aria-hidden="true">
          {initials(s.name)}
        </span>
        <span className="hidden min-w-0 flex-1 leading-tight min-[480px]:grid">
          <span className="truncate text-sm font-medium text-white">{s.name}</span>
        </span>
        <ChevronDown className="hidden size-4 shrink-0 opacity-70 min-[480px]:block" aria-hidden="true" />
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-40 max-sm:bg-black/50" onClick={() => setOpen(false)} aria-hidden="true" />
          <div role="menu" className="opc-acc-panel">
            {pending ? (
              <div className="opc-acc-pin">
                <p className="opc-acc-head">
                  Passcode · {pending.name}
                </p>
                <input
                  value={passcode}
                  onChange={(e) => setPasscode(e.target.value.replace(/\D/g, '').slice(0, 5))}
                  onKeyDown={(e) => { if (e.key === 'Enter') submitPasscode() }}
                  inputMode="numeric"
                  autoFocus
                  placeholder="•••••"
                  aria-label="Passcode 5 angka"
                />
                {err && <p className="opc-acc-err">{err}</p>}
                <div className="opc-acc-row">
                  <button className="opc-acc-btn" onClick={() => { setPending(null); setErr('') }}>
                    Batal
                  </button>
                  <button
                    className="opc-acc-btn primary"
                    onClick={submitPasscode}
                    disabled={passcode.length !== 5 || busy}
                  >
                    {busy ? '…' : 'Masuk'}
                  </button>
                </div>
              </div>
            ) : (
              <>
                <p className="opc-acc-head">Ganti akun</p>
                <div className="opc-acc-list">
                  {accounts.filter((a) => a.active).map((a) => {
                    const active = a.id === s.id && a.role === s.role
                    return (
                      <button
                        key={`${a.role}-${a.id}`}
                        role="menuitem"
                        disabled={busy}
                        onClick={() => pick(a)}
                        aria-current={active || undefined}
                        className={`opc-acc-item${active ? ' active' : ''}`}
                      >
                        <span className="opc-acc-ava" aria-hidden="true">
                          {initials(a.name)}
                        </span>
                        <span className="opc-acc-meta">
                          <strong>{a.name}</strong>
                          <small>{a.role === 'admin' ? 'Admin' : 'Kasir'}</small>
                        </span>
                        {active && <Check className="opc-acc-check" aria-hidden="true" />}
                      </button>
                    )
                  })}
                  {accounts.length === 0 && (
                    <p className="px-2.5 py-2 text-xs text-muted-foreground">Daftar akun tidak tersedia.</p>
                  )}
                </div>
                {err && <p className="opc-acc-err">{err}</p>}
                <div className="opc-acc-div" />
                {(
                  <button
                    role="menuitem"
                    onClick={keluar}
                    disabled={busy}
                    className="opc-acc-logout"
                  >
                    <LogOut aria-hidden="true" />
                    {busy ? 'Keluar…' : 'Keluar'}
                  </button>
                )}
              </>
            )}
          </div>
        </>
      )}
    </div>
  )
}

