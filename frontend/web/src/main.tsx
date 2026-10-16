import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router'
import './index.css'
import App from './App.tsx'

if ('serviceWorker' in navigator) {
  const workerUrl = new URL(`${import.meta.env.BASE_URL}sw.js`, window.location.origin)
  workerUrl.searchParams.set('build', __APP_BUILD_ID__)
  void navigator.serviceWorker.register(workerUrl).catch(() => {})
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
)
