// API fitur finance: permukaan endpoint yang dipakai halaman
// Hutang Piutang dan Pengeluaran. Endpoint-nya sendiri tetap
// implementasi bersama di shared/api per domain (finance untuk
// hutang/piutang/pengeluaran) — file ini hanya mengekspor ulang yang
// dipakai fitur finance, mengikuti pola fitur pos (satu pintu masuk
// API per fitur).
export {
  apiAddPayablePayment, apiAddReceivablePayment, apiCreateExpense, apiDeleteExpense, apiGetPayable, apiGetReceivable,
  apiListExpenses, apiListPayables, apiListReceivables, apiUpdateExpense, fetchAll,
} from '../../shared/api'
export type {
  DebtPaymentInput, DebtStatus, Expense, ExpenseInput, Payable, PayableDetail, Receivable, ReceivableDetail, StandardPayMethod,
} from '../../shared/api'
