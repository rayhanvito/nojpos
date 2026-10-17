import type { ComponentType, ReactNode } from 'react'
import { Link } from 'react-router'
import { Card } from '@/shared/ui/primitives/card'
import { Select } from './forms.tsx'
import type { TrxItem } from '../api.ts'

export function TrxItems({ items, className }: { items: TrxItem[]; className?: string }) {
  const shown = items.slice(0, 2)
  const rest = items.length - shown.length
  return (
    <div className={`flex flex-wrap items-center gap-1.5 ${className ?? ''}`}>
      {shown.map((i) => (
        <span
          key={i.product_id}
          className="inline-flex max-w-full items-center gap-1.5 rounded-md border border-dove bg-surface px-2 py-1"
        >
          <span className="truncate text-[13px] leading-none text-fg">{i.name}</span>
          <span className="rounded-sm bg-paper px-1.5 py-0.5 font-mono text-[11px] font-medium leading-none tabular-nums text-steel ring-1 ring-dove">
            ×{i.qty}
          </span>
        </span>
      ))}
      {rest > 0 && <span className="text-[13px] text-fog">+{rest} lainnya</span>}
    </div>
  )
}

// Logo dua varian: Hitam di mode terang, Putih di mode gelap.
// Varian auto menumpuk kedua gambar + preload (lihat index.html)
// sehingga ganti tema hanya crossfade opacity — tanpa flicker jaringan.

export function Th({ children }: { children?: ReactNode; right?: boolean }) {
  return (
    <th className="whitespace-nowrap border-b border-dove px-2.5 py-2 text-left font-mono text-[11px] font-medium uppercase tracking-wider text-fog">
      {children}
    </th>
  )
}

export function Td({ children, mono, right }: { children: ReactNode; mono?: boolean; right?: boolean }) {
  return (
    <td className={`border-b border-dove px-2.5 py-2.5 text-[13px] text-muted ${mono ? 'font-mono text-xs text-fg' : ''} ${right ? 'tabular-nums' : ''}`}>
      {children}
    </td>
  )
}

// Pembungkus tabel standar tingkat halaman: SATU-satunya wadah tabel daftar/
// laporan (varian sah lain: tabel di dalam Card, tabel mini di modal, dan
// panel drilldown — ketiganya memang konteks berbeda). className tambahan
// untuk varian responsif seperti 'hidden min-[1280px]:block'.
// Tombol aksi baris tabel — SATU-satunya pola aksi di semua tabel (pola
// referensi halaman Produk): ikon saja 32px, tooltip dari title, varian
// danger untuk hapus/tutup paksa (ikon memerah saat hover).
export function RowAction({ title, label, onClick, danger, disabled, children }: {
  title: string
  label?: string
  onClick?: () => void
  danger?: boolean
  disabled?: boolean
  children: ReactNode
}) {
  return (
    <button
      type="button"
      title={title}
      aria-label={label ?? title}
      onClick={onClick}
      disabled={disabled}
      className={`grid size-8 place-items-center rounded-md text-steel transition hover:bg-surface ${danger ? 'hover:text-ember' : 'hover:text-fg'} disabled:pointer-events-none disabled:opacity-40`}
    >
      {children}
    </button>
  )
}

export function TableWrap({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <div className={`overflow-x-auto rounded-2xl bg-cream p-2 ${className}`.trim()}>{children}</div>
}

// Standar kartu data mobile (fase mobile Batch 0, acuan Transaksi/Produk):
// tabel (TableWrap) diberi className "hidden min-[1280px]:block", dan di
// bawah 1280px daftar yang sama tampil sebagai kartu lewat CardGrid +
// DataCard. Satu breakpoint (1280px) untuk semua halaman — Batch 1-3
// mengonversi tabel lain ke pola ini.
export function CardGrid({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <div className={`grid grid-cols-1 gap-2.5 min-[768px]:grid-cols-2 min-[1120px]:grid-cols-3 min-[1280px]:hidden ${className}`.trim()}>{children}</div>
}

export function DataCard({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <div className={`flex flex-col rounded-xl border border-dove bg-paper p-2.5 ${className}`.trim()}>{children}</div>
}

// Bar aksi lengket di dasar modal: tombol tetap terlihat walau isi modal
// panjang dan digulir (fase mobile Batch 0). Margin negatif menembus padding
// kartu modal supaya pita lengket menutup penuh selebar kartu.
// Strip info/peringatan berkotak kertas: ringkasan sesi, galat dengan tombol
// coba lagi, kotak peringatan di modal. Cangkang satu-satunya; tata letak dan
// jarak diatur pemakai lewat className.
export function InfoStrip({ className = '', children }: { className?: string; children: ReactNode }) {
  return <div className={`rounded-xl border border-dove bg-paper ${className}`}>{children}</div>
}

// Strip pesan satu-satunya (galat/sukses) di seluruh app — menggantikan 65+
// salinan tangan `bg-sand … text-ember` yang dulu ditempel per halaman.
export function AlertStrip({ tone = 'error', className = '', children }: {
  tone?: 'error' | 'success'
  className?: string
  children: ReactNode
}) {
  return (
    <p role={tone === 'error' ? 'alert' : 'status'} className={`rounded-lg px-3.5 py-2.5 text-[13px] ${tone === 'error' ? 'bg-sand text-ember' : 'bg-success-bg text-sprout'} ${className}`.trim()}>
      {children}
    </p>
  )
}

// Daftar label–nilai baku untuk modal detail (piutang, transaksi, PO, shift):
// kolom label tetap, nilai sejajar — menggantikan baris inline nempel yang
// membuat nilai mulai di posisi x berbeda-beda per baris.
export function MetaList({ items, className = '' }: {
  items: { label: ReactNode; value: ReactNode; mono?: boolean }[]
  className?: string
}) {
  return (
    <dl className={`grid grid-cols-[7.5rem_1fr] items-baseline gap-x-3 gap-y-2 text-sm ${className}`.trim()}>
      {items.map((it, i) => (
        <div key={i} className="contents">
          <dt className="text-muted">{it.label}</dt>
          <dd className={`min-w-0 text-fg ${it.mono ? 'font-mono' : 'tabular-nums'}`}>{it.value}</dd>
        </div>
      ))}
    </dl>
  )
}

// Pager angka rata kanan: ‹ › ringkas + lompat langsung. Null bila 1 halaman.
const PAGER_SIZE_OPTIONS = [10, 20, 50, 100].map((n) => ({ value: String(n), label: `${n}/halaman` }))

// Bar paginasi baku. Tanpa totalItems = mode lama (tombol halaman saja,
// hilang bila <= 1 halaman). Dengan totalItems: bar info selalu tampil —
// teks "Menampilkan X–Y dari Z data" + pemilih ukuran halaman (10/20/50/100)
// + tombol halaman. Ukuran halaman adalah state lokal halaman pemakai
// (keputusan Rayhan 16 Okt 2026: tidak global, reset ke 10 saat pindah halaman).
export function Pager({ page, total, onChange, className = 'mt-3', totalItems, pageSize, onPageSizeChange }: {
  page: number
  total: number
  onChange: (p: number) => void
  className?: string
  totalItems?: number
  pageSize?: number
  onPageSizeChange?: (n: number) => void
}) {
  const btn = 'h-[38px] rounded-lg border border-dove bg-paper px-3 py-2 text-sm transition hover:border-jet disabled:cursor-not-allowed disabled:opacity-40'
  // Model B (keputusan Rayhan 16 Okt 2026): tanpa deretan angka — cukup
  // panah geser + indikator "halaman/total". Kontrol SELALU tampil agar bar
  // konsisten di semua daftar; yang berubah hanya status nonaktif panah.
  const buttons = (
    <>
      <button aria-label="Halaman sebelumnya" disabled={page === 0} onClick={() => onChange(page - 1)} className={btn}>‹</button>
      <span aria-label={`Halaman ${page + 1} dari ${Math.max(1, total)}`} className="inline-flex h-[38px] min-w-14 items-center justify-center rounded-lg border border-jet bg-jet px-2.5 py-2 text-center font-mono text-[13px] font-medium tabular-nums text-paper">{page + 1}/{Math.max(1, total)}</span>
      <button aria-label="Halaman berikutnya" disabled={page >= total - 1} onClick={() => onChange(page + 1)} className={btn}>›</button>
    </>
  )

  if (totalItems === undefined) {
    if (total <= 1) return null
    return <div className={`flex flex-wrap items-center justify-end gap-1.5 ${className}`}>{buttons}</div>
  }

  const size = pageSize ?? 10
  const from = totalItems === 0 ? 0 : page * size + 1
  const to = Math.min(totalItems, (page + 1) * size)
  return (
    <div className={`flex flex-col gap-2.5 sm:flex-row sm:items-center sm:justify-between ${className}`}>
      <p className="text-[13px] tabular-nums text-muted">Menampilkan {from}–{to} dari {totalItems} data</p>
      <div className="flex flex-wrap items-center gap-2 sm:justify-end">
        {onPageSizeChange && (
          <Select
            label="Data per halaman"
            className="w-[8.75rem] shrink-0" dense
            value={String(size)}
            onChange={(v) => onPageSizeChange(Number(v))}
            options={PAGER_SIZE_OPTIONS}
          />
        )}
        {buttons}
      </div>
    </div>
  )
}

// DatePicker — pemilih tanggal kalender popover (Senin dulu, label Indonesia).
// Nilai string 'YYYY-MM-DD' ('' = kosong), tanpa dependensi tanggal baru.

export function StatCard({ label, value, sub, compare, invert, icon: Icon, href, density = 'kpi' }: {
  label: string
  value: ReactNode
  sub?: string
  compare?: string | null
  invert?: boolean
  icon?: ComponentType<{ className?: string }>
  href?: string
  density?: 'kpi' | 'stat'
}) {
  const dir = invert && compare ? compare.replace(/↑/g, '⇅').replace(/↓/g, '↑').replace(/⇅/g, '↓') : compare
  const up = dir?.startsWith('↑') ?? false
  const down = dir?.startsWith('↓') ?? false
  const tone = !dir ? 'text-muted-foreground' : up ? 'text-sprout' : down ? 'text-ember' : 'text-muted-foreground'
  const stat = density === 'stat'
  return (
    <Card className={stat ? 'opc-stat' : 'opc-kpi'}>
      <div className={stat ? 'opc-stat-body' : 'opc-kpi-body'}>
        {Icon && (
          <span className="opc-kpi-ico" aria-hidden="true">
            <Icon />
          </span>
        )}
        <div className="min-w-0">
          <p className={`${stat ? 'opc-stat-val' : 'opc-kpi-val'} tabular-nums`} title={typeof value === 'string' ? value : undefined}>{value}</p>
          <p className={stat ? 'opc-stat-label' : 'opc-kpi-label'}>{label}</p>
          {(dir ?? sub) && <p className={`mt-0.5 text-xs tabular-nums ${tone}`}>{dir ?? sub}</p>}
        </div>
      </div>
      {href && (
        <Link to={href} className="opc-kpi-foot">
          Lihat detail <span aria-hidden="true">&nbsp;→</span>
        </Link>
      )}
    </Card>
  )
}
