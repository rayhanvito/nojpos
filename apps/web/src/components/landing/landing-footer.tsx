import Image from 'next/image';
import Link from 'next/link';

const footerLinks = [
  { href: '#fitur', label: 'Fitur' },
  { href: '#cara-kerja', label: 'Cara Kerja' },
  { href: '#harga', label: 'Harga' },
  { href: '#faq', label: 'FAQ' },
  { href: '/dashboard', label: 'Demo Admin' },
] as const;

export function LandingFooter() {
  return (
    <footer className="landing-footer">
      <div className="landing-container landing-footer-inner">
        <div className="landing-footer-brand">
          <Image src="/landing/logo/logo-nojpos.png" alt="NojPOS" width={148} height={60} />
          <p>NojPOS — POS modern untuk usaha Indonesia.</p>
          <p className="landing-footer-note">Landing page ini adalah preview UI. Integrasi server belum aktif.</p>
        </div>
        <nav className="landing-footer-links" aria-label="Link footer NojPOS">
          {footerLinks.map((link) =>
            link.href.startsWith('/') ? (
              <Link href={link.href} key={link.href}>
                {link.label}
              </Link>
            ) : (
              <a href={link.href} key={link.href}>
                {link.label}
              </a>
            ),
          )}
        </nav>
      </div>
    </footer>
  );
}
