// Semua tipe data kontrak API (respons server + masukan endpoint).

export type Role = 'admin' | 'cashier'
export type PayMethod = 'Cash' | 'Bank Transfer' | 'QRIS' | 'E-Wallet' | 'Card' | 'Credit'
export type StandardPayMethod = Exclude<PayMethod, 'Credit'>

export interface User {
  id: string
  email: string
  name: string
  role: Role
  active: boolean
  store_id: string
  store_name: string
  outlet_id?: string | null
  outlet_name?: string
  has_passcode: boolean
  created_at?: string
  // Presence (kontrak docs/API-CONTRACT-PRESENCE.md; absen = backend lama).
  online?: boolean
  last_seen_at?: string
}

export interface Outlet {
  id: string
  name: string
  address: string
  is_default: boolean
}

export interface AuthResp {
  access_token: string
  refresh_token: string
  user: User
}

export type UserResponse = Omit<User, 'outlet_id'> & { outlet_id?: string | number | null }
export type OutletResponseItem = Omit<Outlet, 'id'> & { id: string | number }

export function normalizeUser(user: UserResponse): User {
  return { ...user, outlet_id: user.outlet_id == null ? user.outlet_id : String(user.outlet_id) }
}

export function normalizeAuthResponse(response: Omit<AuthResp, 'user'> & { user: UserResponse }): AuthResp {
  return { ...response, user: normalizeUser(response.user) }
}

export function normalizeOutlet(outlet: OutletResponseItem): Outlet {
  return { ...outlet, id: String(outlet.id) }
}

export function outletIdForWire(id: string): string | number {
  const numeric = Number(id)
  return /^\d+$/.test(id) && Number.isSafeInteger(numeric) ? numeric : id
}

export interface Category {
  id: string
  name: string
  active: boolean
  created_at: string
}

export interface Customer {
  id: number
  name: string
  phone: string
  email: string
  address: string
  points: number
  created_at: string
  updated_at: string
}

export interface CustomerTransaction {
  id: string
  total: number
  method: string
  status: Trx['status']
  created_at: string
  points_earned?: number
  points_redeemed?: number
  points_discount?: number
}

export type DebtStatus = 'unpaid' | 'partially_paid' | 'paid'

export interface Payment {
  id: number
  amount: number
  method: StandardPayMethod
  created_at: string
}

export interface Receivable {
  id: number
  customer_id: number
  customer_name?: string
  transaction_id: number
  amount: number
  paid_amount?: number
  remaining: number
  due_date: string
  status: DebtStatus
}

export type ReceivableDetail = Receivable & { payments: Payment[] }
export type ReceivableDetailResponse = Omit<ReceivableDetail, 'payments'> & { payments?: Payment[] | null }

export interface Payable {
  id: number
  supplier_id: number
  supplier_name?: string
  purchase_order_id: number
  po_number?: string
  purchase_order_number?: string
  amount: number
  remaining: number
  status: DebtStatus
  created_at: string
}

export type PayableDetail = Payable & { payments: Payment[] }
export type PayableDetailResponse = Omit<PayableDetail, 'payments'> & { payments?: Payment[] | null }
export type DebtPaymentInput = { amount: number; method: StandardPayMethod }

export interface Supplier {
  id: number
  name: string
  contact: string
  address: string
}

export type SupplierInput = Pick<Supplier, 'name' | 'contact' | 'address'>
export type PurchaseOrderStatus = 'draft' | 'ordered' | 'partially_received' | 'fully_received' | 'cancelled'

export interface PurchaseOrder {
  id: number
  number: string
  supplier_id: number
  supplier_name?: string
  supplier?: Supplier
  status: PurchaseOrderStatus
  created_at: string
  total?: number
}

export interface PurchaseOrderItem {
  id: number
  product_id: number
  product_name?: string
  variant_id?: number | null
  variant_name?: string
  qty_ordered: number
  qty_received: number
  buy_price: number
}

export type PurchaseOrderDetail = PurchaseOrder & { items: PurchaseOrderItem[] }
export type PurchaseOrderInput = {
  supplierId: number
  items: { productId: number; variantId?: number; qty: number; buyPrice: number }[]
}
export type PurchaseOrderReceiptInput = { items: { itemId: number; qty: number }[] }

export type StockOpnameStatus = 'counting' | 'posted' | 'cancelled'

export interface StockOpnameSession {
  id: number
  status: StockOpnameStatus
  note: string
  created_by?: string | number
  created_at: string
  posted_at?: string | null
}

export interface StockOpnameItem {
  id: number
  product_id: string | number
  product_name?: string | null
  variant_id: number | null
  variant_name?: string | null
  system_stock: number
  counted_stock: number | null
  difference: number | null
}

export type StockOpnameDetail = StockOpnameSession & { items: StockOpnameItem[] }
export type StockOpnameCountsInput = { items: { itemId: number; countedStock: number }[] }

export interface Promo {
  id: number
  name: string
  type: 'percent' | 'amount'
  value: number
  scope: 'all' | 'category' | 'product'
  category_id: number | null
  product_id: number | null
  voucher_code: string | null
  starts_at: string
  ends_at: string
  happy_start: string | null
  happy_end: string | null
  min_purchase: number
  max_discount: number
  active: boolean
  combinable: boolean
  created_at?: string
}

export interface PromoPreview {
  promo_id: number
  promo_name: string
  promo_discount: number
  voucher_id: number
  voucher_name: string
  voucher_code: string
  voucher_discount: number
}

export interface CheckoutItem {
  productId: string
  variantId?: number
  qty: number
}

export interface Product {
  id: string
  category_id: string | null
  category_name: string | null
  name: string
  sku: string
  barcode: string
  buy_price: number
  sell_price: number
  stock: number
  min_stock?: number
  unit: string
  active: boolean
  created_at: string
  variants?: Variant[]
  price_tiers?: PriceTier[]
}

export interface PriceTier {
  id: number
  min_qty: number
  sell_price: number
}

export interface Variant {
  id: number
  product_id: number | string
  name: string
  sku: string
  buy_price: number
  sell_price: number
  stock: number
  min_stock: number
  price_tiers?: PriceTier[]
}

export interface TrxItem {
  product_id: string
  name: string
  buy_price: number
  price: number
  qty: number
}

export interface Trx {
  id: string
  seq?: number
  cashier_name: string
  items: TrxItem[]
  subtotal: number
  discount: number
  promo_id?: number | null
  promo_name?: string | null
  promo_discount?: number
  voucher_id?: number | null
  voucher_code?: string | null
  voucher_discount?: number
  tax: number
  total: number
  method: string
  paid: number
  change: number
  status: 'completed' | 'completed_offline_shortage' | 'pending' | 'cancelled' | 'refunded'
  final_stocks?: { product_id: string; variant_id: number | null; stock: number }[]
  customer: string
  customer_id?: number | null
  points_earned?: number
  points_redeemed?: number
  points_discount?: number
  created_at: string
}

export interface Movement {
  id: string
  product_id: string
  product_name: string | null
  type: 'sale' | 'refund' | 'adjust' | 'initial'
  qty: number
  reason: string
  actor: string
  created_at: string
}

export interface StockTransferItemResponse {
  product_id: string | number
  variant_id?: number | null
  qty: number
}

export interface StockTransferResponse {
  id: string | number
  from_outlet_id: string | number
  from_outlet_name?: string | null
  to_outlet_id: string | number
  to_outlet_name?: string | null
  items?: StockTransferItemResponse[] | null
  note?: string | null
  created_by?: string | number | null
  created_by_name?: string | null
  created_at: string
}

export interface StockTransferItem {
  productId: string
  variantId: number | null
  qty: number
}

export interface StockTransfer {
  id: string
  fromOutletId: string
  fromOutletName?: string
  toOutletId: string
  toOutletName?: string
  items: StockTransferItem[]
  note: string
  createdBy: string
  createdAt: string
}

export type StockTransferInput = {
  fromOutletId: string | number
  toOutletId: string | number
  items: { productId: string | number; variantId: number | null; qty: number }[]
  note: string
}

export function normalizeStockTransfer(transfer: StockTransferResponse): StockTransfer {
  return {
    id: String(transfer.id),
    fromOutletId: String(transfer.from_outlet_id),
    fromOutletName: transfer.from_outlet_name ?? undefined,
    toOutletId: String(transfer.to_outlet_id),
    toOutletName: transfer.to_outlet_name ?? undefined,
    items: (transfer.items ?? []).map((item) => ({
      productId: String(item.product_id),
      variantId: item.variant_id ?? null,
      qty: item.qty,
    })),
    note: transfer.note ?? '',
    createdBy: transfer.created_by_name ?? (transfer.created_by == null ? '' : String(transfer.created_by)),
    createdAt: transfer.created_at,
  }
}

export interface StoreHours {
  days: string
  open: string | null
  close: string | null
}

export interface StoreSettings {
  storeName: string
  address: string
  phone: string
  taxEnabled: boolean
  taxPct: number
  receiptHeader: string
  receiptFooter: string
  paper: string
  timezone: string
  // Extended (kontrak docs/API-CONTRACT-SETTINGS-EXTENDED.md, live backend
  // 11 Sep commit 605550d). Opsional agar respons backend lama + adapter
  // offline (tanpa key baru) tetap kompatibel; deteksi dukungan via
  // `'businessType' in settings`. logoUrl, receiptShowLogo &
  // passcodeUpdatedAt sengaja tidak ada (fitur logo dibatalkan,
  // passcodeUpdatedAt ditunda di sisi backend).
  businessType?: string
  email?: string
  city?: string
  province?: string
  currency?: string
  hours?: StoreHours[]
  receiptShowCashier?: boolean
  receiptShowMethod?: boolean
  receiptShowTax?: boolean
  receiptShowDiscount?: boolean
  receiptShowNote?: boolean
  taxName?: string
  taxInclusive?: boolean
  taxRounding?: string
  taxApplyTo?: string
  loyaltyEnabled?: boolean
  loyaltyEarnPer?: number
  loyaltyPointValue?: number
  lowStockDefault?: number
  creditDueDays?: number
}

export type StoreSettingsResponse = StoreSettings & {
  loyalty_enabled?: boolean
  loyalty_earn_per?: number
  loyalty_point_value?: number
  low_stock_default?: number
  credit_due_days?: number
}

export type OutletSettingsOverrides = Partial<Pick<StoreSettings, 'taxPct' | 'receiptFooter' | 'hours'>>
export type OutletSettings = StoreSettings & { overrides: OutletSettingsOverrides }
export type OutletSettingsResponse = StoreSettingsResponse & { overrides?: OutletSettingsOverrides | null }
export type OutletSettingsPatch = {
  taxPct?: number | null
  receiptFooter?: string | null
  hours?: StoreHours[] | null
}

export function normalizeStoreSettings(settings: StoreSettingsResponse): StoreSettings {
  const { loyalty_enabled, loyalty_earn_per, loyalty_point_value, low_stock_default, credit_due_days, ...rest } = settings
  return {
    ...rest,
    loyaltyEnabled: loyalty_enabled ?? settings.loyaltyEnabled,
    loyaltyEarnPer: loyalty_earn_per ?? settings.loyaltyEarnPer,
    loyaltyPointValue: loyalty_point_value ?? settings.loyaltyPointValue,
    lowStockDefault: low_stock_default ?? settings.lowStockDefault,
    creditDueDays: credit_due_days ?? settings.creditDueDays ?? 14,
  }
}

export function normalizeOutletSettings(settings: OutletSettingsResponse): OutletSettings {
  const { overrides, ...effective } = settings
  return { ...normalizeStoreSettings(effective), overrides: overrides ?? {} }
}

export interface Page<T> {
  items: T[]
  total: number
  page: number
  limit: number
}

export interface ShiftSummary {
  sales: number
  trx_count: number
  sales_by_method: Partial<Record<PayMethod, number>>
  refund_total: number
  cash_expenses?: number
  expected_cash: number | null
  closing_cash_actual: number | null
  cash_variance: number | null
  close_note?: string | null
}

export interface CashierShift {
  cash_expenses?: number
  started_at: string
  opening_cash: number
  sales: number
  trx_count: number
  sales_by_method: Partial<Record<PayMethod, number>>
  refund_total: number
  expected_cash: number
}

export interface Shift extends ShiftSummary {
  id: number
  cashier_id: number
  cashier_name: string
  started_at: string
  closed_at: string | null
  opening_cash: number
  closed_by_name?: string | null
  outlet_id?: number | null
}

export interface ShiftDetail {
  shift: Shift
  transactions: Page<Trx>
  hourly?: { hour: number; omzet: number }[]
  top_products?: { product_id: number; name: string; qty: number }[]
}

export interface ShiftCloseResponse {
  message: string
  summary: ShiftSummary
}

export interface DashboardAdmin {
  role: 'admin'
  today: { omzet: number; trx_count: number; items_sold: number; low_stock: number }
  sales7: { date: string; omzet: number }[]
  methods: { method: string; total: number }[]
  top_products: { product_id: string; name: string; qty: number; revenue: number }[]
  recent: { id: string; cashier_name: string; total: number; status: Trx['status']; time: string }[]
}

export interface DashboardCashier {
  role: 'cashier'
  today: { omzet: number; trx_count: number; items_sold: number }
  recent: { id: string; cashier_name: string; total: number; status: Trx['status']; time: string }[]
  // Opsional: mock /demo mengembalikannya (scope kasir sendiri);
  // backend live boleh tidak mengirim — dashboard fallback ke [].
  sales7?: { date: string; omzet: number }[]
  methods?: { method: string; total: number }[]
  top_products?: { product_id: string; name: string; qty: number; revenue: number }[]
}

export interface ReportBundle {
  period: string
  summary: { omzet: number; trx_count: number; items_sold: number; gross_profit: number; promo_discount_total?: number; voucher_discount_total?: number; expense_total?: number }
  by_method: { method: string; total: number }[]
  by_status: { status: string; count: number }[]
  products: { product_id: string; name: string; sku: string; qty: number; revenue: number; profit: number }[]
  transactions: { date: string; id: string; cashier: string; method: string; total: number; hpp: number; profit: number; status: string }[]
  cashiers?: { cashier_id: number; name: string; trx: number; items: number; gross: number; refund: number; net: number; top_product: string; top_product_qty: number }[]
  expenses?: { category: string; total: number; count: number }[]
  stock: { name: string; sku: string; stock: number; buy_price: number; sell_price: number; stock_value: number }[]
}

export interface TaxReport {
  month: string
  omzet_bruto: number
  refund_total: number
  omzet_neto: number
  dpp: number
  ppn_terkumpul: number
  trx_count: number
  refund_count: number
  pph_final_estimate: number
}

// ── auth ─────────────────────────────────────────────────────────────

// Faktor kedua login email/sandi: `passcode` (lama) atau `otp` (baru,
// kontrak docs/API-CONTRACT-LOGIN-OTP.md). Google & switch tidak berubah.

export type PromoInput = {
  name: string
  type: Promo['type']
  value: number
  scope: Promo['scope']
  categoryId?: number | null
  productId?: number | null
  voucherCode?: string | null
  startsAt: string
  endsAt: string
  happyStart?: string | null
  happyEnd?: string | null
  minPurchase: number
  maxDiscount?: number
  active?: boolean
  combinable?: boolean
}

export type CustomerInput = { name: string; phone?: string; email?: string; address?: string }

export interface ProductFilter {
  q?: string
  categoryId?: string
  active?: boolean
  page?: number
  limit?: number
}

export interface Expense {
  id: number
  store_id: number
  outlet_id?: number | null
  date: string
  category: string
  note: string
  amount: number
  method: string
  source: string
  shift_id?: number | null
  created_by_id: number
  created_by_name: string
  created_at: string
}

export interface ExpenseInput {
  date: string
  category: string
  note: string
  amount: number
  method: string
  source: string
}

export interface ActivityItem {
  id: number | string
  actor_id: string
  actor_name: string
  action: string
  detail: string
  reference_type?: string
  reference_id?: string
  created_at: string
}

export interface CheckoutBody {
  items: CheckoutItem[]
  clientTxId?: string
  discount?: number
  voucherCode?: string
  method: string
  paid?: number
  customer?: string
  customerId?: number
  pointsRedeem?: number
}

export type NotifCategory = 'stok' | 'transaksi' | 'sistem'

export interface Notification {
  id: number | string
  title: string
  message: string
  // widened ke string: tipe lama ('info'|'warning'|'alert'|'low_stock')
  // + tipe baru kontrak ('transaction_created', 'refund_created', …).
  type: string
  category?: NotifCategory
  actor_name?: string
  reference_type?: string
  reference_id?: number | string
  read: boolean
  created_at: string
}
