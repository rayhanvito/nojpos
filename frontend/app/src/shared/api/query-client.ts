// Singleton QueryClient TanStack Query untuk seluruh aplikasi.
// Dipakai bersama oleh QueryClientProvider di src/main.tsx dan titik
// pembersihan cache di AppShell (queryClient.clear() di titik yang sama
// dengan clearCache() lama) agar tidak ada dua instans yang berbeda.
import { QueryClient } from '@tanstack/react-query'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
})
