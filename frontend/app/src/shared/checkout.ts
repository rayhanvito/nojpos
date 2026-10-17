export function isCashPaymentValid(paid: string, change: number, exactCash: boolean, redeemingPoints: boolean) {
  if (exactCash) return true
  if (!paid.trim()) return false
  const amount = Number(paid)
  if (!Number.isFinite(amount) || amount < 0) return false
  if (redeemingPoints) return true
  return amount > 0 && change >= 0
}

export function isCreditPaymentValid(customerId: number | null) {
  return Number.isSafeInteger(customerId) && (customerId ?? 0) > 0
}
