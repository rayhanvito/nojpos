import { ActivityLanes } from '@/components/activity-lanes';
import { AdminShell } from '@/components/admin-shell';
import { DetailLinks } from '@/components/detail-links';
import { PaymentStatusCard } from '@/components/payment-status-card';
import { PreviewActionPanel } from '@/components/preview-action-panel';
import { PreviewGate } from '@/components/preview-gate';
import { PreviewMetric } from '@/components/preview-metric';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewTable } from '@/components/preview-table';
import { SubscriptionStatusCard } from '@/components/subscription-status-card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { PageHeader } from '@/components/ui/page-header';
import { PreviewNotice } from '@/components/ui/preview-notice';
import type { DetailPreviewFixture, PreviewActionGroupFixture, PreviewLane, PreviewMetricFixture, PreviewStateFixture, PreviewTableFixture } from '@/fixtures/preview';

type TenantPreviewPageProps = {
  title: string;
  kicker: string;
  description: string;
  action: string;
  lanes: readonly PreviewLane[];
  metrics: readonly PreviewMetricFixture[];
  actionGroup: PreviewActionGroupFixture;
  states: readonly PreviewStateFixture[];
  table: PreviewTableFixture;
  detailLinks?: readonly { href: string; label: string; description: string }[];
  subscription?: { plan: string; status: string; renewal: string; note: string };
  payment?: { label: string; status: string; description: string };
};

export function TenantPreviewPage({ title, kicker, description, action, lanes, metrics, actionGroup, states, table, detailLinks, subscription, payment }: TenantPreviewPageProps) {
  return (
    <AdminShell title={title}>
      <PageHeader
        kicker={kicker}
        title={title}
        description={`${description} Alurnya dibuat sederhana: cek ringkasan, pantau daftar, lalu buka detail saat diperlukan.`}
        badge="Contoh UI"
        actions={<Button variant="gated" size="sm" disabled>{action} dikunci</Button>}
      />
      <PreviewNotice />

      <div className="kpi-grid">
        {metrics.map((metric) => <PreviewMetric key={metric.label} {...metric} />)}
      </div>

      <ActivityLanes title={`${title} workspace`} lanes={lanes} />

      <div className="split-grid">
        <section className="resource card">
          <div className="card-header">
            <div>
              <span className="card-kicker">Daftar contoh</span>
              <h2>{title}</h2>
            </div>
            <Badge variant="neutral">Hanya lihat</Badge>
          </div>
          <p>Data di bawah hanya untuk mengecek layout. Tidak ada data produksi, simpan, ubah, hapus, refund, atau export aktif.</p>
          <PreviewTable
            columns={table.columns}
            rows={table.rows}
            caption={table.caption}
            primaryColumn={table.primaryColumn}
            statusColumn={table.statusColumn}
            metaColumns={table.metaColumns}
            numericColumns={table.numericColumns}
          />
        </section>
        <div className="stack">
          {subscription ? <SubscriptionStatusCard {...subscription} /> : null}
          {payment ? <PaymentStatusCard {...payment} /> : null}
        </div>
      </div>

      <section className="preview-notes-section" aria-labelledby={`${title}-mode-notes`}>
        <div className="preview-notes-header">
          <span className="card-kicker">Catatan mode contoh</span>
          <h2 id={`${title}-mode-notes`}>Keamanan preview tetap aktif</h2>
          <p>Bagian ini disimpan di bawah konten utama supaya user tetap melihat informasi bisnis lebih dulu, sementara tim masih bisa mengecek kondisi UI dan aksi yang dikunci.</p>
        </div>
        <PreviewActionPanel group={actionGroup} compact />
        <PreviewStateBoard title={`Kondisi tampilan ${title}`} description="Loading, kosong, error, hak akses, dan fitur belum tersedia tetap disiapkan sebelum koneksi server aktif." states={states} />
        {detailLinks ? <DetailLinks title={`Detail ${title}`} links={detailLinks} /> : null}
        <PreviewGate />
      </section>
    </AdminShell>
  );
}

export function getDetailParams(details: readonly DetailPreviewFixture[], key: string) {
  return details.map((detail) => ({ [key]: detail.id }));
}
