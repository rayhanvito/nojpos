import * as React from 'react';

import { cn } from '@/lib/utils';

export function TableContainer({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('overflow-x-auto rounded-[var(--radius-sm)] border border-[var(--border)] bg-[var(--card)]', className)} {...props} />;
}

export function Table({ className, ...props }: React.TableHTMLAttributes<HTMLTableElement>) {
  return <table className={cn('w-full min-w-[680px] border-separate border-spacing-0 text-sm', className)} {...props} />;
}

export function TableHeader({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) {
  return <thead className={cn('[&_tr]:border-b', className)} {...props} />;
}

export function TableBody({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) {
  return <tbody className={cn('[&_tr:last-child_td]:border-b-0', className)} {...props} />;
}

export function TableRow({ className, ...props }: React.HTMLAttributes<HTMLTableRowElement>) {
  return <tr className={cn('transition-colors hover:bg-[var(--muted)]/45', className)} {...props} />;
}

export function TableHead({ className, ...props }: React.ThHTMLAttributes<HTMLTableCellElement>) {
  return <th className={cn('border-b border-[var(--border)] bg-[var(--table-header)] px-4 py-3 text-left text-xs font-extrabold uppercase tracking-[0.04em] text-[var(--muted-foreground)] whitespace-nowrap', className)} {...props} />;
}

export function TableCell({ className, ...props }: React.TdHTMLAttributes<HTMLTableCellElement>) {
  return <td className={cn('border-b border-[var(--border)] px-4 py-3 text-[var(--table-foreground)] whitespace-nowrap', className)} {...props} />;
}
