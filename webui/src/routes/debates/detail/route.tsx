import { useEffect, useRef, useState } from 'react';
import type { LoaderFunctionArgs } from 'react-router-dom';
import { useLoaderData, useSearchParams } from 'react-router-dom';
import { AgentPanel } from '@/components/detail/AgentPanel';
import { RoundList } from '@/components/detail/RoundList';
import { SSEStatusBar } from '@/components/detail/SSEStatusBar';
import { useAutoScroll } from '@/hooks/useAutoScroll';
import { useSSE } from '@/hooks/useSSE';
import {
    DEFAULT_TURN_COUNT,
    readStoredSessionConfig,
    writeStoredSessionConfig,
} from '@/lib/debateSessionConfig';
import { getDebate } from '@/services/api/debates';
import { API_PREFIX } from '@/services/api/http';
import { DebateSchema } from '@/services/api/schemas';
import type { Debate, Round } from '@/types';

// eslint-disable-next-line react-refresh/only-export-components
export async function loader({ params }: LoaderFunctionArgs) {
    const data = await getDebate(params.debateId!);
    // Real BE may not echo `id` in the response body — inject it from the URL param
    return DebateSchema.parse({ id: params.debateId, ...data });
}

export function Component() {
    const debate = useLoaderData() as Debate;
    const [searchParams, setSearchParams] = useSearchParams();
    const roundParam = searchParams.get('round');

    // Turn count — Priority: URL param > localStorage > default
    const [activeTurnCount, setActiveTurnCount] = useState(() => {
        if (roundParam) {
            const parsed = Number.parseInt(roundParam, 10);
            if (!Number.isNaN(parsed) && parsed >= 1 && parsed <= 20) {
                return parsed;
            }
        }
        return readStoredSessionConfig(debate.id)?.turnCount ?? DEFAULT_TURN_COUNT;
    });
    const sseUrl = `${API_PREFIX}/debates/${encodeURIComponent(debate.id)}/rounds/sse?n=${activeTurnCount}`;

    // Track whether this SSE session has completed
    const [sseCompleted, setSseCompleted] = useState(false);

    // Track which agent is highlighted (null = none)
    const [highlightedAgentId, setHighlightedAgentId] = useState<string | null>(null);

    // Accumulate complete rounds received via SSE
    const [streamingRounds, setStreamingRounds] = useState<Round[]>([]);

    // Clear streaming rounds when the debate revalidates with new persisted rounds
    const prevRoundsLengthRef = useRef(debate.rounds.length);
    useEffect(() => {
        if (debate.rounds.length !== prevRoundsLengthRef.current) {
            prevRoundsLengthRef.current = debate.rounds.length;
            // eslint-disable-next-line react-hooks/set-state-in-effect
            setStreamingRounds([]);
        }
    }, [debate.rounds.length]);

    const scrollRef = useRef<HTMLDivElement>(null);
    const hasAutoStarted = useRef(false);

    const { isThinking, isStreaming, start, stop } = useSSE({
        url: sseUrl,
        onRound: (round) => {
            setStreamingRounds((prev) => [
                ...prev,
                { agent_id: round.agent_id, message: round.content },
            ]);
        },
        onDone: () => setSseCompleted(true),
    });

    const autoScroll = useAutoScroll({
        containerRef: scrollRef,
        trigger: streamingRounds.length + (isThinking ? 1 : 0),
    });

    // Auto-start SSE when redirected from create with ?round= param
    useEffect(() => {
        if (hasAutoStarted.current) return;
        if (debate.rounds.length === 0 && roundParam) {
            hasAutoStarted.current = true;
            // Clear the search param so refresh doesn't re-trigger
            setSearchParams({}, { replace: true });
            start();
        }
    }, [debate.rounds.length, roundParam, setSearchParams, start]);

    function handleStartNext() {
        writeStoredSessionConfig(debate.id, { turnCount: activeTurnCount });
        setStreamingRounds([]);
        setSseCompleted(false);
        // start() picks up the latest sseUrl via callbacksRef
        start();
    }

    return (
        <section className='flex h-full overflow-hidden bg-bg-base'>
            {/* Rounds column */}
            <div className='flex flex-1 flex-col overflow-hidden border-r border-border'>
                <SSEStatusBar
                    activeRoute={sseUrl}
                    statusText={isThinking ? 'Agent is thinking…' : ''}
                />

                <RoundList
                    debate={debate}
                    scrollRef={scrollRef}
                    streamingRounds={streamingRounds}
                    isThinking={isThinking}
                    autoScrollPaused={autoScroll.paused}
                    onScrollToBottom={autoScroll.resumeNow}
                    highlightedAgentId={highlightedAgentId}
                />
            </div>

            {/* Right side panel */}
            <AgentPanel
                debate={debate}
                streamingCount={streamingRounds.length}
                isStreaming={isStreaming}
                sseCompleted={sseCompleted}
                turnCount={activeTurnCount}
                onTurnCountChange={setActiveTurnCount}
                onStart={start}
                onStop={stop}
                onStartNext={handleStartNext}
                highlightedAgentId={highlightedAgentId}
                onAgentHighlight={setHighlightedAgentId}
            />
        </section>
    );
}

export { DetailErrorBoundary as ErrorBoundary } from './ErrorBoundary';
