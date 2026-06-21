import type { ReactNode } from 'react';

import { cn } from '@/lib/utils';

import { Badge } from './badge';

export type PageHeaderProps = {
  kicker: string;
  title: string;
  description: string;
  badge?: string;
  actions?: ReactNode;
  className?: string;
};

export function PageHeader({ kicker, title, description, badge = 'UI Preview', actions, className }: PageHeaderProps) {
  return (
    <section className={cn('page-header-card', className)}>
      <div className="page-header-copy">
        <span className="card-kicker">{kicker}</span>
        <h2>{title}</h2>
        <p>{description}</p>
      </div>
      <div className="page-header-meta">
        <Badge variant="neutral">{badge}</Badge>
        {actions}
      </div>
    </section>
  );
}
