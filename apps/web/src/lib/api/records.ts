export type ApiRecord = Record<string, unknown>;
export function recordsFromPayload(payload: unknown): ApiRecord[] {
  if (typeof payload !== 'object' || payload === null || !('items' in payload)) return [];
  const items = (payload as { items?: unknown }).items;
  return Array.isArray(items) && items.every((item) => typeof item === 'object' && item !== null) ? items as ApiRecord[] : [];
}
