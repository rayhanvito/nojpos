const plans = [
  {
    name: 'Starter',
    copy: 'Untuk usaha kecil yang ingin mulai lebih rapi.',
    features: ['Kasir dan produk', 'Laporan dasar', '1 outlet'],
  },
  {
    name: 'Growth',
    copy: 'Untuk usaha yang mulai berkembang.',
    features: ['Multi outlet', 'Inventori', 'Tim dan absensi'],
    highlighted: true,
  },
  {
    name: 'Pro',
    copy: 'Untuk operasional yang butuh kontrol lebih.',
    features: ['Kontrol lanjutan', 'Audit aktivitas', 'Support prioritas'],
  },
] as const;

export function LandingPricing() {
  return (
    <section className="landing-section landing-pricing-section" id="harga">
      <div className="landing-container">
        <div className="landing-section-header">
          <p className="landing-eyebrow">Paket</p>
          <h2>Pilih paket sesuai tahap usahamu</h2>
          <p>Harga final akan diumumkan saat layanan siap diluncurkan.</p>
        </div>

        <div className="landing-pricing-grid">
          {plans.map((plan) => (
            <article className={'highlighted' in plan ? 'landing-pricing-card landing-pricing-card-highlight' : 'landing-pricing-card'} key={plan.name}>
              <div>
                <h3>{plan.name}</h3>
                <p>{plan.copy}</p>
              </div>
              <ul>
                {plan.features.map((feature) => (
                  <li key={feature}>{feature}</li>
                ))}
              </ul>
              <button className="landing-disabled-button" disabled type="button">
                Segera hadir
              </button>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
