import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router'
import { ApiError, apiListUsers, apiLogout, apiResetPassword, apiSendPasswordResetOtp, apiSetPasscode, type User } from '../../lib/api'
import { useCache } from '../../lib/cache'
import { fmtDate, fmtTime, setSession, useDB } from '../../lib/store'
import { Check, Info, KeyRound, Lock, LogOut, ShieldCheck, Users, Trash2 } from 'lucide-react'
import { AlertStrip, Button, Modal, Pill, RowAction, SettingsCard, SettingsHead } from '../../lib/ui'
import { MetaRow, PwField, SCORE_BAR, SCORE_LABEL, pwScore, todayStr} from './shared'

export function AkunTab({ busy, setBusy, setMsg, setErr }: {
  busy: boolean
  setBusy: (v: boolean) => void
  setMsg: (v: string) => void
  setErr: (v: string) => void
}) {
  const nav = useNavigate()
  const { session } = useDB()
  const accountList = useCache<User[]>(`users:${session?.id}`, () => apiListUsers())
  const users = accountList.data ?? []
  const today = todayStr()

  const [manageOpen, setManageOpen] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [pcVal, setPcVal] = useState('')
  const [pcBusy, setPcBusy] = useState(false)
  const [confirmResetId, setConfirmResetId] = useState<string | null>(null)

  const [showNew, setShowNew] = useState(false)
  const [showNew2, setShowNew2] = useState(false)
  const [otp, setOtp] = useState('')
  const [otpSent, setOtpSent] = useState(false)
  const [otpSending, setOtpSending] = useState(false)
  const [cooldown, setCooldown] = useState(0)
  const [otpRequested, setOtpRequested] = useState(false)

  useEffect(() => {
    if (cooldown <= 0) return
    const t = setTimeout(() => setCooldown((c) => c - 1), 1000)
    return () => clearTimeout(t)
  }, [cooldown])
  const [newPw, setNewPw] = useState('')
  const [newPw2, setNewPw2] = useState('')
  const [pwdBusy, setPwdBusy] = useState(false)
  const [pwdErr, setPwdErr] = useState('')

  // Cetak uji coba struk.

  async function savePc(u: User) {
    if (!/^\d{5}$/.test(pcVal)) return setErr('Passcode harus 5 angka.')
    setMsg(''); setErr(''); setPcBusy(true)
    try {
      await apiSetPasscode(u.id, pcVal, u.role)
      accountList.mutate(users.map((x) => (x.id === u.id ? { ...x, has_passcode: true } : x)))
      setEditingId(null); setPcVal('')
      setMsg(`Passcode ${u.name} disimpan.`)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal menyimpan passcode.')
    } finally {
      setPcBusy(false)
    }
  }

  async function clearPc(u: User) {
    setMsg(''); setErr(''); setPcBusy(true)
    try {
      await apiSetPasscode(u.id, '', u.role)
      accountList.mutate(users.map((x) => (x.id === u.id ? { ...x, has_passcode: false } : x)))
      setConfirmResetId(null)
      setMsg(`Passcode ${u.name} dinonaktifkan.`)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal menghapus passcode.')
    } finally {
      setPcBusy(false)
    }
  }

  const sendPwOtp = useCallback(async () => {
    const email = session?.email.trim().toLowerCase() ?? ''
    if (!email) return setPwdErr('Email akun tidak ditemukan.')
    setPwdErr(''); setOtp(''); setOtpSending(true); setOtpRequested(true)
    try {
      await apiSendPasswordResetOtp(email)
      setOtpSent(true)
      setCooldown(60)
    } catch (x) {
      if (x instanceof ApiError && x.status === 429) setCooldown(60)
      setOtpSent(false)
      setPwdErr(x instanceof Error ? x.message : 'Gagal mengirim kode. Coba lagi.')
    } finally {
      setOtpSending(false)
    }
  }, [session?.email])

  // Verifikasi muncul otomatis: begitu kata sandi baru valid,
  // kode OTP langsung dikirim tanpa perlu tekan tombol.
  const pwValid = newPw.length >= 8 && newPw === newPw2
  useEffect(() => {
    if (!pwValid || otpRequested || cooldown > 0 || pwdBusy) return
    const t = setTimeout(() => { sendPwOtp() }, 400)
    return () => clearTimeout(t)
  }, [pwValid, otpRequested, cooldown, pwdBusy, sendPwOtp])

  async function submitPwReset() {
    const email = session?.email.trim().toLowerCase() ?? ''
    setPwdErr('')
    if (newPw.length < 8) return setPwdErr('Kata sandi baru minimal 8 karakter.')
    if (newPw !== newPw2) return setPwdErr('Konfirmasi kata sandi tidak cocok.')
    if (otp.length !== 6) return setPwdErr('Masukkan kode 6 digit dari email.')
    setPwdBusy(true)
    try {
      await apiResetPassword(email, otp, newPw)
      setNewPw(''); setNewPw2(''); setOtp(''); setOtpSent(false)
      await apiLogout()
      setSession(null)
      nav('/app', { replace: true })
    } catch (x) {
      if (x instanceof ApiError && (x.status === 410 || x.status === 429)) setCooldown(0)
      if (x instanceof ApiError && x.status === 400) setOtp('')
      setPwdErr(x instanceof Error ? x.message : 'Gagal mengganti kata sandi.')
    } finally {
      setPwdBusy(false)
    }
  }

  async function keluar() {
    setBusy(true)
    await apiLogout()
    setSession(null)
    nav('/app', { replace: true })
  }

  const s = session
  const me = users.find((u) => u.id === s?.id) ?? null
  const cashiers = users.filter((u) => u.role === 'cashier')
  const manageUsers = me ? [me, ...cashiers.filter((c) => c.id !== me.id)] : [...cashiers]

  return (
    <>
    <div className="space-y-4">
      <div className="grid items-start gap-4 lg:grid-cols-2">
        <SettingsCard>
          <SettingsHead kicker="Akun saya" title="Informasi Akun" />
          <div className="flex flex-wrap items-center gap-3">
            <span className="grid size-11 shrink-0 place-items-center rounded-full bg-jet font-mono text-lg text-paper" aria-hidden="true">
              {(session?.name ?? '?').charAt(0).toUpperCase()}
            </span>
            <div className="min-w-0 flex-1 basis-32">
              <p className="break-words text-[15px] font-medium text-fg">{session?.name ?? '—'}</p>
              <p className="break-words text-[13px] text-muted">{session?.email || 'Tanpa email'} · {session?.store ?? '—'}</p>
            </div>
            <span className="flex flex-wrap gap-1.5">
              <Pill tone={session?.role === 'admin' ? 'ok' : 'muted'}>{session?.role === 'admin' ? 'Admin' : 'Kasir'}</Pill>
              <Pill tone={me?.active === false ? 'muted' : 'ok'}>{me?.active === false ? 'Nonaktif' : 'Akun Aktif'}</Pill>
            </span>
          </div>
          <div className="mt-4 divide-y divide-dove border-t border-dove">
            <MetaRow label="Toko" value={session?.store ?? '—'} />
            <MetaRow label="Bergabung" value={me?.created_at ? fmtDate(me.created_at) : '—'} />
            <MetaRow
              label="Terakhir login"
              value={!me?.last_seen_at ? '—' : me.last_seen_at.slice(0, 10) >= today ? `Hari ini, ${fmtTime(me.last_seen_at)}` : fmtDate(me.last_seen_at)}
            />
          </div>
          <div className="mt-5 flex flex-col gap-2 sm:flex-row sm:flex-wrap">
            {session?.role === 'admin' && (
              <Link to="/app/users" className="inline-flex w-full items-center justify-center gap-2 rounded-full border border-dove bg-transparent px-6 py-3 text-[15px] font-medium transition hover:border-jet sm:w-auto">
                <Users className="size-4" /> Kelola Kasir
              </Link>
            )}
            <Button variant="danger" onClick={keluar} disabled={busy} className="w-full sm:w-auto"><LogOut className="size-4" /> Keluar</Button>
          </div>
        </SettingsCard>

        <SettingsCard>
          <div className="mb-4 flex items-center gap-2">
            <Lock className="size-4 text-steel" aria-hidden="true" />
            <div>
              <h2 className="text-[15px] font-medium text-fg">Keamanan &amp; Passcode</h2>
              <p className="mt-0.5 text-[13px] text-muted">Atur passcode untuk akses kasir dan keamanan sistem.</p>
            </div>
          </div>
          <div className="divide-y divide-dove rounded-xl border border-dove px-4">
            <div className="flex items-center justify-between gap-3 py-3 text-sm">
              <span className="text-muted">Status Passcode</span>
              {me == null ? (
                <span className="text-fog">—</span>
              ) : (
                <Pill tone={me.has_passcode ? 'ok' : 'muted'}>{me.has_passcode ? 'Aktif' : 'Tidak aktif'}</Pill>
              )}
            </div>
            <div className="flex items-center justify-between gap-3 py-3 text-sm">
              <span className="text-muted">Akun dilindungi</span>
              <span className="min-w-0 break-words text-right font-medium text-fg">{session?.name ?? '—'}</span>
            </div>
          </div>
          <div className="mt-4 flex flex-wrap gap-2">
            <Button onClick={() => { setManageOpen(true); setEditingId(null); setPcVal(''); setConfirmResetId(null); setErr('') }}>
              <KeyRound className="size-4" /> Manage Passcode
            </Button>
          </div>
          <div className="mt-4 flex items-start gap-2 rounded-lg bg-surface px-3.5 py-3 text-[13px] text-muted">
            <Info className="mt-0.5 size-4 shrink-0 text-steel" aria-hidden="true" />
            <p>Gunakan passcode yang kuat untuk melindungi data transaksi dan batasi akses hanya untuk kasir yang berwenang.</p>
          </div>
        </SettingsCard>
      </div>

      <SettingsCard>
        <SettingsHead title="Kata Sandi Akun" sub="Perbarui kata sandi untuk melindungi akun Anda." />
        <div className="grid gap-5 md:grid-cols-2">
          <div className="space-y-4">
            {pwdErr && <AlertStrip>{pwdErr}</AlertStrip>}
            <PwField label="Kata sandi baru" value={newPw} onChange={setNewPw} show={showNew} onToggle={() => setShowNew((v) => !v)} auto="new-password" />
            {newPw !== '' && (
              <div aria-live="polite">
                <div className="flex gap-1" aria-hidden="true">
                  {[0, 1, 2, 3].map((i) => (
                    <span key={i} className={`h-1.5 flex-1 rounded-full ${i < pwScore(newPw) ? SCORE_BAR[pwScore(newPw)] : 'bg-dove'}`} />
                  ))}
                </div>
                <p className="mt-1.5 text-xs text-muted">Kekuatan: <span className="font-medium text-fg">{SCORE_LABEL[pwScore(newPw)]}</span></p>
                <ul className="mt-2 space-y-1">
                  {[
                    { ok: newPw.length >= 8, label: 'Minimal 8 karakter' },
                    { ok: /\d/.test(newPw), label: 'Mengandung angka' },
                    { ok: /[a-z]/.test(newPw) && /[A-Z]/.test(newPw), label: 'Huruf besar dan kecil' },
                  ].map((r) => (
                    <li key={r.label} className={`flex items-center gap-1.5 text-xs ${r.ok ? 'text-sprout' : 'text-fog'}`}>
                      {r.ok ? <Check className="size-3.5" /> : <span className="grid size-3.5 place-items-center" aria-hidden="true">·</span>}
                      {r.label}
                    </li>
                  ))}
                </ul>
              </div>
            )}
            <PwField label="Ulangi kata sandi baru" value={newPw2} onChange={setNewPw2} show={showNew2} onToggle={() => setShowNew2((v) => !v)} auto="new-password" />
            {(otpSending || otpSent) && (
              <div className="space-y-2" aria-live="polite">
                <p className="text-[13px] text-muted">
                  {otpSent ? (
                    <>Kode 6 digit terkirim ke <strong className="text-fg">{session?.email.trim().toLowerCase()}</strong>. Cek email lalu masukkan di bawah. Kode berlaku 10 menit.</>
                  ) : (
                    'Mengirim kode verifikasi ke email…'
                  )}
                </p>
                <input
                  value={otp}
                  onChange={(e) => setOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  type="text" inputMode="numeric" autoComplete="one-time-code"
                  placeholder="••••••"
                  aria-label="Kode verifikasi 6 digit"
                  className="w-full rounded-md border border-border bg-paper px-3.5 py-3 text-center font-mono text-xl tracking-[0.5em] focus:border-jet focus:outline-none"
                />
                {otpRequested && (
                  <button
                    type="button"
                    onClick={() => sendPwOtp()}
                    disabled={cooldown > 0 || pwdBusy || otpSending}
                    className="text-[13px] font-medium text-jet hover:underline disabled:opacity-50"
                  >
                    {cooldown > 0 ? `Kirim ulang dalam ${cooldown} detik` : 'Kirim ulang kode'}
                  </button>
                )}
              </div>
            )}
            <Button onClick={submitPwReset} disabled={pwdBusy || newPw.length < 8 || newPw !== newPw2 || !otpSent || otp.length !== 6} className="w-full sm:w-auto">
              {pwdBusy ? 'Menyimpan…' : 'Simpan kata sandi baru'}
            </Button>
          </div>
          <div className="h-fit rounded-xl bg-surface p-4 text-[13px]">
            <p className="flex items-center gap-2 font-medium text-fg"><ShieldCheck className="size-4 text-steel" /> Cara mengganti kata sandi</p>
            <ol className="mt-2.5 list-decimal space-y-1.5 pl-5 text-muted">
              <li>Tulis kata sandi baru di form sebelah kiri, lalu ulangi sekali lagi.</li>
              <li>Kode verifikasi terkirim otomatis ke email Anda begitu kata sandi barunya valid. Buka email Anda.</li>
              <li>Masukkan kode 6 digit yang Anda terima, lalu tekan Simpan kata sandi baru.</li>
              <li>Selesai. Anda akan keluar otomatis dan bisa masuk lagi dengan kata sandi yang baru.</li>
            </ol>
            <div className="mt-3 flex items-start gap-2 rounded-lg bg-paper px-3 py-2.5 text-muted">
              <Info className="mt-0.5 size-4 shrink-0 text-steel" aria-hidden="true" />
              <p>Tips: pakai kata sandi yang belum pernah dipakai di aplikasi lain. Kode hanya berlaku 10 menit, jadi langsung dipakai begitu diterima.</p>
            </div>
          </div>
        </div>
      </SettingsCard>
    </div>
    <Modal open={manageOpen} title="Manage Passcode" onClose={() => { setManageOpen(false); setEditingId(null); setPcVal(''); setConfirmResetId(null) }}>
      <p className="mb-3 text-sm text-muted">Atur passcode 5 angka untuk admin dan setiap kasir. Akun tanpa passcode bisa langsung dipakai tanpa PIN.</p>
      {accountList.loading && manageUsers.length === 0 ? (
        <p className="py-4 text-center text-sm text-fog">Memuat daftar akun…</p>
      ) : manageUsers.length === 0 ? (
        <p className="py-4 text-center text-sm text-fog">{accountList.err || 'Daftar akun tidak tersedia.'}</p>
      ) : (
        <div className="space-y-2">
          {manageUsers.map((u) => (
            <div key={u.id} className="rounded-lg border border-dove px-3.5 py-2.5">
              <div className="flex flex-wrap items-center gap-2">
                <span className="min-w-0 flex-1 truncate text-sm font-medium text-fg">
                  {u.name} <span className="font-normal text-fog">· {u.role === 'admin' ? 'Admin' : 'Kasir'}</span>
                </span>
                <Pill tone={u.has_passcode ? 'ok' : 'muted'}>{u.has_passcode ? 'Aktif' : 'Mati'}</Pill>
                {editingId === u.id ? (
                  <button onClick={() => { setEditingId(null); setPcVal('') }} className="text-[13px] text-muted hover:underline">Batal</button>
                ) : (
                  <RowAction title={u.has_passcode ? 'Ubah passcode' : 'Tambah passcode'} label={`${u.has_passcode ? 'Ubah' : 'Tambah'} passcode ${u.name}`} onClick={() => { setEditingId(u.id); setPcVal(''); setConfirmResetId(null); setErr('') }} disabled={pcBusy}><KeyRound className="size-4" /></RowAction>
                )}
                {u.has_passcode && editingId !== u.id && (
                  confirmResetId === u.id ? (
                    <>
                      <button onClick={() => clearPc(u)} disabled={pcBusy} className="text-[13px] font-medium text-ember hover:underline disabled:opacity-40">
                        {pcBusy ? '…' : 'Yakin, hapus'}
                      </button>
                      <button onClick={() => setConfirmResetId(null)} className="text-[13px] text-muted hover:underline">Batal</button>
                    </>
                  ) : (
                    <RowAction title="Hapus passcode" label={`Hapus passcode ${u.name}`} onClick={() => setConfirmResetId(u.id)} disabled={pcBusy} danger><Trash2 className="size-4" /></RowAction>
                  )
                )}
              </div>
              {editingId === u.id && (
                <div className="mt-2.5 flex flex-wrap gap-2">
                  <input
                    value={pcVal}
                    onChange={(e) => setPcVal(e.target.value.replace(/\D/g, '').slice(0, 5))}
                    inputMode="numeric"
                    placeholder="•••••"
                    aria-label={`Passcode baru untuk ${u.name}`}
                    autoFocus
                    className="min-w-0 flex-1 rounded-md border border-border bg-paper px-3 py-2.5 text-center font-mono text-lg tracking-[0.5em] focus:border-jet focus:outline-none"
                  />
                  <Button onClick={() => savePc(u)} disabled={pcBusy || pcVal.length !== 5}>{pcBusy ? '…' : 'Simpan'}</Button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </Modal>
    </>
  )
}
