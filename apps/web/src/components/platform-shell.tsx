'use client';

import {
  Activity,
  Building2,
  Gauge,
  HeartHandshake,
  Layers3,
  Megaphone,
  ReceiptText,
  ShieldCheck,
  type LucideIcon,
} from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { type ReactNode, useState } from 'react';

import { ShellContainer } from '@/components/ui/shell-container';

type PlatformNavigationItem = {
  href: string;
  label: string;
  icon: LucideIcon;
  hint: string;
};

const platformNavigation: PlatformNavigationItem[] = [
  { href: '/platform', label: 'Ringkasan', icon: Gauge, hint: 'Pusat kerja' },
  { href: '/platform/businesses', label: 'Toko', icon: Building2, hint: 'Status toko' },
  { href: '/platform/plans', label: 'Paket', icon: Layers3, hint: 'Konsep paket' },
  { href: '/platform/revenue', label: 'Langganan & Tagihan', icon: ReceiptText, hint: 'Jatuh tempo' },
  { href: '/platform/support', label: 'Bantuan', icon: HeartHandshake, hint: 'Tiket toko' },
  { href: '/platform/audit', label: 'Aktivitas', icon: ShieldCheck, hint: 'Riwayat penting' },
  { href: '/platform/system-health', label: 'Sistem', icon: Activity, hint: 'Layanan inti' },
  { href: '/platform/announcements', label: 'Pengumuman', icon: Megaphone, hint: 'Info toko' },
];

export function PlatformShell({ title, children }: { title: string; children: ReactNode }) {
  const pathname = usePathname();
  const [navOpen, setNavOpen] = useState(false);
  const [collapsed, setCollapsed] = useState(false);
  const active = platformNavigation.find((item) => isPlatformNavActive(pathname, item.href));

  return (
    <div className={collapsed ? 'app-shell platform-shell sidebar-collapsed' : 'app-shell platform-shell'}>
      <button
        className="mobile-nav-button"
        type="button"
        aria-expanded={navOpen}
        aria-controls="platform-sidebar"
        onClick={() => setNavOpen((current) => !current)}
      >
        <span aria-hidden="true">☰</span>
        Platform
      </button>

      {navOpen ? <button className="sidebar-scrim" aria-label="Tutup navigasi platform" type="button" onClick={() => setNavOpen(false)} /> : null}

      <aside id="platform-sidebar" className={navOpen ? 'sidebar platform-sidebar open' : 'sidebar platform-sidebar'} aria-label="Navigasi Platform Super Admin">
        <div className="sidebar-brand-row">
          <Link className="brand" href="/platform" aria-label="NojPOS Platform dashboard" onClick={() => setNavOpen(false)}>
            <span className="brand-mark"><Image src="/assets/nojpos/logo/logo-symbol.png" alt="" width={1254} height={1254} /></span>
            <span className="brand-copy">
              <strong>NojPOS</strong>
              <small>Super Admin</small>
            </span>
          </Link>
          <button
            className="sidebar-collapse-button"
            type="button"
            aria-label={collapsed ? 'Buka sidebar platform' : 'Ciutkan sidebar platform'}
            aria-pressed={collapsed}
            title={collapsed ? 'Buka sidebar' : 'Ciutkan sidebar'}
            onClick={() => setCollapsed((current) => !current)}
          >
            <span aria-hidden="true">{collapsed ? '›' : '‹'}</span>
          </button>
        </div>

        <section className="workspace-card platform-card" aria-label="Konteks platform NojPOS">
          <span className="workspace-avatar">NP</span>
          <div>
            <strong>Tim Internal NojPOS</strong>
            <small>UI preview · data contoh</small>
          </div>
        </section>

        <nav className="sidebar-nav" aria-label="Menu platform owner">
          <section className="sidebar-nav-group" aria-label="Platform">
            <div className="sidebar-section-label">Platform</div>
            <div className="sidebar-nav-list">
              {platformNavigation.map((item) => {
                const isActive = isPlatformNavActive(pathname, item.href);
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
        </nav>

        <section className="sidebar-card" aria-label="Status platform">
          <span className="status-dot" aria-hidden="true" />
          <strong>Preview UI</strong>
          <small>Belum ada API, payment, akses bantuan, atau broadcast real.</small>
        </section>
      </aside>

      <ShellContainer className="platform-shell-container">
        <div className="shell-topbar" aria-label="Konteks halaman platform">
          <span>Platform NojPOS</span>
          <strong>{active?.label ?? title}</strong>
        </div>
        {children}
      </ShellContainer>
    </div>
  );
}

function isPlatformNavActive(pathname: string, href: string) {
  if (href === '/platform') {
    return pathname === href;
  }

  return pathname === href || pathname.startsWith(`${href}/`);
}
