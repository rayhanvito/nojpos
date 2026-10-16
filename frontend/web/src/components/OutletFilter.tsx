import type { Outlet } from '../lib/api'
import { useDB } from '../lib/store'
import { shouldShowOutletFilter } from '../lib/outlet'
import { Select } from '../lib/ui'

export function OutletFilter({ outlets, value, onChange }: {
  outlets: Outlet[]
  value: string
  onChange: (outletId: string) => void
}) {
  const { session } = useDB()
  if (!session || !shouldShowOutletFilter(session.role, outlets.length)) return null

  return (
    <Select
      label="Filter outlet"
      value={value}
      onChange={onChange}
      className="w-full sm:w-44"
      options={[{ value: '', label: 'Semua outlet' }, ...outlets.map((outlet) => ({ value: String(outlet.id), label: outlet.name }))]}
    />
  )
}
