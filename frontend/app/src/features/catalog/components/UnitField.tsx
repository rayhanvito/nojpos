import { useEffect, useMemo, useRef, useState } from 'react'
import { Check, ChevronDown, Plus } from 'lucide-react'
import { apiListProductUnits } from '../api'
import { mergeUnits } from '../units'

// Kolom Satuan: combobox yang mengelola daftarnya sendiri (satuan
// baku + yang sedang dipakai toko, dari GET /products/units).
// Membuka kolom menampilkan SELURUH daftar; penyaringan baru berlaku
// dari ketikan pengguna. Satuan baru bisa ditambah langsung di
// tempat lewat baris "Tambah …" atau cukup diketik lalu tinggalkan
// kolomnya — teks yang tertinggal di kolom adalah satuannya.
export function UnitField({ value, onChange }: {
  value: string
  onChange: (v: string) => void
}) {
  const [used, setUsed] = useState<string[]>([])
  const [open, setOpen] = useState(false)
  const [text, setText] = useState(value)
  const [filtering, setFiltering] = useState(false)
  const rootRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    let alive = true
    apiListProductUnits().then((u) => { if (alive) setUsed(u) }).catch(() => {})
    return () => { alive = false }
  }, [])

  // Nilai dari luar berubah (mis. ganti produk yang diedit) dan kolom
  // sedang tidak dibuka: teks mengikuti.
  useEffect(() => {
    if (!open) setText(value)
  }, [value, open])

  function close(commit: boolean) {
    if (commit && filtering) {
      const v = text.trim()
      if (v) onChange(v)
    }
    setText(commit && filtering && text.trim() ? text.trim() : value)
    setFiltering(false)
    setOpen(false)
  }

  useEffect(() => {
    if (!open) return
    const onDoc = (e: PointerEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) close(true)
    }
    document.addEventListener('pointerdown', onDoc)
    return () => document.removeEventListener('pointerdown', onDoc)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, filtering, text, value])

  const all = useMemo(() => mergeUnits(used), [used])
  const q = filtering ? text.trim().toLowerCase() : ''
  const matches = all.filter((u) => !q || u.toLowerCase().includes(q))
  const canAdd = filtering && q !== '' && !all.some((u) => u.toLowerCase() === q)

  function pick(u: string) {
    onChange(u)
    setText(u)
    setFiltering(false)
    setOpen(false)
  }

  return (
    <div ref={rootRef} className="relative flex flex-col gap-1.5">
      <span className="text-[13px] font-medium text-steel">Satuan</span>
      <div className="relative">
        <input
          value={text}
          placeholder="pcs"
          aria-label="Satuan"
          onChange={(e) => { setText(e.target.value); setFiltering(true); setOpen(true) }}
          onFocus={() => { setOpen(true); setFiltering(false) }}
          onKeyDown={(e) => {
            if (e.key === 'Escape') close(false)
            if (e.key === 'Enter' && open) {
              e.preventDefault()
              if (canAdd) pick(text.trim())
              else if (filtering && matches.length > 0) pick(matches[0])
              else setOpen(false)
            }
          }}
          className="h-[42px] w-full rounded-md border border-border bg-paper px-3.5 py-0 pr-10 text-[15px] text-fg placeholder:text-fog focus:border-jet focus:outline-2 focus:outline-accent-soft"
        />
        <button
          type="button"
          tabIndex={-1}
          aria-label="Tampilkan daftar satuan"
          onClick={() => { if (open) close(true); else { setOpen(true); setFiltering(false) } }}
          className="absolute inset-y-0 right-0 flex w-10 items-center justify-center text-fog"
        >
          <ChevronDown className="size-4" />
        </button>
      </div>
      {open && (matches.length > 0 || canAdd) && (
        <ul className="absolute inset-x-0 top-full z-20 mt-1 max-h-56 overflow-auto rounded-md border border-border bg-paper py-1 shadow-lg">
          {matches.map((u) => (
            <li key={u}>
              <button
                type="button"
                onClick={() => pick(u)}
                className="flex h-10 w-full items-center justify-between px-3.5 text-left text-[15px] text-fg hover:bg-surface"
              >
                {u}
                {u === value && <Check className="size-4 text-jet" />}
              </button>
            </li>
          ))}
          {canAdd && (
            <li>
              <button
                type="button"
                onClick={() => pick(text.trim())}
                className="flex h-10 w-full items-center gap-2 border-t border-dove px-3.5 text-left text-[15px] font-medium text-jet hover:bg-surface"
              >
                <Plus className="size-4" />Tambah &ldquo;{text.trim()}&rdquo;
              </button>
            </li>
          )}
        </ul>
      )}
    </div>
  )
}
