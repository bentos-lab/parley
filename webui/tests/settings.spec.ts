import { expect, test } from '@playwright/test';

test('provider presets fill compatible base URLs and save mapped models', async ({ page }) => {
    let capturedPayload: Record<string, unknown> | null = null;

    await page.route('**/api/debates', async (route) => {
        if (route.request().method() !== 'GET') {
            await route.continue();
            return;
        }

        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify([]),
        });
    });

    await page.route('**/api/config', async (route) => {
        if (route.request().method() === 'GET') {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    llm: {
                        provider: 'openai',
                        openai: {
                            base_url: 'https://api.openai.com/v1',
                            api_key: 'sk-test',
                            model: 'gpt-4.1-mini',
                        },
                    },
                    tts: {
                        provider: 'native',
                        inworld: {
                            api_key: '',
                            model: 'inworld-tts-1.5-max',
                        },
                    },
                }),
            });
            return;
        }

        if (route.request().method() === 'PUT') {
            capturedPayload = route.request().postDataJSON() as Record<string, unknown>;

            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    llm: {
                        provider: 'openai',
                        openai: {
                            base_url: 'https://api.anthropic.com/v1',
                            api_key: 'sk-test',
                            model: 'claude-sonnet-4-6',
                        },
                    },
                    tts: {
                        provider: 'inworld',
                        inworld: {
                            api_key: 'inworld-test',
                            model: 'inworld-tts-1.5-mini',
                        },
                    },
                }),
            });
            return;
        }

        await route.continue();
    });

    await page.goto('/#/settings');

    await page.getByText('Claude', { exact: true }).click();
    await expect(page.getByLabel(/base url/i)).toHaveValue('https://api.anthropic.com/v1');

    await page.getByLabel(/^model$/i).selectOption('claude-sonnet-4-6');

    await page.getByText('Inworld', { exact: true }).click();
    await page.getByLabel(/inworld api key/i).fill('inworld-test');
    await page.getByLabel(/inworld model/i).selectOption('inworld-tts-1.5-mini');

    await page.getByRole('button', { name: /save settings/i }).click();

    await expect(page.getByText('Settings saved successfully.')).toBeVisible();

    expect(capturedPayload).toEqual({
        llm: {
            openai: {
                base_url: 'https://api.anthropic.com/v1',
                api_key: 'sk-test',
                model: 'claude-sonnet-4-6',
            },
        },
        tts: {
            provider: 'inworld',
            inworld: {
                api_key: 'inworld-test',
                model: 'inworld-tts-1.5-mini',
            },
        },
    });
});
