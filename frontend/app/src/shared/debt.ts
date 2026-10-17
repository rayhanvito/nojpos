import type { DebtStatus } from './api'

export type DebtFilter = 'all' | DebtStatus | 'overdue'

function localDate(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

export function isDebtOverdue(status: DebtStatus, dueDate?: string, today = localDate(new Date())) {
  const date = dueDate?.slice(0, 10)
  return status !== 'paid' && !!date && /^\d{4}-\d{2}-\d{2}$/.test(date) && date < today
}

export function debtStatusLabel(status: DebtStatus, dueDate?: string, today?: string) {
  if (isDebtOverdue(status, dueDate, today)) return 'Jatuh Tempo'
  return { unpaid: 'Belum Lunas', partially_paid: 'Sebagian', paid: 'Lunas' }[status]
}

export function filterDebtItems<T extends { status: DebtStatus; due_date?: string }>(items: T[], filter: DebtFilter, today?: string) {
  if (filter === 'all') return items
  if (filter === 'overdue') return items.filter((item) => isDebtOverdue(item.status, item.due_date, today))
  return items.filter((item) => item.status === filter)
}

export function isDebtPaymentAmountValid(value: string, remaining: number) {
  if (!/^\d+$/.test(value)) return false
  const amount = Number(value)
  return Number.isSafeInteger(amount) && amount > 0 && amount <= remaining
}

export function canSubmitDebtPayment(value: string, remaining: number, detailNeedsRefresh: boolean) {
  return !detailNeedsRefresh && isDebtPaymentAmountValid(value, remaining)
}
