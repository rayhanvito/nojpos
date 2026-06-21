type TenantStatusBadgeProps = {
  status: string;
};

export function TenantStatusBadge({ status }: TenantStatusBadgeProps) {
  const normalized = status.toLowerCase();
  const tone = normalized.includes('aktif') || normalized.includes('paid') || normalized.includes('sehat') ? 'success'
    : normalized.includes('gated') || normalized.includes('menunggu') || normalized.includes('risiko') ? 'warning'
      : 'neutral';

  return <span className={`badge ${tone}`}>{status}</span>;
}
