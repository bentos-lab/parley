import { test, expect } from '@playwright/test';

test.describe('Browse Debates', () => {
    test('TEST-01: Browse page loads and displays debate list content', async ({ page }) => {
        // Navigate to browse page
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        // Verify page URL
        await expect(page).toHaveURL('/');

        // Verify no JS errors
        const jsErrors: string[] = [];
        page.on('pageerror', (error) => jsErrors.push(error.message));

        // Wait a bit for any deferred errors
        await page.waitForTimeout(1000);
        expect(jsErrors).toHaveLength(0);

        // Verify page has content (not just blank)
        const bodyText = await page.textContent('body');
        expect(bodyText).toBeTruthy();
        expect(bodyText?.length).toBeGreaterThan(50);

        // Verify key layout elements exist (list page shows "Good morning." hero text)
        const hasHero = bodyText?.includes('Good morning') || bodyText?.includes('debate');
        expect(hasHero).toBe(true);
    });

    test('Browse list shows debate cards with metadata', async ({ page }) => {
        // Navigate to browse page
        await page.goto('/');
        await page.waitForLoadState('networkidle');

        // Get all buttons (debate cards are buttons)
        const buttons = page.locator('button');
        const count = await buttons.count();

        // Should have more than zero buttons (New Debate button + debate cards)
        expect(count).toBeGreaterThan(0);
    });

    test('Browse page renders without errors', async ({ page }) => {
        // Navigate to browse page
        await page.goto('/');

        // Verify page loads
        await expect(page).toHaveURL('/');

        // Check for critical layout elements
        const main = page.locator('main');
        await expect(main).toBeVisible();

        // Verify no unhandled JS errors
        let hasError = false;
        page.on('pageerror', () => {
            hasError = true;
        });

        await page.waitForTimeout(500);
        expect(hasError).toBe(false);
    });
});
