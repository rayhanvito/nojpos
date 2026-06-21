import Link from 'next/link';

import { PreviewActionPanel } from '@/components/preview-action-panel';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardHeader } from '@/components/ui/card';
import { previewActionGroups, type DetailPreviewFixture } from '@/fixtures/preview';

export function DetailPreview({ detail }: { detail: DetailPreviewFixture }) {
  return (
    <Card className="detail-preview" aria-labelledby={`${detail.id}-title`}>
      <div className="detail-header">
        <div>
          <span className="card-kicker">{detail.kicker}</span>
          <h2 id={`${detail.id}-title`}>{detail.title}</h2>
          <p>{detail.description}</p>
        </div>
        <div className="detail-actions" aria-label="Aksi detail contoh nonaktif">
          <Button asChild variant="ghost">
            <Link href={detail.backHref}>{detail.backLabel}</Link>
          </Button>
          {detail.actions.map((action) => (
            <Button disabled key={action.label} type="button" variant="secondary">
              {action.label} · {action.status}
            </Button>
          ))}
        </div>
      </div>

      <div className="detail-layout">
        <div className="detail-section-stack">
          {detail.sections.map((section) => (
            <section className="detail-section" key={section.title}>
              <h3>{section.title}</h3>
              <div className="detail-row-list">
                {section.rows.map((row) => (
                  <div className="detail-row" key={`${section.title}-${row.label}`}>
                    <span>{row.label}</span>
                    <strong>{row.value}</strong>
                    <small>{row.helper}</small>
                  </div>
                ))}
              </div>
            </section>
          ))}
        </div>

        <aside className="detail-aside" aria-label="Timeline contoh">
          <CardHeader>
            <div>
              <span className="card-kicker">Jejak aktivitas contoh</span>
              <h3>Timeline</h3>
            </div>
            <Badge variant="neutral">{detail.badge}</Badge>
          </CardHeader>
          <ol className="detail-timeline">
            {detail.timeline.map((item) => <li key={item}>{item}</li>)}
          </ol>
          <PreviewActionPanel group={previewActionGroups.detail} compact />
        </aside>
      </div>
    </Card>
  );
}
