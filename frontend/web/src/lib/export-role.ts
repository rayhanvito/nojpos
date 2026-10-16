import type { Role } from './api'

export function selectCsvExportHandler(
  role: Role | undefined,
  serverExport: () => void,
  clientExport: () => void,
): () => void {
  return role === 'admin' ? serverExport : clientExport
}
