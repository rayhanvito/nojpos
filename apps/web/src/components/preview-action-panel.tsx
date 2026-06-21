import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import type { PreviewActionGroupFixture, PreviewActionCategory } from '@/fixtures/preview';

const categoryLabels: Record<PreviewActionCategory, string> = {
  read: 'Akses baca belum terhubung',
  write: 'Aksi perubahan dikunci',
  import: 'Import data dikunci',
  export: 'Export data dikunci',
  danger: 'Aksi sensitif dikunci',
};

type PreviewActionPanelProps = {
  group: PreviewActionGroupFixture;
  compact?: boolean;
};

export function PreviewActionPanel({ group, compact = false }: PreviewActionPanelProps) {
  return (
    <Card className={compact ? 'action-panel action-panel-compact' : 'action-panel'} aria-label={group.title}>
      <CardHeader>
        <div>
          <span className="card-kicker">Kontrol aksi</span>
          <CardTitle>{group.title}</CardTitle>
        </div>
        <Badge variant="warning">Dikunci</Badge>
      </CardHeader>
      <CardContent>
        <CardDescription>{group.description}</CardDescription>
        <div className="action-grid">
          {group.actions.map((action) => (
            <article className={`action-item action-${action.category}`} key={`${group.title}-${action.label}`}>
              <div className="action-item-head">
                <span className="action-category">{categoryLabels[action.category]}</span>
                <Button type="button" disabled aria-label={`${action.label} belum aktif`} variant="gated" size="sm">
                  {action.label}
                </Button>
              </div>
              <p>{action.reason}</p>
            </article>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
