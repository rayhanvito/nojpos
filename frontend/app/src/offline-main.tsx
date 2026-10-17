import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { HashRouter, Navigate, Route, Routes } from 'react-router'
import { TooltipProvider } from '@/shared/ui/primitives/tooltip'
import './styles/index.css'
import ErrorBoundary from './shared/ErrorBoundary'
import { getLocalDB, hasAccount } from './shared/localdb'
import { setSession } from './shared/session/store'
import OfflineShell from './offline/OfflineShell'
import Dashboard from './offline/Dashboard'
import OfflinePos from './offline/OfflinePos'
import OfflineProduk from './offline/OfflineProduk'
import Stok from './offline/Stok'
import OfflineTransaksi from './offline/OfflineTransaksi'
import Laporan from './offline/Laporan'
import Pengaturan from './offline/Pengaturan'
import OfflineBackup from './offline/OfflineBackup'

// Sesi offline disiapkan SEBELUM render: halaman (mis. Dashboard) membaca
// session.id di render pertama — bila telat di-set lewat effect, crash
// "Cannot read properties of null (reading 'id')" tiap boot dengan akun.
if (hasAccount()) {
  const s = getLocalDB().settings
  setSession({ id: 'local', email: '', name: s.ownerName, role: 'admin', store: s.storeName })
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <HashRouter>
      <TooltipProvider>
        <ErrorBoundary>
          <Routes>
            <Route path="/" element={<OfflineShell />}>
              <Route index element={<Dashboard />} />
              <Route path="pos" element={<OfflinePos />} />
              <Route path="products" element={<OfflineProduk />} />
              <Route path="produk" element={<Navigate to="/products" replace />} />
              <Route path="stock" element={<Stok />} />
              <Route path="stok" element={<Navigate to="/stock" replace />} />
              <Route path="transactions" element={<OfflineTransaksi />} />
              <Route path="transaksi" element={<Navigate to="/transactions" replace />} />
              <Route path="reports" element={<Laporan />} />
              <Route path="laporan" element={<Navigate to="/reports" replace />} />
              <Route path="settings" element={<Pengaturan />} />
              <Route path="pengaturan" element={<Navigate to="/settings" replace />} />
              <Route path="backup" element={<OfflineBackup />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </ErrorBoundary>
      </TooltipProvider>
    </HashRouter>
  </StrictMode>,
)
