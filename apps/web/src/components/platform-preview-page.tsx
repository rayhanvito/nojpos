import Link from 'next/link';

import { ActivityLanes } from '@/components/activity-lanes';
import { DetailLinks } from '@/components/detail-links';
import { PlatformMetric } from '@/components/platform-metric';
import { PlatformMiniChart } from '@/components/platform/platform-mini-chart';
import { PlatformShell } from '@/components/platform-shell';
import { PreviewActionPanel } from '@/components/preview-action-panel';
import { PreviewGate } from '@/components/preview-gate';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewTable } from '@/components/preview-table';
import { SystemHealthPanel } from '@/components/system-health-panel';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { PageHeader } from '@/components/ui/page-header';
import { PreviewNotice } from '@/components/ui/preview-notice';
import type { PreviewActionGroupFixture, PreviewLane, PreviewMetricFixture, PreviewStateFixture, PreviewTableFixture } from '@/fixtures/preview';

type PlatformRiskItem = {
  title: string;
  description: string;
  tone?: 'neutral' | 'warning' | 'success';
};

type PlatformLinkCard = {
  href: string;
  label: string;
  description: string;
};

type PlatformInfoCard = {
  title: string;
  value: string;
  description: string;
};

type PlatformMetricGroup = {
  title: string;
  description: string;
  metrics: readonly PreviewMetricFixture[];
};

type PlatformChartGroup = {
  title: string;
  description: string;
  data: readonly { name: string; value: number }[];
};

type PlatformPreviewPageProps = {
  title: string;
  kicker: string;
  description: string;
  lanes: readonly PreviewLane[];
  metrics: readonly PreviewMetricFixture[];
  metricGroups?: readonly PlatformMetricGroup[];
  actionGroup: PreviewActionGroupFixture;
  states: readonly PreviewStateFixture[];
  table: PreviewTableFixture;
  detailLinks?: readonly { href: string; label: string; description: string }[];
  healthChecks?: readonly { label: string; status: string; detail: string }[];
  riskQueue?: readonly PlatformRiskItem[];
  operationalLinks?: readonly PlatformLinkCard[];
  infoCards?: readonly PlatformInfoCard[];
  charts?: readonly PlatformChartGroup[];
  tableKicker?: string;
  tableTitle?: string;
  tableDescription?: string;
  noticeChildren?: string;
};

function PlatformRiskQueue({ items }: { items: readonly PlatformRiskItem[] }) {
  return (
    <section className="platform-section" aria-labelledby="platform-risk-queue">
      <div className="section-heading compact">
        <div>
          <span className="card-kicker">Antrian perhatian</span>
          <h2 id="platform-risk-queue">Perlu perhatian</h2>
        </div>
        <Badge variant="warning">Hanya lihat</Badge>
      </div>
      <div className="platform-risk-grid">
        {items.map((item) => (
          <article className="platform-risk-card" key={item.title}>
            <span className={item.tone === 'success' ? 'attention-dot success' : item.tone === 'warning' ? 'attention-dot warning' : 'attention-dot'} />
            <div>
              <h3>{item.title}</h3>
              <p>{item.description}</p>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}

function PlatformOperationalLinks({ links }: { links: readonly PlatformLinkCard[] }) {
  return (
    <section className="platform-section" aria-labelledby="platform-operational-summary">
      <div className="section-heading compact">
        <div>
          <span className="card-kicker">Pusat kontrol</span>
          <h2 id="platform-operational-summary">Ringkasan operasional</h2>
        </div>
        <Badge variant="info">Khusus platform</Badge>
      </div>
      <div className="platform-control-grid">
        {links.map((link) => (
          <Link href={link.href} className="platform-control-card" key={link.href}>
            <strong>{link.label}</strong>
            <small>{link.description}</small>
          </Link>
        ))}
      </div>
    </section>
  );
}

function PlatformInfoCards({ cards }: { cards: readonly PlatformInfoCard[] }) {
  return (
    <div className="platform-info-grid" aria-label="Ringkasan platform preview">
      {cards.map((card) => (
        <article className="platform-info-card" key={card.title}>
          <span>{card.title}</span>
          <strong>{card.value}</strong>
          <p>{card.description}</p>
        </article>
      ))}
    </div>
  );
}

function PlatformMetricGroups({ groups }: { groups: readonly PlatformMetricGroup[] }) {
  return (
    <div className="grid gap-5" aria-label="Grouped Super Admin metrics">
      {groups.map((group) => (
        <section className="platform-section" key={group.title} aria-labelledby={`${group.title.replace(/\s+/g, '-').toLowerCase()}-metrics`}>
          <div className="section-heading compact">
            <div>
              <span className="card-kicker">UI preview · agregat contoh</span>
              <h2 id={`${group.title.replace(/\s+/g, '-').toLowerCase()}-metrics`}>{group.title}</h2>
              <p>{group.description}</p>
            </div>
            <Badge variant="outline">Sample data</Badge>
          </div>
          <div className="kpi-grid platform-metric-strip">
            {group.metrics.map((metric) => <PlatformMetric key={`${group.title}-${metric.label}`} {...metric} />)}
          </div>
        </section>
      ))}
    </div>
  );
}

export function PlatformPreviewPage({
  title,
  kicker,
  description,
  lanes,
  metrics,
  metricGroups,
  actionGroup,
  states,
  table,
  detailLinks,
  healthChecks,
  riskQueue,
  operationalLinks,
  infoCards,
  charts,
  tableKicker = 'Data platform contoh',
  tableTitle = title,
  tableDescription = 'Area ini khusus owner/operator NojPOS. Menu platform tetap terpisah dari admin toko dan tidak memakai konteks outlet.',
  noticeChildren = 'Mode preview internal — data platform belum terhubung dan semua aksi platform masih dikunci.',
}: PlatformPreviewPageProps) {
  return (
    <PlatformShell title={title}>
      <div className="platform-page">
        <PageHeader
          kicker={kicker}
          title={title}
          description={description}
          badge="Owner Platform"
          className="platform-hero"
          actions={(
            <div className="platform-header-actions" aria-label="Status platform preview">
              <Badge variant="info">Preview internal</Badge>
              <Button variant="gated" size="sm" disabled>Aksi platform dikunci</Button>
            </div>
          )}
        />
        <PreviewNotice variant="platform" message={noticeChildren} />

        {metricGroups ? (
          <PlatformMetricGroups groups={metricGroups} />
        ) : (
          <div className="kpi-grid platform-metric-strip">
            {metrics.map((metric) => <PlatformMetric key={metric.label} {...metric} />)}
          </div>
        )}

        {riskQueue ? <PlatformRiskQueue items={riskQueue} /> : null}
        {infoCards ? <PlatformInfoCards cards={infoCards} /> : null}
        {charts ? (
          <section className="platform-section" aria-labelledby="platform-chart-summary">
            <div className="section-heading compact">
              <div>
                <span className="card-kicker">Agregat contoh</span>
                <h2 id="platform-chart-summary">Status ringkas</h2>
                <p>Grafik kecil membantu tim internal membaca pola tanpa membuka detail data toko.</p>
              </div>
              <Badge variant="info">Recharts preview</Badge>
            </div>
            <div className="platform-chart-grid">
              {charts.map((chart) => <PlatformMiniChart key={chart.title} title={chart.title} description={chart.description} data={chart.data} />)}
            </div>
          </section>
        ) : null}
        {operationalLinks ? <PlatformOperationalLinks links={operationalLinks} /> : null}

        <ActivityLanes title="Ringkasan kerja platform" lanes={lanes} />

        <div className="split-grid platform-content-grid">
          <section className="resource card platform-data-card">
            <div className="card-header">
              <div>
                <span className="card-kicker">{tableKicker}</span>
                <h2>{tableTitle}</h2>
              </div>
              <Badge variant="neutral">Hanya lihat</Badge>
            </div>
            <p>{tableDescription}</p>
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
            {healthChecks ? <SystemHealthPanel checks={healthChecks} /> : null}
          </div>
        </div>

        <section className="preview-notes-section platform-preview-notes" aria-labelledby={`${title}-mode-notes`}>
          <div className="preview-notes-header">
            <span className="card-kicker">Catatan mode internal</span>
            <h2 id={`${title}-mode-notes`}>Kontrol platform tetap aman</h2>
            <p>Informasi status, aksi, dan batasan ditempatkan di bawah konten utama agar halaman tetap terasa seperti pusat kontrol platform tanpa mengaktifkan aksi sensitif.</p>
          </div>
          <PreviewActionPanel group={actionGroup} compact />
          <PreviewStateBoard title={`Kondisi tampilan ${title}`} description="Loading, kosong, error, hak akses, dan fitur belum tersedia tetap disiapkan khusus untuk Platform Super Admin." states={states} />
          {detailLinks ? <DetailLinks title={`Detail ${title}`} links={detailLinks} /> : null}
          <PreviewGate title="Koneksi platform belum aktif" />
        </section>
      </div>
    </PlatformShell>
  );
}
