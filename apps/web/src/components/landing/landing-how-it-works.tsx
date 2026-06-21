import Image from 'next/image';

const steps = [
  {
    title: 'Atur profil toko',
    copy: 'Siapkan informasi usaha dan outlet.',
    image: '/landing/mobile/onboarding-store-setup.png',
  },
  {
    title: 'Tambah produk',
    copy: 'Masukkan produk, kategori, harga, dan stok awal.',
    image: '/landing/mobile/onboarding-add-product.png',
  },
  {
    title: 'Mulai transaksi',
    copy: 'Tim kasir bisa mulai mencatat penjualan.',
    image: '/landing/mobile/onboarding-first-sale.png',
  },
  {
    title: 'Pantau laporan',
    copy: 'Owner melihat ringkasan usaha dari admin.',
    image: '/landing/web/feature-reports.png',
  },
] as const;

export function LandingHowItWorks() {
  return (
    <section className="landing-section landing-how-section" id="cara-kerja">
      <div className="landing-container">
        <div className="landing-section-header">
          <p className="landing-eyebrow">Cara kerja</p>
          <h2>Mulai dari setup sampai laporan</h2>
          <p>Alurnya dibuat bertahap agar tim toko dan owner bisa memahami prosesnya dengan mudah.</p>
        </div>

        <div className="landing-step-grid">
          {steps.map((step, index) => (
            <article className="landing-step-card" key={step.title}>
              <span className="landing-step-number">{index + 1}</span>
              <div className="landing-step-image">
                <Image src={step.image} alt="" fill sizes="(max-width: 768px) 88vw, 240px" />
              </div>
              <h3>{step.title}</h3>
              <p>{step.copy}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
