import { useEffect } from 'react'
import { Link } from 'react-router'
import Footer from './Footer'
import Navbar from './Navbar'


const TESTIMONIALS = [
  { initial: 'S', name: 'Bu Sari', role: 'Pemilik toko kelontong · Java', quote: 'Dulu omzet harian saya hitung dari buku kas setiap malam. Sekarang cukup buka dashboard. Stok langsung berkurang tiap transaksi dan struk tercetak otomatis, jadi saya tidak perlu pusing lagi.' },
  { initial: 'J', name: 'Pak Joko', role: 'Pemilik toko sembako · Surabaya', quote: 'Saya sudah tidak mencatat manual lagi. Sekarang tinggal buka ponsel, semua produk dan stok terlihat jelas. Kasir baru juga langsung bisa dipakai tanpa ribet.' },
  { initial: 'R', name: 'Bu Ratna', role: 'Pemilik toko kosmetik · Bandung', quote: 'Refund dulu membuat pusing. Sekarang cukup sekali klik dan stok kembali otomatis. Pelanggan juga senang karena struknya jelas dan rapi.' },
  { initial: 'B', name: 'Pak Bambang', role: 'Pemilik toko elektronik · Semarang', quote: 'Yang paling saya sukai adalah laporannya. Setiap malam saya bisa melihat produk mana yang laku dan mana yang harus diisi ulang, sehingga keputusan belanja lebih pasti.' },
  { initial: 'D', name: 'Bu Dewi', role: 'Pemilik toko pakaian · Yogyakarta', quote: 'Kasirnya cepat, pelanggan tidak menunggu lama saat toko ramai. Pembayaran QRIS, transfer, dan tunai tersedia semua, dan gratis pula.' },
  { initial: 'H', name: 'Pak Hendra', role: 'Pemilik minimarket · Makassar', quote: 'Karyawan saya diberi akses terbatas, hanya bisa bertransaksi. Data tetap aman dan saya masih bisa memantau omzet dari rumah.' },
  { initial: 'S', name: 'Bu Siti', role: 'Pemilik toko kelontong · Malang', quote: 'Dulu saya sering kehabisan stok tanpa sadar. Sekarang stok yang menipis langsung terlihat di dashboard, jadi saya bisa membeli barang sebelum habis.' },
  { initial: 'A', name: 'Pak Agus', role: 'Pemilik toko aksesoris · Denpasar', quote: 'Mudah dipelajari, orang awam seperti saya pun langsung bisa memakainya. Setiap transaksi tercatat otomatis dan tidak ada lagi uang yang terlewat.' },
  { initial: 'M', name: 'Bu Melati', role: 'Pemilik toko kosmetik · Medan', quote: 'Struk bisa dicetak atau dikirim digital. Pelanggan makin percaya dan toko terlihat profesional meskipun hanya toko kecil.' },
  { initial: 'R', name: 'Pak Rudi', role: 'Pemilik toko elektronik · Palembang', quote: 'Seminggu memakai, langsung terbiasa. Import produk dari Excel juga mudah, ratusan barang masuk sekaligus tanpa salah tulis.' },
]





export default function Landing() {

  useEffect(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      document.querySelectorAll('.reveal').forEach((el) => el.classList.add('in'))
      return
    }
    const io = new IntersectionObserver(
      (entries) => entries.forEach((e) => {
        if (e.isIntersecting) {
          e.target.classList.add('in')
          io.unobserve(e.target)
        }
      }),
      { threshold: 0.15 },
    )
    document.querySelectorAll('.reveal').forEach((el) => io.observe(el))
    return () => io.disconnect()
  }, [])

  return (
    <div className="landing-light bg-bg text-fg">
      <Navbar logoTone="light" mobileLandingMenu />
      <main>
        <section id="beranda" className="overflow-hidden pt-[clamp(36px,5vw,92px)] pb-10">
          <div className="container mx-auto max-w-6xl px-5 md:px-8 text-center">
            <h1 className="hero-reveal mx-auto mt-8 text-[clamp(30px,8.5vw,40px)] font-normal leading-[1.08] tracking-[-0.025em] sm:text-[clamp(40px,5.2vw,60px)]">
              Aplikasi kasir gratis<br />buat warung Anda.
            </h1>
            <p className="hero-reveal mx-auto mt-4 max-w-[520px] text-base leading-relaxed text-muted sm:mt-6 sm:text-lg">
              Kelola produk, stok, dan penjualan dari satu dashboard sederhana. Cocok untuk toko kecil dan menengah. Tanpa biaya langganan, gratis seumur hidup.
            </p>
            <div className="hero-reveal mt-6 flex flex-wrap justify-center gap-3 sm:mt-8">
              <Link to="/daftar" className="rounded-full bg-jet px-6 py-3 text-[15px] font-medium text-paper transition hover:bg-[color-mix(in_oklch,var(--t-jet)_82%,white)] active:translate-y-px sm:px-7.5 sm:py-3.5 sm:text-base">
                Buat toko pertama
              </Link>
              <Link to="/masuk" className="rounded-full border border-dove bg-transparent px-6 py-3 text-[15px] font-medium text-jet transition hover:border-jet hover:bg-fg/6 active:translate-y-px sm:px-7.5 sm:py-3.5 sm:text-base">
                Masuk
              </Link>
              <Link to="/unduh" className="rounded-full border border-dove bg-transparent px-6 py-3 text-[15px] font-medium text-jet transition hover:border-jet hover:bg-fg/6 active:translate-y-px sm:px-7.5 sm:py-3.5 sm:text-base">
                Unduh Offline
              </Link>
            </div>
          </div>

        </section>

        <section id="fitur" className="section border-t border-border">
          <div className="container mx-auto max-w-6xl px-5 md:px-8">
            <div className="reveal max-w-[680px]">
              <p className="font-mono text-xs uppercase tracking-[0.08em] text-steel">Fitur</p>
              <h2 className="mt-5 text-[clamp(30px,3.8vw,46px)] font-normal leading-[1.14] tracking-[-0.025em]">
                Fitur yang dibutuhkan toko kecil, tanpa yang berlebihan.
              </h2>
            </div>
            <div className="mt-14 grid gap-8 md:grid-cols-3">
              {[
                { mark: <path d="M13 2 4 14h6l-1 8 9-12h-6l1-8z" />, t: 'Kasir yang cepat', d: 'Cari produk lewat nama, SKU, atau barcode, lalu tambahkan ke keranjang, terima pembayaran, dan cetak struk. Satu transaksi selesai dalam hitungan detik.' },
                { mark: <path d="M12 2 4 6v12l8 4 8-4V6l-8-4zM4 6l8 4 8-4M12 10v10" />, t: 'Produk dan stok real-time', d: 'Stok berkurang otomatis setiap transaksi dan kembali saat refund. Anda tidak akan kehabisan stok tanpa sadar atau menumpuk barang yang tidak terjual.' },
                { mark: <path d="M3 21h18M6 17v-6M11.5 17V8M17 17v-9" />, t: 'Laporan yang rapi', d: 'Omzet harian, produk terlaris, dan profit terlihat langsung dari dashboard. Tidak perlu menghitung manual di buku kas.' },
              ].map((f, i) => (
                <div key={f.t} className="feature reveal group" data-delay={i}>
                  <span className="mb-5 grid h-9 w-9 place-items-center text-fog transition-colors duration-150 group-hover:text-jet">
                    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinejoin="round">{f.mark}</svg>
                  </span>
                  <h3 className="mb-1.5 text-2xl font-medium leading-[1.3] tracking-[-0.01em]">{f.t}</h3>
                  <p className="text-[15px] leading-relaxed text-muted">{f.d}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section id="cara-kerja" className="section border-t border-transparent bg-charcoal">
          <div className="container mx-auto max-w-6xl px-5 md:px-8">
            <div className="reveal max-w-[680px]">
              <p className="font-mono text-xs uppercase tracking-[0.08em] text-white/70">Cara Kerja</p>
              <h2 className="mt-5 text-[clamp(30px,3.8vw,46px)] font-normal leading-[1.14] tracking-[-0.025em] text-white">
                Dari mendaftar sampai transaksi pertama, hanya tiga langkah.
              </h2>
            </div>
            <div className="mt-14 grid gap-8 md:grid-cols-3">
              {[
                { n: '01', t: 'Buat akun dan toko', d: 'Daftar sekali. Akun admin dan data toko langsung dibuat bersamaan. Tanpa kartu kredit, tanpa biaya, dan tanpa masa percobaan.' },
                { n: '02', t: 'Tambah produk dan stok', d: 'Masukkan produk beserta harga dan stok awal, atau impor ratusan baris sekaligus lewat file CSV.' },
                { n: '03', t: 'Mulai berjualan', d: 'Buka menu kasir, cari produk, selesaikan pembayaran, lalu cetak struk. Stok otomatis terbarui di dashboard.' },
              ].map((s, i) => (
                <div key={s.n} className="step reveal" data-delay={i}>
                  <span className="mb-5 block font-mono text-[26px] text-white/60">{s.n}</span>
                  <h3 className="mb-1.5 text-2xl font-medium leading-[1.3] tracking-[-0.01em] text-white">{s.t}</h3>
                  <p className="text-[15px] leading-relaxed text-white/80">{s.d}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section id="cerita" className="section border-t border-border">
          <div className="container mx-auto max-w-6xl px-5 md:px-8">
            <div className="reveal mb-8 max-w-[680px]">
              <p className="font-mono text-xs uppercase tracking-[0.08em] text-steel">Kata mereka</p>
              <h2 className="mt-5 text-[clamp(30px,3.8vw,46px)] font-normal leading-[1.14] tracking-[-0.025em]">
                Dipakai toko-toko kecil di seluruh Indonesia.
              </h2>
            </div>
            <div className="marquee-wrap reveal" data-delay="1">
              <div className="marquee-track">
                {[...TESTIMONIALS, ...TESTIMONIALS].map((t, i) => (
                  <figure key={i} className="marquee-card w-80 flex-none rounded-2xl bg-cream p-5">
                    <div className="mb-2 text-[40px] leading-none text-fog" aria-hidden="true">&ldquo;</div>
                    <blockquote className="text-sm leading-relaxed">{t.quote}</blockquote>
                    <figcaption className="mt-4 flex items-center gap-2.5 text-sm text-steel">
                      <span className="grid h-8 w-8 flex-none place-items-center rounded-full bg-jet font-mono text-xs font-medium text-paper" aria-hidden="true">{t.initial}</span>
                      <span>
                        <span className="font-medium text-jet">{t.name}</span>
                        <span className="mt-0.5 block font-mono text-[11px] text-steel">{t.role}</span>
                      </span>
                    </figcaption>
                  </figure>
                ))}
              </div>
            </div>
          </div>
        </section>

        <section id="daftar" className="section border-t border-border py-16 text-center">
          <div className="container mx-auto max-w-[640px] px-5 md:px-8">
            <h2 className="reveal text-[clamp(30px,3.8vw,46px)] font-normal leading-[1.14] tracking-[-0.025em]">Mulai berjualan hari ini tanpa biaya langganan.</h2>
            <p className="reveal mx-auto mt-4 mb-8 max-w-[520px] text-lg leading-relaxed text-muted" data-delay="1">
              Daftar dalam satu menit. Buat toko, tambah produk, lalu terima pembayaran pertama Anda. Gratis selamanya.
            </p>
            <Link
              to="/daftar"
              className="reveal inline-block rounded-full bg-jet px-7.5 py-3.5 text-base font-medium text-paper transition hover:bg-[color-mix(in_oklch,var(--t-jet)_82%,white)] active:translate-y-px"
              data-delay="2"
            >
              Buat toko Anda sekarang
            </Link>
          </div>
        </section>
      </main>
      <Footer />
    </div>
  )
}

