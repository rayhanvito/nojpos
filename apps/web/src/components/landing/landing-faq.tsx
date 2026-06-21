const faqs = [
  {
    question: 'Apakah NojPOS cocok untuk UMKM?',
    answer: 'Ya. NojPOS dirancang untuk membantu usaha lokal merapikan transaksi, produk, stok, dan laporan dengan tampilan yang mudah dipahami.',
  },
  {
    question: 'Apakah bisa dipakai untuk kafe atau toko retail?',
    answer: 'Bisa. Arah produk NojPOS cocok untuk kafe, restoran kecil, toko retail, booth minuman, laundry, barbershop, dan usaha harian lain.',
  },
  {
    question: 'Apakah bisa multi outlet?',
    answer: 'Landing ini menampilkan arah dukungan multi outlet. Detail data cabang dan akses tetap menunggu integrasi backend yang disetujui.',
  },
  {
    question: 'Apakah bisa pantau stok?',
    answer: 'NojPOS menyiapkan tampilan inventori untuk membantu owner dan tim memantau produk serta stok dengan lebih rapi.',
  },
  {
    question: 'Apakah bisa lihat laporan penjualan?',
    answer: 'Ya, demo UI menampilkan area laporan dan ringkasan usaha. Angka final nantinya harus berasal dari backend produksi.',
  },
  {
    question: 'Apakah demo ini sudah terhubung ke server?',
    answer: 'Belum. Demo saat ini adalah preview UI. Data dan aksi belum terhubung ke server produksi.',
  },
] as const;

export function LandingFaq() {
  return (
    <section className="landing-section landing-faq-section" id="faq">
      <div className="landing-container">
        <div className="landing-section-header">
          <p className="landing-eyebrow">FAQ</p>
          <h2>Pertanyaan yang sering muncul</h2>
          <p>Jawaban singkat agar pemilik usaha memahami posisi demo NojPOS saat ini.</p>
        </div>

        <div className="landing-faq-grid">
          {faqs.map((faq) => (
            <article className="landing-faq-card" key={faq.question}>
              <h3>{faq.question}</h3>
              <p>{faq.answer}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
