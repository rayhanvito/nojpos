// API fitur settings: permukaan endpoint yang dipakai halaman
// Pengaturan (tab Akun/Toko/Struk/Pajak/Outlet). Endpoint-nya sendiri
// tetap implementasi bersama di shared/api per domain (insights untuk
// /settings dan transaksi, auth untuk pengguna/passcode/outlet) — file
// ini hanya mengekspor ulang yang dipakai fitur settings, mengikuti
// pola fitur pos (satu pintu masuk API per fitur).
export {
  ApiError,
  apiCreateOutlet, apiGetSettings, apiListTransactions, apiListUsers, apiLogout,
  apiResetPassword, apiSendPasswordResetOtp, apiSetPasscode, apiUpdateOutlet, apiUpdateSettings,
} from '../../shared/api'
export type { Outlet, Page, StoreSettings, Trx, User } from '../../shared/api'
