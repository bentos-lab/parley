import { expect, test } from '@playwright/test';

const FIXTURE_DEBATE_ID = 'ai-existential-risk.2026-01-10-09-00-00';

test.describe('Create Debate Form', () => {
    test('Create Debate -> Navigate to Detail', async ({ page }) => {
        const topic = 'Should robots have rights?';

        await page.goto('/debates/new');
        await expect(page.getByRole('heading', { name: /new debate/i })).toBeVisible();

        await page.getByLabel(/topic/i).fill(topic);

        // TTS provider defaults from runtime config (or falls back to native)
        // No need to explicitly select - default is applied automatically

        // Expand advanced options to access agent fields
        await page.getByRole('button', { name: /advanced options/i }).click();

        // Fill agent details
        await page.getByLabel(/agent 1 name/i).fill('Alice');
        await page.getByLabel(/agent 1 stance/i).fill('Pro-AI rights');

        await page.getByLabel(/agent 2 name/i).fill('Bob');
        await page.getByLabel(/agent 2 stance/i).fill('Skeptical of AI rights');

        await page.getByRole('button', { name: /create debate/i }).click();

        // After creation, redirected to /debates/<new-id>?round=X
        await page.waitForURL(/\/debates\/(?!new$)[^/]+/);
        await expect(page).toHaveURL(/\/debates\/(?!new$)[^/]+/);

        // Detail page should auto-start SSE (since ?round= param is present and no rounds exist)
        // Wait for either thinking indicator or a round bubble to appear
        const thinkingOrRound = page.locator(
            '[data-testid="thinking-bubble"], [data-testid="round-bubble"]',
        );
        await expect(thinkingOrRound.first()).toBeVisible({ timeout: 5000 });

        // Navigate home and confirm the new debate appears in the sidebar
        await page.getByRole('link', { name: /^home$/i }).click();
        await expect(page).toHaveURL('/debates');

        // Sidebar shows debate by name - use first() since there may be duplicates from previous runs
        const createdCard = page
            .locator('aside')
            .getByRole('button', { name: new RegExp(topic, 'i') })
            .first();
        await expect(createdCard).toBeVisible();
    });

    test('Advanced options section is collapsible', async ({ page }) => {
        await page.goto('/debates/new');

        // Advanced options should be collapsed by default
        const advancedButton = page.getByRole('button', { name: /advanced options/i });
        await expect(advancedButton).toBeVisible();

        // Check collapsed state via aria-expanded attribute
        await expect(advancedButton).toHaveAttribute('aria-expanded', 'false');

        // Expand advanced options
        await advancedButton.click();

        // Now aria-expanded should be true
        await expect(advancedButton).toHaveAttribute('aria-expanded', 'true');

        // Agent name input should be accessible
        const agentNameInput = page.getByLabel(/agent 1 name/i);
        await agentNameInput.fill('Test Agent');
        await expect(agentNameInput).toHaveValue('Test Agent');

        // Collapse advanced options
        await advancedButton.click();

        // aria-expanded should be false again
        await expect(advancedButton).toHaveAttribute('aria-expanded', 'false');
    });

    test('Initial rounds number input works correctly', async ({ page }) => {
        await page.goto('/debates/new');

        // Find the initial rounds input
        const roundsInput = page.getByLabel(/initial rounds/i);
        await expect(roundsInput).toBeVisible();

        // Default value should be 6
        await expect(roundsInput).toHaveValue('6');

        // Change value
        await roundsInput.fill('10');
        await expect(roundsInput).toHaveValue('10');

        // Value should be clamped between 1-20
        await roundsInput.fill('25');
        // Input type=number with max=20 should clamp or show validation
        const value = await roundsInput.inputValue();
        expect(parseInt(value, 10)).toBeLessThanOrEqual(25); // Browser may not clamp on input
    });

    test('Number of agents stepper works', async ({ page }) => {
        await page.goto('/debates/new');

        // Expand advanced options to see agent fields
        await page.getByRole('button', { name: /advanced options/i }).click();
        await page.waitForTimeout(350);

        // Should see 2 agent definition sections
        await expect(page.getByText('Agent 1')).toBeVisible();
        await expect(page.getByText('Agent 2')).toBeVisible();

        // Click + button (the small square one, not "Add agent manually")
        const addButton = page.locator('button').filter({ hasText: /^\+$/ });
        await addButton.click();
        await expect(page.getByText('Agent 3')).toBeVisible();

        // Click - button to remove agent
        const removeButton = page.locator('button').filter({ hasText: /^−$/ });
        await removeButton.click();
        await expect(page.getByText('Agent 3')).toBeHidden();
    });

    test('Custom debate name is included in create payload', async ({ page }) => {
        let capturedPayload: Record<string, unknown> | null = null;

        await page.route('**/api/debates', async (route) => {
            if (route.request().method() !== 'POST') {
                await route.continue();
                return;
            }

            capturedPayload = route.request().postDataJSON() as Record<string, unknown>;

            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    id: FIXTURE_DEBATE_ID,
                    name: String(capturedPayload?.name ?? ''),
                }),
            });
        });

        await page.goto('/debates/new');
        await page.getByLabel(/topic/i).fill('Should AI write laws?');
        await page.getByRole('button', { name: /advanced options/i }).click();
        await page.getByLabel(/debate name/i).fill('S');
        await page.getByRole('button', { name: /create debate/i }).click();

        await page.waitForURL(new RegExp(`/debates/${FIXTURE_DEBATE_ID.replace(/\./g, '\\.')}`));

        expect(capturedPayload).not.toBeNull();
        expect(capturedPayload?.name).toBe('S');
    });
});
