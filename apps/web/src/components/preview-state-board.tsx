import { PreviewState } from '@/components/preview-state';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import type { PreviewStateFixture } from '@/fixtures/preview';

type PreviewStateBoardProps = {
  title: string;
  description: string;
  states: readonly PreviewStateFixture[];
};

export function PreviewStateBoard({ title, description, states }: PreviewStateBoardProps) {
  return (
    <Card className="state-board-card" aria-label={title}>
      <CardHeader>
        <div>
          <span className="card-kicker">Kondisi tampilan contoh</span>
          <CardTitle>{title}</CardTitle>
        </div>
        <Badge variant="neutral">5 kondisi</Badge>
      </CardHeader>
      <CardContent>
        <CardDescription>{description}</CardDescription>
        <div className="state-board-grid">
          {states.map((state) => <PreviewState key={`${state.tone}-${state.title}`} {...state} />)}
        </div>
      </CardContent>
    </Card>
  );
}
