import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

type SystemHealthPanelProps = {
  checks: readonly { label: string; status: string; detail: string }[];
};

function healthTone(status: string) {
  const normalized = status.toLowerCase();

  if (normalized.includes('normal')) return 'success';
  if (normalized.includes('perlu') || normalized.includes('belum') || normalized.includes('dikunci')) return 'warning';

  return 'neutral';
}

export function SystemHealthPanel({ checks }: SystemHealthPanelProps) {
  return (
    <Card className="system-health" aria-labelledby="system-health-title">
      <CardHeader>
        <div>
          <span className="card-kicker">Kesehatan sistem</span>
          <CardTitle id="system-health-title">Kesiapan platform</CardTitle>
        </div>
        <Badge variant="info">Preview internal</Badge>
      </CardHeader>
      <CardContent>
        <CardDescription>Status ini hanya contoh tampilan. Health check asli aktif setelah backend terhubung.</CardDescription>
        <div className="health-grid">
          {checks.map((check) => (
            <div className="health-item" key={check.label} data-tone={healthTone(check.status)}>
              <span>{check.label}</span>
              <strong>{check.status}</strong>
              <small>{check.detail}</small>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
