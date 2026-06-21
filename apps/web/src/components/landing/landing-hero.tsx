import Image from 'next/image';
import Link from 'next/link';

export function LandingHero() {
  return (
    <section className="landing-hero" aria-labelledby="landing-hero-title">
      <div className="landing-container landing-hero-grid">
        <div className="landing-hero-copy">
          <p className="landing-badge">Cocok untuk kafe, retail, laundry, barbershop, dan UMKM</p>
          <h1 id="landing-hero-title">POS modern untuk usaha Indonesia yang ingin lebih rapi</h1>
          <p>
            NojPOS membantu kelola transaksi, produk, stok, pembayaran, pelanggan, pegawai, outlet, dan laporan dalam satu sistem yang mudah dipahami.
          </p>
          <div className="landing-hero-actions">
            <Link className="landing-primary-button" href="/dashboard">
              Lihat Demo Admin
            </Link>
            <a className="landing-secondary-button" href="#fitur">
              Pelajari Fitur
            </a>
          </div>
          <p className="landing-preview-note">UI-first preview. Data dan aksi demo belum terhubung ke server produksi.</p>
        </div>

        <div className="landing-hero-visual" aria-label="Preview dashboard dan kasir NojPOS">
          <div className="landing-hero-dashboard">
            <Image
              src="/landing/web/hero-pos-dashboard.png"
              alt="Preview dashboard NojPOS"
              fill
              priority
              sizes="(max-width: 768px) 92vw, 58vw"
            />
          </div>
          <div className="landing-hero-phone">
            <Image
              src="/landing/web/hero-cashier-mobile.png"
              alt="Preview kasir mobile NojPOS"
              fill
              sizes="(max-width: 768px) 42vw, 220px"
            />
          </div>
        </div>
      </div>
    </section>
  );
}
