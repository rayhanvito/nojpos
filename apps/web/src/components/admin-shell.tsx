'use client';

import {
  Activity,
  BadgePercent,
  BarChart3,
  Boxes,
  Clock,
  CreditCard,
  Crown,
  LayoutDashboard,
  Package,
  ReceiptText,
  Settings,
  Store,
  UserRoundCog,
  Users,
  type LucideIcon,
} from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { type ReactNode, useState } from 'react';

import { ShellContainer } from '@/components/ui/shell-container';

type TenantNavigationItem = {
  href: string;
  label: string;
  icon: LucideIcon;
  hint: string;
};

type TenantNavigationGroup = {
  group: string;
  items: TenantNavigationItem[];
};

const navigationGroups: TenantNavigationGroup[] = [
  {
    group: 'Utama',
    items: [
      { href: '/dashboard', label: 'Dasbor', icon: LayoutDashboard, hint: 'Ringkasan toko' },
      { href: '/transactions', label: 'Transaksi', icon: ReceiptText, hint: 'Riwayat penjualan' },
      { href: '/payments', label: 'Pembayaran Pelanggan', icon: CreditCard, hint: 'Status pembayaran' },
    ],
  },
  {
    group: 'Operasional',
    items: [
      { href: '/catalog', label: 'Produk', icon: Package, hint: 'Menu & kategori' },
      { href: '/inventory', label: 'Stok', icon: Boxes, hint: 'Pantau persediaan' },
      { href: '/outlets', label: 'Outlet', icon: Store, hint: 'Cabang usaha' },
    ],
  },
  {
    group: 'Relasi',
    items: [
      { href: '/customers', label: 'Pelanggan', icon: Users, hint: 'Data pelanggan' },
      { href: '/promotions', label: 'Promo', icon: BadgePercent, hint: 'Voucher & diskon' },
    ],
  },
  {
    group: 'Tim',
    items: [
      { href: '/staff', label: 'Karyawan', icon: UserRoundCog, hint: 'Role & akses' },
      { href: '/attendance', label: 'Absensi', icon: Clock, hint: 'Jam kerja staf' },
    ],
  },
  {
    group: 'Analisis',
    items: [
      { href: '/reports', label: 'Laporan', icon: BarChart3, hint: 'Ringkasan usaha' },
      { href: '/audit', label: 'Aktivitas', icon: Activity, hint: 'Jejak perubahan' },
    ],
  },
  {
    group: 'Akun',
    items: [
      { href: '/subscription', label: 'Langganan NojPOS', icon: Crown, hint: 'Paket & bantuan' },
      { href: '/settings', label: 'Pengaturan', icon: Settings, hint: 'Info toko & struk' },
    ],
  },
];

const navigation = navigationGroups.flatMap((section) => section.items);

export function AdminShell({ title, children }: { title: string; children: ReactNode }) {
  const pathname = usePathname();
  const [navOpen, setNavOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(false);
  const active = navigation.find((item) => pathname === item.href || pathname.startsWith(`${item.href}/`));

  return (
    <div className={collapsed ? 'app-shell tenant-shell sidebar-collapsed' : 'app-shell tenant-shell'}>
      <button
        className="mobile-nav-button"
        type="button"
        aria-expanded={navOpen}
        aria-controls="admin-sidebar"
        onClick={() => setNavOpen((current) => !current)}
      >
        <span aria-hidden="true">☰</span>
        Menu
      </button>

      {navOpen ? <button className="sidebar-scrim" aria-label="Tutup navigasi" type="button" onClick={() => setNavOpen(false)} /> : null}

      <aside id="admin-sidebar" className={navOpen ? 'sidebar open' : 'sidebar'} aria-label="Navigasi admin tenant">
        <div className="sidebar-brand-row">
          <Link className="brand" href="/dashboard" aria-label="NojPOS Admin dashboard" onClick={() => setNavOpen(false)}>
            <span className="brand-mark"><Image src="/assets/nojpos/logo/logo-symbol.png" alt="" width={1254} height={1254} /></span>
            <span className="brand-copy">
              <strong>NojPOS</strong>
              <small>Admin Toko</small>
            </span>
          </Link>
          <button
            className="sidebar-collapse-button"
            type="button"
            aria-label={collapsed ? 'Buka sidebar admin tenant' : 'Ciutkan sidebar admin tenant'}
            aria-pressed={collapsed}
            title={collapsed ? 'Buka sidebar' : 'Ciutkan sidebar'}
            onClick={() => setCollapsed((current) => !current)}
          >
            <span aria-hidden="true">{collapsed ? '›' : '‹'}</span>
          </button>
        </div>

        <section className="workspace-card" aria-label="Konteks admin toko">
          <span className="workspace-avatar">GR</span>
          <div>
            <strong>Owner Toko</strong>
            <small>Semua outlet</small>
          </div>
        </section>

        <nav className="sidebar-nav" aria-label="Menu admin tenant">
          {navigationGroups.map((section) => (
            <section className="sidebar-nav-group" aria-label={section.group} key={section.group}>
              <div className="sidebar-section-label">{section.group}</div>
              <div className="sidebar-nav-list">
                {section.items.map((item) => {
                  const isActive = pathname === item.href || pathname.startsWith(`${item.href}/`);
                  const Icon = item.icon;

                  return (
                    <Link className={isActive ? 'active' : ''} href={item.href} key={item.href} onClick={() => setNavOpen(false)} title={collapsed ? item.label : undefined}>
                      <span className="nav-icon" aria-hidden="true"><Icon size={18} strokeWidth={2.25} /></span>
                      <span className="nav-copy">
                        <strong>{item.label}</strong>
                        <small>{item.hint}</small>
                      </span>
                    </Link>
                  );
                })}
              </div>
            </section>
          ))}
        </nav>

        <section className="sidebar-card" aria-label="Status tampilan contoh">
          <span className="status-dot" aria-hidden="true" />
          <div>
            <strong>Mode contoh UI</strong>
            <p>Tampilan dipakai untuk menilai alur, hirarki, dan rasa visual admin toko sebelum fitur aktif.</p>
          </div>
        </section>
      </aside>

      <div className="workspace">
        <header className="topbar">
          <div className="topbar-main">
            <span className="eyebrow">Admin Toko · {active?.label ?? 'Dasbor'}</span>
            <h1>{title}</h1>
            <div className="breadcrumb" aria-label="Breadcrumb admin toko">
              <span>Admin Toko</span>
              <span aria-hidden="true">/</span>
              <span>{active?.label ?? title}</span>
            </div>
          </div>
          <div className="topbar-actions" aria-label="Konteks halaman admin toko">
            <span className="pill">Mode contoh UI</span>
            <span className="pill muted-pill">Semua outlet</span>
            <span className="avatar" aria-label="Operator contoh GR">GR</span>
          </div>
        </header>

        <main className="content-shell">
          <ShellContainer>
            <div className="command-strip" aria-label="Command strip tampilan contoh">
              <span><strong>Menu</strong> mirip POS: pilih, filter, lihat detail</span>
              <span>Aksi simpan/ubah tetap dikunci</span>
              <span>Bahasa & label disiapkan untuk pengguna Indonesia</span>
            </div>
            {children}
          </ShellContainer>
        </main>
      </div>
    </div>
  );
}
