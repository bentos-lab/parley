import { useState } from 'react';
import type { Debate } from '@/types';

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
    sseCompleted?: boolean;
    turnCount?: number;
    onTurnCountChange?: (n: number) => void;
    onStart?: () => void;
    onStop?: () => void;
    onStartNext?: () => void;
    highlightedAgentId?: string | null;
    onAgentHighlight?: (agentId: string | null) => void;
}

export function AgentPanel({
    debate,
    streamingCount = 0,
    isStreaming = false,
    sseCompleted = false,
    turnCount = 6,
    onTurnCountChange,
    onStart,
    onStop,
    onStartNext,
    highlightedAgentId,
    onAgentHighlight,
}: AgentPanelProps) {
    const [isAgentsExpanded, setIsAgentsExpanded] = useState(true);
    const totalTurns = debate.rounds.length + streamingCount;

    const agentStats = debate.agents.map((agent, index) => {
        const turns = debate.rounds.filter((r) => r.agent_id === agent.id).length;
        const share = totalTurns > 0 ? turns / totalTurns : 1 / debate.agents.length;
        const palette = agentColors[index % agentColors.length];
        return { agent, turns, share, ...palette };
    });

    return (
        <aside
            data-testid='agent-panel'
            className='hidden lg:flex w-[260px] min-w-[260px] flex-col overflow-y-auto bg-bg-panel scrollbar-thin'
        >
            {/* Agents section */}
            <div className='border-b border-border px-4 py-3.5'>
                <button
                    type='button'
                    onClick={() => setIsAgentsExpanded(!isAgentsExpanded)}
                    className='flex w-full items-center justify-between mb-2.5'
                >
                    <span className='text-[9px] font-semibold uppercase tracking-[0.08em] text-text-3'>
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

                {/* Collapsible content */}
                <div
                    className={`space-y-2 transition-all duration-200 ease-in-out overflow-hidden ${
                        isAgentsExpanded ? 'max-h-[500px] opacity-100' : 'max-h-0 opacity-0'
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
                            className={`rounded-lg border p-3 cursor-pointer transition-all ${
                                highlightedAgentId === agent.id
                                    ? 'border-2'
                                    : 'bg-bg-surface hover:border-border-hi'
                            }`}
                            style={{
                                borderColor:
                                    highlightedAgentId === agent.id ? color : 'var(--border)',
                            }}
                        >
                            <div className='flex items-center gap-[7px] mb-1.5'>
                                <span
                                    className='h-[7px] w-[7px] rounded-full shrink-0'
                                    style={{ backgroundColor: color }}
                                />
                                <span className='flex-1 text-xs font-semibold text-text-1 truncate'>
                                    {agent.name}
                                </span>
                                <span className='font-mono text-[9px] text-text-3 bg-bg-elevated px-[5px] py-px rounded-[3px]'>
                                    {turns}
                                </span>
                            </div>

                            {agent.stance && (
                                <p className='text-[11px] text-text-2 leading-[1.4]'>
                                    {agent.stance}
                                </p>
                            )}

                            {/* <div
                                className='mt-2 h-[3px] rounded-full overflow-hidden'
                                style={{ backgroundColor: 'var(--border)' }}
                            >
                                <div
                                    className='h-full rounded-full transition-[width] duration-500'
                                    style={{
                                        width: `${Math.round(share * 100)}%`,
                                        backgroundColor: color,
                                    }}
                                />
                            </div>
                            <p className='mt-1 text-[10px] text-text-3'>
                                {Math.round(share * 100)}% share
                            </p> */}
                        </div>
                    ))}
                </div>
            </div>

            {/* Session section */}
            <div className='border-b border-border px-4 py-3.5'>
                <div className='mb-2.5 text-[9px] font-semibold uppercase tracking-[0.08em] text-text-3'>
                    Session
                </div>

                <div className='space-y-[7px]'>
                    <div className='flex justify-between items-center'>
                        <span className='text-[11px] text-text-3'>Total rounds</span>
                        <span className='font-mono text-[10px] text-text-2'>{totalTurns}</span>
                    </div>
                    <div className='flex justify-between items-center'>
                        <span className='text-[11px] text-text-3'>TTS provider</span>
                        <span className='font-mono text-[10px] text-accent'>
                            {debate.ttsProvider}
                        </span>
                    </div>
                </div>
            </div>

            {/* Controls section */}
            <div className='px-4 py-3.5'>
                <div className='mb-2.5 text-[9px] font-semibold uppercase tracking-[0.08em] text-text-3'>
                    Controls
                </div>

                {/* Turn count slider */}
                <div className='mb-3'>
                    <div className='mb-1.5 flex items-center justify-between'>
                        <span className='text-[11px] text-text-3'>Rounds to generate</span>
                        <span className='font-mono text-[10px] text-text-2'>{turnCount}</span>
                    </div>
                    <input
                        type='range'
                        min={1}
                        max={20}
                        value={turnCount}
                        onChange={(e) => onTurnCountChange?.(Number(e.target.value))}
                        disabled={isStreaming}
                        className='w-full accent-accent disabled:opacity-40'
                    />
                    <div className='flex justify-between text-[9px] text-text-3 mt-0.5'>
                        <span>1</span>
                        <span>20</span>
                    </div>
                </div>

                {/* Action buttons */}
                {!isStreaming && !sseCompleted && (
                    <button
                        type='button'
                        onClick={onStart}
                        className='w-full rounded-md bg-accent px-3 py-1.5 text-xs font-medium text-white hover:opacity-90'
                    >
                        Start
                    </button>
                )}
                {isStreaming && (
                    <button
                        type='button'
                        onClick={onStop}
                        className='w-full rounded-md bg-red-600 px-3 py-1.5 text-xs font-medium text-white hover:opacity-90'
                    >
                        Stop
                    </button>
                )}
                {sseCompleted && (
                    <div className='space-y-2'>
                        <p className='text-[10px] text-text-3 text-center'>
                            Session complete — {totalTurns} turns
                        </p>
                        <button
                            type='button'
                            onClick={onStartNext}
                            className='w-full rounded-md bg-accent px-3 py-1.5 text-xs font-medium text-white hover:opacity-90'
                        >
                            Start Next
                        </button>
                    </div>
                )}
            </div>
        </aside>
    );
}
