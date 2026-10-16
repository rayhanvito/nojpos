import { useCallback, useEffect, useState } from 'react'
import { apiListOutlets, hasToken, type Outlet } from '../lib/api'

export function useOutlets(): { outlets: Outlet[]; loading: boolean; error: string; reload(): void } {
  const [outlets, setOutlets] = useState<Outlet[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [revision, setRevision] = useState(0)

  const reload = useCallback(() => {
    window.dispatchEvent(new Event('op:outlets-changed'))
  }, [])

  useEffect(() => {
    const refresh = () => setRevision((value) => value + 1)
    window.addEventListener('op:outlets-changed', refresh)
    return () => window.removeEventListener('op:outlets-changed', refresh)
  }, [])

  useEffect(() => {
    let dead = false
    if (!hasToken()) {
      setOutlets([])
      setError('')
      setLoading(false)
      return
    }
    setLoading(true)
    setError('')
    apiListOutlets()
      .then((items) => { if (!dead) setOutlets(items) })
      .catch((e) => { if (!dead) setError(e instanceof Error ? e.message : 'Gagal memuat daftar outlet.') })
      .finally(() => { if (!dead) setLoading(false) })
    return () => { dead = true }
  }, [revision])

  return { outlets, loading, error, reload }
}
