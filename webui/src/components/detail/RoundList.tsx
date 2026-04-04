import type { Ref } from 'react';
import type { Agent, Debate, Round } from '@/types';
import { RoundBubble, type RoundBubbleProps } from './RoundBubble';

export interface RoundListProps {
    debate: Debate;
    scrollRef?: Ref<HTMLDivElement>;
    streamingRounds?: Round[];
    isThinking?: boolean;
    autoScrollPaused?: boolean;
    onScrollToBottom?: () => void;
    highlightedAgentId?: string | null;
}

function getTone(agentId: string, agentMap: Map<string, { agent: Agent; index: number }>) {
    if (!agentId) {
        return 'agent-user' satisfies NonNullable<RoundBubbleProps['tone']>;
    }

    const agentEntry = agentMap.get(agentId);
    const toneIndex = ((agentEntry?.index ?? 0) % 4) + 1;

    switch (toneIndex) {
        case 1:
            return 'agent-1' satisfies NonNullable<RoundBubbleProps['tone']>;
        case 2:
            return 'agent-2' satisfies NonNullable<RoundBubbleProps['tone']>;
        case 3:
            return 'agent-3' satisfies NonNullable<RoundBubbleProps['tone']>;
        default:
            return 'agent-4' satisfies NonNullable<RoundBubbleProps['tone']>;
    }
}

const THINKING_AGENT = { id: '', name: 'Agent', stance: '', voiceName: '' };

/**
 * Calculate alignment for each round in a chat-like layout.
 * Messages alternate sides when the agent changes, but stay on the same side
 * for consecutive messages from the same agent.
 */
function calculateAlignments(rounds: Round[]): ('left' | 'right')[] {
    if (rounds.length === 0) return [];

    const alignments: ('left' | 'right')[] = [];
    let currentSide: 'left' | 'right' = 'left';

    for (let i = 0; i < rounds.length; i++) {
        const currentAgentId = rounds[i].agent_id || '__user__';
        const prevAgentId = i > 0 ? rounds[i - 1].agent_id || '__user__' : null;

        // First message or same agent as previous: keep current side
        // Different agent: switch sides
        if (prevAgentId !== null && currentAgentId !== prevAgentId) {
            currentSide = currentSide === 'left' ? 'right' : 'left';
        }

        alignments.push(currentSide);
    }

    return alignments;
}

export function RoundList({
    debate,
    scrollRef,
    streamingRounds = [],
    isThinking = false,
    autoScrollPaused = false,
    onScrollToBottom,
    highlightedAgentId,
}: RoundListProps) {
    const agentMap = new Map<string, { agent: Agent; index: number }>(
        debate.agents.map((agent, index) => [agent.id, { agent, index }]),
    );

    // Track first appearance of each agent for Opening/Rebuttal badges
    const agentFirstAppearance = new Map<string, number>();
    const allRounds = [...debate.rounds, ...streamingRounds];

    allRounds.forEach((round, idx) => {
        if (round.agent_id && !agentFirstAppearance.has(round.agent_id)) {
            agentFirstAppearance.set(round.agent_id, idx);
        }
    });

    // Calculate alignments for all rounds
    const alignments = calculateAlignments(allRounds);

    function getMessageType(
        agentId: string,
        roundIndex: number,
    ): 'Opening' | 'Rebuttal' | undefined {
        if (!agentId) return undefined; // User rounds don't get badges
        return agentFirstAppearance.get(agentId) === roundIndex ? 'Opening' : 'Rebuttal';
    }

    const hasContent = debate.rounds.length > 0 || streamingRounds.length > 0 || isThinking;

    // Recent messages threshold - last 4 messages skip truncation
    const RECENT_THRESHOLD = 4;
    const totalRoundCount = allRounds.length;

    return (
        <div className='relative flex-1 overflow-hidden'>
            <div
                ref={scrollRef}
                className='flex h-full flex-col gap-4 overflow-y-auto p-5 scrollbar-thin'
            >
                {hasContent ? null : (
                    <div className='flex min-h-56 items-center justify-center text-[13px] text-text-3'>
                        Click start to see the debate
                    </div>
                )}

                {/* Persisted rounds from loader */}
                {debate.rounds.map((round, index) => {
                    const agentEntry = agentMap.get(round.agent_id);
                    const isRecent = totalRoundCount - index <= RECENT_THRESHOLD;
                    return (
                        <RoundBubble
                            key={`round-${index}-${round.agent_id || 'user'}`}
                            round={round}
                            agent={agentEntry?.agent}
                            index={index + 1}
                            tone={getTone(round.agent_id, agentMap)}
                            messageType={getMessageType(round.agent_id, index)}
                            highlightedAgentId={highlightedAgentId}
                            alignment={alignments[index]}
                            isRecent={isRecent}
                            debateId={debate.id}
                        />
                    );
                })}

                {/* Live rounds received via SSE */}
                {streamingRounds.map((round, index) => {
                    const agentEntry = agentMap.get(round.agent_id);
                    const globalIndex = debate.rounds.length + index;
                    // Streaming rounds are always recent
                    return (
                        <RoundBubble
                            key={`streaming-${index}`}
                            round={round}
                            agent={agentEntry?.agent}
                            index={globalIndex + 1}
                            tone={getTone(round.agent_id, agentMap)}
                            messageType={getMessageType(round.agent_id, globalIndex)}
                            highlightedAgentId={highlightedAgentId}
                            alignment={alignments[globalIndex]}
                            isRecent
                            debateId={debate.id}
                        />
                    );
                })}

                {/* "Agent is thinking" skeleton between rounds */}
                {isThinking && (
                    <RoundBubble
                        round={{ agent_id: '', message: '' }}
                        agent={THINKING_AGENT}
                        index={debate.rounds.length + streamingRounds.length + 1}
                        thinking
                        tone='agent-user'
                        alignment={
                            alignments.length > 0
                                ? alignments[alignments.length - 1] === 'left'
                                    ? 'right'
                                    : 'left'
                                : 'left'
                        }
                    />
                )}
            </div>

            {autoScrollPaused ? (
                <button
                    type='button'
                    onClick={onScrollToBottom}
                    className='absolute right-4 bottom-4 flex items-center gap-1.5 rounded-full border border-border-mid bg-bg-surface px-3 py-1.5 text-xs text-text-2 shadow-md transition-colors hover:bg-bg-hover hover:text-text-1 cursor-pointer'
                    aria-label='Scroll to bottom'
                >
                    <svg
                        viewBox='0 0 24 24'
                        fill='none'
                        stroke='currentColor'
                        strokeWidth='2'
                        strokeLinecap='round'
                        strokeLinejoin='round'
                        className='w-3.5 h-3.5'
                        aria-hidden='true'
                    >
                        <line x1='12' y1='5' x2='12' y2='19' />
                        <polyline points='19 12 12 19 5 12' />
                    </svg>
                    Scroll to bottom
                </button>
            ) : null}
        </div>
    );
}
