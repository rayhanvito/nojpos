// API fitur shift: permukaan endpoint yang dipakai halaman Shift
// (tab Shift di Laporan dan mode mandiri). Endpoint-nya sendiri tetap
// implementasi bersama di shared/api per domain (insights untuk
// /shifts) — file ini hanya mengekspor ulang yang dipakai fitur shift,
// mengikuti pola fitur pos (satu pintu masuk API per fitur).
export {
  apiExportShifts, apiForceCloseShift, apiGetShift, apiListShifts,
} from '../../shared/api'
export type { Shift, ShiftDetail } from '../../shared/api'
