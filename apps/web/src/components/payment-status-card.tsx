import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

type PaymentStatusCardProps = {
  label: string;
  status: string;
  description: string;
};

export function PaymentStatusCard({ label, status, description }: PaymentStatusCardProps) {
  return (
    <Card className="status-card" aria-label={`${label} pembayaran contoh`}>
      <CardHeader>
        <div>
          <span className="card-kicker">Aksi pembayaran dikunci</span>
          <CardTitle>{label}</CardTitle>
        </div>
        <Badge variant="warning">{status}</Badge>
      </CardHeader>
      <CardContent>
        <CardDescription>{description}</CardDescription>
        <Button disabled type="button" variant="secondary">Aksi pembayaran dikunci</Button>
      </CardContent>
    </Card>
  );
}
