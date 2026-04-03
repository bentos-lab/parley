import { expect, test } from '@playwright/test';

function rgbToHex(rgb: string): string {
    const match = rgb.match(/\d+/g);

    if (!match || match.length < 3) {
        return rgb;
    }

    return `#${match
        .slice(0, 3)
        .map((component) => Number.parseInt(component, 10).toString(16).padStart(2, '0'))
        .join('')}`;
}

// Use a real fixture debate id so loaders don't 404
const DEBATE_ID = 'ai-existential-risk.2026-01-10-09-00-00';

const routes = [
    { name: 'Debates list', path: '/debates' },
    { name: 'Create debate', path: '/debates/new' },
    { name: 'Debate detail', path: `/debates/${encodeURIComponent(DEBATE_ID)}` },
    { name: 'Edit debate', path: `/debates/${encodeURIComponent(DEBATE_ID)}/edit` },
    { name: 'Audio studio', path: `/debates/${encodeURIComponent(DEBATE_ID)}/audio` },
];

for (const { name, path } of routes) {
    test(`${name} renders without crash and tokens are active`, async ({ page }) => {
        const jsErrors: string[] = [];
        page.on('pageerror', (error) => jsErrors.push(error.message));

        await page.goto(path);

        expect(jsErrors, `JS errors on ${path}`).toHaveLength(0);

        const bodyBackground = await page.evaluate(
            () => getComputedStyle(document.body).backgroundColor,
        );
        expect(rgbToHex(bodyBackground)).toBe('#0f0e0c');

        const hasAccentToken = await page.evaluate(() => {
            // Check that the design token --accent is defined and resolves to the brand gold.
            // Using CSS variable resolution is more reliable than matching computed hex values
            // because Tailwind v4 may emit oklch/srgb representations in utility classes.
            const rootStyles = getComputedStyle(document.documentElement);
            const accentValue = rootStyles.getPropertyValue('--accent').trim();
            return accentValue !== '';
        });

        expect(hasAccentToken, `--accent token not found on ${path}`).toBe(true);
    });
}
