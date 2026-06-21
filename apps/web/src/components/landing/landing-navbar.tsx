import Image from 'next/image';
import Link from 'next/link';

const navItems = [
  { href: '#fitur', label: 'Fitur' },
  { href: '#cocok-untuk', label: 'Cocok Untuk' },
  { href: '#cara-kerja', label: 'Cara Kerja' },
  { href: '#harga', label: 'Harga' },
  { href: '#faq', label: 'FAQ' },
] as const;

export function LandingNavbar() {
  return (
    <header className="landing-navbar" aria-label="Navigasi landing NojPOS">
      <div className="landing-container landing-navbar-inner">
        <Link className="landing-brand" href="#top" aria-label="NojPOS landing page">
          <Image src="/landing/logo/logo-nojpos.png" alt="NojPOS" width={180} height={72} priority />
        </Link>

        <nav className="landing-nav-links" aria-label="Menu utama landing">
          {navItems.map((item) => (
            <a href={item.href} key={item.href}>
              {item.label}
            </a>
          ))}
        </nav>

        <div className="landing-navbar-actions">
          <Link className="landing-primary-button landing-navbar-cta" href="/dashboard">
            Lihat Demo Admin
          </Link>
        </div>
      </div>
    </header>
  );
}
