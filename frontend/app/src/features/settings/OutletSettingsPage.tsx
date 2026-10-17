import { useState, type FormEvent } from 'react'
import type { Outlet } from './api'
import { useCreateOutlet, useUpdateOutlet } from './hooks'
import { useOutlets } from '../../hooks/use-outlets'
import { outletPayload } from '../../shared/outlet'
import { ModalActions, AlertStrip, Button, Input, Modal, Pill } from '../../shared/ui'

export default function OutletSettingsPage() {
  const createOutlet = useCreateOutlet()
  const updateOutlet = useUpdateOutlet()
  const { outlets, loading, error: loadError, reload } = useOutlets()
  const [open, setOpen] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState({ name: '', address: '' })
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [message, setMessage] = useState('')

  function addOutlet() {
    setEditingId(null)
    setForm({ name: '', address: '' })
    setError('')
    setOpen(true)
  }

  function editOutlet(outlet: Outlet) {
    setEditingId(outlet.id)
    setForm({ name: outlet.name, address: outlet.address })
    setError('')
    setOpen(true)
  }

  async function saveOutlet(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const body = outletPayload(form.name, form.address)
    if (!body.name) return setError('Nama outlet wajib diisi.')
    setBusy(true)
    setError('')
    setMessage('')
    try {
      if (editingId) await updateOutlet.mutateAsync({ id: editingId, body })
      else await createOutlet.mutateAsync(body)
      setOpen(false)
      setMessage(editingId ? 'Outlet berhasil diperbarui.' : 'Outlet berhasil ditambahkan.')
      reload()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Gagal menyimpan outlet. Coba lagi.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <section className="rounded-2xl bg-cream p-4 sm:p-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-[15px] font-medium text-fg">Outlet</h2>
          <p className="mt-0.5 text-[13px] text-muted">Kelola nama dan alamat outlet toko Anda.</p>
        </div>
        <Button onClick={addOutlet}>Tambah Outlet</Button>
      </div>

      {message && <p className="mt-4 rounded-lg bg-surface px-3.5 py-2.5 text-[13px] text-sprout" role="status">{message}</p>}
      {loadError && <AlertStrip className="mt-4">{loadError}</AlertStrip>}
      <div className="mt-4 divide-y divide-dove border-t border-dove">
        {loading ? (
          <p className="py-4 text-sm text-muted">Memuat outlet…</p>
        ) : outlets.length === 0 ? (
          <p className="py-4 text-sm text-muted">Belum ada outlet.</p>
        ) : outlets.map((outlet) => (
          <div key={outlet.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-2">
                <p className="text-sm font-medium text-fg">{outlet.name}</p>
                {outlet.is_default && <Pill tone="ok">Outlet default</Pill>}
              </div>
              <p className="mt-0.5 break-words text-[13px] text-muted">{outlet.address || 'Alamat belum diisi.'}</p>
            </div>
            <Button variant="ghost" className="px-4 py-2 text-[13px]" onClick={() => editOutlet(outlet)}>Edit</Button>
          </div>
        ))}
      </div>

      <Modal open={open} title={editingId ? 'Edit Outlet' : 'Tambah Outlet'} onClose={() => { if (!busy) setOpen(false) }}>
        <form className="space-y-3" onSubmit={(event) => void saveOutlet(event)}>
          <Input
            label="Nama outlet"
            value={form.name}
            onChange={(name) => setForm((current) => ({ ...current, name }))}
            placeholder="Contoh: Cabang Makassar"
            required
          />
          <Input
            label="Alamat"
            value={form.address}
            onChange={(address) => setForm((current) => ({ ...current, address }))}
            placeholder="Alamat outlet"
          />
          {error && <AlertStrip>{error}</AlertStrip>}
          <ModalActions>
            <Button variant="ghost" disabled={busy} onClick={() => setOpen(false)}>Batal</Button>
            <Button type="submit" disabled={busy}>{busy ? 'Menyimpan…' : 'Simpan Outlet'}</Button>
          </ModalActions>
        </form>
      </Modal>
    </section>
  )
}
