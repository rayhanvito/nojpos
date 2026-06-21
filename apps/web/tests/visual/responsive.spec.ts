import { expect, type Page, test } from '@playwright/test';

const mobileShellRoutes = ['/dashboard', '/platform'] as const;
const mobileTableRoutes = ['/customers', '/transactions', '/payments', '/platform/businesses', '/platform/plans', '/platform/revenue', '/platform/support', '/platform/audit', '/platform/system-health', '/platform/announcements'] as const;
const desktopRoutes = ['/dashboard', '/transactions', '/platform', '/platform/businesses', '/platform/plans', '/platform/revenue', '/platform/support', '/platform/audit', '/platform/system-health', '/platform/announcements'] as const;

const horizontalOverflow = async (page: Page) =>
  page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

test.describe('NojPOS responsive preview QA', () => {
  test.describe('public landing route', () => {
    test('mobile / avoids horizontal overflow and keeps CTA visible', async ({ page }) => {
      await page.setViewportSize({ width: 390, height: 844 });
      await page.goto('/');
      await expect(page.getByRole('heading', { name: /POS modern untuk usaha Indonesia/i })).toBeVisible();
      await expect(page.getByRole('link', { name: /Lihat Demo Admin/i }).first()).toBeVisible();
      expect(await horizontalOverflow(page)).toBeLessThanOrEqual(1);
    });

    test('desktop / avoids horizontal overflow and renders key sections', async ({ page }) => {
      await page.setViewportSize({ width: 1440, height: 1000 });
      await page.goto('/');
      await expect(page.locator('#fitur')).toBeVisible();
      await expect(page.locator('#cara-kerja')).toBeVisible();
      await expect(page.locator('#harga')).toBeVisible();
      expect(await horizontalOverflow(page)).toBeLessThanOrEqual(1);
    });
  });

  test.describe('mobile shell routes', () => {
    test.use({ viewport: { width: 390, height: 844 }, isMobile: true });

    for (const pathname of mobileShellRoutes) {
      test(`${pathname} avoids horizontal overflow on mobile`, async ({ page }) => {
        await page.goto(pathname);
        await expect(page.getByRole('button', { name: /menu/i })).toBeVisible();
        await expect(page.locator('.sidebar')).toHaveCSS('position', 'fixed');
        expect(await horizontalOverflow(page)).toBeLessThanOrEqual(1);
      });
    }
  });

  test.describe('mobile table routes', () => {
    test.use({ viewport: { width: 390, height: 844 }, isMobile: true });

    for (const pathname of mobileTableRoutes) {
      test(`${pathname} keeps drawer responsive and avoids page overflow`, async ({ page }) => {
        await page.goto(pathname);
        await expect(page.getByRole('button', { name: /menu/i })).toBeVisible();
        await expect(page.locator('.sidebar')).toHaveCSS('position', 'fixed');
        await expect(page.locator('.responsive-table-mobile').first()).toBeVisible();
        await expect(page.locator('.mobile-row-card').first()).toBeVisible();
        expect(await horizontalOverflow(page)).toBeLessThanOrEqual(1);
      });
    }
  });

  test.describe('desktop routes', () => {
    test.use({ viewport: { width: 1440, height: 1000 } });

    for (const pathname of desktopRoutes) {
      test(`${pathname} keeps shell and tables inside viewport`, async ({ page }) => {
        await page.goto(pathname);
        await expect(page.locator('.sidebar')).toHaveCSS('position', 'fixed');
        await expect(page.locator('.table-wrap').first()).toBeVisible();
        expect(await horizontalOverflow(page)).toBeLessThanOrEqual(1);
      });
    }
  });
});
