export interface ProviderModelOption {
    value: string;
    label: string;
    description: string;
    deprecated?: boolean;
}

export type LlmProviderPresetId = 'openai' | 'gemini' | 'anthropic' | 'custom';

export interface LlmProviderPreset {
    id: LlmProviderPresetId;
    label: string;
    description: string;
    baseUrl: string;
    defaultModel: string;
    models: ProviderModelOption[];
}

export const LLM_PROVIDER_PRESETS: LlmProviderPreset[] = [
    {
        id: 'openai',
        label: 'OpenAI',
        description: 'Direct OpenAI Responses and chat-completions compatible endpoint.',
        baseUrl: 'https://api.openai.com/v1',
        defaultModel: 'gpt-4.1-mini',
        models: [
            {
                value: 'gpt-5.4',
                label: 'GPT-5.4',
                description: 'Flagship reasoning and coding model.',
            },
            {
                value: 'gpt-5.4-mini',
                label: 'GPT-5.4 mini',
                description: 'Smaller GPT-5.4 variant for lower latency.',
            },
            {
                value: 'gpt-5.4-nano',
                label: 'GPT-5.4 nano',
                description: 'Lowest-cost GPT-5.4 tier for high-volume work.',
            },
            {
                value: 'o3',
                label: 'o3',
                description: 'Heavy reasoning model for difficult tool-using workflows.',
            },
            {
                value: 'o4-mini',
                label: 'o4-mini',
                description: 'Faster reasoning-focused OpenAI model.',
            },
            {
                value: 'gpt-4.1',
                label: 'GPT-4.1',
                description: 'Strong instruction-following and tool-calling model.',
            },
            {
                value: 'gpt-4.1-mini',
                label: 'GPT-4.1 mini',
                description: 'Balanced default for most text generation tasks.',
            },
            {
                value: 'gpt-4.1-nano',
                label: 'GPT-4.1 nano',
                description: 'Low-cost GPT-4.1 family option.',
            },
            {
                value: 'gpt-4o',
                label: 'GPT-4o',
                description: 'General-purpose multimodal model.',
            },
            {
                value: 'gpt-4o-mini',
                label: 'GPT-4o mini',
                description: 'Cheap and fast GPT-4o family option.',
            },
        ],
    },
    {
        id: 'gemini',
        label: 'Gemini',
        description: 'Google Gemini via the OpenAI-compatible compatibility layer.',
        baseUrl: 'https://generativelanguage.googleapis.com/v1beta/openai',
        defaultModel: 'gemini-2.5-flash',
        models: [
            {
                value: 'gemini-3.1-pro-preview',
                label: 'Gemini 3.1 Pro Preview',
                description: "Google's most capable reasoning and coding preview.",
            },
            {
                value: 'gemini-3-flash-preview',
                label: 'Gemini 3 Flash Preview',
                description: 'Fast frontier-class multimodal chat model.',
            },
            {
                value: 'gemini-3.1-flash-lite-preview',
                label: 'Gemini 3.1 Flash-Lite Preview',
                description: 'Cheaper Gemini 3 family option with strong speed.',
            },
            {
                value: 'gemini-2.5-pro',
                label: 'Gemini 2.5 Pro',
                description: 'Deep reasoning model suited to complex debate prompts.',
            },
            {
                value: 'gemini-2.5-flash',
                label: 'Gemini 2.5 Flash',
                description: 'Best price-performance Gemini model for chat tasks.',
            },
            {
                value: 'gemini-2.5-flash-lite',
                label: 'Gemini 2.5 Flash-Lite',
                description: 'Fastest low-cost 2.5 family option.',
            },
            {
                value: 'gemini-flash-latest',
                label: 'Gemini Flash Latest',
                description: 'Alias that tracks the newest Flash release.',
            },
        ],
    },
    {
        id: 'anthropic',
        label: 'Claude',
        description: 'Anthropic Claude through the OpenAI SDK compatibility endpoint.',
        baseUrl: 'https://api.anthropic.com/v1',
        defaultModel: 'claude-sonnet-4-6',
        models: [
            {
                value: 'claude-opus-4-6',
                label: 'Claude Opus 4.6',
                description: 'Highest-capability Claude model for hard reasoning.',
            },
            {
                value: 'claude-sonnet-4-6',
                label: 'Claude Sonnet 4.6',
                description: 'Best balance of speed and intelligence in Claude.',
            },
            {
                value: 'claude-haiku-4-5',
                label: 'Claude Haiku 4.5',
                description: 'Fastest Claude family option exposed via alias.',
            },
        ],
    },
    {
        id: 'custom',
        label: 'Custom endpoint',
        description: 'Manual OpenAI-compatible endpoint for future providers or proxies.',
        baseUrl: '',
        defaultModel: '',
        models: [],
    },
];

export const INWORLD_MODEL_OPTIONS: ProviderModelOption[] = [
    {
        value: 'inworld-tts-1.5-max',
        label: 'Inworld TTS 1.5 Max',
        description: 'Flagship quality-speed balance with enhanced timestamps.',
    },
    {
        value: 'inworld-tts-1.5-mini',
        label: 'Inworld TTS 1.5 Mini',
        description: 'Lowest-latency and lowest-cost Inworld TTS model.',
    },
    {
        value: 'inworld-tts-1-max',
        label: 'Inworld TTS 1 Max',
        description: 'Previous generation high-quality model.',
        deprecated: true,
    },
    {
        value: 'inworld-tts-1',
        label: 'Inworld TTS 1',
        description: 'Previous generation fast model.',
        deprecated: true,
    },
];

function trimTrailingSlash(value: string): string {
    return value.replace(/\/+$/, '');
}

export function normalizeBaseUrl(value: string): string {
    return trimTrailingSlash(value.trim().toLowerCase());
}

export function inferLlmProviderPresetId(baseUrl: string): LlmProviderPresetId {
    const normalized = normalizeBaseUrl(baseUrl);

    if (normalized.includes('generativelanguage.googleapis.com')) {
        return 'gemini';
    }

    if (normalized.includes('api.anthropic.com')) {
        return 'anthropic';
    }

    if (normalized.includes('api.openai.com')) {
        return 'openai';
    }

    return 'custom';
}

export function getLlmProviderPreset(id: LlmProviderPresetId): LlmProviderPreset {
    return LLM_PROVIDER_PRESETS.find((preset) => preset.id === id) ?? LLM_PROVIDER_PRESETS[0];
}
