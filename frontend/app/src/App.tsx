import type { ReactNode } from 'react'
import { useEffect } from 'react'
import { Navigate, Route, Routes, useLocation } from 'react-router'
import { TooltipProvider } from '@/shared/ui/primitives/tooltip'
import ErrorBoundary from './shared/ErrorBoundary'
import { useDB } from './shared/session/store'
import Landing from './pages/Landing'
import Login from './features/auth/LoginPage'
import Register from './features/auth/RegisterPage'
import AccountPicker from './features/auth/AccountPickerPage'
import AppShell from './pages/AppShell'
import Dashboard from './features/dashboard/DashboardPage'
import Customers from './features/customer/CustomersPage'
import Supplier from './features/purchasing/SupplierPage'
import Purchases from './features/purchasing/PurchasesPage'
import Debts from './features/finance/DebtsPage'
import PosPage from './features/pos/PosPage'
import Products from './features/catalog/ProductsPage'
import Promo from './features/promo/PromoPage'
import Stock from './features/inventory/StockPage'
import Transactions from './features/transactions/TransactionsPage'
import Reports from './features/report/ReportsPage'
import Users from './features/users/UsersPage'
import Settings from './features/settings/SettingsPage'

// Halaman khusus admin: kasir yang mengetik URL langsung diarahkan balik.
// API tetap penjaga utama (403); ini pertahanan berlapis + UX.
function AdminOnly({ children }: { children: ReactNode }) {
  const { session } = useDB()
  if (session && session.role !== 'admin') return <Navigate to="/app" replace />
  return <>{children}</>
}

// Redirect rute lama (Bahasa Indonesia) ke rute kanonis Inggris.
// Query string diteruskan persis (?edit=, ?tab=, ?google= dst tidak hilang).
function LegacyRedirect({ to }: { to: string }) {
  const { search } = useLocation()
  return <Navigate to={to + search} replace />
}

// Setiap pindah route mulai dari paling atas halaman,
// bukan posisi scroll terakhir (berlaku /login, /register,
// /download, dan semua route lain; scroll section landing yang
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
        <Route path="/login" element={<Login />} />
        <Route path="/masuk" element={<LegacyRedirect to="/login" />} />
        <Route path="/register" element={<Register />} />
        <Route path="/daftar" element={<LegacyRedirect to="/register" />} />
        <Route path="/select-account" element={<AccountPicker />} />
        <Route path="/pilih-akun" element={<LegacyRedirect to="/select-account" />} />
        <Route path="/download" element={<LegacyRedirect to="/" />} />
        <Route path="/unduh" element={<LegacyRedirect to="/" />} />
        <Route path="/app" element={<AppShell />}>
          <Route index element={<Dashboard />} />
          <Route path="pos" element={<PosPage />} />
          <Route path="products" element={<AdminOnly><Products /></AdminOnly>} />
          <Route path="produk" element={<LegacyRedirect to="/app/products" />} />
          <Route path="promo" element={<AdminOnly><Promo /></AdminOnly>} />
          <Route path="stock" element={<AdminOnly><Stock /></AdminOnly>} />
          <Route path="stok" element={<LegacyRedirect to="/app/stock" />} />
          <Route path="opname" element={<Navigate to="/app/stock?tab=stocktake" replace />} />
          <Route path="transactions" element={<Transactions />} />
          <Route path="transaksi" element={<LegacyRedirect to="/app/transactions" />} />
          <Route path="customers" element={<Customers />} />
          <Route path="pelanggan" element={<LegacyRedirect to="/app/customers" />} />
          <Route path="supplier" element={<AdminOnly><Supplier /></AdminOnly>} />
          <Route path="purchases" element={<AdminOnly><Purchases /></AdminOnly>} />
          <Route path="pembelian" element={<LegacyRedirect to="/app/purchases" />} />
          <Route path="debts" element={<AdminOnly><Debts /></AdminOnly>} />
          <Route path="hutang-piutang" element={<LegacyRedirect to="/app/debts" />} />
          <Route path="shift" element={<Navigate to="/app/reports?tab=shift" replace />} />
          <Route path="reports" element={<AdminOnly><Reports /></AdminOnly>} />
          <Route path="laporan" element={<LegacyRedirect to="/app/reports" />} />
          <Route path="karyawan" element={<Navigate to="/app/reports?tab=cashier" replace />} />
          <Route path="users" element={<AdminOnly><Users /></AdminOnly>} />
          <Route path="settings" element={<AdminOnly><Settings /></AdminOnly>} />
          <Route path="pengaturan" element={<LegacyRedirect to="/app/settings" />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
      </ErrorBoundary>
    </TooltipProvider>
  )
}
