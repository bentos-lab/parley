import { useEffect, useId, useState } from 'react';
import type { Debate } from '@/types';

// Simple volume icons as inline SVGs
function VolumeIcon({ muted, className }: { muted?: boolean; className?: string }) {
    if (muted) {
        return (
            <svg
                className={className}
                fill='none'
                viewBox='0 0 24 24'
                stroke='currentColor'
                strokeWidth={2}
            >
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
            </svg>
        );
    }
    return (
        <svg
            className={className}
            fill='none'
            viewBox='0 0 24 24'
            stroke='currentColor'
            strokeWidth={2}
        >
            <path
                strokeLinecap='round'
                strokeLinejoin='round'
                d='M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z'
            />
        </svg>
    );
}

const agentColors = [
    { color: 'var(--c-a)', bg: 'var(--c-a-bg)', border: 'var(--c-a-border)' },
    { color: 'var(--c-b)', bg: 'var(--c-b-bg)', border: 'var(--c-b-border)' },
    { color: 'var(--c-c)', bg: 'var(--c-c-bg)', border: 'var(--c-c-border)' },
    { color: 'var(--c-d)', bg: 'var(--c-d-bg)', border: 'var(--c-d-border)' },
];

export interface AgentPanelProps {
    debate: Debate;
    streamingCount?: number;
    isStreaming?: boolean;
    isStopping?: boolean;
    sseCompleted?: boolean;
    turnCount?: number;
    onTurnCountChange?: (n: number) => void;
    playAudioImmediately?: boolean;
    onPlayAudioChange?: (enabled: boolean) => void;
    onStart?: () => void;
    onStop?: () => void;
    onStartNext?: () => void;
    highlightedAgentId?: string | null;
    onAgentHighlight?: (agentId: string | null) => void;
    mobileOpen?: boolean;
    onMobileToggle?: () => void;
    onMobileClose?: () => void;
}

export function AgentPanel({
    debate,
    streamingCount = 0,
    isStreaming = false,
    isStopping = false,
    sseCompleted = false,
    turnCount = 6,
    onTurnCountChange,
    playAudioImmediately = false,
    onPlayAudioChange,
    onStart,
    onStop,
    onStartNext,
    highlightedAgentId,
    onAgentHighlight,
    mobileOpen = false,
    onMobileToggle,
    onMobileClose,
}: AgentPanelProps) {
    const [isAgentsExpanded, setIsAgentsExpanded] = useState(true);
    const mobilePanelId = useId();
    const totalTurns = debate.rounds.length + streamingCount;

    const agentStats = debate.agents.map((agent, index) => {
        const turns = debate.rounds.filter((r) => r.agent_id === agent.id).length;
        const share = totalTurns > 0 ? turns / totalTurns : 1 / debate.agents.length;
        const palette = agentColors[index % agentColors.length];
        return { agent, turns, share, ...palette };
    });

    useEffect(() => {
        if (!mobileOpen) {
            return undefined;
        }

        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key === 'Escape') {
                onMobileClose?.();
            }
        };

        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [mobileOpen, onMobileClose]);

    const handleStartClick = () => {
        onStart?.();
        onMobileClose?.();
    };

    const handleStartNextClick = () => {
        onStartNext?.();
        onMobileClose?.();
    };

    const panelContent = (
        <>
            {/* Agents section */}
            <div className='border-b border-border px-4 py-3.5'>
                <button
                    type='button'
                    onClick={() => setIsAgentsExpanded(!isAgentsExpanded)}
                    className='mb-2.5 flex w-full cursor-pointer items-center justify-between'
                >
                    <span className='text-[11px] font-semibold uppercase tracking-[0.08em] text-text-3'>
                        Agents · {debate.agents.length}
                    </span>
                    <svg
                        className={`h-3 w-3 text-text-3 transition-transform duration-200 ${
                            isAgentsExpanded ? 'rotate-0' : '-rotate-90'
                        }`}
                        fill='none'
                        viewBox='0 0 24 24'
                        stroke='currentColor'
                        strokeWidth={2}
                    >
                        <path strokeLinecap='round' strokeLinejoin='round' d='M19 9l-7 7-7-7' />
                    </svg>
                </button>

                <div
                    className={`space-y-2 overflow-hidden transition-all duration-200 ease-in-out ${
                        isAgentsExpanded ? 'max-h-125 opacity-100' : 'max-h-0 opacity-0'
                    }`}
                >
                    {agentStats.map(({ agent, turns, color }) => (
                        <div
                            key={agent.id}
                            data-testid='agent-card'
                            onClick={() => {
                                onAgentHighlight?.(
                                    highlightedAgentId === agent.id ? null : agent.id,
                                );
                            }}
                            className={`cursor-pointer rounded-lg border p-3 transition-all ${
                                highlightedAgentId === agent.id
                                    ? 'border-2'
                                    : 'bg-bg-surface hover:border-border-hi'
                            }`}
                            style={{
                                borderColor:
                                    highlightedAgentId === agent.id ? color : 'var(--border)',
                            }}
                        >
                            <div className='mb-1.5 flex items-center gap-1.75'>
                                <span
                                    className='h-2 w-2 shrink-0 rounded-full'
                                    style={{ backgroundColor: color }}
                                />
                                <span className='flex-1 truncate text-sm font-semibold text-text-1'>
                                    {agent.name}
                                </span>
                                <span className='rounded-[3px] bg-bg-elevated px-1.25 py-px font-mono text-[11px] text-text-3'>
                                    {turns}
                                </span>
                            </div>

                            {agent.stance && (
                                <p className='text-[13px] leading-[1.4] text-text-2'>
                                    {agent.stance}
                                </p>
                            )}
                        </div>
                    ))}
                </div>
            </div>

            {/* Session section */}
            <div className='border-b border-border px-4 py-3.5'>
                <div className='mb-2.5 text-[11px] font-semibold uppercase tracking-[0.08em] text-text-3'>
                    Session
                </div>

                <div className='space-y-1.75'>
                    <div className='flex items-center justify-between'>
                        <span className='text-[13px] text-text-3'>Total rounds</span>
                        <span className='font-mono text-xs text-text-2'>{totalTurns}</span>
                    </div>
                    <div className='flex items-center justify-between'>
                        <span className='text-[13px] text-text-3'>TTS provider</span>
                        <span className='font-mono text-xs text-accent'>{debate.ttsProvider}</span>
                    </div>
                </div>
            </div>

            {/* Controls section */}
            <div className='px-4 py-3.5'>
                <div className='mb-2.5 text-[11px] font-semibold uppercase tracking-[0.08em] text-text-3'>
                    Controls
                </div>

                <div className='mb-3'>
                    <div className='mb-1.5 flex items-center justify-between'>
                        <span className='text-[13px] text-text-3'>Rounds to generate</span>
                        <span className='font-mono text-xs text-text-2'>{turnCount}</span>
                    </div>
                    <input
                        type='range'
                        min={1}
                        max={20}
                        value={turnCount}
                        onChange={(e) => onTurnCountChange?.(Number(e.target.value))}
                        disabled={isStreaming || isStopping}
                        className='w-full accent-accent disabled:opacity-40'
                    />
                    <div className='mt-0.5 flex justify-between text-[11px] text-text-3'>
                        <span>1</span>
                        <span>20</span>
                    </div>
                </div>

                <div className='mb-3'>
                    <button
                        type='button'
                        onClick={() => onPlayAudioChange?.(!playAudioImmediately)}
                        disabled={isStreaming || isStopping}
                        title='Next rounds will play audio automatically'
                        aria-label={`Auto-play audio: ${playAudioImmediately ? 'enabled' : 'disabled'}. Next rounds will play audio automatically.`}
                        className={`flex w-full items-center justify-between rounded-md border px-3 py-2 text-sm transition-colors ${
                            playAudioImmediately
                                ? 'border-accent bg-accent/10 text-accent'
                                : 'border-border bg-bg-surface text-text-2 hover:border-border-hi'
                        } disabled:cursor-not-allowed disabled:opacity-40`}
                    >
                        <span className='flex items-center gap-2'>
                            <VolumeIcon muted={!playAudioImmediately} className='h-4 w-4' />
                            <span>Auto-play audio</span>
                        </span>
                        <span
                            className={`h-4 w-7 rounded-full transition-colors ${
                                playAudioImmediately ? 'bg-accent' : 'bg-border'
                            }`}
                        >
                            <span
                                className={`block h-3 w-3 translate-y-0.5 rounded-full bg-white transition-transform ${
                                    playAudioImmediately ? 'translate-x-3.5' : 'translate-x-0.5'
                                }`}
                            />
                        </span>
                    </button>
                </div>

                {!isStreaming && !sseCompleted && !isStopping && (
                    <button
                        type='button'
                        onClick={handleStartClick}
                        className='w-full rounded-md bg-accent px-3 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90 cursor-pointer'
                    >
                        Start
                    </button>
                )}
                {(isStreaming || isStopping) && (
                    <button
                        type='button'
                        onClick={onStop}
                        disabled={isStopping}
                        className={`flex w-full items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-medium text-white transition-all ${
                            isStopping
                                ? 'cursor-not-allowed bg-red-600/60'
                                : 'cursor-pointer bg-red-600 hover:opacity-90'
                        }`}
                    >
                        {isStopping && (
                            <svg className='h-4 w-4 animate-spin' fill='none' viewBox='0 0 24 24'>
                                <circle
                                    className='opacity-25'
                                    cx='12'
                                    cy='12'
                                    r='10'
                                    stroke='currentColor'
                                    strokeWidth='4'
                                />
                                <path
                                    className='opacity-75'
                                    fill='currentColor'
                                    d='M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z'
                                />
                            </svg>
                        )}
                        {isStopping ? 'Stopping…' : 'Stop'}
                    </button>
                )}
                {sseCompleted && !isStopping && (
                    <div className='space-y-2'>
                        <p className='text-center text-xs text-text-3'>
                            Session complete — {totalTurns} turns
                        </p>
                        <button
                            type='button'
                            onClick={handleStartNextClick}
                            className='w-full rounded-md bg-accent px-3 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90 cursor-pointer'
                        >
                            Start Next
                        </button>
                    </div>
                )}
            </div>
        </>
    );

    return (
        <>
            <button
                type='button'
                aria-label={mobileOpen ? 'Close session controls' : 'Open session controls'}
                aria-expanded={mobileOpen}
                aria-controls={mobilePanelId}
                onClick={onMobileToggle}
                className='fixed bottom-4 right-4 z-30 flex items-center gap-2 rounded-full border border-accent/40 bg-bg-panel/95 px-3 py-2 text-sm text-text-1 shadow-[0_10px_30px_rgba(0,0,0,0.32)] backdrop-blur lg:hidden cursor-pointer transition-colors hover:border-accent hover:bg-bg-elevated'
            >
                <span className='flex h-8 w-8 items-center justify-center rounded-full bg-accent/12 text-accent'>
                    <svg
                        aria-hidden='true'
                        className='h-4 w-4'
                        fill='none'
                        viewBox='0 0 24 24'
                        stroke='currentColor'
                        strokeWidth={2}
                    >
                        <path
                            strokeLinecap='round'
                            strokeLinejoin='round'
                            d='M4 7h16M7 12h10M10 17h4'
                        />
                    </svg>
                </span>
                <span className='flex flex-col items-start leading-none'>
                    <span className='text-[10px] uppercase tracking-[0.12em] text-text-3'>
                        Session
                    </span>
                    <span className='text-sm font-medium text-text-1'>Controls</span>
                </span>
            </button>

            <div
                className={`fixed inset-0 z-40 lg:hidden ${mobileOpen ? '' : 'pointer-events-none'}`}
                aria-hidden={!mobileOpen}
            >
                <button
                    type='button'
                    aria-label='Close session controls'
                    onClick={onMobileClose}
                    className={`absolute inset-0 bg-black/45 transition-opacity duration-200 ${
                        mobileOpen ? 'opacity-100' : 'opacity-0'
                    }`}
                />

                <aside
                    id={mobilePanelId}
                    role='dialog'
                    aria-modal='true'
                    aria-label='Session controls'
                    data-testid='agent-panel-mobile'
                    className={`absolute inset-y-0 right-0 flex w-[min(88vw,320px)] flex-col overflow-y-auto border-l border-border bg-bg-panel shadow-[0_18px_50px_rgba(0,0,0,0.45)] transition-transform duration-200 ease-out ${
                        mobileOpen ? 'translate-x-0' : 'translate-x-full'
                    }`}
                >
                    <div className='sticky top-0 z-10 flex items-center justify-between border-b border-border bg-bg-panel/95 px-4 py-3 backdrop-blur'>
                        <div>
                            <p className='text-[10px] uppercase tracking-[0.12em] text-text-3'>
                                Debate controls
                            </p>
                            <p className='text-sm font-medium text-text-1'>Configure next rounds</p>
                        </div>
                        <button
                            type='button'
                            aria-label='Close session controls'
                            onClick={onMobileClose}
                            className='flex h-8 w-8 items-center justify-center rounded-md text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1 cursor-pointer'
                        >
                            <svg
                                aria-hidden='true'
                                className='h-4 w-4'
                                fill='none'
                                viewBox='0 0 24 24'
                                stroke='currentColor'
                                strokeWidth={2}
                            >
                                <path
                                    strokeLinecap='round'
                                    strokeLinejoin='round'
                                    d='M6 6l12 12M18 6L6 18'
                                />
                            </svg>
                        </button>
                    </div>

                    {panelContent}
                </aside>
            </div>

            <aside
                data-testid='agent-panel'
                className='hidden w-65 min-w-65 flex-col overflow-y-auto bg-bg-panel scrollbar-thin lg:flex'
            >
                {panelContent}
            </aside>
        </>
    );
}
