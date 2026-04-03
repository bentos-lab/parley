import { expect, test } from '@playwright/test';

const DEBATE_ID = 'ai-existential-risk.2026-01-10-09-00-00';
const DEBATE_PATH = `/debates/${encodeURIComponent(DEBATE_ID)}`;
const EDIT_PATH = `/debates/${encodeURIComponent(DEBATE_ID)}/edit`;
const AUDIO_PATH = `/debates/${encodeURIComponent(DEBATE_ID)}/audio`;

// NOTE: These tests require the backend API to be running.
// Skip until MSW mocks are implemented in the frontend.
test.describe('Phase 6 edit debate', () => {
    test.skip('TEST-03: edit topic and agent name, then save to detail route', async ({ page }) => {
        const updatedTopic =
            'Artificial intelligence presents a measurable existential risk this century';
        const updatedAgentName = 'Dr. Aria Chen, PhD';

        await page.goto(EDIT_PATH);

        await expect(
            page.getByRole('heading', { name: /edit ai existential risk/i }),
        ).toBeVisible();

        await page.getByLabel(/topic/i).fill(updatedTopic);
        await page.getByLabel(/agent 1 name/i).fill(updatedAgentName);
        await page.getByTestId('edit-save-button').click();

        await page.waitForURL(DEBATE_PATH);
        await expect(
            page.getByRole('banner').getByText(updatedTopic, { exact: true }),
        ).toBeVisible();
        await expect(page.getByText(updatedAgentName).first()).toBeVisible();
    });

    test.skip('TEST-04: dirty edits trigger blocker before navigation away', async ({ page }) => {
        await page.goto(EDIT_PATH);

        await page.getByLabel(/topic/i).fill('Unsaved blocker test topic');
        await page.getByRole('link', { name: /^home$/i }).click();

        await expect(page.getByRole('dialog', { name: /leave this draft/i })).toBeVisible();
        await page.getByRole('button', { name: /leave and discard/i }).click();

        await page.waitForURL('/debates');
        // Sidebar shows debate name only (no agent count in DebateSummary)
        await expect(
            page.locator('aside').getByRole('button', { name: /ai existential risk/i }),
        ).toBeVisible();
    });

    test.skip('audio route has a back button to the debate detail view', async ({ page }) => {
        await page.goto(AUDIO_PATH);

        await expect(page.getByRole('button', { name: /back to debate/i }).first()).toBeVisible();
        await page
            .getByRole('button', { name: /back to debate/i })
            .first()
            .click();

        await page.waitForURL(DEBATE_PATH);
        await expect(page.getByTestId('sse-status-bar')).toBeVisible();
    });
});
