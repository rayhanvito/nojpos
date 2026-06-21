import Image from 'next/image';
import Link from 'next/link';
import { Building2, Coffee, Home as HomeIcon, Scissors, Shirt, ShoppingBag, Store, Utensils } from 'lucide-react';

import { LandingCta } from '@/components/landing/landing-cta';
import { LandingFaq } from '@/components/landing/landing-faq';
import { LandingFeatureGrid } from '@/components/landing/landing-feature-grid';
import { LandingFooter } from '@/components/landing/landing-footer';
import { LandingHero } from '@/components/landing/landing-hero';
import { LandingHowItWorks } from '@/components/landing/landing-how-it-works';
import { LandingNavbar } from '@/components/landing/landing-navbar';
import { LandingPricing } from '@/components/landing/landing-pricing';
import { LandingSection } from '@/components/landing/landing-section';

const problemCards = [
  {
    title: 'Penjualan masih dicatat manual',
    copy: 'Catatan bisa tercecer saat toko ramai atau saat shift berganti.',
  },
  {
    title: 'Stok sering tidak cocok',
    copy: 'Produk yang hampir habis bisa terlambat diketahui oleh owner dan tim.',
  },
  {
    title: 'Sulit pantau banyak outlet',
    copy: 'Makin banyak cabang, makin sulit membaca kondisi toko dari catatan terpisah.',
  },
  {
    title: 'Laporan harian makan waktu',
    copy: 'Owner perlu rekap yang lebih mudah dibaca tanpa menyusun ulang dari awal.',
  },
] as const;

const solutionPoints = [
  'Transaksi tercatat lebih rapi.',
  'Produk dan stok lebih mudah dipantau.',
  'Tim dan outlet bisa dikelola dari satu tempat.',
  'Owner bisa melihat ringkasan usaha kapan saja.',
] as const;

const audienceCards = [
  { title: 'Kafe & coffee shop', icon: Coffee },
  { title: 'Restoran kecil', icon: Utensils },
  { title: 'Toko retail', icon: ShoppingBag },
  { title: 'Laundry', icon: Shirt },
  { title: 'Barbershop/salon', icon: Scissors },
  { title: 'Booth minuman', icon: Store },
  { title: 'UMKM harian', icon: HomeIcon },
  { title: 'Usaha multi-outlet', icon: Building2 },
] as const;

export default function Home() {
  return (
    <div className="landing-page">
      <LandingNavbar />
      <main id="top">
        <LandingHero />

        <LandingSection
          eyebrow="Masalah umum"
          title="Operasional toko sering terasa ribet?"
          description="Banyak usaha lokal masih mengandalkan catatan terpisah. NojPOS hadir untuk membantu membuat alur kerja lebih rapi tanpa terasa rumit."
          className="landing-problem-section"
        >
          <div className="landing-problem-grid">
            {problemCards.map((problem) => (
              <article className="landing-problem-card" key={problem.title}>
                <span aria-hidden="true">•</span>
                <div>
                  <h3>{problem.title}</h3>
                  <p>{problem.copy}</p>
                </div>
              </article>
            ))}
          </div>
        </LandingSection>

        <LandingSection
          eyebrow="Solusi"
          title="NojPOS menyatukan kasir dan admin toko dalam satu alur kerja"
          description="Tampilan landing ini memperkenalkan arah produk NojPOS sebagai POS dan admin usaha yang mudah dipahami untuk owner dan tim toko."
          className="landing-solution-section"
        >
          <div className="landing-solution-grid">
            <div className="landing-solution-list">
              {solutionPoints.map((point) => (
                <div className="landing-solution-item" key={point}>
                  <span aria-hidden="true">✓</span>
                  <p>{point}</p>
                </div>
              ))}
            </div>
            <div className="landing-solution-image">
              <Image src="/landing/web/hero-pos-dashboard.png" alt="Preview alur admin NojPOS" fill sizes="(max-width: 768px) 92vw, 520px" />
            </div>
          </div>
        </LandingSection>

        <LandingFeatureGrid />

        <LandingSection
          id="cocok-untuk"
          eyebrow="Cocok untuk"
          title="Dibuat untuk berbagai usaha lokal"
          description="NojPOS disiapkan untuk usaha yang ingin transaksi, produk, stok, dan laporan terlihat lebih rapi dari satu tempat."
          className="landing-audience-section"
        >
          <div className="landing-audience-grid">
            {audienceCards.map(({ title, icon: Icon }) => (
              <article className="landing-audience-card" key={title}>
                <Icon aria-hidden="true" size={22} strokeWidth={2.2} />
                <h3>{title}</h3>
              </article>
            ))}
          </div>
        </LandingSection>

        <LandingHowItWorks />

        <LandingSection
          eyebrow="Admin preview"
          title="Lihat tampilan admin NojPOS"
          description="Masuk ke demo admin untuk melihat contoh dashboard, produk, pelanggan, transaksi, pembayaran, dan laporan."
          className="landing-admin-preview-section"
        >
          <div className="landing-admin-preview-card">
            <div className="landing-admin-preview-image">
              <Image src="/landing/web/hero-pos-dashboard.png" alt="Tampilan admin NojPOS" fill sizes="(max-width: 768px) 92vw, 760px" />
            </div>
            <div className="landing-admin-preview-copy">
              <p className="landing-preview-pill">Demo UI preview</p>
              <p>Demo ini masih UI preview. Data dan aksi belum terhubung ke server.</p>
              <Link className="landing-primary-button" href="/dashboard">
                Buka Demo Admin
              </Link>
            </div>
          </div>
        </LandingSection>

        <LandingPricing />
        <LandingFaq />
        <LandingCta />
      </main>
      <LandingFooter />
    </div>
  );
}
