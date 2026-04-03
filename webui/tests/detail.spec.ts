import { expect, test } from '@playwright/test';

const DEBATE_ID = 'ai-existential-risk.2026-01-10-09-00-00';
const DEBATE_PATH = `/debates/${encodeURIComponent(DEBATE_ID)}`;

// Set viewport to ensure side panel is visible (lg: breakpoint is 1024px)
test.use({ viewport: { width: 1280, height: 720 } });

test.describe('Detail Page', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto(DEBATE_PATH);
        await page.waitForLoadState('networkidle');
    });

    test('shows SSE status bar when streaming starts', async ({ page }) => {
        // Status bar is hidden when idle
        await expect(page.getByTestId('sse-status-bar')).toBeHidden();

        // Click Start button if visible
        const startButton = page.getByRole('button', { name: 'Start' });
        if (await startButton.isVisible()) {
            await startButton.click();
            // Status bar should become visible
            await expect(page.getByTestId('sse-status-bar')).toBeVisible({ timeout: 3000 });
        }
    });

    test('streams rounds after clicking Start', async ({ page }) => {
        const initialRounds = await page.locator('[data-testid="round-bubble"]').count();

        // Click Start button
        const startButton = page.getByRole('button', { name: 'Start' });
        if (await startButton.isVisible()) {
            await startButton.click();

            // Wait for at least one new round
            await expect(page.locator('[data-testid="round-bubble"]')).toHaveCount(
                initialRounds + 1,
                { timeout: 15000 },
            );
        }
    });

    // NOTE: This test requires backend API to run. Skip until MSW mocks are implemented.
    // The feature (click-to-highlight agent cards) is implemented in AgentPanel.tsx
    test.skip('agent cards are visible and clickable', async ({ page }) => {
        const agentPanel = page.getByTestId('agent-panel');
        await expect(agentPanel).toBeVisible({ timeout: 5000 });

        const agentCard = page.getByTestId('agent-card').first();
        await expect(agentCard).toBeVisible({ timeout: 5000 });

        // Click to highlight (adds border-2 class)
        await agentCard.click();
        await expect(agentCard).toHaveClass(/border-2/);

        // Click again to clear highlight
        await agentCard.click();
        await expect(agentCard).not.toHaveClass(/border-2/);
    });

    test('round bubbles show message type badges', async ({ page }) => {
        // Ensure we have rounds (start streaming if needed)
        let hasRounds = (await page.locator('[data-testid="round-bubble"]').count()) > 0;

        if (!hasRounds) {
            const startButton = page.getByRole('button', { name: 'Start' });
            if (await startButton.isVisible()) {
                await startButton.click();
                await page.waitForSelector('[data-testid="round-bubble"]', { timeout: 15000 });
                hasRounds = true;
            }
        }

        if (hasRounds) {
            // First round should show "Opening" badge
            const openingBadge = page
                .locator('[data-testid="round-bubble"]')
                .filter({ hasText: 'Opening' });
            await expect(openingBadge.first()).toBeVisible({ timeout: 5000 });
        }
    });
});
