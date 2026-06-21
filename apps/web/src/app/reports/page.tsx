import { ActivityLanes } from '@/components/activity-lanes';
import { AdminShell } from '@/components/admin-shell';
import { PreviewActionPanel } from '@/components/preview-action-panel';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewToolbar } from '@/components/preview-toolbar';
import { previewActionGroups, previewStateMatrix, reportLanes } from '@/fixtures/preview';
import { getReportsPageModel } from '@/lib/server/reports';

export default async function ReportsPage() {
  const model = await getReportsPageModel();

  return (
    <AdminShell title="Laporan">
      <PreviewToolbar />
      <div className="hero-card">
        <div className="hero-head">
          <div>
            <span className="card-kicker">Pusat laporan</span>
            <h2>{model.title}</h2>
          </div>
          <span className={`badge ${model.sourceTone}`}>{model.sourceLabel}</span>
        </div>
        <p>{model.description}</p>
        <p className="helper-copy">Laporan ini read-only. Export, refund, void, reprint, rekonsiliasi kas, dan shift action tetap disabled.</p>
      </div>

      <section className="metric-grid" aria-label="Ringkasan laporan">
        {model.metrics.map((metric) => (
          <article className="metric-card" key={metric.label}>
            <span className="metric-label">{metric.label}</span>
            <strong>{metric.value}</strong>
            <small>{metric.description}</small>
          </article>
        ))}
      </section>

      <div className="data-panel">
        <div className="panel-heading">
          <div>
            <span className="card-kicker">Ringkasan read-only</span>
            <h3>Data laporan utama</h3>
          </div>
          <span className="badge neutral">Server-side</span>
        </div>
        <div className="table-shell">
          <table>
            <thead>
              <tr>
                <th>Bagian</th>
                <th>Jumlah</th>
                <th>Catatan</th>
              </tr>
            </thead>
            <tbody>
              {model.rows.map((row) => (
                <tr key={row.join('-')}>
                  {row.map((cell) => <td key={cell}>{cell}</td>)}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {model.dataNotes.length > 0 ? <p className="helper-copy">{model.dataNotes[0]}</p> : null}
      </div>

      <ActivityLanes title="Status laporan" lanes={reportLanes} />
      <PreviewActionPanel group={previewActionGroups.reports} />
      <PreviewStateBoard
        title="Kondisi tampilan laporan"
        description="Laporan membutuhkan status eksplisit karena filter, periode, export, pagination, dan kebijakan backend semuanya berpengaruh."
        states={previewStateMatrix.reports}
      />
    </AdminShell>
  );
}
