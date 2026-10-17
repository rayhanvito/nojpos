import { useEffect, useMemo, useRef, useState } from 'react'
import type { Customer, PayMethod, Product, Trx, Variant } from '../../shared/api'
import { PageHeader } from '../../shared/PageHeader'
import { fmtRp, useDB } from '../../shared/session/store'
import { Receipt } from '../../shared/receipt'
import { isCashPaymentValid, isCreditPaymentValid } from '../../shared/checkout'
import { createClientTxId, enqueueOfflineCheckout, isNetworkFailure, syncOfflineQueue } from '../../shared/offline-queue'
import { AlertStrip, Button, Modal, ModalActions } from '../../shared/ui'
import { apiCheckout } from './api'
import { useCheckout, useCustomerLookup, usePosProducts, usePosSettings, usePromoPreview, useShiftStatus } from './hooks'
import {
  addToCart, buildCheckoutBody, cartItemCount, cartSubtotal, cartSummaryItems, categoryName,
  clampManualDiscount, computeChange, computeTax, computeTotal, effectiveDiscount as resolveEffectiveDiscount,
  filterProducts, isRedeemingPoints, isVoucherDirty, productCategories, removeLine,
  resolveCheckoutShiftGate, resolveCheckoutVoucherCode, setLineQty, taxPctOf, type CartLine,
} from './logic'
import { CartPanel } from './components/CartPanel'
import { PaymentModal } from './components/PaymentModal'
import { ProductGrid } from './components/ProductGrid'
import { VariantPickerModal } from './components/VariantPickerModal'

export default function PosPage() {
  const { session } = useDB()
  const [q, setQ] = useState('')
  const [cat, setCat] = useState('Semua')
  const prod = usePosProducts()
  const products = prod.data
  const setq = usePosSettings()
  const settings = setq.data
  const { hasShift, refreshShift } = useShiftStatus()
  const checkoutMutation = useCheckout()
  const [cart, setCart] = useState<CartLine[]>([])
  const [discount, setDiscount] = useState(0)
  const [voucherCode, setVoucherCode] = useState('')
  const [appliedVoucherCode, setAppliedVoucherCode] = useState<string | null>(null)
  const [previewRevision, setPreviewRevision] = useState(0)
  const [method, setMethod] = useState<PayMethod>('Cash')
  const [paid, setPaid] = useState('')
  const [exactCash, setExactCash] = useState(false)
  const [customerQuery, setCustomerQuery] = useState('')
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null)
  const [variantProduct, setVariantProduct] = useState<Product | null>(null)
  const [pointsRedeem, setPointsRedeem] = useState('')
  const [payOpen, setPayOpen] = useState(false)
  const [cartOpen, setCartOpen] = useState(false)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const [receipt, setReceipt] = useState<Trx | null>(null)
  const [offlineNotice, setOfflineNotice] = useState('')
  const checkoutAttempt = useRef<{ signature: string; clientTxId: string } | null>(null)

  function load() {
    prod.reload()
  }

  const cats = useMemo(() => productCategories(products ?? []), [products])
  const catName = (id: string | null) => categoryName(products ?? [], id)

  const filtered = filterProducts(products ?? [], cat, q)

  const taxPct = taxPctOf(settings)
  const subtotal = cartSubtotal(cart)
  // Nilai potongan promo berasal dari server preview; POS hanya memakai nilai itu
  // pada rincian yang sudah ada agar jumlah pembayaran mengikuti hasil preview.
  const preview = usePromoPreview(cart, appliedVoucherCode, previewRevision)
  const activePromo = preview.promo
  const promoError = preview.error
  const promoNetworkFailure = preview.networkFailure
  const previewLoading = cart.length > 0 && preview.loading
  const checkoutOnline = navigator.onLine
  const promoBlocksCheckout = checkoutOnline && !promoNetworkFailure && (previewLoading || !!promoError)
  const effectiveDiscount = resolveEffectiveDiscount(activePromo, discount)
  const tax = computeTax(subtotal, effectiveDiscount, taxPct)
  const total = computeTotal(subtotal, effectiveDiscount, tax)
  const change = computeChange(paid, total, exactCash)
  const loyaltyActive = !!settings?.loyaltyEnabled
  const redeemingPoints = isRedeemingPoints(method, loyaltyActive, selectedCustomer, pointsRedeem)
  const customerLookup = useCustomerLookup(customerQuery, payOpen && !selectedCustomer)
  const customerResults = customerLookup.results
  const customerLookupError = customerLookup.error
  const customerLookupBusy = customerLookup.busy

  // Nominal bayar selalu untuk total terkini: tiap keranjang, diskon,
  // atau hasil preview berubah, draft Jumlah dibayar dibuang.
  useEffect(() => { setPaid('') }, [cart, effectiveDiscount])

  const voucherDirty = isVoucherDirty(voucherCode, appliedVoucherCode)

  function applyVoucher() {
    setErr('')
    setAppliedVoucherCode(voucherCode.trim() || null)
    setPreviewRevision((revision) => revision + 1)
  }

  function chooseProduct(p: Product) {
    if (p.variants?.length) setVariantProduct(p)
    else add(p)
  }

  function add(p: Product, variant?: Variant) {
    const blocked = addToCart(cart, p, variant, navigator.onLine).blocked
    setCart((c) => addToCart(c, p, variant, navigator.onLine).cart)
    setErr(blocked ? 'Stok tidak cukup.' : '')
  }

  function setQty(line: CartLine, qty: number) {
    const blocked = setLineQty(cart, line, qty, navigator.onLine).blocked
    setCart((c) => setLineQty(c, line, qty, navigator.onLine).cart)
    if (blocked) setErr('Stok tidak cukup.')
  }

  function clearCheckout(reloadProducts = false) {
    setCart([]); setDiscount(0); setVoucherCode(''); setAppliedVoucherCode(null); setPaid(''); setExactCash(false); setPayOpen(false)
    setSelectedCustomer(null); setCustomerQuery(''); setPointsRedeem('')
    if (reloadProducts) load()
  }

  function selectCustomer(customer: Customer) {
    setSelectedCustomer(customer); setCustomerQuery(''); setPointsRedeem('')
  }

  function clearCustomer() {
    setSelectedCustomer(null); setCustomerQuery(''); setPointsRedeem('')
  }

  function changeMethod(m: PayMethod) {
    setMethod(m)
    if (m === 'Credit') { setPaid('0'); setExactCash(false); setPointsRedeem('') }
    else if (method === 'Credit') setPaid('')
  }

  function changeExactCash(v: boolean) {
    setExactCash(v); if (v) setPaid('')
  }

  function changePointsRedeem(v: string) {
    setPointsRedeem(v)
    if (Number(v) > 0) { setExactCash(false); setPaid('') }
  }

  async function checkout() {
    setErr('')
    setOfflineNotice('')
    if (cart.length === 0) return setErr('Keranjang kosong.')
    // Status online + status shift dari CACHE dibaca SEBELUM refetch
    // apa pun (perbaikan bug offline-penuh): bila perangkat sudah
    // offline saat checkout dimulai, langsung ke jalur antrean offline
    // tanpa menunggu refetch status shift yang tertahan offline.
    // Hanya bila cache menyatakan pasti tidak ada shift, checkout
    // diblokir — backend tetap menegakkan 409 SHIFT_REQUIRED saat
    // antrean tersinkron. Jalur online tidak berubah: refresh dulu,
    // dan hanya hasil refresh `false` yang memblokir.
    const onlineAtStart = navigator.onLine
    const gateBeforeRefresh = resolveCheckoutShiftGate({ online: onlineAtStart, cachedHasShift: hasShift })
    if (gateBeforeRefresh === 'block-no-shift') return setErr('Belum ada shift aktif. Buka shift dulu lewat tombol Status di kanan atas.')
    let online = onlineAtStart
    if (gateBeforeRefresh === 'needs-refresh') {
      const refreshed = await refreshShift()
      const gateAfterRefresh = resolveCheckoutShiftGate({ online: true, cachedHasShift: hasShift, refreshedHasShift: refreshed })
      if (gateAfterRefresh === 'block-no-shift') return setErr('Belum ada shift aktif. Buka shift dulu lewat tombol Status di kanan atas.')
      // Koneksi bisa berubah selama refresh: baca ulang agar
      // perangkat yang baru saja offline tetap masuk antrean.
      online = navigator.onLine
    }
    if (online) {
      if (method === 'Credit' && !isCreditPaymentValid(selectedCustomer?.id ?? null)) return setErr('Pilih pelanggan untuk pembayaran Kredit.')
      if (voucherDirty) return setErr('Terapkan kode voucher sebelum melanjutkan.')
      if (previewLoading && !promoNetworkFailure) return setErr('Pemeriksaan promo belum selesai.')
      if (promoError && !promoNetworkFailure) return setErr(promoError)
      if (method === 'Cash' && !isCashPaymentValid(paid, change, exactCash, redeemingPoints)) {
        return setErr(redeemingPoints ? 'Isi jumlah tunai yang dibayar; isi 0 jika poin menutup seluruh total.' : 'Jumlah bayar kurang dari total.')
      }
    }
    const checkoutVoucherCode = resolveCheckoutVoucherCode(online, voucherCode, appliedVoucherCode)
    const checkoutBody = buildCheckoutBody({
      cart, discount, method, paid, exactCash, total,
      voucherCode: checkoutVoucherCode,
      customerId: selectedCustomer?.id,
      redeemingPoints, pointsRedeem,
    })
    const signature = JSON.stringify(checkoutBody)
    const attempt = checkoutAttempt.current?.signature === signature
      ? checkoutAttempt.current
      : { signature, clientTxId: createClientTxId() }
    checkoutAttempt.current = attempt
    const payload = { ...checkoutBody, clientTxId: attempt.clientTxId }
    const summary = {
      items: cartSummaryItems(cart),
      estimatedTotal: total,
    }

    setBusy(true)
    if (!online) {
      try {
        await enqueueOfflineCheckout(payload, summary, session?.storeId)
        checkoutAttempt.current = null
        setOfflineNotice('Transaksi tersimpan di antrean offline. Keranjang siap untuk transaksi berikutnya.')
        clearCheckout()
      } catch (error) {
        setErr(`Transaksi belum dapat disimpan. Keranjang tetap terbuka. ${error instanceof Error ? error.message : 'Periksa penyimpanan browser lalu coba lagi.'}`)
      } finally {
        setBusy(false)
      }
      return
    }

    try {
      const trx = await checkoutMutation.mutateAsync(payload)
      checkoutAttempt.current = null
      setReceipt(trx)
      clearCheckout(true)
    } catch (error) {
      if (isNetworkFailure(error, navigator.onLine)) {
        try {
          await enqueueOfflineCheckout(payload, summary, session?.storeId)
          checkoutAttempt.current = null
          setOfflineNotice('Transaksi tersimpan di antrean. Penyinkronan akan dicoba otomatis saat jaringan tersedia.')
          clearCheckout()
          if (navigator.onLine) void syncOfflineQueue(apiCheckout, session?.storeId).catch(() => {})
        } catch (storageError) {
          setErr(`Transaksi belum dapat disimpan ke antrean. Keranjang tetap terbuka. ${storageError instanceof Error ? storageError.message : 'Periksa penyimpanan browser lalu coba lagi.'}`)
        }
      } else {
        checkoutAttempt.current = null
        setErr(error instanceof Error ? error.message : 'Gagal menyelesaikan transaksi.')
      }
    } finally {
      setBusy(false)
    }
  }

  const cartPanel = (idSuffix: string, listCls: string) => (
    <CartPanel
      idSuffix={idSuffix}
      listCls={listCls}
      hasShift={hasShift}
      cart={cart}
      onRemoveLine={(line) => setCart((c) => removeLine(c, line))}
      onSetQty={setQty}
      voucherCode={voucherCode}
      onVoucherCodeChange={setVoucherCode}
      onApplyVoucher={applyVoucher}
      previewLoading={previewLoading}
      promoError={promoError}
      promoNetworkFailure={promoNetworkFailure}
      activePromo={activePromo}
      subtotal={subtotal}
      discount={discount}
      onDiscountChange={(r) => setDiscount(clampManualDiscount(r, subtotal))}
      taxPct={taxPct}
      tax={tax}
      total={total}
    />
  )

  return (
    <>
      <PageHeader
        title="Kasir"
       
        crumb="Kasir"
      />

      {offlineNotice && <AlertStrip tone="success" className="mb-4">{offlineNotice}</AlertStrip>}


      <div className="grid gap-5 lg:grid-cols-[1fr_320px]">
        <ProductGrid
          q={q}
          onQueryChange={setQ}
          cats={cats}
          cat={cat}
          onCategoryChange={setCat}
          categoryName={catName}
          err={err}
          products={products}
          filtered={filtered}
          cart={cart}
          onChoose={chooseProduct}
        />

        <aside className="hidden flex-col rounded-2xl bg-cream p-4 lg:flex">
          <h2 className="font-mono text-xs uppercase tracking-wider text-fog">Keranjang</h2>
          {cartPanel('', 'mt-3 flex-1 space-y-2 overflow-y-auto')}
          <Button className="mt-4 w-full" disabled={cart.length === 0 || (checkoutOnline && (promoBlocksCheckout || voucherDirty))} onClick={() => { setPayOpen(true); setErr('') }}>Bayar · {fmtRp(total)}</Button>
        </aside>
      </div>

      {cart.length > 0 && (
        <>
          <div className="h-20 lg:hidden" aria-hidden="true" />
          <div className="fixed inset-x-0 bottom-0 z-40 flex items-center gap-3 border-t border-dove bg-paper px-4 py-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] shadow-[0_-4px_16px_rgba(0,0,0,0.06)] lg:hidden">
            <div className="min-w-0 flex-1">
              <p className="text-[11px] leading-tight text-fog">{cartItemCount(cart)} item</p>
              <p className="truncate font-mono text-[15px] font-semibold leading-tight tabular-nums text-fg">{fmtRp(total)}</p>
            </div>
            <Button
              disabled={checkoutOnline && (promoBlocksCheckout || voucherDirty)}
              onClick={() => setCartOpen(true)}
              className="shrink-0 px-6"
            >
              Lihat Keranjang
            </Button>
          </div>
        </>
      )}

      <VariantPickerModal
        product={variantProduct}
        cart={cart}
        onAdd={add}
        onClose={() => setVariantProduct(null)}
      />

      <Modal open={cartOpen} title="Keranjang" onClose={() => setCartOpen(false)}>
        {cartPanel('-sheet', 'space-y-2')}
        <ModalActions>
          <div className="flex w-full items-center gap-3">
            <div className="min-w-0 flex-1">
              <p className="text-[11px] leading-tight text-fog">Total</p>
              <p className="truncate font-mono text-[15px] font-semibold leading-tight tabular-nums text-fg">{fmtRp(total)}</p>
            </div>
            <Button
              disabled={cart.length === 0 || (checkoutOnline && (promoBlocksCheckout || voucherDirty))}
              onClick={() => { setCartOpen(false); setPayOpen(true); setErr('') }}
              className="shrink-0 px-6"
            >
              Bayar
            </Button>
          </div>
        </ModalActions>
      </Modal>

      <PaymentModal
        open={payOpen}
        onClose={() => setPayOpen(false)}
        err={err}
        promoError={promoError}
        promoNetworkFailure={promoNetworkFailure}
        method={method}
        onMethodChange={changeMethod}
        selectedCustomer={selectedCustomer}
        onSelectCustomer={selectCustomer}
        onClearCustomer={clearCustomer}
        customerQuery={customerQuery}
        onCustomerQueryChange={setCustomerQuery}
        customerResults={customerResults}
        customerLookupBusy={customerLookupBusy}
        customerLookupError={customerLookupError}
        loyaltyActive={loyaltyActive}
        pointsRedeem={pointsRedeem}
        onPointsRedeemChange={changePointsRedeem}
        settings={settings}
        exactCash={exactCash}
        onExactCashChange={changeExactCash}
        redeemingPoints={redeemingPoints}
        paid={paid}
        onPaidChange={setPaid}
        change={change}
        activePromo={activePromo}
        discount={discount}
        taxPct={taxPct}
        tax={tax}
        total={total}
        busy={busy}
        onCheckout={checkout}
      />

      <Modal open={!!receipt} title="Transaksi berhasil" onClose={() => setReceipt(null)}>
        {receipt && <Receipt trx={receipt} settings={settings} onClose={() => setReceipt(null)} />}
      </Modal>
    </>
  )
}
