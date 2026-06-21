import Image from 'next/image';
import Link from 'next/link';

export default function LoginPage() {
  return (
    <main className="login">
      <section className="login-visual">
        <Link className="brand" href="/dashboard" aria-label="NojPOS Admin preview">
          <span className="brand-mark"><Image src="/assets/nojpos/logo/logo-symbol.png" alt="" width={1254} height={1254} /></span>
          <span className="brand-copy">
            <strong>NojPOS</strong>
            <small>Admin Preview</small>
          </span>
        </Link>
        <h1>Admin surface bergaya premium template.</h1>
        <p>
          Login ini hanya entrance ke preview desain. Browser session, token, dan auth persistence belum diaktifkan
          sampai backend integration gate disetujui.
        </p>
        <ul className="notice-list">
          <li>Visual direction dan logo mark diambil selektif dari folder template premium lokal.</li>
          <li>Semua field disabled agar tidak tampak seperti login produksi.</li>
          <li>Halaman preview dapat dibuka tanpa request ke Laravel.</li>
        </ul>
        <div className="login-device" aria-label="Dashboard mockup preview">
          <div className="device-top"><span /><span /><span /></div>
          <div className="device-grid"><span /><span /><span /><span /></div>
          <div className="device-chart"><i /><i /><i /><i /></div>
        </div>
      </section>

      <section className="login-panel" aria-label="Preview login panel">
        <span className="badge neutral">UI Preview</span>
        <h2>Masuk ke admin preview</h2>
        <p>Gunakan tombol di bawah untuk membuka dashboard desain. Form ini sengaja nonaktif.</p>
        <div className="login-form">
          <label>
            Email
            <input placeholder="owner@nojpos.test" disabled />
          </label>
          <label>
            Kode akses
            <input placeholder="••••••••" disabled />
          </label>
        </div>
        <Link className="btn primary" href="/dashboard">Buka preview admin</Link>
      </section>
    </main>
  );
}
