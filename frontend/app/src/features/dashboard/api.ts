// API fitur dashboard: permukaan endpoint yang dipakai halaman Beranda.
// Endpoint-nya sendiri tetap implementasi bersama di shared/api per
// domain (insights untuk /dashboard, sales untuk /transactions) — file
// ini hanya mengekspor ulang yang dipakai fitur dashboard, mengikuti
// pola fitur pos (satu pintu masuk API per fitur).
export { apiGetDashboard, apiListTransactions } from '../../shared/api'
export type { DashboardAdmin, DashboardCashier, Page, Trx } from '../../shared/api'
