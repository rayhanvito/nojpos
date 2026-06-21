import type { PreviewStateTone } from '@/fixtures/preview';

type PreviewStateProps = {
  title: string;
  message: string;
  tone?: PreviewStateTone;
};

const icons = {
  loading: '◌',
  empty: '□',
  error: '!',
  forbidden: '×',
  unavailable: '⌁',
} as const;

export function PreviewState({ title, message, tone = 'unavailable' }: PreviewStateProps) {
  return (
    <section className={`preview-state ${tone}`}>
      <span aria-hidden="true">{icons[tone]}</span>
      <div>
        <small>{stateLabel(tone)}</small>
        <h2>{title}</h2>
        <p>{message}</p>
      </div>
    </section>
  );
}

function stateLabel(tone: PreviewStateProps['tone']) {
  switch (tone) {
    case 'loading':
      return 'Menyiapkan tampilan';
    case 'empty':
      return 'Data kosong';
    case 'error':
      return 'Error contoh';
    case 'forbidden':
      return 'Hak akses';
    default:
      return 'Belum tersedia';
  }
}
