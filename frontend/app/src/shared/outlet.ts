import type { Role } from './api'

export function shouldShowOutletSwitcher(role: Role, count: number): boolean {
  return role === 'admin' && count > 1
}

export function shouldShowCashierOutletLabel(role: Role, count: number): boolean {
  return role === 'cashier' && count > 1
}

export function shouldShowOutletFilter(role: Role, count: number): boolean {
  return role === 'admin' && count > 1
}

export function outletQuery(outletId?: string | null): { outletId?: string } {
  const id = outletId?.trim()
  return id ? { outletId: id } : {}
}

export function outletPayload(name: string, address: string): { name: string; address: string } {
  return { name: name.trim(), address: address.trim() }
}

export function userOutletPayload(name: string, outletId?: string): { name: string; outletId?: string } {
  const id = outletId?.trim()
  return { name: name.trim(), ...(id ? { outletId: id } : {}) }
}

export function outletRouteKey(outletId?: string | null): string {
  return outletId ? `outlet:${outletId}` : 'outlet:none'
}
