import type { ReactNode } from 'react';

import { cn } from '@/lib/utils';

export function ShellContainer({ children, className }: { children: ReactNode; className?: string }) {
  return <div className={cn('shell-container', className)}>{children}</div>;
}
