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
import {
  dashboardAlerts,
  dashboardBranchHighlights,
  dashboardCashierPerformance,
  dashboardKpiCards,
  dashboardLowStockItems,
  dashboardPaymentMethods,
  dashboardQuickLinks,
  dashboardRecentTransactions,
  dashboardSalesLast7Days,
  dashboardTopProducts,
} from '@/fixtures/preview/dashboard';

const kpiIcons = [WalletCards, ReceiptText, ShoppingCart, TrendingUp, Boxes, Banknote] as const;

const quickNavIcons = {
  '/transactions': ReceiptText,
  '/inventory': Boxes,
  '/catalog': Package,
  '/reports': BarChart3,
  '/outlets': Building2,
  '/staff': Users,
} as const;

export default function DashboardPage() {
  return (
    <AdminShell title="Dashboard">
      <main className="owner-dashboard dashboard-overview" aria-labelledby="owner-dashboard-title">
        <section className="owner-dashboard-header card">
          <div>
            <span className="card-kicker">Owner/Admin toko</span>
            <h1 id="owner-dashboard-title">Ringkasan toko hari ini</h1>
            <p>Pantau penjualan, stok, kas, dan aktivitas toko dari satu tampilan.</p>
            <small>Data contoh untuk preview tampilan. Belum terhubung ke server.</small>
          </div>
          <div className="dashboard-context-chips" aria-label="Konteks dashboard">
            <Badge variant="neutral">Hari ini</Badge>
            <Badge variant="neutral">Semua cabang</Badge>
            <Badge variant="warning">Preview UI</Badge>
          </div>
        </section>

        <section className="dashboard-section" aria-labelledby="dashboard-kpi-title">
          <div className="section-heading compact">
            <div>
              <span className="card-kicker">KPI utama</span>
              <h2 id="dashboard-kpi-title">Yang paling penting hari ini</h2>
            </div>
            <Badge variant="neutral">Data mock lokal</Badge>
          </div>
          <div className="dashboard-kpi-grid owner-kpi-grid">
            {dashboardKpiCards.map((metric, index) => {
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
            <Badge variant="neutral">Preview ramah</Badge>
          </div>
          <div className="dashboard-alert-list">
            {dashboardAlerts.map((alert) => (
              <article className={`dashboard-alert-card ${alert.tone}`} key={alert.title}>
                <span className="dashboard-alert-icon" aria-hidden="true">
                  <AlertTriangle />
                </span>
                <div>
                  <h3>{alert.title}</h3>
                  <p>{alert.description}</p>
                </div>
              </article>
            ))}
          </div>
        </section>

        <section className="dashboard-chart-grid" aria-labelledby="dashboard-chart-title">
          <div className="section-heading dashboard-chart-heading">
            <div>
              <span className="card-kicker">Grafik sederhana</span>
              <h2 id="dashboard-chart-title">Tren penjualan dan pembayaran</h2>
            </div>
            <Badge variant="neutral">Recharts preview</Badge>
          </div>

          <article className="dashboard-chart-card sales-card">
            <div className="card-header compact">
              <div>
                <span className="card-kicker">Penjualan</span>
                <h3>Penjualan 7 hari terakhir</h3>
              </div>
              <Badge variant="success">Contoh</Badge>
            </div>
            <DashboardSalesChart data={dashboardSalesLast7Days} />
            <p>Grafik memakai data contoh lokal, bukan transaksi produksi.</p>
          </article>

          <article className="dashboard-chart-card payment-card">
            <div className="card-header compact">
              <div>
                <span className="card-kicker">Pembayaran</span>
                <h3>Metode pembayaran</h3>
              </div>
              <Badge variant="neutral">Mock</Badge>
            </div>
            <DashboardPaymentChart data={dashboardPaymentMethods} />
            <p>Komposisi pembayaran akan mengikuti backend setelah integrasi aktif.</p>
          </article>
        </section>

        <section className="dashboard-section" aria-labelledby="dashboard-ops-title">
          <div className="section-heading compact">
            <div>
              <span className="card-kicker">Operasional toko</span>
              <h2 id="dashboard-ops-title">Ringkasan yang perlu dilihat</h2>
            </div>
            <Badge variant="neutral">Compact list</Badge>
          </div>

          <div className="dashboard-ops-grid">
            <article className="dashboard-ops-card">
              <div className="card-header compact">
                <div>
                  <span className="card-kicker">Produk</span>
                  <h3>Produk terlaris</h3>
                </div>
                <Badge variant="success">5 item</Badge>
              </div>
              <div className="dashboard-product-list">
                {dashboardTopProducts.map((product) => (
                  <div className="dashboard-product-row" key={product.name}>
                    <div>
                      <strong>{product.name}</strong>
                      <span>{product.quantity} · {product.total}</span>
                    </div>
                    <div className="dashboard-product-bar" aria-hidden="true">
                      <span style={{ width: `${product.share}%` }} />
                    </div>
                  </div>
                ))}
              </div>
            </article>

            <article className="dashboard-ops-card">
              <div className="card-header compact">
                <div>
                  <span className="card-kicker">Inventori</span>
                  <h3>Stok kritis</h3>
                </div>
                <Badge variant="warning">Cek stok</Badge>
              </div>
              <div className="dashboard-compact-list">
                {dashboardLowStockItems.map((item) => (
                  <div className="dashboard-list-row" key={item.name}>
                    <div>
                      <strong>{item.name}</strong>
                      <span>Sisa {item.stock}</span>
                    </div>
                    <Badge variant={item.status === 'Habis' ? 'danger' : 'warning'}>{item.status}</Badge>
                  </div>
                ))}
              </div>
            </article>

            <article className="dashboard-ops-card">
              <div className="card-header compact">
                <div>
                  <span className="card-kicker">Transaksi</span>
                  <h3>Transaksi terbaru</h3>
                </div>
                <Badge variant="neutral">5 terakhir</Badge>
              </div>
              <div className="dashboard-compact-list">
                {dashboardRecentTransactions.map((transaction) => (
                  <div className="dashboard-list-row" key={transaction.code}>
                    <div>
                      <strong>{transaction.code}</strong>
                      <span>{transaction.time} · {transaction.method}</span>
                    </div>
                    <b>{transaction.total}</b>
                  </div>
                ))}
              </div>
            </article>

            <article className="dashboard-ops-card">
              <div className="card-header compact">
                <div>
                  <span className="card-kicker">Kasir</span>
                  <h3>Performa kasir</h3>
                </div>
                <Badge variant="neutral">Contoh</Badge>
              </div>
              <div className="dashboard-compact-list">
                {dashboardCashierPerformance.map((cashier) => (
                  <div className="dashboard-list-row stacked" key={cashier.name}>
                    <div>
                      <strong>{cashier.name}</strong>
                      <span>{cashier.transactions} · {cashier.total}</span>
                    </div>
                    <small>{cashier.note}</small>
                  </div>
                ))}
              </div>
            </article>

            <article className="dashboard-ops-card wide">
              <div className="card-header compact">
                <div>
                  <span className="card-kicker">Cabang</span>
                  <h3>Cabang yang perlu dicek</h3>
                </div>
                <Badge variant="neutral">Semua cabang</Badge>
              </div>
              <div className="dashboard-branch-list">
                {dashboardBranchHighlights.map((branch) => (
                  <div className={`dashboard-branch-card ${branch.tone}`} key={branch.name}>
                    <strong>{branch.name}</strong>
                    <span>{branch.status}</span>
                    <p>{branch.summary}</p>
                  </div>
                ))}
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
            <Badge variant="neutral">Link statis</Badge>
          </div>
          <div className="quick-nav-grid dashboard-quick-links">
            {dashboardQuickLinks.map((resource) => {
              const Icon = quickNavIcons[resource.path as keyof typeof quickNavIcons] ?? Package;

              return (
                <Link className="quick-nav-card" href={resource.path} key={resource.path}>
                  <span className="quick-nav-icon" aria-hidden="true"><Icon /></span>
                  <span>
                    <strong>{resource.label}</strong>
                    <small>Buka halaman preview</small>
                  </span>
                </Link>
              );
            })}
          </div>
        </section>

        <section className="dashboard-safety-note" aria-labelledby="dashboard-safety-title">
          <span className="card-kicker">Preview safety note</span>
          <h2 id="dashboard-safety-title">Dashboard ini masih preview UI</h2>
          <p>Data, grafik, dan alert menggunakan contoh lokal. Perhitungan final seperti laba, stok, kas, dan laporan harus berasal dari backend saat integrasi aktif.</p>
          <p>Refund, void, reprint, export, shift close/open, stock adjustment, payment retry, dan cash reconciliation tidak dijalankan di halaman ini.</p>
        </section>
      </main>
    </AdminShell>
  );
}
