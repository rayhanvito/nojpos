import assert from 'node:assert/strict'
import { test } from 'node:test'

const logic = await import('../src/features/pos/logic.ts')
const {
  addToCart, setLineQty, removeLine, lineKey, lineQuantity, availableStock, effectiveStock,
  unitPrice, lineTotal, cartSubtotal, cartItemCount, cartKey, checkoutItems, cartSummaryItems,
  taxPctOf, effectiveDiscount, computeTax, computeTotal, computeChange, clampManualDiscount,
  isVoucherDirty, resolveCheckoutVoucherCode, isRedeemingPoints, addQuickCash, buildCheckoutBody,
  productCategories, categoryName, filterProducts, productDisplayPrice,
} = logic

// Produk/varian contoh minimal — hanya field yang dibaca logic.ts.
function product(over = {}) {
  return {
    id: 'p1', category_id: 'c1', category_name: 'Minuman', name: 'Kopi Susu', sku: 'KS-01',
    barcode: '8991001', buy_price: 8000, sell_price: 15000, stock: 10, unit: 'pcs', active: true,
    created_at: '2026-01-01T00:00:00Z', ...over,
  }
}
function variant(over = {}) {
  return { id: 7, product_id: 'p1', name: 'Besar', sku: 'KS-01-B', buy_price: 10000, sell_price: 20000, stock: 4, min_stock: 0, ...over }
}

test('line total is unit price times qty for a plain product', () => {
  const line = { product: product(), qty: 3 }
  assert.equal(unitPrice(line), 15000)
  assert.equal(lineTotal(line), 45000)
})

test('line total uses the variant price, not the product price', () => {
  const line = { product: product(), variant: variant(), qty: 2 }
  assert.equal(unitPrice(line), 20000)
  assert.equal(lineTotal(line), 40000)
})

test('unit price follows price tiers by qty (highest applicable tier)', () => {
  const p = product({ price_tiers: [{ id: 1, min_qty: 5, sell_price: 12000 }, { id: 2, min_qty: 10, sell_price: 10000 }] })
  assert.equal(unitPrice({ product: p, qty: 4 }), 15000)
  assert.equal(unitPrice({ product: p, qty: 5 }), 12000)
  assert.equal(lineTotal({ product: p, qty: 10 }), 100000)
  const v = variant({ price_tiers: [{ id: 3, min_qty: 3, sell_price: 17000 }] })
  assert.equal(unitPrice({ product: p, variant: v, qty: 2 }), 20000)
  assert.equal(unitPrice({ product: p, variant: v, qty: 3 }), 17000)
})

test('subtotal sums line totals across multiple lines; empty cart is zero', () => {
  const cart = [
    { product: product(), qty: 2 },
    { product: product({ id: 'p2', sell_price: 5000 }), variant: variant({ id: 9, sell_price: 7000 }), qty: 3 },
  ]
  assert.equal(cartSubtotal(cart), 2 * 15000 + 3 * 7000)
  assert.equal(cartSubtotal([]), 0)
  assert.equal(cartItemCount(cart), 5)
  assert.equal(cartItemCount([]), 0)
})

test('manual discount cuts the total and the tax base (actual formula)', () => {
  const subtotal = 100000
  const eff = effectiveDiscount(null, 10000)
  assert.equal(eff, 10000)
  const tax = computeTax(subtotal, eff, 10)
  assert.equal(tax, 9000)
  assert.equal(computeTotal(subtotal, eff, tax), 99000)
  // Tanpa pajak: total = subtotal − diskon persis.
  assert.equal(computeTotal(subtotal, eff, computeTax(subtotal, eff, 0)), 90000)
})

test('active promo preview replaces the manual discount (promo + voucher)', () => {
  const promo = { promo_discount: 5000, voucher_discount: 3000 }
  assert.equal(effectiveDiscount(promo, 99999), 8000)
  assert.equal(effectiveDiscount(null, 1234), 1234)
})

test('tax is Math.round of (subtotal - discount) * pct and taxPct is 0 when disabled', () => {
  assert.equal(computeTax(15555, 0, 10), 1556)
  assert.equal(computeTax(15555, 555, 10), 1500)
  assert.equal(taxPctOf({ taxEnabled: true, taxPct: 11 }), 11)
  assert.equal(taxPctOf({ taxEnabled: false, taxPct: 11 }), 0)
  assert.equal(taxPctOf(null), 0)
  assert.equal(taxPctOf(undefined), 0)
})

test('change is paid minus total, negative when underpaid, zero for exact cash', () => {
  assert.equal(computeChange('50000', 45000, false), 5000)
  assert.equal(computeChange('40000', 45000, false), -5000)
  assert.equal(computeChange('', 45000, false), -45000)
  assert.equal(computeChange('999999', 45000, true), 0)
})

test('adding the same product+variant merges into one line; different variant is a new line', () => {
  const p = product()
  const v = variant()
  let r = addToCart([], p, undefined, true)
  assert.equal(r.blocked, false)
  assert.equal(r.cart.length, 1)
  assert.equal(r.cart[0].qty, 1)
  r = addToCart(r.cart, p, undefined, true)
  assert.equal(r.cart.length, 1)
  assert.equal(r.cart[0].qty, 2)
  r = addToCart(r.cart, p, v, true)
  assert.equal(r.cart.length, 2)
  r = addToCart(r.cart, p, v, true)
  assert.equal(r.cart.length, 2)
  assert.equal(lineQuantity(r.cart, p, v), 2)
  assert.equal(lineKey(p, v), 'p1:7')
  assert.equal(lineKey(p), 'p1:')
})

test('stock guard blocks adds beyond stock only when online', () => {
  const p = product({ stock: 2 })
  let cart = addToCart([], p, undefined, true).cart
  cart = addToCart(cart, p, undefined, true).cart
  const blocked = addToCart(cart, p, undefined, true)
  assert.equal(blocked.blocked, true)
  assert.equal(blocked.cart, cart)
  // Offline: tidak ada penjagaan stok.
  const offline = addToCart(cart, p, undefined, false)
  assert.equal(offline.blocked, false)
  assert.equal(offline.cart[0].qty, 3)
  // Produk habis: tidak bisa ditambah saat online.
  assert.equal(addToCart([], product({ stock: 0 }), undefined, true).blocked, true)
  assert.equal(addToCart([], product({ stock: 0 }), undefined, false).blocked, false)
})

test('setLineQty updates qty, removes below 1, and blocks above stock when online', () => {
  const p = product({ stock: 5 })
  const line = { product: p, qty: 2 }
  const cart = [line]
  assert.equal(setLineQty(cart, line, 4, true).cart[0].qty, 4)
  assert.deepEqual(setLineQty(cart, line, 0, true).cart, [])
  const blocked = setLineQty(cart, line, 6, true)
  assert.equal(blocked.blocked, true)
  assert.equal(blocked.cart[0].qty, 2)
  assert.equal(setLineQty(cart, line, 6, false).cart[0].qty, 6)
})

test('removeLine removes only the matching product+variant line', () => {
  const p = product()
  const v = variant()
  const cart = [{ product: p, qty: 1 }, { product: p, variant: v, qty: 2 }]
  assert.deepEqual(removeLine(cart, cart[0]), [cart[1]])
  assert.deepEqual(removeLine(cart, cart[1]), [cart[0]])
})

test('available and effective stock subtract what is already in the cart', () => {
  const p = product({ stock: 10 })
  const cart = [{ product: p, qty: 3 }]
  assert.equal(availableStock(cart, p), 7)
  assert.equal(effectiveStock(cart, p), 7)
  const pv = product({ stock: 99, variants: [variant({ id: 1, stock: 4 }), variant({ id: 2, stock: 6 })] })
  const cartV = [{ product: pv, variant: pv.variants[0], qty: 4 }, { product: pv, variant: pv.variants[1], qty: 1 }]
  assert.equal(availableStock(cartV, pv, pv.variants[0]), 0)
  assert.equal(effectiveStock(cartV, pv), 5)
})

test('manual discount is clamped between 0 and subtotal', () => {
  assert.equal(clampManualDiscount('5000', 100000), 5000)
  assert.equal(clampManualDiscount('150000', 100000), 100000)
  assert.equal(clampManualDiscount('', 100000), 0)
})

test('cart key, checkout items and summary items follow the cart lines', () => {
  const p = product()
  const v = variant()
  const cart = [{ product: p, qty: 2 }, { product: p, variant: v, qty: 1 }]
  assert.equal(cartKey(cart), JSON.stringify([{ productId: 'p1', variantId: undefined, qty: 2 }, { productId: 'p1', variantId: 7, qty: 1 }]))
  assert.deepEqual(checkoutItems(cart), [{ productId: 'p1', qty: 2 }, { productId: 'p1', variantId: 7, qty: 1 }])
  assert.deepEqual(cartSummaryItems(cart), [{ name: 'Kopi Susu', qty: 2 }, { name: 'Kopi Susu · Besar', qty: 1 }])
})

test('voucher dirty check and checkout voucher resolution (offline keeps unapplied code)', () => {
  assert.equal(isVoucherDirty(' ABC ', 'ABC'), false)
  assert.equal(isVoucherDirty('ABC', null), true)
  assert.equal(isVoucherDirty('', null), false)
  assert.equal(resolveCheckoutVoucherCode(true, 'BARU', 'LAMA'), 'LAMA')
  assert.equal(resolveCheckoutVoucherCode(false, 'BARU', 'LAMA'), 'BARU')
  assert.equal(resolveCheckoutVoucherCode(false, 'LAMA', 'LAMA'), 'LAMA')
  assert.equal(resolveCheckoutVoucherCode(false, '', null), null)
})

test('redeeming points requires non-credit method, loyalty on, a customer and points > 0', () => {
  const cust = { id: 3 }
  assert.equal(isRedeemingPoints('Cash', true, cust, '100'), true)
  assert.equal(isRedeemingPoints('Credit', true, cust, '100'), false)
  assert.equal(isRedeemingPoints('Cash', false, cust, '100'), false)
  assert.equal(isRedeemingPoints('Cash', true, null, '100'), false)
  assert.equal(isRedeemingPoints('Cash', true, cust, '0'), false)
  assert.equal(isRedeemingPoints('Cash', true, cust, ''), false)
})

test('quick cash adds to the current paid amount', () => {
  assert.equal(addQuickCash('', 50000), '50000')
  assert.equal(addQuickCash('20000', 50000), '70000')
  assert.equal(addQuickCash('bukan-angka', 1000), '1000')
})

test('checkout body matches the payload built by the page today', () => {
  const p = product()
  const cart = [{ product: p, qty: 2 }]
  assert.deepEqual(
    buildCheckoutBody({ cart, discount: 1000, method: 'Cash', paid: '50000', exactCash: false, total: 29000, voucherCode: null, redeemingPoints: false, pointsRedeem: '' }),
    { items: [{ productId: 'p1', qty: 2 }], discount: 1000, method: 'Cash', paid: 50000 },
  )
  // Uang pas: paid = total persis.
  assert.equal(buildCheckoutBody({ cart, discount: 0, method: 'Cash', paid: '', exactCash: true, total: 30000, voucherCode: null, redeemingPoints: false, pointsRedeem: '' }).paid, 30000)
  // Non-tunai: paid selalu 0; voucher + pelanggan + poin hanya bila ada.
  assert.deepEqual(
    buildCheckoutBody({ cart, discount: 0, method: 'QRIS', paid: '999', exactCash: false, total: 30000, voucherCode: 'HEMAT', customerId: 3, redeemingPoints: true, pointsRedeem: '25' }),
    { items: [{ productId: 'p1', qty: 2 }], discount: 0, method: 'QRIS', paid: 0, voucherCode: 'HEMAT', customerId: 3, pointsRedeem: 25 },
  )
})

test('catalog helpers: categories, names, search filter and display price', () => {
  const items = [
    product(),
    product({ id: 'p2', category_id: 'c2', category_name: 'Makanan', name: 'Nasi Goreng', sku: 'NG-01', barcode: '8992002' }),
    product({ id: 'p3', category_id: null, category_name: null, name: 'Teh Manis', sku: 'TM-01', barcode: '8993003', variants: [variant({ id: 11, sku: 'TM-01-B', sell_price: 8000 })] }),
  ]
  assert.deepEqual(productCategories(items), ['c1', 'c2'])
  assert.equal(categoryName(items, 'c2'), 'Makanan')
  assert.equal(categoryName(items, 'tidak-ada'), '—')
  assert.equal(filterProducts(items, 'Semua', '').length, 3)
  assert.deepEqual(filterProducts(items, 'c1', '').map((p) => p.id), ['p1'])
  assert.deepEqual(filterProducts(items, 'Semua', 'nasi').map((p) => p.id), ['p2'])
  assert.deepEqual(filterProducts(items, 'Semua', '8993003').map((p) => p.id), ['p3'])
  assert.deepEqual(filterProducts(items, 'Semua', 'tm-01-b').map((p) => p.id), ['p3'])
  assert.equal(productDisplayPrice(items[0]), 15000)
  assert.equal(productDisplayPrice(items[2]), 8000)
})
