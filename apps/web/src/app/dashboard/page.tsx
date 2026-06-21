import Link from 'next/link';
import {
  AlertTriangle,
  Banknote,
  BarChart3,
  Boxes,
  Building2,
  Package,
  ReceiptText,
  ShoppingCart,
  TrendingUp,
  Users,
  WalletCards,
} from 'lucide-react';

import { DashboardPaymentChart } from '@/components/dashboard/dashboard-payment-chart';
import { DashboardSalesChart } from '@/components/dashboard/dashboard-sales-chart';
import { AdminShell } from '@/components/admin-shell';
import { Badge } from '@/components/ui/badge';
import { dashboardQuickLinks } from '@/fixtures/preview/dashboard';
import { getDashboardPageModel } from '@/lib/server/dashboard-summary';

export const dynamic = 'force-dynamic';

const kpiIcons = [WalletCards, ReceiptText, ShoppingCart, TrendingUp, Boxes, Banknote] as const;

const quickNavIcons = {
  '/transactions': ReceiptText,
  '/inventory': Boxes,
  '/catalog': Package,
  '/reports': BarChart3,
  '/outlets': Building2,
  '/staff': Users,
} as const;

export default async function DashboardPage() {
  const model = await getDashboardPageModel();
  const dashboardData = model.data;
  const isRealData = model.state === 'real';
  const hasAlerts = dashboardData.alerts.length > 0;
  const hasTopProducts = dashboardData.topProducts.length > 0;
  const hasLowStock = dashboardData.lowStockItems.length > 0;
  const hasRecentTransactions = dashboardData.recentTransactions.length > 0;
  const hasCashierPerformance = dashboardData.cashierPerformance.length > 0;
  const hasBranchHighlights = dashboardData.branchHighlights.length > 0;

  return (
    <AdminShell title="Dashboard">
      <main className="owner-dashboard dashboard-overview" aria-labelledby="owner-dashboard-title">
        <section className="owner-dashboard-header card">
          <div>
            <span className="card-kicker">Owner/Admin toko</span>
            <h1 id="owner-dashboard-title">{model.title}</h1>
            <p>{model.description}</p>
            <small>{model.detail}</small>
          </div>
          <div className="dashboard-context-chips" aria-label="Konteks dashboard">
            {model.contextBadges.map((badge) => (
              <Badge variant={badge.tone} key={`${badge.label}-${badge.tone}`}>{badge.label}</Badge>
            ))}
          </div>
        </section>

        {model.state !== 'real' ? (
          <section className={`dashboard-safety-note ${model.state === 'forbidden' ? 'danger' : 'warning'}`} aria-labelledby="dashboard-state-title">
            <span className="card-kicker">Status integrasi</span>
            <h2 id="dashboard-state-title">{model.sourceLabel}</h2>
            <p>{model.detail}</p>
            <p>Data yang tampil setelah pesan ini adalah data contoh fallback, bukan data produksi dari backend.</p>
          </section>
        ) : null}

        <section className="dashboard-section" aria-labelledby="dashboard-kpi-title">
          <div className="section-heading compact">
            <div>
              <span className="card-kicker">KPI utama</span>
              <h2 id="dashboard-kpi-title">Yang paling penting hari ini</h2>
            </div>
            <Badge variant={model.sourceTone}>{model.sourceLabel}</Badge>
          </div>
          <div className="dashboard-kpi-grid owner-kpi-grid">
            {dashboardData.kpiCards.map((metric, index) => {
              const Icon = kpiIcons[index] ?? WalletCards;

              return (
                <article className={`owner-kpi-card ${metric.tone}`} key={metric.label}>
                  <div className="owner-kpi-icon" aria-hidden="true">
                    <Icon />
                  </div>
                  <div className="owner-kpi-copy">
                    <span>{metric.label}</span>
                    <strong>{metric.value}</strong>
                    <p>{metric.helper}</p>
                  </div>
                  <small>{metric.trend}</small>
                </article>
              );
            })}
          </div>
        </section>

        <section className="dashboard-section" aria-labelledby="dashboard-alert-title">
          <div className="section-heading compact">
            <div>
              <span className="card-kicker">Perlu dicek</span>
              <h2 id="dashboard-alert-title">Alert penting</h2>
            </div>
            <Badge variant={hasAlerts ? 'warning' : 'success'}>{hasAlerts ? 'Ada catatan' : 'Aman'}</Badge>
          </div>
          <div className="dashboard-alert-list">
            {hasAlerts ? dashboardData.alerts.map((alert) => (
              <article className={`dashboard-alert-card ${alert.tone}`} key={alert.title}>
                <span className="dashboard-alert-icon" aria-hidden="true">
                  <AlertTriangle />
                </span>
                <div>
                  <h3>{alert.title}</h3>
                  <p>{alert.description}</p>
                </div>
              </article>
            )) : (
              <article className="dashboard-alert-card success">
                <span className="dashboard-alert-icon" aria-hidden="true">
                  <AlertTriangle />
                </span>
                <div>
                  <h3>Tidak ada alert penting</h3>
                  <p>{isRealData ? 'Backend tidak mengirim alert untuk tanggal ini.' : 'Fallback preview tetap siap menampilkan alert saat tersedia.'}</p>
                </div>
              </article>
            )}
          </div>
        </section>

        <section className="dashboard-chart-grid" aria-labelledby="dashboard-chart-title">
          <div className="section-heading dashboard-chart-heading">
            <div>
              <span className="card-kicker">Grafik sederhana</span>
              <h2 id="dashboard-chart-title">Tren penjualan dan pembayaran</h2>
            </div>
            <Badge variant={model.sourceTone}>{isRealData ? 'Read-only backend' : 'Fallback preview'}</Badge>
          </div>

          <article className="dashboard-chart-card sales-card">
            <div className="card-header compact">
              <div>
                <span className="card-kicker">Penjualan</span>
                <h3>Penjualan 7 hari terakhir</h3>
              </div>
              <Badge variant={model.sourceTone}>{isRealData ? 'Real' : 'Contoh'}</Badge>
            </div>
            <DashboardSalesChart data={dashboardData.salesLast7Days} />
            <p>{isRealData ? 'Grafik berasal dari endpoint dashboard summary backend.' : 'Grafik memakai data contoh lokal, bukan transaksi produksi.'}</p>
          </article>

          <article className="dashboard-chart-card payment-card">
            <div className="card-header compact">
              <div>
                <span className="card-kicker">Pembayaran</span>
                <h3>Metode pembayaran</h3>
              </div>
              <Badge variant={model.sourceTone}>{isRealData ? 'Real' : 'Contoh'}</Badge>
            </div>
            <DashboardPaymentChart data={dashboardData.paymentMethods} />
            <p>{isRealData ? 'Komposisi pembayaran dibaca dari backend tanpa detail referensi pembayaran.' : 'Komposisi pembayaran fallback diberi label preview.'}</p>
          </article>
        </section>

        <section className="dashboard-section" aria-labelledby="dashboard-ops-title">
          <div className="section-heading compact">
            <div>
              <span className="card-kicker">Operasional toko</span>
              <h2 id="dashboard-ops-title">Ringkasan yang perlu dilihat</h2>
            </div>
            <Badge variant={model.sourceTone}>{isRealData ? 'Backend' : 'Preview'}</Badge>
          </div>

          <div className="dashboard-ops-grid">
            <article className="dashboard-ops-card">
              <div className="card-header compact">
                <div>
                  <span className="card-kicker">Produk</span>
                  <h3>Produk terlaris</h3>
                </div>
                <Badge variant={hasTopProducts ? 'success' : 'neutral'}>{hasTopProducts ? `${dashboardData.topProducts.length} item` : 'Kosong'}</Badge>
              </div>
              <div className="dashboard-product-list">
                {hasTopProducts ? dashboardData.topProducts.map((product) => (
                  <div className="dashboard-product-row" key={product.name}>
                    <div>
                      <strong>{product.name}</strong>
                      <span>{product.quantity} · {product.total}</span>
                    </div>
                    <div className="dashboard-product-bar" aria-hidden="true">
                      <span style={{ width: `${product.share}%` }} />
                    </div>
                  </div>
                )) : <p>Belum ada produk terjual pada periode ini.</p>}
              </div>
            </article>

            <article className="dashboard-ops-card">
              <div className="card-header compact">
                <div>
                  <span className="card-kicker">Inventori</span>
                  <h3>Stok kritis</h3>
                </div>
                <Badge variant={hasLowStock ? 'warning' : 'success'}>{hasLowStock ? 'Cek stok' : 'Aman'}</Badge>
              </div>
              <div className="dashboard-compact-list">
                {hasLowStock ? dashboardData.lowStockItems.map((item) => (
                  <div className="dashboard-list-row" key={item.name}>
                    <div>
                      <strong>{item.name}</strong>
                      <span>Sisa {item.stock}</span>
                    </div>
                    <Badge variant={item.status === 'Habis' ? 'danger' : 'warning'}>{item.status}</Badge>
                  </div>
                )) : <p>Tidak ada stok kritis dari data backend.</p>}
              </div>
            </article>

            <article className="dashboard-ops-card">
              <div className="card-header compact">
                <div>
                  <span className="card-kicker">Transaksi</span>
                  <h3>Transaksi terbaru</h3>
                </div>
                <Badge variant={hasRecentTransactions ? 'neutral' : 'warning'}>{hasRecentTransactions ? '5 terakhir' : 'Kosong'}</Badge>
              </div>
              <div className="dashboard-compact-list">
                {hasRecentTransactions ? dashboardData.recentTransactions.map((transaction) => (
                  <div className="dashboard-list-row" key={transaction.code}>
                    <div>
                      <strong>{transaction.code}</strong>
                      <span>{transaction.time} · {transaction.method}</span>
                    </div>
                    <b>{transaction.total}</b>
                  </div>
                )) : <p>Belum ada transaksi terbaru untuk tanggal ini.</p>}
              </div>
            </article>

            <article className="dashboard-ops-card">
              <div className="card-header compact">
                <div>
                  <span className="card-kicker">Kasir</span>
                  <h3>Performa kasir</h3>
                </div>
                <Badge variant={hasCashierPerformance ? 'neutral' : 'warning'}>{hasCashierPerformance ? 'Read-only' : 'Kosong'}</Badge>
              </div>
              <div className="dashboard-compact-list">
                {hasCashierPerformance ? dashboardData.cashierPerformance.map((cashier) => (
                  <div className="dashboard-list-row stacked" key={cashier.name}>
                    <div>
                      <strong>{cashier.name}</strong>
                      <span>{cashier.transactions} · {cashier.total}</span>
                    </div>
                    <small>{cashier.note}</small>
                  </div>
                )) : <p>Belum ada performa kasir untuk tanggal ini.</p>}
              </div>
            </article>

            <article className="dashboard-ops-card wide">
              <div className="card-header compact">
                <div>
                  <span className="card-kicker">Cabang</span>
                  <h3>Cabang yang perlu dicek</h3>
                </div>
                <Badge variant={hasBranchHighlights ? 'neutral' : 'success'}>{hasBranchHighlights ? 'Semua cabang' : 'Tidak ada catatan'}</Badge>
              </div>
              <div className="dashboard-branch-list">
                {hasBranchHighlights ? dashboardData.branchHighlights.map((branch) => (
                  <div className={`dashboard-branch-card ${branch.tone}`} key={branch.name}>
                    <strong>{branch.name}</strong>
                    <span>{branch.status}</span>
                    <p>{branch.summary}</p>
                  </div>
                )) : <p>Tidak ada cabang yang perlu dicek dari data backend.</p>}
              </div>
            </article>
          </div>
        </section>

        <section className="dashboard-section" aria-labelledby="dashboard-quick-title">
          <div className="section-heading compact">
            <div>
              <span className="card-kicker">Lanjut cek detail</span>
              <h2 id="dashboard-quick-title">Quick links</h2>
            </div>
            <Badge variant="neutral">Link read-only</Badge>
          </div>
          <div className="quick-nav-grid dashboard-quick-links">
            {dashboardQuickLinks.map((resource) => {
              const Icon = quickNavIcons[resource.path as keyof typeof quickNavIcons] ?? Package;

              return (
                <Link className="quick-nav-card" href={resource.path} key={resource.path}>
                  <span className="quick-nav-icon" aria-hidden="true"><Icon /></span>
                  <span>
                    <strong>{resource.label}</strong>
                    <small>Buka halaman admin read-only/preview</small>
                  </span>
                </Link>
              );
            })}
          </div>
        </section>

        <section className="dashboard-safety-note" aria-labelledby="dashboard-safety-title">
          <span className="card-kicker">Read-only safety note</span>
          <h2 id="dashboard-safety-title">Dashboard ini tidak menjalankan aksi sensitif</h2>
          {dashboardData.dataNotes.map((note) => <p key={note}>{note}</p>)}
          <p>Refund, void, reprint, export, shift close/open, stock adjustment, payment retry, dan cash reconciliation tidak dijalankan di halaman ini.</p>
        </section>
      </main>
    </AdminShell>
  );
}
