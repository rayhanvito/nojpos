import Image from 'next/image';

const features = [
  {
    title: 'Kasir cepat',
    copy: 'Bantu tim mencatat penjualan harian dengan alur yang sederhana.',
    image: '/landing/web/hero-cashier-mobile.png',
  },
  {
    title: 'Produk & inventori',
    copy: 'Kelola produk, kategori, harga, dan stok dari satu tampilan.',
    image: '/landing/web/feature-inventory.png',
  },
  {
    title: 'Multi outlet',
    copy: 'Pantau beberapa cabang tanpa harus pindah-pindah catatan.',
    image: '/landing/web/feature-multi-outlet.png',
  },
  {
    title: 'Metode pembayaran',
    copy: 'Siapkan tampilan pembayaran yang rapi untuk berbagai metode.',
    image: '/landing/web/feature-payment-methods.png',
  },
  {
    title: 'Laporan penjualan',
    copy: 'Lihat ringkasan usaha agar keputusan lebih mudah diambil.',
    image: '/landing/web/feature-reports.png',
  },
  {
    title: 'Tim & absensi',
    copy: 'Bantu atur anggota tim, shift, dan aktivitas operasional.',
    image: '/landing/web/feature-team-management.png',
  },
] as const;

export function LandingFeatureGrid() {
  return (
    <section className="landing-section landing-feature-section" id="fitur">
      <div className="landing-container">
        <div className="landing-section-header">
          <p className="landing-eyebrow">Fitur utama</p>
          <h2>Fitur yang membantu operasional harian</h2>
          <p>NojPOS menyiapkan alur kasir dan admin yang mudah dipahami untuk usaha lokal.</p>
        </div>

        <div className="landing-feature-grid">
          {features.map((feature) => (
            <article className="landing-feature-card" key={feature.title}>
              <div className="landing-feature-image">
                <Image src={feature.image} alt="" fill sizes="(max-width: 768px) 90vw, 360px" />
              </div>
              <div>
                <h3>{feature.title}</h3>
                <p>{feature.copy}</p>
              </div>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
