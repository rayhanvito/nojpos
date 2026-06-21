import { existsSync, readFileSync } from 'node:fs';
import { join } from 'node:path';

import { describe, expect, it } from 'vitest';

const requiredRoutes = [
  ['/', 'src/app/page.tsx'],
  ['/login', 'src/app/login/page.tsx'],
  ['/dashboard', 'src/app/dashboard/page.tsx'],
  ['/catalog', 'src/app/catalog/page.tsx'],
  ['/catalog/[productId]', 'src/app/catalog/[productId]/page.tsx'],
  ['/inventory', 'src/app/inventory/page.tsx'],
  ['/inventory/[resourceId]', 'src/app/inventory/[resourceId]/page.tsx'],
  ['/reports', 'src/app/reports/page.tsx'],
  ['/settings', 'src/app/settings/page.tsx'],
  ['/settings/business', 'src/app/settings/business/page.tsx'],
  ['/settings/outlet', 'src/app/settings/outlet/page.tsx'],
  ['/settings/payments', 'src/app/settings/payments/page.tsx'],
  ['/settings/receipt', 'src/app/settings/receipt/page.tsx'],
  ['/staff', 'src/app/staff/page.tsx'],
  ['/staff/[staffId]', 'src/app/staff/[staffId]/page.tsx'],
  ['/promotions', 'src/app/promotions/page.tsx'],
  ['/promotions/[promotionId]', 'src/app/promotions/[promotionId]/page.tsx'],
  ['/audit', 'src/app/audit/page.tsx'],
  ['/audit/[eventId]', 'src/app/audit/[eventId]/page.tsx'],
  ['/customers', 'src/app/customers/page.tsx'],
  ['/customers/[customerId]', 'src/app/customers/[customerId]/page.tsx'],
  ['/attendance', 'src/app/attendance/page.tsx'],
  ['/attendance/[recordId]', 'src/app/attendance/[recordId]/page.tsx'],
  ['/outlets', 'src/app/outlets/page.tsx'],
  ['/outlets/[outletId]', 'src/app/outlets/[outletId]/page.tsx'],
  ['/transactions', 'src/app/transactions/page.tsx'],
  ['/transactions/[transactionId]', 'src/app/transactions/[transactionId]/page.tsx'],
  ['/payments', 'src/app/payments/page.tsx'],
  ['/payments/[paymentId]', 'src/app/payments/[paymentId]/page.tsx'],
  ['/subscription', 'src/app/subscription/page.tsx'],
  ['/platform', 'src/app/platform/page.tsx'],
  ['/platform/businesses', 'src/app/platform/businesses/page.tsx'],
  ['/platform/businesses/[businessId]', 'src/app/platform/businesses/[businessId]/page.tsx'],
  ['/platform/plans', 'src/app/platform/plans/page.tsx'],
  ['/platform/subscriptions', 'src/app/platform/subscriptions/page.tsx'],
  ['/platform/revenue', 'src/app/platform/revenue/page.tsx'],
  ['/platform/support', 'src/app/platform/support/page.tsx'],
  ['/platform/users', 'src/app/platform/users/page.tsx'],
  ['/platform/audit', 'src/app/platform/audit/page.tsx'],
  ['/platform/system-health', 'src/app/platform/system-health/page.tsx'],
  ['/platform/announcements', 'src/app/platform/announcements/page.tsx'],
] as const;

const stateMatrixRoutes = [
  '/catalog',
  '/catalog/[productId]',
  '/inventory',
  '/inventory/[resourceId]',
  '/reports',
  '/settings',
  '/staff',
  '/staff/[staffId]',
  '/promotions',
  '/promotions/[promotionId]',
  '/audit',
  '/audit/[eventId]',
  '/customers',
  '/customers/[customerId]',
  '/attendance',
  '/attendance/[recordId]',
  '/outlets',
  '/outlets/[outletId]',
  '/transactions',
  '/transactions/[transactionId]',
  '/payments',
  '/payments/[paymentId]',
  '/subscription',
  '/platform',
  '/platform/businesses',
  '/platform/businesses/[businessId]',
  '/platform/plans',
  '/platform/subscriptions',
  '/platform/revenue',
  '/platform/support',
  '/platform/users',
  '/platform/audit',
  '/platform/system-health',
  '/platform/announcements',
] as const;

describe('admin web route manifest', () => {
  it('keeps all approved preview routes present', () => {
    for (const [route, file] of requiredRoutes) {
      expect(existsSync(join(process.cwd(), file)), route).toBe(true);
    }
  });

  it('keeps root wired as the public landing page instead of a dashboard redirect', () => {
    const source = readFileSync(join(process.cwd(), 'src/app/page.tsx'), 'utf8');
    expect(source).toContain('LandingHero');
    expect(source).not.toContain("redirect('/dashboard')");
    expect(source).not.toContain('redirect("/dashboard")');
  });

  it('keeps dashboard wired through the approved BFF read-only boundary with fallback safety copy', () => {
    const source = readFileSync(join(process.cwd(), 'src/app/dashboard/page.tsx'), 'utf8');

    expect(source).toContain('getDashboardPageModel');
    expect(source).toContain('data contoh fallback');
    expect(source).toContain('Dashboard ini tidak menjalankan aksi sensitif');
    expect(source).not.toContain('localStorage');
    expect(source).not.toContain('sessionStorage');
    expect(source).not.toContain('Authorization');
    expect(source).not.toContain('Bearer');
  });

  it('keeps primary tenant and platform routes wired to preview state coverage', () => {
    for (const route of stateMatrixRoutes) {
      const file = requiredRoutes.find(([current]) => current === route)?.[1];
      expect(file, route).toBeDefined();

      const source = readFileSync(join(process.cwd(), file as string), 'utf8');
      const usesStateBoard = source.includes('PreviewStateBoard') || source.includes('TenantPreviewPage') || source.includes('PlatformPreviewPage');
      expect(usesStateBoard, route).toBe(true);
      expect(source, route).toContain('previewStateMatrix');
    }
  });

  it('keeps platform routes on a separate shell and out of the tenant sidebar', () => {
    const tenantShell = readFileSync(join(process.cwd(), 'src/components/admin-shell.tsx'), 'utf8');
    expect(tenantShell).not.toContain('/platform');

    const platformShell = readFileSync(join(process.cwd(), 'src/components/platform-shell.tsx'), 'utf8');
    expect(platformShell).toContain('/platform/businesses');
    expect(platformShell).toContain('Platform Super Admin');
  });
});
