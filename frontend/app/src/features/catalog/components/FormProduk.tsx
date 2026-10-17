import { useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { ModalActions, Button, Input, NumInput, RowAction, Select } from '../../../shared/ui'
import { emptyVariant, type Draft, type TierDraft } from './draft'

export function FormProduk({ draft, cats, lowStockDefault, onSave, onCancel }: {
  draft: Draft
  cats: { id: string; name: string }[]
  lowStockDefault?: number
  onSave: (d: Draft) => Promise<Draft | null>
  onCancel: () => void
}) {
  const [d, setD] = useState(draft)
  const [saving, setSaving] = useState(false)
  const set = (k: keyof Draft) => (v: string) => setD({ ...d, [k]: v })
  const defaultHint = lowStockDefault == null ? 'Nilai bawaan toko belum diatur.' : `Nilai bawaan toko: ${lowStockDefault}.`

  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setSaving(true)
    try {
      const updated = await onSave(d)
      if (updated) setD(updated)
    } finally { setSaving(false) }
  }

  function updateTier(tiers: TierDraft[], index: number, key: keyof TierDraft, value: string) {
    return tiers.map((tier, ix) => ix === index ? { ...tier, [key]: value } : tier)
  }

  return (
    <form className="grid gap-4 sm:grid-cols-2" onSubmit={submit}>
      <div className="sm:col-span-2">
        <Input label="Nama produk" value={d.name} onChange={set('name')} required />
      </div>
      <Input label="SKU (unik)" value={d.sku} onChange={set('sku')} required />
      <Input label="Barcode (opsional)" value={d.barcode} onChange={set('barcode')} />
      <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
        Kategori
        <Select label="Kategori" value={d.categoryId} onChange={set('categoryId')} options={[{ value: '', label: 'Tanpa kategori' }, ...cats.map((c) => ({ value: String(c.id), label: c.name }))]} />
      </label>
      <Input label="Satuan" value={d.unit} onChange={set('unit')} />
      <NumInput label="Harga beli (Rp)" value={d.buyPrice} onValue={set('buyPrice')} />
      <NumInput label="Harga jual (Rp)" value={d.sellPrice} onValue={set('sellPrice')} required />
      {!d.id && d.variants.length === 0 && <NumInput label="Stok awal" value={d.stock} onValue={set('stock')} />}
      <NumInput label="Stok minimum" value={d.minStock} onValue={set('minStock')} placeholder={lowStockDefault == null ? '0' : String(lowStockDefault)} hint={defaultHint} />

      <section className="space-y-3 rounded-xl border border-dove p-3 sm:col-span-2">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <h3 className="text-sm font-medium text-fg">Harga bertingkat</h3>
            <p className="text-xs text-fog">Min. jumlah → harga satuan.</p>
          </div>
          <Button type="button" variant="ghost" className="px-3 py-1.5 text-sm" disabled={saving} onClick={() => setD({ ...d, priceTiers: [...d.priceTiers, { minQty: '', sellPrice: '' }] })}><Plus className="size-4" />Tambah harga</Button>
        </div>
        {d.priceTiers.map((tier, index) => (
          <div key={tier.id ?? `product-tier-${index}`} className="grid items-end gap-2 sm:grid-cols-[1fr_1fr_auto]">
            <NumInput label="Min. jumlah" value={tier.minQty} onValue={(v) => setD({ ...d, priceTiers: updateTier(d.priceTiers, index, 'minQty', v) })} required />
            <NumInput label="Harga jual (Rp)" value={tier.sellPrice} onValue={(v) => setD({ ...d, priceTiers: updateTier(d.priceTiers, index, 'sellPrice', v) })} required />
            <button type="button" aria-label={`Hapus harga bertingkat ${index + 1}`} title="Hapus harga bertingkat" disabled={saving} onClick={() => setD({ ...d, priceTiers: d.priceTiers.filter((_, ix) => ix !== index) })} className="mb-1 grid size-9 place-items-center rounded-md text-steel hover:bg-surface hover:text-ember disabled:opacity-40"><Trash2 className="size-4" /></button>
          </div>
        ))}
        {d.priceTiers.length === 0 && <p className="text-[13px] text-fog">Belum ada harga bertingkat.</p>}
      </section>

      <section className="space-y-3 rounded-xl border border-dove p-3 sm:col-span-2">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <h3 className="text-sm font-medium text-fg">Varian</h3>
            <p className="text-xs text-fog">Stok dan harga jual produk bervarian mengikuti data varian.</p>
          </div>
          <Button type="button" variant="ghost" className="px-3 py-1.5 text-sm" disabled={saving} onClick={() => setD({ ...d, variants: [...d.variants, { ...emptyVariant, minStock: lowStockDefault == null ? '' : String(lowStockDefault), priceTiers: [] }] })}><Plus className="size-4" />Tambah varian</Button>
        </div>
        {d.variants.map((variant, index) => (
          <div key={variant.id ?? `variant-${index}`} className="space-y-3 rounded-lg border border-dove bg-paper p-3">
            <div className="flex items-center justify-between gap-2">
              <h4 className="text-[13px] font-medium text-fg">Varian {index + 1}</h4>
              <RowAction title="Hapus varian" label={`Hapus varian ${variant.name || index + 1}`} onClick={() => setD({ ...d, variants: d.variants.filter((_, ix) => ix !== index) })} disabled={saving} danger><Trash2 className="size-4" /></RowAction>
            </div>
            <div className="grid gap-3 sm:grid-cols-2">
              <Input label="Nama varian" value={variant.name} onChange={(v) => setD({ ...d, variants: d.variants.map((row, ix) => ix === index ? { ...row, name: v } : row) })} required />
              <Input label="SKU varian (unik)" value={variant.sku} onChange={(v) => setD({ ...d, variants: d.variants.map((row, ix) => ix === index ? { ...row, sku: v } : row) })} required />
              <NumInput label="Harga beli (Rp)" value={variant.buyPrice} onValue={(v) => setD({ ...d, variants: d.variants.map((row, ix) => ix === index ? { ...row, buyPrice: v } : row) })} />
              <NumInput label="Harga jual (Rp)" value={variant.sellPrice} onValue={(v) => setD({ ...d, variants: d.variants.map((row, ix) => ix === index ? { ...row, sellPrice: v } : row) })} required />
              <NumInput label="Stok" value={variant.stock} onValue={(v) => setD({ ...d, variants: d.variants.map((row, ix) => ix === index ? { ...row, stock: v, stockEdited: true } : row) })} />
              <NumInput label="Stok minimum" value={variant.minStock} onValue={(v) => setD({ ...d, variants: d.variants.map((row, ix) => ix === index ? { ...row, minStock: v } : row) })} placeholder={lowStockDefault == null ? '0' : String(lowStockDefault)} hint={defaultHint} />
            </div>
            <div className="space-y-2 border-t border-dove pt-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <p className="text-[13px] font-medium text-steel">Harga bertingkat varian</p>
                <Button type="button" variant="ghost" className="px-3 py-1.5 text-sm" disabled={saving} onClick={() => setD({ ...d, variants: d.variants.map((row, ix) => ix === index ? { ...row, priceTiers: [...row.priceTiers, { minQty: '', sellPrice: '' }] } : row) })}><Plus className="size-4" />Tambah harga</Button>
              </div>
              {variant.priceTiers.map((tier, tierIndex) => (
                <div key={tier.id ?? `variant-${index}-tier-${tierIndex}`} className="grid items-end gap-2 sm:grid-cols-[1fr_1fr_auto]">
                  <NumInput label="Min. jumlah" value={tier.minQty} onValue={(v) => setD({ ...d, variants: d.variants.map((row, ix) => ix === index ? { ...row, priceTiers: updateTier(row.priceTiers, tierIndex, 'minQty', v) } : row) })} required />
                  <NumInput label="Harga jual (Rp)" value={tier.sellPrice} onValue={(v) => setD({ ...d, variants: d.variants.map((row, ix) => ix === index ? { ...row, priceTiers: updateTier(row.priceTiers, tierIndex, 'sellPrice', v) } : row) })} required />
                  <button type="button" aria-label={`Hapus harga bertingkat varian ${index + 1} baris ${tierIndex + 1}`} title="Hapus harga bertingkat" disabled={saving} onClick={() => setD({ ...d, variants: d.variants.map((row, ix) => ix === index ? { ...row, priceTiers: row.priceTiers.filter((_, ti) => ti !== tierIndex) } : row) })} className="mb-1 grid size-9 place-items-center rounded-md text-steel hover:bg-surface hover:text-ember disabled:opacity-40"><Trash2 className="size-4" /></button>
                </div>
              ))}
              {variant.priceTiers.length === 0 && <p className="text-xs text-fog">Belum ada harga bertingkat.</p>}
            </div>
          </div>
        ))}
        {d.variants.length === 0 && <p className="text-[13px] text-fog">Belum ada varian.</p>}
      </section>

      <ModalActions className="sm:col-span-2">
        <Button type="button" variant="ghost" onClick={onCancel} disabled={saving}>Batal</Button>
        <Button type="submit" disabled={saving}>{saving ? 'Menyimpan…' : 'Simpan'}</Button>
      </ModalActions>
    </form>
  )
}
