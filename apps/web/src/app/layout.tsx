import type { Metadata } from 'next';
import { Inter, Plus_Jakarta_Sans } from 'next/font/google';
import './globals.css';

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-app',
  display: 'swap',
});

const plusJakartaSans = Plus_Jakarta_Sans({
  subsets: ['latin'],
  variable: '--font-brand',
  display: 'swap',
});

export const metadata: Metadata = {
  title: 'NojPOS Preview',
  description: 'Preview UI NojPOS untuk landing, admin toko, dan platform internal.',
  icons: { icon: '/assets/nojpos/logo/logo-app-icon.png' },
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="id" className={`${inter.variable} ${plusJakartaSans.variable}`}>
      <body>{children}</body>
    </html>
  );
}
