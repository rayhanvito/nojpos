// Tab Kasir di Laporan — kinerja per orang per periode (pengganti halaman
// Karyawan). Agregat penjualan datang dari bundle laporan (dihitung server,
// definisi sama dengan ringkasan Laporan); sesi shift diambil dari daftar
// shift periode yang sama untuk jumlah sesi, jam sesi, dan selisih kas.
import { Fragment, useMemo, useState } from 'react'
import { Banknote, ChevronDown, ChevronUp, ReceiptText, Scale, UsersRound } from 'lucide-react'
import { apiGetReport, apiListShifts, apiListUsers, type ReportBundle, type Shift as ApiShift, type User } from '../lib/api'
import { useCache } from '../lib/cache'
import { outletQuery } from '../lib/outlet'
import { durationLabelFromMs, isForceClosedShift, shiftCacheKey, shiftDurationLabel, shiftStatusLabel, shiftVarianceParts } from '../lib/shift'
import { fmtRp, useDB } from '../lib/store'
import { CardGrid, DataCard, Empty, Pill, StatCard, Td, Th, TableWrap } from '../lib/ui'

export interface KasirRow {
  name: string
  cashierId: number
  trx: number
  items: number
  gross: number
  refund: number
  net: number
  avg: number
  topProduct: string
  topProductQty: number
  shiftCount: number
  hoursMs: number
  variance: number
  sessions: ApiShift[]
  prevNet: number | null
  trend: number | null
}

function sessionHoursMs(sessions: ApiShift[], nowMs: number): number {
  return sessions.reduce((n, s) => n + Math.max(0, (s.closed_at ? new Date(s.closed_at).getTime() : nowMs) - new Date(s.started_at).getTime()), 0)
}

export function buildKasirRows(data: ReportBundle, users: User[] | null, shifts: ApiShift[], prev: ReportBundle | null, nowMs: number): KasirRow[] {
  const byName = new Map<string, ApiShift[]>()
  for (const s of shifts) {
    const arr = byName.get(s.cashier_name) ?? []
    arr.push(s)
    byName.set(s.cashier_name, arr)
  }
  const prevById = new Map((prev?.cashiers ?? []).map((c) => [c.cashier_id, c.net]))
  const rows: KasirRow[] = (data.cashiers ?? []).map((c) => {
    const sessions = byName.get(c.name) ?? []
    const prevNet = prevById.get(c.cashier_id) ?? null
    return {
      name: c.name,
      cashierId: c.cashier_id,
      trx: c.trx,
      items: c.items,
      gross: c.gross,
      refund: c.refund,
      net: c.net,
      avg: c.trx > 0 ? Math.round(c.net / c.trx) : 0,
      topProduct: c.top_product,
      topProductQty: c.top_product_qty,
      shiftCount: sessions.length,
      hoursMs: sessionHoursMs(sessions, nowMs),
      variance: sessions.reduce((n, s) => n + (s.cash_variance ?? 0), 0),
      sessions,
      prevNet,
      trend: prevNet !== null && prevNet > 0 ? (c.net - prevNet) / prevNet : null,
    }
  })
  const seen = new Set(rows.map((r) => r.name))
  for (const u of users ?? []) {
    if (u.role !== 'cashier' || !u.active || seen.has(u.name)) continue
    const sessions = byName.get(u.name) ?? []
    rows.push({
      name: u.name, cashierId: 0, trx: 0, items: 0, gross: 0, refund: 0, net: 0, avg: 0,
      topProduct: '', topProductQty: 0,
      shiftCount: sessions.length, hoursMs: sessionHoursMs(sessions, nowMs),
      variance: sessions.reduce((n, s) => n + (s.cash_variance ?? 0), 0),
      sessions, prevNet: null, trend: null,
    })
  }
  return rows.sort((a, b) => b.net - a.net)
}

function isoDay(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

type PrevSpec = { kind: 'yesterday' } | { kind: 'range'; from: string; to: string } | null

function prevSpecFor(period: string, range: { from: string; to: string } | null): PrevSpec {
  if (range) {
    const from = new Date(`${range.from}T00:00:00`)
    const to = new Date(`${range.to}T00:00:00`)
    const lenDays = Math.round((to.getTime() - from.getTime()) / 86400000) + 1
    const prevTo = new Date(from)
    prevTo.setDate(prevTo.getDate() - 1)
    const prevFrom = new Date(prevTo)
    prevFrom.setDate(prevFrom.getDate() - (lenDays - 1))
    return { kind: 'range', from: isoDay(prevFrom), to: isoDay(prevTo) }
  }
  const now = new Date()
  if (period === 'today') return { kind: 'yesterday' }
  if (period === 'yesterday') {
    const d = new Date(now)
    d.setDate(d.getDate() - 2)
    const v = isoDay(d)
    return { kind: 'range', from: v, to: v }
  }
  if (period === 'week') {
    const mon = new Date(now)
    mon.setDate(mon.getDate() - ((mon.getDay() + 6) % 7))
    const prevMon = new Date(mon)
    prevMon.setDate(prevMon.getDate() - 7)
    const prevSun = new Date(prevMon)
    prevSun.setDate(prevSun.getDate() + 6)
    return { kind: 'range', from: isoDay(prevMon), to: isoDay(prevSun) }
  }
  if (period === 'month') {
    const first = new Date(now.getFullYear(), now.getMonth() - 1, 1)
    const last = new Date(now.getFullYear(), now.getMonth(), 0)
    return { kind: 'range', from: isoDay(first), to: isoDay(last) }
  }
  return null
}

function varianceText(value: number): string {
  const v = shiftVarianceParts(value)
  return `${v.sign}${fmtRp(v.amount)}`
}

export default function LaporanKasir({ data, outletId, period, range, from, to }: {
  data: ReportBundle
  outletId: string
  period: string
  range: { from: string; to: string } | null
  from?: string
  to?: string
}) {
  const { session } = useDB()
  const sessionKey = session ? `${session.id}:${session.role}` : 'anonymous'
  const selectedOutlet = outletQuery(outletId).outletId
  const outletKey = selectedOutlet ?? 'all'
  const rangeKey = from && to ? `${from}..${to}` : 'all'
  const shifts = useCache<ApiShift[]>(shiftCacheKey(sessionKey, outletKey, rangeKey), () => apiListShifts({ ...outletQuery(selectedOutlet), ...(from && to ? { from, to } : {}) }), 'Gagal memuat daftar shift.')
  const users = useCache<User[]>(`kasir-users:${sessionKey}`, () => apiListUsers(), 'Gagal memuat daftar kasir.')
  const spec = prevSpecFor(period, range)
  const prevRep = useCache<ReportBundle | null>(
    `reportkasirprev:${outletKey}:${spec ? JSON.stringify(spec) : 'none'}`,
    () => (spec === null
      ? Promise.resolve(null)
      : spec.kind === 'yesterday'
        ? apiGetReport('yesterday', undefined, selectedOutlet)
        : apiGetReport('today', undefined, selectedOutlet, { from: spec.from, to: spec.to })),
    'Gagal memuat periode pembanding.',
  )
  const rows = useMemo(
    () => buildKasirRows(data, users.data, shifts.data ?? [], prevRep.data, Date.now()),
    [data, users.data, shifts.data, prevRep.data],
  )
  const [openName, setOpenName] = useState<string | null>(null)

  const activeCount = rows.filter((r) => r.trx > 0).length
  const totalNet = rows.reduce((n, r) => n + r.net, 0)
  const totalTrx = rows.reduce((n, r) => n + r.trx, 0)
  const totalVariance = rows.reduce((n, r) => n + r.variance, 0)

  if (rows.length === 0) {
    return <Empty title="Belum ada data kasir" sub="Kinerja kasir pada periode ini akan tampil di sini." />
  }

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3 xl:grid-cols-4">
        <StatCard icon={UsersRound} label="Kasir aktif" value={String(activeCount)} sub={`dari ${rows.length} kasir tercatat`} />
        <StatCard icon={Banknote} label="Total omzet tim" value={fmtRp(totalNet)} sub="neto setelah refund, sama dengan ringkasan Laporan" />
        <StatCard icon={ReceiptText} label="Rata-rata per transaksi" value={totalTrx > 0 ? fmtRp(Math.round(totalNet / totalTrx)) : '—'} sub={`${totalTrx} transaksi tim pada periode ini`} />
        <StatCard icon={Scale} label="Total selisih kas" value={varianceText(totalVariance)} sub="dari sesi shift tercatat pada periode ini" />
      </div>

      <TableWrap className="hidden min-[1280px]:block">
        <table className="w-full border-collapse">
          <thead>
            <tr>
              <Th>Kasir</Th><Th right>Transaksi</Th><Th right>Item</Th><Th right>Bruto</Th><Th right>Refund</Th><Th right>Neto</Th><Th right>Rata-rata</Th><Th right>Shift</Th><Th right>Jam Sesi</Th><Th right>Selisih Kas</Th><Th right>Tren</Th>
            </tr>
          </thead>
          <tbody>
            {rows.map((r) => {
              const open = openName === r.name
              return (
                <Fragment key={r.name}>
                  <tr
                    tabIndex={0}
                    aria-label={`Buka rincian sesi ${r.name}`}
                    onClick={() => setOpenName(open ? null : r.name)}
                    onKeyDown={(event) => {
                      if (event.target !== event.currentTarget || (event.key !== 'Enter' && event.key !== ' ')) return
                      event.preventDefault()
                      setOpenName(open ? null : r.name)
                    }}
                    className="cursor-pointer hover:bg-paper/50 focus-visible:outline-2 focus-visible:outline-accent-soft"
                  >
                    <Td>{r.name}</Td>
                    <Td right>{r.trx}</Td>
                    <Td right>{r.items}</Td>
                    <Td right>{fmtRp(r.gross)}</Td>
                    <Td right>{r.refund ? fmtRp(r.refund) : '—'}</Td>
                    <Td right>{fmtRp(r.net)}</Td>
                    <Td right>{r.trx > 0 ? fmtRp(r.avg) : '—'}</Td>
                    <Td right>{r.shiftCount}</Td>
                    <Td right>{r.shiftCount > 0 ? durationLabelFromMs(r.hoursMs) : '—'}</Td>
                    <Td right>{r.sessions.length === 0 ? '—' : r.variance === 0 ? <Pill tone="ok">Pas</Pill> : <Pill tone={r.variance < 0 ? 'warn' : 'muted'}>{varianceText(r.variance)}</Pill>}</Td>
                    <Td right>{r.trend === null ? '—' : <span className={r.trend >= 0 ? 'text-sprout' : 'text-ember'}>{r.trend >= 0 ? '▲' : '▼'} {Math.abs(Math.round(r.trend * 100))}%</span>}</Td>
                  </tr>
                  {open && (
                    <tr>
                      <td colSpan={11} className="border-b border-dove bg-paper/60 px-3.5 py-3">
                        <p className="mb-2 text-xs text-muted">
                          Produk terlaris: <strong className="text-fg">{r.topProduct ? `${r.topProduct} (×${r.topProductQty})` : '—'}</strong>
                          {' · '}Pembanding periode lalu: <strong className="text-fg">{r.prevNet === null ? '—' : fmtRp(r.prevNet)}</strong>
                        </p>
                        {r.sessions.length === 0 ? (
                          <p className="text-sm text-muted">Tidak ada sesi shift tercatat pada periode ini.</p>
                        ) : (
                          <div className="overflow-x-auto">
                            <table className="w-full border-collapse">
                              <thead>
                                <tr><Th>Mulai</Th><Th>Selesai</Th><Th>Durasi</Th><Th right>Penjualan</Th><Th right>Kas Diharapkan</Th><Th right>Kas Aktual</Th><Th right>Selisih</Th><Th>Status</Th></tr>
                              </thead>
                              <tbody>
                                {r.sessions.map((s) => (
                                  <tr key={s.id}>
                                    <Td mono>{new Date(s.started_at).toLocaleString('id-ID')}</Td>
                                    <Td mono>{s.closed_at ? new Date(s.closed_at).toLocaleString('id-ID') : '—'}</Td>
                                    <Td mono>{shiftDurationLabel(s.started_at, s.closed_at, Date.now())}</Td>
                                    <Td right>{fmtRp(s.sales)}</Td>
                                    <Td right>{s.expected_cash === null ? '—' : fmtRp(s.expected_cash)}</Td>
                                    <Td right>{s.closing_cash_actual === null || s.closing_cash_actual === undefined ? '—' : fmtRp(s.closing_cash_actual)}</Td>
                                    <Td right>{s.cash_variance === null || s.cash_variance === undefined ? '—' : varianceText(s.cash_variance)}</Td>
                                    <Td>{shiftStatusLabel(s)}{isForceClosedShift(s) && s.closed_by_name ? ` · oleh ${s.closed_by_name}` : ''}</Td>
                                  </tr>
                                ))}
                              </tbody>
                            </table>
                          </div>
                        )}
                      </td>
                    </tr>
                  )}
                </Fragment>
              )
            })}
          </tbody>
        </table>
      </TableWrap>

      <CardGrid>
        {rows.map((r) => {
          const open = openName === r.name
          return (
            <DataCard key={r.name}>
              <div className="flex items-center justify-between gap-2">
                <p className="min-w-0 truncate font-medium text-fg">{r.name}</p>
                {r.trend === null ? <span className="text-xs text-fog">tren —</span> : <span className={`text-xs font-medium ${r.trend >= 0 ? 'text-sprout' : 'text-ember'}`}>{r.trend >= 0 ? '▲' : '▼'} {Math.abs(Math.round(r.trend * 100))}%</span>}
              </div>
              <p className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-fg">{fmtRp(r.net)}</p>
              <p className="mt-0.5 text-right text-xs text-fog">omzet neto</p>
              <div className="mt-2 grid grid-cols-3 gap-2 text-xs">
                <p><span className="block text-fog">Transaksi</span><span className="font-medium tabular-nums text-fg">{r.trx}</span></p>
                <p><span className="block text-fog">Item</span><span className="font-medium tabular-nums text-fg">{r.items}</span></p>
                <p><span className="block text-fog">Rata-rata</span><span className="font-medium tabular-nums text-fg">{r.trx > 0 ? fmtRp(r.avg) : '—'}</span></p>
                <p><span className="block text-fog">Bruto</span><span className="font-medium tabular-nums text-fg">{fmtRp(r.gross)}</span></p>
                <p><span className="block text-fog">Refund</span><span className="font-medium tabular-nums text-fg">{r.refund ? fmtRp(r.refund) : '—'}</span></p>
                <p><span className="block text-fog">Selisih kas</span><span className="font-medium tabular-nums text-fg">{r.sessions.length === 0 ? '—' : varianceText(r.variance)}</span></p>
                <p><span className="block text-fog">Shift</span><span className="font-medium tabular-nums text-fg">{r.shiftCount}</span></p>
                <p><span className="block text-fog">Jam sesi</span><span className="font-medium tabular-nums text-fg">{r.shiftCount > 0 ? durationLabelFromMs(r.hoursMs) : '—'}</span></p>
              </div>
              <button
                type="button" aria-expanded={open} aria-label={`Rincian sesi ${r.name}`}
                onClick={() => setOpenName(open ? null : r.name)}
                className="mt-3 flex w-full items-center justify-center gap-1.5 rounded-lg border border-dove bg-paper py-2 text-[13px] font-medium transition hover:border-jet"
              >
                {open ? 'Tutup rincian' : 'Rincian sesi'}
                {open ? <ChevronUp className="size-4" /> : <ChevronDown className="size-4" />}
              </button>
              {open && (
                <div className="mt-3 space-y-2 border-t border-dove pt-3">
                  <p className="text-xs text-muted">
                    Produk terlaris: <strong className="text-fg">{r.topProduct ? `${r.topProduct} (×${r.topProductQty})` : '—'}</strong>
                    {' · '}Pembanding periode lalu: <strong className="text-fg">{r.prevNet === null ? '—' : fmtRp(r.prevNet)}</strong>
                  </p>
                  {r.sessions.length === 0 ? (
                    <p className="text-sm text-muted">Tidak ada sesi shift tercatat pada periode ini.</p>
                  ) : r.sessions.map((s) => (
                    <div key={s.id} className="rounded-lg border border-dove bg-paper p-2.5">
                      <div className="flex items-center justify-between gap-2">
                        <p className="min-w-0 truncate text-xs text-fog">{new Date(s.started_at).toLocaleString('id-ID')} – {s.closed_at ? new Date(s.closed_at).toLocaleString('id-ID') : 'berjalan'}</p>
                        <Pill tone="muted">{shiftStatusLabel(s)}</Pill>
                      </div>
                      <p className="mt-1 text-xs leading-snug text-fog">
                        Penjualan <span className="font-medium tabular-nums text-fg">{fmtRp(s.sales)}</span> · Selisih <span className="font-medium tabular-nums text-fg">{s.cash_variance === null || s.cash_variance === undefined ? '—' : varianceText(s.cash_variance)}</span> · {shiftDurationLabel(s.started_at, s.closed_at, Date.now())}
                        {isForceClosedShift(s) && s.closed_by_name ? ` · ditutup ${s.closed_by_name}` : ''}
                      </p>
                      <p className="mt-0.5 text-xs leading-snug text-fog">Kas diharapkan {s.expected_cash === null ? '—' : fmtRp(s.expected_cash)} · Kas aktual {s.closing_cash_actual === null || s.closing_cash_actual === undefined ? '—' : fmtRp(s.closing_cash_actual)}</p>
                    </div>
                  ))}
                </div>
              )}
            </DataCard>
          )
        })}
      </CardGrid>

      <p className="text-xs text-muted">
        Jam sesi dan selisih kas dihitung dari sesi shift yang tercatat pada periode ini; penjualan tanpa shift terbuka tidak masuk hitungan jam.
        Tren membandingkan omzet neto dengan periode sebelumnya yang sama panjang.
      </p>
    </div>
  )
}
