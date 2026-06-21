import { Badge, type BadgeProps } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import type { PreviewMetricFixture } from '@/fixtures/preview';

type PlatformMetricTone = Extract<PreviewMetricFixture['tone'], BadgeProps['variant']>;

export function PlatformMetric({ label, value, description, badge, tone = 'neutral' }: PreviewMetricFixture) {
  const badgeTone: PlatformMetricTone = tone;

  return (
    <Card className={`metric-card platform-metric-card ${tone}`} aria-label={label}>
      <div className="metric-card-top">
        <span>{label}</span>
        {badge ? <Badge variant={badgeTone}>{badge}</Badge> : null}
      </div>
      <p>{value}</p>
      <small>{description}</small>
    </Card>
  );
}
