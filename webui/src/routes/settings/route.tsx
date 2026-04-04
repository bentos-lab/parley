import { useEffect, useRef, useState } from 'react';
import { useLoaderData, useRevalidator } from 'react-router-dom';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import {
    getLlmProviderPreset,
    inferLlmProviderPresetId,
    INWORLD_MODEL_OPTIONS,
    LLM_PROVIDER_PRESETS,
    type LlmProviderPresetId,
    type ProviderModelOption,
} from '@/lib/providerCatalog';
import { getConfig, updateConfig } from '@/services/api/config';
import { getErrorMessage } from '@/services/api/http';
import { ConfigResponseSchema } from '@/services/api/schemas';
import type { ConfigResponse, ConfigUpdatePayload } from '@/types';

// Icons as inline SVGs
function EyeIcon({ className }: { className?: string }) {
    return (
        <svg
            className={className ?? 'w-4 h-4'}
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='2'
            strokeLinecap='round'
            strokeLinejoin='round'
            aria-hidden='true'
        >
            <path d='M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z' />
            <circle cx='12' cy='12' r='3' />
        </svg>
    );
}

function EyeOffIcon({ className }: { className?: string }) {
    return (
        <svg
            className={className ?? 'w-4 h-4'}
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='2'
            strokeLinecap='round'
            strokeLinejoin='round'
            aria-hidden='true'
        >
            <path d='M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24' />
            <line x1='1' y1='1' x2='23' y2='23' />
        </svg>
    );
}

function CheckIcon({ className }: { className?: string }) {
    return (
        <svg
            className={className ?? 'w-4 h-4'}
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='2'
            strokeLinecap='round'
            strokeLinejoin='round'
            aria-hidden='true'
        >
            <polyline points='20 6 9 17 4 12' />
        </svg>
    );
}

const TTS_PROVIDERS = [
    { value: 'native', label: 'Native' },
    { value: 'inworld', label: 'Inworld' },
];

const CUSTOM_MODEL_VALUE = '__custom_model__';

interface SelectFieldProps {
    id: string;
    value: string;
    onChange: (value: string) => void;
    options: ProviderModelOption[];
    disabled?: boolean;
}

function SelectField({ id, value, onChange, options, disabled = false }: SelectFieldProps) {
    return (
        <div className='relative'>
            <select
                id={id}
                value={value}
                onChange={(event) => onChange(event.target.value)}
                disabled={disabled}
                className='w-full appearance-none rounded border border-border bg-bg-surface px-3 py-1.5 pr-9 text-xs font-sans text-text-1 focus:border-accent-dim focus:outline-none focus:ring-1 focus:ring-accent/30 disabled:cursor-not-allowed disabled:opacity-50'
            >
                {options.map((option) => (
                    <option key={option.value} value={option.value}>
                        {option.deprecated ? `${option.label} (deprecated)` : option.label}
                    </option>
                ))}
                <option value={CUSTOM_MODEL_VALUE}>Custom model</option>
            </select>
            <svg
                className='pointer-events-none absolute right-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-text-3'
                viewBox='0 0 16 16'
                fill='none'
                aria-hidden='true'
            >
                <path
                    d='M4 6l4 4 4-4'
                    stroke='currentColor'
                    strokeWidth='1.4'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                />
            </svg>
        </div>
    );
}

// eslint-disable-next-line react-refresh/only-export-components
export async function loader() {
    const data = await getConfig();
    return ConfigResponseSchema.parse(data);
}

export function Component() {
    const config = useLoaderData() as ConfigResponse;
    const revalidator = useRevalidator();
    const successTimeoutRef = useRef<number | null>(null);
    const initialProviderPreset = inferLlmProviderPresetId(config.llm.openai.base_url);

    // LLM state
    const [llmProviderPresetId, setLlmProviderPresetId] =
        useState<LlmProviderPresetId>(initialProviderPreset);
    const [llmBaseUrl, setLlmBaseUrl] = useState(config.llm.openai.base_url);
    const [llmApiKey, setLlmApiKey] = useState(config.llm.openai.api_key);
    const [llmModel, setLlmModel] = useState(config.llm.openai.model);
    const [showLlmApiKey, setShowLlmApiKey] = useState(false);

    // TTS state
    const [ttsProvider, setTtsProvider] = useState(config.tts.provider);
    const [inworldApiKey, setInworldApiKey] = useState(config.tts.inworld.api_key);
    const [inworldModel, setInworldModel] = useState(config.tts.inworld.model);
    const [showInworldApiKey, setShowInworldApiKey] = useState(false);

    // UI state
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [success, setSuccess] = useState(false);

    const llmPreset = getLlmProviderPreset(llmProviderPresetId);
    const llmUsesPresetBaseUrl = llmProviderPresetId !== 'custom';
    const hasPresetLlmModel = llmPreset.models.some((model) => model.value === llmModel);
    const llmModelSelectionValue = hasPresetLlmModel ? llmModel : CUSTOM_MODEL_VALUE;
    const hasKnownInworldModel = INWORLD_MODEL_OPTIONS.some(
        (model) => model.value === inworldModel,
    );
    const inworldModelSelectionValue = hasKnownInworldModel ? inworldModel : CUSTOM_MODEL_VALUE;
    const selectedInworldModel = INWORLD_MODEL_OPTIONS.find(
        (model) => model.value === inworldModel,
    );

    useEffect(() => {
        return () => {
            if (successTimeoutRef.current != null) {
                window.clearTimeout(successTimeoutRef.current);
            }
        };
    }, []);

    async function handleSave() {
        setSaving(true);
        setError(null);
        setSuccess(false);

        if (successTimeoutRef.current != null) {
            window.clearTimeout(successTimeoutRef.current);
            successTimeoutRef.current = null;
        }

        const payload: ConfigUpdatePayload = {
            llm: {
                openai: {
                    base_url: llmUsesPresetBaseUrl ? llmPreset.baseUrl : llmBaseUrl,
                    api_key: llmApiKey,
                    model: llmModel,
                },
            },
            tts: {
                provider: ttsProvider,
                inworld: {
                    api_key: inworldApiKey,
                    model: inworldModel,
                },
            },
        };

        try {
            await updateConfig(payload);
            revalidator.revalidate();
            setSuccess(true);
            // Auto-hide success message after 3 seconds
            successTimeoutRef.current = window.setTimeout(() => {
                setSuccess(false);
                successTimeoutRef.current = null;
            }, 3000);
        } catch (err) {
            const message = getErrorMessage(err, 'Failed to save settings. Please try again.');
            setError(message);
        } finally {
            setSaving(false);
        }
    }

    function handleProviderPresetChange(nextPresetId: LlmProviderPresetId) {
        setLlmProviderPresetId(nextPresetId);

        if (nextPresetId === 'custom') {
            return;
        }

        const nextPreset = getLlmProviderPreset(nextPresetId);
        setLlmBaseUrl(nextPreset.baseUrl);
        setLlmModel((currentModel) => {
            if (nextPreset.models.some((model) => model.value === currentModel)) {
                return currentModel;
            }

            return nextPreset.defaultModel;
        });
    }

    function handleLlmModelSelection(nextValue: string) {
        if (nextValue === CUSTOM_MODEL_VALUE) {
            setLlmModel((currentModel) =>
                llmPreset.models.some((model) => model.value === currentModel) ? '' : currentModel,
            );
            return;
        }

        setLlmModel(nextValue);
    }

    function handleInworldModelSelection(nextValue: string) {
        if (nextValue === CUSTOM_MODEL_VALUE) {
            setInworldModel((currentModel) =>
                INWORLD_MODEL_OPTIONS.some((model) => model.value === currentModel)
                    ? ''
                    : currentModel,
            );
            return;
        }

        setInworldModel(nextValue);
    }

    return (
        <div className='mx-auto w-full max-w-3xl p-6 sm:p-8 font-sans text-text-1'>
            <h1 className='font-display text-xl text-text-1 mb-5'>Settings</h1>

            {/* Error banner */}
            {error && (
                <div className='mb-4 rounded-lg border border-error/30 bg-error/10 px-4 py-3 text-sm text-error'>
                    {error}
                </div>
            )}

            {/* Success banner */}
            {success && (
                <div className='mb-4 rounded-lg border border-accent/30 bg-accent/10 px-4 py-3 text-sm text-accent flex items-center gap-2'>
                    <CheckIcon className='w-4 h-4' />
                    Settings saved successfully.
                </div>
            )}

            <div className='space-y-6'>
                {/* LLM Settings Section */}
                <section className='rounded-lg border border-border bg-bg-surface p-4'>
                    <h2 className='text-sm font-semibold text-text-1 mb-4 flex items-center gap-2'>
                        <svg
                            className='w-4 h-4 opacity-70'
                            viewBox='0 0 24 24'
                            fill='none'
                            stroke='currentColor'
                            strokeWidth='2'
                            strokeLinecap='round'
                            strokeLinejoin='round'
                            aria-hidden='true'
                        >
                            <path d='M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5' />
                        </svg>
                        LLM Configuration
                    </h2>

                    <div className='space-y-4'>
                        <div>
                            <label className='mb-2 block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'>
                                Provider preset
                            </label>
                            <div className='grid gap-2 sm:grid-cols-2'>
                                {LLM_PROVIDER_PRESETS.map((preset) => (
                                    <label
                                        key={preset.id}
                                        className={`cursor-pointer rounded border px-3 py-2.5 transition-colors ${
                                            llmProviderPresetId === preset.id
                                                ? 'border-accent bg-bg-elevated text-text-1'
                                                : 'border-border bg-bg-base text-text-2 hover:border-border-mid hover:text-text-1'
                                        }`}
                                    >
                                        <input
                                            className='sr-only'
                                            type='radio'
                                            name='llm_provider_preset'
                                            value={preset.id}
                                            checked={llmProviderPresetId === preset.id}
                                            onChange={() => handleProviderPresetChange(preset.id)}
                                        />
                                        <span className='block text-xs font-semibold'>
                                            {preset.label}
                                        </span>
                                        <span className='mt-1 block text-[11px] leading-4 text-text-3'>
                                            {preset.description}
                                        </span>
                                    </label>
                                ))}
                            </div>
                            <p className='mt-1.5 text-[11px] text-text-3'>
                                Parley uses OpenAI-compatible chat endpoints. These presets fill the
                                correct base URL and model list for common providers.
                            </p>
                        </div>

                        {/* Base URL */}
                        <div>
                            <label
                                className='mb-2 block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'
                                htmlFor='llm-base-url'
                            >
                                Base URL
                            </label>
                            <Input
                                id='llm-base-url'
                                type='url'
                                value={llmUsesPresetBaseUrl ? llmPreset.baseUrl : llmBaseUrl}
                                onChange={(e) => setLlmBaseUrl(e.target.value)}
                                placeholder='https://api.openai.com/v1'
                                disabled={llmUsesPresetBaseUrl}
                            />
                            <p className='mt-1.5 text-[11px] text-text-3'>
                                {llmUsesPresetBaseUrl
                                    ? 'Filled automatically from the selected provider preset.'
                                    : 'Manual OpenAI-compatible base URL for custom providers or proxies.'}
                            </p>
                        </div>

                        {/* API Key */}
                        <div>
                            <label
                                className='mb-2 block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'
                                htmlFor='llm-api-key'
                            >
                                API Key
                            </label>
                            <div className='relative'>
                                <Input
                                    id='llm-api-key'
                                    type={showLlmApiKey ? 'text' : 'password'}
                                    value={llmApiKey}
                                    onChange={(e) => setLlmApiKey(e.target.value)}
                                    placeholder='sk-...'
                                    className='pr-10'
                                />
                                <button
                                    type='button'
                                    onClick={() => setShowLlmApiKey(!showLlmApiKey)}
                                    className='absolute right-2 top-1/2 -translate-y-1/2 p-1 text-text-3 hover:text-text-1 transition-colors cursor-pointer'
                                    aria-label={showLlmApiKey ? 'Hide API key' : 'Show API key'}
                                >
                                    {showLlmApiKey ? (
                                        <EyeOffIcon className='w-3.5 h-3.5' />
                                    ) : (
                                        <EyeIcon className='w-3.5 h-3.5' />
                                    )}
                                </button>
                            </div>
                            <p className='mt-1.5 text-[11px] text-text-3'>
                                Your OpenAI API key. Stored locally on your machine.
                            </p>
                        </div>

                        {/* Model */}
                        <div>
                            <label
                                className='mb-2 block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'
                                htmlFor='llm-model'
                            >
                                Model
                            </label>
                            {llmPreset.models.length > 0 ? (
                                <SelectField
                                    id='llm-model'
                                    value={llmModelSelectionValue}
                                    onChange={handleLlmModelSelection}
                                    options={llmPreset.models}
                                />
                            ) : (
                                <Input
                                    id='llm-model'
                                    type='text'
                                    value={llmModel}
                                    onChange={(e) => setLlmModel(e.target.value)}
                                    placeholder='provider/model-or-alias'
                                />
                            )}
                            {llmPreset.models.length > 0 &&
                            llmModelSelectionValue === CUSTOM_MODEL_VALUE ? (
                                <Input
                                    className='mt-2'
                                    type='text'
                                    value={llmModel}
                                    onChange={(e) => setLlmModel(e.target.value)}
                                    placeholder='Enter a custom model identifier'
                                />
                            ) : null}
                            <p className='mt-1.5 text-[11px] text-text-3'>
                                Select a published chat/completions model or enter a custom
                                identifier.
                            </p>
                        </div>
                    </div>
                </section>

                {/* TTS Settings Section */}
                <section className='rounded-lg border border-border bg-bg-surface p-4'>
                    <h2 className='text-sm font-semibold text-text-1 mb-4 flex items-center gap-2'>
                        <svg
                            className='w-4 h-4 opacity-70'
                            viewBox='0 0 24 24'
                            fill='none'
                            stroke='currentColor'
                            strokeWidth='2'
                            strokeLinecap='round'
                            strokeLinejoin='round'
                            aria-hidden='true'
                        >
                            <path d='M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z' />
                            <path d='M19 10v2a7 7 0 0 1-14 0v-2' />
                            <line x1='12' y1='19' x2='12' y2='23' />
                            <line x1='8' y1='23' x2='16' y2='23' />
                        </svg>
                        TTS Configuration
                    </h2>

                    <div className='space-y-4'>
                        {/* Provider */}
                        <div>
                            <label className='mb-2 block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'>
                                Provider
                            </label>
                            <div className='grid gap-2 grid-cols-2'>
                                {TTS_PROVIDERS.map((p) => (
                                    <label
                                        key={p.value}
                                        className={`cursor-pointer rounded border px-3 py-2 text-xs transition-colors ${
                                            ttsProvider === p.value
                                                ? 'border-accent bg-bg-elevated text-text-1'
                                                : 'border-border bg-bg-base text-text-2 hover:border-border-mid hover:text-text-1'
                                        }`}
                                    >
                                        <input
                                            className='sr-only'
                                            type='radio'
                                            name='tts_provider'
                                            value={p.value}
                                            checked={ttsProvider === p.value}
                                            onChange={() => setTtsProvider(p.value)}
                                        />
                                        <span className='block text-xs font-semibold'>
                                            {p.label}
                                        </span>
                                    </label>
                                ))}
                            </div>
                            <p className='mt-1.5 text-[11px] text-text-3'>
                                Select the text-to-speech provider.
                            </p>
                        </div>

                        {/* Inworld settings (only show when inworld is selected) */}
                        {ttsProvider === 'inworld' && (
                            <>
                                {/* Inworld API Key */}
                                <div>
                                    <label
                                        className='mb-2 block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'
                                        htmlFor='inworld-api-key'
                                    >
                                        Inworld API Key
                                    </label>
                                    <div className='relative'>
                                        <Input
                                            id='inworld-api-key'
                                            type={showInworldApiKey ? 'text' : 'password'}
                                            value={inworldApiKey}
                                            onChange={(e) => setInworldApiKey(e.target.value)}
                                            placeholder='Enter your Inworld API key'
                                            className='pr-10'
                                        />
                                        <button
                                            type='button'
                                            onClick={() => setShowInworldApiKey(!showInworldApiKey)}
                                            className='absolute right-2 top-1/2 -translate-y-1/2 p-1 text-text-3 hover:text-text-1 transition-colors cursor-pointer'
                                            aria-label={
                                                showInworldApiKey ? 'Hide API key' : 'Show API key'
                                            }
                                        >
                                            {showInworldApiKey ? (
                                                <EyeOffIcon className='w-3.5 h-3.5' />
                                            ) : (
                                                <EyeIcon className='w-3.5 h-3.5' />
                                            )}
                                        </button>
                                    </div>
                                    <p className='mt-1.5 text-[11px] text-text-3'>
                                        Your Inworld TTS API key.
                                    </p>
                                </div>

                                {/* Inworld Model */}
                                <div>
                                    <label
                                        className='mb-2 block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'
                                        htmlFor='inworld-model'
                                    >
                                        Inworld Model
                                    </label>
                                    <SelectField
                                        id='inworld-model'
                                        value={inworldModelSelectionValue}
                                        onChange={handleInworldModelSelection}
                                        options={INWORLD_MODEL_OPTIONS}
                                    />
                                    {inworldModelSelectionValue === CUSTOM_MODEL_VALUE ? (
                                        <Input
                                            className='mt-2'
                                            type='text'
                                            value={inworldModel}
                                            onChange={(e) => setInworldModel(e.target.value)}
                                            placeholder='Enter a custom Inworld model identifier'
                                        />
                                    ) : null}
                                    <p className='mt-1.5 text-[11px] text-text-3'>
                                        Pick from current Inworld TTS releases or enter a custom
                                        model identifier.
                                    </p>
                                    {selectedInworldModel?.deprecated ? (
                                        <div className='mt-2 rounded-lg border border-stale-border/50 bg-stale-bg px-3 py-2 text-[11px] text-stale'>
                                            This Inworld model is deprecated. Prefer the 1.5 Max or
                                            1.5 Mini family for new setups.
                                        </div>
                                    ) : null}
                                </div>
                            </>
                        )}
                    </div>
                </section>

                {/* Save Button */}
                <div className='flex justify-end'>
                    <Button variant='accent' onClick={handleSave} disabled={saving}>
                        {saving ? (
                            <>
                                <span className='h-3 w-3 animate-spin rounded-full border border-bg-base/40 border-t-bg-base' />
                                Saving...
                            </>
                        ) : (
                            'Save settings'
                        )}
                    </Button>
                </div>
            </div>
        </div>
    );
}

export function ErrorBoundary() {
    return (
        <div className='mx-auto w-full max-w-3xl p-6 sm:p-8 font-sans'>
            <h1 className='font-display text-xl text-text-1 mb-4'>Settings</h1>
            <div className='rounded-lg border border-error/30 bg-error/10 px-4 py-3 text-sm text-error'>
                Failed to load settings. Please check your connection and try again.
            </div>
        </div>
    );
}
