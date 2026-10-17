/// <reference types="vite/client" />

declare const __APP_BUILD_ID__: string

interface Window {
  offline?: {
    isElectron: boolean
    close: () => void
    saveFile: (filename: string, data: ArrayBuffer) => Promise<{ saved: boolean; path?: string }>
  }
}
