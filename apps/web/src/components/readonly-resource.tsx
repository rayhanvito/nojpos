import { PreviewTable } from '@/components/preview-table';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { readonlyResourceTableFor } from '@/fixtures/preview';

type ReadonlyResourceProps = {
  path: string;
  label: string;
  description?: string;
};

export function ReadonlyResource({ path, label, description }: ReadonlyResourceProps) {
  const table = readonlyResourceTableFor(path);

  return (
    <Card className="readonly-resource-card">
      <CardHeader>
        <div>
          <span className="card-kicker">Data contoh terkunci</span>
          <CardTitle>{label}</CardTitle>
        </div>
        <Badge variant="neutral">Contoh UI</Badge>
      </CardHeader>
      <CardContent>
        {description ? <CardDescription>{description}</CardDescription> : null}
        {table.rows.length > 0 ? (
          <PreviewTable columns={table.columns} rows={table.rows} />
        ) : (
          <p className="muted-note">Belum ada data contoh untuk tampilan ini.</p>
        )}
      </CardContent>
    </Card>
  );
}
