// Produk — Operate surface. Katalog + filter kategori + CRUD + Import/Export.
// Token font/warna milik sistem (tidak ada token baru di file ini).
import { useEffect, useMemo, useRef, useState } from 'react'
import { useSearchParams } from 'react-router'
import { FolderPlus, Pencil, Plus, Power, Trash2 } from 'lucide-react'
import { apiCreateCategory, apiCreateProduct, apiCreateProductPriceTier, apiCreateVariant, apiCreateVariantPriceTier, apiDeleteCategory, apiDeleteProduct, apiDeleteProductPriceTier, apiDeleteVariant, apiDeleteVariantPriceTier, apiExportProducts, apiGetProduct, apiGetSettings, apiImportProducts, apiListCategories, apiListProducts, apiSetProductActive, apiUpdateProduct, apiUpdateProductPriceTier, apiUpdateVariant, apiUpdateVariantPriceTier, fetchAll, type Category, type PriceTier, type Product, type ProductImportResult, type StoreSettings, type Variant } from '../lib/api'
import { selectCsvExportHandler } from '../lib/export-role'
import { PageHeader } from '../lib/PageHeader'
import { useCache } from '../lib/cache'
import { downloadBlob, exportCSV, exportExcel, fmtRp, parseImportFile, useDB } from '../lib/store'
import { ModalActions, AlertStrip, Button, CardGrid, DataCard, Empty, ExportMenu, FilterChips, FilterField, FilterPanel, ImportMenu, Modal, Pager, Pill, RowAction, SearchInput, Select, SkeletonRows, TableWrap, Td, Th } from '../lib/ui'
import { Skeleton } from '@/components/ui/skeleton'
import { FormProduk } from './produk/FormProduk'
import { ImportProdukModal } from './produk/ImportProdukModal'
import { KategoriModal } from './produk/KategoriModal'
import { StockCell, kondisiOf } from './produk/StockCell'
import { emptyDraft, hasDuplicateTierQty, productDraft, tierDraft, variantDraft, type Draft, type TierDraft, type VariantDraft } from './produk/draft'

const NONE = '__none__'

const STATUS_OPTS = [{ value: '', label: 'Semua status' }, { value: 'aktif', label: 'Aktif' }, { value: 'nonaktif', label: 'Nonaktif' }]
const KONDISI_OPTS = [{ value: '', label: 'Semua kondisi' }, { value: 'aman', label: 'Aman' }, { value: 'menipis', label: 'Menipis' }, { value: 'habis', label: 'Habis' }]


// Bentuk chip filter kategori untuk tablet (768-1279). Di >=1280 rail pakai
// kelas aslinya (w-full, px-3 py-2 text-sm) dan tidak tersentuh helper ini.

export default function Produk() {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const [q, setQ] = useState('')
  const [catFilter, setCatFilter] = useState('')
  const [statusF, setStatusF] = useState('')
  const [kondisiF, setKondisiF] = useState('')
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const prod = useCache<Product[]>(`products:${who}:${q.trim()}`, () => fetchAll<Product>((page) => apiListProducts({ q: q.trim() || undefined, page, limit: 200 })), 'Gagal memuat produk.')
  const products = prod.data
  const settings = useCache<StoreSettings>(`settings:${session?.id}`, () => apiGetSettings()).data
  const catq = useCache<Category[]>(`cats:${who}`, () => apiListCategories())
  const cats = catq.data ?? []
  const [err, setErr] = useState('')
  const [editing, setEditing] = useState<Draft | null>(null)
  const [saveBusy, setSaveBusy] = useState(false)
  const [deleteFor, setDeleteFor] = useState<Product | null>(null)
  const [catName, setCatName] = useState('')
  const [catOpen, setCatOpen] = useState(false)
  const [catBusy, setCatBusy] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [importBusy, setImportBusy] = useState(false)
  const [importResult, setImportResult] = useState<ProductImportResult | null>(null)
  const [importError, setImportError] = useState('')

  const counts = useMemo(() => {
    const m = new Map<string, number>()
    let none = 0
    for (const p of products ?? []) {
      if (p.category_id) m.set(p.category_id, (m.get(p.category_id) ?? 0) + 1)
      else none++
    }
    return { m, none }
  }, [products])

  const filtered = useMemo(() => {
    if (!products) return null
    return products.filter((p) => {
      if (catFilter === NONE ? p.category_id : catFilter && p.category_id !== catFilter) return false
      if (statusF === 'aktif' && !p.active) return false
      if (statusF === 'nonaktif' && p.active) return false
      if (kondisiF && kondisiOf(p) !== kondisiF) return false
      return true
    })
  }, [products, catFilter, statusF, kondisiF])

  useEffect(() => { setPage(0) }, [q, catFilter, statusF, kondisiF])
  const totalPages = Math.max(1, Math.ceil((filtered?.length ?? 0) / pageSize))
  const safePage = Math.min(page, totalPages - 1)
  const pageItems = filtered?.slice(safePage * pageSize, safePage * pageSize + pageSize)

  // Deep link dari halaman Stok: ?edit=<id> membuka editor produk itu sekali,
  // lalu parameter dibersihkan agar refresh tidak membukanya lagi.
  const [searchParams, setSearchParams] = useSearchParams()
  const editLinkOpened = useRef(false)
  useEffect(() => {
    const id = searchParams.get('edit')
    if (!id || editLinkOpened.current || !products) return
    const target = products.find((p) => String(p.id) === id)
    if (target) {
      editLinkOpened.current = true
      editProduct(target)
      setSearchParams({}, { replace: true })
    }
  }, [searchParams, products])

  const activeCats = cats.filter((c) => c.active)

  function loadProducts() {
    prod.reload()
  }

  function editProduct(product?: Product) {
    setErr('')
    setEditing(product ? productDraft(product) : {
      ...emptyDraft, variants: [], priceTiers: [],
      minStock: settings?.lowStockDefault == null ? '' : String(settings.lowStockDefault),
    })
  }

  async function save(d: Draft): Promise<Draft | null> {
    setErr('')
    const sell = Number(d.sellPrice)
    if (!d.name.trim() || !d.sku.trim() || !Number.isFinite(sell)) return null
    const tiers = [...d.priceTiers, ...d.variants.flatMap((v) => v.priceTiers)]
    if (tiers.some((t) => !Number.isInteger(Number(t.minQty)) || Number(t.minQty) < 1 || t.sellPrice === '' || !Number.isFinite(Number(t.sellPrice)))) {
      setErr('Setiap harga bertingkat harus memiliki min. jumlah dan harga yang valid.')
      return null
    }
    if (hasDuplicateTierQty(d.priceTiers) || d.variants.some((variant) => hasDuplicateTierQty(variant.priceTiers))) {
      setErr('Min. jumlah harga bertingkat tidak boleh duplikat dalam produk atau varian yang sama.')
      return null
    }
    if (d.variants.some((v) => !v.name.trim() || !v.sku.trim() || v.sellPrice === '' || !Number.isFinite(Number(v.sellPrice)))) {
      setErr('Nama, SKU, dan harga jual setiap varian wajib diisi.')
      return null
    }
    let working: Draft = {
      ...d,
      priceTiers: d.priceTiers.map((tier) => ({ ...tier })),
      variants: d.variants.map((variant) => ({ ...variant, priceTiers: variant.priceTiers.map((tier) => ({ ...tier })) })),
    }
    const isNew = !d.id
    const productMinStock = d.minStock !== '' ? Number(d.minStock) : isNew ? settings?.lowStockDefault : undefined
    const body = {
      name: d.name.trim(), sku: d.sku.trim(), barcode: d.barcode.trim(),
      categoryId: d.categoryId || null, buyPrice: Number(d.buyPrice) || 0,
      sellPrice: sell, unit: d.unit || 'pcs',
      ...(productMinStock != null ? { minStock: productMinStock } : {}),
    }
    let savedProduct: Product | undefined
    setSaveBusy(true)
    try {
      savedProduct = d.id
        ? await apiUpdateProduct(d.id, body)
        : await apiCreateProduct({ ...body, stock: d.variants.length ? 0 : Number(d.stock) || 0 })
      working = {
        ...working,
        id: savedProduct.id,
        minStock: working.minStock || (savedProduct.min_stock == null ? '' : String(savedProduct.min_stock)),
      }
      const latest = await apiGetProduct(savedProduct.id)
      const productTiers = await syncPriceTiers(
        savedProduct.id, undefined, latest.price_tiers ?? [], new Set(working.initialPriceTierIds), working.priceTiers,
        (priceTiers) => {
          working = {
            ...working, priceTiers,
            initialPriceTierIds: [...new Set([...working.initialPriceTierIds, ...priceTiers.flatMap((tier) => tier.id === undefined ? [] : [tier.id])])],
          }
        },
      )
      working = {
        ...working, priceTiers: productTiers,
        initialPriceTierIds: [...new Set([...working.initialPriceTierIds, ...productTiers.flatMap((tier) => tier.id === undefined ? [] : [tier.id])])],
      }

      const initialVariantIds = new Set(working.initialVariantIds)
      const desiredVariantIds = new Set(working.variants.flatMap((variant) => variant.id === undefined ? [] : [variant.id]))
      const newlySeen = (latest.variants ?? []).filter((variant) => !initialVariantIds.has(variant.id) && !desiredVariantIds.has(variant.id))
      if (newlySeen.length) {
        working = {
          ...working,
          variants: [...working.variants, ...newlySeen.map(variantDraft)],
          initialVariantIds: [...new Set([...working.initialVariantIds, ...newlySeen.map((variant) => variant.id)])],
        }
      }

      for (const old of latest.variants ?? []) {
        if (!initialVariantIds.has(old.id) || desiredVariantIds.has(old.id)) continue
        for (const tier of old.price_tiers ?? []) await apiDeleteVariantPriceTier(savedProduct.id, old.id, tier.id)
        await apiDeleteVariant(savedProduct.id, old.id)
      }
      for (let index = 0; index < working.variants.length; index++) {
        const variant = working.variants[index]
        let old = latest.variants?.find((item) => item.id === variant.id)
        if (variant.id !== undefined && !old) throw new Error(`Varian ${variant.name} sudah berubah di server. Muat ulang produk sebelum menyimpan.`)
        const minStock = variant.minStock !== '' ? Number(variant.minStock) : old ? undefined : settings?.lowStockDefault
        const makeVariantBody = (current?: Variant) => ({
          name: variant.name.trim(), sku: variant.sku.trim(),
          buyPrice: Number(variant.buyPrice) || 0, sellPrice: Number(variant.sellPrice),
          stock: current && !variant.stockEdited ? current.stock : Number(variant.stock) || 0,
          ...(minStock != null ? { minStock } : {}),
        })
        let variantBody = makeVariantBody(old)
        let changed = !old || old.name !== variantBody.name || old.sku !== variantBody.sku
          || old.buy_price !== variantBody.buyPrice || old.sell_price !== variantBody.sellPrice
          || old.stock !== variantBody.stock || (minStock != null && old.min_stock !== minStock)
        if (old && changed) {
          const current = await apiGetProduct(savedProduct.id)
          old = current.variants?.find((item) => item.id === variant.id)
          if (!old) throw new Error(`Varian ${variant.name} sudah berubah di server. Muat ulang produk sebelum menyimpan.`)
          variantBody = makeVariantBody(old)
          changed = old.name !== variantBody.name || old.sku !== variantBody.sku
            || old.buy_price !== variantBody.buyPrice || old.sell_price !== variantBody.sellPrice
            || old.stock !== variantBody.stock || (minStock != null && old.min_stock !== minStock)
        }
        const savedVariant = !old
          ? await apiCreateVariant(savedProduct.id, variantBody)
          : changed ? await apiUpdateVariant(savedProduct.id, old.id, variantBody) : old
        const savedVariantDraft: VariantDraft = {
          ...variant, id: savedVariant.id, stock: String(savedVariant.stock),
          stockEdited: false,
          minStock: variant.minStock || String(savedVariant.min_stock ?? ''),
        }
        working = {
          ...working,
          initialVariantIds: [...new Set([...working.initialVariantIds, savedVariant.id])],
          variants: working.variants.map((row, ix) => ix === index ? savedVariantDraft : row),
        }
        const latestTiers = old?.price_tiers ?? []
        const updatedPriceTiers = await syncPriceTiers(
          savedProduct.id, savedVariant.id, latestTiers, new Set(variant.initialPriceTierIds), variant.priceTiers,
          (priceTiers) => {
            working = {
              ...working,
              variants: working.variants.map((row, ix) => ix === index ? {
                ...row, priceTiers,
                initialPriceTierIds: [...new Set([...row.initialPriceTierIds, ...priceTiers.flatMap((tier) => tier.id === undefined ? [] : [tier.id])])],
              } : row),
            }
          },
        )
        working = { ...working, variants: working.variants.map((row, ix) => ix === index ? { ...row, priceTiers: updatedPriceTiers, initialPriceTierIds: updatedPriceTiers.flatMap((tier) => tier.id === undefined ? [] : [tier.id]) } : row) }
      }
      setEditing(null)
      loadProducts()
      return null
    } catch (e) {
      if (savedProduct) {
        loadProducts()
        setEditing(working)
        setErr(`Sebagian data tersimpan. Draft tetap terbuka; periksa lalu simpan kembali. ${e instanceof Error ? e.message : ''}`.trim())
        return working
      } else {
        setErr(e instanceof Error ? e.message : 'Gagal menyimpan produk.')
        return null
      }
    } finally {
      setSaveBusy(false)
    }
  }

  async function syncPriceTiers(
    productId: string,
    variantId: number | undefined,
    current: PriceTier[],
    visibleIds: Set<number>,
    desired: TierDraft[],
    onChange: (tiers: TierDraft[]) => void,
  ): Promise<TierDraft[]> {
    let tiers = desired.map((tier) => ({ ...tier }))
    const desiredIds = new Set(tiers.flatMap((tier) => tier.id === undefined ? [] : [tier.id]))
    const unseen = current.filter((tier) => !visibleIds.has(tier.id) && !desiredIds.has(tier.id))
    tiers = [...tiers, ...unseen.map(tierDraft)]
    const byId = (id: number) => tiers.find((tier) => tier.id === id)

    for (const old of current) {
      if (!visibleIds.has(old.id) && !desiredIds.has(old.id)) continue
      const target = byId(old.id)
      if (target && Number(target.minQty) === old.min_qty) continue
      if (variantId === undefined) await apiDeleteProductPriceTier(productId, old.id)
      else await apiDeleteVariantPriceTier(productId, variantId, old.id)
      if (target) tiers = tiers.map((tier) => tier === target ? { ...tier, id: undefined } : tier)
      onChange(tiers)
    }
    for (let index = 0; index < tiers.length; index++) {
      let tier = tiers[index]
      const old = tier.id === undefined ? undefined : current.find((item) => item.id === tier.id)
      if (old) {
        const input = { minQty: Number(tier.minQty), sellPrice: Number(tier.sellPrice) }
        if (old.sell_price !== input.sellPrice) {
          if (variantId === undefined) await apiUpdateProductPriceTier(productId, old.id, input)
          else await apiUpdateVariantPriceTier(productId, variantId, old.id, input)
          onChange(tiers)
        }
        continue
      }
      if (tier.id !== undefined) {
        tier = { ...tier, id: undefined }
        tiers = tiers.map((row, ix) => ix === index ? tier : row)
        onChange(tiers)
      }
      const input = { minQty: Number(tier.minQty), sellPrice: Number(tier.sellPrice) }
      const created = variantId === undefined
        ? await apiCreateProductPriceTier(productId, input)
        : await apiCreateVariantPriceTier(productId, variantId, input)
      tier = { ...tier, id: created.id }
      tiers = tiers.map((row, ix) => ix === index ? tier : row)
      visibleIds.add(created.id)
      onChange(tiers)
    }
    return tiers
  }

  async function toggleActive(p: Product) {
    try {
      await apiSetProductActive(p.id, !p.active)
      loadProducts()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal mengubah status.')
    }
  }

  async function removeProduct() {
    if (!deleteFor) return
    try {
      await apiDeleteProduct(deleteFor.id)
      setDeleteFor(null)
      loadProducts()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal menghapus produk.')
    }
  }

  async function addCat() {
    const n = catName.trim()
    if (!n) return
    try {
      await apiCreateCategory(n)
      catq.reload()
      setCatName('')
      setCatOpen(false)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal menambah kategori.')
    }
  }

  async function removeCategory(c: Category) {
    if (catBusy) return
    const n = counts.m.get(c.id) ?? 0
    if (!confirm(`Hapus kategori "${c.name}"?${n > 0 ? ` ${n} produk di dalamnya jadi Tanpa Kategori.` : ''}`)) return
    setErr(''); setCatBusy(true)
    try {
      // Pindah semua produk ke Tanpa Kategori dulu supaya DELETE
      // menghapus permanen (backend soft-delete bila masih dipakai).
      const all = await fetchAll<Product>((page) => apiListProducts({ page, limit: 200 }))
      let fail = 0
      for (const p of all.filter((x) => x.category_id === c.id)) {
        try {
          await apiUpdateProduct(p.id, {
            name: p.name, sku: p.sku, barcode: p.barcode, categoryId: null,
            buyPrice: p.buy_price, sellPrice: p.sell_price, unit: p.unit, minStock: p.min_stock,
          })
        } catch {
          fail++
        }
      }
      if (fail > 0) {
        setErr(`${fail} produk gagal dipindah. Kategori tidak dihapus.`)
        return
      }
      await apiDeleteCategory(c.id)
      if (c.id === catFilter) setCatFilter('')
      catq.reload()
      loadProducts()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal menghapus kategori.')
    } finally {
      setCatBusy(false)
    }
  }

  function productRows(all: Product[]): string[][] {
    return [
      ['nama', 'sku', 'barcode', 'kategori', 'harga_beli', 'harga_jual', 'stok', 'unit', 'aktif'],
      ...all.map((p) => [
        p.name, p.sku, p.barcode, p.category_name ?? '', String(p.buy_price),
        String(p.sell_price), String(p.stock), p.unit, p.active ? '1' : '0',
      ]),
    ]
  }

  async function exportListCSV() {
    const all = await fetchAll<Product>((page) => apiListProducts({ page, limit: 200 }))
    exportCSV('produk.csv', productRows(all))
  }

  async function exportServerCSV() {
    setErr('')
    try {
      const file = await apiExportProducts()
      const today = new Date().toISOString().slice(0, 10)
      await downloadBlob(file.filename ?? `produk-${today}.csv`, file.blob)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal mengekspor produk.')
    }
  }

  async function exportListExcel() {
    const all = await fetchAll<Product>((page) => apiListProducts({ page, limit: 200 }))
    await exportExcel('produk.xlsx', [{
      name: 'Produk',
      title: 'Daftar Produk',
      subtitle: `${all.length} produk`,
      columns: [
        { header: 'Nama' },
        { header: 'SKU' },
        { header: 'Barcode' },
        { header: 'Kategori' },
        { header: 'Harga Beli (Rp)', money: true },
        { header: 'Harga Jual (Rp)', money: true },
        { header: 'Stok', align: 'right' },
        { header: 'Satuan', align: 'center' },
        { header: 'Status', align: 'center' },
      ],
      rows: all.map((p) => [
        p.name, p.sku, p.barcode, p.category_name ?? '', p.buy_price,
        p.sell_price, p.stock, p.unit, p.active ? 'Aktif' : 'Nonaktif',
      ]),
    }])
  }

  const fileRef = useRef<HTMLInputElement>(null)
  function pickImport(kind: 'xlsx' | 'csv') {
    if (kind === 'csv') {
      setImportError('')
      setImportResult(null)
      setImportOpen(true)
      return
    }
    if (fileRef.current) fileRef.current.accept = '.xlsx,.xls'
    fileRef.current?.click()
  }

  function pickCsvFile() {
    if (fileRef.current) fileRef.current.accept = '.csv'
    fileRef.current?.click()
  }

  async function onImportFile(f: File) {
    setImportOpen(true)
    setImportError('')
    setImportResult(null)
    try {
      let file = f
      if (/\.(xlsx|xls)$/i.test(f.name)) {
        const lines = await parseImportFile(f)
        const csvText = lines.map((row) => row.map((c) => (/[",\n]/.test(c) ? '"' + c.replace(/"/g, '""') + '"' : c)).join(',')).join('\n')
        file = new File([csvText], 'import.csv', { type: 'text/csv' })
      }
      await uploadImport(file)
    } catch (e) {
      setImportError(e instanceof Error ? e.message : 'Gagal membaca file.')
    }
  }

  async function uploadImport(file: File) {
    setImportBusy(true)
    setImportError('')
    try {
      const res = await apiImportProducts(file)
      setImportResult(res)
      loadProducts()
    } catch (e) {
      setImportError(e instanceof Error ? e.message : 'Gagal mengimpor produk.')
    } finally {
      setImportBusy(false)
    }
  }

  function downloadTemplate() {
    const rows = [
      ['id', 'nama', 'sku', 'barcode', 'kategori', 'harga beli', 'harga jual', 'stok', 'satuan', 'aktif'],
      ['', 'Contoh Kopi Susu', 'KOPI-001', '8991234567890', 'Minuman', '8000', '12000', '24', 'pcs', 'true'],
    ]
    downloadBlob('template-import-produk.csv', new Blob(['\ufeff' + rows.map((r) => r.join(',')).join('\n')], { type: 'text/csv;charset=utf-8' }))
  }

  return (
    <>
      <PageHeader
        title="Produk & Kategori"
       
        crumb="Produk & Kategori"
        actions={(
          <div className="opc-head-actions">
            <ExportMenu onCSV={selectCsvExportHandler(session?.role, exportServerCSV, exportListCSV)} onExcel={exportListExcel} />
            <ImportMenu onPick={pickImport} />
            <Button variant="ghost" onClick={() => { setCatName(''); setErr(''); setCatOpen(true) }}><FolderPlus className="size-4" />Kategori</Button>
            <Button onClick={() => editProduct()}><Plus className="size-4" />Tambah Produk</Button>
            <input ref={fileRef} type="file" accept=".csv,.xlsx,.xls" hidden onChange={(e) => { const f = e.target.files?.[0]; if (f) onImportFile(f); e.target.value = '' }} />
          </div>
        )}
      />

      {(err || prod.err) && !editing && <AlertStrip className="mb-4">{err || prod.err}</AlertStrip>}

      <div className="mb-5">
        <div className="min-w-0">
          <div className="mb-4 flex flex-col gap-3 sm:flex-row">
            <SearchInput value={q} onChange={(e) => setQ(e.target.value)} placeholder="Cari nama atau SKU…" aria-label="Cari produk" />
            <Select
              label="Filter kategori"
              value={catFilter}
              onChange={setCatFilter}
              className="shrink-0 sm:w-56"
              options={[
                { value: '', label: `Semua kategori · ${products?.length ?? 0}` },
                ...activeCats.map((c) => ({ value: c.id, label: `${c.name} · ${counts.m.get(c.id) ?? 0}` })),
                ...(counts.none > 0 ? [{ value: NONE, label: `Tanpa kategori · ${counts.none}` }] : []),
              ]}
            />
            <FilterPanel
              className="shrink-0 sm:w-44"
              activeCount={(statusF ? 1 : 0) + (kondisiF ? 1 : 0)}
              onReset={() => { setStatusF(''); setKondisiF('') }}
            >
              <FilterField label="Status">
                <Select label="Filter status produk" value={statusF} onChange={setStatusF} options={STATUS_OPTS} />
              </FilterField>
              <FilterField label="Kondisi stok">
                <Select label="Filter kondisi stok" value={kondisiF} onChange={setKondisiF} options={KONDISI_OPTS} />
              </FilterField>
            </FilterPanel>
          </div>
          <FilterChips items={[
            ...(statusF ? [{ key: 'status', label: `Status: ${STATUS_OPTS.find((o) => o.value === statusF)?.label ?? statusF}`, onClear: () => setStatusF('') }] : []),
            ...(kondisiF ? [{ key: 'kondisi', label: `Kondisi: ${KONDISI_OPTS.find((o) => o.value === kondisiF)?.label ?? kondisiF}`, onClear: () => setKondisiF('') }] : []),
          ]} />
          {/* Tabel 8 kolom terukur 608px; di bawah 1280 konten hanya 438-890px
              sehingga tabel harus digeser horizontal. ≥1280 tetap tabel apa
              adanya; kartu di bawah yang menutup celah itu. */}
          <TableWrap className="hidden min-[1280px]:block">
            {!filtered || filtered.length > 0 ? (
              <table className="w-full border-collapse">
                <thead>
                  <tr>
                    <Th>Nama</Th><Th>SKU</Th><Th>Kategori</Th><Th right>Beli</Th><Th right>Jual</Th><Th right>Stok</Th><Th>Status</Th><Th>Aksi</Th>
                  </tr>
                </thead>
                {!filtered ? (
                  <SkeletonRows cols={8} />
                ) : (
                <tbody>
                  {(pageItems ?? []).map((p) => (
                    <tr key={p.id}>
                      <Td><span className="font-medium text-fg">{p.name}</span></Td>
                      <Td mono>{p.sku}</Td>
                      <Td>{p.category_name ?? 'Tanpa kategori'}</Td>
                      <Td right>{fmtRp(p.buy_price)}</Td>
                      <Td right><span className="font-medium text-fg">{fmtRp(p.sell_price)}</span></Td>
                      <Td right><StockCell product={p} /></Td>
                      <Td><Pill tone={p.active ? 'ok' : 'muted'}>{p.active ? 'Aktif' : 'Nonaktif'}</Pill></Td>
                      <Td>
                        <div className="flex items-center gap-1">
                          <RowAction title="Ubah" label={`Ubah ${p.name}`} onClick={() => editProduct(p)}><Pencil className="size-4" /></RowAction>
                          <RowAction title={p.active ? 'Nonaktifkan' : 'Aktifkan'} label={`${p.active ? 'Nonaktifkan' : 'Aktifkan'} ${p.name}`} onClick={() => toggleActive(p)}><Power className="size-4" /></RowAction>
                          <RowAction title="Hapus" label={`Hapus ${p.name}`} onClick={() => setDeleteFor(p)} danger><Trash2 className="size-4" /></RowAction>
                        </div>
                      </Td>
                    </tr>
                  ))}
                </tbody>
                )}
              </table>
            ) : products && products.length > 0 ? (
              <Empty title="Tidak ada produk yang cocok" sub="Coba kata kunci lain atau ubah filter kategori." action={<Button onClick={() => { setCatFilter(''); setQ('') }}>Tampilkan semua</Button>} />
            ) : (
              <Empty title="Belum ada produk" sub="Tambah produk pertama untuk mulai berjualan." action={<Button onClick={() => editProduct()}>+ Tambah Produk</Button>} />
            )}
          </TableWrap>

          {/* Di bawah 1280 kartu menggantikan tabel: 1 kolom <768, 2 kolom
              768-1119 (kartu ~238-362px), 3 kolom >=1120 (rail hilang, konten
              ~890px). Harga + ikon aksi satu blok footer — kalau tidak muat
              satu baris (kartu 238px di 768) ikon turun ke baris kedua dan
              tetap rata kanan, bukan terpotong. >=1280 kembali ke tabel. */}
          <CardGrid>
            {!filtered ? (
              [0, 1, 2, 3].map((i) => <Skeleton key={i} className="h-[104px] w-full rounded-xl" />)
            ) : filtered.length > 0 ? (
              (pageItems ?? []).map((p) => (
                <DataCard key={p.id}>
                  <div className="flex items-start justify-between gap-2">
                    <p className="min-w-0 flex-1 truncate text-[15px] font-medium text-fg">{p.name}</p>
                    <Pill tone={p.active ? 'ok' : 'muted'}>{p.active ? 'Aktif' : 'Nonaktif'}</Pill>
                  </div>
                  <p className="mt-1 flex items-center gap-1.5 text-xs text-fog">
                    <span className="min-w-0 truncate font-mono">{p.sku}</span>
                    <span className="shrink-0" aria-hidden="true">·</span>
                    <span className="min-w-0 truncate">{p.category_name ?? 'Tanpa kategori'}</span>
                    <span className="ml-auto shrink-0 whitespace-nowrap"><StockCell product={p} /></span>
                  </p>
                  {/* Harga di kiri, ikon aksi di kanan dalam satu blok. Kalau
                      tidak muat satu baris, blok membungkus dan ikon tetap
                      rata kanan — bukan terpotong. */}
                  <div className="mt-auto flex flex-wrap items-center justify-between gap-x-2 gap-y-1.5 pt-2.5">
                    <dl className="text-[13px] leading-snug">
                      <div className="flex items-baseline gap-1.5">
                        <dt className="shrink-0 text-fog">Beli</dt>
                        <dd className="font-mono tabular-nums text-muted">{fmtRp(p.buy_price)}</dd>
                      </div>
                      <div className="mt-0.5 flex items-baseline gap-1.5">
                        <dt className="shrink-0 text-fog">Jual</dt>
                        <dd className="font-mono text-[15px] font-medium tabular-nums text-fg">{fmtRp(p.sell_price)}</dd>
                      </div>
                    </dl>
                    <div className="ml-auto flex shrink-0 items-center gap-1">
                      <RowAction title="Ubah" label={`Ubah ${p.name}`} onClick={() => editProduct(p)}><Pencil className="size-4" /></RowAction>
                      <RowAction title={p.active ? 'Nonaktifkan' : 'Aktifkan'} label={`${p.active ? 'Nonaktifkan' : 'Aktifkan'} ${p.name}`} onClick={() => toggleActive(p)}><Power className="size-4" /></RowAction>
                      <RowAction title="Hapus" label={`Hapus ${p.name}`} onClick={() => setDeleteFor(p)} danger><Trash2 className="size-4" /></RowAction>
                    </div>
                  </div>
                </DataCard>
              ))
            ) : products && products.length > 0 ? (
              <div className="col-span-full"><Empty title="Tidak ada produk yang cocok" sub="Coba kata kunci lain atau ubah filter kategori." action={<Button onClick={() => { setCatFilter(''); setQ('') }}>Tampilkan semua</Button>} /></div>
            ) : (
              <div className="col-span-full"><Empty title="Belum ada produk" sub="Tambah produk pertama untuk mulai berjualan." action={<Button onClick={() => editProduct()}>+ Tambah Produk</Button>} /></div>
            )}
          </CardGrid>
          <Pager page={safePage} total={totalPages} onChange={setPage} totalItems={filtered?.length ?? 0} pageSize={pageSize} onPageSizeChange={(n) => { setPageSize(n); setPage(0) }} />
        </div>
      </div>

      <KategoriModal open={catOpen} onClose={() => setCatOpen(false)} catName={catName} onCatName={setCatName} busy={catBusy} onAdd={addCat} cats={activeCats} counts={counts.m} onRemove={removeCategory} />

      <Modal open={!!editing} title={editing?.id ? 'Ubah Produk' : 'Tambah Produk'} onClose={() => { if (!saveBusy) setEditing(null) }} wide>
        {editing && (
          <>
          {err && <AlertStrip className="mb-4">{err}</AlertStrip>}
          <FormProduk
            key={editing.id ?? 'new-product'}
            draft={editing}
            cats={cats.filter((c) => c.active)}
            lowStockDefault={settings?.lowStockDefault}
            onSave={save}
            onCancel={() => setEditing(null)}
          />
          </>
        )}
      </Modal>

      <Modal open={!!deleteFor} title={`Hapus Produk · ${deleteFor?.name ?? ''}`} onClose={() => setDeleteFor(null)}>
        <div className="space-y-4">
          <p className="text-sm text-muted">
            Tindakan ini <strong className="text-ember">permanen</strong>. Produk tidak bisa dikembalikan setelah dihapus.
          </p>
          <ModalActions>
            <Button variant="ghost" onClick={() => setDeleteFor(null)}>Batal</Button>
            <Button variant="danger" onClick={removeProduct}>Hapus Produk</Button>
          </ModalActions>
        </div>
      </Modal>

      <ImportProdukModal open={importOpen} onClose={() => { if (!importBusy) setImportOpen(false) }} busy={importBusy} result={importResult} error={importError} onDownloadTemplate={downloadTemplate} onPickFile={pickCsvFile} onImportAgain={() => { setImportResult(null); setImportError('') }} onDone={() => setImportOpen(false)} />
    </>
  )
}
