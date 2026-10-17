import type { ReactNode } from 'react'
import { useEffect, useRef, useState } from 'react'
import { ChevronDown, Download, FileSpreadsheet, FileText, Upload, X } from 'lucide-react'
import { useSheetMode, SHEET_BACKDROP } from './sheet.ts'

// Menu format berkas bersama (Export/Import): di desktop popover melayang,
// di mobile (<640px) bottom sheet berjudul — pola yang sama dengan Select
// dan FilterPanel agar semua popup konsisten (keputusan Rayhan 16 Okt 2026).
function FileMenu({ label, icon, disabled, items }: {
  label: string
  icon: ReactNode
  disabled?: boolean
  items: { label: string; icon: ReactNode; onClick: () => void }[]
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

  const triggerCls = 'inline-flex h-[42px] items-center justify-center gap-2 rounded-full border border-dove bg-transparent px-6 py-0 text-sm font-medium text-fg transition outline-none hover:border-jet active:translate-y-px disabled:cursor-not-allowed disabled:opacity-40 focus-visible:ring-2 focus-visible:ring-ring'

  return (
    <div ref={box} className="opc-filemenu relative inline-block">
      <button type="button" disabled={disabled} aria-label={`${label} data`} aria-expanded={open} aria-haspopup="menu" onClick={() => setOpen((o) => !o)} className={triggerCls}>
        {icon}{label}<ChevronDown className="size-4 text-fog" />
      </button>
      {open && sheet && (
        <div className={SHEET_BACKDROP} onClick={() => setOpen(false)} aria-hidden="true" />
      )}
      {open && (
        <div role="menu" aria-label={label} className={sheet ? 'fixed inset-x-0 bottom-0 z-50 flex max-h-[85dvh] flex-col rounded-t-3xl bg-paper pb-5 shadow-xl' : 'absolute right-0 top-full z-50 mt-2 w-56 rounded-xl border border-dove bg-paper p-1.5 shadow-lg'}>
          {sheet && (
            <div className="flex items-center justify-between px-5 pb-1 pt-4">
              <p className="text-[15px] font-bold text-fg">{label}</p>
              <button type="button" onClick={() => setOpen(false)} aria-label="Tutup" className="grid size-8 shrink-0 place-items-center rounded-md text-muted transition outline-none hover:bg-surface hover:text-fg">
                <X className="size-4" aria-hidden="true" />
              </button>
            </div>
          )}
          <p className={`px-5 pb-1 pt-2 text-xs font-medium text-fog ${sheet ? '' : 'px-3'}`}>Format file</p>
          <div className={sheet ? 'min-h-0 overflow-y-auto px-2' : ''}>
            {items.map((it) => (
              <button
                key={it.label} type="button" role="menuitem"
                onClick={() => { setOpen(false); it.onClick() }}
                className={`flex w-full items-center gap-2.5 rounded-lg px-3 py-2.5 text-left text-sm transition outline-none hover:bg-surface focus-visible:ring-2 focus-visible:ring-ring ${sheet ? 'min-h-[44px] text-[15px]' : ''}`}
              >
                {it.icon}
                <span className="min-w-0 flex-1 truncate">{it.label}</span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

export function ExportMenu({ onCSV, onExcel, onTaxCSV, label = 'Export', disabled }: {
  onCSV: () => void
  onExcel?: () => void
  onTaxCSV?: () => void
  label?: string
  disabled?: boolean
}) {
  return (
    <FileMenu
      label={label}
      disabled={disabled}
      icon={<Download className="size-4" />}
      items={[
        ...(onExcel ? [{ label: 'Excel (.xlsx)', icon: <FileSpreadsheet className="size-4" />, onClick: onExcel }] : []),
        { label: 'CSV (.csv)', icon: <FileText className="size-4" />, onClick: onCSV },
        ...(onTaxCSV ? [{ label: 'CSV Pajak bulan terpilih', icon: <FileText className="size-4" />, onClick: onTaxCSV }] : []),
      ]}
    />
  )
}

// Satu tombol dropdown untuk import dual-format. Pilihan menentukan filter
// file picker; parseImportFile tetap auto-deteksi isi apa pun yang dipilih.
export function ImportMenu({ onPick, label = 'Import', disabled }: {
  onPick: (kind: 'xlsx' | 'csv') => void
  label?: string
  disabled?: boolean
}) {
  return (
    <FileMenu
      label={label}
      disabled={disabled}
      icon={<Upload className="size-4" />}
      items={[
        { label: 'Excel (.xlsx)', icon: <FileSpreadsheet className="size-4" />, onClick: () => onPick('xlsx') },
        { label: 'CSV (.csv)', icon: <FileText className="size-4" />, onClick: () => onPick('csv') },
      ]}
    />
  )
}

export function Modal({
  open, title, onClose, children, wide, variant: _variant = 'center',
}: {
  open: boolean
  title: string
  onClose: () => void
  children: ReactNode
  wide?: boolean
  // Pola modal mobile (keputusan 15 Okt 2026): 'sheet' = bottom sheet untuk aksi pendek;
  // 'full' = layar penuh untuk form padat; desktop selalu di tengah untuk semua variant.
  variant?: 'center' | 'sheet' | 'full'
}) {
  if (!open) return null
  // Keputusan Rayhan 15 Okt 17:18 (mencabut taksonomi 16:28): SEMUA modal di mobile
  // adalah bottom sheet; desktop tetap kartu di tengah. variant dipertahankan untuk
  // kompatibilitas pemanggil lama dan tidak lagi mengubah bentuk di mobile.
  const wrapCls = 'fixed inset-0 z-50 grid place-items-end overflow-y-auto bg-black/50 sm:place-items-center sm:p-4'
  const cardCls = `max-h-[calc(100dvh-1.5rem)] w-full ${wide ? 'max-w-2xl' : 'max-w-md'} overflow-y-auto rounded-t-3xl bg-cream p-4 text-fg shadow-xl sm:rounded-2xl sm:p-6`
  return (
    <div className={wrapCls} onClick={onClose} role="dialog" aria-modal="true" aria-label={title}>
      <div
        className={cardCls}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="sticky -top-4 z-10 -mx-4 -mt-4 flex items-center justify-between bg-cream px-4 pt-4 pb-4 sm:-top-6 sm:-mx-6 sm:-mt-6 sm:px-6 sm:pt-6">
          <h3 className="text-lg font-bold tracking-tight">{title}</h3>
          <button className="rounded-full p-1 text-fog hover:text-fg" onClick={onClose} aria-label="Tutup">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M18 6 6 18M6 6l12 12" /></svg>
          </button>
        </div>
        {children}
      </div>
    </div>
  )
}

// Kartu seksi + kepala seksi bersama (semula lokal di Pengaturan; dipindah ke
// sini supaya definisinya satu untuk semua halaman — fase penyatuan komponen).

export function ModalActions({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <div className={`sticky -bottom-4 z-10 -mx-4 -mb-4 mt-4 flex justify-end gap-2 bg-cream px-4 pt-3 pb-4 sm:-bottom-6 sm:-mx-6 sm:-mb-6 sm:px-6 sm:pb-6 ${className}`.trim()}>{children}</div>
}

// Nomor halaman ringkas: semua bila <=7, else 1 … jendela … akhir.
