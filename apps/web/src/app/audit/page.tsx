import { ActivityLanes } from '@/components/activity-lanes';
import { AdminShell } from '@/components/admin-shell';
import { DetailLinks } from '@/components/detail-links';
import { PreviewActionPanel } from '@/components/preview-action-panel';
import { PreviewGate } from '@/components/preview-gate';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewTable } from '@/components/preview-table';
import { auditDetailLinks, auditLanes, auditTrailTable, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function AuditPage() {
  return (
    <AdminShell title="Audit">
      <ActivityLanes title="Ringkasan audit" lanes={auditLanes} />
      <PreviewActionPanel group={previewActionGroups.audit} />
      <PreviewStateBoard
        title="Kondisi tampilan audit"
        description="Jejak audit wajib berasal dari server, jadi semua status disiapkan tanpa membuat log palsu di browser."
        states={previewStateMatrix.audit}
      />
      <DetailLinks title="Preview detail audit" links={auditDetailLinks} />
      <div className="split-grid">
        <section className="resource card">
          <div className="card-header">
            <div>
              <span className="card-kicker">Audit trail</span>
              <h2>Log aktivitas</h2>
            </div>
            <span className="badge warning">Menunggu backend</span>
          </div>
          <p>Preview ini sengaja tidak membuat data audit palsu. Log asli wajib berasal dari backend.</p>
          <PreviewTable columns={auditTrailTable.columns} rows={auditTrailTable.rows} />
        </section>
        <PreviewGate title="Log audit belum aktif" />
      </div>
    </AdminShell>
  );
}
