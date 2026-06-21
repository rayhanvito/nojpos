import { Info } from 'lucide-react';

import { cn } from '@/lib/utils';

type PreviewNoticeVariant = 'tenant' | 'platform';

type PreviewNoticeProps = {
  variant?: PreviewNoticeVariant;
  className?: string;
  message?: string;
};

const copy: Record<PreviewNoticeVariant, string> = {
  tenant: 'Mode contoh UI — data belum terhubung ke server dan semua aksi perubahan dikunci.',
  platform: 'Mode preview internal — data platform belum terhubung dan aksi platform masih dikunci.',
};

export function PreviewNotice({ variant = 'tenant', className, message }: PreviewNoticeProps) {
  const notice = message ?? copy[variant];

  return (
    <aside className={cn('preview-notice', `preview-notice-${variant}`, className)} aria-label={notice}>
      <Info aria-hidden="true" />
      <p>{notice}</p>
    </aside>
  );
}
