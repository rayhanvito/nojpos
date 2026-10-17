import { Trash2 } from 'lucide-react'
import type { Category } from '../../../shared/api'
import { Button, Input, Modal, RowAction } from '../../../shared/ui'

function CountBadge({ n, active }: { n: number; active: boolean }) {
  return (
    <span className={`shrink-0 rounded-full px-2 py-0.5 font-mono text-[11px] tabular-nums ${active ? 'bg-jet text-paper' : 'bg-surface text-muted'}`}>
      {n}
    </span>
  )
}

export function KategoriModal({ open, onClose, catName, onCatName, busy, onAdd, cats, counts, onRemove }: {
  open: boolean
  onClose: () => void
  catName: string
  onCatName: (v: string) => void
  busy: boolean
  onAdd: () => void
  cats: Category[]
  counts: Map<string, number>
  onRemove: (c: Category) => void
}) {
  return (
<Modal open={open} title="Kelola Kategori" onClose={onClose}>
  <form
    className="flex flex-col gap-3 sm:flex-row sm:gap-2"
    onSubmit={(e) => { e.preventDefault(); onAdd() }}
  >
    <div className="min-w-0 flex-1">
    <Input label="Nama kategori baru" value={catName} onChange={onCatName} placeholder="cth: Minuman" />
    </div>
    <div className="flex items-end">
      <Button type="submit" disabled={busy} className="w-full sm:w-auto">Simpan</Button>
    </div>
  </form>
  <div className="mt-5 border-t border-dove pt-4">
    <h3 className="text-sm font-semibold">Daftar kategori</h3>
    <div className="mt-2 max-h-64 space-y-1.5 overflow-y-auto">
      {cats.map((c) => (
        <div key={c.id} className="flex items-center justify-between gap-2 rounded-lg border border-dove bg-paper px-3 py-2 text-sm">
          <span className="min-w-0 flex-1 truncate">{c.name}</span>
          <CountBadge n={counts.get(c.id) ?? 0} active={false} />
          <RowAction title="Hapus kategori" label={`Hapus kategori ${c.name}`} onClick={() => onRemove(c)} disabled={busy} danger><Trash2 className="size-4" /></RowAction>
        </div>
      ))}
      {cats.length === 0 && <p className="text-sm text-fog">Belum ada kategori.</p>}
    </div>
  </div>
</Modal>
  )
}
