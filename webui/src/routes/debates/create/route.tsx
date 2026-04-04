import { useEffect, useState, type FormEvent } from 'react';
import {
    Form,
    redirect,
    useActionData,
    useBlocker,
    useNavigation,
    useRouteError,
} from 'react-router-dom';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { useToast } from '@/components/ui/ToastProvider';
import { writeStoredSessionConfig, DEFAULT_TURN_COUNT } from '@/lib/debateSessionConfig';
import { ApiError, getErrorMessage } from '@/services/api/http';
import { createDebate, type CreateDebatePayload } from '@/services/api/debates';
import { getConfig } from '@/services/api/config';

type AgentFormValue = {
    name: string;
    stance: string;
};

type ActionData = {
    error?: string;
};

const TTS_PROVIDERS = [
    { value: 'native', label: 'Native (local)' },
    { value: 'inworld', label: 'Inworld (cloud)' },
];

const EMPTY_AGENT: AgentFormValue = { name: '', stance: '' };

function getAgentFieldKey(index: number, field: keyof AgentFormValue) {
    return `agents_${index}_${field}`;
}

function validateFields(input: { topic: string; name: string; agents: AgentFormValue[] }) {
    const errors: Record<string, string> = {};
    const topic = input.topic.trim();
    const name = input.name.trim();

    if (!topic) {
        errors.topic = 'Topic is required';
    } else if (topic.length < 3) {
        errors.topic = 'Topic must be at least 3 characters';
    } else if (topic.length > 200) {
        errors.topic = 'Topic must be less than 200 characters';
    }

    if (name.length > 100) {
        errors.name = 'Name must be less than 100 characters';
    }

    input.agents.forEach((agent, index) => {
        if (agent.name.trim().length > 100) {
            errors[getAgentFieldKey(index, 'name')] = 'Name must be less than 100 characters';
        }
        if (agent.stance.trim().length > 120) {
            errors[getAgentFieldKey(index, 'stance')] = 'Stance must be less than 120 characters';
        }
    });

    return { valid: Object.keys(errors).length === 0, errors };
}

function parseAgentsFromFormData(formData: FormData) {
    const agentsByIndex = new Map<number, AgentFormValue>();

    for (const [key, rawValue] of formData.entries()) {
        if (typeof rawValue !== 'string') continue;
        const match = key.match(/^agent-(\d+)-(name|stance)$/);
        if (!match) continue;
        const index = Number.parseInt(match[1], 10);
        const field = match[2] as keyof AgentFormValue;
        const current = agentsByIndex.get(index) ?? { ...EMPTY_AGENT };
        current[field] = rawValue;
        agentsByIndex.set(index, current);
    }

    const sorted = Array.from(agentsByIndex.entries())
        .sort(([a], [b]) => a - b)
        .map(([, v]) => v);

    return sorted.length < 2 ? [{ ...EMPTY_AGENT }, { ...EMPTY_AGENT }] : sorted.slice(0, 4);
}

// eslint-disable-next-line react-refresh/only-export-components
export async function action({ request }: { request: Request }) {
    const formData = await request.formData();
    const topic = (formData.get('topic')?.toString() ?? '').trim();
    const name = (formData.get('name')?.toString() ?? '').trim();
    const ttsProvider = (formData.get('tts_provider')?.toString() ?? 'native').trim();
    const turnCount = Number.parseInt(
        formData.get('turn_count')?.toString() ?? String(DEFAULT_TURN_COUNT),
        10,
    );
    const playAudio = formData.get('play_audio')?.toString() === '1';
    const agents = parseAgentsFromFormData(formData);

    const validation = validateFields({ topic, name, agents });
    if (!validation.valid) {
        // Return field-level errors as a banner message
        const firstError = Object.values(validation.errors)[0];
        return { error: firstError };
    }

    const payload: CreateDebatePayload = {
        topic,
        name: name || undefined,
        tts_provider: ttsProvider,
    };

    const hasAgentData = agents.some((a) => Boolean(a.name.trim() || a.stance.trim()));
    if (hasAgentData) {
        payload.agents = agents
            .filter((a) => Boolean(a.name.trim() || a.stance.trim()))
            .map((a) => ({
                name: a.name.trim() || undefined,
                stance: a.stance.trim() || undefined,
            }));
    } else {
        // When no agents are specified, request server to generate agents based on form count
        payload.num_agents = agents.length;
    }

    try {
        const created = await createDebate(payload);
        writeStoredSessionConfig(created.id, { turnCount });
        const redirectUrl = `/debates/${encodeURIComponent(created.id)}?round=${turnCount}${playAudio ? '&autoplay=1' : ''}`;
        return redirect(redirectUrl);
    } catch (error) {
        if (error instanceof ApiError && error.status === 400) {
            return { error: getErrorMessage(error, 'Something went wrong') };
        }
        throw error;
    }
}

export function Component() {
    const actionData = useActionData() as ActionData | undefined;
    const navigation = useNavigation();
    const toast = useToast();

    const [topic, setTopic] = useState('');
    const [name, setName] = useState('');
    const [turnCount, setTurnCount] = useState(DEFAULT_TURN_COUNT);
    const [ttsProvider, setTtsProvider] = useState('');
    const [playAudioImmediately, setPlayAudioImmediately] = useState(false);
    const [configLoaded, setConfigLoaded] = useState(false);
    const [agents, setAgents] = useState<AgentFormValue[]>([
        { ...EMPTY_AGENT },
        { ...EMPTY_AGENT },
    ]);
    const [isAdvancedOpen, setIsAdvancedOpen] = useState(false);
    const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
    const [touched, setTouched] = useState<Record<string, boolean>>({});
    const [isPending, setIsPending] = useState(false);
    const [isDirty, setIsDirty] = useState(false);
    const [allowNavigation, setAllowNavigation] = useState(false);

    const blocker = useBlocker(
        ({ currentLocation, nextLocation }) =>
            isDirty &&
            !allowNavigation &&
            navigation.state === 'idle' &&
            currentLocation.pathname !== nextLocation.pathname,
    );

    useEffect(() => {
        if (actionData?.error) {
            // eslint-disable-next-line react-hooks/set-state-in-effect
            setIsPending(false);
            setAllowNavigation(false);
        }
    }, [actionData]);

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        setIsPending(navigation.state === 'submitting');
        if (navigation.state === 'loading' && !actionData?.error) setIsDirty(false);
        if (navigation.state === 'idle' && !actionData?.error) setAllowNavigation(false);
    }, [actionData?.error, navigation.state]);

    useEffect(() => {
        let cancelled = false;
        getConfig()
            .then((cfg) => {
                if (cancelled) return;
                // Use config TTS provider if set, otherwise default to native
                const provider = cfg.tts?.provider?.trim() || 'native';

                setTtsProvider(provider);

                setConfigLoaded(true);
            })
            .catch(() => {
                if (cancelled) return;
                // On error, default to native

                setTtsProvider('native');

                setConfigLoaded(true);
                toast.error(
                    'Could not load runtime config. Using Native as the fallback TTS provider.',
                    {
                        title: 'Using fallback config',
                    },
                );
            });
        return () => {
            cancelled = true;
        };
    }, [toast]);

    function markDirty() {
        if (!isDirty) setIsDirty(true);
    }

    function updateAgents(updater: (c: AgentFormValue[]) => AgentFormValue[]) {
        setAgents((c) => updater(c));
        markDirty();
    }

    function handleBlur(field: string) {
        const validation = validateFields({ topic, name, agents });
        setFieldErrors(validation.errors);
        setTouched((p) => ({ ...p, [field]: true }));
    }

    function handleSubmit(event: FormEvent<HTMLFormElement>) {
        const allTouched: Record<string, boolean> = { topic: true, name: true };
        agents.forEach((_, i) => {
            allTouched[getAgentFieldKey(i, 'name')] = true;
            allTouched[getAgentFieldKey(i, 'stance')] = true;
        });

        const validation = validateFields({ topic, name, agents });
        setFieldErrors(validation.errors);
        setTouched(allTouched);

        if (!validation.valid) {
            event.preventDefault();
            setIsPending(false);
            return;
        }

        setIsPending(true);
        setAllowNavigation(true);
    }

    function addAgent() {
        updateAgents((c) => (c.length >= 4 ? c : [...c, { ...EMPTY_AGENT }]));
    }

    function removeAgent(index: number) {
        updateAgents((c) => (c.length <= 2 ? c : c.filter((_, i) => i !== index)));
    }

    function fieldError(field: string) {
        return touched[field] ? fieldErrors[field] : undefined;
    }

    return (
        <>
            {navigation.state === 'submitting' && (
                <div className='fixed inset-0 z-50 flex flex-col items-center justify-center bg-(--bg-base)/95'>
                    <span
                        className='text-lg text-(--text-1)'
                        style={{ fontFamily: 'var(--font-display)' }}
                    >
                        Creating debate...
                    </span>
                    <div className='mt-4 flex gap-1.5'>
                        <span className='h-1.5 w-1.5 animate-[bounce_1s_ease-in-out_0ms_infinite] rounded-full bg-accent' />
                        <span className='h-1.5 w-1.5 animate-[bounce_1s_ease-in-out_150ms_infinite] rounded-full bg-accent' />
                        <span className='h-1.5 w-1.5 animate-[bounce_1s_ease-in-out_300ms_infinite] rounded-full bg-accent' />
                    </div>
                </div>
            )}
            <div className='mx-auto w-full max-w-5xl p-6 sm:p-8 font-sans text-text-1'>
                <h1 className='font-display text-xl text-text-1 mb-5'>New debate</h1>

                {actionData?.error ? (
                    <div className='mb-4 rounded-lg border border-error/30 bg-error/10 px-4 py-3 text-sm text-error'>
                        {actionData.error}
                    </div>
                ) : null}

                <div className='max-w-xl'>
                    <Form
                        method='post'
                        className='rounded-lg border border-border bg-bg-surface p-4'
                        onSubmit={handleSubmit}
                    >
                        <div className='space-y-6'>
                            <input type='hidden' name='name' value={name} />

                            {/* Topic */}
                            <div>
                                <label
                                    className='mb-2 block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'
                                    htmlFor='topic'
                                >
                                    Topic *
                                </label>
                                <textarea
                                    id='topic'
                                    name='topic'
                                    required
                                    rows={3}
                                    value={topic}
                                    placeholder="Enter the proposition to debate — e.g. 'Remote work is more productive than office work'"
                                    onChange={(e) => {
                                        setTopic(e.target.value);
                                        markDirty();
                                    }}
                                    onBlur={() => handleBlur('topic')}
                                    className='w-full resize-none rounded-lg border border-border-mid bg-bg-surface px-3.5 py-2.5 text-[13px] text-text-1 outline-none transition-colors placeholder:text-text-3 focus:border-accent-dim leading-[1.6]'
                                />
                                <p className='mt-1.5 text-[11px] text-text-3'>
                                    State it as a clear, arguable proposition.
                                </p>
                                {fieldError('topic') ? (
                                    <p className='mt-1 text-xs text-error'>{fieldError('topic')}</p>
                                ) : null}
                            </div>

                            {/* TTS Provider */}
                            <div>
                                <label
                                    className='mb-2 block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'
                                    htmlFor='tts_provider'
                                >
                                    TTS provider
                                </label>
                                {!configLoaded ? (
                                    <div className='h-10 flex items-center text-xs text-text-3'>
                                        Loading…
                                    </div>
                                ) : (
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
                                                    onChange={() => {
                                                        setTtsProvider(p.value);
                                                        markDirty();
                                                    }}
                                                />
                                                <span className='block text-xs font-semibold'>
                                                    {p.label}
                                                </span>
                                            </label>
                                        ))}
                                    </div>
                                )}
                            </div>

                            {/* Initial rounds */}
                            <div>
                                <label
                                    className='mb-2 block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'
                                    htmlFor='turn_count'
                                >
                                    Initial rounds
                                </label>
                                <input type='hidden' name='turn_count' value={turnCount} />
                                <input
                                    id='turn_count'
                                    type='number'
                                    min={1}
                                    max={20}
                                    value={turnCount}
                                    onChange={(e) => {
                                        const v = Number.parseInt(e.target.value, 10);
                                        if (!Number.isNaN(v) && v >= 1 && v <= 20) {
                                            setTurnCount(v);
                                            markDirty();
                                        }
                                    }}
                                    className='w-24 rounded-lg border border-border-mid bg-bg-surface px-3 py-2 text-center font-mono text-[13px] text-text-1 outline-none transition-colors focus:border-accent-dim'
                                />
                                <p className='mt-1.5 text-[11px] text-text-3'>
                                    Number of debate rounds to generate (1–20).
                                </p>
                            </div>

                            {/* Play audio immediately toggle */}
                            <div>
                                <input
                                    type='hidden'
                                    name='play_audio'
                                    value={playAudioImmediately ? '1' : '0'}
                                />
                                <button
                                    type='button'
                                    onClick={() => setPlayAudioImmediately(!playAudioImmediately)}
                                    className={`flex w-full items-center justify-between rounded-md border px-3 py-2 text-sm transition-colors ${
                                        playAudioImmediately
                                            ? 'border-accent bg-accent/10 text-accent'
                                            : 'border-border bg-bg-surface text-text-2 hover:border-border-mid'
                                    }`}
                                >
                                    <span className='flex items-center gap-2'>
                                        <svg
                                            className='h-4 w-4'
                                            fill='none'
                                            viewBox='0 0 24 24'
                                            stroke='currentColor'
                                            strokeWidth={2}
                                        >
                                            {playAudioImmediately ? (
                                                <path
                                                    strokeLinecap='round'
                                                    strokeLinejoin='round'
                                                    d='M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z'
                                                />
                                            ) : (
                                                <>
                                                    <path
                                                        strokeLinecap='round'
                                                        strokeLinejoin='round'
                                                        d='M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z'
                                                    />
                                                    <path
                                                        strokeLinecap='round'
                                                        strokeLinejoin='round'
                                                        d='M17 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2'
                                                    />
                                                </>
                                            )}
                                        </svg>
                                        <span>Play audio immediately</span>
                                    </span>
                                    <span
                                        className={`h-4 w-7 rounded-full transition-colors ${
                                            playAudioImmediately ? 'bg-accent' : 'bg-border'
                                        }`}
                                    >
                                        <span
                                            className={`block h-3 w-3 translate-y-0.5 rounded-full bg-white transition-transform ${
                                                playAudioImmediately
                                                    ? 'translate-x-3.5'
                                                    : 'translate-x-0.5'
                                            }`}
                                        />
                                    </span>
                                </button>
                                <p className='mt-1.5 text-[11px] text-text-3'>
                                    Automatically play each round&apos;s audio as it completes.
                                </p>
                            </div>

                            {/* Agent count */}
                            <div className='space-y-2'>
                                <label className='block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'>
                                    Number of agents
                                </label>
                                <div className='flex items-center gap-1.5'>
                                    <button
                                        type='button'
                                        className='flex h-7 w-7 items-center justify-center rounded-[5px] border border-border-mid bg-bg-surface text-text-1 text-[16px] cursor-pointer transition-colors hover:bg-bg-hover'
                                        onClick={() =>
                                            updateAgents((c) =>
                                                c.length <= 2 ? c : c.slice(0, -1),
                                            )
                                        }
                                    >
                                        −
                                    </button>
                                    <span className='w-8 text-center font-mono text-[16px] text-accent'>
                                        {agents.length}
                                    </span>
                                    <button
                                        type='button'
                                        className='flex h-7 w-7 items-center justify-center rounded-[5px] border border-border-mid bg-bg-surface text-text-1 text-[16px] cursor-pointer transition-colors hover:bg-bg-hover'
                                        onClick={addAgent}
                                    >
                                        +
                                    </button>
                                    <span className='ml-1.5 text-[11px] text-text-3'>
                                        Auto-generated from topic if not defined.
                                    </span>
                                </div>
                            </div>

                            {/* Advanced options (collapsible) */}
                            <div className='border-t border-border pt-4'>
                                <button
                                    type='button'
                                    className='flex w-full items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2 cursor-pointer'
                                    onClick={() => setIsAdvancedOpen((o) => !o)}
                                    aria-expanded={isAdvancedOpen}
                                >
                                    <svg
                                        className={`h-3.5 w-3.5 transition-transform duration-200 ${isAdvancedOpen ? 'rotate-90' : ''}`}
                                        viewBox='0 0 16 16'
                                        fill='none'
                                        stroke='currentColor'
                                        strokeWidth='2'
                                        strokeLinecap='round'
                                        strokeLinejoin='round'
                                    >
                                        <path d='M6 4l4 4-4 4' />
                                    </svg>
                                    Advanced options
                                </button>

                                <div
                                    className='overflow-hidden transition-[max-height] duration-300 ease-in-out'
                                    style={{ maxHeight: isAdvancedOpen ? '2000px' : '0px' }}
                                >
                                    <div className='space-y-6 pt-4'>
                                        {/* Name */}
                                        <div>
                                            <label
                                                className='mb-2 block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'
                                                htmlFor='name'
                                            >
                                                Debate name (optional)
                                            </label>
                                            <Input
                                                id='name'
                                                value={name}
                                                placeholder='Auto-generated from topic if left empty'
                                                onChange={(e) => {
                                                    setName(e.target.value);
                                                    markDirty();
                                                }}
                                                onBlur={() => handleBlur('name')}
                                            />
                                            {fieldError('name') ? (
                                                <p className='mt-1 text-xs text-error'>
                                                    {fieldError('name')}
                                                </p>
                                            ) : null}
                                        </div>

                                        {/* Agent definitions */}
                                        <section className='space-y-2'>
                                            <label className='block text-[11px] font-semibold uppercase tracking-[0.04em] text-text-2'>
                                                Agent definitions (optional)
                                            </label>
                                            {agents.map((agent, index) => (
                                                <div
                                                    key={`agent-${index}`}
                                                    className='space-y-3 rounded border border-border p-3'
                                                >
                                                    <div className='flex items-center justify-between'>
                                                        <p className='text-xs text-text-2'>
                                                            Agent {index + 1}
                                                        </p>
                                                        <Button
                                                            type='button'
                                                            variant='ghost'
                                                            disabled={agents.length <= 2}
                                                            onClick={() => removeAgent(index)}
                                                        >
                                                            Remove
                                                        </Button>
                                                    </div>
                                                    <div className='space-y-3'>
                                                        {(['name', 'stance'] as const).map(
                                                            (field) => (
                                                                <div key={field}>
                                                                    <label
                                                                        className='mb-1 block text-xs text-text-2'
                                                                        htmlFor={`agent-${index}-${field}`}
                                                                    >
                                                                        {field
                                                                            .charAt(0)
                                                                            .toUpperCase() +
                                                                            field.slice(1)}
                                                                    </label>
                                                                    <Input
                                                                        id={`agent-${index}-${field}`}
                                                                        name={`agent-${index}-${field}`}
                                                                        aria-label={`Agent ${index + 1} ${field}`}
                                                                        placeholder={
                                                                            field === 'name'
                                                                                ? 'E.g., Alice'
                                                                                : 'E.g., Pro-innovation'
                                                                        }
                                                                        value={agent[field]}
                                                                        onChange={(e) => {
                                                                            updateAgents((c) => {
                                                                                const next = [...c];
                                                                                next[index] = {
                                                                                    ...next[index],
                                                                                    [field]:
                                                                                        e.target
                                                                                            .value,
                                                                                };
                                                                                return next;
                                                                            });
                                                                        }}
                                                                        onBlur={() =>
                                                                            handleBlur(
                                                                                getAgentFieldKey(
                                                                                    index,
                                                                                    field,
                                                                                ),
                                                                            )
                                                                        }
                                                                    />
                                                                    {fieldError(
                                                                        getAgentFieldKey(
                                                                            index,
                                                                            field,
                                                                        ),
                                                                    ) ? (
                                                                        <p className='mt-1 text-xs text-error'>
                                                                            {fieldError(
                                                                                getAgentFieldKey(
                                                                                    index,
                                                                                    field,
                                                                                ),
                                                                            )}
                                                                        </p>
                                                                    ) : null}
                                                                </div>
                                                            ),
                                                        )}
                                                    </div>
                                                </div>
                                            ))}
                                            <Button
                                                type='button'
                                                variant='ghost'
                                                disabled={agents.length >= 4}
                                                onClick={addAgent}
                                                className='text-[11px]'
                                            >
                                                + Add agent manually
                                            </Button>
                                        </section>
                                    </div>
                                </div>
                            </div>

                            <div className='pt-2'>
                                <button
                                    type='submit'
                                    disabled={isPending || !topic.trim()}
                                    className='w-full h-9 rounded-[5px] border border-accent bg-accent text-bg-base font-semibold text-[13px] cursor-pointer transition-colors hover:bg-[#d4ae5a] disabled:opacity-50 disabled:cursor-not-allowed'
                                >
                                    {isPending ? 'Creating…' : 'Create debate'}
                                </button>
                            </div>
                        </div>
                    </Form>
                </div>

                {blocker.state === 'blocked' ? (
                    <div
                        className='fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4'
                        role='presentation'
                        onClick={(e) => {
                            if (e.target === e.currentTarget) blocker.reset();
                        }}
                        onKeyDown={(e) => {
                            if (e.key === 'Escape') blocker.reset();
                        }}
                    >
                        <div className='w-full max-w-md rounded-lg border border-border bg-bg-surface p-4 text-text-1'>
                            <h2 className='font-display text-lg'>Unsaved Changes</h2>
                            <p className='mt-2 text-sm text-text-2'>
                                You have unsaved form changes. Are you sure you want to leave?
                            </p>
                            <div className='mt-4 flex justify-end gap-2'>
                                <Button
                                    type='button'
                                    variant='secondary'
                                    onClick={() => blocker.reset()}
                                >
                                    Cancel
                                </Button>
                                <Button
                                    type='button'
                                    variant='danger'
                                    onClick={() => {
                                        setIsDirty(false);
                                        blocker.proceed();
                                    }}
                                >
                                    Leave
                                </Button>
                            </div>
                        </div>
                    </div>
                ) : null}
            </div>
        </>
    );
}

export function ErrorBoundary() {
    const error = useRouteError();
    const message = error instanceof Error ? error.message : 'Create debate failed unexpectedly.';
    return (
        <div className='p-4 text-sm text-error'>
            <strong>Unable to load create debate view:</strong> {message}
        </div>
    );
}
