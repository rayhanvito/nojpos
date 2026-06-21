import * as React from 'react';

import { cn } from '@/lib/utils';

export function Card({ className, ...props }: React.HTMLAttributes<HTMLElement>) {
  return <section data-slot="card" className={cn('rounded-[var(--radius)] border border-[var(--border)] bg-[var(--card)] text-[var(--card-foreground)] shadow-[var(--shadow-card)]', className)} {...props} />;
}

export function CardHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div data-slot="card-header" className={cn('flex flex-col gap-1.5 p-5 sm:flex-row sm:items-start sm:justify-between', className)} {...props} />;
}

export function CardTitle({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>) {
  return <h2 data-slot="card-title" className={cn('text-base font-bold tracking-[-0.02em] text-[var(--mono)]', className)} {...props} />;
}

export function CardDescription({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>) {
  return <p data-slot="card-description" className={cn('text-sm leading-6 text-[var(--muted-foreground)]', className)} {...props} />;
}

export function CardContent({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div data-slot="card-content" className={cn('p-5 pt-0', className)} {...props} />;
}
