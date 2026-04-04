import { expect, test, type Page } from '@playwright/test';

const DEBATES_PATH = '/debates';

// Set viewport to ensure side panel is visible (lg: breakpoint is 1024px)
test.use({ viewport: { width: 1280, height: 720 } });

async function ensureRoundsVisible(page: Page) {
    const roundBubbles = page.locator('[data-testid="round-bubble"]');
    const initialCount = await roundBubbles.count();
    if (initialCount > 0) {
        return initialCount;
    }

    const startButton = page.getByRole('button', { name: 'Start' });
    await expect(startButton).toBeVisible();
    await startButton.click();
    await expect(roundBubbles.first()).toBeVisible({ timeout: 15000 });

    return roundBubbles.count();
}

test.describe('Detail Page', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto(DEBATES_PATH);
        await page.waitForLoadState('networkidle');

        const firstDebateCard = page
            .locator('button', {
                has: page.locator('h3'),
            })
            .first();

        await expect(firstDebateCard).toBeVisible();
        await firstDebateCard.click();
        await page.waitForLoadState('networkidle');
    });

    test('shows generation controls when round generation starts', async ({ page }) => {
        const startButton = page.getByRole('button', { name: 'Start' });
        await expect(startButton).toBeVisible();

        await startButton.click();

        await expect(page.getByRole('button', { name: 'Stop' })).toBeVisible({ timeout: 5000 });
    });

    test('adds rounds after clicking Start', async ({ page }) => {
        const initialRounds = await page.locator('[data-testid="round-bubble"]').count();

        const startButton = page.getByRole('button', { name: 'Start' });
        await expect(startButton).toBeVisible();
        await startButton.click();

        await expect(page.locator('[data-testid="round-bubble"]')).toHaveCount(initialRounds + 1, {
            timeout: 15000,
        });
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
        await ensureRoundsVisible(page);

        const openingBadge = page
            .locator('[data-testid="round-bubble"]')
            .filter({ hasText: 'Opening' });
        await expect(openingBadge.first()).toBeVisible({ timeout: 5000 });
    });

    test('round bubbles show metadata when available', async ({ page }) => {
        await ensureRoundsVisible(page);

        const metadataSection = page.getByTestId('round-metadata');
        const summaryMetadata = page.getByTestId('round-summary');
        const voiceMetadata = page.getByTestId('round-voice');

        await expect(metadataSection.first()).toBeVisible();
        await expect(summaryMetadata.first()).toBeVisible();
        await expect(voiceMetadata.first()).toBeVisible();
    });
});

test('shows a blocking error modal when round generation fails', async ({ page }) => {
    const mockedDebateId = 'round-error-fixture';

    await page.route('**/api/debates', async (route) => {
        if (route.request().method() !== 'GET') {
            await route.continue();
            return;
        }

        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify([
                {
                    id: mockedDebateId,
                    name: 'Round Error Fixture',
                    topic: 'Can mock debates verify UI regressions?',
                },
            ]),
        });
    });

    await page.route(`**/api/debates/${mockedDebateId}`, async (route) => {
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
                name: 'Round Error Fixture',
                normalized_name: 'round-error-fixture',
                topic: 'Can mock debates verify UI regressions?',
                agents: [
                    {
                        id: 'agent-1',
                        name: 'Ada',
                        stance: 'Yes',
                        voice_name: 'alloy',
                    },
                    {
                        id: 'agent-2',
                        name: 'Linus',
                        stance: 'Sometimes',
                        voice_name: 'verse',
                    },
                ],
                rounds: [],
                tts_provider: 'native',
            }),
        });
    });

    await page.route(`**/api/debates/${mockedDebateId}/rounds`, async (route) => {
        if (route.request().method() !== 'POST') {
            await route.continue();
            return;
        }

        await route.fulfill({
            status: 400,
            contentType: 'application/json',
            body: JSON.stringify({ error: 'Invalid API key for the configured LLM provider.' }),
        });
    });

    await page.goto(`/#/debates/${mockedDebateId}`);
    await page.waitForLoadState('networkidle');

    const startButton = page.getByRole('button', { name: 'Start' });
    await expect(startButton).toBeVisible();
    await startButton.click();

    await expect(page.getByRole('alertdialog')).toBeVisible();
    await expect(page.getByText('Round generation failed')).toBeVisible();
    await expect(page.getByText('Invalid API key for the configured LLM provider.')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Open settings' })).toBeVisible();
});
