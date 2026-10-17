import type { ReactNode } from 'react'
import { Skeleton } from '@/shared/ui/primitives/skeleton'

export function Logo({ className = 'h-8 w-auto', tone = 'auto' }: { className?: string; tone?: 'auto' | 'light' | 'dark' }) {
  // BASE_URL = '/' di cloud, './' di bundle offline (file://) — path absolut
  // '/Logo-*.png' tak resolve di Electron, jadi prefix ikut base.
  const base = import.meta.env.BASE_URL || '/'
  if (tone === 'dark') return <img src={`${base}Logo-Putih.png`} alt="NOJPOS" draggable={false} className={className} />
  if (tone === 'light') return <img src={`${base}Logo-Hitam.png`} alt="NOJPOS" draggable={false} className={className} />
  return (
    <span className={`relative inline-flex shrink-0 ${className}`} role="img" aria-label="NOJPOS">
      <img src={`${base}Logo-Hitam.png`} alt="" aria-hidden="true" draggable={false} className="h-full w-auto transition-opacity duration-150 dark:opacity-0" />
      <img src={`${base}Logo-Putih.png`} alt="" aria-hidden="true" draggable={false} className="absolute inset-0 h-full w-auto opacity-0 transition-opacity duration-150 dark:opacity-100" />
    </span>
  )
}

export function Button({
  children, onClick, type = 'button', variant = 'primary', disabled, className = '',
}: {
  children: ReactNode
  onClick?: () => void
  type?: 'button' | 'submit'
  variant?: 'primary' | 'ghost' | 'danger'
  disabled?: boolean
  className?: string
}) {
  const styles = {
    primary: 'bg-jet text-paper border-jet hover:opacity-85',
    ghost: 'bg-transparent text-fg border-dove hover:border-jet',
    danger: 'bg-transparent text-ember border-dove hover:border-ember',
  }[variant]
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className={`inline-flex h-[42px] items-center justify-center gap-2 rounded-full border px-6 py-0 text-sm font-medium transition active:translate-y-px disabled:cursor-not-allowed disabled:opacity-40 ${styles} ${className}`}
    >
      {children}
    </button>
  )
}

// Satu tombol dropdown untuk export dual-format (Excel + CSV).
// Tampil persis seperti Button ghost agar konsisten di semua halaman.

export function SettingsCard({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <section className={`min-w-0 rounded-2xl bg-cream p-4 sm:p-6 ${className}`}>{children}</section>
}

export function SettingsHead({ kicker, title, sub }: { kicker?: string; title: string; sub?: string }) {
  return (
    <div className="mb-4">
      {kicker && <p className="font-mono text-[11px] uppercase tracking-wider text-fog">{kicker}</p>}
      <h2 className="mt-0.5 text-[15px] font-medium text-fg">{title}</h2>
      {sub && <p className="mt-0.5 text-[13px] text-muted">{sub}</p>}
    </div>
  )
}

export function Pill({ children, tone = 'ok' }: { children: ReactNode; tone?: 'ok' | 'warn' | 'muted' }) {
  const styles = {
    ok: 'bg-success-bg text-sprout',
    warn: 'bg-sand text-steel',
    muted: 'bg-surface text-muted',
  }[tone]
  return <span className={`inline-block rounded-full px-2.5 py-0.5 text-[11px] font-medium ${styles}`}>{children}</span>
}

export function StatusPill({ status }: { status: string }) {
  const map: Record<string, { label: string; cls: string }> = {
    completed: { label: 'Selesai', cls: 'bg-success-bg text-sprout' },
    pending: { label: 'Proses', cls: 'bg-sand text-steel' },
    cancelled: { label: 'Dibatalkan', cls: 'bg-surface text-muted' },
    refunded: { label: 'Refund', cls: 'bg-sand text-sunbeam' },
    completed_offline_shortage: { label: 'Selesai · Stok Kurang', cls: 'bg-sand text-sunbeam' },
  }
  const s = map[status] ?? { label: status, cls: 'bg-surface text-muted' }
  return <span className={`inline-block rounded-full px-2.5 py-0.5 text-[11px] font-medium ${s.cls}`}>{s.label}</span>
}

export function Empty({ title, sub, action }: { title: string; sub: string; action?: ReactNode }) {
  return (
    <div className="rounded-2xl border border-dashed border-dove px-6 py-14 text-center">
      <p className="font-medium text-fg">{title}</p>
      <p className="mx-auto mt-1 max-w-sm text-sm text-muted">{sub}</p>
      {action && <div className="mt-5">{action}</div>}
    </div>
  )
}

export function PageHead({ title, right }: { title: string; sub?: string; right?: ReactNode }) {
  return (
    <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 className="text-[clamp(26px,3vw,34px)] font-normal leading-[1.1] tracking-[-0.025em]">{title}</h1>
      </div>
      {right}
    </div>
  )
}

// Standar tabel (keputusan Rayhan 2026-10-15): SEMUA header & sel rata kiri,
// termasuk angka — prop `right` lama dipertahankan di API hanya agar
// pemanggil lama tidak berubah; efeknya kini cuma tabular-nums di Td.

export function SkeletonRows({ cols, rows = 8 }: { cols: number; rows?: number }) {
  return (
    <tbody aria-hidden="true">
      {Array.from({ length: rows }).map((_, r) => (
        <tr key={r}>
          {Array.from({ length: cols }).map((_, c) => (
            <td key={c} className="border-b border-dove px-2.5 py-2.5">
              <Skeleton className="h-4 w-full max-w-36" />
            </td>
          ))}
        </tr>
      ))}
    </tbody>
  )
}

// Kartu statistik bersama untuk semua strip ringkasan (Beranda, Laporan,
// Stok, Transaksi, Pengaturan). Visual = keluarga opc-kpi/opc-stat dari
// classic.css: ikon di chip kiri, nilai besar, label, lalu baris sub ATAU
// baris perbandingan (compare menggantikan sub, dengan arah ↑↓ berwarna).
// density 'stat' = varian ringkas (Stok/Transaksi); 'kpi' = varian besar.
