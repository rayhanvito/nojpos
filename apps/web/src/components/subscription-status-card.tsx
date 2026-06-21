import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

type SubscriptionStatusCardProps = {
  plan: string;
  status: string;
  renewal: string;
  note: string;
};

export function SubscriptionStatusCard({ plan, status, renewal, note }: SubscriptionStatusCardProps) {
  return (
    <Card className="status-card" aria-label="Status langganan contoh">
      <CardHeader>
        <div>
          <span className="card-kicker">Langganan contoh</span>
          <CardTitle>{plan}</CardTitle>
        </div>
        <Badge variant="warning">{status}</Badge>
      </CardHeader>
      <CardContent>
        <div className="status-card-grid">
          <div>
            <span>Perpanjangan</span>
            <strong>{renewal}</strong>
          </div>
          <div>
            <span>Status</span>
            <strong>Dikunci</strong>
          </div>
        </div>
        <CardDescription>{note}</CardDescription>
      </CardContent>
    </Card>
  );
}
