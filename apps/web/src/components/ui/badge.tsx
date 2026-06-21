import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';

import { cn } from '@/lib/utils';

const badgeVariants = cva('inline-flex min-h-6 items-center justify-center rounded-full px-2.5 py-0.5 text-xs font-bold leading-none whitespace-nowrap', {
  variants: {
    variant: {
      neutral: 'bg-[var(--primary-soft)] text-[var(--primary)]',
      muted: 'bg-[var(--muted)] text-[var(--muted-foreground)]',
      success: 'bg-[var(--success-soft)] text-[var(--success-foreground)]',
      warning: 'bg-[var(--warning-soft)] text-[var(--warning-foreground)]',
      danger: 'bg-[var(--danger-soft)] text-[var(--danger-foreground)]',
      info: 'bg-[var(--info-soft)] text-[var(--info-foreground)]',
      outline: 'border border-[var(--border)] bg-transparent text-[var(--muted-foreground)]',
    },
  },
  defaultVariants: {
    variant: 'neutral',
  },
});

export type BadgeProps = React.HTMLAttributes<HTMLSpanElement> & VariantProps<typeof badgeVariants>;

export function Badge({ className, variant, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant, className }))} {...props} />;
}

export { badgeVariants };
