// API fitur customer: permukaan endpoint yang dipakai halaman
// Pelanggan. Endpoint-nya sendiri tetap implementasi bersama di
// shared/api per domain (finance untuk pelanggan/piutang) — file ini
// hanya mengekspor ulang yang dipakai fitur customer, mengikuti pola
// fitur pos (satu pintu masuk API per fitur).
export {
  apiCreateCustomer, apiCustomerTransactions, apiDeleteCustomer, apiGetCustomer, apiListCustomers, apiListReceivables, apiUpdateCustomer,
} from '../../shared/api'
export type { Customer, CustomerInput, CustomerTransaction, Page, Receivable } from '../../shared/api'
