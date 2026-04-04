import { useMemo, useState, useCallback, useEffect, useRef } from 'react';
import { Markdown } from '@/components/ui/Markdown';
import { useToast } from '@/components/ui/ToastProvider';
import {
    useRoundAudioCache,
    useRoundAudioStatus,
    useIsRoundPlaying,
} from '@/hooks/useRoundAudioCache';
import type { Agent, Round } from '@/types';

type RoundTone = 'agent-1' | 'agent-2' | 'agent-3' | 'agent-4' | 'agent-user';

export interface RoundBubbleProps {
    round: Round;
    agent: Agent | undefined;
    index: number;
    streaming?: boolean;
    thinking?: boolean;
    preparingSpeech?: boolean;
    tone?: RoundTone;
    messageType?: 'Opening' | 'Rebuttal';
    highlightedAgentId?: string | null;
    alignment?: 'left' | 'right';
    /** If true, skip truncation (for recent messages) */
    isRecent?: boolean;
    /** Debate ID for audio playback */
    debateId?: string;
}

// ---------------------------------------------------------------------------
// Audio control icons
// ---------------------------------------------------------------------------

function IconPlaySmall({ className }: { className?: string }) {
    return (
        <svg viewBox='0 0 24 24' fill='currentColor' className={className ?? 'w-3 h-3'} aria-hidden>
            <path d='M6 4.75a.75.75 0 0 1 1.14-.64l12 7.25a.75.75 0 0 1 0 1.28l-12 7.25A.75.75 0 0 1 6 19.25v-14.5z' />
        </svg>
    );
}

function IconPauseSmall({ className }: { className?: string }) {
    return (
        <svg viewBox='0 0 24 24' fill='currentColor' className={className ?? 'w-3 h-3'} aria-hidden>
            <rect x='5' y='4' width='4' height='16' rx='1.25' />
            <rect x='15' y='4' width='4' height='16' rx='1.25' />
        </svg>
    );
}

function IconSpinner({ className }: { className?: string }) {
    return (
        <svg
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='2.5'
            strokeLinecap='round'
            className={`${className ?? 'w-3 h-3'} animate-spin`}
            aria-hidden
        >
            <circle cx='12' cy='12' r='9' strokeOpacity='0.25' />
            <path d='M12 3a9 9 0 0 1 9 9' />
        </svg>
    );
}

function IconError({ className }: { className?: string }) {
    return (
        <svg
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='2'
            strokeLinecap='round'
            className={className ?? 'w-3 h-3'}
            aria-hidden
        >
            <path d='M18 6L6 18M6 6l12 12' />
        </svg>
    );
}

function IconReset({ className }: { className?: string }) {
    return (
        <svg
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='2'
            strokeLinecap='round'
            strokeLinejoin='round'
            className={className ?? 'w-3 h-3'}
            aria-hidden
        >
            <path d='M3 12a9 9 0 1 0 .75-3.6' />
            <polyline points='3 4.5 3 8.5 7 8.5' />
        </svg>
    );
}

function formatDuration(seconds: number): string {
    const m = Math.floor(seconds / 60);
    const s = Math.floor(seconds % 60);
    return `${m}:${String(s).padStart(2, '0')}`;
}

const toneStyles: Record<RoundTone, { color: string; bg: string; border: string }> = {
    'agent-1': { color: 'var(--c-a)', bg: 'var(--c-a-bg)', border: 'rgba(78, 143, 104, 0.3)' },
    'agent-2': { color: 'var(--c-b)', bg: 'var(--c-b-bg)', border: 'rgba(176, 90, 60, 0.3)' },
    'agent-3': { color: 'var(--c-c)', bg: 'var(--c-c-bg)', border: 'rgba(82, 120, 168, 0.3)' },
    'agent-4': { color: 'var(--c-d)', bg: 'var(--c-d-bg)', border: 'rgba(138, 98, 192, 0.3)' },
    'agent-user': {
        color: 'var(--c-user)',
        bg: 'var(--c-user-bg)',
        border: 'rgba(122, 144, 96, 0.3)',
    },
};

const WORD_LIMIT = 80;

function getInitials(agent: Agent | undefined) {
    if (!agent) {
        return 'YOU';
    }

    return agent.name
        .split(' ')
        .map((word) => word[0] ?? '')
        .join('')
        .slice(0, 2)
        .toUpperCase();
}

function normalizeMessage(text: string) {
    return text.replace(/<pause\d+>/g, ' ');
}

function countWords(text: string): number {
    return text.split(/\s+/).filter(Boolean).length;
}

export function RoundBubble({
    round,
    agent,
    index,
    streaming = false,
    thinking = false,
    preparingSpeech = false,
    tone = 'agent-user',
    messageType,
    highlightedAgentId,
    alignment = 'left',
    isRecent = false,
    debateId,
}: RoundBubbleProps) {
    const styles = toneStyles[tone];
    const toast = useToast();
    const label = agent?.name ?? 'You';
    const isUser = !agent;
    const isDimmed = highlightedAgentId && agent?.id !== highlightedAgentId;
    const isHighlighted = highlightedAgentId && agent?.id === highlightedAgentId;
    const isRight = alignment === 'right';

    // Audio playback (index is 1-based display, convert to 0-based for API)
    const roundIndex = index - 1;
    const audioState = useRoundAudioStatus(debateId, roundIndex);
    const previousAudioStatusRef = useRef(audioState.status);
    const isPlaying = useIsRoundPlaying(debateId ?? '', roundIndex);
    const { playNow, reset, togglePause } = useRoundAudioCache();
    const canPlayAudio =
        !!debateId && roundIndex >= 0 && !streaming && !thinking && !preparingSpeech;

    const handlePlayClick = useCallback(
        (e: React.MouseEvent) => {
            e.stopPropagation();
            if (!debateId || !canPlayAudio) return;

            // If this round is already playing, toggle pause
            if (isPlaying) {
                togglePause();
            } else {
                // Play this round immediately (interrupts queue)
                void playNow(debateId, roundIndex);
            }
        },
        [debateId, roundIndex, canPlayAudio, isPlaying, playNow, togglePause],
    );

    const handleResetClick = useCallback(
        (e: React.MouseEvent) => {
            e.stopPropagation();
            if (debateId && canPlayAudio) {
                reset(debateId, roundIndex);
            }
        },
        [debateId, roundIndex, canPlayAudio, reset],
    );

    // Compute progress percentage for progress bar
    const progressPercent =
        audioState.duration && audioState.duration > 0
            ? (audioState.currentTime / audioState.duration) * 100
            : 0;

    useEffect(() => {
        const previousStatus = previousAudioStatusRef.current;

        if (audioState.status === 'error' && previousStatus !== 'error' && canPlayAudio) {
            toast.error(audioState.errorMessage ?? 'Something went wrong', {
                title: 'Round audio failed',
                id: `round-audio-error-${debateId ?? 'unknown'}-${roundIndex}`,
            });
        }

        previousAudioStatusRef.current = audioState.status;
    }, [audioState.errorMessage, audioState.status, canPlayAudio, debateId, roundIndex, toast]);

    const normalizedMessage = useMemo(() => normalizeMessage(round.message), [round.message]);
    const wordCount = useMemo(() => countWords(normalizedMessage), [normalizedMessage]);
    const shouldTruncate = !isRecent && !streaming && wordCount > WORD_LIMIT;
    const [isExpanded, setIsExpanded] = useState(false);

    const displayMessage = useMemo(() => {
        if (!shouldTruncate || isExpanded) {
            return round.message;
        }
        const words = round.message.split(/\s+/);
        return `${words.slice(0, WORD_LIMIT).join(' ')}…`;
    }, [round.message, shouldTruncate, isExpanded]);

    const Avatar = (
        <div
            className='flex h-7 w-7 shrink-0 items-center justify-center rounded-md font-mono text-[10px] font-semibold mt-px'
            style={{ color: styles.color, backgroundColor: styles.bg }}
            aria-hidden='true'
        >
            {getInitials(agent)}
        </div>
    );

    return (
        <div
            className={`flex gap-2.5 items-start animate-[fadeSlide_0.25s_ease_both] transition-all duration-200 max-w-[85%] ${
                isDimmed ? 'opacity-25' : ''
            } ${isRight ? 'ml-auto flex-row-reverse' : ''}`}
            data-testid={
                thinking
                    ? 'thinking-bubble'
                    : preparingSpeech
                      ? 'preparing-speech-bubble'
                      : streaming
                        ? 'streaming-bubble'
                        : 'round-bubble'
            }
        >
            {Avatar}

            <div
                className={`min-w-0 flex-1 rounded-xl px-4 py-3 border transition-all duration-200 ${
                    shouldTruncate && !isExpanded ? 'cursor-pointer hover:border-opacity-60' : ''
                }`}
                style={{
                    backgroundColor: 'var(--bg-surface)',
                    borderColor: isHighlighted ? styles.color : styles.border,
                    borderWidth: isHighlighted ? '2px' : '1px',
                    boxShadow: isHighlighted ? `0 0 12px ${styles.border}` : 'none',
                }}
                onClick={shouldTruncate ? () => setIsExpanded(!isExpanded) : undefined}
            >
                <div className='mb-2 flex items-baseline gap-2'>
                    <span className='text-[11px] font-semibold' style={{ color: styles.color }}>
                        {label}
                    </span>
                    <span className='font-mono text-[9px] text-text-3'>#{index}</span>
                    {messageType && (
                        <span
                            className={`text-[9px] px-1.25 py-px rounded-[3px] ${
                                messageType === 'Opening'
                                    ? 'bg-accent/20 text-accent'
                                    : 'bg-(--c-b)/20 text-agent-2'
                            }`}
                        >
                            {messageType}
                        </span>
                    )}
                    {isUser && (
                        <span
                            className='text-[9px] px-1.25 py-px rounded-[3px] border'
                            style={{
                                background: 'var(--c-user-bg)',
                                color: 'var(--c-user)',
                                borderColor: '#2a3820',
                            }}
                        >
                            injected
                        </span>
                    )}
                </div>

                {thinking ? (
                    <div className='flex items-center gap-2 py-1' data-testid='thinking-indicator'>
                        <div className='flex gap-1'>
                            <span
                                className='inline-block h-1.5 w-1.5 rounded-full animate-[pulse_1.2s_ease-in-out_infinite] opacity-60'
                                style={{ backgroundColor: styles.color }}
                            />
                            <span
                                className='inline-block h-1.5 w-1.5 rounded-full animate-[pulse_1.2s_ease-in-out_0.2s_infinite] opacity-60'
                                style={{ backgroundColor: styles.color }}
                            />
                            <span
                                className='inline-block h-1.5 w-1.5 rounded-full animate-[pulse_1.2s_ease-in-out_0.4s_infinite] opacity-60'
                                style={{ backgroundColor: styles.color }}
                            />
                        </div>
                        <span className='text-[12px] text-text-3 italic'>{label} is thinking…</span>
                    </div>
                ) : preparingSpeech ? (
                    <div
                        className='flex items-center gap-2 py-1'
                        data-testid='preparing-speech-indicator'
                    >
                        <svg
                            className='h-4 w-4 animate-spin text-accent'
                            fill='none'
                            viewBox='0 0 24 24'
                        >
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
                        <span className='text-[12px] text-accent italic'>
                            Preparing agent speech…
                        </span>
                    </div>
                ) : (
                    <div className='space-y-3'>
                        {/* Primary: Metadata + Audio control row */}
                        <div className='flex items-start gap-3' data-testid='round-metadata'>
                            {/* Metadata: voice & summary */}
                            <div className='flex flex-wrap items-center gap-x-4 gap-y-1.5 flex-1 min-w-0'>
                                {round.summary && (
                                    <div className='flex items-start gap-1.5 flex-1 min-w-0'>
                                        <svg
                                            className='h-3.5 w-3.5 shrink-0 mt-0.5'
                                            style={{ color: styles.color }}
                                            fill='none'
                                            viewBox='0 0 24 24'
                                            stroke='currentColor'
                                            strokeWidth={2}
                                        >
                                            <path
                                                strokeLinecap='round'
                                                strokeLinejoin='round'
                                                d='M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z'
                                            />
                                        </svg>
                                        <span
                                            className='text-[13px] font-medium leading-snug text-text-1'
                                            data-testid='round-summary'
                                        >
                                            {round.summary}
                                        </span>
                                    </div>
                                )}
                            </div>

                            {/* Audio play button */}
                            {canPlayAudio && (
                                <div className='flex items-center gap-1 shrink-0'>
                                    {/* Reset button (only show when audio has been loaded/played) */}
                                    {(audioState.status === 'ready' || isPlaying) &&
                                        audioState.currentTime > 0 && (
                                            <button
                                                type='button'
                                                onClick={handleResetClick}
                                                className='flex items-center justify-center w-6 h-6 rounded-full border transition-colors cursor-pointer hover:border-accent hover:text-accent'
                                                style={{
                                                    borderColor: styles.border,
                                                    color: 'var(--text-3)',
                                                }}
                                                title='Reset to beginning'
                                                aria-label={`Reset round ${index}`}
                                                data-testid='round-reset-button'
                                            >
                                                <IconReset className='w-2.5 h-2.5' />
                                            </button>
                                        )}
                                    {/* Play/Pause button */}
                                    <button
                                        type='button'
                                        onClick={handlePlayClick}
                                        disabled={audioState.status === 'loading'}
                                        className='flex items-center gap-1.5 rounded-full border px-2 py-1 text-[10px] font-mono transition-colors cursor-pointer disabled:cursor-wait'
                                        style={{
                                            borderColor:
                                                audioState.status === 'error'
                                                    ? 'var(--error)'
                                                    : styles.border,
                                            color:
                                                audioState.status === 'error'
                                                    ? 'var(--error)'
                                                    : isPlaying
                                                      ? styles.color
                                                      : 'var(--text-3)',
                                            backgroundColor: isPlaying ? styles.bg : 'transparent',
                                        }}
                                        title={
                                            audioState.status === 'error'
                                                ? 'Audio unavailable'
                                                : audioState.status === 'loading'
                                                  ? 'Loading audio…'
                                                  : isPlaying
                                                    ? 'Pause'
                                                    : 'Play round audio'
                                        }
                                        aria-label={
                                            isPlaying
                                                ? `Pause round ${index}`
                                                : `Play round ${index}`
                                        }
                                        data-testid='round-play-button'
                                    >
                                        {audioState.status === 'loading' ? (
                                            <IconSpinner className='w-2.5 h-2.5' />
                                        ) : audioState.status === 'error' ? (
                                            <IconError className='w-2.5 h-2.5' />
                                        ) : isPlaying ? (
                                            <IconPauseSmall className='w-2.5 h-2.5' />
                                        ) : (
                                            <IconPlaySmall className='w-2.5 h-2.5' />
                                        )}
                                        {audioState.duration != null && (
                                            <span>
                                                {isPlaying
                                                    ? `${formatDuration(audioState.currentTime)} / ${formatDuration(audioState.duration)}`
                                                    : formatDuration(audioState.duration)}
                                            </span>
                                        )}
                                    </button>
                                </div>
                            )}
                        </div>

                        {/* Audio progress bar (shown when audio is loaded and playing/ready) */}
                        {canPlayAudio &&
                            audioState.status !== 'idle' &&
                            audioState.status !== 'loading' &&
                            audioState.status !== 'error' &&
                            audioState.duration != null && (
                                <div
                                    className='h-[3px] rounded-full overflow-hidden'
                                    style={{ backgroundColor: styles.border }}
                                    data-testid='round-progress-bar'
                                >
                                    <div
                                        className='h-full rounded-full transition-[width] duration-200'
                                        style={{
                                            width: `${progressPercent}%`,
                                            backgroundColor: styles.color,
                                        }}
                                    />
                                </div>
                            )}

                        {/* Secondary: Message content */}
                        <div
                            className='text-[13px] leading-relaxed text-text-2 border-l-2 pl-3'
                            style={{ borderColor: styles.border }}
                        >
                            <Markdown content={displayMessage} streaming={streaming} />
                        </div>
                        {shouldTruncate && (
                            <button
                                type='button'
                                className='text-[10px] text-accent hover:text-accent-hi transition-colors cursor-pointer'
                                onClick={(e) => {
                                    e.stopPropagation();
                                    setIsExpanded(!isExpanded);
                                }}
                            >
                                {isExpanded ? '▲ Show less' : `▼ Show more (${wordCount} words)`}
                            </button>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
}
