import { useEffect, useMemo, useRef, useState } from 'react'
import { ApiError, apiCheckout, apiGetCashierShift, apiGetSettings, apiListCustomers, apiListProducts, apiPreviewPromo, fetchAll, type CheckoutBody, type Customer, type PayMethod, type Product, type PromoPreview, type StoreSettings, type Trx, type Variant } from '../lib/api'
import { PageHeader } from '../lib/PageHeader'
import { useCache } from '../lib/cache'
import { fmtRp, useDB } from '../lib/store'
import { Receipt } from '../lib/receipt'
import { isCashPaymentValid, isCreditPaymentValid } from '../lib/checkout'
import { createClientTxId, enqueueOfflineCheckout, isNetworkFailure, syncOfflineQueue } from '../lib/offline-queue'
import { getUnitPrice } from '../lib/price-tier'
import { CheckRow, AlertStrip, NumInput, Button, Modal, ModalActions } from '../lib/ui'
import { Skeleton } from '@/components/ui/skeleton'

// Nominal cepat tunai: ketuk untuk menambah ke Jumlah dibayar.
const CASH_QUICK: { v: number; label: string }[] = [
  { v: 500, label: '500' }, { v: 1000, label: '1rb' },
  { v: 2000, label: '2rb' }, { v: 5000, label: '5rb' },
  { v: 10000, label: '10rb' }, { v: 20000, label: '20rb' },
  { v: 50000, label: '50rb' }, { v: 100000, label: '100rb' },
]

const METHODS: PayMethod[] = ['Cash', 'Bank Transfer', 'QRIS', 'E-Wallet', 'Card', 'Credit']

interface CartLine { product: Product; variant?: Variant; qty: number }

export default function Pos() {
  const { session } = useDB()
  const [q, setQ] = useState('')
  const [cat, setCat] = useState('Semua')
  const prod = useCache<Product[]>(`pos-products:${session?.id}`, () => fetchAll<Product>((page) => apiListProducts({ active: true, page, limit: 200 })))
  const products = prod.data
  const setq = useCache<StoreSettings>(`settings:${session?.id}`, () => apiGetSettings())
  const settings = setq.data
  const [cart, setCart] = useState<CartLine[]>([])
  const [discount, setDiscount] = useState(0)
  const [voucherCode, setVoucherCode] = useState('')
  const [appliedVoucherCode, setAppliedVoucherCode] = useState<string | null>(null)
  const [previewRevision, setPreviewRevision] = useState(0)
  const [promoPreview, setPromoPreview] = useState<{
    cartKey: string
    voucherCode: string | null
    revision: number
    promo: PromoPreview | null
    error: string
    loading: boolean
    networkFailure: boolean
  } | null>(null)
  const [method, setMethod] = useState<PayMethod>('Cash')
  const [paid, setPaid] = useState('')
  const [exactCash, setExactCash] = useState(false)
  const [customerQuery, setCustomerQuery] = useState('')
  const [customerResults, setCustomerResults] = useState<Customer[]>([])
  const [customerLookupError, setCustomerLookupError] = useState('')
  const [customerLookupBusy, setCustomerLookupBusy] = useState(false)
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null)
  const [variantProduct, setVariantProduct] = useState<Product | null>(null)
  const [pointsRedeem, setPointsRedeem] = useState('')
  const [payOpen, setPayOpen] = useState(false)
  const [cartOpen, setCartOpen] = useState(false)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const [receipt, setReceipt] = useState<Trx | null>(null)
  const [offlineNotice, setOfflineNotice] = useState('')
  const [hasShift, setHasShift] = useState<boolean | null>(null)
  const checkoutAttempt = useRef<{ signature: string; clientTxId: string } | null>(null)

  // Gerbang shift (2026-10-16): backend menolak checkout tanpa shift aktif.
  // FE memeriksa lebih dulu agar kasir diarahkan membuka shift, bukan gagal
  // di tengah pembayaran. 404 = pasti belum ada shift; kegagalan jaringan
  // = status tak diketahui (null) dan backend tetap penegak terakhir.
  async function refreshShift(): Promise<boolean | null> {
    try {
      const { shift } = await apiGetCashierShift()
      const active = !!shift
      setHasShift(active)
      return active
    } catch (e) {
      if (e instanceof ApiError && e.status === 404) {
        setHasShift(false)
        return false
      }
      return hasShift
    }
  }
  useEffect(() => {
    void refreshShift()
    // Status bisa berubah di luar halaman ini (shift dibuka/ditutup lewat
    // popup Status di header) — segarkan saat jendela kembali fokus dan
    // berkala, agar pemberitahuan tidak basi.
    const onFocus = () => { void refreshShift() }
    window.addEventListener('pos:shift-changed', onFocus)
    window.addEventListener('focus', onFocus)
    document.addEventListener('visibilitychange', onFocus)
    const timer = setInterval(onFocus, 30000)
    return () => {
      window.removeEventListener('pos:shift-changed', onFocus)
      window.removeEventListener('focus', onFocus)
      document.removeEventListener('visibilitychange', onFocus)
      clearInterval(timer)
    }
  }, [])

  function load() {
    prod.reload()
  }

  const cats = useMemo(() => [...new Set((products ?? []).map((p) => p.category_id).filter((x): x is string => !!x))], [products])
  const catName = (id: string | null) => (products ?? []).find((p) => p.category_id === id)?.category_name ?? '—'

  const filtered = (products ?? []).filter((p) => {
    if (cat !== 'Semua' && p.category_id !== cat) return false
    const s2 = q.toLowerCase()
    return !s2 || p.name.toLowerCase().includes(s2) || p.sku.toLowerCase().includes(s2) || p.barcode.toLowerCase().includes(s2) || (p.variants ?? []).some((variant) => variant.sku.toLowerCase().includes(s2))
  })

  const lineKey = (product: Product, variant?: Variant) => `${product.id}:${variant?.id ?? ''}`
  const lineQuantity = (product: Product, variant?: Variant) => cart.find((line) => lineKey(line.product, line.variant) === lineKey(product, variant))?.qty ?? 0
  const availableStock = (product: Product, variant?: Variant) => (variant?.stock ?? product.stock) - lineQuantity(product, variant)
  const effStock = (product: Product) => product.variants?.length
    ? product.variants.reduce((sum, variant) => sum + Math.max(0, availableStock(product, variant)), 0)
    : availableStock(product)
  const unitPrice = (line: CartLine) => getUnitPrice(
    line.qty,
    (line.variant ?? line.product).sell_price,
    (line.variant ?? line.product).price_tiers ?? [],
  )

  const taxPct = settings?.taxEnabled ? settings.taxPct : 0
  const subtotal = cart.reduce((n, line) => n + unitPrice(line) * line.qty, 0)
  const cartKey = JSON.stringify(cart.map((line) => ({ productId: line.product.id, variantId: line.variant?.id, qty: line.qty })))
  const previewMatches = promoPreview?.cartKey === cartKey && promoPreview.voucherCode === appliedVoucherCode && promoPreview.revision === previewRevision
  const activePromo = previewMatches ? promoPreview?.promo ?? null : null
  const promoError = previewMatches ? promoPreview?.error ?? '' : ''
  const promoNetworkFailure = previewMatches && !!promoPreview?.networkFailure
  const previewLoading = cart.length > 0 && (!previewMatches || !!promoPreview?.loading)
  const checkoutOnline = navigator.onLine
  const promoBlocksCheckout = checkoutOnline && !promoNetworkFailure && (previewLoading || !!promoError)
  const effectiveDiscount = activePromo ? activePromo.promo_discount + activePromo.voucher_discount : discount
  const tax = Math.round((subtotal - effectiveDiscount) * (taxPct / 100))
  const total = subtotal - effectiveDiscount + tax
  const change = exactCash ? 0 : Number(paid) - total
  const loyaltyActive = !!settings?.loyaltyEnabled
  const redeemingPoints = method !== 'Credit' && loyaltyActive && !!selectedCustomer && Number(pointsRedeem) > 0

  // Nilai potongan promo berasal dari server preview; POS hanya memakai nilai itu
  // pada rincian yang sudah ada agar jumlah pembayaran mengikuti hasil preview.
  useEffect(() => {
    let dead = false
    const currentKey = JSON.stringify(cart.map((line) => ({ productId: line.product.id, variantId: line.variant?.id, qty: line.qty })))
    if (cart.length === 0) {
      setPromoPreview({ cartKey: currentKey, voucherCode: appliedVoucherCode, revision: previewRevision, promo: null, error: '', loading: false, networkFailure: false })
      return
    }
    setPromoPreview({ cartKey: currentKey, voucherCode: appliedVoucherCode, revision: previewRevision, promo: null, error: '', loading: true, networkFailure: false })
    apiPreviewPromo({
      items: cart.map((line) => ({ productId: line.product.id, ...(line.variant ? { variantId: line.variant.id } : {}), qty: line.qty })),
      ...(appliedVoucherCode ? { voucherCode: appliedVoucherCode } : {}),
    })
      .then((promo) => { if (!dead) setPromoPreview({ cartKey: currentKey, voucherCode: appliedVoucherCode, revision: previewRevision, promo, error: '', loading: false, networkFailure: false }) })
      .catch((e) => { if (!dead) setPromoPreview({ cartKey: currentKey, voucherCode: appliedVoucherCode, revision: previewRevision, promo: null, error: e instanceof Error ? e.message : 'Gagal memeriksa promo.', loading: false, networkFailure: isNetworkFailure(e, navigator.onLine) }) })
    return () => { dead = true }
  }, [cart, appliedVoucherCode, previewRevision])

  useEffect(() => {
    const search = customerQuery.trim()
    if (!payOpen || selectedCustomer || !search) {
      setCustomerResults([])
      setCustomerLookupError('')
      setCustomerLookupBusy(false)
      return
    }
    let dead = false
    const timer = setTimeout(() => {
      setCustomerLookupBusy(true)
      setCustomerLookupError('')
      apiListCustomers({ search, page: 1 })
        .then((res) => { if (!dead) setCustomerResults(res.items ?? []) })
        .catch((e) => { if (!dead) setCustomerLookupError(e instanceof Error ? e.message : 'Gagal mencari pelanggan.') })
        .finally(() => { if (!dead) setCustomerLookupBusy(false) })
    }, 250)
    return () => { dead = true; clearTimeout(timer) }
  }, [payOpen, customerQuery, selectedCustomer])

  // Nominal bayar selalu untuk total terkini: tiap keranjang, diskon,
  // atau hasil preview berubah, draft Jumlah dibayar dibuang.
  useEffect(() => { setPaid('') }, [cart, effectiveDiscount])

  const voucherDirty = voucherCode.trim() !== (appliedVoucherCode ?? '')

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
    const key = lineKey(p, variant)
    const stock = variant?.stock ?? p.stock
    setCart((c) => {
      const ex = c.find((l) => lineKey(l.product, l.variant) === key)
      if (ex) {
        if (navigator.onLine && ex.qty >= stock) { setErr('Stok tidak cukup.'); return c }
        return c.map((l) => (lineKey(l.product, l.variant) === key ? { ...l, qty: l.qty + 1 } : l))
      }
      if (navigator.onLine && stock < 1) { setErr('Stok tidak cukup.'); return c }
      return [...c, { product: p, ...(variant ? { variant } : {}), qty: 1 }]
    })
    setErr('')
  }

  function setQty(line: CartLine, qty: number) {
    const key = lineKey(line.product, line.variant)
    if (qty < 1) return setCart((c) => c.filter((item) => lineKey(item.product, item.variant) !== key))
    setCart((c) => c.map((l) => {
      if (lineKey(l.product, l.variant) !== key) return l
      if (navigator.onLine && qty > (l.variant?.stock ?? l.product.stock)) { setErr('Stok tidak cukup.'); return l }
      return { ...l, qty }
    }))
  }

  function clearCheckout(reloadProducts = false) {
    setCart([]); setDiscount(0); setVoucherCode(''); setAppliedVoucherCode(null); setPaid(''); setExactCash(false); setPayOpen(false)
    setSelectedCustomer(null); setCustomerQuery(''); setCustomerResults([]); setPointsRedeem('')
    if (reloadProducts) load()
  }

  async function checkout() {
    setErr('')
    setOfflineNotice('')
    if (cart.length === 0) return setErr('Keranjang kosong.')
    if ((await refreshShift()) === false) return setErr('Belum ada shift aktif. Buka shift dulu lewat tombol Status di kanan atas.')
    const online = navigator.onLine
    if (online) {
      if (method === 'Credit' && !isCreditPaymentValid(selectedCustomer?.id ?? null)) return setErr('Pilih pelanggan untuk pembayaran Kredit.')
      if (voucherDirty) return setErr('Terapkan kode voucher sebelum melanjutkan.')
      if (previewLoading && !promoNetworkFailure) return setErr('Pemeriksaan promo belum selesai.')
      if (promoError && !promoNetworkFailure) return setErr(promoError)
      if (method === 'Cash' && !isCashPaymentValid(paid, change, exactCash, redeemingPoints)) {
        return setErr(redeemingPoints ? 'Isi jumlah tunai yang dibayar; isi 0 jika poin menutup seluruh total.' : 'Jumlah bayar kurang dari total.')
      }
    }
    const checkoutVoucherCode = !online && voucherDirty ? voucherCode.trim() || null : appliedVoucherCode
    const checkoutBody: CheckoutBody = {
      items: cart.map((line) => ({ productId: line.product.id, ...(line.variant ? { variantId: line.variant.id } : {}), qty: line.qty })),
      discount,
      method,
      paid: method === 'Cash' ? (exactCash ? total : Number(paid)) : 0,
      ...(checkoutVoucherCode ? { voucherCode: checkoutVoucherCode } : {}),
      ...(selectedCustomer ? { customerId: selectedCustomer.id } : {}),
      ...(redeemingPoints ? { pointsRedeem: Number(pointsRedeem) } : {}),
    }
    const signature = JSON.stringify(checkoutBody)
    const attempt = checkoutAttempt.current?.signature === signature
      ? checkoutAttempt.current
      : { signature, clientTxId: createClientTxId() }
    checkoutAttempt.current = attempt
    const payload = { ...checkoutBody, clientTxId: attempt.clientTxId }
    const summary = {
      items: cart.map((line) => ({ name: `${line.product.name}${line.variant ? ` · ${line.variant.name}` : ''}`, qty: line.qty })),
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
      const trx = await apiCheckout(payload)
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

  // Badan keranjang bersama: dipakai aside desktop & sheet mobile (satu-satunya
  // perbedaan: kelas daftar item + sufiks id voucher agar tidak ganda di DOM).
  const cartBody = (idSuffix: string, listCls: string) => (
    <>
          {hasShift === false && (
            <div className="mb-3 rounded-lg border border-sunbeam/60 bg-sunbeam/10 px-3 py-2 text-[13px] leading-snug text-fg">
              Belum ada shift aktif. Buka shift dulu lewat tombol <b>Status</b> di kanan atas untuk bisa transaksi.
            </div>
          )}
          <div className={listCls}>
            {cart.length === 0 && <p className="py-6 text-center text-sm text-fog">Belum ada item. Klik produk untuk menambah.</p>}
            {cart.map((l) => (
              <div key={lineKey(l.product, l.variant)} className="rounded-lg border border-dove bg-paper p-2.5">
                <div className="flex items-start justify-between gap-2">
                  <div>
                    <p className="text-[13px] font-medium leading-snug">{l.product.name}{l.variant ? ` · ${l.variant.name}` : ''}</p>
                    {l.variant && <p className="mt-0.5 font-mono text-[11px] text-fog">SKU {l.variant.sku}</p>}
                  </div>
                  <button onClick={() => setCart((c) => c.filter((x) => lineKey(x.product, x.variant) !== lineKey(l.product, l.variant)))} aria-label={`Hapus ${l.product.name}${l.variant ? ` ${l.variant.name}` : ''}`} className="-mr-1 grid size-8 place-items-center rounded-full text-fog hover:text-ember">
                    <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M18 6 6 18M6 6l12 12" /></svg>
                  </button>
                </div>
                <div className="mt-2 flex items-center justify-between">
                  <div className="flex items-center gap-1.5">
                    <button onClick={() => setQty(l, l.qty - 1)} aria-label={`Kurangi ${l.product.name}`} className="grid h-9 w-9 place-items-center rounded-full border border-dove text-base hover:border-jet">−</button>
                    <span className="w-8 text-center font-mono text-sm tabular-nums">{l.qty}</span>
                    <button onClick={() => setQty(l, l.qty + 1)} aria-label={`Tambah ${l.product.name}`} className="grid h-9 w-9 place-items-center rounded-full border border-dove text-base hover:border-jet">+</button>
                  </div>
                  <span className="text-right">
                    <span className="block font-mono text-[11px] text-muted">{fmtRp(unitPrice(l))} / satuan</span>
                    <span className="font-mono text-[13px] tabular-nums">{fmtRp(unitPrice(l) * l.qty)}</span>
                  </span>
                </div>
              </div>
            ))}
          </div>

          <div className="mt-3 border-t border-dove pt-3">
            <label htmlFor={`voucher-code${idSuffix}`} className="mb-1.5 block text-[13px] font-medium text-steel">Kode voucher</label>
            <div className="flex gap-2">
              <input
                id={`voucher-code${idSuffix}`}
                value={voucherCode}
                onChange={(e) => setVoucherCode(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); applyVoucher() } }}
                placeholder="Masukkan kode"
                className="min-w-0 flex-1 rounded-md border border-border bg-paper px-3 py-2 text-sm focus:border-jet focus:outline-none"
              />
              <Button onClick={applyVoucher} disabled={navigator.onLine && previewLoading} className="px-4 py-2 text-sm">Terapkan</Button>
            </div>
            {previewLoading && <p className="mt-2 text-xs text-muted">Memeriksa promo…</p>}
            {promoError && !promoNetworkFailure && <AlertStrip className="mt-2">{promoError}</AlertStrip>}
          </div>

          <div className="mt-3 space-y-1.5 border-t border-dove pt-3 text-sm">
            <div className="flex justify-between"><span className="text-muted">Subtotal</span><span className="font-mono tabular-nums">{fmtRp(subtotal)}</span></div>
            {activePromo ? (
              <>
                {(activePromo.promo_discount > 0 || activePromo.voucher_discount === 0) && (
                  <div className="flex justify-between"><span className="min-w-0 truncate text-muted">Promo · {activePromo.promo_name || activePromo.voucher_name}</span><span className="shrink-0 font-mono tabular-nums">−{fmtRp(activePromo.promo_discount)}</span></div>
                )}
                {activePromo.voucher_discount > 0 && (
                  <div className="flex justify-between"><span className="min-w-0 truncate text-muted">Voucher · {activePromo.voucher_code}</span><span className="shrink-0 font-mono tabular-nums">−{fmtRp(activePromo.voucher_discount)}</span></div>
                )}
              </>
            ) : (
              <div className="flex items-center justify-between">
                <span className="text-muted">Diskon</span>
                <NumInput
                  value={discount}
                  onValue={(r) => setDiscount(Math.min(subtotal, Math.max(0, Number(r || 0))))}
                  className="w-24 rounded border border-dove bg-paper px-2 py-0.5 text-right font-mono text-[13px] tabular-nums focus:border-jet focus:outline-none"
                />
              </div>
            )}
            {taxPct > 0 && (
              <div className="flex justify-between"><span className="text-muted">Pajak ({taxPct}%)</span><span className="font-mono tabular-nums">{fmtRp(tax)}</span></div>
            )}
            <div className="flex justify-between text-base font-medium"><span>Total</span><span className="font-mono tabular-nums">{fmtRp(total)}</span></div>
          </div>
    </>
  )

  return (
    <>
      <PageHeader
        title="Kasir"
       
        crumb="Kasir"
      />

      {offlineNotice && <AlertStrip tone="success" className="mb-4">{offlineNotice}</AlertStrip>}


      <div className="grid gap-5 lg:grid-cols-[1fr_320px]">
        <section>
          <input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Cari produk, SKU, atau barcode…"
            aria-label="Cari produk"
            className="mb-3 w-full rounded-md border border-border bg-paper px-3.5 py-3 text-[15px] focus:border-jet focus:outline-none"
          />
          <div className="mb-4 flex flex-wrap gap-2">
            <button
              onClick={() => setCat('Semua')}
              className={`rounded-full border px-3 py-1 text-xs ${cat === 'Semua' ? 'border-jet bg-jet text-paper' : 'border-dove text-muted hover:border-jet'}`}
            >
              Semua
            </button>
            {cats.map((id) => (
              <button
                key={id}
                onClick={() => setCat(id)}
                className={`rounded-full border px-3 py-1 text-xs ${cat === id ? 'border-jet bg-jet text-paper' : 'border-dove text-muted hover:border-jet'}`}
              >
                {catName(id)}
              </button>
            ))}
          </div>
          {err && <AlertStrip className="mb-3">{err}</AlertStrip>}
          {!products ? (
            <div aria-busy="true" aria-label="Memuat katalog">
              <div className="mb-4 flex flex-wrap gap-2" aria-hidden="true">
                {[0, 1, 2, 3].map((i) => (
                  <Skeleton key={i} className="h-6 w-16 rounded-full" />
                ))}
              </div>
              <div className="grid max-h-[62vh] grid-cols-2 gap-2.5 overflow-y-auto pr-1 md:grid-cols-3 xl:grid-cols-4" aria-hidden="true">
                {Array.from({ length: 8 }).map((_, i) => (
                  <div key={i} className="space-y-2 rounded-xl border border-dove p-3.5">
                    <Skeleton className="h-4 w-3/4" />
                    <Skeleton className="h-3 w-1/3" />
                    <Skeleton className="h-4 w-1/2" />
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="grid max-h-[62vh] grid-cols-2 gap-2.5 overflow-y-auto pr-1 md:grid-cols-3 xl:grid-cols-4">
              {filtered.map((p) => {
                const price = p.variants?.length
                  ? Math.min(...p.variants.map((variant) => getUnitPrice(1, variant.sell_price, variant.price_tiers ?? [])))
                  : p.sell_price
                const remaining = effStock(p)
                const stockBlocked = navigator.onLine && remaining <= 0
                return (
                  <button
                    key={p.id}
                    onClick={() => chooseProduct(p)}
                    disabled={stockBlocked}
                    className={`rounded-xl border p-3.5 text-left transition ${stockBlocked ? 'cursor-not-allowed border-dove opacity-40' : 'border-dove hover:border-jet'}`}
                  >
                    <p className="text-[13px] font-medium leading-snug">{p.name}</p>
                    <p className="mt-1 font-mono text-[11px] text-fog">{remaining} stok{p.variants?.length ? ` · ${p.variants.length} varian` : ''}</p>
                    <p className="mt-1 font-mono text-sm font-medium tabular-nums">{p.variants?.length ? 'Mulai ' : ''}{fmtRp(price)}</p>
                  </button>
                )
              })}
              {filtered.length === 0 && <p className="col-span-full py-10 text-center text-sm text-fog">Tidak ada produk ditemukan.</p>}
            </div>
          )}
        </section>

        <aside className="hidden flex-col rounded-2xl bg-cream p-4 lg:flex">
          <h2 className="font-mono text-xs uppercase tracking-wider text-fog">Keranjang</h2>
          {cartBody('', 'mt-3 flex-1 space-y-2 overflow-y-auto')}
          <Button className="mt-4 w-full" disabled={cart.length === 0 || (checkoutOnline && (promoBlocksCheckout || voucherDirty))} onClick={() => { setPayOpen(true); setErr('') }}>Bayar · {fmtRp(total)}</Button>
        </aside>
      </div>

      {cart.length > 0 && (
        <>
          <div className="h-20 lg:hidden" aria-hidden="true" />
          <div className="fixed inset-x-0 bottom-0 z-40 flex items-center gap-3 border-t border-dove bg-paper px-4 py-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] shadow-[0_-4px_16px_rgba(0,0,0,0.06)] lg:hidden">
            <div className="min-w-0 flex-1">
              <p className="text-[11px] leading-tight text-fog">{cart.reduce((n, l) => n + l.qty, 0)} item</p>
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

      <Modal open={!!variantProduct} title={`Pilih Varian · ${variantProduct?.name ?? ''}`} onClose={() => setVariantProduct(null)}>
        <div className="space-y-2">
          {variantProduct?.variants?.map((variant) => {
            const remaining = availableStock(variantProduct, variant)
            const price = getUnitPrice(1, variant.sell_price, variant.price_tiers ?? [])
            const stockBlocked = navigator.onLine && remaining <= 0
            return (
              <button
                key={variant.id}
                type="button"
                disabled={stockBlocked}
                onClick={() => { add(variantProduct, variant); setVariantProduct(null) }}
                className="flex w-full items-center justify-between gap-3 rounded-lg border border-dove bg-paper px-3.5 py-3 text-left transition hover:border-jet disabled:cursor-not-allowed disabled:opacity-40"
              >
                <span className="min-w-0">
                  <span className="block truncate text-sm font-medium text-fg">{variant.name}</span>
                  <span className="block truncate font-mono text-xs text-muted">{variant.sku} · stok {remaining}</span>
                </span>
                <span className="shrink-0 font-mono text-sm tabular-nums">{fmtRp(price)}</span>
              </button>
            )
          })}
          {variantProduct?.variants?.length === 0 && <p className="text-sm text-muted">Belum ada varian untuk produk ini.</p>}
        </div>
      </Modal>

      <Modal open={cartOpen} title="Keranjang" onClose={() => setCartOpen(false)}>
        {cartBody('-sheet', 'space-y-2')}
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

      <Modal open={payOpen} title="Pembayaran" onClose={() => setPayOpen(false)}>
        <div className="space-y-4">
          {err && <AlertStrip>{err}</AlertStrip>}
          {promoError && !promoNetworkFailure && <AlertStrip>{promoError}</AlertStrip>}
          <div className="space-y-2">
            <label htmlFor="checkout-customer" className="block text-[13px] font-medium text-steel">Pelanggan {method === 'Credit' ? '(wajib)' : '(opsional)'}</label>
            {selectedCustomer ? (
              <div className="flex items-center justify-between gap-3 rounded-lg border border-dove bg-paper px-3.5 py-2.5">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-fg">{selectedCustomer.name}</p>
                  <p className="truncate text-xs text-muted">{selectedCustomer.phone || 'Tanpa nomor telepon'}</p>
                </div>
                <button
                  type="button"
                  onClick={() => { setSelectedCustomer(null); setCustomerQuery(''); setCustomerResults([]); setCustomerLookupBusy(false); setPointsRedeem('') }}
                  className="shrink-0 rounded-md px-2 py-1 text-xs text-steel hover:bg-surface hover:text-fg"
                >
                  Kosongkan
                </button>
              </div>
            ) : (
              <>
                <input
                  id="checkout-customer"
                  value={customerQuery}
                  onChange={(e) => { setCustomerQuery(e.target.value); setCustomerResults([]); setCustomerLookupError(''); setCustomerLookupBusy(!!e.target.value.trim()) }}
                  placeholder="Cari nama atau telepon…"
                  autoComplete="off"
                  className="w-full rounded-md border border-border bg-paper px-3.5 py-2.5 text-[15px] focus:border-jet focus:outline-none"
                />
                {customerLookupBusy && <p className="text-xs text-muted" role="status">Mencari pelanggan…</p>}
                {customerLookupError && <p className="text-xs text-ember" role="alert">{customerLookupError}</p>}
                {customerResults.length > 0 && (
                  <div className="max-h-40 overflow-y-auto rounded-lg border border-dove bg-paper">
                    {customerResults.map((customer) => (
                      <button
                        key={customer.id}
                        type="button"
                        onClick={() => { setSelectedCustomer(customer); setCustomerQuery(''); setCustomerResults([]); setCustomerLookupBusy(false); setPointsRedeem('') }}
                        className="flex w-full items-center justify-between gap-3 border-b border-dove px-3 py-2 text-left last:border-b-0 hover:bg-surface"
                      >
                        <span className="min-w-0">
                          <span className="block truncate text-sm font-medium text-fg">{customer.name}</span>
                          <span className="block truncate text-xs text-muted">{customer.phone || 'Tanpa nomor telepon'}</span>
                        </span>
                        <span className="shrink-0 font-mono text-xs text-muted">{customer.points.toLocaleString('id-ID')} poin</span>
                      </button>
                    ))}
                  </div>
                )}
                {customerQuery.trim() && !customerLookupBusy && !customerLookupError && customerResults.length === 0 && (
                  <p className="text-xs text-muted">Pelanggan tidak ditemukan.</p>
                )}
              </>
            )}
            {method === 'Credit' && <p className="text-xs text-muted" role="status">Pilih pelanggan untuk mencatat piutang. Jumlah dibayar untuk Kredit adalah Rp0.</p>}
            {method !== 'Credit' && loyaltyActive && selectedCustomer && (
              <div className="space-y-2 rounded-lg bg-surface px-3.5 py-3">
                <p className="text-sm text-fg">Saldo poin: <span className="font-mono font-medium tabular-nums">{selectedCustomer.points.toLocaleString('id-ID')} poin</span></p>
                <label htmlFor="points-redeem" className="block text-[13px] font-medium text-steel">Tukar poin</label>
                <input
                  id="points-redeem"
                  type="number"
                  min="0"
                  max={selectedCustomer.points}
                  step="1"
                  value={pointsRedeem}
                  disabled={selectedCustomer.points < 1}
                  onChange={(e) => {
                    setPointsRedeem(e.target.value)
                    if (Number(e.target.value) > 0) { setExactCash(false); setPaid('') }
                  }}
                  placeholder={selectedCustomer.points > 0 ? 'Jumlah poin' : 'Belum ada poin'}
                  className="w-full rounded-md border border-border bg-paper px-3.5 py-2.5 text-[15px] focus:border-jet focus:outline-none disabled:opacity-50"
                />
                {Number(pointsRedeem) > 0 && (
                  <p className="text-xs text-muted">Nilai tukar {Number(pointsRedeem).toLocaleString('id-ID')} poin × {fmtRp(settings?.loyaltyPointValue ?? 100)} = {fmtRp(Number(pointsRedeem) * (settings?.loyaltyPointValue ?? 100))}.</p>
                )}
                <p className="text-xs text-muted">Nilai akhir dan kembalian setelah tukar poin ditampilkan dari hasil transaksi.</p>
              </div>
            )}
          </div>
          <div>
            <p className="mb-2 text-[13px] font-medium text-steel">Metode pembayaran</p>
            <div className="flex flex-wrap gap-2">
              {METHODS.map((m) => (
                <button
                  key={m}
                  onClick={() => {
                    setMethod(m)
                    if (m === 'Credit') { setPaid('0'); setExactCash(false); setPointsRedeem('') }
                    else if (method === 'Credit') setPaid('')
                  }}
                  className={`rounded-full border px-3.5 py-1.5 text-xs ${method === m ? 'border-jet bg-jet text-paper' : 'border-dove text-muted hover:border-jet'}`}
                >
                  {m === 'Credit' ? 'Kredit' : m}
                </button>
              ))}
            </div>
          </div>
          {method === 'Cash' ? (
            <div className="space-y-3">
              <CheckRow
                checked={exactCash}
                disabled={redeemingPoints}
                onChange={(v) => { setExactCash(v); if (v) setPaid('') }}
                className="text-sm font-normal text-fg"
              >
                Uang Pas{redeemingPoints ? ' (nonaktif saat tukar poin)' : ''}
              </CheckRow>
              {!exactCash && (
                <>
                  <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
                    Jumlah dibayar
                    <NumInput
                      value={paid}
                      onValue={setPaid}
                      placeholder="0"
                      autoFocus
                      hint={redeemingPoints ? 'Isi 0 bila poin menutup seluruh total; sisa tunai divalidasi server.' : undefined}
                      className="rounded-md border border-border bg-paper px-3.5 py-2.5 text-[15px] focus:border-jet focus:outline-none"
                    />
                  </label>
                  <div>
                    <p className="mb-2 text-[13px] font-medium text-steel">Tambah cepat</p>
                    <div className="grid grid-cols-4 gap-2">
                      {CASH_QUICK.map((q) => (
                        <button
                          key={q.v} type="button"
                          onClick={() => setPaid(String((Number(paid) || 0) + q.v))}
                          className="rounded-lg border border-dove bg-paper px-2 py-2 font-mono text-xs tabular-nums text-fg transition outline-none hover:border-jet hover:text-jet focus-visible:ring-2 focus-visible:ring-ring active:translate-y-px"
                        >
                          +{q.label}
                        </button>
                      ))}
                    </div>
                  </div>
                </>
              )}
            </div>
          ) : (
            <p className="text-sm text-muted">
              {method === 'Credit'
                ? `Total ${fmtRp(total)} akan dicatat sebagai piutang dengan jatuh tempo ${settings?.creditDueDays ?? 14} hari.`
                : redeemingPoints
                ? `Total sebelum tukar poin: ${fmtRp(total)}. Nominal akhir ditentukan setelah transaksi.`
                : `Total ${fmtRp(total)} akan dicatat sebagai pembayaran ${method}.`}
            </p>
          )}
          {method === 'Cash' && !redeemingPoints && (exactCash ? change === 0 : paid) && change >= 0 && (
            <div className="flex justify-between rounded-lg bg-surface px-3.5 py-3 text-sm">
              <span className="text-muted">Kembalian</span>
              <span className="font-mono font-medium tabular-nums">{fmtRp(change)}</span>
            </div>
          )}
          <div className="space-y-1.5 text-sm">
            {activePromo ? (
              <div className="flex justify-between"><span className="text-muted">Promo · {activePromo.promo_name}</span><span className="font-mono tabular-nums">−{fmtRp(activePromo.promo_discount)}</span></div>
            ) : (
              <div className="flex justify-between"><span className="text-muted">Diskon</span><span className="font-mono tabular-nums">&minus;{fmtRp(discount)}</span></div>
            )}
            <div className="flex justify-between">
              <span className="text-muted">Pajak{taxPct > 0 ? ` (${taxPct}%)` : ''}</span>
              <span className="font-mono tabular-nums">+{fmtRp(tax)}</span>
            </div>
          </div>
          <div className="flex justify-between border-t border-dove pt-3 text-[15px]">
            <span>Total</span>
            <span className="font-mono font-medium tabular-nums">{fmtRp(total)}</span>
          </div>
          <ModalActions>
            <Button className="w-full" onClick={checkout} disabled={busy || (method === 'Credit' && !isCreditPaymentValid(selectedCustomer?.id ?? null))}>{busy ? 'Memproses…' : 'Selesaikan Transaksi'}</Button>
          </ModalActions>
        </div>
      </Modal>

      <Modal open={!!receipt} title="Transaksi berhasil" onClose={() => setReceipt(null)}>
        {receipt && <Receipt trx={receipt} settings={settings} onClose={() => setReceipt(null)} />}
      </Modal>
    </>
  )
}
