import { useState, type FormEvent } from 'react'
import { Copy, Trash2 } from 'lucide-react'
import type { ApiKey, Webhook, WebhookEvent } from './api'
import { fmtDate } from '../../shared/session/store'
import { AlertStrip, Button, DataCard, Empty, Input, Modal, ModalActions, Pill, SettingsCard, SettingsHead } from '../../shared/ui'
import {
  useCreateApiKey, useCreateWebhook, useDeleteApiKey, useDeleteWebhook, useIntegrationApiKeys, useIntegrationWebhooks,
} from './hooks'

const WEBHOOK_EVENTS: { value: WebhookEvent; label: string }[] = [
  { value: 'transaction.created', label: 'Transaksi dibuat' },
  { value: 'transaction.refunded', label: 'Transaksi direfund' },
  { value: 'stock.updated', label: 'Stok berubah' },
]

function dateLabel(value: string) {
  return fmtDate(value)
}

function prefixLabel(prefix: string) {
  return `${prefix.slice(0, 8)}…`
}

type PendingDelete = { kind: 'key'; item: ApiKey } | { kind: 'webhook'; item: Webhook }

export default function IntegrationsSettings() {
  const keys = useIntegrationApiKeys()
  const webhooks = useIntegrationWebhooks()
  const createKey = useCreateApiKey()
  const deleteKey = useDeleteApiKey()
  const createWebhook = useCreateWebhook()
  const deleteWebhook = useDeleteWebhook()
  const [keyOpen, setKeyOpen] = useState(false)
  const [keyName, setKeyName] = useState('')
  const [keySecret, setKeySecret] = useState('')
  const [keyError, setKeyError] = useState('')
  const [keyBusy, setKeyBusy] = useState(false)
  const [webhookOpen, setWebhookOpen] = useState(false)
  const [webhookUrl, setWebhookUrl] = useState('')
  const [webhookEvents, setWebhookEvents] = useState<WebhookEvent[]>([])
  const [webhookSecret, setWebhookSecret] = useState('')
  const [webhookError, setWebhookError] = useState('')
  const [webhookBusy, setWebhookBusy] = useState(false)
  const [copyMessage, setCopyMessage] = useState('')
  const [message, setMessage] = useState('')
  const [pendingDelete, setPendingDelete] = useState<PendingDelete | null>(null)
  const [deleteError, setDeleteError] = useState('')
  const [deleteBusy, setDeleteBusy] = useState(false)

  function openKeyForm() {
    setKeyName('')
    setKeySecret('')
    setKeyError('')
    setCopyMessage('')
    setKeyOpen(true)
  }

  function closeKeyForm() {
    if (keyBusy) return
    setKeyOpen(false)
    setKeySecret('')
    setCopyMessage('')
    createKey.reset()
  }

  async function submitKey(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const name = keyName.trim()
    if (!name) return setKeyError('Nama API Key wajib diisi.')
    setKeyBusy(true)
    setKeyError('')
    try {
      const result = await createKey.mutateAsync(name)
      setKeySecret(result.apiKey)
    } catch (e) {
      setKeyError(e instanceof Error ? e.message : 'Gagal membuat API Key.')
    } finally {
      setKeyBusy(false)
    }
  }

  function openWebhookForm() {
    setWebhookUrl('')
    setWebhookEvents([])
    setWebhookSecret('')
    setWebhookError('')
    setCopyMessage('')
    setWebhookOpen(true)
  }

  function closeWebhookForm() {
    if (webhookBusy) return
    setWebhookOpen(false)
    setWebhookSecret('')
    setCopyMessage('')
    createWebhook.reset()
  }

  async function submitWebhook(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const url = webhookUrl.trim()
    const events = WEBHOOK_EVENTS.filter((choice) => webhookEvents.includes(choice.value)).map((choice) => choice.value)
    if (!events.length) return setWebhookError('Pilih minimal satu event webhook.')
    setWebhookBusy(true)
    setWebhookError('')
    try {
      const result = await createWebhook.mutateAsync({ url, events })
      setWebhookSecret(result.secret)
    } catch (e) {
      setWebhookError(e instanceof Error ? e.message : 'Gagal membuat webhook.')
    } finally {
      setWebhookBusy(false)
    }
  }

  async function copySecret(secret: string) {
    try {
      await navigator.clipboard.writeText(secret)
      setCopyMessage('Berhasil disalin.')
    } catch {
      setCopyMessage('Tidak dapat menyalin otomatis. Pilih dan salin secret secara manual.')
    }
  }

  async function confirmDelete() {
    if (!pendingDelete || deleteBusy) return
    setDeleteBusy(true)
    setDeleteError('')
    try {
      if (pendingDelete.kind === 'key') {
        await deleteKey.mutateAsync(pendingDelete.item.id)
        setMessage('API Key berhasil dihapus.')
      } else {
        await deleteWebhook.mutateAsync(pendingDelete.item.id)
        setMessage('Webhook berhasil dihapus.')
      }
      setPendingDelete(null)
    } catch (e) {
      setDeleteError(e instanceof Error ? e.message : 'Gagal menghapus integrasi.')
    } finally {
      setDeleteBusy(false)
    }
  }

  return (
    <div className="space-y-4">
      {message && <p className="rounded-lg bg-surface px-3.5 py-2.5 text-[13px] text-sprout" role="status">{message}</p>}
      <SettingsCard>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <SettingsHead title="API Keys" sub="Gunakan kunci untuk menghubungkan aplikasi atau layanan lain." />
          <Button onClick={openKeyForm}>Buat API Key</Button>
        </div>
        {keys.err && <AlertStrip className="mt-4">{keys.err}</AlertStrip>}
        {keys.loading ? <p className="mt-4 text-sm text-muted">Memuat API Keys…</p> : (keys.data?.items.length ?? 0) === 0 ? (
          <div className="mt-4"><Empty title="Belum ada API Key" sub="Buat kunci untuk menghubungkan layanan lain." /></div>
        ) : (
          <div className="mt-4 grid grid-cols-1 gap-2.5 min-[768px]:grid-cols-2">
            {keys.data!.items.map((key) => (
              <DataCard key={key.id}>
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0 flex-1">
                    <p className="break-words text-sm font-semibold text-fg">{key.name}</p>
                    <p className="mt-1 break-all font-mono text-xs text-fog">Prefix: {prefixLabel(key.prefix)}</p>
                    <p className="mt-1 text-xs text-muted">Dibuat {dateLabel(key.created_at)}</p>
                    <p className="mt-1 text-xs text-muted">Terakhir dipakai: {key.last_used_at ? dateLabel(key.last_used_at) : 'Belum pernah'}</p>
                  </div>
                  <Button variant="ghost" className="h-[42px] shrink-0 px-2 text-xs" onClick={() => { setDeleteError(''); setPendingDelete({ kind: 'key', item: key }) }}>
                    <Trash2 className="size-3.5" />Hapus
                  </Button>
                </div>
              </DataCard>
            ))}
          </div>
        )}
      </SettingsCard>

      <SettingsCard>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <SettingsHead title="Webhooks" sub="Kirim event toko ke URL layanan yang Anda kelola." />
          <Button onClick={openWebhookForm}>Tambah Webhook</Button>
        </div>
        {webhooks.err && <AlertStrip className="mt-4">{webhooks.err}</AlertStrip>}
        {webhooks.loading ? <p className="mt-4 text-sm text-muted">Memuat webhook…</p> : (webhooks.data?.items.length ?? 0) === 0 ? (
          <div className="mt-4"><Empty title="Belum ada webhook" sub="Tambahkan URL untuk menerima event toko." /></div>
        ) : (
          <div className="mt-4 grid grid-cols-1 gap-2.5 min-[768px]:grid-cols-2">
            {webhooks.data!.items.map((webhook) => (
              <DataCard key={webhook.id}>
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0 flex-1">
                    <p className="break-all text-sm font-medium text-fg">{webhook.url}</p>
                    <div className="mt-2 flex flex-wrap gap-1.5">
                      {webhook.events.map((event) => <Pill key={event} tone="muted">{WEBHOOK_EVENTS.find((choice) => choice.value === event)?.label ?? event}</Pill>)}
                    </div>
                    <p className="mt-2 text-xs text-muted">Dibuat {dateLabel(webhook.created_at)}</p>
                  </div>
                  <Button variant="ghost" className="h-[42px] shrink-0 px-2 text-xs" onClick={() => { setDeleteError(''); setPendingDelete({ kind: 'webhook', item: webhook }) }}>
                    <Trash2 className="size-3.5" />Hapus
                  </Button>
                </div>
              </DataCard>
            ))}
          </div>
        )}
      </SettingsCard>

      <Modal open={keyOpen} title={keySecret ? 'API Key dibuat' : 'Buat API Key'} onClose={closeKeyForm}>
        {keySecret ? (
          <div className="space-y-3">
            <p className="text-sm text-muted">Salin dan simpan kunci ini sekarang. Kunci tidak dapat dilihat lagi setelah sheet ditutup.</p>
            <code className="block break-all rounded-lg bg-surface p-3 font-mono text-sm text-fg">{keySecret}</code>
            {copyMessage && <p className="text-[13px] text-muted" role="status">{copyMessage}</p>}
            <Button className="w-full" onClick={() => void copySecret(keySecret)}><Copy className="size-4" />Salin API Key</Button>
            <ModalActions><Button onClick={closeKeyForm}>Selesai</Button></ModalActions>
          </div>
        ) : (
          <form className="space-y-3" onSubmit={(event) => void submitKey(event)}>
            <Input label="Nama API Key" value={keyName} onChange={setKeyName} required placeholder="Contoh: Integrasi laporan" />
            {keyError && <AlertStrip>{keyError}</AlertStrip>}
            <ModalActions>
              <Button type="button" variant="ghost" disabled={keyBusy} onClick={closeKeyForm}>Batal</Button>
              <Button type="submit" disabled={keyBusy}>{keyBusy ? 'Membuat…' : 'Buat API Key'}</Button>
            </ModalActions>
          </form>
        )}
      </Modal>

      <Modal open={webhookOpen} title={webhookSecret ? 'Webhook dibuat' : 'Tambah Webhook'} onClose={closeWebhookForm}>
        {webhookSecret ? (
          <div className="space-y-3">
            <p className="text-sm text-muted">Simpan secret ini sekarang. Gunakan untuk memverifikasi tanda tangan X-NOJPOS-Signature (HMAC-SHA256).</p>
            <code className="block break-all rounded-lg bg-surface p-3 font-mono text-sm text-fg">{webhookSecret}</code>
            {copyMessage && <p className="text-[13px] text-muted" role="status">{copyMessage}</p>}
            <Button className="w-full" onClick={() => void copySecret(webhookSecret)}><Copy className="size-4" />Salin Secret</Button>
            <ModalActions><Button onClick={closeWebhookForm}>Selesai</Button></ModalActions>
          </div>
        ) : (
          <form className="space-y-3" onSubmit={(event) => void submitWebhook(event)}>
            <Input label="URL endpoint" type="url" value={webhookUrl} onChange={setWebhookUrl} required placeholder="https://contoh.com/webhook" />
            <fieldset className="space-y-1">
              <legend className="mb-1 text-[13px] font-medium text-steel">Event yang dikirim</legend>
              {WEBHOOK_EVENTS.map((choice) => (
                <label key={choice.value} className="flex min-h-[42px] cursor-pointer items-center justify-between gap-3 rounded-lg px-2 text-sm text-fg hover:bg-surface">
                  {choice.label}
                  <input type="checkbox" checked={webhookEvents.includes(choice.value)} onChange={(event) => setWebhookEvents((current) => event.target.checked ? [...current, choice.value] : current.filter((value) => value !== choice.value))} className="h-4 w-4 accent-jet" />
                </label>
              ))}
            </fieldset>
            {webhookError && <AlertStrip>{webhookError}</AlertStrip>}
            <ModalActions>
              <Button type="button" variant="ghost" disabled={webhookBusy} onClick={closeWebhookForm}>Batal</Button>
              <Button type="submit" disabled={webhookBusy}>{webhookBusy ? 'Menyimpan…' : 'Simpan Webhook'}</Button>
            </ModalActions>
          </form>
        )}
      </Modal>

      <Modal
        open={!!pendingDelete}
        title={pendingDelete?.kind === 'key' ? 'Hapus API Key?' : 'Hapus Webhook?'}
        onClose={() => { if (!deleteBusy) { setPendingDelete(null); setDeleteError('') } }}
      >
        <div className="space-y-3">
          <p className="text-sm text-muted">{pendingDelete?.kind === 'key' ? 'Kunci yang dihapus langsung berhenti bekerja.' : 'Webhook yang dihapus langsung berhenti menerima event.'}</p>
          {deleteError && <AlertStrip>{deleteError}</AlertStrip>}
          <ModalActions>
            <Button variant="ghost" disabled={deleteBusy} onClick={() => setPendingDelete(null)}>Batal</Button>
            <Button variant="danger" disabled={deleteBusy} onClick={() => void confirmDelete()}>{deleteBusy ? 'Menghapus…' : 'Hapus'}</Button>
          </ModalActions>
        </div>
      </Modal>
    </div>
  )
}
