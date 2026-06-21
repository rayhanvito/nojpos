import { Badge, type BadgeProps } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';

type PreviewMetricProps = {
  label: string;
  value: string;
  description: string;
  badge?: string;
  tone?: BadgeProps['variant'];
};

export function PreviewMetric({ label, value, description, badge = 'Mode contoh', tone = 'neutral' }: PreviewMetricProps) {
  return (
    <Card className="metric-card preview-metric-card">
      <div className="metric-card-header">
        <div>
          <span className="card-kicker">{label}</span>
          <span className="stat-value">{value}</span>
        </div>
        <Badge variant={tone}>{badge}</Badge>
      </div>
      <p>{description}</p>
    </Card>
  );
}
