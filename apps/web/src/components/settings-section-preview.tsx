import Link from 'next/link';

import { AdminShell } from '@/components/admin-shell';
import { DisabledFormPreview } from '@/components/disabled-form-preview';
import { PreviewGate } from '@/components/preview-gate';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewNotice } from '@/components/ui/preview-notice';
import { previewStateMatrix, type SettingsSectionFixture } from '@/fixtures/preview';

export function SettingsSectionPreview({ section }: { section: SettingsSectionFixture }) {
  return (
    <AdminShell title={section.title}>
      <div className="detail-header card">
        <div>
          <span className="card-kicker">{section.kicker}</span>
          <h2>{section.title}</h2>
          <p>{section.description}</p>
        </div>
        <div className="detail-actions" aria-label="Aksi pengaturan contoh nonaktif">
          <span className={section.badge === 'Backend gated' ? 'badge warning' : 'badge neutral'}>{section.badge === 'Backend gated' ? 'Dikunci' : section.badge}</span>
          <span className="btn" aria-disabled="true">Simpan contoh</span>
          <Link className="btn ghost" href="/settings">Kembali</Link>
        </div>
      </div>

      <PreviewNotice />

      <div className="split-grid">
        <DisabledFormPreview
          title={`Form ${section.title}`}
          description="Form ini hanya menunjukkan layout, spacing, helper text, dan kondisi terkunci sebelum koneksi server aktif."
          fields={section.fields}
        />
      </div>

      <section className="preview-notes-section" aria-labelledby={`${section.title}-mode-notes`}>
        <div className="preview-notes-header">
          <span className="card-kicker">Catatan mode contoh</span>
          <h2 id={`${section.title}-mode-notes`}>Kontrol pengaturan tetap aman</h2>
          <p>Panel ini menjelaskan status koneksi dan kondisi tampilan tanpa mendominasi form utama.</p>
        </div>
        <PreviewGate title={`${section.title} belum terhubung`} />
        <PreviewStateBoard
          title={`Kondisi tampilan ${section.title}`}
          description="Loading, kosong, error, hak akses, dan fitur belum tersedia tetap konsisten di area pengaturan."
          states={previewStateMatrix.settings}
        />
      </section>
    </AdminShell>
  );
}
