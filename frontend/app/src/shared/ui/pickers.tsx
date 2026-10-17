import { useEffect, useRef, useState } from 'react'
import { CalendarDays, ChevronDown, ChevronLeft, ChevronRight, Clock, X } from 'lucide-react'
import { fmtDate } from '../session/store.ts'
import { Button } from './primitives.tsx'
import { useSheetMode, SHEET_BACKDROP } from './sheet.ts'

const MONTHS_ID = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
const WEEKDAYS_ID = ['Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab', 'Min']

function isoDay(y: number, m: number, d: number) {
  return `${y}-${String(m + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`
}

function parseISODate(s: string): [number, number, number] | null {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(s)
  if (!m) return null
  return [Number(m[1]), Number(m[2]) - 1, Number(m[3])]
}

function todayParts(): [number, number, number] {
  const t = new Date()
  return [t.getFullYear(), t.getMonth(), t.getDate()]
}

export function DatePicker({ value, onChange, label = 'Pilih tanggal', placeholder = 'Pilih tanggal' }: {
  value: string
  onChange: (v: string) => void
  label?: string
  placeholder?: string
}) {
  const [open, setOpen] = useState(false)
  const today = todayParts()
  const parsed = parseISODate(value)
  const [view, setView] = useState<[number, number]>([parsed?.[0] ?? today[0], parsed?.[1] ?? today[1]])
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
  const [vy, vm] = view
  const todayISO = isoDay(today[0], today[1], today[2])
  const offset = (new Date(vy, vm, 1).getDay() + 6) % 7
  const dim = new Date(vy, vm + 1, 0).getDate()
  const prevDim = new Date(vy, vm, 0).getDate()
  const cells: { key: string; d: number; iso: string; outside: boolean }[] = []
  for (let i = offset - 1; i >= 0; i--) {
    const d = prevDim - i
    cells.push({ key: `p${d}`, d, iso: isoDay(vm === 0 ? vy - 1 : vy, (vm + 11) % 12, d), outside: true })
  }
  for (let d = 1; d <= dim; d++) cells.push({ key: `c${d}`, d, iso: isoDay(vy, vm, d), outside: false })
  for (let d = 1; cells.length % 7 !== 0; d++) {
    cells.push({ key: `n${d}`, d, iso: isoDay(vm === 11 ? vy + 1 : vy, (vm + 1) % 12, d), outside: true })
  }

  function pick(iso: string) {
    onChange(iso)
    setOpen(false)
  }

  function openMenu() {
    const p = parseISODate(value)
    setView([p?.[0] ?? today[0], p?.[1] ?? today[1]])
    setOpen((o) => !o)
  }

  return (
    <div ref={box} className="relative">
      <button
        type="button" aria-label={label} aria-expanded={open} onClick={openMenu}
        className="flex h-[42px] w-full items-center gap-2.5 rounded-md border border-border bg-paper px-3.5 py-0 text-sm transition outline-none hover:border-jet focus-visible:ring-2 focus-visible:ring-ring"
      >
        <CalendarDays className="size-4 shrink-0 text-fog" />
        <span className={`min-w-0 flex-1 truncate text-left tabular-nums ${value ? 'text-fg' : 'text-fog'}`}>
          {value ? fmtDate(value) : placeholder}
        </span>
        {value ? (
          <span
            role="button" tabIndex={0} aria-label="Hapus tanggal"
            onClick={(e) => { e.stopPropagation(); onChange(''); }}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); e.stopPropagation(); onChange(''); } }}
            className="grid shrink-0 place-items-center rounded-full p-0.5 text-fog transition outline-none hover:bg-surface hover:text-fg focus-visible:ring-2 focus-visible:ring-ring"
          >
            <X className="size-3.5" />
          </span>
        ) : (
          <ChevronDown className="size-4 shrink-0 text-fog" />
        )}
      </button>

      {open && sheet && (
        <div className={SHEET_BACKDROP} onClick={() => setOpen(false)} aria-hidden="true" />
      )}
      {open && (
        <div role="dialog" aria-label="Kalender" className={sheet ? 'fixed inset-x-0 bottom-0 z-50 max-h-[88dvh] overflow-y-auto rounded-t-3xl bg-paper p-4 pb-6 shadow-xl' : 'absolute right-0 top-full z-50 mt-2 w-72 max-w-[calc(100vw-2rem)] rounded-xl border border-dove bg-paper p-3 shadow-lg'}>
          {sheet && (
            <div className="mb-2 flex items-center justify-between">
              <p className="text-[15px] font-bold text-fg">{label}</p>
              <button type="button" onClick={() => setOpen(false)} aria-label="Tutup" className="grid size-8 shrink-0 place-items-center rounded-md text-muted transition hover:bg-surface hover:text-fg">
                <X className="size-4" aria-hidden="true" />
              </button>
            </div>
          )}
          <div className="flex items-center justify-between">
            <button
              type="button" aria-label="Bulan sebelumnya"
              onClick={() => setView([vm === 0 ? vy - 1 : vy, (vm + 11) % 12])}
              className="grid size-8 place-items-center rounded-lg text-muted transition outline-none hover:bg-surface hover:text-fg focus-visible:ring-2 focus-visible:ring-ring"
            >
              <ChevronLeft className="size-4" />
            </button>
            <p className="text-sm font-medium">{MONTHS_ID[vm]} {vy}</p>
            <button
              type="button" aria-label="Bulan berikutnya"
              onClick={() => setView([vm === 11 ? vy + 1 : vy, (vm + 1) % 12])}
              className="grid size-8 place-items-center rounded-lg text-muted transition outline-none hover:bg-surface hover:text-fg focus-visible:ring-2 focus-visible:ring-ring"
            >
              <ChevronRight className="size-4" />
            </button>
          </div>
          <div className="mt-2 grid grid-cols-7 text-center font-mono text-[11px] uppercase text-fog">
            {WEEKDAYS_ID.map((d) => <span key={d} className="py-1">{d}</span>)}
          </div>
          <div className="grid grid-cols-7 gap-0.5">
            {cells.map((c) => {
              const selected = c.iso === value
              const isToday = c.iso === todayISO
              return (
                <button
                  key={c.key} type="button" aria-label={fmtDate(c.iso)} aria-pressed={selected}
                  onClick={() => { setView([Number(c.iso.slice(0, 4)), Number(c.iso.slice(5, 7)) - 1]); pick(c.iso) }}
                  className={`grid size-9 place-items-center rounded-full text-[13px] tabular-nums transition outline-none focus-visible:ring-2 focus-visible:ring-ring ${selected
                    ? 'bg-jet font-medium text-paper'
                    : isToday
                      ? 'border border-jet font-medium text-fg hover:bg-surface'
                      : `border border-transparent hover:bg-surface ${c.outside ? 'text-fog/60' : 'text-fg'}`}`}
                >
                  {c.d}
                </button>
              )
            })}
          </div>
          {value && (
            <div className="mt-2 flex items-center justify-end border-t border-dove pt-2">
              <button type="button" onClick={() => { onChange(''); setOpen(false) }} className="text-[13px] text-ember hover:underline">
                Hapus filter
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// Baris skeleton dalam tabel asli (pakai setelah <thead> saat data null)
// agar kolom & padding sama persis dengan isi sebenarnya.

function monthCells(y: number, m: number) {
  const offset = (new Date(y, m, 1).getDay() + 6) % 7
  const dim = new Date(y, m + 1, 0).getDate()
  const prevDim = new Date(y, m, 0).getDate()
  const cells: { key: string; d: number; iso: string; outside: boolean }[] = []
  for (let i = offset - 1; i >= 0; i--) {
    const d = prevDim - i
    cells.push({ key: `p${d}`, d, iso: isoDay(m === 0 ? y - 1 : y, (m + 11) % 12, d), outside: true })
  }
  for (let d = 1; d <= dim; d++) cells.push({ key: `c${d}`, d, iso: isoDay(y, m, d), outside: false })
  for (let d = 1; cells.length % 7 !== 0; d++) {
    cells.push({ key: `n${d}`, d, iso: isoDay(m === 11 ? y + 1 : y, (m + 1) % 12, d), outside: true })
  }
  return cells
}

// Pemilih periode model kalender rentang dua bulan + rel preset (patokan visual
// dari Rayhan 15 Okt 2026, gaya diselaraskan ke tema project). Klik tanggal mulai
// lalu tanggal akhir; preset di rel kiri langsung menerapkan.

// Mode sheet untuk panel popover di layar kecil (fase mobile): Select,
// FilterPanel, dan PeriodPicker membuka sebagai bottom sheet pada <640px —
// bahasa yang sama dengan Modal — dan popover mengambang di layar besar.

// Kolom tanggal+jam berlabel baku untuk form (mis. masa berlaku promo):
// tanggal memakai DatePicker bersama (bukan input natif mm/dd/yyyy), jam
// tetap input waktu dalam gaya field yang sama. Nilai 'YYYY-MM-DDTHH:mm'.
export function DateTimeField({ label, value, onChange, required = false }: {
  label: string
  value: string
  onChange: (v: string) => void
  required?: boolean
}) {
  const date = value ? value.slice(0, 10) : ''
  const time = value && value.length >= 16 ? value.slice(11, 16) : ''
  function emit(d: string, tm: string) {
    onChange(d ? `${d}T${tm || '00:00'}` : '')
  }
  return (
    <div className="flex flex-col gap-1.5">
      <span className="text-[13px] font-medium text-steel">{label}</span>
      <div className="flex gap-2">
        <div className="min-w-0 flex-1">
          <DatePicker value={date} onChange={(d) => emit(d, time)} label={label} placeholder="Pilih tanggal" />
        </div>
        <input
          type="time"
          value={time}
          required={required && !!date}
          onChange={(e) => emit(date, e.target.value)}
          aria-label={`${label} (jam)`}
          className="h-[42px] w-32 shrink-0 rounded-md border border-border bg-paper px-3 py-0 text-[15px] text-fg focus:border-jet focus:outline-2 focus:outline-accent-soft"
        />
      </div>
    </div>
  )
}


// Pemilih jam bersama (pengganti <input type="time"> natif, mis. jam mulai/
// berakhir harian di form Promo) — bahasa visual sama dengan DatePicker:
// tombol berbingkai 42px + panel (bottom sheet di <640px, popover di
// desktop). Nilai string 'HH:MM' persis ('' = kosong); pasangan field tetap
// opsional dan urutan nilai tidak pernah divalidasi di sini — pemanggil
// boleh menyimpan mulai > berakhir (lintas tengah malam). Memilih jam lalu
// menit (atau sebaliknya) langsung menerapkan; pilihan sebelah yang sudah
// ada dipertahankan, jadi mengubah jam saja tidak mengutak-atik menit.
// Kolom menit berisi 00–59 penuh agar presisi nilai sama seperti input
// natif yang digantikan. Kosongkan lewat ikon X di pemicu atau tombol
// "Hapus jam" di panel.
export function TimePicker({ value, onChange, label = 'Pilih jam', placeholder = 'Pilih jam' }: {
  value: string
  onChange: (v: string) => void
  label?: string
  placeholder?: string
}) {
  const [open, setOpen] = useState(false)
  const [draftH, setDraftH] = useState<number | null>(null)
  const [draftM, setDraftM] = useState<number | null>(null)
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

  const parsed = /^(\d{2}):(\d{2})$/.exec(value)
  const curH = parsed ? Number(parsed[1]) : null
  const curM = parsed ? Number(parsed[2]) : null

  function toggle() {
    if (!open) {
      setDraftH(curH)
      setDraftM(curM)
    }
    setOpen((o) => !o)
  }

  function pick(h: number | null, m: number | null) {
    setDraftH(h)
    setDraftM(m)
    if (h !== null && m !== null) {
      onChange(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`)
      setOpen(false)
    }
  }

  function column(title: string, count: number, selected: number | null, onPick: (n: number) => void) {
    return (
      <div className="flex min-w-0 flex-1 flex-col">
        <p className="mb-1.5 text-center font-mono text-[11px] uppercase tracking-wider text-fog">{title}</p>
        <div role="listbox" aria-label={`${label} — ${title}`} className="max-h-56 overflow-y-auto">
          {Array.from({ length: count }, (_, n) => {
            const isSel = n === selected
            return (
              <button
                key={n} type="button" role="option" aria-selected={isSel}
                ref={isSel ? (el) => { if (el) { const c = el.parentElement; if (c) c.scrollTop = el.offsetTop - c.clientHeight / 2 + el.offsetHeight / 2 } } : undefined}
                onClick={() => onPick(n)}
                className={`flex w-full items-center justify-center rounded-lg px-3 py-2.5 text-sm tabular-nums transition outline-none hover:bg-surface focus-visible:ring-2 focus-visible:ring-ring ${sheet ? 'min-h-[44px] text-[15px]' : ''} ${isSel ? 'bg-jet font-medium text-paper hover:bg-jet' : 'text-fg/80'}`}
              >
                {String(n).padStart(2, '0')}
              </button>
            )
          })}
        </div>
      </div>
    )
  }

  return (
    <div ref={box} className="relative">
      <button
        type="button" aria-label={label} aria-expanded={open} onClick={toggle}
        className="flex h-[42px] w-full items-center gap-2.5 rounded-md border border-border bg-paper px-3.5 py-0 text-sm transition outline-none hover:border-jet focus-visible:ring-2 focus-visible:ring-ring"
      >
        <Clock className="size-4 shrink-0 text-fog" />
        <span className={`min-w-0 flex-1 truncate text-left tabular-nums ${value ? 'text-fg' : 'text-fog'}`}>
          {value || placeholder}
        </span>
        {value ? (
          <span
            role="button" tabIndex={0} aria-label="Hapus jam"
            onClick={(e) => { e.stopPropagation(); onChange('') }}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); e.stopPropagation(); onChange('') } }}
            className="grid shrink-0 place-items-center rounded-full p-0.5 text-fog transition outline-none hover:bg-surface hover:text-fg focus-visible:ring-2 focus-visible:ring-ring"
          >
            <X className="size-3.5" />
          </span>
        ) : (
          <ChevronDown className="size-4 shrink-0 text-fog" />
        )}
      </button>

      {open && sheet && (
        <div className={SHEET_BACKDROP} onClick={() => setOpen(false)} aria-hidden="true" />
      )}
      {open && (
        <div role="dialog" aria-label={label} className={sheet ? 'fixed inset-x-0 bottom-0 z-50 max-h-[88dvh] overflow-y-auto rounded-t-3xl bg-paper p-4 pb-6 shadow-xl' : 'absolute right-0 top-full z-50 mt-2 w-64 max-w-[calc(100vw-2rem)] rounded-xl border border-dove bg-paper p-3 shadow-lg'}>
          {sheet && (
            <div className="mb-2 flex items-center justify-between">
              <p className="text-[15px] font-bold text-fg">{label}</p>
              <button type="button" onClick={() => setOpen(false)} aria-label="Tutup" className="grid size-8 shrink-0 place-items-center rounded-md text-muted transition hover:bg-surface hover:text-fg">
                <X className="size-4" aria-hidden="true" />
              </button>
            </div>
          )}
          <div className="flex gap-3">
            {column('Jam', 24, draftH, (n) => pick(n, draftM))}
            {column('Menit', 60, draftM, (n) => pick(draftH, n))}
          </div>
          <div className="mt-2 flex items-center justify-between gap-2 border-t border-dove pt-2">
            <p className="text-[13px] tabular-nums text-muted">
              {draftH !== null || draftM !== null
                ? `${draftH !== null ? String(draftH).padStart(2, '0') : '--'}:${draftM !== null ? String(draftM).padStart(2, '0') : '--'}`
                : 'Pilih jam dan menit'}
            </p>
            {value && (
              <button type="button" onClick={() => { onChange(''); setOpen(false) }} className="text-[13px] text-ember hover:underline">
                Hapus jam
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export function PeriodPicker({ options, period, range, onPreset, onRange, label = 'Pilih periode', single = false }: {
  options: { id: string; label: string }[]
  period: string
  range: { from: string; to: string } | null
  onPreset: (id: string) => void
  onRange: (from: string, to: string) => void
  label?: string
  single?: boolean
}) {
  const [open, setOpen] = useState(false)
  const [draftFrom, setDraftFrom] = useState('')
  const [draftTo, setDraftTo] = useState('')
  const tp = todayParts()
  const [view, setView] = useState<[number, number]>([tp[0], tp[1]])
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

  const activeLabel = range
    ? range.from === range.to
      ? fmtDate(range.from)
      : `${fmtDate(range.from)} – ${fmtDate(range.to)}`
    : (options.find((o) => o.id === period)?.label ?? 'Pilih periode')
  const todayISO = isoDay(tp[0], tp[1], tp[2])

  function toggle() {
    if (!open) {
      setDraftFrom(range?.from ?? '')
      setDraftTo(range?.to ?? '')
      const base = range ? parseISODate(range.from) : null
      setView(base ? [base[0], base[1]] : [tp[0], tp[1]])
    }
    setOpen((o) => !o)
  }

  function pickDay(iso: string) {
    if (single) {
      setDraftFrom(iso)
      setDraftTo(iso)
      return
    }
    if (!draftFrom || (draftFrom && draftTo)) {
      setDraftFrom(iso)
      setDraftTo('')
    } else if (iso < draftFrom) {
      setDraftTo(draftFrom)
      setDraftFrom(iso)
    } else {
      setDraftTo(iso)
    }
  }

  function shiftMonth(delta: number) {
    setView(([y, m]) => {
      const nm = m + delta
      return [y + Math.floor(nm / 12), ((nm % 12) + 12) % 12]
    })
  }

  const months: [number, number][] = [view, view[1] === 11 ? [view[0] + 1, 0] : [view[0], view[1] + 1]]

  function MonthGrid({ y, m, first }: { y: number; m: number; first: boolean }) {
    return (
      <div className={sheet ? 'w-full max-w-80 shrink-0' : 'w-64 shrink-0'}>
        <div className="flex items-center justify-between px-1">
          {first ? (
            <button type="button" aria-label="Bulan sebelumnya" onClick={() => shiftMonth(-1)} className="grid size-8 place-items-center rounded-lg text-muted transition outline-none hover:bg-surface hover:text-fg focus-visible:ring-2 focus-visible:ring-ring">
              <ChevronLeft className="size-4" />
            </button>
          ) : <span className="size-8" />}
          <p className="text-sm font-medium">{MONTHS_ID[m]} {y}</p>
          {!first ? (
            <button type="button" aria-label="Bulan berikutnya" onClick={() => shiftMonth(1)} className="grid size-8 place-items-center rounded-lg text-muted transition outline-none hover:bg-surface hover:text-fg focus-visible:ring-2 focus-visible:ring-ring">
              <ChevronRight className="size-4" />
            </button>
          ) : <span className="size-8" />}
        </div>
        <div className="mt-2 grid grid-cols-7 text-center font-mono text-[11px] uppercase text-fog">
          {WEEKDAYS_ID.map((d) => <span key={d} className="py-1">{d}</span>)}
        </div>
        <div className="grid grid-cols-7 gap-0.5">
          {monthCells(y, m).map((c) => {
            const isStart = draftFrom !== '' && c.iso === draftFrom
            const isEnd = draftTo !== '' && c.iso === draftTo
            const inRange = draftFrom !== '' && draftTo !== '' && c.iso > draftFrom && c.iso < draftTo
            const isToday = c.iso === todayISO
            return (
              <button
                key={c.key} type="button" onClick={() => pickDay(c.iso)} aria-label={fmtDate(c.iso)}
                className={`grid ${sheet ? 'h-10' : 'h-8'} place-items-center rounded-md text-[13px] tabular-nums transition outline-none focus-visible:ring-2 focus-visible:ring-ring ${isStart || isEnd ? 'bg-jet font-medium text-paper' : inRange ? 'bg-surface text-fg' : c.outside ? 'text-fog/50 hover:bg-surface' : 'text-fg hover:bg-surface'} ${isToday && !isStart && !isEnd ? 'ring-1 ring-jet' : ''}`}
              >
                {c.d}
              </button>
            )
          })}
        </div>
      </div>
    )
  }

  return (
    <div ref={box} className="relative">
      <button
        type="button" aria-label={label} aria-expanded={open} onClick={toggle}
        className="flex h-[42px] w-full items-center gap-2.5 rounded-md border border-border bg-paper px-3.5 py-0 text-sm transition outline-none hover:border-jet focus-visible:ring-2 focus-visible:ring-ring"
      >
        <CalendarDays className="size-4 shrink-0 text-fog" />
        <span className="min-w-0 flex-1 truncate text-left tabular-nums text-fg">{activeLabel}</span>
        <ChevronDown className="size-4 shrink-0 text-fog" />
      </button>

      {open && sheet && (
        <div className={SHEET_BACKDROP} onClick={() => setOpen(false)} aria-hidden="true" />
      )}
      {open && (
        <div role="dialog" aria-label={label} className={sheet ? 'fixed inset-x-0 bottom-0 z-50 max-h-[88dvh] overflow-y-auto rounded-t-3xl bg-paper p-4 pb-6 shadow-xl' : 'absolute right-0 top-full z-50 mt-2 max-h-[calc(100dvh-7rem)] max-w-[calc(100vw-2rem)] overflow-y-auto rounded-xl border border-dove bg-paper p-3 shadow-lg'}>
          {sheet && (
            <div className="mb-2 flex items-center justify-between">
              <p className="text-[15px] font-bold text-fg">{label}</p>
              <button type="button" onClick={() => setOpen(false)} aria-label="Tutup" className="grid size-8 shrink-0 place-items-center rounded-md text-muted transition hover:bg-surface hover:text-fg">
                <X className="size-4" aria-hidden="true" />
              </button>
            </div>
          )}
          <div className="flex flex-col gap-3 sm:flex-row">
            <div className="flex shrink-0 flex-row flex-wrap gap-1 sm:w-32 sm:flex-col sm:flex-nowrap sm:border-r sm:border-dove sm:pr-3" role="listbox" aria-label="Periode cepat">
              {options.map((o) => (
                <button
                  key={o.id} type="button" role="option" aria-selected={!range && period === o.id}
                  onClick={() => { onPreset(o.id); setOpen(false) }}
                  className={`rounded-lg px-3 py-2 text-left text-sm whitespace-nowrap transition outline-none hover:bg-surface focus-visible:ring-2 focus-visible:ring-ring ${!range && period === o.id ? 'bg-surface font-medium text-fg' : 'text-fg/80'}`}
                >
                  {o.label}
                </button>
              ))}
            </div>
            <div className={sheet ? 'flex flex-col items-center gap-4' : 'flex flex-col gap-4 min-[920px]:flex-row'}>
              <MonthGrid y={months[0][0]} m={months[0][1]} first />
              {!sheet && <MonthGrid y={months[1][0]} m={months[1][1]} first={false} />}
            </div>
          </div>
          <div className="mt-3 flex flex-wrap items-center gap-2 border-t border-dove pt-3">
            <p className="min-w-0 flex-1 truncate text-sm tabular-nums text-muted">
              {draftFrom
                ? single
                  ? fmtDate(draftFrom)
                  : draftTo
                    ? `${fmtDate(draftFrom)} – ${fmtDate(draftTo)}`
                    : `${fmtDate(draftFrom)} – pilih tanggal akhir`
                : single
                  ? 'Klik tanggal yang ingin dilihat.'
                  : 'Klik tanggal mulai, lalu tanggal akhir.'}
            </p>
            <Button variant="ghost" onClick={() => setOpen(false)}>Batal</Button>
            <Button disabled={!draftFrom} onClick={() => { onRange(draftFrom, draftTo || draftFrom); setOpen(false) }}>
              Terapkan
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

// Dropdown pilihan tunggal bersama (pengganti <select> native untuk filter &
// pemilih di header/modal) — bahasa visual sama dengan PeriodPicker: tombol
// berbingkai + panel listbox dengan tanda centang pada pilihan aktif.
