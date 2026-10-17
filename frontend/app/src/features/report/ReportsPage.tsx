import { useEffect, useMemo, useState } from 'react'
import { Loader2 } from 'lucide-react'
import {
  apiExportShifts, apiExportTaxReport, apiListShifts, apiListUsers,
  type Shift as ApiShift, type User,
} from './api'
import {
  useReport, useReportCatalog, useReportHourly, useReportMovements, useReportPrev, useTaxReport,
} from './hooks'
import Shift from '../shift/ShiftPage'
import { TaxSection } from './components/TaxSection'
import { LaporanSkeleton } from './components/LaporanSkeleton'
import { ProductsTab } from './components/ProductsTab'
import { ProfitTab } from './components/ProfitTab'
import { SalesTab } from './components/SalesTab'
import { StockTab } from './components/StockTab'
import { PERIODS, catDonutFrom, catMapFrom, catValueFrom, dailyFrom, localDayISO, lowProductsFrom, lowStockFrom, prodPageData, prodProfitFrom, prodSortedFrom, profitDailyFrom, profitInsightsFrom, profitPageData, rowsProductsFrom, rowsProfitFrom, rowsSalesFrom, rowsStockFrom, stockRowsFrom, stockStatusFrom, topProductsFrom, type Period } from './components/derive'
import CashierReport, { buildKasirRows } from './CashierReportPage'
import { durationLabelFromMs } from '../../shared/shift'
import { PageHeader } from '../../shared/PageHeader'
import { downloadBlob, exportCSV, exportExcel, fmtDate, fmtTime } from '../../shared/session/store'
import { AlertStrip, ExportMenu, PeriodPicker } from '../../shared/ui'
import { Tabs, TabsList, TabsTrigger } from '@/shared/ui/primitives/tabs'
import { taxMonthsForYear } from '../../shared/tax-report'
import { REPORT_TABS, REPORT_TAB_ALIASES, type ReportTab } from '../../shared/tab-params'
import { useTabParam } from '../../shared/use-tab-param'
import { OutletFilter } from '../../components/OutletFilter'
import { useOutlets } from '../../hooks/use-outlets'
import { outletQuery } from '../../shared/outlet'

export default function Reports() {
  const [period, setPeriod] = useState<Period>('today')
  const [tab, setTab] = useTabParam<ReportTab>(REPORT_TABS, REPORT_TAB_ALIASES, 'sales')
  const now = new Date()
  const taxYear = now.getFullYear()
  const currentTaxMonth = `${taxYear}-${String(now.getMonth() + 1).padStart(2, '0')}`
  const taxMonths = useMemo(() => taxMonthsForYear(taxYear), [taxYear])
  const [selectedTaxMonth, setSelectedTaxMonth] = useState(currentTaxMonth)
  const [taxExportErr, setTaxExportErr] = useState('')
  // Rentang kustom dari pemilih periode (null = ikut preset). Backend menerima
  // from/to inklusif per hari lokal; rentang 1 hari = laporan tanggal itu.
  const [range, setRange] = useState<{ from: string; to: string } | null>(null)
  const [outletId, setOutletId] = useState('')
  const { outlets } = useOutlets()
  const selectedOutlet = outletQuery(outletId).outletId
  // Animasi chart hidup hanya saat mount; refresh/pindah tab tak me-restartnya.
  const [animate, setAnimate] = useState(true)
  useEffect(() => {
    const t = setTimeout(() => setAnimate(false), 900)
    return () => clearTimeout(t)
  }, [])
  // Tab Shift digerakkan kontrol halaman yang sama: rentang efektif dari
  // pemilih periode (preset dipetakan ke rentang tanggal; 'all' = tanpa batas).
  const shiftRange = useMemo(() => {
    if (range) return range
    const now = new Date()
    const iso = (d: Date) => `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    const today = iso(now)
    if (period === 'today') return { from: today, to: today }
    if (period === 'yesterday') { const y = new Date(now); y.setDate(y.getDate() - 1); const v = iso(y); return { from: v, to: v } }
    if (period === 'week') { const m = new Date(now); m.setDate(m.getDate() - ((m.getDay() + 6) % 7)); return { from: iso(m), to: today } }
    if (period === 'month') return { from: iso(new Date(now.getFullYear(), now.getMonth(), 1)), to: today }
    return null
  }, [period, range])
  const exportShiftCsv = async () => {
    try {
      const file = await apiExportShifts({ ...(outletId ? { outletId } : {}), ...(shiftRange ?? {}) })
      await downloadBlob(file.filename ?? 'laporan-shift.csv', file.blob)
    } catch {
      // Ekspor gagal: daftar shift tetap tampil; pengguna bisa mencoba lagi.
    }
  }

  const rep = useReport(period, selectedOutlet, range)
  const data = rep.data
  const err = rep.err
  const taxRep = useTaxReport(taxYear, taxMonths, selectedOutlet)
  // Pembanding "kemarin" hanya terdefinisi untuk periode hari ini (API tak punya
  // periode lalu untuk kemarin/minggu/bulan/semua) — dipakai tab sales & profit.
  const needPrev = (tab === 'sales' || tab === 'profit') && period === 'today' && !range
  const prevRep = useReportPrev(needPrev, tab, period, selectedOutlet)
  const prev = needPrev ? prevRep.data : null
  // Kata pembanding mengikuti mode: periode = "kemarin", tanggal kustom = "sebelumnya" (H-1).
  const banding = 'kemarin'
  // Katalog untuk join kategori + hitung produk aktif (tab Produk & Stok).
  const catRep = useReportCatalog(tab)
  // Riwayat global 5 terbaru (API tak dukung filter periode) — tab Stok saja.
  const movRep = useReportMovements(tab)
  const [prodAll, setProdAll] = useState(false)
  const [prodPage, setProdPage] = useState(0)
  const [profitPage, setProfitPage] = useState(0)
  const [prodPageSize, setProdPageSize] = useState(10)
  const [profitPageSize, setProfitPageSize] = useState(10)
  useEffect(() => { setProdAll(false); setProdPage(0); setProfitPage(0) }, [period, range])
  const catMap = useMemo(() => catMapFrom(catRep.data), [catRep.data])

  // Data per jam untuk periode 1 hari: t.date laporan hanya tanggal, jadi ambil
  // jam dari daftar transaksi harian (created_at) — endpoint existing, tanpa API baru.
  // HPP per slot = Σ buy_price × qty per item (TrxItem existing).
  const needHourly = (tab === 'sales' || tab === 'profit') && !range && (period === 'today' || period === 'yesterday')
  const dayStr = period === 'yesterday' ? localDayISO(1) : localDayISO(0)
  const hourRep = useReportHourly(needHourly, tab, period, dayStr)
  const hourly = needHourly ? hourRep.data : undefined
  // Indikator switching: ganti periode (fetch ulang) MAUPUN pindah tab yang butuh
  // data tambahan (katalog) — pola sama seperti kasus periode Hari ini dkk.
  const tabLabel = tab === 'sales' ? 'Penjualan' : tab === 'products' ? 'Produk' : tab === 'profit' ? 'Profit' : tab === 'cashier' ? 'Kasir' : tab === 'tax' ? 'Pajak' : 'Stok'
  const auxLoading = (tab === 'products' || tab === 'stock') && catRep.loading && !catRep.data
  const switching = (rep.loading && data !== null) || auxLoading
  const switchingText = auxLoading && !(rep.loading && data !== null)
    ? `Memuat tab ${tabLabel}…`
    : range ? `Memuat rentang ${fmtDate(range.from)} – ${fmtDate(range.to)}…` : `Memuat periode ${PERIODS.find((p) => p.id === period)?.label}…`

  const daily = useMemo(() => dailyFrom(data, needHourly, hourly), [data, needHourly, hourly])

  const topProducts = useMemo(() => topProductsFrom(data), [data])
  const prodSorted = useMemo(() => prodSortedFrom(data), [data])
  const lowProducts = useMemo(() => lowProductsFrom(data), [data])
  const prodProfit = useMemo(() => prodProfitFrom(data), [data])

  const { prodPages, prodPageSafe, prodShown } = prodPageData(prodSorted, prodAll, prodPage, prodPageSize)
  const { profitPages, profitPageSafe } = profitPageData(prodProfit, profitPage, profitPageSize)

  const profitDaily = useMemo(() => profitDailyFrom(data, needHourly, hourly), [data, needHourly, hourly])

  // Insight valid dari data existing (maks 5, tanpa klaim buatan).
  const profitInsights = useMemo(() => profitInsightsFrom(data, period, prev, catMap, prodProfit, banding), [data, period, range, prev, catMap, prodProfit])

  // Donut kategori: gabung agregat laporan + kategori katalog; 4 teratas + Lainnya.
  const catDonut = useMemo(() => catDonutFrom(data, catMap), [data, catMap])

  // Baris stok + kategori katalog (cocok SKU lalu nama).
  const stockRows = useMemo(() => stockRowsFrom(data, catRep.data), [data, catRep.data])

  const stockStatus = useMemo(() => stockStatusFrom(stockRows), [stockRows])

  const lowStock = useMemo(() => lowStockFrom(stockRows), [stockRows])

  const catValue = useMemo(() => catValueFrom(stockRows), [stockRows])

  const periodName = range ? `${range.from}_${range.to}` : period

  async function exportKasirCSV() {
    if (!data) return
    const [users, shifts] = await Promise.all([
      apiListUsers().catch(() => [] as User[]),
      apiListShifts({ ...(outletId ? { outletId } : {}), ...(shiftRange ?? {}) }).catch(() => [] as ApiShift[]),
    ])
    const rows = buildKasirRows(data, users, shifts, null, Date.now())
    exportCSV(`laporan-kasir-${periodName}.csv`, [
      ['kasir', 'transaksi', 'item', 'bruto', 'refund', 'neto', 'rata_rata', 'jumlah_shift', 'jam_sesi', 'selisih_kas', 'produk_terlaris'],
      ...rows.map((r) => [r.name, String(r.trx), String(r.items), String(r.gross), String(r.refund), String(r.net), String(r.avg), String(r.shiftCount), durationLabelFromMs(r.hoursMs), String(r.variance), r.topProduct]),
    ])
  }

  function exportTabCSV() {
    if (tab === 'shift') { void exportShiftCsv(); return }
    if (tab === 'cashier') { void exportKasirCSV(); return }
    if (tab === 'tax') { void exportTaxCSV(); return }
    if (!data) return
    if (tab === 'sales') exportCSV(`laporan-penjualan-${periodName}.csv`, rowsSalesFrom(data))
    else if (tab === 'products') exportCSV(`laporan-produk-${periodName}.csv`, rowsProductsFrom(data))
    else if (tab === 'stock') exportCSV(`laporan-stok.csv`, rowsStockFrom(data))
    else exportCSV(`laporan-profit-${periodName}.csv`, rowsProfitFrom(data))
  }

  // Excel selalu 1 workbook berisi SEMUA tab sekaligus (termasuk Shift) — CSV tetap per tab aktif.
  async function exportWorkbookExcel() {
    if (!data) return
    const shifts = await apiListShifts({ ...(outletId ? { outletId } : {}), ...(shiftRange ?? {}) }).catch(() => [] as ApiShift[])
    const periodLabel = range ? `${fmtDate(range.from)} – ${fmtDate(range.to)}` : (PERIODS.find((p) => p.id === period)?.label ?? period)
    const kasirUsers = await apiListUsers().catch(() => [] as User[])
    const kasirRows = buildKasirRows(data, kasirUsers, shifts, null, Date.now())
    await exportExcel(`laporan-${periodName}.xlsx`, [
      {
        name: 'Penjualan', title: 'Laporan Penjualan', subtitle: `Periode: ${periodLabel} · ${data.transactions.length} transaksi`,
        columns: [
          { header: 'Tanggal' }, { header: 'ID' }, { header: 'Kasir' }, { header: 'Metode' },
          { header: 'Total (Rp)', money: true },
        ],
        rows: data.transactions.map((t) => [t.date, t.id, t.cashier, t.method, t.total]),
        sumCols: [4],
      },
      {
        name: 'Produk', title: 'Laporan Produk', subtitle: `Periode: ${periodLabel} · ${data.products.length} produk terjual`,
        columns: [
          { header: 'Produk' }, { header: 'SKU' }, { header: 'Qty Terjual', align: 'right' },
          { header: 'Pendapatan (Rp)', money: true }, { header: 'Profit (Rp)', money: true },
        ],
        rows: data.products.map((p) => [p.name, p.sku, p.qty, p.revenue, p.profit]),
        sumCols: [2, 3, 4],
      },
      {
        name: 'Stok', title: 'Laporan Stok', subtitle: `${data.stock.length} produk`,
        columns: [
          { header: 'Produk' }, { header: 'SKU' }, { header: 'Stok', align: 'right' },
          { header: 'Harga Beli (Rp)', money: true }, { header: 'Harga Jual (Rp)', money: true },
          { header: 'Nilai Stok (Rp)', money: true },
        ],
        rows: data.stock.map((s) => [s.name, s.sku, s.stock, s.buy_price, s.sell_price, s.stock_value]),
        sumCols: [2, 5],
      },
      {
        name: 'Profit', title: 'Laporan Profit', subtitle: `Periode: ${periodLabel}`,
        columns: [
          { header: 'Tanggal' }, { header: 'ID' }, { header: 'Kasir' },
          { header: 'Total (Rp)', money: true }, { header: 'HPP (Rp)', money: true }, { header: 'Profit (Rp)', money: true },
        ],
        rows: data.transactions.map((t) => [t.date, t.id, t.cashier, t.total, t.hpp, t.profit]),
        sumCols: [3, 4, 5],
      },
      ...(data.expenses && data.expenses.length > 0 ? [{
        name: 'Pengeluaran', title: 'Pengeluaran Operasional', subtitle: `Periode: ${periodLabel} · ${data.expenses.length} kategori`,
        columns: [
          { header: 'Kategori' }, { header: 'Jumlah Catatan', align: 'right' as const }, { header: 'Total (Rp)', money: true },
        ],
        rows: data.expenses.map((row) => [row.category, row.count, row.total]),
        sumCols: [1, 2],
      }] : []),
      {
        name: 'Shift', title: 'Laporan Shift Kasir', subtitle: `Periode: ${periodLabel} · ${shifts.length} shift`,
        columns: [
          { header: 'Kasir' }, { header: 'Mulai' }, { header: 'Selesai' },
          { header: 'Modal Awal (Rp)', money: true },
          { header: 'Penjualan Cash (Rp)', money: true }, { header: 'Penjualan Bank Transfer (Rp)', money: true },
          { header: 'Penjualan QRIS (Rp)', money: true }, { header: 'Penjualan E-Wallet (Rp)', money: true },
          { header: 'Penjualan Card (Rp)', money: true }, { header: 'Penjualan Credit (Rp)', money: true },
          { header: 'Refund (Rp)', money: true }, { header: 'Total Penjualan (Rp)', money: true },
          { header: 'Jumlah Transaksi', align: 'right' as const },
          { header: 'Kas Diharapkan (Rp)', money: true }, { header: 'Kas Aktual (Rp)', money: true }, { header: 'Selisih (Rp)', money: true },
        ],
        rows: shifts.map((sh) => [
          sh.cashier_name,
          `${fmtDate(sh.started_at)} ${fmtTime(sh.started_at)}`,
          sh.closed_at ? `${fmtDate(sh.closed_at)} ${fmtTime(sh.closed_at)}` : '',
          sh.opening_cash,
          sh.sales_by_method.Cash ?? 0, sh.sales_by_method['Bank Transfer'] ?? 0,
          sh.sales_by_method.QRIS ?? 0, sh.sales_by_method['E-Wallet'] ?? 0,
          sh.sales_by_method.Card ?? 0, sh.sales_by_method.Credit ?? 0,
          sh.refund_total, sh.sales, sh.trx_count,
          sh.expected_cash ?? '', sh.closing_cash_actual ?? '', sh.cash_variance ?? '',
        ]),
        sumCols: [3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 14, 15],
      },
      {
        name: 'Kasir', title: 'Laporan Kasir', subtitle: `Periode: ${periodLabel} · kinerja per kasir (neto setelah refund)`,
        columns: [
          { header: 'Kasir' }, { header: 'Transaksi' }, { header: 'Item Terjual' },
          { header: 'Bruto (Rp)', money: true }, { header: 'Refund (Rp)', money: true }, { header: 'Neto (Rp)', money: true },
          { header: 'Rata-rata (Rp)', money: true }, { header: 'Jumlah Shift' }, { header: 'Jam Sesi' },
          { header: 'Selisih Kas (Rp)', money: true }, { header: 'Produk Terlaris' },
        ],
        rows: kasirRows.map((r) => [r.name, r.trx, r.items, r.gross, r.refund, r.net, r.avg, r.shiftCount, durationLabelFromMs(r.hoursMs), r.variance, r.topProduct ? `${r.topProduct} (×${r.topProductQty})` : '']),
      },
    ])
  }

  async function exportTaxCSV() {
    setTaxExportErr('')
    try {
      const file = await apiExportTaxReport(selectedTaxMonth, selectedOutlet)
      await downloadBlob(file.filename ?? `laporan-pajak-${selectedTaxMonth}.csv`, file.blob)
    } catch (e) {
      setTaxExportErr(e instanceof Error ? e.message : 'Gagal mengekspor laporan pajak.')
    }
  }

  function taxSection() {
    return (
      <TaxSection
        year={taxYear}
        currentMonth={currentTaxMonth}
        selectedMonth={selectedTaxMonth}
        rows={taxRep.data}
        loading={taxRep.loading}
        error={taxRep.err}
        exportError={taxExportErr}
        onSelectMonth={(month) => { setSelectedTaxMonth(month); setTaxExportErr('') }}
      />
    )
  }

  if (err && !data) return (
    <>
      <PageHeader title="Laporan" crumb="Laporan" actions={(
        <div className="opc-head-actions-stack">
          <OutletFilter outlets={outlets} value={outletId} onChange={setOutletId} />
          <div className="w-full sm:w-72">
            <PeriodPicker options={[...PERIODS, { id: 'year', label: 'Tahun ini' }]} period={period} range={range} onPreset={(id) => { if (id === 'year') { const y = new Date().getFullYear(); setRange({ from: `${y}-01-01`, to: `${y}-12-31` }) } else { setRange(null); setPeriod(id as Period) } }} onRange={(from, to) => setRange({ from, to })} label="Pilih periode laporan" />
          </div>
          <ExportMenu onCSV={exportTabCSV} onExcel={exportWorkbookExcel} />
        </div>
      )} />
      <AlertStrip>{err}</AlertStrip>
      {tab === 'tax' && taxSection()}
    </>
  )

  return (
    <>
      <PageHeader
        title="Laporan"
       
        crumb="Laporan"
        actions={(
          <div className="opc-head-actions-stack">
            <OutletFilter outlets={outlets} value={outletId} onChange={setOutletId} />
            <div className="w-full sm:w-72">
              <PeriodPicker options={[...PERIODS, { id: 'year', label: 'Tahun ini' }]} period={period} range={range} onPreset={(id) => { if (id === 'year') { const y = new Date().getFullYear(); setRange({ from: `${y}-01-01`, to: `${y}-12-31` }) } else { setRange(null); setPeriod(id as Period) } }} onRange={(from, to) => setRange({ from, to })} label="Pilih periode laporan" />
            </div>
            <ExportMenu onCSV={exportTabCSV} onExcel={exportWorkbookExcel} />
          </div>
        )}
      />

      <div className="mb-4">
        <Tabs value={tab} onValueChange={(v) => setTab(v as typeof tab)} className="mt-2.5">
          <TabsList variant="line" aria-label="Jenis laporan">
            {REPORT_TABS.map((t) => (
              <TabsTrigger key={t} value={t}>
                {t === 'sales' ? 'Penjualan' : t === 'products' ? 'Produk' : t === 'profit' ? 'Profit' : t === 'stock' ? 'Stok' : t === 'shift' ? 'Shift' : t === 'tax' ? 'Pajak' : 'Kasir'}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </div>

      {!data ? (
        <LaporanSkeleton />
      ) : (
        <div aria-busy={switching || undefined} className={switching ? 'pointer-events-none opacity-60 transition-opacity' : undefined}>
          {switching && (
            <p className="mb-3 flex items-center gap-2 text-[13px] text-muted" role="status">
              <Loader2 className="size-4 animate-spin" aria-hidden="true" />
              {switchingText}
            </p>
          )}
          {tab === 'sales' && (
            <SalesTab data={data} daily={daily} topProducts={topProducts} prev={prev} banding={banding} animate={animate} needHourly={needHourly} hourly={hourly} setTab={setTab} />
          )}
          {tab === 'products' && (
            <ProductsTab data={data} prodSorted={prodSorted} prodShown={prodShown} prodAll={prodAll} setProdAll={setProdAll} setProdPage={setProdPage} prodPages={prodPages} prodPageSafe={prodPageSafe} prodPageSize={prodPageSize} onProdPageSize={(n) => { setProdPageSize(n); setProdPage(0) }} prodTotal={prodSorted.length} topProducts={topProducts} lowProducts={lowProducts} catMap={catMap} animate={animate} catDonut={catDonut} catData={catRep.data} />
          )}
          {tab === 'profit' && (
            <ProfitTab data={data} profitDaily={profitDaily} profitInsights={profitInsights} prodProfit={prodProfit} profitPages={profitPages} profitPageSafe={profitPageSafe} setProfitPage={setProfitPage} profitPageSize={profitPageSize} onProfitPageSize={(n) => { setProfitPageSize(n); setProfitPage(0) }} profitTotal={prodProfit.length} prev={prev} banding={banding} animate={animate} />
          )}
          {tab === 'stock' && (
            <StockTab data={data} stockRows={stockRows} stockStatus={stockStatus} lowStock={lowStock} catValue={catValue} movements={movRep.data} catMap={catMap} movLoading={movRep.loading} animate={animate} catData={catRep.data} />
          )}
        </div>
      )}
      {tab === 'shift' && (
        <div className="mt-1">
          <Shift embedded outletId={outletId} from={shiftRange?.from} to={shiftRange?.to} />
        </div>
      )}

      {tab === 'cashier' && data && (
        <div className="mt-1">
          <CashierReport data={data} outletId={outletId} period={period} range={range} from={shiftRange?.from} to={shiftRange?.to} />
        </div>
      )}

      {tab === 'tax' && taxSection()}
    </>
  )
}
