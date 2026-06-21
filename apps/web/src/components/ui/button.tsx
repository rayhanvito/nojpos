import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';

import { cn } from '@/lib/utils';

const buttonVariants = cva(
  'inline-flex min-h-10 items-center justify-center gap-2 rounded-[var(--radius-sm)] border px-4 py-2 text-sm font-bold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--ring)] focus-visible:ring-offset-2 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-60 aria-disabled:pointer-events-none aria-disabled:cursor-not-allowed aria-disabled:opacity-65',
  {
    variants: {
      variant: {
        default: 'border-[var(--primary)] bg-[var(--primary)] text-white shadow-[var(--shadow-button)] hover:bg-[var(--primary-strong)]',
        secondary: 'border-[var(--border)] bg-[var(--card)] text-[var(--foreground)] hover:bg-[var(--muted)]',
        ghost: 'border-transparent bg-transparent text-[var(--muted-foreground)] hover:bg-[var(--muted)] hover:text-[var(--foreground)]',
        gated: 'border-[color-mix(in_srgb,var(--warning)_34%,transparent)] bg-[var(--warning-soft)] text-[var(--warning-foreground)]',
        danger: 'border-[color-mix(in_srgb,var(--danger)_32%,transparent)] bg-[var(--danger-soft)] text-[var(--danger-foreground)]',
      },
      size: {
        sm: 'min-h-8 px-3 text-xs',
        default: 'min-h-10 px-4',
        lg: 'min-h-11 px-5',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
);

export type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  };

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(({ className, variant, size, asChild = false, ...props }, ref) => {
  const Comp = asChild ? Slot : 'button';

  return <Comp className={cn(buttonVariants({ variant, size, className }))} ref={ref} {...props} />;
});

Button.displayName = 'Button';

export { buttonVariants };
