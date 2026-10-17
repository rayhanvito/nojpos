// Breadcrumb gaya dashboard: ikon Home, di atas judul halaman, rata kiri.
export function Crumb({ page }: { page: string }) {
  return (
    <div className="opc-crumb-right">
      <nav className="opc-crumb" aria-label="Breadcrumb">
        <span>Home</span>
        <span aria-hidden="true">›</span>
        <span aria-current="page" className="text-foreground">{page}</span>
      </nav>
    </div>
  )
}
