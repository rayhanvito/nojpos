import { ActivityLanes } from '@/components/activity-lanes';
import { AdminShell } from '@/components/admin-shell';
import { PreviewActionPanel } from '@/components/preview-action-panel';
import { PreviewGate } from '@/components/preview-gate';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewToolbar } from '@/components/preview-toolbar';
import { ReadonlyResource } from '@/components/readonly-resource';
import { previewActionGroups, previewStateMatrix, reportLanes, reportResources } from '@/fixtures/preview';

export default function ReportsPage() {
  return (
    <AdminShell title="Laporan">
      <PreviewToolbar />
      <div className="hero-card">
        <div className="hero-head">
          <div>
            <span className="card-kicker">Pusat laporan</span>
            <h2>Filter, ekspor, dan periode masih menunggu backend</h2>
          </div>
          <span className="badge neutral">Hanya lihat</span>
        </div>
        <p>
          Desain mengikuti pola laporan dan tabel NojPOS, tetapi angka, ekspor, pagination, dan visibilitas peran
          menunggu kontrak backend agar tidak ada hasil kalkulasi palsu di browser.
        </p>
      </div>
      <ActivityLanes title="Status preview laporan" lanes={reportLanes} />
      <PreviewActionPanel group={previewActionGroups.reports} />
      <PreviewStateBoard
        title="Kondisi tampilan laporan"
        description="Laporan membutuhkan status eksplisit karena filter, periode, ekspor, pagination, dan kebijakan backend semuanya berpengaruh."
        states={previewStateMatrix.reports}
      />
      <div className="split-grid">
        <div className="panel-grid compact-panel-grid">
          {reportResources.map((resource) => <ReadonlyResource key={resource.path} {...resource} />)}
        </div>
        <PreviewGate title="Data laporan belum aktif" />
      </div>
    </AdminShell>
  );
}
