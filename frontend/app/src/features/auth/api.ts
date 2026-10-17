// API fitur auth: permukaan endpoint yang dipakai halaman Masuk,
// Daftar, dan Pilih Akun. Endpoint-nya sendiri tetap implementasi
// bersama di shared/api per domain (auth untuk sesi/OTP/passcode,
// insights untuk pengaturan toko pada onboarding Google) — file ini
// hanya mengekspor ulang yang dipakai fitur auth, mengikuti pola
// fitur pos (satu pintu masuk API per fitur).
export {
  ApiError, hasToken,
  apiGetSettings, apiGoogleLogin, apiHasActiveCashiers, apiListUsers, apiLogin, apiLogout, apiMe,
  apiRegister, apiResetPassword, apiSendOtp, apiSendPasswordResetOtp, apiSetPasscode, apiSwitchAccount,
  apiUpdateSettings, apiVerifyOtp,
} from '../../shared/api'
export type { User } from '../../shared/api'
