import { useState, type FormEvent } from 'react'
import { Pencil, Trash2 } from 'lucide-react'
import type { Category } from '../../../shared/api'
import { AlertStrip, Button, Input, Modal, ModalActions, RowAction } from '../../../shared/ui'

function CountBadge({ n, active }: { n: number; active: boolean }) {
  return (
    <span className={`shrink-0 rounded-full px-2 py-0.5 font-mono text-[11px] tabular-nums ${active ? 'bg-jet text-paper' : 'bg-surface text-muted'}`}>
      {n}
    </span>
  )
}

export function KategoriModal({ open, onClose, catName, onCatName, busy, onAdd, cats, counts, onRename, onRemove }: {
  open: boolean
  onClose: () => void
  catName: string
  onCatName: (v: string) => void
  busy: boolean
  onAdd: () => void
  cats: Category[]
  counts: Map<string, number>
  onRename: (category: Category, name: string) => Promise<void>
  onRemove: (c: Category) => void
}) {
  const [editing, setEditing] = useState<Category | null>(null)
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [renameBusy, setRenameBusy] = useState(false)

  function close() {
    if (renameBusy) return
    setEditing(null)
    setError('')
    onClose()
  }

  async function saveName(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!editing) return
    const nextName = name.trim()
    if (!nextName) return setError('Nama kategori wajib diisi.')
    setRenameBusy(true)
    setError('')
    try {
      await onRename(editing, nextName)
      setEditing(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Gagal mengubah nama kategori.')
    } finally {
      setRenameBusy(false)
    }
  }

  return (
<Modal open={open} title={editing ? 'Ubah Nama Kategori' : 'Kelola Kategori'} onClose={close}>
  {editing ? (
    <form className="space-y-3" onSubmit={(event) => void saveName(event)}>
      <Input label="Nama kategori" value={name} onChange={setName} required />
      {error && <AlertStrip>{error}</AlertStrip>}
      <ModalActions>
        <Button type="button" variant="ghost" disabled={busy || renameBusy} onClick={() => { setEditing(null); setError('') }}>Batal</Button>
        <Button type="submit" disabled={busy || renameBusy}>{renameBusy ? 'Menyimpan…' : 'Simpan Nama'}</Button>
      </ModalActions>
    </form>
  ) : (
    <>
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
        <div key={c.id} className="flex items-center gap-2 rounded-lg border border-dove bg-paper px-3 py-2 text-sm">
          <span className="min-w-0 flex-1 truncate">{c.name}</span>
          <CountBadge n={counts.get(c.id) ?? 0} active={false} />
          <Button
            variant="ghost"
            className="h-[42px] gap-1.5 px-2 text-xs"
            onClick={() => { setEditing(c); setName(c.name); setError('') }}
            disabled={busy}
          ><Pencil className="size-3.5" />Ubah nama</Button>
          <RowAction title="Hapus kategori" label={`Hapus kategori ${c.name}`} onClick={() => onRemove(c)} disabled={busy} danger><Trash2 className="size-4" /></RowAction>
        </div>
      ))}
      {cats.length === 0 && <p className="text-sm text-fog">Belum ada kategori.</p>}
    </div>
  </div>
    </>
  )}
</Modal>
  )
}
