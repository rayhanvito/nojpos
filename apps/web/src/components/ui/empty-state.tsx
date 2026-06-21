import { Inbox } from 'lucide-react';
import type { ReactNode } from 'react';

import { cn } from '@/lib/utils';

import { Badge } from './badge';

export type EmptyStateProps = {
  title: string;
  description: string;
  label?: string;
  icon?: ReactNode;
  className?: string;
};

export function EmptyState({ title, description, label = 'Data kosong contoh', icon, className }: EmptyStateProps) {
  return (
    <section className={cn('grid place-items-center rounded-[var(--radius)] border border-dashed border-[var(--border-strong)] bg-[var(--muted)]/45 p-6 text-center', className)}>
      <div className="grid max-w-md justify-items-center gap-3">
        <span className="grid size-12 place-items-center rounded-2xl bg-[var(--card)] text-[var(--primary)] shadow-[var(--shadow-soft)]" aria-hidden="true">
          {icon ?? <Inbox size={20} />}
        </span>
        <Badge variant="muted">{label}</Badge>
        <div>
          <h2 className="m-0 text-base font-bold text-[var(--mono)]">{title}</h2>
          <p className="mt-2 text-sm leading-6 text-[var(--muted-foreground)]">{description}</p>
        </div>
      </div>
    </section>
  );
}
