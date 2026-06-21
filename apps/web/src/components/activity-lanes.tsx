import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

type ActivityLane = {
  label: string;
  items: readonly string[];
};

type ActivityLanesProps = {
  title: string;
  lanes: readonly ActivityLane[];
  kicker?: string;
  badge?: string;
};

export function ActivityLanes({ title, lanes, kicker = 'Aktivitas contoh', badge = 'Mode contoh' }: ActivityLanesProps) {
  return (
    <Card className="lanes-card">
      <CardHeader>
        <div>
          <span className="card-kicker">{kicker}</span>
          <CardTitle>{title}</CardTitle>
        </div>
        <Badge variant="neutral">{badge}</Badge>
      </CardHeader>
      <CardContent>
        <div className="lanes-grid">
          {lanes.map((lane) => (
            <div className="lane" key={lane.label}>
              <strong>{lane.label}</strong>
              {lane.items.map((item) => (
                <span key={item}>{item}</span>
              ))}
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
