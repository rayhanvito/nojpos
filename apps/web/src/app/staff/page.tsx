import { ActivityLanes } from '@/components/activity-lanes';
import { AdminShell } from '@/components/admin-shell';
import { DetailLinks } from '@/components/detail-links';
import { PreviewActionPanel } from '@/components/preview-action-panel';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewTable } from '@/components/preview-table';
import { PreviewToolbar } from '@/components/preview-toolbar';
import { previewActionGroups, previewStateMatrix, staffDetailLinks, staffDirectoryTable, staffLanes } from '@/fixtures/preview';

export default function StaffPage() {
  return (
    <AdminShell title="Manajemen Staf">
      <PreviewToolbar action="Undang staf preview" />
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
        <p>Preview hanya lihat. Peran, outlet, dan status aktif akan mengikuti otorisasi backend.</p>
        <PreviewTable columns={staffDirectoryTable.columns} rows={staffDirectoryTable.rows} />
      </section>
    </AdminShell>
  );
}
