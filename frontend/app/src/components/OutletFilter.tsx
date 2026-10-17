import type { Outlet } from '../shared/api'
import { useDB } from '../shared/session/store'
import { shouldShowOutletFilter } from '../shared/outlet'
import { Select } from '../shared/ui'

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
