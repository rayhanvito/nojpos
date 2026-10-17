import type { ReactNode } from 'react'
import { Crumb } from './Crumb'

// Pola header persis Dashboard: judul + sub kiri, breadcrumb + aksi kanan.
export function PageHeader({ title, crumb, actions }: {
  title: string
  // sub TIDAK lagi dirender (keputusan Rayhan 16 Okt 2026: header cukup judul).
  // Masih diterima di tipe hanya demi kompilasi halaman offline lama.
  sub?: ReactNode
  crumb: string
  actions?: ReactNode
}) {
  return (
    <div className="opc-pagehead">
      <Crumb page={crumb} />
      <div className="opc-pagehead-main">
        <div className="min-w-0">
          <h1>{title}</h1>
        </div>
        {actions && <div className="opc-pagehead-right">{actions}</div>}
      </div>
    </div>
  )
}
