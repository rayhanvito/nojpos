import { useCallback, useEffect, useRef, useState } from 'react'
import { useLocation } from 'react-router'
import { WifiOff } from 'lucide-react'
import { ApiError, apiCloseCashierShift, apiGetCashierShift, apiStartCashierShift, type CashierShift, type ShiftSummary } from '../shared/api'
import { fmtRp, useDB } from '../shared/session/store'
import { cashierShiftCloseBody, shiftMethodRows, shiftVarianceParts } from '../shared/shift'
import type { OfflineQueueItem } from '../shared/offline-queue'
import { TextArea, AlertStrip, Button, Modal, ModalActions, NumInput } from '../shared/ui'

function fmtVariance(v: number) {
  const { sign, amount } = shiftVarianceParts(v)
  if (sign === '') return { text: 'Pas', cls: 'text-sprout' }
  if (sign === '+') return { text: `Lebih ${fmtRp(amount)}`, cls: 'text-sprout' }
  return { text: `Kurang ${fmtRp(amount)}`, cls: 'text-ember' }
}

function sejak(iso: string) {
  return new Date(iso).toLocaleString('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}

const QUEUE_STATUS_LABEL: Record<OfflineQueueItem['status'], string> = {
  menunggu: 'Menunggu',
  tersinkron: 'Tersinkron',
  'gagal-perlu-tindakan': 'Gagal · perlu tindakan',
}

// Pil status tunggal di header (desain final Rayhan 16 Okt 2026): shift dan
// koneksi/antrean digabung jadi SATU pil + SATU popup Status.
// Warna pil = masalah paling genting: offline → merah (ember), online tapi
// shift belum dibuka → oranye (sunbeam), shift aktif + online → hijau
// (sprout). Label selalu soal shift, TANPA angka Rp (detail di popup).
// Badge jumlah antrean selalu tampil bila ada. Klik pil selalu membuka
// popup Status yang sama (perilaku tidak berubah-ubah).
export default function ShiftControl({ online, waitingCount, queueItems, queueError, queueAction, onRetryItem, onDismissItem, onRefreshQueue }: {
  online: boolean
  waitingCount: number
  queueItems: OfflineQueueItem[]
  queueError: string
  queueAction: string | null
  onRetryItem: (clientTxId: string) => void
  onDismissItem: (clientTxId: string) => void
  onRefreshQueue: () => void
}) {
  const { session } = useDB()
  const loc = useLocation()
  const [shift, setShift] = useState<CashierShift | null>(null)
  const shiftRef = useRef<CashierShift | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadErr, setLoadErr] = useState('')

  const [statusOpen, setStatusOpen] = useState(false)
  const [openingCash, setOpeningCash] = useState('')
  const [closingCash, setClosingCash] = useState('')
  const [closeNote, setCloseNote] = useState('')
  const [busy, setBusy] = useState(false)
  const [actionErr, setActionErr] = useState('')
  const [result, setResult] = useState<ShiftSummary | null>(null)

  const load = useCallback(async () => {
    try {
      const r = await apiGetCashierShift()
      shiftRef.current = r.shift
      setShift(r.shift)
      setLoadErr('')
    } catch (e) {
      if (e instanceof ApiError && e.status === 404) {
        shiftRef.current = null
        setShift(null)
        setLoadErr('')
      } else if (!shiftRef.current) {
        // Galat muat (mis. offline) hanya ditampilkan bila belum ada data
        // shift; shift yang sudah termuat tetap ditampilkan apa adanya —
        // status koneksi sendiri sudah diwakili seksi Koneksi di popup.
        setLoadErr(e instanceof Error ? e.message : 'Gagal memuat status shift.')
      }
    } finally {
      setLoading(false)
    }
  }, [])

  // Muat ulang saat akun berganti dan setiap pindah halaman.
  useEffect(() => { void load() }, [load, session?.id, loc.pathname])

  async function doOpen() {
    setBusy(true); setActionErr('')
    try {
      const r = await apiStartCashierShift(Number(openingCash) || 0)
      setShift(r.shift)
      window.dispatchEvent(new Event('pos:shift-changed'))
      setOpeningCash('')
    } catch (e) {
      setActionErr(e instanceof Error ? e.message : 'Gagal membuka shift.')
    } finally { setBusy(false) }
  }

  async function doClose() {
    const body = cashierShiftCloseBody(closingCash, closeNote)
    if (!body) { setActionErr('Uang fisik harus bilangan bulat dan tidak boleh negatif.'); return }
    setBusy(true); setActionErr('')
    try {
      const r = await apiCloseCashierShift(body)
      setResult(r.summary)
      setShift(null)
      window.dispatchEvent(new Event('pos:shift-changed'))
    } catch (e) {
      setActionErr(e instanceof Error ? e.message : 'Gagal menutup shift.')
    } finally { setBusy(false) }
  }

  function closeStatus() {
    if (busy) return
    if (result) { setResult(null); setClosingCash(''); setCloseNote(''); void load() }
    setStatusOpen(false)
  }

  function openStatus() {
    setActionErr('')
    void load()
    onRefreshQueue()
    setStatusOpen(true)
  }

  const offline = !online
  const dotCls = offline ? 'bg-ember' : shift ? 'bg-sprout' : 'bg-sunbeam'
  const label = loading || loadErr ? 'Shift' : shift ? 'Shift aktif' : 'Buka Shift'
  const closingParsed = cashierShiftCloseBody(closingCash, '')
  const closingNum = closingParsed ? closingParsed.closingCash : null
  const liveVariance = shift && closingNum !== null && !Number.isNaN(closingNum) ? closingNum - shift.expected_cash : null

  return (
    <>
      <button
        type="button"
        onClick={openStatus}
        aria-label={`Status: ${offline ? 'Offline' : 'Online'}; ${shift ? `shift aktif sejak ${sejak(shift.started_at)}` : 'shift belum dibuka'}${waitingCount > 0 ? `; ${waitingCount} transaksi menunggu sinkronisasi` : ''}. Buka status`}
        title={offline ? 'Offline' : shift ? `Shift aktif · sejak ${sejak(shift.started_at)}` : 'Shift belum dibuka'}
        className={`inline-flex h-9 items-center gap-1.5 whitespace-nowrap rounded px-1.5 text-[12px] font-medium text-white transition hover:bg-white/10 ${!loading && !shift && online ? 'bg-white/10' : ''}`}
      >
        <span className={`size-2 rounded-full ${dotCls}`} aria-hidden="true" />
        {offline && <WifiOff className="size-3.5" aria-hidden="true" />}
        {label}
        {waitingCount > 0 && (
          <span className="rounded bg-black/25 px-1.5 py-0.5 font-mono tabular-nums">{waitingCount}</span>
        )}
      </button>

      <Modal open={statusOpen} title="Status" onClose={closeStatus} wide>
        <div className="space-y-5">
          <section aria-label="Koneksi dan antrean transaksi">
            <h4 className="mb-2 text-sm font-semibold">Koneksi &amp; Antrean</h4>
            <p className="rounded-lg bg-surface px-3.5 py-2.5 text-sm" role="status">
              Koneksi: <strong>{online ? 'Online' : 'Offline'}</strong>
              {waitingCount > 0 && <span> · {waitingCount} transaksi menunggu sinkronisasi</span>}
            </p>
            {queueError && <AlertStrip className="mt-2">{queueError}</AlertStrip>}
            {queueItems.length === 0 ? (
              <p className="mt-2 rounded-lg border border-dashed border-dove px-4 py-5 text-center text-sm text-muted">Belum ada transaksi dalam antrean.</p>
            ) : (
              <div className="mt-2 max-h-[40dvh] space-y-2 overflow-y-auto pr-1">
                {queueItems.map((item) => {
                  const statusClass = item.status === 'tersinkron'
                    ? 'bg-success-bg text-sprout'
                    : item.status === 'gagal-perlu-tindakan'
                      ? 'bg-sand text-ember'
                      : 'bg-surface text-muted'
                  return (
                    <article key={item.clientTxId} className="rounded-lg border border-dove bg-paper p-3.5">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <span className={`rounded-full px-2.5 py-0.5 text-[11px] font-medium ${statusClass}`}>{QUEUE_STATUS_LABEL[item.status]}</span>
                        <time className="text-xs text-muted">{new Date(item.createdAt).toLocaleString('id-ID')}</time>
                      </div>
                      <div className="mt-2 flex flex-wrap gap-1.5">
                        {item.summary.items.map((line, index) => (
                          <span key={`${line.name}-${index}`} className="inline-flex items-center gap-1.5 rounded-md border border-dove bg-surface px-2 py-1 text-[13px]">
                            <span>{line.name}</span><span className="font-mono text-[11px] text-muted">×{line.qty}</span>
                          </span>
                        ))}
                      </div>
                      <p className="mt-2 text-sm font-medium">
                        {typeof item.serverTotal === 'number' ? 'Total server' : 'Perkiraan total'}: {fmtRp(item.serverTotal ?? item.summary.estimatedTotal)}
                      </p>
                      {item.serverStatus === 'completed_offline_shortage' && (
                        <p className="mt-1 text-xs text-sunbeam" role="status">Transaksi diterima server, stok kurang saat sinkron (completed_offline_shortage).</p>
                      )}
                      {item.serverStatus && item.serverStatus !== 'completed_offline_shortage' && (
                        <p className="mt-1 text-xs text-muted">Status server: {item.serverStatus}</p>
                      )}
                      {item.error && <p className="mt-1 text-xs text-ember" role="alert">{item.error}</p>}
                      <div className="mt-3 flex justify-end gap-2">
                        {item.status === 'gagal-perlu-tindakan' && (
                          <button
                            type="button"
                            disabled={!online || queueAction !== null}
                            onClick={() => onRetryItem(item.clientTxId)}
                            className="rounded-md border border-dove px-3 py-1.5 text-xs font-medium text-fg hover:border-jet disabled:opacity-40"
                          >
                            Coba lagi
                          </button>
                        )}
                        {item.status !== 'menunggu' && (
                          <button
                            type="button"
                            disabled={queueAction !== null}
                            onClick={() => onDismissItem(item.clientTxId)}
                            className="rounded-md border border-dove px-3 py-1.5 text-xs font-medium text-fg hover:border-jet disabled:opacity-40"
                          >
                            {item.status === 'tersinkron' ? 'Tutup' : 'Abaikan'}
                          </button>
                        )}
                      </div>
                    </article>
                  )
                })}
              </div>
            )}
          </section>

          <section aria-label="Shift">
            <h4 className="mb-2 text-sm font-semibold">Shift</h4>
            {result ? (
              <div className="space-y-4">
                <AlertStrip tone="success">Shift ditutup.</AlertStrip>
                <div className="rounded-lg border border-dove p-4 text-center">
                  <p className="text-xs text-muted">Selisih kas</p>
                  <p className={`mt-1 text-2xl font-semibold tabular-nums ${fmtVariance(result.cash_variance ?? 0).cls}`}>{fmtVariance(result.cash_variance ?? 0).text}</p>
                  <p className="mt-2 text-[13px] text-muted">
                    Diharapkan {fmtRp(result.expected_cash ?? 0)} · Aktual {fmtRp(result.closing_cash_actual ?? 0)} · Penjualan {fmtRp(result.sales)} ({result.trx_count} transaksi)
                  </p>
                </div>
                <ModalActions>
                  <Button onClick={closeStatus}>Selesai</Button>
                </ModalActions>
              </div>
            ) : loadErr && !shift ? (
              <AlertStrip>{loadErr}</AlertStrip>
            ) : shift ? (
              <div className="flex flex-col gap-4">
                <p className="text-[13px] text-muted">Dibuka {sejak(shift.started_at)} · {shift.trx_count} transaksi sejauh ini</p>
                <div className="rounded-lg border border-dove">
                  <dl className="divide-y divide-dove text-sm">
                    {shiftMethodRows(shift.sales_by_method).filter((r) => r.total !== 0).map((r) => (
                      <div key={r.method} className="flex items-center justify-between px-3.5 py-2">
                        <dt className="text-muted">{r.method}</dt>
                        <dd className="font-medium tabular-nums">{fmtRp(r.total)}</dd>
                      </div>
                    ))}
                    <div className="flex items-center justify-between px-3.5 py-2">
                      <dt className="text-muted">Total penjualan</dt>
                      <dd className="font-medium tabular-nums">{fmtRp(shift.sales)}</dd>
                    </div>
                    <div className="flex items-center justify-between px-3.5 py-2">
                      <dt className="text-muted">Refund</dt>
                      <dd className="font-medium tabular-nums">{fmtRp(shift.refund_total)}</dd>
                    </div>
                    {(shift.cash_expenses ?? 0) > 0 && (
                      <div className="flex items-center justify-between px-3.5 py-2">
                        <dt className="text-muted">Pengeluaran laci</dt>
                        <dd className="font-medium tabular-nums">−{fmtRp(shift.cash_expenses ?? 0)}</dd>
                      </div>
                    )}
                    <div className="flex items-center justify-between px-3.5 py-2">
                      <dt className="text-muted">Modal awal</dt>
                      <dd className="font-medium tabular-nums">{fmtRp(shift.opening_cash)}</dd>
                    </div>
                    <div className="flex items-center justify-between bg-surface px-3.5 py-2.5">
                      <dt className="font-medium">Kas yang diharapkan</dt>
                      <dd className="text-base font-semibold tabular-nums">{fmtRp(shift.expected_cash)}</dd>
                    </div>
                  </dl>
                </div>
                <NumInput label="Uang fisik di laci (Rp)" value={closingCash} onValue={setClosingCash} placeholder="0" autoFocus required />
                {liveVariance !== null && (
                  <p className={`text-sm font-medium ${fmtVariance(liveVariance).cls}`} role="status">Selisih sementara: {fmtVariance(liveVariance).text}</p>
                )}
                <TextArea label="Catatan (opsional)" value={closeNote} onChange={setCloseNote} rows={2} placeholder="Mis. uang pas, tidak ada selisih" />
                {offline && <p className="text-[13px] text-ember" role="status">Offline — tutup shift butuh koneksi.</p>}
                {actionErr && <AlertStrip>{actionErr}</AlertStrip>}
                <ModalActions>
                  <Button variant="ghost" disabled={busy} onClick={closeStatus}>Batal</Button>
                  <Button disabled={busy || closingNum === null || closingNum < 0 || offline} onClick={() => void doClose()}>{busy ? 'Menutup…' : 'Tutup Shift'}</Button>
                </ModalActions>
              </div>
            ) : (
              <div className="space-y-4">
                <p className="text-sm text-muted">Shift belum dibuka. Buka shift untuk mencatat modal awal laci.</p>
                <NumInput label="Modal awal di laci (Rp)" value={openingCash} onValue={setOpeningCash} placeholder="0" autoFocus required />
                {actionErr && <AlertStrip>{actionErr}</AlertStrip>}
                <ModalActions>
                  <Button variant="ghost" disabled={busy} onClick={closeStatus}>Batal</Button>
                  <Button disabled={busy} onClick={() => void doOpen()}>{busy ? 'Membuka…' : 'Buka Shift'}</Button>
                </ModalActions>
              </div>
            )}
          </section>
        </div>
      </Modal>
    </>
  )
}
