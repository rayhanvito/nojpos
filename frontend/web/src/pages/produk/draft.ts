import type { PriceTier, Product, Variant } from '../../lib/api'

export interface Draft {
  id?: string
  name: string
  sku: string
  barcode: string
  categoryId: string
  buyPrice: string
  sellPrice: string
  stock: string
  minStock: string
  unit: string
  variants: VariantDraft[]
  priceTiers: TierDraft[]
  initialVariantIds: number[]
  initialPriceTierIds: number[]
}

export interface TierDraft { id?: number; minQty: string; sellPrice: string }
export interface VariantDraft {
  id?: number
  name: string
  sku: string
  buyPrice: string
  sellPrice: string
  stock: string
  stockEdited?: boolean
  minStock: string
  priceTiers: TierDraft[]
  initialPriceTierIds: number[]
}

export const emptyDraft: Draft = { name: '', sku: '', barcode: '', categoryId: '', buyPrice: '', sellPrice: '', stock: '', minStock: '', unit: 'pcs', variants: [], priceTiers: [], initialVariantIds: [], initialPriceTierIds: [] }
export const emptyVariant: VariantDraft = { name: '', sku: '', buyPrice: '', sellPrice: '', stock: '0', minStock: '', priceTiers: [], initialPriceTierIds: [] }
export const hasDuplicateTierQty = (tiers: TierDraft[]) => new Set(tiers.map((tier) => Number(tier.minQty))).size !== tiers.length

export function tierDraft(tier: PriceTier): TierDraft {
  return { id: tier.id, minQty: String(tier.min_qty), sellPrice: String(tier.sell_price) }
}

export function variantDraft(variant: Variant): VariantDraft {
  return {
    id: variant.id, name: variant.name, sku: variant.sku,
    buyPrice: String(variant.buy_price), sellPrice: String(variant.sell_price),
    stock: String(variant.stock), minStock: String(variant.min_stock ?? ''),
    priceTiers: (variant.price_tiers ?? []).map(tierDraft),
    initialPriceTierIds: (variant.price_tiers ?? []).map((tier) => tier.id),
  }
}

export function productDraft(product: Product): Draft {
  return {
    id: product.id, name: product.name, sku: product.sku, barcode: product.barcode,
    categoryId: product.category_id ?? '', buyPrice: String(product.buy_price),
    sellPrice: String(product.sell_price), stock: String(product.stock),
    minStock: product.min_stock == null ? '' : String(product.min_stock), unit: product.unit,
    variants: (product.variants ?? []).map(variantDraft),
    priceTiers: (product.price_tiers ?? []).map(tierDraft),
    initialVariantIds: (product.variants ?? []).map((variant) => variant.id),
    initialPriceTierIds: (product.price_tiers ?? []).map((tier) => tier.id),
  }
}
