// Mode sheet bawah untuk popover bersama (Select, FilterPanel, pemilih
// tanggal) di viewport <640px; desktop memakai popover melayang biasa.

import { useEffect, useState } from 'react'

export function useSheetMode(): boolean {
  const [sheet, setSheet] = useState(() => typeof window !== 'undefined' && window.matchMedia('(max-width: 639px)').matches)
  useEffect(() => {
    const mq = window.matchMedia('(max-width: 639px)')
    const on = () => setSheet(mq.matches)
    mq.addEventListener('change', on)
    return () => mq.removeEventListener('change', on)
  }, [])
  return sheet
}

export const SHEET_BACKDROP = 'fixed inset-0 z-40 bg-black/50'
