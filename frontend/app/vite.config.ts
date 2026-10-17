import path from 'node:path'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  define: { __APP_BUILD_ID__: JSON.stringify(Date.now().toString(36)) },
  resolve: {
    alias: { '@': path.resolve(import.meta.dirname, 'src') },
  },
  build: {
    rollupOptions: {
      input: {
        main: path.resolve(import.meta.dirname, 'index.html'),
        offline: path.resolve(import.meta.dirname, 'offline.html'),
      },
    },
  },
  server: {
    proxy: { '/api': 'http://localhost:8080' },
  },
})
