import { expect, test } from '@playwright/test';

const DEBATE_ID = 'ai-existential-risk.2026-01-10-09-00-00';
const AUDIO_PATH = `/debates/${encodeURIComponent(DEBATE_ID)}/audio`;

test.describe('Audio Studio — mini chapter controls', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto(AUDIO_PATH);
        await page.waitForLoadState('networkidle');
    });

    test('renders skip-back and skip-forward buttons for each chapter row', async ({ page }) => {
        // At least one chapter row must be present for a debate with rounds
        const skipBackButtons = page.getByRole('button', { name: 'Skip back 15 seconds' });
        const skipForwardButtons = page.getByRole('button', { name: 'Skip forward 15 seconds' });

        await expect(skipBackButtons.first()).toBeVisible();
        await expect(skipForwardButtons.first()).toBeVisible();
    });

    test('skip-back and skip-forward buttons are disabled before playback starts', async ({
        page,
    }) => {
        // Before any round is played the buttons must be disabled (isActive is false)
        const skipBackButton = page.getByRole('button', { name: 'Skip back 15 seconds' }).first();
        const skipForwardButton = page
            .getByRole('button', { name: 'Skip forward 15 seconds' })
            .first();

        await expect(skipBackButton).toBeDisabled();
        await expect(skipForwardButton).toBeDisabled();
    });

    test('each chapter row has a play button alongside the seek controls', async ({ page }) => {
        // The play button aria-label contains "Play round N"
        const playButton = page.getByRole('button', { name: /Play round \d+/ }).first();
        await expect(playButton).toBeVisible();
    });
});
