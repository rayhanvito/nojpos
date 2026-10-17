// Endpoint keuangan: pelanggan, hutang/piutang, supplier, PO, pengeluaran.

import { request, qs } from './core.ts'
import type { Customer, CustomerTransaction, Receivable, ReceivableDetailResponse, Payable, PayableDetailResponse, DebtPaymentInput, Supplier, SupplierInput, PurchaseOrder, PurchaseOrderDetail, PurchaseOrderInput, PurchaseOrderReceiptInput, Page, CustomerInput, Expense, ExpenseInput } from './types.ts'

export function apiListCustomers(f: { search?: string; page?: number; limit?: number } = {}) {
  // Backend membaca parameter "q" (bukan "search") dan respons membawa total;
  // keduanya menentukan pencarian & status tombol muat-sebanyaknya di pemanggil.
  return request<Page<Customer>>('GET', '/customers' + qs({ q: f.search, page: f.page, limit: f.limit }))
}

export function apiGetCustomer(id: number) {
  return request<Customer>('GET', `/customers/${id}`)
}

export function apiCreateCustomer(body: CustomerInput) {
  return request<Customer>('POST', '/customers', body)
}

export function apiUpdateCustomer(id: number, body: CustomerInput) {
  return request<Customer>('PUT', `/customers/${id}`, body)
}

export function apiDeleteCustomer(id: number) {
  return request<{ message: string }>('DELETE', `/customers/${id}`)
}

export function apiCustomerTransactions(id: number) {
  return request<{ items: CustomerTransaction[] | null }>('GET', `/customers/${id}/transactions`).then((d) => d.items ?? [])
}

export function apiListReceivables() {
  return request<{ items: Receivable[] | null }>('GET', '/receivables').then((d) => d.items ?? [])
}

export function apiGetReceivable(id: number) {
  return request<ReceivableDetailResponse>('GET', `/receivables/${id}`).then((detail) => ({
    ...detail,
    payments: detail.payments ?? [],
  }))
}

export function apiAddReceivablePayment(id: number, body: DebtPaymentInput): Promise<void> {
  return request<unknown>('POST', `/receivables/${id}/payments`, body).then(() => undefined)
}

export function apiListPayables() {
  return request<{ items: Payable[] | null }>('GET', '/payables').then((d) => d.items ?? [])
}

export function apiGetPayable(id: number) {
  return request<PayableDetailResponse>('GET', `/payables/${id}`).then((detail) => ({
    ...detail,
    payments: detail.payments ?? [],
  }))
}

export function apiAddPayablePayment(id: number, body: DebtPaymentInput): Promise<void> {
  return request<unknown>('POST', `/payables/${id}/payments`, body).then(() => undefined)
}

export function apiListSuppliers() {
  return request<{ items: Supplier[] | null }>('GET', '/suppliers').then((d) => d.items ?? [])
}

export function apiCreateSupplier(body: SupplierInput) {
  return request<Supplier>('POST', '/suppliers', body)
}

export function apiUpdateSupplier(id: number, body: SupplierInput) {
  return request<Supplier>('PUT', `/suppliers/${id}`, body)
}

export function apiDeleteSupplier(id: number) {
  return request<{ message: string }>('DELETE', `/suppliers/${id}`)
}

export function apiListPurchaseOrders() {
  return request<{ items: PurchaseOrder[] | null }>('GET', '/purchase-orders').then((d) => d.items ?? [])
}

export function apiGetPurchaseOrder(id: number) {
  return request<PurchaseOrderDetail>('GET', `/purchase-orders/${id}`)
}

export function apiCreatePurchaseOrder(body: PurchaseOrderInput) {
  return request<PurchaseOrderDetail>('POST', '/purchase-orders', body)
}

export function apiUpdatePurchaseOrder(id: number, body: PurchaseOrderInput) {
  return request<PurchaseOrderDetail>('PUT', `/purchase-orders/${id}`, body)
}

export function apiDeletePurchaseOrder(id: number) {
  return request<{ message: string }>('DELETE', `/purchase-orders/${id}`)
}

export function apiOrderPurchaseOrder(id: number) {
  return request<PurchaseOrderDetail>('POST', `/purchase-orders/${id}/order`)
}

export function apiReceivePurchaseOrder(id: number, body: PurchaseOrderReceiptInput) {
  return request<PurchaseOrderDetail>('POST', `/purchase-orders/${id}/receive`, body)
}

export function apiCancelPurchaseOrder(id: number) {
  return request<PurchaseOrderDetail>('POST', `/purchase-orders/${id}/cancel`)
}

export function apiListExpenses(f: { from?: string; to?: string; category?: string; q?: string; page?: number; limit?: number } = {}) {
  return request<Page<Expense>>('GET', '/expenses' + qs(f))
}

export function apiCreateExpense(body: ExpenseInput) {
  return request<Expense>('POST', '/expenses', body)
}

export function apiUpdateExpense(id: number, body: ExpenseInput) {
  return request<Expense>('PUT', `/expenses/${id}`, body)
}

export function apiDeleteExpense(id: number) {
  return request<{ ok: boolean }>('DELETE', `/expenses/${id}`)
}
