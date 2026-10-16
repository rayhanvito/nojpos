type StockOpnameCountItem = { id: number; system_stock: number; counted_stock: number | null }
type StockOpnameDraftItem = { id: number; counted_stock: number | null }

export function stockOpnameChangedItems<T extends StockOpnameDraftItem>(items: T[], counts: Record<number, string>): T[] {
  return items.filter((item) => (counts[item.id] ?? '') !== (item.counted_stock === null ? '' : String(item.counted_stock)))
}

export function stockOpnameCount(value: number | string | null): number | null {
  if (value === null || (typeof value === 'string' && value.trim() === '')) return null
  const count = typeof value === 'number' ? value : Number(value)
  return Number.isSafeInteger(count) && count >= 0 ? count : null
}

export function stockOpnameDifference(systemStock: number, countedStock: number | string | null): number | null {
  const count = stockOpnameCount(countedStock)
  return count === null ? null : count - systemStock
}

export function countStockOpnameDifferences(items: StockOpnameCountItem[], counts?: Record<number, string>): number {
  return items.reduce((total, item) => {
    const count = counts?.[item.id] === undefined ? item.counted_stock : counts[item.id]
    const difference = stockOpnameDifference(item.system_stock, count)
    return difference !== null && difference !== 0 ? total + 1 : total
  }, 0)
}
