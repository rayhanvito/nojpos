import { ActivityLanes } from '@/components/activity-lanes';
import { AdminShell } from '@/components/admin-shell';
import { DetailLinks } from '@/components/detail-links';
import { PreviewActionPanel } from '@/components/preview-action-panel';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewTable } from '@/components/preview-table';
import { previewActionGroups, previewStateMatrix, promotionDetailLinks, promotionLanes, promotionTable } from '@/fixtures/preview';

export default function PromotionsPage() {
  return (
    <AdminShell title="Promosi">
      <ActivityLanes title="Preview pengelolaan promosi" lanes={promotionLanes} />
      <PreviewActionPanel group={previewActionGroups.promotions} />
      <PreviewStateBoard
        title="Kondisi tampilan promosi"
        description="Campaign, aturan, publikasi, dan validasi disiapkan tanpa menghitung diskon atau total di browser."
        states={previewStateMatrix.promotions}
      />
      <DetailLinks title="Preview detail promosi" links={promotionDetailLinks} />
      <section className="resource card">
        <div className="card-header">
          <div>
            <span className="card-kicker">Manajemen promosi</span>
            <h2>Daftar promosi</h2>
          </div>
          <span className="badge warning">Menunggu backend</span>
        </div>
        <p>Preview tidak menghitung quote, diskon, pajak, service, rounding, atau total.</p>
        <PreviewTable columns={promotionTable.columns} rows={promotionTable.rows} />
      </section>
    </AdminShell>
  );
}
