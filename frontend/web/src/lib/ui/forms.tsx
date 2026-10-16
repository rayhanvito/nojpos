import type { ReactNode } from 'react'
import { useEffect, useRef, useState } from 'react'
import { Check, ChevronDown, Search, SlidersHorizontal, X } from 'lucide-react'
import { useSheetMode, SHEET_BACKDROP } from './sheet.ts'

export function Input({  label, value, onChange, type = 'text', placeholder, required, hint,
}: {
  label: string
  value: string | number
  onChange: (v: string) => void
  type?: string
  placeholder?: string
  required?: boolean
  hint?: string
}) {
  return (
    <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
      {label}
      <input
        type={type}
        value={value}
        required={required}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        className="h-[42px] w-full rounded-md border border-border bg-paper px-3.5 py-0 text-[15px] text-fg placeholder:text-fog focus:border-jet focus:outline-2 focus:outline-accent-soft"
      />
      {hint && <span className="text-xs font-normal text-fog">{hint}</span>}
    </label>
  )
}

// Input angka dengan pengelompokan ribuan otomatis (1000 → 1.000).
// Titik/koma yang diketik manual diabaikan sebagai pemisah ribuan; nilai
// mentah (hanya digit, desimal pakai titik) diteruskan via onValue.
// allowDecimal = dukung desimal (mis. pajak 11,5%) — tampil pakai koma.
export function NumInput({
  label, value, onValue, placeholder, required, disabled, hint, allowDecimal, autoFocus, className,
}: {
  label?: string
  value: string | number
  onValue: (raw: string) => void
  placeholder?: string
  required?: boolean
  disabled?: boolean
  hint?: string
  allowDecimal?: boolean
  autoFocus?: boolean
  className?: string
}) {
  const raw = String(value ?? '')
  const parts = raw.split('.')
  const intDigits = (parts[0] ?? '').replace(/\D/g, '')
  const fracDigits = (parts[1] ?? '').replace(/\D/g, '').slice(0, 2)
  const display = intDigits.replace(/\B(?=(\d{3})+(?!\d))/g, '.') + (allowDecimal && fracDigits ? `,${fracDigits}` : '')

  function handle(e: React.ChangeEvent<HTMLInputElement>) {
    let v = e.target.value.replace(/[^\d.,]/g, '')
    if (!allowDecimal) {
      onValue(v.replace(/\D/g, ''))
      return
    }
    v = v.replace(/,/g, '.')
    const seg = v.split('.')
    if (seg.length > 2) v = seg[0] + '.' + seg.slice(1).join('')
    onValue(v)
  }

  return (
    <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
      {label}
      <input
        inputMode={allowDecimal ? 'decimal' : 'numeric'}
        value={display}
        placeholder={placeholder}
        required={required}
        disabled={disabled}
        autoFocus={autoFocus}
        onChange={handle}
        className={`h-[42px] w-full rounded-md border border-border bg-paper px-3.5 py-0 text-[15px] text-fg placeholder:text-fog focus:border-jet focus:outline-none disabled:opacity-50 ${className ?? ''}`}
      />
      {hint && <span className="text-xs font-normal text-fog">{hint}</span>}
    </label>
  )
}

export function Select({ value, options, onChange, label, placeholder = 'Pilih', className = '', disabled = false, dense = false }: {
  value: string
  options: { value: string; label: string }[]
  onChange: (value: string) => void
  label: string
  placeholder?: string
  className?: string
  disabled?: boolean
  dense?: boolean
}) {
  const [open, setOpen] = useState(false)
  const box = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const close = (e: PointerEvent) => { if (!box.current?.contains(e.target as Node)) setOpen(false) }
    const esc = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false) }
    document.addEventListener('pointerdown', close)
    document.addEventListener('keydown', esc)
    return () => {
      document.removeEventListener('pointerdown', close)
      document.removeEventListener('keydown', esc)
    }
  }, [open ])

  const sheet = useSheetMode()
  const active = options.find((o) => o.value === value)
  // Auto-flip (keputusan Rayhan 16 Okt 2026): popover desktop membuka ke atas
  // bila ruang di bawah pemicu sempit (mis. Select di dasar halaman/modal).
  const [flip, setFlip] = useState(false)
  const toggle = () => {
    if (!open && !sheet && box.current) {
      const r = box.current.getBoundingClientRect()
      const below = window.innerHeight - r.bottom
      setFlip(below < 256 && r.top > below)
    }
    setOpen((o) => !o)
  }

  return (
    <div ref={box} className={`relative ${className}`}>
      <button
        type="button" aria-label={label} aria-expanded={open} aria-haspopup="listbox" disabled={disabled}
        onClick={toggle}
        className={`flex w-full items-center gap-2.5 rounded-md border border-border bg-paper transition outline-none hover:border-jet focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-60 ${dense ? 'h-[38px] px-3 py-2 text-[13px]' : 'h-[42px] px-3.5 py-0 text-sm'}`}
      >
        <span className={`min-w-0 flex-1 truncate text-left ${active ? 'text-fg' : 'text-fog'}`}>{active?.label ?? placeholder}</span>
        <ChevronDown className="size-4 shrink-0 text-fog" />
      </button>
      {open && sheet && (
        <div className={SHEET_BACKDROP} onClick={() => setOpen(false)} aria-hidden="true" />
      )}
      {open && (
        <div className={sheet ? 'fixed inset-x-0 bottom-0 z-50 flex max-h-[85dvh] flex-col rounded-t-3xl bg-paper pb-5 shadow-xl' : `absolute right-0 z-50 w-full min-w-44 rounded-xl border border-dove bg-paper shadow-lg ${flip ? 'bottom-full mb-2' : 'top-full mt-2'}`}>
          {sheet && (
            <div className="flex items-center justify-between px-5 pb-1 pt-4">
              <p className="text-[15px] font-bold text-fg">{label}</p>
              <button type="button" onClick={() => setOpen(false)} aria-label="Tutup" className="grid size-8 shrink-0 place-items-center rounded-md text-muted transition outline-none hover:bg-surface hover:text-fg">
                <X className="size-4" aria-hidden="true" />
              </button>
            </div>
          )}
          <div role="listbox" aria-label={label} className={sheet ? 'min-h-0 overflow-y-auto px-2' : 'max-h-64 overflow-y-auto p-1.5'}>
            {options.map((o) => (
              <button
                key={o.value} type="button" role="option" aria-selected={o.value === value}
                onClick={() => { onChange(o.value); setOpen(false) }}
                className={`flex w-full items-center gap-2 rounded-lg px-3 py-2.5 text-left text-sm transition outline-none hover:bg-surface focus-visible:ring-2 focus-visible:ring-ring ${sheet ? 'min-h-[44px] text-[15px]' : ''} ${o.value === value ? 'bg-surface font-medium text-fg' : 'text-fg/80'}`}
              >
                <span className="min-w-0 flex-1 truncate">{o.label}</span>
                {o.value === value && <Check className="size-4 shrink-0" aria-hidden="true" />}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

// Pola filter lanjutan bersama (keputusan Rayhan 2026-10-15): satu halaman
// maksimal punya search + 1 Select utama + tombol Filter ini. Dimensi
// pendamping tinggal di dalam panel; status aktif terlihat dari badge di
// tombol dan chip di bawah toolbar (FilterChips).
// Kolom pencarian toolbar bersama: ikon + input polos (pola yang dulu disalin
// di 12 tempat). Membawa atribut input apa adanya (value/onChange/placeholder/
// aria-label) agar perilaku tiap halaman tidak berubah.

export function SearchInput({ className = '', ...props }: React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <div className={`relative min-w-0 flex-1 ${className}`}>
      <Search className="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-fog" />
      <input
        type="search"
        {...props}
        className="h-[42px] w-full rounded-md border border-border bg-paper py-0 pl-10 pr-3.5 text-sm text-fg placeholder:text-fog focus:border-jet focus:outline-none"
      />
    </div>
  )
}

export function FilterPanel({ activeCount, onReset, children, className = '' }: {
  activeCount: number
  onReset: () => void
  children: ReactNode
  className?: string
}) {
  const [open, setOpen] = useState(false)
  const box = useRef<HTMLDivElement>(null)
  const sheet = useSheetMode()

  useEffect(() => {
    if (!open) return
    const close = (e: PointerEvent) => { if (!box.current?.contains(e.target as Node)) setOpen(false) }
    const esc = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false) }
    document.addEventListener('pointerdown', close)
    document.addEventListener('keydown', esc)
    return () => {
      document.removeEventListener('pointerdown', close)
      document.removeEventListener('keydown', esc)
    }
  }, [open ])

  return (
    <div ref={box} className={`relative ${className}`}>
      <button
        type="button" aria-label="Filter lanjutan" aria-expanded={open} aria-haspopup="dialog"
        onClick={() => setOpen((o) => !o)}
        className="flex h-[42px] w-full items-center gap-2.5 rounded-md border border-border bg-paper px-3.5 py-0 text-sm transition outline-none hover:border-jet focus-visible:ring-2 focus-visible:ring-ring"
      >
        <SlidersHorizontal className="size-4 shrink-0 text-fog" />
        <span className="min-w-0 flex-1 truncate text-left text-fg">Filter</span>
        {activeCount > 0 && (
          <span className="shrink-0 rounded-full bg-jet px-1.5 py-0.5 font-mono text-[11px] leading-none text-paper">{activeCount}</span>
        )}
        <ChevronDown className="size-4 shrink-0 text-fog" />
      </button>
      {open && sheet && (
        <div className={SHEET_BACKDROP} onClick={() => setOpen(false)} aria-hidden="true" />
      )}
      {open && (
        <div role="dialog" aria-label="Filter lanjutan" className={sheet ? 'fixed inset-x-0 bottom-0 z-50 max-h-[85dvh] overflow-y-auto rounded-t-3xl bg-paper p-4 pb-6 shadow-xl' : 'absolute right-0 top-full z-50 mt-2 w-72 max-w-[calc(100vw-2rem)] rounded-xl border border-dove bg-paper p-3.5 shadow-lg'}>
          {sheet && (
            <div className="mb-3 flex items-center justify-between">
              <p className="text-[15px] font-bold text-fg">Filter</p>
              <button type="button" onClick={() => setOpen(false)} aria-label="Tutup" className="grid size-8 shrink-0 place-items-center rounded-md text-muted transition hover:bg-surface hover:text-fg">
                <X className="size-4" aria-hidden="true" />
              </button>
            </div>
          )}
          <div className="space-y-3.5">{children}</div>
          <button
            type="button" onClick={onReset} disabled={activeCount === 0}
            className="mt-4 w-full rounded-md border border-dove px-3 py-2 text-sm text-muted transition hover:border-jet hover:text-fg disabled:opacity-40"
          >
            Reset filter
          </button>
        </div>
      )}
    </div>
  )
}

export function FilterField({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div>
      <p className="mb-1.5 font-mono text-[11px] uppercase tracking-wider text-fog">{label}</p>
      {children}
    </div>
  )
}

export function FilterChips({ items }: { items: { key: string; label: string; onClear: () => void }[] }) {
  if (items.length === 0) return null
  return (
    <div className="mb-4 flex flex-wrap items-center gap-2">
      {items.map((it) => (
        <span key={it.key} className="inline-flex items-center gap-1.5 rounded-full border border-dove bg-paper py-1 pl-3 pr-1.5 text-[13px] text-fg">
          {it.label}
          <button
            type="button" aria-label={`Hapus filter ${it.label}`} onClick={it.onClear}
            className="grid size-5 place-items-center rounded-full text-fog transition hover:bg-surface hover:text-fg"
          >
            <X className="size-3.5" />
          </button>
        </span>
      ))}
    </div>
  )
}

// Textarea berlabel baku — pasangan Input untuk teks panjang (catatan,
// keterangan). Dulu ditulis tangan per halaman dengan dua varian padding.
export function TextArea({ label, value, onChange, rows = 3, placeholder, required = false }: {
  label: string
  value: string
  onChange: (v: string) => void
  rows?: number
  placeholder?: string
  required?: boolean
}) {
  return (
    <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
      {label}
      <textarea
        value={value}
        required={required}
        placeholder={placeholder}
        rows={rows}
        onChange={(e) => onChange(e.target.value)}
        className="h-[42px] w-full rounded-md border border-border bg-paper px-3.5 py-0 text-[15px] font-normal text-fg placeholder:text-fog focus:border-jet focus:outline-2 focus:outline-accent-soft"
      />
    </label>
  )
}

// Baris centang baku (kotak + label, sejajar tengah/atas) untuk modal & form —
// menggantikan label checkbox tulis-tangan dengan jarak/ukuran yang sama.
export function CheckRow({ checked, onChange, disabled = false, align = 'center', className = '', children }: {
  checked: boolean
  onChange: (v: boolean) => void
  disabled?: boolean
  align?: 'center' | 'start'
  className?: string
  children: ReactNode
}) {
  return (
    <label className={`flex cursor-pointer gap-2.5 text-[13px] font-medium text-steel ${align === 'start' ? 'items-start' : 'items-center'} ${className}`.trim()}>
      <input
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
        className={`${align === 'start' ? 'mt-0.5 ' : ''}h-4 w-4 shrink-0 accent-jet`}
      />
      <span>{children}</span>
    </label>
  )
}
