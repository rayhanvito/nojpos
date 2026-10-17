// API fitur users: permukaan endpoint yang dipakai halaman Pengguna.
// Endpoint-nya sendiri tetap implementasi bersama di shared/api per
// domain (auth untuk pengguna, insights untuk laporan/aktivitas, sales
// untuk transaksi) — file ini hanya mengekspor ulang yang dipakai fitur
// users, mengikuti pola fitur pos (satu pintu masuk API per fitur).
export {
  apiCreateUser, apiDeleteUser, apiGetReport, apiListActivity, apiListTransactions, apiListUsers, apiRenameUser, apiSetUserActive,
  setCachedAccounts,
} from '../../shared/api'
export type { ActivityItem, Page, ReportBundle, Trx, User } from '../../shared/api'
