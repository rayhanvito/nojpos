import { useEffect, useState, type Dispatch, type SetStateAction } from 'react'
import { useSearchParams } from 'react-router'
import { normalizeTabParam } from './tab-params'

// Titik baca bersama untuk parameter ?tab=: membaca nilai awal dari URL,
// mengaktifkan tab kanonis, dan BILA nilai yang datang adalah alias lama,
// menormalkan URL ke nilai kanonis dengan replace (tanpa entri riwayat
// baru) sambil mempertahankan parameter lain persis. Klik tab sesudah itu
// hanya mengubah state lokal, sama seperti perilaku sebelum rename
// (URL tidak ditulis ulang saat klik).
export function useTabParam<T extends string>(
  canonical: readonly T[],
  aliases: Readonly<Record<string, T>>,
  fallback: T,
): [T, Dispatch<SetStateAction<T>>] {
  const [searchParams, setSearchParams] = useSearchParams()
  const [tab, setTab] = useState<T>(() => normalizeTabParam(searchParams.get('tab'), canonical, aliases, fallback).tab)

  const raw = searchParams.get('tab')
  const aliasedTo = raw ? normalizeTabParam(raw, canonical, aliases, fallback) : null
  useEffect(() => {
    if (aliasedTo?.aliased) {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev)
        next.set('tab', aliasedTo.tab)
        return next
      }, { replace: true })
    }
    // Hanya normalisasi saat mount dari URL awal; perubahan ?tab= eksternal
    // sesudah mount tidak dikejar (perilaku lama: tab dibaca sekali).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return [tab, setTab]
}
