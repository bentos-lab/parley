import { useMemo, useState } from 'react';
import { Markdown } from '@/components/ui/Markdown';
import type { Agent, Round } from '@/types';

type RoundTone = 'agent-1' | 'agent-2' | 'agent-3' | 'agent-4' | 'agent-user';

export interface RoundBubbleProps {
    round: Round;
    agent: Agent | undefined;
    index: number;
    streaming?: boolean;
    thinking?: boolean;
    tone?: RoundTone;
    messageType?: 'Opening' | 'Rebuttal';
    highlightedAgentId?: string | null;
    alignment?: 'left' | 'right';
    /** If true, skip truncation (for recent messages) */
    isRecent?: boolean;
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
    tone = 'agent-user',
    messageType,
    highlightedAgentId,
    alignment = 'left',
    isRecent = false,
}: RoundBubbleProps) {
    const styles = toneStyles[tone];
    const label = agent?.name ?? 'You';
    const isUser = !agent;
    const isDimmed = highlightedAgentId && agent?.id !== highlightedAgentId;
    const isHighlighted = highlightedAgentId && agent?.id === highlightedAgentId;
    const isRight = alignment === 'right';

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
                thinking ? 'thinking-bubble' : streaming ? 'streaming-bubble' : 'round-bubble'
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
                ) : (
                    <div className='space-y-1'>
                        <Markdown content={displayMessage} streaming={streaming} />
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
