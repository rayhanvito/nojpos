import { describe, expect, it } from 'vitest';

import {
  attendanceDetailPreviews,
  attendanceTable,
  auditDetailPreviews,
  auditTrailTable,
  catalogDetailPreviews,
  catalogProductsTable,
  customerDetailPreviews,
  customersTable,
  inventoryDetailPreviews,
  inventoryMetrics,
  outletDetailPreviews,
  outletsTable,
  paymentDetailPreviews,
  paymentsTable,
  platformAuditTable,
  platformBusinessDetailPreviews,
  platformBusinessesTable,
  platformOverviewTable,
  platformPlansTable,
  platformRevenueTable,
  platformSubscriptionsTable,
  platformSystemHealthTable,
  platformUsersTable,
  previewActionGroups,
  previewStateMatrix,
  promotionDetailPreviews,
  promotionTable,
  reportResources,
  requiredPreviewStateTones,
  settingsResources,
  settingsSections,
  settingsSummaryTable,
  staffDetailPreviews,
  staffDirectoryTable,
  subscriptionMetrics,
  subscriptionTable,
  transactionDetailPreviews,
  transactionsTable,
} from '.';

const tables = [
  catalogProductsTable,
  settingsSummaryTable,
  staffDirectoryTable,
  promotionTable,
  auditTrailTable,
  customersTable,
  attendanceTable,
  outletsTable,
  transactionsTable,
  paymentsTable,
  subscriptionTable,
  platformOverviewTable,
  platformBusinessesTable,
  platformPlansTable,
  platformSubscriptionsTable,
  platformRevenueTable,
  platformUsersTable,
  platformAuditTable,
  platformSystemHealthTable,
] as const;

const detailFixtures = [
  ...catalogDetailPreviews,
  ...inventoryDetailPreviews,
  ...staffDetailPreviews,
  ...promotionDetailPreviews,
  ...auditDetailPreviews,
  ...customerDetailPreviews,
  ...attendanceDetailPreviews,
  ...outletDetailPreviews,
  ...transactionDetailPreviews,
  ...paymentDetailPreviews,
  ...platformBusinessDetailPreviews,
] as const;

const requiredFixtureModules = [
  'dashboard',
  'catalog',
  'inventory',
  'reports',
  'settings',
  'staff',
  'promotions',
  'audit',
  'customers',
  'attendance',
  'outlets',
  'transactions',
  'payments',
  'subscription',
  'platform',
] as const;

describe('preview fixtures', () => {
  it('keeps preview table rows aligned with columns', () => {
    for (const table of tables) {
      for (const row of table.rows) {
        expect(row).toHaveLength(table.columns.length);
      }
    }
  });

  it('keeps metrics as placeholder-safe preview values', () => {
    const allMetrics = [...inventoryMetrics, ...subscriptionMetrics];
    const safeValueTerms = ['—', 'Read-only', 'Menunggu data', 'Belum terhubung', 'Mode contoh', 'Menunggu sinkronisasi'];

    expect(allMetrics.every((metric) => safeValueTerms.some((term) => metric.value.includes(term)))).toBe(true);
  });

  it('keeps required fixture modules present in state and action matrices', () => {
    for (const fixtureModule of requiredFixtureModules) {
      expect(previewStateMatrix[fixtureModule], fixtureModule).toBeDefined();
      expect(previewActionGroups[fixtureModule], fixtureModule).toBeDefined();
    }
  });

  it('keeps settings detail routes mirrored in overview resources', () => {
    const resourcePaths = new Set(settingsResources.map((resource) => resource.path));

    for (const section of settingsSections) {
      expect(resourcePaths.has(section.path)).toBe(true);
      expect(section.fields.length).toBeGreaterThan(0);
    }
  });

  it('keeps report resources route-like and read-only', () => {
    expect(reportResources.every((resource) => resource.path.startsWith('/reports/'))).toBe(true);
  });

  it('keeps every module covered by the required preview state tones', () => {
    for (const [fixtureModule, states] of Object.entries(previewStateMatrix)) {
      expect(states).toHaveLength(requiredPreviewStateTones.length);
      expect(states.map((state) => state.tone).sort(), fixtureModule).toEqual([...requiredPreviewStateTones].sort());
      expect(states.every((state) => state.title.length > 0 && state.message.length > 0), fixtureModule).toBe(true);
    }
  });

  it('keeps action groups locked with user-friendly server reason before API integration', () => {
    const safeReasonTerms = ['menunggu', 'membutuhkan', 'dikunci', 'server', 'audit', 'izin'];

    for (const [fixtureModule, group] of Object.entries(previewActionGroups)) {
      expect(group.title.length, fixtureModule).toBeGreaterThan(0);
      expect(group.description.length, fixtureModule).toBeGreaterThan(0);
      expect(group.actions.length, fixtureModule).toBeGreaterThan(0);

      for (const action of group.actions) {
        const reason = action.reason.toLowerCase();

        expect(action.label.length, `${fixtureModule}:${action.label}`).toBeGreaterThan(0);
        expect(safeReasonTerms.some((term) => reason.includes(term)), `${fixtureModule}:${action.label}`).toBe(true);
        expect(['read', 'write', 'import', 'export', 'danger']).toContain(action.category);
      }
    }
  });

  it('keeps detail previews static, linked, and action-gated', () => {
    const ids = new Set<string>();

    for (const detail of detailFixtures) {
      expect(ids.has(detail.id), detail.id).toBe(false);
      ids.add(detail.id);
      expect(detail.backHref.startsWith('/')).toBe(true);
      expect(detail.sections.length).toBeGreaterThan(0);
      expect(detail.timeline.length).toBeGreaterThan(0);
      expect(detail.actions.every((action) => action.status.length > 0)).toBe(true);
      expect(detail.sections.every((section) => section.rows.length > 0)).toBe(true);
    }
  });
});
