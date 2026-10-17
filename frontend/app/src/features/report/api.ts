// API fitur report: permukaan endpoint yang dipakai halaman Laporan
// dan tab Kasir. Endpoint-nya sendiri tetap implementasi bersama di
// shared/api per domain (insights untuk laporan/shift/pajak, catalog
// untuk produk, stock untuk pergerakan, sales untuk transaksi, auth
// untuk daftar pengguna) — file ini hanya mengekspor ulang yang dipakai
// fitur report, mengikuti pola fitur pos (satu pintu masuk API per
// fitur).
export {
  apiExportShifts, apiExportTaxReport, apiGetReport, apiGetTaxReport, apiListMovements, apiListProducts,
  apiListShifts, apiListTransactions, apiListUsers, fetchAll,
} from '../../shared/api'
export type { Movement, Product, ReportBundle, Shift, TaxReport, Trx, User } from '../../shared/api'
