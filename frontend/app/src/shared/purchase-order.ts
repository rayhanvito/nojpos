function nonnegativeInteger(value: string | number): number {
  const number = Number(value)
  return Number.isFinite(number) ? Math.max(0, Math.trunc(number)) : 0
}

export function purchaseOrderProductId(value: string | number): number | null {
  const id = Number(value)
  return Number.isSafeInteger(id) && id > 0 ? id : null
}

export function isCurrentPurchaseOrderRequest(requestToken: number, currentToken: number): boolean {
  return requestToken === currentToken
}

export function purchaseOrderTotal(items: { qty: string | number; buyPrice: string | number }[]): number {
  return items.reduce((total, item) => total + nonnegativeInteger(item.qty) * nonnegativeInteger(item.buyPrice), 0)
}

export function remainingToReceive(item: { qty_ordered: number; qty_received: number }): number {
  return Math.max(0, item.qty_ordered - item.qty_received)
}

export function purchaseOrderItemName(item: {
  product_id: string | number
  product_name?: string
  variant_id?: number | null
  variant_name?: string
}, products: { id: string | number; name: string; variants?: { id: number; name: string }[] }[] = []): string {
  const productId = purchaseOrderProductId(item.product_id)
  const product = products.find((entry) => purchaseOrderProductId(entry.id) === productId)
  const productName = item.product_name || product?.name || `Produk #${item.product_id}`
  if (item.variant_id == null) return productName
  const variant = product?.variants?.find((entry) => entry.id === item.variant_id)
  return `${productName} · ${item.variant_name || variant?.name || `Varian #${item.variant_id}`}`
}
