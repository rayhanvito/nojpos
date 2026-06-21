type PreviewToolbarProps = {
  action?: string;
  label?: string;
};

export function PreviewToolbar({ action = 'Export contoh', label = 'Filter contoh' }: PreviewToolbarProps) {
  return (
    <div className="preview-toolbar" aria-label={label}>
      <label>
        <span>Cari</span>
        <input aria-label="Cari" placeholder="Cari data contoh…" disabled />
      </label>
      <label>
        <span>Outlet</span>
        <select aria-label="Outlet" disabled>
          <option>Semua outlet</option>
          <option>Outlet Sudirman</option>
        </select>
      </label>
      <label>
        <span>Periode</span>
        <select aria-label="Periode" disabled>
          <option>Periode dari server</option>
        </select>
      </label>
      <button type="button" disabled>{action}</button>
    </div>
  );
}
