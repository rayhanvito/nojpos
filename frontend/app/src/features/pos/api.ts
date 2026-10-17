// API fitur pos: permukaan endpoint yang dipakai halaman Kasir.
// Endpoint-nya sendiri tetap implementasi bersama di shared/api per
// domain (katalog, penjualan, keuangan, insights) — file ini hanya
// mengekspor ulang yang dipakai fitur pos, mengikuti pola fitur promo
// (satu pintu masuk API per fitur).
export {
  apiCheckout,
  apiGetCashierShift,
  apiGetSettings,
  apiListCustomers,
  apiListProducts,
  apiPreviewPromo,
  fetchAll,
} from '../../shared/api'
export type { CheckoutBody } from '../../shared/api'
