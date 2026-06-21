import { Badge, type BadgeProps } from '@/components/ui/badge';
import { EmptyState } from '@/components/ui/empty-state';
import { Table, TableBody, TableCell, TableContainer, TableHead, TableHeader, TableRow } from '@/components/ui/table';

export type PreviewTableProps = {
  columns: readonly string[];
  rows: readonly (readonly string[])[];
  caption?: string;
  primaryColumn?: string | number;
  statusColumn?: string | number;
  metaColumns?: readonly (string | number)[];
  numericColumns?: readonly (string | number)[];
  rowHrefs?: readonly (string | undefined)[];
};

type ResolvedTableConfig = {
  primaryIndex: number;
  statusIndex: number;
  metaIndexes: readonly number[];
  numericIndexes: readonly number[];
};

export function PreviewTable(props: PreviewTableProps) {
  return <ResponsivePreviewTable {...props} />;
}

export function ResponsivePreviewTable({ columns, rows, caption, primaryColumn, statusColumn, metaColumns, numericColumns, rowHrefs }: PreviewTableProps) {
  const config = resolveTableConfig({ columns, primaryColumn, statusColumn, metaColumns, numericColumns });

  return (
    <div className="data-grid-card">
      {caption ? <p className="table-caption">{caption}</p> : null}
      {rows.length === 0 ? (
        <EmptyState title="Belum ada data contoh untuk bagian ini." description="List akan terisi saat data contoh atau koneksi server sudah tersedia." />
      ) : (
        <>
          <TableContainer className="table-wrap responsive-table-desktop" data-testid="responsive-preview-table-desktop">
            <Table>
              <TableHeader>
                <TableRow>{columns.map((column, index) => <TableHead key={column} className={config.numericIndexes.includes(index) ? 'table-cell-numeric' : undefined}>{column}</TableHead>)}</TableRow>
              </TableHeader>
              <TableBody>
                {rows.map((row, index) => (
                  <TableRow key={rowKey(row, index)}>
                    {row.map((value, cell) => (
                      <TableCell key={`${index}-${cell}`} className={config.numericIndexes.includes(cell) ? 'table-cell-numeric' : undefined}>
                        {renderCell(value, cell === config.statusIndex)}
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>

          <div className="responsive-table-mobile" data-testid="responsive-preview-table-mobile" aria-label={caption ?? 'Daftar data contoh'}>
            {rows.map((row, index) => renderMobileRow(row, index, columns, config, rowHrefs?.[index]))}
          </div>
        </>
      )}
    </div>
  );
}

function renderMobileRow(row: readonly string[], index: number, columns: readonly string[], config: ResolvedTableConfig, href?: string) {
  const title = row[config.primaryIndex] ?? row[0] ?? 'Data contoh';
  const status = row[config.statusIndex];
  const metaIndexes = config.metaIndexes.filter((cell) => cell !== config.primaryIndex && cell !== config.statusIndex && row[cell] !== undefined);
  const content = (
    <>
      <div className="mobile-row-main">
        <div className="mobile-row-title-group">
          <span className="mobile-row-kicker">{columns[config.primaryIndex] ?? 'Data'}</span>
          <strong>{title}</strong>
        </div>
        {status ? <span className="mobile-row-status">{renderStatusBadge(status)}</span> : null}
      </div>
      <dl className="mobile-row-meta">
        {metaIndexes.map((cell) => (
          <div key={`${index}-${cell}`}>
            <dt>{columns[cell]}</dt>
            <dd>{renderCell(row[cell], false)}</dd>
          </div>
        ))}
      </dl>
      {href ? <span className="mobile-row-hint">Buka detail contoh</span> : null}
    </>
  );

  if (href) {
    return <a key={rowKey(row, index)} className="mobile-row-card is-clickable" href={href}>{content}</a>;
  }

  return <article key={rowKey(row, index)} className="mobile-row-card">{content}</article>;
}

function resolveTableConfig({ columns, primaryColumn, statusColumn, metaColumns, numericColumns }: Pick<PreviewTableProps, 'columns' | 'primaryColumn' | 'statusColumn' | 'metaColumns' | 'numericColumns'>): ResolvedTableConfig {
  const statusIndex = columnToIndex(columns, statusColumn) ?? findColumnIndex(columns, ['status', 'state', 'kondisi', 'signal']) ?? Math.max(columns.length - 1, 0);
  const primaryIndex = columnToIndex(columns, primaryColumn) ?? 0;
  const metaIndexes = (metaColumns?.map((column) => columnToIndex(columns, column)).filter((index): index is number => typeof index === 'number') ?? columns.map((_, index) => index)).filter((index) => index >= 0 && index < columns.length);
  const numericIndexes = numericColumns?.map((column) => columnToIndex(columns, column)).filter((index): index is number => typeof index === 'number') ?? columns.map((column, index) => isNumericColumn(column) ? index : -1).filter((index) => index >= 0);

  return { primaryIndex, statusIndex, metaIndexes, numericIndexes };
}

function columnToIndex(columns: readonly string[], column?: string | number) {
  if (typeof column === 'number') {
    return column >= 0 && column < columns.length ? column : undefined;
  }

  if (!column) return undefined;
  const normalized = normalize(column);
  const index = columns.findIndex((candidate) => normalize(candidate) === normalized);
  return index >= 0 ? index : undefined;
}

function findColumnIndex(columns: readonly string[], keywords: readonly string[]) {
  return columns.findIndex((column) => keywords.some((keyword) => normalize(column).includes(keyword)));
}

function isNumericColumn(column: string) {
  const normalized = normalize(column);
  return ['total', 'value', 'nilai', 'amount', 'nominal', 'mrr', 'revenue', 'pendapatan'].some((keyword) => normalized.includes(keyword));
}

function renderCell(value: string, forceBadge: boolean) {
  if (!forceBadge && !isStatusLike(value)) {
    return value;
  }

  return renderStatusBadge(value);
}

function renderStatusBadge(value: string) {
  return <Badge variant={statusTone(value)}>{value}</Badge>;
}

export function statusTone(value: string): BadgeProps['variant'] {
  const normalized = normalize(value);

  if (matchesAny(normalized, ['aktif', 'active', 'selesai', 'paid', 'sehat', 'healthy', 'normal', 'tersedia', 'available', 'success'])) return 'success';
  if (matchesAny(normalized, ['pending', 'trial', 'perlu dicek', 'menunggu', 'waiting', 'review', 'contract pending', 'gated', 'dikunci'])) return 'warning';
  if (matchesAny(normalized, ['gagal', 'failed', 'overdue', 'error', 'blocked', 'danger', 'nonaktif', 'suspended'])) return 'danger';
  if (matchesAny(normalized, ['preview', 'read-only', 'read only', 'contoh', 'masked', 'safe'])) return 'info';
  if (matchesAny(normalized, ['unavailable', 'belum tersedia', 'kosong', 'empty', 'not connected', 'disabled', '—'])) return 'muted';

  return 'neutral';
}

function isStatusLike(value: string) {
  const normalized = normalize(value);
  return matchesAny(normalized, [
    'aktif', 'active', 'selesai', 'paid', 'sehat', 'healthy', 'normal', 'tersedia', 'available', 'success',
    'pending', 'trial', 'perlu dicek', 'menunggu', 'waiting', 'review', 'contract pending', 'gated', 'dikunci',
    'gagal', 'failed', 'overdue', 'error', 'blocked', 'danger', 'nonaktif', 'suspended',
    'preview', 'read-only', 'read only', 'contoh', 'masked', 'safe',
    'unavailable', 'belum tersedia', 'kosong', 'empty', 'not connected', 'disabled', '—',
  ]);
}

function matchesAny(value: string, keywords: readonly string[]) {
  return keywords.some((keyword) => value.includes(keyword));
}

function normalize(value: string) {
  return value.trim().toLowerCase();
}

function rowKey(row: readonly string[], index: number) {
  return `${index}-${row.join('|')}`;
}
