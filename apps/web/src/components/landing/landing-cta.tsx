import Image from 'next/image';
import Link from 'next/link';

export function LandingCta() {
  return (
    <section className="landing-section landing-final-cta" aria-labelledby="landing-final-cta-title">
      <div className="landing-container">
        <div className="landing-cta-card">
          <div>
            <p className="landing-eyebrow">Preview NojPOS</p>
            <h2 id="landing-final-cta-title">Siap bikin operasional toko lebih rapi?</h2>
            <p>Coba lihat tampilan NojPOS dan bayangkan alur kerja toko kamu jadi lebih mudah dipantau.</p>
            <div className="landing-hero-actions">
              <Link className="landing-primary-button" href="/dashboard">
                Lihat Demo Admin
              </Link>
              <a className="landing-secondary-button" href="#top">
                Kembali ke Atas
              </a>
            </div>
          </div>
          <div className="landing-cta-image">
            <Image src="/landing/web/cta-start-selling.png" alt="Ilustrasi mulai menggunakan NojPOS" fill sizes="(max-width: 768px) 88vw, 380px" />
          </div>
        </div>
      </div>
    </section>
  );
}
