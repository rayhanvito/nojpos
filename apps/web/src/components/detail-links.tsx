import Link from 'next/link';

import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

type DetailLink = {
  href: string;
  label: string;
  description: string;
};

export function DetailLinks({ title, links }: { title: string; links: readonly DetailLink[] }) {
  return (
    <Card className="detail-links-card" aria-labelledby={`${title.replaceAll(' ', '-').toLowerCase()}-links`}>
      <CardHeader>
        <div>
          <span className="card-kicker">Detail route preview</span>
          <CardTitle id={`${title.replaceAll(' ', '-').toLowerCase()}-links`}>{title}</CardTitle>
        </div>
        <Badge variant="neutral">Static preview</Badge>
      </CardHeader>
      <CardContent>
        <div className="detail-link-grid">
          {links.map((link) => (
            <Link className="detail-link" href={link.href} key={link.href}>
              <strong>{link.label}</strong>
              <span>{link.description}</span>
            </Link>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
