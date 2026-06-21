import { AdminShell } from '@/components/admin-shell';
import { DisabledFormPreview } from '@/components/disabled-form-preview';
import { PreviewActionPanel } from '@/components/preview-action-panel';
import { PreviewGate } from '@/components/preview-gate';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewTable } from '@/components/preview-table';
import { ReadonlyResource } from '@/components/readonly-resource';
import { businessProfileFields, previewActionGroups, previewStateMatrix, settingsResources, settingsSummaryTable } from '@/fixtures/preview';

export default function SettingsPage() {
  return (
    <AdminShell title="Pengaturan">
      <PreviewActionPanel group={previewActionGroups.settings} />
      <PreviewStateBoard
        title="Kondisi tampilan pengaturan"
        description="Konfigurasi bisnis, outlet, pembayaran, dan struk punya status lengkap sebelum aksi simpan dibuka."
        states={previewStateMatrix.settings}
      />
      <div className="split-grid">
        <section className="resource card">
          <div className="card-header">
            <div>
              <span className="card-kicker">Pengaturan tenant</span>
              <h2>Konfigurasi bisnis dan outlet</h2>
            </div>
            <span className="badge neutral">Preview</span>
          </div>
          <p>Struktur mengikuti pola kartu pengaturan NojPOS, tetapi semua form masih hanya lihat.</p>
          <PreviewTable columns={settingsSummaryTable.columns} rows={settingsSummaryTable.rows} />
        </section>
        <div className="stack">
          <DisabledFormPreview
            title="Profil bisnis"
            description="Form mengikuti pola pengaturan NojPOS, tetapi semua input dikunci untuk preview desain."
            fields={businessProfileFields}
          />
          <div className="panel-grid compact-panel-grid">
            {settingsResources.map((resource) => <ReadonlyResource key={resource.path} {...resource} />)}
          </div>
          <PreviewGate title="Aksi simpan pengaturan belum aktif" />
        </div>
      </div>
    </AdminShell>
  );
}
