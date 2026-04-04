import { useEffect, useMemo, useState } from 'react';
import type { LoaderFunctionArgs } from 'react-router-dom';
import {
    Form,
    redirect,
    useActionData,
    useBlocker,
    useLoaderData,
    useNavigate,
    useNavigation,
} from 'react-router-dom';
import { AgentEditCard } from '@/components/edit/AgentEditCard';
import { BlockerModal } from '@/components/edit/BlockerModal';
import { RoundEditor } from '@/components/edit/RoundEditor';
import { Button } from '@/components/ui/Button';
import { ApiError, getErrorMessage } from '@/services/api/http';
import { getDebate, updateDebate } from '@/services/api/debates';
import { DebateSchema } from '@/services/api/schemas';
import type { Agent, Debate, Round } from '@/types';

type AgentDraft = Pick<Agent, 'id' | 'name' | 'stance'>;

type RoundDraft = Round;

type ActionData = {
    error?: string;
};

function getAgentFieldKey(index: number, field: keyof Omit<AgentDraft, 'id'>) {
    return `agents_${index}_${field}`;
}

function getRoundFieldKey(index: number) {
    return `rounds_${index}_message`;
}

function normalizeText(value: string) {
    return value.replace(/\r\n/g, '\n').trim();
}

function parseAgentsFromFormData(formData: FormData) {
    const agentsByIndex = new Map<number, AgentDraft>();

    for (const [key, rawValue] of formData.entries()) {
        if (typeof rawValue !== 'string') continue;
        const match = key.match(/^agent-(\d+)-(id|name|stance)$/);
        if (!match) continue;
        const index = Number.parseInt(match[1], 10);
        const field = match[2] as keyof AgentDraft;
        const current = agentsByIndex.get(index) ?? { id: '', name: '', stance: '' };
        current[field] = rawValue;
        agentsByIndex.set(index, current);
    }

    return Array.from(agentsByIndex.entries())
        .sort(([a], [b]) => a - b)
        .map(([, agent]) => ({
            id: agent.id,
            name: normalizeText(agent.name),
            stance: normalizeText(agent.stance),
        }));
}

function parseRoundsFromFormData(formData: FormData) {
    const roundsByIndex = new Map<number, RoundDraft>();

    for (const [key, rawValue] of formData.entries()) {
        if (typeof rawValue !== 'string') continue;
        const match = key.match(/^round-(\d+)-(agent-id|message)$/);
        if (!match) continue;
        const index = Number.parseInt(match[1], 10);
        const field = match[2] === 'agent-id' ? 'agent_id' : 'message';
        const current = roundsByIndex.get(index) ?? { agent_id: '', message: '' };
        current[field] = field === 'message' ? rawValue.replace(/\r\n/g, '\n') : rawValue;
        roundsByIndex.set(index, current);
    }

    return Array.from(roundsByIndex.entries())
        .sort(([a], [b]) => a - b)
        .map(([, round]) => round);
}

// eslint-disable-next-line react-refresh/only-export-components
export async function loader({ params }: LoaderFunctionArgs) {
    if (!params.debateId) throw new Response('Debate not found', { status: 404 });
    const data = await getDebate(params.debateId);
    return DebateSchema.parse({ id: params.debateId, ...data });
}

// eslint-disable-next-line react-refresh/only-export-components
export async function action({ request, params }: LoaderFunctionArgs) {
    if (!params.debateId) throw new Response('Debate not found', { status: 404 });

    const formData = await request.formData();
    const topic = normalizeText(formData.get('topic')?.toString() ?? '');
    const formAgents = parseAgentsFromFormData(formData);
    const rounds = parseRoundsFromFormData(formData);

    // Re-fetch to get non-editable fields (name, ttsProvider, voiceName etc.)
    const rawDebate = await getDebate(params.debateId);
    const fullDebate = DebateSchema.parse({ id: params.debateId, ...rawDebate }) as Debate;

    const updated: Debate = {
        ...fullDebate,
        topic,
        agents: formAgents.map((formAgent, i) => ({
            ...fullDebate.agents[i],
            id: formAgent.id,
            name: formAgent.name,
            stance: formAgent.stance,
        })),
        rounds,
    };

    try {
        await updateDebate(params.debateId, updated);
        return redirect(`/debates/${encodeURIComponent(params.debateId)}`);
    } catch (error) {
        if (error instanceof ApiError && error.status === 400) {
            return { error: getErrorMessage(error, 'Something went wrong') };
        }
        throw error;
    }
}

function getRoundTone(
    round: Round,
    debate: Debate,
): 'agent-1' | 'agent-2' | 'agent-3' | 'agent-4' | 'agent-user' {
    const agentIndex = debate.agents.findIndex((a) => a.id === round.agent_id);
    if (agentIndex === -1) return 'agent-user';
    return (['agent-1', 'agent-2', 'agent-3', 'agent-4'][agentIndex] ?? 'agent-user') as
        | 'agent-1'
        | 'agent-2'
        | 'agent-3'
        | 'agent-4';
}

export function Component() {
    const debate = useLoaderData() as Debate;
    const actionData = useActionData() as ActionData | undefined;
    const navigate = useNavigate();
    const navigation = useNavigation();
    const [topic, setTopic] = useState(() => debate.topic);
    const [agents, setAgents] = useState<AgentDraft[]>(() => debate.agents.map((a) => ({ ...a })));
    const [rounds, setRounds] = useState<RoundDraft[]>(() => debate.rounds.map((r) => ({ ...r })));
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [allowNavigation, setAllowNavigation] = useState(false);

    const detailPath = `/debates/${encodeURIComponent(debate.id)}`;

    const topicDirty = useMemo(
        () => normalizeText(topic) !== normalizeText(debate.topic),
        [topic, debate.topic],
    );

    const agentsDirty = useMemo(
        () =>
            agents.length !== debate.agents.length ||
            agents.some((a, i) => {
                const init = debate.agents[i];
                return (
                    !init ||
                    normalizeText(a.name) !== normalizeText(init.name) ||
                    normalizeText(a.stance) !== normalizeText(init.stance)
                );
            }),
        [agents, debate.agents],
    );

    const roundsDirty = useMemo(
        () =>
            rounds.length !== debate.rounds.length ||
            rounds.some((r, i) => {
                const init = debate.rounds[i];
                return !init || r.agent_id !== init.agent_id || r.message !== init.message;
            }),
        [rounds, debate.rounds],
    );

    const anyDirty = topicDirty || agentsDirty || roundsDirty;

    const blocker = useBlocker(
        ({ currentLocation, nextLocation }) =>
            anyDirty &&
            !allowNavigation &&
            navigation.state === 'idle' &&
            currentLocation.pathname !== nextLocation.pathname,
    );

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        if (actionData?.error) setAllowNavigation(false);
    }, [actionData]);

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        if (navigation.state === 'idle' && !actionData?.error) setAllowNavigation(false);
    }, [actionData?.error, navigation.state]);

    function clearError(field: string) {
        setErrors((c) => {
            if (!(field in c)) return c;
            const next = { ...c };
            delete next[field];
            return next;
        });
    }

    function updateAgent(index: number, field: keyof Omit<AgentDraft, 'id'>, value: string) {
        setAgents((c) => {
            const next = [...c];
            next[index] = { ...next[index], [field]: value };
            return next;
        });
        clearError(getAgentFieldKey(index, field));
    }

    function commitRound(index: number, message: string) {
        setRounds((c) => {
            const next = [...c];
            next[index] = { ...next[index], message };
            return next;
        });
        clearError(getRoundFieldKey(index));
    }

    return (
        <section className='min-h-full bg-bg-base'>
            <Form
                className='mx-auto flex w-full max-w-5xl flex-col gap-8 px-4 py-5 pb-32 sm:px-6 sm:py-6'
                method='post'
                onSubmit={() => setAllowNavigation(true)}
            >
                {actionData?.error ? (
                    <div className='rounded-lg border border-error/30 bg-error/10 px-4 py-3 text-sm text-error'>
                        {actionData.error}
                    </div>
                ) : null}

                <header className=''>
                    <div className='mt-3 flex flex-wrap items-end justify-between gap-4'>
                        <div>
                            <h1 className='font-display text-2xl sm:text-3xl leading-none text-text-1'>
                                Configuration
                            </h1>
                        </div>
                    </div>
                </header>

                <section className='space-y-4'>
                    <div>
                        <p className='font-mono text-[10px] uppercase tracking-[0.18em] text-text-3'>
                            Topic
                        </p>
                    </div>
                    <div>
                        <textarea
                            id='topic'
                            name='topic'
                            value={topic}
                            rows={5}
                            className='w-full max-w-[680px] rounded border border-border bg-bg-surface px-3 py-2 text-sm text-text-1 placeholder:text-text-3 focus:border-accent-dim focus:outline-none focus:ring-1 focus:ring-accent/30'
                            onChange={(e) => {
                                setTopic(e.target.value);
                                clearError('topic');
                            }}
                        />
                        {errors.topic ? (
                            <p className='mt-1 text-xs text-error'>{errors.topic}</p>
                        ) : null}
                    </div>
                </section>

                <section className='space-y-4'>
                    <div>
                        <p className='font-mono text-[10px] uppercase tracking-[0.18em] text-text-3'>
                            Agents
                        </p>
                        <h2 className='mt-2 font-display text-2xl text-text-1'>Speaker cards</h2>
                    </div>
                    <div className='grid gap-4 lg:grid-cols-2'>
                        {agents.map((agent, index) => (
                            <AgentEditCard
                                key={agent.id}
                                index={index}
                                agent={agent}
                                errors={errors}
                                onChange={(field, value) => updateAgent(index, field, value)}
                                onBlur={(field) => clearError(getAgentFieldKey(index, field))}
                            />
                        ))}
                    </div>
                </section>

                <section className='space-y-4'>
                    <div>
                        <p className='font-mono text-[10px] uppercase tracking-[0.18em] text-text-3'>
                            Rounds
                        </p>
                        <h2 className='mt-2 font-display text-2xl text-text-1'>Round transcript</h2>
                    </div>
                    <div className='space-y-4'>
                        {rounds.map((round, index) => {
                            const speaker = debate.agents.find((a) => a.id === round.agent_id);
                            return (
                                <RoundEditor
                                    key={`${round.agent_id || 'user'}-${index}`}
                                    index={index}
                                    agentId={round.agent_id}
                                    message={round.message}
                                    speakerLabel={speaker?.name ?? 'Moderator prompt'}
                                    tone={getRoundTone(round, debate)}
                                    error={errors[getRoundFieldKey(index)]}
                                    onCommit={(message) => commitRound(index, message)}
                                />
                            );
                        })}
                    </div>
                </section>

                <div className='fixed bottom-6 right-6 z-20 flex items-center gap-3 rounded-2xl border border-border bg-bg-panel/95 px-4 py-3 shadow-2xl backdrop-blur-sm'>
                    <div className='hidden text-right sm:block'>
                        <p className='text-xs text-text-2'>
                            {anyDirty ? 'Unsaved edits' : 'No local changes'}
                        </p>
                        <p className='font-mono text-[10px] uppercase tracking-[0.16em] text-text-3'>
                            {topicDirty ? 'Topic ' : ''}
                            {agentsDirty ? 'Agents ' : ''}
                            {roundsDirty ? 'Rounds' : ''}
                        </p>
                    </div>
                    <Button type='button' variant='ghost' onClick={() => navigate(detailPath)}>
                        Back to debate
                    </Button>
                    <Button
                        type='submit'
                        variant='accent'
                        disabled={navigation.state !== 'idle'}
                        data-testid='edit-save-button'
                    >
                        {navigation.state === 'submitting' ? 'Saving...' : 'Save changes'}
                    </Button>
                </div>
            </Form>

            <BlockerModal
                open={blocker.state === 'blocked'}
                onStay={() => blocker.reset?.()}
                onLeave={() => blocker.proceed?.()}
            />
        </section>
    );
}

export { EditErrorBoundary as ErrorBoundary } from './ErrorBoundary';
