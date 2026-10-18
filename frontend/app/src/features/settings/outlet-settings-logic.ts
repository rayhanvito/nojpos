import type { OutletSettingsOverrides } from '../../shared/api/types.ts'

export function hasOutletOverride<K extends keyof OutletSettingsOverrides>(overrides: OutletSettingsOverrides, field: K): boolean {
  return Object.prototype.hasOwnProperty.call(overrides, field)
}
