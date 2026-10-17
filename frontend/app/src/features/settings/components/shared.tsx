import type { ReactNode } from 'react'
import { Eye, EyeOff } from 'lucide-react'
import type { StoreHours } from '../../../shared/api'
import { Button, Pill } from '../../../shared/ui'

export const TIMEZONES = [
  { value: 'Asia/Jakarta', label: 'WIB (UTC+7)' },
  { value: 'Asia/Makassar', label: 'WITA (UTC+8)' },
  { value: 'Asia/Jayapura', label: 'WIT (UTC+9)' },
]

// Label ramah untuk nilai IANA (dipakai di badge preview); fallback ke nilai mentah.
export function tzLabel(v: string) {
  return TIMEZONES.find((t) => t.value === v)?.label ?? v
}

export const INPUT_CLS = 'h-[42px] w-full rounded-md border border-border bg-paper px-3.5 py-0 text-[15px] text-fg placeholder:text-fog focus:border-jet focus:outline-2 focus:outline-accent-soft disabled:opacity-60'
export const LABEL_CLS = 'flex flex-col gap-1.5 text-[13px] font-medium text-steel'
// Settings extended (kontrak docs/API-CONTRACT-SETTINGS-EXTENDED.md, live
// backend 11 Sep): opsi + rumus pajak server (inclusive + rounding).
export const BUSINESS_TYPES = [
  { value: '', label: 'Pilih jenis usaha…' },
  { value: 'retail', label: 'Retail / Kelontong' },
  { value: 'fnb', label: 'Makanan & Minuman' },
  { value: 'fashion', label: 'Fashion' },
  { value: 'jasa', label: 'Jasa' },
  { value: 'lainnya', label: 'Lainnya' },
]

export const HOURS_LABELS = ['Senin – Jumat', 'Sabtu', 'Minggu']

// Rumus pajak = backend repo/transaction.go: inclusive → total tetap = base,
// tax = base − base/(1+pct/100); rounding none = round_half_up, down = floor,
// up = ceil.
export function calcTax(base: number, pct: number, inclusive: boolean, rounding: string) {
  if (pct <= 0) return { tax: 0, total: base }
  const raw = inclusive ? base - base / (1 + pct / 100) : (base * pct) / 100
  const tax = rounding === 'down' ? Math.floor(raw) : rounding === 'up' ? Math.ceil(raw) : Math.round(raw)
  return { tax, total: inclusive ? base : base + tax }
}

export function defaultHours(): StoreHours[] {
  return [
    { days: HOURS_LABELS[0], open: '08:00', close: '21:00' },
    { days: HOURS_LABELS[1], open: '08:00', close: '21:00' },
    { days: HOURS_LABELS[2], open: null, close: null },
  ]
}

export function todayStr() {
  const t = new Date()
  return `${t.getFullYear()}-${String(t.getMonth() + 1).padStart(2, '0')}-${String(t.getDate()).padStart(2, '0')}`
}

export function MetaRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 py-2.5 text-sm">
      <span className="shrink-0 text-muted">{label}</span>
      <span className="min-w-0 break-words text-right font-medium text-fg">{value}</span>
    </div>
  )
}

// Field yang belum didukung backend: tampil nonaktif + kontrak
// docs/API-CONTRACT-SETTINGS-EXTENDED.md.
export function Soon({ children }: { children: ReactNode }) {
  return (
    <div className="opacity-80">
      <div className="mb-1.5 flex items-center gap-2">
        <div className="flex-1">{children}</div>
        <Pill tone="muted">Segera</Pill>
      </div>
      <p className="-mt-0.5 mb-1 text-xs font-normal text-fog">Segera hadir — menunggu backend.</p>
    </div>
  )
}

export function FormActions({ dirty, busy, onReset, onSave }: { dirty: boolean; busy: boolean; onReset: () => void; onSave: () => void }) {
  return (
    <div className="mt-5 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
      <Button variant="ghost" onClick={onReset} disabled={busy || !dirty}>Reset</Button>
      <Button onClick={onSave} disabled={busy || !dirty}>{busy ? 'Menyimpan…' : 'Simpan Perubahan'}</Button>
    </div>
  )
}

export function PwField({ label, value, onChange, show, onToggle, auto }: { label: string; value: string; onChange: (v: string) => void; show: boolean; onToggle: () => void; auto: string }) {
  return (
    <label className={LABEL_CLS}>
      {label}
      <span className="relative block">
        <input
          type={show ? 'text' : 'password'}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          autoComplete={auto}
          className={`${INPUT_CLS} pr-11`}
        />
        <button
          type="button"
          onClick={onToggle}
          aria-label={show ? `Sembunyikan ${label}` : `Tampilkan ${label}`}
          className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1.5 text-fog hover:text-fg"
        >
          {show ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
        </button>
      </span>
    </label>
  )
}

export function pwScore(pw: string) {
  let s = 0
  if (pw.length >= 8) s += 1
  if (pw.length >= 12) s += 1
  if (/\d/.test(pw) && /[a-zA-Z]/.test(pw)) s += 1
  if (/[a-z]/.test(pw) && /[A-Z]/.test(pw) && /[^a-zA-Z0-9]/.test(pw)) s += 1
  return Math.min(s, 4)
}

export const SCORE_LABEL = ['Lemah', 'Lemah', 'Cukup', 'Kuat', 'Sangat kuat']
export const SCORE_BAR = ['bg-ember', 'bg-ember', 'bg-sunbeam', 'bg-sprout', 'bg-sprout']

export const SAMPLE_ITEMS = [
  { name: 'Indomie Goreng', qty: 2, price: 12000 },
  { name: 'Aqua 600ml', qty: 1, price: 5000 },
  { name: 'Roti Tawar', qty: 1, price: 15000 },
]
