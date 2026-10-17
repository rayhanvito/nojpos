export interface PriceTierPrice {
  min_qty: number
  sell_price: number
}

export function getUnitPrice(qty: number, sellPrice: number, tiers: readonly PriceTierPrice[] = []) {
  const tier = tiers.reduce<PriceTierPrice | undefined>(
    (best, current) => current.min_qty <= qty && (!best || current.min_qty > best.min_qty) ? current : best,
    undefined,
  )
  return tier?.sell_price ?? sellPrice
}
