import type { ReactNode } from 'react'
import { useEffect } from 'react'
import { Navigate, Route, Routes, useLocation } from 'react-router'
import { TooltipProvider } from '@/components/ui/tooltip'
import ErrorBoundary from './lib/ErrorBoundary'
import { useDB } from './lib/store'
import Landing from './pages/Landing'
import Masuk from './pages/Masuk'
import Daftar from './pages/Daftar'
import PilihAkun from './pages/PilihAkun'
import AppShell from './pages/AppShell'
import Dashboard from './pages/Dashboard'
import Pelanggan from './pages/Pelanggan'
import Supplier from './pages/Supplier'
import Pembelian from './pages/Pembelian'
import HutangPiutang from './pages/HutangPiutang'
import Pos from './pages/Pos'
import Produk from './pages/Produk'
import Promo from './pages/Promo'
import Stok from './pages/Stok'
import Transaksi from './pages/Transaksi'
import Laporan from './pages/Laporan'
import Users from './pages/Users'
import Pengaturan from './pages/Pengaturan'
import Unduh from './pages/Unduh'

// Halaman khusus admin: kasir yang mengetik URL langsung diarahkan balik.
// API tetap penjaga utama (403); ini pertahanan berlapis + UX.
function AdminOnly({ children }: { children: ReactNode }) {
  const { session } = useDB()
  if (session && session.role !== 'admin') return <Navigate to="/app" replace />
  return <>{children}</>
}

// Setiap pindah route mulai dari paling atas halaman,
// bukan posisi scroll terakhir (berlaku /masuk, /daftar,
// /unduh, dan semua route lain; scroll section landing yang
// dipicu setelah navigasi tetap jalan karena dieksekusi belakangan).
function ScrollToTop() {
  const { pathname } = useLocation()
  useEffect(() => {
    try {
      window.scrollTo({ top: 0, left: 0, behavior: 'instant' as ScrollBehavior })
    } catch {
      window.scrollTo(0, 0)
    }
  }, [pathname])
  return null
}

export default function App() {
  return (
    <TooltipProvider>
      <ErrorBoundary>
        <ScrollToTop />
        <Routes>
        <Route path="/" element={<Landing />} />
        <Route path="/masuk" element={<Masuk />} />
        <Route path="/daftar" element={<Daftar />} />
        <Route path="/pilih-akun" element={<PilihAkun />} />
        <Route path="/unduh" element={<Unduh />} />
        <Route path="/app" element={<AppShell />}>
          <Route index element={<Dashboard />} />
          <Route path="pos" element={<Pos />} />
          <Route path="produk" element={<AdminOnly><Produk /></AdminOnly>} />
          <Route path="promo" element={<AdminOnly><Promo /></AdminOnly>} />
          <Route path="stok" element={<AdminOnly><Stok /></AdminOnly>} />
          <Route path="opname" element={<Navigate to="/app/stok?tab=opname" replace />} />
          <Route path="transaksi" element={<Transaksi />} />
          <Route path="pelanggan" element={<Pelanggan />} />
          <Route path="supplier" element={<AdminOnly><Supplier /></AdminOnly>} />
          <Route path="pembelian" element={<AdminOnly><Pembelian /></AdminOnly>} />
          <Route path="hutang-piutang" element={<AdminOnly><HutangPiutang /></AdminOnly>} />
          <Route path="shift" element={<Navigate to="/app/laporan?tab=shift" replace />} />
          <Route path="laporan" element={<AdminOnly><Laporan /></AdminOnly>} />
          <Route path="karyawan" element={<Navigate to="/app/laporan?tab=kasir" replace />} />
          <Route path="users" element={<AdminOnly><Users /></AdminOnly>} />
          <Route path="pengaturan" element={<AdminOnly><Pengaturan /></AdminOnly>} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
      </ErrorBoundary>
    </TooltipProvider>
  )
}
