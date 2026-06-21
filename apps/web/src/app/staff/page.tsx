import { ActivityLanes } from '@/components/activity-lanes';
import { AdminShell } from '@/components/admin-shell';
import { DetailLinks } from '@/components/detail-links';
import { PreviewActionPanel } from '@/components/preview-action-panel';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewTable } from '@/components/preview-table';
import { PreviewToolbar } from '@/components/preview-toolbar';
import { PreviewMetric } from '@/components/preview-metric';
import { previewActionGroups, previewStateMatrix, staffDetailLinks, staffLanes } from '@/fixtures/preview';
import { getStaffPageModel } from '@/lib/server/staff';

export default async function StaffPage() {
  const model = await getStaffPageModel();

  return (
    <AdminShell title="Manajemen Staf">
      <PreviewToolbar action="Undang staf preview" />
      <div className="hero-card">
        <div className="hero-head">
          <div>
            <span className="card-kicker">Direktori user</span>
            <h2>{model.title}</h2>
          </div>
          <span className={`badge ${model.sourceTone}`}>{model.sourceLabel}</span>
        </div>
        <p>{model.description}</p>
      </div>
      <section className="metric-grid" aria-label="Ringkasan staf">
        {model.metrics.map((metric) => <PreviewMetric key={metric.label} {...metric} />)}
      </section>
      <ActivityLanes title="Status direktori staf" lanes={staffLanes} />
      <PreviewActionPanel group={previewActionGroups.staff} />
      <PreviewStateBoard
        title="Kondisi tampilan staf"
        description="Direktori, peran, undangan, dan jejak aktivitas staf disiapkan sebagai UI hanya lihat sebelum autentikasi browser dibuka."
        states={previewStateMatrix.staff}
      />
      <DetailLinks title="Preview detail staf" links={staffDetailLinks} />
      <section className="resource card">
        <div className="card-header">
          <div>
            <span className="card-kicker">Direktori user</span>
            <h2>Daftar staf</h2>
          </div>
          <span className="badge neutral">Hanya lihat</span>
        </div>
        <p>{model.dataNotes[0] ?? 'Preview hanya lihat. Peran, outlet, dan status aktif akan mengikuti otorisasi backend.'}</p>
        <PreviewTable columns={model.table.columns} rows={model.table.rows} caption={model.table.caption} />
      </section>
    </AdminShell>
  );
}
