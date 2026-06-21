import { ShieldCheck } from 'lucide-react';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

type PreviewGateProps = {
  title?: string;
  items?: readonly string[];
};

const defaultItems = [
  'Tidak ada koneksi server aktif dari halaman contoh ini.',
  'Aksi perubahan tetap dikunci sampai login asli dan hak akses server siap.',
  'Server tetap menjadi sumber data saat fase integrasi disetujui.',
] as const;

export function PreviewGate({ title = 'Koneksi server belum aktif', items = defaultItems }: PreviewGateProps) {
  return (
    <Card className="gate-panel" aria-label={title}>
      <CardHeader className="gate-head">
        <span className="gate-icon" aria-hidden="true">
          <ShieldCheck size={18} strokeWidth={2.4} />
        </span>
        <div>
          <span className="card-kicker">Catatan keamanan mode contoh</span>
          <CardTitle>{title}</CardTitle>
        </div>
      </CardHeader>
      <CardContent>
        <div className="gate-list">
          {items.map((item) => (
            <span key={item}>
              <i aria-hidden="true" />
              {item}
            </span>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
